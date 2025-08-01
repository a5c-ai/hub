package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/a5c-ai/hub/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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

// TestGetEmailHealth_Stats verifies that email health stats are computed correctly
func TestGetEmailHealth_Stats(t *testing.T) {
	// Setup in-memory SQLite database and schema for job_queue
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	assert.NoError(t, err)
	// Minimal schema matching columns used in metrics queries
	execSQL := `CREATE TABLE job_queue (
		id TEXT PRIMARY KEY,
		job_id TEXT,
		workflow_run_id TEXT,
		status TEXT,
		data JSON,
		created_at TIMESTAMP,
		updated_at TIMESTAMP
	)`
	err = db.Exec(execSQL).Error
	assert.NoError(t, err)

	// Insert test records: 2 sent today, 1 failed today, 3 sent a week ago
	now := time.Now().UTC()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	older := startOfDay.AddDate(0, 0, -7)
	for i := 0; i < 2; i++ {
		err = db.Exec(
			`INSERT INTO job_queue (id, job_id, workflow_run_id, status, data, created_at, updated_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?)`,
			uuid.New().String(), uuid.New().String(), uuid.New().String(), "completed", "{}", startOfDay.Add(time.Hour), startOfDay.Add(time.Hour),
		).Error
		assert.NoError(t, err)
	}
	for i := 0; i < 1; i++ {
		err = db.Exec(
			`INSERT INTO job_queue (id, job_id, workflow_run_id, status, data, created_at, updated_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?)`,
			uuid.New().String(), uuid.New().String(), uuid.New().String(), "failed", "{}", startOfDay.Add(2*time.Hour), startOfDay.Add(2*time.Hour),
		).Error
		assert.NoError(t, err)
	}
	for i := 0; i < 3; i++ {
		err = db.Exec(
			`INSERT INTO job_queue (id, job_id, workflow_run_id, status, data, created_at, updated_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?)`,
			uuid.New().String(), uuid.New().String(), uuid.New().String(), "completed", "{}", older, older,
		).Error
		assert.NoError(t, err)
	}

	// Invoke handler
	cfg := &config.Config{}
	logger := logrus.New()
	h := NewAdminEmailHandlers(db, cfg, logger)
	ctx, rec := setupTestContext(http.MethodGet, "/", "")
	ctx.Set("user_id", uuid.New())
	ctx.Set("is_admin", true)
	h.GetEmailHealth(ctx)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Parse and validate response
	var resp map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	stats, ok := resp["stats"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, float64(2), stats["emails_sent_today"])
	assert.Equal(t, float64(3), stats["emails_sent_this_week"])
	assert.Equal(t, float64(1), stats["failed_emails_today"])
	assert.InDelta(t, 66.666, stats["success_rate"], 0.01)
	// last_check should be valid RFC3339 timestamp
	_, err = time.Parse(time.RFC3339, resp["last_check"].(string))
	assert.NoError(t, err)
}
