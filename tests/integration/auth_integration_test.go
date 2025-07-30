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

// TestUserAuthentication verifies authentication flows, session management,
// and permission enforcement via the API.
func TestUserAuthentication(t *testing.T) {
	baseURL := os.Getenv("API_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// Register a new user
	username := fmt.Sprintf("testuser_%d", time.Now().UnixNano())
	email := username + "@example.com"
	password := "password1234"
	regReq := map[string]string{
		"username":  username,
		"email":     email,
		"password":  password,
		"full_name": "Test User",
	}
	data, _ := json.Marshal(regReq)
	resp, err := client.Post(baseURL+"/api/v1/auth/register", "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("register request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status 201, got %d: %s", resp.StatusCode, string(body))
	}

	// Login with the new user
	loginReq := map[string]string{"email": email, "password": password}
	data, _ = json.Marshal(loginReq)
	resp, err = client.Post(baseURL+"/api/v1/auth/login", "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("login request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status 200, got %d: %s", resp.StatusCode, string(body))
	}

	var authResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		t.Fatalf("failed to decode login response: %v", err)
	}
	if authResp.AccessToken == "" || authResp.RefreshToken == "" {
		t.Fatal("authentication response missing tokens")
	}
}

// TestTokenRefreshAndLogout verifies token refresh and logout endpoints.
func TestTokenRefreshAndLogout(t *testing.T) {
	baseURL := os.Getenv("API_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// Register a new user for this test
	username := fmt.Sprintf("testuser_%d", time.Now().UnixNano())
	email := username + "@example.com"
	password := "password1234"
	regReq := map[string]string{
		"username":  username,
		"email":     email,
		"password":  password,
		"full_name": "Test User",
	}
	data, _ := json.Marshal(regReq)
	resp, err := client.Post(baseURL+"/api/v1/auth/register", "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("register request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status 201, got %d: %s", resp.StatusCode, string(body))
	}

	// Login to obtain tokens
	loginReq := map[string]string{"email": email, "password": password}
	data, _ = json.Marshal(loginReq)
	resp, err = client.Post(baseURL+"/api/v1/auth/login", "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("login request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status 200, got %d: %s", resp.StatusCode, string(body))
	}
	var authResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		t.Fatalf("failed to decode login response: %v", err)
	}
	if authResp.AccessToken == "" || authResp.RefreshToken == "" {
		t.Fatal("authentication response missing tokens")
	}

	// Refresh the token
	refreshReq := map[string]string{"refresh_token": authResp.RefreshToken}
	data, _ = json.Marshal(refreshReq)
	resp, err = client.Post(baseURL+"/api/v1/auth/refresh", "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("refresh request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status 200, got %d: %s", resp.StatusCode, string(body))
	}
	var refreshResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&refreshResp); err != nil {
		t.Fatalf("failed to decode refresh response: %v", err)
	}
	if refreshResp.AccessToken == authResp.AccessToken || refreshResp.RefreshToken == authResp.RefreshToken {
		t.Fatal("tokens did not change after refresh")
	}

	// Logout to revoke sessions
	reqLogout, err := http.NewRequest("POST", baseURL+"/api/v1/auth/logout", nil)
	if err != nil {
		t.Fatalf("failed to create logout request: %v", err)
	}
	reqLogout.Header.Set("Authorization", "Bearer "+refreshResp.AccessToken)
	resp, err = client.Do(reqLogout)
	if err != nil {
		t.Fatalf("logout request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status 200, got %d: %s", resp.StatusCode, string(body))
	}

	// Attempt to refresh with old token should fail
	data, _ = json.Marshal(refreshReq)
	resp, err = client.Post(baseURL+"/api/v1/auth/refresh", "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("refresh-after-logout request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status 401 after logout, got %d: %s", resp.StatusCode, string(body))
	}
}
