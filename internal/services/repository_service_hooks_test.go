package services

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/a5c-ai/hub/internal/git"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)
	return db
}

func TestSetupRepositoryHooks(t *testing.T) {
	tmp := t.TempDir()
	repoPath := filepath.Join(tmp, "repo.git")
	assert.NoError(t, os.MkdirAll(repoPath, 0755))

	logger := logrus.New()
	db := setupTestDB(t)
	gitService := git.NewGitService(logger)
	svc := NewRepositoryService(db, gitService, logger, tmp).(*repositoryService)

	err := svc.setupRepositoryHooks(context.Background(), repoPath)
	assert.NoError(t, err)

	hooksDir := filepath.Join(repoPath, "hooks")
	preDir := filepath.Join(hooksDir, "pre-receive.d")
	postDir := filepath.Join(hooksDir, "post-receive.d")
	assert.DirExists(t, preDir)
	assert.DirExists(t, postDir)

	preWrapper := filepath.Join(hooksDir, "pre-receive")
	assert.FileExists(t, preWrapper)
	content, err := os.ReadFile(preWrapper)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "for hook in "+preDir+"/*.sh")

	postWrapper := filepath.Join(hooksDir, "post-receive")
	assert.FileExists(t, postWrapper)
	contentPost, err := os.ReadFile(postWrapper)
	assert.NoError(t, err)
	assert.Contains(t, string(contentPost), "while read oldrev newrev ref; do")
}
