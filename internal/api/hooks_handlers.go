package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/a5c-ai/hub/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// HooksHandlers contains handlers for webhook-related endpoints
type HooksHandlers struct {
	repositoryService services.RepositoryService
	logger           *logrus.Logger
}

// NewHooksHandlers creates a new hooks handlers instance
func NewHooksHandlers(repositoryService services.RepositoryService, logger *logrus.Logger) *HooksHandlers {
	return &HooksHandlers{
		repositoryService: repositoryService,
		logger:           logger,
	}
}

// Webhook represents a repository webhook
type Webhook struct {
	ID          int                    `json:"id"`
	Name        string                 `json:"name"`
	Config      map[string]interface{} `json:"config"`
	Events      []string               `json:"events"`
	Active      bool                   `json:"active"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	PingURL     string                 `json:"ping_url,omitempty"`
	TestURL     string                 `json:"test_url,omitempty"`
	LastResponse *WebhookResponse      `json:"last_response,omitempty"`
}

// WebhookResponse represents a webhook delivery response
type WebhookResponse struct {
	Code    int    `json:"code"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// DeployKey represents a repository deploy key
type DeployKey struct {
	ID        int       `json:"id"`
	Key       string    `json:"key"`
	URL       string    `json:"url"`
	Title     string    `json:"title"`
	Verified  bool      `json:"verified"`
	CreatedAt time.Time `json:"created_at"`
	ReadOnly  bool      `json:"read_only"`
}

// ListWebhooks handles GET /api/v1/repositories/{owner}/{repo}/hooks
func (h *HooksHandlers) ListWebhooks(c *gin.Context) {
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

	// For now, return mock webhook data
	// In a full implementation, this would query webhooks from the database
	webhooks := []Webhook{
		{
			ID:   1,
			Name: "web",
			Config: map[string]interface{}{
				"url":          "https://example.com/webhook",
				"content_type": "json",
				"insecure_ssl": "0",
			},
			Events:    []string{"push", "pull_request"},
			Active:    true,
			CreatedAt: time.Now().Add(-24 * time.Hour),
			UpdatedAt: time.Now().Add(-1 * time.Hour),
			PingURL:   "/api/v1/repositories/" + owner + "/" + repoName + "/hooks/1/pings",
			TestURL:   "/api/v1/repositories/" + owner + "/" + repoName + "/hooks/1/test",
		},
	}

	h.logger.WithFields(logrus.Fields{
		"repo_id": repo.ID,
		"count":   len(webhooks),
	}).Info("Listed repository webhooks")

	c.JSON(http.StatusOK, webhooks)
}

// CreateWebhook handles POST /api/v1/repositories/{owner}/{repo}/hooks
func (h *HooksHandlers) CreateWebhook(c *gin.Context) {
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

	var req struct {
		Name   string                 `json:"name"`
		Config map[string]interface{} `json:"config"`
		Events []string               `json:"events"`
		Active *bool                  `json:"active,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Validate required fields
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
		return
	}

	if req.Config == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Config is required"})
		return
	}

	// Set default values
	active := true
	if req.Active != nil {
		active = *req.Active
	}

	if len(req.Events) == 0 {
		req.Events = []string{"push"}
	}

	// Create webhook (mock implementation)
	webhook := Webhook{
		ID:        2, // Mock ID
		Name:      req.Name,
		Config:    req.Config,
		Events:    req.Events,
		Active:    active,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		PingURL:   "/api/v1/repositories/" + owner + "/" + repoName + "/hooks/2/pings",
		TestURL:   "/api/v1/repositories/" + owner + "/" + repoName + "/hooks/2/test",
	}

	h.logger.WithFields(logrus.Fields{
		"repo_id":     repo.ID,
		"webhook_id":  webhook.ID,
		"webhook_url": req.Config["url"],
		"events":      req.Events,
	}).Info("Created repository webhook")

	c.JSON(http.StatusCreated, webhook)
}

// GetWebhook handles GET /api/v1/repositories/{owner}/{repo}/hooks/{hook_id}
func (h *HooksHandlers) GetWebhook(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")
	hookIDStr := c.Param("hook_id")

	if owner == "" || repoName == "" || hookIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner, repository name, and hook ID are required"})
		return
	}

	hookID, err := strconv.Atoi(hookIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid hook ID"})
		return
	}

	// Get repository first
	_, err = h.repositoryService.Get(c.Request.Context(), owner, repoName)
	if err != nil {
		if err.Error() == "repository not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository"})
		}
		return
	}

	// For now, return mock webhook data
	// In a full implementation, this would query the specific webhook from the database
	if hookID == 1 {
		webhook := Webhook{
			ID:   1,
			Name: "web",
			Config: map[string]interface{}{
				"url":          "https://example.com/webhook",
				"content_type": "json",
				"insecure_ssl": "0",
			},
			Events:    []string{"push", "pull_request"},
			Active:    true,
			CreatedAt: time.Now().Add(-24 * time.Hour),
			UpdatedAt: time.Now().Add(-1 * time.Hour),
			PingURL:   "/api/v1/repositories/" + owner + "/" + repoName + "/hooks/1/pings",
			TestURL:   "/api/v1/repositories/" + owner + "/" + repoName + "/hooks/1/test",
			LastResponse: &WebhookResponse{
				Code:    200,
				Status:  "success",
				Message: "OK",
			},
		}

		c.JSON(http.StatusOK, webhook)
		return
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Webhook not found"})
}

// UpdateWebhook handles PATCH /api/v1/repositories/{owner}/{repo}/hooks/{hook_id}
func (h *HooksHandlers) UpdateWebhook(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")
	hookIDStr := c.Param("hook_id")

	if owner == "" || repoName == "" || hookIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner, repository name, and hook ID are required"})
		return
	}

	hookID, err := strconv.Atoi(hookIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid hook ID"})
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
		Config map[string]interface{} `json:"config,omitempty"`
		Events []string               `json:"events,omitempty"`
		Active *bool                  `json:"active,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// For now, return updated mock webhook data
	// In a full implementation, this would update the webhook in the database
	webhook := Webhook{
		ID:   hookID,
		Name: "web",
		Config: map[string]interface{}{
			"url":          "https://example.com/webhook",
			"content_type": "json",
			"insecure_ssl": "0",
		},
		Events:    []string{"push", "pull_request"},
		Active:    true,
		CreatedAt: time.Now().Add(-24 * time.Hour),
		UpdatedAt: time.Now(),
		PingURL:   "/api/v1/repositories/" + owner + "/" + repoName + "/hooks/" + hookIDStr + "/pings",
		TestURL:   "/api/v1/repositories/" + owner + "/" + repoName + "/hooks/" + hookIDStr + "/test",
	}

	// Apply updates
	if req.Config != nil {
		for k, v := range req.Config {
			webhook.Config[k] = v
		}
	}
	if req.Events != nil {
		webhook.Events = req.Events
	}
	if req.Active != nil {
		webhook.Active = *req.Active
	}

	h.logger.WithFields(logrus.Fields{
		"repo_id":    repo.ID,
		"webhook_id": hookID,
	}).Info("Updated repository webhook")

	c.JSON(http.StatusOK, webhook)
}

// DeleteWebhook handles DELETE /api/v1/repositories/{owner}/{repo}/hooks/{hook_id}
func (h *HooksHandlers) DeleteWebhook(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")
	hookIDStr := c.Param("hook_id")

	if owner == "" || repoName == "" || hookIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner, repository name, and hook ID are required"})
		return
	}

	hookID, err := strconv.Atoi(hookIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid hook ID"})
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

	// For now, just log the deletion
	// In a full implementation, this would delete the webhook from the database
	h.logger.WithFields(logrus.Fields{
		"repo_id":    repo.ID,
		"webhook_id": hookID,
	}).Info("Deleted repository webhook")

	c.JSON(http.StatusNoContent, nil)
}

// PingWebhook handles POST /api/v1/repositories/{owner}/{repo}/hooks/{hook_id}/pings
func (h *HooksHandlers) PingWebhook(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")
	hookIDStr := c.Param("hook_id")

	if owner == "" || repoName == "" || hookIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner, repository name, and hook ID are required"})
		return
	}

	hookID, err := strconv.Atoi(hookIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid hook ID"})
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
	// In a full implementation, this would send a ping to the webhook URL
	h.logger.WithFields(logrus.Fields{
		"repo_id":    repo.ID,
		"webhook_id": hookID,
	}).Info("Pinged repository webhook")

	c.JSON(http.StatusOK, gin.H{
		"message": "Webhook ping sent successfully",
	})
}

// ListDeployKeys handles GET /api/v1/repositories/{owner}/{repo}/keys
func (h *HooksHandlers) ListDeployKeys(c *gin.Context) {
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

	// For now, return mock deploy keys data
	// In a full implementation, this would query deploy keys from the database
	deployKeys := []DeployKey{
		{
			ID:        1,
			Key:       "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC7Ht...",
			URL:       "/api/v1/repositories/" + owner + "/" + repoName + "/keys/1",
			Title:     "Deploy Key for CI/CD",
			Verified:  true,
			CreatedAt: time.Now().Add(-7 * 24 * time.Hour),
			ReadOnly:  true,
		},
	}

	h.logger.WithFields(logrus.Fields{
		"repo_id": repo.ID,
		"count":   len(deployKeys),
	}).Info("Listed repository deploy keys")

	c.JSON(http.StatusOK, deployKeys)
}

// CreateDeployKey handles POST /api/v1/repositories/{owner}/{repo}/keys
func (h *HooksHandlers) CreateDeployKey(c *gin.Context) {
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

	var req struct {
		Title    string `json:"title"`
		Key      string `json:"key"`
		ReadOnly *bool  `json:"read_only,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Validate required fields
	if req.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title is required"})
		return
	}

	if req.Key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Key is required"})
		return
	}

	// Set default values
	readOnly := true
	if req.ReadOnly != nil {
		readOnly = *req.ReadOnly
	}

	// Create deploy key (mock implementation)
	deployKey := DeployKey{
		ID:        2, // Mock ID
		Key:       req.Key,
		URL:       "/api/v1/repositories/" + owner + "/" + repoName + "/keys/2",
		Title:     req.Title,
		Verified:  true, // Mock as verified
		CreatedAt: time.Now(),
		ReadOnly:  readOnly,
	}

	h.logger.WithFields(logrus.Fields{
		"repo_id":      repo.ID,
		"deploy_key_id": deployKey.ID,
		"title":        req.Title,
		"read_only":    readOnly,
	}).Info("Created repository deploy key")

	c.JSON(http.StatusCreated, deployKey)
}

// GetDeployKey handles GET /api/v1/repositories/{owner}/{repo}/keys/{key_id}
func (h *HooksHandlers) GetDeployKey(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")
	keyIDStr := c.Param("key_id")

	if owner == "" || repoName == "" || keyIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner, repository name, and key ID are required"})
		return
	}

	keyID, err := strconv.Atoi(keyIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid key ID"})
		return
	}

	// Get repository first
	_, err = h.repositoryService.Get(c.Request.Context(), owner, repoName)
	if err != nil {
		if err.Error() == "repository not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository"})
		}
		return
	}

	// For now, return mock deploy key data
	// In a full implementation, this would query the specific deploy key from the database
	if keyID == 1 {
		deployKey := DeployKey{
			ID:        1,
			Key:       "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC7Ht...",
			URL:       "/api/v1/repositories/" + owner + "/" + repoName + "/keys/1",
			Title:     "Deploy Key for CI/CD",
			Verified:  true,
			CreatedAt: time.Now().Add(-7 * 24 * time.Hour),
			ReadOnly:  true,
		}

		c.JSON(http.StatusOK, deployKey)
		return
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Deploy key not found"})
}

// DeleteDeployKey handles DELETE /api/v1/repositories/{owner}/{repo}/keys/{key_id}
func (h *HooksHandlers) DeleteDeployKey(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")
	keyIDStr := c.Param("key_id")

	if owner == "" || repoName == "" || keyIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner, repository name, and key ID are required"})
		return
	}

	keyID, err := strconv.Atoi(keyIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid key ID"})
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

	// For now, just log the deletion
	// In a full implementation, this would delete the deploy key from the database
	h.logger.WithFields(logrus.Fields{
		"repo_id":       repo.ID,
		"deploy_key_id": keyID,
	}).Info("Deleted repository deploy key")

	c.JSON(http.StatusNoContent, nil)
}