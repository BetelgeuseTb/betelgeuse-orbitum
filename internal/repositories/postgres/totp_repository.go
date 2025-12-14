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

type TOTPRepository struct {
	pool   *pgxpool.Pool
	logger *logger.Logger
}

func NewTOTPRepository(pool *pgxpool.Pool, l *logger.Logger) *TOTPRepository {
	l.Info("totp repository initialized")
	return &TOTPRepository{pool: pool, logger: l}
}

func (r *TOTPRepository) Create(ctx context.Context, t *models.TOTP) (*models.TOTP, error) {
	r.logger.Trace(fmt.Sprintf("totp create user=%d", t.UserID))

	now := time.Now().UTC()
	if t.CreatedAt.IsZero() {
		t.CreatedAt = now
	}
	if t.UpdatedAt.IsZero() {
		t.UpdatedAt = now
	}

	row := r.pool.QueryRow(ctx, sql_scripts.InsertTOTP,
		t.CreatedAt,
		t.UpdatedAt,
		t.UserID,
		t.OrbitID,
		t.SecretCipher,
		t.Algorithm,
		t.Digits,
		t.Period,
		t.Issuer,
		t.Label,
		t.LastUsedStep,
		t.IsConfirmed,
		t.Name,
	)

	var id int64
	var createdAt, updatedAt time.Time
	if err := row.Scan(&id, &createdAt, &updatedAt); err != nil {
		r.logger.Error(fmt.Sprintf("totp insert failed: %v", err))
		return nil, fmt.Errorf("insert totp: %w", err)
	}
	t.ID = id
	t.CreatedAt = createdAt
	t.UpdatedAt = updatedAt
	r.logger.Debug(fmt.Sprintf("totp created id=%d user=%d", t.ID, t.UserID))
	return t, nil
}

func (r *TOTPRepository) GetByID(ctx context.Context, id int64) (*models.TOTP, error) {
	r.logger.Trace(fmt.Sprintf("get totp by id=%d", id))

	row := r.pool.QueryRow(ctx, sql_scripts.SelectTOTPByID, id)
	var t models.TOTP
	if err := row.Scan(&t.ID, &t.UserID, &t.OrbitID, &t.SecretCipher, &t.Algorithm, &t.Digits, &t.Period, &t.Issuer, &t.Label, &t.LastUsedStep, &t.IsConfirmed, &t.Name, &t.CreatedAt, &t.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, common.ErrNotFound
		}
		r.logger.Error(fmt.Sprintf("select totp failed: %v", err))
		return nil, fmt.Errorf("select totp: %w", err)
	}
	return &t, nil
}

func (r *TOTPRepository) Update(ctx context.Context, t *models.TOTP) error {
	r.logger.Trace(fmt.Sprintf("update totp id=%d", t.ID))

	t.UpdatedAt = time.Now().UTC()

	row := r.pool.QueryRow(ctx, sql_scripts.UpdateTOTP,
		t.ID,
		t.SecretCipher,
		t.Algorithm,
		t.Digits,
		t.Period,
		t.Issuer,
		t.Label,
		t.LastUsedStep,
		t.IsConfirmed,
		t.Name,
		t.UpdatedAt,
	)

	var id int64
	var updatedAt time.Time
	if err := row.Scan(&id, &updatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return common.ErrNotFound
		}
		r.logger.Error(fmt.Sprintf("totp update failed: %v", err))
		return fmt.Errorf("update totp: %w", err)
	}

	t.UpdatedAt = updatedAt
	r.logger.Debug(fmt.Sprintf("totp updated id=%d", t.ID))
	return nil
}

func (r *TOTPRepository) Delete(ctx context.Context, id int64) error {
	r.logger.Trace(fmt.Sprintf("delete totp id=%d", id))

	row := r.pool.QueryRow(ctx, sql_scripts.DeleteTOTP, id)
	var returnedID int64
	if err := row.Scan(&returnedID); err != nil {
		if err == pgx.ErrNoRows {
			return common.ErrNotFound
		}
		r.logger.Error(fmt.Sprintf("delete totp failed: %v", err))
		return fmt.Errorf("delete totp: %w", err)
	}
	r.logger.Debug(fmt.Sprintf("totp deleted id=%d", id))
	return nil
}

func (r *TOTPRepository) ListByUser(ctx context.Context, userID int64) ([]models.TOTP, error) {
	r.logger.Trace(fmt.Sprintf("list totps user=%d", userID))

	rows, err := r.pool.Query(ctx, sql_scripts.ListTOTPsByUser, userID)
	if err != nil {
		r.logger.Error(fmt.Sprintf("list totps query failed: %v", err))
		return nil, fmt.Errorf("list totps: %w", err)
	}
	defer rows.Close()

	var res []models.TOTP
	for rows.Next() {
		var t models.TOTP
		if err := rows.Scan(&t.ID, &t.UserID, &t.OrbitID, &t.SecretCipher, &t.Algorithm, &t.Digits, &t.Period, &t.Issuer, &t.Label, &t.LastUsedStep, &t.IsConfirmed, &t.Name, &t.CreatedAt, &t.UpdatedAt); err != nil {
			r.logger.Error(fmt.Sprintf("scan totp failed: %v", err))
			return nil, fmt.Errorf("scan totp: %w", err)
		}
		res = append(res, t)
	}
	if err := rows.Err(); err != nil {
		r.logger.Error(fmt.Sprintf("rows error: %v", err))
		return nil, fmt.Errorf("rows error: %w", err)
	}
	r.logger.Debug(fmt.Sprintf("listed totps count=%d user=%d", len(res), userID))
	return res, nil
}
