package ssh

import (
	"context"

	"github.com/a5c-ai/hub/internal/models"
	"github.com/a5c-ai/hub/internal/services"
	"github.com/google/uuid"
)

// repositoryServiceAdapter adapts the existing repository service to the SSH interface
type repositoryServiceAdapter struct {
	repoService services.RepositoryService
}

// NewRepositoryServiceAdapter creates a new adapter
func NewRepositoryServiceAdapter(repoService services.RepositoryService) RepositoryService {
	return &repositoryServiceAdapter{
		repoService: repoService,
	}
}

// Get retrieves a repository by owner and name
func (a *repositoryServiceAdapter) Get(ctx context.Context, owner, name string) (*models.Repository, error) {
	return a.repoService.Get(ctx, owner, name)
}

// GetRepositoryPath returns the filesystem path for a repository
func (a *repositoryServiceAdapter) GetRepositoryPath(ctx context.Context, repoID uuid.UUID) (string, error) {
	return a.repoService.GetRepositoryPath(ctx, repoID)
}
