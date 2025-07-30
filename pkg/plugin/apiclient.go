package plugin

import (
	"context"
	"net/http"
)

// APIClient enables plugins to call core platform APIs.
type APIClient struct {
	token  string
	client *http.Client
}

// NewAPIClient constructs a new APIClient with the given bearer token.
func NewAPIClient(token string) *APIClient {
	return &APIClient{token: token, client: http.DefaultClient}
}

// tokenHeader returns the authorization header value.
func (c *APIClient) tokenHeader() string {
	return "Bearer " + c.token
}

// GetRepository retrieves repository details (stub).
func (c *APIClient) GetRepository(ctx context.Context, owner, repo string) (interface{}, error) {
	// TODO: implement API call
	return nil, nil
}

// ListRepositories lists repositories (stub).
func (c *APIClient) ListRepositories(ctx context.Context, opts interface{}) ([]interface{}, error) {
	// TODO: implement API call
	return nil, nil
}

// CreatePRComment posts a comment on a pull request (stub).
func (c *APIClient) CreatePRComment(ctx context.Context, owner, repo string, number int, body string) error {
	// TODO: implement API call
	return nil
}

// CreateStatusCheck creates a status check (stub).
func (c *APIClient) CreateStatusCheck(ctx context.Context, owner, repo, sha string, check *StatusCheck) error {
	// TODO: implement API call
	return nil
}

// StatusCheck represents a status check result (stub).
type StatusCheck struct {
	State       string
	Context     string
	Description string
}
