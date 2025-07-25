package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/a5c-ai/hub/internal/git"
	"github.com/a5c-ai/hub/internal/models"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// RepositoryService provides repository management operations
type RepositoryService interface {
	// Repository CRUD operations
	Create(ctx context.Context, req CreateRepositoryRequest) (*models.Repository, error)
	Get(ctx context.Context, owner, name string) (*models.Repository, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Repository, error)
	Update(ctx context.Context, id uuid.UUID, req UpdateRepositoryRequest) (*models.Repository, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filters RepositoryFilters) ([]*models.Repository, int64, error)

	// Repository operations
	Fork(ctx context.Context, id uuid.UUID, req ForkRequest) (*models.Repository, error)
	Transfer(ctx context.Context, id uuid.UUID, req TransferRequest) error
	Archive(ctx context.Context, id uuid.UUID) error
	Unarchive(ctx context.Context, id uuid.UUID) error

	// Git operations
	InitializeGitRepository(ctx context.Context, repoID uuid.UUID) error
	GetRepositoryPath(ctx context.Context, repoID uuid.UUID) (string, error)
	SyncCommits(ctx context.Context, repoID uuid.UUID) error

	// Statistics and language detection
	UpdateRepositoryStats(ctx context.Context, repoID uuid.UUID) error
	GetLanguages(ctx context.Context, repoID uuid.UUID) (map[string]git.LanguageStats, error)
	GetRepositoryStatistics(ctx context.Context, repoID uuid.UUID) (*git.RepositoryStats, error)

	// Git hooks management
	CreateGitHook(ctx context.Context, repoID uuid.UUID, req CreateGitHookRequest) (*models.GitHook, error)
	UpdateGitHook(ctx context.Context, hookID uuid.UUID, req UpdateGitHookRequest) (*models.GitHook, error)
	DeleteGitHook(ctx context.Context, hookID uuid.UUID) error
	GetGitHooks(ctx context.Context, repoID uuid.UUID) ([]*models.GitHook, error)
	
	// Repository templates
	CreateTemplate(ctx context.Context, repoID uuid.UUID, req CreateTemplateRequest) (*models.RepositoryTemplate, error)
	GetTemplates(ctx context.Context, filters TemplateFilters) ([]*models.RepositoryTemplate, error)
	UseTemplate(ctx context.Context, templateID uuid.UUID, req CreateRepositoryRequest) (*models.Repository, error)
}

// CreateRepositoryRequest represents a request to create a repository
type CreateRepositoryRequest struct {
	OwnerID             uuid.UUID         `json:"owner_id"`
	OwnerType           models.OwnerType  `json:"owner_type"`
	Name                string            `json:"name"`
	Description         string            `json:"description,omitempty"`
	DefaultBranch       string            `json:"default_branch,omitempty"`
	Visibility          models.Visibility `json:"visibility"`
	IsTemplate          bool              `json:"is_template,omitempty"`
	HasIssues           bool              `json:"has_issues"`
	HasProjects         bool              `json:"has_projects"`
	HasWiki             bool              `json:"has_wiki"`
	HasDownloads        bool              `json:"has_downloads"`
	AllowMergeCommit    bool              `json:"allow_merge_commit"`
	AllowSquashMerge    bool              `json:"allow_squash_merge"`
	AllowRebaseMerge    bool              `json:"allow_rebase_merge"`
	DeleteBranchOnMerge bool              `json:"delete_branch_on_merge"`
	AutoInit            bool              `json:"auto_init"` // Initialize with README
}

// UpdateRepositoryRequest represents a request to update a repository
type UpdateRepositoryRequest struct {
	Name                *string            `json:"name,omitempty"`
	Description         *string            `json:"description,omitempty"`
	DefaultBranch       *string            `json:"default_branch,omitempty"`
	Visibility          *models.Visibility `json:"visibility,omitempty"`
	IsTemplate          *bool              `json:"is_template,omitempty"`
	HasIssues           *bool              `json:"has_issues,omitempty"`
	HasProjects         *bool              `json:"has_projects,omitempty"`
	HasWiki             *bool              `json:"has_wiki,omitempty"`
	HasDownloads        *bool              `json:"has_downloads,omitempty"`
	AllowMergeCommit    *bool              `json:"allow_merge_commit,omitempty"`
	AllowSquashMerge    *bool              `json:"allow_squash_merge,omitempty"`
	AllowRebaseMerge    *bool              `json:"allow_rebase_merge,omitempty"`
	DeleteBranchOnMerge *bool              `json:"delete_branch_on_merge,omitempty"`
}

// ForkRequest represents a request to fork a repository
type ForkRequest struct {
	Name      string           `json:"name,omitempty"` // New name for the fork
	OwnerID   uuid.UUID        `json:"owner_id"`       // New owner
	OwnerType models.OwnerType `json:"owner_type"`
}

// TransferRequest represents a request to transfer a repository
type TransferRequest struct {
	NewOwnerID   uuid.UUID        `json:"new_owner_id"`
	NewOwnerType models.OwnerType `json:"new_owner_type"`
}

// RepositoryFilters represents filters for listing repositories
type RepositoryFilters struct {
	OwnerID    *uuid.UUID         `json:"owner_id,omitempty"`
	OwnerType  *models.OwnerType  `json:"owner_type,omitempty"`
	Visibility *models.Visibility `json:"visibility,omitempty"`
	IsTemplate *bool              `json:"is_template,omitempty"`
	IsArchived *bool              `json:"is_archived,omitempty"`
	IsFork     *bool              `json:"is_fork,omitempty"`
	Search     string             `json:"search,omitempty"` // Search in name and description
	Language   string             `json:"language,omitempty"`
	Sort       string             `json:"sort,omitempty"`      // name, created, updated, pushed, stars, forks
	Direction  string             `json:"direction,omitempty"` // asc, desc
	Page       int                `json:"page,omitempty"`
	PerPage    int                `json:"per_page,omitempty"`
}

// CreateGitHookRequest represents a request to create a Git hook
type CreateGitHookRequest struct {
	HookType  string `json:"hook_type"`  // pre-receive, post-receive, update, etc.
	Script    string `json:"script"`     // Hook script content
	Language  string `json:"language"`   // bash, python, etc.
	IsEnabled bool   `json:"is_enabled"` // Whether the hook is enabled
	Order     int    `json:"order"`      // Execution order for multiple hooks
}

