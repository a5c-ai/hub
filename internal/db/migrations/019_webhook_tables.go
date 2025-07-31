package migrations

import (
	"github.com/a5c-ai/hub/internal/models"
	"gorm.io/gorm"
)

func init() {
	registerMigration("019_webhook_tables", migrate019Up, migrate019Down)
}

func migrate019Up(db *gorm.DB) error {
	// Create webhook tables
	return db.AutoMigrate(
		&models.Webhook{},
		&models.WebhookDelivery{},
		&models.DeployKey{},
		&models.WebhookEvent{},
	)
}

func migrate019Down(db *gorm.DB) error {
	// Drop webhook tables in reverse order
	return db.Migrator().DropTable(
		&models.WebhookEvent{},
		&models.DeployKey{},
		&models.WebhookDelivery{},
		&models.Webhook{},
	)
}
