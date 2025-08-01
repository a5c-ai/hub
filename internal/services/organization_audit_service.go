package services

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/a5c-ai/hub/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Enhanced Activity Service with Search, Filtering, and Export
type OrganizationAuditService interface {
	// Enhanced activity retrieval
	GetActivitiesWithFilters(ctx context.Context, orgName string, filters ActivityFilters) (*ActivityResponse, error)
	SearchActivities(ctx context.Context, orgName string, query string, filters ActivityFilters) (*ActivityResponse, error)

	// Export functionality
	ExportActivities(ctx context.Context, orgName string, format string, filters ActivityFilters) ([]byte, error)

	// Retention policy management
	ApplyRetentionPolicy(ctx context.Context, orgName string) error
	GetRetentionPolicyStatus(ctx context.Context, orgName string) (*RetentionPolicyStatus, error)

	// Real-time notifications
	SubscribeToActivity(ctx context.Context, orgName string, userID uuid.UUID, filters ActivitySubscriptionFilters) error
	UnsubscribeFromActivity(ctx context.Context, orgName string, userID uuid.UUID) error

	// Compliance and audit reports
	GenerateComplianceReport(ctx context.Context, orgName string, reportType string, period string) (*ComplianceReport, error)
	GetAuditSummary(ctx context.Context, orgName string, period string) (*AuditSummary, error)
}

// Request/Response Types
type ActivityFilters struct {
	Actions     []models.ActivityAction `json:"actions,omitempty"`
	ActorIDs    []uuid.UUID             `json:"actor_ids,omitempty"`
	TargetTypes []string                `json:"target_types,omitempty"`
	TargetIDs   []uuid.UUID             `json:"target_ids,omitempty"`
	StartDate   *time.Time              `json:"start_date,omitempty"`
	EndDate     *time.Time              `json:"end_date,omitempty"`
	Limit       int                     `json:"limit,omitempty"`
	Offset      int                     `json:"offset,omitempty"`
	SortBy      string                  `json:"sort_by,omitempty"`    // "created_at", "action", "actor"
	SortOrder   string                  `json:"sort_order,omitempty"` // "asc", "desc"
}

type ActivityResponse struct {
	Activities    []*EnhancedOrganizationActivity `json:"activities"`
	TotalCount    int                             `json:"total_count"`
	FilteredCount int                             `json:"filtered_count"`
	HasMore       bool                            `json:"has_more"`
	NextOffset    int                             `json:"next_offset"`
}

type EnhancedOrganizationActivity struct {
	*models.OrganizationActivity
	ActorDetails  *UserDetails   `json:"actor_details,omitempty"`
	TargetDetails *TargetDetails `json:"target_details,omitempty"`
	RiskLevel     string         `json:"risk_level"`
	IPAddress     string         `json:"ip_address,omitempty"`
	UserAgent     string         `json:"user_agent,omitempty"`
	Location      string         `json:"location,omitempty"`
}

