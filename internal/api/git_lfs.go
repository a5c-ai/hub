package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// LFSHandlers provides Git LFS endpoint handlers.
// TODO: implement Git LFS support with Azure Blob storage and other backends.
type LFSHandlers struct{}

// NewLFSHandlers creates a new LFSHandlers instance.
func NewLFSHandlers() *LFSHandlers {
	return &LFSHandlers{}
}

// Batch handles Git LFS batch API requests (upload/download actions).
func (h *LFSHandlers) Batch(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Git LFS batch endpoint not implemented yet"})
}

// Upload handles Git LFS object upload requests.
func (h *LFSHandlers) Upload(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Git LFS upload endpoint not implemented yet"})
}

// Download handles Git LFS object download requests.
func (h *LFSHandlers) Download(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Git LFS download endpoint not implemented yet"})
}

// Verify handles Git LFS object existence verification (HEAD request).
func (h *LFSHandlers) Verify(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Git LFS verify endpoint not implemented yet"})
}
