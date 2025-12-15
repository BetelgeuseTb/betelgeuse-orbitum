package services

import (
	"context"
	"errors"
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

var ErrUserAlreadyExists = errors.New("user with given identity already exists")

type userRepoRead interface {
	GetByID(ctx context.Context, id int64) (*models.User, error)
	GetByUsername(ctx context.Context, orbitID int64, username string) (*models.User, error)
	GetByEmail(ctx context.Context, orbitID int64, email string) (*models.User, error)
	ListByOrbit(ctx context.Context, orbitID int64, limit, offset int) ([]*models.User, error)
}

type UserService struct {
	db        *db.DB
	readRepo  userRepoRead
	cacheMan  cache.Manager
	logger    zerolog.Logger
	tracer    trace.Tracer
	cacheTTL  time.Duration
	cacheName string
}

func NewUserService(dbConn *db.DB, readRepo userRepoRead, cacheManager cache.Manager, logger zerolog.Logger) *UserService {
	return &UserService{
		db:        dbConn,
		readRepo:  readRepo,
		cacheMan:  cacheManager,
		logger:    logger,
		tracer:    otel.Tracer("service.user"),
		cacheTTL:  5 * time.Minute,
		cacheName: "users",
	}
}

func (s *UserService) cacheKeyByID(id int64) string {
	return fmt.Sprintf("id:%d", id)
}

func (s *UserService) cacheKeyByIdentity(orbitID int64, identity string) string {
	return fmt.Sprintf("orbit:%d:identity:%s", orbitID, identity)
}

func (s *UserService) Create(ctx context.Context, user *models.User) (*models.User, error) {
	ctx, span := s.tracer.Start(ctx, "Create")
	defer span.End()

	var created *models.User
	err := s.db.WithTx(ctx, func(tx pgx.Tx) error {
		txRepo := repositories.NewUserRepository(tx, s.logger)

		existingByUsername, err := txRepo.GetByUsername(ctx, user.OrbitID, user.Username)
		if err != nil {
			return err
		}
		if existingByUsername != nil {
			return ErrUserAlreadyExists
		}

		existingByEmail, err := txRepo.GetByEmail(ctx, user.OrbitID, user.Email)
		if err != nil {
			return err
		}
		if existingByEmail != nil {
			return ErrUserAlreadyExists
		}

		created, err = txRepo.Create(ctx, user)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		s.logger.Error().Err(err).Str("username", user.Username).Str("email", user.Email).Msg("user create failed")
		return nil, err
	}

	cacheObj := s.cacheMan.Cache(s.cacheName)
	_ = cacheObj.Set(ctx, s.cacheKeyByID(created.ID), created, s.cacheTTL)
	_ = cacheObj.Set(ctx, s.cacheKeyByIdentity(created.OrbitID, created.Username), created, s.cacheTTL)
	s.logger.Debug().Int64("user_id", created.ID).Str("username", created.Username).Msg("user created and cached")
	return created, nil
}

func (s *UserService) GetByID(ctx context.Context, id int64) (*models.User, error) {
	ctx, span := s.tracer.Start(ctx, "GetByID")
	defer span.End()

	c := s.cacheMan.Cache(s.cacheName)
	var u models.User
	if err := c.Get(ctx, s.cacheKeyByID(id), &u); err == nil {
		s.logger.Debug().Int64("user_id", id).Msg("user cache hit by id")
		return &u, nil
	}

	user, err := s.readRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error().Err(err).Int64("user_id", id).Msg("user get by id failed")
		return nil, err
	}
	if user != nil {
		_ = c.Set(ctx, s.cacheKeyByID(id), user, s.cacheTTL)
		_ = c.Set(ctx, s.cacheKeyByIdentity(user.OrbitID, user.Username), user, s.cacheTTL)
	}
	return user, nil
}

func (s *UserService) GetByIdentity(ctx context.Context, orbitID int64, identity string) (*models.User, error) {
	ctx, span := s.tracer.Start(ctx, "GetByIdentity")
	defer span.End()

	c := s.cacheMan.Cache(s.cacheName)
	var u models.User
	key := s.cacheKeyByIdentity(orbitID, identity)
	if err := c.Get(ctx, key, &u); err == nil {
		s.logger.Debug().Str("identity", identity).Int64("orbit_id", orbitID).Msg("user cache hit by identity")
		return &u, nil
	}

	user, err := s.readRepo.GetByUsername(ctx, orbitID, identity)
	if err != nil {
		return nil, err
	}
	if user == nil {
		user, err = s.readRepo.GetByEmail(ctx, orbitID, identity)
		if err != nil {
			return nil, err
		}
	}
	if user != nil {
		_ = c.Set(ctx, key, user, s.cacheTTL)
		_ = c.Set(ctx, s.cacheKeyByID(user.ID), user, s.cacheTTL)
	}
	return user, nil
}

func (s *UserService) Update(ctx context.Context, user *models.User) (*models.User, error) {
	ctx, span := s.tracer.Start(ctx, "Update")
	defer span.End()

	var updated *models.User
	err := s.db.WithTx(ctx, func(tx pgx.Tx) error {
		txRepo := repositories.NewUserRepository(tx, s.logger)
		u, err := txRepo.Update(ctx, user)
		if err != nil {
			return err
		}
		updated = u
		return nil
	})
	if err != nil {
		s.logger.Error().Err(err).Int64("user_id", user.ID).Msg("user update failed")
		return nil, err
	}

	c := s.cacheMan.Cache(s.cacheName)
	_ = c.Delete(ctx, s.cacheKeyByID(user.ID))
	_ = c.Delete(ctx, s.cacheKeyByIdentity(user.OrbitID, user.Username))
	if updated != nil {
		_ = c.Set(ctx, s.cacheKeyByID(updated.ID), updated, s.cacheTTL)
		_ = c.Set(ctx, s.cacheKeyByIdentity(updated.OrbitID, updated.Username), updated, s.cacheTTL)
	}
	s.logger.Debug().Int64("user_id", user.ID).Msg("user updated and cache invalidated/refreshed")
	return updated, nil
}

func (s *UserService) Delete(ctx context.Context, id int64) error {
	ctx, span := s.tracer.Start(ctx, "Delete")
	defer span.End()

	var orbitID int64
	err := s.db.WithTx(ctx, func(tx pgx.Tx) error {
		txRepo := repositories.NewUserRepository(tx, s.logger)
		u, err := txRepo.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if u == nil {
			return nil
		}
		orbitID = u.OrbitID
		return txRepo.Delete(ctx, id)
	})
	if err != nil {
		s.logger.Error().Err(err).Int64("user_id", id).Msg("user delete failed")
		return err
	}

	c := s.cacheMan.Cache(s.cacheName)
	_ = c.Delete(ctx, s.cacheKeyByID(id))
	_ = c.Delete(ctx, s.cacheKeyByIdentity(orbitID, ""))
	s.logger.Info().Int64("user_id", id).Msg("user soft-deleted and cache invalidated")
	return nil
}

func (s *UserService) ListByOrbit(ctx context.Context, orbitID int64, limit, offset int) ([]*models.User, error) {
	ctx, span := s.tracer.Start(ctx, "ListByOrbit")
	defer span.End()

	users, err := s.readRepo.ListByOrbit(ctx, orbitID, limit, offset)
	if err != nil {
		s.logger.Error().Err(err).Int64("orbit_id", orbitID).Msg("list users by orbit failed")
		return nil, err
	}
	return users, nil
}
