package ssh

import (
	"context"

	"github.com/a5c-ai/hub/internal/models"
	"github.com/google/uuid"
)

// RepositoryService defines the interface needed by SSH server for repository operations
type RepositoryService interface {
	Get(ctx context.Context, owner, name string) (*models.Repository, error)
	GetRepositoryPath(ctx context.Context, repoID uuid.UUID) (string, error)
}
