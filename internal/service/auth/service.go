package auth

import (
	"context"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/model"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/repository"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/service/cache"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/service/security"
	"github.com/google/uuid"
)

type Service struct {
	users        repository.UserRepository
	sessions     repository.SessionRepository
	accessTokens repository.AccessTokenRepository
	revoked      repository.RevokedTokenRepository
	userRoles    interface {
		AssignRole(userID string, roleID int) error
		RemoveRole(userID string, roleID int) error
		RemoveAll(userID string) error
		GetUserRoles(userID string) ([]model.UserRole, error)
		HasRole(userID string, roleID int) (bool, error)
	}
	jwt         security.JWTManager
	hasher      security.PasswordHasher
	cache       cache.Cache
	issuer      string
	audienceTTL time.Duration
	accessTTL   time.Duration
	refreshTTL  time.Duration
}

func NewService(
	users repository.UserRepository,
	sessions repository.SessionRepository,
	accessTokens repository.AccessTokenRepository,
	revoked repository.RevokedTokenRepository,
	userRoles interface {
		AssignRole(userID string, roleID int) error
		RemoveRole(userID string, roleID int) error
		RemoveAll(userID string) error
		GetUserRoles(userID string) ([]model.UserRole, error)
		HasRole(userID string, roleID int) (bool, error)
	},
	jwt security.JWTManager,
	hasher security.PasswordHasher,
	cache cache.Cache,
	issuer string,
) *Service {
	return &Service{
		users:        users,
		sessions:     sessions,
		accessTokens: accessTokens,
		revoked:      revoked,
		userRoles:    userRoles,
		jwt:          jwt,
		hasher:       hasher,
		cache:        cache,
		issuer:       issuer,
		accessTTL:    15 * time.Minute,
		refreshTTL:   30 * 24 * time.Hour,
	}
}

func (s *Service) Register(ctx context.Context, in RegisterInput) (*model.User, error) {
	hash, err := s.hasher.Hash([]byte(in.Password))
	if err != nil {
		return nil, err
	}
	u := &model.User{
		Email:        in.Email,
		PasswordHash: hash,
		IsActive:     true,
	}
	if err := s.users.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Service) Login(ctx context.Context, in LoginInput) (*LoginResult, error) {
	u, err := s.users.GetByEmail(ctx, in.Email)
	if err != nil {
		return nil, err
	}
	if !u.IsActive {
		return nil, errors.New("inactive")
	}
	if err := security.VerifyOrError(s.hasher, []byte(in.Password), u.PasswordHash); err != nil {
		return nil, err
	}
	roleRecords, err := s.userRoles.GetUserRoles(u.ID)
	if err != nil {
		return nil, err
	}
	roles := make([]int, 0, len(roleRecords))
	for _, r := range roleRecords {
		roles = append(roles, r.RoleID)
	}
	sid := uuid.NewString()
	secret, err := security.RandomBytes(32)
	if err != nil {
		return nil, err
	}
	refreshOpaque := sid + "." + base64.RawURLEncoding.EncodeToString(secret)
	refreshHash, err := s.hasher.Hash([]byte(refreshOpaque))
	if err != nil {
		return nil, err
	}
	session := &model.Session{
		ID:               sid,
		UserID:           u.ID,
		RefreshTokenHash: refreshHash,
		UserAgent:        in.UserAgent,
		IPAddress:        in.IP,
		ExpiresAt:        time.Now().Add(s.refreshTTL),
	}
	if err := s.sessions.Create(ctx, session); err != nil {
		return nil, err
	}
	claims := security.Claims{
		Sub:     u.ID,
		Aud:     in.ClientID,
		Scope:   in.Scopes,
		Roles:   roles,
		Iss:     s.issuer,
		Client:  in.ClientID,
		Session: session.ID,
	}
	access, exp, err := s.jwt.SignAccessToken(claims)
	if err != nil {
		return nil, err
	}
	rec := &model.AccessTokenRecord{
		TokenID:   uuid.NewString(),
		ClientID:  in.ClientID,
		UserID:    u.ID,
		Scopes:    in.Scopes,
		IssuedAt:  time.Now(),
		ExpiresAt: exp,
		JTI:       claims.JTI,
	}
	if err := s.accessTokens.Record(ctx, rec); err != nil {
		return nil, err
	}
	return &LoginResult{
		AccessToken:  access,
		AccessExpiry: exp,
		RefreshToken: refreshOpaque,
		SessionID:    session.ID,
		UserID:       u.ID,
		Scopes:       in.Scopes,
		Roles:        roles,
	}, nil
}

