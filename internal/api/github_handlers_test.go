package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/a5c-ai/hub/internal/services"
)

func TestInitiateGitHubCreate_NotImplemented(t *testing.T) {
	router := gin.New()
	githubService := services.NewGitHubService(logrus.New())
	handlers := NewGitHubHandlers(githubService, logrus.New())
	router.POST("/api/v1/repositories/github/create", handlers.InitiateGitHubCreate)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/repositories/github/create", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusNotImplemented {
		t.Errorf("expected 501 Not Implemented, got %d", w.Code)
	}
}

func TestInitiateGitHubImport_NotImplemented(t *testing.T) {
	router := gin.New()
	githubService := services.NewGitHubService(logrus.New())
	handlers := NewGitHubHandlers(githubService, logrus.New())
	router.POST("/api/v1/repositories/github/import", handlers.InitiateGitHubImport)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/repositories/github/import", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusNotImplemented {
		t.Errorf("expected 501 Not Implemented, got %d", w.Code)
	}
}
