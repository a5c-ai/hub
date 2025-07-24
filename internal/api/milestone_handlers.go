package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/a5c-ai/hub/internal/models"
	"github.com/a5c-ai/hub/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type MilestoneHandlers struct {
	milestoneService  services.MilestoneService
	repositoryService services.RepositoryService
	logger            *logrus.Logger
}

func NewMilestoneHandlers(
	milestoneService services.MilestoneService,
	repositoryService services.RepositoryService,
	logger *logrus.Logger,
) *MilestoneHandlers {
	return &MilestoneHandlers{
		milestoneService:  milestoneService,
		repositoryService: repositoryService,
		logger:            logger,
	}
}

// CreateMilestone godoc
// @Summary Create a milestone
// @Description Create a new milestone for a repository
// @Tags milestones
// @Accept json
// @Produce json
// @Param owner path string true "Repository owner"
// @Param repo path string true "Repository name"
// @Param milestone body CreateMilestoneRequest true "Milestone data"
// @Success 201 {object} models.Milestone
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /repositories/{owner}/{repo}/milestones [post]
func (h *MilestoneHandlers) CreateMilestone(c *gin.Context) {
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
	var req CreateMilestoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}
	
	// Convert to service request
	serviceReq := services.CreateMilestoneRequest{
		RepositoryID: repo.ID,
		Title:        req.Title,
		Description:  req.Description,
		DueOn:        req.DueOn,
	}
	
	// Create milestone
	milestone, err := h.milestoneService.Create(c.Request.Context(), serviceReq)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create milestone")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create milestone", "details": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, milestone)
}

// GetMilestone godoc
// @Summary Get a milestone
// @Description Get a specific milestone by number
// @Tags milestones
// @Produce json
// @Param owner path string true "Repository owner"
// @Param repo path string true "Repository name"
// @Param number path int true "Milestone number"
// @Success 200 {object} models.Milestone
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /repositories/{owner}/{repo}/milestones/{number} [get]
func (h *MilestoneHandlers) GetMilestone(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")
	numberStr := c.Param("number")
	
	number, err := strconv.Atoi(numberStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid milestone number"})
		return
	}
	
	milestone, err := h.milestoneService.Get(c.Request.Context(), owner, repoName, number)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"owner":  owner,
			"repo":   repoName,
			"number": number,
		}).Error("Failed to get milestone")
		if err.Error() == "milestone not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Milestone not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get milestone"})
		}
		return
	}
	
	c.JSON(http.StatusOK, milestone)
}

// UpdateMilestone godoc
// @Summary Update a milestone
// @Description Update an existing milestone
// @Tags milestones
// @Accept json
// @Produce json
// @Param owner path string true "Repository owner"
// @Param repo path string true "Repository name"
// @Param number path int true "Milestone number"
// @Param milestone body UpdateMilestoneRequest true "Milestone update data"
// @Success 200 {object} models.Milestone
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /repositories/{owner}/{repo}/milestones/{number} [patch]
func (h *MilestoneHandlers) UpdateMilestone(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")
	numberStr := c.Param("number")
	
	number, err := strconv.Atoi(numberStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid milestone number"})
		return
	}
	
	// Get existing milestone
	milestone, err := h.milestoneService.Get(c.Request.Context(), owner, repoName, number)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Milestone not found"})
		return
	}
	
	// Parse request
	var req UpdateMilestoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}
	
	// Convert to service request
	serviceReq := services.UpdateMilestoneRequest{
		Title:       req.Title,
		Description: req.Description,
		State:       req.State,
		DueOn:       req.DueOn,
	}
	
	// Update milestone
	updatedMilestone, err := h.milestoneService.Update(c.Request.Context(), milestone.ID, serviceReq)
	if err != nil {
		h.logger.WithError(err).Error("Failed to update milestone")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update milestone", "details": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, updatedMilestone)
}

// DeleteMilestone godoc
// @Summary Delete a milestone
// @Description Delete a milestone from a repository
// @Tags milestones
// @Param owner path string true "Repository owner"
// @Param repo path string true "Repository name"
// @Param number path int true "Milestone number"
// @Success 204
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /repositories/{owner}/{repo}/milestones/{number} [delete]
func (h *MilestoneHandlers) DeleteMilestone(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")
	numberStr := c.Param("number")
	
	number, err := strconv.Atoi(numberStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid milestone number"})
		return
	}
	
	// Get existing milestone
	milestone, err := h.milestoneService.Get(c.Request.Context(), owner, repoName, number)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Milestone not found"})
		return
	}
	
	// Delete milestone
	if err := h.milestoneService.Delete(c.Request.Context(), milestone.ID); err != nil {
		h.logger.WithError(err).Error("Failed to delete milestone")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete milestone"})
		return
	}
	
	c.Status(http.StatusNoContent)
}

