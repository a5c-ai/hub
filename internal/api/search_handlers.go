package api

import (
	"net/http"
	"strconv"

	"github.com/a5c-ai/hub/internal/services"
	"github.com/gin-gonic/gin"
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

// GlobalSearch handles GET /api/v1/search
func (h *SearchHandlers) GlobalSearch(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
		return
	}

	searchType := c.DefaultQuery("type", "")
	page := 1
	perPage := 30

	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if pp := c.Query("per_page"); pp != "" {
		if parsed, err := strconv.Atoi(pp); err == nil && parsed > 0 && parsed <= 100 {
			perPage = parsed
		}
	}

	filter := services.SearchFilter{
		Query:   query,
		Type:    searchType,
		Page:    page,
		PerPage: perPage,
	}

	// Perform search
	results, err := h.searchService.GlobalSearch(c.Request.Context(), filter)
	if err != nil {
		h.logger.WithError(err).Error("Failed to perform global search")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to perform search",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": results,
		"meta": gin.H{
			"query":    query,
			"type":     filter.Type,
			"page":     filter.Page,
			"per_page": filter.PerPage,
			"total":    results.TotalCount,
		},
	})
}