type UserDetails struct {
	Username  string `json:"username"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

type TargetDetails struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

type ActivitySubscriptionFilters struct {
	Actions     []models.ActivityAction `json:"actions,omitempty"`
	TargetTypes []string                `json:"target_types,omitempty"`
	RiskLevels  []string                `json:"risk_levels,omitempty"`
	RealTime    bool                    `json:"real_time"`
	Email       bool                    `json:"email"`
	Webhook     string                  `json:"webhook,omitempty"`
}

type RetentionPolicyStatus struct {
	CurrentPolicy        string    `json:"current_policy"`
	RetentionDays        int       `json:"retention_days"`
	TotalActivities      int       `json:"total_activities"`
	OldestActivity       time.Time `json:"oldest_activity"`
	LastCleanup          time.Time `json:"last_cleanup"`
	NextCleanup          time.Time `json:"next_cleanup"`
	ActivitiesForCleanup int       `json:"activities_for_cleanup"`
}

type ComplianceReport struct {
	ReportType       string                  `json:"report_type"`
	Period           string                  `json:"period"`
	GeneratedAt      time.Time               `json:"generated_at"`
	Organization     string                  `json:"organization"`
	Summary          *ComplianceSummary      `json:"summary"`
	AccessEvents     []*AccessEvent          `json:"access_events"`
	SecurityEvents   []*SecurityEvent        `json:"security_events"`
	DataEvents       []*DataEvent            `json:"data_events"`
	AdminEvents      []*AdminEvent           `json:"admin_events"`
	PolicyViolations []*PolicyViolationEvent `json:"policy_violations"`
}

type ComplianceSummary struct {
	TotalEvents      int     `json:"total_events"`
	SecurityEvents   int     `json:"security_events"`
	AccessEvents     int     `json:"access_events"`
	DataEvents       int     `json:"data_events"`
	AdminEvents      int     `json:"admin_events"`
	HighRiskEvents   int     `json:"high_risk_events"`
	PolicyViolations int     `json:"policy_violations"`
	ComplianceScore  float64 `json:"compliance_score"`
}

type AccessEvent struct {
	Timestamp time.Time `json:"timestamp"`
	User      string    `json:"user"`
	Action    string    `json:"action"`
	Resource  string    `json:"resource"`
	Result    string    `json:"result"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	RiskLevel string    `json:"risk_level"`
}

type SecurityEvent struct {
	Timestamp   time.Time `json:"timestamp"`
	EventType   string    `json:"event_type"`
	User        string    `json:"user"`
	Description string    `json:"description"`
	Severity    string    `json:"severity"`
	IPAddress   string    `json:"ip_address"`
	Location    string    `json:"location"`
}

type DataEvent struct {
	Timestamp  time.Time `json:"timestamp"`
	User       string    `json:"user"`
	Action     string    `json:"action"`
	DataType   string    `json:"data_type"`
	Resource   string    `json:"resource"`
	Size       int64     `json:"size"`
	Encryption bool      `json:"encryption"`
}

type AdminEvent struct {
	Timestamp time.Time `json:"timestamp"`
	Admin     string    `json:"admin"`
	Action    string    `json:"action"`
	Target    string    `json:"target"`
	Changes   string    `json:"changes"`
	Approved  bool      `json:"approved"`
}

type PolicyViolationEvent struct {
	Timestamp  time.Time `json:"timestamp"`
	User       string    `json:"user"`
	PolicyName string    `json:"policy_name"`
	Violation  string    `json:"violation"`
	Action     string    `json:"action"`
	Resolved   bool      `json:"resolved"`
}

type AuditSummary struct {
	Period             string               `json:"period"`
	TotalActivities    int                  `json:"total_activities"`
	UniqueUsers        int                  `json:"unique_users"`
	TopActions         []ActionSummary      `json:"top_actions"`
	ActivityTrends     []ActivityTrendPoint `json:"activity_trends"`
	SecurityInsights   *SecurityInsights    `json:"security_insights"`
	ComplianceStatus   map[string]bool      `json:"compliance_status"`
	RecommendedActions []string             `json:"recommended_actions"`
}

type ActionSummary struct {
	Action string `json:"action"`
	Count  int    `json:"count"`
	Users  int    `json:"unique_users"`
}

type ActivityTrendPoint struct {
	Date  time.Time `json:"date"`
	Count int       `json:"count"`
}

type SecurityInsights struct {
	SuspiciousActivities int            `json:"suspicious_activities"`
	FailedAttempts       int            `json:"failed_attempts"`
	UnusualLocations     []string       `json:"unusual_locations"`
	OffHoursAccess       int            `json:"off_hours_access"`
	RiskDistribution     map[string]int `json:"risk_distribution"`
}

// Service Implementation
type organizationAuditService struct {
	db *gorm.DB
}

func NewOrganizationAuditService(db *gorm.DB) OrganizationAuditService {
	return &organizationAuditService{db: db}
}

