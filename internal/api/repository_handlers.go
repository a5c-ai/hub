package api

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/a5c-ai/hub/internal/git"
	"github.com/a5c-ai/hub/internal/models"
	"github.com/a5c-ai/hub/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// RepositoryResponse represents a repository with additional fields for API responses
type RepositoryResponse struct {
	models.Repository
	FullName        string     `json:"full_name"`
	Owner           *OwnerInfo `json:"owner,omitempty"`
	Private         bool       `json:"private"`
	Fork            bool       `json:"fork"`
	Language        *string    `json:"language,omitempty"`
	StargazersCount int        `json:"stargazers_count"`
	ForksCount      int        `json:"forks_count"`
	WatchersCount   int        `json:"watchers_count"`
	OpenIssuesCount int        `json:"open_issues_count"`
	CloneURL        string     `json:"clone_url"`
	SSHURL          string     `json:"ssh_url"`
	Size            int64      `json:"size"`
	PushedAt        *string    `json:"pushed_at,omitempty"`
}

// OwnerInfo represents repository owner information
type OwnerInfo struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Type      string    `json:"type"`
	AvatarURL *string   `json:"avatar_url,omitempty"`
}

// RepositoryHandlers contains handlers for repository-related endpoints
type RepositoryHandlers struct {
	repositoryService services.RepositoryService
	branchService     services.BranchService
	gitService        git.GitService
	logger            *logrus.Logger
	db                *gorm.DB
}

// NewRepositoryHandlers creates a new repository handlers instance
func NewRepositoryHandlers(repositoryService services.RepositoryService, branchService services.BranchService, gitService git.GitService, logger *logrus.Logger, db *gorm.DB) *RepositoryHandlers {
	return &RepositoryHandlers{
		repositoryService: repositoryService,
		branchService:     branchService,
		gitService:        gitService,
		logger:            logger,
		db:                db,
	}
}

// convertToRepositoryResponse converts a repository model to a response DTO
func (h *RepositoryHandlers) convertToRepositoryResponse(repo *models.Repository) (*RepositoryResponse, error) {
	// Get owner information
	owner, err := h.getOwnerInfo(repo.OwnerID, repo.OwnerType)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get owner info")
		// Create a fallback owner info
		owner = &OwnerInfo{
			ID:       repo.OwnerID,
			Username: "unknown",
			Type:     string(repo.OwnerType),
		}
	}

	// Construct full_name
	fullName := fmt.Sprintf("%s/%s", owner.Username, repo.Name)

	// Convert pushed_at to string if present
	var pushedAtStr *string
	if repo.PushedAt != nil {
		pushedAtTime := repo.PushedAt.Format("2006-01-02T15:04:05Z")
		pushedAtStr = &pushedAtTime
	}

	// Get primary language from repository statistics
	var primaryLanguage *string
	if _, err := h.repositoryService.GetRepositoryStatistics(context.Background(), repo.ID); err == nil {
		var repoStats models.RepositoryStatistics
		if err := h.db.Where("repository_id = ?", repo.ID).First(&repoStats).Error; err == nil && repoStats.PrimaryLanguage != "" {
			primaryLanguage = &repoStats.PrimaryLanguage
		}
	}

	// Count open issues
	var openIssuesCount int64
	h.db.Model(&models.Issue{}).Where("repository_id = ? AND state = ?", repo.ID, models.IssueStateOpen).Count(&openIssuesCount)

	return &RepositoryResponse{
		Repository:      *repo,
		FullName:        fullName,
		Owner:           owner,
		Private:         repo.Visibility != models.VisibilityPublic,
		Fork:            repo.IsFork,
		Language:        primaryLanguage,
		StargazersCount: repo.StarsCount,
		ForksCount:      repo.ForksCount,
		WatchersCount:   repo.WatchersCount,
		OpenIssuesCount: int(openIssuesCount),
		CloneURL:        fmt.Sprintf("http://localhost:8080/%s/%s.git", owner.Username, repo.Name),
		SSHURL:          fmt.Sprintf("git@localhost:%s/%s.git", owner.Username, repo.Name),
		Size:            repo.SizeKB,
		PushedAt:        pushedAtStr,
	}, nil
}

