package postgres

import (
	"context"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/model"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/repository"
	sqlq "github.com/BetelgeuseTb/betelgeuse-orbitum/internal/repository/postgres/sql"
	"github.com/jackc/pgx/v5/pgxpool"
)

type accessTokenRepoPG struct {
	db *pgxpool.Pool
}

func NewAccessTokenRepository(db *pgxpool.Pool) repository.AccessTokenRepository {
	return &accessTokenRepoPG{db: db}
}

func (r *accessTokenRepoPG) Record(ctx context.Context, t *model.AccessTokenRecord) error {
	_, err := r.db.Exec(
		ctx,
		sqlq.AccessTokenInsert,
		t.TokenID,
		t.ClientID,
		t.UserID,
		t.Scopes,
		t.IssuedAt,
		t.ExpiresAt,
		t.JTI,
	)
	return err
}

func (r *accessTokenRepoPG) GetByJTI(ctx context.Context, jti model.UUID) (*model.AccessTokenRecord, error) {
	row := r.db.QueryRow(ctx, sqlq.AccessTokenGetByJTI, jti)

	var rec model.AccessTokenRecord
	var userID *string

	err := row.Scan(
		&rec.TokenID,
		&rec.ClientID,
		&userID,
		&rec.Scopes,
		&rec.IssuedAt,
		&rec.ExpiresAt,
		&rec.JTI,
	)
	if err != nil {
		return nil, err
	}

	if userID != nil {
		rec.UserID = *userID
	}

	return &rec, nil
}

func (r *accessTokenRepoPG) DeleteExpired(ctx context.Context) error {
	_, err := r.db.Exec(ctx, sqlq.AccessTokenDeleteExpired)
	return err
}
