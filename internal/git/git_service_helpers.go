package git

import (
	"context"
	"fmt"

	"github.com/go-git/go-git/v5"
)

// Helper methods for bare repository operations

func (s *gitService) createFileInBareRepo(ctx context.Context, repo *git.Repository, req CreateFileRequest) (*Commit, error) {
	// This is a complex operation for bare repositories
	// For now, return an error indicating it's not supported
	return nil, fmt.Errorf("file operations in bare repositories require more complex implementation")
}

func (s *gitService) updateFileInBareRepo(ctx context.Context, repo *git.Repository, req UpdateFileRequest) (*Commit, error) {
	// This is a complex operation for bare repositories
	// For now, return an error indicating it's not supported
	return nil, fmt.Errorf("file operations in bare repositories require more complex implementation")
}

func (s *gitService) deleteFileInBareRepo(ctx context.Context, repo *git.Repository, req DeleteFileRequest) (*Commit, error) {
	// This is a complex operation for bare repositories
	// For now, return an error indicating it's not supported
	return nil, fmt.Errorf("file operations in bare repositories require more complex implementation")
}