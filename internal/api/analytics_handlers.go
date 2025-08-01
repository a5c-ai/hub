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

	c.JSON(http.StatusOK, gin.H{"error": "Issue analytics not available"})
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

	uid, err := parseUserID(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get user contributions analytics
	contributions, err := h.getUserContributions(c.Request.Context(), uid)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get user contributions")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user contributions"})
		return
	}

	c.JSON(http.StatusOK, contributions)
}

// GetUserRepositories handles GET /api/v1/user/analytics/repositories
func (h *AnalyticsHandlers) GetUserRepositories(c *gin.Context) {
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
		h.logger.WithError(err).Error("Failed to get user repository analytics")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user repository analytics"})
		return
	}

	// Return just the repository stats portion
	c.JSON(http.StatusOK, insights.RepositoryStats)
}

// GetPublicUserAnalytics handles GET /api/v1/users/:username/analytics/public
func (h *AnalyticsHandlers) GetPublicUserAnalytics(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username is required"})
		return
	}

	// Get user ID from username
	var user models.User
	if err := h.db.WithContext(c.Request.Context()).Where("username = ?", username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		h.logger.WithError(err).Error("Failed to find user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find user"})
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

	// Get public user analytics (limited dataset)
	insights, err := h.analyticsService.GetUserInsights(c.Request.Context(), user.ID, filters)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get public user analytics")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get public user analytics"})
		return
	}

	// Return only public information
	publicData := gin.H{
		"username":           user.Username,
		"contribution_stats": insights.ContributionStats,
		"repository_stats": gin.H{
			"total_repositories": insights.RepositoryStats.TotalRepositories,
			"total_stars":        insights.RepositoryStats.TotalStars,
			"total_forks":        insights.RepositoryStats.TotalForks,
		},
	}

	c.JSON(http.StatusOK, publicData)
}

// Organization Analytics Endpoints

// GetOrganizationAnalytics handles GET /api/v1/organizations/:org/analytics/overview
func (h *AnalyticsHandlers) GetOrganizationAnalytics(c *gin.Context) {
	orgName := c.Param("org")
	if orgName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Organization name is required"})
		return
	}

	// Get organization ID from name
	orgID, err := h.getOrganizationID(c.Request.Context(), orgName)
	if err != nil {
		h.logger.WithError(err).Error("Failed to resolve organization")
		c.JSON(http.StatusNotFound, gin.H{"error": "Organization not found"})
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

	insights, err := h.analyticsService.GetOrganizationInsights(c.Request.Context(), orgID, filters)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get organization analytics")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get organization analytics"})
		return
	}

	c.JSON(http.StatusOK, insights)
}

// GetOrganizationMembers handles GET /api/v1/organizations/:org/analytics/members
func (h *AnalyticsHandlers) GetOrganizationMembers(c *gin.Context) {
	orgName := c.Param("org")
	if orgName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Organization name is required"})
		return
	}

	// Get organization ID from name
	orgID, err := h.getOrganizationID(c.Request.Context(), orgName)
	if err != nil {
		h.logger.WithError(err).Error("Failed to resolve organization")
		c.JSON(http.StatusNotFound, gin.H{"error": "Organization not found"})
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

	insights, err := h.analyticsService.GetOrganizationInsights(c.Request.Context(), orgID, filters)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get organization member analytics")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get organization member analytics"})
		return
	}

	// Return just the member stats portion
	c.JSON(http.StatusOK, insights.MemberStats)
}

// GetOrganizationRepositories handles GET /api/v1/organizations/:org/analytics/repositories
func (h *AnalyticsHandlers) GetOrganizationRepositories(c *gin.Context) {
	orgName := c.Param("org")
	if orgName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Organization name is required"})
		return
	}

	// Get organization ID from name
	orgID, err := h.getOrganizationID(c.Request.Context(), orgName)
	if err != nil {
		h.logger.WithError(err).Error("Failed to resolve organization")
		c.JSON(http.StatusNotFound, gin.H{"error": "Organization not found"})
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

	insights, err := h.analyticsService.GetOrganizationInsights(c.Request.Context(), orgID, filters)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get organization repository analytics")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get organization repository analytics"})
		return
	}

	// Return just the repository stats portion
	c.JSON(http.StatusOK, insights.RepositoryStats)
}

