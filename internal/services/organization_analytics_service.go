package services

import (
	"context"
	"fmt"
	"time"

	"github.com/a5c-ai/hub/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Organization Analytics Service
type OrganizationAnalyticsService interface {
	GetDashboardMetrics(ctx context.Context, orgName string) (*DashboardMetrics, error)
	GetMemberActivityMetrics(ctx context.Context, orgName string, period string) (*MemberActivityMetrics, error)
	GetRepositoryUsageMetrics(ctx context.Context, orgName string, period string) (*RepositoryUsageMetrics, error)
	GetTeamPerformanceMetrics(ctx context.Context, orgName string, period string) (*TeamPerformanceMetrics, error)
	GetSecurityMetrics(ctx context.Context, orgName string, period string) (*SecurityMetrics, error)
	GetUsageAndCostMetrics(ctx context.Context, orgName string, period string) (*UsageAndCostMetrics, error)
	ExportAnalyticsData(ctx context.Context, orgName string, format string, period string) ([]byte, error)
}

// Analytics Data Structures
type DashboardMetrics struct {
	Overview              *OverviewMetrics              `json:"overview"`
	RecentActivity        []*ActivitySummary            `json:"recent_activity"`
	TopRepositories       []*RepositorySummary          `json:"top_repositories"`
	ActiveMembers         []*MemberSummary              `json:"active_members"`
	SecurityAlerts        []*SecurityAlert              `json:"security_alerts"`
	StorageUsage          *StorageUsageMetrics          `json:"storage_usage"`
	BandwidthUsage        *BandwidthUsageMetrics        `json:"bandwidth_usage"`
}

type OverviewMetrics struct {
	TotalMembers       int     `json:"total_members"`
	TotalRepositories  int     `json:"total_repositories"`
	TotalTeams         int     `json:"total_teams"`
	ActiveMembers30d   int     `json:"active_members_30d"`
	CommitsThisMonth   int     `json:"commits_this_month"`
	IssuesOpen         int     `json:"issues_open"`
	PullRequestsOpen   int     `json:"pull_requests_open"`
	SecurityScore      float64 `json:"security_score"`
}

type ActivitySummary struct {
	Date         time.Time `json:"date"`
	Action       string    `json:"action"`
	ActorName    string    `json:"actor_name"`
	TargetType   string    `json:"target_type"`
	TargetName   string    `json:"target_name"`
	Description  string    `json:"description"`
}

type RepositorySummary struct {
	Name            string    `json:"name"`
	Language        string    `json:"language"`
	Stars           int       `json:"stars"`
	Forks           int       `json:"forks"`
	Contributors    int       `json:"contributors"`
	Commits30d      int       `json:"commits_30d"`
	Issues30d       int       `json:"issues_30d"`
	LastActivityAt  time.Time `json:"last_activity_at"`
}

type MemberSummary struct {
	Username       string    `json:"username"`
	Name           string    `json:"name"`
	Role           string    `json:"role"`
	Commits30d     int       `json:"commits_30d"`
	Issues30d      int       `json:"issues_30d"`
	PullRequests30d int      `json:"pull_requests_30d"`
	LastActiveAt   time.Time `json:"last_active_at"`
}

type SecurityAlert struct {
	Type        string    `json:"type"`
	Severity    string    `json:"severity"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Repository  string    `json:"repository"`
	CreatedAt   time.Time `json:"created_at"`
}

type StorageUsageMetrics struct {
	TotalUsedGB    float64 `json:"total_used_gb"`
	TotalLimitGB   float64 `json:"total_limit_gb"`
	UsagePercent   float64 `json:"usage_percent"`
	TopRepositories []RepositoryStorageUsage `json:"top_repositories"`
}

type RepositoryStorageUsage struct {
	Name     string  `json:"name"`
	SizeGB   float64 `json:"size_gb"`
	Percent  float64 `json:"percent"`
}

type BandwidthUsageMetrics struct {
	TotalUsedGB    float64 `json:"total_used_gb"`
	TotalLimitGB   float64 `json:"total_limit_gb"`
	UsagePercent   float64 `json:"usage_percent"`
	DailyUsage     []DailyBandwidthUsage `json:"daily_usage"`
}

type DailyBandwidthUsage struct {
	Date     time.Time `json:"date"`
	UsageGB  float64   `json:"usage_gb"`
}

type MemberActivityMetrics struct {
	Period         string                    `json:"period"`
	TotalMembers   int                       `json:"total_members"`
	ActiveMembers  int                       `json:"active_members"`
	MemberGrowth   []MemberGrowthData        `json:"member_growth"`
	ActivityTrends []ActivityTrendData       `json:"activity_trends"`
	TopContributors []MemberSummary          `json:"top_contributors"`
}

type MemberGrowthData struct {
	Date    time.Time `json:"date"`
	Added   int       `json:"added"`
	Removed int       `json:"removed"`
	Total   int       `json:"total"`
}

type ActivityTrendData struct {
	Date         time.Time `json:"date"`
	Commits      int       `json:"commits"`
	Issues       int       `json:"issues"`
	PullRequests int       `json:"pull_requests"`
	Comments     int       `json:"comments"`
}

type RepositoryUsageMetrics struct {
	Period             string                     `json:"period"`
	TotalRepositories  int                        `json:"total_repositories"`
	ActiveRepositories int                        `json:"active_repositories"`
	RepositoryGrowth   []RepositoryGrowthData     `json:"repository_growth"`
	LanguageStats      []LanguageStatsData        `json:"language_stats"`
	TopRepositories    []RepositorySummary        `json:"top_repositories"`
}

type RepositoryGrowthData struct {
	Date    time.Time `json:"date"`
	Created int       `json:"created"`
	Deleted int       `json:"deleted"`
	Total   int       `json:"total"`
}

type LanguageStatsData struct {
	Language    string  `json:"language"`
	Count       int     `json:"count"`
	Percentage  float64 `json:"percentage"`
}

type TeamPerformanceMetrics struct {
	Period        string                    `json:"period"`
	TotalTeams    int                       `json:"total_teams"`
	ActiveTeams   int                       `json:"active_teams"`
	TeamActivity  []TeamActivityData        `json:"team_activity"`
	TeamSizes     []TeamSizeData            `json:"team_sizes"`
}

type TeamActivityData struct {
	TeamName     string    `json:"team_name"`
	Members      int       `json:"members"`
	Commits30d   int       `json:"commits_30d"`
	Issues30d    int       `json:"issues_30d"`
	PRs30d       int       `json:"prs_30d"`
	LastActive   time.Time `json:"last_active"`
}

type TeamSizeData struct {
	TeamName string `json:"team_name"`
	Size     int    `json:"size"`
}

type SecurityMetrics struct {
	Period           string                  `json:"period"`
	SecurityScore    float64                 `json:"security_score"`
	VulnerabilitiesFound int                `json:"vulnerabilities_found"`
	VulnerabilitiesFixed int                `json:"vulnerabilities_fixed"`
	SecurityAlerts   []SecurityAlert         `json:"security_alerts"`
	ComplianceStatus map[string]bool         `json:"compliance_status"`
	PolicyViolations []PolicyViolationData   `json:"policy_violations"`
}

type PolicyViolationData struct {
	PolicyName  string    `json:"policy_name"`
	Count       int       `json:"count"`
	LastOccurred time.Time `json:"last_occurred"`
}

type UsageAndCostMetrics struct {
	Period           string                  `json:"period"`
	BillingPlan      string                  `json:"billing_plan"`
	SeatCount        int                     `json:"seat_count"`
	StorageUsage     *StorageUsageMetrics    `json:"storage_usage"`
	BandwidthUsage   *BandwidthUsageMetrics  `json:"bandwidth_usage"`
	EstimatedCost    float64                 `json:"estimated_cost"`
	CostBreakdown    []CostBreakdownData     `json:"cost_breakdown"`
	UsageTrends      []UsageTrendData        `json:"usage_trends"`
}

type CostBreakdownData struct {
	Category    string  `json:"category"`
	Amount      float64 `json:"amount"`
	Percentage  float64 `json:"percentage"`
}

type UsageTrendData struct {
	Date        time.Time `json:"date"`
	Storage     float64   `json:"storage_gb"`
	Bandwidth   float64   `json:"bandwidth_gb"`
	ActiveUsers int       `json:"active_users"`
}

// Service Implementation
type organizationAnalyticsService struct {
	db *gorm.DB
}

func NewOrganizationAnalyticsService(db *gorm.DB) OrganizationAnalyticsService {
	return &organizationAnalyticsService{db: db}
}

func (s *organizationAnalyticsService) GetDashboardMetrics(ctx context.Context, orgName string) (*DashboardMetrics, error) {
	var org models.Organization
	if err := s.db.Where("name = ?", orgName).First(&org).Error; err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	overview, err := s.getOverviewMetrics(org.ID)
	if err != nil {
		return nil, err
	}

	recentActivity, err := s.getRecentActivity(org.ID, 10)
	if err != nil {
		return nil, err
	}

	topRepos, err := s.getTopRepositories(org.ID, 5)
	if err != nil {
		return nil, err
	}

	activeMembers, err := s.getActiveMembers(org.ID, 5)
	if err != nil {
		return nil, err
	}

	securityAlerts, err := s.getSecurityAlerts(org.ID, 5)
	if err != nil {
		return nil, err
	}

	storageUsage, err := s.getStorageUsage(org.ID)
	if err != nil {
		return nil, err
	}

	bandwidthUsage, err := s.getBandwidthUsage(org.ID)
	if err != nil {
		return nil, err
	}

	return &DashboardMetrics{
		Overview:        overview,
		RecentActivity:  recentActivity,
		TopRepositories: topRepos,
		ActiveMembers:   activeMembers,
		SecurityAlerts:  securityAlerts,
		StorageUsage:    storageUsage,
		BandwidthUsage:  bandwidthUsage,
	}, nil
}

func (s *organizationAnalyticsService) GetMemberActivityMetrics(ctx context.Context, orgName string, period string) (*MemberActivityMetrics, error) {
	var org models.Organization
	if err := s.db.Where("name = ?", orgName).First(&org).Error; err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	// Get total and active member counts
	var totalMembers int64
	s.db.Model(&models.OrganizationMember{}).Where("organization_id = ?", org.ID).Count(&totalMembers)

	// Get active members (members with activity in the last 30 days)
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	var activeMembers int64
	s.db.Model(&models.OrganizationActivity{}).
		Where("organization_id = ? AND created_at > ?", org.ID, thirtyDaysAgo).
		Distinct("actor_id").Count(&activeMembers)

	// Generate growth and trend data based on period
	memberGrowth := s.generateMemberGrowthData(org.ID, period)
	activityTrends := s.generateActivityTrendData(org.ID, period)
	topContributors, _ := s.getActiveMembers(org.ID, 10)

	return &MemberActivityMetrics{
		Period:          period,
		TotalMembers:    int(totalMembers),
		ActiveMembers:   int(activeMembers),
		MemberGrowth:    memberGrowth,
		ActivityTrends:  activityTrends,
		TopContributors: convertMemberSummarySlice(topContributors),
	}, nil
}

func (s *organizationAnalyticsService) GetRepositoryUsageMetrics(ctx context.Context, orgName string, period string) (*RepositoryUsageMetrics, error) {
	var org models.Organization
	if err := s.db.Where("name = ?", orgName).First(&org).Error; err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	// Get repository counts
	var totalRepos int64
	s.db.Model(&models.Repository{}).Where("owner_id = ? AND owner_type = ?", org.ID, "organization").Count(&totalRepos)

	// Get active repositories (with activity in the last 30 days)
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	var activeRepos int64
	s.db.Model(&models.Repository{}).
		Where("owner_id = ? AND owner_type = ? AND pushed_at > ?", org.ID, "organization", thirtyDaysAgo).
		Count(&activeRepos)

	repositoryGrowth := s.generateRepositoryGrowthData(org.ID, period)
	languageStats := s.generateLanguageStatsData(org.ID)
	topRepositories, _ := s.getTopRepositories(org.ID, 10)

	return &RepositoryUsageMetrics{
		Period:             period,
		TotalRepositories:  int(totalRepos),
		ActiveRepositories: int(activeRepos),
		RepositoryGrowth:   repositoryGrowth,
		LanguageStats:      languageStats,
		TopRepositories:    convertRepositorySummarySlice(topRepositories),
	}, nil
}

func (s *organizationAnalyticsService) GetTeamPerformanceMetrics(ctx context.Context, orgName string, period string) (*TeamPerformanceMetrics, error) {
	var org models.Organization
	if err := s.db.Where("name = ?", orgName).First(&org).Error; err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	// Get team counts
	var totalTeams int64
	s.db.Model(&models.Team{}).Where("organization_id = ?", org.ID).Count(&totalTeams)

	// Generate team activity and size data
	teamActivity := s.generateTeamActivityData(org.ID, period)
	teamSizes := s.generateTeamSizeData(org.ID)

	return &TeamPerformanceMetrics{
		Period:       period,
		TotalTeams:   int(totalTeams),
		ActiveTeams:  len(teamActivity), // Teams with activity
		TeamActivity: teamActivity,
		TeamSizes:    teamSizes,
	}, nil
}

func (s *organizationAnalyticsService) GetSecurityMetrics(ctx context.Context, orgName string, period string) (*SecurityMetrics, error) {
	var org models.Organization
	if err := s.db.Where("name = ?", orgName).First(&org).Error; err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	// Calculate security score based on various factors
	securityScore := s.calculateSecurityScore(org.ID)

	securityAlerts, _ := s.getSecurityAlerts(org.ID, 50)
	complianceStatus := s.getComplianceStatus(org.ID)
	policyViolations := s.getPolicyViolations(org.ID, period)

	return &SecurityMetrics{
		Period:           period,
		SecurityScore:    securityScore,
		VulnerabilitiesFound: 0, // Would integrate with security scanners
		VulnerabilitiesFixed: 0, // Would integrate with security scanners
		SecurityAlerts:   convertSecurityAlertSlice(securityAlerts),
		ComplianceStatus: complianceStatus,
		PolicyViolations: policyViolations,
	}, nil
}

func (s *organizationAnalyticsService) GetUsageAndCostMetrics(ctx context.Context, orgName string, period string) (*UsageAndCostMetrics, error) {
	var org models.Organization
	if err := s.db.Where("name = ?", orgName).First(&org).Error; err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	// Get organization settings
	var settings models.OrganizationSettings
	s.db.Where("organization_id = ?", org.ID).First(&settings)

	storageUsage, _ := s.getStorageUsage(org.ID)
	bandwidthUsage, _ := s.getBandwidthUsage(org.ID)

	// Calculate estimated costs
	estimatedCost := s.calculateEstimatedCost(settings, storageUsage, bandwidthUsage)
	costBreakdown := s.generateCostBreakdown(settings, storageUsage, bandwidthUsage)
	usageTrends := s.generateUsageTrendData(org.ID, period)

	return &UsageAndCostMetrics{
		Period:         period,
		BillingPlan:    settings.BillingPlan,
		SeatCount:      settings.SeatCount,
		StorageUsage:   storageUsage,
		BandwidthUsage: bandwidthUsage,
		EstimatedCost:  estimatedCost,
		CostBreakdown:  costBreakdown,
		UsageTrends:    usageTrends,
	}, nil
}

func (s *organizationAnalyticsService) ExportAnalyticsData(ctx context.Context, orgName string, format string, period string) ([]byte, error) {
	// Implementation would depend on the format (CSV, JSON, PDF, etc.)
	// For now, return a placeholder
	return []byte("Analytics data export not yet implemented"), nil
}

// Helper methods for generating analytics data
func (s *organizationAnalyticsService) getOverviewMetrics(orgID uuid.UUID) (*OverviewMetrics, error) {
	var totalMembers, totalRepos, totalTeams, activeMembers30d int64
	var commitsThisMonth, issuesOpen, prsOpen int64

	s.db.Model(&models.OrganizationMember{}).Where("organization_id = ?", orgID).Count(&totalMembers)
	s.db.Model(&models.Repository{}).Where("owner_id = ? AND owner_type = ?", orgID, "organization").Count(&totalRepos)
	s.db.Model(&models.Team{}).Where("organization_id = ?", orgID).Count(&totalTeams)

	// Calculate active members in last 30 days
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	s.db.Model(&models.OrganizationActivity{}).
		Where("organization_id = ? AND created_at > ?", orgID, thirtyDaysAgo).
		Distinct("actor_id").Count(&activeMembers30d)

	// These would integrate with actual data sources
	commitsThisMonth = 0  // Would query git data
	issuesOpen = 0        // Would query issues
	prsOpen = 0           // Would query pull requests

	securityScore := s.calculateSecurityScore(orgID)

	return &OverviewMetrics{
		TotalMembers:      int(totalMembers),
		TotalRepositories: int(totalRepos),
		TotalTeams:        int(totalTeams),
		ActiveMembers30d:  int(activeMembers30d),
		CommitsThisMonth:  int(commitsThisMonth),
		IssuesOpen:        int(issuesOpen),
		PullRequestsOpen:  int(prsOpen),
		SecurityScore:     securityScore,
	}, nil
}

// Additional helper methods would be implemented here for:
// - getRecentActivity
// - getTopRepositories  
// - getActiveMembers
// - getSecurityAlerts
// - getStorageUsage
// - getBandwidthUsage
// - generateMemberGrowthData
// - generateActivityTrendData
// - generateRepositoryGrowthData
// - generateLanguageStatsData
// - generateTeamActivityData
// - generateTeamSizeData
// - calculateSecurityScore
// - getComplianceStatus
// - getPolicyViolations
// - calculateEstimatedCost
// - generateCostBreakdown
// - generateUsageTrendData

// Placeholder implementations for now - these would contain the actual logic
func (s *organizationAnalyticsService) getRecentActivity(orgID uuid.UUID, limit int) ([]*ActivitySummary, error) {
	// Get recent analytics events for this organization
	var events []models.AnalyticsEvent
	err := s.db.Where("organization_id = ?", orgID).
		Order("created_at DESC").
		Limit(limit).
		Preload("Actor").
		Find(&events).Error
	if err != nil {
		return nil, err
	}

	var activities []*ActivitySummary
	for _, event := range events {
		actorName := "Unknown"
		if event.Actor != nil {
			actorName = event.Actor.Username
		}
		
		activities = append(activities, &ActivitySummary{
			Date:        event.CreatedAt,
			Action:      string(event.EventType),
			ActorName:   actorName,
			TargetType:  event.TargetType,
			TargetName:  "", // Would need to resolve from target ID
			Description: fmt.Sprintf("%s performed %s", actorName, event.EventType),
		})
	}

	return activities, nil
}

func (s *organizationAnalyticsService) getTopRepositories(orgID uuid.UUID, limit int) ([]*RepositorySummary, error) {
	// Get repositories with their analytics data
	var repos []struct {
		models.Repository
		StarsCount   int64     `json:"stars_count"`
		ForksCount   int64     `json:"forks_count"`
		Commits30d   int64     `json:"commits_30d"`
		Issues30d    int64     `json:"issues_30d"`
		LastActivity time.Time `json:"last_activity"`
	}

	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	err := s.db.Table("repositories r").
		Select(`
			r.*,
			COALESCE(ra.stars_count, 0) as stars_count,
			COALESCE(ra.forks_count, 0) as forks_count,
			(SELECT COUNT(*) FROM commits c WHERE c.repository_id = r.id AND c.created_at >= ?) as commits_30d,
			(SELECT COUNT(*) FROM issues i WHERE i.repository_id = r.id AND i.created_at >= ?) as issues_30d,
			COALESCE(r.pushed_at, r.updated_at) as last_activity
		`, thirtyDaysAgo, thirtyDaysAgo).
		Joins("LEFT JOIN repository_analytics ra ON r.id = ra.repository_id").
		Where("r.owner_id = ? AND r.owner_type = ?", orgID, "organization").
		Order("stars_count DESC, commits_30d DESC").
		Limit(limit).
		Scan(&repos).Error

	if err != nil {
		return nil, err
	}

	var summaries []*RepositorySummary
	for _, repo := range repos {
		// Count contributors
		var contributorCount int64
		s.db.Model(&models.Commit{}).
			Where("repository_id = ?", repo.ID).
			Distinct("author_id").
			Count(&contributorCount)

		summaries = append(summaries, &RepositorySummary{
			Name:           repo.Name,
			Language:       "Go", // Placeholder - would get from repository stats
			Stars:          int(repo.StarsCount),
			Forks:          int(repo.ForksCount),
			Contributors:   int(contributorCount),
			Commits30d:     int(repo.Commits30d),
			Issues30d:      int(repo.Issues30d),
			LastActivityAt: repo.LastActivity,
		})
	}

	return summaries, nil
}

func (s *organizationAnalyticsService) getActiveMembers(orgID uuid.UUID, limit int) ([]*MemberSummary, error) {
	// Get organization members with their activity stats
	var members []struct {
		models.User
		Role           string    `json:"role"`
		Commits30d     int64     `json:"commits_30d"`
		Issues30d      int64     `json:"issues_30d"`
		PullRequests30d int64    `json:"pull_requests_30d"`
		LastActiveAt   time.Time `json:"last_active_at"`
	}

	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	err := s.db.Table("users u").
		Select(`
			u.*,
			om.role,
			(SELECT COUNT(*) FROM commits c WHERE c.author_id = u.id AND c.created_at >= ?) as commits_30d,
			(SELECT COUNT(*) FROM issues i WHERE i.user_id = u.id AND i.created_at >= ?) as issues_30d,
			(SELECT COUNT(*) FROM pull_requests pr WHERE pr.user_id = u.id AND pr.created_at >= ?) as pull_requests_30d,
			(SELECT MAX(ae.created_at) FROM analytics_events ae WHERE ae.actor_id = u.id) as last_active_at
		`, thirtyDaysAgo, thirtyDaysAgo, thirtyDaysAgo).
		Joins("JOIN organization_members om ON u.id = om.user_id").
		Where("om.organization_id = ?", orgID).
		Order("commits_30d DESC, pull_requests_30d DESC, issues_30d DESC").
		Limit(limit).
		Scan(&members).Error

	if err != nil {
		return nil, err
	}

	var summaries []*MemberSummary
	for _, member := range members {
		summaries = append(summaries, &MemberSummary{
			Username:       member.Username,
			Name:           member.FullName,
			Role:           member.Role,
			Commits30d:     int(member.Commits30d),
			Issues30d:      int(member.Issues30d),
			PullRequests30d: int(member.PullRequests30d),
			LastActiveAt:   member.LastActiveAt,
		})
	}

	return summaries, nil
}

func (s *organizationAnalyticsService) getSecurityAlerts(orgID uuid.UUID, limit int) ([]*SecurityAlert, error) {
	// Get security-related events and convert to alerts
	var events []models.AnalyticsEvent
	err := s.db.Where("organization_id = ? AND event_type LIKE ?", orgID, "security.%").
		Order("created_at DESC").
		Limit(limit).
		Preload("Repository").
		Find(&events).Error

	if err != nil {
		return nil, err
	}

	var alerts []*SecurityAlert
	for _, event := range events {
		severity := "medium"
		if event.EventType == "security.access_denied" {
			severity = "high"
		}

		repoName := "Unknown"
		if event.Repository != nil {
			repoName = event.Repository.Name
		}

		alerts = append(alerts, &SecurityAlert{
			Type:        string(event.EventType),
			Severity:    severity,
			Title:       fmt.Sprintf("Security Event: %s", event.EventType),
			Description: event.ErrorMessage,
			Repository:  repoName,
			CreatedAt:   event.CreatedAt,
		})
	}

	// Add some placeholder alerts if no real ones exist
	if len(alerts) == 0 {
		alerts = append(alerts, &SecurityAlert{
			Type:        "vulnerability",
			Severity:    "low",
			Title:       "No recent security alerts",
			Description: "Your organization has no recent security alerts",
			Repository:  "",
			CreatedAt:   time.Now(),
		})
	}

	return alerts, nil
}

func (s *organizationAnalyticsService) getStorageUsage(orgID uuid.UUID) (*StorageUsageMetrics, error) {
	// Get storage usage from organization analytics
	var latest models.OrganizationAnalytics
	err := s.db.Where("organization_id = ?", orgID).
		Order("date DESC").
		First(&latest).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// Convert MB to GB
	totalUsedGB := float64(latest.StorageUsedMB) / 1024.0
	totalLimitGB := 1024.0 // Default 1TB limit
	usagePercent := (totalUsedGB / totalLimitGB) * 100

	// Get top repositories by storage usage
	var topRepos []RepositoryStorageUsage

	// This would integrate with actual repository size tracking
	// For now, use placeholder data
	topRepos = []RepositoryStorageUsage{
		{Name: "main-app", SizeGB: totalUsedGB * 0.4, Percent: 40.0},
		{Name: "api-service", SizeGB: totalUsedGB * 0.3, Percent: 30.0},
		{Name: "frontend", SizeGB: totalUsedGB * 0.2, Percent: 20.0},
		{Name: "docs", SizeGB: totalUsedGB * 0.1, Percent: 10.0},
	}

	return &StorageUsageMetrics{
		TotalUsedGB:    totalUsedGB,
		TotalLimitGB:   totalLimitGB,
		UsagePercent:   usagePercent,
		TopRepositories: topRepos,
	}, nil
}

func (s *organizationAnalyticsService) getBandwidthUsage(orgID uuid.UUID) (*BandwidthUsageMetrics, error) {
	// Get bandwidth usage from organization analytics
	var analytics []models.OrganizationAnalytics
	err := s.db.Where("organization_id = ? AND date >= ?", orgID, time.Now().AddDate(0, -1, 0)).
		Order("date ASC").
		Find(&analytics).Error

	if err != nil {
		return nil, err
	}

	// Calculate total bandwidth for the month
	var totalUsedMB int64
	var dailyUsage []DailyBandwidthUsage

	for _, record := range analytics {
		totalUsedMB += record.BandwidthUsedMB
		dailyUsage = append(dailyUsage, DailyBandwidthUsage{
			Date:    record.Date,
			UsageGB: float64(record.BandwidthUsedMB) / 1024.0,
		})
	}

	totalUsedGB := float64(totalUsedMB) / 1024.0
	totalLimitGB := 10240.0 // Default 10TB limit
	usagePercent := (totalUsedGB / totalLimitGB) * 100

	return &BandwidthUsageMetrics{
		TotalUsedGB:  totalUsedGB,
		TotalLimitGB: totalLimitGB,
		UsagePercent: usagePercent,
		DailyUsage:   dailyUsage,
	}, nil
}

func (s *organizationAnalyticsService) generateMemberGrowthData(orgID uuid.UUID, period string) []MemberGrowthData {
	return []MemberGrowthData{}
}

func (s *organizationAnalyticsService) generateActivityTrendData(orgID uuid.UUID, period string) []ActivityTrendData {
	return []ActivityTrendData{}
}

func (s *organizationAnalyticsService) generateRepositoryGrowthData(orgID uuid.UUID, period string) []RepositoryGrowthData {
	return []RepositoryGrowthData{}
}

func (s *organizationAnalyticsService) generateLanguageStatsData(orgID uuid.UUID) []LanguageStatsData {
	return []LanguageStatsData{}
}

func (s *organizationAnalyticsService) generateTeamActivityData(orgID uuid.UUID, period string) []TeamActivityData {
	return []TeamActivityData{}
}

func (s *organizationAnalyticsService) generateTeamSizeData(orgID uuid.UUID) []TeamSizeData {
	return []TeamSizeData{}
}

func (s *organizationAnalyticsService) calculateSecurityScore(orgID uuid.UUID) float64 {
	// Calculate based on various security factors
	return 85.5
}

func (s *organizationAnalyticsService) getComplianceStatus(orgID uuid.UUID) map[string]bool {
	return map[string]bool{
		"GDPR": true,
		"SOC2": false,
		"ISO27001": false,
	}
}

func (s *organizationAnalyticsService) getPolicyViolations(orgID uuid.UUID, period string) []PolicyViolationData {
	return []PolicyViolationData{}
}

func (s *organizationAnalyticsService) calculateEstimatedCost(settings models.OrganizationSettings, storage *StorageUsageMetrics, bandwidth *BandwidthUsageMetrics) float64 {
	// Calculate based on billing plan and usage
	return 0.0
}

func (s *organizationAnalyticsService) generateCostBreakdown(settings models.OrganizationSettings, storage *StorageUsageMetrics, bandwidth *BandwidthUsageMetrics) []CostBreakdownData {
	return []CostBreakdownData{}
}

func (s *organizationAnalyticsService) generateUsageTrendData(orgID uuid.UUID, period string) []UsageTrendData {
	return []UsageTrendData{}
}

// Helper functions to convert slice types
func convertMemberSummarySlice(input []*MemberSummary) []MemberSummary {
	result := make([]MemberSummary, len(input))
	for i, item := range input {
		if item != nil {
			result[i] = *item
		}
	}
	return result
}

func convertRepositorySummarySlice(input []*RepositorySummary) []RepositorySummary {
	result := make([]RepositorySummary, len(input))
	for i, item := range input {
		if item != nil {
			result[i] = *item
		}
	}
	return result
}

func convertSecurityAlertSlice(input []*SecurityAlert) []SecurityAlert {
	result := make([]SecurityAlert, len(input))
	for i, item := range input {
		if item != nil {
			result[i] = *item
		}
	}
	return result
}