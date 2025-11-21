package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/model"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/repository"
	sqlq "github.com/BetelgeuseTb/betelgeuse-orbitum/internal/repository/postgres/sql"
	"github.com/jackc/pgx/v5/pgxpool"
)

type authorizationCodeRepoPG struct {
	db *pgxpool.Pool
}

func NewAuthorizationCodeRepoPG(db *pgxpool.Pool) repository.AuthorizationCodeRepository {
	return &authorizationCodeRepoPG{db: db}
}

func (r *authorizationCodeRepoPG) Create(ctx context.Context, ac *model.AuthorizationCode) error {
	_, err := r.db.Exec(ctx, sqlq.AuthorizationCodeCreate,
		ac.Code,
		ac.ClientID,
		ac.RedirectURI,
		ac.ExpiresAt,
		ac.UserID,
	)
	return err
}

func (r *authorizationCodeRepoPG) Get(ctx context.Context, code string) (*model.AuthorizationCode, error) {
	row := r.db.QueryRow(ctx, sqlq.AuthorizationCodeGet, code)

	var ac model.AuthorizationCode
	err := row.Scan(
		&ac.Code,
		&ac.ClientID,
		&ac.RedirectURI,
		&ac.ExpiresAt,
		&ac.UserID,
		&ac.Used,
	)
	if err != nil {
		return nil, err
	}

	return &ac, nil
}

func (r *authorizationCodeRepoPG) MarkUsed(ctx context.Context, code string) error {
	tag, err := r.db.Exec(ctx, sqlq.AuthorizationCodeMarkUsed, code)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("authorization code not found or already used")
	}
	return nil
}

func (r *authorizationCodeRepoPG) DeleteExpired(ctx context.Context) error {
	_, err := r.db.Exec(ctx, sqlq.AuthorizationCodeDeleteExpired, time.Now())
	return err
}
