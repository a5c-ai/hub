package migrations

import (
	"github.com/a5c-ai/hub/internal/models"
	"gorm.io/gorm"
)

func init() {
	registerMigration("001_5_add_missing_tables", migrate001_5Up, migrate001_5Down)
}

func migrate001_5Up(db *gorm.DB) error {
	// Create missing tables that should have been in 001_initial_schema
	// but were added later to fix migration 002 dependencies

	// Create tables in proper order due to foreign key dependencies
	if err := db.AutoMigrate(&models.Label{}); err != nil {
		return err
	}

	if err := db.AutoMigrate(&models.Issue{}); err != nil {
		return err
	}

	if err := db.AutoMigrate(&models.IssueLabel{}); err != nil {
		return err
	}

	if err := db.AutoMigrate(&models.Comment{}); err != nil {
		return err
	}

	// Add the IssueID field to pull_requests table if it doesn't exist
	if !db.Migrator().HasColumn(&models.PullRequest{}, "issue_id") {
		if err := db.Migrator().AddColumn(&models.PullRequest{}, "issue_id"); err != nil {
			return err
		}
	}

	return nil
}

func migrate001_5Down(db *gorm.DB) error {
	// Drop the tables in reverse order
	return db.Migrator().DropTable(
		&models.Comment{},
		&models.IssueLabel{},
		&models.Issue{},
		&models.Label{},
	)
}
