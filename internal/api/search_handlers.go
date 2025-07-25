package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/a5c-ai/hub/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type SearchHandlers struct {
	searchService *services.SearchService
	logger        *logrus.Logger
}

func NewSearchHandlers(searchService *services.SearchService, logger *logrus.Logger) *SearchHandlers {
	return &SearchHandlers{
		searchService: searchService,
		logger:        logger,
	}
}

// GlobalSearch handles global search across all content types
// GET /api/v1/search
func (h *SearchHandlers) GlobalSearch(c *gin.Context) {
	// Parse query parameters
	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Search query is required",
		})
		return
	}

	// Build search filter
	filter := services.SearchFilter{
		Query:     query,
		Type:      c.Query("type"),
		Sort:      c.Query("sort"),
		Direction: c.Query("order"),
		Page:      1,
		PerPage:   30,
	}

	// Parse pagination
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filter.Page = page
		}
	}
	if perPageStr := c.Query("per_page"); perPageStr != "" {
		if perPage, err := strconv.Atoi(perPageStr); err == nil && perPage > 0 && perPage <= 100 {
			filter.PerPage = perPage
		}
	}

	// Get user ID from context if authenticated
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(uuid.UUID); ok {
			filter.UserID = &uid
		}
	}

	// Perform search
	results, err := h.searchService.GlobalSearch(filter)
	if err != nil {
		h.logger.WithError(err).Error("Failed to perform global search")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to perform search",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    results,
		"meta": gin.H{
			"query":     query,
			"type":      filter.Type,
			"page":      filter.Page,
			"per_page":  filter.PerPage,
			"total":     results.TotalCount,
		},
	})
}

// SearchRepositories handles repository-specific search
// GET /api/v1/search/repositories
func (h *SearchHandlers) SearchRepositories(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Search query is required",
		})
		return
	}

	// Build repository search filter
	filter := services.RepositorySearchFilter{
		SearchFilter: services.SearchFilter{
			Query:     query,
			Sort:      c.Query("sort"),
			Direction: c.Query("order"),
			Page:      1,
			PerPage:   30,
		},
		Owner:      c.Query("user"),
		Language:   c.Query("language"),
		Visibility: c.Query("visibility"),
		Stars:      c.Query("stars"),
		Forks:      c.Query("forks"),
		Size:       c.Query("size"),
		Created:    c.Query("created"),
		Updated:    c.Query("updated"),
		Pushed:     c.Query("pushed"),
	}

	// Parse pagination
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filter.Page = page
		}
	}
	if perPageStr := c.Query("per_page"); perPageStr != "" {
		if perPage, err := strconv.Atoi(perPageStr); err == nil && perPage > 0 && perPage <= 100 {
			filter.PerPage = perPage
		}
	}

	// Get user ID from context if authenticated
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(uuid.UUID); ok {
			filter.UserID = &uid
		}
	}

	// Perform repository search - use GlobalSearch with repository type
	searchFilter := services.SearchFilter{
		Query:     filter.Query,
		Type:      "repository",
		Sort:      filter.Sort,
		Direction: filter.Direction,
		Page:      filter.Page,
		PerPage:   filter.PerPage,
		UserID:    filter.UserID,
	}
	results, err := h.searchService.GlobalSearch(searchFilter)
	if err != nil {
		h.logger.WithError(err).Error("Failed to search repositories")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to search repositories",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    results.Repositories,
		"meta": gin.H{
			"query":     query,
			"page":      filter.Page,
			"per_page":  filter.PerPage,
			"total":     len(results.Repositories),
		},
	})
}

// SearchIssues handles issue and pull request search
// GET /api/v1/search/issues
func (h *SearchHandlers) SearchIssues(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Search query is required",
		})
		return
	}

	// Build issue search filter
	filter := services.IssueSearchFilter{
		SearchFilter: services.SearchFilter{
			Query:     query,
			Sort:      c.Query("sort"),
			Direction: c.Query("order"),
			Page:      1,
			PerPage:   30,
		},
		State:     c.Query("state"),
		Author:    c.Query("author"),
		Assignee:  c.Query("assignee"),
		Milestone: c.Query("milestone"),
	}

	// Parse labels
	if labelsStr := c.Query("labels"); labelsStr != "" {
		filter.Labels = strings.Split(labelsStr, ",")
	}

	// Parse is_pr flag
	if isPRStr := c.Query("is_pr"); isPRStr != "" {
		isPR := isPRStr == "true"
		filter.IsPR = &isPR
	}

	// Parse pagination
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filter.Page = page
		}
	}
	if perPageStr := c.Query("per_page"); perPageStr != "" {
		if perPage, err := strconv.Atoi(perPageStr); err == nil && perPage > 0 && perPage <= 100 {
			filter.PerPage = perPage
		}
	}

	// Get user ID from context if authenticated
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(uuid.UUID); ok {
			filter.UserID = &uid
		}
	}

	// Perform issue search - use GlobalSearch with issue type
	searchFilter := services.SearchFilter{
		Query:     filter.Query,
		Type:      "issue",
		Sort:      filter.Sort,
		Direction: filter.Direction,
		Page:      filter.Page,
		PerPage:   filter.PerPage,
		UserID:    filter.UserID,
	}
	results, err := h.searchService.GlobalSearch(searchFilter)
	if err != nil {
		h.logger.WithError(err).Error("Failed to search issues")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to search issues",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    results.Issues,
		"meta": gin.H{
			"query":     query,
			"page":      filter.Page,
			"per_page":  filter.PerPage,
			"total":     len(results.Issues),
		},
	})
}

