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

// RunnerService manages workflow runners
type RunnerService struct {
	db     *gorm.DB
	logger *logrus.Logger
}

// NewRunnerService creates a new runner service
func NewRunnerService(db *gorm.DB, logger *logrus.Logger) *RunnerService {
	return &RunnerService{
		db:     db,
		logger: logger,
	}
}

// RegisterRunnerRequest represents a request to register a new runner
type RegisterRunnerRequest struct {
	Name           string             `json:"name" binding:"required"`
	Labels         []string           `json:"labels" binding:"required"`
	Type           models.RunnerType  `json:"type" binding:"required"`
	Version        *string            `json:"version,omitempty"`
	OS             *string            `json:"os,omitempty"`
	Architecture   *string            `json:"architecture,omitempty"`
	RepositoryID   *uuid.UUID         `json:"repository_id,omitempty"`
	OrganizationID *uuid.UUID         `json:"organization_id,omitempty"`
}

// UpdateRunnerRequest represents a request to update a runner
type UpdateRunnerRequest struct {
	Name         *string            `json:"name,omitempty"`
	Labels       *[]string          `json:"labels,omitempty"`
	Status       *models.RunnerStatus `json:"status,omitempty"`
	Version      *string            `json:"version,omitempty"`
	OS           *string            `json:"os,omitempty"`
	Architecture *string            `json:"architecture,omitempty"`
}

// ListRunnersRequest represents a request to list runners
type ListRunnersRequest struct {
	RepositoryID   *uuid.UUID            `json:"repository_id,omitempty"`
	OrganizationID *uuid.UUID            `json:"organization_id,omitempty"`
	Status         *models.RunnerStatus  `json:"status,omitempty"`
	Type           *models.RunnerType    `json:"type,omitempty"`
	Labels         []string              `json:"labels,omitempty"`
	Limit          int                   `json:"limit,omitempty"`
	Offset         int                   `json:"offset,omitempty"`
}

// RegisterRunner registers a new runner
func (s *RunnerService) RegisterRunner(ctx context.Context, req RegisterRunnerRequest) (*models.Runner, error) {
	runner := &models.Runner{
		Name:           req.Name,
		Labels:         req.Labels,
		Status:         models.RunnerStatusOnline,
		Type:           req.Type,
		Version:        req.Version,
		OS:             req.OS,
		Architecture:   req.Architecture,
		RepositoryID:   req.RepositoryID,
		OrganizationID: req.OrganizationID,
		LastSeenAt:     &[]time.Time{time.Now()}[0],
	}

	if err := s.db.WithContext(ctx).Create(runner).Error; err != nil {
		return nil, fmt.Errorf("failed to register runner: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"runner_id":   runner.ID,
		"name":        runner.Name,
		"type":        runner.Type,
		"labels":      runner.Labels,
	}).Info("Runner registered")

	return runner, nil
}

// GetRunner retrieves a runner by ID
func (s *RunnerService) GetRunner(ctx context.Context, id uuid.UUID) (*models.Runner, error) {
	var runner models.Runner
	err := s.db.WithContext(ctx).
		Preload("Repository").
		Preload("Organization").
		First(&runner, "id = ?", id).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get runner: %w", err)
	}

	return &runner, nil
}

// ListRunners lists runners based on filters
func (s *RunnerService) ListRunners(ctx context.Context, req ListRunnersRequest) ([]models.Runner, int64, error) {
	query := s.db.WithContext(ctx).Model(&models.Runner{})

	// Apply filters
	if req.RepositoryID != nil {
		query = query.Where("repository_id = ?", *req.RepositoryID)
	}

	if req.OrganizationID != nil {
		query = query.Where("organization_id = ?", *req.OrganizationID)
	}

	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	if req.Type != nil {
		query = query.Where("type = ?", *req.Type)
	}

	// Filter by labels (runner must have all specified labels)
	if len(req.Labels) > 0 {
		for _, label := range req.Labels {
			query = query.Where("labels ? ?", label)
		}
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count runners: %w", err)
	}

	// Apply pagination
	if req.Limit > 0 {
		query = query.Limit(req.Limit)
	} else {
		query = query.Limit(50) // Default limit
	}

	if req.Offset > 0 {
		query = query.Offset(req.Offset)
	}

	var runners []models.Runner
	err := query.Preload("Repository").
		Preload("Organization").
		Order("created_at DESC").
		Find(&runners).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list runners: %w", err)
	}

	return runners, total, nil
}

