package migrations

import (
	"gorm.io/gorm"
)

func init() {
	registerMigration("002_indexes_and_constraints", migrate002Up, migrate002Down)
}

func migrate002Up(db *gorm.DB) error {
	// Performance indexes for frequent queries
	indexes := []string{
		// User indexes
		"CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)",
		"CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)",
		"CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at)",
		"CREATE INDEX IF NOT EXISTS idx_users_last_login_at ON users(last_login_at)",

		// SSH Key indexes
		"CREATE INDEX IF NOT EXISTS idx_ssh_keys_user_id ON ssh_keys(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_ssh_keys_fingerprint ON ssh_keys(fingerprint)",

		// Organization indexes
		"CREATE INDEX IF NOT EXISTS idx_organizations_name ON organizations(name)",
		"CREATE INDEX IF NOT EXISTS idx_organizations_created_at ON organizations(created_at)",

		// Organization member indexes
		"CREATE INDEX IF NOT EXISTS idx_org_members_org_id ON organization_members(organization_id)",
		"CREATE INDEX IF NOT EXISTS idx_org_members_user_id ON organization_members(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_org_members_role ON organization_members(role)",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_org_members_unique ON organization_members(organization_id, user_id)",

		// Team indexes
		"CREATE INDEX IF NOT EXISTS idx_teams_org_id ON teams(organization_id)",
		"CREATE INDEX IF NOT EXISTS idx_teams_name ON teams(name)",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_teams_org_name ON teams(organization_id, name)",

		// Team member indexes
		"CREATE INDEX IF NOT EXISTS idx_team_members_team_id ON team_members(team_id)",
		"CREATE INDEX IF NOT EXISTS idx_team_members_user_id ON team_members(user_id)",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_team_members_unique ON team_members(team_id, user_id)",

		// Repository indexes
		"CREATE INDEX IF NOT EXISTS idx_repositories_owner ON repositories(owner_id, owner_type)",
		"CREATE INDEX IF NOT EXISTS idx_repositories_name ON repositories(name)",
		"CREATE INDEX IF NOT EXISTS idx_repositories_visibility ON repositories(visibility)",
		"CREATE INDEX IF NOT EXISTS idx_repositories_created_at ON repositories(created_at)",
		"CREATE INDEX IF NOT EXISTS idx_repositories_updated_at ON repositories(updated_at)",
		"CREATE INDEX IF NOT EXISTS idx_repositories_pushed_at ON repositories(pushed_at)",
		"CREATE INDEX IF NOT EXISTS idx_repositories_parent_id ON repositories(parent_id)",
		"CREATE INDEX IF NOT EXISTS idx_repositories_is_fork ON repositories(is_fork)",
		"CREATE INDEX IF NOT EXISTS idx_repositories_is_template ON repositories(is_template)",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_repositories_owner_name ON repositories(owner_id, owner_type, name)",

		// Repository collaborator indexes
		"CREATE INDEX IF NOT EXISTS idx_repo_collaborators_repo_id ON repository_collaborators(repository_id)",
		"CREATE INDEX IF NOT EXISTS idx_repo_collaborators_user_id ON repository_collaborators(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_repo_collaborators_permission ON repository_collaborators(permission)",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_repo_collaborators_unique ON repository_collaborators(repository_id, user_id)",

		// Branch indexes
		"CREATE INDEX IF NOT EXISTS idx_branches_repo_id ON branches(repository_id)",
		"CREATE INDEX IF NOT EXISTS idx_branches_name ON branches(name)",
		"CREATE INDEX IF NOT EXISTS idx_branches_is_default ON branches(is_default)",
		"CREATE INDEX IF NOT EXISTS idx_branches_is_protected ON branches(is_protected)",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_branches_repo_name ON branches(repository_id, name)",

		// Branch protection rule indexes
		"CREATE INDEX IF NOT EXISTS idx_branch_protection_repo_id ON branch_protection_rules(repository_id)",
		"CREATE INDEX IF NOT EXISTS idx_branch_protection_pattern ON branch_protection_rules(pattern)",

		// Label indexes
		"CREATE INDEX IF NOT EXISTS idx_labels_repo_id ON labels(repository_id)",
		"CREATE INDEX IF NOT EXISTS idx_labels_name ON labels(name)",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_labels_repo_name ON labels(repository_id, name)",

		"CREATE INDEX IF NOT EXISTS idx_comments_user_id ON comments(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_comments_created_at ON comments(created_at)",

		// Pull request indexes
		"CREATE INDEX IF NOT EXISTS idx_pull_requests_issue_id ON pull_requests(issue_id)",
		"CREATE INDEX IF NOT EXISTS idx_pull_requests_head_repo_id ON pull_requests(head_repository_id)",
		"CREATE INDEX IF NOT EXISTS idx_pull_requests_base_repo_id ON pull_requests(base_repository_id)",
		"CREATE INDEX IF NOT EXISTS idx_pull_requests_merged_by_id ON pull_requests(merged_by_id)",
		"CREATE INDEX IF NOT EXISTS idx_pull_requests_merged ON pull_requests(merged)",
		"CREATE INDEX IF NOT EXISTS idx_pull_requests_draft ON pull_requests(draft)",
		"CREATE INDEX IF NOT EXISTS idx_pull_requests_merged_at ON pull_requests(merged_at)",
	}

	for _, index := range indexes {
		if err := db.Exec(index).Error; err != nil {
			return err
		}
	}

	return nil
}

