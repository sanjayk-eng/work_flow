package postgres

import (
	"context"

	"github/sanjay-khandelwal/internal/shared/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool *pgxpool.Pool
}

func New(ctx context.Context, cfg config.DBConfig) (*DB, error) {
	poolConfig, err := pgxpool.ParseConfig(BuildDSN(cfg))
	if err != nil {
		return nil, err
	}

	poolConfig.MaxConns = DefaultMaxConns
	poolConfig.MinConns = DefaultMinConns
	poolConfig.MaxConnLifetime = DefaultMaxConnLifetime
	poolConfig.MaxConnIdleTime = DefaultMaxConnIdleTime
	poolConfig.HealthCheckPeriod = DefaultHealthCheckPeriod

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	return &DB{Pool: pool}, nil
}

func (db *DB) Close() {
	db.Pool.Close()
}