// UpdateRunner updates a runner
func (s *RunnerService) UpdateRunner(ctx context.Context, id uuid.UUID, req UpdateRunnerRequest) (*models.Runner, error) {
	runner, err := s.GetRunner(ctx, id)
	if err != nil {
		return nil, err
	}

	updates := make(map[string]interface{})

	if req.Name != nil {
		updates["name"] = *req.Name
	}

	if req.Labels != nil {
		updates["labels"] = *req.Labels
	}

	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if req.Version != nil {
		updates["version"] = *req.Version
	}

	if req.OS != nil {
		updates["os"] = *req.OS
	}

	if req.Architecture != nil {
		updates["architecture"] = *req.Architecture
	}

	if len(updates) > 0 {
		updates["updated_at"] = time.Now()
		if err := s.db.WithContext(ctx).Model(runner).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("failed to update runner: %w", err)
		}
	}

	// Reload runner with updates
	return s.GetRunner(ctx, id)
}

// DeleteRunner deletes a runner
func (s *RunnerService) DeleteRunner(ctx context.Context, id uuid.UUID) error {
	result := s.db.WithContext(ctx).Delete(&models.Runner{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete runner: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("runner not found")
	}

	s.logger.WithField("runner_id", id).Info("Runner deleted")
	return nil
}

// UpdateRunnerHeartbeat updates the last seen timestamp for a runner
func (s *RunnerService) UpdateRunnerHeartbeat(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	err := s.db.WithContext(ctx).Model(&models.Runner{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"last_seen_at": &now,
			"status":       models.RunnerStatusOnline,
			"updated_at":   now,
		}).Error

	if err != nil {
		return fmt.Errorf("failed to update runner heartbeat: %w", err)
	}

	return nil
}

// FindAvailableRunner finds an available runner for a job
func (s *RunnerService) FindAvailableRunner(ctx context.Context, requiredLabels []string, repositoryID *uuid.UUID, organizationID *uuid.UUID) (*models.Runner, error) {
	query := s.db.WithContext(ctx).Model(&models.Runner{}).
		Where("status = ?", models.RunnerStatusOnline)

	// Check for repository-specific runners first
	if repositoryID != nil {
		repoQuery := query.Where("repository_id = ?", *repositoryID)
		
		// Apply label filters
		for _, label := range requiredLabels {
			repoQuery = repoQuery.Where("labels ? ?", label)
		}

		var runner models.Runner
		if err := repoQuery.First(&runner).Error; err == nil {
			return &runner, nil
		}
	}

	// Check for organization runners
	if organizationID != nil {
		orgQuery := query.Where("organization_id = ? AND repository_id IS NULL", *organizationID)
		
		// Apply label filters
		for _, label := range requiredLabels {
			orgQuery = orgQuery.Where("labels ? ?", label)
		}

		var runner models.Runner
		if err := orgQuery.First(&runner).Error; err == nil {
			return &runner, nil
		}
	}

	// Check for global runners (no repository or organization)
	globalQuery := query.Where("repository_id IS NULL AND organization_id IS NULL")
	
	// Apply label filters
	for _, label := range requiredLabels {
		globalQuery = globalQuery.Where("labels ? ?", label)
	}

	var runner models.Runner
	if err := globalQuery.First(&runner).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // No available runners
		}
		return nil, fmt.Errorf("failed to find available runner: %w", err)
	}

	return &runner, nil
}

