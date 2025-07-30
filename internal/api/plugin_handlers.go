package api

import (
	"net/http"

	"github.com/a5c-ai/hub/internal/services"
	"github.com/gin-gonic/gin"
)

// PluginHandlers handles HTTP requests for plugin operations.
type PluginHandlers struct {
	service services.PluginService
}

// NewPluginHandlers constructs handlers for plugin endpoints.
func NewPluginHandlers(s services.PluginService) *PluginHandlers {
	return &PluginHandlers{service: s}
}

// ListPlugins returns the plugin marketplace.
func (h *PluginHandlers) ListPlugins(c *gin.Context) {
	list, err := h.service.ListMarketplace(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// installRequest represents plugin installation settings.
type installRequest struct {
	Settings map[string]interface{} `json:"settings"`
}

// InstallOrgPlugin installs a plugin for an organization.
func (h *PluginHandlers) InstallOrgPlugin(c *gin.Context) {
	var req installRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	org := c.Param("org")
	name := c.Param("name")
	if err := h.service.InstallOrgPlugin(c.Request.Context(), org, name, req.Settings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// UninstallOrgPlugin removes a plugin from an organization.
func (h *PluginHandlers) UninstallOrgPlugin(c *gin.Context) {
	org := c.Param("org")
	name := c.Param("name")
	if err := h.service.UninstallOrgPlugin(c.Request.Context(), org, name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// InstallRepoPlugin installs a plugin for a repository.
func (h *PluginHandlers) InstallRepoPlugin(c *gin.Context) {
	var req installRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	owner := c.Param("owner")
	repo := c.Param("repo")
	name := c.Param("name")
	if err := h.service.InstallRepoPlugin(c.Request.Context(), owner, repo, name, req.Settings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// UninstallRepoPlugin removes a plugin from a repository.
func (h *PluginHandlers) UninstallRepoPlugin(c *gin.Context) {
	owner := c.Param("owner")
	repo := c.Param("repo")
	name := c.Param("name")
	if err := h.service.UninstallRepoPlugin(c.Request.Context(), owner, repo, name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