// getOwnerInfo retrieves owner information based on owner ID and type
func (h *RepositoryHandlers) getOwnerInfo(ownerID uuid.UUID, ownerType models.OwnerType) (*OwnerInfo, error) {
	switch ownerType {
	case models.OwnerTypeUser:
		var user models.User
		if err := h.db.Where("id = ?", ownerID).First(&user).Error; err != nil {
			return nil, fmt.Errorf("failed to get user: %w", err)
		}
		return &OwnerInfo{
			ID:        user.ID,
			Username:  user.Username,
			Type:      "user",
			AvatarURL: &user.AvatarURL,
		}, nil
	case models.OwnerTypeOrganization:
		var org models.Organization
		if err := h.db.Where("id = ?", ownerID).First(&org).Error; err != nil {
			return nil, fmt.Errorf("failed to get organization: %w", err)
		}
		return &OwnerInfo{
			ID:        org.ID,
			Username:  org.Name, // Use organization name as username
			Type:      "organization",
			AvatarURL: &org.AvatarURL,
		}, nil
	default:
		return nil, fmt.Errorf("unknown owner type: %s", ownerType)
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

	// Convert to response DTO with full_name
	repoResponse, err := h.convertToRepositoryResponse(repo)
	if err != nil {
		h.logger.WithError(err).Error("Failed to convert repository to response")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process repository"})
		return
	}

	c.JSON(http.StatusCreated, repoResponse)
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

	// Convert to response DTO with full_name
	repoResponse, err := h.convertToRepositoryResponse(repo)
	if err != nil {
		h.logger.WithError(err).Error("Failed to convert repository to response")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process repository"})
		return
	}

	c.JSON(http.StatusOK, repoResponse)
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

	// Convert repositories to response DTOs with full_name
	var repoResponses []*RepositoryResponse
	for _, repo := range repositories {
		repoResponse, err := h.convertToRepositoryResponse(repo)
		if err != nil {
			h.logger.WithError(err).WithField("repo_id", repo.ID).Warn("Failed to convert repository to response")
			continue // Skip this repository instead of failing the entire request
		}
		repoResponses = append(repoResponses, repoResponse)
	}

	c.Header("X-Total-Count", strconv.FormatInt(total, 10))
	c.JSON(http.StatusOK, repoResponses)
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

// Git Content Endpoints

// GetCommits handles GET /api/v1/repositories/{owner}/{repo}/commits
func (h *RepositoryHandlers) GetCommits(c *gin.Context) {
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

	// Get repository path
	repoPath, err := h.repositoryService.GetRepositoryPath(c.Request.Context(), repo.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository path"})
		return
	}

	// Parse query parameters
	var opts git.CommitOptions
	opts.Branch = c.Query("sha") // Git reference (branch, tag, or commit SHA)
	if opts.Branch == "" {
		opts.Branch = repo.DefaultBranch
	}

	if page := c.Query("page"); page != "" {
		if val, err := strconv.Atoi(page); err == nil && val > 0 {
			opts.Page = val - 1 // Convert to 0-based
		}
	}

	if perPage := c.Query("per_page"); perPage != "" {
		if val, err := strconv.Atoi(perPage); err == nil && val > 0 && val <= 100 {
			opts.PerPage = val
		}
	} else {
		opts.PerPage = 30
	}

	commits, err := h.gitService.GetCommits(c.Request.Context(), repoPath, opts)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get commits")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get commits"})
		return
	}

	c.JSON(http.StatusOK, commits)
}

// GetCommit handles GET /api/v1/repositories/{owner}/{repo}/commits/{sha}
func (h *RepositoryHandlers) GetCommit(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")
	sha := c.Param("sha")

	if owner == "" || repoName == "" || sha == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner, repository name, and commit SHA are required"})
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

	// Get repository path
	repoPath, err := h.repositoryService.GetRepositoryPath(c.Request.Context(), repo.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository path"})
		return
	}

	commit, err := h.gitService.GetCommit(c.Request.Context(), repoPath, sha)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get commit")
		c.JSON(http.StatusNotFound, gin.H{"error": "Commit not found"})
		return
	}

	c.JSON(http.StatusOK, commit)
}

