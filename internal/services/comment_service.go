package services

import (
	"context"
	"fmt"

	"github.com/a5c-ai/hub/internal/models"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// CommentService provides comment management operations
type CommentService interface {
	// Comment CRUD operations
	Create(ctx context.Context, req CreateCommentRequest) (*models.Comment, error)
	Get(ctx context.Context, id uuid.UUID) (*models.Comment, error)
	Update(ctx context.Context, id uuid.UUID, req UpdateCommentRequest) (*models.Comment, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, issueID uuid.UUID, filters CommentFilters) ([]*models.Comment, int64, error)
}

// CreateCommentRequest represents a request to create a comment
type CreateCommentRequest struct {
	IssueID uuid.UUID  `json:"issue_id" validate:"required"`
	UserID  *uuid.UUID `json:"user_id" validate:"required"`
	Body    string     `json:"body" validate:"required"`
}

// UpdateCommentRequest represents a request to update a comment
type UpdateCommentRequest struct {
	Body string `json:"body" validate:"required"`
}

// CommentFilters represents filters for listing comments
type CommentFilters struct {
	Page    int `json:"page,omitempty"`
	PerPage int `json:"per_page,omitempty"`
}

type commentService struct {
	db     *gorm.DB
	logger *logrus.Logger
}

// NewCommentService creates a new comment service
func NewCommentService(db *gorm.DB, logger *logrus.Logger) CommentService {
	return &commentService{
		db:     db,
		logger: logger,
	}
}

func (s *commentService) Create(ctx context.Context, req CreateCommentRequest) (*models.Comment, error) {
	logger := s.logger.WithField("method", "Create")
	
	comment := &models.Comment{
		IssueID: req.IssueID,
		UserID:  req.UserID,
		Body:    req.Body,
	}
	
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create the comment
		if err := tx.Create(comment).Error; err != nil {
			return fmt.Errorf("failed to create comment: %w", err)
		}
		
		// Update the issue's comment count
		if err := tx.Model(&models.Issue{}).
			Where("id = ?", req.IssueID).
			UpdateColumn("comments_count", gorm.Expr("comments_count + 1")).Error; err != nil {
			return fmt.Errorf("failed to update comment count: %w", err)
		}
		
		return nil
	})
	
	if err != nil {
		logger.WithError(err).Error("Failed to create comment")
		return nil, err
	}
	
	// Load the comment with user relationship
	return s.loadCommentWithRelations(ctx, comment.ID)
}

func (s *commentService) Get(ctx context.Context, id uuid.UUID) (*models.Comment, error) {
	return s.loadCommentWithRelations(ctx, id)
}

func (s *commentService) Update(ctx context.Context, id uuid.UUID, req UpdateCommentRequest) (*models.Comment, error) {
	logger := s.logger.WithField("method", "Update")
	
	err := s.db.WithContext(ctx).Model(&models.Comment{}).
		Where("id = ?", id).
		Update("body", req.Body).Error
	
	if err != nil {
		logger.WithError(err).Error("Failed to update comment")
		return nil, fmt.Errorf("failed to update comment: %w", err)
	}
	
	return s.loadCommentWithRelations(ctx, id)
}

func (s *commentService) Delete(ctx context.Context, id uuid.UUID) error {
	logger := s.logger.WithField("method", "Delete")
	
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Get the comment to get the issue ID
		var comment models.Comment
		if err := tx.First(&comment, "id = ?", id).Error; err != nil {
			return fmt.Errorf("comment not found: %w", err)
		}
		
		// Delete the comment
		if err := tx.Delete(&comment).Error; err != nil {
			return fmt.Errorf("failed to delete comment: %w", err)
		}
		
		// Update the issue's comment count
		if err := tx.Model(&models.Issue{}).
			Where("id = ?", comment.IssueID).
			UpdateColumn("comments_count", gorm.Expr("comments_count - 1")).Error; err != nil {
			return fmt.Errorf("failed to update comment count: %w", err)
		}
		
		return nil
	})
	
	if err != nil {
		logger.WithError(err).Error("Failed to delete comment")
		return err
	}
	
	return nil
}

func (s *commentService) List(ctx context.Context, issueID uuid.UUID, filters CommentFilters) ([]*models.Comment, int64, error) {
	logger := s.logger.WithField("method", "List")
	
	query := s.db.WithContext(ctx).Where("issue_id = ?", issueID)
	
	// Count total records
	var total int64
	if err := query.Model(&models.Comment{}).Count(&total).Error; err != nil {
		logger.WithError(err).Error("Failed to count comments")
		return nil, 0, fmt.Errorf("failed to count comments: %w", err)
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
	
	var comments []*models.Comment
	err := query.
		Preload("User").
		Order("created_at ASC").
		Find(&comments).Error
	
	if err != nil {
		logger.WithError(err).Error("Failed to list comments")
		return nil, 0, fmt.Errorf("failed to list comments: %w", err)
	}
	
	return comments, total, nil
}

// Helper methods

func (s *commentService) loadCommentWithRelations(ctx context.Context, id uuid.UUID) (*models.Comment, error) {
	var comment models.Comment
	err := s.db.WithContext(ctx).
		Preload("User").
		Preload("Issue").
		First(&comment, "id = ?", id).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("comment not found")
		}
		return nil, fmt.Errorf("failed to load comment: %w", err)
	}
	
	return &comment, nil
}