package migrations

import (
	"github.com/a5c-ai/hub/internal/models"
	"gorm.io/gorm"
)

func init() {
	registerMigration("001_initial_schema", migrate001Up, migrate001Down)
}

func migrate001Up(db *gorm.DB) error {
	// Enable UUID extension
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		return err
	}

	// Create tables in proper order due to foreign key dependencies
	return db.AutoMigrate(
		&models.User{},
		&models.SSHKey{},
		&models.Organization{},
		&models.OrganizationMember{},
		&models.Team{},
		&models.TeamMember{},
		&models.Repository{},
		&models.RepositoryCollaborator{},
		&models.Branch{},
		&models.BranchProtectionRule{},
		&models.Release{},
		&models.Milestone{},
		&models.Label{},
		&models.Issue{},
		&models.IssueLabel{},
		&models.Comment{},
		&models.PullRequest{},
	)
}

func migrate001Down(db *gorm.DB) error {
	// Drop tables in reverse order
	return db.Migrator().DropTable(
		&models.PullRequest{},
		&models.Comment{},
		&models.IssueLabel{},
		&models.Issue{},
		&models.Label{},
		&models.Milestone{},
		&models.Release{},
		&models.BranchProtectionRule{},
		&models.Branch{},
		&models.RepositoryCollaborator{},
		&models.Repository{},
		&models.TeamMember{},
		&models.Team{},
		&models.OrganizationMember{},
		&models.Organization{},
		&models.SSHKey{},
		&models.User{},
	)
}