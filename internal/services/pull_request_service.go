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

type PullRequestService interface {
	Create(ctx context.Context, repoID uuid.UUID, userID uuid.UUID, req CreatePullRequestRequest) (*models.PullRequest, error)
	Get(ctx context.Context, owner, repo string, number int) (*models.PullRequest, error)
	List(ctx context.Context, repoID uuid.UUID, filter PullRequestFilter) ([]*models.PullRequest, error)
	Update(ctx context.Context, id uuid.UUID, req UpdatePullRequestRequest) (*models.PullRequest, error)
	Close(ctx context.Context, id uuid.UUID) error
	Merge(ctx context.Context, id uuid.UUID, req MergePullRequestRequest) error
}

type pullRequestService struct {
	db           *gorm.DB
	gitService   git.GitService
	repoService  RepositoryService
	logger       *logrus.Logger
	repoBasePath string
}

type CreatePullRequestRequest struct {
	Title               string     `json:"title" binding:"required"`
	Body                string     `json:"body"`
	Head                string     `json:"head" binding:"required"`
	Base                string     `json:"base" binding:"required"`
	HeadRepositoryID    *uuid.UUID `json:"head_repository_id"`
	Draft               bool       `json:"draft"`
	MaintainerCanModify bool       `json:"maintainer_can_modify"`
}

type UpdatePullRequestRequest struct {
	Title *string `json:"title,omitempty"`
	Body  *string `json:"body,omitempty"`
	State *string `json:"state,omitempty"`
}

type MergePullRequestRequest struct {
	CommitTitle   string `json:"commit_title,omitempty"`
	CommitMessage string `json:"commit_message,omitempty"`
	MergeMethod   string `json:"merge_method,omitempty"` // merge, squash, rebase
}

type PullRequestFilter struct {
	State    *string    `json:"state,omitempty"`
	Head     *string    `json:"head,omitempty"`
	Base     *string    `json:"base,omitempty"`
	UserID   *uuid.UUID `json:"user_id,omitempty"`
	Page     int        `json:"page,omitempty"`
	PageSize int        `json:"page_size,omitempty"`
}

func NewPullRequestService(db *gorm.DB, gitService git.GitService, repoService RepositoryService, logger *logrus.Logger, repoBasePath string) PullRequestService {
	return &pullRequestService{
		db:           db,
		gitService:   gitService,
		repoService:  repoService,
		logger:       logger,
		repoBasePath: repoBasePath,
	}
}

func (s *pullRequestService) Create(ctx context.Context, repoID uuid.UUID, userID uuid.UUID, req CreatePullRequestRequest) (*models.PullRequest, error) {
	// Get repository
	var repo models.Repository
	if err := s.db.First(&repo, "id = ?", repoID).Error; err != nil {
		return nil, fmt.Errorf("repository not found: %w", err)
	}

	// Get the next PR number
	nextNumber, err := s.getNextPRNumber(repoID)
	if err != nil {
		return nil, err
	}

	// Set head repository ID
	headRepoID := repoID
	if req.HeadRepositoryID != nil {
		headRepoID = *req.HeadRepositoryID
	}

	// Create the pull request
	pr := models.PullRequest{
		RepositoryID:     repoID,
		Number:           nextNumber,
		Title:            req.Title,
		Body:             req.Body,
		UserID:           &userID,
		HeadRepositoryID: &headRepoID,
		HeadBranch:       req.Head,
		BaseBranch:       req.Base,
		State:            models.PullRequestStateOpen,
		Draft:            req.Draft,
	}

	if err := s.db.Create(&pr).Error; err != nil {
		return nil, err
	}

	return &pr, nil
}

func (s *pullRequestService) Get(ctx context.Context, owner, repo string, number int) (*models.PullRequest, error) {
	var pr models.PullRequest
	err := s.db.Preload("Repository").Preload("User").
		Joins("JOIN repositories ON repositories.id = pull_requests.repository_id").
		Joins("JOIN users ON users.id = repositories.owner_id").
		Where("users.username = ? AND repositories.name = ? AND pull_requests.number = ?", owner, repo, number).
		First(&pr).Error

	if err != nil {
		return nil, err
	}

	return &pr, nil
}

func (s *pullRequestService) List(ctx context.Context, repoID uuid.UUID, filter PullRequestFilter) ([]*models.PullRequest, error) {
	query := s.db.Where("repository_id = ?", repoID)

	if filter.State != nil {
		query = query.Where("state = ?", *filter.State)
	}

	if filter.UserID != nil {
		query = query.Where("user_id = ?", *filter.UserID)
	}

	// Pagination
	pageSize := 30
	if filter.PageSize > 0 {
		pageSize = filter.PageSize
	}

	offset := 0
	if filter.Page > 1 {
		offset = (filter.Page - 1) * pageSize
	}

	var prs []*models.PullRequest
	err := query.Preload("User").Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&prs).Error
	return prs, err
}

func (s *pullRequestService) Update(ctx context.Context, id uuid.UUID, req UpdatePullRequestRequest) (*models.PullRequest, error) {
	var pr models.PullRequest
	if err := s.db.First(&pr, "id = ?", id).Error; err != nil {
		return nil, err
	}

	updates := make(map[string]interface{})
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Body != nil {
		updates["body"] = *req.Body
	}
	if req.State != nil {
		updates["state"] = *req.State
	}

	if len(updates) > 0 {
		if err := s.db.Model(&pr).Updates(updates).Error; err != nil {
			return nil, err
		}
	}

	return &pr, nil
}

func (s *pullRequestService) Close(ctx context.Context, id uuid.UUID) error {
	return s.db.Model(&models.PullRequest{}).Where("id = ?", id).
		Update("state", models.PullRequestStateClosed).Error
}

func (s *pullRequestService) Merge(ctx context.Context, id uuid.UUID, req MergePullRequestRequest) error {
	// Simplified merge - just update state
	return s.db.Model(&models.PullRequest{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"state":  models.PullRequestStateMerged,
			"merged": true,
		}).Error
}

func (s *pullRequestService) getNextPRNumber(repoID uuid.UUID) (int, error) {
	var lastNumber int
	err := s.db.Model(&models.PullRequest{}).
		Where("repository_id = ?", repoID).
		Order("number DESC").
		Limit(1).
		Pluck("number", &lastNumber).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return 0, err
	}

	return lastNumber + 1, nil
}
