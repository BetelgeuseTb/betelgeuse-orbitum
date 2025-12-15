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

type RefreshTokenService struct {
	db         *db.DB
	cacheMan   cache.Manager
	logger     zerolog.Logger
	tracer     trace.Tracer
	cacheName  string
	tokenTTL   time.Duration
	introspect cache.Cache
}

func NewRefreshTokenService(dbConn *db.DB, cacheManager cache.Manager, logger zerolog.Logger) *RefreshTokenService {
	return &RefreshTokenService{
		db:         dbConn,
		cacheMan:   cacheManager,
		logger:     logger,
		tracer:     otel.Tracer("service.refresh_token"),
		cacheName:  "refresh_tokens",
		tokenTTL:   24 * time.Hour,
		introspect: cacheManager.Cache("introspection"),
	}
}

func (s *RefreshTokenService) Rotate(ctx context.Context, oldRefreshID int64, newRefresh *models.RefreshToken, newAccess *models.AccessToken) (*models.RefreshToken, *models.AccessToken, error) {
	ctx, span := s.tracer.Start(ctx, "Rotate")
	defer span.End()

	var createdRefresh *models.RefreshToken
	var createdAccess *models.AccessToken
	err := s.db.WithTx(ctx, func(tx pgx.Tx) error {
		rtRepo := repositories.NewRefreshTokenRepository(tx, s.logger)
		atRepo := repositories.NewAccessTokenRepository(tx, s.logger)

		cr, err := rtRepo.Create(ctx, newRefresh)
		if err != nil {
			return err
		}
		createdRefresh = cr

		if err := rtRepo.Rotate(ctx, oldRefreshID, createdRefresh.ID); err != nil {
			return err
		}

		ca, err := atRepo.Create(ctx, newAccess)
		if err != nil {
			return err
		}
		createdAccess = ca

		return nil
	})
	if err != nil {
		s.logger.Error().Err(err).Int64("old_refresh_id", oldRefreshID).Msg("refresh token rotate failed")
		return nil, nil, err
	}

	_ = s.cacheMan.Cache(s.cacheName).Delete(ctx, createdRefresh.JTI)
	_ = s.introspect.Delete(ctx, createdAccess.JTI)
	return createdRefresh, createdAccess, nil
}

func (s *RefreshTokenService) TouchUsage(ctx context.Context, tokenID int64) error {
	ctx, span := s.tracer.Start(ctx, "TouchUsage")
	defer span.End()

	return s.db.WithTx(ctx, func(tx pgx.Tx) error {
		rtRepo := repositories.NewRefreshTokenRepository(tx, s.logger)
		now := time.Now().UTC()
		_, err := rtRepo.Update(ctx, &models.RefreshToken{
			ID:         tokenID,
			LastUsedAt: &now,
		})
		return err
	})
}
