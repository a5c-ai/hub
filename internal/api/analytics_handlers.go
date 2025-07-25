package api

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/a5c-ai/hub/internal/models"
	"github.com/a5c-ai/hub/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// AnalyticsHandlers contains handlers for analytics-related endpoints
type AnalyticsHandlers struct {
	analyticsService services.AnalyticsService
	logger           *logrus.Logger
	db               *gorm.DB
}

// NewAnalyticsHandlers creates a new analytics handlers instance
func NewAnalyticsHandlers(analyticsService services.AnalyticsService, logger *logrus.Logger, db *gorm.DB) *AnalyticsHandlers {
	return &AnalyticsHandlers{
		analyticsService: analyticsService,
		logger:           logger,
		db:               db,
	}
}

// Repository Analytics Endpoints

// GetRepositoryAnalytics handles GET /api/v1/repositories/:owner/:repo/analytics
func (h *AnalyticsHandlers) GetRepositoryAnalytics(c *gin.Context) {
	owner := c.Param("owner")
	repo := c.Param("repo")

	if owner == "" || repo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner and repository name are required"})
		return
	}

	// Parse query parameters
	period := services.Period(c.DefaultQuery("period", "daily"))
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	var startDate, endDate *time.Time
	if startDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			startDate = &parsed
		}
	}
	if endDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			endDate = &parsed
		}
	}

	// Resolve repository ID from owner/repo
	repoID, err := h.getRepositoryID(c.Request.Context(), owner, repo)
	if err != nil {
		h.logger.WithError(err).Error("Failed to resolve repository")
		c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		return
	}

	filters := services.InsightFilters{
		StartDate: startDate,
		EndDate:   endDate,
		Period:    period,
	}

	insights, err := h.analyticsService.GetRepositoryInsights(c.Request.Context(), repoID, filters)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get repository analytics")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository analytics"})
		return
	}

	c.JSON(http.StatusOK, insights)
}

// GetRepositoryCodeStats handles GET /api/v1/repositories/:owner/:repo/analytics/code-stats
func (h *AnalyticsHandlers) GetRepositoryCodeStats(c *gin.Context) {
	owner := c.Param("owner")
	repo := c.Param("repo")

	if owner == "" || repo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner and repository name are required"})
		return
	}

	// Resolve repository ID from owner/repo
	repoID, err := h.getRepositoryID(c.Request.Context(), owner, repo)
	if err != nil {
		h.logger.WithError(err).Error("Failed to resolve repository")
		c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		return
	}

	codeStats, err := h.analyticsService.GetRepositoryCodeStats(c.Request.Context(), repoID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get repository code stats")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get code statistics"})
		return
	}

	c.JSON(http.StatusOK, codeStats)
}

// GetRepositoryContributors handles GET /api/v1/repositories/:owner/:repo/analytics/contributors
func (h *AnalyticsHandlers) GetRepositoryContributors(c *gin.Context) {
	owner := c.Param("owner")
	repo := c.Param("repo")

	if owner == "" || repo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner and repository name are required"})
		return
	}

	// Parse query parameters
	period := services.Period(c.DefaultQuery("period", "daily"))
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	var startDate, endDate *time.Time
	if startDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			startDate = &parsed
		}
	}
	if endDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			endDate = &parsed
		}
	}

	// Resolve repository ID from owner/repo
	repoID, err := h.getRepositoryID(c.Request.Context(), owner, repo)
	if err != nil {
		h.logger.WithError(err).Error("Failed to resolve repository")
		c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		return
	}

	filters := services.InsightFilters{
		StartDate: startDate,
		EndDate:   endDate,
		Period:    period,
	}

	contributorStats, err := h.analyticsService.GetRepositoryContributorStats(c.Request.Context(), repoID, filters)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get repository contributor stats")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get contributor analytics"})
		return
	}

	c.JSON(http.StatusOK, contributorStats)
}

// GetRepositoryActivity handles GET /api/v1/repositories/:owner/:repo/analytics/activity
func (h *AnalyticsHandlers) GetRepositoryActivity(c *gin.Context) {
	owner := c.Param("owner")
	repo := c.Param("repo")

	if owner == "" || repo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner and repository name are required"})
		return
	}

	// Parse query parameters
	period := services.Period(c.DefaultQuery("period", "daily"))
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	var startDate, endDate *time.Time
	if startDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			startDate = &parsed
		}
	}
	if endDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			endDate = &parsed
		}
	}

	// Resolve repository ID from owner/repo
	repoID, err := h.getRepositoryID(c.Request.Context(), owner, repo)
	if err != nil {
		h.logger.WithError(err).Error("Failed to resolve repository")
		c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		return
	}

	filters := services.InsightFilters{
		StartDate: startDate,
		EndDate:   endDate,
		Period:    period,
	}

	activityStats, err := h.analyticsService.GetRepositoryActivityStats(c.Request.Context(), repoID, filters)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get repository activity stats")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get activity analytics"})
		return
	}

	c.JSON(http.StatusOK, activityStats)
}

