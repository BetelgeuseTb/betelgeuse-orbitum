package postgres

import (
	"context"
	"fmt"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/repositories/common"
	"time"

	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/models"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/repositories/postgres/sql_scripts"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/utils/logger"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AuditLogRepository struct {
	pool   *pgxpool.Pool
	logger *logger.Logger
}

func NewAuditLogRepository(pool *pgxpool.Pool, l *logger.Logger) *AuditLogRepository {
	l.Info("audit log repository initialized")
	return &AuditLogRepository{pool: pool, logger: l}
}

func (r *AuditLogRepository) Create(ctx context.Context, a *models.AuditLog) (*models.AuditLog, error) {
	r.logger.Trace(fmt.Sprintf("audit log create action=%s orbit=%d", a.Action, a.OrbitID))

	if a.CreatedAt.IsZero() {
		a.CreatedAt = time.Now().UTC()
	}

	row := r.pool.QueryRow(ctx, sql_scripts.InsertAuditLog, a.CreatedAt, a.ActorUserID, a.ActorClientID, a.Action, a.Result, a.IP, a.OrbitID, a.Details)
	var id int64
	var createdAt time.Time
	if err := row.Scan(&id, &createdAt); err != nil {
		r.logger.Error(fmt.Sprintf("audit log insert failed: %v", err))
		return nil, fmt.Errorf("insert audit log: %w", err)
	}
	a.ID = id
	a.CreatedAt = createdAt
	r.logger.Debug(fmt.Sprintf("audit log created id=%d action=%s", a.ID, a.Action))
	return a, nil
}

func (r *AuditLogRepository) ListByOrbit(ctx context.Context, orbitID int64, p common.Pagination) ([]models.AuditLog, error) {
	p = p.Normalize()
	r.logger.Trace(fmt.Sprintf("list audit logs orbit=%d limit=%d offset=%d", orbitID, p.Limit, p.Offset))

	rows, err := r.pool.Query(ctx, sql_scripts.ListAuditLogsByOrbit, orbitID, p.Limit, p.Offset)
	if err != nil {
		r.logger.Error(fmt.Sprintf("list audit logs query failed: %v", err))
		return nil, fmt.Errorf("list audit logs: %w", err)
	}
	defer rows.Close()

	var res []models.AuditLog
	for rows.Next() {
		var a models.AuditLog
		var details []byte
		if err := rows.Scan(&a.ID, &a.CreatedAt, &a.ActorUserID, &a.ActorClientID, &a.Action, &a.Result, &a.IP, &a.OrbitID, &details); err != nil {
			r.logger.Error(fmt.Sprintf("scan audit log failed: %v", err))
			return nil, fmt.Errorf("scan audit log: %w", err)
		}
		if details != nil {
			a.Details = details
		}
		res = append(res, a)
	}
	if err := rows.Err(); err != nil {
		r.logger.Error(fmt.Sprintf("rows error: %v", err))
		return nil, fmt.Errorf("rows error: %w", err)
	}
	r.logger.Debug(fmt.Sprintf("listed audit logs count=%d orbit=%d", len(res), orbitID))
	return res, nil
}
