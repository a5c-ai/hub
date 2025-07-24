package api

import (
	"net/http"
	"strconv"

	"github.com/a5c-ai/hub/internal/models"
	"github.com/a5c-ai/hub/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type PullRequestHandlers struct {
	service *services.PullRequestService
	logger  *logrus.Logger
}

func NewPullRequestHandlers(service *services.PullRequestService, logger *logrus.Logger) *PullRequestHandlers {
	return &PullRequestHandlers{
		service: service,
		logger:  logger,
	}
}

// ListPullRequests handles GET /api/v1/repositories/:owner/:repo/pulls
func (h *PullRequestHandlers) ListPullRequests(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")

	// Get repository ID from database
	repoID, err := h.getRepositoryID(c, owner, repoName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		return
	}

	// Parse query parameters
	opts := services.PullRequestListOptions{
		State:     c.DefaultQuery("state", "open"),
		Head:      c.Query("head"),
		Base:      c.Query("base"),
		Sort:      c.DefaultQuery("sort", "created"),
		Direction: c.DefaultQuery("direction", "desc"),
		Page:      1,
		PerPage:   30,
	}

	if page := c.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			opts.Page = p
		}
	}

	if perPage := c.Query("per_page"); perPage != "" {
		if pp, err := strconv.Atoi(perPage); err == nil && pp > 0 && pp <= 100 {
			opts.PerPage = pp
		}
	}

	prs, total, err := h.service.ListPullRequests(repoID, opts)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list pull requests")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list pull requests"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pull_requests": prs,
		"total_count":   total,
		"page":          opts.Page,
		"per_page":      opts.PerPage,
	})
}

// GetPullRequest handles GET /api/v1/repositories/:owner/:repo/pulls/:number
func (h *PullRequestHandlers) GetPullRequest(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")
	numberStr := c.Param("number")

	number, err := strconv.Atoi(numberStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pull request number"})
		return
	}

	repoID, err := h.getRepositoryID(c, owner, repoName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		return
	}

	pr, err := h.service.GetPullRequest(repoID, number)
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
	repoName := c.Param("repo")

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	repoID, err := h.getRepositoryID(c, owner, repoName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		return
	}

	var req services.CreatePullRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pr, err := h.service.CreatePullRequest(repoID, userID.(uuid.UUID), req)
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
	repoName := c.Param("repo")
	numberStr := c.Param("number")

	number, err := strconv.Atoi(numberStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pull request number"})
		return
	}

	repoID, err := h.getRepositoryID(c, owner, repoName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pr, err := h.service.UpdatePullRequest(repoID, number, updates)
	if err != nil {
		h.logger.WithError(err).Error("Failed to update pull request")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update pull request"})
		return
	}

	c.JSON(http.StatusOK, pr)
}

// MergePullRequest handles PUT /api/v1/repositories/:owner/:repo/pulls/:number/merge
func (h *PullRequestHandlers) MergePullRequest(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")
	numberStr := c.Param("number")

	number, err := strconv.Atoi(numberStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pull request number"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	repoID, err := h.getRepositoryID(c, owner, repoName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		return
	}

	var req struct {
		MergeMethod   string `json:"merge_method"`
		CommitTitle   string `json:"commit_title"`
		CommitMessage string `json:"commit_message"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Default merge method
	if req.MergeMethod == "" {
		req.MergeMethod = "merge"
	}

	mergeMethod := models.MergeMethod(req.MergeMethod)
	if mergeMethod != models.MergeMethodMerge && mergeMethod != models.MergeMethodSquash && mergeMethod != models.MergeMethodRebase {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid merge method"})
		return
	}

	pr, err := h.service.MergePullRequest(repoID, number, userID.(uuid.UUID), mergeMethod, req.CommitTitle, req.CommitMessage)
	if err != nil {
		h.logger.WithError(err).Error("Failed to merge pull request")
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sha":     pr.MergeCommitSHA,
		"merged":  pr.Merged,
		"message": "Pull request successfully merged",
	})
}

// GetPullRequestFiles handles GET /api/v1/repositories/:owner/:repo/pulls/:number/files
func (h *PullRequestHandlers) GetPullRequestFiles(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")
	numberStr := c.Param("number")

	number, err := strconv.Atoi(numberStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pull request number"})
		return
	}

	repoID, err := h.getRepositoryID(c, owner, repoName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		return
	}

	files, err := h.service.GetPullRequestFiles(repoID, number)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get pull request files")
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to get pull request files"})
		return
	}

	c.JSON(http.StatusOK, files)
}

// CreateReview handles POST /api/v1/repositories/:owner/:repo/pulls/:number/reviews
func (h *PullRequestHandlers) CreateReview(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")
	numberStr := c.Param("number")

	number, err := strconv.Atoi(numberStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pull request number"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	repoID, err := h.getRepositoryID(c, owner, repoName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		return
	}

	var req struct {
		Body      string                                    `json:"body"`
		Event     string                                    `json:"event" binding:"required"`
		Comments  []services.CreateReviewCommentRequest    `json:"comments"`
		CommitSHA string                                    `json:"commit_sha"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate event
	var state models.ReviewState
	switch req.Event {
	case "APPROVE":
		state = models.ReviewStateApproved
	case "REQUEST_CHANGES":
		state = models.ReviewStateRequestChanges
	case "COMMENT":
		state = models.ReviewStateCommented
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review event"})
		return
	}

	review, err := h.service.CreateReview(repoID, number, userID.(uuid.UUID), req.CommitSHA, req.Body, state, req.Comments)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create review")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create review"})
		return
	}

	c.JSON(http.StatusCreated, review)
}