// GetRepositoryPerformance handles GET /api/v1/repositories/:owner/:repo/analytics/performance
func (h *AnalyticsHandlers) GetRepositoryPerformance(c *gin.Context) {
	owner := c.Param("owner")
	repo := c.Param("repo")

	if owner == "" || repo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner and repository name are required"})
		return
	}

	// Parse query parameters
	period := services.Period(c.DefaultQuery("period", "daily"))
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	var startDate, endDate *time.Time
	if startDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			startDate = &parsed
		}
	}
	if endDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			endDate = &parsed
		}
	}

	// Resolve repository ID from owner/repo
	repoID, err := h.getRepositoryID(c.Request.Context(), owner, repo)
	if err != nil {
		h.logger.WithError(err).Error("Failed to resolve repository")
		c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		return
	}

	filters := services.InsightFilters{
		StartDate: startDate,
		EndDate:   endDate,
		Period:    period,
	}

	performanceStats, err := h.analyticsService.GetRepositoryPerformanceStats(c.Request.Context(), repoID, filters)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get repository performance stats")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get performance analytics"})
		return
	}

	c.JSON(http.StatusOK, performanceStats)
}

// GetRepositoryIssues handles GET /api/v1/repositories/:owner/:repo/analytics/issues
func (h *AnalyticsHandlers) GetRepositoryIssues(c *gin.Context) {
	owner := c.Param("owner")
	repo := c.Param("repo")

	if owner == "" || repo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner and repository name are required"})
		return
	}

	// Parse query parameters
	period := services.Period(c.DefaultQuery("period", "daily"))
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	var startDate, endDate *time.Time
	if startDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			startDate = &parsed
		}
	}
	if endDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			endDate = &parsed
		}
	}

	// Resolve repository ID from owner/repo
	repoID, err := h.getRepositoryID(c.Request.Context(), owner, repo)
	if err != nil {
		h.logger.WithError(err).Error("Failed to resolve repository")
		c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		return
	}

	filters := services.InsightFilters{
		StartDate: startDate,
		EndDate:   endDate,
		Period:    period,
	}

	issueStats, err := h.analyticsService.GetRepositoryIssueStats(c.Request.Context(), repoID, filters)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get repository issue stats")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get issue analytics"})
		return
	}

	c.JSON(http.StatusOK, issueStats)
}

// GetRepositoryPulls handles GET /api/v1/repositories/:owner/:repo/analytics/pulls
func (h *AnalyticsHandlers) GetRepositoryPulls(c *gin.Context) {
	owner := c.Param("owner")
	repo := c.Param("repo")

	if owner == "" || repo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner and repository name are required"})
		return
	}

	// Parse query parameters
	period := services.Period(c.DefaultQuery("period", "daily"))
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	var startDate, endDate *time.Time
	if startDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			startDate = &parsed
		}
	}
	if endDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			endDate = &parsed
		}
	}

	// Resolve repository ID from owner/repo
	repoID, err := h.getRepositoryID(c.Request.Context(), owner, repo)
	if err != nil {
		h.logger.WithError(err).Error("Failed to resolve repository")
		c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		return
	}

	filters := services.InsightFilters{
		StartDate: startDate,
		EndDate:   endDate,
		Period:    period,
	}

	prStats, err := h.analyticsService.GetRepositoryPRStats(c.Request.Context(), repoID, filters)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get repository PR stats")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get pull request analytics"})
		return
	}

	c.JSON(http.StatusOK, prStats)
}

// User Analytics Endpoints

// GetUserAnalytics handles GET /api/v1/user/analytics/activity
func (h *AnalyticsHandlers) GetUserAnalytics(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	uid, err := parseUserID(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Parse query parameters
	period := services.Period(c.DefaultQuery("period", "daily"))
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	var startDate, endDate *time.Time
	if startDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			startDate = &parsed
		}
	}
	if endDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			endDate = &parsed
		}
	}

	filters := services.InsightFilters{
		StartDate: startDate,
		EndDate:   endDate,
		Period:    period,
	}

	insights, err := h.analyticsService.GetUserInsights(c.Request.Context(), uid, filters)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get user analytics")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user analytics"})
		return
	}

	c.JSON(http.StatusOK, insights)
}

// GetUserContributions handles GET /api/v1/user/analytics/contributions
func (h *AnalyticsHandlers) GetUserContributions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	_, err := parseUserID(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// TODO: Implementation for user contributions
	c.JSON(http.StatusNotImplemented, gin.H{"error": "User contributions analytics not implemented yet"})
}