// GetOrganizationTeams handles GET /api/v1/organizations/:org/analytics/teams
func (h *AnalyticsHandlers) GetOrganizationTeams(c *gin.Context) {
	orgName := c.Param("org")
	if orgName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Organization name is required"})
		return
	}

	// Get organization ID from name
	orgID, err := h.getOrganizationID(c.Request.Context(), orgName)
	if err != nil {
		h.logger.WithError(err).Error("Failed to resolve organization")
		c.JSON(http.StatusNotFound, gin.H{"error": "Organization not found"})
		return
	}

	// Get team analytics - for now return placeholder data since team insights aren't in OrganizationInsights
	var teamStats []gin.H
	var teams []models.Team
	if err := h.db.WithContext(c.Request.Context()).Where("organization_id = ?", orgID).Find(&teams).Error; err != nil {
		h.logger.WithError(err).Error("Failed to get teams")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get team analytics"})
		return
	}

	for _, team := range teams {
		// Get member count for each team
		var memberCount int64
		h.db.WithContext(c.Request.Context()).Model(&models.TeamMember{}).Where("team_id = ?", team.ID).Count(&memberCount)

		teamStats = append(teamStats, gin.H{
			"id":           team.ID,
			"name":         team.Name,
			"description":  team.Description,
			"member_count": memberCount,
			"created_at":   team.CreatedAt,
			"updated_at":   team.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"total_teams": len(teams),
		"teams":       teamStats,
	})
}

// GetOrganizationSecurity handles GET /api/v1/organizations/:org/analytics/security
func (h *AnalyticsHandlers) GetOrganizationSecurity(c *gin.Context) {
	orgName := c.Param("org")
	if orgName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Organization name is required"})
		return
	}

	// Get organization ID from name
	orgID, err := h.getOrganizationID(c.Request.Context(), orgName)
	if err != nil {
		h.logger.WithError(err).Error("Failed to resolve organization")
		c.JSON(http.StatusNotFound, gin.H{"error": "Organization not found"})
		return
	}

	// Parse query parameters for filtering
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

	// Get security events and metrics
	query := h.db.WithContext(c.Request.Context()).Model(&models.AnalyticsEvent{}).Where("organization_id = ?", orgID)

	if startDate != nil {
		query = query.Where("created_at >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("created_at <= ?", *endDate)
	}

	// Count security-related events
	var securityEvents int64
	query.Where("event_type LIKE ?", "security.%").Count(&securityEvents)

	// Count access denied events
	var accessDeniedEvents int64
	query.Where("event_type = ?", "security.access_denied").Count(&accessDeniedEvents)

	// Count MFA events
	var mfaEvents int64
	query.Where("event_type = ?", "security.mfa_enabled").Count(&mfaEvents)

	// Get recent security alerts (placeholder data)
	securityAlerts := []gin.H{
		{
			"type":        "vulnerability",
			"severity":    "medium",
			"title":       "Outdated dependency detected",
			"description": "A repository contains outdated dependencies with known vulnerabilities",
			"created_at":  time.Now().Add(-24 * time.Hour),
		},
	}

	// Calculate security score (simplified)
	securityScore := 85.0
	if accessDeniedEvents > 10 {
		securityScore -= 10.0
	}
	if mfaEvents > 0 {
		securityScore += 5.0
	}

	c.JSON(http.StatusOK, gin.H{
		"security_score":        securityScore,
		"total_security_events": securityEvents,
		"access_denied_events":  accessDeniedEvents,
		"mfa_enabled_events":    mfaEvents,
		"security_alerts":       securityAlerts,
		"compliance_status": gin.H{
			"two_factor_required": true,
			"sso_enabled":         false,
			"audit_logs_enabled":  true,
		},
	})
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
		h.logger.WithError(err).Error("Failed to get usage analytics")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get usage analytics"})
		return
	}

	// Calculate additional usage metrics
	var totalEvents int64
	h.db.WithContext(c.Request.Context()).Model(&models.AnalyticsEvent{}).Count(&totalEvents)

	var totalPerfLogs int64
	h.db.WithContext(c.Request.Context()).Model(&models.PerformanceLog{}).Count(&totalPerfLogs)

	usageData := gin.H{
		"system_insights":        insights,
		"total_events":           totalEvents,
		"total_performance_logs": totalPerfLogs,
		"period":                 period,
	}

	c.JSON(http.StatusOK, usageData)
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

	// Parse query parameters
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	var startDate, endDate *time.Time
	sinceDefault := time.Now().AddDate(0, -1, 0) // Default to last month
	if startDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			startDate = &parsed
		}
	} else {
		startDate = &sinceDefault
	}
	if endDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			endDate = &parsed
		}
	}

	// Get cost-related data from organization analytics
	query := h.db.WithContext(c.Request.Context()).Model(&models.OrganizationAnalytics{})
	if startDate != nil {
		query = query.Where("date >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("date <= ?", *endDate)
	}

	// Aggregate cost data
	var costSummary struct {
		TotalStorageMB   int64   `json:"total_storage_mb"`
		TotalBandwidthMB int64   `json:"total_bandwidth_mb"`
		TotalComputeMin  int64   `json:"total_compute_minutes"`
		EstimatedCost    float64 `json:"estimated_cost"`
	}

	err := query.Select(
		"COALESCE(SUM(storage_used_mb), 0) as total_storage_mb, " +
			"COALESCE(SUM(bandwidth_used_mb), 0) as total_bandwidth_mb, " +
			"COALESCE(SUM(compute_time_minutes), 0) as total_compute_minutes, " +
			"COALESCE(SUM(estimated_cost), 0) as estimated_cost",
	).Scan(&costSummary).Error

	if err != nil {
		h.logger.WithError(err).Error("Failed to get cost analytics")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get cost analytics"})
		return
	}

	// Calculate cost breakdown
	costBreakdown := []gin.H{
		{
			"category":   "Storage",
			"usage_mb":   costSummary.TotalStorageMB,
			"cost":       float64(costSummary.TotalStorageMB) * 0.0001, // $0.0001 per MB
			"percentage": 35.0,
		},
		{
			"category":   "Bandwidth",
			"usage_mb":   costSummary.TotalBandwidthMB,
			"cost":       float64(costSummary.TotalBandwidthMB) * 0.00005, // $0.00005 per MB
			"percentage": 25.0,
		},
		{
			"category":      "Compute",
			"usage_minutes": costSummary.TotalComputeMin,
			"cost":          float64(costSummary.TotalComputeMin) * 0.01, // $0.01 per minute
			"percentage":    40.0,
		},
	}

	costData := gin.H{
		"period":         fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), time.Now().Format("2006-01-02")),
		"total_cost":     costSummary.EstimatedCost,
		"cost_breakdown": costBreakdown,
		"usage_summary":  costSummary,
	}

	c.JSON(http.StatusOK, costData)
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
	userID, exists := c.Get("user_id")
	if !exists {
		return false
	}

	uid, err := parseUserID(userID)
	if err != nil {
		return false
	}

	// Check if user is admin
	var user models.User
	if err := h.db.WithContext(c.Request.Context()).Where("id = ?", uid).First(&user).Error; err != nil {
		return false
	}

	// Check admin status using the IsAdmin field
	return user.IsAdmin
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

