package storage

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilesystemBackend_UploadDownload(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "filesystem_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create filesystem backend
	backend, err := NewFilesystemBackend(FilesystemConfig{
		BasePath: tempDir,
	})
	require.NoError(t, err)

	ctx := context.Background()
	testPath := "test/file.txt"
	testContent := "Hello, World!"
	
	// Test upload
	reader := strings.NewReader(testContent)
	err = backend.Upload(ctx, testPath, reader, int64(len(testContent)))
	require.NoError(t, err)

	// Test download
	downloadReader, err := backend.Download(ctx, testPath)
	require.NoError(t, err)
	defer downloadReader.Close()

	downloadedContent, err := io.ReadAll(downloadReader)
	require.NoError(t, err)
	
	assert.Equal(t, testContent, string(downloadedContent))
}

func TestFilesystemBackend_Exists(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "filesystem_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create filesystem backend
	backend, err := NewFilesystemBackend(FilesystemConfig{
		BasePath: tempDir,
	})
	require.NoError(t, err)

	ctx := context.Background()
	testPath := "test/file.txt"
	testContent := "Hello, World!"

	// Test file doesn't exist initially
	exists, err := backend.Exists(ctx, testPath)
	require.NoError(t, err)
	assert.False(t, exists)

	// Upload file
	reader := strings.NewReader(testContent)
	err = backend.Upload(ctx, testPath, reader, int64(len(testContent)))
	require.NoError(t, err)

	// Test file exists after upload
	exists, err = backend.Exists(ctx, testPath)
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestFilesystemBackend_Delete(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "filesystem_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create filesystem backend
	backend, err := NewFilesystemBackend(FilesystemConfig{
		BasePath: tempDir,
	})
	require.NoError(t, err)

	ctx := context.Background()
	testPath := "test/file.txt"
	testContent := "Hello, World!"

	// Upload file
	reader := strings.NewReader(testContent)
	err = backend.Upload(ctx, testPath, reader, int64(len(testContent)))
	require.NoError(t, err)

	// Verify file exists
	exists, err := backend.Exists(ctx, testPath)
	require.NoError(t, err)
	assert.True(t, exists)

	// Delete file
	err = backend.Delete(ctx, testPath)
	require.NoError(t, err)

	// Verify file no longer exists
	exists, err = backend.Exists(ctx, testPath)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestFilesystemBackend_GetSize(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "filesystem_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create filesystem backend
	backend, err := NewFilesystemBackend(FilesystemConfig{
		BasePath: tempDir,
	})
	require.NoError(t, err)

	ctx := context.Background()
	testPath := "test/file.txt"
	testContent := "Hello, World!"

	// Upload file
	reader := strings.NewReader(testContent)
	err = backend.Upload(ctx, testPath, reader, int64(len(testContent)))
	require.NoError(t, err)

	// Get file size
	size, err := backend.GetSize(ctx, testPath)
	require.NoError(t, err)
	assert.Equal(t, int64(len(testContent)), size)
}

func TestFilesystemBackend_List(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "filesystem_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create filesystem backend
	backend, err := NewFilesystemBackend(FilesystemConfig{
		BasePath: tempDir,
	})
	require.NoError(t, err)

	ctx := context.Background()

	// Upload multiple files
	files := map[string]string{
		"artifacts/file1.txt": "Content 1",
		"artifacts/file2.txt": "Content 2",
		"logs/log1.txt":       "Log content 1",
		"other/other.txt":     "Other content",
	}

	for path, content := range files {
		reader := strings.NewReader(content)
		err = backend.Upload(ctx, path, reader, int64(len(content)))
		require.NoError(t, err)
	}

	// List files with "artifacts/" prefix
	artifactFiles, err := backend.List(ctx, "artifacts/")
	require.NoError(t, err)
	assert.Len(t, artifactFiles, 2)
	assert.Contains(t, artifactFiles, "artifacts/file1.txt")
	assert.Contains(t, artifactFiles, "artifacts/file2.txt")

	// List files with "logs/" prefix
	logFiles, err := backend.List(ctx, "logs/")
	require.NoError(t, err)
	assert.Len(t, logFiles, 1)
	assert.Contains(t, logFiles, "logs/log1.txt")
}