package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/a5c-ai/hub/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// HooksHandlers contains handlers for webhook-related endpoints
type HooksHandlers struct {
	repositoryService      services.RepositoryService
	webhookDeliveryService *services.WebhookDeliveryService
	deployKeyService       *services.DeployKeyService
	logger                 *logrus.Logger
}

// NewHooksHandlers creates a new hooks handlers instance
func NewHooksHandlers(repositoryService services.RepositoryService, webhookDeliveryService *services.WebhookDeliveryService, deployKeyService *services.DeployKeyService, logger *logrus.Logger) *HooksHandlers {
	return &HooksHandlers{
		repositoryService:      repositoryService,
		webhookDeliveryService: webhookDeliveryService,
		deployKeyService:       deployKeyService,
		logger:                 logger,
	}
}

// Webhook represents a repository webhook
type Webhook struct {
	ID           int                    `json:"id"`
	Name         string                 `json:"name"`
	Config       map[string]interface{} `json:"config"`
	Events       []string               `json:"events"`
	Active       bool                   `json:"active"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	PingURL      string                 `json:"ping_url,omitempty"`
	TestURL      string                 `json:"test_url,omitempty"`
	LastResponse *WebhookResponse       `json:"last_response,omitempty"`
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

	// Get webhooks from database
	dbWebhooks, err := h.webhookDeliveryService.ListWebhooks(c.Request.Context(), repo.ID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list webhooks")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list webhooks"})
		return
	}

	// Convert to API format
	webhooks := make([]Webhook, len(dbWebhooks))
	for i, dbWebhook := range dbWebhooks {
		webhooks[i] = Webhook{
			ID:   int(dbWebhook.ID.ID()), // Convert UUID to int for API compatibility
			Name: dbWebhook.Name,
			Config: map[string]interface{}{
				"url":          dbWebhook.URL,
				"content_type": dbWebhook.ContentType,
				"insecure_ssl": func() string {
					if dbWebhook.InsecureSSL {
						return "1"
					} else {
						return "0"
					}
				}(),
			},
			Events:    dbWebhook.GetEventsSlice(),
			Active:    dbWebhook.Active,
			CreatedAt: dbWebhook.CreatedAt,
			UpdatedAt: dbWebhook.UpdatedAt,
			PingURL:   "/api/v1/repositories/" + owner + "/" + repoName + "/hooks/" + dbWebhook.ID.String() + "/pings",
			TestURL:   "/api/v1/repositories/" + owner + "/" + repoName + "/hooks/" + dbWebhook.ID.String() + "/test",
		}
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

	// Extract URL from config
	url, ok := req.Config["url"].(string)
	if !ok || url == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL is required in config"})
		return
	}

	// Extract content type
	contentType := "application/json"
	if ct, ok := req.Config["content_type"].(string); ok {
		if ct == "form" {
			contentType = "application/x-www-form-urlencoded"
		} else {
			contentType = "application/json"
		}
	}

	// Extract insecure SSL setting
	insecureSSL := false
	if ssl, ok := req.Config["insecure_ssl"].(string); ok {
		insecureSSL = ssl == "1"
	}

	// Extract secret
	secret := ""
	if s, ok := req.Config["secret"].(string); ok {
		secret = s
	}

	// Create webhook in database
	dbWebhook, err := h.webhookDeliveryService.CreateWebhook(
		c.Request.Context(),
		repo.ID,
		req.Name,
		url,
		secret,
		req.Events,
		contentType,
		insecureSSL,
		active,
	)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create webhook")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create webhook"})
		return
	}

	// Convert to API format
	webhook := Webhook{
		ID:   int(dbWebhook.ID.ID()), // Convert UUID to int for API compatibility
		Name: dbWebhook.Name,
		Config: map[string]interface{}{
			"url": dbWebhook.URL,
			"content_type": func() string {
				if dbWebhook.ContentType == "application/json" {
					return "json"
				} else {
					return "form"
				}
			}(),
			"insecure_ssl": func() string {
				if dbWebhook.InsecureSSL {
					return "1"
				} else {
					return "0"
				}
			}(),
		},
		Events:    dbWebhook.GetEventsSlice(),
		Active:    dbWebhook.Active,
		CreatedAt: dbWebhook.CreatedAt,
		UpdatedAt: dbWebhook.UpdatedAt,
		PingURL:   "/api/v1/repositories/" + owner + "/" + repoName + "/hooks/" + dbWebhook.ID.String() + "/pings",
		TestURL:   "/api/v1/repositories/" + owner + "/" + repoName + "/hooks/" + dbWebhook.ID.String() + "/test",
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

	// Parse hook ID as UUID
	hookID, err := uuid.Parse(hookIDStr)
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

	// Get webhook from database
	dbWebhook, err := h.webhookDeliveryService.GetWebhook(c.Request.Context(), hookID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Webhook not found"})
		} else {
			h.logger.WithError(err).Error("Failed to get webhook")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get webhook"})
		}
		return
	}

	// Get latest delivery for last response
	deliveries, err := h.webhookDeliveryService.GetDeliveries(c.Request.Context(), hookID, 1, 0)
	var lastResponse *WebhookResponse
	if err == nil && len(deliveries) > 0 {
		latest := deliveries[0]
		lastResponse = &WebhookResponse{
			Code: latest.StatusCode,
			Status: func() string {
				if latest.Success {
					return "success"
				} else {
					return "failed"
				}
			}(),
			Message: latest.ErrorMessage,
		}
		if latest.Success {
			lastResponse.Message = "OK"
		}
	}

	// Convert to API format
	webhook := Webhook{
		ID:   int(dbWebhook.ID.ID()),
		Name: dbWebhook.Name,
		Config: map[string]interface{}{
			"url": dbWebhook.URL,
			"content_type": func() string {
				if dbWebhook.ContentType == "application/json" {
					return "json"
				} else {
					return "form"
				}
			}(),
			"insecure_ssl": func() string {
				if dbWebhook.InsecureSSL {
					return "1"
				} else {
					return "0"
				}
			}(),
		},
		Events:       dbWebhook.GetEventsSlice(),
		Active:       dbWebhook.Active,
		CreatedAt:    dbWebhook.CreatedAt,
		UpdatedAt:    dbWebhook.UpdatedAt,
		PingURL:      "/api/v1/repositories/" + owner + "/" + repoName + "/hooks/" + dbWebhook.ID.String() + "/pings",
		TestURL:      "/api/v1/repositories/" + owner + "/" + repoName + "/hooks/" + dbWebhook.ID.String() + "/test",
		LastResponse: lastResponse,
	}

	c.JSON(http.StatusOK, webhook)
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

	// Parse hook ID as UUID
	hookID, err := uuid.Parse(hookIDStr)
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

	// Prepare updates map
	updates := make(map[string]interface{})

	// Handle config updates
	if req.Config != nil {
		if url, ok := req.Config["url"].(string); ok && url != "" {
			updates["url"] = url
		}
		if contentType, ok := req.Config["content_type"].(string); ok {
			if contentType == "form" {
				updates["content_type"] = "application/x-www-form-urlencoded"
			} else {
				updates["content_type"] = "application/json"
			}
		}
		if insecureSSL, ok := req.Config["insecure_ssl"].(string); ok {
			updates["insecure_ssl"] = insecureSSL == "1"
		}
		if secret, ok := req.Config["secret"].(string); ok {
			updates["secret"] = secret
		}
	}

	// Handle events update
	if req.Events != nil {
		// In a full implementation, this would properly serialize events
		updates["events"] = strings.Join(req.Events, ",")
	}

	// Handle active status update
	if req.Active != nil {
		updates["active"] = *req.Active
	}

	// Update webhook in database
	dbWebhook, err := h.webhookDeliveryService.UpdateWebhook(c.Request.Context(), hookID, updates)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Webhook not found"})
		} else {
			h.logger.WithError(err).Error("Failed to update webhook")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update webhook"})
		}
		return
	}

	// Convert to API format
	webhook := Webhook{
		ID:   int(dbWebhook.ID.ID()),
		Name: dbWebhook.Name,
		Config: map[string]interface{}{
			"url": dbWebhook.URL,
			"content_type": func() string {
				if dbWebhook.ContentType == "application/json" {
					return "json"
				} else {
					return "form"
				}
			}(),
			"insecure_ssl": func() string {
				if dbWebhook.InsecureSSL {
					return "1"
				} else {
					return "0"
				}
			}(),
		},
		Events:    dbWebhook.GetEventsSlice(),
		Active:    dbWebhook.Active,
		CreatedAt: dbWebhook.CreatedAt,
		UpdatedAt: dbWebhook.UpdatedAt,
		PingURL:   "/api/v1/repositories/" + owner + "/" + repoName + "/hooks/" + dbWebhook.ID.String() + "/pings",
		TestURL:   "/api/v1/repositories/" + owner + "/" + repoName + "/hooks/" + dbWebhook.ID.String() + "/test",
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

	// Parse hook ID as UUID
	hookID, err := uuid.Parse(hookIDStr)
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

	// Delete webhook from database
	err = h.webhookDeliveryService.DeleteWebhook(c.Request.Context(), hookID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Webhook not found"})
		} else {
			h.logger.WithError(err).Error("Failed to delete webhook")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete webhook"})
		}
		return
	}

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

	// Parse hook ID as UUID
	hookID, err := uuid.Parse(hookIDStr)
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

	// Send ping webhook
	err = h.webhookDeliveryService.PingWebhook(c.Request.Context(), hookID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Webhook not found"})
		} else {
			h.logger.WithError(err).Error("Failed to ping webhook")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to ping webhook"})
		}
		return
	}

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

	// Get deploy keys from database
	dbDeployKeys, err := h.deployKeyService.ListDeployKeys(c.Request.Context(), repo.ID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list deploy keys")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list deploy keys"})
		return
	}

	// Convert to API format
	deployKeys := make([]DeployKey, len(dbDeployKeys))
	for i, dbKey := range dbDeployKeys {
		deployKeys[i] = DeployKey{
			ID:        int(dbKey.ID.ID()),
			Key:       dbKey.Key,
			URL:       "/api/v1/repositories/" + owner + "/" + repoName + "/keys/" + dbKey.ID.String(),
			Title:     dbKey.Title,
			Verified:  dbKey.Verified,
			CreatedAt: dbKey.CreatedAt,
			ReadOnly:  dbKey.ReadOnly,
		}
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

	// Validate the SSH key
	if err := h.deployKeyService.ValidateSSHKey(req.Key); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid SSH key: " + err.Error()})
		return
	}

	// Create deploy key in database
	dbDeployKey, err := h.deployKeyService.CreateDeployKey(
		c.Request.Context(),
		repo.ID,
		req.Title,
		req.Key,
		readOnly,
	)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, gin.H{"error": "Deploy key already exists"})
		} else {
			h.logger.WithError(err).Error("Failed to create deploy key")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create deploy key"})
		}
		return
	}

	// Convert to API format
	deployKey := DeployKey{
		ID:        int(dbDeployKey.ID.ID()),
		Key:       dbDeployKey.Key,
		URL:       "/api/v1/repositories/" + owner + "/" + repoName + "/keys/" + dbDeployKey.ID.String(),
		Title:     dbDeployKey.Title,
		Verified:  dbDeployKey.Verified,
		CreatedAt: dbDeployKey.CreatedAt,
		ReadOnly:  dbDeployKey.ReadOnly,
	}

	h.logger.WithFields(logrus.Fields{
		"repo_id":       repo.ID,
		"deploy_key_id": deployKey.ID,
		"title":         req.Title,
		"read_only":     readOnly,
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

	// Parse key ID as UUID
	keyID, err := uuid.Parse(keyIDStr)
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

	// Get deploy key from database
	dbDeployKey, err := h.deployKeyService.GetDeployKey(c.Request.Context(), keyID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Deploy key not found"})
		} else {
			h.logger.WithError(err).Error("Failed to get deploy key")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get deploy key"})
		}
		return
	}

	// Convert to API format
	deployKey := DeployKey{
		ID:        int(dbDeployKey.ID.ID()),
		Key:       dbDeployKey.Key,
		URL:       "/api/v1/repositories/" + owner + "/" + repoName + "/keys/" + dbDeployKey.ID.String(),
		Title:     dbDeployKey.Title,
		Verified:  dbDeployKey.Verified,
		CreatedAt: dbDeployKey.CreatedAt,
		ReadOnly:  dbDeployKey.ReadOnly,
	}

	c.JSON(http.StatusOK, deployKey)
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

	// Parse key ID as UUID
	keyID, err := uuid.Parse(keyIDStr)
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

	// Delete deploy key from database
	err = h.deployKeyService.DeleteDeployKey(c.Request.Context(), keyID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Deploy key not found"})
		} else {
			h.logger.WithError(err).Error("Failed to delete deploy key")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete deploy key"})
		}
		return
	}

	h.logger.WithFields(logrus.Fields{
		"repo_id":       repo.ID,
		"deploy_key_id": keyID,
	}).Info("Deleted repository deploy key")

	c.JSON(http.StatusNoContent, nil)
}