// UpdateGitHookRequest represents a request to update a Git hook
type UpdateGitHookRequest struct {
	Script    *string `json:"script,omitempty"`
	Language  *string `json:"language,omitempty"`
	IsEnabled *bool   `json:"is_enabled,omitempty"`
	Order     *int    `json:"order,omitempty"`
}

// CreateTemplateRequest represents a request to create a repository template
type CreateTemplateRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
	IsFeatured  bool     `json:"is_featured"`
	IsPublic    bool     `json:"is_public"`
}

// TemplateFilters represents filters for listing repository templates
type TemplateFilters struct {
	Category   string `json:"category,omitempty"`
	IsFeatured *bool  `json:"is_featured,omitempty"`
	IsPublic   *bool  `json:"is_public,omitempty"`
	Search     string `json:"search,omitempty"` // Search in name and description
	Sort       string `json:"sort,omitempty"`   // name, created, usage_count
	Direction  string `json:"direction,omitempty"` // asc, desc
	Page       int    `json:"page,omitempty"`
	PerPage    int    `json:"per_page,omitempty"`
}

// repositoryService implements the RepositoryService interface
type repositoryService struct {
	db           *gorm.DB
	gitService   git.GitService
	logger       *logrus.Logger
	repoBasePath string // Base path where repositories are stored
}

// NewRepositoryService creates a new repository service
func NewRepositoryService(db *gorm.DB, gitService git.GitService, logger *logrus.Logger, repoBasePath string) RepositoryService {
	return &repositoryService{
		db:           db,
		gitService:   gitService,
		logger:       logger,
		repoBasePath: repoBasePath,
	}
}

// Create creates a new repository
func (s *repositoryService) Create(ctx context.Context, req CreateRepositoryRequest) (*models.Repository, error) {
	s.logger.WithFields(logrus.Fields{
		"owner_id":   req.OwnerID,
		"owner_type": req.OwnerType,
		"name":       req.Name,
		"visibility": req.Visibility,
	}).Info("Creating repository")

	// Validate request
	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	// Check if repository already exists
	var existing models.Repository
	err := s.db.Where("owner_id = ? AND owner_type = ? AND name = ?", req.OwnerID, req.OwnerType, req.Name).First(&existing).Error
	if err == nil {
		return nil, fmt.Errorf("repository %s already exists", req.Name)
	} else if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check existing repository: %w", err)
	}

	// Set defaults
	if req.DefaultBranch == "" {
		req.DefaultBranch = "main"
	}

	// Create repository model
	repo := &models.Repository{
		OwnerID:             req.OwnerID,
		OwnerType:           req.OwnerType,
		Name:                req.Name,
		Description:         req.Description,
		DefaultBranch:       req.DefaultBranch,
		Visibility:          req.Visibility,
		IsTemplate:          req.IsTemplate,
		HasIssues:           req.HasIssues,
		HasProjects:         req.HasProjects,
		HasWiki:             req.HasWiki,
		HasDownloads:        req.HasDownloads,
		AllowMergeCommit:    req.AllowMergeCommit,
		AllowSquashMerge:    req.AllowSquashMerge,
		AllowRebaseMerge:    req.AllowRebaseMerge,
		DeleteBranchOnMerge: req.DeleteBranchOnMerge,
	}

	// Create in database
	if err := s.db.Create(repo).Error; err != nil {
		return nil, fmt.Errorf("failed to create repository in database: %w", err)
	}

	// Initialize Git repository
	if err := s.InitializeGitRepository(ctx, repo.ID); err != nil {
		// Rollback database changes if Git initialization fails
		s.db.Delete(repo)
		return nil, fmt.Errorf("failed to initialize Git repository: %w", err)
	}

	// Auto-initialize with README if requested
	if req.AutoInit {
		if err := s.createInitialCommit(ctx, repo); err != nil {
			s.logger.WithError(err).Warn("Failed to create initial commit")
		}
	}

	return repo, nil
}

// Get retrieves a repository by owner and name
func (s *repositoryService) Get(ctx context.Context, owner, name string) (*models.Repository, error) {
	// First, resolve the owner name to owner ID and type
	var ownerID uuid.UUID
	var ownerType models.OwnerType
	var ownerEntity *models.OwnerEntity

	// Try to find a user with this username
	var user models.User
	err := s.db.Where("username = ?", owner).First(&user).Error
	if err == nil {
		ownerID = user.ID
		ownerType = models.OwnerTypeUser
		ownerEntity = &models.OwnerEntity{
			ID:        user.ID,
			Username:  user.Username,
			Name:      user.FullName,
			AvatarURL: user.AvatarURL,
			Type:      models.OwnerTypeUser,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		}
	} else if err == gorm.ErrRecordNotFound {
		// Try to find an organization with this name
		var org models.Organization
		err = s.db.Where("name = ?", owner).First(&org).Error
		if err == nil {
			ownerID = org.ID
			ownerType = models.OwnerTypeOrganization
			ownerEntity = &models.OwnerEntity{
				ID:        org.ID,
				Username:  org.Name,
				Name:      org.DisplayName,
				AvatarURL: org.AvatarURL,
				Type:      models.OwnerTypeOrganization,
				CreatedAt: org.CreatedAt,
				UpdatedAt: org.UpdatedAt,
			}
		} else if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("repository not found")
		} else {
			return nil, fmt.Errorf("failed to find organization: %w", err)
		}
	} else {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Now find the repository with the resolved owner ID
	var repo models.Repository
	err = s.db.Where("owner_id = ? AND owner_type = ? AND name = ?", ownerID, ownerType, name).First(&repo).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("repository not found")
		}
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}

	// Populate the owner relationship
	repo.Owner = ownerEntity

	return &repo, nil
}

