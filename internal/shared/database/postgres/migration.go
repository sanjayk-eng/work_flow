package postgres

import (
	"database/sql"

	"github/sanjay-khandelwal/internal/shared/config"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// openDB creates a new DB connection for migrations
func openDB(cfg config.DBConfig) (*sql.DB, error) {
	pgConfig, err := pgx.ParseConfig(BuildDSN(cfg))
	if err != nil {
		return nil, err
	}

	return stdlib.OpenDB(*pgConfig), nil
}

// set goose dialect once
func init() {
	
	_ = goose.SetDialect("postgres")
}

// MigrateUp runs all up migrations
func MigrateUp(cfg config.DBConfig, dir string) error {
	db, err := openDB(cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	return goose.Up(db, dir)
}

// MigrateDown rolls back all migrations
func MigrateDown(cfg config.DBConfig, dir string) error {
	db, err := openDB(cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	return goose.Down(db, dir)
}

// MigrateTo migrates to a specific version (recommended)
func MigrateTo(cfg config.DBConfig, dir string, version int64) error {
	db, err := openDB(cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	return goose.UpTo(db, dir, version)
}
