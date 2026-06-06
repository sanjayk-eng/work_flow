package main

import (
	"context"
	"github/sanjay-khandelwal/internal/shared/core/config"
	"github/sanjay-khandelwal/internal/shared/core/database/postgres"
	"github/sanjay-khandelwal/internal/shared/core/server"
	"log"
)

func main() {
	// load the configuration
	config := config.Get()

	//Db connection
	db, err := postgres.New(context.Background(), config.DB)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// migrate the database
	if err := postgres.MigrateUp(config.DB, "../../migrations"); err != nil {
		log.Fatal(err)
	}

	// create and run the server
	srv := server.New()
	if err := srv.Run(config.App); err != nil {
		log.Fatal(err)
	}
}
