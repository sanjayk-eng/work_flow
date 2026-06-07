package main

import (
	"context"
	"log"

	auth "github/sanjay-khandelwal/internal/modules/auth"
	"github/sanjay-khandelwal/internal/modules/user"
	"github/sanjay-khandelwal/internal/shared/core/config"
	"github/sanjay-khandelwal/internal/shared/core/database/postgres"
	"github/sanjay-khandelwal/internal/shared/core/server"
)

func main() {
	// load env config
	cfg := config.Get()

	// Database
	db, err := postgres.New(context.Background(), cfg.DB)
	if err != nil {
		log.Fatalf("failed to connect database %v", err)
	}
	defer db.Close()

	// Migrations — path from config, not hardcoded
	if err := postgres.MigrateUp(cfg.DB, "../../migrations"); err != nil {
		log.Fatalf("failed to run migrations %v", err)

	}

	// Server
	srv := server.New()
	api := srv.Group("/api/v1")

	userModule := user.New(db)

	// Mount modules
	auth.New(db, userModule.Service).Mount(api)

	// Start
	if err := srv.Run(cfg.App); err != nil {
		log.Fatalf("server error %v", err)
	}
}