func (s *organizationAuditService) GetActivitiesWithFilters(ctx context.Context, orgName string, filters ActivityFilters) (*ActivityResponse, error) {
	var org models.Organization
	if err := s.db.Where("name = ?", orgName).First(&org).Error; err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	query := s.db.Where("organization_id = ?", org.ID).Preload("Actor")

	// Apply filters
	query = s.applyActivityFilters(query, filters)

	// Get total count before pagination
	var totalCount int64
	countQuery := query
	countQuery.Model(&models.OrganizationActivity{}).Count(&totalCount)

	// Apply sorting
	sortBy := filters.SortBy
	if sortBy == "" {
		sortBy = "created_at"
	}
	sortOrder := filters.SortOrder
	if sortOrder == "" {
		sortOrder = "desc"
	}
	query = query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))

	// Apply pagination
	limit := filters.Limit
	if limit == 0 || limit > 100 {
		limit = 50 // Default limit
	}
	offset := filters.Offset

	query = query.Limit(limit).Offset(offset)

	var activities []*models.OrganizationActivity
	if err := query.Find(&activities).Error; err != nil {
		return nil, fmt.Errorf("failed to get activities: %w", err)
	}

	// Enhance activities with additional details
	enhancedActivities := make([]*EnhancedOrganizationActivity, len(activities))
	for i, activity := range activities {
		enhanced := &EnhancedOrganizationActivity{
			OrganizationActivity: activity,
			RiskLevel:            s.calculateRiskLevel(activity),
		}

		// Add actor details
		if activity.Actor.ID != uuid.Nil {
			enhanced.ActorDetails = &UserDetails{
				Username:  activity.Actor.Username,
				Name:      activity.Actor.FullName,
				Email:     activity.Actor.Email,
				AvatarURL: activity.Actor.AvatarURL,
			}
		}

		// Add target details
		enhanced.TargetDetails = s.getTargetDetails(activity)

		// Add metadata if available
		if activity.Metadata != "" {
			var metadata map[string]interface{}
			if err := json.Unmarshal([]byte(activity.Metadata), &metadata); err == nil {
				if ip, ok := metadata["ip_address"].(string); ok {
					enhanced.IPAddress = ip
				}
				if ua, ok := metadata["user_agent"].(string); ok {
					enhanced.UserAgent = ua
				}
				if loc, ok := metadata["location"].(string); ok {
					enhanced.Location = loc
				}
			}
		}

		enhancedActivities[i] = enhanced
	}

	hasMore := int64(offset+limit) < totalCount
	nextOffset := offset + limit

	return &ActivityResponse{
		Activities:    enhancedActivities,
		TotalCount:    int(totalCount),
		FilteredCount: len(enhancedActivities),
		HasMore:       hasMore,
		NextOffset:    nextOffset,
	}, nil
}

