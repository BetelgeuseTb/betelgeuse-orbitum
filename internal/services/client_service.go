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

type ClientService struct {
	db         *db.DB
	cache      cache.Manager
	logger     zerolog.Logger
	tracer     trace.Tracer
	ttl        time.Duration
	prefix     string
	introspect cache.Cache
}

func NewClientService(dbConn *db.DB, cacheManager cache.Manager, logger zerolog.Logger) *ClientService {
	return &ClientService{
		db:         dbConn,
		cache:      cacheManager,
		logger:     logger,
		tracer:     otel.Tracer("service.client"),
		ttl:        30 * time.Minute,
		prefix:     "clients",
		introspect: cacheManager.Cache("introspection"),
	}
}

func (s *ClientService) keyByClientID(orbitID int64, clientID string) string {
	return "orbit:" + strconv.FormatInt(orbitID, 10) + ":client:" + clientID
}

func (s *ClientService) Create(ctx context.Context, c *models.Client) (*models.Client, error) {
	ctx, span := s.tracer.Start(ctx, "Create")
	defer span.End()

	var created *models.Client
	err := s.db.WithTx(ctx, func(tx pgx.Tx) error {
		repo := repositories.NewClientRepository(tx, s.logger)
		var err error
		created, err = repo.Create(ctx, c)
		return err
	})
	if err != nil {
		s.logger.Error().Err(err).Str("client_id", c.ClientID).Msg("client create failed")
		return nil, err
	}
	_ = s.cache.Cache(s.prefix).Set(ctx, s.keyByClientID(created.OrbitID, created.ClientID), created, s.ttl)
	return created, nil
}

func (s *ClientService) GetByID(ctx context.Context, id int64) (*models.Client, error) {
	ctx, span := s.tracer.Start(ctx, "GetByID")
	defer span.End()

	var client *models.Client
	err := s.db.WithTx(ctx, func(tx pgx.Tx) error {
		repo := repositories.NewClientRepository(tx, s.logger)
		var err error
		client, err = repo.GetByID(ctx, id)
		return err
	})
	if err != nil {
		s.logger.Error().Err(err).Int64("client_db_id", id).Msg("client get by id failed")
		return nil, err
	}
	return client, nil
}

func (s *ClientService) GetByClientID(ctx context.Context, orbitID int64, clientID string) (*models.Client, error) {
	ctx, span := s.tracer.Start(ctx, "GetByClientID")
	defer span.End()

	key := s.keyByClientID(orbitID, clientID)
	var cached models.Client
	if err := s.cache.Cache(s.prefix).Get(ctx, key, &cached); err == nil {
		return &cached, nil
	}

	var client *models.Client
	err := s.db.WithTx(ctx, func(tx pgx.Tx) error {
		repo := repositories.NewClientRepository(tx, s.logger)
		var err error
		client, err = repo.GetByClientID(ctx, orbitID, clientID)
		return err
	})
	if err != nil {
		s.logger.Error().Err(err).Str("client_id", clientID).Msg("client get by client_id failed")
		return nil, err
	}
	if client != nil {
		_ = s.cache.Cache(s.prefix).Set(ctx, key, client, s.ttl)
	}
	return client, nil
}

func (s *ClientService) Update(ctx context.Context, c *models.Client) (*models.Client, error) {
	ctx, span := s.tracer.Start(ctx, "Update")
	defer span.End()

	var updated *models.Client
	err := s.db.WithTx(ctx, func(tx pgx.Tx) error {
		repo := repositories.NewClientRepository(tx, s.logger)
		var err error
		updated, err = repo.Update(ctx, c)
		return err
	})
	if err != nil {
		s.logger.Error().Err(err).Int64("client_db_id", c.ID).Msg("client update failed")
		return nil, err
	}
	_ = s.cache.Cache(s.prefix).Delete(ctx, s.keyByClientID(c.OrbitID, c.ClientID))
	if updated != nil {
		_ = s.cache.Cache(s.prefix).Set(ctx, s.keyByClientID(updated.OrbitID, updated.ClientID), updated, s.ttl)
	}
	_ = s.introspect.Delete(ctx, "") // best-effort: clear introspection cache
	return updated, nil
}

func (s *ClientService) Delete(ctx context.Context, id int64) error {
	ctx, span := s.tracer.Start(ctx, "Delete")
	defer span.End()

	var orbitID int64
	var clientID string
	err := s.db.WithTx(ctx, func(tx pgx.Tx) error {
		repo := repositories.NewClientRepository(tx, s.logger)
		c, err := repo.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if c == nil {
			return nil
		}
		orbitID = c.OrbitID
		clientID = c.ClientID
		return repo.Delete(ctx, id)
	})
	if err != nil {
		s.logger.Error().Err(err).Int64("client_db_id", id).Msg("client delete failed")
		return err
	}
	_ = s.cache.Cache(s.prefix).Delete(ctx, s.keyByClientID(orbitID, clientID))
	return nil
}

func (s *ClientService) ListByOrbit(ctx context.Context, orbitID int64, limit, offset int) ([]*models.Client, error) {
	ctx, span := s.tracer.Start(ctx, "ListByOrbit")
	defer span.End()

	var list []*models.Client
	err := s.db.WithTx(ctx, func(tx pgx.Tx) error {
		repo := repositories.NewClientRepository(tx, s.logger)
		var err error
		list, err = repo.ListByOrbit(ctx, orbitID, limit, offset)
		return err
	})
	if err != nil {
		s.logger.Error().Err(err).Int64("orbit_id", orbitID).Msg("client list failed")
		return nil, err
	}
	return list, nil
}
