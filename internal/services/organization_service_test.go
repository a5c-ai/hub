package services

import (
	"context"
	"testing"

	"github.com/a5c-ai/hub/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Mock ActivityService for testing
type mockActivityService struct {
	mock.Mock
}

func (m *mockActivityService) LogActivity(ctx context.Context, orgID, actorID uuid.UUID, action models.ActivityAction, targetType string, targetID *uuid.UUID, metadata map[string]interface{}) error {
	args := m.Called(ctx, orgID, actorID, action, targetType, targetID, metadata)
	return args.Error(0)
}

func (m *mockActivityService) GetActivity(ctx context.Context, orgName string, limit, offset int) ([]*models.OrganizationActivity, error) {
	args := m.Called(ctx, orgName, limit, offset)
	return args.Get(0).([]*models.OrganizationActivity), args.Error(1)
}

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// For SQLite, manually create simplified tables to avoid UUID issues
	err = db.Exec(`
		CREATE TABLE users (
			id TEXT PRIMARY KEY,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			username TEXT UNIQUE NOT NULL,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			full_name TEXT,
			avatar_url TEXT,
			bio TEXT,
			location TEXT,
			website TEXT,
			company TEXT,
			email_verified BOOLEAN DEFAULT FALSE,
			two_factor_enabled BOOLEAN DEFAULT FALSE,
			is_active BOOLEAN DEFAULT TRUE,
			is_admin BOOLEAN DEFAULT FALSE,
			last_login_at DATETIME
		);
		
		CREATE TABLE organizations (
			id TEXT PRIMARY KEY,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			name TEXT UNIQUE NOT NULL,
			display_name TEXT NOT NULL,
			description TEXT,
			avatar_url TEXT,
			website TEXT,
			location TEXT,
			email TEXT,
			billing_email TEXT
		);
		
		CREATE TABLE organization_members (
			id TEXT PRIMARY KEY,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			organization_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			role TEXT NOT NULL,
			public_member BOOLEAN DEFAULT FALSE,
			FOREIGN KEY (organization_id) REFERENCES organizations(id),
			FOREIGN KEY (user_id) REFERENCES users(id)
		);
	`).Error
	assert.NoError(t, err)

	return db
}

func TestOrganizationService_Create(t *testing.T) {
	db := setupTestDB(t)
	mockAS := new(mockActivityService)
	service := NewOrganizationService(db, mockAS)

	ownerID := uuid.New()
	req := CreateOrganizationRequest{
		Login:       "test-org",
		Name:        "Test Organization",
		Description: "A test organization",
		Email:       "test@example.com",
	}

	// Create test user first using direct SQL for SQLite compatibility
	db.Exec("INSERT INTO users (id, username, email, password_hash, created_at, updated_at) VALUES (?, ?, ?, ?, datetime('now'), datetime('now'))",
		ownerID.String(), "testuser", "test@example.com", "hash")

	// Mock activity service
	mockAS.On("LogActivity", mock.Anything, mock.AnythingOfType("uuid.UUID"), ownerID, models.ActivityMemberAdded, "organization", mock.AnythingOfType("*uuid.UUID"), mock.Anything).Return(nil)

	org, err := service.Create(context.Background(), req, ownerID)

	assert.NoError(t, err)
	assert.NotNil(t, org)
	assert.Equal(t, req.Login, org.Name)
	assert.Equal(t, req.Name, org.DisplayName)
	assert.Equal(t, req.Description, org.Description)
	assert.Equal(t, req.Email, org.Email)

	// Verify organization member was created using direct SQL
	var count int
	db.Raw("SELECT COUNT(*) FROM organization_members WHERE organization_id = ? AND user_id = ? AND role = ?",
		org.ID.String(), ownerID.String(), string(models.OrgRoleOwner)).Scan(&count)
	assert.Equal(t, 1, count)
}

func TestOrganizationService_Get(t *testing.T) {
	db := setupTestDB(t)
	service := NewOrganizationService(db, nil)

	// Create test organization using direct SQL
	orgID := uuid.New()
	db.Exec("INSERT INTO organizations (id, name, display_name, description, created_at, updated_at) VALUES (?, ?, ?, ?, datetime('now'), datetime('now'))",
		orgID.String(), "test-org", "Test Organization", "A test organization")

	result, err := service.Get(context.Background(), "test-org")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test-org", result.Name)
	assert.Equal(t, "Test Organization", result.DisplayName)
}

func TestOrganizationService_Update(t *testing.T) {
	db := setupTestDB(t)
	service := NewOrganizationService(db, nil)

	// Create test organization
	org := &models.Organization{
		Name:        "test-org",
		DisplayName: "Test Organization",
		Description: "A test organization",
	}
	db.Create(org)

	newDisplayName := "Updated Organization"
	newDescription := "Updated description"
	req := UpdateOrganizationRequest{
		DisplayName: &newDisplayName,
		Description: &newDescription,
	}

	result, err := service.Update(context.Background(), "test-org", req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, newDisplayName, result.DisplayName)
	assert.Equal(t, newDescription, result.Description)
}

func TestOrganizationService_Delete(t *testing.T) {
	db := setupTestDB(t)
	service := NewOrganizationService(db, nil)

	// Create test organization
	org := &models.Organization{
		Name:        "test-org",
		DisplayName: "Test Organization",
	}
	db.Create(org)

	err := service.Delete(context.Background(), "test-org")
	assert.NoError(t, err)

	// Verify organization was deleted
	var result models.Organization
	err = db.Where("name = ?", "test-org").First(&result).Error
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

func TestOrganizationService_List(t *testing.T) {
	db := setupTestDB(t)
	service := NewOrganizationService(db, nil)

	// Create test organizations
	orgs := []*models.Organization{
		{Name: "org1", DisplayName: "Organization 1"},
		{Name: "org2", DisplayName: "Organization 2"},
		{Name: "org3", DisplayName: "Organization 3"},
	}

	for _, org := range orgs {
		db.Create(org)
	}

	filters := OrganizationFilters{Limit: 2}
	result, err := service.List(context.Background(), filters)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestOrganizationService_GetUserOrganizations(t *testing.T) {
	db := setupTestDB(t)
	service := NewOrganizationService(db, nil)

	userID := uuid.New()

	// Create test user
	user := &models.User{
		ID:       userID,
		Username: "testuser",
		Email:    "test@example.com",
	}
	db.Create(user)

	// Create test organizations
	org1 := &models.Organization{Name: "org1", DisplayName: "Organization 1"}
	org2 := &models.Organization{Name: "org2", DisplayName: "Organization 2"}
	org3 := &models.Organization{Name: "org3", DisplayName: "Organization 3"}

	db.Create(org1)
	db.Create(org2)
	db.Create(org3)

	// Create memberships for user in org1 and org2
	member1 := &models.OrganizationMember{
		OrganizationID: org1.ID,
		UserID:         userID,
		Role:           models.OrgRoleMember,
	}
	member2 := &models.OrganizationMember{
		OrganizationID: org2.ID,
		UserID:         userID,
		Role:           models.OrgRoleAdmin,
	}

	db.Create(member1)
	db.Create(member2)

	result, err := service.GetUserOrganizations(context.Background(), userID)

	assert.NoError(t, err)
	assert.Len(t, result, 2)

	// Verify correct organizations are returned
	orgNames := make([]string, len(result))
	for i, org := range result {
		orgNames[i] = org.Name
	}
	assert.Contains(t, orgNames, "org1")
	assert.Contains(t, orgNames, "org2")
	assert.NotContains(t, orgNames, "org3")
}
