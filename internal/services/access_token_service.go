package services

import (
	"context"
	"time"

	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/models"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/repositories"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/repositories/db"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/services/cache"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type AccessTokenService struct {
	db               *db.DB
	cacheMan         cache.Manager
	logger           zerolog.Logger
	tracer           trace.Tracer
	introspection    cache.Cache
	revocationName   string
	introspectionTTL time.Duration
}

func NewAccessTokenService(dbConn *db.DB, cacheManager cache.Manager, logger zerolog.Logger) *AccessTokenService {
	return &AccessTokenService{
		db:               dbConn,
		cacheMan:         cacheManager,
		logger:           logger,
		tracer:           otel.Tracer("service.access_token"),
		introspection:    cacheManager.Cache("introspection"),
		revocationName:   "revoked_tokens",
		introspectionTTL: 30 * time.Second,
	}
}

func (s *AccessTokenService) Issue(ctx context.Context, token *models.AccessToken) (*models.AccessToken, error) {
	ctx, span := s.tracer.Start(ctx, "Issue")
	defer span.End()

	var created *models.AccessToken
	err := s.db.WithTx(ctx, func(tx pgx.Tx) error {
		atRepo := repositories.NewAccessTokenRepository(tx, s.logger)
		var err error
		created, err = atRepo.Create(ctx, token)
		return err
	})
	if err != nil {
		s.logger.Error().Err(err).Str("jti", token.JTI).Msg("issue access token failed")
		return nil, err
	}
	_ = s.introspection.Delete(ctx, token.JTI)
	return created, nil
}

func (s *AccessTokenService) Revoke(ctx context.Context, orbitID int64, jti string, reason string) error {
	ctx, span := s.tracer.Start(ctx, "Revoke")
	defer span.End()

	err := s.db.WithTx(ctx, func(tx pgx.Tx) error {
		atRepo := repositories.NewAccessTokenRepository(tx, s.logger)
		rtRepo := repositories.NewRevokedTokenRepository(tx, s.logger)
		trRepo := repositories.NewTokenRevocationRepository(tx, s.logger)

		revoked, err := atRepo.RevokeByJTI(ctx, jti)
		if err != nil {
			return err
		}
		if !revoked {
			return nil
		}

		_, err = rtRepo.Create(ctx, &models.RevokedToken{
			JTI:       jti,
			OrbitID:   orbitID,
			ExpiresAt: time.Now().Add(24 * time.Hour),
			Reason:    reason,
		})
		if err != nil {
			return err
		}

		_, err = trRepo.Create(ctx, &models.TokenRevocation{
			OrbitID:   orbitID,
			TokenJTI:  jti,
			TokenType: "access_token",
			Reason:    reason,
			RevokedAt: time.Now().UTC(),
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		s.logger.Error().Err(err).Str("jti", jti).Msg("revoke access token failed")
		return err
	}
	_ = s.introspection.Delete(ctx, jti)
	return nil
}

func (s *AccessTokenService) Introspect(ctx context.Context, jti string) (*models.AccessToken, bool, error) {
	ctx, span := s.tracer.Start(ctx, "Introspect")
	defer span.End()

	var at models.AccessToken
	if err := s.introspection.Get(ctx, jti, &at); err == nil {
		return &at, !at.Revoked, nil
	}

	var token *models.AccessToken
	err := s.db.WithTx(ctx, func(tx pgx.Tx) error {
		atRepo := repositories.NewAccessTokenRepository(tx, s.logger)
		var err error
		token, err = atRepo.GetByJTI(ctx, jti)
		return err
	})
	if err != nil {
		s.logger.Error().Err(err).Str("jti", jti).Msg("introspection DB query failed")
		return nil, false, err
	}
	if token == nil {
		return nil, false, nil
	}
	_ = s.introspection.Set(ctx, jti, token, s.introspectionTTL)
	return token, !token.Revoked, nil
}
