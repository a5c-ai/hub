package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/a5c-ai/hub/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// setupTestContext initializes a gin Context and ResponseRecorder for testing
func setupTestContext(method, path, body string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	ctx.Request = req
	return ctx, rec
}

func TestGetEmailConfig_AccessControl(t *testing.T) {
	h := NewAdminEmailHandlers(nil, &config.Config{}, logrus.New())
	uid := uuid.New()

	// Unauthorized: not admin
	ctx, rec := setupTestContext(http.MethodGet, "/", "")
	ctx.Set("user_id", uid)
	ctx.Set("is_admin", false)
	h.GetEmailConfig(ctx)
	assert.Equal(t, http.StatusForbidden, rec.Code)

	// Unauthorized: not authenticated
	ctx, rec = setupTestContext(http.MethodGet, "/", "")
	h.GetEmailConfig(ctx)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	// Authorized: admin
	ctx, rec = setupTestContext(http.MethodGet, "/", "")
	ctx.Set("user_id", uid)
	ctx.Set("is_admin", true)
	h.GetEmailConfig(ctx)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestUpdateEmailConfig_AccessControl(t *testing.T) {
	h := NewAdminEmailHandlers(nil, &config.Config{}, logrus.New())
	uid := uuid.New()
	body := `{}`

	// Unauthorized: not admin
	ctx, rec := setupTestContext(http.MethodPut, "/", body)
	ctx.Set("user_id", uid)
	ctx.Set("is_admin", false)
	h.UpdateEmailConfig(ctx)
	assert.Equal(t, http.StatusForbidden, rec.Code)

	// Unauthorized: not authenticated
	ctx, rec = setupTestContext(http.MethodPut, "/", body)
	h.UpdateEmailConfig(ctx)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	// Authorized: admin
	ctx, rec = setupTestContext(http.MethodPut, "/", body)
	ctx.Set("user_id", uid)
	ctx.Set("is_admin", true)
	h.UpdateEmailConfig(ctx)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestTestEmailConfig_AccessControl(t *testing.T) {
	h := NewAdminEmailHandlers(nil, &config.Config{}, logrus.New())
	uid := uuid.New()
	body := `{"to":"user@example.com"}`

	// Unauthorized: not admin
	ctx, rec := setupTestContext(http.MethodPost, "/", body)
	ctx.Set("user_id", uid)
	ctx.Set("is_admin", false)
	h.TestEmailConfig(ctx)
	assert.Equal(t, http.StatusForbidden, rec.Code)

	// Unauthorized: not authenticated
	ctx, rec = setupTestContext(http.MethodPost, "/", body)
	h.TestEmailConfig(ctx)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	// Authorized: admin
	ctx, rec = setupTestContext(http.MethodPost, "/", body)
	ctx.Set("user_id", uid)
	ctx.Set("is_admin", true)
	h.TestEmailConfig(ctx)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestGetEmailLogs_AccessControl(t *testing.T) {
	h := NewAdminEmailHandlers(nil, &config.Config{}, logrus.New())
	uid := uuid.New()

	// Unauthorized: not admin
	ctx, rec := setupTestContext(http.MethodGet, "/", "")
	ctx.Set("user_id", uid)
	ctx.Set("is_admin", false)
	h.GetEmailLogs(ctx)
	assert.Equal(t, http.StatusForbidden, rec.Code)

	// Unauthorized: not authenticated
	ctx, rec = setupTestContext(http.MethodGet, "/", "")
	h.GetEmailLogs(ctx)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	// Authorized: admin
	ctx, rec = setupTestContext(http.MethodGet, "/?page=1&per_page=10&type=verification", "")
	ctx.Set("user_id", uid)
	ctx.Set("is_admin", true)
	h.GetEmailLogs(ctx)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestGetEmailHealth_AccessControl(t *testing.T) {
	h := NewAdminEmailHandlers(nil, &config.Config{}, logrus.New())
	uid := uuid.New()

	// Unauthorized: not admin
	ctx, rec := setupTestContext(http.MethodGet, "/", "")
	ctx.Set("user_id", uid)
	ctx.Set("is_admin", false)
	h.GetEmailHealth(ctx)
	assert.Equal(t, http.StatusForbidden, rec.Code)

	// Unauthorized: not authenticated
	ctx, rec = setupTestContext(http.MethodGet, "/", "")
	h.GetEmailHealth(ctx)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	// Authorized: admin
	ctx, rec = setupTestContext(http.MethodGet, "/", "")
	ctx.Set("user_id", uid)
	ctx.Set("is_admin", true)
	h.GetEmailHealth(ctx)
	assert.Equal(t, http.StatusOK, rec.Code)
}
