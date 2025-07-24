package services

import (
	"context"
	"fmt"
	"time"

	"github.com/a5c-ai/hub/internal/models"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// MilestoneService provides milestone management operations
type MilestoneService interface {
	// Milestone CRUD operations
	Create(ctx context.Context, req CreateMilestoneRequest) (*models.Milestone, error)
	Get(ctx context.Context, repoOwner, repoName string, number int) (*models.Milestone, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Milestone, error)
	Update(ctx context.Context, id uuid.UUID, req UpdateMilestoneRequest) (*models.Milestone, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, repoOwner, repoName string, filters MilestoneFilters) ([]*models.Milestone, int64, error)
	
	// Milestone operations
	Close(ctx context.Context, id uuid.UUID) (*models.Milestone, error)
	Reopen(ctx context.Context, id uuid.UUID) (*models.Milestone, error)
}

// CreateMilestoneRequest represents a request to create a milestone
type CreateMilestoneRequest struct {
	RepositoryID uuid.UUID  `json:"repository_id" validate:"required"`
	Title        string     `json:"title" validate:"required,max=255"`
	Description  string     `json:"description"`
	DueOn        *time.Time `json:"due_on"`
}

// UpdateMilestoneRequest represents a request to update a milestone
type UpdateMilestoneRequest struct {
	Title       *string    `json:"title,omitempty" validate:"omitempty,max=255"`
	Description *string    `json:"description,omitempty"`
	State       *string    `json:"state,omitempty" validate:"omitempty,oneof=open closed"`
	DueOn       *time.Time `json:"due_on,omitempty"`
}

// MilestoneFilters represents filters for listing milestones
type MilestoneFilters struct {
	State   *string `json:"state,omitempty"`
	Page    int     `json:"page,omitempty"`
	PerPage int     `json:"per_page,omitempty"`
}

type milestoneService struct {
	db     *gorm.DB
	logger *logrus.Logger
}

// NewMilestoneService creates a new milestone service
func NewMilestoneService(db *gorm.DB, logger *logrus.Logger) MilestoneService {
	return &milestoneService{
		db:     db,
		logger: logger,
	}
}

func (s *milestoneService) Create(ctx context.Context, req CreateMilestoneRequest) (*models.Milestone, error) {
	logger := s.logger.WithField("method", "Create")
	
	// Get next milestone number for this repository
	var lastMilestone models.Milestone
	err := s.db.WithContext(ctx).
		Where("repository_id = ?", req.RepositoryID).
		Order("number DESC").
		First(&lastMilestone).Error
	
	nextNumber := 1
	if err == nil {
		nextNumber = lastMilestone.Number + 1
	} else if err != gorm.ErrRecordNotFound {
		logger.WithError(err).Error("Failed to get last milestone number")
		return nil, fmt.Errorf("failed to get next milestone number: %w", err)
	}
	
	milestone := &models.Milestone{
		RepositoryID: req.RepositoryID,
		Number:       nextNumber,
		Title:        req.Title,
		Description:  req.Description,
		State:        "open",
		DueOn:        req.DueOn,
	}
	
	if err := s.db.WithContext(ctx).Create(milestone).Error; err != nil {
		logger.WithError(err).Error("Failed to create milestone")
		return nil, fmt.Errorf("failed to create milestone: %w", err)
	}
	
	return s.loadMilestoneWithRelations(ctx, milestone.ID)
}

func (s *milestoneService) Get(ctx context.Context, repoOwner, repoName string, number int) (*models.Milestone, error) {
	logger := s.logger.WithField("method", "Get")
	
	var milestone models.Milestone
	err := s.db.WithContext(ctx).
		Joins("JOIN repositories ON repositories.id = milestones.repository_id").
		Joins("JOIN users owner ON owner.id = repositories.owner_id").
		Where("owner.username = ? AND repositories.name = ? AND milestones.number = ?", repoOwner, repoName, number).
		Preload("Repository").
		First(&milestone).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("milestone not found")
		}
		logger.WithError(err).Error("Failed to get milestone")
		return nil, fmt.Errorf("failed to get milestone: %w", err)
	}
	
	return &milestone, nil
}

