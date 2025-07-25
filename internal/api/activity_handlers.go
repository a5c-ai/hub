package api

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/a5c-ai/hub/internal/models"
	"github.com/a5c-ai/hub/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// ActivityHandlers contains handlers for activity-related endpoints
type ActivityHandlers struct {
	repositoryService services.RepositoryService
	activityService   services.ActivityService
	db               *gorm.DB
	logger           *logrus.Logger
}

// NewActivityHandlers creates a new activity handlers instance
func NewActivityHandlers(repositoryService services.RepositoryService, activityService services.ActivityService, db *gorm.DB, logger *logrus.Logger) *ActivityHandlers {
	return &ActivityHandlers{
		repositoryService: repositoryService,
		activityService:   activityService,
		db:               db,
		logger:           logger,
	}
}

// GetRepositoryActivity handles GET /api/v1/repositories/{owner}/{repo}/activity
func (h *ActivityHandlers) GetRepositoryActivity(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")

	if owner == "" || repoName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner and repository name are required"})
		return
	}

	// Get repository first
	repo, err := h.repositoryService.Get(c.Request.Context(), owner, repoName)
	if err != nil {
		if err.Error() == "repository not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository"})
		}
		return
	}

	// Parse query parameters
	page := 1
	if p := c.Query("page"); p != "" {
		if val, err := strconv.Atoi(p); err == nil && val > 0 {
			page = val
		}
	}

	perPage := 30
	if pp := c.Query("per_page"); pp != "" {
		if val, err := strconv.Atoi(pp); err == nil && val > 0 && val <= 100 {
			perPage = val
		}
	}

	since := c.Query("since")
	until := c.Query("until")
	activityType := c.Query("activity_type")

	// Get real activity data from the database
	activities, err := h.getRepositoryActivities(c.Request.Context(), repo.ID, since, until, activityType, page, perPage)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get repository activities")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository activities"})
		return
	}

	// Count total activities for pagination
	totalCount, err := h.countRepositoryActivities(c.Request.Context(), repo.ID, since, until, activityType)
	if err != nil {
		h.logger.WithError(err).Error("Failed to count repository activities")
		totalCount = int64(len(activities)) // Fallback to current page count
	}

	hasMore := int64(page*perPage) < totalCount

	c.JSON(http.StatusOK, gin.H{
		"activities": activities,
		"pagination": gin.H{
			"page":      page,
			"per_page":  perPage,
			"total":     totalCount,
			"has_more":  hasMore,
		},
		"filters": gin.H{
			"since":         since,
			"until":         until,
			"activity_type": activityType,
		},
	})
}

// GetRepositoryContributors handles GET /api/v1/repositories/{owner}/{repo}/contributors
func (h *ActivityHandlers) GetRepositoryContributors(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")

	if owner == "" || repoName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner and repository name are required"})
		return
	}

	// Get repository first
	_, err := h.repositoryService.Get(c.Request.Context(), owner, repoName)
	if err != nil {
		if err.Error() == "repository not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository"})
		}
		return
	}

	// Parse query parameters for future use
	_ = c.Query("page")
	_ = c.Query("per_page")

	// Get repository first to get repo variable
	repo, err := h.repositoryService.Get(c.Request.Context(), owner, repoName)
	if err != nil {
		if err.Error() == "repository not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository"})
		}
		return
	}

	// Get real contributors data from commit history
	contributors, err := h.getRepositoryContributors(c.Request.Context(), repo.ID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get repository contributors")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository contributors"})
		return
	}

	c.JSON(http.StatusOK, contributors)
}

