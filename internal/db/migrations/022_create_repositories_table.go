package migrations

import (
	"github.com/a5c-ai/hub/internal/models"
	"gorm.io/gorm"
)

func init() {
	registerMigration("022_create_repositories_table", migrate022Up, migrate022Down)
}

func migrate022Up(db *gorm.DB) error {
	return db.AutoMigrate(&models.Repository{})
}

func migrate022Down(db *gorm.DB) error {
	return db.Migrator().DropTable(&models.Repository{})
}
