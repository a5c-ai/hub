package plugin

import (
	"context"
	"fmt"
	"net/http"
)

// APIClient provides methods for plugin operations.
type APIClient struct {
	baseURL string
	token   string
	client  *http.Client
}

// NewAPIClient creates a new APIClient with the given base URL and token.
func NewAPIClient(baseURL, token string) *APIClient {
	return &APIClient{baseURL: baseURL, token: token, client: http.DefaultClient}
}

// InstallPlugin installs a plugin by name, version, or URL.
func (c *APIClient) InstallPlugin(ctx context.Context, name, version, url string) error {
	return fmt.Errorf("InstallPlugin not implemented")
}

// ConfigPlugin configures settings for a plugin at org or repo scope.
func (c *APIClient) ConfigPlugin(ctx context.Context, name, org, repo string, settings map[string]string) error {
	return fmt.Errorf("ConfigPlugin not implemented")
}

// EnablePlugin enables a plugin for an organization or repository.
func (c *APIClient) EnablePlugin(ctx context.Context, name, org, repo string) error {
	return fmt.Errorf("EnablePlugin not implemented")
}
