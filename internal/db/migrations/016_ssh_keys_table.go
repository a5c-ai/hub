package migrations

import (
	"gorm.io/gorm"
)

// Migration016_ssh_keys_table adds indexes for SSH keys table
func Migration016_ssh_keys_table(db *gorm.DB) error {
	// The SSH keys table is already created by the models auto-migration
	// This migration adds additional indexes and constraints for performance

	// Add index on user_id for faster lookups
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_ssh_keys_user_id ON ssh_keys(user_id)").Error; err != nil {
		return err
	}

	// Add index on fingerprint for uniqueness and fast lookups
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_ssh_keys_fingerprint ON ssh_keys(fingerprint)").Error; err != nil {
		return err
	}

	// Add index on last_used_at for analytics
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_ssh_keys_last_used_at ON ssh_keys(last_used_at)").Error; err != nil {
		return err
	}

	return nil
}
