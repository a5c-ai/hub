package api

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/a5c-ai/hub/internal/services"
	"github.com/a5c-ai/hub/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// AdminStorageHandlers handles admin storage-related HTTP requests
type AdminStorageHandlers struct {
	artifactService *services.ArtifactService
	storageBackend  storage.Backend
	logger          *logrus.Logger
}

// NewAdminStorageHandlers creates a new admin storage handlers instance
func NewAdminStorageHandlers(artifactService *services.ArtifactService, storageBackend storage.Backend, logger *logrus.Logger) *AdminStorageHandlers {
	return &AdminStorageHandlers{
		artifactService: artifactService,
		storageBackend:  storageBackend,
		logger:          logger,
	}
}

// StorageConfigRequest represents a storage configuration request
type StorageConfigRequest struct {
	Backend       string                 `json:"backend"`
	Azure         storage.AzureConfig    `json:"azure,omitempty"`
	S3            storage.S3Config       `json:"s3,omitempty"`
	Filesystem    storage.FilesystemConfig `json:"filesystem,omitempty"`
	MaxSizeMB     int64                  `json:"max_size_mb"`
	RetentionDays int                    `json:"retention_days"`
}

// StorageConfigResponse represents storage configuration response
type StorageConfigResponse struct {
	Backend       string `json:"backend"`
	MaxSizeMB     int64  `json:"max_size_mb"`
	RetentionDays int    `json:"retention_days"`
	Health        string `json:"health"`
}

// GetStorageConfig handles GET /api/v1/admin/storage/config
func (h *AdminStorageHandlers) GetStorageConfig(c *gin.Context) {
	// In a real implementation, this would come from configuration
	// For now, we'll return the current artifact service settings
	config := StorageConfigResponse{
		Backend:       "filesystem", // This should be read from actual config
		MaxSizeMB:     100,          // This should be read from artifact service
		RetentionDays: h.artifactService.GetRetentionPolicy(),
		Health:        "healthy",    // TODO: Implement actual health check
	}

	c.JSON(http.StatusOK, config)
}

// UpdateStorageConfig handles PUT /api/v1/admin/storage/config
func (h *AdminStorageHandlers) UpdateStorageConfig(c *gin.Context) {
	var req StorageConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate backend type
	validBackends := []string{"filesystem", "azure", "s3"}
	isValid := false
	for _, backend := range validBackends {
		if strings.ToLower(req.Backend) == backend {
			isValid = true
			break
		}
	}

	if !isValid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid storage backend"})
		return
	}

	// Validate retention days
	if req.RetentionDays < 1 || req.RetentionDays > 365 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Retention days must be between 1 and 365"})
		return
	}

	// Validate max size
	if req.MaxSizeMB < 1 || req.MaxSizeMB > 10240 { // Max 10GB
		c.JSON(http.StatusBadRequest, gin.H{"error": "Max size must be between 1MB and 10GB"})
		return
	}

	// Update artifact service settings
	h.artifactService.SetRetentionDays(req.RetentionDays)
	h.artifactService.SetMaxSizeMB(req.MaxSizeMB)

	h.logger.WithFields(logrus.Fields{
		"backend":        req.Backend,
		"retention_days": req.RetentionDays,
		"max_size_mb":    req.MaxSizeMB,
	}).Info("Storage configuration updated")

	c.JSON(http.StatusOK, gin.H{"message": "Storage configuration updated successfully"})
}

