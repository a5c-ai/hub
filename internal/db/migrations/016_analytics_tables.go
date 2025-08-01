package migrations

import (
	"github.com/a5c-ai/hub/internal/models"
	"gorm.io/gorm"
)

func init() {
	registerMigration("016_analytics_tables", migrate016Up, migrate016Down)
}

func migrate016Up(db *gorm.DB) error {
	// Create analytics tables in proper order
	return db.AutoMigrate(
		&models.AnalyticsEvent{},
		&models.AnalyticsMetric{},
		&models.RepositoryAnalytics{},
		&models.UserAnalytics{},
		&models.OrganizationAnalytics{},
		&models.SystemAnalytics{},
		&models.PerformanceLog{},
	)
}

func migrate016Down(db *gorm.DB) error {
	// Drop tables in reverse order
	return db.Migrator().DropTable(
		&models.PerformanceLog{},
		&models.SystemAnalytics{},
		&models.OrganizationAnalytics{},
		&models.UserAnalytics{},
		&models.RepositoryAnalytics{},
		&models.AnalyticsMetric{},
		&models.AnalyticsEvent{},
	)
}
