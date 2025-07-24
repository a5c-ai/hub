package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/a5c-ai/hub/internal/models"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// IssueService provides issue management operations
type IssueService interface {
	// Issue CRUD operations
	Create(ctx context.Context, req CreateIssueRequest) (*models.Issue, error)
	Get(ctx context.Context, repoOwner, repoName string, number int) (*models.Issue, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Issue, error)
	Update(ctx context.Context, id uuid.UUID, req UpdateIssueRequest) (*models.Issue, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, repoOwner, repoName string, filters IssueFilters) ([]*models.Issue, int64, error)
	
	// Issue operations
	Close(ctx context.Context, id uuid.UUID, reason string) (*models.Issue, error)
	Reopen(ctx context.Context, id uuid.UUID) (*models.Issue, error)
	Lock(ctx context.Context, id uuid.UUID, reason string) (*models.Issue, error)
	Unlock(ctx context.Context, id uuid.UUID) (*models.Issue, error)
	
	// Assignment operations
	Assign(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.Issue, error)
	Unassign(ctx context.Context, id uuid.UUID) (*models.Issue, error)
	
	// Label operations
	AddLabels(ctx context.Context, id uuid.UUID, labelIDs []uuid.UUID) (*models.Issue, error)
	RemoveLabels(ctx context.Context, id uuid.UUID, labelIDs []uuid.UUID) (*models.Issue, error)
	SetLabels(ctx context.Context, id uuid.UUID, labelIDs []uuid.UUID) (*models.Issue, error)
	
	// Milestone operations
	SetMilestone(ctx context.Context, id uuid.UUID, milestoneID *uuid.UUID) (*models.Issue, error)
	
	// Search operations
	Search(ctx context.Context, repoOwner, repoName string, query string, filters IssueFilters) ([]*models.Issue, int64, error)
}

// CreateIssueRequest represents a request to create an issue
type CreateIssueRequest struct {
	RepositoryID uuid.UUID      `json:"repository_id" validate:"required"`
	Title        string         `json:"title" validate:"required,max=255"`
	Body         string         `json:"body"`
	UserID       *uuid.UUID     `json:"user_id"`
	AssigneeID   *uuid.UUID     `json:"assignee_id"`
	MilestoneID  *uuid.UUID     `json:"milestone_id"`
	LabelIDs     []uuid.UUID    `json:"label_ids"`
}

// UpdateIssueRequest represents a request to update an issue
type UpdateIssueRequest struct {
	Title        *string        `json:"title,omitempty" validate:"omitempty,max=255"`
	Body         *string        `json:"body,omitempty"`
	State        *models.IssueState `json:"state,omitempty"`
	StateReason  *string        `json:"state_reason,omitempty"`
	AssigneeID   *uuid.UUID     `json:"assignee_id,omitempty"`
	MilestoneID  *uuid.UUID     `json:"milestone_id,omitempty"`
	LabelIDs     []uuid.UUID    `json:"label_ids,omitempty"`
}

// IssueFilters represents filters for listing issues
type IssueFilters struct {
	State       *models.IssueState `json:"state,omitempty"`
	AssigneeID  *uuid.UUID         `json:"assignee_id,omitempty"`
	CreatorID   *uuid.UUID         `json:"creator_id,omitempty"`
	MilestoneID *uuid.UUID         `json:"milestone_id,omitempty"`
	LabelIDs    []uuid.UUID        `json:"label_ids,omitempty"`
	Since       *time.Time         `json:"since,omitempty"`
	Sort        string             `json:"sort,omitempty"` // created, updated, comments
	Direction   string             `json:"direction,omitempty"` // asc, desc
	Page        int                `json:"page,omitempty"`
	PerPage     int                `json:"per_page,omitempty"`
}

type issueService struct {
	db     *gorm.DB
	logger *logrus.Logger
}

// NewIssueService creates a new issue service
func NewIssueService(db *gorm.DB, logger *logrus.Logger) IssueService {
	return &issueService{
		db:     db,
		logger: logger,
	}
}

