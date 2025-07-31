package migrations

import (
	"github.com/a5c-ai/hub/internal/models"
	"gorm.io/gorm"
)

func init() {
	registerMigration("018_repository_features_tables", migrate018Up, migrate018Down)
}

func migrate018Up(db *gorm.DB) error {
	// Create repository_languages table
	err := db.AutoMigrate(&models.RepositoryLanguage{})
	if err != nil {
		return err
	}

	// Create repository_statistics table
	err = db.AutoMigrate(&models.RepositoryStatistics{})
	if err != nil {
		return err
	}

	// Create repository_templates table
	err = db.AutoMigrate(&models.RepositoryTemplate{})
	if err != nil {
		return err
	}

	// Create git_hooks table
	err = db.AutoMigrate(&models.GitHook{})
	if err != nil {
		return err
	}

	// Create repository_imports table
	err = db.AutoMigrate(&models.RepositoryImport{})
	if err != nil {
		return err
	}

	// Add indexes for performance

	// Repository languages indexes
	err = db.Exec("CREATE INDEX IF NOT EXISTS idx_repository_languages_repo_id ON repository_languages(repository_id)").Error
	if err != nil {
		return err
	}

	err = db.Exec("CREATE INDEX IF NOT EXISTS idx_repository_languages_language ON repository_languages(language)").Error
	if err != nil {
		return err
	}

	// Repository statistics indexes
	err = db.Exec("CREATE INDEX IF NOT EXISTS idx_repository_statistics_repo_id ON repository_statistics(repository_id)").Error
	if err != nil {
		return err
	}

	err = db.Exec("CREATE INDEX IF NOT EXISTS idx_repository_statistics_primary_language ON repository_statistics(primary_language)").Error
	if err != nil {
		return err
	}

	// Repository templates indexes
	err = db.Exec("CREATE INDEX IF NOT EXISTS idx_repository_templates_repo_id ON repository_templates(repository_id)").Error
	if err != nil {
		return err
	}

	err = db.Exec("CREATE INDEX IF NOT EXISTS idx_repository_templates_category ON repository_templates(category)").Error
	if err != nil {
		return err
	}

	err = db.Exec("CREATE INDEX IF NOT EXISTS idx_repository_templates_featured ON repository_templates(is_featured)").Error
	if err != nil {
		return err
	}

	err = db.Exec("CREATE INDEX IF NOT EXISTS idx_repository_templates_public ON repository_templates(is_public)").Error
	if err != nil {
		return err
	}

	// Git hooks indexes
	err = db.Exec("CREATE INDEX IF NOT EXISTS idx_git_hooks_repo_id ON git_hooks(repository_id)").Error
	if err != nil {
		return err
	}

	err = db.Exec("CREATE INDEX IF NOT EXISTS idx_git_hooks_type ON git_hooks(hook_type)").Error
	if err != nil {
		return err
	}

	err = db.Exec("CREATE INDEX IF NOT EXISTS idx_git_hooks_enabled ON git_hooks(is_enabled)").Error
	if err != nil {
		return err
	}

	// Repository imports indexes
	err = db.Exec("CREATE INDEX IF NOT EXISTS idx_repository_imports_repo_id ON repository_imports(repository_id)").Error
	if err != nil {
		return err
	}

	err = db.Exec("CREATE INDEX IF NOT EXISTS idx_repository_imports_status ON repository_imports(status)").Error
	if err != nil {
		return err
	}

	err = db.Exec("CREATE INDEX IF NOT EXISTS idx_repository_imports_source_type ON repository_imports(source_type)").Error
	if err != nil {
		return err
	}

	return nil
}

func migrate018Down(db *gorm.DB) error {
	// Drop tables in reverse order
	err := db.Migrator().DropTable(&models.RepositoryImport{})
	if err != nil {
		return err
	}

	err = db.Migrator().DropTable(&models.GitHook{})
	if err != nil {
		return err
	}

	err = db.Migrator().DropTable(&models.RepositoryTemplate{})
	if err != nil {
		return err
	}

	err = db.Migrator().DropTable(&models.RepositoryStatistics{})
	if err != nil {
		return err
	}

	err = db.Migrator().DropTable(&models.RepositoryLanguage{})
	if err != nil {
		return err
	}

	return nil
}
