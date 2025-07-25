package api

import (
	"net/http"
	"strconv"

	"github.com/a5c-ai/hub/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ActivityHandlers contains handlers for activity-related endpoints
type ActivityHandlers struct {
	repositoryService services.RepositoryService
	activityService   services.ActivityService
	logger           *logrus.Logger
}

// NewActivityHandlers creates a new activity handlers instance
func NewActivityHandlers(repositoryService services.RepositoryService, activityService services.ActivityService, logger *logrus.Logger) *ActivityHandlers {
	return &ActivityHandlers{
		repositoryService: repositoryService,
		activityService:   activityService,
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

	// For now, return mock activity data
	// In a full implementation, this would query actual activity from the database
	activities := []gin.H{
		{
			"id":          1,
			"type":        "push",
			"actor": gin.H{
				"id":         1,
				"username":   "john.doe",
				"avatar_url": "/avatars/john.doe.png",
			},
			"repository": gin.H{
				"id":        repo.ID,
				"name":      repo.Name,
				"full_name": owner + "/" + repoName,
			},
			"payload": gin.H{
				"ref":     "refs/heads/main",
				"commits": []gin.H{
					{
						"sha":     "abc123",
						"message": "Update README.md",
						"author": gin.H{
							"name":  "John Doe",
							"email": "john.doe@example.com",
						},
					},
				},
			},
			"created_at": "2024-01-15T10:30:00Z",
		},
		{
			"id":   2,
			"type": "issues",
			"actor": gin.H{
				"id":         2,
				"username":   "jane.smith",
				"avatar_url": "/avatars/jane.smith.png",
			},
			"repository": gin.H{
				"id":        repo.ID,
				"name":      repo.Name,
				"full_name": owner + "/" + repoName,
			},
			"payload": gin.H{
				"action": "opened",
				"issue": gin.H{
					"id":     1,
					"number": 1,
					"title":  "Bug in login functionality",
					"state":  "open",
				},
			},
			"created_at": "2024-01-14T15:45:00Z",
		},
	}

	// Apply filters (mock implementation)
	filteredActivities := activities
	if activityType != "" {
		var filtered []gin.H
		for _, activity := range activities {
			if activity["type"] == activityType {
				filtered = append(filtered, activity)
			}
		}
		filteredActivities = filtered
	}

	c.JSON(http.StatusOK, gin.H{
		"activities": filteredActivities,
		"pagination": gin.H{
			"page":      page,
			"per_page":  perPage,
			"total":     len(filteredActivities),
			"has_more":  false,
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

	// For now, return mock contributors data
	// In a full implementation, this would analyze Git history to find actual contributors
	contributors := []gin.H{
		{
			"id":           1,
			"username":     "john.doe",
			"name":         "John Doe",
			"email":        "john.doe@example.com",
			"avatar_url":   "/avatars/john.doe.png",
			"contributions": 45,
			"type":         "user",
		},
		{
			"id":           2,
			"username":     "jane.smith",
			"name":         "Jane Smith",
			"email":        "jane.smith@example.com",
			"avatar_url":   "/avatars/jane.smith.png",
			"contributions": 23,
			"type":         "user",
		},
		{
			"id":           3,
			"username":     "bob.wilson",
			"name":         "Bob Wilson",
			"email":        "bob.wilson@example.com",
			"avatar_url":   "/avatars/bob.wilson.png",
			"contributions": 12,
			"type":         "user",
		},
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

	// For now, return mock subscription data
	// In a full implementation, this would query the actual subscription
	c.JSON(http.StatusOK, gin.H{
		"subscribed":  true,
		"ignored":     false,
		"reason":      "subscribed",
		"created_at":  "2024-01-10T08:00:00Z",
		"url":         "/api/v1/repositories/" + owner + "/" + repoName + "/subscription",
		"repository_url": "/api/v1/repositories/" + owner + "/" + repoName,
	})
}