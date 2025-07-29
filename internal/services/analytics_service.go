package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/a5c-ai/hub/internal/models"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// AnalyticsService provides analytics and monitoring operations
type AnalyticsService interface {
	// Event tracking
	RecordEvent(ctx context.Context, event *models.AnalyticsEvent) error
	GetEvents(ctx context.Context, filters EventFilters) ([]*models.AnalyticsEvent, int64, error)

	// Metrics recording and querying
	RecordMetric(ctx context.Context, metric *models.AnalyticsMetric) error
	GetMetrics(ctx context.Context, filters MetricFilters) ([]*models.AnalyticsMetric, error)

	// Repository analytics
	GetRepositoryAnalytics(ctx context.Context, repoID uuid.UUID, period Period) (*models.RepositoryAnalytics, error)
	UpdateRepositoryAnalytics(ctx context.Context, repoID uuid.UUID, date time.Time) error
	GetRepositoryInsights(ctx context.Context, repoID uuid.UUID, filters InsightFilters) (*RepositoryInsights, error)
	GetRepositoryCodeStats(ctx context.Context, repoID uuid.UUID) (*CodeStatistics, error)
	GetRepositoryContributorStats(ctx context.Context, repoID uuid.UUID, filters InsightFilters) (*ContributorStatistics, error)
	GetRepositoryActivityStats(ctx context.Context, repoID uuid.UUID, filters InsightFilters) (*ActivityStatistics, error)

	GetRepositoryPRStats(ctx context.Context, repoID uuid.UUID, filters InsightFilters) (*PullRequestStatistics, error)
	GetRepositoryPerformanceStats(ctx context.Context, repoID uuid.UUID, filters InsightFilters) (*PerformanceStatistics, error)

	// User analytics
	GetUserAnalytics(ctx context.Context, userID uuid.UUID, period Period) (*models.UserAnalytics, error)
	UpdateUserAnalytics(ctx context.Context, userID uuid.UUID, date time.Time) error
	GetUserInsights(ctx context.Context, userID uuid.UUID, filters InsightFilters) (*UserInsights, error)

	// Organization analytics
	GetOrganizationAnalytics(ctx context.Context, orgID uuid.UUID, period Period) (*models.OrganizationAnalytics, error)
	UpdateOrganizationAnalytics(ctx context.Context, orgID uuid.UUID, date time.Time) error
	GetOrganizationInsights(ctx context.Context, orgID uuid.UUID, filters InsightFilters) (*OrganizationInsights, error)

	// System analytics
	GetSystemAnalytics(ctx context.Context, period Period) (*models.SystemAnalytics, error)
	UpdateSystemAnalytics(ctx context.Context, date time.Time) error
	GetSystemInsights(ctx context.Context, filters InsightFilters) (*SystemInsights, error)

	// Performance monitoring
	RecordPerformanceLog(ctx context.Context, log *models.PerformanceLog) error
	GetPerformanceLogs(ctx context.Context, filters PerformanceFilters) ([]*models.PerformanceLog, int64, error)
	GetPerformanceMetrics(ctx context.Context, filters PerformanceFilters) (*PerformanceMetrics, error)

	// Data aggregation and reporting
	AggregateMetrics(ctx context.Context, period Period) error
	GenerateReport(ctx context.Context, reportType ReportType, filters ReportFilters) (*Report, error)
	ExportData(ctx context.Context, exportType ExportType, filters ExportFilters) ([]byte, error)
}

// Period represents time periods for analytics
type Period string

const (
	PeriodHourly  Period = "hourly"
	PeriodDaily   Period = "daily"
	PeriodWeekly  Period = "weekly"
	PeriodMonthly Period = "monthly"
	PeriodYearly  Period = "yearly"
)

// EventFilters for filtering analytics events
type EventFilters struct {
	EventTypes     []models.EventType `json:"event_types,omitempty"`
	ActorID        *uuid.UUID         `json:"actor_id,omitempty"`
	RepositoryID   *uuid.UUID         `json:"repository_id,omitempty"`
	OrganizationID *uuid.UUID         `json:"organization_id,omitempty"`
	StartDate      *time.Time         `json:"start_date,omitempty"`
	EndDate        *time.Time         `json:"end_date,omitempty"`
	Status         string             `json:"status,omitempty"`
	Limit          int                `json:"limit,omitempty"`
	Offset         int                `json:"offset,omitempty"`
}

// MetricFilters for filtering analytics metrics
type MetricFilters struct {
	Names          []string   `json:"names,omitempty"`
	MetricType     string     `json:"metric_type,omitempty"`
	RepositoryID   *uuid.UUID `json:"repository_id,omitempty"`
	OrganizationID *uuid.UUID `json:"organization_id,omitempty"`
	UserID         *uuid.UUID `json:"user_id,omitempty"`
	Period         Period     `json:"period,omitempty"`
	StartDate      *time.Time `json:"start_date,omitempty"`
	EndDate        *time.Time `json:"end_date,omitempty"`
}

// InsightFilters for filtering insight data
type InsightFilters struct {
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	Period    Period     `json:"period,omitempty"`
}

// PerformanceFilters for filtering performance logs
type PerformanceFilters struct {
	Methods        []string   `json:"methods,omitempty"`
	Paths          []string   `json:"paths,omitempty"`
	StatusCodes    []int      `json:"status_codes,omitempty"`
	UserID         *uuid.UUID `json:"user_id,omitempty"`
	RepositoryID   *uuid.UUID `json:"repository_id,omitempty"`
	OrganizationID *uuid.UUID `json:"organization_id,omitempty"`
	MinDuration    *int64     `json:"min_duration,omitempty"`
	MaxDuration    *int64     `json:"max_duration,omitempty"`
	StartDate      *time.Time `json:"start_date,omitempty"`
	EndDate        *time.Time `json:"end_date,omitempty"`
	Limit          int        `json:"limit,omitempty"`
	Offset         int        `json:"offset,omitempty"`
}

// Repository Insights
type RepositoryInsights struct {
	Repository       *models.Repository            `json:"repository"`
	Analytics        []*models.RepositoryAnalytics `json:"analytics"`
	CodeStats        *CodeStatistics               `json:"code_stats"`
	ActivityStats    *ActivityStatistics           `json:"activity_stats"`
	ContributorStats *ContributorStatistics        `json:"contributor_stats"`

	PullRequestStats *PullRequestStatistics `json:"pull_request_stats"`
	PerformanceStats *PerformanceStatistics `json:"performance_stats"`
}

type CodeStatistics struct {
	TotalLinesOfCode  int64             `json:"total_lines_of_code"`
	TotalFiles        int64             `json:"total_files"`
	TotalCommits      int64             `json:"total_commits"`
	TotalBranches     int64             `json:"total_branches"`
	LanguageBreakdown map[string]int64  `json:"language_breakdown"`
	CommitActivity    []TimeSeriesPoint `json:"commit_activity"`
}

type ActivityStatistics struct {
	TotalViews    int64             `json:"total_views"`
	TotalClones   int64             `json:"total_clones"`
	TotalForks    int64             `json:"total_forks"`
	TotalStars    int64             `json:"total_stars"`
	TotalWatchers int64             `json:"total_watchers"`
	ActivityTrend []TimeSeriesPoint `json:"activity_trend"`
}

type ContributorStatistics struct {
	TotalContributors   int64             `json:"total_contributors"`
	ActiveContributors  int64             `json:"active_contributors"`
	TopContributors     []ContributorStat `json:"top_contributors"`
	ContributorActivity []TimeSeriesPoint `json:"contributor_activity"`
}

type ContributorStat struct {
	UserID       uuid.UUID `json:"user_id"`
	Username     string    `json:"username"`
	CommitCount  int64     `json:"commit_count"`
	LinesAdded   int64     `json:"lines_added"`
	LinesDeleted int64     `json:"lines_deleted"`
}

type PullRequestStatistics struct {
	TotalPullRequests  int64             `json:"total_pull_requests"`
	OpenPullRequests   int64             `json:"open_pull_requests"`
	MergedPullRequests int64             `json:"merged_pull_requests"`
	ClosedPullRequests int64             `json:"closed_pull_requests"`
	AvgTimeToMerge     *float64          `json:"avg_time_to_merge"`
	PRActivity         []TimeSeriesPoint `json:"pr_activity"`
}

type PerformanceStatistics struct {
	AvgResponseTime *float64          `json:"avg_response_time"`
	P95ResponseTime *float64          `json:"p95_response_time"`
	ErrorRate       *float64          `json:"error_rate"`
	ThroughputTrend []TimeSeriesPoint `json:"throughput_trend"`
}

// User Insights
type UserInsights struct {
	User              *models.User            `json:"user"`
	Analytics         []*models.UserAnalytics `json:"analytics"`
	ActivityStats     *UserActivityStats      `json:"activity_stats"`
	ContributionStats *UserContributionStats  `json:"contribution_stats"`
	RepositoryStats   *UserRepositoryStats    `json:"repository_stats"`
}

type UserActivityStats struct {
	TotalLogins    int64             `json:"total_logins"`
	TotalSessions  int64             `json:"total_sessions"`
	AvgSessionTime *float64          `json:"avg_session_time"`
	TotalPageViews int64             `json:"total_page_views"`
	ActivityTrend  []TimeSeriesPoint `json:"activity_trend"`
}

type UserContributionStats struct {
	TotalCommits      int64             `json:"total_commits"`
	TotalPullRequests int64             `json:"total_pull_requests"`
	TotalComments     int64             `json:"total_comments"`
	ContributionTrend []TimeSeriesPoint `json:"contribution_trend"`
}

type UserRepositoryStats struct {
	TotalRepositories int64             `json:"total_repositories"`
	TotalStars        int64             `json:"total_stars"`
	TotalForks        int64             `json:"total_forks"`
	RepositoryTrend   []TimeSeriesPoint `json:"repository_trend"`
}

