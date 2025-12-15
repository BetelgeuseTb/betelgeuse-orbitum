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

type RoleService struct {
	db     *db.DB
	cache  cache.Manager
	logger zerolog.Logger
	tracer trace.Tracer
	ttl    time.Duration
	prefix string
}

func NewRoleService(dbConn *db.DB, cacheManager cache.Manager, logger zerolog.Logger) *RoleService {
	return &RoleService{
		db:     dbConn,
		cache:  cacheManager,
		logger: logger,
		tracer: otel.Tracer("service.role"),
		ttl:    30 * time.Minute,
		prefix: "roles",
	}
}

func (s *RoleService) key(orbitID int64, roleID int64) string {
	return fmt.Sprintf("orbit:%d:role:%d", orbitID, roleID)
}

func (s *RoleService) Create(ctx context.Context, r *models.Role) (*models.Role, error) {
	ctx, span := s.tracer.Start(ctx, "Create")
	defer span.End()

	var created *models.Role
	err := s.db.WithTx(ctx, func(tx pgx.Tx) error {
		repo := repositories.NewRoleRepository(tx, s.logger)
		var err error
		created, err = repo.Create(ctx, r)
		return err
	})
	if err != nil {
		s.logger.Error().Err(err).Msg("role create failed")
		return nil, err
	}

	_ = s.cache.Cache(s.prefix).Set(ctx, s.key(created.OrbitID, created.ID), created, s.ttl)
	return created, nil
}

func (s *RoleService) GetByID(ctx context.Context, orbitID, roleID int64) (*models.Role, error) {
	ctx, span := s.tracer.Start(ctx, "GetByID")
	defer span.End()

	key := s.key(orbitID, roleID)
	var cached models.Role
	if err := s.cache.Cache(s.prefix).Get(ctx, key, &cached); err == nil {
		return &cached, nil
	}

	var r *models.Role
	err := s.db.WithTx(ctx, func(tx pgx.Tx) error {
		repo := repositories.NewRoleRepository(tx, s.logger)
		var err error
		r, err = repo.GetByID(ctx, roleID)
		return err
	})
	if err != nil {
		s.logger.Error().Err(err).Int64("role_id", roleID).Msg("role get failed")
		return nil, err
	}
	if r != nil {
		_ = s.cache.Cache(s.prefix).Set(ctx, key, r, s.ttl)
	}
	return r, nil
}

func (s *RoleService) Delete(ctx context.Context, orbitID, roleID int64) error {
	ctx, span := s.tracer.Start(ctx, "Delete")
	defer span.End()

	err := s.db.WithTx(ctx, func(tx pgx.Tx) error {
		repo := repositories.NewRoleRepository(tx, s.logger)
		return repo.Delete(ctx, roleID)
	})
	if err != nil {
		s.logger.Error().Err(err).Int64("role_id", roleID).Msg("role delete failed")
		return err
	}

	_ = s.cache.Cache(s.prefix).Delete(ctx, s.key(orbitID, roleID))
	return nil
}

func (s *RoleService) ListByOrbit(ctx context.Context, orbitID int64, limit, offset int) ([]*models.Role, error) {
	ctx, span := s.tracer.Start(ctx, "ListByOrbit")
	defer span.End()

	var list []*models.Role
	err := s.db.WithTx(ctx, func(tx pgx.Tx) error {
		repo := repositories.NewRoleRepository(tx, s.logger)
		var err error
		list, err = repo.ListByOrbit(ctx, orbitID, limit, offset)
		return err
	})
	if err != nil {
		s.logger.Error().Err(err).Int64("orbit_id", orbitID).Msg("role list failed")
		return nil, err
	}
	return list, nil
}
