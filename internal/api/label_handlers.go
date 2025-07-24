package api

import (
	"net/http"
	"strconv"

	"github.com/a5c-ai/hub/internal/models"
	"github.com/a5c-ai/hub/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type LabelHandlers struct {
	labelService      services.LabelService
	repositoryService services.RepositoryService
	logger            *logrus.Logger
}

func NewLabelHandlers(
	labelService services.LabelService,
	repositoryService services.RepositoryService,
	logger *logrus.Logger,
) *LabelHandlers {
	return &LabelHandlers{
		labelService:      labelService,
		repositoryService: repositoryService,
		logger:            logger,
	}
}

// CreateLabel godoc
// @Summary Create a label
// @Description Create a new label for a repository
// @Tags labels
// @Accept json
// @Produce json
// @Param owner path string true "Repository owner"
// @Param repo path string true "Repository name"
// @Param label body CreateLabelRequest true "Label data"
// @Success 201 {object} models.Label
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /repositories/{owner}/{repo}/labels [post]
func (h *LabelHandlers) CreateLabel(c *gin.Context) {
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
	var req CreateLabelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}
	
	// Convert to service request
	serviceReq := services.CreateLabelRequest{
		RepositoryID: repo.ID,
		Name:         req.Name,
		Color:        req.Color,
		Description:  req.Description,
	}
	
	// Create label
	label, err := h.labelService.Create(c.Request.Context(), serviceReq)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create label")
		if err.Error() == "label with name '"+req.Name+"' already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create label", "details": err.Error()})
		}
		return
	}
	
	c.JSON(http.StatusCreated, label)
}

// GetLabel godoc
// @Summary Get a label
// @Description Get a specific label by name
// @Tags labels
// @Produce json
// @Param owner path string true "Repository owner"
// @Param repo path string true "Repository name"
// @Param name path string true "Label name"
// @Success 200 {object} models.Label
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /repositories/{owner}/{repo}/labels/{name} [get]
func (h *LabelHandlers) GetLabel(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")
	name := c.Param("name")
	
	label, err := h.labelService.Get(c.Request.Context(), owner, repoName, name)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"owner": owner,
			"repo":  repoName,
			"name":  name,
		}).Error("Failed to get label")
		if err.Error() == "label not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Label not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get label"})
		}
		return
	}
	
	c.JSON(http.StatusOK, label)
}

// UpdateLabel godoc
// @Summary Update a label
// @Description Update an existing label
// @Tags labels
// @Accept json
// @Produce json
// @Param owner path string true "Repository owner"
// @Param repo path string true "Repository name"
// @Param name path string true "Label name"
// @Param label body UpdateLabelRequest true "Label update data"
// @Success 200 {object} models.Label
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /repositories/{owner}/{repo}/labels/{name} [patch]
func (h *LabelHandlers) UpdateLabel(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")
	name := c.Param("name")
	
	// Get existing label
	label, err := h.labelService.Get(c.Request.Context(), owner, repoName, name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Label not found"})
		return
	}
	
	// Parse request
	var req UpdateLabelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}
	
	// Convert to service request
	serviceReq := services.UpdateLabelRequest{
		Name:        req.Name,
		Color:       req.Color,
		Description: req.Description,
	}
	
	// Update label
	updatedLabel, err := h.labelService.Update(c.Request.Context(), label.ID, serviceReq)
	if err != nil {
		h.logger.WithError(err).Error("Failed to update label")
		if err.Error() == "label with name '"+*req.Name+"' already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update label", "details": err.Error()})
		}
		return
	}
	
	c.JSON(http.StatusOK, updatedLabel)
}

// DeleteLabel godoc
// @Summary Delete a label
// @Description Delete a label from a repository
// @Tags labels
// @Param owner path string true "Repository owner"
// @Param repo path string true "Repository name"
// @Param name path string true "Label name"
// @Success 204
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /repositories/{owner}/{repo}/labels/{name} [delete]
func (h *LabelHandlers) DeleteLabel(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")
	name := c.Param("name")
	
	// Get existing label
	label, err := h.labelService.Get(c.Request.Context(), owner, repoName, name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Label not found"})
		return
	}
	
	// Delete label
	if err := h.labelService.Delete(c.Request.Context(), label.ID); err != nil {
		h.logger.WithError(err).Error("Failed to delete label")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete label"})
		return
	}
	
	c.Status(http.StatusNoContent)
}

// ListLabels godoc
// @Summary List labels
// @Description Get all labels for a repository
// @Tags labels
// @Produce json
// @Param owner path string true "Repository owner"
// @Param repo path string true "Repository name"
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(100)
// @Success 200 {object} ListLabelsResponse
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /repositories/{owner}/{repo}/labels [get]
func (h *LabelHandlers) ListLabels(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")
	
	// Parse query parameters
	filters := h.parseLabelFilters(c)
	
	// Get labels
	labels, total, err := h.labelService.List(c.Request.Context(), owner, repoName, filters)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"owner": owner,
			"repo":  repoName,
		}).Error("Failed to list labels")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list labels"})
		return
	}
	
	response := ListLabelsResponse{
		Labels:  labels,
		Total:   total,
		Page:    filters.Page,
		PerPage: filters.PerPage,
	}
	
	c.JSON(http.StatusOK, response)
}

// Helper methods

func (h *LabelHandlers) parseLabelFilters(c *gin.Context) services.LabelFilters {
	filters := services.LabelFilters{
		Page:    1,
		PerPage: 100,
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

type CreateLabelRequest struct {
	Name        string `json:"name" binding:"required,max=255"`
	Color       string `json:"color" binding:"required,len=7"`
	Description string `json:"description"`
}

type UpdateLabelRequest struct {
	Name        *string `json:"name,omitempty" binding:"omitempty,max=255"`
	Color       *string `json:"color,omitempty" binding:"omitempty,len=7"`
	Description *string `json:"description,omitempty"`
}

type ListLabelsResponse struct {
	Labels  []*models.Label `json:"labels"`
	Total   int64           `json:"total"`
	Page    int             `json:"page"`
	PerPage int             `json:"per_page"`
}