// GetByID retrieves a repository by ID
func (s *repositoryService) GetByID(ctx context.Context, id uuid.UUID) (*models.Repository, error) {
	var repo models.Repository
	err := s.db.Where("id = ?", id).First(&repo).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("repository not found")
		}
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}

	// Load the owner information
	var ownerEntity *models.OwnerEntity
	if repo.OwnerType == models.OwnerTypeUser {
		var user models.User
		err = s.db.Where("id = ?", repo.OwnerID).First(&user).Error
		if err == nil {
			ownerEntity = &models.OwnerEntity{
				ID:        user.ID,
				Username:  user.Username,
				Name:      user.FullName,
				AvatarURL: user.AvatarURL,
				Type:      models.OwnerTypeUser,
				CreatedAt: user.CreatedAt,
				UpdatedAt: user.UpdatedAt,
			}
		}
	} else if repo.OwnerType == models.OwnerTypeOrganization {
		var org models.Organization
		err = s.db.Where("id = ?", repo.OwnerID).First(&org).Error
		if err == nil {
			ownerEntity = &models.OwnerEntity{
				ID:        org.ID,
				Username:  org.Name,
				Name:      org.DisplayName,
				AvatarURL: org.AvatarURL,
				Type:      models.OwnerTypeOrganization,
				CreatedAt: org.CreatedAt,
				UpdatedAt: org.UpdatedAt,
			}
		}
	}

	// Populate the owner relationship
	repo.Owner = ownerEntity

	return &repo, nil
}

// Update updates a repository
func (s *repositoryService) Update(ctx context.Context, id uuid.UUID, req UpdateRepositoryRequest) (*models.Repository, error) {
	repo, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.DefaultBranch != nil {
		updates["default_branch"] = *req.DefaultBranch
	}
	if req.Visibility != nil {
		updates["visibility"] = *req.Visibility
	}
	if req.IsTemplate != nil {
		updates["is_template"] = *req.IsTemplate
	}
	if req.HasIssues != nil {
		updates["has_issues"] = *req.HasIssues
	}
	if req.HasProjects != nil {
		updates["has_projects"] = *req.HasProjects
	}
	if req.HasWiki != nil {
		updates["has_wiki"] = *req.HasWiki
	}
	if req.HasDownloads != nil {
		updates["has_downloads"] = *req.HasDownloads
	}
	if req.AllowMergeCommit != nil {
		updates["allow_merge_commit"] = *req.AllowMergeCommit
	}
	if req.AllowSquashMerge != nil {
		updates["allow_squash_merge"] = *req.AllowSquashMerge
	}
	if req.AllowRebaseMerge != nil {
		updates["allow_rebase_merge"] = *req.AllowRebaseMerge
	}
	if req.DeleteBranchOnMerge != nil {
		updates["delete_branch_on_merge"] = *req.DeleteBranchOnMerge
	}

	if len(updates) > 0 {
		updates["updated_at"] = time.Now()
		if err := s.db.Model(repo).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("failed to update repository: %w", err)
		}
	}

	return repo, nil
}

// Delete deletes a repository
func (s *repositoryService) Delete(ctx context.Context, id uuid.UUID) error {
	repo, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Delete Git repository from filesystem
	repoPath, err := s.GetRepositoryPath(ctx, id)
	if err == nil {
		if err := s.gitService.DeleteRepository(ctx, repoPath); err != nil {
			s.logger.WithError(err).Warn("Failed to delete Git repository from filesystem")
		}
	}

	// Delete from database (soft delete)
	if err := s.db.Delete(repo).Error; err != nil {
		return fmt.Errorf("failed to delete repository: %w", err)
	}

	return nil
}

// List lists repositories with filters
func (s *repositoryService) List(ctx context.Context, filters RepositoryFilters) ([]*models.Repository, int64, error) {
	query := s.db.Model(&models.Repository{})

	// Apply filters
	if filters.OwnerID != nil {
		query = query.Where("owner_id = ?", *filters.OwnerID)
	}
	if filters.OwnerType != nil {
		query = query.Where("owner_type = ?", *filters.OwnerType)
	}
	if filters.Visibility != nil {
		query = query.Where("visibility = ?", *filters.Visibility)
	}
	if filters.IsTemplate != nil {
		query = query.Where("is_template = ?", *filters.IsTemplate)
	}
	if filters.IsArchived != nil {
		query = query.Where("is_archived = ?", *filters.IsArchived)
	}
	if filters.IsFork != nil {
		query = query.Where("is_fork = ?", *filters.IsFork)
	}
	if filters.Search != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ?", "%"+filters.Search+"%", "%"+filters.Search+"%")
	}

	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count repositories: %w", err)
	}

	// Apply sorting
	orderBy := "created_at DESC"
	if filters.Sort != "" {
		direction := "DESC"
		if filters.Direction == "asc" {
			direction = "ASC"
		}

		switch filters.Sort {
		case "name":
			orderBy = fmt.Sprintf("name %s", direction)
		case "created":
			orderBy = fmt.Sprintf("created_at %s", direction)
		case "updated":
			orderBy = fmt.Sprintf("updated_at %s", direction)
		case "pushed":
			orderBy = fmt.Sprintf("pushed_at %s", direction)
		case "stars":
			orderBy = fmt.Sprintf("stars_count %s", direction)
		case "forks":
			orderBy = fmt.Sprintf("forks_count %s", direction)
		}
	}
	query = query.Order(orderBy)

	// Apply pagination
	if filters.PerPage <= 0 {
		filters.PerPage = 30
	}
	if filters.Page < 0 {
		filters.Page = 0
	}

	offset := filters.Page * filters.PerPage
	query = query.Offset(offset).Limit(filters.PerPage)

	// Execute query
	var repositories []*models.Repository
	if err := query.Find(&repositories).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list repositories: %w", err)
	}

	return repositories, total, nil
}

// InitializeGitRepository initializes the Git repository on filesystem
func (s *repositoryService) InitializeGitRepository(ctx context.Context, repoID uuid.UUID) error {
	repoPath, err := s.GetRepositoryPath(ctx, repoID)
	if err != nil {
		return err
	}

	// Create parent directories if they don't exist
	if err := os.MkdirAll(filepath.Dir(repoPath), 0755); err != nil {
		return fmt.Errorf("failed to create repository directory: %w", err)
	}

	// Check if repository already exists
	if _, err := os.Stat(repoPath); err == nil {
		s.logger.WithField("path", repoPath).Info("Repository already exists on filesystem")
		return nil
	}

	// Initialize bare repository
	if err := s.gitService.InitRepository(ctx, repoPath, true); err != nil {
		return fmt.Errorf("failed to initialize Git repository: %w", err)
	}

	// Set up Git hooks and configuration for the repository
	if err := s.setupRepositoryHooks(ctx, repoPath); err != nil {
		s.logger.WithError(err).Warn("Failed to setup repository hooks")
		// Don't fail the entire operation for hook setup failure
	}

	s.logger.WithField("path", repoPath).Info("Git repository initialized successfully")
	return nil
}

