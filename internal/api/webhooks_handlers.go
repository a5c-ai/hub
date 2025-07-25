package api

import (
	"net/http"

	"github.com/a5c-ai/hub/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// WebhooksHandlers handles webhook events for Actions
type WebhooksHandlers struct {
	actionsEventService *services.ActionsEventService
	logger              *logrus.Logger
}

// NewWebhooksHandlers creates a new webhooks handlers instance
func NewWebhooksHandlers(actionsEventService *services.ActionsEventService, logger *logrus.Logger) *WebhooksHandlers {
	return &WebhooksHandlers{
		actionsEventService: actionsEventService,
		logger:              logger,
	}
}

// HandlePushWebhook handles push webhook events
// POST /api/v1/webhooks/push
func (h *WebhooksHandlers) HandlePushWebhook(c *gin.Context) {
	var event services.GitPushEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		h.logger.WithError(err).Error("Invalid push webhook payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	if err := h.actionsEventService.HandlePushEvent(c.Request.Context(), event); err != nil {
		h.logger.WithError(err).WithField("repository_id", event.RepositoryID).
			Error("Failed to handle push event")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process push event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Push event processed"})
}

// HandlePullRequestWebhook handles pull request webhook events
// POST /api/v1/webhooks/pull_request
func (h *WebhooksHandlers) HandlePullRequestWebhook(c *gin.Context) {
	var event services.PullRequestEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		h.logger.WithError(err).Error("Invalid pull request webhook payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	if err := h.actionsEventService.HandlePullRequestEvent(c.Request.Context(), event); err != nil {
		h.logger.WithError(err).WithField("repository_id", event.RepositoryID).
			Error("Failed to handle pull request event")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process pull request event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pull request event processed"})
}

// HandleIssuesWebhook handles issues webhook events
// POST /api/v1/webhooks/issues
func (h *WebhooksHandlers) HandleIssuesWebhook(c *gin.Context) {
	var event services.IssuesEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		h.logger.WithError(err).Error("Invalid issues webhook payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	if err := h.actionsEventService.HandleIssuesEvent(c.Request.Context(), event); err != nil {
		h.logger.WithError(err).WithField("repository_id", event.RepositoryID).
			Error("Failed to handle issues event")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process issues event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Issues event processed"})
}

// HandleReleaseWebhook handles release webhook events
// POST /api/v1/webhooks/release
func (h *WebhooksHandlers) HandleReleaseWebhook(c *gin.Context) {
	var event services.ReleaseEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		h.logger.WithError(err).Error("Invalid release webhook payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	if err := h.actionsEventService.HandleReleaseEvent(c.Request.Context(), event); err != nil {
		h.logger.WithError(err).WithField("repository_id", event.RepositoryID).
			Error("Failed to handle release event")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process release event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Release event processed"})
}

// HandleScheduledWebhook handles scheduled workflow triggers
// POST /api/v1/webhooks/scheduled
func (h *WebhooksHandlers) HandleScheduledWebhook(c *gin.Context) {
	var req struct {
		RepositoryID uuid.UUID `json:"repository_id" binding:"required"`
		WorkflowID   uuid.UUID `json:"workflow_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid scheduled webhook payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	if err := h.actionsEventService.HandleScheduledEvent(c.Request.Context(), req.RepositoryID, req.WorkflowID); err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"repository_id": req.RepositoryID,
			"workflow_id":   req.WorkflowID,
		}).Error("Failed to handle scheduled event")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process scheduled event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Scheduled event processed"})
}

// HandleGenericWebhook handles generic webhook events (for custom integrations)
// POST /api/v1/webhooks/generic
func (h *WebhooksHandlers) HandleGenericWebhook(c *gin.Context) {
	// Extract event type from header or query parameter
	eventType := c.GetHeader("X-Hub-Event")
	if eventType == "" {
		eventType = c.Query("event")
	}

	if eventType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing event type"})
		return
	}

	// Route to appropriate handler based on event type
	switch eventType {
	case "push":
		h.HandlePushWebhook(c)
	case "pull_request":
		h.HandlePullRequestWebhook(c)
	case "issues":
		h.HandleIssuesWebhook(c)
	case "release":
		h.HandleReleaseWebhook(c)
	default:
		h.logger.WithField("event_type", eventType).Warn("Unknown webhook event type")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unknown event type"})
	}
}