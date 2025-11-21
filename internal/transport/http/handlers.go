package httptransport

import (
	"encoding/base64"
	"net/http"
	"strings"
	"time"

	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/model"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/repository"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/service/auth"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/service/oauthflow"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/service/security"
	api "github.com/BetelgeuseTb/betelgeuse-orbitum/pkg/api"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Handlers struct {
	AuthSvc         *auth.Service
	UserRepo        repository.UserRepository
	SessionRepo     repository.SessionRepository
	OAuthClientRepo repository.OAuthClientRepository
	AuthCodeRepo    repository.AuthorizationCodeRepository
	AccessTokenRepo repository.AccessTokenRepository
	RevokedRepo     repository.RevokedTokenRepository
	RolesRepo       interface {
		GetUserRoles(userID string) ([]model.UserRole, error)
	}
	OauthStore  *oauthflow.Store
	JWTManager  security.JWTManager
	KeyProvider security.KeyProvider
	Hasher      security.PasswordHasher
	Issuer      string
}

func (h *Handlers) RegisterUser(ctx echo.Context) error {
	var req api.RegisterUserRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, api.ErrorResponse{Error: "invalid_request"})
	}
	u, err := h.AuthSvc.Register(ctx.Request().Context(), auth.RegisterInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, api.ErrorResponse{Error: "invalid_request", ErrorDescription: err.Error()})
	}
	return ctx.JSON(http.StatusCreated, api.RegisterUserResponse{
		UserId:    u.ID,
		Email:     u.Email,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
	})
}

func (h *Handlers) RegisterClient(ctx echo.Context) error {
	var req api.RegisterClientRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, api.ErrorResponse{Error: "invalid_request"})
	}
	id := uuid.NewString()
	var secretPlain *string
	var secretHash string
	public := req.Public
	method := req.TokenEndpointAuthMethod
	if !public && method != nil && (*method == "client_secret_basic" || *method == "client_secret_post") {
		s := uuid.NewString() + "." + uuid.NewString()
		secretPlain = &s
		ph, err := h.Hasher.Hash([]byte(s))
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, api.ErrorResponse{Error: "invalid_request"})
		}
		secretHash = ph
	}
	oc := &model.OAuthClient{
		ClientID:                id,
		ClientSecretHash:        secretHash,
		ClientName:              req.ClientName,
		RedirectURIs:            req.RedirectUris,
		Scopes:                  req.Scopes,
		GrantTypes:              req.GrantTypes,
		TokenEndpointAuthMethod: strOrDefault(method, "none"),
		Public:                  public,
	}
	if err := h.OAuthClientRepo.Create(ctx.Request().Context(), oc); err != nil {
		return ctx.JSON(http.StatusBadRequest, api.ErrorResponse{Error: "invalid_request", ErrorDescription: err.Error()})
	}
	resp := api.RegisterClientResponse{
		ClientId:  id,
		CreatedAt: time.Now().Format(time.RFC3339),
	}
	if secretPlain != nil {
		resp.ClientSecret = secretPlain
	}
	return ctx.JSON(http.StatusCreated, resp)
}

