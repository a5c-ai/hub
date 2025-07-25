package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/a5c-ai/hub/internal/config"
	"github.com/a5c-ai/hub/internal/db"
	"github.com/a5c-ai/hub/internal/git"
	"github.com/a5c-ai/hub/internal/services"
	"github.com/a5c-ai/hub/internal/ssh"
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

	// Initialize database
	database, err := db.Connect(cfg.Database)
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to database")
	}

	// Run migrations if needed
	if err := database.Migrate(); err != nil {
		logger.WithError(err).Fatal("Failed to run migrations")
	}

	// Initialize services
	gitService := git.NewGitService(logger)
	repoBasePath := cfg.Storage.RepositoryPath
	if repoBasePath == "" {
		repoBasePath = "./repositories"
	}

	repositoryService := services.NewRepositoryService(database.DB, gitService, logger, repoBasePath)

	// Initialize git shell service
	gitShell := ssh.NewGitShellService(logger)

	// Configure SSH server
	sshConfig := ssh.SSHServerConfig{
		Port:        cfg.SSH.Port,
		HostKeyPath: cfg.SSH.HostKeyPath,
	}

	// Create SSH server adapter
	sshRepoService := ssh.NewRepositoryServiceAdapter(repositoryService)

	// Create SSH server
	sshServer, err := ssh.NewSSHServer(
		sshConfig,
		sshRepoService,
		gitShell,
		logger,
		database.DB,
	)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize SSH server")
	}

	// Context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start SSH server
	go func() {
		logger.WithField("port", cfg.SSH.Port).Info("Starting SSH server")
		if err := sshServer.Start(ctx); err != nil {
			logger.WithError(err).Error("SSH server error")
		}
	}()

	fmt.Printf("🚀 SSH Git Server started on port %d\n", cfg.SSH.Port)
	fmt.Printf("📝 To test SSH clone:\n")
	fmt.Printf("   1. Add your SSH key via the API\n")
	fmt.Printf("   2. git clone ssh://username@localhost:%d/owner/repo.git\n", cfg.SSH.Port)
	fmt.Printf("💡 Use Ctrl+C to stop\n\n")

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down SSH server...")

	// Cancel context to stop SSH server
	cancel()

	// Stop SSH server
	if err := sshServer.Stop(); err != nil {
		logger.WithError(err).Error("Error stopping SSH server")
	}

	logger.Info("SSH server stopped")
}