// GetUserRepositories handles GET /api/v1/user/analytics/repositories
func (h *AnalyticsHandlers) GetUserRepositories(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	_, err := parseUserID(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// TODO: Implementation for user repository analytics
	c.JSON(http.StatusNotImplemented, gin.H{"error": "User repository analytics not implemented yet"})
}

// GetPublicUserAnalytics handles GET /api/v1/users/:username/analytics/public
func (h *AnalyticsHandlers) GetPublicUserAnalytics(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username is required"})
		return
	}

	// TODO: Get user ID from username and implement public analytics
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Public user analytics not implemented yet"})
}

// Organization Analytics Endpoints

// GetOrganizationAnalytics handles GET /api/v1/organizations/:org/analytics/overview
func (h *AnalyticsHandlers) GetOrganizationAnalytics(c *gin.Context) {
	orgName := c.Param("org")
	if orgName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Organization name is required"})
		return
	}

	// TODO: Get organization ID from name and implement analytics
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Organization analytics not implemented yet"})
}

// GetOrganizationMembers handles GET /api/v1/organizations/:org/analytics/members
func (h *AnalyticsHandlers) GetOrganizationMembers(c *gin.Context) {
	orgName := c.Param("org")
	if orgName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Organization name is required"})
		return
	}

	// TODO: Implementation for organization member analytics
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Organization member analytics not implemented yet"})
}

// GetOrganizationRepositories handles GET /api/v1/organizations/:org/analytics/repositories
func (h *AnalyticsHandlers) GetOrganizationRepositories(c *gin.Context) {
	orgName := c.Param("org")
	if orgName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Organization name is required"})
		return
	}

	// TODO: Implementation for organization repository analytics
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Organization repository analytics not implemented yet"})
}

// GetOrganizationTeams handles GET /api/v1/organizations/:org/analytics/teams
func (h *AnalyticsHandlers) GetOrganizationTeams(c *gin.Context) {
	orgName := c.Param("org")
	if orgName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Organization name is required"})
		return
	}

	// TODO: Implementation for organization team analytics
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Organization team analytics not implemented yet"})
}

// GetOrganizationSecurity handles GET /api/v1/organizations/:org/analytics/security
func (h *AnalyticsHandlers) GetOrganizationSecurity(c *gin.Context) {
	orgName := c.Param("org")
	if orgName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Organization name is required"})
		return
	}

	// TODO: Implementation for organization security analytics
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Organization security analytics not implemented yet"})
}

// Admin Analytics Endpoints

// GetPlatformAnalytics handles GET /api/v1/admin/analytics/platform
func (h *AnalyticsHandlers) GetPlatformAnalytics(c *gin.Context) {
	// Check if user is admin
	if !h.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	// Parse query parameters
	period := services.Period(c.DefaultQuery("period", "daily"))
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	var startDate, endDate *time.Time
	if startDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			startDate = &parsed
		}
	}
	if endDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			endDate = &parsed
		}
	}

	filters := services.InsightFilters{
		StartDate: startDate,
		EndDate:   endDate,
		Period:    period,
	}

	insights, err := h.analyticsService.GetSystemInsights(c.Request.Context(), filters)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get platform analytics")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get platform analytics"})
		return
	}

	c.JSON(http.StatusOK, insights)
}

// GetUsageAnalytics handles GET /api/v1/admin/analytics/usage
func (h *AnalyticsHandlers) GetUsageAnalytics(c *gin.Context) {
	if !h.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	// TODO: Implementation for usage analytics
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Usage analytics not implemented yet"})
}

// GetPerformanceAnalytics handles GET /api/v1/admin/analytics/performance
func (h *AnalyticsHandlers) GetPerformanceAnalytics(c *gin.Context) {
	if !h.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	// Parse query parameters for performance filters
	limitStr := c.DefaultQuery("limit", "100")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	filters := services.PerformanceFilters{
		Limit:  limit,
		Offset: offset,
	}

	// Add date filters if provided
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			filters.StartDate = &parsed
		}
	}
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			filters.EndDate = &parsed
		}
	}

	metrics, err := h.analyticsService.GetPerformanceMetrics(c.Request.Context(), filters)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get performance analytics")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get performance analytics"})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetCostAnalytics handles GET /api/v1/admin/analytics/costs
func (h *AnalyticsHandlers) GetCostAnalytics(c *gin.Context) {
	if !h.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	// TODO: Implementation for cost analytics
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Cost analytics not implemented yet"})
}