func (h *Handlers) OAuthAuthorize(ctx echo.Context, params api.OauthAuthorizeParams) error {
	client, err := h.OAuthClientRepo.GetByID(ctx.Request().Context(), params.ClientId)
	if err != nil || client == nil {
		return ctx.JSON(http.StatusBadRequest, api.ErrorResponse{Error: "invalid_client"})
	}
	if !contains(client.RedirectURIs, params.RedirectUri) {
		return ctx.JSON(http.StatusBadRequest, api.ErrorResponse{Error: "invalid_request", ErrorDescription: "redirect_uri"})
	}
	if client.Public && (params.CodeChallenge == nil || params.CodeChallengeMethod == nil || *params.CodeChallengeMethod != "S256") {
		return ctx.JSON(http.StatusBadRequest, api.ErrorResponse{Error: "invalid_request", ErrorDescription: "pkce_required"})
	}
	// In a real app, this should come from authenticated session/cookie
	userID := ctx.Request().Header.Get("X-User-ID")
	if userID == "" {
		return ctx.JSON(http.StatusBadRequest, api.ErrorResponse{Error: "login_required"})
	}
	code := uuid.NewString()
	scopes := strings.Fields(params.Scope)
	meta := oauthflow.CodeMeta{
		ClientID:            params.ClientId,
		UserID:              userID,
		RedirectURI:         params.RedirectUri,
		Scopes:              scopes,
		CodeChallenge:       params.CodeChallenge,
		CodeChallengeMethod: params.CodeChallengeMethod,
		ExpiresAt:           time.Now().Add(5 * time.Minute),
	}
	if err := h.OauthStore.Save(ctx.Request().Context(), code, meta, 5*time.Minute); err != nil {
		return ctx.JSON(http.StatusBadRequest, api.ErrorResponse{Error: "server_error"})
	}
	ac := &model.AuthorizationCode{
		Code:        code,
		ClientID:    params.ClientId,
		UserID:      userID,
		RedirectURI: params.RedirectUri,
		ExpiresAt:   meta.ExpiresAt,
		Used:        false,
		CreatedAt:   time.Now(),
	}
	_ = h.AuthCodeRepo.Create(ctx.Request().Context(), ac)
	loc := params.RedirectUri + "?code=" + code + "&state=" + params.State
	return ctx.Redirect(http.StatusFound, loc)
}

