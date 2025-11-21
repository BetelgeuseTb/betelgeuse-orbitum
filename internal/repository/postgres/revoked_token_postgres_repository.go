package postgres

import (
	"context"

	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/model"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/repository"
	sqlq "github.com/BetelgeuseTb/betelgeuse-orbitum/internal/repository/postgres/sql"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type revokedTokenRepoPG struct {
	db *pgxpool.Pool
}

func NewRevokedTokenRepoPG(db *pgxpool.Pool) repository.RevokedTokenRepository {
	return &revokedTokenRepoPG{db: db}
}

func (r *revokedTokenRepoPG) Add(ctx context.Context, t *model.RevokedToken) error {
	_, err := r.db.Exec(ctx, sqlq.RevokedTokenInsert,
		t.JTI,
		t.TokenType,
		t.RevokedAt,
		t.Reason,
	)
	return err
}

func (r *revokedTokenRepoPG) IsRevoked(ctx context.Context, jti model.UUID) (bool, error) {
	row := r.db.QueryRow(ctx, sqlq.RevokedTokenCheck, jti)

	var tmp int
	err := row.Scan(&tmp)

	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
