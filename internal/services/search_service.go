package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/a5c-ai/hub/internal/models"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type SearchService struct {
	db     *gorm.DB
	logger *logrus.Logger
}

// SearchResults represents the aggregated search results
type SearchResults struct {
	Users         []models.User         `json:"users"`
	Repositories  []models.Repository   `json:"repositories"`
	Organizations []models.Organization `json:"organizations"`
	Commits       []models.Commit       `json:"commits"`
	TotalCount    int64                 `json:"total_count"`
}

// SearchFilter represents search filtering options
type SearchFilter struct {
	Query     string     `json:"query"`
	Type      string     `json:"type"`      // user, repository, organization, commit, code
	Sort      string     `json:"sort"`      // relevance, created, updated, stars, forks
	Direction string     `json:"direction"` // asc, desc
	Page      int        `json:"page"`
	PerPage   int        `json:"per_page"`
	UserID    *uuid.UUID `json:"user_id,omitempty"` // For permission filtering
}

func NewSearchService(db *gorm.DB, elasticsearch interface{}, logger *logrus.Logger) *SearchService {
	return &SearchService{
		db:     db,
		logger: logger,
	}
}

// GlobalSearch performs a global search across all content types
func (s *SearchService) GlobalSearch(ctx context.Context, filter SearchFilter) (*SearchResults, error) {
	if filter.Query == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}
	results := &SearchResults{}

	// Simple database search
	switch filter.Type {
	case "user":
		users, err := s.searchUsers(filter, (filter.Page-1)*filter.PerPage)
		if err != nil {
			return nil, err
		}
		results.Users = users
	case "repository":
		repos, err := s.searchRepositories(filter, (filter.Page-1)*filter.PerPage)
		if err != nil {
			return nil, err
		}
		results.Repositories = repos
	case "organization":
		orgs, err := s.searchOrganizations(filter, (filter.Page-1)*filter.PerPage)
		if err != nil {
			return nil, err
		}
		results.Organizations = orgs
	case "commit":
		commits, err := s.searchCommits(filter, (filter.Page-1)*filter.PerPage)
		if err != nil {
			return nil, err
		}
		results.Commits = commits
	default:
		// Search all types for a general search
		users, _ := s.searchUsers(filter, 0)
		repos, _ := s.searchRepositories(filter, 0)
		orgs, _ := s.searchOrganizations(filter, 0)
		commits, _ := s.searchCommits(filter, 0)

		// Limit results for overview
		if len(users) > 5 {
			users = users[:5]
		}
		if len(repos) > 5 {
			repos = repos[:5]
		}
		if len(orgs) > 5 {
			orgs = orgs[:5]
		}
		if len(commits) > 5 {
			commits = commits[:5]
		}

		results.Users = users
		results.Repositories = repos
		results.Organizations = orgs
		results.Commits = commits
	}

	results.TotalCount = int64(len(results.Users) + len(results.Repositories) + len(results.Organizations) + len(results.Commits))
	return results, nil
}

func (s *SearchService) searchUsers(filter SearchFilter, offset int) ([]models.User, error) {
	var users []models.User
	query := s.db.Model(&models.User{})

	if filter.Query != "" {
		q := "%" + strings.ToLower(filter.Query) + "%"
		query = query.Where(
			"lower(username) LIKE ? OR lower(full_name) LIKE ? OR lower(email) LIKE ? OR lower(bio) LIKE ? OR lower(company) LIKE ?",
			q, q, q, q, q,
		)
	}

	switch filter.Sort {
	case "created":
		if filter.Direction == "asc" {
			query = query.Order("created_at ASC")
		} else {
			query = query.Order("created_at DESC")
		}
	default:
		query = query.Order("username ASC")
	}

	query = query.Offset(offset).Limit(filter.PerPage)
	return users, query.Find(&users).Error
}

func (s *SearchService) searchRepositories(filter SearchFilter, offset int) ([]models.Repository, error) {
	var repos []models.Repository
	query := s.db.Model(&models.Repository{})

	// Only show public repositories for unauthenticated users
	if filter.UserID == nil {
		query = query.Where("visibility = 'public'")
	} else {
		query = query.Where("visibility = 'public' OR owner_id = ?", *filter.UserID)
	}

	if filter.Query != "" {
		q := "%" + strings.ToLower(filter.Query) + "%"
		query = query.Where(
			"lower(name) LIKE ? OR lower(description) LIKE ?",
			q, q,
		)
	}

	switch filter.Sort {
	case "updated":
		if filter.Direction == "asc" {
			query = query.Order("updated_at ASC")
		} else {
			query = query.Order("updated_at DESC")
		}
	case "stars":
		query = query.Order("stars_count DESC")
	case "forks":
		query = query.Order("forks_count DESC")
	default:
		query = query.Order("name ASC")
	}

	query = query.Offset(offset).Limit(filter.PerPage)
	return repos, query.Find(&repos).Error
}

func (s *SearchService) searchOrganizations(filter SearchFilter, offset int) ([]models.Organization, error) {
	var orgs []models.Organization
	query := s.db.Model(&models.Organization{})

	if filter.Query != "" {
		q := "%" + strings.ToLower(filter.Query) + "%"
		query = query.Where(
			"lower(name) LIKE ? OR lower(description) LIKE ?",
			q, q,
		)
	}

	query = query.Order("name ASC")
	query = query.Offset(offset).Limit(filter.PerPage)
	return orgs, query.Find(&orgs).Error
}

func (s *SearchService) searchCommits(filter SearchFilter, offset int) ([]models.Commit, error) {
	var commits []models.Commit
	query := s.db.Model(&models.Commit{})

	if filter.Query != "" {
		q := "%" + strings.ToLower(filter.Query) + "%"
		query = query.Where(
			"lower(message) LIKE ? OR lower(author_name) LIKE ?",
			q, q,
		)
	}

	query = query.Order("created_at DESC")
	query = query.Offset(offset).Limit(filter.PerPage)
	return commits, query.Preload("Repository").Find(&commits).Error
}