// SearchUsers handles user search
// GET /api/v1/search/users
func (h *SearchHandlers) SearchUsers(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Search query is required",
		})
		return
	}

	// Build search filter
	filter := services.SearchFilter{
		Query:     query,
		Sort:      c.Query("sort"),
		Direction: c.Query("order"),
		Page:      1,
		PerPage:   30,
	}

	// Parse pagination
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filter.Page = page
		}
	}
	if perPageStr := c.Query("per_page"); perPageStr != "" {
		if perPage, err := strconv.Atoi(perPageStr); err == nil && perPage > 0 && perPage <= 100 {
			filter.PerPage = perPage
		}
	}

	// Perform user search - use GlobalSearch with user type
	filter.Type = "user"
	results, err := h.searchService.GlobalSearch(filter)
	if err != nil {
		h.logger.WithError(err).Error("Failed to search users")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to search users",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    results.Users,
		"meta": gin.H{
			"query":     query,
			"page":      filter.Page,
			"per_page":  filter.PerPage,
			"total":     len(results.Users),
		},
	})
}

// SearchCommits handles commit search
// GET /api/v1/search/commits
func (h *SearchHandlers) SearchCommits(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Search query is required",
		})
		return
	}

	// Build search filter
	filter := services.SearchFilter{
		Query:     query,
		Sort:      c.Query("sort"),
		Direction: c.Query("order"),
		Page:      1,
		PerPage:   30,
	}

	// Parse pagination
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filter.Page = page
		}
	}
	if perPageStr := c.Query("per_page"); perPageStr != "" {
		if perPage, err := strconv.Atoi(perPageStr); err == nil && perPage > 0 && perPage <= 100 {
			filter.PerPage = perPage
		}
	}

	// Get user ID from context if authenticated
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(uuid.UUID); ok {
			filter.UserID = &uid
		}
	}

	// Perform commit search - use GlobalSearch with commit type
	filter.Type = "commit"
	results, err := h.searchService.GlobalSearch(filter)
	if err != nil {
		h.logger.WithError(err).Error("Failed to search commits")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to search commits",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    results.Commits,
		"meta": gin.H{
			"query":     query,
			"page":      filter.Page,
			"per_page":  filter.PerPage,
			"total":     len(results.Commits),
		},
	})
}

// SearchInRepository handles repository-specific search
// GET /api/v1/repos/:owner/:repo/search
func (h *SearchHandlers) SearchInRepository(c *gin.Context) {
	owner := c.Param("owner")
	repo := c.Param("repo")
	query := strings.TrimSpace(c.Query("q"))

	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Search query is required",
		})
		return
	}

	// First, find the repository
	// This would need to be implemented with a repository service lookup
	// For now, we'll use a placeholder repository ID
	repositoryID := uuid.New() // This should be looked up from owner/repo

	// Build search filter
	filter := services.SearchFilter{
		Query:     query,
		Type:      c.Query("type"),
		Sort:      c.Query("sort"),
		Direction: c.Query("order"),
		Page:      1,
		PerPage:   30,
	}

	// Parse pagination
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filter.Page = page
		}
	}
	if perPageStr := c.Query("per_page"); perPageStr != "" {
		if perPage, err := strconv.Atoi(perPageStr); err == nil && perPage > 0 && perPage <= 100 {
			filter.PerPage = perPage
		}
	}

	// Get user ID from context if authenticated
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(uuid.UUID); ok {
			filter.UserID = &uid
		}
	}

	// Perform repository-specific search
	results, err := h.searchService.SearchInRepository(repositoryID, filter)
	if err != nil {
		h.logger.WithError(err).Error("Failed to search in repository")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to search in repository",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    results,
		"meta": gin.H{
			"repository": owner + "/" + repo,
			"query":      query,
			"type":       filter.Type,
			"page":       filter.Page,
			"per_page":   filter.PerPage,
			"total":      results.TotalCount,
		},
	})
}