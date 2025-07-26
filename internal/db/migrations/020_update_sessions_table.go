package migrations

import (
	"gorm.io/gorm"
)

func init() {
	registerMigration("020_update_sessions_table", migrate020Up, migrate020Down)
}

func migrate020Up(db *gorm.DB) error {
	// Add missing columns to sessions table
	err := db.Exec(`
		ALTER TABLE sessions 
		ADD COLUMN IF NOT EXISTS device_name VARCHAR(255),
		ADD COLUMN IF NOT EXISTS location_info VARCHAR(255),
		ADD COLUMN IF NOT EXISTS is_remembered BOOLEAN DEFAULT FALSE,
		ADD COLUMN IF NOT EXISTS security_flags INTEGER DEFAULT 0
	`).Error

	if err != nil {
		return err
	}

	return nil
}

func migrate020Down(db *gorm.DB) error {
	// Remove the added columns
	err := db.Exec(`
		ALTER TABLE sessions 
		DROP COLUMN IF EXISTS device_name,
		DROP COLUMN IF EXISTS location_info,
		DROP COLUMN IF EXISTS is_remembered,
		DROP COLUMN IF EXISTS security_flags
	`).Error

	return err
}