func (s *issueService) Create(ctx context.Context, req CreateIssueRequest) (*models.Issue, error) {
	logger := s.logger.WithField("method", "Create")
	
	// Get next issue number for this repository
	var lastIssue models.Issue
	err := s.db.WithContext(ctx).
		Where("repository_id = ?", req.RepositoryID).
		Order("number DESC").
		First(&lastIssue).Error
	
	nextNumber := 1
	if err == nil {
		nextNumber = lastIssue.Number + 1
	} else if err != gorm.ErrRecordNotFound {
		logger.WithError(err).Error("Failed to get last issue number")
		return nil, fmt.Errorf("failed to get next issue number: %w", err)
	}
	
	// Create the issue
	issue := &models.Issue{
		RepositoryID: req.RepositoryID,
		Number:       nextNumber,
		Title:        req.Title,
		Body:         req.Body,
		UserID:       req.UserID,
		AssigneeID:   req.AssigneeID,
		MilestoneID:  req.MilestoneID,
		State:        models.IssueStateOpen,
	}
	
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create the issue
		if err := tx.Create(issue).Error; err != nil {
			return fmt.Errorf("failed to create issue: %w", err)
		}
		
		// Add labels if provided
		if len(req.LabelIDs) > 0 {
			if err := s.addLabelsInTx(tx, issue.ID, req.LabelIDs); err != nil {
				return fmt.Errorf("failed to add labels: %w", err)
			}
		}
		
		return nil
	})
	
	if err != nil {
		logger.WithError(err).Error("Failed to create issue")
		return nil, err
	}
	
	// Load the complete issue with relationships
	return s.loadIssueWithRelations(ctx, issue.ID)
}

func (s *issueService) Get(ctx context.Context, repoOwner, repoName string, number int) (*models.Issue, error) {
	logger := s.logger.WithField("method", "Get")
	
	var issue models.Issue
	err := s.db.WithContext(ctx).
		Joins("JOIN repositories ON repositories.id = issues.repository_id").
		Joins("JOIN users owner ON owner.id = repositories.owner_id").
		Where("owner.username = ? AND repositories.name = ? AND issues.number = ?", repoOwner, repoName, number).
		Preload("Repository").
		Preload("User").
		Preload("Assignee").
		Preload("Milestone").
		Preload("Labels").
		Preload("Comments.User").
		Preload("PullRequest").
		First(&issue).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("issue not found")
		}
		logger.WithError(err).Error("Failed to get issue")
		return nil, fmt.Errorf("failed to get issue: %w", err)
	}
	
	return &issue, nil
}

func (s *issueService) GetByID(ctx context.Context, id uuid.UUID) (*models.Issue, error) {
	return s.loadIssueWithRelations(ctx, id)
}

func (s *issueService) Update(ctx context.Context, id uuid.UUID, req UpdateIssueRequest) (*models.Issue, error) {
	logger := s.logger.WithField("method", "Update")
	
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Get the current issue
		var issue models.Issue
		if err := tx.First(&issue, "id = ?", id).Error; err != nil {
			return fmt.Errorf("issue not found: %w", err)
		}
		
		// Update fields
		updates := make(map[string]interface{})
		
		if req.Title != nil {
			updates["title"] = *req.Title
		}
		if req.Body != nil {
			updates["body"] = *req.Body
		}
		if req.State != nil {
			updates["state"] = *req.State
			if *req.State == models.IssueStateClosed {
				now := time.Now()
				updates["closed_at"] = &now
			} else {
				updates["closed_at"] = nil
			}
		}
		if req.StateReason != nil {
			updates["state_reason"] = *req.StateReason
		}
		if req.AssigneeID != nil {
			updates["assignee_id"] = *req.AssigneeID
		}
		if req.MilestoneID != nil {
			updates["milestone_id"] = *req.MilestoneID
		}
		
		if len(updates) > 0 {
			if err := tx.Model(&issue).Updates(updates).Error; err != nil {
				return fmt.Errorf("failed to update issue: %w", err)
			}
		}
		
		// Update labels if provided
		if req.LabelIDs != nil {
			if err := s.setLabelsInTx(tx, id, req.LabelIDs); err != nil {
				return fmt.Errorf("failed to update labels: %w", err)
			}
		}
		
		return nil
	})
	
	if err != nil {
		logger.WithError(err).Error("Failed to update issue")
		return nil, err
	}
	
	return s.loadIssueWithRelations(ctx, id)
}