func (h *Handlers) OAuthToken(ctx echo.Context) error {
	if err := ctx.Request().ParseForm(); err != nil {
		return ctx.JSON(http.StatusBadRequest, api.ErrorResponse{Error: "invalid_request"})
	}
	grant := ctx.Request().Form.Get("grant_type")
	clientID, clientSecret, basicErr := ParseBasicAuth(ctx.Request().Header.Get("Authorization"))
	if basicErr != nil {
		clientID = ctx.Request().Form.Get("client_id")
		clientSecret = ctx.Request().Form.Get("client_secret")
	}
	client, err := h.OAuthClientRepo.GetByID(ctx.Request().Context(), clientID)
	if err != nil || client == nil {
		return ctx.JSON(http.StatusUnauthorized, api.ErrorResponse{Error: "invalid_client"})
	}
	if client.TokenEndpointAuthMethod != "none" {
		if clientSecret == "" {
			return ctx.JSON(http.StatusUnauthorized, api.ErrorResponse{Error: "invalid_client"})
		}
		if err := security.VerifyOrError(h.Hasher, []byte(clientSecret), client.ClientSecretHash); err != nil {
			return ctx.JSON(http.StatusUnauthorized, api.ErrorResponse{Error: "invalid_client"})
		}
	}

	switch grant {
	case "authorization_code":
		code := ctx.Request().Form.Get("code")
		redirectURI := ctx.Request().Form.Get("redirect_uri")
		codeVerifier := ctx.Request().Form.Get("code_verifier")
		meta, ok, err := h.OauthStore.Get(ctx.Request().Context(), code)
		if err != nil || !ok {
			return ctx.JSON(http.StatusBadRequest, api.ErrorResponse{Error: "invalid_grant"})
		}
		if meta.ClientID != clientID || meta.RedirectURI != redirectURI || time.Now().After(meta.ExpiresAt) {
			return ctx.JSON(http.StatusBadRequest, api.ErrorResponse{Error: "invalid_grant"})
		}
		if meta.CodeChallenge != nil {
			if codeVerifier == "" {
				return ctx.JSON(http.StatusBadRequest, api.ErrorResponse{Error: "invalid_grant", ErrorDescription: "missing_code_verifier"})
			}
			if hmac := PKCEChallengeS256(codeVerifier); hmac != *meta.CodeChallenge {
				return ctx.JSON(http.StatusBadRequest, api.ErrorResponse{Error: "invalid_grant"})
			}
		}
		u, err := h.UserRepo.GetByID(ctx.Request().Context(), meta.UserID)
		if err != nil || u == nil || !u.IsActive {
			return ctx.JSON(http.StatusBadRequest, api.ErrorResponse{Error: "invalid_request"})
		}
		roleRecords, err := h.RolesRepo.GetUserRoles(u.ID)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, api.ErrorResponse{Error: "server_error"})
		}
		roles := make([]int, 0, len(roleRecords))
		for _, r := range roleRecords {
			roles = append(roles, r.RoleID)
		}
		claims := security.Claims{
			Sub:    u.ID,
			Aud:    clientID,
			Scope:  meta.Scopes,
			Roles:  roles,
			Iss:    h.Issuer,
			Client: clientID,
		}
		token, exp, err := h.JWTManager.SignAccessToken(claims)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, api.ErrorResponse{Error: "server_error"})
		}
		rec := &model.AccessTokenRecord{
			TokenID:   uuid.NewString(),
			ClientID:  clientID,
			UserID:    u.ID,
			Scopes:    meta.Scopes,
			IssuedAt:  time.Now(),
			ExpiresAt: exp,
			JTI:       claims.JTI,
		}
		_ = h.AccessTokenRepo.Record(ctx.Request().Context(), rec)
		_ = h.AuthCodeRepo.MarkUsed(ctx.Request().Context(), code)
		_ = h.OauthStore.Delete(ctx.Request().Context(), code)
		return ctx.JSON(http.StatusOK, api.TokenResponse{
			AccessToken: token,
			TokenType:   "Bearer",
			ExpiresIn:   int32(exp.Sub(time.Now()).Seconds()),
			Scope:       strings.Join(meta.Scopes, " "),
		})

	case "client_credentials":
		scopeStr := ctx.Request().Form.Get("scope")
		scopes := strings.Fields(scopeStr)
		claims := security.Claims{
			Sub:    clientID,
			Aud:    clientID,
			Scope:  scopes,
			Roles:  []int{},
			Iss:    h.Issuer,
			Client: clientID,
		}
		token, exp, err := h.JWTManager.SignAccessToken(claims)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, api.ErrorResponse{Error: "server_error"})
		}
		rec := &model.AccessTokenRecord{
			TokenID:   uuid.NewString(),
			ClientID:  clientID,
			Scopes:    scopes,
			IssuedAt:  time.Now(),
			ExpiresAt: exp,
			JTI:       claims.JTI,
		}
		_ = h.AccessTokenRepo.Record(ctx.Request().Context(), rec)
		return ctx.JSON(http.StatusOK, api.TokenResponse{
			AccessToken: token,
			TokenType:   "Bearer",
			ExpiresIn:   int32(exp.Sub(time.Now()).Seconds()),
			Scope:       scopeStr,
		})

	case "refresh_token":
		refresh := ctx.Request().Form.Get("refresh_token")
		res, err := h.AuthSvc.Refresh(ctx.Request().Context(), auth.RefreshInput{
			RefreshToken: refresh,
			ClientID:     clientID,
			UserAgent:    ctx.Request().UserAgent(),
			IP:           ctx.RealIP(),
		})
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, api.ErrorResponse{Error: "invalid_grant"})
		}
		return ctx.JSON(http.StatusOK, api.TokenResponse{
			AccessToken: res.AccessToken,
			TokenType:   "Bearer",
			ExpiresIn:   int32(res.AccessExpiry.Sub(time.Now()).Seconds()),
		})

	default:
		return ctx.JSON(http.StatusBadRequest, api.ErrorResponse{Error: "unsupported_grant_type"})
	}
}

