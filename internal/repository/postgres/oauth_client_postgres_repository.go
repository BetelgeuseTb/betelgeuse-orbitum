package postgres

import (
	"context"
	"errors"

	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/model"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/repository"
	sqlq "github.com/BetelgeuseTb/betelgeuse-orbitum/internal/repository/postgres/sql"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type oauthClientRepoPG struct {
	db *pgxpool.Pool
}

func NewOAuthClientRepoPG(db *pgxpool.Pool) repository.OAuthClientRepository {
	return &oauthClientRepoPG{db: db}
}

func (r *oauthClientRepoPG) GetByID(ctx context.Context, id model.UUID) (*model.OAuthClient, error) {
	row := r.db.QueryRow(ctx, sqlq.OAuthClientGetByID, id)

	var c model.OAuthClient
	err := row.Scan(
		&c.ClientID,
		&c.ClientSecretHash,
		&c.ClientName,
		&c.RedirectURIs,
		&c.Scopes,
		&c.GrantTypes,
		&c.TokenEndpointAuthMethod,
		&c.Public,
		&c.CreatedAt,
		&c.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // not found is not an error
		}
		return nil, err
	}

	return &c, nil
}

func (r *oauthClientRepoPG) Create(ctx context.Context, c *model.OAuthClient) error {
	_, err := r.db.Exec(ctx, sqlq.OAuthClientCreate,
		c.ClientID,
		c.ClientSecretHash,
		c.ClientName,
		c.RedirectURIs,
		c.Scopes,
		c.GrantTypes,
		c.TokenEndpointAuthMethod,
		c.Public,
	)
	return err
}

func (r *oauthClientRepoPG) Update(ctx context.Context, c *model.OAuthClient) error {
	tag, err := r.db.Exec(ctx, sqlq.OAuthClientUpdate,
		c.ClientSecretHash,
		c.ClientName,
		c.RedirectURIs,
		c.Scopes,
		c.GrantTypes,
		c.TokenEndpointAuthMethod,
		c.Public,
		c.ClientID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("oauth client not found")
	}
	return nil
}

func (r *oauthClientRepoPG) Delete(ctx context.Context, id model.UUID) error {
	tag, err := r.db.Exec(ctx, sqlq.OAuthClientDelete, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("oauth client not found")
	}
	return nil
}
