package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EventType represents different types of analytics events
type EventType string

const (
	// Repository Events
	EventRepositoryCreated     EventType = "repository.created"
	EventRepositoryDeleted     EventType = "repository.deleted"
	EventRepositoryPush        EventType = "repository.push"
	EventRepositoryClone       EventType = "repository.clone"
	EventRepositoryFork        EventType = "repository.fork"
	EventRepositoryStar        EventType = "repository.star"
	EventRepositoryWatch       EventType = "repository.watch"
	EventRepositoryPullRequest EventType = "repository.pull_request"

	// User Events
	EventUserLogin         EventType = "user.login"
	EventUserLogout        EventType = "user.logout"
	EventUserRegistration  EventType = "user.registration"
	EventUserProfileUpdate EventType = "user.profile_update"
	EventUserPasswordReset EventType = "user.password_reset"

	// Organization Events
	EventOrgCreated       EventType = "organization.created"
	EventOrgMemberAdded   EventType = "organization.member_added"
	EventOrgMemberRemoved EventType = "organization.member_removed"
	EventOrgTeamCreated   EventType = "organization.team_created"
	EventOrgRepositoryAdd EventType = "organization.repository_add"

	// CI/CD Events

	EventJobStarted   EventType = "job.started"
	EventJobCompleted EventType = "job.completed"
	EventDeployment   EventType = "deployment"

	// Security Events
	EventSecurityScan EventType = "security.scan"
	EventAccessDenied EventType = "security.access_denied"
	EventAPIKeyUsed   EventType = "security.api_key_used"
	EventMFAEnabled   EventType = "security.mfa_enabled"

	// Performance Events
	EventAPICall     EventType = "api.call"
	EventPageView    EventType = "page.view"
	EventSearchQuery EventType = "search.query"
)

// AnalyticsEvent stores individual analytics events
type AnalyticsEvent struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	CreatedAt time.Time      `json:"created_at" gorm:"index"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	EventType EventType  `json:"event_type" gorm:"type:varchar(100);not null;index"`
	ActorID   *uuid.UUID `json:"actor_id,omitempty" gorm:"type:uuid;index"`
	ActorType string     `json:"actor_type" gorm:"type:varchar(50);index"` // user, system, anonymous

	// Target information
	TargetType string     `json:"target_type" gorm:"type:varchar(50);index"` // repository, user, organization
	TargetID   *uuid.UUID `json:"target_id,omitempty" gorm:"type:uuid;index"`

	// Context information
	RepositoryID   *uuid.UUID `json:"repository_id,omitempty" gorm:"type:uuid;index"`
	OrganizationID *uuid.UUID `json:"organization_id,omitempty" gorm:"type:uuid;index"`

	// Request/session information
	UserAgent string `json:"user_agent" gorm:"type:text"`
	IPAddress string `json:"ip_address" gorm:"type:varchar(45);index"`
	SessionID string `json:"session_id" gorm:"type:varchar(255);index"`
	RequestID string `json:"request_id" gorm:"type:varchar(255);index"`

	// Event metadata
	Metadata     string `json:"metadata" gorm:"type:jsonb"`           // Additional event-specific data
	Duration     *int64 `json:"duration,omitempty"`                   // Duration in milliseconds
	Size         *int64 `json:"size,omitempty"`                       // Size in bytes (for transfers)
	Status       string `json:"status" gorm:"type:varchar(50);index"` // success, error, pending
	ErrorMessage string `json:"error_message,omitempty" gorm:"type:text"`

	// Relationships
	Actor        *User         `json:"actor,omitempty" gorm:"foreignKey:ActorID"`
	Repository   *Repository   `json:"repository,omitempty" gorm:"foreignKey:RepositoryID"`
	Organization *Organization `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
}

func (ae *AnalyticsEvent) TableName() string {
	return "analytics_events"
}

// MetricType represents different types of aggregated metrics
type MetricType string

const (
	MetricTypeCounter   MetricType = "counter"
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeHistogram MetricType = "histogram"
	MetricTypeSummary   MetricType = "summary"
)

