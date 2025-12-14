package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/models"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/repositories/common"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/repositories/postgres/sql_scripts"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/utils/logger"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthCodeRepository struct {
	pool   *pgxpool.Pool
	logger *logger.Logger
}

func NewAuthCodeRepository(pool *pgxpool.Pool, l *logger.Logger) *AuthCodeRepository {
	l.Info("auth code repository initialized")
	return &AuthCodeRepository{
		pool:   pool,
		logger: l,
	}
}

func (r *AuthCodeRepository) Create(ctx context.Context, ac *models.AuthCode) (*models.AuthCode, error) {
	r.logger.Trace(fmt.Sprintf("auth code create start orbit=%d client=%d", ac.OrbitID, ac.ClientID))

	now := time.Now().UTC()
	if ac.CreatedAt.IsZero() {
		ac.CreatedAt = now
	}
	if ac.ExpiresAt.IsZero() {
		return nil, fmt.Errorf("expires_at is required")
	}

	row := r.pool.QueryRow(
		ctx,
		sql_scripts.InsertAuthCode,
		ac.Code,
		ac.OrbitID,
		ac.ClientID,
		ac.UserID,
		ac.RedirectURI,
		ac.Scope,
		ac.CodeChallenge,
		ac.CodeChallengeMethod,
		ac.Used,
		ac.Metadata,
		ac.CreatedAt,
		ac.ExpiresAt,
	)

	var created models.AuthCode
	if err := row.Scan(&created.ID, &created.CreatedAt, &created.ExpiresAt); err != nil {
		r.logger.Error(fmt.Sprintf("auth code insert failed: %v", err))
		return nil, fmt.Errorf("insert auth code: %w", err)
	}

	created.Code = ac.Code
	created.OrbitID = ac.OrbitID
	created.ClientID = ac.ClientID
	created.UserID = ac.UserID
	created.RedirectURI = ac.RedirectURI
	created.Scope = ac.Scope
	created.CodeChallenge = ac.CodeChallenge
	created.CodeChallengeMethod = ac.CodeChallengeMethod
	created.Used = ac.Used
	created.Metadata = ac.Metadata

	r.logger.Debug(fmt.Sprintf("auth code created id=%d code=%s", created.ID, created.Code))
	return &created, nil
}

func (r *AuthCodeRepository) GetByID(ctx context.Context, id int64) (*models.AuthCode, error) {
	r.logger.Trace(fmt.Sprintf("auth code get by id=%d", id))

	row := r.pool.QueryRow(ctx, sql_scripts.SelectAuthCodeByID, id)
	ac, err := scanAuthCode(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, common.ErrNotFound
		}
		r.logger.Error(fmt.Sprintf("get auth code by id failed: %v", err))
		return nil, fmt.Errorf("get auth code by id: %w", err)
	}
	return ac, nil
}

func (r *AuthCodeRepository) GetByCode(ctx context.Context, code string) (*models.AuthCode, error) {
	r.logger.Trace(fmt.Sprintf("auth code get by code=%s", code))

	row := r.pool.QueryRow(ctx, sql_scripts.SelectAuthCodeByCode, code)
	ac, err := scanAuthCode(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, common.ErrNotFound
		}
		r.logger.Error(fmt.Sprintf("get auth code by code failed: %v", err))
		return nil, fmt.Errorf("get auth code by code: %w", err)
	}
	return ac, nil
}

func (r *AuthCodeRepository) Update(ctx context.Context, ac *models.AuthCode) error {
	r.logger.Trace(fmt.Sprintf("auth code update started id=%d", ac.ID))
	// auth_codes table in current migrations is essentially immutable aside from 'used' flag.
	// If you later add updateable fields — implement here.
	return nil
}

func (r *AuthCodeRepository) Delete(ctx context.Context, id int64) error {
	r.logger.Trace(fmt.Sprintf("auth code delete started id=%d", id))
	// auth codes are short-lived; physical delete can be implemented as needed.
	return nil
}

func (r *AuthCodeRepository) MarkUsed(ctx context.Context, code string) error {
	r.logger.Trace(fmt.Sprintf("auth code mark used code=%s", code))

	row := r.pool.QueryRow(ctx, sql_scripts.UpdateAuthCodeSetUsedByCode, code)
	var id int64
	if err := row.Scan(&id); err != nil {
		if err == pgx.ErrNoRows {
			return common.ErrNotFound
		}
		r.logger.Error(fmt.Sprintf("mark auth code used failed: %v", err))
		return fmt.Errorf("mark auth code used: %w", err)
	}

	r.logger.Debug(fmt.Sprintf("auth code marked used id=%d code=%s", id, code))
	return nil
}

func scanAuthCode(row pgx.Row) (*models.AuthCode, error) {
	var ac models.AuthCode
	err := row.Scan(
		&ac.ID,
		&ac.Code,
		&ac.OrbitID,
		&ac.ClientID,
		&ac.UserID,
		&ac.RedirectURI,
		&ac.Scope,
		&ac.CodeChallenge,
		&ac.CodeChallengeMethod,
		&ac.Used,
		&ac.Metadata,
		&ac.CreatedAt,
		&ac.ExpiresAt,
	)
	if err != nil {
		return nil, err
	}
	return &ac, nil
}
