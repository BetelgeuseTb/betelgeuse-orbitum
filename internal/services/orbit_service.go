package services

import (
	"context"
	"fmt"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/repositories"
	"time"

	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/models"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/repositories/db"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/services/cache"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type orbitRepoRead interface {
	GetByID(ctx context.Context, id int64) (*models.Orbit, error)
	List(ctx context.Context, limit, offset int) ([]*models.Orbit, error)
}

type OrbitService struct {
	db       *db.DB
	readRepo orbitRepoRead
	cacheMan cache.Manager
	logger   zerolog.Logger
	tracer   trace.Tracer
	ttl      time.Duration
	name     string
}

func NewOrbitService(dbConn *db.DB, readRepo orbitRepoRead, cacheManager cache.Manager, logger zerolog.Logger) *OrbitService {
	return &OrbitService{
		db:       dbConn,
		readRepo: readRepo,
		cacheMan: cacheManager,
		logger:   logger,
		tracer:   otel.Tracer("service.orbit"),
		ttl:      10 * time.Minute,
		name:     "orbits",
	}
}

func (s *OrbitService) key(id int64) string {
	return fmt.Sprintf("id:%d", id)
}

func (s *OrbitService) Create(ctx context.Context, o *models.Orbit) (*models.Orbit, error) {
	ctx, span := s.tracer.Start(ctx, "Create")
	defer span.End()

	var created *models.Orbit
	err := s.db.WithTx(ctx, func(tx pgx.Tx) error {
		txRepo := repositories.NewOrbitRepository(tx, s.logger)
		var err error
		created, err = txRepo.Create(ctx, o)
		return err
	})
	if err != nil {
		s.logger.Error().Err(err).Str("name", o.Name).Msg("orbit create failed")
		return nil, err
	}

	c := s.cacheMan.Cache(s.name)
	_ = c.Set(ctx, s.key(created.ID), created, s.ttl)
	s.logger.Info().Int64("orbit_id", created.ID).Str("name", created.Name).Msg("orbit created and cached")
	return created, nil
}

func (s *OrbitService) GetByID(ctx context.Context, id int64) (*models.Orbit, error) {
	ctx, span := s.tracer.Start(ctx, "GetByID")
	defer span.End()

	c := s.cacheMan.Cache(s.name)
	var o models.Orbit
	if err := c.Get(ctx, s.key(id), &o); err == nil {
		s.logger.Debug().Int64("orbit_id", id).Msg("orbit cache hit")
		return &o, nil
	}

	orbit, err := s.readRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error().Err(err).Int64("orbit_id", id).Msg("orbit get failed")
		return nil, err
	}
	if orbit != nil {
		_ = c.Set(ctx, s.key(id), orbit, s.ttl)
	}
	return orbit, nil
}

func (s *OrbitService) Update(ctx context.Context, o *models.Orbit) (*models.Orbit, error) {
	ctx, span := s.tracer.Start(ctx, "Update")
	defer span.End()

	var updated *models.Orbit
	err := s.db.WithTx(ctx, func(tx pgx.Tx) error {
		txRepo := repositories.NewOrbitRepository(tx, s.logger)
		var err error
		updated, err = txRepo.Update(ctx, o)
		return err
	})
	if err != nil {
		s.logger.Error().Err(err).Int64("orbit_id", o.ID).Msg("orbit update failed")
		return nil, err
	}

	c := s.cacheMan.Cache(s.name)
	_ = c.Delete(ctx, s.key(o.ID))
	if updated != nil {
		_ = c.Set(ctx, s.key(updated.ID), updated, s.ttl)
	}
	s.logger.Info().Int64("orbit_id", o.ID).Msg("orbit updated and cache refreshed")
	return updated, nil
}

func (s *OrbitService) Delete(ctx context.Context, id int64) error {
	ctx, span := s.tracer.Start(ctx, "Delete")
	defer span.End()

	err := s.db.WithTx(ctx, func(tx pgx.Tx) error {
		txRepo := repositories.NewOrbitRepository(tx, s.logger)
		return txRepo.Delete(ctx, id)
	})
	if err != nil {
		s.logger.Error().Err(err).Int64("orbit_id", id).Msg("orbit delete failed")
		return err
	}

	_ = s.cacheMan.Cache(s.name).Delete(ctx, s.key(id))
	s.logger.Info().Int64("orbit_id", id).Msg("orbit soft-deleted and cache invalidated")
	return nil
}

func (s *OrbitService) List(ctx context.Context, limit, offset int) ([]*models.Orbit, error) {
	ctx, span := s.tracer.Start(ctx, "List")
	defer span.End()

	return s.readRepo.List(ctx, limit, offset)
}
