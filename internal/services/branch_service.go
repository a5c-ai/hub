package services

import (
	"context"
	"fmt"

	"github.com/a5c-ai/hub/internal/git"
	"github.com/a5c-ai/hub/internal/models"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// BranchService provides branch management operations
type BranchService interface {
	List(ctx context.Context, repoID uuid.UUID) ([]*models.Branch, error)
	Get(ctx context.Context, repoID uuid.UUID, branchName string) (*models.Branch, error)
	Create(ctx context.Context, repoID uuid.UUID, req CreateBranchRequest) (*models.Branch, error)
	Delete(ctx context.Context, repoID uuid.UUID, branchName string) error
	SetDefault(ctx context.Context, repoID uuid.UUID, branchName string) error
	
	// Branch protection
	GetProtectionRule(ctx context.Context, repoID uuid.UUID, pattern string) (*models.BranchProtectionRule, error)
	CreateProtectionRule(ctx context.Context, repoID uuid.UUID, req CreateBranchProtectionRequest) (*models.BranchProtectionRule, error)
	UpdateProtectionRule(ctx context.Context, ruleID uuid.UUID, req UpdateBranchProtectionRequest) (*models.BranchProtectionRule, error)
	DeleteProtectionRule(ctx context.Context, ruleID uuid.UUID) error
	ListProtectionRules(ctx context.Context, repoID uuid.UUID) ([]*models.BranchProtectionRule, error)
	
	// Sync operations
	SyncBranchesFromGit(ctx context.Context, repoID uuid.UUID) error
}

// CreateBranchRequest represents a request to create a branch
type CreateBranchRequest struct {
	Name    string `json:"name"`
	FromRef string `json:"from_ref,omitempty"` // Branch or commit to create from
}

// CreateBranchProtectionRequest represents a request to create a branch protection rule
type CreateBranchProtectionRequest struct {
	Pattern                        string                       `json:"pattern"`
	RequiredStatusChecks           *RequiredStatusChecks        `json:"required_status_checks,omitempty"`
	EnforceAdmins                  bool                         `json:"enforce_admins"`
	RequiredPullRequestReviews     *RequiredPullRequestReviews  `json:"required_pull_request_reviews,omitempty"`
	Restrictions                   *BranchRestrictions          `json:"restrictions,omitempty"`
}

// UpdateBranchProtectionRequest represents a request to update a branch protection rule
type UpdateBranchProtectionRequest struct {
	Pattern                        *string                      `json:"pattern,omitempty"`
	RequiredStatusChecks           *RequiredStatusChecks        `json:"required_status_checks,omitempty"`
	EnforceAdmins                  *bool                        `json:"enforce_admins,omitempty"`
	RequiredPullRequestReviews     *RequiredPullRequestReviews  `json:"required_pull_request_reviews,omitempty"`
	Restrictions                   *BranchRestrictions          `json:"restrictions,omitempty"`
}

// RequiredStatusChecks represents required status checks for branch protection
type RequiredStatusChecks struct {
	Strict   bool     `json:"strict"`
	Contexts []string `json:"contexts"`
}

// RequiredPullRequestReviews represents required pull request reviews for branch protection
type RequiredPullRequestReviews struct {
	RequiredApprovingReviewCount   int    `json:"required_approving_review_count"`
	DismissStaleReviews           bool   `json:"dismiss_stale_reviews"`
	RequireCodeOwnerReviews       bool   `json:"require_code_owner_reviews"`
	RestrictPushesToCodeOwners    bool   `json:"restrict_pushes_to_code_owners"`
}

// BranchRestrictions represents push restrictions for branch protection
type BranchRestrictions struct {
	Users []string `json:"users"`
	Teams []string `json:"teams"`
}

// branchService implements the BranchService interface
type branchService struct {
	db               *gorm.DB
	gitService       git.GitService
	repositoryService RepositoryService
	logger           *logrus.Logger
}

// NewBranchService creates a new branch service
func NewBranchService(db *gorm.DB, gitService git.GitService, repositoryService RepositoryService, logger *logrus.Logger) BranchService {
	return &branchService{
		db:               db,
		gitService:       gitService,
		repositoryService: repositoryService,
		logger:           logger,
	}
}

// List retrieves all branches for a repository
func (s *branchService) List(ctx context.Context, repoID uuid.UUID) ([]*models.Branch, error) {
	s.logger.WithField("repo_id", repoID).Info("Listing branches")

	// First sync branches from Git
	if err := s.SyncBranchesFromGit(ctx, repoID); err != nil {
		s.logger.WithError(err).Warn("Failed to sync branches from Git")
	}

	var branches []*models.Branch
	err := s.db.Where("repository_id = ?", repoID).Order("name ASC").Find(&branches).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}

	return branches, nil
}

