package migrations

import (
	"github.com/a5c-ai/hub/internal/models"
	"gorm.io/gorm"
)

func init() {
	registerMigration("003_git_tables", migrate003Up, migrate003Down)
}

func migrate003Up(db *gorm.DB) error {
	// Create commits table
	if err := db.AutoMigrate(&models.Commit{}); err != nil {
		return err
	}

	// Create commit_files table
	if err := db.AutoMigrate(&models.CommitFile{}); err != nil {
		return err
	}

	// Create tags table
	if err := db.AutoMigrate(&models.Tag{}); err != nil {
		return err
	}

	// Create git_refs table
	if err := db.AutoMigrate(&models.GitRef{}); err != nil {
		return err
	}

	// Create repository_hooks table
	if err := db.AutoMigrate(&models.RepositoryHook{}); err != nil {
		return err
	}

	// Create repository_clones table for analytics
	if err := db.AutoMigrate(&models.RepositoryClone{}); err != nil {
		return err
	}

	// Create repository_views table for analytics
	if err := db.AutoMigrate(&models.RepositoryView{}); err != nil {
		return err
	}

	// Add additional indexes for performance
	
	// Indexes for commits table
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_commits_repository_author_date ON commits(repository_id, author_date DESC)").Error; err != nil {
		return err
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_commits_sha_lookup ON commits(repository_id, sha)").Error; err != nil {
		return err
	}

	// Indexes for commit_files table
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_commit_files_path ON commit_files(commit_id, path)").Error; err != nil {
		return err
	}

	// Indexes for tags table
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_tags_repository_name ON tags(repository_id, name)").Error; err != nil {
		return err
	}

	// Indexes for git_refs table
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_git_refs_repository_type_name ON git_refs(repository_id, type, name)").Error; err != nil {
		return err
	}

	// Indexes for analytics tables
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_repo_clones_date ON repository_clones(repository_id, created_at DESC)").Error; err != nil {
		return err
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_repo_views_date ON repository_views(repository_id, created_at DESC)").Error; err != nil {
		return err
	}

	return nil
}

func migrate003Down(db *gorm.DB) error {
	// Drop tables in reverse order
	return db.Migrator().DropTable(
		&models.RepositoryView{},
		&models.RepositoryClone{},
		&models.RepositoryHook{},
		&models.GitRef{},
		&models.Tag{},
		&models.CommitFile{},
		&models.Commit{},
	)
}