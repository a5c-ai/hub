//go:build integration
// +build integration

package integration

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
)

// TestHealthEndpoint verifies that the HTTP health endpoint responds with status 200.
func TestHealthEndpoint(t *testing.T) {
	baseURL := os.Getenv("API_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	// Retry GET /api/health for up to 30 seconds
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get(fmt.Sprintf("%s/api/health", baseURL))
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(time.Second)
	}
	t.Fatalf("health endpoint did not return 200 within timeout")
}
