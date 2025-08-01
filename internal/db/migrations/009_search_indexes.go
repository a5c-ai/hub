package migrations

import (
	"gorm.io/gorm"
)

func init() {
	registerMigration("009_search_indexes", migrate009Up, migrate009Down)
}

func migrate009Up(db *gorm.DB) error {
	// Create full-text search indexes for better search performance

	// Users table - search index for username, full_name, email, bio, company
	if err := db.Exec(`
		CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_search_vector 
		ON users USING gin(to_tsvector('english', 
			coalesce(username, '') || ' ' || 
			coalesce(full_name, '') || ' ' || 
			coalesce(email, '') || ' ' || 
			coalesce(bio, '') || ' ' || 
			coalesce(company, '')
		))
	`).Error; err != nil {
		return err
	}

	// Repositories table - search index for name and description
	if err := db.Exec(`
		CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_repositories_search_vector 
		ON repositories USING gin(to_tsvector('english', 
			coalesce(name, '') || ' ' || 
			coalesce(description, '')
		))
	`).Error; err != nil {
		return err
	}

	// Issues table - search index for title and body
	if err := db.Exec(`
		CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_issues_search_vector 
		ON issues USING gin(to_tsvector('english', 
			coalesce(title, '') || ' ' || 
			coalesce(body, '')
		))
	`).Error; err != nil {
		return err
	}

	// Organizations table - search index for name, display_name, and description
	if err := db.Exec(`
		CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_organizations_search_vector 
		ON organizations USING gin(to_tsvector('english', 
			coalesce(name, '') || ' ' || 
			coalesce(display_name, '') || ' ' || 
			coalesce(description, '')
		))
	`).Error; err != nil {
		return err
	}

	// Commits table - search index for message, author_name, committer_name
	if err := db.Exec(`
		CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_commits_search_vector 
		ON commits USING gin(to_tsvector('english', 
			coalesce(message, '') || ' ' || 
			coalesce(author_name, '') || ' ' || 
			coalesce(committer_name, '')
		))
	`).Error; err != nil {
		return err
	}

	// Additional indexes for filtering and sorting
	if err := db.Exec(`
		CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_repositories_visibility_stars 
		ON repositories (visibility, stars_count DESC)
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_repositories_owner_visibility 
		ON repositories (owner_id, visibility)
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_issues_repository_state 
		ON issues (repository_id, state)
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_commits_repository_date 
		ON commits (repository_id, committer_date DESC)
	`).Error; err != nil {
		return err
	}

	return nil
}

func migrate009Down(db *gorm.DB) error {
	// Drop the search indexes
	indexes := []string{
		"idx_users_search_vector",
		"idx_repositories_search_vector",
		"idx_issues_search_vector",
		"idx_organizations_search_vector",
		"idx_commits_search_vector",
		"idx_repositories_visibility_stars",
		"idx_repositories_owner_visibility",
		"idx_issues_repository_state",
		"idx_commits_repository_date",
	}

	for _, index := range indexes {
		if err := db.Exec("DROP INDEX IF EXISTS " + index).Error; err != nil {
			return err
		}
	}

	return nil
}
