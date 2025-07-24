package services

import (
	"context"
	"fmt"

	"github.com/a5c-ai/hub/internal/models"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// LabelService provides label management operations
type LabelService interface {
	// Label CRUD operations
	Create(ctx context.Context, req CreateLabelRequest) (*models.Label, error)
	Get(ctx context.Context, repoOwner, repoName, name string) (*models.Label, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Label, error)
	Update(ctx context.Context, id uuid.UUID, req UpdateLabelRequest) (*models.Label, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, repoOwner, repoName string, filters LabelFilters) ([]*models.Label, int64, error)
}

// CreateLabelRequest represents a request to create a label
type CreateLabelRequest struct {
	RepositoryID uuid.UUID `json:"repository_id" validate:"required"`
	Name         string    `json:"name" validate:"required,max=255"`
	Color        string    `json:"color" validate:"required,len=7"`
	Description  string    `json:"description"`
}

// UpdateLabelRequest represents a request to update a label
type UpdateLabelRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,max=255"`
	Color       *string `json:"color,omitempty" validate:"omitempty,len=7"`
	Description *string `json:"description,omitempty"`
}

// LabelFilters represents filters for listing labels
type LabelFilters struct {
	Page    int `json:"page,omitempty"`
	PerPage int `json:"per_page,omitempty"`
}

type labelService struct {
	db     *gorm.DB
	logger *logrus.Logger
}

// NewLabelService creates a new label service
func NewLabelService(db *gorm.DB, logger *logrus.Logger) LabelService {
	return &labelService{
		db:     db,
		logger: logger,
	}
}

func (s *labelService) Create(ctx context.Context, req CreateLabelRequest) (*models.Label, error) {
	logger := s.logger.WithField("method", "Create")
	
	// Check if label name already exists in repository
	var existingLabel models.Label
	err := s.db.WithContext(ctx).
		Where("repository_id = ? AND name = ?", req.RepositoryID, req.Name).
		First(&existingLabel).Error
	
	if err == nil {
		return nil, fmt.Errorf("label with name '%s' already exists", req.Name)
	} else if err != gorm.ErrRecordNotFound {
		logger.WithError(err).Error("Failed to check existing label")
		return nil, fmt.Errorf("failed to check existing label: %w", err)
	}
	
	label := &models.Label{
		RepositoryID: req.RepositoryID,
		Name:         req.Name,
		Color:        req.Color,
		Description:  req.Description,
	}
	
	if err := s.db.WithContext(ctx).Create(label).Error; err != nil {
		logger.WithError(err).Error("Failed to create label")
		return nil, fmt.Errorf("failed to create label: %w", err)
	}
	
	return s.loadLabelWithRelations(ctx, label.ID)
}

func (s *labelService) Get(ctx context.Context, repoOwner, repoName, name string) (*models.Label, error) {
	logger := s.logger.WithField("method", "Get")
	
	var label models.Label
	err := s.db.WithContext(ctx).
		Joins("JOIN repositories ON repositories.id = labels.repository_id").
		Joins("JOIN users owner ON owner.id = repositories.owner_id").
		Where("owner.username = ? AND repositories.name = ? AND labels.name = ?", repoOwner, repoName, name).
		Preload("Repository").
		First(&label).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("label not found")
		}
		logger.WithError(err).Error("Failed to get label")
		return nil, fmt.Errorf("failed to get label: %w", err)
	}
	
	return &label, nil
}

func (s *labelService) GetByID(ctx context.Context, id uuid.UUID) (*models.Label, error) {
	return s.loadLabelWithRelations(ctx, id)
}

func (s *labelService) Update(ctx context.Context, id uuid.UUID, req UpdateLabelRequest) (*models.Label, error) {
	logger := s.logger.WithField("method", "Update")
	
	updates := make(map[string]interface{})
	
	if req.Name != nil {
		// Check if the new name conflicts with existing labels in the same repository
		var label models.Label
		if err := s.db.WithContext(ctx).First(&label, "id = ?", id).Error; err != nil {
			return nil, fmt.Errorf("label not found: %w", err)
		}
		
		var existingLabel models.Label
		err := s.db.WithContext(ctx).
			Where("repository_id = ? AND name = ? AND id != ?", label.RepositoryID, *req.Name, id).
			First(&existingLabel).Error
		
		if err == nil {
			return nil, fmt.Errorf("label with name '%s' already exists", *req.Name)
		} else if err != gorm.ErrRecordNotFound {
			logger.WithError(err).Error("Failed to check existing label")
			return nil, fmt.Errorf("failed to check existing label: %w", err)
		}
		
		updates["name"] = *req.Name
	}
	
	if req.Color != nil {
		updates["color"] = *req.Color
	}
	
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	
	if len(updates) > 0 {
		if err := s.db.WithContext(ctx).Model(&models.Label{}).
			Where("id = ?", id).
			Updates(updates).Error; err != nil {
			logger.WithError(err).Error("Failed to update label")
			return nil, fmt.Errorf("failed to update label: %w", err)
		}
	}
	
	return s.loadLabelWithRelations(ctx, id)
}

func (s *labelService) Delete(ctx context.Context, id uuid.UUID) error {
	logger := s.logger.WithField("method", "Delete")
	
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Remove label from all issues
		if err := tx.Where("label_id = ?", id).Delete(&models.IssueLabel{}).Error; err != nil {
			return fmt.Errorf("failed to remove label from issues: %w", err)
		}
		
		// Delete the label
		if err := tx.Delete(&models.Label{}, "id = ?", id).Error; err != nil {
			return fmt.Errorf("failed to delete label: %w", err)
		}
		
		return nil
	})
	
	if err != nil {
		logger.WithError(err).Error("Failed to delete label")
		return err
	}
	
	return nil
}

func (s *labelService) List(ctx context.Context, repoOwner, repoName string, filters LabelFilters) ([]*models.Label, int64, error) {
	logger := s.logger.WithField("method", "List")
	
	query := s.db.WithContext(ctx).
		Joins("JOIN repositories ON repositories.id = labels.repository_id").
		Joins("JOIN users owner ON owner.id = repositories.owner_id").
		Where("owner.username = ? AND repositories.name = ?", repoOwner, repoName)
	
	// Count total records
	var total int64
	if err := query.Model(&models.Label{}).Count(&total).Error; err != nil {
		logger.WithError(err).Error("Failed to count labels")
		return nil, 0, fmt.Errorf("failed to count labels: %w", err)
	}
	
	// Apply pagination
	page := 1
	if filters.Page > 0 {
		page = filters.Page
	}
	
	perPage := 100
	if filters.PerPage > 0 && filters.PerPage <= 100 {
		perPage = filters.PerPage
	}
	
	offset := (page - 1) * perPage
	query = query.Offset(offset).Limit(perPage)
	
	var labels []*models.Label
	err := query.
		Preload("Repository").
		Order("name ASC").
		Find(&labels).Error
	
	if err != nil {
		logger.WithError(err).Error("Failed to list labels")
		return nil, 0, fmt.Errorf("failed to list labels: %w", err)
	}
	
	return labels, total, nil
}

// Helper methods

func (s *labelService) loadLabelWithRelations(ctx context.Context, id uuid.UUID) (*models.Label, error) {
	var label models.Label
	err := s.db.WithContext(ctx).
		Preload("Repository").
		First(&label, "id = ?", id).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("label not found")
		}
		return nil, fmt.Errorf("failed to load label: %w", err)
	}
	
	return &label, nil
}