package migrations

import (
	"github.com/a5c-ai/hub/internal/models"
	"gorm.io/gorm"
)

func init() {
	registerMigration("001_initial_schema", migrate001Up, migrate001Down)
}

func migrate001Up(db *gorm.DB) error {
	// Enable UUID extensions for UUID generation
	// pgcrypto provides gen_random_uuid(), and uuid-ossp provides uuid_generate_v4()
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"pgcrypto\"").Error; err != nil {
		return err
	}
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		return err
	}

	// Create tables in proper order due to foreign key dependencies
	return db.AutoMigrate(
		&models.User{},
		&models.SSHKey{},
		&models.Organization{},
		&models.OrganizationMember{},
		&models.OrganizationInvitation{},
		&models.OrganizationActivity{},
		&models.Team{},
		&models.TeamMember{},
		&models.Repository{},
		&models.RepositoryCollaborator{},
		&models.RepositoryPermission{},
		&models.Branch{},
		&models.BranchProtectionRule{},
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
		&models.BranchProtectionRule{},
		&models.Branch{},
		&models.RepositoryPermission{},
		&models.RepositoryCollaborator{},
		&models.Repository{},
		&models.TeamMember{},
		&models.Team{},
		&models.OrganizationActivity{},
		&models.OrganizationInvitation{},
		&models.OrganizationMember{},
		&models.Organization{},
		&models.SSHKey{},
		&models.User{},
	)
}