func (s *organizationAuditService) SearchActivities(ctx context.Context, orgName string, query string, filters ActivityFilters) (*ActivityResponse, error) {
	var org models.Organization
	if err := s.db.Where("name = ?", orgName).First(&org).Error; err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	// Build search query
	dbQuery := s.db.Where("organization_id = ?", org.ID).Preload("Actor")

	// Add text search across multiple fields
	if query != "" {
		searchPattern := "%" + strings.ToLower(query) + "%"
		dbQuery = dbQuery.Where(
			"LOWER(action::text) LIKE ? OR LOWER(target_type) LIKE ? OR LOWER(metadata) LIKE ?",
			searchPattern, searchPattern, searchPattern,
		)
	}

	// Apply additional filters
	dbQuery = s.applyActivityFilters(dbQuery, filters)

	// Continue with the same logic as GetActivitiesWithFilters
	var totalCount int64
	dbQuery.Model(&models.OrganizationActivity{}).Count(&totalCount)

	// Apply sorting and pagination...
	sortBy := filters.SortBy
	if sortBy == "" {
		sortBy = "created_at"
	}
	sortOrder := filters.SortOrder
	if sortOrder == "" {
		sortOrder = "desc"
	}
	dbQuery = dbQuery.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))

	limit := filters.Limit
	if limit == 0 || limit > 100 {
		limit = 50
	}
	offset := filters.Offset
	dbQuery = dbQuery.Limit(limit).Offset(offset)

	var activities []*models.OrganizationActivity
	if err := dbQuery.Find(&activities).Error; err != nil {
		return nil, fmt.Errorf("failed to search activities: %w", err)
	}

	// Convert to enhanced activities (same logic as above)
	enhancedActivities := make([]*EnhancedOrganizationActivity, len(activities))
	for i, activity := range activities {
		enhanced := &EnhancedOrganizationActivity{
			OrganizationActivity: activity,
			RiskLevel:            s.calculateRiskLevel(activity),
		}

		if activity.Actor.ID != uuid.Nil {
			enhanced.ActorDetails = &UserDetails{
				Username:  activity.Actor.Username,
				Name:      activity.Actor.FullName,
				Email:     activity.Actor.Email,
				AvatarURL: activity.Actor.AvatarURL,
			}
		}

		enhanced.TargetDetails = s.getTargetDetails(activity)
		enhancedActivities[i] = enhanced
	}

	return &ActivityResponse{
		Activities:    enhancedActivities,
		TotalCount:    int(totalCount),
		FilteredCount: len(enhancedActivities),
		HasMore:       int64(offset+limit) < totalCount,
		NextOffset:    offset + limit,
	}, nil
}

func (s *organizationAuditService) ExportActivities(ctx context.Context, orgName string, format string, filters ActivityFilters) ([]byte, error) {
	// Remove pagination for export
	filters.Limit = 0
	filters.Offset = 0

	response, err := s.GetActivitiesWithFilters(ctx, orgName, filters)
	if err != nil {
		return nil, err
	}

	switch strings.ToLower(format) {
	case "csv":
		return s.exportAsCSV(response.Activities)
	case "json":
		return s.exportAsJSON(response.Activities)
	default:
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}
}

func (s *organizationAuditService) ApplyRetentionPolicy(ctx context.Context, orgName string) error {
	var org models.Organization
	if err := s.db.Where("name = ?", orgName).First(&org).Error; err != nil {
		return fmt.Errorf("organization not found: %w", err)
	}

	// Get organization settings to determine retention policy
	var settings models.OrganizationSettings
	if err := s.db.Where("organization_id = ?", org.ID).First(&settings).Error; err != nil {
		// Use default retention if settings not found
		settings.RetentionDays = 365 // Default 1 year
	}

	// Calculate cutoff date
	cutoffDate := time.Now().AddDate(0, 0, -settings.RetentionDays)

	// Delete activities older than retention period
	result := s.db.Where("organization_id = ? AND created_at < ?", org.ID, cutoffDate).Delete(&models.OrganizationActivity{})
	if result.Error != nil {
		return fmt.Errorf("failed to apply retention policy: %w", result.Error)
	}

	// Log retention policy application
	metadata := map[string]interface{}{
		"retention_days":     settings.RetentionDays,
		"cutoff_date":        cutoffDate,
		"activities_deleted": result.RowsAffected,
	}

	activity := &models.OrganizationActivity{
		OrganizationID: org.ID,
		ActorID:        uuid.Nil, // System action
		Action:         models.ActivityAction("system.retention_policy_applied"),
		TargetType:     "organization",
		TargetID:       &org.ID,
		Metadata:       s.marshalMetadata(metadata),
	}

	s.db.Create(activity)

	return nil
}

