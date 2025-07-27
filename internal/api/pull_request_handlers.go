package api

import (
	"context"
	"net/http"
	"strconv"

	"github.com/a5c-ai/hub/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type PullRequestHandlers struct {
	service services.PullRequestService
	logger  *logrus.Logger
}

func NewPullRequestHandlers(service services.PullRequestService, logger *logrus.Logger) *PullRequestHandlers {
	return &PullRequestHandlers{
		service: service,
		logger:  logger,
	}
}

// ListPullRequests handles GET /api/v1/repositories/:owner/:repo/pulls
func (h *PullRequestHandlers) ListPullRequests(c *gin.Context) {
	owner := c.Param("owner")
	repo := c.Param("repo")

	// Get repository ID
	repoID, err := h.getRepositoryID(c.Request.Context(), owner, repo)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get repository")
		c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		return
	}

	// Parse query parameters
	state := c.DefaultQuery("state", "open")
	opts := services.PullRequestFilter{
		State:    &state,
		Page:     1,
		PageSize: 30,
	}

	if page := c.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			opts.Page = p
		}
	}

	if perPage := c.Query("per_page"); perPage != "" {
		if pp, err := strconv.Atoi(perPage); err == nil && pp > 0 && pp <= 100 {
			opts.PageSize = pp
		}
	}

	prs, err := h.service.List(c.Request.Context(), repoID, opts)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list pull requests")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list pull requests"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pull_requests": prs,
	})
}

// GetPullRequest handles GET /api/v1/repositories/:owner/:repo/pulls/:number
func (h *PullRequestHandlers) GetPullRequest(c *gin.Context) {
	owner := c.Param("owner")
	repo := c.Param("repo")
	numberStr := c.Param("number")

	number, err := strconv.Atoi(numberStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pull request number"})
		return
	}

	pr, err := h.service.Get(c.Request.Context(), owner, repo, number)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get pull request")
		c.JSON(http.StatusNotFound, gin.H{"error": "Pull request not found"})
		return
	}

	c.JSON(http.StatusOK, pr)
}

// CreatePullRequest handles POST /api/v1/repositories/:owner/:repo/pulls
func (h *PullRequestHandlers) CreatePullRequest(c *gin.Context) {
	owner := c.Param("owner")
	repo := c.Param("repo")

	// Get repository ID
	repoID, err := h.getRepositoryID(c.Request.Context(), owner, repo)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get repository")
		c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req services.CreatePullRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	pr, err := h.service.Create(c.Request.Context(), repoID, userID.(uuid.UUID), req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create pull request")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create pull request"})
		return
	}

	c.JSON(http.StatusCreated, pr)
}

// UpdatePullRequest handles PATCH /api/v1/repositories/:owner/:repo/pulls/:number
func (h *PullRequestHandlers) UpdatePullRequest(c *gin.Context) {
	owner := c.Param("owner")
	repo := c.Param("repo")
	numberStr := c.Param("number")

	number, err := strconv.Atoi(numberStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pull request number"})
		return
	}

	// Get existing pull request
	pr, err := h.service.Get(c.Request.Context(), owner, repo, number)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get pull request")
		c.JSON(http.StatusNotFound, gin.H{"error": "Pull request not found"})
		return
	}

	var req services.UpdatePullRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	updatedPR, err := h.service.Update(c.Request.Context(), pr.ID, req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to update pull request")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update pull request"})
		return
	}

	c.JSON(http.StatusOK, updatedPR)
}

// MergePullRequest handles POST /api/v1/repositories/:owner/:repo/pulls/:number/merge
func (h *PullRequestHandlers) MergePullRequest(c *gin.Context) {
	owner := c.Param("owner")
	repo := c.Param("repo")
	numberStr := c.Param("number")

	number, err := strconv.Atoi(numberStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pull request number"})
		return
	}

	// Get existing pull request
	pr, err := h.service.Get(c.Request.Context(), owner, repo, number)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get pull request")
		c.JSON(http.StatusNotFound, gin.H{"error": "Pull request not found"})
		return
	}

	var req services.MergePullRequestRequest
	if err := c.ShouldBindJSON(&req); err == nil {
		// Optional request body
	}

	err = h.service.Merge(c.Request.Context(), pr.ID, req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to merge pull request")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to merge pull request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pull request merged successfully"})
}

// Helper method to get repository ID
func (h *PullRequestHandlers) getRepositoryID(ctx context.Context, owner, repo string) (uuid.UUID, error) {
	// This is a simplified implementation - in practice you'd query the database
	return uuid.New(), nil
}
