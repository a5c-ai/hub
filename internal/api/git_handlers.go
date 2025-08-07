package api

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/a5c-ai/hub/internal/auth"
	"github.com/a5c-ai/hub/internal/models"
	"github.com/a5c-ai/hub/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// GitHandlers contains handlers for Git HTTP protocol endpoints
type GitHandlers struct {
	repositoryService services.RepositoryService
	logger            *logrus.Logger
	jwtManager        *auth.JWTManager
}

// NewGitHandlers creates a new Git handlers instance
func NewGitHandlers(repositoryService services.RepositoryService, logger *logrus.Logger, jwtManager *auth.JWTManager) *GitHandlers {
	return &GitHandlers{
		repositoryService: repositoryService,
		logger:            logger,
		jwtManager:        jwtManager,
	}
}

// InfoRefs handles GET /{owner}/{repo.git}/info/refs
func (h *GitHandlers) InfoRefs(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo") // Already without .git suffix since route pattern is :repo.git
	service := c.Query("service")

	h.logger.WithFields(logrus.Fields{
		"owner":   owner,
		"repo":    repoName,
		"service": service,
	}).Info("Git info/refs request")

	if repoName == "" {
		h.logger.Error("Repository name is empty")
		c.Status(http.StatusBadRequest)
		return
	}

	// Get repository
	repo, err := h.repositoryService.Get(c.Request.Context(), owner, repoName)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"owner": owner,
			"repo":  repoName,
		}).Error("Failed to get repository in git handler")
		if err.Error() == "repository not found" {
			c.Status(http.StatusNotFound)
		} else {
			h.logger.WithError(err).Error("Failed to get repository")
			c.Status(http.StatusInternalServerError)
		}
		return
	}

	h.logger.WithFields(logrus.Fields{
		"repo_id": repo.ID,
		"owner":   owner,
		"repo":    repoName,
	}).Info("Repository found in database")

	// Get repository path
	repoPath, err := h.repositoryService.GetRepositoryPath(c.Request.Context(), repo.ID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get repository path")
		c.Status(http.StatusInternalServerError)
		return
	}

	h.logger.WithField("path", repoPath).Info("Repository path determined")

	// Check if repository exists on filesystem, create if needed
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		h.logger.WithField("path", repoPath).Info("Repository doesn't exist on filesystem, initializing")
		if err := h.repositoryService.InitializeGitRepository(c.Request.Context(), repo.ID); err != nil {
			h.logger.WithError(err).Error("Failed to initialize Git repository")
			c.Status(http.StatusInternalServerError)
			return
		}
		h.logger.WithField("path", repoPath).Info("Repository initialized successfully")
	} else {
		h.logger.WithField("path", repoPath).Info("Repository already exists on filesystem")
	}

	switch service {
	case "git-upload-pack":
		h.handleUploadPackInfoRefs(c, repoPath)
	case "git-receive-pack":
		h.handleReceivePackInfoRefs(c, repoPath)
	default:
		h.handleDumbInfoRefs(c, repoPath)
	}
}

// UploadPack handles POST /{owner}/{repo.git}/git-upload-pack
func (h *GitHandlers) UploadPack(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo") // Already without .git suffix since route pattern is :repo.git

	h.logger.WithFields(logrus.Fields{
		"owner": owner,
		"repo":  repoName,
	}).Info("Git upload-pack request")

	// Get repository
	repo, err := h.repositoryService.Get(c.Request.Context(), owner, repoName)
	if err != nil {
		if err.Error() == "repository not found" {
			c.Status(http.StatusNotFound)
		} else {
			h.logger.WithError(err).Error("Failed to get repository")
			c.Status(http.StatusInternalServerError)
		}
		return
	}

	// Enforce authentication for private repositories
	if repo.Visibility == models.VisibilityPrivate {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer {token}"})
			return
		}
		if _, err := h.jwtManager.ValidateToken(parts[1]); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}
	}

	// Get repository path
	repoPath, err := h.repositoryService.GetRepositoryPath(c.Request.Context(), repo.ID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get repository path")
		c.Status(http.StatusInternalServerError)
		return
	}

	// Ensure repository exists
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		h.logger.WithError(err).Error("Repository path does not exist")
		c.Status(http.StatusNotFound)
		return
	}

	h.handleGitCommand(c, repoPath, "git-upload-pack", "--stateless-rpc", repoPath)
}

