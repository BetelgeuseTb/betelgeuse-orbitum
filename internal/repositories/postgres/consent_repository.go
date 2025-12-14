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

type ConsentRepository struct {
	pool   *pgxpool.Pool
	logger *logger.Logger
}

func NewConsentRepository(pool *pgxpool.Pool, l *logger.Logger) *ConsentRepository {
	l.Info("consent repository initialized")
	return &ConsentRepository{pool: pool, logger: l}
}

func (r *ConsentRepository) Create(ctx context.Context, c *models.Consent) (*models.Consent, error) {
	r.logger.Trace(fmt.Sprintf("consent create orbit=%d user=%d client=%d", c.OrbitID, c.UserID, c.ClientID))

	now := time.Now().UTC()
	if c.CreatedAt.IsZero() {
		c.CreatedAt = now
	}
	if c.UpdatedAt.IsZero() {
		c.UpdatedAt = now
	}
	if c.GrantedAt.IsZero() {
		c.GrantedAt = now
	}

	row := r.pool.QueryRow(ctx, sql_scripts.InsertConsent,
		c.CreatedAt,
		c.UpdatedAt,
		c.OrbitID,
		c.UserID,
		c.ClientID,
		c.Scopes,
		c.GrantedAt,
		c.ExpiresAt,
		c.Revoked,
	)

	var id int64
	var createdAt, updatedAt time.Time
	if err := row.Scan(&id, &createdAt, &updatedAt); err != nil {
		r.logger.Error(fmt.Sprintf("consent insert failed: %v", err))
		return nil, fmt.Errorf("insert consent: %w", err)
	}

	c.ID = id
	c.CreatedAt = createdAt
	c.UpdatedAt = updatedAt

	r.logger.Debug(fmt.Sprintf("consent created id=%d user=%d client=%d", c.ID, c.UserID, c.ClientID))
	return c, nil
}

func (r *ConsentRepository) GetByID(ctx context.Context, id int64) (*models.Consent, error) {
	r.logger.Trace(fmt.Sprintf("get consent by id=%d", id))

	row := r.pool.QueryRow(ctx, sql_scripts.SelectConsentByID, id)
	c, err := scanConsent(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, common.ErrNotFound
		}
		r.logger.Error(fmt.Sprintf("select consent failed: %v", err))
		return nil, fmt.Errorf("select consent: %w", err)
	}
	return c, nil
}

func (r *ConsentRepository) GetByUserAndClient(ctx context.Context, userID, clientID int64) (*models.Consent, error) {
	r.logger.Trace(fmt.Sprintf("get consent user=%d client=%d", userID, clientID))

	row := r.pool.QueryRow(ctx, sql_scripts.SelectConsentByUserAndClient, userID, clientID)
	c, err := scanConsent(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, common.ErrNotFound
		}
		r.logger.Error(fmt.Sprintf("select consent failed: %v", err))
		return nil, fmt.Errorf("select consent: %w", err)
	}
	return c, nil
}

func (r *ConsentRepository) Update(ctx context.Context, c *models.Consent) error {
	r.logger.Trace(fmt.Sprintf("update consent id=%d", c.ID))

	c.UpdatedAt = time.Now().UTC()

	row := r.pool.QueryRow(ctx, sql_scripts.UpdateConsent,
		c.ID,
		c.Scopes,
		c.GrantedAt,
		c.ExpiresAt,
		c.Revoked,
		c.UpdatedAt,
	)

	var id int64
	var updatedAt time.Time
	if err := row.Scan(&id, &updatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return common.ErrNotFound
		}
		r.logger.Error(fmt.Sprintf("consent update failed: %v", err))
		return fmt.Errorf("update consent: %w", err)
	}

	c.UpdatedAt = updatedAt
	r.logger.Debug(fmt.Sprintf("consent updated id=%d", c.ID))
	return nil
}

func (r *ConsentRepository) Delete(ctx context.Context, id int64) error {
	r.logger.Trace(fmt.Sprintf("revoke consent id=%d", id))

	now := time.Now().UTC()
	row := r.pool.QueryRow(ctx, sql_scripts.RevokeConsent, id, now)
	var returnedID int64
	if err := row.Scan(&returnedID); err != nil {
		if err == pgx.ErrNoRows {
			return common.ErrNotFound
		}
		r.logger.Error(fmt.Sprintf("revoke consent failed: %v", err))
		return fmt.Errorf("revoke consent: %w", err)
	}
	r.logger.Debug(fmt.Sprintf("consent revoked id=%d", id))
	return nil
}

func (r *ConsentRepository) ListByOrbit(ctx context.Context, orbitID int64, p common.Pagination) ([]models.Consent, error) {
	p = p.Normalize()
	r.logger.Trace(fmt.Sprintf("list consents orbit=%d limit=%d offset=%d", orbitID, p.Limit, p.Offset))

	rows, err := r.pool.Query(ctx, sql_scripts.ListConsentsByOrbit, orbitID, p.Limit, p.Offset)
	if err != nil {
		r.logger.Error(fmt.Sprintf("list consents query failed: %v", err))
		return nil, fmt.Errorf("list consents: %w", err)
	}
	defer rows.Close()

	var res []models.Consent
	for rows.Next() {
		c, err := scanConsent(rows)
		if err != nil {
			r.logger.Error(fmt.Sprintf("scan consent failed: %v", err))
			return nil, fmt.Errorf("scan consent: %w", err)
		}
		res = append(res, *c)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error(fmt.Sprintf("rows error: %v", err))
		return nil, fmt.Errorf("rows error: %w", err)
	}

	r.logger.Debug(fmt.Sprintf("listed consents count=%d orbit=%d", len(res), orbitID))
	return res, nil
}

func scanConsent(row pgx.Row) (*models.Consent, error) {
	var c models.Consent
	var scopesBytes, metadataBytes []byte
	var expiresPtr *time.Time

	err := row.Scan(
		&c.ID,
		&c.CreatedAt,
		&c.UpdatedAt,
		&c.OrbitID,
		&c.UserID,
		&c.ClientID,
		&scopesBytes,
		&c.GrantedAt,
		&expiresPtr,
		&c.Revoked,
	)
	if err != nil {
		return nil, err
	}

	if scopesBytes != nil {
		c.Scopes = scopesBytes
	}
	if expiresPtr != nil {
		c.ExpiresAt = expiresPtr
	}
	_ = metadataBytes
	return &c, nil
}
