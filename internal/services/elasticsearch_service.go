package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/a5c-ai/hub/internal/config"
	"github.com/a5c-ai/hub/internal/models"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type ElasticsearchService struct {
	client      *elasticsearch.Client
	config      *config.Elasticsearch
	logger      *logrus.Logger
	indexPrefix string
}

// Index names
const (
	IndexUsers         = "users"
	IndexRepositories  = "repositories"
	IndexIssues        = "issues"
	IndexCommits       = "commits"
	IndexOrganizations = "organizations"
	IndexCode          = "code"
)

// Document types for Elasticsearch
type UserDocument struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	FullName     string    `json:"full_name"`
	Email        string    `json:"email"`
	Bio          string    `json:"bio"`
	Company      string    `json:"company"`
	Location     string    `json:"location"`
	Website      string    `json:"website"`
	AvatarURL    string    `json:"avatar_url"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Repositories int       `json:"repositories_count"`
	Followers    int       `json:"followers_count"`
	Following    int       `json:"following_count"`
}

type RepositoryDocument struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	FullName        string    `json:"full_name"`
	Description     string    `json:"description"`
	OwnerID         string    `json:"owner_id"`
	OwnerUsername   string    `json:"owner_username"`
	OwnerType       string    `json:"owner_type"`
	Visibility      string    `json:"visibility"`
	PrimaryLanguage string    `json:"primary_language"`
	Languages       []string  `json:"languages"`
	Topics          []string  `json:"topics"`
	StarsCount      int       `json:"stars_count"`
	ForksCount      int       `json:"forks_count"`
	WatchersCount   int       `json:"watchers_count"`
	Size            int       `json:"size"`
	DefaultBranch   string    `json:"default_branch"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	PushedAt        time.Time `json:"pushed_at"`
	IsTemplate      bool      `json:"is_template"`
	IsArchived      bool      `json:"is_archived"`
	IsFork          bool      `json:"is_fork"`
	HasIssues       bool      `json:"has_issues"`
	HasWiki         bool      `json:"has_wiki"`
	HasPages        bool      `json:"has_pages"`
}

type IssueDocument struct {
	ID           string    `json:"id"`
	Number       int       `json:"number"`
	Title        string    `json:"title"`
	Body         string    `json:"body"`
	State        string    `json:"state"`
	RepositoryID string    `json:"repository_id"`
	Repository   string    `json:"repository_name"`
	UserID       string    `json:"user_id"`
	Username     string    `json:"username"`
	AssigneeID   string    `json:"assignee_id"`
	AssigneeName string    `json:"assignee_name"`
	Labels       []string  `json:"labels"`
	Milestone    string    `json:"milestone"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	ClosedAt     *time.Time `json:"closed_at,omitempty"`
	CommentsCount int      `json:"comments_count"`
	IsPullRequest bool     `json:"is_pull_request"`
}

type CommitDocument struct {
	ID             string    `json:"id"`
	SHA            string    `json:"sha"`
	Message        string    `json:"message"`
	AuthorName     string    `json:"author_name"`
	AuthorEmail    string    `json:"author_email"`
	CommitterName  string    `json:"committer_name"`
	CommitterEmail string    `json:"committer_email"`
	AuthorDate     time.Time `json:"author_date"`
	CommitterDate  time.Time `json:"committer_date"`
	RepositoryID   string    `json:"repository_id"`
	Repository     string    `json:"repository_name"`
	Branch         string    `json:"branch"`
	ParentSHAs     []string  `json:"parent_shas"`
	FileChanges    int       `json:"file_changes"`
	Additions      int       `json:"additions"`
	Deletions      int       `json:"deletions"`
}

type OrganizationDocument struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	Website     string    `json:"website"`
	Email       string    `json:"email"`
	AvatarURL   string    `json:"avatar_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Members     int       `json:"members_count"`
	Teams       int       `json:"teams_count"`
	Repositories int      `json:"repositories_count"`
}

