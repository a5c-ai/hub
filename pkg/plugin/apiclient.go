package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// APIClient enables plugins to call core platform APIs.
type APIClient struct {
	// baseURL is the base endpoint for API requests (e.g., server URL or "/api/v1").
	baseURL string
	token   string
	client  *http.Client
}

// NewAPIClient constructs a new APIClient with the given bearer token.
// It reads the base API URL from the HUB_API_URL environment variable,
// defaulting to "/api/v1" when not set.
func NewAPIClient(token string) *APIClient {
	base := os.Getenv("HUB_API_URL")
	if base == "" {
		base = "/api/v1"
	}
	return &APIClient{baseURL: base, token: token, client: http.DefaultClient}
}

// tokenHeader returns the authorization header value.
func (c *APIClient) tokenHeader() string {
	return "Bearer " + c.token
}

// GetRepository retrieves repository details (stub).
// GetRepository retrieves a repository by owner/name.
func (c *APIClient) GetRepository(ctx context.Context, owner, repo string) (interface{}, error) {
	url := c.baseURL + fmt.Sprintf("/repositories/%s/%s", owner, repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", c.tokenHeader())

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

// ListRepositories lists repositories (stub).
// ListRepositories lists repositories. opts is currently unused.
func (c *APIClient) ListRepositories(ctx context.Context, opts interface{}) ([]interface{}, error) {
	url := c.baseURL + "/repositories"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", c.tokenHeader())

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result []interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

// CreatePRComment posts a comment on a pull request (stub).
// CreatePRComment posts a comment on a pull request.
func (c *APIClient) CreatePRComment(ctx context.Context, owner, repo string, number int, body string) error {
	url := c.baseURL + fmt.Sprintf("/repositories/%s/%s/pulls/%d/comments", owner, repo, number)
	payload := map[string]string{"body": body}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", c.tokenHeader())
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}

// CreateStatusCheck creates a status check (stub).
// CreateStatusCheck creates a status check for a commit SHA.
func (c *APIClient) CreateStatusCheck(ctx context.Context, owner, repo, sha string, check *StatusCheck) error {
	url := c.baseURL + fmt.Sprintf("/repositories/%s/%s/statuses/%s", owner, repo, sha)
	data, err := json.Marshal(check)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", c.tokenHeader())
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}

// StatusCheck represents a status check result (stub).
// StatusCheck represents a status check result.
type StatusCheck struct {
	State       string `json:"state"`
	Context     string `json:"context"`
	Description string `json:"description"`
}
