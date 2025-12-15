package services

import (
	"context"
	"errors"
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

var ErrAuthCodeNotFound = errors.New("auth code not found or expired")
var ErrAuthCodeAlreadyUsed = errors.New("auth code already used")

type AuthCodeService struct {
	db       *db.DB
	cacheMan cache.Manager
	logger   zerolog.Logger
	tracer   trace.Tracer
	ttl      time.Duration
	cacheKey string
}

func NewAuthCodeService(dbConn *db.DB, cacheManager cache.Manager, logger zerolog.Logger) *AuthCodeService {
	return &AuthCodeService{
		db:       dbConn,
		cacheMan: cacheManager,
		logger:   logger,
		tracer:   otel.Tracer("service.auth_code"),
		ttl:      30 * time.Second,
		cacheKey: "auth_code",
	}
}

func (s *AuthCodeService) cacheKeyFor(code string) string {
	return s.cacheKey + ":" + code
}

func (s *AuthCodeService) Create(ctx context.Context, ac *models.AuthCode) (*models.AuthCode, error) {
	ctx, span := s.tracer.Start(ctx, "Create")
	defer span.End()

	var created *models.AuthCode
	err := s.db.WithTx(ctx, func(tx pgx.Tx) error {
		repo := repositories.NewAuthCodeRepository(tx, s.logger)
		var err error
		created, err = repo.Create(ctx, ac)
		return err
	})
	if err != nil {
		s.logger.Error().Err(err).Msg("auth code create failed")
		return nil, err
	}
	return created, nil
}

func (s *AuthCodeService) GetByCode(ctx context.Context, code string) (*models.AuthCode, error) {
	ctx, span := s.tracer.Start(ctx, "GetByCode")
	defer span.End()

	cacheObj := s.cacheMan.Cache("volatile")
	var cached models.AuthCode
	if err := cacheObj.Get(ctx, s.cacheKeyFor(code), &cached); err == nil {
		return &cached, nil
	}

	var ac *models.AuthCode
	err := s.db.WithTx(ctx, func(tx pgx.Tx) error {
		repo := repositories.NewAuthCodeRepository(tx, s.logger)
		var err error
		ac, err = repo.GetByCode(ctx, code)
		return err
	})
	if err != nil {
		s.logger.Error().Err(err).Str("code", code).Msg("get auth code failed")
		return nil, err
	}
	if ac == nil {
		return nil, nil
	}
	_ = cacheObj.Set(ctx, s.cacheKeyFor(code), ac, s.ttl)
	return ac, nil
}

func (s *AuthCodeService) Consume(ctx context.Context, code string) (*models.AuthCode, error) {
	ctx, span := s.tracer.Start(ctx, "Consume")
	defer span.End()

	var consumed *models.AuthCode
	err := s.db.WithTx(ctx, func(tx pgx.Tx) error {
		repo := repositories.NewAuthCodeRepository(tx, s.logger)

		ac, err := repo.GetByCode(ctx, code)
		if err != nil {
			return err
		}
		if ac == nil {
			return ErrAuthCodeNotFound
		}
		if ac.Used {
			return ErrAuthCodeAlreadyUsed
		}
		ok, err := repo.SetUsedByCode(ctx, code)
		if err != nil {
			return err
		}
		if !ok {
			return ErrAuthCodeAlreadyUsed
		}
		consumed = ac
		return nil
	})
	if err != nil {
		s.logger.Error().Err(err).Str("code", code).Msg("consume auth code failed")
		return nil, err
	}
	_ = s.cacheMan.Cache("volatile").Delete(ctx, s.cacheKeyFor(code))
	return consumed, nil
}