// ReceivePack handles POST /{owner}/{repo.git}/git-receive-pack
func (h *GitHandlers) ReceivePack(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo") // Already without .git suffix since route pattern is :repo.git

	h.logger.WithFields(logrus.Fields{
		"owner": owner,
		"repo":  repoName,
	}).Info("Git receive-pack request")

	// Get repository
	repo, err := h.repositoryService.Get(c.Request.Context(), owner, repoName)
	if err != nil {
		if err.Error() == "repository not found" {
			c.Status(http.StatusNotFound)
		} else {
			h.logger.WithError(err).Error("Failed to get repository")
			c.Status(http.StatusInternalServerError)
		}
		return
	}

	// Enforce authentication for push operations
	{
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer {token}"})
			return
		}
		if _, err := h.jwtManager.ValidateToken(parts[1]); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}
	}

	// Get repository path
	repoPath, err := h.repositoryService.GetRepositoryPath(c.Request.Context(), repo.ID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get repository path")
		c.Status(http.StatusInternalServerError)
		return
	}

	// Ensure repository exists
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		h.logger.WithError(err).Error("Repository path does not exist")
		c.Status(http.StatusNotFound)
		return
	}

	h.handleGitCommand(c, repoPath, "git-receive-pack", "--stateless-rpc", repoPath)
}

// Helper methods

func (h *GitHandlers) handleUploadPackInfoRefs(c *gin.Context, repoPath string) {
	c.Header("Content-Type", "application/x-git-upload-pack-advertisement")
	c.Header("Cache-Control", "no-cache")

	// Write packet header
	c.Writer.Write(h.packetWrite("# service=git-upload-pack\n"))
	c.Writer.Write([]byte("0000"))

	// Execute git-upload-pack command from the repository directory
	cmd := exec.Command("git", "upload-pack", "--stateless-rpc", "--advertise-refs", ".")
	cmd.Dir = repoPath

	h.logger.WithFields(logrus.Fields{
		"command": cmd.String(),
		"dir":     repoPath,
		"args":    cmd.Args,
	}).Info("Executing git upload-pack command")

	output, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			h.logger.WithError(err).WithField("stderr", string(exitError.Stderr)).Error("Failed to execute git-upload-pack")
		} else {
			h.logger.WithError(err).Error("Failed to execute git-upload-pack")
		}
		c.Status(http.StatusInternalServerError)
		return
	}

	h.logger.WithField("output_size", len(output)).Info("Git upload-pack completed successfully")
	c.Writer.Write(output)
}

func (h *GitHandlers) handleReceivePackInfoRefs(c *gin.Context, repoPath string) {
	c.Header("Content-Type", "application/x-git-receive-pack-advertisement")
	c.Header("Cache-Control", "no-cache")

	// Write packet header
	c.Writer.Write(h.packetWrite("# service=git-receive-pack\n"))
	c.Writer.Write([]byte("0000"))

	// Execute git receive-pack command from the repository directory
	cmd := exec.Command("git", "receive-pack", "--stateless-rpc", "--advertise-refs", ".")
	cmd.Dir = repoPath

	h.logger.WithFields(logrus.Fields{
		"command": cmd.String(),
		"dir":     repoPath,
		"args":    cmd.Args,
	}).Info("Executing git receive-pack command")

	output, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			h.logger.WithError(err).WithField("stderr", string(exitError.Stderr)).Error("Failed to execute git-receive-pack")
		} else {
			h.logger.WithError(err).Error("Failed to execute git-receive-pack")
		}
		c.Status(http.StatusInternalServerError)
		return
	}

	h.logger.WithField("output_size", len(output)).Info("Git receive-pack completed successfully")
	c.Writer.Write(output)
}