// GetTree handles GET /api/v1/repositories/{owner}/{repo}/contents/{path}
func (h *RepositoryHandlers) GetTree(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")
	path := c.Param("path")

	// Clean up the path - remove leading slash if present
	if strings.HasPrefix(path, "/") {
		path = strings.TrimPrefix(path, "/")
	}

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

	// Get repository path
	repoPath, err := h.repositoryService.GetRepositoryPath(c.Request.Context(), repo.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository path"})
		return
	}

	// Get reference (branch, tag, or commit SHA)
	ref := c.Query("ref")
	if ref == "" {
		ref = repo.DefaultBranch
	}

	// First try to get as a tree (directory)
	tree, err := h.gitService.GetTree(c.Request.Context(), repoPath, ref, path)
	if err != nil {
		// If that fails, try to get as a file
		file, fileErr := h.gitService.GetFile(c.Request.Context(), repoPath, ref, path)
		if fileErr != nil {
			// If both fail, return the original tree error
			h.logger.WithError(err).Error("Failed to get tree or file")
			c.JSON(http.StatusNotFound, gin.H{"error": "Path not found"})
			return
		}
		// Return the file content
		c.JSON(http.StatusOK, file)
		return
	}

	// Return the tree content
	c.JSON(http.StatusOK, tree)
}

// GetFile handles GET /api/v1/repositories/{owner}/{repo}/contents/{path} (for files)
func (h *RepositoryHandlers) GetFile(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")
	path := c.Param("path")

	// Clean up the path - remove leading slash if present
	if strings.HasPrefix(path, "/") {
		path = strings.TrimPrefix(path, "/")
	}

	if owner == "" || repoName == "" || path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner, repository name, and file path are required"})
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

	// Get repository path
	repoPath, err := h.repositoryService.GetRepositoryPath(c.Request.Context(), repo.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository path"})
		return
	}

	// Get reference (branch, tag, or commit SHA)
	ref := c.Query("ref")
	if ref == "" {
		ref = repo.DefaultBranch
	}

	file, err := h.gitService.GetFile(c.Request.Context(), repoPath, ref, path)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get file")
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	c.JSON(http.StatusOK, file)
}

// GetRepositoryInfo handles GET /api/v1/repositories/{owner}/{repo}/info
func (h *RepositoryHandlers) GetRepositoryInfo(c *gin.Context) {
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

	// Get repository path
	repoPath, err := h.repositoryService.GetRepositoryPath(c.Request.Context(), repo.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository path"})
		return
	}

	info, err := h.gitService.GetRepositoryInfo(c.Request.Context(), repoPath)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get repository info")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository info"})
		return
	}

	c.JSON(http.StatusOK, info)
}

// CreateFile handles POST /api/v1/repositories/{owner}/{repo}/contents/{path}
func (h *RepositoryHandlers) CreateFile(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")
	path := c.Param("path")

	// Clean up the path - remove leading slash if present
	if strings.HasPrefix(path, "/") {
		path = strings.TrimPrefix(path, "/")
	}

	if owner == "" || repoName == "" || path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner, repository name, and file path are required"})
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

	// Parse request body
	var req git.CreateFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Set path from URL parameter
	req.Path = path

	// Get repository path
	repoPath, err := h.repositoryService.GetRepositoryPath(c.Request.Context(), repo.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository path"})
		return
	}

	// Create the file
	commit, err := h.gitService.CreateFile(c.Request.Context(), repoPath, req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create file")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create file", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"content": gin.H{
			"name":     path,
			"path":     path,
			"sha":      commit.SHA,
			"size":     len(req.Content),
			"type":     "file",
			"encoding": req.Encoding,
		},
		"commit": commit,
	})
}