// GetRepositoryPath returns the filesystem path for a repository
func (s *repositoryService) GetRepositoryPath(ctx context.Context, repoID uuid.UUID) (string, error) {
	repo, err := s.GetByID(ctx, repoID)
	if err != nil {
		return "", err
	}

	// Generate path: /repos/{owner_type}/{owner_id}/{repo_name}.git
	return filepath.Join(s.repoBasePath, string(repo.OwnerType), repo.OwnerID.String(), repo.Name+".git"), nil
}

// Helper methods

func (s *repositoryService) validateCreateRequest(req CreateRepositoryRequest) error {
	if req.OwnerID == uuid.Nil {
		return fmt.Errorf("owner_id is required")
	}
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}
	if req.OwnerType == "" {
		return fmt.Errorf("owner_type is required")
	}
	if req.Visibility == "" {
		return fmt.Errorf("visibility is required")
	}
	return nil
}

func (s *repositoryService) validateForkRequest(ctx context.Context, sourceRepo *models.Repository, req ForkRequest) error {
	if req.OwnerID == uuid.Nil {
		return fmt.Errorf("owner_id is required")
	}
	if req.OwnerType == "" {
		return fmt.Errorf("owner_type is required")
	}

	// Check if trying to fork to the same owner
	if sourceRepo.OwnerID == req.OwnerID && sourceRepo.OwnerType == req.OwnerType {
		return fmt.Errorf("cannot fork repository to the same owner")
	}

	// Validate owner exists
	if req.OwnerType == models.OwnerTypeUser {
		var user models.User
		if err := s.db.Where("id = ?", req.OwnerID).First(&user).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("user not found")
			}
			return fmt.Errorf("failed to validate user: %w", err)
		}
	} else if req.OwnerType == models.OwnerTypeOrganization {
		var org models.Organization
		if err := s.db.Where("id = ?", req.OwnerID).First(&org).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("organization not found")
			}
			return fmt.Errorf("failed to validate organization: %w", err)
		}
	}

	return nil
}

func (s *repositoryService) validateTransferRequest(ctx context.Context, repo *models.Repository, req TransferRequest) error {
	if req.NewOwnerID == uuid.Nil {
		return fmt.Errorf("new_owner_id is required")
	}
	if req.NewOwnerType == "" {
		return fmt.Errorf("new_owner_type is required")
	}

	// Check if trying to transfer to the same owner
	if repo.OwnerID == req.NewOwnerID && repo.OwnerType == req.NewOwnerType {
		return fmt.Errorf("cannot transfer repository to the same owner")
	}

	// Validate new owner exists
	if req.NewOwnerType == models.OwnerTypeUser {
		var user models.User
		if err := s.db.Where("id = ?", req.NewOwnerID).First(&user).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("new owner user not found")
			}
			return fmt.Errorf("failed to validate new owner user: %w", err)
		}
	} else if req.NewOwnerType == models.OwnerTypeOrganization {
		var org models.Organization
		if err := s.db.Where("id = ?", req.NewOwnerID).First(&org).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("new owner organization not found")
			}
			return fmt.Errorf("failed to validate new owner organization: %w", err)
		}
	}

	return nil
}

func (s *repositoryService) moveRepository(ctx context.Context, oldPath, newPath string) error {
	// Create parent directories for the new path
	if err := os.MkdirAll(filepath.Dir(newPath), 0755); err != nil {
		return fmt.Errorf("failed to create new repository directory: %w", err)
	}

	// Move the repository directory
	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("failed to move repository directory: %w", err)
	}

	// Clean up empty parent directories in the old path
	s.cleanupEmptyDirectories(filepath.Dir(oldPath))

	s.logger.WithFields(logrus.Fields{
		"old_path": oldPath,
		"new_path": newPath,
	}).Info("Repository moved successfully")

	return nil
}

func (s *repositoryService) cleanupEmptyDirectories(dirPath string) {
	// Don't clean up the base repository path
	if dirPath == s.repoBasePath || dirPath == filepath.Dir(s.repoBasePath) {
		return
	}

	// Check if directory is empty
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return
	}

	if len(entries) == 0 {
		// Remove empty directory
		if err := os.Remove(dirPath); err == nil {
			s.logger.WithField("path", dirPath).Debug("Removed empty directory")
			// Recursively clean up parent directories
			s.cleanupEmptyDirectories(filepath.Dir(dirPath))
		}
	}
}

func (s *repositoryService) cloneRepository(ctx context.Context, sourceRepo, forkRepo *models.Repository) error {
	sourceRepoPath, err := s.GetRepositoryPath(ctx, sourceRepo.ID)
	if err != nil {
		return fmt.Errorf("failed to get source repository path: %w", err)
	}

	forkRepoPath, err := s.GetRepositoryPath(ctx, forkRepo.ID)
	if err != nil {
		return fmt.Errorf("failed to get fork repository path: %w", err)
	}

	// Create parent directories for the fork
	if err := os.MkdirAll(filepath.Dir(forkRepoPath), 0755); err != nil {
		return fmt.Errorf("failed to create fork repository directory: %w", err)
	}

	// Clone the source repository to the fork location
	cloneOptions := git.CloneOptions{
		Bare:   true,
		Mirror: false,
	}
	if err := s.gitService.CloneRepository(ctx, sourceRepoPath, forkRepoPath, cloneOptions); err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"source_path": sourceRepoPath,
		"fork_path":   forkRepoPath,
	}).Info("Repository cloned successfully")

	return nil
}

