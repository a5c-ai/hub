package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/a5c-ai/hub/internal/models"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupSearchTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Create tables
	err = db.AutoMigrate(
		&models.User{},
		&models.Organization{},
		&models.Repository{},
		&models.Commit{},
	)
	require.NoError(t, err)

	return db
}

func TestSearchService_GlobalSearch(t *testing.T) {
	db := setupSearchTestDB(t)
	service := NewSearchService(db, nil, logrus.New())

	// Create test data
	user := models.User{
		ID:       uuid.New(),
		Username: "testuser",
		FullName: "Test User",
		Email:    "test@example.com",
		Bio:      "Test user bio",
	}
	require.NoError(t, db.Create(&user).Error)

	org := models.Organization{
		ID:          uuid.New(),
		Name:        "testorg",
		DisplayName: "Test Organization",
		Description: "Test organization description",
	}
	require.NoError(t, db.Create(&org).Error)

	repo := models.Repository{
		ID:          uuid.New(),
		Name:        "testrepo",
		Description: "Test repository description",
		OwnerID:     user.ID,
		OwnerType:   "user",
		Visibility:  "public",
	}
	require.NoError(t, db.Create(&repo).Error)

	commit := models.Commit{
		ID:           uuid.New(),
		SHA:          "abcd1234567890abcd1234567890abcd12345678",
		Message:      "Test commit message",
		AuthorName:   "Test Author",
		AuthorEmail:  "author@example.com",
		RepositoryID: repo.ID,
	}
	require.NoError(t, db.Create(&commit).Error)

	tests := []struct {
		name          string
		filter        SearchFilter
		expectUsers   bool
		expectRepos   bool
		expectOrgs    bool
		expectCommits bool
	}{
		{
			name: "global search for test",
			filter: SearchFilter{
				Query:   "test",
				Page:    1,
				PerPage: 30,
			},
			expectUsers:   true,
			expectRepos:   true,
			expectOrgs:    true,
			expectCommits: true,
		},
		{
			name: "search users only",
			filter: SearchFilter{
				Query:   "test",
				Type:    "user",
				Page:    1,
				PerPage: 30,
			},
			expectUsers: true,
		},
		{
			name: "search repositories only",
			filter: SearchFilter{
				Query:   "test",
				Type:    "repository",
				Page:    1,
				PerPage: 30,
			},
			expectRepos: true,
		},
		{
			name: "search organizations only",
			filter: SearchFilter{
				Query:   "test",
				Type:    "organization",
				Page:    1,
				PerPage: 30,
			},
			expectOrgs: true,
		},
		{
			name: "search commits only",
			filter: SearchFilter{
				Query:   "test",
				Type:    "commit",
				Page:    1,
				PerPage: 30,
			},
			expectCommits: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := service.GlobalSearch(context.Background(), tt.filter)
			require.NoError(t, err)
			require.NotNil(t, results)

			if tt.expectUsers {
				assert.Greater(t, len(results.Users), 0, "Expected to find users")
			} else {
				assert.Equal(t, 0, len(results.Users), "Expected no users")
			}

			if tt.expectRepos {
				assert.Greater(t, len(results.Repositories), 0, "Expected to find repositories")
			} else {
				assert.Equal(t, 0, len(results.Repositories), "Expected no repositories")
			}

			if tt.expectOrgs {
				assert.Greater(t, len(results.Organizations), 0, "Expected to find organizations")
			} else {
				assert.Equal(t, 0, len(results.Organizations), "Expected no organizations")
			}

			if tt.expectCommits {
				assert.Greater(t, len(results.Commits), 0, "Expected to find commits")
			} else {
				assert.Equal(t, 0, len(results.Commits), "Expected no commits")
			}
		})
	}
}

