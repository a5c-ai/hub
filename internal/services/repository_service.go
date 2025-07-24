package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
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
}

// CreateRepositoryRequest represents a request to create a repository
type CreateRepositoryRequest struct {
	OwnerID               uuid.UUID           `json:"owner_id"`
	OwnerType             models.OwnerType    `json:"owner_type"`
	Name                  string              `json:"name"`
	Description           string              `json:"description,omitempty"`
	DefaultBranch         string              `json:"default_branch,omitempty"`
	Visibility            models.Visibility   `json:"visibility"`
	IsTemplate            bool                `json:"is_template,omitempty"`
	HasIssues             bool                `json:"has_issues"`
	HasProjects           bool                `json:"has_projects"`
	HasWiki               bool                `json:"has_wiki"`
	HasDownloads          bool                `json:"has_downloads"`
	AllowMergeCommit      bool                `json:"allow_merge_commit"`
	AllowSquashMerge      bool                `json:"allow_squash_merge"`
	AllowRebaseMerge      bool                `json:"allow_rebase_merge"`
	DeleteBranchOnMerge   bool                `json:"delete_branch_on_merge"`
	AutoInit              bool                `json:"auto_init"` // Initialize with README
}

// UpdateRepositoryRequest represents a request to update a repository
type UpdateRepositoryRequest struct {
	Name                  *string             `json:"name,omitempty"`
	Description           *string             `json:"description,omitempty"`
	DefaultBranch         *string             `json:"default_branch,omitempty"`
	Visibility            *models.Visibility  `json:"visibility,omitempty"`
	IsTemplate            *bool               `json:"is_template,omitempty"`
	HasIssues             *bool               `json:"has_issues,omitempty"`
	HasProjects           *bool               `json:"has_projects,omitempty"`
	HasWiki               *bool               `json:"has_wiki,omitempty"`
	HasDownloads          *bool               `json:"has_downloads,omitempty"`
	AllowMergeCommit      *bool               `json:"allow_merge_commit,omitempty"`
	AllowSquashMerge      *bool               `json:"allow_squash_merge,omitempty"`
	AllowRebaseMerge      *bool               `json:"allow_rebase_merge,omitempty"`
	DeleteBranchOnMerge   *bool               `json:"delete_branch_on_merge,omitempty"`
}

// ForkRequest represents a request to fork a repository
type ForkRequest struct {
	Name         string           `json:"name,omitempty"` // New name for the fork
	OwnerID      uuid.UUID        `json:"owner_id"`       // New owner
	OwnerType    models.OwnerType `json:"owner_type"`
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
	Sort       string             `json:"sort,omitempty"`   // name, created, updated, pushed, stars, forks
	Direction  string             `json:"direction,omitempty"` // asc, desc
	Page       int                `json:"page,omitempty"`
	PerPage    int                `json:"per_page,omitempty"`
}

// repositoryService implements the RepositoryService interface
type repositoryService struct {
	db         *gorm.DB
	gitService git.GitService
	logger     *logrus.Logger
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
		"owner_id":    req.OwnerID,
		"owner_type":  req.OwnerType,
		"name":        req.Name,
		"visibility":  req.Visibility,
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
		OwnerID:               req.OwnerID,
		OwnerType:             req.OwnerType,
		Name:                  req.Name,
		Description:           req.Description,
		DefaultBranch:         req.DefaultBranch,
		Visibility:            req.Visibility,
		IsTemplate:            req.IsTemplate,
		HasIssues:             req.HasIssues,
		HasProjects:           req.HasProjects,
		HasWiki:               req.HasWiki,
		HasDownloads:          req.HasDownloads,
		AllowMergeCommit:      req.AllowMergeCommit,
		AllowSquashMerge:      req.AllowSquashMerge,
		AllowRebaseMerge:      req.AllowRebaseMerge,
		DeleteBranchOnMerge:   req.DeleteBranchOnMerge,
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
	var repo models.Repository
	err := s.db.Where("owner_id = ? AND name = ?", owner, name).First(&repo).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("repository not found")
		}
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}

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

	// Initialize bare repository
	return s.gitService.InitRepository(ctx, repoPath, true)
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
	// TODO: Implement repository forking
	return nil, fmt.Errorf("Fork not yet implemented")
}

func (s *repositoryService) Transfer(ctx context.Context, id uuid.UUID, req TransferRequest) error {
	// TODO: Implement repository transfer
	return fmt.Errorf("Transfer not yet implemented")
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