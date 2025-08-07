package services

import (
	"github.com/sirupsen/logrus"
)

// GitHubService provides operations for interacting with GitHub API for repository creation and import.
type GitHubService struct {
	logger *logrus.Logger
}

// NewGitHubService creates a new GitHubService.
func NewGitHubService(logger *logrus.Logger) *GitHubService {
	return &GitHubService{logger: logger}
}