// GetStorageUsage handles GET /api/v1/admin/storage/usage
func (h *AdminStorageHandlers) GetStorageUsage(c *gin.Context) {
	stats, err := h.artifactService.GetStorageStats(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to get storage stats")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get storage usage"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetRetentionPolicy handles GET /api/v1/admin/storage/retention
func (h *AdminStorageHandlers) GetRetentionPolicy(c *gin.Context) {
	policy := map[string]interface{}{
		"retention_days": h.artifactService.GetRetentionPolicy(),
	}

	c.JSON(http.StatusOK, policy)
}

// UpdateRetentionPolicy handles PUT /api/v1/admin/storage/retention
func (h *AdminStorageHandlers) UpdateRetentionPolicy(c *gin.Context) {
	var req struct {
		RetentionDays int `json:"retention_days"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.RetentionDays < 1 || req.RetentionDays > 365 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Retention days must be between 1 and 365"})
		return
	}

	h.artifactService.SetRetentionDays(req.RetentionDays)
	
	h.logger.WithField("retention_days", req.RetentionDays).Info("Retention policy updated")

	c.JSON(http.StatusOK, gin.H{
		"message":        "Retention policy updated successfully",
		"retention_days": req.RetentionDays,
	})
}

// ManualCleanup handles DELETE /api/v1/admin/storage/cleanup
func (h *AdminStorageHandlers) ManualCleanup(c *gin.Context) {
	// Cleanup expired artifacts
	if err := h.artifactService.CleanupExpiredArtifacts(c.Request.Context()); err != nil {
		h.logger.WithError(err).Error("Failed to cleanup expired artifacts")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cleanup expired artifacts"})
		return
	}

	// Cleanup old build logs
	if err := h.artifactService.CleanupBuildLogs(c.Request.Context()); err != nil {
		h.logger.WithError(err).Error("Failed to cleanup build logs")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cleanup build logs"})
		return
	}

	h.logger.Info("Manual storage cleanup completed")

	c.JSON(http.StatusOK, gin.H{"message": "Storage cleanup completed successfully"})
}

// SearchBuildLogs handles GET /api/v1/repos/{owner}/{repo}/actions/logs/search
func (h *AdminStorageHandlers) SearchBuildLogs(c *gin.Context) {
	owner := c.Param("owner")
	repo := c.Param("repo")
	query := c.Query("q")
	limitStr := c.DefaultQuery("limit", "50")

	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query is required"})
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 50
	}

	// This is a placeholder - in a real implementation, you would:
	// 1. Get repository ID from owner/repo
	// 2. Call the artifact service search method
	// For now, return a basic response

	results := []map[string]interface{}{
		{
			"message": fmt.Sprintf("Search functionality for repository %s/%s", owner, repo),
			"query":   query,
			"limit":   limit,
			"note":    "Build log search requires Elasticsearch integration",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"results": results,
		"query":   query,
		"total":   len(results),
	})
}

// BatchDeleteArtifacts handles POST /api/v1/admin/storage/artifacts/batch-delete
func (h *AdminStorageHandlers) BatchDeleteArtifacts(c *gin.Context) {
	var req struct {
		ArtifactIDs []string `json:"artifact_ids"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if len(req.ArtifactIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No artifact IDs provided"})
		return
	}

	if len(req.ArtifactIDs) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete more than 100 artifacts at once"})
		return
	}

	var results []map[string]interface{}
	successCount := 0
	errorCount := 0

	for _, artifactIDStr := range req.ArtifactIDs {
		artifactID, err := uuid.Parse(artifactIDStr)
		if err != nil {
			results = append(results, map[string]interface{}{
				"artifact_id": artifactIDStr,
				"success":     false,
				"error":       "Invalid artifact ID format",
			})
			errorCount++
			continue
		}

		if err := h.artifactService.DeleteArtifact(c.Request.Context(), artifactID); err != nil {
			results = append(results, map[string]interface{}{
				"artifact_id": artifactIDStr,
				"success":     false,
				"error":       err.Error(),
			})
			errorCount++
		} else {
			results = append(results, map[string]interface{}{
				"artifact_id": artifactIDStr,
				"success":     true,
			})
			successCount++
		}
	}

	h.logger.WithFields(logrus.Fields{
		"success_count": successCount,
		"error_count":   errorCount,
		"total_count":   len(req.ArtifactIDs),
	}).Info("Batch artifact deletion completed")

	c.JSON(http.StatusOK, gin.H{
		"message":       fmt.Sprintf("Processed %d artifacts: %d successful, %d errors", len(req.ArtifactIDs), successCount, errorCount),
		"success_count": successCount,
		"error_count":   errorCount,
		"results":       results,
	})
}

// GetStorageHealth handles GET /api/v1/admin/storage/health
func (h *AdminStorageHandlers) GetStorageHealth(c *gin.Context) {
	ctx := context.WithTimeout(c.Request.Context(), 10*context.Second)
	
	health := map[string]interface{}{
		"status": "healthy",
		"checks": map[string]interface{}{},
	}

	// Test storage backend connectivity
	testPath := "health-check/test.txt"
	testContent := "health check"
	
	// Try to upload a test file
	if err := h.storageBackend.Upload(ctx, testPath, strings.NewReader(testContent), int64(len(testContent))); err != nil {
		health["status"] = "unhealthy"
		health["checks"].(map[string]interface{})["upload"] = map[string]interface{}{
			"status": "fail",
			"error":  err.Error(),
		}
	} else {
		health["checks"].(map[string]interface{})["upload"] = map[string]interface{}{
			"status": "pass",
		}

		// Try to download the test file
		if reader, err := h.storageBackend.Download(ctx, testPath); err != nil {
			health["status"] = "unhealthy"
			health["checks"].(map[string]interface{})["download"] = map[string]interface{}{
				"status": "fail",
				"error":  err.Error(),
			}
		} else {
			reader.Close()
			health["checks"].(map[string]interface{})["download"] = map[string]interface{}{
				"status": "pass",
			}

			// Clean up test file
			h.storageBackend.Delete(ctx, testPath)
		}
	}

	c.JSON(http.StatusOK, health)
}