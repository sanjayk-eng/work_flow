package main

import (
	"context"
	"log"

	"github/sanjay-khandelwal/internal/modules/iam"
	"github/sanjay-khandelwal/internal/shared/core/config"
	"github/sanjay-khandelwal/internal/shared/core/database/postgres"
	"github/sanjay-khandelwal/internal/shared/core/server"
)

func main() {
	cfg := config.Get()

	db, err := postgres.New(context.Background(), cfg.DB)
	if err != nil {
		log.Fatalf("failed to connect database %v", err)
	}
	defer db.Close()

	if err := postgres.MigrateUp(cfg.DB, "../../migrations"); err != nil {
		log.Fatalf("failed to run migrations %v", err)
	}

	srv := server.New()
	api := srv.Group("/api/v1")

	iam.New(db, cfg.JWT).Mount(api)

	if err := srv.Run(cfg.App); err != nil {
		log.Fatalf("server error %v", err)
	}
}