func (s *issueService) Delete(ctx context.Context, id uuid.UUID) error {
	logger := s.logger.WithField("method", "Delete")
	
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete label associations
		if err := tx.Where("issue_id = ?", id).Delete(&models.IssueLabel{}).Error; err != nil {
			return fmt.Errorf("failed to delete issue labels: %w", err)
		}
		
		// Delete comments
		if err := tx.Where("issue_id = ?", id).Delete(&models.Comment{}).Error; err != nil {
			return fmt.Errorf("failed to delete issue comments: %w", err)
		}
		
		// Delete pull request (if exists)
		if err := tx.Where("issue_id = ?", id).Delete(&models.PullRequest{}).Error; err != nil {
			return fmt.Errorf("failed to delete pull request: %w", err)
		}
		
		// Delete the issue
		if err := tx.Delete(&models.Issue{}, "id = ?", id).Error; err != nil {
			return fmt.Errorf("failed to delete issue: %w", err)
		}
		
		return nil
	})
	
	if err != nil {
		logger.WithError(err).Error("Failed to delete issue")
		return err
	}
	
	return nil
}

func (s *issueService) List(ctx context.Context, repoOwner, repoName string, filters IssueFilters) ([]*models.Issue, int64, error) {
	logger := s.logger.WithField("method", "List")
	
	query := s.db.WithContext(ctx).
		Joins("JOIN repositories ON repositories.id = issues.repository_id").
		Joins("JOIN users owner ON owner.id = repositories.owner_id").
		Where("owner.username = ? AND repositories.name = ?", repoOwner, repoName)
	
	// Apply filters
	query = s.applyFilters(query, filters)
	
	// Count total records
	var total int64
	if err := query.Model(&models.Issue{}).Count(&total).Error; err != nil {
		logger.WithError(err).Error("Failed to count issues")
		return nil, 0, fmt.Errorf("failed to count issues: %w", err)
	}
	
	// Apply pagination and sorting
	query = s.applySortingAndPagination(query, filters)
	
	var issues []*models.Issue
	err := query.
		Preload("Repository").
		Preload("User").
		Preload("Assignee").
		Preload("Milestone").
		Preload("Labels").
		Find(&issues).Error
	
	if err != nil {
		logger.WithError(err).Error("Failed to list issues")
		return nil, 0, fmt.Errorf("failed to list issues: %w", err)
	}
	
	return issues, total, nil
}

func (s *issueService) Close(ctx context.Context, id uuid.UUID, reason string) (*models.Issue, error) {
	state := models.IssueStateClosed
	
	return s.Update(ctx, id, UpdateIssueRequest{
		State:       &state,
		StateReason: &reason,
	})
}

func (s *issueService) Reopen(ctx context.Context, id uuid.UUID) (*models.Issue, error) {
	state := models.IssueStateOpen
	
	return s.Update(ctx, id, UpdateIssueRequest{
		State: &state,
	})
}

func (s *issueService) Lock(ctx context.Context, id uuid.UUID, reason string) (*models.Issue, error) {
	logger := s.logger.WithField("method", "Lock")
	
	err := s.db.WithContext(ctx).Model(&models.Issue{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"locked":       true,
			"state_reason": reason,
		}).Error
	
	if err != nil {
		logger.WithError(err).Error("Failed to lock issue")
		return nil, fmt.Errorf("failed to lock issue: %w", err)
	}
	
	return s.loadIssueWithRelations(ctx, id)
}