// UpdateFile handles PUT /api/v1/repositories/{owner}/{repo}/contents/{path}
func (h *RepositoryHandlers) UpdateFile(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")
	path := c.Param("path")

	// Clean up the path - remove leading slash if present
	if strings.HasPrefix(path, "/") {
		path = strings.TrimPrefix(path, "/")
	}

	if owner == "" || repoName == "" || path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner, repository name, and file path are required"})
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

	// Parse request body
	var req git.UpdateFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Set path from URL parameter
	req.Path = path

	// Get repository path
	repoPath, err := h.repositoryService.GetRepositoryPath(c.Request.Context(), repo.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository path"})
		return
	}

	// Update the file
	commit, err := h.gitService.UpdateFile(c.Request.Context(), repoPath, req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to update file")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update file", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"content": gin.H{
			"name":     path,
			"path":     path,
			"sha":      commit.SHA,
			"size":     len(req.Content),
			"type":     "file",
			"encoding": req.Encoding,
		},
		"commit": commit,
	})
}

// DeleteFile handles DELETE /api/v1/repositories/{owner}/{repo}/contents/{path}
func (h *RepositoryHandlers) DeleteFile(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")
	path := c.Param("path")

	// Clean up the path - remove leading slash if present
	if strings.HasPrefix(path, "/") {
		path = strings.TrimPrefix(path, "/")
	}

	if owner == "" || repoName == "" || path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner, repository name, and file path are required"})
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

	// Parse request body
	var req git.DeleteFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Set path from URL parameter
	req.Path = path

	// Get repository path
	repoPath, err := h.repositoryService.GetRepositoryPath(c.Request.Context(), repo.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository path"})
		return
	}

	// Delete the file
	commit, err := h.gitService.DeleteFile(c.Request.Context(), repoPath, req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to delete file")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"commit": commit,
	})
}

// GetRepositoryStats handles GET /api/v1/repositories/{owner}/{repo}/stats
func (h *RepositoryHandlers) GetRepositoryStats(c *gin.Context) {
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

	// Get repository path
	repoPath, err := h.repositoryService.GetRepositoryPath(c.Request.Context(), repo.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository path"})
		return
	}

	stats, err := h.gitService.GetRepositoryStats(c.Request.Context(), repoPath)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get repository stats")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetRepositoryLanguages handles GET /api/v1/repositories/{owner}/{repo}/languages
func (h *RepositoryHandlers) GetRepositoryLanguages(c *gin.Context) {
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

	// Get repository path
	repoPath, err := h.repositoryService.GetRepositoryPath(c.Request.Context(), repo.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository path"})
		return
	}

	stats, err := h.gitService.GetRepositoryStats(c.Request.Context(), repoPath)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get repository stats")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository stats"})
		return
	}

	c.JSON(http.StatusOK, stats.Languages)
}

// GetRepositoryTags handles GET /api/v1/repositories/{owner}/{repo}/tags
func (h *RepositoryHandlers) GetRepositoryTags(c *gin.Context) {
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

	// Get repository path
	repoPath, err := h.repositoryService.GetRepositoryPath(c.Request.Context(), repo.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository path"})
		return
	}

	tags, err := h.gitService.GetTags(c.Request.Context(), repoPath)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get repository tags")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository tags"})
		return
	}

	c.JSON(http.StatusOK, tags)
}