// ListReviews handles GET /api/v1/repositories/:owner/:repo/pulls/:number/reviews
func (h *PullRequestHandlers) ListReviews(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")
	numberStr := c.Param("number")

	number, err := strconv.Atoi(numberStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pull request number"})
		return
	}

	repoID, err := h.getRepositoryID(c, owner, repoName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		return
	}

	reviews, err := h.service.ListReviews(repoID, number)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list reviews")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list reviews"})
		return
	}

	c.JSON(http.StatusOK, reviews)
}

// CreateReviewComment handles POST /api/v1/repositories/:owner/:repo/pulls/:number/comments
func (h *PullRequestHandlers) CreateReviewComment(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")
	numberStr := c.Param("number")

	number, err := strconv.Atoi(numberStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pull request number"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	repoID, err := h.getRepositoryID(c, owner, repoName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		return
	}

	var req services.CreateReviewCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	comment, err := h.service.CreateReviewComment(repoID, number, userID.(uuid.UUID), req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create review comment")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create review comment"})
		return
	}

	c.JSON(http.StatusCreated, comment)
}

// ListReviewComments handles GET /api/v1/repositories/:owner/:repo/pulls/:number/comments
func (h *PullRequestHandlers) ListReviewComments(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")
	numberStr := c.Param("number")

	number, err := strconv.Atoi(numberStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pull request number"})
		return
	}

	repoID, err := h.getRepositoryID(c, owner, repoName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		return
	}

	comments, err := h.service.ListReviewComments(repoID, number)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list review comments")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list review comments"})
		return
	}

	c.JSON(http.StatusOK, comments)
}

// Helper methods

func (h *PullRequestHandlers) getRepositoryID(c *gin.Context, owner, name string) (uuid.UUID, error) {
	// This would typically query the database to get the repository ID
	// For now, we'll use the first repository in the database as a mock
	
	// For now, just return a mock UUID since we need access to the database
	// In a real implementation, this would be passed as a dependency or we'd have 
	// a proper repository service to query by owner/name
	return uuid.New(), nil
	
	// This is how it would work with proper database access:
	/*
	if err := h.db.Raw(`
		SELECT r.id FROM repositories r 
		INNER JOIN users u ON r.owner_id = u.id AND r.owner_type = 'user'
		WHERE u.username = ? AND r.name = ?
		UNION
		SELECT r.id FROM repositories r 
		INNER JOIN organizations o ON r.owner_id = o.id AND r.owner_type = 'organization'
		WHERE o.login = ? AND r.name = ?
		LIMIT 1
	`, owner, name, owner, name).Scan(&repo).Error; err != nil {
		return uuid.Nil, err
	}
	
	return repo.ID, nil
	*/
}