package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/a5c-ai/hub/internal/services"
)

// GitHubHandlers handles GitHub-specific repository creation and import endpoints.
type GitHubHandlers struct {
	githubService *services.GitHubService
	logger        *logrus.Logger
}

// NewGitHubHandlers initializes GitHubHandlers with the given service and logger.
func NewGitHubHandlers(githubService *services.GitHubService, logger *logrus.Logger) *GitHubHandlers {
	return &GitHubHandlers{githubService: githubService, logger: logger}
}

// InitiateGitHubCreate handles POST /api/v1/repositories/github/create
// Response: 501 Not Implemented until feature is supported.
func (h *GitHubHandlers) InitiateGitHubCreate(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "GitHub repository creation not implemented"})
}

// InitiateGitHubImport handles POST /api/v1/repositories/github/import
// Response: 501 Not Implemented until feature is supported.
func (h *GitHubHandlers) InitiateGitHubImport(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "GitHub repository import not implemented"})
}