// CompareBranches handles GET /api/v1/repositories/{owner}/{repo}/compare/{base}...{head}
func (h *RepositoryHandlers) CompareBranches(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")
	base := c.Param("base")
	head := c.Param("head")

	if owner == "" || repoName == "" || base == "" || head == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner, repository name, base, and head are required"})
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

	// Get repository path
	repoPath, err := h.repositoryService.GetRepositoryPath(c.Request.Context(), repo.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository path"})
		return
	}

	comparison, err := h.gitService.CompareRefs(repoPath, base, head)
	if err != nil {
		h.logger.WithError(err).Error("Failed to compare branches")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to compare branches", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, comparison)
}

// GetMergeBase handles GET /api/v1/repositories/{owner}/{repo}/compare/{base}...HEAD
func (h *RepositoryHandlers) GetMergeBase(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")
	base := c.Param("base")

	if owner == "" || repoName == "" || base == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner, repository name, and base are required"})
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

	// Get repository path
	repoPath, err := h.repositoryService.GetRepositoryPath(c.Request.Context(), repo.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository path"})
		return
	}

	comparison, err := h.gitService.CompareRefs(repoPath, base, "HEAD")
	if err != nil {
		h.logger.WithError(err).Error("Failed to compare with HEAD")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to compare with HEAD", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, comparison)
}

// StarRepository handles PUT /api/v1/repositories/{owner}/{repo}/star
func (h *RepositoryHandlers) StarRepository(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")

	if owner == "" || repoName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner and repository name are required"})
		return
	}

	// Get authenticated user
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get repository
	repo, err := h.repositoryService.Get(c.Request.Context(), owner, repoName)
	if err != nil {
		if err.Error() == "repository not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository"})
		}
		return
	}

	// Create star record
	star := models.Star{
		UserID:       userID.(uuid.UUID),
		RepositoryID: repo.ID,
	}

	if err := h.db.Create(&star).Error; err != nil {
		// Check if it's a duplicate key error (already starred)
		if strings.Contains(err.Error(), "unique_user_repository_star") || strings.Contains(err.Error(), "duplicate key") {
			c.JSON(http.StatusOK, gin.H{"message": "Repository already starred"})
			return
		}
		h.logger.WithError(err).Error("Failed to star repository")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to star repository"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Repository starred successfully"})
}

// UnstarRepository handles DELETE /api/v1/repositories/{owner}/{repo}/star
func (h *RepositoryHandlers) UnstarRepository(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")

	if owner == "" || repoName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner and repository name are required"})
		return
	}

	// Get authenticated user
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get repository
	repo, err := h.repositoryService.Get(c.Request.Context(), owner, repoName)
	if err != nil {
		if err.Error() == "repository not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository"})
		}
		return
	}

	// Delete star record
	result := h.db.Where("user_id = ? AND repository_id = ?", userID.(uuid.UUID), repo.ID).Delete(&models.Star{})
	if result.Error != nil {
		h.logger.WithError(result.Error).Error("Failed to unstar repository")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unstar repository"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Repository not starred by user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Repository unstarred successfully"})
}

// CheckStarred handles GET /api/v1/repositories/{owner}/{repo}/star
func (h *RepositoryHandlers) CheckStarred(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")

	if owner == "" || repoName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner and repository name are required"})
		return
	}

	// Get authenticated user
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get repository
	repo, err := h.repositoryService.Get(c.Request.Context(), owner, repoName)
	if err != nil {
		if err.Error() == "repository not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository"})
		}
		return
	}

	// Check if user has starred the repository
	var count int64
	if err := h.db.Model(&models.Star{}).Where("user_id = ? AND repository_id = ?", userID.(uuid.UUID), repo.ID).Count(&count).Error; err != nil {
		h.logger.WithError(err).Error("Failed to check star status")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check star status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"starred": count > 0})
}

// ForkRepository handles POST /api/v1/repositories/{owner}/{repo}/fork
func (h *RepositoryHandlers) ForkRepository(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")

	if owner == "" || repoName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner and repository name are required"})
		return
	}

	// Get authenticated user
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get repository
	repo, err := h.repositoryService.Get(c.Request.Context(), owner, repoName)
	if err != nil {
		if err.Error() == "repository not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository"})
		}
		return
	}

	// Parse request body for optional fork parameters
	var forkReq struct {
		Name         string `json:"name,omitempty"`
		Organization string `json:"organization,omitempty"`
	}
	if err := c.ShouldBindJSON(&forkReq); err != nil {
		// Use default values if JSON parsing fails
	}

	// Use original repository name if no new name provided
	if forkReq.Name == "" {
		forkReq.Name = repo.Name
	}

	// Create fork request
	forkRequest := services.ForkRequest{
		Name:      forkReq.Name,
		OwnerID:   userID.(uuid.UUID),
		OwnerType: models.OwnerTypeUser,
	}

	// Use the repository service to fork the repository
	fork, err := h.repositoryService.Fork(c.Request.Context(), repo.ID, forkRequest)
	if err != nil {
		h.logger.WithError(err).Error("Failed to fork repository")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fork repository: " + err.Error()})
		return
	}

	// Return the created fork
	response, err := h.convertToRepositoryResponse(fork)
	if err != nil {
		h.logger.WithError(err).Error("Failed to convert fork to response")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to format fork response"})
		return
	}
	c.JSON(http.StatusCreated, response)
}

