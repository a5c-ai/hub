package services

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/a5c-ai/hub/internal/models"
	"github.com/a5c-ai/hub/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupArtifactServiceTest(t *testing.T) (*ArtifactService, *gorm.DB, func()) {
	// Create in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Migrate tables
	err = db.AutoMigrate(&models.Artifact{})
	require.NoError(t, err)

	// Create temporary directory for filesystem storage
	tempDir, err := os.MkdirTemp("", "artifact_test_*")
	require.NoError(t, err)

	// Create filesystem storage backend
	storageConfig := storage.Config{
		Backend: "filesystem",
		Filesystem: storage.FilesystemConfig{
			BasePath: tempDir,
		},
	}

	storageBackend, err := storage.NewBackend(storageConfig)
	require.NoError(t, err)

	// Create logger
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise in tests

	// Create artifact service
	service := NewArtifactService(db, logger, storageBackend, 100, 30) // 100MB max, 30 days retention

	// Cleanup function
	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return service, db, cleanup
}

func TestArtifactService_UploadAndDownload(t *testing.T) {
	service, _, cleanup := setupArtifactServiceTest(t)
	defer cleanup()

	ctx := context.Background()
	workflowRunID := uuid.New()
	artifactName := "test-artifact.txt"
	content := "This is test artifact content"
	reader := strings.NewReader(content)

	// Test upload
	artifact, err := service.UploadArtifact(ctx, workflowRunID, artifactName, reader, int64(len(content)))
	require.NoError(t, err)
	assert.NotNil(t, artifact)
	assert.Equal(t, workflowRunID, artifact.WorkflowRunID)
	assert.Equal(t, artifactName, artifact.Name)
	assert.Equal(t, int64(len(content)), artifact.SizeBytes)
	assert.False(t, artifact.Expired)
	assert.NotNil(t, artifact.ExpiresAt)

	// Test download
	downloadReader, downloadedArtifact, err := service.DownloadArtifact(ctx, artifact.ID)
	require.NoError(t, err)
	assert.Equal(t, artifact.ID, downloadedArtifact.ID)
	
	downloadedContent, err := io.ReadAll(downloadReader)
	require.NoError(t, err)
	downloadReader.Close()
	
	assert.Equal(t, content, string(downloadedContent))
}

func TestArtifactService_ListArtifacts(t *testing.T) {
	service, _, cleanup := setupArtifactServiceTest(t)
	defer cleanup()

	ctx := context.Background()
	workflowRunID := uuid.New()

	// Upload multiple artifacts
	artifacts := []string{"artifact1.txt", "artifact2.txt", "artifact3.txt"}
	for _, name := range artifacts {
		content := "Content for " + name
		reader := strings.NewReader(content)
		_, err := service.UploadArtifact(ctx, workflowRunID, name, reader, int64(len(content)))
		require.NoError(t, err)
	}

	// List artifacts
	listedArtifacts, err := service.ListArtifacts(ctx, workflowRunID)
	require.NoError(t, err)
	assert.Len(t, listedArtifacts, 3)

	// Check that all artifact names are present
	names := make(map[string]bool)
	for _, artifact := range listedArtifacts {
		names[artifact.Name] = true
	}
	
	for _, expectedName := range artifacts {
		assert.True(t, names[expectedName], "Expected artifact %s not found", expectedName)
	}
}

func TestArtifactService_DeleteArtifact(t *testing.T) {
	service, _, cleanup := setupArtifactServiceTest(t)
	defer cleanup()

	ctx := context.Background()
	workflowRunID := uuid.New()
	artifactName := "test-artifact.txt"
	content := "This is test artifact content"
	reader := strings.NewReader(content)

	// Upload artifact
	artifact, err := service.UploadArtifact(ctx, workflowRunID, artifactName, reader, int64(len(content)))
	require.NoError(t, err)

	// Delete artifact
	err = service.DeleteArtifact(ctx, artifact.ID)
	require.NoError(t, err)

	// Try to download deleted artifact - should fail
	_, _, err = service.DownloadArtifact(ctx, artifact.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "has expired")
}

