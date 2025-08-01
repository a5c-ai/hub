package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/a5c-ai/hub/internal/auth"
	"github.com/a5c-ai/hub/internal/config"
	"github.com/a5c-ai/hub/internal/models"
	"github.com/a5c-ai/hub/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// fakeRepoService implements RepositoryService for testing selected methods
// fakeRepoService implements RepositoryService for testing selected methods
type fakeRepoService struct {
	services.RepositoryService
	repo *models.Repository
	path string
}

// Get returns the configured repository
func (f *fakeRepoService) Get(ctx context.Context, owner, name string) (*models.Repository, error) {
	return f.repo, nil
}

// GetRepositoryPath returns the configured filesystem path
func (f *fakeRepoService) GetRepositoryPath(ctx context.Context, id uuid.UUID) (string, error) {
	return f.path, nil
}

func setupHandler(t *testing.T, repo *models.Repository, makePath bool) (*GitHandlers, string) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	jwtMgr := auth.NewJWTManager(cfg.JWT)
	tmpDir := t.TempDir()
	if makePath {
		repoPath := filepath.Join(tmpDir, repo.ID.String())
		if err := os.Mkdir(repoPath, 0755); err != nil {
			t.Fatalf("failed to create temp repo path: %v", err)
		}
		tmpDir = repoPath
	}
	fakeSvc := &fakeRepoService{repo: repo, path: tmpDir}
	logger := logrus.New()
	handler := NewGitHandlers(fakeSvc, logger, jwtMgr)
	return handler, tmpDir
}

func TestUploadPack_PrivateRepo_Auth(t *testing.T) {
	repo := &models.Repository{ID: uuid.New(), Visibility: models.VisibilityPrivate}
	handler, _ := setupHandler(t, repo, false)
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name     string
		header   string
		wantCode int
	}{
		{"no auth", "", http.StatusUnauthorized},
		{"bad format", "Basic foo", http.StatusUnauthorized},
		{"invalid token", "Bearer invalid", http.StatusUnauthorized},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/owner/repo/git-upload-pack", strings.NewReader(""))
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			handler.UploadPack(c)
			if w.Code != tt.wantCode {
				t.Errorf("got code %d, want %d", w.Code, tt.wantCode)
			}
		})
	}
}

func TestUploadPack_PrivateRepo_ValidToken(t *testing.T) {
	user := &models.User{ID: uuid.New(), Username: "u", Email: "e", IsAdmin: false}
	repo := &models.Repository{ID: uuid.New(), Visibility: models.VisibilityPrivate}
	handler, _ := setupHandler(t, repo, false)
	// generate valid token
	token, err := handler.jwtManager.GenerateToken(user)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}
	gin.SetMode(gin.TestMode)
	req := httptest.NewRequest(http.MethodPost, "/owner/repo/git-upload-pack", strings.NewReader(""))
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	handler.UploadPack(c)
	if w.Code == http.StatusUnauthorized {
		t.Errorf("expected auth success, got unauthorized")
	}
}

func TestReceivePack_Auth(t *testing.T) {
	repo := &models.Repository{ID: uuid.New(), Visibility: models.VisibilityPublic}
	handler, _ := setupHandler(t, repo, false)
	gin.SetMode(gin.TestMode)
	req := httptest.NewRequest(http.MethodPost, "/owner/repo/git-receive-pack", strings.NewReader(""))
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	handler.ReceivePack(c)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("got code %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestReceivePack_ValidToken(t *testing.T) {
	user := &models.User{ID: uuid.New(), Username: "u", Email: "e", IsAdmin: false}
	repo := &models.Repository{ID: uuid.New(), Visibility: models.VisibilityPublic}
	handler, _ := setupHandler(t, repo, false)
	token, err := handler.jwtManager.GenerateToken(user)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}
	gin.SetMode(gin.TestMode)
	req := httptest.NewRequest(http.MethodPost, "/owner/repo/git-receive-pack", strings.NewReader(""))
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	handler.ReceivePack(c)
	if w.Code == http.StatusUnauthorized {
		t.Errorf("expected auth success, got unauthorized")
	}
}