func (s *repositoryService) createInitialCommit(ctx context.Context, repo *models.Repository) error {
	repoPath, err := s.GetRepositoryPath(ctx, repo.ID)
	if err != nil {
		return err
	}

	// Create README content
	readmeContent := fmt.Sprintf("# %s\n\n%s\n", repo.Name, repo.Description)

	// Create initial commit with README
	createReq := git.CreateFileRequest{
		Path:    "README.md",
		Content: readmeContent,
		Message: "Initial commit",
		Branch:  repo.DefaultBranch,
		Author: git.CommitAuthor{
			Name:  "System",
			Email: "noreply@hub.local",
			Date:  time.Now(),
		},
	}

	_, err = s.gitService.CreateFile(ctx, repoPath, createReq)
	return err
}

// Placeholder implementations for methods that need more complex logic

func (s *repositoryService) Fork(ctx context.Context, id uuid.UUID, req ForkRequest) (*models.Repository, error) {
	s.logger.WithFields(logrus.Fields{
		"source_repo_id": id,
		"new_owner_id":   req.OwnerID,
		"new_owner_type": req.OwnerType,
		"fork_name":      req.Name,
	}).Info("Forking repository")

	// Get the source repository
	sourceRepo, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get source repository: %w", err)
	}

	// Validate fork request
	if err := s.validateForkRequest(ctx, sourceRepo, req); err != nil {
		return nil, err
	}

	// Set fork name (default to source repo name if not provided)
	forkName := req.Name
	if forkName == "" {
		forkName = sourceRepo.Name
	}

	// Check if fork already exists
	var existing models.Repository
	err = s.db.Where("owner_id = ? AND owner_type = ? AND name = ?", req.OwnerID, req.OwnerType, forkName).First(&existing).Error
	if err == nil {
		return nil, fmt.Errorf("repository %s already exists for this owner", forkName)
	} else if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check existing repository: %w", err)
	}

	// Create fork repository in database
	fork := &models.Repository{
		OwnerID:             req.OwnerID,
		OwnerType:           req.OwnerType,
		Name:                forkName,
		Description:         sourceRepo.Description,
		DefaultBranch:       sourceRepo.DefaultBranch,
		Visibility:          sourceRepo.Visibility,
		IsFork:              true,
		ParentID:            &sourceRepo.ID,
		IsTemplate:          false, // Forks cannot be templates
		HasIssues:           sourceRepo.HasIssues,
		HasProjects:         sourceRepo.HasProjects,
		HasWiki:             sourceRepo.HasWiki,
		HasDownloads:        sourceRepo.HasDownloads,
		AllowMergeCommit:    sourceRepo.AllowMergeCommit,
		AllowSquashMerge:    sourceRepo.AllowSquashMerge,
		AllowRebaseMerge:    sourceRepo.AllowRebaseMerge,
		DeleteBranchOnMerge: sourceRepo.DeleteBranchOnMerge,
	}

	// Create fork in database
	if err := s.db.Create(fork).Error; err != nil {
		return nil, fmt.Errorf("failed to create fork in database: %w", err)
	}

	// Clone the Git repository
	if err := s.cloneRepository(ctx, sourceRepo, fork); err != nil {
		// Rollback database changes if Git cloning fails
		s.db.Delete(fork)
		return nil, fmt.Errorf("failed to clone Git repository: %w", err)
	}

	// Update source repository fork count
	if err := s.db.Model(sourceRepo).Update("forks_count", gorm.Expr("forks_count + 1")).Error; err != nil {
		s.logger.WithError(err).Warn("Failed to update source repository fork count")
	}

	s.logger.WithFields(logrus.Fields{
		"fork_id":   fork.ID,
		"fork_name": fork.Name,
	}).Info("Repository forked successfully")

	return fork, nil
}

func (s *repositoryService) Transfer(ctx context.Context, id uuid.UUID, req TransferRequest) error {
	s.logger.WithFields(logrus.Fields{
		"repo_id":          id,
		"new_owner_id":     req.NewOwnerID,
		"new_owner_type":   req.NewOwnerType,
	}).Info("Transferring repository")

	// Get the repository
	repo, err := s.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get repository: %w", err)
	}

	// Validate transfer request
	if err := s.validateTransferRequest(ctx, repo, req); err != nil {
		return err
	}

	// Check if a repository with the same name already exists for the new owner
	var existing models.Repository
	err = s.db.Where("owner_id = ? AND owner_type = ? AND name = ?", req.NewOwnerID, req.NewOwnerType, repo.Name).First(&existing).Error
	if err == nil {
		return fmt.Errorf("repository %s already exists for the new owner", repo.Name)
	} else if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to check existing repository: %w", err)
	}

	oldOwnerID := repo.OwnerID
	oldOwnerType := repo.OwnerType

	// Get the old and new repository paths
	oldRepoPath, err := s.GetRepositoryPath(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get old repository path: %w", err)
	}

	// Update repository ownership in database
	updates := map[string]interface{}{
		"owner_id":   req.NewOwnerID,
		"owner_type": req.NewOwnerType,
		"updated_at": time.Now(),
	}

	if err := s.db.Model(repo).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update repository ownership: %w", err)
	}

	// Get the new repository path after ownership change
	newRepoPath, err := s.GetRepositoryPath(ctx, id)
	if err != nil {
		// Rollback ownership change
		s.db.Model(repo).Updates(map[string]interface{}{
			"owner_id":   oldOwnerID,
			"owner_type": oldOwnerType,
		})
		return fmt.Errorf("failed to get new repository path: %w", err)
	}

	// Move the Git repository on filesystem if paths are different
	if oldRepoPath != newRepoPath {
		if err := s.moveRepository(ctx, oldRepoPath, newRepoPath); err != nil {
			// Rollback ownership change
			s.db.Model(repo).Updates(map[string]interface{}{
				"owner_id":   oldOwnerID,
				"owner_type": oldOwnerType,
			})
			return fmt.Errorf("failed to move repository on filesystem: %w", err)
		}
	}

	s.logger.WithFields(logrus.Fields{
		"repo_id":      id,
		"old_path":     oldRepoPath,
		"new_path":     newRepoPath,
	}).Info("Repository transferred successfully")

	return nil
}

func (s *repositoryService) Archive(ctx context.Context, id uuid.UUID) error {
	return s.db.Model(&models.Repository{}).Where("id = ?", id).Update("is_archived", true).Error
}

func (s *repositoryService) Unarchive(ctx context.Context, id uuid.UUID) error {
	return s.db.Model(&models.Repository{}).Where("id = ?", id).Update("is_archived", false).Error
}

