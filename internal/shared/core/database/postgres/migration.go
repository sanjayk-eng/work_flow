package postgres

import (
	"database/sql"

	"github/sanjay-khandelwal/internal/shared/core/config"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func setDialect() error {
	return goose.SetDialect("postgres")
}

func openMigrationDB(cfg config.DBConfig) (*sql.DB, error) {
	pgConfig, err := pgx.ParseConfig(BuildDSN(cfg))
	if err != nil {
		return nil, err
	}
	return stdlib.OpenDB(*pgConfig), nil
}

func MigrateUp(cfg config.DBConfig, dir string) error {
	if err := setDialect(); err != nil {
		return err
	}
	db, err := openMigrationDB(cfg)
	if err != nil {
		return err
	}
	defer db.Close()
	return goose.Up(db, dir)
}

func MigrateDown(cfg config.DBConfig, dir string) error {
	if err := setDialect(); err != nil {
		return err
	}
	db, err := openMigrationDB(cfg)
	if err != nil {
		return err
	}
	defer db.Close()
	return goose.Down(db, dir)
}
