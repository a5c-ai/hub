package services

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/a5c-ai/hub/internal/models"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// ArtifactService handles artifact management operations
type ArtifactService struct {
	db           *gorm.DB
	logger       *logrus.Logger
	storagePath  string
	maxSizeMB    int64
	retentionDays int
}

// NewArtifactService creates a new artifact service
func NewArtifactService(db *gorm.DB, logger *logrus.Logger, storagePath string) *ArtifactService {
	return &ArtifactService{
		db:            db,
		logger:        logger,
		storagePath:   storagePath,
		maxSizeMB:     1024, // 1GB default max size per artifact
		retentionDays: 90,   // 90 days default retention
	}
}

// UploadArtifact uploads an artifact for a workflow run
func (s *ArtifactService) UploadArtifact(ctx context.Context, workflowRunID uuid.UUID, name string, reader io.Reader, sizeBytes int64) (*models.Artifact, error) {
	// Validate size
	if sizeBytes > s.maxSizeMB*1024*1024 {
		return nil, fmt.Errorf("artifact size %d bytes exceeds maximum allowed size of %d MB", sizeBytes, s.maxSizeMB)
	}

	// Generate storage path
	storagePath := s.generateStoragePath(workflowRunID, name)

	// Create artifact record
	artifact := &models.Artifact{
		WorkflowRunID: workflowRunID,
		Name:          name,
		Path:          storagePath,
		SizeBytes:     sizeBytes,
		ExpiresAt:     timePtr(time.Now().AddDate(0, 0, s.retentionDays)),
	}

	// TODO: In a real implementation, you would:
	// 1. Stream the data to storage (Azure Blob, S3, etc.)
	// 2. Verify the upload was successful
	// 3. Handle compression if needed

	// Save artifact metadata to database
	if err := s.db.WithContext(ctx).Create(artifact).Error; err != nil {
		return nil, fmt.Errorf("failed to create artifact record: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"artifact_id":     artifact.ID,
		"workflow_run_id": workflowRunID,
		"name":            name,
		"size_bytes":      sizeBytes,
	}).Info("Artifact uploaded successfully")

	return artifact, nil
}

// DownloadArtifact downloads an artifact
func (s *ArtifactService) DownloadArtifact(ctx context.Context, artifactID uuid.UUID) (io.ReadCloser, *models.Artifact, error) {
	// Get artifact metadata
	var artifact models.Artifact
	err := s.db.WithContext(ctx).First(&artifact, "id = ?", artifactID).Error
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get artifact: %w", err)
	}

	// Check if artifact has expired
	if artifact.Expired || (artifact.ExpiresAt != nil && artifact.ExpiresAt.Before(time.Now())) {
		return nil, nil, fmt.Errorf("artifact has expired")
	}

	// TODO: In a real implementation, you would:
	// 1. Open the file from storage (Azure Blob, S3, etc.)
	// 2. Return a ReadCloser to stream the content
	// For now, we'll return a placeholder

	s.logger.WithFields(logrus.Fields{
		"artifact_id": artifactID,
		"name":        artifact.Name,
	}).Info("Artifact download requested")

	// Placeholder - return nil for now since we don't have actual storage
	return nil, &artifact, fmt.Errorf("artifact download not implemented - storage backend needed")
}

// ListArtifacts lists artifacts for a workflow run
func (s *ArtifactService) ListArtifacts(ctx context.Context, workflowRunID uuid.UUID) ([]models.Artifact, error) {
	var artifacts []models.Artifact
	err := s.db.WithContext(ctx).
		Where("workflow_run_id = ? AND expired = false", workflowRunID).
		Order("created_at DESC").
		Find(&artifacts).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list artifacts: %w", err)
	}

	return artifacts, nil
}

// DeleteArtifact deletes an artifact
func (s *ArtifactService) DeleteArtifact(ctx context.Context, artifactID uuid.UUID) error {
	// Get artifact metadata
	var artifact models.Artifact
	err := s.db.WithContext(ctx).First(&artifact, "id = ?", artifactID).Error
	if err != nil {
		return fmt.Errorf("failed to get artifact: %w", err)
	}

	// TODO: In a real implementation, you would:
	// 1. Delete the file from storage (Azure Blob, S3, etc.)
	// 2. Only mark as deleted in DB after successful storage deletion

	// Mark artifact as expired
	if err := s.db.WithContext(ctx).Model(&artifact).Updates(map[string]interface{}{
		"expired":    true,
		"updated_at": time.Now(),
	}).Error; err != nil {
		return fmt.Errorf("failed to delete artifact: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"artifact_id": artifactID,
		"name":        artifact.Name,
	}).Info("Artifact marked as deleted")

	return nil
}