func (s *repositoryService) SyncCommits(ctx context.Context, repoID uuid.UUID) error {
	// TODO: Implement commit synchronization from Git to database
	return fmt.Errorf("SyncCommits not yet implemented")
}

// UpdateRepositoryStats updates repository statistics including language detection
func (s *repositoryService) UpdateRepositoryStats(ctx context.Context, repoID uuid.UUID) error {
	s.logger.WithField("repo_id", repoID).Info("Updating repository statistics")

	repoPath, err := s.GetRepositoryPath(ctx, repoID)
	if err != nil {
		return fmt.Errorf("failed to get repository path: %w", err)
	}

	// Get repository statistics from Git
	stats, err := s.gitService.GetRepositoryStats(ctx, repoPath)
	if err != nil {
		return fmt.Errorf("failed to get repository stats: %w", err)
	}

	// Update or create repository statistics record
	var repoStats models.RepositoryStatistics
	err = s.db.Where("repository_id = ?", repoID).First(&repoStats).Error
	if err == gorm.ErrRecordNotFound {
		// Create new statistics record
		repoStats = models.RepositoryStatistics{
			RepositoryID: repoID,
		}
	} else if err != nil {
		return fmt.Errorf("failed to query repository statistics: %w", err)
	}

	// Update statistics
	repoStats.SizeBytes = stats.Size
	repoStats.CommitCount = stats.CommitCount
	repoStats.BranchCount = stats.BranchCount
	repoStats.TagCount = stats.TagCount
	repoStats.Contributors = stats.Contributors
	repoStats.LastActivity = &stats.LastActivity
	repoStats.LanguageCount = len(stats.Languages)

	// Determine primary language (language with most bytes)
	var primaryLanguage string
	var maxBytes int64
	for lang, langStats := range stats.Languages {
		if langStats.Bytes > maxBytes {
			maxBytes = langStats.Bytes
			primaryLanguage = lang
		}
	}
	repoStats.PrimaryLanguage = primaryLanguage

	// Save or update repository statistics
	if err := s.db.Save(&repoStats).Error; err != nil {
		return fmt.Errorf("failed to save repository statistics: %w", err)
	}

	// Update repository languages
	if err := s.updateRepositoryLanguages(ctx, repoID, stats.Languages); err != nil {
		return fmt.Errorf("failed to update repository languages: %w", err)
	}

	// Update repository size in the main repository record
	if err := s.db.Model(&models.Repository{}).Where("id = ?", repoID).Update("size_kb", stats.Size/1024).Error; err != nil {
		s.logger.WithError(err).Warn("Failed to update repository size")
	}

	s.logger.WithField("repo_id", repoID).Info("Repository statistics updated successfully")
	return nil
}

// GetLanguages returns the programming languages used in a repository
func (s *repositoryService) GetLanguages(ctx context.Context, repoID uuid.UUID) (map[string]git.LanguageStats, error) {
	var languages []models.RepositoryLanguage
	if err := s.db.Where("repository_id = ?", repoID).Find(&languages).Error; err != nil {
		return nil, fmt.Errorf("failed to get repository languages: %w", err)
	}

	result := make(map[string]git.LanguageStats)
	for _, lang := range languages {
		result[lang.Language] = git.LanguageStats{
			Bytes:      lang.Bytes,
			Percentage: lang.Percentage,
		}
	}

	return result, nil
}

// GetRepositoryStatistics returns comprehensive statistics for a repository
func (s *repositoryService) GetRepositoryStatistics(ctx context.Context, repoID uuid.UUID) (*git.RepositoryStats, error) {
	var repoStats models.RepositoryStatistics
	err := s.db.Where("repository_id = ?", repoID).First(&repoStats).Error
	if err == gorm.ErrRecordNotFound {
		// Statistics not yet calculated, trigger update
		if err := s.UpdateRepositoryStats(ctx, repoID); err != nil {
			return nil, fmt.Errorf("failed to update repository statistics: %w", err)
		}
		// Retry getting statistics
		err = s.db.Where("repository_id = ?", repoID).First(&repoStats).Error
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get repository statistics: %w", err)
	}

	// Get languages
	languages, err := s.GetLanguages(ctx, repoID)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository languages: %w", err)
	}

	// Convert to git.RepositoryStats format
	result := &git.RepositoryStats{
		Size:         repoStats.SizeBytes,
		CommitCount:  repoStats.CommitCount,
		BranchCount:  repoStats.BranchCount,
		TagCount:     repoStats.TagCount,
		Contributors: repoStats.Contributors,
		Languages:    languages,
	}

	if repoStats.LastActivity != nil {
		result.LastActivity = *repoStats.LastActivity
	}

	return result, nil
}

// updateRepositoryLanguages updates the language statistics for a repository
func (s *repositoryService) updateRepositoryLanguages(ctx context.Context, repoID uuid.UUID, languages map[string]git.LanguageStats) error {
	// Delete existing language records
	if err := s.db.Where("repository_id = ?", repoID).Delete(&models.RepositoryLanguage{}).Error; err != nil {
		return fmt.Errorf("failed to delete existing language records: %w", err)
	}

	// Insert new language records
	for language, stats := range languages {
		langRecord := models.RepositoryLanguage{
			RepositoryID: repoID,
			Language:     language,
			Bytes:        stats.Bytes,
			Percentage:   stats.Percentage,
		}

		if err := s.db.Create(&langRecord).Error; err != nil {
			return fmt.Errorf("failed to create language record for %s: %w", language, err)
		}
	}

	return nil
}

// Git hooks management methods