func (h *Handlers) OAuthIntrospect(ctx echo.Context) error {
	if err := ctx.Request().ParseForm(); err != nil {
		return ctx.JSON(http.StatusBadRequest, api.ErrorResponse{Error: "invalid_request"})
	}
	clientID, clientSecret, err := ParseBasicAuth(ctx.Request().Header.Get("Authorization"))
	if err != nil {
		return ctx.JSON(http.StatusUnauthorized, api.ErrorResponse{Error: "invalid_client"})
	}
	client, err := h.OAuthClientRepo.GetByID(ctx.Request().Context(), clientID)
	if err != nil || client == nil {
		return ctx.JSON(http.StatusUnauthorized, api.ErrorResponse{Error: "invalid_client"})
	}
	if err := security.VerifyOrError(h.Hasher, []byte(clientSecret), client.ClientSecretHash); err != nil {
		return ctx.JSON(http.StatusUnauthorized, api.ErrorResponse{Error: "invalid_client"})
	}
	token := ctx.Request().Form.Get("token")
	hint := ctx.Request().Form.Get("token_type_hint")
	if hint == "refresh_token" && strings.Contains(token, ".") {
		parts := strings.Split(token, ".")
		if len(parts) == 2 {
			sessionID := parts[0]
			sess, _ := h.SessionRepo.GetByID(ctx.Request().Context(), sessionID)
			active := sess != nil && !sess.Revoked && time.Now().Before(sess.ExpiresAt)
			return ctx.JSON(http.StatusOK, api.IntrospectionResponse{
				Active:    active,
				TokenType: strPtr("refresh_token"),
				ClientId:  &clientID,
				Exp:       int64OrNil(active, sess.ExpiresAt.Unix()),
				Sub:       strPtrOrNil(active, sess.UserID),
				Iss:       strPtr(h.Issuer),
			})
		}
	}
	vr, err := h.AuthSvc.ValidateToken(ctx.Request().Context(), token, clientID)
	if err != nil {
		return ctx.JSON(http.StatusOK, api.IntrospectionResponse{Active: false})
	}
	return ctx.JSON(http.StatusOK, api.IntrospectionResponse{
		Active:    true,
		Scope:     strPtr(strings.Join(vr.Scopes, " ")),
		ClientId:  &vr.Client,
		TokenType: strPtr("access_token"),
		Exp:       &vr.Exp,
		Iat:       int64Ptr(vr.Exp - 900),
		Sub:       &vr.UserID,
		Aud:       &[]string{vr.Client},
		Iss:       strPtr(h.Issuer),
		Jti:       &vr.JTI,
	})
}

func (h *Handlers) OAuthRevoke(ctx echo.Context) error {
	if err := ctx.Request().ParseForm(); err != nil {
		return ctx.JSON(http.StatusBadRequest, api.ErrorResponse{Error: "invalid_request"})
	}
	clientID, clientSecret, err := ParseBasicAuth(ctx.Request().Header.Get("Authorization"))
	if err != nil {
		return ctx.JSON(http.StatusUnauthorized, api.ErrorResponse{Error: "invalid_client"})
	}
	client, err := h.OAuthClientRepo.GetByID(ctx.Request().Context(), clientID)
	if err != nil || client == nil {
		return ctx.JSON(http.StatusUnauthorized, api.ErrorResponse{Error: "invalid_client"})
	}
	if err := security.VerifyOrError(h.Hasher, []byte(clientSecret), client.ClientSecretHash); err != nil {
		return ctx.JSON(http.StatusUnauthorized, api.ErrorResponse{Error: "invalid_client"})
	}
	token := ctx.Request().Form.Get("token")
	hint := ctx.Request().Form.Get("token_type_hint")
	if hint == "refresh_token" && strings.Contains(token, ".") {
		parts := strings.Split(token, ".")
		sessionID := parts[0]
		_ = h.AuthSvc.Logout(ctx.Request().Context(), sessionID)
		return ctx.NoContent(http.StatusOK)
	}
	vr, err := h.AuthSvc.ValidateToken(ctx.Request().Context(), token, clientID)
	if err == nil {
		_ = h.AuthSvc.RevokeAccessToken(ctx.Request().Context(), vr.JTI, nil)
	}
	return ctx.NoContent(http.StatusOK)
}

