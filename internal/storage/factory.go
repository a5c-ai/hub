package storage

import (
	"fmt"
	"strings"
)

// NewBackend creates a new storage backend based on the configuration
func NewBackend(config Config) (Backend, error) {
	switch strings.ToLower(config.Backend) {
	case "filesystem", "local", "":
		return NewFilesystemBackend(config.Filesystem)
	case "azure", "azureblob":
		return NewAzureBackend(config.Azure)
	case "s3", "aws":
		return NewS3Backend(config.S3)
	default:
		return nil, fmt.Errorf("unsupported storage backend: %s", config.Backend)
	}
}
