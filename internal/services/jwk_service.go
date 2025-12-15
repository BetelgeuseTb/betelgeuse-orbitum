package services

import (
	"context"
	"strconv"
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

type JWKService struct {
	db       *db.DB
	cacheMan cache.Manager
	logger   zerolog.Logger
	tracer   trace.Tracer
	cache    cache.Cache
	ttl      time.Duration
	name     string
}

func NewJWKService(dbConn *db.DB, cacheManager cache.Manager, logger zerolog.Logger) *JWKService {
	return &JWKService{
		db:       dbConn,
		cacheMan: cacheManager,
		logger:   logger,
		tracer:   otel.Tracer("service.jwk"),
		cache:    cacheManager.Cache("jwks"),
		ttl:      60 * time.Minute,
		name:     "jwks",
	}
}

func (s *JWKService) cacheKey(orbitID int64, kid string) string {
	return "orbit:" + fmtID(orbitID) + ":kid:" + kid
}

func (s *JWKService) Create(ctx context.Context, jwk *models.JWKey) (*models.JWKey, error) {
	ctx, span := s.tracer.Start(ctx, "Create")
	defer span.End()

	var created *models.JWKey
	err := s.db.WithTx(ctx, func(tx pgx.Tx) error {
		repo := repositories.NewJWKRepository(tx, s.logger)
		var err error
		created, err = repo.Create(ctx, jwk)
		return err
	})
	if err != nil {
		s.logger.Error().Err(err).Str("kid", jwk.Kid).Msg("jwk create failed")
		return nil, err
	}
	_ = s.cache.Set(ctx, s.cacheKey(created.OrbitID, created.Kid), created, s.ttl)
	return created, nil
}

func (s *JWKService) GetByID(ctx context.Context, id int64) (*models.JWKey, error) {
	ctx, span := s.tracer.Start(ctx, "GetByID")
	defer span.End()

	var jwk *models.JWKey
	err := s.db.WithTx(ctx, func(tx pgx.Tx) error {
		repo := repositories.NewJWKRepository(tx, s.logger)
		var err error
		jwk, err = repo.GetByID(ctx, id)
		return err
	})
	if err != nil {
		s.logger.Error().Err(err).Int64("jwk_id", id).Msg("jwk get by id failed")
		return nil, err
	}
	return jwk, nil
}

func (s *JWKService) GetByOrbitAndKid(ctx context.Context, orbitID int64, kid string) (*models.JWKey, error) {
	ctx, span := s.tracer.Start(ctx, "GetByOrbitAndKid")
	defer span.End()

	key := s.cacheKey(orbitID, kid)
	var cached models.JWKey
	if err := s.cache.Get(ctx, key, &cached); err == nil {
		return &cached, nil
	}

	var jwk *models.JWKey
	err := s.db.WithTx(ctx, func(tx pgx.Tx) error {
		repo := repositories.NewJWKRepository(tx, s.logger)
		var err error
		jwk, err = repo.GetByOrbitAndKid(ctx, orbitID, kid)
		return err
	})
	if err != nil {
		s.logger.Error().Err(err).Int64("orbit_id", orbitID).Str("kid", kid).Msg("jwk get by orbit/kid failed")
		return nil, err
	}
	if jwk != nil {
		_ = s.cache.Set(ctx, key, jwk, s.ttl)
	}
	return jwk, nil
}

func (s *JWKService) Update(ctx context.Context, jwk *models.JWKey) (*models.JWKey, error) {
	ctx, span := s.tracer.Start(ctx, "Update")
	defer span.End()

	var updated *models.JWKey
	err := s.db.WithTx(ctx, func(tx pgx.Tx) error {
		repo := repositories.NewJWKRepository(tx, s.logger)
		var err error
		updated, err = repo.Update(ctx, jwk)
		return err
	})
	if err != nil {
		s.logger.Error().Err(err).Int64("jwk_id", jwk.ID).Msg("jwk update failed")
		return nil, err
	}
	if updated != nil {
		_ = s.cache.Delete(ctx, s.cacheKey(updated.OrbitID, updated.Kid))
		_ = s.cache.Set(ctx, s.cacheKey(updated.OrbitID, updated.Kid), updated, s.ttl)
	}
	return updated, nil
}

func (s *JWKService) ListByOrbit(ctx context.Context, orbitID int64, limit, offset int) ([]*models.JWKey, error) {
	ctx, span := s.tracer.Start(ctx, "ListByOrbit")
	defer span.End()

	var items []*models.JWKey
	err := s.db.WithTx(ctx, func(tx pgx.Tx) error {
		repo := repositories.NewJWKRepository(tx, s.logger)
		var err error
		items, err = repo.ListByOrbit(ctx, orbitID, limit, offset)
		return err
	})
	if err != nil {
		s.logger.Error().Err(err).Int64("orbit_id", orbitID).Msg("jwk list failed")
		return nil, err
	}
	return items, nil
}

func (s *JWKService) Delete(ctx context.Context, id int64, orbitID int64) error {
	ctx, span := s.tracer.Start(ctx, "Delete")
	defer span.End()

	err := s.db.WithTx(ctx, func(tx pgx.Tx) error {
		repo := repositories.NewJWKRepository(tx, s.logger)
		return repo.Delete(ctx, id, orbitID)
	})
	if err != nil {
		s.logger.Error().Err(err).Int64("jwk_id", id).Int64("orbit_id", orbitID).Msg("jwk delete failed")
		return err
	}
	_ = s.cache.Delete(ctx, s.cacheKey(orbitID, ""))
	return nil
}

func fmtID(id int64) string {
	return strconv.FormatInt(id, 10)
}
