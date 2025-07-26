//go:build integration
// +build integration

package integration

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// TestPostgresConnection verifies that the PostgreSQL service is reachable.
func TestPostgresConnection(t *testing.T) {
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "hub"
	}
	pass := os.Getenv("DB_PASSWORD")
	if pass == "" {
		pass = "password"
	}
	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = "hub"
	}
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, pass, dbname)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("failed to open connection: %v", err)
	}
	defer db.Close()

	// Retry ping for up to 30 seconds
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		if err = db.Ping(); err == nil {
			return
		}
		time.Sleep(time.Second)
	}
	t.Fatalf("could not connect to Postgres within timeout: %v", err)
}
