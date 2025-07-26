//go:build integration
// +build integration

package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
)

// TestElasticsearchHealth verifies that the Elasticsearch service is reachable.
func TestElasticsearchHealth(t *testing.T) {
	addr := os.Getenv("ES_URL")
	if addr == "" {
		addr = "http://localhost:9200"
	}
	es, err := elasticsearch.NewClient(elasticsearch.Config{Addresses: []string{addr}})
	if err != nil {
		t.Fatalf("failed to create Elasticsearch client: %v", err)
	}

	// Retry cluster health for up to 30 seconds
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		res, err := es.Cluster.Health(es.Cluster.Health.WithContext(context.Background()))
		if err == nil && res.StatusCode == 200 {
			res.Body.Close()
			return
		}
		time.Sleep(time.Second)
	}
	t.Fatalf("Elasticsearch cluster health did not return OK within timeout")
}
