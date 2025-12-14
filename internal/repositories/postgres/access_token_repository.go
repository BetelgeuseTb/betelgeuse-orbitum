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

type AccessTokenRepository struct {
	db *DB
}

func NewAccessTokenRepository(pool *pgxpool.Pool, l *logger.Logger) *AccessTokenRepository {
	l.Info("access token repository initialized")
	return &AccessTokenRepository{db: NewDB(pool, l)}
}

func (r *AccessTokenRepository) Create(ctx context.Context, toCreate *models.AccessToken) (*models.AccessToken, error) {
	r.db.logger.Trace(fmt.Sprintf("begin: save access token orbit=%d client=%d jti=%s", toCreate.OrbitID, toCreate.ClientID, toCreate.JTI))

	res, err := InTransaction(r.db, ctx, func(tx pgx.Tx) (*models.AccessToken, error) {
		row := tx.QueryRow(ctx, sql_scripts.InsertAccessToken,
			toCreate.JTI,
			toCreate.OrbitID,
			toCreate.ClientID,
			toCreate.UserID,
			toCreate.IsJWT,
			toCreate.TokenString,
			toCreate.Scope,
			toCreate.IssuedAt,
			toCreate.TokenType,
			toCreate.Revoked,
			toCreate.Metadata,
			toCreate.RefreshTokenID,
			toCreate.CreatedAt,
			toCreate.ExpiresAt,
		)

		return scanAccessToken(row)
	})

	if err != nil {
		r.db.logger.Error(fmt.Sprintf("access token insert failed: %v", err))
		return nil, fmt.Errorf("insert access token: %w", err)
	}

	r.db.logger.Trace(fmt.Sprintf("finished: save access token orbit=%d client=%d jti=%s with id=%d",
		res.OrbitID, res.ClientID, res.JTI, res.ID))
	return res, nil
}

func (r *AccessTokenRepository) GetByID(ctx context.Context, id int64) (*models.AccessToken, error) {
	r.db.logger.Trace(fmt.Sprintf("begin: get access token by id=%d", id))

	row := r.db.pool.QueryRow(ctx, sql_scripts.SelectAccessTokenByID, id)
	res, err := scanAccessToken(row)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, common.ErrNotFound
		}
		r.db.logger.Error(fmt.Sprintf("get access token by id failed: %v", err))
		return nil, fmt.Errorf("get access token by id: %w", err)
	}

	r.db.logger.Trace(fmt.Sprintf("finished: get access token by id=%d", id))
	return res, nil
}

func (r *AccessTokenRepository) GetByJTI(ctx context.Context, jti string) (*models.AccessToken, error) {
	r.db.logger.Trace(fmt.Sprintf("begin: get access token by jti=%s start", jti))

	row := r.db.pool.QueryRow(ctx, sql_scripts.SelectAccessTokenByJTI, jti)
	at, err := scanAccessToken(row)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, common.ErrNotFound
		}
		r.db.logger.Error(fmt.Sprintf("get access token by jti failed: %v", err))
		return nil, fmt.Errorf("get access token by jti: %w", err)
	}

	r.db.logger.Trace(fmt.Sprintf("finished: get access token by jti=%s", jti))
	return at, nil
}

func (r *AccessTokenRepository) Update(ctx context.Context, toUpdate *models.AccessToken) (*models.AccessToken, error) {
	r.db.logger.Trace(fmt.Sprintf("begin: access token update with id=%d", toUpdate.ID))

	res, err := InTransaction(r.db, ctx, func(tx pgx.Tx) (*models.AccessToken, error) {
		row := tx.QueryRow(ctx, sql_scripts.UpdateAccessToken,
			toUpdate.ID,
			toUpdate.TokenString,
			toUpdate.Revoked,
			toUpdate.Metadata,
			toUpdate.RefreshTokenID,
			toUpdate.IssuedAt,
			toUpdate.ExpiresAt,
		)
		return scanAccessToken(row)
	})

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, common.ErrNotFound
		}
		r.db.logger.Error(fmt.Sprintf("access token update failed: %v", err))
		return nil, fmt.Errorf("update access token: %w", err)
	}

	r.db.logger.Debug(fmt.Sprintf("finished: access token updated id=%d success", toUpdate.ID))
	return res, nil
}

func (r *AccessTokenRepository) RevokeByJTI(ctx context.Context, jti string) error {
	r.db.logger.Trace(fmt.Sprintf("begin: revoke access token jti=%s", jti))

	_, err := InTransaction(r.db, ctx, func(tx pgx.Tx) (any, error) {
		tx.QueryRow(ctx, sql_scripts.RevokeAccessTokenByJTI, jti)
		return nil, nil
	})

	if err != nil {
		if err == pgx.ErrNoRows {
			return common.ErrNotFound
		}
		r.db.logger.Error(fmt.Sprintf("revoke access token failed: %v", err))
		return fmt.Errorf("revoke access token: %w", err)
	}

	r.db.logger.Debug(fmt.Sprintf("finished: access token revoked jti=%s", jti))
	return nil
}

func scanAccessToken(row pgx.Row) (*models.AccessToken, error) {
	var result models.AccessToken

	var (
		userIDPtr      *int64
		refreshIDPtr   *int64
		tokenStringPtr *string
		issuedAtPtr    *time.Time
		tokenTypePtr   *string
		revokedPtr     *bool
		scopeBytes     []byte
		metadataBytes  []byte
	)

	err := row.Scan(
		&result.ID,
		&result.JTI,
		&result.OrbitID,
		&result.ClientID,
		&userIDPtr,
		&result.IsJWT,
		&tokenStringPtr,
		&scopeBytes,
		&issuedAtPtr,
		&tokenTypePtr,
		&revokedPtr,
		&metadataBytes,
		&refreshIDPtr,
		&result.CreatedAt,
		&result.ExpiresAt,
	)
	if err != nil {
		return nil, err
	}

	if userIDPtr != nil {
		result.UserID = userIDPtr
	} else {
		result.UserID = nil
	}
	if refreshIDPtr != nil {
		result.RefreshTokenID = refreshIDPtr
	} else {
		result.RefreshTokenID = nil
	}
	if tokenStringPtr != nil {
		result.TokenString = *tokenStringPtr
	}
	if issuedAtPtr != nil {
		result.IssuedAt = *issuedAtPtr
	}
	if tokenTypePtr != nil {
		result.TokenType = *tokenTypePtr
	}
	if revokedPtr != nil {
		result.Revoked = *revokedPtr
	} else {
		result.Revoked = false
	}
	if scopeBytes != nil {
		result.Scope = scopeBytes
	}
	if metadataBytes != nil {
		result.Metadata = metadataBytes
	}

	return &result, nil
}