// AnalyticsMetric stores aggregated metrics data
type AnalyticsMetric struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	Name       string     `json:"name" gorm:"type:varchar(255);not null;index"`
	MetricType MetricType `json:"metric_type" gorm:"type:varchar(50);not null"`
	Value      float64    `json:"value" gorm:"not null"`
	Timestamp  time.Time  `json:"timestamp" gorm:"not null;index"`

	// Scope information
	RepositoryID   *uuid.UUID `json:"repository_id,omitempty" gorm:"type:uuid;index"`
	OrganizationID *uuid.UUID `json:"organization_id,omitempty" gorm:"type:uuid;index"`
	UserID         *uuid.UUID `json:"user_id,omitempty" gorm:"type:uuid;index"`

	// Aggregation information
	Period string `json:"period" gorm:"type:varchar(50);index"` // hourly, daily, weekly, monthly
	Tags   string `json:"tags" gorm:"type:jsonb"`               // Additional tags for filtering

	// Statistical data for histograms/summaries
	Count        *int64   `json:"count,omitempty"`
	Sum          *float64 `json:"sum,omitempty"`
	Min          *float64 `json:"min,omitempty"`
	Max          *float64 `json:"max,omitempty"`
	Average      *float64 `json:"average,omitempty"`
	Percentile50 *float64 `json:"percentile_50,omitempty"`
	Percentile95 *float64 `json:"percentile_95,omitempty"`
	Percentile99 *float64 `json:"percentile_99,omitempty"`

	// Relationships
	Repository   *Repository   `json:"repository,omitempty" gorm:"foreignKey:RepositoryID"`
	Organization *Organization `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
	User         *User         `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (am *AnalyticsMetric) TableName() string {
	return "analytics_metrics"
}

// RepositoryAnalytics stores repository-specific analytics
type RepositoryAnalytics struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:(gen_random_uuid())"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	RepositoryID uuid.UUID `json:"repository_id" gorm:"type:uuid;not null;uniqueIndex"`
	Date         time.Time `json:"date" gorm:"type:date;not null;index"`

	// Code statistics
	LinesOfCode      int64 `json:"lines_of_code" gorm:"default:0"`
	FileCount        int64 `json:"file_count" gorm:"default:0"`
	CommitCount      int64 `json:"commit_count" gorm:"default:0"`
	BranchCount      int64 `json:"branch_count" gorm:"default:0"`
	ContributorCount int64 `json:"contributor_count" gorm:"default:0"`

	// Activity metrics
	ViewsCount    int64 `json:"views_count" gorm:"default:0"`
	ClonesCount   int64 `json:"clones_count" gorm:"default:0"`
	ForksCount    int64 `json:"forks_count" gorm:"default:0"`
	StarsCount    int64 `json:"stars_count" gorm:"default:0"`
	WatchersCount int64 `json:"watchers_count" gorm:"default:0"`

	// PR metrics
	PullRequestsOpened int64 `json:"pull_requests_opened" gorm:"default:0"`
	PullRequestsClosed int64 `json:"pull_requests_closed" gorm:"default:0"`
	PullRequestsMerged int64 `json:"pull_requests_merged" gorm:"default:0"`

	// Performance metrics
	AveragePRMergeTime *float64 `json:"average_pr_merge_time,omitempty"` // in hours

	BuildSuccessRate *float64 `json:"build_success_rate,omitempty"` // percentage

	// Language breakdown (stored as JSON)
	LanguageStats string `json:"language_stats" gorm:"type:jsonb"`

	// Relationships
	Repository Repository `json:"repository,omitempty" gorm:"foreignKey:RepositoryID"`
}

func (ra *RepositoryAnalytics) TableName() string {
	return "repository_analytics"
}

// UserAnalytics stores user-specific analytics
type UserAnalytics struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:(gen_random_uuid())"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	UserID uuid.UUID `json:"user_id" gorm:"type:uuid;not null;uniqueIndex"`
	Date   time.Time `json:"date" gorm:"type:date;not null;index"`

	// Activity metrics
	LoginCount          int64 `json:"login_count" gorm:"default:0"`
	CommitsCount        int64 `json:"commits_count" gorm:"default:0"`
	PullRequestsCreated int64 `json:"pull_requests_created" gorm:"default:0"`

	CommentsCreated int64 `json:"comments_created" gorm:"default:0"`

	// Repository interactions
	RepositoriesCreated int64 `json:"repositories_created" gorm:"default:0"`
	RepositoriesStarred int64 `json:"repositories_starred" gorm:"default:0"`
	RepositoriesForked  int64 `json:"repositories_forked" gorm:"default:0"`

	// Session information
	SessionDuration *float64 `json:"session_duration,omitempty"` // average in minutes
	PageViews       int64    `json:"page_views" gorm:"default:0"`
	UniquePageViews int64    `json:"unique_page_views" gorm:"default:0"`

	// Performance metrics
	AverageResponseTime *float64 `json:"average_response_time,omitempty"` // in milliseconds

	// Relationships
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (ua *UserAnalytics) TableName() string {
	return "user_analytics"
}