// Get retrieves a single branch by repository ID and name
func (s *branchService) Get(ctx context.Context, repoID uuid.UUID, branchName string) (*models.Branch, error) {
	var branch models.Branch
	err := s.db.Where("repository_id = ? AND name = ?", repoID, branchName).First(&branch).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Try to sync from Git and retry
			if syncErr := s.SyncBranchesFromGit(ctx, repoID); syncErr == nil {
				err = s.db.Where("repository_id = ? AND name = ?", repoID, branchName).First(&branch).Error
			}
		}
		
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, fmt.Errorf("branch not found")
			}
			return nil, fmt.Errorf("failed to get branch: %w", err)
		}
	}

	return &branch, nil
}

// Create creates a new branch
func (s *branchService) Create(ctx context.Context, repoID uuid.UUID, req CreateBranchRequest) (*models.Branch, error) {
	s.logger.WithFields(logrus.Fields{
		"repo_id":   repoID,
		"name":      req.Name,
		"from_ref":  req.FromRef,
	}).Info("Creating branch")

	// Validate request
	if req.Name == "" {
		return nil, fmt.Errorf("branch name is required")
	}

	// Get repository path
	repoPath, err := s.repositoryService.GetRepositoryPath(ctx, repoID)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository path: %w", err)
	}

	// Create branch in Git
	if err := s.gitService.CreateBranch(ctx, repoPath, req.Name, req.FromRef); err != nil {
		return nil, fmt.Errorf("failed to create branch in Git: %w", err)
	}

	// Get the branch from Git to get the SHA
	gitBranch, err := s.gitService.GetBranch(ctx, repoPath, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get created branch from Git: %w", err)
	}

	// Create branch in database
	branch := &models.Branch{
		RepositoryID: repoID,
		Name:         req.Name,
		SHA:          gitBranch.SHA,
		IsProtected:  false,
		IsDefault:    false,
	}

	if err := s.db.Create(branch).Error; err != nil {
		// If database creation fails, try to clean up Git branch
		s.gitService.DeleteBranch(ctx, repoPath, req.Name)
		return nil, fmt.Errorf("failed to create branch in database: %w", err)
	}

	return branch, nil
}

// Delete deletes a branch
func (s *branchService) Delete(ctx context.Context, repoID uuid.UUID, branchName string) error {
	s.logger.WithFields(logrus.Fields{
		"repo_id": repoID,
		"name":    branchName,
	}).Info("Deleting branch")

	// Check if branch exists and is not the default branch
	branch, err := s.Get(ctx, repoID, branchName)
	if err != nil {
		return err
	}

	if branch.IsDefault {
		return fmt.Errorf("cannot delete default branch")
	}

	// Get repository path
	repoPath, err := s.repositoryService.GetRepositoryPath(ctx, repoID)
	if err != nil {
		return fmt.Errorf("failed to get repository path: %w", err)
	}

	// Delete branch from Git
	if err := s.gitService.DeleteBranch(ctx, repoPath, branchName); err != nil {
		return fmt.Errorf("failed to delete branch from Git: %w", err)
	}

	// Delete branch from database
	if err := s.db.Where("repository_id = ? AND name = ?", repoID, branchName).Delete(&models.Branch{}).Error; err != nil {
		s.logger.WithError(err).Warn("Failed to delete branch from database")
		return fmt.Errorf("failed to delete branch from database: %w", err)
	}

	return nil
}

// SetDefault sets a branch as the default branch
func (s *branchService) SetDefault(ctx context.Context, repoID uuid.UUID, branchName string) error {
	s.logger.WithFields(logrus.Fields{
		"repo_id": repoID,
		"name":    branchName,
	}).Info("Setting default branch")

	// Verify branch exists
	_, err := s.Get(ctx, repoID, branchName)
	if err != nil {
		return err
	}

	// Update repository default branch
	updateReq := UpdateRepositoryRequest{
		DefaultBranch: &branchName,
	}

	_, err = s.repositoryService.Update(ctx, repoID, updateReq)
	if err != nil {
		return fmt.Errorf("failed to update repository default branch: %w", err)
	}

	// Update branch flags in database
	tx := s.db.Begin()

	// Unset all default flags for this repository
	if err := tx.Model(&models.Branch{}).Where("repository_id = ?", repoID).Update("is_default", false).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to unset default branches: %w", err)
	}

	// Set the new default branch
	if err := tx.Model(&models.Branch{}).Where("repository_id = ? AND name = ?", repoID, branchName).Update("is_default", true).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to set default branch: %w", err)
	}

	tx.Commit()
	return nil
}

