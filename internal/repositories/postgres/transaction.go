package postgres

import (
	"context"
	"fmt"

	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/utils/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	pool   *pgxpool.Pool
	logger *logger.Logger
}

func NewDB(pool *pgxpool.Pool, l *logger.Logger) *DB {
	return &DB{pool: pool, logger: l}
}

func InTransaction[T any](db *DB, ctx context.Context, fn func(tx pgx.Tx) (T, error)) (T, error) {
	var zero T

	tx, err := db.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		db.logger.Error(fmt.Sprintf("begin tx failed: %v", err))
		return zero, err
	}

	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback(ctx)
			panic(r)
		}
	}()

	res, err := fn(tx)
	if err != nil {
		_ = tx.Rollback(ctx)
		db.logger.Error(fmt.Sprintf("tx function failed: %v", err))
		return zero, err
	}

	if err := tx.Commit(ctx); err != nil {
		db.logger.Error(fmt.Sprintf("tx commit failed: %v", err))
		return zero, err
	}

	return res, nil
}