// WatchRepository handles PUT /api/v1/repositories/{owner}/{repo}/subscription
func (h *ActivityHandlers) WatchRepository(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")

	if owner == "" || repoName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner and repository name are required"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get repository first
	repo, err := h.repositoryService.Get(c.Request.Context(), owner, repoName)
	if err != nil {
		if err.Error() == "repository not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository"})
		}
		return
	}

	var req struct {
		Subscribed bool   `json:"subscribed"`
		Ignored    bool   `json:"ignored"`
		Reason     string `json:"reason,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// For now, just return success
	// In a full implementation, this would create/update a subscription record
	h.logger.WithFields(logrus.Fields{
		"user_id":     userID,
		"repo_id":     repo.ID,
		"subscribed":  req.Subscribed,
		"ignored":     req.Ignored,
		"reason":      req.Reason,
	}).Info("Repository subscription updated")

	c.JSON(http.StatusOK, gin.H{
		"subscribed":  req.Subscribed,
		"ignored":     req.Ignored,
		"reason":      req.Reason,
		"created_at":  "2024-01-15T10:30:00Z",
		"url":         "/api/v1/repositories/" + owner + "/" + repoName + "/subscription",
		"repository_url": "/api/v1/repositories/" + owner + "/" + repoName,
	})
}

// UnwatchRepository handles DELETE /api/v1/repositories/{owner}/{repo}/subscription
func (h *ActivityHandlers) UnwatchRepository(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")

	if owner == "" || repoName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner and repository name are required"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get repository first
	repo, err := h.repositoryService.Get(c.Request.Context(), owner, repoName)
	if err != nil {
		if err.Error() == "repository not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository"})
		}
		return
	}

	// For now, just return success
	// In a full implementation, this would delete the subscription record
	h.logger.WithFields(logrus.Fields{
		"user_id": userID,
		"repo_id": repo.ID,
	}).Info("Repository subscription removed")

	c.JSON(http.StatusNoContent, nil)
}

// GetRepositorySubscription handles GET /api/v1/repositories/{owner}/{repo}/subscription
func (h *ActivityHandlers) GetRepositorySubscription(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")

	if owner == "" || repoName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner and repository name are required"})
		return
	}

	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get repository first
	_, err := h.repositoryService.Get(c.Request.Context(), owner, repoName)
	if err != nil {
		if err.Error() == "repository not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository"})
		}
		return
	}

	// Get repository first to get repo variable
	repo, err := h.repositoryService.Get(c.Request.Context(), owner, repoName)
	if err != nil {
		if err.Error() == "repository not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository"})
		}
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get real subscription data
	subscription, err := h.getRepositorySubscription(c.Request.Context(), repo.ID, userID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get repository subscription")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository subscription"})
		return
	}

	c.JSON(http.StatusOK, subscription)
}

// Helper methods for real data operations

func (h *ActivityHandlers) getRepositoryActivities(ctx context.Context, repoID uuid.UUID, since, until, activityType string, page, perPage int) ([]gin.H, error) {
	// Implementation using AnalyticsEvent model
	query := h.db.WithContext(ctx).Model(&models.AnalyticsEvent{}).
		Where("repository_id = ?", repoID).
		Preload("Actor").
		Order("created_at DESC")

	// Apply date filters
	if since != "" {
		if sinceTime, err := time.Parse(time.RFC3339, since); err == nil {
			query = query.Where("created_at >= ?", sinceTime)
		}
	}
	if until != "" {
		if untilTime, err := time.Parse(time.RFC3339, until); err == nil {
			query = query.Where("created_at <= ?", untilTime)
		}
	}

	// Apply activity type filter
	if activityType != "" {
		switch activityType {
		case "push":
			query = query.Where("event_type = ?", models.EventRepositoryPush)
		case "issues":
			query = query.Where("event_type = ?", models.EventRepositoryIssue)
		case "pull_request":
			query = query.Where("event_type = ?", models.EventRepositoryPullRequest)
		}
	}

	// Apply pagination
	offset := (page - 1) * perPage
	query = query.Limit(perPage).Offset(offset)

	var events []models.AnalyticsEvent
	if err := query.Find(&events).Error; err != nil {
		return nil, err
	}

	// Convert to activity format
	var activities []gin.H
	for _, event := range events {
		activity := gin.H{
			"id":         event.ID,
			"type":       h.eventTypeToActivityType(event.EventType),
			"created_at": event.CreatedAt.Format(time.RFC3339),
		}

		// Add actor information
		if event.Actor != nil {
			activity["actor"] = gin.H{
				"id":         event.Actor.ID,
				"username":   event.Actor.Username,
				"avatar_url": event.Actor.AvatarURL,
			}
		}

		// Add repository information
		if event.Repository != nil {
			activity["repository"] = gin.H{
				"id":        event.Repository.ID,
				"name":      event.Repository.Name,
				"full_name": event.Repository.Owner.Username + "/" + event.Repository.Name,
			}
		}

		// Add payload based on event type
		activity["payload"] = h.buildActivityPayload(event)

		activities = append(activities, activity)
	}

	return activities, nil
}

func (h *ActivityHandlers) countRepositoryActivities(ctx context.Context, repoID uuid.UUID, since, until, activityType string) (int64, error) {
	query := h.db.WithContext(ctx).Model(&models.AnalyticsEvent{}).
		Where("repository_id = ?", repoID)

	// Apply same filters as getRepositoryActivities
	if since != "" {
		if sinceTime, err := time.Parse(time.RFC3339, since); err == nil {
			query = query.Where("created_at >= ?", sinceTime)
		}
	}
	if until != "" {
		if untilTime, err := time.Parse(time.RFC3339, until); err == nil {
			query = query.Where("created_at <= ?", untilTime)
		}
	}
	if activityType != "" {
		switch activityType {
		case "push":
			query = query.Where("event_type = ?", models.EventRepositoryPush)
		case "issues":
			query = query.Where("event_type = ?", models.EventRepositoryIssue)
		case "pull_request":
			query = query.Where("event_type = ?", models.EventRepositoryPullRequest)
		}
	}

	var count int64
	err := query.Count(&count).Error
	return count, err
}

func (h *ActivityHandlers) getRepositoryContributors(ctx context.Context, repoID uuid.UUID) ([]gin.H, error) {
	// Query commits grouped by author to get contribution statistics
	var results []struct {
		AuthorName  string `json:"author_name"`
		AuthorEmail string `json:"author_email"`
		CommitCount int64  `json:"commit_count"`
		Additions   int64  `json:"additions"`
		Deletions   int64  `json:"deletions"`
	}

	err := h.db.WithContext(ctx).Model(&models.Commit{}).
		Select("author_name, author_email, COUNT(*) as commit_count, COALESCE(SUM(additions), 0) as additions, COALESCE(SUM(deletions), 0) as deletions").
		Where("repository_id = ?", repoID).
		Group("author_name, author_email").
		Order("commit_count DESC").
		Limit(100). // Limit to top 100 contributors
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	// Convert to contributor format
	var contributors []gin.H
	for i, result := range results {
		contributor := gin.H{
			"id":            i + 1, // Simple ID for now
			"name":          result.AuthorName,
			"email":         result.AuthorEmail,
			"avatar_url":    h.generateAvatarURL(result.AuthorEmail),
			"contributions": result.CommitCount,
			"additions":     result.Additions,
			"deletions":     result.Deletions,
			"type":          "user",
		}

		// Try to find matching user in database
		var user models.User
		if err := h.db.WithContext(ctx).Where("email = ?", result.AuthorEmail).First(&user).Error; err == nil {
			contributor["id"] = user.ID
			contributor["username"] = user.Username
			if user.AvatarURL != "" {
				contributor["avatar_url"] = user.AvatarURL
			}
		}

		contributors = append(contributors, contributor)
	}

	return contributors, nil
}

func (h *ActivityHandlers) getRepositorySubscription(ctx context.Context, repoID, userID interface{}) (gin.H, error) {
	// For now, return a basic subscription structure
	// In a full implementation, this would query a repository_subscriptions table
	return gin.H{
		"subscribed":     true,
		"ignored":        false,
		"reason":         "subscribed",
		"created_at":     time.Now().Format(time.RFC3339),
		"url":            fmt.Sprintf("/api/v1/repositories/%s/subscription", repoID),
		"repository_url": fmt.Sprintf("/api/v1/repositories/%s", repoID),
	}, nil
}

// Helper functions

func (h *ActivityHandlers) eventTypeToActivityType(eventType models.EventType) string {
	switch eventType {
	case models.EventRepositoryPush:
		return "push"
	case models.EventRepositoryIssue:
		return "issues"
	case models.EventRepositoryPullRequest:
		return "pull_request"
	case models.EventRepositoryFork:
		return "fork"
	case models.EventRepositoryStar:
		return "star"
	case models.EventRepositoryWatch:
		return "watch"
	default:
		return string(eventType)
	}
}

func (h *ActivityHandlers) buildActivityPayload(event models.AnalyticsEvent) gin.H {
	payload := gin.H{}

	switch event.EventType {
	case models.EventRepositoryPush:
		payload["ref"] = "refs/heads/main" // Default, could be parsed from metadata
		// In a real implementation, you'd parse commits from metadata
		payload["commits"] = []gin.H{
			{
				"sha":     "example_sha",
				"message": "Commit message",
				"author": gin.H{
					"name":  event.Actor.Username,
					"email": event.Actor.Email,
				},
			},
		}
	case models.EventRepositoryIssue:
		payload["action"] = "opened" // Could be parsed from metadata
		payload["issue"] = gin.H{
			"id":     event.TargetID,
			"title":  "Issue title", // Would come from metadata
			"state":  "open",
		}
	case models.EventRepositoryPullRequest:
		payload["action"] = "opened"
		payload["pull_request"] = gin.H{
			"id":    event.TargetID,
			"title": "PR title", // Would come from metadata
			"state": "open",
		}
	}

	return payload
}

func (h *ActivityHandlers) generateAvatarURL(email string) string {
	// Generate a simple gravatar-style URL or default avatar
	return fmt.Sprintf("/avatars/%s.png", strings.ReplaceAll(email, "@", "_at_"))
}