// CreateGitHook creates a new Git hook for a repository
func (s *repositoryService) CreateGitHook(ctx context.Context, repoID uuid.UUID, req CreateGitHookRequest) (*models.GitHook, error) {
	s.logger.WithFields(logrus.Fields{
		"repo_id":   repoID,
		"hook_type": req.HookType,
	}).Info("Creating Git hook")

	// Validate repository exists
	if _, err := s.GetByID(ctx, repoID); err != nil {
		return nil, fmt.Errorf("repository not found: %w", err)
	}

	// Validate hook type
	validHookTypes := []string{"pre-receive", "post-receive", "update", "pre-push", "post-update"}
	isValid := false
	for _, validType := range validHookTypes {
		if req.HookType == validType {
			isValid = true
			break
		}
	}
	if !isValid {
		return nil, fmt.Errorf("invalid hook type: %s", req.HookType)
	}

	// Set default language if not provided
	if req.Language == "" {
		req.Language = "bash"
	}

	// Create hook record
	hook := &models.GitHook{
		RepositoryID: repoID,
		HookType:     req.HookType,
		Script:       req.Script,
		Language:     req.Language,
		IsEnabled:    req.IsEnabled,
		Order:        req.Order,
	}

	if err := s.db.Create(hook).Error; err != nil {
		return nil, fmt.Errorf("failed to create Git hook: %w", err)
	}

	// Install the hook on the filesystem
	if err := s.installGitHook(ctx, repoID, hook); err != nil {
		// Rollback database changes if filesystem installation fails
		s.db.Delete(hook)
		return nil, fmt.Errorf("failed to install Git hook: %w", err)
	}

	s.logger.WithField("hook_id", hook.ID).Info("Git hook created successfully")
	return hook, nil
}

// UpdateGitHook updates an existing Git hook
func (s *repositoryService) UpdateGitHook(ctx context.Context, hookID uuid.UUID, req UpdateGitHookRequest) (*models.GitHook, error) {
	var hook models.GitHook
	if err := s.db.Where("id = ?", hookID).First(&hook).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("Git hook not found")
		}
		return nil, fmt.Errorf("failed to get Git hook: %w", err)
	}

	// Update fields if provided
	updates := make(map[string]interface{})
	if req.Script != nil {
		updates["script"] = *req.Script
	}
	if req.Language != nil {
		updates["language"] = *req.Language
	}
	if req.IsEnabled != nil {
		updates["is_enabled"] = *req.IsEnabled
	}
	if req.Order != nil {
		updates["order"] = *req.Order
	}

	if len(updates) > 0 {
		updates["updated_at"] = time.Now()
		if err := s.db.Model(&hook).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("failed to update Git hook: %w", err)
		}
	}

	// Reinstall the hook on the filesystem
	if err := s.installGitHook(ctx, hook.RepositoryID, &hook); err != nil {
		return nil, fmt.Errorf("failed to reinstall Git hook: %w", err)
	}

	return &hook, nil
}

// DeleteGitHook deletes a Git hook
func (s *repositoryService) DeleteGitHook(ctx context.Context, hookID uuid.UUID) error {
	var hook models.GitHook
	if err := s.db.Where("id = ?", hookID).First(&hook).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("Git hook not found")
		}
		return fmt.Errorf("failed to get Git hook: %w", err)
	}

	// Remove hook from filesystem
	if err := s.uninstallGitHook(ctx, hook.RepositoryID, &hook); err != nil {
		s.logger.WithError(err).Warn("Failed to uninstall Git hook from filesystem")
	}

	// Delete from database
	if err := s.db.Delete(&hook).Error; err != nil {
		return fmt.Errorf("failed to delete Git hook: %w", err)
	}

	return nil
}

// GetGitHooks returns all Git hooks for a repository
func (s *repositoryService) GetGitHooks(ctx context.Context, repoID uuid.UUID) ([]*models.GitHook, error) {
	var hooks []*models.GitHook
	if err := s.db.Where("repository_id = ?", repoID).Order("hook_type, \"order\"").Find(&hooks).Error; err != nil {
		return nil, fmt.Errorf("failed to get Git hooks: %w", err)
	}
	return hooks, nil
}

// installGitHook installs a Git hook on the filesystem
func (s *repositoryService) installGitHook(ctx context.Context, repoID uuid.UUID, hook *models.GitHook) error {
	repoPath, err := s.GetRepositoryPath(ctx, repoID)
	if err != nil {
		return err
	}

	hooksDir := filepath.Join(repoPath, "hooks")
	hookPath := filepath.Join(hooksDir, hook.HookType)

	// Create hooks directory if it doesn't exist
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return fmt.Errorf("failed to create hooks directory: %w", err)
	}

	// Write hook script
	hookContent := fmt.Sprintf("#!/bin/%s\n%s\n", hook.Language, hook.Script)
	if err := os.WriteFile(hookPath, []byte(hookContent), 0755); err != nil {
		return fmt.Errorf("failed to write hook script: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"hook_path": hookPath,
		"hook_type": hook.HookType,
	}).Debug("Git hook installed")

	return nil
}

// uninstallGitHook removes a Git hook from the filesystem
func (s *repositoryService) uninstallGitHook(ctx context.Context, repoID uuid.UUID, hook *models.GitHook) error {
	repoPath, err := s.GetRepositoryPath(ctx, repoID)
	if err != nil {
		return err
	}

	hookPath := filepath.Join(repoPath, "hooks", hook.HookType)
	if err := os.Remove(hookPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove hook script: %w", err)
	}

	return nil
}

// Repository template methods

// CreateTemplate creates a new repository template
func (s *repositoryService) CreateTemplate(ctx context.Context, repoID uuid.UUID, req CreateTemplateRequest) (*models.RepositoryTemplate, error) {
	s.logger.WithFields(logrus.Fields{
		"repo_id": repoID,
		"name":    req.Name,
	}).Info("Creating repository template")

	// Validate repository exists and is a template
	repo, err := s.GetByID(ctx, repoID)
	if err != nil {
		return nil, fmt.Errorf("repository not found: %w", err)
	}

	if !repo.IsTemplate {
		return nil, fmt.Errorf("repository is not marked as a template")
	}

	// Convert tags to JSON
	tagsJSON := "[]"
	if len(req.Tags) > 0 {
		tagsJSON = fmt.Sprintf(`["%s"]`, strings.Join(req.Tags, `","`))
	}

	// Create template record
	template := &models.RepositoryTemplate{
		RepositoryID: repoID,
		Name:         req.Name,
		Description:  req.Description,
		Category:     req.Category,
		Tags:         tagsJSON,
		IsFeatured:   req.IsFeatured,
		IsPublic:     req.IsPublic,
	}

	if err := s.db.Create(template).Error; err != nil {
		return nil, fmt.Errorf("failed to create repository template: %w", err)
	}

	s.logger.WithField("template_id", template.ID).Info("Repository template created successfully")
	return template, nil
}

