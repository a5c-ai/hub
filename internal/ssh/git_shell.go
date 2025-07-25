package ssh

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/sirupsen/logrus"
)

// gitShellService implements GitShellService for handling git commands
type gitShellService struct {
	logger *logrus.Logger
}

// NewGitShellService creates a new git shell service
func NewGitShellService(logger *logrus.Logger) GitShellService {
	return &gitShellService{
		logger: logger,
	}
}

// HandleGitCommand executes a git command for SSH access
func (g *gitShellService) HandleGitCommand(
	ctx context.Context,
	command string,
	repoPath string,
	stdin io.Reader,
	stdout, stderr io.Writer,
) error {
	g.logger.WithFields(logrus.Fields{
		"command":   command,
		"repo_path": repoPath,
	}).Info("Executing git command via SSH")

	// Validate repository exists
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		return fmt.Errorf("repository not found: %s", repoPath)
	}

	// Prepare git command
	var cmd *exec.Cmd
	switch command {
	case "git-upload-pack":
		cmd = exec.CommandContext(ctx, "git", "upload-pack", "--stateless-rpc", ".")
	case "git-receive-pack":
		cmd = exec.CommandContext(ctx, "git", "receive-pack", "--stateless-rpc", ".")
	default:
		return fmt.Errorf("unsupported git command: %s", command)
	}

	// Set working directory to repository path
	cmd.Dir = repoPath

	// Connect streams
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	g.logger.WithFields(logrus.Fields{
		"command": cmd.String(),
		"dir":     cmd.Dir,
		"args":    cmd.Args,
	}).Debug("Starting git command")

	// Execute command
	if err := cmd.Run(); err != nil {
		g.logger.WithError(err).Error("Git command failed")
		return fmt.Errorf("git command failed: %w", err)
	}

	g.logger.Info("Git command completed successfully")
	return nil
}
