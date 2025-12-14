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

type JWKSRepository struct {
	pool   *pgxpool.Pool
	logger *logger.Logger
}

func NewJWKSRepository(pool *pgxpool.Pool, l *logger.Logger) *JWKSRepository {
	l.Info("jwks repository initialized")
	return &JWKSRepository{pool: pool, logger: l}
}

func (r *JWKSRepository) Create(ctx context.Context, jwk *models.JWKey) (*models.JWKey, error) {
	r.logger.Trace(fmt.Sprintf("jwk create orbit=%d kid=%s", jwk.OrbitID, jwk.Kid))

	now := time.Now().UTC()
	if jwk.CreatedAt.IsZero() {
		jwk.CreatedAt = now
	}
	jwk.UpdatedAt = now

	row := r.pool.QueryRow(ctx, sql_scripts.InsertJWK,
		jwk.OrbitID,
		jwk.Kid,
		jwk.Use,
		jwk.Alg,
		jwk.Kty,
		jwk.PublicKeyJWK,
		jwk.PrivateKeyCipher,
		jwk.IsActive,
		jwk.NotBefore,
		jwk.ExpiresAt,
		jwk.Metadata,
		jwk.CreatedAt,
		jwk.UpdatedAt,
	)

	var id int64
	var createdAt time.Time
	var updatedAt time.Time
	if err := row.Scan(&id, &createdAt, &updatedAt); err != nil {
		r.logger.Error(fmt.Sprintf("jwk insert failed: %v", err))
		return nil, fmt.Errorf("insert jwk: %w", err)
	}

	jwk.ID = id
	jwk.CreatedAt = createdAt
	jwk.UpdatedAt = updatedAt

	r.logger.Debug(fmt.Sprintf("jwk created id=%d kid=%s", jwk.ID, jwk.Kid))
	return jwk, nil
}

func (r *JWKSRepository) GetByID(ctx context.Context, id int64) (*models.JWKey, error) {
	r.logger.Trace(fmt.Sprintf("jwk get by id=%d", id))

	row := r.pool.QueryRow(ctx, sql_scripts.SelectJWKByID, id)
	jwk, err := scanJWK(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, common.ErrNotFound
		}
		r.logger.Error(fmt.Sprintf("jwk select failed: %v", err))
		return nil, fmt.Errorf("select jwk: %w", err)
	}
	return jwk, nil
}

func (r *JWKSRepository) GetByOrbitAndKid(ctx context.Context, orbitID int64, kid string) (*models.JWKey, error) {
	r.logger.Trace(fmt.Sprintf("jwk get by orbit=%d kid=%s", orbitID, kid))

	row := r.pool.QueryRow(ctx, sql_scripts.SelectJWKByOrbitAndKid, orbitID, kid)
	jwk, err := scanJWK(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, common.ErrNotFound
		}
		r.logger.Error(fmt.Sprintf("jwk select by orbit/kid failed: %v", err))
		return nil, fmt.Errorf("select jwk by orbit/kid: %w", err)
	}
	return jwk, nil
}

func (r *JWKSRepository) Update(ctx context.Context, jwk *models.JWKey) error {
	r.logger.Trace(fmt.Sprintf("jwk update id=%d", jwk.ID))

	jwk.UpdatedAt = time.Now().UTC()

	row := r.pool.QueryRow(ctx, sql_scripts.UpdateJWK,
		jwk.ID,
		jwk.OrbitID,
		jwk.PublicKeyJWK,
		jwk.PrivateKeyCipher,
		jwk.IsActive,
		jwk.NotBefore,
		jwk.ExpiresAt,
		jwk.Metadata,
		jwk.UpdatedAt,
	)

	var id int64
	var updatedAt time.Time
	if err := row.Scan(&id, &updatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return common.ErrNotFound
		}
		r.logger.Error(fmt.Sprintf("jwk update failed: %v", err))
		return fmt.Errorf("update jwk: %w", err)
	}

	jwk.UpdatedAt = updatedAt
	r.logger.Debug(fmt.Sprintf("jwk updated id=%d", jwk.ID))
	return nil
}

func (r *JWKSRepository) ListByOrbit(ctx context.Context, orbitID int64, limit, offset int) ([]models.JWKey, error) {
	r.logger.Trace(fmt.Sprintf("jwk list orbit=%d limit=%d offset=%d", orbitID, limit, offset))

	rows, err := r.pool.Query(ctx, sql_scripts.ListJWKsByOrbit, orbitID, limit, offset)
	if err != nil {
		r.logger.Error(fmt.Sprintf("jwk list query failed: %v", err))
		return nil, fmt.Errorf("list jwks: %w", err)
	}
	defer rows.Close()

	var res []models.JWKey
	for rows.Next() {
		jwk, err := scanJWK(rows)
		if err != nil {
			r.logger.Error(fmt.Sprintf("scan jwk failed: %v", err))
			return nil, fmt.Errorf("scan jwk: %w", err)
		}
		res = append(res, *jwk)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error(fmt.Sprintf("rows error: %v", err))
		return nil, fmt.Errorf("rows error: %w", err)
	}

	r.logger.Debug(fmt.Sprintf("listed jwks count=%d orbit=%d", len(res), orbitID))
	return res, nil
}

func scanJWK(row pgx.Row) (*models.JWKey, error) {
	var j models.JWKey
	var publicKey, metadata []byte
	var notBeforePtr, expiresAtPtr *time.Time

	err := row.Scan(
		&j.ID,
		&j.OrbitID,
		&j.Kid,
		&j.Use,
		&j.Alg,
		&j.Kty,
		&publicKey,
		&j.PrivateKeyCipher,
		&j.IsActive,
		&notBeforePtr,
		&expiresAtPtr,
		&metadata,
		&j.CreatedAt,
		&j.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if publicKey != nil {
		j.PublicKeyJWK = publicKey
	}
	if metadata != nil {
		j.Metadata = metadata
	}
	j.NotBefore = notBeforePtr
	j.ExpiresAt = expiresAtPtr

	return &j, nil
}
