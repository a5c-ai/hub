package migrations

import (
	"github.com/a5c-ai/hub/internal/models"
	"gorm.io/gorm"
)

func init() {
	registerMigration("023_create_organizations_table", migrate023Up, migrate023Down)
}

func migrate023Up(db *gorm.DB) error {
	return db.AutoMigrate(&models.Organization{})
}

func migrate023Down(db *gorm.DB) error {
	return db.Migrator().DropTable(&models.Organization{})
}