// OrganizationAnalytics stores organization-specific analytics
type OrganizationAnalytics struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:(gen_random_uuid())"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	OrganizationID uuid.UUID `json:"organization_id" gorm:"type:uuid;not null;uniqueIndex"`
	Date           time.Time `json:"date" gorm:"type:date;not null;index"`

	// Organization metrics
	MemberCount     int64 `json:"member_count" gorm:"default:0"`
	TeamCount       int64 `json:"team_count" gorm:"default:0"`
	RepositoryCount int64 `json:"repository_count" gorm:"default:0"`

	// Activity metrics
	TotalCommits      int64 `json:"total_commits" gorm:"default:0"`
	TotalPullRequests int64 `json:"total_pull_requests" gorm:"default:0"`

	// Resource usage
	StorageUsedMB      int64 `json:"storage_used_mb" gorm:"default:0"`
	BandwidthUsedMB    int64 `json:"bandwidth_used_mb" gorm:"default:0"`
	ComputeTimeMinutes int64 `json:"compute_time_minutes" gorm:"default:0"`

	// Cost metrics (optional)
	EstimatedCost *float64 `json:"estimated_cost,omitempty"`

	// Relationships
	Organization Organization `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
}

func (oa *OrganizationAnalytics) TableName() string {
	return "organization_analytics"
}

// SystemAnalytics stores platform-wide analytics
type SystemAnalytics struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:(gen_random_uuid())"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	Date time.Time `json:"date" gorm:"type:date;not null;uniqueIndex"`

	// System-wide metrics
	TotalUsers         int64 `json:"total_users" gorm:"default:0"`
	ActiveUsers        int64 `json:"active_users" gorm:"default:0"`
	TotalOrganizations int64 `json:"total_organizations" gorm:"default:0"`
	TotalRepositories  int64 `json:"total_repositories" gorm:"default:0"`

	// Performance metrics
	AverageResponseTime *float64 `json:"average_response_time,omitempty"` // in milliseconds
	P95ResponseTime     *float64 `json:"p95_response_time,omitempty"`     // in milliseconds
	ErrorRate           *float64 `json:"error_rate,omitempty"`            // percentage
	Uptime              *float64 `json:"uptime,omitempty"`                // percentage

	// Resource metrics
	CPUUsage     *float64 `json:"cpu_usage,omitempty"`    // percentage
	MemoryUsage  *float64 `json:"memory_usage,omitempty"` // percentage
	DiskUsage    *float64 `json:"disk_usage,omitempty"`   // percentage
	NetworkInMB  int64    `json:"network_in_mb" gorm:"default:0"`
	NetworkOutMB int64    `json:"network_out_mb" gorm:"default:0"`

	// Business metrics
	NewRegistrations int64    `json:"new_registrations" gorm:"default:0"`
	ChurnRate        *float64 `json:"churn_rate,omitempty"`  // percentage
	GrowthRate       *float64 `json:"growth_rate,omitempty"` // percentage
}

func (sa *SystemAnalytics) TableName() string {
	return "system_analytics"
}

// PerformanceLog stores detailed performance information
type PerformanceLog struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:(gen_random_uuid())"`
	CreatedAt time.Time      `json:"created_at" gorm:"index"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	RequestID    string `json:"request_id" gorm:"type:varchar(255);index"`
	Method       string `json:"method" gorm:"type:varchar(10);not null"`
	Path         string `json:"path" gorm:"type:varchar(500);not null;index"`
	StatusCode   int    `json:"status_code" gorm:"not null;index"`
	Duration     int64  `json:"duration" gorm:"not null"` // in milliseconds
	ResponseSize int64  `json:"response_size" gorm:"default:0"`

	// User context
	UserID    *uuid.UUID `json:"user_id,omitempty" gorm:"type:uuid;index"`
	IPAddress string     `json:"ip_address" gorm:"type:varchar(45);index"`
	UserAgent string     `json:"user_agent" gorm:"type:text"`

	// Additional context
	RepositoryID   *uuid.UUID `json:"repository_id,omitempty" gorm:"type:uuid;index"`
	OrganizationID *uuid.UUID `json:"organization_id,omitempty" gorm:"type:uuid;index"`

	// Error information
	ErrorMessage string `json:"error_message,omitempty" gorm:"type:text"`
	StackTrace   string `json:"stack_trace,omitempty" gorm:"type:text"`

	// Relationships
	User         *User         `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Repository   *Repository   `json:"repository,omitempty" gorm:"foreignKey:RepositoryID"`
	Organization *Organization `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
}

func (pl *PerformanceLog) TableName() string {
	return "performance_logs"
}