func (h *Handlers) OAuthJwks(ctx echo.Context) error {
	pub := h.KeyProvider.PrivateKey().PublicKey
	n := base64url(pub.N.Bytes())
	e := base64url(intToBytes(pub.E))
	resp := api.JwksResponse{
		Keys: []api.Jwk{{
			Kty: "RSA",
			Use: strPtr("sig"),
			Kid: h.KeyProvider.KID(),
			Alg: "RS256",
			N:   n,
			E:   e,
		}},
	}
	return ctx.JSON(http.StatusOK, resp)
}

func (h *Handlers) OAuthUserinfo(ctx echo.Context) error {
	token, err := ParseBearer(ctx.Request())
	if err != nil {
		return ctx.JSON(http.StatusUnauthorized, api.ErrorResponse{Error: "invalid_token"})
	}
	vr, err := h.AuthSvc.ValidateToken(ctx.Request().Context(), token, "")
	if err != nil {
		return ctx.JSON(http.StatusUnauthorized, api.ErrorResponse{Error: "invalid_token"})
	}
	u, err := h.UserRepo.GetByID(ctx.Request().Context(), vr.UserID)
	if err != nil || u == nil {
		return ctx.JSON(http.StatusUnauthorized, api.ErrorResponse{Error: "invalid_token"})
	}
	return ctx.JSON(http.StatusOK, api.UserInfo{
		Sub:           u.ID,
		Email:         &u.Email,
		EmailVerified: boolPtr(true),
	})
}

// Adapter and helpers

var _ api.ServerInterface = (*Adapter)(nil)

type Adapter struct {
	h *Handlers
}

func NewAdapter(h *Handlers) *Adapter { return &Adapter{h: h} }

func (a *Adapter) PostClients(ctx echo.Context) error { return a.h.RegisterClient(ctx) }
func (a *Adapter) GetOauthAuthorize(ctx echo.Context, params api.OauthAuthorizeParams) error {
	return a.h.OAuthAuthorize(ctx, params)
}
func (a *Adapter) PostOauthToken(ctx echo.Context) error      { return a.h.OAuthToken(ctx) }
func (a *Adapter) PostOauthIntrospect(ctx echo.Context) error { return a.h.OAuthIntrospect(ctx) }
func (a *Adapter) PostOauthRevoke(ctx echo.Context) error     { return a.h.OAuthRevoke(ctx) }
func (a *Adapter) GetOauthJwks(ctx echo.Context) error        { return a.h.OAuthJwks(ctx) }
func (a *Adapter) GetOauthUserinfo(ctx echo.Context) error    { return a.h.OAuthUserinfo(ctx) }
func (a *Adapter) PostUsers(ctx echo.Context) error           { return a.h.RegisterUser(ctx) }

func strOrDefault(p *string, def string) string {
	if p == nil || *p == "" {
		return def
	}
	return *p
}

func contains(arr []string, v string) bool {
	for _, s := range arr {
		if s == v {
			return true
		}
	}
	return false
}

func base64url(b []byte) string { return base64.RawURLEncoding.EncodeToString(b) }

func intToBytes(e int) []byte {
	if e == 0 {
		return []byte{0}
	}
	var out []byte
	for e > 0 {
		out = append([]byte{byte(e & 0xff)}, out...)
		e >>= 8
	}
	return out
}

func strPtr(s string) *string { return &s }
func boolPtr(b bool) *bool    { return &b }
func int64Ptr(i int64) *int64 { return &i }

func strPtrOrNil(active bool, v string) *string {
	if !active || v == "" {
		return nil
	}
	return &v
}

func int64OrNil(active bool, v int64) *int64 {
	if !active {
		return nil
	}
	return &v
}
