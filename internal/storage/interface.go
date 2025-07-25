package storage

import (
	"context"
	"io"
	"time"
)

// Backend defines the interface for artifact storage backends
type Backend interface {
	// Upload uploads a file to the storage backend
	Upload(ctx context.Context, path string, reader io.Reader, size int64) error
	
	// Download downloads a file from the storage backend
	Download(ctx context.Context, path string) (io.ReadCloser, error)
	
	// Delete deletes a file from the storage backend
	Delete(ctx context.Context, path string) error
	
	// Exists checks if a file exists in the storage backend
	Exists(ctx context.Context, path string) (bool, error)
	
	// GetSize returns the size of a file in bytes
	GetSize(ctx context.Context, path string) (int64, error)
	
	// GetLastModified returns the last modified time of a file
	GetLastModified(ctx context.Context, path string) (time.Time, error)
	
	// List lists files with the given prefix
	List(ctx context.Context, prefix string) ([]string, error)
	
	// GetURL returns a presigned URL for downloading (if supported)
	GetURL(ctx context.Context, path string, expiry time.Duration) (string, error)
}

// Config holds the configuration for storage backends
type Config struct {
	Backend       string
	Azure         AzureConfig
	S3            S3Config
	Filesystem    FilesystemConfig
}

type AzureConfig struct {
	AccountName   string
	AccountKey    string
	ContainerName string
	EndpointURL   string
}

type S3Config struct {
	Region          string
	Bucket          string
	AccessKeyID     string
	SecretAccessKey string
	EndpointURL     string
	UseSSL          bool
}

type FilesystemConfig struct {
	BasePath string
}