// AssignJobToRunner assigns a job to a runner
func (s *RunnerService) AssignJobToRunner(ctx context.Context, jobID, runnerID uuid.UUID) error {
	now := time.Now()
	
	// Update job with runner assignment
	if err := s.db.WithContext(ctx).Model(&models.Job{}).
		Where("id = ?", jobID).
		Updates(map[string]interface{}{
			"runner_id":  runnerID,
			"status":     models.JobStatusInProgress,
			"started_at": &now,
			"updated_at": now,
		}).Error; err != nil {
		return fmt.Errorf("failed to assign job to runner: %w", err)
	}

	// Update runner status to busy
	if err := s.db.WithContext(ctx).Model(&models.Runner{}).
		Where("id = ?", runnerID).
		Updates(map[string]interface{}{
			"status":     models.RunnerStatusBusy,
			"updated_at": now,
		}).Error; err != nil {
		return fmt.Errorf("failed to update runner status: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"job_id":    jobID,
		"runner_id": runnerID,
	}).Info("Job assigned to runner")

	return nil
}

// ReleaseRunner releases a runner from a job
func (s *RunnerService) ReleaseRunner(ctx context.Context, runnerID uuid.UUID) error {
	now := time.Now()
	
	if err := s.db.WithContext(ctx).Model(&models.Runner{}).
		Where("id = ?", runnerID).
		Updates(map[string]interface{}{
			"status":     models.RunnerStatusOnline,
			"updated_at": now,
		}).Error; err != nil {
		return fmt.Errorf("failed to release runner: %w", err)
	}

	s.logger.WithField("runner_id", runnerID).Info("Runner released")
	return nil
}

// MarkRunnerOffline marks runners as offline if they haven't sent heartbeat recently
func (s *RunnerService) MarkRunnerOffline(ctx context.Context, timeout time.Duration) error {
	cutoff := time.Now().Add(-timeout)
	
	result := s.db.WithContext(ctx).Model(&models.Runner{}).
		Where("last_seen_at < ? AND status != ?", cutoff, models.RunnerStatusOffline).
		Updates(map[string]interface{}{
			"status":     models.RunnerStatusOffline,
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to mark runners offline: %w", result.Error)
	}

	if result.RowsAffected > 0 {
		s.logger.WithFields(logrus.Fields{
			"count":   result.RowsAffected,
			"timeout": timeout,
		}).Info("Marked runners as offline")
	}

	return nil
}

// GetRunnerStats returns statistics about runners
func (s *RunnerService) GetRunnerStats(ctx context.Context, repositoryID *uuid.UUID, organizationID *uuid.UUID) (map[string]interface{}, error) {
	query := s.db.WithContext(ctx).Model(&models.Runner{})

	if repositoryID != nil {
		query = query.Where("repository_id = ?", *repositoryID)
	} else if organizationID != nil {
		query = query.Where("organization_id = ? OR (organization_id IS NULL AND repository_id IS NULL)", *organizationID)
	}

	var stats struct {
		Total   int64 `gorm:"column:total"`
		Online  int64 `gorm:"column:online"`
		Busy    int64 `gorm:"column:busy"`
		Offline int64 `gorm:"column:offline"`
	}

	err := query.Select(`
		COUNT(*) as total,
		COUNT(CASE WHEN status = 'online' THEN 1 END) as online,
		COUNT(CASE WHEN status = 'busy' THEN 1 END) as busy,
		COUNT(CASE WHEN status = 'offline' THEN 1 END) as offline
	`).Scan(&stats).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get runner stats: %w", err)
	}

	return map[string]interface{}{
		"total":     stats.Total,
		"online":    stats.Online,
		"busy":      stats.Busy,
		"offline":   stats.Offline,
		"available": stats.Online,
		"timestamp": time.Now(),
	}, nil
}