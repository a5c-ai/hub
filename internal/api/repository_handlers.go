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

// RepositoryHandlers contains handlers for repository-related endpoints
type RepositoryHandlers struct {
	repositoryService services.RepositoryService
	branchService     services.BranchService
	logger           *logrus.Logger
}

// NewRepositoryHandlers creates a new repository handlers instance
func NewRepositoryHandlers(repositoryService services.RepositoryService, branchService services.BranchService, logger *logrus.Logger) *RepositoryHandlers {
	return &RepositoryHandlers{
		repositoryService: repositoryService,
		branchService:     branchService,
		logger:           logger,
	}
}

// CreateRepository handles POST /api/v1/repositories
func (h *RepositoryHandlers) CreateRepository(c *gin.Context) {
	var req services.CreateRepositoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Set owner ID from authenticated user if not provided
	if req.OwnerID == uuid.Nil {
		if uid, ok := userID.(uuid.UUID); ok {
			req.OwnerID = uid
			req.OwnerType = "user"
		} else if uidStr, ok := userID.(string); ok {
			if uid, err := uuid.Parse(uidStr); err == nil {
				req.OwnerID = uid
				req.OwnerType = "user"
			}
		}
	}

	repo, err := h.repositoryService.Create(c.Request.Context(), req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create repository")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create repository", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, repo)
}

// GetRepository handles GET /api/v1/repositories/{owner}/{repo}
func (h *RepositoryHandlers) GetRepository(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")

	if owner == "" || repoName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner and repository name are required"})
		return
	}

	repo, err := h.repositoryService.Get(c.Request.Context(), owner, repoName)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"owner": owner,
			"repo":  repoName,
		}).Error("Failed to get repository")
		
		if err.Error() == "repository not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository"})
		}
		return
	}

	c.JSON(http.StatusOK, repo)
}

// UpdateRepository handles PATCH /api/v1/repositories/{owner}/{repo}
func (h *RepositoryHandlers) UpdateRepository(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")

	if owner == "" || repoName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner and repository name are required"})
		return
	}

	// Get repository first to get ID
	repo, err := h.repositoryService.Get(c.Request.Context(), owner, repoName)
	if err != nil {
		if err.Error() == "repository not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository"})
		}
		return
	}

	var req services.UpdateRepositoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	updatedRepo, err := h.repositoryService.Update(c.Request.Context(), repo.ID, req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to update repository")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update repository", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedRepo)
}

// DeleteRepository handles DELETE /api/v1/repositories/{owner}/{repo}
func (h *RepositoryHandlers) DeleteRepository(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")

	if owner == "" || repoName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner and repository name are required"})
		return
	}

	// Get repository first to get ID
	repo, err := h.repositoryService.Get(c.Request.Context(), owner, repoName)
	if err != nil {
		if err.Error() == "repository not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository"})
		}
		return
	}

	if err := h.repositoryService.Delete(c.Request.Context(), repo.ID); err != nil {
		h.logger.WithError(err).Error("Failed to delete repository")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete repository", "details": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ListRepositories handles GET /api/v1/repositories
func (h *RepositoryHandlers) ListRepositories(c *gin.Context) {
	var filters services.RepositoryFilters

	// Parse query parameters
	if ownerID := c.Query("owner_id"); ownerID != "" {
		if uid, err := uuid.Parse(ownerID); err == nil {
			filters.OwnerID = &uid
		}
	}

	if ownerType := c.Query("owner_type"); ownerType != "" {
		if ot := parseOwnerType(ownerType); ot != "" {
			filters.OwnerType = &ot
		}
	}

	if visibility := c.Query("visibility"); visibility != "" {
		if v := parseVisibility(visibility); v != "" {
			filters.Visibility = &v
		}
	}

	if isTemplate := c.Query("is_template"); isTemplate != "" {
		if val, err := strconv.ParseBool(isTemplate); err == nil {
			filters.IsTemplate = &val
		}
	}

	if isArchived := c.Query("is_archived"); isArchived != "" {
		if val, err := strconv.ParseBool(isArchived); err == nil {
			filters.IsArchived = &val
		}
	}

	if isFork := c.Query("is_fork"); isFork != "" {
		if val, err := strconv.ParseBool(isFork); err == nil {
			filters.IsFork = &val
		}
	}

	filters.Search = c.Query("q")
	filters.Language = c.Query("language")
	filters.Sort = c.Query("sort")
	filters.Direction = c.Query("direction")

	if page := c.Query("page"); page != "" {
		if val, err := strconv.Atoi(page); err == nil && val > 0 {
			filters.Page = val - 1 // Convert to 0-based
		}
	}

	if perPage := c.Query("per_page"); perPage != "" {
		if val, err := strconv.Atoi(perPage); err == nil && val > 0 && val <= 100 {
			filters.PerPage = val
		}
	}

	repositories, total, err := h.repositoryService.List(c.Request.Context(), filters)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list repositories")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list repositories"})
		return
	}

	c.Header("X-Total-Count", strconv.FormatInt(total, 10))
	c.JSON(http.StatusOK, repositories)
}