// TransferRepository handles POST /api/v1/repositories/{owner}/{repo}/transfer
func (h *RepositoryHandlers) TransferRepository(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")

	if owner == "" || repoName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner and repository name are required"})
		return
	}

	// Get repository
	repo, err := h.repositoryService.Get(c.Request.Context(), owner, repoName)
	if err != nil {
		if err.Error() == "repository not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository"})
		}
		return
	}

	var transferReq struct {
		NewOwnerID   uuid.UUID        `json:"new_owner_id" binding:"required"`
		NewOwnerType models.OwnerType `json:"new_owner_type" binding:"required"`
	}
	if err := c.ShouldBindJSON(&transferReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Create transfer request
	transferRequest := services.TransferRequest{
		NewOwnerID:   transferReq.NewOwnerID,
		NewOwnerType: transferReq.NewOwnerType,
	}

	// Transfer the repository
	if err := h.repositoryService.Transfer(c.Request.Context(), repo.ID, transferRequest); err != nil {
		h.logger.WithError(err).Error("Failed to transfer repository")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to transfer repository: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Repository transferred successfully"})
}

// UpdateRepositoryStats handles POST /api/v1/repositories/{owner}/{repo}/stats/update
func (h *RepositoryHandlers) UpdateRepositoryStats(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")

	if owner == "" || repoName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner and repository name are required"})
		return
	}

	// Get repository
	repo, err := h.repositoryService.Get(c.Request.Context(), owner, repoName)
	if err != nil {
		if err.Error() == "repository not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository"})
		}
		return
	}

	// Update repository statistics
	if err := h.repositoryService.UpdateRepositoryStats(c.Request.Context(), repo.ID); err != nil {
		h.logger.WithError(err).Error("Failed to update repository statistics")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update repository statistics: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Repository statistics updated successfully"})
}

// GetRepositoryStatistics handles GET /api/v1/repositories/{owner}/{repo}/stats
func (h *RepositoryHandlers) GetRepositoryStatistics(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")

	if owner == "" || repoName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner and repository name are required"})
		return
	}

	// Get repository
	repo, err := h.repositoryService.Get(c.Request.Context(), owner, repoName)
	if err != nil {
		if err.Error() == "repository not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository"})
		}
		return
	}

	// Get repository statistics
	stats, err := h.repositoryService.GetRepositoryStatistics(c.Request.Context(), repo.ID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get repository statistics")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository statistics: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// CreateGitHook handles POST /api/v1/repositories/{owner}/{repo}/hooks
func (h *RepositoryHandlers) CreateGitHook(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")

	if owner == "" || repoName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner and repository name are required"})
		return
	}

	// Get repository
	repo, err := h.repositoryService.Get(c.Request.Context(), owner, repoName)
	if err != nil {
		if err.Error() == "repository not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository"})
		}
		return
	}

	var hookReq services.CreateGitHookRequest
	if err := c.ShouldBindJSON(&hookReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Create Git hook
	hook, err := h.repositoryService.CreateGitHook(c.Request.Context(), repo.ID, hookReq)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create Git hook")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Git hook: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, hook)
}

// GetGitHooks handles GET /api/v1/repositories/{owner}/{repo}/hooks
func (h *RepositoryHandlers) GetGitHooks(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")

	if owner == "" || repoName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner and repository name are required"})
		return
	}

	// Get repository
	repo, err := h.repositoryService.Get(c.Request.Context(), owner, repoName)
	if err != nil {
		if err.Error() == "repository not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository"})
		}
		return
	}

	// Get Git hooks
	hooks, err := h.repositoryService.GetGitHooks(c.Request.Context(), repo.ID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get Git hooks")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Git hooks: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, hooks)
}