// Organization Insights
type OrganizationInsights struct {
	Organization    *models.Organization            `json:"organization"`
	Analytics       []*models.OrganizationAnalytics `json:"analytics"`
	MemberStats     *OrganizationMemberStats        `json:"member_stats"`
	RepositoryStats *OrganizationRepositoryStats    `json:"repository_stats"`
	ActivityStats   *OrganizationActivityStats      `json:"activity_stats"`
	ResourceStats   *OrganizationResourceStats      `json:"resource_stats"`
}

type OrganizationMemberStats struct {
	TotalMembers  int64             `json:"total_members"`
	ActiveMembers int64             `json:"active_members"`
	TotalTeams    int64             `json:"total_teams"`
	MemberTrend   []TimeSeriesPoint `json:"member_trend"`
}

type OrganizationRepositoryStats struct {
	TotalRepositories   int64             `json:"total_repositories"`
	PublicRepositories  int64             `json:"public_repositories"`
	PrivateRepositories int64             `json:"private_repositories"`
	RepositoryTrend     []TimeSeriesPoint `json:"repository_trend"`
}

type OrganizationActivityStats struct {
	TotalCommits      int64             `json:"total_commits"`
	TotalPullRequests int64             `json:"total_pull_requests"`
	ActivityTrend     []TimeSeriesPoint `json:"activity_trend"`
}

type OrganizationResourceStats struct {
	StorageUsedMB      int64             `json:"storage_used_mb"`
	BandwidthUsedMB    int64             `json:"bandwidth_used_mb"`
	ComputeTimeMinutes int64             `json:"compute_time_minutes"`
	EstimatedCost      *float64          `json:"estimated_cost"`
	ResourceTrend      []TimeSeriesPoint `json:"resource_trend"`
}

// System Insights
type SystemInsights struct {
	Analytics        []*models.SystemAnalytics `json:"analytics"`
	UserStats        *SystemUserStats          `json:"user_stats"`
	RepositoryStats  *SystemRepositoryStats    `json:"repository_stats"`
	PerformanceStats *SystemPerformanceStats   `json:"performance_stats"`
	ResourceStats    *SystemResourceStats      `json:"resource_stats"`
}

type SystemUserStats struct {
	TotalUsers       int64             `json:"total_users"`
	ActiveUsers      int64             `json:"active_users"`
	NewRegistrations int64             `json:"new_registrations"`
	ChurnRate        *float64          `json:"churn_rate"`
	UserTrend        []TimeSeriesPoint `json:"user_trend"`
}

type SystemRepositoryStats struct {
	TotalRepositories   int64             `json:"total_repositories"`
	PublicRepositories  int64             `json:"public_repositories"`
	PrivateRepositories int64             `json:"private_repositories"`
	TotalOrganizations  int64             `json:"total_organizations"`
	RepositoryTrend     []TimeSeriesPoint `json:"repository_trend"`
}

type SystemPerformanceStats struct {
	AvgResponseTime  *float64          `json:"avg_response_time"`
	P95ResponseTime  *float64          `json:"p95_response_time"`
	ErrorRate        *float64          `json:"error_rate"`
	Uptime           *float64          `json:"uptime"`
	PerformanceTrend []TimeSeriesPoint `json:"performance_trend"`
}

type SystemResourceStats struct {
	CPUUsage      *float64          `json:"cpu_usage"`
	MemoryUsage   *float64          `json:"memory_usage"`
	DiskUsage     *float64          `json:"disk_usage"`
	NetworkInMB   int64             `json:"network_in_mb"`
	NetworkOutMB  int64             `json:"network_out_mb"`
	ResourceTrend []TimeSeriesPoint `json:"resource_trend"`
}

// Performance Metrics
type PerformanceMetrics struct {
	AvgResponseTime   *float64          `json:"avg_response_time"`
	P50ResponseTime   *float64          `json:"p50_response_time"`
	P95ResponseTime   *float64          `json:"p95_response_time"`
	P99ResponseTime   *float64          `json:"p99_response_time"`
	TotalRequests     int64             `json:"total_requests"`
	ErrorRate         *float64          `json:"error_rate"`
	ThroughputPerMin  *float64          `json:"throughput_per_min"`
	ResponseTimeTrend []TimeSeriesPoint `json:"response_time_trend"`
	ErrorRateTrend    []TimeSeriesPoint `json:"error_rate_trend"`
}

// Common types
type TimeSeriesPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// Report types
type ReportType string

const (
	ReportTypeRepository   ReportType = "repository"
	ReportTypeUser         ReportType = "user"
	ReportTypeOrganization ReportType = "organization"
	ReportTypeSystem       ReportType = "system"
	ReportTypePerformance  ReportType = "performance"
)

type ReportFilters struct {
	Type           ReportType `json:"type"`
	TargetID       *uuid.UUID `json:"target_id,omitempty"`
	StartDate      *time.Time `json:"start_date,omitempty"`
	EndDate        *time.Time `json:"end_date,omitempty"`
	Period         Period     `json:"period,omitempty"`
	IncludeTrends  bool       `json:"include_trends,omitempty"`
	IncludeDetails bool       `json:"include_details,omitempty"`
}

type Report struct {
	Type        ReportType  `json:"type"`
	TargetID    *uuid.UUID  `json:"target_id,omitempty"`
	Period      Period      `json:"period"`
	StartDate   time.Time   `json:"start_date"`
	EndDate     time.Time   `json:"end_date"`
	GeneratedAt time.Time   `json:"generated_at"`
	Data        interface{} `json:"data"`
}

// Export types
type ExportType string

const (
	ExportTypeCSV  ExportType = "csv"
	ExportTypeJSON ExportType = "json"
	ExportTypeXLSX ExportType = "xlsx"
)

type ExportFilters struct {
	Type           ExportType `json:"type"`
	DataType       string     `json:"data_type"` // events, metrics, analytics
	TargetID       *uuid.UUID `json:"target_id,omitempty"`
	StartDate      *time.Time `json:"start_date,omitempty"`
	EndDate        *time.Time `json:"end_date,omitempty"`
	IncludeHeaders bool       `json:"include_headers,omitempty"`
}

// analyticsService implements AnalyticsService
type analyticsService struct {
	db     *gorm.DB
	logger *logrus.Logger
}

// NewAnalyticsService creates a new analytics service
func NewAnalyticsService(db *gorm.DB, logger *logrus.Logger) AnalyticsService {
	return &analyticsService{
		db:     db,
		logger: logger,
	}
}

// RecordEvent records an analytics event
func (s *analyticsService) RecordEvent(ctx context.Context, event *models.AnalyticsEvent) error {
	if err := s.db.WithContext(ctx).Create(event).Error; err != nil {
		s.logger.WithError(err).Error("Failed to record analytics event")
		return fmt.Errorf("failed to record analytics event: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"event_type": event.EventType,
		"actor_id":   event.ActorID,
		"target_id":  event.TargetID,
	}).Debug("Analytics event recorded")

	return nil
}

// GetEvents retrieves analytics events based on filters
func (s *analyticsService) GetEvents(ctx context.Context, filters EventFilters) ([]*models.AnalyticsEvent, int64, error) {
	query := s.db.WithContext(ctx).Model(&models.AnalyticsEvent{})

	// Apply filters
	if len(filters.EventTypes) > 0 {
		query = query.Where("event_type IN ?", filters.EventTypes)
	}
	if filters.ActorID != nil {
		query = query.Where("actor_id = ?", *filters.ActorID)
	}
	if filters.RepositoryID != nil {
		query = query.Where("repository_id = ?", *filters.RepositoryID)
	}
	if filters.OrganizationID != nil {
		query = query.Where("organization_id = ?", *filters.OrganizationID)
	}
	if filters.StartDate != nil {
		query = query.Where("created_at >= ?", *filters.StartDate)
	}
	if filters.EndDate != nil {
		query = query.Where("created_at <= ?", *filters.EndDate)
	}
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count events: %w", err)
	}

	// Apply pagination
	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}
	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	// Get events
	var events []*models.AnalyticsEvent
	if err := query.Order("created_at DESC").Find(&events).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get events: %w", err)
	}

	return events, total, nil
}

// RecordMetric records an analytics metric
func (s *analyticsService) RecordMetric(ctx context.Context, metric *models.AnalyticsMetric) error {
	if err := s.db.WithContext(ctx).Create(metric).Error; err != nil {
		s.logger.WithError(err).Error("Failed to record analytics metric")
		return fmt.Errorf("failed to record analytics metric: %w", err)
	}

	return nil
}

// GetMetrics retrieves analytics metrics based on filters
func (s *analyticsService) GetMetrics(ctx context.Context, filters MetricFilters) ([]*models.AnalyticsMetric, error) {
	query := s.db.WithContext(ctx).Model(&models.AnalyticsMetric{})

	// Apply filters
	if len(filters.Names) > 0 {
		query = query.Where("name IN ?", filters.Names)
	}
	if filters.MetricType != "" {
		query = query.Where("metric_type = ?", filters.MetricType)
	}
	if filters.RepositoryID != nil {
		query = query.Where("repository_id = ?", *filters.RepositoryID)
	}
	if filters.OrganizationID != nil {
		query = query.Where("organization_id = ?", *filters.OrganizationID)
	}
	if filters.UserID != nil {
		query = query.Where("user_id = ?", *filters.UserID)
	}
	if filters.Period != "" {
		query = query.Where("period = ?", filters.Period)
	}
	if filters.StartDate != nil {
		query = query.Where("timestamp >= ?", *filters.StartDate)
	}
	if filters.EndDate != nil {
		query = query.Where("timestamp <= ?", *filters.EndDate)
	}

	// Get metrics
	var metrics []*models.AnalyticsMetric
	if err := query.Order("timestamp DESC").Find(&metrics).Error; err != nil {
		return nil, fmt.Errorf("failed to get metrics: %w", err)
	}

	return metrics, nil
}

