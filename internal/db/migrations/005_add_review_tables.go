package migrations

import (
	"github.com/a5c-ai/hub/internal/models"
	"gorm.io/gorm"
)

func init() {
	registerMigration("005_add_review_tables", migrate005Up, migrate005Down)
}

func migrate005Up(db *gorm.DB) error {
	// Create review-related tables
	return db.AutoMigrate(
		&models.Review{},
		&models.ReviewComment{},
		&models.PullRequestFile{},
		&models.PullRequestMerge{},
	)
}

func migrate005Down(db *gorm.DB) error {
	// Drop tables in reverse order
	return db.Migrator().DropTable(
		&models.PullRequestMerge{},
		&models.PullRequestFile{},
		&models.ReviewComment{},
		&models.Review{},
	)
}