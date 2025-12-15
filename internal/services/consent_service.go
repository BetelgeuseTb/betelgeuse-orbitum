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

type ConsentService struct {
	db     *db.DB
	cache  cache.Manager
	logger zerolog.Logger
	tracer trace.Tracer
	ttl    time.Duration
	prefix string
}

func NewConsentService(dbConn *db.DB, cacheManager cache.Manager, logger zerolog.Logger) *ConsentService {
	return &ConsentService{
		db:     dbConn,
		cache:  cacheManager,
		logger: logger,
		tracer: otel.Tracer("service.consent"),
		ttl:    10 * time.Minute,
		prefix: "consent",
	}
}

func (s *ConsentService) key(orbitID, userID, clientID int64) string {
	return "orbit:" + strconv.FormatInt(orbitID, 10) + ":user:" + strconv.FormatInt(userID, 10) + ":client:" + strconv.FormatInt(clientID, 10)
}

func (s *ConsentService) Create(ctx context.Context, c *models.Consent) (*models.Consent, error) {
	ctx, span := s.tracer.Start(ctx, "Create")
	defer span.End()

	var created *models.Consent
	err := s.db.WithTx(ctx, func(tx pgx.Tx) error {
		repo := repositories.NewConsentRepository(tx, s.logger)
		var err error
		created, err = repo.Create(ctx, c)
		return err
	})
	if err != nil {
		s.logger.Error().Err(err).Msg("consent create failed")
		return nil, err
	}
	_ = s.cache.Cache(s.prefix).Set(ctx, s.key(created.OrbitID, created.UserID, created.ClientID), created, s.ttl)
	return created, nil
}

func (s *ConsentService) Get(ctx context.Context, orbitID, userID, clientID int64) (*models.Consent, error) {
	ctx, span := s.tracer.Start(ctx, "Get")
	defer span.End()

	key := s.key(orbitID, userID, clientID)
	var cached models.Consent
	if err := s.cache.Cache(s.prefix).Get(ctx, key, &cached); err == nil {
		return &cached, nil
	}

	var c *models.Consent
	err := s.db.WithTx(ctx, func(tx pgx.Tx) error {
		repo := repositories.NewConsentRepository(tx, s.logger)
		var err error
		c, err = repo.Get(ctx, orbitID, userID, clientID)
		return err
	})
	if err != nil {
		s.logger.Error().Err(err).Msg("consent get failed")
		return nil, err
	}
	if c != nil {
		_ = s.cache.Cache(s.prefix).Set(ctx, key, c, s.ttl)
	}
	return c, nil
}

func (s *ConsentService) Revoke(ctx context.Context, consentID int64) error {
	ctx, span := s.tracer.Start(ctx, "Revoke")
	defer span.End()

	err := s.db.WithTx(ctx, func(tx pgx.Tx) error {
		repo := repositories.NewConsentRepository(tx, s.logger)
		var err error
		_, err = repo.Get(ctx, 0, 0, 0)
		_ = err
		return repo.Revoke(ctx, consentID)
	})
	if err != nil {
		s.logger.Error().Err(err).Int64("consent_id", consentID).Msg("consent revoke failed")
		return err
	}
	_ = s.cache.Cache(s.prefix).Delete(ctx, "")
	return nil
}