// Placeholder implementations for other methods (to be implemented)
func (s *analyticsService) GetRepositoryAnalytics(ctx context.Context, repoID uuid.UUID, period Period) (*models.RepositoryAnalytics, error) {
	var analytics models.RepositoryAnalytics
	err := s.db.WithContext(ctx).Where("repository_id = ?", repoID).
		Order("date DESC").First(&analytics).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("repository analytics not found")
		}
		return nil, fmt.Errorf("failed to get repository analytics: %w", err)
	}
	return &analytics, nil
}

func (s *analyticsService) GetRepositoryCodeStats(ctx context.Context, repoID uuid.UUID) (*CodeStatistics, error) {
	return s.getRepositoryCodeStats(ctx, repoID)
}

func (s *analyticsService) GetRepositoryContributorStats(ctx context.Context, repoID uuid.UUID, filters InsightFilters) (*ContributorStatistics, error) {
	return s.getRepositoryContributorStats(ctx, repoID, filters)
}

func (s *analyticsService) GetRepositoryActivityStats(ctx context.Context, repoID uuid.UUID, filters InsightFilters) (*ActivityStatistics, error) {
	return s.getRepositoryActivityStats(ctx, repoID, filters)
}

func (s *analyticsService) GetRepositoryPRStats(ctx context.Context, repoID uuid.UUID, filters InsightFilters) (*PullRequestStatistics, error) {
	return s.getRepositoryPRStats(ctx, repoID, filters)
}

func (s *analyticsService) GetRepositoryPerformanceStats(ctx context.Context, repoID uuid.UUID, filters InsightFilters) (*PerformanceStatistics, error) {
	return s.getRepositoryPerformanceStats(ctx, repoID, filters)
}

func (s *analyticsService) UpdateRepositoryAnalytics(ctx context.Context, repoID uuid.UUID, date time.Time) error {
	// Implementation will be added
	return fmt.Errorf("not implemented yet")
}

func (s *analyticsService) GetRepositoryInsights(ctx context.Context, repoID uuid.UUID, filters InsightFilters) (*RepositoryInsights, error) {
	// Get repository details
	var repository models.Repository
	if err := s.db.WithContext(ctx).Where("id = ?", repoID).First(&repository).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("repository not found")
		}
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}

	// Get repository analytics data
	analytics, err := s.getRepositoryAnalyticsData(ctx, repoID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository analytics: %w", err)
	}

	// Get code statistics
	codeStats, err := s.getRepositoryCodeStats(ctx, repoID)
	if err != nil {
		return nil, fmt.Errorf("failed to get code stats: %w", err)
	}

	// Get activity statistics
	activityStats, err := s.getRepositoryActivityStats(ctx, repoID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity stats: %w", err)
	}

	// Get contributor statistics
	contributorStats, err := s.getRepositoryContributorStats(ctx, repoID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get contributor stats: %w", err)
	}

	// Get pull request statistics
	prStats, err := s.getRepositoryPRStats(ctx, repoID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get PR stats: %w", err)
	}

	// Get performance statistics
	perfStats, err := s.getRepositoryPerformanceStats(ctx, repoID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get performance stats: %w", err)
	}

	return &RepositoryInsights{
		Repository:       &repository,
		Analytics:        analytics,
		CodeStats:        codeStats,
		ActivityStats:    activityStats,
		ContributorStats: contributorStats,

		PullRequestStats: prStats,
		PerformanceStats: perfStats,
	}, nil
}

func (s *analyticsService) GetUserAnalytics(ctx context.Context, userID uuid.UUID, period Period) (*models.UserAnalytics, error) {
	// Implementation will be added
	return nil, fmt.Errorf("not implemented yet")
}

func (s *analyticsService) UpdateUserAnalytics(ctx context.Context, userID uuid.UUID, date time.Time) error {
	// Implementation will be added
	return fmt.Errorf("not implemented yet")
}

func (s *analyticsService) GetUserInsights(ctx context.Context, userID uuid.UUID, filters InsightFilters) (*UserInsights, error) {
	// Get user details
	var user models.User
	if err := s.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Get user analytics data
	analytics, err := s.getUserAnalyticsData(ctx, userID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get user analytics: %w", err)
	}

	// Get activity statistics
	activityStats, err := s.getUserActivityStats(ctx, userID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get user activity stats: %w", err)
	}

	// Get contribution statistics
	contributionStats, err := s.getUserContributionStats(ctx, userID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get user contribution stats: %w", err)
	}

	// Get repository statistics
	repositoryStats, err := s.getUserRepositoryStats(ctx, userID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get user repository stats: %w", err)
	}

	return &UserInsights{
		User:              &user,
		Analytics:         analytics,
		ActivityStats:     activityStats,
		ContributionStats: contributionStats,
		RepositoryStats:   repositoryStats,
	}, nil
}

func (s *analyticsService) GetOrganizationAnalytics(ctx context.Context, orgID uuid.UUID, period Period) (*models.OrganizationAnalytics, error) {
	// Implementation will be added
	return nil, fmt.Errorf("not implemented yet")
}

func (s *analyticsService) UpdateOrganizationAnalytics(ctx context.Context, orgID uuid.UUID, date time.Time) error {
	// Implementation will be added
	return fmt.Errorf("not implemented yet")
}

func (s *analyticsService) GetOrganizationInsights(ctx context.Context, orgID uuid.UUID, filters InsightFilters) (*OrganizationInsights, error) {
	// Get organization details
	var organization models.Organization
	if err := s.db.WithContext(ctx).Where("id = ?", orgID).First(&organization).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("organization not found")
		}
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	// Get organization analytics data
	analytics, err := s.getOrganizationAnalyticsData(ctx, orgID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization analytics: %w", err)
	}

	// Get member statistics
	memberStats, err := s.getOrganizationMemberStats(ctx, orgID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get member stats: %w", err)
	}

	// Get repository statistics
	repositoryStats, err := s.getOrganizationRepositoryStats(ctx, orgID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository stats: %w", err)
	}

	// Get activity statistics
	activityStats, err := s.getOrganizationActivityStats(ctx, orgID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity stats: %w", err)
	}

	// Get resource statistics
	resourceStats, err := s.getOrganizationResourceStats(ctx, orgID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource stats: %w", err)
	}

	return &OrganizationInsights{
		Organization:    &organization,
		Analytics:       analytics,
		MemberStats:     memberStats,
		RepositoryStats: repositoryStats,
		ActivityStats:   activityStats,
		ResourceStats:   resourceStats,
	}, nil
}

func (s *analyticsService) GetSystemAnalytics(ctx context.Context, period Period) (*models.SystemAnalytics, error) {
	// Implementation will be added
	return nil, fmt.Errorf("not implemented yet")
}

func (s *analyticsService) UpdateSystemAnalytics(ctx context.Context, date time.Time) error {
	// Implementation will be added
	return fmt.Errorf("not implemented yet")
}

func (s *analyticsService) GetSystemInsights(ctx context.Context, filters InsightFilters) (*SystemInsights, error) {
	// Get system analytics data
	analytics, err := s.getSystemAnalyticsData(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get system analytics: %w", err)
	}

	// Get user statistics
	userStats, err := s.getSystemUserStats(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}

	// Get repository statistics
	repoStats, err := s.getSystemRepositoryStats(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository stats: %w", err)
	}

	// Get performance statistics
	perfStats, err := s.getSystemPerformanceStats(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get performance stats: %w", err)
	}

	// Get resource statistics
	resourceStats, err := s.getSystemResourceStats(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource stats: %w", err)
	}

	return &SystemInsights{
		Analytics:        analytics,
		UserStats:        userStats,
		RepositoryStats:  repoStats,
		PerformanceStats: perfStats,
		ResourceStats:    resourceStats,
	}, nil
}

func (s *analyticsService) RecordPerformanceLog(ctx context.Context, log *models.PerformanceLog) error {
	if err := s.db.WithContext(ctx).Create(log).Error; err != nil {
		s.logger.WithError(err).Error("Failed to record performance log")
		return fmt.Errorf("failed to record performance log: %w", err)
	}

	return nil
}

func (s *analyticsService) GetPerformanceLogs(ctx context.Context, filters PerformanceFilters) ([]*models.PerformanceLog, int64, error) {
	query := s.db.WithContext(ctx).Model(&models.PerformanceLog{})

	// Apply filters
	if len(filters.Methods) > 0 {
		query = query.Where("method IN ?", filters.Methods)
	}
	if len(filters.Paths) > 0 {
		query = query.Where("path IN ?", filters.Paths)
	}
	if len(filters.StatusCodes) > 0 {
		query = query.Where("status_code IN ?", filters.StatusCodes)
	}
	if filters.UserID != nil {
		query = query.Where("user_id = ?", *filters.UserID)
	}
	if filters.RepositoryID != nil {
		query = query.Where("repository_id = ?", *filters.RepositoryID)
	}
	if filters.OrganizationID != nil {
		query = query.Where("organization_id = ?", *filters.OrganizationID)
	}
	if filters.MinDuration != nil {
		query = query.Where("duration >= ?", *filters.MinDuration)
	}
	if filters.MaxDuration != nil {
		query = query.Where("duration <= ?", *filters.MaxDuration)
	}
	if filters.StartDate != nil {
		query = query.Where("created_at >= ?", *filters.StartDate)
	}
	if filters.EndDate != nil {
		query = query.Where("created_at <= ?", *filters.EndDate)
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count performance logs: %w", err)
	}

	// Apply pagination
	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}
	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	// Get logs
	var logs []*models.PerformanceLog
	if err := query.Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get performance logs: %w", err)
	}

	return logs, total, nil
}