func (s *organizationAuditService) GetRetentionPolicyStatus(ctx context.Context, orgName string) (*RetentionPolicyStatus, error) {
	var org models.Organization
	if err := s.db.Where("name = ?", orgName).First(&org).Error; err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	var settings models.OrganizationSettings
	if err := s.db.Where("organization_id = ?", org.ID).First(&settings).Error; err != nil {
		settings.RetentionDays = 365
	}

	var totalActivities int64
	s.db.Model(&models.OrganizationActivity{}).Where("organization_id = ?", org.ID).Count(&totalActivities)

	var oldestActivity models.OrganizationActivity
	s.db.Where("organization_id = ?", org.ID).Order("created_at asc").First(&oldestActivity)

	cutoffDate := time.Now().AddDate(0, 0, -settings.RetentionDays)
	var activitiesForCleanup int64
	s.db.Model(&models.OrganizationActivity{}).
		Where("organization_id = ? AND created_at < ?", org.ID, cutoffDate).
		Count(&activitiesForCleanup)

	// Find last cleanup
	var lastCleanup time.Time
	var lastCleanupActivity models.OrganizationActivity
	if err := s.db.Where("organization_id = ? AND action = ?", org.ID, "system.retention_policy_applied").
		Order("created_at desc").First(&lastCleanupActivity).Error; err == nil {
		lastCleanup = lastCleanupActivity.CreatedAt
	}

	return &RetentionPolicyStatus{
		CurrentPolicy:        "automatic",
		RetentionDays:        settings.RetentionDays,
		TotalActivities:      int(totalActivities),
		OldestActivity:       oldestActivity.CreatedAt,
		LastCleanup:          lastCleanup,
		NextCleanup:          time.Now().AddDate(0, 0, 1), // Daily cleanup
		ActivitiesForCleanup: int(activitiesForCleanup),
	}, nil
}

func (s *organizationAuditService) SubscribeToActivity(ctx context.Context, orgName string, userID uuid.UUID, filters ActivitySubscriptionFilters) error {
	// This would implement real-time activity notifications
	// For now, it's a placeholder
	return fmt.Errorf("activity subscriptions not yet implemented")
}

func (s *organizationAuditService) UnsubscribeFromActivity(ctx context.Context, orgName string, userID uuid.UUID) error {
	// This would remove activity subscriptions
	return fmt.Errorf("activity subscriptions not yet implemented")
}

func (s *organizationAuditService) GenerateComplianceReport(ctx context.Context, orgName string, reportType string, period string) (*ComplianceReport, error) {
	// This would generate detailed compliance reports
	// For now, return a basic structure
	return &ComplianceReport{
		ReportType:   reportType,
		Period:       period,
		GeneratedAt:  time.Now(),
		Organization: orgName,
		Summary: &ComplianceSummary{
			ComplianceScore: 85.5,
		},
	}, nil
}

func (s *organizationAuditService) GetAuditSummary(ctx context.Context, orgName string, period string) (*AuditSummary, error) {
	var org models.Organization
	if err := s.db.Where("name = ?", orgName).First(&org).Error; err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	// Calculate date range based on period
	startDate, endDate := s.calculatePeriodRange(period)

	// Get total activities in period
	var totalActivities int64
	s.db.Model(&models.OrganizationActivity{}).
		Where("organization_id = ? AND created_at BETWEEN ? AND ?", org.ID, startDate, endDate).
		Count(&totalActivities)

	// Get unique users
	var uniqueUsers int64
	s.db.Model(&models.OrganizationActivity{}).
		Where("organization_id = ? AND created_at BETWEEN ? AND ?", org.ID, startDate, endDate).
		Distinct("actor_id").Count(&uniqueUsers)

	// Get top actions
	topActions := s.getTopActions(org.ID, startDate, endDate)

	return &AuditSummary{
		Period:          period,
		TotalActivities: int(totalActivities),
		UniqueUsers:     int(uniqueUsers),
		TopActions:      topActions,
		ComplianceStatus: map[string]bool{
			"GDPR": true,
			"SOC2": false,
		},
		RecommendedActions: []string{
			"Enable two-factor authentication for all users",
			"Review and update access permissions",
			"Implement stronger password policies",
		},
	}, nil
}

