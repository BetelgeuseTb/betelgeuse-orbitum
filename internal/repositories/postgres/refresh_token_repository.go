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

type RefreshTokenRepository struct {
	db *DB
}

func NewRefreshTokenRepository(pool *pgxpool.Pool, l *logger.Logger) *RefreshTokenRepository {
	l.Info("refresh token repository initialized")
	return &RefreshTokenRepository{db: NewDB(pool, l)}
}

func (r *RefreshTokenRepository) Create(ctx context.Context, rt *models.RefreshToken) (*models.RefreshToken, error) {
	r.db.logger.Trace(fmt.Sprintf("refresh token create start orbit=%d client=%d jti=%s", rt.OrbitID, rt.ClientID, rt.JTI))

	if rt.ExpiresAt.IsZero() {
		return nil, fmt.Errorf("expires_at is required")
	}
	now := time.Now().UTC()
	if rt.CreatedAt.IsZero() {
		rt.CreatedAt = now
	}

	var created models.RefreshToken
	err := r.db.WithTx(ctx, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, sql_scripts.InsertRefreshToken,
			rt.ExpiresAt,
			rt.TokenString,
			rt.JTI,
			rt.OrbitID,
			rt.ClientID,
			rt.UserID,
			rt.Revoked,
			rt.RotatedFromID,
			rt.RotatedToID,
			rt.Scopes,
			rt.Metadata,
			rt.LastUsedAt,
			rt.UseCount,
			rt.CreatedAt,
		)
		return row.Scan(&created.ID, &created.CreatedAt, &created.ExpiresAt)
	})
	if err != nil {
		r.db.logger.Error(fmt.Sprintf("refresh token insert failed: %v", err))
		return nil, fmt.Errorf("insert refresh token: %w", err)
	}

	created.TokenString = rt.TokenString
	created.JTI = rt.JTI
	created.OrbitID = rt.OrbitID
	created.ClientID = rt.ClientID
	created.UserID = rt.UserID
	created.Revoked = rt.Revoked
	created.RotatedFromID = rt.RotatedFromID
	created.RotatedToID = rt.RotatedToID
	created.Scopes = rt.Scopes
	created.Metadata = rt.Metadata
	created.LastUsedAt = rt.LastUsedAt
	created.UseCount = rt.UseCount

	r.db.logger.Debug(fmt.Sprintf("refresh token created id=%d jti=%s", created.ID, created.JTI))
	return &created, nil
}

func (r *RefreshTokenRepository) GetByID(ctx context.Context, id int64) (*models.RefreshToken, error) {
	r.db.logger.Trace(fmt.Sprintf("get refresh token by id=%d", id))

	row := r.db.pool.QueryRow(ctx, sql_scripts.SelectRefreshTokenByID, id)
	rt, err := scanRefreshToken(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, common.ErrNotFound
		}
		r.db.logger.Error(fmt.Sprintf("get refresh token failed: %v", err))
		return nil, fmt.Errorf("get refresh token: %w", err)
	}
	return rt, nil
}

func (r *RefreshTokenRepository) GetByJTI(ctx context.Context, jti string) (*models.RefreshToken, error) {
	r.db.logger.Trace(fmt.Sprintf("get refresh token by jti=%s", jti))

	row := r.db.pool.QueryRow(ctx, sql_scripts.SelectRefreshTokenByJTI, jti)
	rt, err := scanRefreshToken(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, common.ErrNotFound
		}
		r.db.logger.Error(fmt.Sprintf("get refresh token by jti failed: %v", err))
		return nil, fmt.Errorf("get refresh token by jti: %w", err)
	}
	return rt, nil
}

func (r *RefreshTokenRepository) Update(ctx context.Context, rt *models.RefreshToken) error {
	r.db.logger.Trace(fmt.Sprintf("refresh token update id=%d", rt.ID))

	var id int64
	var expires time.Time

	err := r.db.WithTx(ctx, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, sql_scripts.UpdateRefreshToken,
			rt.ID,
			rt.TokenString,
			rt.Revoked,
			rt.RotatedFromID,
			rt.RotatedToID,
			rt.Scopes,
			rt.Metadata,
			rt.LastUsedAt,
			rt.UseCount,
			rt.ExpiresAt,
		)
		return row.Scan(&id, &expires)
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return common.ErrNotFound
		}
		r.db.logger.Error(fmt.Sprintf("refresh token update failed: %v", err))
		return fmt.Errorf("update refresh token: %w", err)
	}

	rt.ExpiresAt = expires
	r.db.logger.Debug(fmt.Sprintf("refresh token updated id=%d", rt.ID))
	return nil
}

func (r *RefreshTokenRepository) RevokeByJTI(ctx context.Context, jti string) error {
	r.db.logger.Trace(fmt.Sprintf("revoke refresh token jti=%s", jti))

	err := r.db.WithTx(ctx, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, sql_scripts.RevokeRefreshTokenByJTI, jti)
		var id int64
		return row.Scan(&id)
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return common.ErrNotFound
		}
		r.db.logger.Error(fmt.Sprintf("revoke refresh token failed: %v", err))
		return fmt.Errorf("revoke refresh token: %w", err)
	}
	r.db.logger.Debug(fmt.Sprintf("refresh token revoked jti=%s", jti))
	return nil
}

func (r *RefreshTokenRepository) Rotate(ctx context.Context, id int64, rotatedToID int64) error {
	r.db.logger.Trace(fmt.Sprintf("rotate refresh token id=%d rotatedTo=%d", id, rotatedToID))

	err := r.db.WithTx(ctx, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, sql_scripts.RotateRefreshToken, id, rotatedToID)
		var returnedID int64
		return row.Scan(&returnedID)
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return common.ErrNotFound
		}
		r.db.logger.Error(fmt.Sprintf("rotate refresh token failed: %v", err))
		return fmt.Errorf("rotate refresh token: %w", err)
	}
	r.db.logger.Debug(fmt.Sprintf("refresh token rotated id=%d to=%d", id, rotatedToID))
	return nil
}

func scanRefreshToken(row pgx.Row) (*models.RefreshToken, error) {
	var rt models.RefreshToken

	var (
		userIDPtr     *int64
		rotFromPtr    *int64
		rotToPtr      *int64
		scopesBytes   []byte
		metadataBytes []byte
		lastUsedPtr   *time.Time
	)

	err := row.Scan(
		&rt.ID,
		&rt.ExpiresAt,
		&rt.TokenString,
		&rt.JTI,
		&rt.OrbitID,
		&rt.ClientID,
		&userIDPtr,
		&rt.Revoked,
		&rotFromPtr,
		&rotToPtr,
		&scopesBytes,
		&metadataBytes,
		&lastUsedPtr,
		&rt.UseCount,
		&rt.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if userIDPtr != nil {
		rt.UserID = userIDPtr
	}
	if rotFromPtr != nil {
		rt.RotatedFromID = rotFromPtr
	}
	if rotToPtr != nil {
		rt.RotatedToID = rotToPtr
	}
	if scopesBytes != nil {
		rt.Scopes = scopesBytes
	}
	if metadataBytes != nil {
		rt.Metadata = metadataBytes
	}
	if lastUsedPtr != nil {
		rt.LastUsedAt = lastUsedPtr
	}

	return &rt, nil
}