func (s *analyticsService) GetPerformanceMetrics(ctx context.Context, filters PerformanceFilters) (*PerformanceMetrics, error) {
	query := s.db.WithContext(ctx).Model(&models.PerformanceLog{})

	// Apply filters
	if len(filters.Methods) > 0 {
		query = query.Where("method IN ?", filters.Methods)
	}
	if len(filters.Paths) > 0 {
		query = query.Where("path IN ?", filters.Paths)
	}
	if len(filters.StatusCodes) > 0 {
		query = query.Where("status_code IN ?", filters.StatusCodes)
	}
	if filters.UserID != nil {
		query = query.Where("user_id = ?", *filters.UserID)
	}
	if filters.RepositoryID != nil {
		query = query.Where("repository_id = ?", *filters.RepositoryID)
	}
	if filters.OrganizationID != nil {
		query = query.Where("organization_id = ?", *filters.OrganizationID)
	}
	if filters.MinDuration != nil {
		query = query.Where("duration >= ?", *filters.MinDuration)
	}
	if filters.MaxDuration != nil {
		query = query.Where("duration <= ?", *filters.MaxDuration)
	}
	if filters.StartDate != nil {
		query = query.Where("created_at >= ?", *filters.StartDate)
	}
	if filters.EndDate != nil {
		query = query.Where("created_at <= ?", *filters.EndDate)
	}

	// Calculate performance metrics
	var metrics struct {
		AvgResponseTime  float64 `json:"avg_response_time"`
		P50ResponseTime  float64 `json:"p50_response_time"`
		P95ResponseTime  float64 `json:"p95_response_time"`
		P99ResponseTime  float64 `json:"p99_response_time"`
		TotalRequests    int64   `json:"total_requests"`
		ErrorRequests    int64   `json:"error_requests"`
		ThroughputPerMin float64 `json:"throughput_per_min"`
	}

	// Get basic metrics
	err := query.Select(`
		AVG(duration) as avg_response_time,
		PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY duration) as p50_response_time,
		PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY duration) as p95_response_time,
		PERCENTILE_CONT(0.99) WITHIN GROUP (ORDER BY duration) as p99_response_time,
		COUNT(*) as total_requests
	`).Scan(&metrics).Error

	if err != nil {
		return nil, fmt.Errorf("failed to calculate performance metrics: %w", err)
	}

	// Get error count
	query.Where("status_code >= 400").Count(&metrics.ErrorRequests)

	// Calculate error rate
	var errorRate *float64
	if metrics.TotalRequests > 0 {
		rate := float64(metrics.ErrorRequests) / float64(metrics.TotalRequests) * 100
		errorRate = &rate
	}

	// Calculate throughput per minute
	if filters.StartDate != nil && filters.EndDate != nil {
		duration := filters.EndDate.Sub(*filters.StartDate).Minutes()
		if duration > 0 {
			metrics.ThroughputPerMin = float64(metrics.TotalRequests) / duration
		}
	}

	// Get time series data
	responseTimeTrend, err := s.getResponseTimeTrend(ctx, filters)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get response time trend")
		responseTimeTrend = []TimeSeriesPoint{}
	}

	errorRateTrend, err := s.getErrorRateTrend(ctx, filters)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get error rate trend")
		errorRateTrend = []TimeSeriesPoint{}
	}

	return &PerformanceMetrics{
		AvgResponseTime:   &metrics.AvgResponseTime,
		P50ResponseTime:   &metrics.P50ResponseTime,
		P95ResponseTime:   &metrics.P95ResponseTime,
		P99ResponseTime:   &metrics.P99ResponseTime,
		TotalRequests:     metrics.TotalRequests,
		ErrorRate:         errorRate,
		ThroughputPerMin:  &metrics.ThroughputPerMin,
		ResponseTimeTrend: responseTimeTrend,
		ErrorRateTrend:    errorRateTrend,
	}, nil
}

func (s *analyticsService) AggregateMetrics(ctx context.Context, period Period) error {
	// Implementation will be added
	return fmt.Errorf("not implemented yet")
}

func (s *analyticsService) GenerateReport(ctx context.Context, reportType ReportType, filters ReportFilters) (*Report, error) {
	// Implementation will be added
	return nil, fmt.Errorf("not implemented yet")
}

func (s *analyticsService) ExportData(ctx context.Context, exportType ExportType, filters ExportFilters) ([]byte, error) {
	var data interface{}

	// Get data based on data type
	switch filters.DataType {
	case "events":
		eventFilters := EventFilters{
			StartDate:      filters.StartDate,
			EndDate:        filters.EndDate,
			RepositoryID:   filters.TargetID,
			OrganizationID: filters.TargetID,
		}
		events, _, eventErr := s.GetEvents(ctx, eventFilters)
		if eventErr != nil {
			return nil, fmt.Errorf("failed to get events: %w", eventErr)
		}
		data = events

	case "metrics":
		metricFilters := MetricFilters{
			StartDate:      filters.StartDate,
			EndDate:        filters.EndDate,
			RepositoryID:   filters.TargetID,
			OrganizationID: filters.TargetID,
		}
		metrics, metricErr := s.GetMetrics(ctx, metricFilters)
		if metricErr != nil {
			return nil, fmt.Errorf("failed to get metrics: %w", metricErr)
		}
		data = metrics

	case "performance":
		perfFilters := PerformanceFilters{
			StartDate:      filters.StartDate,
			EndDate:        filters.EndDate,
			RepositoryID:   filters.TargetID,
			OrganizationID: filters.TargetID,
			Limit:          1000, // Reasonable limit for export
		}
		logs, _, perfErr := s.GetPerformanceLogs(ctx, perfFilters)
		if perfErr != nil {
			return nil, fmt.Errorf("failed to get performance logs: %w", perfErr)
		}
		data = logs

	default:
		return nil, fmt.Errorf("unsupported data type: %s", filters.DataType)
	}

	// Export in the requested format
	switch exportType {
	case ExportTypeJSON:
		return json.Marshal(data)

	case ExportTypeCSV:
		return s.exportToCSV(data, filters.IncludeHeaders)

	case ExportTypeXLSX:
		return s.exportToXLSX(data, filters.IncludeHeaders)

	default:
		return nil, fmt.Errorf("unsupported export type: %s", exportType)
	}
}

// Helper methods for repository analytics

func (s *analyticsService) getRepositoryAnalyticsData(ctx context.Context, repoID uuid.UUID, filters InsightFilters) ([]*models.RepositoryAnalytics, error) {
	query := s.db.WithContext(ctx).Model(&models.RepositoryAnalytics{}).Where("repository_id = ?", repoID)

	if filters.StartDate != nil {
		query = query.Where("date >= ?", *filters.StartDate)
	}
	if filters.EndDate != nil {
		query = query.Where("date <= ?", *filters.EndDate)
	}

	var analytics []*models.RepositoryAnalytics
	if err := query.Order("date ASC").Find(&analytics).Error; err != nil {
		return nil, fmt.Errorf("failed to get repository analytics: %w", err)
	}

	return analytics, nil
}

func parseLanguageStats(raw string) (map[string]int64, error) {
	var result map[string]int64
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *analyticsService) getRepositoryCodeStats(ctx context.Context, repoID uuid.UUID) (*CodeStatistics, error) {
	// Get latest repository analytics for code stats
	var latest models.RepositoryAnalytics
	err := s.db.WithContext(ctx).Where("repository_id = ?", repoID).
		Order("date DESC").First(&latest).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to get repository analytics: %w", err)
	}

	// Get commits count from database
	var totalCommits int64
	if err := s.db.WithContext(ctx).Model(&models.Commit{}).Where("repository_id = ?", repoID).Count(&totalCommits).Error; err != nil {
		s.logger.WithError(err).Warn("Failed to count commits")
	}

	// Get branches count from database
	var totalBranches int64
	if err := s.db.WithContext(ctx).Model(&models.Branch{}).Where("repository_id = ?", repoID).Count(&totalBranches).Error; err != nil {
		s.logger.WithError(err).Warn("Failed to count branches")
	}

	// Get commit activity for the last 30 days
	commitActivity, err := s.getCommitActivity(ctx, repoID, 30)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get commit activity")
		commitActivity = []TimeSeriesPoint{}
	}

	// Parse language stats if available
	languageBreakdown := make(map[string]int64)
	if latest.LanguageStats != "" {
		// Parse JSON language stats
		parsed, err := parseLanguageStats(latest.LanguageStats)
		if err != nil {
			s.logger.WithError(err).Warn("Failed to parse language stats JSON")
		} else {
			languageBreakdown = parsed
		}
	}

	return &CodeStatistics{
		TotalLinesOfCode:  latest.LinesOfCode,
		TotalFiles:        latest.FileCount,
		TotalCommits:      totalCommits,
		TotalBranches:     totalBranches,
		LanguageBreakdown: languageBreakdown,
		CommitActivity:    commitActivity,
	}, nil
}

func (s *analyticsService) getRepositoryActivityStats(ctx context.Context, repoID uuid.UUID, filters InsightFilters) (*ActivityStatistics, error) {
	// Get latest repository analytics for activity stats
	var latest models.RepositoryAnalytics
	err := s.db.WithContext(ctx).Where("repository_id = ?", repoID).
		Order("date DESC").First(&latest).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to get repository analytics: %w", err)
	}

	// Get activity trend data
	activityTrend, err := s.getActivityTrend(ctx, repoID, filters)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get activity trend")
		activityTrend = []TimeSeriesPoint{}
	}

	return &ActivityStatistics{
		TotalViews:    latest.ViewsCount,
		TotalClones:   latest.ClonesCount,
		TotalForks:    latest.ForksCount,
		TotalStars:    latest.StarsCount,
		TotalWatchers: latest.WatchersCount,
		ActivityTrend: activityTrend,
	}, nil
}

