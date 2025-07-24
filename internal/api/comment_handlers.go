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

type CommentHandlers struct {
	commentService services.CommentService
	issueService   services.IssueService
	logger         *logrus.Logger
}

func NewCommentHandlers(
	commentService services.CommentService,
	issueService services.IssueService,
	logger *logrus.Logger,
) *CommentHandlers {
	return &CommentHandlers{
		commentService: commentService,
		issueService:   issueService,
		logger:         logger,
	}
}

// CreateComment godoc
// @Summary Create a comment
// @Description Create a new comment on an issue
// @Tags comments
// @Accept json
// @Produce json
// @Param owner path string true "Repository owner"
// @Param repo path string true "Repository name"
// @Param number path int true "Issue number"
// @Param comment body CreateCommentRequest true "Comment data"
// @Success 201 {object} models.Comment
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /repositories/{owner}/{repo}/issues/{number}/comments [post]
func (h *CommentHandlers) CreateComment(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")
	numberStr := c.Param("number")
	
	number, err := strconv.Atoi(numberStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid issue number"})
		return
	}
	
	// Get issue to validate existence
	issue, err := h.issueService.Get(c.Request.Context(), owner, repoName, number)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"owner":  owner,
			"repo":   repoName,
			"number": number,
		}).Error("Failed to get issue for comment")
		c.JSON(http.StatusNotFound, gin.H{"error": "Issue not found"})
		return
	}
	
	// Check if issue is locked
	if issue.Locked {
		c.JSON(http.StatusForbidden, gin.H{"error": "Issue is locked"})
		return
	}
	
	// Parse request
	var req CreateCommentRequest
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
	serviceReq := services.CreateCommentRequest{
		IssueID: issue.ID,
		UserID:  &userIDUUID,
		Body:    req.Body,
	}
	
	// Create comment
	comment, err := h.commentService.Create(c.Request.Context(), serviceReq)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create comment")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create comment", "details": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, comment)
}

// GetComment godoc
// @Summary Get a comment
// @Description Get a specific comment by ID
// @Tags comments
// @Produce json
// @Param owner path string true "Repository owner"
// @Param repo path string true "Repository name"
// @Param number path int true "Issue number"
// @Param comment_id path string true "Comment ID"
// @Success 200 {object} models.Comment
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /repositories/{owner}/{repo}/issues/{number}/comments/{comment_id} [get]
func (h *CommentHandlers) GetComment(c *gin.Context) {
	commentIDStr := c.Param("comment_id")
	
	commentID, err := uuid.Parse(commentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid comment ID"})
		return
	}
	
	comment, err := h.commentService.Get(c.Request.Context(), commentID)
	if err != nil {
		h.logger.WithError(err).WithField("comment_id", commentID).Error("Failed to get comment")
		if err.Error() == "comment not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get comment"})
		}
		return
	}
	
	c.JSON(http.StatusOK, comment)
}

// UpdateComment godoc
// @Summary Update a comment
// @Description Update an existing comment
// @Tags comments
// @Accept json
// @Produce json
// @Param owner path string true "Repository owner"
// @Param repo path string true "Repository name"
// @Param number path int true "Issue number"
// @Param comment_id path string true "Comment ID"
// @Param comment body UpdateCommentRequest true "Comment update data"
// @Success 200 {object} models.Comment
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /repositories/{owner}/{repo}/issues/{number}/comments/{comment_id} [patch]
func (h *CommentHandlers) UpdateComment(c *gin.Context) {
	commentIDStr := c.Param("comment_id")
	
	commentID, err := uuid.Parse(commentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid comment ID"})
		return
	}
	
	// Get existing comment to check ownership
	comment, err := h.commentService.Get(c.Request.Context(), commentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
		return
	}
	
	// Check if user owns the comment or is admin
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
	
	isAdmin, _ := c.Get("is_admin")
	if comment.UserID == nil || (*comment.UserID != userIDUUID && !isAdmin.(bool)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}
	
	// Parse request
	var req UpdateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}
	
	// Convert to service request
	serviceReq := services.UpdateCommentRequest{
		Body: req.Body,
	}
	
	// Update comment
	updatedComment, err := h.commentService.Update(c.Request.Context(), commentID, serviceReq)
	if err != nil {
		h.logger.WithError(err).Error("Failed to update comment")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update comment", "details": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, updatedComment)
}

// DeleteComment godoc
// @Summary Delete a comment
// @Description Delete a comment (owner or admin only)
// @Tags comments
// @Param owner path string true "Repository owner"
// @Param repo path string true "Repository name"
// @Param number path int true "Issue number"
// @Param comment_id path string true "Comment ID"
// @Success 204
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /repositories/{owner}/{repo}/issues/{number}/comments/{comment_id} [delete]
func (h *CommentHandlers) DeleteComment(c *gin.Context) {
	commentIDStr := c.Param("comment_id")
	
	commentID, err := uuid.Parse(commentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid comment ID"})
		return
	}
	
	// Get existing comment to check ownership
	comment, err := h.commentService.Get(c.Request.Context(), commentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
		return
	}
	
	// Check if user owns the comment or is admin
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
	
	isAdmin, _ := c.Get("is_admin")
	if comment.UserID == nil || (*comment.UserID != userIDUUID && !isAdmin.(bool)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}
	
	// Delete comment
	if err := h.commentService.Delete(c.Request.Context(), commentID); err != nil {
		h.logger.WithError(err).Error("Failed to delete comment")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete comment"})
		return
	}
	
	c.Status(http.StatusNoContent)
}

// ListComments godoc
// @Summary List comments
// @Description Get all comments for an issue
// @Tags comments
// @Produce json
// @Param owner path string true "Repository owner"
// @Param repo path string true "Repository name"
// @Param number path int true "Issue number"
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(30)
// @Success 200 {object} ListCommentsResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /repositories/{owner}/{repo}/issues/{number}/comments [get]
func (h *CommentHandlers) ListComments(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")
	numberStr := c.Param("number")
	
	number, err := strconv.Atoi(numberStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid issue number"})
		return
	}
	
	// Get issue to validate existence and get ID
	issue, err := h.issueService.Get(c.Request.Context(), owner, repoName, number)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Issue not found"})
		return
	}
	
	// Parse query parameters
	filters := h.parseCommentFilters(c)
	
	// Get comments
	comments, total, err := h.commentService.List(c.Request.Context(), issue.ID, filters)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"owner":    owner,
			"repo":     repoName,
			"number":   number,
			"issue_id": issue.ID,
		}).Error("Failed to list comments")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list comments"})
		return
	}
	
	response := ListCommentsResponse{
		Comments: comments,
		Total:    total,
		Page:     filters.Page,
		PerPage:  filters.PerPage,
	}
	
	c.JSON(http.StatusOK, response)
}

// Helper methods

func (h *CommentHandlers) parseCommentFilters(c *gin.Context) services.CommentFilters {
	filters := services.CommentFilters{
		Page:    1,
		PerPage: 30,
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
	
	return filters
}

// Request/Response types

type CreateCommentRequest struct {
	Body string `json:"body" binding:"required"`
}

type UpdateCommentRequest struct {
	Body string `json:"body" binding:"required"`
}

type ListCommentsResponse struct {
	Comments []*models.Comment `json:"comments"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PerPage  int               `json:"per_page"`
}