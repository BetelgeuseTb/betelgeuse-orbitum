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

type ScopeRepository struct {
	db *DB
}

func NewScopeRepository(pool *pgxpool.Pool, l *logger.Logger) *ScopeRepository {
	l.Info("scope repository initialized")
	return &ScopeRepository{
		db: NewDB(pool, l),
	}
}

func (r *ScopeRepository) Create(ctx context.Context, scope *models.Scope) (*models.Scope, error) {
	r.db.logger.Trace(fmt.Sprintf("scope create started orbit=%d name=%s", scope.OrbitID, scope.Name))

	now := time.Now().UTC()
	if scope.CreatedAt.IsZero() {
		scope.CreatedAt = now
	}
	scope.UpdatedAt = now

	var createdID int64
	var createdAt, updatedAt time.Time

	err := r.db.WithTx(ctx, func(tx pgx.Tx) error {
		row := tx.QueryRow(
			ctx,
			sql_scripts.InsertScope,
			scope.OrbitID,
			scope.Name,
			scope.Description,
			scope.IsDefault,
			scope.IsActive,
			scope.Metadata,
			scope.CreatedAt,
			scope.UpdatedAt,
		)
		return row.Scan(&createdID, &createdAt, &updatedAt)
	})
	if err != nil {
		r.db.logger.Error(fmt.Sprintf("scope insert failed: %v", err))
		return nil, fmt.Errorf("insert scope: %w", err)
	}

	scope.ID = createdID
	scope.CreatedAt = createdAt
	scope.UpdatedAt = updatedAt

	r.db.logger.Debug(fmt.Sprintf("scope created id=%d name=%s", scope.ID, scope.Name))
	return scope, nil
}

func (r *ScopeRepository) GetByID(ctx context.Context, id int64) (*models.Scope, error) {
	r.db.logger.Trace(fmt.Sprintf("get scope by id=%d", id))

	row := r.db.pool.QueryRow(ctx, sql_scripts.SelectScopeByID, id)
	scope, err := scanScope(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, common.ErrNotFound
		}
		r.db.logger.Error(fmt.Sprintf("get scope by id failed: %v", err))
		return nil, fmt.Errorf("get scope by id: %w", err)
	}

	return scope, nil
}

func (r *ScopeRepository) GetByName(ctx context.Context, orbitID int64, name string) (*models.Scope, error) {
	r.db.logger.Trace(fmt.Sprintf("get scope by name orbit=%d name=%s", orbitID, name))

	row := r.db.pool.QueryRow(ctx, sql_scripts.SelectScopeByName, orbitID, name)
	scope, err := scanScope(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, common.ErrNotFound
		}
		r.db.logger.Error(fmt.Sprintf("get scope by name failed: %v", err))
		return nil, fmt.Errorf("get scope by name: %w", err)
	}
	return scope, nil
}

func (r *ScopeRepository) ListByOrbit(ctx context.Context, orbitID int64, p common.Pagination) ([]models.Scope, error) {
	p = p.Normalize()
	r.db.logger.Trace(fmt.Sprintf("list scopes orbit=%d limit=%d offset=%d", orbitID, p.Limit, p.Offset))

	rows, err := r.db.pool.Query(ctx, sql_scripts.ListScopesByOrbit, orbitID, p.Limit, p.Offset)
	if err != nil {
		r.db.logger.Error(fmt.Sprintf("list scopes query failed: %v", err))
		return nil, fmt.Errorf("list scopes: %w", err)
	}
	defer rows.Close()

	var result []models.Scope

	for rows.Next() {
		scope, err := scanScope(rows)
		if err != nil {
			r.db.logger.Error(fmt.Sprintf("scan scope failed: %v", err))
			return nil, fmt.Errorf("scan scope: %w", err)
		}
		result = append(result, *scope)
	}

	if err := rows.Err(); err != nil {
		r.db.logger.Error(fmt.Sprintf("rows error: %v", err))
		return nil, fmt.Errorf("rows error: %w", err)
	}

	r.db.logger.Debug(fmt.Sprintf("listed scopes count=%d orbit=%d", len(result), orbitID))
	return result, nil
}

func (r *ScopeRepository) Update(ctx context.Context, scope *models.Scope) error {
	r.db.logger.Trace(fmt.Sprintf("scope update started id=%d", scope.ID))

	scope.UpdatedAt = time.Now().UTC()

	var id int64
	var updatedAt time.Time

	err := r.db.WithTx(ctx, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, sql_scripts.UpdateScope,
			scope.ID,
			scope.OrbitID,
			scope.Description,
			scope.IsDefault,
			scope.IsActive,
			scope.Metadata,
			scope.UpdatedAt,
		)
		return row.Scan(&id, &updatedAt)
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return common.ErrNotFound
		}
		r.db.logger.Error(fmt.Sprintf("scope update failed: %v", err))
		return fmt.Errorf("update scope: %w", err)
	}

	scope.UpdatedAt = updatedAt
	r.db.logger.Debug(fmt.Sprintf("scope updated id=%d", scope.ID))
	return nil
}

func (r *ScopeRepository) Delete(ctx context.Context, id int64) error {
	r.db.logger.Trace(fmt.Sprintf("scope delete started id=%d", id))

	now := time.Now().UTC()
	err := r.db.WithTx(ctx, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, sql_scripts.SoftDeleteScope, id, now)
		var deletedID int64
		return row.Scan(&deletedID)
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return common.ErrNotFound
		}
		r.db.logger.Error(fmt.Sprintf("scope delete failed: %v", err))
		return fmt.Errorf("delete scope: %w", err)
	}

	r.db.logger.Debug(fmt.Sprintf("scope deleted id=%d", id))
	return nil
}

func scanScope(row pgx.Row) (*models.Scope, error) {
	var scope models.Scope

	err := row.Scan(
		&scope.ID,
		&scope.OrbitID,
		&scope.Name,
		&scope.Description,
		&scope.IsDefault,
		&scope.IsActive,
		&scope.Metadata,
		&scope.CreatedAt,
		&scope.UpdatedAt,
		&scope.DeletedAt,
	)
	if err != nil {
		return nil, err
	}

	return &scope, nil
}
