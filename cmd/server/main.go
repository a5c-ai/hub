package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/a5c-ai/hub/internal/api"
	"github.com/a5c-ai/hub/internal/config"
	"github.com/a5c-ai/hub/internal/db"
	"github.com/a5c-ai/hub/internal/git"
	"github.com/a5c-ai/hub/internal/services"
	"github.com/a5c-ai/hub/internal/ssh"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
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

	// Run migrations
	if err := database.Migrate(); err != nil {
		logger.WithError(err).Fatal("Failed to run migrations")
	}

	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Setup HTTP router
	router := gin.Default()

	// Setup CORS middleware
	router.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		for _, allowedOrigin := range cfg.CORS.AllowedOrigins {
			if origin == allowedOrigin {
				c.Header("Access-Control-Allow-Origin", origin)
				break
			}
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// Setup API routes
	api.SetupRoutes(router, database, logger)

	// Create HTTP server
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: router,
	}

	// Initialize SSH server if enabled
	var sshServer *ssh.SSHServer
	if cfg.SSH.Enabled {
		// Initialize services with distributed config support
		distributedConfig := &git.DistributedConfig{
			Enabled:             cfg.Storage.Distributed.Enabled,
			NodeID:              cfg.Storage.Distributed.NodeID,
			ReplicationCount:    cfg.Storage.Distributed.ReplicationCount,
			ConsistentHashing:   cfg.Storage.Distributed.ConsistentHashing,
			HealthCheckInterval: parseHealthCheckInterval(cfg.Storage.Distributed.HealthCheckInterval),
		}

		// Convert storage nodes from config format to git format
		for _, node := range cfg.Storage.Distributed.StorageNodes {
			distributedConfig.StorageNodes = append(distributedConfig.StorageNodes, git.StorageNode{
				ID:      node.ID,
				Address: node.Address,
				Weight:  node.Weight,
			})
		}

		gitService := git.NewGitServiceWithConfig(distributedConfig, logger)
		repoBasePath := cfg.Storage.RepositoryPath
		if repoBasePath == "" {
			repoBasePath = "./repositories"
		}

		repositoryService := services.NewRepositoryService(database.DB, gitService, logger, repoBasePath)

		// Initialize git shell service
		gitShell := ssh.NewGitShellService(logger)

		sshConfig := ssh.SSHServerConfig{
			Port:        cfg.SSH.Port,
			HostKeyPath: cfg.SSH.HostKeyPath,
		}

		// Create SSH server adapter
		sshRepoService := ssh.NewRepositoryServiceAdapter(repositoryService)

		sshServer, err = ssh.NewSSHServer(
			sshConfig,
			sshRepoService,
			gitShell,
			logger,
			database.DB,
		)
		if err != nil {
			logger.WithError(err).Fatal("Failed to initialize SSH server")
		}
	}

	// Context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start SSH server if enabled
	if sshServer != nil {
		go func() {
			logger.WithField("port", cfg.SSH.Port).Info("Starting SSH server")
			if err := sshServer.Start(ctx); err != nil {
				logger.WithError(err).Error("SSH server error")
			}
		}()
	}

	// Start HTTP server
	go func() {
		logger.WithField("port", cfg.Server.Port).Info("Starting HTTP server")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("HTTP server error")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down servers...")

	// Cancel context to stop SSH server
	cancel()

	// Graceful shutdown of HTTP server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.WithError(err).Error("HTTP server forced to shutdown")
	}

	// Stop SSH server if running
	if sshServer != nil {
		if err := sshServer.Stop(); err != nil {
			logger.WithError(err).Error("Error stopping SSH server")
		}
	}

	logger.Info("Servers stopped")
}

// parseHealthCheckInterval parses a string duration or returns a default
func parseHealthCheckInterval(interval string) time.Duration {
	if interval == "" {
		return 30 * time.Second
	}
	
	duration, err := time.ParseDuration(interval)
	if err != nil {
		return 30 * time.Second
	}
	
	return duration
}
