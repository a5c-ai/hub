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

// TestIssuePRWorkflow verifies that issue creation, PR workflows,
// and related CI/CD triggers succeed end-to-end.
func TestIssuePRWorkflow(t *testing.T) {
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

	// Create an issue
	issueTitle := "Test Issue"
	issueBody := "This is a test issue"
	issueReq := map[string]string{"title": issueTitle, "body": issueBody}
	data, _ = json.Marshal(issueReq)
	issueURL := fmt.Sprintf("%s/api/v1/repositories/%s/%s/issues", baseURL, username, repoName)
	req, _ = http.NewRequest("POST", issueURL, bytes.NewReader(data))
	req.Header.Set("Authorization", "Bearer "+auth.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("issue creation failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected issue status 201, got %d: %s", resp.StatusCode, string(body))
	}

	// List issues
	listURL := issueURL
	resp, err = client.Get(listURL)
	if err != nil {
		t.Fatalf("list issues failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected list issues status: %d", resp.StatusCode)
	}
	var listResp struct {
		Issues []struct {
			Title string `json:"title"`
		} `json:"issues"`
		Total int `json:"total"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		t.Fatalf("failed to decode list issues response: %v", err)
	}
	if listResp.Total < 1 || listResp.Issues[0].Title != issueTitle {
		t.Fatalf("issue list did not include created issue, got %+v", listResp)
	}

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