// GetBranches handles GET /api/v1/repositories/{owner}/{repo}/branches
func (h *RepositoryHandlers) GetBranches(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")

	if owner == "" || repoName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner and repository name are required"})
		return
	}

	// Get repository first to get ID
	repo, err := h.repositoryService.Get(c.Request.Context(), owner, repoName)
	if err != nil {
		if err.Error() == "repository not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository"})
		}
		return
	}

	branches, err := h.branchService.List(c.Request.Context(), repo.ID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list branches")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list branches"})
		return
	}

	c.JSON(http.StatusOK, branches)
}

// GetBranch handles GET /api/v1/repositories/{owner}/{repo}/branches/{branch}
func (h *RepositoryHandlers) GetBranch(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")
	branchName := c.Param("branch")

	if owner == "" || repoName == "" || branchName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner, repository name, and branch name are required"})
		return
	}

	// Get repository first to get ID
	repo, err := h.repositoryService.Get(c.Request.Context(), owner, repoName)
	if err != nil {
		if err.Error() == "repository not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository"})
		}
		return
	}

	branch, err := h.branchService.Get(c.Request.Context(), repo.ID, branchName)
	if err != nil {
		if err.Error() == "branch not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Branch not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get branch"})
		}
		return
	}

	c.JSON(http.StatusOK, branch)
}

// CreateBranch handles POST /api/v1/repositories/{owner}/{repo}/branches
func (h *RepositoryHandlers) CreateBranch(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")

	if owner == "" || repoName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner and repository name are required"})
		return
	}

	// Get repository first to get ID
	repo, err := h.repositoryService.Get(c.Request.Context(), owner, repoName)
	if err != nil {
		if err.Error() == "repository not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository"})
		}
		return
	}

	var req services.CreateBranchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	branch, err := h.branchService.Create(c.Request.Context(), repo.ID, req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create branch")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create branch", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, branch)
}

// DeleteBranch handles DELETE /api/v1/repositories/{owner}/{repo}/branches/{branch}
func (h *RepositoryHandlers) DeleteBranch(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")
	branchName := c.Param("branch")

	if owner == "" || repoName == "" || branchName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner, repository name, and branch name are required"})
		return
	}

	// Get repository first to get ID
	repo, err := h.repositoryService.Get(c.Request.Context(), owner, repoName)
	if err != nil {
		if err.Error() == "repository not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository"})
		}
		return
	}

	if err := h.branchService.Delete(c.Request.Context(), repo.ID, branchName); err != nil {
		h.logger.WithError(err).Error("Failed to delete branch")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete branch", "details": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// Helper functions
func parseOwnerType(s string) models.OwnerType {
	switch s {
	case "user":
		return models.OwnerTypeUser
	case "organization":
		return models.OwnerTypeOrganization
	default:
		return ""
	}
}

func parseVisibility(s string) models.Visibility {
	switch s {
	case "public":
		return models.VisibilityPublic
	case "private":
		return models.VisibilityPrivate
	case "internal":
		return models.VisibilityInternal
	default:
		return ""
	}
}