// UpdateGitHook handles PUT /api/v1/repositories/{owner}/{repo}/hooks/{hookId}
func (h *RepositoryHandlers) UpdateGitHook(c *gin.Context) {
	hookIDStr := c.Param("hookId")
	hookID, err := uuid.Parse(hookIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid hook ID"})
		return
	}

	var updateReq services.UpdateGitHookRequest
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Update Git hook
	hook, err := h.repositoryService.UpdateGitHook(c.Request.Context(), hookID, updateReq)
	if err != nil {
		h.logger.WithError(err).Error("Failed to update Git hook")
		if err.Error() == "Git hook not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Git hook not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update Git hook: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, hook)
}

// DeleteGitHook handles DELETE /api/v1/repositories/{owner}/{repo}/hooks/{hookId}
func (h *RepositoryHandlers) DeleteGitHook(c *gin.Context) {
	hookIDStr := c.Param("hookId")
	hookID, err := uuid.Parse(hookIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid hook ID"})
		return
	}

	// Delete Git hook
	if err := h.repositoryService.DeleteGitHook(c.Request.Context(), hookID); err != nil {
		h.logger.WithError(err).Error("Failed to delete Git hook")
		if err.Error() == "Git hook not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Git hook not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete Git hook: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Git hook deleted successfully"})
}

// CreateTemplate handles POST /api/v1/repositories/{owner}/{repo}/template
func (h *RepositoryHandlers) CreateTemplate(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")

	if owner == "" || repoName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner and repository name are required"})
		return
	}

	// Get repository
	repo, err := h.repositoryService.Get(c.Request.Context(), owner, repoName)
	if err != nil {
		if err.Error() == "repository not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository"})
		}
		return
	}

	var templateReq services.CreateTemplateRequest
	if err := c.ShouldBindJSON(&templateReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Create template
	template, err := h.repositoryService.CreateTemplate(c.Request.Context(), repo.ID, templateReq)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create template")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create template: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, template)
}

// GetTemplates handles GET /api/v1/templates
func (h *RepositoryHandlers) GetTemplates(c *gin.Context) {
	// Parse query parameters
	filters := services.TemplateFilters{}
	if category := c.Query("category"); category != "" {
		filters.Category = category
	}
	if featured := c.Query("featured"); featured != "" {
		if featured == "true" {
			filters.IsFeatured = &[]bool{true}[0]
		} else if featured == "false" {
			filters.IsFeatured = &[]bool{false}[0]
		}
	}
	if public := c.Query("public"); public != "" {
		if public == "true" {
			filters.IsPublic = &[]bool{true}[0]
		} else if public == "false" {
			filters.IsPublic = &[]bool{false}[0]
		}
	}
	if search := c.Query("search"); search != "" {
		filters.Search = search
	}
	if sort := c.Query("sort"); sort != "" {
		filters.Sort = sort
	}
	if direction := c.Query("direction"); direction != "" {
		filters.Direction = direction
	}
	if page := c.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p >= 0 {
			filters.Page = p
		}
	}
	if perPage := c.Query("per_page"); perPage != "" {
		if pp, err := strconv.Atoi(perPage); err == nil && pp > 0 && pp <= 100 {
			filters.PerPage = pp
		}
	}

	// Get templates
	templates, err := h.repositoryService.GetTemplates(c.Request.Context(), filters)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get templates")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get templates: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, templates)
}

// UseTemplate handles POST /api/v1/templates/{templateId}/use
func (h *RepositoryHandlers) UseTemplate(c *gin.Context) {
	templateIDStr := c.Param("templateId")
	templateID, err := uuid.Parse(templateIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	var createReq services.CreateRepositoryRequest
	if err := c.ShouldBindJSON(&createReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Use template to create repository
	repo, err := h.repositoryService.UseTemplate(c.Request.Context(), templateID, createReq)
	if err != nil {
		h.logger.WithError(err).Error("Failed to use template")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to use template: " + err.Error()})
		return
	}

	// Return the created repository
	response, err := h.convertToRepositoryResponse(repo)
	if err != nil {
		h.logger.WithError(err).Error("Failed to convert repository to response")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to format repository response"})
		return
	}

	c.JSON(http.StatusCreated, response)
}
