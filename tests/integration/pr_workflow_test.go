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

// TestPRWorkflow verifies that PR workflows
// and related CI/CD triggers succeed end-to-end.
func TestPRWorkflow(t *testing.T) {
	baseURL := os.Getenv("API_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	client := &http.Client{Timeout: 15 * time.Second}

	// Register and login
	username := fmt.Sprintf("pruser_%d", time.Now().UnixNano())
	email := username + "@example.com"
	password := "password1234"
	reg := map[string]string{"username": username, "email": email, "password": password, "full_name": "PR Tester"}
	data, _ := json.Marshal(reg)
	resp, err := client.Post(baseURL+"/api/v1/auth/register", "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	resp.Body.Close()

	login := map[string]string{"email": email, "password": password}
	data, _ = json.Marshal(login)
	resp, err = client.Post(baseURL+"/api/v1/auth/login", "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("login failed: %v", err)
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
		t.Fatalf("login response decode failed: %v", err)
	}

	// Create repository for PR workflow
	repoName := fmt.Sprintf("pr-repo-%d", time.Now().UnixNano())
	createReq := map[string]interface{}{"name": repoName, "visibility": "public", "auto_init": true}
	data, _ = json.Marshal(createReq)
	req, _ := http.NewRequest("POST", baseURL+"/api/v1/repositories", bytes.NewReader(data))
	req.Header.Set("Authorization", "Bearer "+auth.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("repo creation failed: %v", err)
	}
	resp.Body.Close()

	// Create a pull request
	prTitle := "Test Pull Request"
	prBody := "This is a test PR"
	prReq := map[string]string{"title": prTitle, "body": prBody, "head": "main", "base": "main"}
	data, _ = json.Marshal(prReq)
	prURL := fmt.Sprintf("%s/api/v1/repositories/%s/%s/pulls", baseURL, username, repoName)
	req, _ = http.NewRequest("POST", prURL, bytes.NewReader(data))
	req.Header.Set("Authorization", "Bearer "+auth.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("pull request creation failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected PR status 201, got %d: %s", resp.StatusCode, string(body))
	}

	// List pull requests
	resp, err = client.Get(prURL)
	if err != nil {
		t.Fatalf("list PRs failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected list PRs status: %d", resp.StatusCode)
	}
	var prList struct {
		TotalCount int `json:"total_count"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&prList); err != nil {
		t.Fatalf("failed to decode list PRs response: %v", err)
	}
	if prList.TotalCount < 1 {
		t.Fatalf("expected at least one PR, got %d", prList.TotalCount)
	}
}
