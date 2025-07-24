package seeds

import (
	"fmt"
	"time"

	"github.com/a5c-ai/hub/internal/models"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SeedDatabase creates initial development data
func SeedDatabase(db *gorm.DB) error {
	fmt.Println("Seeding database with development data...")

	// Create admin user
	adminPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	adminUser := &models.User{
		ID:            uuid.New(),
		Username:      "admin",
		Email:         "admin@hub.local",
		PasswordHash:  string(adminPassword),
		FullName:      "Hub Administrator",
		Bio:           "Default administrator account for Hub git hosting service",
		EmailVerified: true,
		IsActive:      true,
		IsAdmin:       true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := db.FirstOrCreate(adminUser, models.User{Username: "admin"}).Error; err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	// Create test user
	testPassword, _ := bcrypt.GenerateFromPassword([]byte("test123"), bcrypt.DefaultCost)
	testUser := &models.User{
		ID:            uuid.New(),
		Username:      "testuser",
		Email:         "test@hub.local",
		PasswordHash:  string(testPassword),
		FullName:      "Test User",
		Bio:           "Test user account for development",
		EmailVerified: true,
		IsActive:      true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := db.FirstOrCreate(testUser, models.User{Username: "testuser"}).Error; err != nil {
		return fmt.Errorf("failed to create test user: %w", err)
	}

	// Create test organization
	testOrg := &models.Organization{
		ID:          uuid.New(),
		Name:        "test-org",
		DisplayName: "Test Organization",
		Description: "A test organization for development purposes",
		Email:       "org@hub.local",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := db.FirstOrCreate(testOrg, models.Organization{Name: "test-org"}).Error; err != nil {
		return fmt.Errorf("failed to create test organization: %w", err)
	}

	// Make admin user owner of test organization
	orgMember := &models.OrganizationMember{
		ID:             uuid.New(),
		OrganizationID: testOrg.ID,
		UserID:         adminUser.ID,
		Role:           models.OrgRoleOwner,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := db.FirstOrCreate(orgMember, models.OrganizationMember{
		OrganizationID: testOrg.ID,
		UserID:         adminUser.ID,
	}).Error; err != nil {
		return fmt.Errorf("failed to create organization membership: %w", err)
	}

	// Create test team
	testTeam := &models.Team{
		ID:             uuid.New(),
		OrganizationID: testOrg.ID,
		Name:           "developers",
		Description:    "Development team for the test organization",
		Privacy:        models.TeamPrivacyClosed,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := db.FirstOrCreate(testTeam, models.Team{
		OrganizationID: testOrg.ID,
		Name:           "developers",
	}).Error; err != nil {
		return fmt.Errorf("failed to create test team: %w", err)
	}

	// Add test user to team
	teamMember := &models.TeamMember{
		ID:        uuid.New(),
		TeamID:    testTeam.ID,
		UserID:    testUser.ID,
		Role:      models.TeamRoleMember,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := db.FirstOrCreate(teamMember, models.TeamMember{
		TeamID: testTeam.ID,
		UserID: testUser.ID,
	}).Error; err != nil {
		return fmt.Errorf("failed to create team membership: %w", err)
	}

	// Create sample repository
	testRepo := &models.Repository{
		ID:            uuid.New(),
		OwnerID:       adminUser.ID,
		OwnerType:     models.OwnerTypeUser,
		Name:          "sample-project",
		Description:   "A sample project repository for testing",
		DefaultBranch: "main",
		Visibility:    models.VisibilityPublic,
		HasIssues:     true,
		HasProjects:   true,
		HasWiki:       true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := db.FirstOrCreate(testRepo, models.Repository{
		OwnerID:   adminUser.ID,
		OwnerType: models.OwnerTypeUser,
		Name:      "sample-project",
	}).Error; err != nil {
		return fmt.Errorf("failed to create test repository: %w", err)
	}

	// Create main branch for repository
	mainBranch := &models.Branch{
		ID:           uuid.New(),
		RepositoryID: testRepo.ID,
		Name:         "main",
		SHA:          "abc123def456", // Placeholder SHA
		IsDefault:    true,
		IsProtected:  true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := db.FirstOrCreate(mainBranch, models.Branch{
		RepositoryID: testRepo.ID,
		Name:         "main",
	}).Error; err != nil {
		return fmt.Errorf("failed to create main branch: %w", err)
	}

	// Create branch protection rule
	branchProtection := &models.BranchProtectionRule{
		ID:                         uuid.New(),
		RepositoryID:               testRepo.ID,
		Pattern:                    "main",
		RequiredStatusChecks:       `{"strict":true,"contexts":["ci/tests"]}`,
		EnforceAdmins:              false,
		RequiredPullRequestReviews: `{"required_approving_review_count":1,"dismiss_stale_reviews":true}`,
		Restrictions:               `{}`,
		CreatedAt:                  time.Now(),
		UpdatedAt:                  time.Now(),
	}

	if err := db.FirstOrCreate(branchProtection, models.BranchProtectionRule{
		RepositoryID: testRepo.ID,
		Pattern:      "main",
	}).Error; err != nil {
		return fmt.Errorf("failed to create branch protection rule: %w", err)
	}

	// Create sample labels
	labels := []models.Label{
		{
			ID:           uuid.New(),
			RepositoryID: testRepo.ID,
			Name:         "bug",
			Color:        "#d73a49",
			Description:  "Something isn't working",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           uuid.New(),
			RepositoryID: testRepo.ID,
			Name:         "enhancement",
			Color:        "#a2eeef",
			Description:  "New feature or request",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           uuid.New(),
			RepositoryID: testRepo.ID,
			Name:         "documentation",
			Color:        "#0075ca",
			Description:  "Improvements or additions to documentation",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
	}

	for _, label := range labels {
		if err := db.FirstOrCreate(&label, models.Label{
			RepositoryID: testRepo.ID,
			Name:         label.Name,
		}).Error; err != nil {
			return fmt.Errorf("failed to create label %s: %w", label.Name, err)
		}
	}

	// Create sample issue
	testIssue := &models.Issue{
		ID:           uuid.New(),
		RepositoryID: testRepo.ID,
		Number:       1,
		Title:        "Welcome to Hub!",
		Body:         "This is a sample issue to demonstrate the issue tracking functionality.",
		UserID:       &adminUser.ID,
		State:        models.IssueStateOpen,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := db.FirstOrCreate(testIssue, models.Issue{
		RepositoryID: testRepo.ID,
		Number:       1,
	}).Error; err != nil {
		return fmt.Errorf("failed to create test issue: %w", err)
	}

	// Create sample comment
	testComment := &models.Comment{
		ID:        uuid.New(),
		IssueID:   testIssue.ID,
		UserID:    &testUser.ID,
		Body:      "Thanks for creating Hub! This looks great.",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := db.FirstOrCreate(testComment, models.Comment{
		IssueID: testIssue.ID,
		UserID:  &testUser.ID,
	}).Error; err != nil {
		return fmt.Errorf("failed to create test comment: %w", err)
	}

	fmt.Println("Database seeding completed successfully!")
	return nil
}