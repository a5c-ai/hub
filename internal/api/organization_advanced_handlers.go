package api

import (
	"net/http"
	"strconv"

	"github.com/a5c-ai/hub/internal/models"
	"github.com/a5c-ai/hub/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Organization Advanced Handlers
type OrganizationAdvancedHandlers struct {
	customRoleService   services.CustomRoleService
	policyService       services.OrganizationPolicyService
	templateService     services.OrganizationTemplateService
	settingsService     services.OrganizationSettingsService
	analyticsService    services.OrganizationAnalyticsService
	auditService        services.OrganizationAuditService
	teamTemplateService services.TeamTemplateService
}

func NewOrganizationAdvancedHandlers(
	customRoleService services.CustomRoleService,
	policyService services.OrganizationPolicyService,
	templateService services.OrganizationTemplateService,
	settingsService services.OrganizationSettingsService,
	analyticsService services.OrganizationAnalyticsService,
	auditService services.OrganizationAuditService,
	teamTemplateService services.TeamTemplateService,
) *OrganizationAdvancedHandlers {
	return &OrganizationAdvancedHandlers{
		customRoleService:   customRoleService,
		policyService:       policyService,
		templateService:     templateService,
		settingsService:     settingsService,
		analyticsService:    analyticsService,
		auditService:        auditService,
		teamTemplateService: teamTemplateService,
	}
}

// Custom Roles Handlers
func (h *OrganizationAdvancedHandlers) CreateCustomRole(c *gin.Context) {
	orgName := c.Param("org")

	var req services.CreateCustomRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	role, err := h.customRoleService.CreateCustomRole(c.Request.Context(), orgName, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, role)
}

func (h *OrganizationAdvancedHandlers) GetCustomRole(c *gin.Context) {
	orgName := c.Param("org")
	roleIDStr := c.Param("role_id")

	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role ID"})
		return
	}

	role, err := h.customRoleService.GetCustomRole(c.Request.Context(), orgName, roleID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, role)
}

func (h *OrganizationAdvancedHandlers) UpdateCustomRole(c *gin.Context) {
	orgName := c.Param("org")
	roleIDStr := c.Param("role_id")

	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role ID"})
		return
	}

	var req services.UpdateCustomRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	role, err := h.customRoleService.UpdateCustomRole(c.Request.Context(), orgName, roleID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, role)
}

func (h *OrganizationAdvancedHandlers) DeleteCustomRole(c *gin.Context) {
	orgName := c.Param("org")
	roleIDStr := c.Param("role_id")

	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role ID"})
		return
	}

	if err := h.customRoleService.DeleteCustomRole(c.Request.Context(), orgName, roleID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *OrganizationAdvancedHandlers) ListCustomRoles(c *gin.Context) {
	orgName := c.Param("org")

	roles, err := h.customRoleService.ListCustomRoles(c.Request.Context(), orgName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"roles": roles})
}

// Policy Handlers
func (h *OrganizationAdvancedHandlers) CreatePolicy(c *gin.Context) {
	orgName := c.Param("org")

	var req services.CreatePolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	policy, err := h.policyService.CreatePolicy(c.Request.Context(), orgName, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, policy)
}

func (h *OrganizationAdvancedHandlers) GetPolicy(c *gin.Context) {
	orgName := c.Param("org")
	policyIDStr := c.Param("policy_id")

	policyID, err := uuid.Parse(policyIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid policy ID"})
		return
	}

	policy, err := h.policyService.GetPolicy(c.Request.Context(), orgName, policyID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, policy)
}

func (h *OrganizationAdvancedHandlers) UpdatePolicy(c *gin.Context) {
	orgName := c.Param("org")
	policyIDStr := c.Param("policy_id")

	policyID, err := uuid.Parse(policyIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid policy ID"})
		return
	}

	var req services.UpdatePolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	policy, err := h.policyService.UpdatePolicy(c.Request.Context(), orgName, policyID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, policy)
}

func (h *OrganizationAdvancedHandlers) DeletePolicy(c *gin.Context) {
	orgName := c.Param("org")
	policyIDStr := c.Param("policy_id")

	policyID, err := uuid.Parse(policyIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid policy ID"})
		return
	}

	if err := h.policyService.DeletePolicy(c.Request.Context(), orgName, policyID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *OrganizationAdvancedHandlers) ListPolicies(c *gin.Context) {
	orgName := c.Param("org")
	policyTypeStr := c.Query("type")

	var policyType *models.PolicyType
	if policyTypeStr != "" {
		pt := models.PolicyType(policyTypeStr)
		policyType = &pt
	}

	policies, err := h.policyService.ListPolicies(c.Request.Context(), orgName, policyType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"policies": policies})
}

