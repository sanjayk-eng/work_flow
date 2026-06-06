package main

import (
	"context"

	"github/sanjay-khandelwal/internal/shared/core/config"
	"github/sanjay-khandelwal/internal/shared/core/database/postgres"
	"github/sanjay-khandelwal/internal/shared/core/logger"
	"github/sanjay-khandelwal/internal/shared/core/server"
)

func main() {
	cfg := config.Get()

	log := logger.New(cfg.Log.Level)

	db, err := postgres.New(context.Background(), cfg.DB)
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		panic(err)
	}
	defer db.Close()

	if err := postgres.MigrateUp(cfg.DB, "../../migrations"); err != nil {
		log.Error("failed to run migrations", "error", err)
		panic(err)
	}

	srv := server.New(log)
	if err := srv.Run(cfg.App); err != nil {
		log.Error("server error", "error", err)
		panic(err)
	}
}