func migrate002Down(db *gorm.DB) error {
	// Drop indexes (PostgreSQL will handle cascading)
	indexes := []string{
		"DROP INDEX IF EXISTS idx_users_username",
		"DROP INDEX IF EXISTS idx_users_email",
		"DROP INDEX IF EXISTS idx_users_created_at",
		"DROP INDEX IF EXISTS idx_users_last_login_at",
		"DROP INDEX IF EXISTS idx_ssh_keys_user_id",
		"DROP INDEX IF EXISTS idx_ssh_keys_fingerprint",
		"DROP INDEX IF EXISTS idx_organizations_name",
		"DROP INDEX IF EXISTS idx_organizations_created_at",
		"DROP INDEX IF EXISTS idx_org_members_org_id",
		"DROP INDEX IF EXISTS idx_org_members_user_id",
		"DROP INDEX IF EXISTS idx_org_members_role",
		"DROP INDEX IF EXISTS idx_org_members_unique",
		"DROP INDEX IF EXISTS idx_teams_org_id",
		"DROP INDEX IF EXISTS idx_teams_name",
		"DROP INDEX IF EXISTS idx_teams_org_name",
		"DROP INDEX IF EXISTS idx_team_members_team_id",
		"DROP INDEX IF EXISTS idx_team_members_user_id",
		"DROP INDEX IF EXISTS idx_team_members_unique",
		"DROP INDEX IF EXISTS idx_repositories_owner",
		"DROP INDEX IF EXISTS idx_repositories_name",
		"DROP INDEX IF EXISTS idx_repositories_visibility",
		"DROP INDEX IF EXISTS idx_repositories_created_at",
		"DROP INDEX IF EXISTS idx_repositories_updated_at",
		"DROP INDEX IF EXISTS idx_repositories_pushed_at",
		"DROP INDEX IF EXISTS idx_repositories_parent_id",
		"DROP INDEX IF EXISTS idx_repositories_is_fork",
		"DROP INDEX IF EXISTS idx_repositories_is_template",
		"DROP INDEX IF EXISTS idx_repositories_owner_name",
		"DROP INDEX IF EXISTS idx_repo_collaborators_repo_id",
		"DROP INDEX IF EXISTS idx_repo_collaborators_user_id",
		"DROP INDEX IF EXISTS idx_repo_collaborators_permission",
		"DROP INDEX IF EXISTS idx_repo_collaborators_unique",
		"DROP INDEX IF EXISTS idx_branches_repo_id",
		"DROP INDEX IF EXISTS idx_branches_name",
		"DROP INDEX IF EXISTS idx_branches_is_default",
		"DROP INDEX IF EXISTS idx_branches_is_protected",
		"DROP INDEX IF EXISTS idx_branches_repo_name",
		"DROP INDEX IF EXISTS idx_branch_protection_repo_id",
		"DROP INDEX IF EXISTS idx_branch_protection_pattern",

		"DROP INDEX IF EXISTS idx_labels_repo_id",
		"DROP INDEX IF EXISTS idx_labels_name",
		"DROP INDEX IF EXISTS idx_labels_repo_name",
		"DROP INDEX IF EXISTS idx_issues_repo_id",
		"DROP INDEX IF EXISTS idx_issues_user_id",
		"DROP INDEX IF EXISTS idx_issues_assignee_id",

		"DROP INDEX IF EXISTS idx_issues_state",
		"DROP INDEX IF EXISTS idx_issues_created_at",
		"DROP INDEX IF EXISTS idx_issues_updated_at",
		"DROP INDEX IF EXISTS idx_issues_closed_at",
		"DROP INDEX IF EXISTS idx_issues_repo_number",
		"DROP INDEX IF EXISTS idx_issue_labels_issue_id",
		"DROP INDEX IF EXISTS idx_issue_labels_label_id",
		"DROP INDEX IF EXISTS idx_comments_issue_id",
		"DROP INDEX IF EXISTS idx_comments_user_id",
		"DROP INDEX IF EXISTS idx_comments_created_at",
		"DROP INDEX IF EXISTS idx_pull_requests_issue_id",
		"DROP INDEX IF EXISTS idx_pull_requests_head_repo_id",
		"DROP INDEX IF EXISTS idx_pull_requests_base_repo_id",
		"DROP INDEX IF EXISTS idx_pull_requests_merged_by_id",
		"DROP INDEX IF EXISTS idx_pull_requests_merged",
		"DROP INDEX IF EXISTS idx_pull_requests_draft",
		"DROP INDEX IF EXISTS idx_pull_requests_merged_at",
	}

	for _, index := range indexes {
		if err := db.Exec(index).Error; err != nil {
			return err
		}
	}

	return nil
}
