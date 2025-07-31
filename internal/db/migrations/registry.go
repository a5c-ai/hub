package migrations

var allMigrations []MigrationItem

// registerMigration registers a migration to be run
func registerMigration(version string, up, down MigrationFunc) {
	allMigrations = append(allMigrations, MigrationItem{
		Version: version,
		Up:      up,
		Down:    down,
	})
}

// getAllMigrations returns all registered migrations
func getAllMigrations() []MigrationItem {
	return allMigrations
}