// getOrganizationID resolves organization ID from organization name
func (h *AnalyticsHandlers) getOrganizationID(ctx context.Context, orgName string) (uuid.UUID, error) {
	var org struct {
		ID uuid.UUID `json:"id"`
	}
	err := h.db.WithContext(ctx).
		Model(&models.Organization{}).Select("id").Where("name = ?", orgName).First(&org).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return uuid.Nil, fmt.Errorf("organization not found")
		}
		return uuid.Nil, fmt.Errorf("failed to find organization: %w", err)
	}

	return org.ID, nil
}

// getUserContributions gets user contributions across all repositories
func (h *AnalyticsHandlers) getUserContributions(ctx context.Context, userID uuid.UUID) (gin.H, error) {
	// First get the user's email to match against commits
	var user models.User
	if err := h.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Get user's commits across all repositories
	var commitStats struct {
		TotalCommits   int64 `json:"total_commits"`
		TotalAdditions int64 `json:"total_additions"`
		TotalDeletions int64 `json:"total_deletions"`
	}

	err := h.db.WithContext(ctx).Model(&models.Commit{}).
		Select("COUNT(*) as total_commits, COALESCE(SUM(additions), 0) as total_additions, COALESCE(SUM(deletions), 0) as total_deletions").
		Where("author_email = ?", user.Email).
		Scan(&commitStats).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get commit stats: %w", err)
	}

	// Get user's pull requests count
	var prCount int64
	h.db.WithContext(ctx).Model(&models.PullRequest{}).Where("user_id = ?", userID).Count(&prCount)

	// Get repositories user has contributed to
	var repoCount int64
	h.db.WithContext(ctx).Model(&models.Commit{}).
		Where("author_email = ?", user.Email).
		Distinct("repository_id").Count(&repoCount)

	// Get contribution activity for the last 12 months
	since := time.Now().AddDate(-1, 0, 0)
	var monthlyContributions []struct {
		Month string `json:"month"`
		Count int64  `json:"count"`
	}

	err = h.db.WithContext(ctx).Model(&models.Commit{}).
		Select("DATE_TRUNC('month', created_at) as month, COUNT(*) as count").
		Where("author_email = ? AND created_at >= ?", user.Email, since).
		Group("DATE_TRUNC('month', created_at)").
		Order("month ASC").
		Scan(&monthlyContributions).Error

	if err != nil {
		h.logger.WithError(err).Warn("Failed to get monthly contributions")
		monthlyContributions = []struct {
			Month string `json:"month"`
			Count int64  `json:"count"`
		}{}
	}

	return gin.H{
		"total_commits":   commitStats.TotalCommits,
		"total_additions": commitStats.TotalAdditions,
		"total_deletions": commitStats.TotalDeletions,

		"total_pull_requests":      prCount,
		"repositories_contributed": repoCount,
		"monthly_contributions":    monthlyContributions,
		"last_updated":             time.Now().Format(time.RFC3339),
	}, nil
}