// SyncBranchesFromGit synchronizes branches from Git repository to database
func (s *branchService) SyncBranchesFromGit(ctx context.Context, repoID uuid.UUID) error {
	// Get repository path
	repoPath, err := s.repositoryService.GetRepositoryPath(ctx, repoID)
	if err != nil {
		return fmt.Errorf("failed to get repository path: %w", err)
	}

	// Get branches from Git
	gitBranches, err := s.gitService.GetBranches(ctx, repoPath)
	if err != nil {
		return fmt.Errorf("failed to get branches from Git: %w", err)
	}

	// Get existing branches from database
	var dbBranches []*models.Branch
	if err := s.db.Where("repository_id = ?", repoID).Find(&dbBranches).Error; err != nil {
		return fmt.Errorf("failed to get branches from database: %w", err)
	}

	// Create maps for easier lookup
	dbBranchMap := make(map[string]*models.Branch)
	for _, branch := range dbBranches {
		dbBranchMap[branch.Name] = branch
	}

	gitBranchMap := make(map[string]*git.Branch)
	for _, branch := range gitBranches {
		gitBranchMap[branch.Name] = branch
	}

	// Sync branches
	for _, gitBranch := range gitBranches {
		if dbBranch, exists := dbBranchMap[gitBranch.Name]; exists {
			// Update existing branch if SHA changed
			if dbBranch.SHA != gitBranch.SHA {
				dbBranch.SHA = gitBranch.SHA
				if err := s.db.Save(dbBranch).Error; err != nil {
					s.logger.WithError(err).WithField("branch", gitBranch.Name).Warn("Failed to update branch")
				}
			}
		} else {
			// Create new branch
			newBranch := &models.Branch{
				RepositoryID: repoID,
				Name:         gitBranch.Name,
				SHA:          gitBranch.SHA,
				IsProtected:  false,
				IsDefault:    gitBranch.IsDefault,
			}
			if err := s.db.Create(newBranch).Error; err != nil {
				s.logger.WithError(err).WithField("branch", gitBranch.Name).Warn("Failed to create branch")
			}
		}
	}

	// Remove branches that no longer exist in Git
	for _, dbBranch := range dbBranches {
		if _, exists := gitBranchMap[dbBranch.Name]; !exists {
			if err := s.db.Delete(dbBranch).Error; err != nil {
				s.logger.WithError(err).WithField("branch", dbBranch.Name).Warn("Failed to delete branch")
			}
		}
	}

	return nil
}

// Placeholder implementations for branch protection methods

func (s *branchService) GetProtectionRule(ctx context.Context, repoID uuid.UUID, pattern string) (*models.BranchProtectionRule, error) {
	var rule models.BranchProtectionRule
	err := s.db.Where("repository_id = ? AND pattern = ?", repoID, pattern).First(&rule).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("protection rule not found")
		}
		return nil, fmt.Errorf("failed to get protection rule: %w", err)
	}
	return &rule, nil
}

func (s *branchService) CreateProtectionRule(ctx context.Context, repoID uuid.UUID, req CreateBranchProtectionRequest) (*models.BranchProtectionRule, error) {
	// TODO: Implement branch protection rule creation
	return nil, fmt.Errorf("CreateProtectionRule not yet implemented")
}

func (s *branchService) UpdateProtectionRule(ctx context.Context, ruleID uuid.UUID, req UpdateBranchProtectionRequest) (*models.BranchProtectionRule, error) {
	// TODO: Implement branch protection rule update
	return nil, fmt.Errorf("UpdateProtectionRule not yet implemented")
}

func (s *branchService) DeleteProtectionRule(ctx context.Context, ruleID uuid.UUID) error {
	// TODO: Implement branch protection rule deletion
	return fmt.Errorf("DeleteProtectionRule not yet implemented")
}

func (s *branchService) ListProtectionRules(ctx context.Context, repoID uuid.UUID) ([]*models.BranchProtectionRule, error) {
	var rules []*models.BranchProtectionRule
	err := s.db.Where("repository_id = ?", repoID).Find(&rules).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list protection rules: %w", err)
	}
	return rules, nil
}