func TestSearchService_SearchUsers(t *testing.T) {
	db := setupSearchTestDB(t)
	service := NewSearchService(db, nil, logrus.New())

	// Create test users
	user1 := models.User{
		ID:       uuid.New(),
		Username: "johndoe",
		FullName: "John Doe",
		Email:    "john@example.com",
		Bio:      "Software developer",
		Company:  "Acme Corp",
	}
	require.NoError(t, db.Create(&user1).Error)

	user2 := models.User{
		ID:       uuid.New(),
		Username: "janedoe",
		FullName: "Jane Doe",
		Email:    "jane@example.com",
		Bio:      "Product manager",
		Company:  "Tech Inc",
	}
	require.NoError(t, db.Create(&user2).Error)

	tests := []struct {
		name          string
		query         string
		expectedCount int
	}{
		{
			name:          "search by username",
			query:         "john",
			expectedCount: 1,
		},
		{
			name:          "search by full name",
			query:         "Jane Doe",
			expectedCount: 1,
		},
		{
			name:          "search by bio",
			query:         "developer",
			expectedCount: 1,
		},
		{
			name:          "search by company",
			query:         "Acme",
			expectedCount: 1,
		},
		{
			name:          "search with no results",
			query:         "nonexistent",
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users, err := service.searchUsers(SearchFilter{
				Query:   tt.query,
				Page:    1,
				PerPage: 30,
			}, 0)
			require.NoError(t, err)
			assert.Len(t, users, tt.expectedCount)
		})
	}
}

func TestSearchService_SearchRepositories(t *testing.T) {
	db := setupSearchTestDB(t)
	service := NewSearchService(db, nil, logrus.New())

	// Create test user
	user := models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
	}
	require.NoError(t, db.Create(&user).Error)

	// Create test repositories
	repo1 := models.Repository{
		ID:          uuid.New(),
		Name:        "awesome-project",
		Description: "An awesome web application",
		OwnerID:     user.ID,
		OwnerType:   "user",
		Visibility:  "public",
		StarsCount:  100,
		ForksCount:  25,
	}
	require.NoError(t, db.Create(&repo1).Error)

	repo2 := models.Repository{
		ID:          uuid.New(),
		Name:        "api-server",
		Description: "REST API server in Go",
		OwnerID:     user.ID,
		OwnerType:   "user",
		Visibility:  "private",
		StarsCount:  50,
		ForksCount:  10,
	}
	require.NoError(t, db.Create(&repo2).Error)

	tests := []struct {
		name          string
		query         string
		userID        *uuid.UUID
		expectedCount int
	}{
		{
			name:          "search by name",
			query:         "awesome",
			expectedCount: 1,
		},
		{
			name:          "search by description",
			query:         "API",
			expectedCount: 1,
		},
		{
			name:          "search with user authentication (should see private repos)",
			query:         "server",
			userID:        &user.ID,
			expectedCount: 1,
		},
		{
			name:          "search without authentication (only public repos)",
			query:         "server",
			expectedCount: 0, // private repo not visible
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repos, err := service.searchRepositories(SearchFilter{
				Query:   tt.query,
				Page:    1,
				PerPage: 30,
				UserID:  tt.userID,
			}, 0)
			require.NoError(t, err)
			assert.Len(t, repos, tt.expectedCount)
		})
	}
}

func TestSearchService_EmptyQuery(t *testing.T) {
	db := setupSearchTestDB(t)
	service := NewSearchService(db, nil, logrus.New())

	results, err := service.GlobalSearch(context.Background(), SearchFilter{
		Query:   "",
		Page:    1,
		PerPage: 30,
	})

	assert.Error(t, err)
	assert.Nil(t, results)
	assert.Contains(t, err.Error(), "search query cannot be empty")
}

func TestSearchService_Pagination(t *testing.T) {
	db := setupSearchTestDB(t)
	service := NewSearchService(db, nil, logrus.New())

	// Create multiple test users
	for i := 0; i < 35; i++ {
		user := models.User{
			ID:       uuid.New(),
			Username: fmt.Sprintf("testuser%d", i),
			FullName: fmt.Sprintf("Test User %d", i),
			Email:    fmt.Sprintf("test%d@example.com", i),
		}
		require.NoError(t, db.Create(&user).Error)
	}

	// Test first page
	results1, err := service.GlobalSearch(context.Background(), SearchFilter{
		Query:   "test",
		Type:    "user",
		Page:    1,
		PerPage: 30,
	})
	require.NoError(t, err)
	assert.Len(t, results1.Users, 30)

	// Test second page
	results2, err := service.GlobalSearch(context.Background(), SearchFilter{
		Query:   "test",
		Type:    "user",
		Page:    2,
		PerPage: 30,
	})
	require.NoError(t, err)
	assert.Len(t, results2.Users, 5) // Remaining 5 users

	// Ensure users are different between pages
	user1IDs := make(map[uuid.UUID]bool)
	for _, user := range results1.Users {
		user1IDs[user.ID] = true
	}

	for _, user := range results2.Users {
		assert.False(t, user1IDs[user.ID], "User should not appear on both pages")
	}
}
