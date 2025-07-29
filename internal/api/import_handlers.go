package api

import (
	"encoding/json"
	"net/http"

	"github.com/a5c-ai/hub/internal/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ImportHandlers handles repository import endpoints and enqueues jobs.
type ImportHandlers struct {
	db *db.Database
}

// NewImportHandlers creates a new ImportHandlers with database access.
func NewImportHandlers(database *db.Database) *ImportHandlers {
	return &ImportHandlers{db: database}
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
	// enqueue background import job using job queue (database fallback)
	data, err := json.Marshal(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal job data"})
		return
	}
	if err := h.db.DB.Exec(
		`INSERT INTO job_queue (job_id, workflow_run_id, data) VALUES (?, ?, ?);`,
		jobID, jobID, data,
	).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to enqueue job"})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"job_id": jobID})
}

// GetImportStatus returns the status of an import job.
func (h *ImportHandlers) GetImportStatus(c *gin.Context) {
	jobID := c.Param("job_id")
	// fetch job status from job queue (database fallback)
	var status string
	err := h.db.DB.Raw(
		"SELECT status FROM job_queue WHERE job_id = ?;",
		jobID,
	).Row().Scan(&status)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"job_id": jobID, "status": status})
}

// ExportHandlers handles repository export endpoints and enqueues jobs.
type ExportHandlers struct {
	db *db.Database
}

// NewExportHandlers creates a new ExportHandlers with database access.
func NewExportHandlers(database *db.Database) *ExportHandlers {
	return &ExportHandlers{db: database}
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
	// enqueue background export job using job queue (database fallback)
	data, err := json.Marshal(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal job data"})
		return
	}
	if err := h.db.DB.Exec(
		`INSERT INTO job_queue (job_id, workflow_run_id, data) VALUES (?, ?, ?);`,
		jobID, jobID, data,
	).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to enqueue job"})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"job_id": jobID})
}

// GetExportStatus returns the status of an export job.

func (h *ExportHandlers) GetExportStatus(c *gin.Context) {
	jobID := c.Param("job_id")
	// fetch job status from job queue (database fallback)
	var status string
	err := h.db.DB.Raw(
		"SELECT status FROM job_queue WHERE job_id = ?;",
		jobID,
	).Row().Scan(&status)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"job_id": jobID, "status": status})
}
