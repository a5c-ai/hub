package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/a5c-ai/hub/internal/models"
	"github.com/a5c-ai/hub/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type IssueHandlers struct {
	issueService      services.IssueService
	commentService    services.CommentService
	labelService      services.LabelService
	milestoneService  services.MilestoneService
	repositoryService services.RepositoryService
	logger            *logrus.Logger
}

func NewIssueHandlers(
	issueService services.IssueService,
	commentService services.CommentService,
	labelService services.LabelService,
	milestoneService services.MilestoneService,
	repositoryService services.RepositoryService,
	logger *logrus.Logger,
) *IssueHandlers {
	return &IssueHandlers{
		issueService:      issueService,
		commentService:    commentService,
		labelService:      labelService,
		milestoneService:  milestoneService,
		repositoryService: repositoryService,
		logger:            logger,
	}
}

// CreateIssue godoc
// @Summary Create a new issue
// @Description Create a new issue in a repository
// @Tags issues
// @Accept json
// @Produce json
// @Param owner path string true "Repository owner"
// @Param repo path string true "Repository name"
// @Param issue body CreateIssueAPIRequest true "Issue data"
// @Success 201 {object} models.Issue
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /repositories/{owner}/{repo}/issues [post]
func (h *IssueHandlers) CreateIssue(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")
	
	// Get repository to validate access and existence
	repo, err := h.repositoryService.Get(c.Request.Context(), owner, repoName)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"owner": owner,
			"repo":  repoName,
		}).Error("Failed to get repository")
		c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		return
	}
	
	// Parse request
	var req CreateIssueAPIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}
	
	// Get authenticated user
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	
	userIDUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}
	
	// Convert to service request
	serviceReq := services.CreateIssueRequest{
		RepositoryID: repo.ID,
		Title:        req.Title,
		Body:         req.Body,
		UserID:       &userIDUUID,
		AssigneeID:   req.AssigneeID,
		MilestoneID:  req.MilestoneID,
		LabelIDs:     req.LabelIDs,
	}
	
	// Create issue
	issue, err := h.issueService.Create(c.Request.Context(), serviceReq)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create issue")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create issue", "details": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, issue)
}

// GetIssue godoc
// @Summary Get an issue
// @Description Get a specific issue by number
// @Tags issues
// @Produce json
// @Param owner path string true "Repository owner"
// @Param repo path string true "Repository name"
// @Param number path int true "Issue number"
// @Success 200 {object} models.Issue
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /repositories/{owner}/{repo}/issues/{number} [get]
func (h *IssueHandlers) GetIssue(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")
	numberStr := c.Param("number")
	
	number, err := strconv.Atoi(numberStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid issue number"})
		return
	}
	
	issue, err := h.issueService.Get(c.Request.Context(), owner, repoName, number)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"owner":  owner,
			"repo":   repoName,
			"number": number,
		}).Error("Failed to get issue")
		if err.Error() == "issue not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Issue not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get issue"})
		}
		return
	}
	
	c.JSON(http.StatusOK, issue)
}

// UpdateIssue godoc
// @Summary Update an issue
// @Description Update an existing issue
// @Tags issues
// @Accept json
// @Produce json
// @Param owner path string true "Repository owner"
// @Param repo path string true "Repository name"
// @Param number path int true "Issue number"
// @Param issue body UpdateIssueAPIRequest true "Issue update data"
// @Success 200 {object} models.Issue
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /repositories/{owner}/{repo}/issues/{number} [patch]
func (h *IssueHandlers) UpdateIssue(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")
	numberStr := c.Param("number")
	
	number, err := strconv.Atoi(numberStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid issue number"})
		return
	}
	
	// Get existing issue
	issue, err := h.issueService.Get(c.Request.Context(), owner, repoName, number)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get issue for update")
		c.JSON(http.StatusNotFound, gin.H{"error": "Issue not found"})
		return
	}
	
	// Parse request
	var req UpdateIssueAPIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}
	
	// Convert to service request
	serviceReq := services.UpdateIssueRequest{
		Title:        req.Title,
		Body:         req.Body,
		State:        req.State,
		StateReason:  req.StateReason,
		AssigneeID:   req.AssigneeID,
		MilestoneID:  req.MilestoneID,
		LabelIDs:     req.LabelIDs,
	}
	
	// Update issue
	updatedIssue, err := h.issueService.Update(c.Request.Context(), issue.ID, serviceReq)
	if err != nil {
		h.logger.WithError(err).Error("Failed to update issue")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update issue", "details": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, updatedIssue)
}

