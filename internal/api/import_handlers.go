package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ImportHandlers handles repository import endpoints.
type ImportHandlers struct{}

// NewImportHandlers creates a new ImportHandlers.
func NewImportHandlers() *ImportHandlers {
	return &ImportHandlers{}
}

// InitiateImport starts an import job from an external Git service.
func (h *ImportHandlers) InitiateImport(c *gin.Context) {
	type request struct {
		URL   string `json:"url" binding:"required"`
		Token string `json:"token"`
	}
	var req request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	jobID := uuid.New().String()
	// TODO: enqueue background import job using job queue
	c.JSON(http.StatusAccepted, gin.H{"job_id": jobID})
}

// GetImportStatus returns the status of an import job.
func (h *ImportHandlers) GetImportStatus(c *gin.Context) {
	jobID := c.Param("job_id")
	// TODO: fetch job status from job queue or job store
	c.JSON(http.StatusOK, gin.H{"job_id": jobID, "status": "pending"})
}

// ExportHandlers handles repository export endpoints.
type ExportHandlers struct{}

// NewExportHandlers creates a new ExportHandlers.
func NewExportHandlers() *ExportHandlers {
	return &ExportHandlers{}
}

// InitiateExport starts an export job to an external Git service.
func (h *ExportHandlers) InitiateExport(c *gin.Context) {
	type request struct {
		RemoteURL string `json:"remote_url" binding:"required"`
		Token     string `json:"token"`
	}
	var req request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	jobID := uuid.New().String()
	// TODO: enqueue background export job using job queue
	c.JSON(http.StatusAccepted, gin.H{"job_id": jobID})
}

// GetExportStatus returns the status of an export job.
func (h *ExportHandlers) GetExportStatus(c *gin.Context) {
	jobID := c.Param("job_id")
	// TODO: fetch job status from job queue or job store
	c.JSON(http.StatusOK, gin.H{"job_id": jobID, "status": "pending"})
}