func (s *Service) Refresh(ctx context.Context, in RefreshInput) (*RefreshResult, error) {
	parts := strings.Split(in.RefreshToken, ".")
	if len(parts) != 2 {
		return nil, errors.New("bad refresh token")
	}
	sid := parts[0]
	session, err := s.sessions.GetByID(ctx, sid)
	if err != nil {
		return nil, err
	}
	if session == nil || session.Revoked {
		return nil, errors.New("revoked")
	}
	if time.Now().After(session.ExpiresAt) {
		return nil, errors.New("expired")
	}
	if err := security.VerifyOrError(s.hasher, []byte(in.RefreshToken), session.RefreshTokenHash); err != nil {
		return nil, err
	}
	u, err := s.users.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, err
	}
	if u == nil || !u.IsActive {
		return nil, errors.New("inactive")
	}
	roleRecords, err := s.userRoles.GetUserRoles(u.ID)
	if err != nil {
		return nil, err
	}
	roles := make([]int, 0, len(roleRecords))
	for _, r := range roleRecords {
		roles = append(roles, r.RoleID)
	}
	claims := security.Claims{
		Sub:     u.ID,
		Aud:     in.ClientID,
		Scope:   []string{},
		Roles:   roles,
		Iss:     s.issuer,
		Client:  in.ClientID,
		Session: session.ID,
	}
	access, exp, err := s.jwt.SignAccessToken(claims)
	if err != nil {
		return nil, err
	}
	rec := &model.AccessTokenRecord{
		TokenID:   uuid.NewString(),
		ClientID:  in.ClientID,
		UserID:    u.ID,
		Scopes:    []string{},
		IssuedAt:  time.Now(),
		ExpiresAt: exp,
		JTI:       claims.JTI,
	}
	if err := s.accessTokens.Record(ctx, rec); err != nil {
		return nil, err
	}
	return &RefreshResult{
		AccessToken:  access,
		AccessExpiry: exp,
		SessionID:    session.ID,
	}, nil
}

func (s *Service) Logout(ctx context.Context, sessionID string) error {
	return s.sessions.Revoke(ctx, sessionID)
}

func (s *Service) ValidateToken(ctx context.Context, token string, audience string) (*ValidateResult, error) {
	claims, err := s.jwt.Validate(token, audience, s.issuer)
	if err != nil {
		return nil, err
	}
	key := "revoked:" + claims.JTI
	if s.cache != nil {
		if _, ok, _ := s.cache.Get(key); ok {
			return nil, errors.New("revoked")
		}
	}
	rv, err := s.revoked.IsRevoked(ctx, claims.JTI)
	if err != nil {
		return nil, err
	}
	if rv {
		if s.cache != nil {
			_ = s.cache.Set(key, "1", time.Hour)
		}
		return nil, errors.New("revoked")
	}
	return &ValidateResult{
		UserID: claims.Sub,
		Client: claims.Client,
		Scopes: claims.Scope,
		Roles:  claims.Roles,
		JTI:    claims.JTI,
		Exp:    claims.Exp,
	}, nil
}

func (s *Service) RevokeAccessToken(ctx context.Context, jti string, reason *string) error {
	rt := &model.RevokedToken{
		JTI:       jti,
		TokenType: "access",
		RevokedAt: time.Now(),
		Reason:    reason,
	}
	if err := s.revoked.Add(ctx, rt); err != nil {
		return err
	}
	if s.cache != nil {
		_ = s.cache.Set("revoked:"+jti, "1", 24*time.Hour)
	}
	return nil
}
