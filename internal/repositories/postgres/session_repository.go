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

type SessionRepository struct {
	db *DB
}

func NewSessionRepository(pool *pgxpool.Pool, l *logger.Logger) *SessionRepository {
	l.Info("session repository initialized")
	return &SessionRepository{db: NewDB(pool, l)}
}

func (r *SessionRepository) Create(ctx context.Context, s *models.Session) (*models.Session, error) {
	r.db.logger.Trace(fmt.Sprintf("session create orbit=%d user=%d", s.OrbitID, s.UserID))

	now := time.Now().UTC()
	if s.CreatedAt.IsZero() {
		s.CreatedAt = now
	}
	if s.UpdatedAt.IsZero() {
		s.UpdatedAt = now
	}
	if s.StartedAt.IsZero() {
		s.StartedAt = now
	}
	if s.LastActiveAt.IsZero() {
		s.LastActiveAt = now
	}

	var id int64
	var createdAt, updatedAt time.Time

	err := r.db.WithTx(ctx, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, sql_scripts.InsertSession,
			s.OrbitID,
			s.UserID,
			s.ClientID,
			s.StartedAt,
			s.LastActiveAt,
			s.ExpiresAt,
			s.Revoked,
			s.DeviceInfo,
			s.IP,
			s.Metadata,
			s.CreatedAt,
			s.UpdatedAt,
		)
		return row.Scan(&id, &createdAt, &updatedAt)
	})
	if err != nil {
		r.db.logger.Error(fmt.Sprintf("session insert failed: %v", err))
		return nil, fmt.Errorf("insert session: %w", err)
	}

	s.ID = id
	s.CreatedAt = createdAt
	s.UpdatedAt = updatedAt

	r.db.logger.Debug(fmt.Sprintf("session created id=%d user=%d", s.ID, s.UserID))
	return s, nil
}

func (r *SessionRepository) GetByID(ctx context.Context, id int64) (*models.Session, error) {
	r.db.logger.Trace(fmt.Sprintf("get session by id=%d", id))

	row := r.db.pool.QueryRow(ctx, sql_scripts.SelectSessionByID, id)
	s, err := scanSession(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, common.ErrNotFound
		}
		r.db.logger.Error(fmt.Sprintf("get session failed: %v", err))
		return nil, fmt.Errorf("get session: %w", err)
	}
	return s, nil
}

func (r *SessionRepository) Update(ctx context.Context, s *models.Session) error {
	r.db.logger.Trace(fmt.Sprintf("update session id=%d", s.ID))

	s.UpdatedAt = time.Now().UTC()

	var id int64
	var updatedAt time.Time

	err := r.db.WithTx(ctx, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, sql_scripts.UpdateSession,
			s.ID,
			s.LastActiveAt,
			s.ExpiresAt,
			s.Revoked,
			s.Metadata,
			s.UpdatedAt,
		)
		return row.Scan(&id, &updatedAt)
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return common.ErrNotFound
		}
		r.db.logger.Error(fmt.Sprintf("session update failed: %v", err))
		return fmt.Errorf("update session: %w", err)
	}

	s.UpdatedAt = updatedAt
	r.db.logger.Debug(fmt.Sprintf("session updated id=%d", s.ID))
	return nil
}

func (r *SessionRepository) Delete(ctx context.Context, id int64) error {
	r.db.logger.Trace(fmt.Sprintf("revoke session id=%d", id))

	now := time.Now().UTC()
	err := r.db.WithTx(ctx, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, sql_scripts.RevokeSession, id, now)
		var returnedID int64
		return row.Scan(&returnedID)
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return common.ErrNotFound
		}
		r.db.logger.Error(fmt.Sprintf("session revoke failed: %v", err))
		return fmt.Errorf("revoke session: %w", err)
	}

	r.db.logger.Debug(fmt.Sprintf("session revoked id=%d", id))
	return nil
}

func (r *SessionRepository) ListByUser(ctx context.Context, userID int64, p common.Pagination) ([]models.Session, error) {
	p = p.Normalize()
	r.db.logger.Trace(fmt.Sprintf("list sessions user=%d limit=%d offset=%d", userID, p.Limit, p.Offset))

	rows, err := r.db.pool.Query(ctx, sql_scripts.ListSessionsByUser, userID, p.Limit, p.Offset)
	if err != nil {
		r.db.logger.Error(fmt.Sprintf("list sessions query failed: %v", err))
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	defer rows.Close()

	var res []models.Session
	for rows.Next() {
		s, err := scanSession(rows)
		if err != nil {
			r.db.logger.Error(fmt.Sprintf("scan session failed: %v", err))
			return nil, fmt.Errorf("scan session: %w", err)
		}
		res = append(res, *s)
	}

	if err := rows.Err(); err != nil {
		r.db.logger.Error(fmt.Sprintf("rows error: %v", err))
		return nil, fmt.Errorf("rows error: %w", err)
	}

	r.db.logger.Debug(fmt.Sprintf("listed sessions count=%d user=%d", len(res), userID))
	return res, nil
}

func scanSession(row pgx.Row) (*models.Session, error) {
	var s models.Session
	var clientIDPtr *int64
	var expiresPtr *time.Time
	var metadataBytes []byte

	err := row.Scan(
		&s.ID,
		&s.OrbitID,
		&s.UserID,
		&clientIDPtr,
		&s.StartedAt,
		&s.LastActiveAt,
		&expiresPtr,
		&s.Revoked,
		&s.DeviceInfo,
		&s.IP,
		&metadataBytes,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if clientIDPtr != nil {
		s.ClientID = clientIDPtr
	}
	if expiresPtr != nil {
		s.ExpiresAt = expiresPtr
	}
	if metadataBytes != nil {
		s.Metadata = metadataBytes
	}

	return &s, nil
}