func (s *issueService) Unlock(ctx context.Context, id uuid.UUID) (*models.Issue, error) {
	logger := s.logger.WithField("method", "Unlock")
	
	err := s.db.WithContext(ctx).Model(&models.Issue{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"locked":       false,
			"state_reason": "",
		}).Error
	
	if err != nil {
		logger.WithError(err).Error("Failed to unlock issue")
		return nil, fmt.Errorf("failed to unlock issue: %w", err)
	}
	
	return s.loadIssueWithRelations(ctx, id)
}

func (s *issueService) Assign(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.Issue, error) {
	return s.Update(ctx, id, UpdateIssueRequest{
		AssigneeID: &userID,
	})
}

func (s *issueService) Unassign(ctx context.Context, id uuid.UUID) (*models.Issue, error) {
	logger := s.logger.WithField("method", "Unassign")
	
	err := s.db.WithContext(ctx).Model(&models.Issue{}).
		Where("id = ?", id).
		Update("assignee_id", nil).Error
	
	if err != nil {
		logger.WithError(err).Error("Failed to unassign issue")
		return nil, fmt.Errorf("failed to unassign issue: %w", err)
	}
	
	return s.loadIssueWithRelations(ctx, id)
}

func (s *issueService) AddLabels(ctx context.Context, id uuid.UUID, labelIDs []uuid.UUID) (*models.Issue, error) {
	logger := s.logger.WithField("method", "AddLabels")
	
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return s.addLabelsInTx(tx, id, labelIDs)
	})
	
	if err != nil {
		logger.WithError(err).Error("Failed to add labels")
		return nil, err
	}
	
	return s.loadIssueWithRelations(ctx, id)
}

func (s *issueService) RemoveLabels(ctx context.Context, id uuid.UUID, labelIDs []uuid.UUID) (*models.Issue, error) {
	logger := s.logger.WithField("method", "RemoveLabels")
	
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Where("issue_id = ? AND label_id IN ?", id, labelIDs).
			Delete(&models.IssueLabel{}).Error
	})
	
	if err != nil {
		logger.WithError(err).Error("Failed to remove labels")
		return nil, fmt.Errorf("failed to remove labels: %w", err)
	}
	
	return s.loadIssueWithRelations(ctx, id)
}

func (s *issueService) SetLabels(ctx context.Context, id uuid.UUID, labelIDs []uuid.UUID) (*models.Issue, error) {
	logger := s.logger.WithField("method", "SetLabels")
	
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return s.setLabelsInTx(tx, id, labelIDs)
	})
	
	if err != nil {
		logger.WithError(err).Error("Failed to set labels")
		return nil, err
	}
	
	return s.loadIssueWithRelations(ctx, id)
}

func (s *issueService) SetMilestone(ctx context.Context, id uuid.UUID, milestoneID *uuid.UUID) (*models.Issue, error) {
	return s.Update(ctx, id, UpdateIssueRequest{
		MilestoneID: milestoneID,
	})
}

func (s *issueService) Search(ctx context.Context, repoOwner, repoName string, query string, filters IssueFilters) ([]*models.Issue, int64, error) {
	logger := s.logger.WithField("method", "Search")
	
	dbQuery := s.db.WithContext(ctx).
		Joins("JOIN repositories ON repositories.id = issues.repository_id").
		Joins("JOIN users owner ON owner.id = repositories.owner_id").
		Where("owner.username = ? AND repositories.name = ?", repoOwner, repoName)
	
	// Add text search
	if query != "" {
		searchQuery := fmt.Sprintf("%%%s%%", strings.ToLower(query))
		dbQuery = dbQuery.Where("LOWER(issues.title) LIKE ? OR LOWER(issues.body) LIKE ?", searchQuery, searchQuery)
	}
	
	// Apply filters
	dbQuery = s.applyFilters(dbQuery, filters)
	
	// Count total records
	var total int64
	if err := dbQuery.Model(&models.Issue{}).Count(&total).Error; err != nil {
		logger.WithError(err).Error("Failed to count search results")
		return nil, 0, fmt.Errorf("failed to count search results: %w", err)
	}
	
	// Apply pagination and sorting
	dbQuery = s.applySortingAndPagination(dbQuery, filters)
	
	var issues []*models.Issue
	err := dbQuery.
		Preload("Repository").
		Preload("User").
		Preload("Assignee").
		Preload("Milestone").
		Preload("Labels").
		Find(&issues).Error
	
	if err != nil {
		logger.WithError(err).Error("Failed to search issues")
		return nil, 0, fmt.Errorf("failed to search issues: %w", err)
	}
	
	return issues, total, nil
}

