package api

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"github.com/a5c-ai/hub/internal/config"
	"github.com/a5c-ai/hub/internal/storage"
)

// LFSHandlers provides Git LFS endpoint handlers.
// Implements basic Git LFS Batch, Upload, Download, and Verify using storage backends.
type LFSHandlers struct {
	backend storage.Backend
}

// NewLFSHandlers creates a new LFSHandlers with the given LFS config and repository base path.
// repoBasePath is used as the root for filesystem-based LFS storage.
func NewLFSHandlers(cfg config.LFS, repoBasePath string) (*LFSHandlers, error) {
	// prepare storage configuration for LFS
	var stCfg storage.Config
	stCfg.Backend = cfg.Backend
	// Azure settings
	stCfg.Azure.AccountName = cfg.Azure.AccountName
	stCfg.Azure.AccountKey = cfg.Azure.AccountKey
	stCfg.Azure.ContainerName = cfg.Azure.ContainerName
	// filesystem backend uses an lfs subdirectory under repoBasePath
	stCfg.Filesystem.BasePath = filepath.Join(repoBasePath, "lfs")
	backend, err := storage.NewBackend(stCfg)
	if err != nil {
		return nil, err
	}
	return &LFSHandlers{backend: backend}, nil
}

// Batch handles Git LFS batch API requests (upload/download actions).
func (h *LFSHandlers) Batch(c *gin.Context) {
	var req struct {
		Operation string `json:"operation"`
		Objects   []struct {
			Oid  string `json:"oid"`
			Size int64  `json:"size"`
		} `json:"objects"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid batch request"})
		return
	}
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	base := fmt.Sprintf("%s://%s/api/v1/git-lfs/objects", scheme, c.Request.Host)
	// build response objects
	respObjs := make([]gin.H, 0, len(req.Objects))
	for _, obj := range req.Objects {
		entry := gin.H{"oid": obj.Oid, "size": obj.Size, "authenticated": false}
		actions := gin.H{}
		switch req.Operation {
		case "upload":
			actions["upload"] = gin.H{"href": fmt.Sprintf("%s/%s", base, obj.Oid)}
		case "download":
			exists, err := h.backend.Exists(c, obj.Oid)
			if err != nil || !exists {
				entry["error"] = gin.H{"code": http.StatusNotFound, "message": "object not found"}
			} else {
				actions["download"] = gin.H{"href": fmt.Sprintf("%s/%s", base, obj.Oid)}
			}
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported operation"})
			return
		}
		entry["actions"] = actions
		respObjs = append(respObjs, entry)
	}
	// return LFS batch response
	c.Header("Content-Type", "application/vnd.git-lfs+json")
	c.JSON(http.StatusOK, gin.H{"objects": respObjs})
}

// Upload handles Git LFS object upload requests (POST with raw content).
func (h *LFSHandlers) Upload(c *gin.Context) {
	oid := c.Param("oid")
	// determine size if provided
	size := c.Request.ContentLength
	if err := h.backend.Upload(c, oid, c.Request.Body, size); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "upload failed"})
		return
	}
	c.Status(http.StatusOK)
}

// Download handles Git LFS object download requests (streams raw content).
func (h *LFSHandlers) Download(c *gin.Context) {
	oid := c.Param("oid")
	reader, err := h.backend.Download(c, oid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "object not found"})
		return
	}
	defer reader.Close()
	size, _ := h.backend.GetSize(c, oid)
	c.DataFromReader(http.StatusOK, size, "application/octet-stream", reader, nil)
}

// Verify handles Git LFS object existence verification (HEAD request).
func (h *LFSHandlers) Verify(c *gin.Context) {
	oid := c.Param("oid")
	exists, err := h.backend.Exists(c, oid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "verification failed"})
		return
	}
	if !exists {
		c.Status(http.StatusNotFound)
		return
	}
	c.Status(http.StatusOK)
}
