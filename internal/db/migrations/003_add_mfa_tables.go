package migrations

import (
	"github.com/a5c-ai/hub/internal/auth"
	"gorm.io/gorm"
)

func init() {
	registerMigration("003_add_mfa_tables", migrate003Up, migrate003Down)
}

func migrate003Up(db *gorm.DB) error {
	// Create backup_codes table
	return db.AutoMigrate(&auth.BackupCode{})
}

func migrate003Down(db *gorm.DB) error {
	return db.Migrator().DropTable(&auth.BackupCode{})
}
