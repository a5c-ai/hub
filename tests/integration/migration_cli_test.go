//go:build integration
// +build integration

package integration

import (
	"os"
	"os/exec"
	"testing"
)

// TestMigrationCLIUpAndDown verifies that the migration CLI can apply and rollback migrations without errors.
func TestMigrationCLIUpAndDown(t *testing.T) {
	// Apply all pending migrations
	cmd := exec.Command("go", "run", "cmd/migrate/main.go")
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("migration CLI up failed: %v\nOutput:\n%s", err, out)
	}

	// Rollback last migration
	cmd = exec.Command("go", "run", "cmd/migrate/main.go", "-rollback")
	cmd.Env = os.Environ()
	out, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("migration CLI rollback failed: %v\nOutput:\n%s", err, out)
	}
}