// DeleteIssue godoc
// @Summary Delete an issue
// @Description Delete an issue (admin only)
// @Tags issues
// @Param owner path string true "Repository owner"
// @Param repo path string true "Repository name"
// @Param number path int true "Issue number"
// @Success 204
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /repositories/{owner}/{repo}/issues/{number} [delete]
func (h *IssueHandlers) DeleteIssue(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")
	numberStr := c.Param("number")
	
	number, err := strconv.Atoi(numberStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid issue number"})
		return
	}
	
	// Get existing issue
	issue, err := h.issueService.Get(c.Request.Context(), owner, repoName, number)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get issue for deletion")
		c.JSON(http.StatusNotFound, gin.H{"error": "Issue not found"})
		return
	}
	
	// Delete issue
	if err := h.issueService.Delete(c.Request.Context(), issue.ID); err != nil {
		h.logger.WithError(err).Error("Failed to delete issue")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete issue"})
		return
	}
	
	c.Status(http.StatusNoContent)
}

// ListIssues godoc
// @Summary List issues
// @Description Get a list of issues for a repository
// @Tags issues
// @Produce json
// @Param owner path string true "Repository owner"
// @Param repo path string true "Repository name"
// @Param state query string false "Issue state" Enums(open, closed)
// @Param assignee query string false "Filter by assignee username"
// @Param creator query string false "Filter by creator username"
// @Param milestone query string false "Filter by milestone number"
// @Param labels query string false "Filter by label names (comma-separated)"
// @Param sort query string false "Sort by" Enums(created, updated, comments) default(created)
// @Param direction query string false "Sort direction" Enums(asc, desc) default(desc)
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(30)
// @Success 200 {object} ListIssuesResponse
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /repositories/{owner}/{repo}/issues [get]
func (h *IssueHandlers) ListIssues(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")
	
	// Parse query parameters
	filters := h.parseIssueFilters(c)
	
	// Get issues
	issues, total, err := h.issueService.List(c.Request.Context(), owner, repoName, filters)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"owner": owner,
			"repo":  repoName,
		}).Error("Failed to list issues")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list issues"})
		return
	}
	
	response := ListIssuesResponse{
		Issues: issues,
		Total:  total,
		Page:   filters.Page,
		PerPage: filters.PerPage,
	}
	
	c.JSON(http.StatusOK, response)
}

// SearchIssues godoc
// @Summary Search issues
// @Description Search issues in a repository
// @Tags issues
// @Produce json
// @Param owner path string true "Repository owner"
// @Param repo path string true "Repository name"
// @Param q query string true "Search query"
// @Param state query string false "Issue state" Enums(open, closed)
// @Param sort query string false "Sort by" Enums(created, updated, comments) default(created)
// @Param direction query string false "Sort direction" Enums(asc, desc) default(desc)
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(30)
// @Success 200 {object} ListIssuesResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /repositories/{owner}/{repo}/issues/search [get]
func (h *IssueHandlers) SearchIssues(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")
	query := c.Query("q")
	
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query is required"})
		return
	}
	
	// Parse query parameters
	filters := h.parseIssueFilters(c)
	
	// Search issues
	issues, total, err := h.issueService.Search(c.Request.Context(), owner, repoName, query, filters)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"owner": owner,
			"repo":  repoName,
			"query": query,
		}).Error("Failed to search issues")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search issues"})
		return
	}
	
	response := ListIssuesResponse{
		Issues: issues,
		Total:  total,
		Page:   filters.Page,
		PerPage: filters.PerPage,
	}
	
	c.JSON(http.StatusOK, response)
}

// Issue operations

// CloseIssue godoc
// @Summary Close an issue
// @Description Close an issue with optional reason
// @Tags issues
// @Accept json
// @Produce json
// @Param owner path string true "Repository owner"
// @Param repo path string true "Repository name"
// @Param number path int true "Issue number"
// @Param body body CloseIssueRequest true "Close reason"
// @Success 200 {object} models.Issue
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /repositories/{owner}/{repo}/issues/{number}/close [post]
func (h *IssueHandlers) CloseIssue(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")
	numberStr := c.Param("number")
	
	number, err := strconv.Atoi(numberStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid issue number"})
		return
	}
	
	// Get existing issue
	issue, err := h.issueService.Get(c.Request.Context(), owner, repoName, number)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Issue not found"})
		return
	}
	
	var req CloseIssueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}
	
	// Close issue
	updatedIssue, err := h.issueService.Close(c.Request.Context(), issue.ID, req.Reason)
	if err != nil {
		h.logger.WithError(err).Error("Failed to close issue")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to close issue"})
		return
	}
	
	c.JSON(http.StatusOK, updatedIssue)
}

// ReopenIssue godoc
// @Summary Reopen an issue
// @Description Reopen a closed issue
// @Tags issues
// @Produce json
// @Param owner path string true "Repository owner"
// @Param repo path string true "Repository name"
// @Param number path int true "Issue number"
// @Success 200 {object} models.Issue
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /repositories/{owner}/{repo}/issues/{number}/reopen [post]
func (h *IssueHandlers) ReopenIssue(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")
	numberStr := c.Param("number")
	
	number, err := strconv.Atoi(numberStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid issue number"})
		return
	}
	
	// Get existing issue
	issue, err := h.issueService.Get(c.Request.Context(), owner, repoName, number)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Issue not found"})
		return
	}
	
	// Reopen issue
	updatedIssue, err := h.issueService.Reopen(c.Request.Context(), issue.ID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to reopen issue")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reopen issue"})
		return
	}
	
	c.JSON(http.StatusOK, updatedIssue)
}