func (s *analyticsService) getRepositoryContributorStats(ctx context.Context, repoID uuid.UUID, filters InsightFilters) (*ContributorStatistics, error) {
	// Get contributor data from commits
	type contributorData struct {
		UserID       uuid.UUID `json:"user_id"`
		Username     string    `json:"username"`
		CommitCount  int64     `json:"commit_count"`
		LinesAdded   int64     `json:"lines_added"`
		LinesDeleted int64     `json:"lines_deleted"`
	}

	query := `
		SELECT 
			u.id as user_id,
			u.username,
			COUNT(c.id) as commit_count,
			COALESCE(SUM(c.additions), 0) as lines_added,
			COALESCE(SUM(c.deletions), 0) as lines_deleted
		FROM commits c
		JOIN users u ON c.author_id = u.id
		WHERE c.repository_id = ?
	`

	if filters.StartDate != nil {
		query += " AND c.created_at >= ?"
	}
	if filters.EndDate != nil {
		query += " AND c.created_at <= ?"
	}

	query += " GROUP BY u.id, u.username ORDER BY commit_count DESC LIMIT 10"

	var args []interface{}
	args = append(args, repoID)
	if filters.StartDate != nil {
		args = append(args, *filters.StartDate)
	}
	if filters.EndDate != nil {
		args = append(args, *filters.EndDate)
	}

	var contributors []contributorData
	if err := s.db.WithContext(ctx).Raw(query, args...).Scan(&contributors).Error; err != nil {
		return nil, fmt.Errorf("failed to get contributors: %w", err)
	}

	// Convert to ContributorStat format
	var topContributors []ContributorStat
	for _, c := range contributors {
		topContributors = append(topContributors, ContributorStat{
			UserID:       c.UserID,
			Username:     c.Username,
			CommitCount:  c.CommitCount,
			LinesAdded:   c.LinesAdded,
			LinesDeleted: c.LinesDeleted,
		})
	}

	// Get total and active contributor counts
	var totalContributors int64
	var activeContributors int64

	s.db.WithContext(ctx).Model(&models.Commit{}).
		Where("repository_id = ?", repoID).
		Distinct("author_id").Count(&totalContributors)

	// Active contributors in the last 30 days
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	s.db.WithContext(ctx).Model(&models.Commit{}).
		Where("repository_id = ? AND created_at >= ?", repoID, thirtyDaysAgo).
		Distinct("author_id").Count(&activeContributors)

	contributorActivity, err := s.getContributorActivity(ctx, repoID, filters)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get contributor activity")
		contributorActivity = []TimeSeriesPoint{}
	}

	return &ContributorStatistics{
		TotalContributors:   totalContributors,
		ActiveContributors:  activeContributors,
		TopContributors:     topContributors,
		ContributorActivity: contributorActivity,
	}, nil
}

func (s *analyticsService) getRepositoryPRStats(ctx context.Context, repoID uuid.UUID, filters InsightFilters) (*PullRequestStatistics, error) {
	var totalPRs, openPRs, mergedPRs, closedPRs int64

	// Total pull requests
	s.db.WithContext(ctx).Model(&models.PullRequest{}).Where("repository_id = ?", repoID).Count(&totalPRs)

	// Open pull requests
	s.db.WithContext(ctx).Model(&models.PullRequest{}).Where("repository_id = ? AND state = ?", repoID, "open").Count(&openPRs)

	// Merged pull requests
	s.db.WithContext(ctx).Model(&models.PullRequest{}).Where("repository_id = ? AND state = ?", repoID, "merged").Count(&mergedPRs)

	// Closed pull requests (excluding merged)
	s.db.WithContext(ctx).Model(&models.PullRequest{}).Where("repository_id = ? AND state = ?", repoID, "closed").Count(&closedPRs)

	// Average time to merge (in hours)
	var avgTimeToMerge *float64
	var avgDuration float64
	err := s.db.WithContext(ctx).Model(&models.PullRequest{}).
		Select("AVG(EXTRACT(EPOCH FROM (merged_at - created_at))/3600) as avg_duration").
		Where("repository_id = ? AND merged_at IS NOT NULL", repoID).
		Scan(&avgDuration).Error
	if err == nil && avgDuration > 0 {
		avgTimeToMerge = &avgDuration
	}

	// Get PR activity trend
	prActivity, err := s.getPRActivity(ctx, repoID, filters)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get PR activity")
		prActivity = []TimeSeriesPoint{}
	}

	return &PullRequestStatistics{
		TotalPullRequests:  totalPRs,
		OpenPullRequests:   openPRs,
		MergedPullRequests: mergedPRs,
		ClosedPullRequests: closedPRs,
		AvgTimeToMerge:     avgTimeToMerge,
		PRActivity:         prActivity,
	}, nil
}

func (s *analyticsService) getRepositoryPerformanceStats(ctx context.Context, repoID uuid.UUID, filters InsightFilters) (*PerformanceStatistics, error) {
	var avgResponseTime, p95ResponseTime, errorRate *float64
	var avgResp, p95Resp, errRate float64

	// Get performance metrics from performance logs
	query := s.db.WithContext(ctx).Model(&models.PerformanceLog{}).Where("repository_id = ?", repoID)

	if filters.StartDate != nil {
		query = query.Where("created_at >= ?", *filters.StartDate)
	}
	if filters.EndDate != nil {
		query = query.Where("created_at <= ?", *filters.EndDate)
	}

	// Average response time
	err := query.Select("AVG(duration)").Scan(&avgResp).Error
	if err == nil && avgResp > 0 {
		avgResponseTime = &avgResp
	}

	// 95th percentile response time
	err = query.Select("PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY duration)").Scan(&p95Resp).Error
	if err == nil && p95Resp > 0 {
		p95ResponseTime = &p95Resp
	}

	// Error rate
	var totalRequests, errorRequests int64
	query.Count(&totalRequests)
	query.Where("status_code >= 400").Count(&errorRequests)

	if totalRequests > 0 {
		errRate = float64(errorRequests) / float64(totalRequests) * 100
		errorRate = &errRate
	}

	// Get throughput trend
	throughputTrend, err := s.getThroughputTrend(ctx, repoID, filters)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get throughput trend")
		throughputTrend = []TimeSeriesPoint{}
	}

	return &PerformanceStatistics{
		AvgResponseTime: avgResponseTime,
		P95ResponseTime: p95ResponseTime,
		ErrorRate:       errorRate,
		ThroughputTrend: throughputTrend,
	}, nil
}

// Helper methods for time series data

func (s *analyticsService) getCommitActivity(ctx context.Context, repoID uuid.UUID, days int) ([]TimeSeriesPoint, error) {
	since := time.Now().AddDate(0, 0, -days)

	var results []struct {
		Date  time.Time `json:"date"`
		Count int64     `json:"count"`
	}

	err := s.db.WithContext(ctx).Model(&models.Commit{}).
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("repository_id = ? AND created_at >= ?", repoID, since).
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	var activity []TimeSeriesPoint
	for _, r := range results {
		activity = append(activity, TimeSeriesPoint{
			Timestamp: r.Date,
			Value:     float64(r.Count),
		})
	}

	return activity, nil
}

func (s *analyticsService) getActivityTrend(ctx context.Context, repoID uuid.UUID, filters InsightFilters) ([]TimeSeriesPoint, error) {
	// Get analytics events for repository views/clones
	since := time.Now().AddDate(0, 0, -30)
	if filters.StartDate != nil {
		since = *filters.StartDate
	}

	var results []struct {
		Date  time.Time `json:"date"`
		Count int64     `json:"count"`
	}

	err := s.db.WithContext(ctx).Model(&models.AnalyticsEvent{}).
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("repository_id = ? AND created_at >= ? AND event_type IN (?)",
			repoID, since, []string{"repository.clone", "repository.view"}).
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	var trend []TimeSeriesPoint
	for _, r := range results {
		trend = append(trend, TimeSeriesPoint{
			Timestamp: r.Date,
			Value:     float64(r.Count),
		})
	}

	return trend, nil
}

func (s *analyticsService) getContributorActivity(ctx context.Context, repoID uuid.UUID, filters InsightFilters) ([]TimeSeriesPoint, error) {
	since := time.Now().AddDate(0, 0, -30)
	if filters.StartDate != nil {
		since = *filters.StartDate
	}

	var results []struct {
		Date  time.Time `json:"date"`
		Count int64     `json:"count"`
	}

	err := s.db.WithContext(ctx).Model(&models.Commit{}).
		Select("DATE(created_at) as date, COUNT(DISTINCT author_id) as count").
		Where("repository_id = ? AND created_at >= ?", repoID, since).
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	var activity []TimeSeriesPoint
	for _, r := range results {
		activity = append(activity, TimeSeriesPoint{
			Timestamp: r.Date,
			Value:     float64(r.Count),
		})
	}

	return activity, nil
}

func (s *analyticsService) getPRActivity(ctx context.Context, repoID uuid.UUID, filters InsightFilters) ([]TimeSeriesPoint, error) {
	since := time.Now().AddDate(0, 0, -30)
	if filters.StartDate != nil {
		since = *filters.StartDate
	}

	var results []struct {
		Date  time.Time `json:"date"`
		Count int64     `json:"count"`
	}

	err := s.db.WithContext(ctx).Model(&models.PullRequest{}).
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("repository_id = ? AND created_at >= ?", repoID, since).
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	var activity []TimeSeriesPoint
	for _, r := range results {
		activity = append(activity, TimeSeriesPoint{
			Timestamp: r.Date,
			Value:     float64(r.Count),
		})
	}

	return activity, nil
}

