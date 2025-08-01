package main

import (
	"flag"
	"log"

	"github.com/a5c-ai/hub/internal/config"
	"github.com/a5c-ai/hub/internal/db"
	"github.com/a5c-ai/hub/internal/db/seeds"
)

func main() {
	var (
		rollback = flag.Bool("rollback", false, "Rollback the last migration")
		seed     = flag.Bool("seed", false, "Seed the database with development data")
	)
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	database, err := db.Connect(cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()

	if *rollback {
		log.Println("Rolling back last migration...")
		if err := database.Rollback(); err != nil {
			log.Fatal("Failed to rollback migration:", err)
		}
		log.Println("Migration rollback completed successfully")
		return
	}

	log.Println("Running database migrations...")
	if err := database.Migrate(); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}
	log.Println("Database migrations completed successfully")

	if *seed {
		log.Println("Seeding database with development data...")
		if err := seeds.SeedDatabase(database.DB); err != nil {
			log.Fatal("Failed to seed database:", err)
		}
		log.Println("Database seeding completed successfully")
	}
}
