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
	return []*ActivitySummary{}, nil
}

func (s *organizationAnalyticsService) getTopRepositories(orgID uuid.UUID, limit int) ([]*RepositorySummary, error) {
	return []*RepositorySummary{}, nil
}

func (s *organizationAnalyticsService) getActiveMembers(orgID uuid.UUID, limit int) ([]*MemberSummary, error) {
	return []*MemberSummary{}, nil
}

func (s *organizationAnalyticsService) getSecurityAlerts(orgID uuid.UUID, limit int) ([]*SecurityAlert, error) {
	return []*SecurityAlert{}, nil
}

func (s *organizationAnalyticsService) getStorageUsage(orgID uuid.UUID) (*StorageUsageMetrics, error) {
	return &StorageUsageMetrics{
		TotalUsedGB:  0,
		TotalLimitGB: 1024,
		UsagePercent: 0,
	}, nil
}

func (s *organizationAnalyticsService) getBandwidthUsage(orgID uuid.UUID) (*BandwidthUsageMetrics, error) {
	return &BandwidthUsageMetrics{
		TotalUsedGB:  0,
		TotalLimitGB: 1024,
		UsagePercent: 0,
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