// Analytics Handlers
func (h *OrganizationAdvancedHandlers) GetDashboardMetrics(c *gin.Context) {
	orgName := c.Param("org")

	metrics, err := h.analyticsService.GetDashboardMetrics(c.Request.Context(), orgName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

func (h *OrganizationAdvancedHandlers) GetMemberActivityMetrics(c *gin.Context) {
	orgName := c.Param("org")
	period := c.DefaultQuery("period", "30d")

	metrics, err := h.analyticsService.GetMemberActivityMetrics(c.Request.Context(), orgName, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

func (h *OrganizationAdvancedHandlers) GetRepositoryUsageMetrics(c *gin.Context) {
	orgName := c.Param("org")
	period := c.DefaultQuery("period", "30d")

	metrics, err := h.analyticsService.GetRepositoryUsageMetrics(c.Request.Context(), orgName, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

func (h *OrganizationAdvancedHandlers) GetTeamPerformanceMetrics(c *gin.Context) {
	orgName := c.Param("org")
	period := c.DefaultQuery("period", "30d")

	metrics, err := h.analyticsService.GetTeamPerformanceMetrics(c.Request.Context(), orgName, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

func (h *OrganizationAdvancedHandlers) GetSecurityMetrics(c *gin.Context) {
	orgName := c.Param("org")
	period := c.DefaultQuery("period", "30d")

	metrics, err := h.analyticsService.GetSecurityMetrics(c.Request.Context(), orgName, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

func (h *OrganizationAdvancedHandlers) GetUsageAndCostMetrics(c *gin.Context) {
	orgName := c.Param("org")
	period := c.DefaultQuery("period", "30d")

	metrics, err := h.analyticsService.GetUsageAndCostMetrics(c.Request.Context(), orgName, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// Audit Handlers
func (h *OrganizationAdvancedHandlers) GetActivitiesWithFilters(c *gin.Context) {
	orgName := c.Param("org")

	var filters services.ActivityFilters

	// Parse query parameters into filters
	if actions := c.QueryArray("actions"); len(actions) > 0 {
		// Convert string array to ActivityAction array
		// This is a simplified implementation
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filters.Limit = limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			filters.Offset = offset
		}
	}

	filters.SortBy = c.DefaultQuery("sort_by", "created_at")
	filters.SortOrder = c.DefaultQuery("sort_order", "desc")

	response, err := h.auditService.GetActivitiesWithFilters(c.Request.Context(), orgName, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *OrganizationAdvancedHandlers) SearchActivities(c *gin.Context) {
	orgName := c.Param("org")
	query := c.Query("q")

	var filters services.ActivityFilters
	// Parse filters from query parameters (similar to above)

	response, err := h.auditService.SearchActivities(c.Request.Context(), orgName, query, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *OrganizationAdvancedHandlers) ExportActivities(c *gin.Context) {
	orgName := c.Param("org")
	format := c.DefaultQuery("format", "csv")

	var filters services.ActivityFilters
	// Parse filters from query parameters

	data, err := h.auditService.ExportActivities(c.Request.Context(), orgName, format, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Set appropriate headers for file download
	switch format {
	case "csv":
		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", "attachment; filename=activities.csv")
	case "json":
		c.Header("Content-Type", "application/json")
		c.Header("Content-Disposition", "attachment; filename=activities.json")
	}

	c.Data(http.StatusOK, c.GetHeader("Content-Type"), data)
}

func (h *OrganizationAdvancedHandlers) GetAuditSummary(c *gin.Context) {
	orgName := c.Param("org")
	period := c.DefaultQuery("period", "30d")

	summary, err := h.auditService.GetAuditSummary(c.Request.Context(), orgName, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}

func (h *OrganizationAdvancedHandlers) GetRetentionPolicyStatus(c *gin.Context) {
	orgName := c.Param("org")

	status, err := h.auditService.GetRetentionPolicyStatus(c.Request.Context(), orgName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

// Team Template Handlers
func (h *OrganizationAdvancedHandlers) CreateTeamFromTemplate(c *gin.Context) {
	orgName := c.Param("org")
	templateIDStr := c.Param("template_id")

	templateID, err := uuid.Parse(templateIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template ID"})
		return
	}

	var req services.CreateTeamFromTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	team, err := h.teamTemplateService.CreateTeamFromTemplate(c.Request.Context(), orgName, templateID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, team)
}

func (h *OrganizationAdvancedHandlers) GetTeamTemplates(c *gin.Context) {
	orgName := c.Param("org")

	templates, err := h.teamTemplateService.GetTeamTemplates(c.Request.Context(), orgName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"templates": templates})
}

func (h *OrganizationAdvancedHandlers) GetTeamPerformanceMetricsForTeam(c *gin.Context) {
	orgName := c.Param("org")
	teamName := c.Param("team")
	period := c.DefaultQuery("period", "30d")

	metrics, err := h.teamTemplateService.GetTeamPerformanceMetrics(c.Request.Context(), orgName, teamName, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metrics)
}
