package services

import (
	"fmt"
	"strings"

	"github.com/a5c-ai/hub/internal/models"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type SearchService struct {
	db             *gorm.DB
	elasticsearch  *ElasticsearchService
	logger         *logrus.Logger
}

// SearchResults represents the aggregated search results
type SearchResults struct {
	Users         []models.User         `json:"users"`
	Repositories  []models.Repository   `json:"repositories"`
	Issues        []models.Issue        `json:"issues"`
	Organizations []models.Organization `json:"organizations"`
	Commits       []models.Commit       `json:"commits"`
	TotalCount    int64                 `json:"total_count"`
}

// SearchFilter represents search filtering options
type SearchFilter struct {
	Query      string `json:"query"`
	Type       string `json:"type"`       // user, repository, issue, organization, commit, code
	Sort       string `json:"sort"`       // relevance, created, updated, stars, forks
	Direction  string `json:"direction"`  // asc, desc
	Page       int    `json:"page"`
	PerPage    int    `json:"per_page"`
	UserID     *uuid.UUID `json:"user_id,omitempty"` // For permission filtering
}

// Repository-specific search filters
type RepositorySearchFilter struct {
	SearchFilter
	Owner      string `json:"owner"`
	Language   string `json:"language"`
	Visibility string `json:"visibility"` // public, private, internal
	Stars      string `json:"stars"`      // ">100", "<50", etc.
	Forks      string `json:"forks"`
	Size       string `json:"size"`
	Created    string `json:"created"`
	Updated    string `json:"updated"`
	Pushed     string `json:"pushed"`
}

// Code search filters
type CodeSearchFilter struct {
	SearchFilter
	RepositoryID *uuid.UUID `json:"repository_id,omitempty"`
	Path         string     `json:"path"`
	Extension    string     `json:"extension"`
	Language     string     `json:"language"`
	Size         string     `json:"size"`
}

// Issue search filters
type IssueSearchFilter struct {
	SearchFilter
	RepositoryID *uuid.UUID `json:"repository_id,omitempty"`
	State        string     `json:"state"`        // open, closed
	Author       string     `json:"author"`
	Assignee     string     `json:"assignee"`
	Labels       []string   `json:"labels"`
	Milestone    string     `json:"milestone"`
	IsPR         *bool      `json:"is_pr,omitempty"`
}

func NewSearchService(db *gorm.DB, elasticsearch *ElasticsearchService, logger *logrus.Logger) *SearchService {
	return &SearchService{
		db:            db,
		elasticsearch: elasticsearch,
		logger:        logger,
	}
}

// GlobalSearch performs a search across all searchable content
func (s *SearchService) GlobalSearch(filter SearchFilter) (*SearchResults, error) {
	if filter.Query == "" {
		return &SearchResults{}, fmt.Errorf("search query cannot be empty")
	}

	// Set defaults
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage < 1 || filter.PerPage > 100 {
		filter.PerPage = 30
	}

	// Try Elasticsearch first if enabled
	if s.elasticsearch != nil && s.elasticsearch.IsEnabled() {
		return s.searchWithElasticsearch(filter)
	}

	// Fallback to PostgreSQL search
	return s.searchWithPostgreSQL(filter)
}

// searchWithElasticsearch uses Elasticsearch for advanced search
func (s *SearchService) searchWithElasticsearch(filter SearchFilter) (*SearchResults, error) {
	// Build Elasticsearch request
	req := SearchRequest{
		Query: filter.Query,
		Size:  filter.PerPage,
		From:  (filter.Page - 1) * filter.PerPage,
	}

	// Add type-specific indices
	if filter.Type != "" {
		switch filter.Type {
		case "user":
			req.Indices = []string{s.elasticsearch.getIndexName(IndexUsers)}
		case "repository":
			req.Indices = []string{s.elasticsearch.getIndexName(IndexRepositories)}
		case "issue":
			req.Indices = []string{s.elasticsearch.getIndexName(IndexIssues)}
		case "commit":
			req.Indices = []string{s.elasticsearch.getIndexName(IndexCommits)}
		case "organization":
			req.Indices = []string{s.elasticsearch.getIndexName(IndexOrganizations)}
		case "code":
			req.Indices = []string{s.elasticsearch.getIndexName(IndexCode)}
		}
	}

	// Add filters
	filters := make(map[string]interface{})
	
	// Permission filtering for repositories and issues
	if filter.UserID != nil {
		// For repositories: public or owned by user
		if filter.Type == "repository" || filter.Type == "" {
			// This would need more complex query building in practice
		}
	} else {
		// Only public content for unauthenticated users
		if filter.Type == "repository" || filter.Type == "" {
			filters["visibility"] = "public"
		}
	}

	if len(filters) > 0 {
		req.Filters = filters
	}

	// Add sorting
	if filter.Sort != "" {
		switch filter.Sort {
		case "created":
			order := "desc"
			if filter.Direction == "asc" {
				order = "asc"
			}
			req.Sort = []map[string]interface{}{
				{"created_at": map[string]string{"order": order}},
			}
		case "updated":
			order := "desc"
			if filter.Direction == "asc" {
				order = "asc"
			}
			req.Sort = []map[string]interface{}{
				{"updated_at": map[string]string{"order": order}},
			}
		case "stars":
			order := "desc"
			if filter.Direction == "asc" {
				order = "asc"
			}
			req.Sort = []map[string]interface{}{
				{"stars_count": map[string]string{"order": order}},
			}
		}
	}

	// Execute search
	response, err := s.elasticsearch.Search(req)
	if err != nil {
		s.logger.WithError(err).Error("Elasticsearch search failed, falling back to PostgreSQL")
		return s.searchWithPostgreSQL(filter)
	}

	// Convert results
	return s.elasticsearch.ConvertSearchResults(response, filter)
}

// searchWithPostgreSQL uses PostgreSQL for fallback search
func (s *SearchService) searchWithPostgreSQL(filter SearchFilter) (*SearchResults, error) {
	results := &SearchResults{}
	offset := (filter.Page - 1) * filter.PerPage

	// Search across different types based on filter
	switch filter.Type {
	case "user":
		users, err := s.searchUsers(filter, offset)
		if err != nil {
			return nil, err
		}
		results.Users = users
	case "repository":
		repos, err := s.searchRepositories(RepositorySearchFilter{SearchFilter: filter}, offset)
		if err != nil {
			return nil, err
		}
		results.Repositories = repos
	case "issue":
		issues, err := s.searchIssues(IssueSearchFilter{SearchFilter: filter}, offset)
		if err != nil {
			return nil, err
		}
		results.Issues = issues
	case "organization":
		orgs, err := s.searchOrganizations(filter, offset)
		if err != nil {
			return nil, err
		}
		results.Organizations = orgs
	case "commit":
		commits, err := s.searchCommits(filter, offset)
		if err != nil {
			return nil, err
		}
		results.Commits = commits
	default:
		// Search all types with limited results per type
		users, _ := s.searchUsers(filter, 0)
		repos, _ := s.searchRepositories(RepositorySearchFilter{SearchFilter: filter}, 0)
		issues, _ := s.searchIssues(IssueSearchFilter{SearchFilter: filter}, 0)
		orgs, _ := s.searchOrganizations(filter, 0)
		commits, _ := s.searchCommits(filter, 0)

		// Limit results per type for global search
		if len(users) > 5 {
			users = users[:5]
		}
		if len(repos) > 5 {
			repos = repos[:5]
		}
		if len(issues) > 5 {
			issues = issues[:5]
		}
		if len(orgs) > 5 {
			orgs = orgs[:5]
		}
		if len(commits) > 5 {
			commits = commits[:5]
		}

		results.Users = users
		results.Repositories = repos
		results.Issues = issues
		results.Organizations = orgs
		results.Commits = commits
	}

	results.TotalCount = int64(len(results.Users) + len(results.Repositories) + len(results.Issues) + len(results.Organizations) + len(results.Commits))

	return results, nil
}

// Indexing methods for automatic document indexing
func (s *SearchService) IndexUser(user *models.User) error {
	if s.elasticsearch != nil && s.elasticsearch.IsEnabled() {
		return s.elasticsearch.IndexUser(user)
	}
	return nil
}

func (s *SearchService) IndexRepository(repo *models.Repository) error {
	if s.elasticsearch != nil && s.elasticsearch.IsEnabled() {
		return s.elasticsearch.IndexRepository(repo)
	}
	return nil
}

func (s *SearchService) IndexIssue(issue *models.Issue) error {
	if s.elasticsearch != nil && s.elasticsearch.IsEnabled() {
		return s.elasticsearch.IndexIssue(issue)
	}
	return nil
}

func (s *SearchService) IndexCommit(commit *models.Commit) error {
	if s.elasticsearch != nil && s.elasticsearch.IsEnabled() {
		return s.elasticsearch.IndexCommit(commit)
	}
	return nil
}

func (s *SearchService) IndexOrganization(org *models.Organization) error {
	if s.elasticsearch != nil && s.elasticsearch.IsEnabled() {
		return s.elasticsearch.IndexOrganization(org)
	}
	return nil
}

func (s *SearchService) DeleteUser(userID uuid.UUID) error {
	if s.elasticsearch != nil && s.elasticsearch.IsEnabled() {
		return s.elasticsearch.DeleteDocument(IndexUsers, userID.String())
	}
	return nil
}

func (s *SearchService) DeleteRepository(repoID uuid.UUID) error {
	if s.elasticsearch != nil && s.elasticsearch.IsEnabled() {
		return s.elasticsearch.DeleteDocument(IndexRepositories, repoID.String())
	}
	return nil
}

func (s *SearchService) DeleteIssue(issueID uuid.UUID) error {
	if s.elasticsearch != nil && s.elasticsearch.IsEnabled() {
		return s.elasticsearch.DeleteDocument(IndexIssues, issueID.String())
	}
	return nil
}

func (s *SearchService) DeleteCommit(commitID uuid.UUID) error {
	if s.elasticsearch != nil && s.elasticsearch.IsEnabled() {
		return s.elasticsearch.DeleteDocument(IndexCommits, commitID.String())
	}
	return nil
}

func (s *SearchService) DeleteOrganization(orgID uuid.UUID) error {
	if s.elasticsearch != nil && s.elasticsearch.IsEnabled() {
		return s.elasticsearch.DeleteDocument(IndexOrganizations, orgID.String())
	}
	return nil
}

// Code search methods
func (s *SearchService) SearchCode(query string, repoID *uuid.UUID, language string, page, perPage int) (*SearchResults, error) {
	if s.elasticsearch == nil || !s.elasticsearch.IsEnabled() {
		return &SearchResults{}, fmt.Errorf("code search requires Elasticsearch")
	}

	req := SearchRequest{
		Query:   query,
		Indices: []string{s.elasticsearch.getIndexName(IndexCode)},
		Size:    perPage,
		From:    (page - 1) * perPage,
	}

	filters := make(map[string]interface{})
	if repoID != nil {
		filters["repository_id"] = repoID.String()
	}
	if language != "" {
		filters["language"] = language
	}

	if len(filters) > 0 {
		req.Filters = filters
	}

	response, err := s.elasticsearch.Search(req)
	if err != nil {
		return nil, err
	}

	// For code search, we'll return a simplified result structure
	// In practice, you might want a separate CodeSearchResult type
	results := &SearchResults{
		TotalCount: int64(response.Hits.Total.Value),
	}

	return results, nil
}

func (s *SearchService) IndexCodeFile(repoID uuid.UUID, repoName, filePath, content, language, branch, sha string) error {
	if s.elasticsearch != nil && s.elasticsearch.IsEnabled() {
		return s.elasticsearch.IndexCode(repoID, repoName, filePath, content, language, branch, sha)
	}
	return nil
}

func (s *SearchService) DeleteCodeFile(repoID uuid.UUID, branch, filePath string) error {
	if s.elasticsearch != nil && s.elasticsearch.IsEnabled() {
		documentID := fmt.Sprintf("%s:%s:%s", repoID.String(), branch, filePath)
		return s.elasticsearch.DeleteDocument(IndexCode, documentID)
	}
	return nil
}

// Bulk indexing for initial data import
func (s *SearchService) BulkIndexUsers(users []models.User) error {
	if s.elasticsearch == nil || !s.elasticsearch.IsEnabled() {
		return nil
	}

	operations := make([]map[string]interface{}, 0, len(users)*2)
	for _, user := range users {
		// Index operation metadata
		operations = append(operations, map[string]interface{}{
			"index": map[string]interface{}{
				"_index": s.elasticsearch.getIndexName(IndexUsers),
				"_id":    user.ID.String(),
			},
		})
		
		// Document data
		doc := UserDocument{
			ID:        user.ID.String(),
			Username:  user.Username,
			FullName:  user.FullName,
			Email:     user.Email,
			Bio:       user.Bio,
			Company:   user.Company,
			Location:  user.Location,
			Website:   user.Website,
			AvatarURL: user.AvatarURL,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		}
		operations = append(operations, map[string]interface{}(map[string]interface{}{
			"id":         doc.ID,
			"username":   doc.Username,
			"full_name":  doc.FullName,
			"email":      doc.Email,
			"bio":        doc.Bio,
			"company":    doc.Company,
			"location":   doc.Location,
			"website":    doc.Website,
			"avatar_url": doc.AvatarURL,
			"created_at": doc.CreatedAt,
			"updated_at": doc.UpdatedAt,
		}))
	}

	return s.elasticsearch.BulkIndex(operations)
}

func (s *SearchService) BulkIndexRepositories(repos []models.Repository) error {
	if s.elasticsearch == nil || !s.elasticsearch.IsEnabled() {
		return nil
	}

	operations := make([]map[string]interface{}, 0, len(repos)*2)
	for _, repo := range repos {
		// Index operation metadata
		operations = append(operations, map[string]interface{}{
			"index": map[string]interface{}{
				"_index": s.elasticsearch.getIndexName(IndexRepositories),
				"_id":    repo.ID.String(),
			},
		})
		
		// Document data
		doc := RepositoryDocument{
			ID:              repo.ID.String(),
			Name:            repo.Name,
			FullName:        fmt.Sprintf("%s/%s", repo.Owner.Username, repo.Name),
			Description:     repo.Description,
			OwnerID:         repo.OwnerID.String(),
			OwnerUsername:   repo.Owner.Username,
			OwnerType:       repo.OwnerType,
			Visibility:      repo.Visibility,
			PrimaryLanguage: repo.PrimaryLanguage,
			StarsCount:      repo.StarsCount,
			ForksCount:      repo.ForksCount,
			WatchersCount:   repo.WatchersCount,
			Size:            repo.Size,
			DefaultBranch:   repo.DefaultBranch,
			CreatedAt:       repo.CreatedAt,
			UpdatedAt:       repo.UpdatedAt,
			PushedAt:        repo.PushedAt,
			IsTemplate:      repo.IsTemplate,
			IsArchived:      repo.IsArchived,
			IsFork:          repo.IsFork,
			HasIssues:       repo.HasIssues,
			HasWiki:         repo.HasWiki,
			HasPages:        repo.HasPages,
		}
		
		operations = append(operations, map[string]interface{}(map[string]interface{}{
			"id":               doc.ID,
			"name":             doc.Name,
			"full_name":        doc.FullName,
			"description":      doc.Description,
			"owner_id":         doc.OwnerID,
			"owner_username":   doc.OwnerUsername,
			"owner_type":       doc.OwnerType,
			"visibility":       doc.Visibility,
			"primary_language": doc.PrimaryLanguage,
			"stars_count":      doc.StarsCount,
			"forks_count":      doc.ForksCount,
			"watchers_count":   doc.WatchersCount,
			"size":             doc.Size,
			"default_branch":   doc.DefaultBranch,
			"created_at":       doc.CreatedAt,
			"updated_at":       doc.UpdatedAt,
			"pushed_at":        doc.PushedAt,
			"is_template":      doc.IsTemplate,
			"is_archived":      doc.IsArchived,
			"is_fork":          doc.IsFork,
			"has_issues":       doc.HasIssues,
			"has_wiki":         doc.HasWiki,
			"has_pages":        doc.HasPages,
		}))
	}

	return s.elasticsearch.BulkIndex(operations)
}

// Reindex all data from database to Elasticsearch
func (s *SearchService) ReindexAll() error {
	if s.elasticsearch == nil || !s.elasticsearch.IsEnabled() {
		return fmt.Errorf("Elasticsearch is not enabled")
	}

	s.logger.Info("Starting full reindex of search data")

	// Reindex users
	var users []models.User
	if err := s.db.Find(&users).Error; err != nil {
		return fmt.Errorf("failed to fetch users: %w", err)
	}
	if err := s.BulkIndexUsers(users); err != nil {
		return fmt.Errorf("failed to reindex users: %w", err)
	}
	s.logger.WithField("count", len(users)).Info("Reindexed users")

	// Reindex repositories
	var repos []models.Repository
	if err := s.db.Preload("Owner").Find(&repos).Error; err != nil {
		return fmt.Errorf("failed to fetch repositories: %w", err)
	}
	if err := s.BulkIndexRepositories(repos); err != nil {
		return fmt.Errorf("failed to reindex repositories: %w", err)
	}
	s.logger.WithField("count", len(repos)).Info("Reindexed repositories")

	// Reindex issues
	var issues []models.Issue
	if err := s.db.Preload("User").Preload("Repository").Preload("Labels").Preload("Assignee").Find(&issues).Error; err != nil {
		return fmt.Errorf("failed to fetch issues: %w", err)
	}
	for _, issue := range issues {
		if err := s.IndexIssue(&issue); err != nil {
			s.logger.WithError(err).WithField("issue_id", issue.ID).Warn("Failed to index issue")
		}
	}
	s.logger.WithField("count", len(issues)).Info("Reindexed issues")

	// Reindex organizations
	var orgs []models.Organization
	if err := s.db.Find(&orgs).Error; err != nil {
		return fmt.Errorf("failed to fetch organizations: %w", err)
	}
	for _, org := range orgs {
		if err := s.IndexOrganization(&org); err != nil {
			s.logger.WithError(err).WithField("org_id", org.ID).Warn("Failed to index organization")
		}
	}
	s.logger.WithField("count", len(orgs)).Info("Reindexed organizations")

	// Reindex commits
	var commits []models.Commit
	if err := s.db.Preload("Repository").Find(&commits).Error; err != nil {
		return fmt.Errorf("failed to fetch commits: %w", err)
	}
	for _, commit := range commits {
		if err := s.IndexCommit(&commit); err != nil {
			s.logger.WithError(err).WithField("commit_id", commit.ID).Warn("Failed to index commit")
		}
	}
	s.logger.WithField("count", len(commits)).Info("Reindexed commits")

	s.logger.Info("Completed full reindex of search data")
	return nil
}

// searchUsers searches for users using PostgreSQL full-text search
func (s *SearchService) searchUsers(filter SearchFilter, offset int) ([]models.User, error) {
	var users []models.User
	
	query := s.db.Model(&models.User{})
	
	// Use PostgreSQL full-text search on searchable fields
	searchQuery := strings.TrimSpace(filter.Query)
	if searchQuery != "" {
		// Create a full-text search vector from multiple fields
		query = query.Where(`
			to_tsvector('english', coalesce(username, '') || ' ' || 
			                     coalesce(full_name, '') || ' ' || 
			                     coalesce(email, '') || ' ' || 
			                     coalesce(bio, '') || ' ' || 
			                     coalesce(company, '')) 
			@@ plainto_tsquery('english', ?)`, searchQuery)
	}

	// Apply sorting
	switch filter.Sort {
	case "created":
		if filter.Direction == "asc" {
			query = query.Order("created_at ASC")
		} else {
			query = query.Order("created_at DESC")
		}
	case "updated":
		if filter.Direction == "asc" {
			query = query.Order("updated_at ASC")
		} else {
			query = query.Order("updated_at DESC")
		}
	default:
		// Default to relevance-based ordering using ts_rank
		if searchQuery != "" {
			query = query.Select("*, ts_rank(to_tsvector('english', coalesce(username, '') || ' ' || coalesce(full_name, '') || ' ' || coalesce(bio, '')), plainto_tsquery('english', ?)) as relevance", searchQuery)
			query = query.Order("relevance DESC")
		} else {
			query = query.Order("created_at DESC")
		}
	}

	// Apply pagination
	if filter.PerPage > 0 {
		query = query.Limit(filter.PerPage).Offset(offset)
	}

	return users, query.Find(&users).Error
}

// searchRepositories searches for repositories
func (s *SearchService) searchRepositories(filter RepositorySearchFilter, offset int) ([]models.Repository, error) {
	var repositories []models.Repository
	
	query := s.db.Model(&models.Repository{})
	
	// Apply visibility filter based on user permissions
	if filter.UserID != nil {
		// Complex query to filter based on permissions - for now, simplified
		query = query.Where("visibility = 'public' OR owner_id = ?", *filter.UserID)
	} else {
		// Only public repositories for unauthenticated users
		query = query.Where("visibility = 'public'")
	}

	searchQuery := strings.TrimSpace(filter.Query)
	if searchQuery != "" {
		// Use full-text search on name and description
		query = query.Where(`
			to_tsvector('english', coalesce(name, '') || ' ' || coalesce(description, '')) 
			@@ plainto_tsquery('english', ?)`, searchQuery)
	}

	// Apply additional filters
	if filter.Owner != "" {
		query = query.Where("owner_id IN (SELECT id FROM users WHERE username = ? UNION SELECT id FROM organizations WHERE name = ?)", filter.Owner, filter.Owner)
	}
	if filter.Language != "" {
		query = query.Where("primary_language = ?", filter.Language)
	}
	if filter.Visibility != "" && (filter.Visibility == "public" || filter.Visibility == "private" || filter.Visibility == "internal") {
		query = query.Where("visibility = ?", filter.Visibility)
	}

	// Apply sorting
	switch filter.Sort {
	case "name":
		if filter.Direction == "asc" {
			query = query.Order("name ASC")
		} else {
			query = query.Order("name DESC")
		}
	case "created":
		if filter.Direction == "asc" {
			query = query.Order("created_at ASC")
		} else {
			query = query.Order("created_at DESC")
		}
	case "updated":
		if filter.Direction == "asc" {
			query = query.Order("updated_at ASC")
		} else {
			query = query.Order("updated_at DESC")
		}
	case "pushed":
		if filter.Direction == "asc" {
			query = query.Order("pushed_at ASC")
		} else {
			query = query.Order("pushed_at DESC")
		}
	case "stars":
		if filter.Direction == "asc" {
			query = query.Order("stars_count ASC")
		} else {
			query = query.Order("stars_count DESC")
		}
	case "forks":
		if filter.Direction == "asc" {
			query = query.Order("forks_count ASC")
		} else {
			query = query.Order("forks_count DESC")
		}
	default:
		// Default to relevance-based ordering
		if searchQuery != "" {
			query = query.Select("*, ts_rank(to_tsvector('english', coalesce(name, '') || ' ' || coalesce(description, '')), plainto_tsquery('english', ?)) as relevance", searchQuery)
			query = query.Order("relevance DESC")
		} else {
			query = query.Order("stars_count DESC")
		}
	}

	// Apply pagination
	if filter.PerPage > 0 {
		query = query.Limit(filter.PerPage).Offset(offset)
	}

	// Preload owner information
	query = query.Preload("Owner")

	return repositories, query.Find(&repositories).Error
}

// searchIssues searches for issues and pull requests
func (s *SearchService) searchIssues(filter IssueSearchFilter, offset int) ([]models.Issue, error) {
	var issues []models.Issue
	
	query := s.db.Model(&models.Issue{})

	// Apply repository filter
	if filter.RepositoryID != nil {
		query = query.Where("repository_id = ?", *filter.RepositoryID)
	} else if filter.UserID != nil {
		// Only show issues from accessible repositories
		query = query.Where(`repository_id IN (
			SELECT id FROM repositories 
			WHERE visibility = 'public' OR owner_id = ?
		)`, *filter.UserID)
	} else {
		// Only public repositories for unauthenticated users
		query = query.Where(`repository_id IN (
			SELECT id FROM repositories WHERE visibility = 'public'
		)`)
	}

	searchQuery := strings.TrimSpace(filter.Query)
	if searchQuery != "" {
		query = query.Where(`
			to_tsvector('english', coalesce(title, '') || ' ' || coalesce(body, '')) 
			@@ plainto_tsquery('english', ?)`, searchQuery)
	}

	// Apply filters
	if filter.State != "" {
		query = query.Where("state = ?", filter.State)
	}
	if filter.Author != "" {
		query = query.Where("user_id IN (SELECT id FROM users WHERE username = ?)", filter.Author)
	}
	if filter.Assignee != "" {
		query = query.Where("assignee_id IN (SELECT id FROM users WHERE username = ?)", filter.Assignee)
	}
	if filter.IsPR != nil {
		if *filter.IsPR {
			query = query.Where("id IN (SELECT issue_id FROM pull_requests)")
		} else {
			query = query.Where("id NOT IN (SELECT issue_id FROM pull_requests)")
		}
	}

	// Apply labels filter
	if len(filter.Labels) > 0 {
		query = query.Where(`id IN (
			SELECT issue_id FROM issue_labels il 
			JOIN labels l ON il.label_id = l.id 
			WHERE l.name IN ?
			GROUP BY issue_id 
			HAVING COUNT(DISTINCT l.name) = ?
		)`, filter.Labels, len(filter.Labels))
	}

	// Apply sorting
	switch filter.Sort {
	case "created":
		if filter.Direction == "asc" {
			query = query.Order("created_at ASC")
		} else {
			query = query.Order("created_at DESC")
		}
	case "updated":
		if filter.Direction == "asc" {
			query = query.Order("updated_at ASC")
		} else {
			query = query.Order("updated_at DESC")
		}
	case "comments":
		if filter.Direction == "asc" {
			query = query.Order("comments_count ASC")
		} else {
			query = query.Order("comments_count DESC")
		}
	default:
		if searchQuery != "" {
			query = query.Select("*, ts_rank(to_tsvector('english', coalesce(title, '') || ' ' || coalesce(body, '')), plainto_tsquery('english', ?)) as relevance", searchQuery)
			query = query.Order("relevance DESC")
		} else {
			query = query.Order("created_at DESC")
		}
	}

	// Apply pagination
	if filter.PerPage > 0 {
		query = query.Limit(filter.PerPage).Offset(offset)
	}

	// Preload related data
	query = query.Preload("User").Preload("Repository").Preload("Labels")

	return issues, query.Find(&issues).Error
}

// searchOrganizations searches for organizations
func (s *SearchService) searchOrganizations(filter SearchFilter, offset int) ([]models.Organization, error) {
	var organizations []models.Organization
	
	query := s.db.Model(&models.Organization{})

	searchQuery := strings.TrimSpace(filter.Query)
	if searchQuery != "" {
		query = query.Where(`
			to_tsvector('english', coalesce(name, '') || ' ' || 
			                     coalesce(display_name, '') || ' ' || 
			                     coalesce(description, '')) 
			@@ plainto_tsquery('english', ?)`, searchQuery)
	}

	// Apply sorting
	switch filter.Sort {
	case "created":
		if filter.Direction == "asc" {
			query = query.Order("created_at ASC")
		} else {
			query = query.Order("created_at DESC")
		}
	case "updated":
		if filter.Direction == "asc" {
			query = query.Order("updated_at ASC")
		} else {
			query = query.Order("updated_at DESC")
		}
	default:
		if searchQuery != "" {
			query = query.Select("*, ts_rank(to_tsvector('english', coalesce(name, '') || ' ' || coalesce(display_name, '') || ' ' || coalesce(description, '')), plainto_tsquery('english', ?)) as relevance", searchQuery)
			query = query.Order("relevance DESC")
		} else {
			query = query.Order("created_at DESC")
		}
	}

	// Apply pagination
	if filter.PerPage > 0 {
		query = query.Limit(filter.PerPage).Offset(offset)
	}

	return organizations, query.Find(&organizations).Error
}

// searchCommits searches for commits
func (s *SearchService) searchCommits(filter SearchFilter, offset int) ([]models.Commit, error) {
	var commits []models.Commit
	
	query := s.db.Model(&models.Commit{})

	// Only search in accessible repositories
	if filter.UserID != nil {
		query = query.Where(`repository_id IN (
			SELECT id FROM repositories 
			WHERE visibility = 'public' OR owner_id = ?
		)`, *filter.UserID)
	} else {
		query = query.Where(`repository_id IN (
			SELECT id FROM repositories WHERE visibility = 'public'
		)`)
	}

	searchQuery := strings.TrimSpace(filter.Query)
	if searchQuery != "" {
		query = query.Where(`
			to_tsvector('english', coalesce(message, '') || ' ' || 
			                     coalesce(author_name, '') || ' ' || 
			                     coalesce(committer_name, '')) 
			@@ plainto_tsquery('english', ?)`, searchQuery)
	}

	// Apply sorting
	switch filter.Sort {
	case "author-date":
		if filter.Direction == "asc" {
			query = query.Order("author_date ASC")
		} else {
			query = query.Order("author_date DESC")
		}
	case "committer-date":
		if filter.Direction == "asc" {
			query = query.Order("committer_date ASC")
		} else {
			query = query.Order("committer_date DESC")
		}
	default:
		if searchQuery != "" {
			query = query.Select("*, ts_rank(to_tsvector('english', coalesce(message, '') || ' ' || coalesce(author_name, '')), plainto_tsquery('english', ?)) as relevance", searchQuery)
			query = query.Order("relevance DESC")
		} else {
			query = query.Order("committer_date DESC")
		}
	}

	// Apply pagination
	if filter.PerPage > 0 {
		query = query.Limit(filter.PerPage).Offset(offset)
	}

	// Preload repository information
	query = query.Preload("Repository")

	return commits, query.Find(&commits).Error
}

// SearchInRepository performs a repository-specific search
func (s *SearchService) SearchInRepository(repositoryID uuid.UUID, filter SearchFilter) (*SearchResults, error) {
	results := &SearchResults{}
	
	// Set repository ID in filter
	issueFilter := IssueSearchFilter{SearchFilter: filter, RepositoryID: &repositoryID}
	
	// Search within the repository
	switch filter.Type {
	case "issue":
		issues, err := s.searchIssues(issueFilter, (filter.Page-1)*filter.PerPage)
		if err != nil {
			return nil, err
		}
		results.Issues = issues
	case "commit":
		filter.Query = fmt.Sprintf("repo:%s %s", repositoryID.String(), filter.Query)
		commits, err := s.searchCommits(filter, (filter.Page-1)*filter.PerPage)
		if err != nil {
			return nil, err
		}
		results.Commits = commits
	default:
		// Search all types within repository
		issues, _ := s.searchIssues(issueFilter, 0)
		commits, _ := s.searchCommits(filter, 0)
		
		// Limit results
		if len(issues) > 10 {
			issues = issues[:10]
		}
		if len(commits) > 10 {
			commits = commits[:10]
		}
		
		results.Issues = issues
		results.Commits = commits
	}
	
	results.TotalCount = int64(len(results.Issues) + len(results.Commits))
	
	return results, nil
}