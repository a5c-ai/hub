package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ImportHandlers handles repository import endpoints.
type ImportHandlers struct{}

// NewImportHandlers creates a new ImportHandlers.
func NewImportHandlers() *ImportHandlers {
	return &ImportHandlers{}
}

// InitiateImport starts an import job from an external Git service.
func (h *ImportHandlers) InitiateImport(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Import functionality not implemented"})
}

// GetImportStatus returns the status of an import job.
func (h *ImportHandlers) GetImportStatus(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Import functionality not implemented"})
}

// ExportHandlers handles repository export endpoints.
type ExportHandlers struct{}

// NewExportHandlers creates a new ExportHandlers.
func NewExportHandlers() *ExportHandlers {
	return &ExportHandlers{}
}

// InitiateExport starts an export job to an external Git service.
func (h *ExportHandlers) InitiateExport(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Export functionality not implemented"})
}

// GetExportStatus returns the status of an export job.
func (h *ExportHandlers) GetExportStatus(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Export functionality not implemented"})
}
