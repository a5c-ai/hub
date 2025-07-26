package services

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/a5c-ai/hub/internal/models"
	"github.com/a5c-ai/hub/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// ArtifactService handles artifact management operations
type ArtifactService struct {
	db            *gorm.DB
	logger        *logrus.Logger
	storage       storage.Backend
	maxSizeMB     int64
	retentionDays int
}

// NewArtifactService creates a new artifact service
func NewArtifactService(db *gorm.DB, logger *logrus.Logger, storageBackend storage.Backend, maxSizeMB int64, retentionDays int) *ArtifactService {
	return &ArtifactService{
		db:            db,
		logger:        logger,
		storage:       storageBackend,
		maxSizeMB:     maxSizeMB,
		retentionDays: retentionDays,
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

	// Create artifact record with generated ID
	artifact := &models.Artifact{
		ID:            uuid.New(),
		WorkflowRunID: workflowRunID,
		Name:          name,
		Path:          storagePath,
		SizeBytes:     sizeBytes,
		ExpiresAt:     timePtr(time.Now().AddDate(0, 0, s.retentionDays)),
	}

	// Upload to configured storage backend
	s.logger.WithFields(logrus.Fields{
		"workflow_run_id": workflowRunID,
		"name":            name,
		"size_bytes":      sizeBytes,
		"path":            storagePath,
	}).Info("Uploading artifact to storage backend")

	// Upload to storage backend
	if err := s.storage.Upload(ctx, storagePath, reader, sizeBytes); err != nil {
		return nil, fmt.Errorf("failed to upload artifact to storage: %w", err)
	}

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

	s.logger.WithFields(logrus.Fields{
		"artifact_id": artifactID,
		"name":        artifact.Name,
		"path":        artifact.Path,
	}).Info("Artifact download requested")

	// Download from storage backend
	reader, err := s.storage.Download(ctx, artifact.Path)
	if err != nil {
		return nil, &artifact, fmt.Errorf("failed to download artifact from storage: %w", err)
	}

	return reader, &artifact, nil
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

	// Delete from storage backend first
	if err := s.storage.Delete(ctx, artifact.Path); err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"artifact_id": artifactID,
			"path":        artifact.Path,
		}).Warn("Failed to delete artifact from storage, marking as expired anyway")
	}

	// Mark artifact as expired in database
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

// EnforceRetentionPolicy enforces retention policies for artifacts
func (s *ArtifactService) EnforceRetentionPolicy(ctx context.Context) error {
	// Get all non-expired artifacts that should be expired based on retention policy
	cutoffDate := time.Now().AddDate(0, 0, -s.retentionDays)

	var artifactsToExpire []models.Artifact
	err := s.db.WithContext(ctx).
		Where("expired = false AND created_at < ?", cutoffDate).
		Find(&artifactsToExpire).Error
	if err != nil {
		return fmt.Errorf("failed to find artifacts for retention policy: %w", err)
	}

	// Mark them as expired
	for _, artifact := range artifactsToExpire {
		if err := s.DeleteArtifact(ctx, artifact.ID); err != nil {
			s.logger.WithError(err).WithField("artifact_id", artifact.ID).
				Error("Failed to enforce retention policy on artifact")
		}
	}

	if len(artifactsToExpire) > 0 {
		s.logger.WithFields(logrus.Fields{
			"count":          len(artifactsToExpire),
			"retention_days": s.retentionDays,
		}).Info("Enforced retention policy on artifacts")
	}

	return nil
}

// SetRetentionPolicy updates the retention policy for new artifacts
func (s *ArtifactService) SetRetentionPolicy(days int) {
	s.retentionDays = days
	s.logger.WithField("retention_days", days).Info("Updated artifact retention policy")
}

// GetRetentionPolicy returns the current retention policy in days
func (s *ArtifactService) GetRetentionPolicy() int {
	return s.retentionDays
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

// Build Log Storage Methods

// StoreBuildLog stores build logs for a job
func (s *ArtifactService) StoreBuildLog(ctx context.Context, jobID uuid.UUID, logContent string) error {
	// Generate storage path for build log
	logPath := s.generateBuildLogPath(jobID)

	// Store the log content
	reader := strings.NewReader(logContent)
	if err := s.storage.Upload(ctx, logPath, reader, int64(len(logContent))); err != nil {
		return fmt.Errorf("failed to store build log: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"job_id":   jobID,
		"log_path": logPath,
		"log_size": len(logContent),
	}).Info("Build log stored successfully")

	return nil
}

// GetBuildLog retrieves build logs for a job
func (s *ArtifactService) GetBuildLog(ctx context.Context, jobID uuid.UUID) (string, error) {
	logPath := s.generateBuildLogPath(jobID)

	// Check if log exists
	exists, err := s.storage.Exists(ctx, logPath)
	if err != nil {
		return "", fmt.Errorf("failed to check build log existence: %w", err)
	}

	if !exists {
		return "", fmt.Errorf("build log not found for job %s", jobID)
	}

	// Download the log
	reader, err := s.storage.Download(ctx, logPath)
	if err != nil {
		return "", fmt.Errorf("failed to download build log: %w", err)
	}
	defer reader.Close()

	// Read the content
	logBytes, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("failed to read build log content: %w", err)
	}

	return string(logBytes), nil
}

// SearchBuildLogs searches through build logs for specific content
func (s *ArtifactService) SearchBuildLogs(ctx context.Context, repositoryID uuid.UUID, query string, limit int) ([]map[string]interface{}, error) {
	// TODO: For now, this is a basic implementation
	// In a production environment, you would want to:
	// 1. Use Elasticsearch or similar search engine for indexing logs
	// 2. Store log metadata in database for efficient searching
	// 3. Implement advanced search features (regex, filters, etc.)

	s.logger.WithFields(logrus.Fields{
		"repository_id": repositoryID,
		"query":         query,
		"limit":         limit,
	}).Info("Build log search requested")

	// For now, return empty results with a note
	return []map[string]interface{}{
		{
			"message": "Build log search functionality requires Elasticsearch integration",
			"query":   query,
			"status":  "not_implemented",
		},
	}, nil
}

// CleanupBuildLogs removes old build logs based on retention policy
func (s *ArtifactService) CleanupBuildLogs(ctx context.Context) error {
	// List all build logs
	buildLogs, err := s.storage.List(ctx, "build-logs/")
	if err != nil {
		return fmt.Errorf("failed to list build logs for cleanup: %w", err)
	}

	cutoffDate := time.Now().AddDate(0, 0, -s.retentionDays)
	cleanedCount := 0

	for _, logPath := range buildLogs {
		// Get last modified time
		modTime, err := s.storage.GetLastModified(ctx, logPath)
		if err != nil {
			s.logger.WithError(err).WithField("log_path", logPath).
				Warn("Failed to get build log modification time")
			continue
		}

		// Delete if older than retention period
		if modTime.Before(cutoffDate) {
			if err := s.storage.Delete(ctx, logPath); err != nil {
				s.logger.WithError(err).WithField("log_path", logPath).
					Error("Failed to delete old build log")
			} else {
				cleanedCount++
			}
		}
	}

	if cleanedCount > 0 {
		s.logger.WithFields(logrus.Fields{
			"count":          cleanedCount,
			"retention_days": s.retentionDays,
		}).Info("Cleaned up old build logs")
	}

	return nil
}

// generateBuildLogPath generates a storage path for build logs
func (s *ArtifactService) generateBuildLogPath(jobID uuid.UUID) string {
	return filepath.Join("build-logs", jobID.String()+".log")
}