// Helper methods

func (s *issueService) loadIssueWithRelations(ctx context.Context, id uuid.UUID) (*models.Issue, error) {
	var issue models.Issue
	err := s.db.WithContext(ctx).
		Preload("Repository").
		Preload("User").
		Preload("Assignee").
		Preload("Milestone").
		Preload("Labels").
		Preload("Comments.User").
		Preload("PullRequest").
		First(&issue, "id = ?", id).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("issue not found")
		}
		return nil, fmt.Errorf("failed to load issue: %w", err)
	}
	
	return &issue, nil
}

func (s *issueService) addLabelsInTx(tx *gorm.DB, issueID uuid.UUID, labelIDs []uuid.UUID) error {
	for _, labelID := range labelIDs {
		issueLabel := &models.IssueLabel{
			IssueID: issueID,
			LabelID: labelID,
		}
		
		// Use FirstOrCreate to avoid duplicates
		if err := tx.FirstOrCreate(issueLabel, issueLabel).Error; err != nil {
			return fmt.Errorf("failed to add label %s: %w", labelID, err)
		}
	}
	return nil
}

func (s *issueService) setLabelsInTx(tx *gorm.DB, issueID uuid.UUID, labelIDs []uuid.UUID) error {
	// Remove all existing labels
	if err := tx.Where("issue_id = ?", issueID).Delete(&models.IssueLabel{}).Error; err != nil {
		return fmt.Errorf("failed to remove existing labels: %w", err)
	}
	
	// Add new labels
	if len(labelIDs) > 0 {
		return s.addLabelsInTx(tx, issueID, labelIDs)
	}
	
	return nil
}

func (s *issueService) applyFilters(query *gorm.DB, filters IssueFilters) *gorm.DB {
	if filters.State != nil {
		query = query.Where("issues.state = ?", *filters.State)
	}
	
	if filters.AssigneeID != nil {
		query = query.Where("issues.assignee_id = ?", *filters.AssigneeID)
	}
	
	if filters.CreatorID != nil {
		query = query.Where("issues.user_id = ?", *filters.CreatorID)
	}
	
	if filters.MilestoneID != nil {
		query = query.Where("issues.milestone_id = ?", *filters.MilestoneID)
	}
	
	if len(filters.LabelIDs) > 0 {
		query = query.
			Joins("JOIN issue_labels ON issue_labels.issue_id = issues.id").
			Where("issue_labels.label_id IN ?", filters.LabelIDs).
			Group("issues.id")
	}
	
	if filters.Since != nil {
		query = query.Where("issues.created_at >= ?", *filters.Since)
	}
	
	return query
}

func (s *issueService) applySortingAndPagination(query *gorm.DB, filters IssueFilters) *gorm.DB {
	// Apply sorting
	sort := "created"
	if filters.Sort != "" {
		sort = filters.Sort
	}
	
	direction := "desc"
	if filters.Direction != "" {
		direction = filters.Direction
	}
	
	switch sort {
	case "created":
		query = query.Order(fmt.Sprintf("issues.created_at %s", direction))
	case "updated":
		query = query.Order(fmt.Sprintf("issues.updated_at %s", direction))
	case "comments":
		query = query.Order(fmt.Sprintf("issues.comments_count %s", direction))
	default:
		query = query.Order(fmt.Sprintf("issues.created_at %s", direction))
	}
	
	// Apply pagination
	page := 1
	if filters.Page > 0 {
		page = filters.Page
	}
	
	perPage := 30
	if filters.PerPage > 0 && filters.PerPage <= 100 {
		perPage = filters.PerPage
	}
	
	offset := (page - 1) * perPage
	query = query.Offset(offset).Limit(perPage)
	
	return query
}