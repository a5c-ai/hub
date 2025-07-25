package migrations

import (
	"github.com/a5c-ai/hub/internal/models"
	"gorm.io/gorm"
)

func init() {
	registerMigration("018", up018OrganizationEnhancements, down018OrganizationEnhancements)
}

func up018OrganizationEnhancements(db *gorm.DB) error {
	// Create custom roles table
	if err := db.AutoMigrate(&models.CustomRole{}); err != nil {
		return err
	}

	// Create organization policies table
	if err := db.AutoMigrate(&models.OrganizationPolicy{}); err != nil {
		return err
	}

	// Create organization templates table
	if err := db.AutoMigrate(&models.OrganizationTemplate{}); err != nil {
		return err
	}

	// Create organization settings table
	if err := db.AutoMigrate(&models.OrganizationSettings{}); err != nil {
		return err
	}

	// Add custom_role_id column to organization_members
	if err := db.Exec("ALTER TABLE organization_members ADD COLUMN IF NOT EXISTS custom_role_id UUID REFERENCES custom_roles(id)").Error; err != nil {
		return err
	}

	// Add custom_role_id column to organization_invitations
	if err := db.Exec("ALTER TABLE organization_invitations ADD COLUMN IF NOT EXISTS custom_role_id UUID REFERENCES custom_roles(id)").Error; err != nil {
		return err
	}

	// Update check constraints to include custom role
	if err := db.Exec("ALTER TABLE organization_members DROP CONSTRAINT IF EXISTS organization_members_role_check").Error; err != nil {
		return err
	}
	if err := db.Exec("ALTER TABLE organization_members ADD CONSTRAINT organization_members_role_check CHECK (role IN ('owner','admin','member','billing','custom'))").Error; err != nil {
		return err
	}

	if err := db.Exec("ALTER TABLE organization_invitations DROP CONSTRAINT IF EXISTS organization_invitations_role_check").Error; err != nil {
		return err
	}
	if err := db.Exec("ALTER TABLE organization_invitations ADD CONSTRAINT organization_invitations_role_check CHECK (role IN ('owner','admin','member','billing','custom'))").Error; err != nil {
		return err
	}

	// Create indexes for better performance
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_custom_roles_organization_id ON custom_roles(organization_id)").Error; err != nil {
		return err
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_organization_policies_org_type ON organization_policies(organization_id, policy_type)").Error; err != nil {
		return err
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_organization_templates_org_type ON organization_templates(organization_id, template_type)").Error; err != nil {
		return err
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_organization_members_custom_role ON organization_members(custom_role_id)").Error; err != nil {
		return err
	}

	return nil
}

func down018OrganizationEnhancements(db *gorm.DB) error {
	// Drop indexes
	db.Exec("DROP INDEX IF EXISTS idx_organization_members_custom_role")
	db.Exec("DROP INDEX IF EXISTS idx_organization_templates_org_type")
	db.Exec("DROP INDEX IF EXISTS idx_organization_policies_org_type")
	db.Exec("DROP INDEX IF EXISTS idx_custom_roles_organization_id")

	// Remove custom_role_id columns
	db.Exec("ALTER TABLE organization_invitations DROP COLUMN IF EXISTS custom_role_id")
	db.Exec("ALTER TABLE organization_members DROP COLUMN IF EXISTS custom_role_id")

	// Restore original check constraints
	db.Exec("ALTER TABLE organization_invitations DROP CONSTRAINT IF EXISTS organization_invitations_role_check")
	db.Exec("ALTER TABLE organization_invitations ADD CONSTRAINT organization_invitations_role_check CHECK (role IN ('owner','admin','member','billing'))")

	db.Exec("ALTER TABLE organization_members DROP CONSTRAINT IF EXISTS organization_members_role_check")
	db.Exec("ALTER TABLE organization_members ADD CONSTRAINT organization_members_role_check CHECK (role IN ('owner','admin','member','billing'))")

	// Drop tables
	db.Migrator().DropTable(&models.OrganizationSettings{})
	db.Migrator().DropTable(&models.OrganizationTemplate{})
	db.Migrator().DropTable(&models.OrganizationPolicy{})
	db.Migrator().DropTable(&models.CustomRole{})

	return nil
}