func TestArtifactService_GetStorageStats(t *testing.T) {
	service, _, cleanup := setupArtifactServiceTest(t)
	defer cleanup()

	ctx := context.Background()
	workflowRunID := uuid.New()

	// Upload some artifacts
	content1 := "Small content"
	content2 := "Larger content with more text"
	
	reader1 := strings.NewReader(content1)
	_, err := service.UploadArtifact(ctx, workflowRunID, "small.txt", reader1, int64(len(content1)))
	require.NoError(t, err)

	reader2 := strings.NewReader(content2)
	_, err = service.UploadArtifact(ctx, workflowRunID, "large.txt", reader2, int64(len(content2)))
	require.NoError(t, err)

	// Get storage stats
	stats, err := service.GetStorageStats(ctx)
	require.NoError(t, err)

	assert.Equal(t, int64(2), stats["total_artifacts"])
	assert.Equal(t, int64(len(content1)+len(content2)), stats["total_size_bytes"])
	assert.Equal(t, int64(0), stats["expired_count"])
	assert.Equal(t, 30, stats["retention_days"])
	assert.Equal(t, int64(100), stats["max_size_mb"])
}

func TestArtifactService_BuildLogStorage(t *testing.T) {
	service, _, cleanup := setupArtifactServiceTest(t)
	defer cleanup()

	ctx := context.Background()
	jobID := uuid.New()
	logContent := "2023-01-01 10:00:00 INFO Starting build\n2023-01-01 10:00:01 INFO Build completed successfully"

	// Store build log
	err := service.StoreBuildLog(ctx, jobID, logContent)
	require.NoError(t, err)

	// Retrieve build log
	retrievedLog, err := service.GetBuildLog(ctx, jobID)
	require.NoError(t, err)
	assert.Equal(t, logContent, retrievedLog)

	// Try to get non-existent log
	nonExistentJobID := uuid.New()
	_, err = service.GetBuildLog(ctx, nonExistentJobID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "build log not found")
}

func TestArtifactService_RetentionPolicy(t *testing.T) {
	service, db, cleanup := setupArtifactServiceTest(t)
	defer cleanup()

	ctx := context.Background()
	workflowRunID := uuid.New()

	// Create an old artifact by directly inserting to database
	oldArtifact := &models.Artifact{
		WorkflowRunID: workflowRunID,
		Name:          "old-artifact.txt",
		Path:          "artifacts/old/old-artifact.txt",
		SizeBytes:     100,
		Expired:       false,
		CreatedAt:     time.Now().AddDate(0, 0, -40), // 40 days old
		ExpiresAt:     timePtr(time.Now().AddDate(0, 0, -10)), // Expired 10 days ago
	}
	
	err := db.Create(oldArtifact).Error
	require.NoError(t, err)

	// Run cleanup
	err = service.CleanupExpiredArtifacts(ctx)
	require.NoError(t, err)

	// Check that artifact is marked as expired
	var updatedArtifact models.Artifact
	err = db.First(&updatedArtifact, oldArtifact.ID).Error
	require.NoError(t, err)
	assert.True(t, updatedArtifact.Expired)
}

func TestArtifactService_SizeValidation(t *testing.T) {
	service, _, cleanup := setupArtifactServiceTest(t)
	defer cleanup()

	ctx := context.Background()
	workflowRunID := uuid.New()
	artifactName := "large-artifact.txt"
	
	// Create content larger than max size (100MB limit in test setup)
	largeContent := strings.Repeat("x", 101*1024*1024) // 101MB
	reader := strings.NewReader(largeContent)

	// Try to upload - should fail
	_, err := service.UploadArtifact(ctx, workflowRunID, artifactName, reader, int64(len(largeContent)))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds maximum allowed size")
}

// timePtr helper function - using the one from artifact_service.go