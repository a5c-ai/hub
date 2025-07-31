package migrations

import (
	"fmt"
	"sort"
	"time"

	"gorm.io/gorm"
)

// Migration represents a database migration
type Migration struct {
	Version   string    `gorm:"primaryKey"`
	AppliedAt time.Time `gorm:"not null"`
}

// MigrationFunc is the type for migration functions
type MigrationFunc func(*gorm.DB) error

// MigrationItem represents a single migration
type MigrationItem struct {
	Version string
	Up      MigrationFunc
	Down    MigrationFunc
}

// Migrator handles database migrations
type Migrator struct {
	db         *gorm.DB
	migrations []MigrationItem
}

// NewMigrator creates a new migrator instance
func NewMigrator(db *gorm.DB) *Migrator {
	return &Migrator{
		db:         db,
		migrations: getAllMigrations(),
	}
}

// Migrate runs all pending migrations
func (m *Migrator) Migrate() error {
	// Create migrations table if it doesn't exist
	if err := m.db.AutoMigrate(&Migration{}); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get applied migrations
	var appliedMigrations []Migration
	if err := m.db.Order("version").Find(&appliedMigrations).Error; err != nil {
		return fmt.Errorf("failed to query applied migrations: %w", err)
	}

	appliedVersions := make(map[string]bool)
	for _, migration := range appliedMigrations {
		appliedVersions[migration.Version] = true
	}

	// Sort migrations by version
	sort.Slice(m.migrations, func(i, j int) bool {
		return m.migrations[i].Version < m.migrations[j].Version
	})

	// Apply pending migrations
	for _, migration := range m.migrations {
		if appliedVersions[migration.Version] {
			continue
		}

		fmt.Printf("Applying migration %s...\n", migration.Version)

		if err := migration.Up(m.db); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", migration.Version, err)
		}

		// Record migration as applied
		migrationRecord := Migration{
			Version:   migration.Version,
			AppliedAt: time.Now(),
		}
		if err := m.db.Create(&migrationRecord).Error; err != nil {
			return fmt.Errorf("failed to record migration %s: %w", migration.Version, err)
		}

		fmt.Printf("Migration %s applied successfully\n", migration.Version)
	}

	return nil
}

// Rollback rolls back the last migration
func (m *Migrator) Rollback() error {
	// Get the last applied migration
	var lastMigration Migration
	if err := m.db.Order("version DESC").First(&lastMigration).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			fmt.Println("No migrations to rollback")
			return nil
		}
		return fmt.Errorf("failed to get last migration: %w", err)
	}

	// Find the migration to rollback
	var migrationToRollback *MigrationItem
	for _, migration := range m.migrations {
		if migration.Version == lastMigration.Version {
			migrationToRollback = &migration
			break
		}
	}

	if migrationToRollback == nil {
		return fmt.Errorf("migration %s not found in migration list", lastMigration.Version)
	}

	fmt.Printf("Rolling back migration %s...\n", migrationToRollback.Version)

	if err := migrationToRollback.Down(m.db); err != nil {
		return fmt.Errorf("failed to rollback migration %s: %w", migrationToRollback.Version, err)
	}

	// Remove migration record
	if err := m.db.Delete(&lastMigration).Error; err != nil {
		return fmt.Errorf("failed to remove migration record %s: %w", migrationToRollback.Version, err)
	}

	fmt.Printf("Migration %s rolled back successfully\n", migrationToRollback.Version)
	return nil
}
