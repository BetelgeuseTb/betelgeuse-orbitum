package services

import (
	"context"
	"fmt"
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

type PermissionService struct {
	db     *db.DB
	cache  cache.Manager
	logger zerolog.Logger
	tracer trace.Tracer
	ttl    time.Duration
	prefix string
}

func NewPermissionService(dbConn *db.DB, cacheManager cache.Manager, logger zerolog.Logger) *PermissionService {
	return &PermissionService{
		db:     dbConn,
		cache:  cacheManager,
		logger: logger,
		tracer: otel.Tracer("service.permission"),
		ttl:    30 * time.Minute,
		prefix: "permissions",
	}
}

func (s *PermissionService) key(orbitID, permissionID int64) string {
	return fmt.Sprintf("orbit:%d:perm:%d", orbitID, permissionID)
}

func (s *PermissionService) Create(ctx context.Context, p *models.Permission) (*models.Permission, error) {
	ctx, span := s.tracer.Start(ctx, "Create")
	defer span.End()

	var created *models.Permission
	err := s.db.WithTx(ctx, func(tx pgx.Tx) error {
		repo := repositories.NewPermissionRepository(tx, s.logger)
		var err error
		created, err = repo.Create(ctx, p)
		return err
	})
	if err != nil {
		s.logger.Error().Err(err).Msg("permission create failed")
		return nil, err
	}

	_ = s.cache.Cache(s.prefix).Set(ctx, s.key(created.OrbitID, created.ID), created, s.ttl)
	return created, nil
}

func (s *PermissionService) GetByID(ctx context.Context, orbitID, permissionID int64) (*models.Permission, error) {
	ctx, span := s.tracer.Start(ctx, "GetByID")
	defer span.End()

	key := s.key(orbitID, permissionID)
	var cached models.Permission
	if err := s.cache.Cache(s.prefix).Get(ctx, key, &cached); err == nil {
		return &cached, nil
	}

	var perm *models.Permission
	err := s.db.WithTx(ctx, func(tx pgx.Tx) error {
		repo := repositories.NewPermissionRepository(tx, s.logger)
		var err error
		perm, err = repo.GetByID(ctx, permissionID)
		return err
	})
	if err != nil {
		s.logger.Error().Err(err).Int64("perm_id", permissionID).Msg("permission get failed")
		return nil, err
	}
	if perm != nil {
		_ = s.cache.Cache(s.prefix).Set(ctx, key, perm, s.ttl)
	}
	return perm, nil
}

func (s *PermissionService) Delete(ctx context.Context, orbitID, permissionID int64) error {
	ctx, span := s.tracer.Start(ctx, "Delete")
	defer span.End()

	err := s.db.WithTx(ctx, func(tx pgx.Tx) error {
		repo := repositories.NewPermissionRepository(tx, s.logger)
		return repo.Delete(ctx, permissionID)
	})
	if err != nil {
		s.logger.Error().Err(err).Int64("perm_id", permissionID).Msg("permission delete failed")
		return err
	}
	_ = s.cache.Cache(s.prefix).Delete(ctx, s.key(orbitID, permissionID))
	return nil
}

func (s *PermissionService) ListByOrbit(ctx context.Context, orbitID int64, limit, offset int) ([]*models.Permission, error) {
	ctx, span := s.tracer.Start(ctx, "ListByOrbit")
	defer span.End()

	var list []*models.Permission
	err := s.db.WithTx(ctx, func(tx pgx.Tx) error {
		repo := repositories.NewPermissionRepository(tx, s.logger)
		var err error
		list, err = repo.ListByOrbit(ctx, orbitID, limit, offset)
		return err
	})
	if err != nil {
		s.logger.Error().Err(err).Int64("orbit_id", orbitID).Msg("permission list failed")
		return nil, err
	}
	return list, nil
}