func (s *analyticsService) getThroughputTrend(ctx context.Context, repoID uuid.UUID, filters InsightFilters) ([]TimeSeriesPoint, error) {
	since := time.Now().AddDate(0, 0, -30)
	if filters.StartDate != nil {
		since = *filters.StartDate
	}

	var results []struct {
		Date  time.Time `json:"date"`
		Count int64     `json:"count"`
	}

	err := s.db.WithContext(ctx).Model(&models.PerformanceLog{}).
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("repository_id = ? AND created_at >= ?", repoID, since).
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	var trend []TimeSeriesPoint
	for _, r := range results {
		trend = append(trend, TimeSeriesPoint{
			Timestamp: r.Date,
			Value:     float64(r.Count),
		})
	}

	return trend, nil
}

// Helper methods for user analytics

func (s *analyticsService) getUserAnalyticsData(ctx context.Context, userID uuid.UUID, filters InsightFilters) ([]*models.UserAnalytics, error) {
	query := s.db.WithContext(ctx).Model(&models.UserAnalytics{}).Where("user_id = ?", userID)

	if filters.StartDate != nil {
		query = query.Where("date >= ?", *filters.StartDate)
	}
	if filters.EndDate != nil {
		query = query.Where("date <= ?", *filters.EndDate)
	}

	var analytics []*models.UserAnalytics
	if err := query.Order("date ASC").Find(&analytics).Error; err != nil {
		return nil, fmt.Errorf("failed to get user analytics: %w", err)
	}

	return analytics, nil
}

func (s *analyticsService) getUserActivityStats(ctx context.Context, userID uuid.UUID, filters InsightFilters) (*UserActivityStats, error) {
	// Get latest user analytics for activity stats
	var latest models.UserAnalytics
	err := s.db.WithContext(ctx).Where("user_id = ?", userID).
		Order("date DESC").First(&latest).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to get user analytics: %w", err)
	}

	// Get activity trend data
	activityTrend, err := s.getUserActivityTrend(ctx, userID, filters)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get user activity trend")
		activityTrend = []TimeSeriesPoint{}
	}

	return &UserActivityStats{
		TotalLogins:    latest.LoginCount,
		TotalSessions:  latest.LoginCount, // Assuming 1 session per login for simplicity
		AvgSessionTime: latest.SessionDuration,
		TotalPageViews: latest.PageViews,
		ActivityTrend:  activityTrend,
	}, nil
}

func (s *analyticsService) getUserContributionStats(ctx context.Context, userID uuid.UUID, filters InsightFilters) (*UserContributionStats, error) {
	// Get contribution data from various sources
	var totalCommits, totalPullRequests, totalComments int64

	// Count commits by user
	commitQuery := s.db.WithContext(ctx).Model(&models.Commit{}).Where("author_id = ?", userID)
	if filters.StartDate != nil {
		commitQuery = commitQuery.Where("created_at >= ?", *filters.StartDate)
	}
	if filters.EndDate != nil {
		commitQuery = commitQuery.Where("created_at <= ?", *filters.EndDate)
	}
	commitQuery.Count(&totalCommits)

	// Count pull requests by user
	prQuery := s.db.WithContext(ctx).Model(&models.PullRequest{}).Where("user_id = ?", userID)
	if filters.StartDate != nil {
		prQuery = prQuery.Where("created_at >= ?", *filters.StartDate)
	}
	if filters.EndDate != nil {
		prQuery = prQuery.Where("created_at <= ?", *filters.EndDate)
	}
	prQuery.Count(&totalPullRequests)

	// Count comments by user (assuming there's a comments table)
	// For now, using a placeholder
	totalComments = totalPullRequests // Rough estimate

	// Get contribution trend
	contributionTrend, err := s.getUserContributionTrend(ctx, userID, filters)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get user contribution trend")
		contributionTrend = []TimeSeriesPoint{}
	}

	return &UserContributionStats{
		TotalCommits:      totalCommits,
		TotalPullRequests: totalPullRequests,
		TotalComments:     totalComments,
		ContributionTrend: contributionTrend,
	}, nil
}

func (s *analyticsService) getUserRepositoryStats(ctx context.Context, userID uuid.UUID, filters InsightFilters) (*UserRepositoryStats, error) {
	// Count repositories owned by user
	var totalRepositories int64
	s.db.WithContext(ctx).Model(&models.Repository{}).
		Where("owner_id = ? AND owner_type = ?", userID, "user").
		Count(&totalRepositories)

	// Get total stars and forks for user's repositories
	var totalStars, totalForks int64

	// Sum stars and forks from repository analytics
	var starsForks struct {
		TotalStars int64 `json:"total_stars"`
		TotalForks int64 `json:"total_forks"`
	}

	err := s.db.WithContext(ctx).Model(&models.RepositoryAnalytics{}).
		Select("COALESCE(SUM(stars_count), 0) as total_stars, COALESCE(SUM(forks_count), 0) as total_forks").
		Joins("JOIN repositories r ON repository_analytics.repository_id = r.id").
		Where("r.owner_id = ? AND r.owner_type = ?", userID, "user").
		Scan(&starsForks).Error

	if err == nil {
		totalStars = starsForks.TotalStars
		totalForks = starsForks.TotalForks
	}

	// Get repository trend
	repositoryTrend, err := s.getUserRepositoryTrend(ctx, userID, filters)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get user repository trend")
		repositoryTrend = []TimeSeriesPoint{}
	}

	return &UserRepositoryStats{
		TotalRepositories: totalRepositories,
		TotalStars:        totalStars,
		TotalForks:        totalForks,
		RepositoryTrend:   repositoryTrend,
	}, nil
}

// Helper methods for user time series data

func (s *analyticsService) getUserActivityTrend(ctx context.Context, userID uuid.UUID, filters InsightFilters) ([]TimeSeriesPoint, error) {
	since := time.Now().AddDate(0, 0, -30)
	if filters.StartDate != nil {
		since = *filters.StartDate
	}

	var results []struct {
		Date  time.Time `json:"date"`
		Count int64     `json:"count"`
	}

	err := s.db.WithContext(ctx).Model(&models.AnalyticsEvent{}).
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("actor_id = ? AND created_at >= ? AND event_type IN (?)",
			userID, since, []string{"user.login", "page.view"}).
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	var trend []TimeSeriesPoint
	for _, r := range results {
		trend = append(trend, TimeSeriesPoint{
			Timestamp: r.Date,
			Value:     float64(r.Count),
		})
	}

	return trend, nil
}

func (s *analyticsService) getUserContributionTrend(ctx context.Context, userID uuid.UUID, filters InsightFilters) ([]TimeSeriesPoint, error) {
	since := time.Now().AddDate(0, 0, -30)
	if filters.StartDate != nil {
		since = *filters.StartDate
	}

	var results []struct {
		Date  time.Time `json:"date"`
		Count int64     `json:"count"`
	}

	err := s.db.WithContext(ctx).Model(&models.Commit{}).
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("author_id = ? AND created_at >= ?", userID, since).
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	var trend []TimeSeriesPoint
	for _, r := range results {
		trend = append(trend, TimeSeriesPoint{
			Timestamp: r.Date,
			Value:     float64(r.Count),
		})
	}

	return trend, nil
}

func (s *analyticsService) getUserRepositoryTrend(ctx context.Context, userID uuid.UUID, filters InsightFilters) ([]TimeSeriesPoint, error) {
	since := time.Now().AddDate(0, 0, -30)
	if filters.StartDate != nil {
		since = *filters.StartDate
	}

	var results []struct {
		Date  time.Time `json:"date"`
		Count int64     `json:"count"`
	}

	err := s.db.WithContext(ctx).Model(&models.Repository{}).
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("owner_id = ? AND owner_type = ? AND created_at >= ?", userID, "user", since).
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	var trend []TimeSeriesPoint
	for _, r := range results {
		trend = append(trend, TimeSeriesPoint{
			Timestamp: r.Date,
			Value:     float64(r.Count),
		})
	}

	return trend, nil
}

// Helper methods for system analytics

func (s *analyticsService) getSystemAnalyticsData(ctx context.Context, filters InsightFilters) ([]*models.SystemAnalytics, error) {
	query := s.db.WithContext(ctx).Model(&models.SystemAnalytics{})

	if filters.StartDate != nil {
		query = query.Where("date >= ?", *filters.StartDate)
	}
	if filters.EndDate != nil {
		query = query.Where("date <= ?", *filters.EndDate)
	}

	var analytics []*models.SystemAnalytics
	if err := query.Order("date ASC").Find(&analytics).Error; err != nil {
		return nil, fmt.Errorf("failed to get system analytics: %w", err)
	}

	return analytics, nil
}

func (s *analyticsService) getSystemUserStats(ctx context.Context, filters InsightFilters) (*SystemUserStats, error) {
	var totalUsers, activeUsers, newRegistrations int64

	// Count total users
	s.db.WithContext(ctx).Model(&models.User{}).Count(&totalUsers)

	// Count active users in the last 30 days
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	s.db.WithContext(ctx).Model(&models.AnalyticsEvent{}).
		Where("event_type = ? AND created_at >= ?", "user.login", thirtyDaysAgo).
		Distinct("actor_id").Count(&activeUsers)

	// Count new registrations in the filter period
	regQuery := s.db.WithContext(ctx).Model(&models.User{})
	if filters.StartDate != nil {
		regQuery = regQuery.Where("created_at >= ?", *filters.StartDate)
	} else {
		regQuery = regQuery.Where("created_at >= ?", thirtyDaysAgo)
	}
	if filters.EndDate != nil {
		regQuery = regQuery.Where("created_at <= ?", *filters.EndDate)
	}
	regQuery.Count(&newRegistrations)

	// Get user trend
	userTrend, err := s.getSystemUserTrend(ctx, filters)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get user trend")
		userTrend = []TimeSeriesPoint{}
	}

	// Calculate churn rate (placeholder)
	var churnRate *float64
	if totalUsers > 0 {
		rate := float64(newRegistrations) / float64(totalUsers) * 5 // Rough estimate
		churnRate = &rate
	}

	return &SystemUserStats{
		TotalUsers:       totalUsers,
		ActiveUsers:      activeUsers,
		NewRegistrations: newRegistrations,
		ChurnRate:        churnRate,
		UserTrend:        userTrend,
	}, nil
}

