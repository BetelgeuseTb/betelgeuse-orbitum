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

type RevokedTokenRepository struct {
	pool   *pgxpool.Pool
	logger *logger.Logger
}

func NewRevokedTokenRepository(pool *pgxpool.Pool, l *logger.Logger) *RevokedTokenRepository {
	l.Info("revoked token repository initialized")
	return &RevokedTokenRepository{pool: pool, logger: l}
}

func (r *RevokedTokenRepository) Create(ctx context.Context, rt *models.RevokedToken) (*models.RevokedToken, error) {
	r.logger.Trace(fmt.Sprintf("revoked token create jti=%s", rt.JTI))

	now := time.Now().UTC()
	if rt.CreatedAt.IsZero() {
		rt.CreatedAt = now
	}

	row := r.pool.QueryRow(ctx, sql_scripts.InsertRevokedToken, rt.CreatedAt, rt.JTI, rt.ExpiresAt, rt.OrbitID, rt.Reason)
	var id int64
	var createdAt time.Time
	if err := row.Scan(&id, &createdAt); err != nil {
		r.logger.Error(fmt.Sprintf("revoked token insert failed: %v", err))
		return nil, fmt.Errorf("insert revoked token: %w", err)
	}
	rt.ID = id
	rt.CreatedAt = createdAt
	r.logger.Debug(fmt.Sprintf("revoked token created id=%d jti=%s", rt.ID, rt.JTI))
	return rt, nil
}

func (r *RevokedTokenRepository) GetByJTI(ctx context.Context, jti string) (*models.RevokedToken, error) {
	r.logger.Trace(fmt.Sprintf("get revoked token jti=%s", jti))

	row := r.pool.QueryRow(ctx, sql_scripts.SelectRevokedByJTI, jti)
	var rt models.RevokedToken
	if err := row.Scan(&rt.ID, &rt.CreatedAt, &rt.JTI, &rt.ExpiresAt, &rt.OrbitID, &rt.Reason); err != nil {
		if err == pgx.ErrNoRows {
			return nil, common.ErrNotFound
		}
		r.logger.Error(fmt.Sprintf("select revoked token failed: %v", err))
		return nil, fmt.Errorf("select revoked token: %w", err)
	}
	return &rt, nil
}

func (r *RevokedTokenRepository) IsRevoked(ctx context.Context, jti string) (bool, error) {
	r.logger.Trace(fmt.Sprintf("check revoked jti=%s", jti))

	row := r.pool.QueryRow(ctx, sql_scripts.CountRevokedByJTI, jti)
	var count int
	if err := row.Scan(&count); err != nil {
		r.logger.Error(fmt.Sprintf("count revoked tokens failed: %v", err))
		return false, fmt.Errorf("count revoked tokens: %w", err)
	}
	return count > 0, nil
}
