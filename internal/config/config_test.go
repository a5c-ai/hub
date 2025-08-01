package config

import (
	"os"
	"testing"
)

func TestLoadDefault(t *testing.T) {
	// Clear environment to ensure defaults are used
	oldPort := os.Getenv("PORT")
	oldEnv := os.Getenv("ENVIRONMENT")
	defer func() {
		os.Setenv("PORT", oldPort)
		os.Setenv("ENVIRONMENT", oldEnv)
	}()
	os.Unsetenv("PORT")
	os.Unsetenv("ENVIRONMENT")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load default config: %v", err)
	}

	if cfg.Environment != "development" {
		t.Errorf("Expected environment to be 'development', got %s", cfg.Environment)
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("Expected server port to be 8080, got %d", cfg.Server.Port)
	}

	if cfg.Database.Host != "localhost" {
		t.Errorf("Expected database host to be 'localhost', got %s", cfg.Database.Host)
	}
}

func TestLoadFromEnv(t *testing.T) {
	oldPort := os.Getenv("PORT")
	oldEnv := os.Getenv("ENVIRONMENT")

	defer func() {
		os.Setenv("PORT", oldPort)
		os.Setenv("ENVIRONMENT", oldEnv)
	}()

	os.Setenv("PORT", "9000")
	os.Setenv("ENVIRONMENT", "production")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config from env: %v", err)
	}

	if cfg.Server.Port != 9000 {
		t.Errorf("Expected server port to be 9000, got %d", cfg.Server.Port)
	}

	if cfg.Environment != "production" {
		t.Errorf("Expected environment to be 'production', got %s", cfg.Environment)
	}
}
