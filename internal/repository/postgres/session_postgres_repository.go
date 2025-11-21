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

type sessionRepoPG struct {
	db *pgxpool.Pool
}

func NewSessionRepoPG(db *pgxpool.Pool) repository.SessionRepository {
	return &sessionRepoPG{db: db}
}

func (r *sessionRepoPG) Create(ctx context.Context, s *model.Session) error {
	return r.db.QueryRow(ctx, sqlq.SessionInsert,
		s.UserID,
		s.RefreshTokenHash,
		s.UserAgent,
		s.IPAddress,
		s.ExpiresAt,
	).Scan(
		&s.ID,
		&s.UserID,
		&s.RefreshTokenHash,
		&s.UserAgent,
		&s.IPAddress,
		&s.ExpiresAt,
		&s.CreatedAt,
		&s.UpdatedAt,
		&s.Revoked,
	)
}

func (r *sessionRepoPG) GetByID(ctx context.Context, id model.UUID) (*model.Session, error) {
	row := r.db.QueryRow(ctx, sqlq.SessionGetByID, id)

	var s model.Session
	err := row.Scan(
		&s.ID,
		&s.UserID,
		&s.RefreshTokenHash,
		&s.UserAgent,
		&s.IPAddress,
		&s.ExpiresAt,
		&s.CreatedAt,
		&s.UpdatedAt,
		&s.Revoked,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &s, nil
}

func (r *sessionRepoPG) GetByUser(ctx context.Context, userID model.UUID) ([]model.Session, error) {
	rows, err := r.db.Query(ctx, sqlq.SessionGetByUser, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []model.Session

	for rows.Next() {
		var s model.Session
		err := rows.Scan(
			&s.ID,
			&s.UserID,
			&s.RefreshTokenHash,
			&s.UserAgent,
			&s.IPAddress,
			&s.ExpiresAt,
			&s.CreatedAt,
			&s.UpdatedAt,
			&s.Revoked,
		)
		if err != nil {
			return nil, err
		}

		sessions = append(sessions, s)
	}

	return sessions, rows.Err()
}

func (r *sessionRepoPG) Revoke(ctx context.Context, id model.UUID) error {
	_, err := r.db.Exec(ctx, sqlq.SessionRevoke, id)
	return err
}

func (r *sessionRepoPG) DeleteExpired(ctx context.Context) error {
	_, err := r.db.Exec(ctx, sqlq.SessionDeleteExpired)
	return err
}