// CleanupExpiredArtifacts removes expired artifacts
func (s *ArtifactService) CleanupExpiredArtifacts(ctx context.Context) error {
	// Find expired artifacts
	var expiredArtifacts []models.Artifact
	err := s.db.WithContext(ctx).
		Where("expired = false AND expires_at < ?", time.Now()).
		Find(&expiredArtifacts).Error
	if err != nil {
		return fmt.Errorf("failed to find expired artifacts: %w", err)
	}

	// Mark them as expired
	for _, artifact := range expiredArtifacts {
		if err := s.DeleteArtifact(ctx, artifact.ID); err != nil {
			s.logger.WithError(err).WithField("artifact_id", artifact.ID).
				Error("Failed to cleanup expired artifact")
		}
	}

	if len(expiredArtifacts) > 0 {
		s.logger.WithField("count", len(expiredArtifacts)).
			Info("Cleaned up expired artifacts")
	}

	return nil
}

// GetArtifact retrieves an artifact by ID
func (s *ArtifactService) GetArtifact(ctx context.Context, artifactID uuid.UUID) (*models.Artifact, error) {
	var artifact models.Artifact
	err := s.db.WithContext(ctx).
		Preload("WorkflowRun").
		First(&artifact, "id = ?", artifactID).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get artifact: %w", err)
	}

	return &artifact, nil
}

// generateStoragePath generates a unique storage path for an artifact
func (s *ArtifactService) generateStoragePath(workflowRunID uuid.UUID, name string) string {
	// Create a path like: artifacts/{workflow_run_id}/{timestamp}_{name}
	timestamp := time.Now().Format("20060102-150405")
	sanitizedName := strings.ReplaceAll(name, " ", "_")
	sanitizedName = strings.ReplaceAll(sanitizedName, "/", "_")
	
	return filepath.Join(
		"artifacts",
		workflowRunID.String(),
		fmt.Sprintf("%s_%s", timestamp, sanitizedName),
	)
}

// Helper function to create time pointer
func timePtr(t time.Time) *time.Time {
	return &t
}

// SetRetentionDays sets the artifact retention period
func (s *ArtifactService) SetRetentionDays(days int) {
	s.retentionDays = days
}

// SetMaxSizeMB sets the maximum artifact size in MB
func (s *ArtifactService) SetMaxSizeMB(sizeMB int64) {
	s.maxSizeMB = sizeMB
}

// GetStorageStats returns storage statistics
func (s *ArtifactService) GetStorageStats(ctx context.Context) (map[string]interface{}, error) {
	var stats struct {
		TotalArtifacts int64 `json:"total_artifacts"`
		TotalSizeBytes int64 `json:"total_size_bytes"`
		ExpiredCount   int64 `json:"expired_count"`
	}

	// Count total artifacts
	err := s.db.WithContext(ctx).Model(&models.Artifact{}).
		Where("expired = false").
		Count(&stats.TotalArtifacts).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count artifacts: %w", err)
	}

	// Sum total size
	err = s.db.WithContext(ctx).Model(&models.Artifact{}).
		Where("expired = false").
		Select("COALESCE(SUM(size_bytes), 0)").
		Scan(&stats.TotalSizeBytes).Error
	if err != nil {
		return nil, fmt.Errorf("failed to sum artifact sizes: %w", err)
	}

	// Count expired artifacts
	err = s.db.WithContext(ctx).Model(&models.Artifact{}).
		Where("expired = true OR expires_at < ?", time.Now()).
		Count(&stats.ExpiredCount).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count expired artifacts: %w", err)
	}

	return map[string]interface{}{
		"total_artifacts":  stats.TotalArtifacts,
		"total_size_bytes": stats.TotalSizeBytes,
		"total_size_mb":    stats.TotalSizeBytes / (1024 * 1024),
		"expired_count":    stats.ExpiredCount,
		"retention_days":   s.retentionDays,
		"max_size_mb":      s.maxSizeMB,
	}, nil
}