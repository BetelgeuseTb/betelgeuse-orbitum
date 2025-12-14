package postgres

import (
	"context"
	"fmt"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/models"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/repositories/common"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/repositories/postgres/sql_scripts"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/utils/logger"
)

type OrbitRepository struct {
	db *DB
}

func NewOrbitRepository(pool *pgxpool.Pool, log *logger.Logger) *OrbitRepository {
	log.Info("orbit repository initialized")
	return &OrbitRepository{
		db: NewDB(pool, log),
	}
}

func (r *OrbitRepository) Create(ctx context.Context, orbit *models.Orbit) (*models.Orbit, error) {
	r.db.logger.Trace("orbit create started")

	now := time.Now().UTC()
	if orbit.CreatedAt.IsZero() {
		orbit.CreatedAt = now
	}
	orbit.UpdatedAt = now

	var created models.Orbit

	err := r.db.WithTx(ctx, func(tx pgx.Tx) error {
		row := tx.QueryRow(
			ctx,
			sql_scripts.InsertOrbit,
			orbit.Name,
			orbit.DisplayName,
			orbit.Description,
			orbit.Issuer,
			orbit.Domain,
			orbit.Config,
			orbit.DefaultScopes,
			orbit.CreatedAt,
			orbit.UpdatedAt,
		)
		return row.Scan(&created.ID, &created.CreatedAt, &created.UpdatedAt, &created.DeletedAt)
	})
	if err != nil {
		r.db.logger.Error(fmt.Sprintf("orbit create failed: %v", err))
		return nil, fmt.Errorf("create orbit: %w", err)
	}

	created.Name = orbit.Name
	created.DisplayName = orbit.DisplayName
	created.Description = orbit.Description
	created.Issuer = orbit.Issuer
	created.Domain = orbit.Domain
	created.Config = orbit.Config
	created.DefaultScopes = orbit.DefaultScopes

	r.db.logger.Debug(fmt.Sprintf("orbit created id=%d name=%s", created.ID, created.Name))
	return &created, nil
}

func (r *OrbitRepository) GetByID(ctx context.Context, id int64) (*models.Orbit, error) {
	r.db.logger.Trace(fmt.Sprintf("get orbit by id=%d", id))

	row := r.db.pool.QueryRow(ctx, sql_scripts.SelectOrbitByID, id)
	orbit, err := scanOrbit(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, common.ErrNotFound
		}
		r.db.logger.Error(fmt.Sprintf("get orbit by id failed: %v", err))
		return nil, fmt.Errorf("get orbit by id: %w", err)
	}

	return orbit, nil
}

func (r *OrbitRepository) GetByName(ctx context.Context, name string) (*models.Orbit, error) {
	r.db.logger.Trace(fmt.Sprintf("get orbit by name=%s", name))

	row := r.db.pool.QueryRow(ctx, sql_scripts.SelectOrbitByName, name)
	orbit, err := scanOrbit(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, common.ErrNotFound
		}
		r.db.logger.Error(fmt.Sprintf("get orbit by name failed: %v", err))
		return nil, fmt.Errorf("get orbit by name: %w", err)
	}

	return orbit, nil
}

func (r *OrbitRepository) Update(ctx context.Context, orbit *models.Orbit) error {
	r.db.logger.Trace(fmt.Sprintf("orbit update started id=%d", orbit.ID))

	orbit.UpdatedAt = time.Now().UTC()

	var id int64
	var updatedAt time.Time

	err := r.db.WithTx(ctx, func(tx pgx.Tx) error {
		row := tx.QueryRow(
			ctx,
			sql_scripts.UpdateOrbit,
			orbit.ID,
			orbit.Name,
			orbit.DisplayName,
			orbit.Description,
			orbit.Issuer,
			orbit.Domain,
			orbit.Config,
			orbit.DefaultScopes,
			orbit.UpdatedAt,
		)
		return row.Scan(&id, &updatedAt)
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return common.ErrNotFound
		}
		r.db.logger.Error(fmt.Sprintf("orbit update failed: %v", err))
		return fmt.Errorf("update orbit: %w", err)
	}

	orbit.UpdatedAt = updatedAt
	r.db.logger.Debug(fmt.Sprintf("orbit updated id=%d", orbit.ID))
	return nil
}

func (r *OrbitRepository) Delete(ctx context.Context, id int64) error {
	r.db.logger.Trace(fmt.Sprintf("orbit delete started id=%d", id))

	now := time.Now().UTC()
	err := r.db.WithTx(ctx, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, sql_scripts.SoftDeleteOrbit, id, now)
		var deletedID int64
		return row.Scan(&deletedID)
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return common.ErrNotFound
		}
		r.db.logger.Error(fmt.Sprintf("orbit delete failed: %v", err))
		return fmt.Errorf("delete orbit: %w", err)
	}

	r.db.logger.Debug(fmt.Sprintf("orbit deleted id=%d", id))
	return nil
}

func (r *OrbitRepository) List(ctx context.Context, p common.Pagination) ([]models.Orbit, error) {
	p = p.Normalize()
	r.db.logger.Trace(fmt.Sprintf("list orbits limit=%d offset=%d", p.Limit, p.Offset))

	rows, err := r.db.pool.Query(ctx, sql_scripts.ListOrbits, p.Limit, p.Offset)
	if err != nil {
		r.db.logger.Error(fmt.Sprintf("list orbits query failed: %v", err))
		return nil, fmt.Errorf("list orbits: %w", err)
	}
	defer rows.Close()

	var result []models.Orbit

	for rows.Next() {
		orbit, err := scanOrbit(rows)
		if err != nil {
			r.db.logger.Error(fmt.Sprintf("scan orbit failed: %v", err))
			return nil, fmt.Errorf("scan orbit: %w", err)
		}
		result = append(result, *orbit)
	}

	if err := rows.Err(); err != nil {
		r.db.logger.Error(fmt.Sprintf("rows error: %v", err))
		return nil, fmt.Errorf("rows error: %w", err)
	}

	r.db.logger.Debug(fmt.Sprintf("list orbits returned %d items", len(result)))
	return result, nil
}

func scanOrbit(row pgx.Row) (*models.Orbit, error) {
	var orbit models.Orbit
	err := row.Scan(
		&orbit.ID,
		&orbit.Name,
		&orbit.DisplayName,
		&orbit.Description,
		&orbit.Issuer,
		&orbit.Domain,
		&orbit.Config,
		&orbit.DefaultScopes,
		&orbit.CreatedAt,
		&orbit.UpdatedAt,
		&orbit.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return &orbit, nil
}
