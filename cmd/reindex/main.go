package main

import (
	"flag"
	"log"

	"github.com/a5c-ai/hub/internal/config"
	"github.com/a5c-ai/hub/internal/db"
	"github.com/a5c-ai/hub/internal/services"
	"github.com/sirupsen/logrus"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "", "Path to config file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Setup logger
	logger := logrus.New()
	logger.SetLevel(logrus.Level(cfg.LogLevel))
	logger.SetFormatter(&logrus.JSONFormatter{})

	// Initialize database
	database, err := db.Connect(cfg.Database)
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to database")
	}
	defer database.Close()

	// Initialize Elasticsearch service
	elasticsearchService, err := services.NewElasticsearchService(&cfg.Elasticsearch, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize Elasticsearch service")
	}

	if !elasticsearchService.IsEnabled() {
		logger.Fatal("Elasticsearch is not enabled in configuration")
	}

	// Initialize search service
	searchService := services.NewSearchService(database.DB, elasticsearchService, logger)

	// Perform reindex
	logger.Info("Starting reindex operation...")
	if err := searchService.ReindexAll(); err != nil {
		logger.WithError(err).Fatal("Reindex operation failed")
	}

	logger.Info("Reindex operation completed successfully")
}