// LockIssue godoc
// @Summary Lock an issue
// @Description Lock an issue to prevent further comments
// @Tags issues
// @Accept json
// @Produce json
// @Param owner path string true "Repository owner"
// @Param repo path string true "Repository name"
// @Param number path int true "Issue number"
// @Param body body LockIssueRequest true "Lock reason"
// @Success 200 {object} models.Issue
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /repositories/{owner}/{repo}/issues/{number}/lock [post]
func (h *IssueHandlers) LockIssue(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")
	numberStr := c.Param("number")
	
	number, err := strconv.Atoi(numberStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid issue number"})
		return
	}
	
	// Get existing issue
	issue, err := h.issueService.Get(c.Request.Context(), owner, repoName, number)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Issue not found"})
		return
	}
	
	var req LockIssueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}
	
	// Lock issue
	updatedIssue, err := h.issueService.Lock(c.Request.Context(), issue.ID, req.Reason)
	if err != nil {
		h.logger.WithError(err).Error("Failed to lock issue")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to lock issue"})
		return
	}
	
	c.JSON(http.StatusOK, updatedIssue)
}

// UnlockIssue godoc
// @Summary Unlock an issue
// @Description Unlock an issue to allow comments
// @Tags issues
// @Produce json
// @Param owner path string true "Repository owner"
// @Param repo path string true "Repository name"
// @Param number path int true "Issue number"
// @Success 200 {object} models.Issue
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /repositories/{owner}/{repo}/issues/{number}/unlock [post]
func (h *IssueHandlers) UnlockIssue(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")
	numberStr := c.Param("number")
	
	number, err := strconv.Atoi(numberStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid issue number"})
		return
	}
	
	// Get existing issue
	issue, err := h.issueService.Get(c.Request.Context(), owner, repoName, number)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Issue not found"})
		return
	}
	
	// Unlock issue
	updatedIssue, err := h.issueService.Unlock(c.Request.Context(), issue.ID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to unlock issue")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unlock issue"})
		return
	}
	
	c.JSON(http.StatusOK, updatedIssue)
}

// Helper methods

func (h *IssueHandlers) parseIssueFilters(c *gin.Context) services.IssueFilters {
	filters := services.IssueFilters{
		Page:    1,
		PerPage: 30,
	}
	
	// State filter
	if state := c.Query("state"); state != "" {
		if state == "open" || state == "closed" {
			issueState := models.IssueState(state)
			filters.State = &issueState
		}
	}
	
	// Sort and direction
	if sort := c.Query("sort"); sort != "" {
		filters.Sort = sort
	}
	if direction := c.Query("direction"); direction != "" {
		filters.Direction = direction
	}
	
	// Pagination
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filters.Page = page
		}
	}
	if perPageStr := c.Query("per_page"); perPageStr != "" {
		if perPage, err := strconv.Atoi(perPageStr); err == nil && perPage > 0 && perPage <= 100 {
			filters.PerPage = perPage
		}
	}
	
	// Since filter
	if sinceStr := c.Query("since"); sinceStr != "" {
		if since, err := time.Parse(time.RFC3339, sinceStr); err == nil {
			filters.Since = &since
		}
	}
	
	return filters
}

// Request/Response types

type CreateIssueAPIRequest struct {
	Title       string       `json:"title" binding:"required,max=255"`
	Body        string       `json:"body"`
	AssigneeID  *uuid.UUID   `json:"assignee_id"`
	MilestoneID *uuid.UUID   `json:"milestone_id"`
	LabelIDs    []uuid.UUID  `json:"label_ids"`
}

type UpdateIssueAPIRequest struct {
	Title        *string             `json:"title,omitempty" binding:"omitempty,max=255"`
	Body         *string             `json:"body,omitempty"`
	State        *models.IssueState  `json:"state,omitempty"`
	StateReason  *string             `json:"state_reason,omitempty"`
	AssigneeID   *uuid.UUID          `json:"assignee_id,omitempty"`
	MilestoneID  *uuid.UUID          `json:"milestone_id,omitempty"`
	LabelIDs     []uuid.UUID         `json:"label_ids,omitempty"`
}

type CloseIssueRequest struct {
	Reason string `json:"reason"`
}

type LockIssueRequest struct {
	Reason string `json:"reason"`
}

type ListIssuesResponse struct {
	Issues  []*models.Issue `json:"issues"`
	Total   int64          `json:"total"`
	Page    int            `json:"page"`
	PerPage int            `json:"per_page"`
}