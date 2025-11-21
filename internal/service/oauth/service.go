package oauth

import (
	"context"
	"errors"
	"time"

	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/model"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/repository"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/service/security"
	"github.com/google/uuid"
)

type Service struct {
	clients repository.OAuthClientRepository
	codes   repository.AuthorizationCodeRepository
	access  repository.AccessTokenRepository
	revoked repository.RevokedTokenRepository
	users   repository.UserRepository
	roles   interface {
		GetUserRoles(userID string) ([]model.UserRole, error)
	}
	jwt    security.JWTManager
	issuer string
}

func NewService(
	clients repository.OAuthClientRepository,
	codes repository.AuthorizationCodeRepository,
	access repository.AccessTokenRepository,
	revoked repository.RevokedTokenRepository,
	users repository.UserRepository,
	roles interface {
		GetUserRoles(userID string) ([]model.UserRole, error)
	},
	jwt security.JWTManager,
	issuer string,
) *Service {
	return &Service{
		clients: clients,
		codes:   codes,
		access:  access,
		revoked: revoked,
		users:   users,
		roles:   roles,
		jwt:     jwt,
		issuer:  issuer,
	}
}

type CreateCodeInput struct {
	Code                string
	ClientID            string
	UserID              string
	RedirectURI         string
	Scopes              []string
	CodeChallenge       *string
	CodeChallengeMethod *string
	Expiry              time.Time
}

type ExchangeInput struct {
	Code         string
	ClientID     string
	RedirectURI  string
	CodeVerifier *string
}

type ExchangeResult struct {
	AccessToken  string
	AccessExpiry time.Time
	JTI          string
}

func (s *Service) CreateCode(ctx context.Context, in CreateCodeInput) error {
	ac := &model.AuthorizationCode{
		Code:                in.Code,
		ClientID:            in.ClientID,
		UserID:              in.UserID,
		Scopes:              in.Scopes,
		RedirectURI:         in.RedirectURI,
		CodeChallenge:       in.CodeChallenge,
		CodeChallengeMethod: in.CodeChallengeMethod,
		ExpiresAt:           in.Expiry,
		Used:                false,
		CreatedAt:           time.Now(),
	}
	return s.codes.Create(ctx, ac)
}

func (s *Service) Exchange(ctx context.Context, in ExchangeInput) (*ExchangeResult, error) {
	ac, err := s.codes.Get(ctx, in.Code)
	if err != nil {
		return nil, err
	}
	if ac == nil || ac.Used {
		return nil, errors.New("invalid code")
	}
	if time.Now().After(ac.ExpiresAt) {
		return nil, errors.New("expired")
	}
	if ac.RedirectURI != in.RedirectURI || ac.ClientID != in.ClientID {
		return nil, errors.New("mismatch")
	}
	if ac.CodeChallenge != nil && ac.CodeChallengeMethod != nil && in.CodeVerifier != nil {
		switch *ac.CodeChallengeMethod {
		case "S256":
			// PKCE validation is expected here if persistence supports it
		default:
			return nil, errors.New("unsupported pkce")
		}
	}
	u, err := s.users.GetByID(ctx, ac.UserID)
	if err != nil {
		return nil, err
	}
	if u == nil || !u.IsActive {
		return nil, errors.New("inactive")
	}
	roleRecords, err := s.roles.GetUserRoles(u.ID)
	if err != nil {
		return nil, err
	}
	roles := make([]int, 0, len(roleRecords))
	for _, r := range roleRecords {
		roles = append(roles, r.RoleID)
	}
	claims := security.Claims{
		Sub:    u.ID,
		Aud:    ac.ClientID,
		Scope:  ac.Scopes,
		Roles:  roles,
		Iss:    s.issuer,
		Client: ac.ClientID,
	}
	token, exp, err := s.jwt.SignAccessToken(claims)
	if err != nil {
		return nil, err
	}
	rec := &model.AccessTokenRecord{
		TokenID:   uuid.NewString(),
		ClientID:  ac.ClientID,
		UserID:    u.ID,
		Scopes:    ac.Scopes,
		IssuedAt:  time.Now(),
		ExpiresAt: exp,
		JTI:       claims.JTI,
	}
	if err := s.access.Record(ctx, rec); err != nil {
		return nil, err
	}
	if err := s.codes.MarkUsed(ctx, in.Code); err != nil {
		return nil, err
	}
	return &ExchangeResult{
		AccessToken:  token,
		AccessExpiry: exp,
		JTI:          claims.JTI,
	}, nil
}