func (s *analyticsService) getSystemRepositoryStats(ctx context.Context, filters InsightFilters) (*SystemRepositoryStats, error) {
	var totalRepos, publicRepos, privateRepos, totalOrgs int64

	// Count repositories
	s.db.WithContext(ctx).Model(&models.Repository{}).Count(&totalRepos)
	s.db.WithContext(ctx).Model(&models.Repository{}).Where("visibility = ?", "public").Count(&publicRepos)
	s.db.WithContext(ctx).Model(&models.Repository{}).Where("visibility = ?", "private").Count(&privateRepos)

	// Count organizations
	s.db.WithContext(ctx).Model(&models.Organization{}).Count(&totalOrgs)

	// Get repository trend
	repoTrend, err := s.getSystemRepositoryTrend(ctx, filters)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get repository trend")
		repoTrend = []TimeSeriesPoint{}
	}

	return &SystemRepositoryStats{
		TotalRepositories:   totalRepos,
		PublicRepositories:  publicRepos,
		PrivateRepositories: privateRepos,
		TotalOrganizations:  totalOrgs,
		RepositoryTrend:     repoTrend,
	}, nil
}

func (s *analyticsService) getSystemPerformanceStats(ctx context.Context, filters InsightFilters) (*SystemPerformanceStats, error) {
	// Get performance metrics from performance logs
	query := s.db.WithContext(ctx).Model(&models.PerformanceLog{})

	if filters.StartDate != nil {
		query = query.Where("created_at >= ?", *filters.StartDate)
	}
	if filters.EndDate != nil {
		query = query.Where("created_at <= ?", *filters.EndDate)
	}

	var avgResponseTime, p95ResponseTime, errorRate, uptime *float64
	var avgResp, p95Resp, errRate float64

	// Average response time
	err := query.Select("AVG(duration)").Scan(&avgResp).Error
	if err == nil && avgResp > 0 {
		avgResponseTime = &avgResp
	}

	// 95th percentile response time
	err = query.Select("PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY duration)").Scan(&p95Resp).Error
	if err == nil && p95Resp > 0 {
		p95ResponseTime = &p95Resp
	}

	// Error rate
	var totalRequests, errorRequests int64
	query.Count(&totalRequests)
	query.Where("status_code >= 400").Count(&errorRequests)

	if totalRequests > 0 {
		errRate = float64(errorRequests) / float64(totalRequests) * 100
		errorRate = &errRate
	}

	// Uptime (placeholder - 99.9%)
	uptimeVal := 99.9
	uptime = &uptimeVal

	// Get performance trend
	perfTrend, err := s.getSystemPerformanceTrend(ctx, filters)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get performance trend")
		perfTrend = []TimeSeriesPoint{}
	}

	return &SystemPerformanceStats{
		AvgResponseTime:  avgResponseTime,
		P95ResponseTime:  p95ResponseTime,
		ErrorRate:        errorRate,
		Uptime:           uptime,
		PerformanceTrend: perfTrend,
	}, nil
}

func (s *analyticsService) getSystemResourceStats(ctx context.Context, filters InsightFilters) (*SystemResourceStats, error) {
	// Placeholder resource stats (would integrate with actual monitoring)
	var cpuUsage, memoryUsage, diskUsage float64 = 45.2, 67.8, 23.4
	var networkInMB, networkOutMB int64 = 1024, 2048

	// Get resource trend
	resourceTrend, err := s.getSystemResourceTrend(ctx, filters)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get resource trend")
		resourceTrend = []TimeSeriesPoint{}
	}

	return &SystemResourceStats{
		CPUUsage:      &cpuUsage,
		MemoryUsage:   &memoryUsage,
		DiskUsage:     &diskUsage,
		NetworkInMB:   networkInMB,
		NetworkOutMB:  networkOutMB,
		ResourceTrend: resourceTrend,
	}, nil
}

// Helper methods for system time series data

func (s *analyticsService) getSystemUserTrend(ctx context.Context, filters InsightFilters) ([]TimeSeriesPoint, error) {
	since := time.Now().AddDate(0, 0, -30)
	if filters.StartDate != nil {
		since = *filters.StartDate
	}

	var results []struct {
		Date  time.Time `json:"date"`
		Count int64     `json:"count"`
	}

	err := s.db.WithContext(ctx).Model(&models.User{}).
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("created_at >= ?", since).
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	var trend []TimeSeriesPoint
	for _, r := range results {
		trend = append(trend, TimeSeriesPoint{
			Timestamp: r.Date,
			Value:     float64(r.Count),
		})
	}

	return trend, nil
}

func (s *analyticsService) getSystemRepositoryTrend(ctx context.Context, filters InsightFilters) ([]TimeSeriesPoint, error) {
	since := time.Now().AddDate(0, 0, -30)
	if filters.StartDate != nil {
		since = *filters.StartDate
	}

	var results []struct {
		Date  time.Time `json:"date"`
		Count int64     `json:"count"`
	}

	err := s.db.WithContext(ctx).Model(&models.Repository{}).
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("created_at >= ?", since).
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	var trend []TimeSeriesPoint
	for _, r := range results {
		trend = append(trend, TimeSeriesPoint{
			Timestamp: r.Date,
			Value:     float64(r.Count),
		})
	}

	return trend, nil
}

func (s *analyticsService) getSystemPerformanceTrend(ctx context.Context, filters InsightFilters) ([]TimeSeriesPoint, error) {
	since := time.Now().AddDate(0, 0, -30)
	if filters.StartDate != nil {
		since = *filters.StartDate
	}

	var results []struct {
		Date time.Time `json:"date"`
		Avg  float64   `json:"avg"`
	}

	err := s.db.WithContext(ctx).Model(&models.PerformanceLog{}).
		Select("DATE(created_at) as date, AVG(duration) as avg").
		Where("created_at >= ?", since).
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	var trend []TimeSeriesPoint
	for _, r := range results {
		trend = append(trend, TimeSeriesPoint{
			Timestamp: r.Date,
			Value:     r.Avg,
		})
	}

	return trend, nil
}

func (s *analyticsService) getSystemResourceTrend(ctx context.Context, filters InsightFilters) ([]TimeSeriesPoint, error) {
	// Placeholder implementation - would integrate with actual monitoring
	since := time.Now().AddDate(0, 0, -30)
	if filters.StartDate != nil {
		since = *filters.StartDate
	}

	var trend []TimeSeriesPoint
	// Generate some sample data points
	for d := since; d.Before(time.Now()); d = d.AddDate(0, 0, 1) {
		trend = append(trend, TimeSeriesPoint{
			Timestamp: d,
			Value:     45.0 + float64(len(trend)%10), // Sample CPU usage
		})
	}

	return trend, nil
}

// Export helper functions

func (s *analyticsService) exportToCSV(data interface{}, includeHeaders bool) ([]byte, error) {
	// Simple CSV export implementation
	// In a real implementation, you would use a proper CSV library
	return []byte("CSV export not fully implemented"), nil
}

func (s *analyticsService) exportToXLSX(data interface{}, includeHeaders bool) ([]byte, error) {
	// Simple XLSX export implementation
	// In a real implementation, you would use a library like excelize
	return []byte("XLSX export not fully implemented"), nil
}

// Organization analytics helper functions

func (s *analyticsService) getOrganizationAnalyticsData(ctx context.Context, orgID uuid.UUID, filters InsightFilters) ([]*models.OrganizationAnalytics, error) {
	query := s.db.WithContext(ctx).Model(&models.OrganizationAnalytics{}).Where("organization_id = ?", orgID)

	if filters.StartDate != nil {
		query = query.Where("date >= ?", *filters.StartDate)
	}
	if filters.EndDate != nil {
		query = query.Where("date <= ?", *filters.EndDate)
	}

	var analytics []*models.OrganizationAnalytics
	if err := query.Order("date ASC").Find(&analytics).Error; err != nil {
		return nil, fmt.Errorf("failed to get organization analytics: %w", err)
	}

	return analytics, nil
}

func (s *analyticsService) getOrganizationMemberStats(ctx context.Context, orgID uuid.UUID, filters InsightFilters) (*OrganizationMemberStats, error) {
	var totalMembers, activeMembers, totalTeams int64

	// Get total members
	s.db.WithContext(ctx).Model(&models.OrganizationMember{}).Where("organization_id = ?", orgID).Count(&totalMembers)

	// Get active members (those with activity in the last 30 days)
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	s.db.WithContext(ctx).Model(&models.AnalyticsEvent{}).
		Where("organization_id = ? AND created_at >= ? AND actor_type = ?", orgID, thirtyDaysAgo, "user").
		Distinct("actor_id").Count(&activeMembers)

	// Get total teams
	s.db.WithContext(ctx).Model(&models.Team{}).Where("organization_id = ?", orgID).Count(&totalTeams)

	// Get member trend
	memberTrend, err := s.getOrganizationMemberTrend(ctx, orgID, filters)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get member trend")
		memberTrend = []TimeSeriesPoint{}
	}

	return &OrganizationMemberStats{
		TotalMembers:  totalMembers,
		ActiveMembers: activeMembers,
		TotalTeams:    totalTeams,
		MemberTrend:   memberTrend,
	}, nil
}