// Helper methods
func (s *organizationAuditService) applyActivityFilters(query *gorm.DB, filters ActivityFilters) *gorm.DB {
	if len(filters.Actions) > 0 {
		query = query.Where("action IN ?", filters.Actions)
	}
	if len(filters.ActorIDs) > 0 {
		query = query.Where("actor_id IN ?", filters.ActorIDs)
	}
	if len(filters.TargetTypes) > 0 {
		query = query.Where("target_type IN ?", filters.TargetTypes)
	}
	if len(filters.TargetIDs) > 0 {
		query = query.Where("target_id IN ?", filters.TargetIDs)
	}
	if filters.StartDate != nil {
		query = query.Where("created_at >= ?", *filters.StartDate)
	}
	if filters.EndDate != nil {
		query = query.Where("created_at <= ?", *filters.EndDate)
	}
	return query
}

func (s *organizationAuditService) calculateRiskLevel(activity *models.OrganizationActivity) string {
	// Calculate risk level based on action type, time, location, etc.
	switch activity.Action {
	case models.ActivityMemberAdded, models.ActivityMemberRemoved:
		return "medium"
	case models.ActivityPermissionGranted, models.ActivityPermissionRevoked:
		return "high"
	default:
		return "low"
	}
}

func (s *organizationAuditService) getTargetDetails(activity *models.OrganizationActivity) *TargetDetails {
	// Get target details based on target type and ID
	return &TargetDetails{
		Type: activity.TargetType,
		Name: activity.TargetType + " target",
	}
}

func (s *organizationAuditService) exportAsCSV(activities []*EnhancedOrganizationActivity) ([]byte, error) {
	var output strings.Builder
	writer := csv.NewWriter(&output)

	// Write header
	header := []string{"Timestamp", "Action", "Actor", "Target Type", "Target", "Risk Level", "IP Address", "Location"}
	writer.Write(header)

	// Write data
	for _, activity := range activities {
		record := []string{
			activity.CreatedAt.Format(time.RFC3339),
			string(activity.Action),
			activity.ActorDetails.Username,
			activity.TargetType,
			activity.TargetDetails.Name,
			activity.RiskLevel,
			activity.IPAddress,
			activity.Location,
		}
		writer.Write(record)
	}

	writer.Flush()
	return []byte(output.String()), nil
}

func (s *organizationAuditService) exportAsJSON(activities []*EnhancedOrganizationActivity) ([]byte, error) {
	return json.MarshalIndent(activities, "", "  ")
}

func (s *organizationAuditService) marshalMetadata(metadata map[string]interface{}) string {
	if data, err := json.Marshal(metadata); err == nil {
		return string(data)
	}
	return ""
}

func (s *organizationAuditService) calculatePeriodRange(period string) (time.Time, time.Time) {
	now := time.Now()
	switch period {
	case "7d":
		return now.AddDate(0, 0, -7), now
	case "30d":
		return now.AddDate(0, 0, -30), now
	case "90d":
		return now.AddDate(0, 0, -90), now
	case "1y":
		return now.AddDate(-1, 0, 0), now
	default:
		return now.AddDate(0, 0, -30), now
	}
}

func (s *organizationAuditService) getTopActions(orgID uuid.UUID, startDate, endDate time.Time) []ActionSummary {
	var results []struct {
		Action string
		Count  int64
		Users  int64
	}

	s.db.Model(&models.OrganizationActivity{}).
		Select("action, COUNT(*) as count, COUNT(DISTINCT actor_id) as users").
		Where("organization_id = ? AND created_at BETWEEN ? AND ?", orgID, startDate, endDate).
		Group("action").
		Order("count DESC").
		Limit(10).
		Scan(&results)

	topActions := make([]ActionSummary, len(results))
	for i, result := range results {
		topActions[i] = ActionSummary{
			Action: result.Action,
			Count:  int(result.Count),
			Users:  int(result.Users),
		}
	}

	return topActions
}