type CodeDocument struct {
	ID           string    `json:"id"`
	RepositoryID string    `json:"repository_id"`
	Repository   string    `json:"repository_name"`
	FilePath     string    `json:"file_path"`
	FileName     string    `json:"file_name"`
	FileExtension string   `json:"file_extension"`
	Language     string    `json:"language"`
	Content      string    `json:"content"`
	LineCount    int       `json:"line_count"`
	Size         int       `json:"size"`
	Branch       string    `json:"branch"`
	SHA          string    `json:"sha"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type SearchRequest struct {
	Query      string            `json:"query"`
	Indices    []string          `json:"indices,omitempty"`
	Size       int               `json:"size,omitempty"`
	From       int               `json:"from,omitempty"`
	Sort       []map[string]interface{} `json:"sort,omitempty"`
	Filters    map[string]interface{} `json:"filters,omitempty"`
	Highlight  map[string]interface{} `json:"highlight,omitempty"`
}

type SearchResponse struct {
	Hits struct {
		Total struct {
			Value int `json:"value"`
		} `json:"total"`
		Hits []struct {
			Index  string          `json:"_index"`
			ID     string          `json:"_id"`
			Score  float64         `json:"_score"`
			Source json.RawMessage `json:"_source"`
			Highlight map[string][]string `json:"highlight,omitempty"`
		} `json:"hits"`
	} `json:"hits"`
	Aggregations map[string]interface{} `json:"aggregations,omitempty"`
}

func NewElasticsearchService(cfg *config.Elasticsearch, logger *logrus.Logger) (*ElasticsearchService, error) {
	if !cfg.Enabled {
		logger.Info("Elasticsearch is disabled")
		return &ElasticsearchService{
			config:      cfg,
			logger:      logger,
			indexPrefix: cfg.IndexPrefix,
		}, nil
	}

	esConfig := elasticsearch.Config{
		Addresses: cfg.Addresses,
	}

	// Configure authentication
	if cfg.Username != "" && cfg.Password != "" {
		esConfig.Username = cfg.Username
		esConfig.Password = cfg.Password
	}

	if cfg.APIKey != "" {
		esConfig.APIKey = cfg.APIKey
	}

	if cfg.CloudID != "" {
		esConfig.CloudID = cfg.CloudID
	}

	client, err := elasticsearch.NewClient(esConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Elasticsearch client: %w", err)
	}

	// Test connection
	res, err := client.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Elasticsearch: %w", err)
	}
	res.Body.Close()

	logger.Info("Successfully connected to Elasticsearch")

	service := &ElasticsearchService{
		client:      client,
		config:      cfg,
		logger:      logger,
		indexPrefix: cfg.IndexPrefix,
	}

	// Initialize indices
	if err := service.InitializeIndices(); err != nil {
		logger.WithError(err).Warn("Failed to initialize Elasticsearch indices")
	}

	return service, nil
}

func (es *ElasticsearchService) IsEnabled() bool {
	return es.config.Enabled && es.client != nil
}

func (es *ElasticsearchService) getIndexName(index string) string {
	return fmt.Sprintf("%s_%s", es.indexPrefix, index)
}

// InitializeIndices creates all necessary indices with proper mappings
func (es *ElasticsearchService) InitializeIndices() error {
	if !es.IsEnabled() {
		return nil
	}

	indices := map[string]string{
		IndexUsers:         es.getUserMapping(),
		IndexRepositories:  es.getRepositoryMapping(),
		IndexIssues:        es.getIssueMapping(),
		IndexCommits:       es.getCommitMapping(),
		IndexOrganizations: es.getOrganizationMapping(),
		IndexCode:          es.getCodeMapping(),
	}

	for index, mapping := range indices {
		indexName := es.getIndexName(index)
		
		// Check if index exists
		req := esapi.IndicesExistsRequest{
			Index: []string{indexName},
		}
		
		res, err := req.Do(context.Background(), es.client)
		if err != nil {
			return fmt.Errorf("failed to check index %s: %w", indexName, err)
		}
		res.Body.Close()

		// Create index if it doesn't exist
		if res.StatusCode == 404 {
			req := esapi.IndicesCreateRequest{
				Index: indexName,
				Body:  strings.NewReader(mapping),
			}
			
			res, err := req.Do(context.Background(), es.client)
			if err != nil {
				return fmt.Errorf("failed to create index %s: %w", indexName, err)
			}
			res.Body.Close()

			if res.IsError() {
				return fmt.Errorf("failed to create index %s: %s", indexName, res.String())
			}

			es.logger.WithField("index", indexName).Info("Created Elasticsearch index")
		}
	}

	return nil
}

// Mapping definitions
func (es *ElasticsearchService) getUserMapping() string {
	return `{
		"mappings": {
			"properties": {
				"id": {"type": "keyword"},
				"username": {"type": "text", "analyzer": "standard"},
				"full_name": {"type": "text", "analyzer": "standard"},
				"email": {"type": "keyword"},
				"bio": {"type": "text", "analyzer": "standard"},
				"company": {"type": "text", "analyzer": "standard"},
				"location": {"type": "text", "analyzer": "standard"},
				"website": {"type": "keyword"},
				"avatar_url": {"type": "keyword"},
				"created_at": {"type": "date"},
				"updated_at": {"type": "date"},
				"repositories_count": {"type": "integer"},
				"followers_count": {"type": "integer"},
				"following_count": {"type": "integer"}
			}
		},
		"settings": {
			"number_of_shards": 1,
			"number_of_replicas": 0
		}
	}`
}

func (es *ElasticsearchService) getRepositoryMapping() string {
	return `{
		"mappings": {
			"properties": {
				"id": {"type": "keyword"},
				"name": {"type": "text", "analyzer": "standard"},
				"full_name": {"type": "text", "analyzer": "standard"},
				"description": {"type": "text", "analyzer": "standard"},
				"owner_id": {"type": "keyword"},
				"owner_username": {"type": "keyword"},
				"owner_type": {"type": "keyword"},
				"visibility": {"type": "keyword"},
				"primary_language": {"type": "keyword"},
				"languages": {"type": "keyword"},
				"topics": {"type": "keyword"},
				"stars_count": {"type": "integer"},
				"forks_count": {"type": "integer"},
				"watchers_count": {"type": "integer"},
				"size": {"type": "integer"},
				"default_branch": {"type": "keyword"},
				"created_at": {"type": "date"},
				"updated_at": {"type": "date"},
				"pushed_at": {"type": "date"},
				"is_template": {"type": "boolean"},
				"is_archived": {"type": "boolean"},
				"is_fork": {"type": "boolean"},
				"has_issues": {"type": "boolean"},
				"has_wiki": {"type": "boolean"},
				"has_pages": {"type": "boolean"}
			}
		},
		"settings": {
			"number_of_shards": 1,
			"number_of_replicas": 0
		}
	}`
}

func (es *ElasticsearchService) getIssueMapping() string {
	return `{
		"mappings": {
			"properties": {
				"id": {"type": "keyword"},
				"number": {"type": "integer"},
				"title": {"type": "text", "analyzer": "standard"},
				"body": {"type": "text", "analyzer": "standard"},
				"state": {"type": "keyword"},
				"repository_id": {"type": "keyword"},
				"repository_name": {"type": "keyword"},
				"user_id": {"type": "keyword"},
				"username": {"type": "keyword"},
				"assignee_id": {"type": "keyword"},
				"assignee_name": {"type": "keyword"},
				"labels": {"type": "keyword"},
				"milestone": {"type": "keyword"},
				"created_at": {"type": "date"},
				"updated_at": {"type": "date"},
				"closed_at": {"type": "date"},
				"comments_count": {"type": "integer"},
				"is_pull_request": {"type": "boolean"}
			}
		},
		"settings": {
			"number_of_shards": 1,
			"number_of_replicas": 0
		}
	}`
}

func (es *ElasticsearchService) getCommitMapping() string {
	return `{
		"mappings": {
			"properties": {
				"id": {"type": "keyword"},
				"sha": {"type": "keyword"},
				"message": {"type": "text", "analyzer": "standard"},
				"author_name": {"type": "text", "analyzer": "standard"},
				"author_email": {"type": "keyword"},
				"committer_name": {"type": "text", "analyzer": "standard"},
				"committer_email": {"type": "keyword"},
				"author_date": {"type": "date"},
				"committer_date": {"type": "date"},
				"repository_id": {"type": "keyword"},
				"repository_name": {"type": "keyword"},
				"branch": {"type": "keyword"},
				"parent_shas": {"type": "keyword"},
				"file_changes": {"type": "integer"},
				"additions": {"type": "integer"},
				"deletions": {"type": "integer"}
			}
		},
		"settings": {
			"number_of_shards": 1,
			"number_of_replicas": 0
		}
	}`
}

func (es *ElasticsearchService) getOrganizationMapping() string {
	return `{
		"mappings": {
			"properties": {
				"id": {"type": "keyword"},
				"name": {"type": "text", "analyzer": "standard"},
				"display_name": {"type": "text", "analyzer": "standard"},
				"description": {"type": "text", "analyzer": "standard"},
				"location": {"type": "text", "analyzer": "standard"},
				"website": {"type": "keyword"},
				"email": {"type": "keyword"},
				"avatar_url": {"type": "keyword"},
				"created_at": {"type": "date"},
				"updated_at": {"type": "date"},
				"members_count": {"type": "integer"},
				"teams_count": {"type": "integer"},
				"repositories_count": {"type": "integer"}
			}
		},
		"settings": {
			"number_of_shards": 1,
			"number_of_replicas": 0
		}
	}`
}

func (es *ElasticsearchService) getCodeMapping() string {
	return `{
		"mappings": {
			"properties": {
				"id": {"type": "keyword"},
				"repository_id": {"type": "keyword"},
				"repository_name": {"type": "keyword"},
				"file_path": {"type": "keyword"},
				"file_name": {"type": "text", "analyzer": "standard"},
				"file_extension": {"type": "keyword"},
				"language": {"type": "keyword"},
				"content": {"type": "text", "analyzer": "standard"},
				"line_count": {"type": "integer"},
				"size": {"type": "integer"},
				"branch": {"type": "keyword"},
				"sha": {"type": "keyword"},
				"updated_at": {"type": "date"}
			}
		},
		"settings": {
			"number_of_shards": 1,
			"number_of_replicas": 0,
			"analysis": {
				"analyzer": {
					"code_analyzer": {
						"tokenizer": "standard",
						"filter": ["lowercase", "code_filter"]
					}
				},
				"filter": {
					"code_filter": {
						"type": "pattern_replace",
						"pattern": "([a-z])([A-Z])",
						"replacement": "$1 $2"
					}
				}
			}
		}
	}`
}

// Search methods
func (es *ElasticsearchService) Search(req SearchRequest) (*SearchResponse, error) {
	if !es.IsEnabled() {
		return nil, fmt.Errorf("Elasticsearch is not enabled")
	}

	// Build Elasticsearch query
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  req.Query,
				"fields": []string{"*"},
				"type":   "best_fields",
				"fuzziness": "AUTO",
			},
		},
	}

	// Add filters
	if len(req.Filters) > 0 {
		filters := make([]map[string]interface{}, 0)
		for field, value := range req.Filters {
			filters = append(filters, map[string]interface{}{
				"term": map[string]interface{}{
					field: value,
				},
			})
		}

		query["query"] = map[string]interface{}{
			"bool": map[string]interface{}{
				"must": query["query"],
				"filter": filters,
			},
		}
	}

	// Add sorting
	if len(req.Sort) > 0 {
		query["sort"] = req.Sort
	} else {
		query["sort"] = []map[string]interface{}{
			{"_score": map[string]string{"order": "desc"}},
		}
	}

	// Add pagination
	if req.Size > 0 {
		query["size"] = req.Size
	} else {
		query["size"] = 20
	}

	if req.From > 0 {
		query["from"] = req.From
	}

	// Add highlighting
	if len(req.Highlight) > 0 {
		query["highlight"] = req.Highlight
	} else {
		query["highlight"] = map[string]interface{}{
			"fields": map[string]interface{}{
				"*": map[string]interface{}{},
			},
			"pre_tags":  []string{"<mark>"},
			"post_tags": []string{"</mark>"},
		}
	}

	// Determine indices to search
	indices := req.Indices
	if len(indices) == 0 {
		indices = []string{
			es.getIndexName(IndexUsers),
			es.getIndexName(IndexRepositories),
			es.getIndexName(IndexIssues),
			es.getIndexName(IndexCommits),
			es.getIndexName(IndexOrganizations),
			es.getIndexName(IndexCode),
		}
	}

	// Execute search
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, fmt.Errorf("failed to encode query: %w", err)
	}

	res, err := es.client.Search(
		es.client.Search.WithContext(context.Background()),
		es.client.Search.WithIndex(indices...),
		es.client.Search.WithBody(&buf),
		es.client.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("search error: %s", res.String())
	}

	var response SearchResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// Document indexing methods
func (es *ElasticsearchService) IndexUser(user *models.User) error {
	if !es.IsEnabled() {
		return nil
	}

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

	return es.indexDocument(IndexUsers, user.ID.String(), doc)
}

func (es *ElasticsearchService) IndexRepository(repo *models.Repository) error {
	if !es.IsEnabled() {
		return nil
	}

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

	return es.indexDocument(IndexRepositories, repo.ID.String(), doc)
}

func (es *ElasticsearchService) IndexIssue(issue *models.Issue) error {
	if !es.IsEnabled() {
		return nil
	}

	var labels []string
	for _, label := range issue.Labels {
		labels = append(labels, label.Name)
	}

	doc := IssueDocument{
		ID:           issue.ID.String(),
		Number:       issue.Number,
		Title:        issue.Title,
		Body:         issue.Body,
		State:        issue.State,
		RepositoryID: issue.RepositoryID.String(),
		Repository:   issue.Repository.Name,
		UserID:       issue.UserID.String(),
		Username:     issue.User.Username,
		Labels:       labels,
		CreatedAt:    issue.CreatedAt,
		UpdatedAt:    issue.UpdatedAt,
		ClosedAt:     issue.ClosedAt,
		CommentsCount: issue.CommentsCount,
	}

	if issue.AssigneeID != nil {
		doc.AssigneeID = issue.AssigneeID.String()
		if issue.Assignee != nil {
			doc.AssigneeName = issue.Assignee.Username
		}
	}

	return es.indexDocument(IndexIssues, issue.ID.String(), doc)
}

func (es *ElasticsearchService) IndexCommit(commit *models.Commit) error {
	if !es.IsEnabled() {
		return nil
	}

	doc := CommitDocument{
		ID:             commit.ID.String(),
		SHA:            commit.SHA,
		Message:        commit.Message,
		AuthorName:     commit.AuthorName,
		AuthorEmail:    commit.AuthorEmail,
		CommitterName:  commit.CommitterName,
		CommitterEmail: commit.CommitterEmail,
		AuthorDate:     commit.AuthorDate,
		CommitterDate:  commit.CommitterDate,
		RepositoryID:   commit.RepositoryID.String(),
		Repository:     commit.Repository.Name,
	}

	return es.indexDocument(IndexCommits, commit.ID.String(), doc)
}

func (es *ElasticsearchService) IndexOrganization(org *models.Organization) error {
	if !es.IsEnabled() {
		return nil
	}

	doc := OrganizationDocument{
		ID:          org.ID.String(),
		Name:        org.Name,
		DisplayName: org.DisplayName,
		Description: org.Description,
		Location:    org.Location,
		Website:     org.Website,
		Email:       org.Email,
		AvatarURL:   org.AvatarURL,
		CreatedAt:   org.CreatedAt,
		UpdatedAt:   org.UpdatedAt,
	}

	return es.indexDocument(IndexOrganizations, org.ID.String(), doc)
}

func (es *ElasticsearchService) IndexCode(repoID uuid.UUID, repoName, filePath, content, language, branch, sha string) error {
	if !es.IsEnabled() {
		return nil
	}

	doc := CodeDocument{
		ID:           fmt.Sprintf("%s:%s:%s", repoID.String(), branch, filePath),
		RepositoryID: repoID.String(),
		Repository:   repoName,
		FilePath:     filePath,
		FileName:     getFileName(filePath),
		FileExtension: getFileExtension(filePath),
		Language:     language,
		Content:      content,
		LineCount:    strings.Count(content, "\n") + 1,
		Size:         len(content),
		Branch:       branch,
		SHA:          sha,
		UpdatedAt:    time.Now(),
	}

	return es.indexDocument(IndexCode, doc.ID, doc)
}

// Helper methods
func (es *ElasticsearchService) indexDocument(index, id string, doc interface{}) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(doc); err != nil {
		return fmt.Errorf("failed to encode document: %w", err)
	}

	req := esapi.IndexRequest{
		Index:      es.getIndexName(index),
		DocumentID: id,
		Body:       &buf,
		Refresh:    "wait_for",
	}

	res, err := req.Do(context.Background(), es.client)
	if err != nil {
		return fmt.Errorf("failed to index document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("indexing error: %s", res.String())
	}

	return nil
}

func (es *ElasticsearchService) DeleteDocument(index, id string) error {
	if !es.IsEnabled() {
		return nil
	}

	req := esapi.DeleteRequest{
		Index:      es.getIndexName(index),
		DocumentID: id,
		Refresh:    "wait_for",
	}

	res, err := req.Do(context.Background(), es.client)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() && res.StatusCode != 404 {
		return fmt.Errorf("delete error: %s", res.String())
	}

	return nil
}

func (es *ElasticsearchService) BulkIndex(operations []map[string]interface{}) error {
	if !es.IsEnabled() {
		return nil
	}

	var buf bytes.Buffer
	for _, op := range operations {
		if err := json.NewEncoder(&buf).Encode(op); err != nil {
			return fmt.Errorf("failed to encode operation: %w", err)
		}
	}

	req := esapi.BulkRequest{
		Body:    &buf,
		Refresh: "wait_for",
	}

	res, err := req.Do(context.Background(), es.client)
	if err != nil {
		return fmt.Errorf("failed to execute bulk request: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("bulk error: %s", res.String())
	}

	return nil
}

// Utility functions
func getFileName(path string) string {
	parts := strings.Split(path, "/")
	return parts[len(parts)-1]
}

func getFileExtension(path string) string {
	parts := strings.Split(path, ".")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return ""
}

// Search result conversion helpers
func (es *ElasticsearchService) ConvertSearchResults(response *SearchResponse, filter SearchFilter) (*SearchResults, error) {
	results := &SearchResults{
		TotalCount: int64(response.Hits.Total.Value),
	}

	for _, hit := range response.Hits.Hits {
		switch {
		case strings.Contains(hit.Index, IndexUsers):
			var user UserDocument
			if err := json.Unmarshal(hit.Source, &user); err != nil {
				continue
			}
			results.Users = append(results.Users, es.convertUserDocument(user))

		case strings.Contains(hit.Index, IndexRepositories):
			var repo RepositoryDocument
			if err := json.Unmarshal(hit.Source, &repo); err != nil {
				continue
			}
			results.Repositories = append(results.Repositories, es.convertRepositoryDocument(repo))

		case strings.Contains(hit.Index, IndexIssues):
			var issue IssueDocument
			if err := json.Unmarshal(hit.Source, &issue); err != nil {
				continue
			}
			results.Issues = append(results.Issues, es.convertIssueDocument(issue))

		case strings.Contains(hit.Index, IndexCommits):
			var commit CommitDocument
			if err := json.Unmarshal(hit.Source, &commit); err != nil {
				continue
			}
			results.Commits = append(results.Commits, es.convertCommitDocument(commit))

		case strings.Contains(hit.Index, IndexOrganizations):
			var org OrganizationDocument
			if err := json.Unmarshal(hit.Source, &org); err != nil {
				continue
			}
			results.Organizations = append(results.Organizations, es.convertOrganizationDocument(org))
		}
	}

	return results, nil
}

func (es *ElasticsearchService) convertUserDocument(doc UserDocument) models.User {
	id, _ := uuid.Parse(doc.ID)
	return models.User{
		ID:        id,
		Username:  doc.Username,
		FullName:  doc.FullName,
		Email:     doc.Email,
		Bio:       doc.Bio,
		Company:   doc.Company,
		Location:  doc.Location,
		Website:   doc.Website,
		AvatarURL: doc.AvatarURL,
		CreatedAt: doc.CreatedAt,
		UpdatedAt: doc.UpdatedAt,
	}
}

func (es *ElasticsearchService) convertRepositoryDocument(doc RepositoryDocument) models.Repository {
	id, _ := uuid.Parse(doc.ID)
	ownerID, _ := uuid.Parse(doc.OwnerID)
	return models.Repository{
		ID:              id,
		Name:            doc.Name,
		Description:     doc.Description,
		OwnerID:         ownerID,
		OwnerType:       doc.OwnerType,
		Visibility:      doc.Visibility,
		PrimaryLanguage: doc.PrimaryLanguage,
		StarsCount:      doc.StarsCount,
		ForksCount:      doc.ForksCount,
		WatchersCount:   doc.WatchersCount,
		Size:            doc.Size,
		DefaultBranch:   doc.DefaultBranch,
		CreatedAt:       doc.CreatedAt,
		UpdatedAt:       doc.UpdatedAt,
		PushedAt:        doc.PushedAt,
		IsTemplate:      doc.IsTemplate,
		IsArchived:      doc.IsArchived,
		IsFork:          doc.IsFork,
		HasIssues:       doc.HasIssues,
		HasWiki:         doc.HasWiki,
		HasPages:        doc.HasPages,
		Owner:           models.User{Username: doc.OwnerUsername},
	}
}

func (es *ElasticsearchService) convertIssueDocument(doc IssueDocument) models.Issue {
	id, _ := uuid.Parse(doc.ID)
	repoID, _ := uuid.Parse(doc.RepositoryID)
	userID, _ := uuid.Parse(doc.UserID)
	
	issue := models.Issue{
		ID:           id,
		Number:       doc.Number,
		Title:        doc.Title,
		Body:         doc.Body,
		State:        doc.State,
		RepositoryID: repoID,
		UserID:       userID,
		CreatedAt:    doc.CreatedAt,
		UpdatedAt:    doc.UpdatedAt,
		ClosedAt:     doc.ClosedAt,
		CommentsCount: doc.CommentsCount,
		Repository:   models.Repository{Name: doc.Repository},
		User:         models.User{Username: doc.Username},
	}

	if doc.AssigneeID != "" {
		assigneeID, _ := uuid.Parse(doc.AssigneeID)
		issue.AssigneeID = &assigneeID
		if doc.AssigneeName != "" {
			issue.Assignee = &models.User{Username: doc.AssigneeName}
		}
	}

	// Convert labels
	for _, labelName := range doc.Labels {
		issue.Labels = append(issue.Labels, models.Label{Name: labelName})
	}

	return issue
}

func (es *ElasticsearchService) convertCommitDocument(doc CommitDocument) models.Commit {
	id, _ := uuid.Parse(doc.ID)
	repoID, _ := uuid.Parse(doc.RepositoryID)
	return models.Commit{
		ID:             id,
		SHA:            doc.SHA,
		Message:        doc.Message,
		AuthorName:     doc.AuthorName,
		AuthorEmail:    doc.AuthorEmail,
		CommitterName:  doc.CommitterName,
		CommitterEmail: doc.CommitterEmail,
		AuthorDate:     doc.AuthorDate,
		CommitterDate:  doc.CommitterDate,
		RepositoryID:   repoID,
		Repository:     models.Repository{Name: doc.Repository},
	}
}

func (es *ElasticsearchService) convertOrganizationDocument(doc OrganizationDocument) models.Organization {
	id, _ := uuid.Parse(doc.ID)
	return models.Organization{
		ID:          id,
		Name:        doc.Name,
		DisplayName: doc.DisplayName,
		Description: doc.Description,
		Location:    doc.Location,
		Website:     doc.Website,
		Email:       doc.Email,
		AvatarURL:   doc.AvatarURL,
		CreatedAt:   doc.CreatedAt,
		UpdatedAt:   doc.UpdatedAt,
	}
}