// GetTemplates returns repository templates based on filters
func (s *repositoryService) GetTemplates(ctx context.Context, filters TemplateFilters) ([]*models.RepositoryTemplate, error) {
	query := s.db.Model(&models.RepositoryTemplate{})

	// Apply filters
	if filters.Category != "" {
		query = query.Where("category = ?", filters.Category)
	}
	if filters.IsFeatured != nil {
		query = query.Where("is_featured = ?", *filters.IsFeatured)
	}
	if filters.IsPublic != nil {
		query = query.Where("is_public = ?", *filters.IsPublic)
	}
	if filters.Search != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ?", "%"+filters.Search+"%", "%"+filters.Search+"%")
	}

	// Apply sorting
	orderBy := "created_at DESC"
	if filters.Sort != "" {
		direction := "DESC"
		if filters.Direction == "asc" {
			direction = "ASC"
		}

		switch filters.Sort {
		case "name":
			orderBy = fmt.Sprintf("name %s", direction)
		case "created":
			orderBy = fmt.Sprintf("created_at %s", direction)
		case "usage_count":
			orderBy = fmt.Sprintf("usage_count %s", direction)
		}
	}
	query = query.Order(orderBy)

	// Apply pagination
	if filters.PerPage <= 0 {
		filters.PerPage = 30
	}
	if filters.Page < 0 {
		filters.Page = 0
	}

	offset := filters.Page * filters.PerPage
	query = query.Offset(offset).Limit(filters.PerPage)

	// Execute query
	var templates []*models.RepositoryTemplate
	if err := query.Find(&templates).Error; err != nil {
		return nil, fmt.Errorf("failed to list repository templates: %w", err)
	}

	return templates, nil
}

// UseTemplate creates a new repository from a template
func (s *repositoryService) UseTemplate(ctx context.Context, templateID uuid.UUID, req CreateRepositoryRequest) (*models.Repository, error) {
	s.logger.WithFields(logrus.Fields{
		"template_id": templateID,
		"repo_name":   req.Name,
	}).Info("Creating repository from template")

	// Get template
	var template models.RepositoryTemplate
	err := s.db.Where("id = ?", templateID).First(&template).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("template not found")
		}
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	// Get template repository to verify it exists
	_, err = s.GetByID(ctx, template.RepositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get template repository: %w", err)
	}

	// Create new repository
	newRepo, err := s.Create(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}

	// Clone template repository content to new repository
	templateRepoPath, err := s.GetRepositoryPath(ctx, template.RepositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get template repository path: %w", err)
	}

	newRepoPath, err := s.GetRepositoryPath(ctx, newRepo.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get new repository path: %w", err)
	}

	// Clone template repository content
	cloneOptions := git.CloneOptions{
		Bare:   true,
		Mirror: false,
	}
	if err := s.gitService.CloneRepository(ctx, templateRepoPath, newRepoPath, cloneOptions); err != nil {
		// Clean up the created repository on failure
		s.Delete(ctx, newRepo.ID)
		return nil, fmt.Errorf("failed to clone template repository: %w", err)
	}

	// Increment template usage count
	if err := s.db.Model(&template).Update("usage_count", gorm.Expr("usage_count + 1")).Error; err != nil {
		s.logger.WithError(err).Warn("Failed to update template usage count")
	}

	s.logger.WithFields(logrus.Fields{
		"new_repo_id":  newRepo.ID,
		"template_id":  templateID,
	}).Info("Repository created from template successfully")

	return newRepo, nil
}

// setupRepositoryHooks sets up Git hooks for the repository
func (s *repositoryService) setupRepositoryHooks(ctx context.Context, repoPath string) error {
	hooksDir := filepath.Join(repoPath, "hooks")

	// Create hooks directory if it doesn't exist
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return fmt.Errorf("failed to create hooks directory: %w", err)
	}

	// TODO: Add pre-receive and post-receive hooks for:
	// - Authentication/authorization checks
	// - Commit validation
	// - Webhook notifications
	// - Database synchronization

	s.logger.WithField("hooks_dir", hooksDir).Debug("Repository hooks directory created")
	return nil
}

// CleanupRepositoryStorage removes orphaned repository directories
func (s *repositoryService) CleanupRepositoryStorage(ctx context.Context) error {
	s.logger.Info("Starting repository storage cleanup")

	// Walk through the repository base path and check for orphaned directories
	err := filepath.Walk(s.repoBasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip if not a .git directory
		if !info.IsDir() || !strings.HasSuffix(path, ".git") {
			return nil
		}

		// Extract repository info from path
		relPath, err := filepath.Rel(s.repoBasePath, path)
		if err != nil {
			return nil // Skip this path
		}

		parts := strings.Split(relPath, string(filepath.Separator))
		if len(parts) != 3 { // owner_type/owner_id/repo_name.git
			return nil // Skip malformed paths
		}

		ownerType := parts[0]
		ownerIDStr := parts[1]
		repoName := strings.TrimSuffix(parts[2], ".git")

		// Parse owner ID
		ownerID, err := uuid.Parse(ownerIDStr)
		if err != nil {
			return nil // Skip invalid UUIDs
		}

		// Check if repository exists in database
		var count int64
		err = s.db.Model(&models.Repository{}).
			Where("owner_id = ? AND owner_type = ? AND name = ?", ownerID, ownerType, repoName).
			Count(&count).Error
		if err != nil {
			s.logger.WithError(err).Warn("Failed to check repository existence")
			return nil
		}

		// Remove orphaned directory
		if count == 0 {
			s.logger.WithField("path", path).Info("Removing orphaned repository directory")
			if err := os.RemoveAll(path); err != nil {
				s.logger.WithError(err).Error("Failed to remove orphaned directory")
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to cleanup repository storage: %w", err)
	}

	s.logger.Info("Repository storage cleanup completed")
	return nil
}

// GetRepositorySize calculates the size of a repository on disk
func (s *repositoryService) GetRepositorySize(ctx context.Context, repoID uuid.UUID) (int64, error) {
	repoPath, err := s.GetRepositoryPath(ctx, repoID)
	if err != nil {
		return 0, err
	}

	var size int64
	err = filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("failed to calculate repository size: %w", err)
	}

	return size, nil
}
