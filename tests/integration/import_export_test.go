//go:build integration
// +build integration

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
)

// waitForServer pings health endpoint until ready or timeout.
func waitForServer(t *testing.T, baseURL string) {
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get(fmt.Sprintf("%s/health", baseURL))
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(time.Second)
	}
	t.Fatalf("server did not respond with healthy status within timeout")
}

// setupUser registers and logs in a new user, returning client, token, and username.
func setupUser(t *testing.T, baseURL string) (*http.Client, string, string) {
	client := &http.Client{Timeout: 10 * time.Second}
	username := fmt.Sprintf("user_%d", time.Now().UnixNano())
	email := username + "@example.com"
	password := "password1234"
	// register
	reg := map[string]string{"username": username, "email": email, "password": password, "full_name": "Job Tester"}
	data, _ := json.Marshal(reg)
	resp, err := client.Post(baseURL+"/api/v1/auth/register", "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("failed to register user: %v", err)
	}
	resp.Body.Close()
	// login
	login := map[string]string{"email": email, "password": password}
	data, _ = json.Marshal(login)
	resp, err = client.Post(baseURL+"/api/v1/auth/login", "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("failed to login: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("login failed: status %d, body %s", resp.StatusCode, string(body))
	}
	var auth struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&auth); err != nil {
		t.Fatalf("failed to parse login response: %v", err)
	}
	return client, auth.AccessToken, username
}

// TestImportJobQueue verifies enqueuing and status retrieval of import jobs.
func TestImportJobQueue(t *testing.T) {
	baseURL := os.Getenv("API_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	waitForServer(t, baseURL)
	client, token, _ := setupUser(t, baseURL)
	// initiate import
	reqBody := map[string]string{"url": "https://github.com/example/repo.git", "token": "token123"}
	data, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", baseURL+"/api/v1/repositories/import", bytes.NewReader(data))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to initiate import: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status 202 Accepted, got %d: %s", resp.StatusCode, string(body))
	}
	var initResp struct {
		JobID string `json:"job_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&initResp); err != nil {
		t.Fatalf("failed to decode import response: %v", err)
	}
	if _, err := uuid.Parse(initResp.JobID); err != nil {
		t.Fatalf("invalid job_id returned: %v", err)
	}
	// get import status
	statusURL := fmt.Sprintf("%s/api/v1/repositories/import/%s", baseURL, initResp.JobID)
	req, _ = http.NewRequest("GET", statusURL, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("failed to get import status: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", resp.StatusCode)
	}
	var statusResp struct {
		JobID, Status string `json:"job_id" json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&statusResp); err != nil {
		t.Fatalf("failed to decode status response: %v", err)
	}
	if statusResp.Status != "pending" {
		t.Errorf("expected status 'pending', got '%s'", statusResp.Status)
	}
}

// TestExportJobQueue verifies enqueuing and status retrieval of export jobs.
func TestExportJobQueue(t *testing.T) {
	baseURL := os.Getenv("API_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	waitForServer(t, baseURL)
	client, token, username := setupUser(t, baseURL)
	// create repository
	repoName := fmt.Sprintf("repo_%d", time.Now().UnixNano())
	createReq := map[string]interface{}{
		"name":                   repoName,
		"visibility":             "public",
		"has_wiki":               true,
		"has_downloads":          true,
		"allow_merge_commit":     true,
		"allow_squash_merge":     true,
		"allow_rebase_merge":     true,
		"delete_branch_on_merge": false,
		"auto_init":              true,
	}
	data, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", baseURL+"/api/v1/repositories", bytes.NewReader(data))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("repository creation failed: %d %s", resp.StatusCode, string(body))
	}
	// initiate export
	reqBody := map[string]string{"remote_url": "https://github.com/example/remote.git", "token": "token123"}
	data, _ := json.Marshal(reqBody)
	exportURL := fmt.Sprintf("%s/api/v1/repositories/%s/%s/export", baseURL, username, repoName)
	req, _ = http.NewRequest("POST", exportURL, bytes.NewReader(data))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("failed to initiate export: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status 202 Accepted, got %d: %s", resp.StatusCode, string(body))
	}
	var initResp struct {
		JobID string `json:"job_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&initResp); err != nil {
		t.Fatalf("failed to decode export response: %v", err)
	}
	if _, err := uuid.Parse(initResp.JobID); err != nil {
		t.Fatalf("invalid job_id returned: %v", err)
	}
	// get export status
	statusURL := fmt.Sprintf("%s/api/v1/repositories/%s/%s/export/%s", baseURL, username, repoName, initResp.JobID)
	req, _ = http.NewRequest("GET", statusURL, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("failed to get export status: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", resp.StatusCode)
	}
	var statusResp struct {
		JobID, Status string `json:"job_id" json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&statusResp); err != nil {
		t.Fatalf("failed to decode status response: %v", err)
	}
	if statusResp.Status != "pending" {
		t.Errorf("expected status 'pending', got '%s'", statusResp.Status)
	}
}
