package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/models"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/repositories/postgres/sql_scripts"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/utils/logger"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ClientRepository struct {
	db *DB
}

func NewClientRepository(pool *pgxpool.Pool, l *logger.Logger) *ClientRepository {
	l.Info("client repository created")
	return &ClientRepository{db: NewDB(pool, l)}
}

func (r *ClientRepository) Create(ctx context.Context, c *models.Client) (*models.Client, error) {
	r.db.logger.Trace(fmt.Sprintf("client create start orbit=%d client_id=%s", c.OrbitID, c.ClientID))

	now := time.Now().UTC()
	if c.CreatedAt.IsZero() {
		c.CreatedAt = now
	}
	c.UpdatedAt = now

	var id int64
	var createdAt, updatedAt time.Time
	var deletedAt *time.Time

	err := r.db.WithTx(ctx, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, sql_scripts.InsertClient,
			c.OrbitID,
			c.ClientID,
			c.ClientSecretHash,
			c.Name,
			c.Description,
			c.RedirectURIs,
			c.PostLogoutRedirectURIs,
			c.GrantTypes,
			c.ResponseTypes,
			c.TokenEndpointAuthMethod,
			c.Contacts,
			c.LogoURI,
			c.AppType,
			c.IsPublic,
			c.IsActive,
			c.AllowedCORSOrigins,
			c.AllowedScopes,
			c.Metadata,
			c.CreatedAt,
			c.UpdatedAt,
		)
		return row.Scan(&id, &createdAt, &updatedAt, &deletedAt)
	})

	if err != nil {
		r.db.logger.Error(fmt.Sprintf("client insert failed: %v", err))
		return nil, fmt.Errorf("insert client: %w", err)
	}

	c.ID = id
	c.CreatedAt = createdAt
	c.UpdatedAt = updatedAt
	c.DeletedAt = deletedAt

	r.db.logger.Info(fmt.Sprintf("client created id=%d client_id=%s", c.ID, c.ClientID))
	return c, nil
}

func (r *ClientRepository) GetByID(ctx context.Context, id int64) (*models.Client, error) {
	r.db.logger.Trace(fmt.Sprintf("get client by id=%d", id))

	row := r.db.pool.QueryRow(ctx, sql_scripts.SelectClientByID, id)
	client, err := scanClient(row)
	if err != nil {
		r.db.logger.Error(fmt.Sprintf("client select by id failed: %v", err))
		return nil, fmt.Errorf("select client: %w", err)
	}

	return client, nil
}

func (r *ClientRepository) GetByClientID(ctx context.Context, orbitID int64, clientID string) (*models.Client, error) {
	r.db.logger.Trace(fmt.Sprintf("client get by client_id orbit=%d client_id=%s", orbitID, clientID))

	row := r.db.pool.QueryRow(ctx, sql_scripts.SelectClientByClientID, orbitID, clientID)
	client, err := scanClient(row)
	if err != nil {
		r.db.logger.Error(fmt.Sprintf("client select failed: %v", err))
		return nil, fmt.Errorf("select client: %w", err)
	}

	return client, nil
}

func (r *ClientRepository) Update(ctx context.Context, c *models.Client) error {
	r.db.logger.Trace(fmt.Sprintf("client update id=%d", c.ID))

	c.UpdatedAt = time.Now().UTC()

	var id int64
	var updatedAt time.Time

	err := r.db.WithTx(ctx, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, sql_scripts.UpdateClient,
			c.ID,
			c.ClientSecretHash,
			c.Name,
			c.Description,
			c.RedirectURIs,
			c.PostLogoutRedirectURIs,
			c.GrantTypes,
			c.ResponseTypes,
			c.TokenEndpointAuthMethod,
			c.Contacts,
			c.LogoURI,
			c.AppType,
			c.IsPublic,
			c.IsActive,
			c.AllowedCORSOrigins,
			c.AllowedScopes,
			c.Metadata,
			c.UpdatedAt,
		)
		return row.Scan(&id, &updatedAt)
	})

	if err != nil {
		r.db.logger.Error(fmt.Sprintf("client update failed: %v", err))
		return fmt.Errorf("update client: %w", err)
	}

	c.UpdatedAt = updatedAt
	r.db.logger.Debug(fmt.Sprintf("client updated id=%d", c.ID))
	return nil
}

func (r *ClientRepository) Delete(ctx context.Context, id int64) error {
	r.db.logger.Trace(fmt.Sprintf("client delete id=%d", id))

	now := time.Now().UTC()
	err := r.db.WithTx(ctx, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, sql_scripts.SoftDeleteClient, id, now)
		var deletedID int64
		return row.Scan(&deletedID)
	})

	if err != nil {
		r.db.logger.Error(fmt.Sprintf("client delete failed: %v", err))
		return fmt.Errorf("delete client: %w", err)
	}

	r.db.logger.Debug(fmt.Sprintf("client deleted id=%d", id))
	return nil
}

func scanClient(row pgx.Row) (*models.Client, error) {
	var c models.Client

	err := row.Scan(
		&c.ID,
		&c.OrbitID,
		&c.ClientID,
		&c.ClientSecretHash,
		&c.Name,
		&c.Description,
		&c.RedirectURIs,
		&c.PostLogoutRedirectURIs,
		&c.GrantTypes,
		&c.ResponseTypes,
		&c.TokenEndpointAuthMethod,
		&c.Contacts,
		&c.LogoURI,
		&c.AppType,
		&c.IsPublic,
		&c.IsActive,
		&c.AllowedCORSOrigins,
		&c.AllowedScopes,
		&c.Metadata,
		&c.CreatedAt,
		&c.UpdatedAt,
		&c.DeletedAt,
	)

	if err != nil {
		return nil, err
	}

	return &c, nil
}
