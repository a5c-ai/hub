package migrations

import (
	"github.com/a5c-ai/hub/internal/models"
	"gorm.io/gorm"
)

func init() {
	registerMigration("021_create_users_table", migrate021Up, migrate021Down)
}

func migrate021Up(db *gorm.DB) error {
	return db.AutoMigrate(&models.User{})
}

func migrate021Down(db *gorm.DB) error {
	return db.Migrator().DropTable(&models.User{})
}
