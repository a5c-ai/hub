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
)

// TestRepositoryDataFlow verifies that git repository actions are correctly persisted
// and metadata is synchronized to the database.
func TestRepositoryDataFlow(t *testing.T) {
	baseURL := os.Getenv("API_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	client := &http.Client{Timeout: 10 * time.Second}

	// Register and login to obtain access token
	username := fmt.Sprintf("repouser_%d", time.Now().UnixNano())
	email := username + "@example.com"
	password := "password1234"
	reg := map[string]string{"username": username, "email": email, "password": password, "full_name": "Repo Tester"}
	data, _ := json.Marshal(reg)
	resp, err := client.Post(baseURL+"/api/v1/auth/register", "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("failed to register user: %v", err)
	}
	resp.Body.Close()

	login := map[string]string{"email": email, "password": password}
	data, _ = json.Marshal(login)
	resp, err = client.Post(baseURL+"/api/v1/auth/login", "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("failed to login: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("unexpected login status: %d %s", resp.StatusCode, string(body))
	}
	var auth struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&auth); err != nil {
		t.Fatalf("failed to parse login response: %v", err)
	}

	// Create a new repository
	repoName := fmt.Sprintf("test-repo-%d", time.Now().UnixNano())
	createReq := map[string]interface{}{
		"name":                   repoName,
		"visibility":             "public",
		"has_issues":             true,
		"has_projects":           true,
		"has_wiki":               true,
		"has_downloads":          true,
		"allow_merge_commit":     true,
		"allow_squash_merge":     true,
		"allow_rebase_merge":     true,
		"delete_branch_on_merge": false,
		"auto_init":              true,
	}
	data, _ = json.Marshal(createReq)
	req, _ := http.NewRequest("POST", baseURL+"/api/v1/repositories", bytes.NewReader(data))
	req.Header.Set("Authorization", "Bearer "+auth.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("repository creation failed: %d %s", resp.StatusCode, string(body))
	}
	var repoResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&repoResp); err != nil {
		t.Fatalf("failed to decode repo creation response: %v", err)
	}
	if repoResp["name"] != repoName {
		t.Fatalf("expected repo name %s, got %v", repoName, repoResp["name"])
	}
	if repoResp["default_branch"] != "main" {
		t.Fatalf("expected default branch 'main', got %v", repoResp["default_branch"])
	}

	// Retrieve repository info
	getURL := fmt.Sprintf("%s/api/v1/repositories/%s/%s", baseURL, username, repoName)
	resp, err = client.Get(getURL)
	if err != nil {
		t.Fatalf("failed to get repository: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected get status: %d", resp.StatusCode)
	}
	var getResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&getResp); err != nil {
		t.Fatalf("failed to decode get response: %v", err)
	}
	if getResp["name"] != repoName {
		t.Fatalf("expected get repo name %s, got %v", repoName, getResp["name"])
	}
}