// ListMilestones godoc
// @Summary List milestones
// @Description Get all milestones for a repository
// @Tags milestones
// @Produce json
// @Param owner path string true "Repository owner"
// @Param repo path string true "Repository name"
// @Param state query string false "Milestone state" Enums(open, closed)
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(30)
// @Success 200 {object} ListMilestonesResponse
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /repositories/{owner}/{repo}/milestones [get]
func (h *MilestoneHandlers) ListMilestones(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")
	
	// Parse query parameters
	filters := h.parseMilestoneFilters(c)
	
	// Get milestones
	milestones, total, err := h.milestoneService.List(c.Request.Context(), owner, repoName, filters)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"owner": owner,
			"repo":  repoName,
		}).Error("Failed to list milestones")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list milestones"})
		return
	}
	
	response := ListMilestonesResponse{
		Milestones: milestones,
		Total:      total,
		Page:       filters.Page,
		PerPage:    filters.PerPage,
	}
	
	c.JSON(http.StatusOK, response)
}

// CloseMilestone godoc
// @Summary Close a milestone
// @Description Close a milestone
// @Tags milestones
// @Produce json
// @Param owner path string true "Repository owner"
// @Param repo path string true "Repository name"
// @Param number path int true "Milestone number"
// @Success 200 {object} models.Milestone
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /repositories/{owner}/{repo}/milestones/{number}/close [post]
func (h *MilestoneHandlers) CloseMilestone(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")
	numberStr := c.Param("number")
	
	number, err := strconv.Atoi(numberStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid milestone number"})
		return
	}
	
	// Get existing milestone
	milestone, err := h.milestoneService.Get(c.Request.Context(), owner, repoName, number)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Milestone not found"})
		return
	}
	
	// Close milestone
	updatedMilestone, err := h.milestoneService.Close(c.Request.Context(), milestone.ID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to close milestone")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to close milestone"})
		return
	}
	
	c.JSON(http.StatusOK, updatedMilestone)
}

// ReopenMilestone godoc
// @Summary Reopen a milestone
// @Description Reopen a closed milestone
// @Tags milestones
// @Produce json
// @Param owner path string true "Repository owner"
// @Param repo path string true "Repository name"
// @Param number path int true "Milestone number"
// @Success 200 {object} models.Milestone
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /repositories/{owner}/{repo}/milestones/{number}/reopen [post]
func (h *MilestoneHandlers) ReopenMilestone(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")
	numberStr := c.Param("number")
	
	number, err := strconv.Atoi(numberStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid milestone number"})
		return
	}
	
	// Get existing milestone
	milestone, err := h.milestoneService.Get(c.Request.Context(), owner, repoName, number)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Milestone not found"})
		return
	}
	
	// Reopen milestone
	updatedMilestone, err := h.milestoneService.Reopen(c.Request.Context(), milestone.ID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to reopen milestone")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reopen milestone"})
		return
	}
	
	c.JSON(http.StatusOK, updatedMilestone)
}

// Helper methods

func (h *MilestoneHandlers) parseMilestoneFilters(c *gin.Context) services.MilestoneFilters {
	filters := services.MilestoneFilters{
		Page:    1,
		PerPage: 30,
	}
	
	// State filter
	if state := c.Query("state"); state != "" {
		if state == "open" || state == "closed" {
			filters.State = &state
		}
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

type CreateMilestoneRequest struct {
	Title       string     `json:"title" binding:"required,max=255"`
	Description string     `json:"description"`
	DueOn       *time.Time `json:"due_on"`
}

type UpdateMilestoneRequest struct {
	Title       *string    `json:"title,omitempty" binding:"omitempty,max=255"`
	Description *string    `json:"description,omitempty"`
	State       *string    `json:"state,omitempty" binding:"omitempty,oneof=open closed"`
	DueOn       *time.Time `json:"due_on,omitempty"`
}

type ListMilestonesResponse struct {
	Milestones []*models.Milestone `json:"milestones"`
	Total      int64               `json:"total"`
	Page       int                 `json:"page"`
	PerPage    int                 `json:"per_page"`
}