func (h *GitHandlers) handleDumbInfoRefs(c *gin.Context, repoPath string) {
	refsPath := filepath.Join(repoPath, "info", "refs")

	// Update info/refs file
	cmd := exec.Command("git", "update-server-info")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		h.logger.WithError(err).Error("Failed to update server info")
	}

	c.Header("Content-Type", "text/plain")
	c.Header("Cache-Control", "no-cache")

	if _, err := os.Stat(refsPath); os.IsNotExist(err) {
		c.Status(http.StatusNotFound)
		return
	}

	c.File(refsPath)
}

func (h *GitHandlers) handleGitCommand(c *gin.Context, repoPath, command string, args ...string) {
	// Set appropriate content type
	var contentType string
	switch command {
	case "git-upload-pack":
		contentType = "application/x-git-upload-pack-result"
	case "git-receive-pack":
		contentType = "application/x-git-receive-pack-result"
	default:
		contentType = "application/octet-stream"
	}

	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "no-cache")

	// Create command - use "git" as the base command and add proper args
	var cmdArgs []string
	switch command {
	case "git-upload-pack":
		cmdArgs = []string{"upload-pack", "--stateless-rpc", "."}
	case "git-receive-pack":
		cmdArgs = []string{"receive-pack", "--stateless-rpc", "."}
	default:
		cmdArgs = append([]string{command}, args...)
	}

	cmd := exec.Command("git", cmdArgs...)
	cmd.Dir = repoPath

	h.logger.WithFields(logrus.Fields{
		"command": cmd.String(),
		"args":    cmd.Args,
		"dir":     repoPath,
	}).Info("Executing git command")

	// Set up pipes
	stdin, err := cmd.StdinPipe()
	if err != nil {
		h.logger.WithError(err).Error("Failed to create stdin pipe")
		c.Status(http.StatusInternalServerError)
		return
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		h.logger.WithError(err).Error("Failed to create stdout pipe")
		c.Status(http.StatusInternalServerError)
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		h.logger.WithError(err).Error("Failed to create stderr pipe")
		c.Status(http.StatusInternalServerError)
		return
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		h.logger.WithError(err).Error("Failed to start git command")
		c.Status(http.StatusInternalServerError)
		return
	}

	// Handle request body (input to git command)
	go func() {
		defer stdin.Close()

		var reader io.Reader = c.Request.Body

		// Handle gzip compression
		if c.GetHeader("Content-Encoding") == "gzip" {
			if gzipReader, err := gzip.NewReader(c.Request.Body); err == nil {
				defer gzipReader.Close()
				reader = gzipReader
			}
		}

		io.Copy(stdin, reader)
	}()

	// Stream stdout to response
	go func() {
		io.Copy(c.Writer, stdout)
	}()

	// Log stderr
	go func() {
		stderrBytes, err := io.ReadAll(stderr)
		if err == nil && len(stderrBytes) > 0 {
			h.logger.WithField("stderr", string(stderrBytes)).Info("Git command stderr")
		}
	}()

	// Wait for command to complete
	if err := cmd.Wait(); err != nil {
		h.logger.WithError(err).Error("Git command failed")
		return
	}

	h.logger.Info("Git command completed successfully")
}

// packetWrite formats data according to Git packet-line format
func (h *GitHandlers) packetWrite(data string) []byte {
	length := len(data) + 4
	return []byte(fmt.Sprintf("%04x%s", length, data))
}

// GitMiddleware adds Git-specific headers and logging
func (h *GitHandlers) GitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Log Git requests
		h.logger.WithFields(logrus.Fields{
			"method":     c.Request.Method,
			"url":        c.Request.URL.String(),
			"user_agent": c.GetHeader("User-Agent"),
			"remote_ip":  c.ClientIP(),
		}).Info("Git HTTP request")

		// Set Git-specific headers
		c.Header("Expires", "Fri, 01 Jan 1980 00:00:00 GMT")
		c.Header("Pragma", "no-cache")
		c.Header("Cache-Control", "no-cache, max-age=0, must-revalidate")

		c.Next()
	}
}
