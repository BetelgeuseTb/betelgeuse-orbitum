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

type UserRepository struct {
	db *DB
}

func NewUserRepository(pool *pgxpool.Pool, log *logger.Logger) *UserRepository {
	log.Info("user repository initialized")
	return &UserRepository{db: NewDB(pool, log)}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	r.db.logger.Trace(fmt.Sprintf("user create started orbit=%d username=%s", user.OrbitID, user.Username))

	now := time.Now().UTC()
	createdAt := user.CreatedAt
	if createdAt.IsZero() {
		createdAt = now
	}

	var created models.User
	err := r.db.WithTx(ctx, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, sql_scripts.InsertUser,
			user.OrbitID,
			user.Username,
			user.Email,
			user.EmailVerified,
			user.PasswordHash,
			user.PasswordAlgo,
			user.LastPasswordChange,
			user.DisplayName,
			user.Profile,
			user.IsActive,
			user.IsLocked,
			user.MFAEnabled,
			user.Metadata,
			createdAt,
			now,
		)
		return row.Scan(&created.ID, &created.CreatedAt, &created.UpdatedAt, &created.DeletedAt)
	})
	if err != nil {
		r.db.logger.Error(fmt.Sprintf("user create failed: %v", err))
		return nil, fmt.Errorf("create user: %w", err)
	}

	created.OrbitID = user.OrbitID
	created.Username = user.Username
	created.Email = user.Email
	created.EmailVerified = user.EmailVerified
	created.PasswordHash = user.PasswordHash
	created.PasswordAlgo = user.PasswordAlgo
	created.LastPasswordChange = user.LastPasswordChange
	created.DisplayName = user.DisplayName
	created.Profile = user.Profile
	created.IsActive = user.IsActive
	created.IsLocked = user.IsLocked
	created.MFAEnabled = user.MFAEnabled
	created.Metadata = user.Metadata

	r.db.logger.Debug(fmt.Sprintf("user created id=%d orbit=%d username=%s", created.ID, created.OrbitID, created.Username))
	return &created, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	r.db.logger.Trace(fmt.Sprintf("get user by id=%d", id))
	row := r.db.pool.QueryRow(ctx, sql_scripts.SelectUserByID, id)
	user, err := scanUser(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, common.ErrNotFound
		}
		r.db.logger.Error(fmt.Sprintf("get user by id failed: %v", err))
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, orbitID int64, email string) (*models.User, error) {
	r.db.logger.Trace(fmt.Sprintf("get user by email orbit=%d email=%s", orbitID, email))
	row := r.db.pool.QueryRow(ctx, sql_scripts.SelectUserByEmail, orbitID, email)
	user, err := scanUser(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, common.ErrNotFound
		}
		r.db.logger.Error(fmt.Sprintf("get user by email failed: %v", err))
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return user, nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, orbitID int64, username string) (*models.User, error) {
	r.db.logger.Trace(fmt.Sprintf("get user by username orbit=%d username=%s", orbitID, username))
	row := r.db.pool.QueryRow(ctx, sql_scripts.SelectUserByUsername, orbitID, username)
	user, err := scanUser(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, common.ErrNotFound
		}
		r.db.logger.Error(fmt.Sprintf("get user by username failed: %v", err))
		return nil, fmt.Errorf("get user by username: %w", err)
	}
	return user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	r.db.logger.Trace(fmt.Sprintf("user update started id=%d", user.ID))

	user.UpdatedAt = time.Now().UTC()

	err := r.db.WithTx(ctx, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, sql_scripts.UpdateUser,
			user.ID,
			user.Username,
			user.Email,
			user.EmailVerified,
			user.PasswordHash,
			user.PasswordAlgo,
			user.LastPasswordChange,
			user.DisplayName,
			user.Profile,
			user.IsActive,
			user.IsLocked,
			user.MFAEnabled,
			user.Metadata,
			user.UpdatedAt,
		)
		var id int64
		var updatedAt time.Time
		if err := row.Scan(&id, &updatedAt); err != nil {
			return err
		}
		user.UpdatedAt = updatedAt
		return nil
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return common.ErrNotFound
		}
		r.db.logger.Error(fmt.Sprintf("user update failed: %v", err))
		return fmt.Errorf("update user: %w", err)
	}

	r.db.logger.Debug(fmt.Sprintf("user updated id=%d", user.ID))
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	r.db.logger.Trace(fmt.Sprintf("user delete started id=%d", id))

	now := time.Now().UTC()
	err := r.db.WithTx(ctx, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, sql_scripts.SoftDeleteUser, id, now)
		var deletedID int64
		return row.Scan(&deletedID)
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return common.ErrNotFound
		}
		r.db.logger.Error(fmt.Sprintf("user delete failed: %v", err))
		return fmt.Errorf("delete user: %w", err)
	}

	r.db.logger.Debug(fmt.Sprintf("user deleted id=%d", id))
	return nil
}

func (r *UserRepository) ListByOrbit(ctx context.Context, orbitID int64, p common.Pagination) ([]models.User, error) {
	p = p.Normalize()
	r.db.logger.Trace(fmt.Sprintf("list users orbit=%d limit=%d offset=%d", orbitID, p.Limit, p.Offset))

	rows, err := r.db.pool.Query(ctx, sql_scripts.ListUsersByOrbit, orbitID, p.Limit, p.Offset)
	if err != nil {
		r.db.logger.Error(fmt.Sprintf("list users query failed: %v", err))
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var result []models.User

	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			r.db.logger.Error(fmt.Sprintf("scan user failed: %v", err))
			return nil, fmt.Errorf("scan user: %w", err)
		}
		result = append(result, *user)
	}

	if err := rows.Err(); err != nil {
		r.db.logger.Error(fmt.Sprintf("rows error: %v", err))
		return nil, fmt.Errorf("rows error: %w", err)
	}

	r.db.logger.Debug(fmt.Sprintf("listed users count=%d orbit=%d", len(result), orbitID))
	return result, nil
}

func scanUser(row pgx.Row) (*models.User, error) {
	var user models.User

	err := row.Scan(
		&user.ID,
		&user.OrbitID,
		&user.Username,
		&user.Email,
		&user.EmailVerified,
		&user.PasswordHash,
		&user.PasswordAlgo,
		&user.LastPasswordChange,
		&user.DisplayName,
		&user.Profile,
		&user.IsActive,
		&user.IsLocked,
		&user.MFAEnabled,
		&user.Metadata,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