// ExportAnalytics handles GET /api/v1/admin/analytics/export
func (h *AnalyticsHandlers) ExportAnalytics(c *gin.Context) {
	if !h.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	exportType := services.ExportType(c.DefaultQuery("format", "json"))
	dataType := c.DefaultQuery("data_type", "events")

	filters := services.ExportFilters{
		Type:           exportType,
		DataType:       dataType,
		IncludeHeaders: true,
	}

	// Add date filters if provided
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			filters.StartDate = &parsed
		}
	}
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			filters.EndDate = &parsed
		}
	}

	data, err := h.analyticsService.ExportData(c.Request.Context(), exportType, filters)
	if err != nil {
		h.logger.WithError(err).Error("Failed to export analytics data")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to export analytics data"})
		return
	}

	// Set appropriate content type based on export format
	var contentType string
	var filename string
	switch exportType {
	case services.ExportTypeCSV:
		contentType = "text/csv"
		filename = "analytics.csv"
	case services.ExportTypeJSON:
		contentType = "application/json"
		filename = "analytics.json"
	case services.ExportTypeXLSX:
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
		filename = "analytics.xlsx"
	default:
		contentType = "application/octet-stream"
		filename = "analytics.data"
	}

	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, contentType, data)
}

// Event Recording Endpoints (for internal use)

// RecordEvent handles POST /api/v1/analytics/events (internal)
func (h *AnalyticsHandlers) RecordEvent(c *gin.Context) {
	var event models.AnalyticsEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event data", "details": err.Error()})
		return
	}

	if err := h.analyticsService.RecordEvent(c.Request.Context(), &event); err != nil {
		h.logger.WithError(err).Error("Failed to record analytics event")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record event"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Event recorded successfully"})
}

// RecordMetric handles POST /api/v1/analytics/metrics (internal)
func (h *AnalyticsHandlers) RecordMetric(c *gin.Context) {
	var metric models.AnalyticsMetric
	if err := c.ShouldBindJSON(&metric); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid metric data", "details": err.Error()})
		return
	}

	if err := h.analyticsService.RecordMetric(c.Request.Context(), &metric); err != nil {
		h.logger.WithError(err).Error("Failed to record analytics metric")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record metric"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Metric recorded successfully"})
}

// RecordPerformanceLog handles POST /api/v1/analytics/performance (internal)
func (h *AnalyticsHandlers) RecordPerformanceLog(c *gin.Context) {
	var log models.PerformanceLog
	if err := c.ShouldBindJSON(&log); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid performance log data", "details": err.Error()})
		return
	}

	if err := h.analyticsService.RecordPerformanceLog(c.Request.Context(), &log); err != nil {
		h.logger.WithError(err).Error("Failed to record performance log")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record performance log"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Performance log recorded successfully"})
}

// Helper functions

// parseUserID parses user ID from context
func parseUserID(userID interface{}) (uuid.UUID, error) {
	switch v := userID.(type) {
	case uuid.UUID:
		return v, nil
	case string:
		return uuid.Parse(v)
	default:
		return uuid.Nil, fmt.Errorf("invalid user ID type")
	}
}

// isAdmin checks if the current user is an admin
func (h *AnalyticsHandlers) isAdmin(c *gin.Context) bool {
	// TODO: Implement admin check - this should check user role from context or database
	// For now, return false as a placeholder
	return false
}

// getRepositoryID resolves repository ID from owner and repository name
func (h *AnalyticsHandlers) getRepositoryID(ctx context.Context, owner, name string) (uuid.UUID, error) {
	// This needs to integrate with repository service
	// For now, we'll do a direct database lookup
	
	// First, resolve the owner name to owner ID and type
	var ownerID uuid.UUID
	var ownerType string
	
	// Try to find a user with this username
	var user struct {
		ID uuid.UUID `json:"id"`
	}
	err := h.db.WithContext(ctx).
		Model(&models.User{}).Select("id").Where("username = ?", owner).First(&user).Error
	if err == nil {
		ownerID = user.ID
		ownerType = "user"
	} else if err == gorm.ErrRecordNotFound {
		// Try to find an organization with this name
		var org struct {
			ID uuid.UUID `json:"id"`
		}
		err = h.db.WithContext(ctx).
			Model(&models.Organization{}).Select("id").Where("name = ?", owner).First(&org).Error
		if err == nil {
			ownerID = org.ID
			ownerType = "organization"
		} else if err == gorm.ErrRecordNotFound {
			return uuid.Nil, fmt.Errorf("owner not found")
		} else {
			return uuid.Nil, fmt.Errorf("failed to find organization: %w", err)
		}
	} else {
		return uuid.Nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Now find the repository
	var repo struct {
		ID uuid.UUID `json:"id"`
	}
	
	query := h.db.WithContext(ctx).
		Model(&models.Repository{}).Select("id").
		Where("name = ? AND owner_id = ?", name, ownerID)
		
	if ownerType == "user" {
		query = query.Where("owner_type = ?", "user")
	} else {
		query = query.Where("owner_type = ?", "organization")
	}
	
	err = query.First(&repo).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return uuid.Nil, fmt.Errorf("repository not found")
		}
		return uuid.Nil, fmt.Errorf("failed to find repository: %w", err)
	}

	return repo.ID, nil
}
