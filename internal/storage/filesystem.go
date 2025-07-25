package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FilesystemBackend implements the Backend interface using local filesystem
type FilesystemBackend struct {
	basePath string
}

// NewFilesystemBackend creates a new filesystem storage backend
func NewFilesystemBackend(config FilesystemConfig) (*FilesystemBackend, error) {
	if config.BasePath == "" {
		return nil, fmt.Errorf("base path is required for filesystem backend")
	}

	// Ensure the base directory exists
	if err := os.MkdirAll(config.BasePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	return &FilesystemBackend{
		basePath: config.BasePath,
	}, nil
}

// Upload uploads a file to the filesystem
func (f *FilesystemBackend) Upload(ctx context.Context, path string, reader io.Reader, size int64) error {
	fullPath := filepath.Join(f.basePath, path)
	
	// Ensure the directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Create the file
	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", fullPath, err)
	}
	defer file.Close()

	// Copy the data
	written, err := io.Copy(file, reader)
	if err != nil {
		// Clean up on error
		os.Remove(fullPath)
		return fmt.Errorf("failed to write file %s: %w", fullPath, err)
	}

	// Verify size if provided
	if size > 0 && written != size {
		os.Remove(fullPath)
		return fmt.Errorf("size mismatch: expected %d bytes, wrote %d bytes", size, written)
	}

	return nil
}

// Download downloads a file from the filesystem
func (f *FilesystemBackend) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	fullPath := filepath.Join(f.basePath, path)
	
	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", path)
		}
		return nil, fmt.Errorf("failed to open file %s: %w", fullPath, err)
	}

	return file, nil
}

// Delete deletes a file from the filesystem
func (f *FilesystemBackend) Delete(ctx context.Context, path string) error {
	fullPath := filepath.Join(f.basePath, path)
	
	err := os.Remove(fullPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file %s: %w", fullPath, err)
	}

	return nil
}

// Exists checks if a file exists in the filesystem
func (f *FilesystemBackend) Exists(ctx context.Context, path string) (bool, error) {
	fullPath := filepath.Join(f.basePath, path)
	
	_, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check file existence %s: %w", fullPath, err)
	}

	return true, nil
}

// GetSize returns the size of a file in bytes
func (f *FilesystemBackend) GetSize(ctx context.Context, path string) (int64, error) {
	fullPath := filepath.Join(f.basePath, path)
	
	stat, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, fmt.Errorf("file not found: %s", path)
		}
		return 0, fmt.Errorf("failed to get file size %s: %w", fullPath, err)
	}

	return stat.Size(), nil
}

// GetLastModified returns the last modified time of a file
func (f *FilesystemBackend) GetLastModified(ctx context.Context, path string) (time.Time, error) {
	fullPath := filepath.Join(f.basePath, path)
	
	stat, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return time.Time{}, fmt.Errorf("file not found: %s", path)
		}
		return time.Time{}, fmt.Errorf("failed to get file modification time %s: %w", fullPath, err)
	}

	return stat.ModTime(), nil
}

// List lists files with the given prefix
func (f *FilesystemBackend) List(ctx context.Context, prefix string) ([]string, error) {
	var files []string
	
	err := filepath.Walk(f.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if the path starts with the prefix
		relPath, err := filepath.Rel(f.basePath, path)
		if err != nil {
			return err
		}

		// Convert to forward slashes for consistency
		relPath = strings.ReplaceAll(relPath, "\\", "/")
		
		if strings.HasPrefix(relPath, prefix) {
			files = append(files, relPath)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list files with prefix %s: %w", prefix, err)
	}

	return files, nil
}

// GetURL returns a presigned URL for downloading (not supported for filesystem)
func (f *FilesystemBackend) GetURL(ctx context.Context, path string, expiry time.Duration) (string, error) {
	return "", fmt.Errorf("presigned URLs not supported for filesystem backend")
}