func (s *analyticsService) getOrganizationRepositoryStats(ctx context.Context, orgID uuid.UUID, filters InsightFilters) (*OrganizationRepositoryStats, error) {
	var totalRepos, publicRepos, privateRepos int64

	// Count repositories
	s.db.WithContext(ctx).Model(&models.Repository{}).Where("owner_id = ? AND owner_type = ?", orgID, "organization").Count(&totalRepos)
	s.db.WithContext(ctx).Model(&models.Repository{}).Where("owner_id = ? AND owner_type = ? AND visibility = ?", orgID, "organization", "public").Count(&publicRepos)
	s.db.WithContext(ctx).Model(&models.Repository{}).Where("owner_id = ? AND owner_type = ? AND visibility = ?", orgID, "organization", "private").Count(&privateRepos)

	// Get repository trend
	repoTrend, err := s.getOrganizationRepositoryTrend(ctx, orgID, filters)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get repository trend")
		repoTrend = []TimeSeriesPoint{}
	}

	return &OrganizationRepositoryStats{
		TotalRepositories:   totalRepos,
		PublicRepositories:  publicRepos,
		PrivateRepositories: privateRepos,
		RepositoryTrend:     repoTrend,
	}, nil
}

func (s *analyticsService) getOrganizationActivityStats(ctx context.Context, orgID uuid.UUID, filters InsightFilters) (*OrganizationActivityStats, error) {
	var totalCommits, totalPRs int64

	// Count activity across organization repositories
	s.db.WithContext(ctx).Raw(`
		SELECT COUNT(c.id) FROM commits c 
		JOIN repositories r ON c.repository_id = r.id 
		WHERE r.owner_id = ? AND r.owner_type = ?
	`, orgID, "organization").Scan(&totalCommits)

	s.db.WithContext(ctx).Raw(`
		SELECT COUNT(pr.id) FROM pull_requests pr 
		JOIN repositories r ON pr.repository_id = r.id 
		WHERE r.owner_id = ? AND r.owner_type = ?
	`, orgID, "organization").Scan(&totalPRs)

	// Get activity trend
	activityTrend, err := s.getOrganizationActivityTrend(ctx, orgID, filters)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get activity trend")
		activityTrend = []TimeSeriesPoint{}
	}

	return &OrganizationActivityStats{
		TotalCommits:      totalCommits,
		TotalPullRequests: totalPRs,
		ActivityTrend:     activityTrend,
	}, nil
}

func (s *analyticsService) getOrganizationResourceStats(ctx context.Context, orgID uuid.UUID, filters InsightFilters) (*OrganizationResourceStats, error) {
	// Get latest resource usage from organization analytics
	var latest models.OrganizationAnalytics
	err := s.db.WithContext(ctx).Where("organization_id = ?", orgID).
		Order("date DESC").First(&latest).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to get organization analytics: %w", err)
	}

	// Get resource trend
	resourceTrend, err := s.getOrganizationResourceTrend(ctx, orgID, filters)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get resource trend")
		resourceTrend = []TimeSeriesPoint{}
	}

	// Calculate estimated cost (simplified)
	var estimatedCost *float64
	cost := float64(latest.StorageUsedMB)*0.0001 + float64(latest.BandwidthUsedMB)*0.00005 + float64(latest.ComputeTimeMinutes)*0.01
	estimatedCost = &cost

	return &OrganizationResourceStats{
		StorageUsedMB:      latest.StorageUsedMB,
		BandwidthUsedMB:    latest.BandwidthUsedMB,
		ComputeTimeMinutes: latest.ComputeTimeMinutes,
		EstimatedCost:      estimatedCost,
		ResourceTrend:      resourceTrend,
	}, nil
}

// Organization trend helper functions

func (s *analyticsService) getOrganizationMemberTrend(ctx context.Context, orgID uuid.UUID, filters InsightFilters) ([]TimeSeriesPoint, error) {
	since := time.Now().AddDate(0, 0, -30)
	if filters.StartDate != nil {
		since = *filters.StartDate
	}

	var results []struct {
		Date  time.Time `json:"date"`
		Count int64     `json:"count"`
	}

	err := s.db.WithContext(ctx).Model(&models.OrganizationMember{}).
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("organization_id = ? AND created_at >= ?", orgID, since).
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	var trend []TimeSeriesPoint
	for _, r := range results {
		trend = append(trend, TimeSeriesPoint{
			Timestamp: r.Date,
			Value:     float64(r.Count),
		})
	}

	return trend, nil
}

func (s *analyticsService) getOrganizationRepositoryTrend(ctx context.Context, orgID uuid.UUID, filters InsightFilters) ([]TimeSeriesPoint, error) {
	since := time.Now().AddDate(0, 0, -30)
	if filters.StartDate != nil {
		since = *filters.StartDate
	}

	var results []struct {
		Date  time.Time `json:"date"`
		Count int64     `json:"count"`
	}

	err := s.db.WithContext(ctx).Model(&models.Repository{}).
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("owner_id = ? AND owner_type = ? AND created_at >= ?", orgID, "organization", since).
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	var trend []TimeSeriesPoint
	for _, r := range results {
		trend = append(trend, TimeSeriesPoint{
			Timestamp: r.Date,
			Value:     float64(r.Count),
		})
	}

	return trend, nil
}

func (s *analyticsService) getOrganizationActivityTrend(ctx context.Context, orgID uuid.UUID, filters InsightFilters) ([]TimeSeriesPoint, error) {
	since := time.Now().AddDate(0, 0, -30)
	if filters.StartDate != nil {
		since = *filters.StartDate
	}

	var results []struct {
		Date  time.Time `json:"date"`
		Count int64     `json:"count"`
	}

	err := s.db.WithContext(ctx).Model(&models.AnalyticsEvent{}).
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("organization_id = ? AND created_at >= ?", orgID, since).
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	var trend []TimeSeriesPoint
	for _, r := range results {
		trend = append(trend, TimeSeriesPoint{
			Timestamp: r.Date,
			Value:     float64(r.Count),
		})
	}

	return trend, nil
}

func (s *analyticsService) getOrganizationResourceTrend(ctx context.Context, orgID uuid.UUID, filters InsightFilters) ([]TimeSeriesPoint, error) {
	since := time.Now().AddDate(0, 0, -30)
	if filters.StartDate != nil {
		since = *filters.StartDate
	}

	var results []struct {
		Date    time.Time `json:"date"`
		Storage int64     `json:"storage"`
	}

	err := s.db.WithContext(ctx).Model(&models.OrganizationAnalytics{}).
		Select("date, storage_used_mb as storage").
		Where("organization_id = ? AND date >= ?", orgID, since).
		Order("date ASC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	var trend []TimeSeriesPoint
	for _, r := range results {
		trend = append(trend, TimeSeriesPoint{
			Timestamp: r.Date,
			Value:     float64(r.Storage),
		})
	}

	return trend, nil
}

// Performance trend helper functions

func (s *analyticsService) getResponseTimeTrend(ctx context.Context, filters PerformanceFilters) ([]TimeSeriesPoint, error) {
	since := time.Now().AddDate(0, 0, -7) // Last 7 days by default
	if filters.StartDate != nil {
		since = *filters.StartDate
	}

	query := s.db.WithContext(ctx).Model(&models.PerformanceLog{}).
		Select("DATE(created_at) as date, AVG(duration) as avg_duration").
		Where("created_at >= ?", since)

	// Apply additional filters
	if len(filters.Methods) > 0 {
		query = query.Where("method IN ?", filters.Methods)
	}
	if len(filters.Paths) > 0 {
		query = query.Where("path IN ?", filters.Paths)
	}
	if filters.RepositoryID != nil {
		query = query.Where("repository_id = ?", *filters.RepositoryID)
	}
	if filters.OrganizationID != nil {
		query = query.Where("organization_id = ?", *filters.OrganizationID)
	}

	var results []struct {
		Date        time.Time `json:"date"`
		AvgDuration float64   `json:"avg_duration"`
	}

	err := query.Group("DATE(created_at)").Order("date ASC").Scan(&results).Error
	if err != nil {
		return nil, err
	}

	var trend []TimeSeriesPoint
	for _, r := range results {
		trend = append(trend, TimeSeriesPoint{
			Timestamp: r.Date,
			Value:     r.AvgDuration,
		})
	}

	return trend, nil
}

func (s *analyticsService) getErrorRateTrend(ctx context.Context, filters PerformanceFilters) ([]TimeSeriesPoint, error) {
	since := time.Now().AddDate(0, 0, -7) // Last 7 days by default
	if filters.StartDate != nil {
		since = *filters.StartDate
	}

	query := s.db.WithContext(ctx).Model(&models.PerformanceLog{}).
		Select(`
			DATE(created_at) as date, 
			COUNT(*) as total_requests,
			COUNT(CASE WHEN status_code >= 400 THEN 1 END) as error_requests
		`).
		Where("created_at >= ?", since)

	// Apply additional filters
	if len(filters.Methods) > 0 {
		query = query.Where("method IN ?", filters.Methods)
	}
	if len(filters.Paths) > 0 {
		query = query.Where("path IN ?", filters.Paths)
	}
	if filters.RepositoryID != nil {
		query = query.Where("repository_id = ?", *filters.RepositoryID)
	}
	if filters.OrganizationID != nil {
		query = query.Where("organization_id = ?", *filters.OrganizationID)
	}

	var results []struct {
		Date          time.Time `json:"date"`
		TotalRequests int64     `json:"total_requests"`
		ErrorRequests int64     `json:"error_requests"`
	}

	err := query.Group("DATE(created_at)").Order("date ASC").Scan(&results).Error
	if err != nil {
		return nil, err
	}

	var trend []TimeSeriesPoint
	for _, r := range results {
		errorRate := float64(0)
		if r.TotalRequests > 0 {
			errorRate = float64(r.ErrorRequests) / float64(r.TotalRequests) * 100
		}
		trend = append(trend, TimeSeriesPoint{
			Timestamp: r.Date,
			Value:     errorRate,
		})
	}

	return trend, nil
}