func (s *milestoneService) GetByID(ctx context.Context, id uuid.UUID) (*models.Milestone, error) {
	return s.loadMilestoneWithRelations(ctx, id)
}

func (s *milestoneService) Update(ctx context.Context, id uuid.UUID, req UpdateMilestoneRequest) (*models.Milestone, error) {
	logger := s.logger.WithField("method", "Update")
	
	updates := make(map[string]interface{})
	
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	
	if req.State != nil {
		updates["state"] = *req.State
		if *req.State == "closed" {
			now := time.Now()
			updates["closed_at"] = &now
		} else {
			updates["closed_at"] = nil
		}
	}
	
	if req.DueOn != nil {
		updates["due_on"] = req.DueOn
	}
	
	if len(updates) > 0 {
		if err := s.db.WithContext(ctx).Model(&models.Milestone{}).
			Where("id = ?", id).
			Updates(updates).Error; err != nil {
			logger.WithError(err).Error("Failed to update milestone")
			return nil, fmt.Errorf("failed to update milestone: %w", err)
		}
	}
	
	return s.loadMilestoneWithRelations(ctx, id)
}

func (s *milestoneService) Delete(ctx context.Context, id uuid.UUID) error {
	logger := s.logger.WithField("method", "Delete")
	
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Remove milestone from all issues
		if err := tx.Model(&models.Issue{}).
			Where("milestone_id = ?", id).
			Update("milestone_id", nil).Error; err != nil {
			return fmt.Errorf("failed to remove milestone from issues: %w", err)
		}
		
		// Delete the milestone
		if err := tx.Delete(&models.Milestone{}, "id = ?", id).Error; err != nil {
			return fmt.Errorf("failed to delete milestone: %w", err)
		}
		
		return nil
	})
	
	if err != nil {
		logger.WithError(err).Error("Failed to delete milestone")
		return err
	}
	
	return nil
}

func (s *milestoneService) List(ctx context.Context, repoOwner, repoName string, filters MilestoneFilters) ([]*models.Milestone, int64, error) {
	logger := s.logger.WithField("method", "List")
	
	query := s.db.WithContext(ctx).
		Joins("JOIN repositories ON repositories.id = milestones.repository_id").
		Joins("JOIN users owner ON owner.id = repositories.owner_id").
		Where("owner.username = ? AND repositories.name = ?", repoOwner, repoName)
	
	// Apply filters
	if filters.State != nil {
		query = query.Where("milestones.state = ?", *filters.State)
	}
	
	// Count total records
	var total int64
	if err := query.Model(&models.Milestone{}).Count(&total).Error; err != nil {
		logger.WithError(err).Error("Failed to count milestones")
		return nil, 0, fmt.Errorf("failed to count milestones: %w", err)
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
	
	var milestones []*models.Milestone
	err := query.
		Preload("Repository").
		Order("due_on ASC NULLS LAST, created_at DESC").
		Find(&milestones).Error
	
	if err != nil {
		logger.WithError(err).Error("Failed to list milestones")
		return nil, 0, fmt.Errorf("failed to list milestones: %w", err)
	}
	
	return milestones, total, nil
}

func (s *milestoneService) Close(ctx context.Context, id uuid.UUID) (*models.Milestone, error) {
	state := "closed"
	return s.Update(ctx, id, UpdateMilestoneRequest{
		State: &state,
	})
}

func (s *milestoneService) Reopen(ctx context.Context, id uuid.UUID) (*models.Milestone, error) {
	state := "open"
	return s.Update(ctx, id, UpdateMilestoneRequest{
		State: &state,
	})
}

// Helper methods

func (s *milestoneService) loadMilestoneWithRelations(ctx context.Context, id uuid.UUID) (*models.Milestone, error) {
	var milestone models.Milestone
	err := s.db.WithContext(ctx).
		Preload("Repository").
		First(&milestone, "id = ?", id).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("milestone not found")
		}
		return nil, fmt.Errorf("failed to load milestone: %w", err)
	}
	
	return &milestone, nil
}