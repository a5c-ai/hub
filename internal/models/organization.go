package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Organization struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:(gen_random_uuid())"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	Name         string `json:"name" gorm:"uniqueIndex;not null;size:255"`
	DisplayName  string `json:"display_name" gorm:"not null;size:255"`
	Description  string `json:"description" gorm:"type:text"`
	AvatarURL    string `json:"avatar_url" gorm:"type:text"`
	Website      string `json:"website" gorm:"size:255"`
	Location     string `json:"location" gorm:"size:255"`
	Email        string `json:"email" gorm:"size:255"`
	BillingEmail string `json:"billing_email" gorm:"size:255"`

	// Relationships
	Members      []OrganizationMember `json:"members,omitempty" gorm:"foreignKey:OrganizationID"`
	Teams        []Team               `json:"teams,omitempty" gorm:"foreignKey:OrganizationID"`
	Repositories []Repository         `json:"repositories,omitempty" gorm:"polymorphic:Owner"`
}

func (o *Organization) TableName() string {
	return "organizations"
}

type OrganizationRole string

const (
	OrgRoleOwner   OrganizationRole = "owner"
	OrgRoleAdmin   OrganizationRole = "admin"
	OrgRoleMember  OrganizationRole = "member"
	OrgRoleBilling OrganizationRole = "billing"
	OrgRoleCustom  OrganizationRole = "custom"
)

// Custom Role System
type CustomRole struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:(gen_random_uuid())"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	OrganizationID uuid.UUID `json:"organization_id" gorm:"type:uuid;not null;index"`
	Name           string    `json:"name" gorm:"not null;size:255"`
	Description    string    `json:"description" gorm:"type:text"`
	Permissions    string    `json:"permissions" gorm:"type:jsonb"`
	Color          string    `json:"color" gorm:"size:7;default:'#6b7280'"`
	IsDefault      bool      `json:"is_default" gorm:"default:false"`

	// Relationships
	Organization Organization `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
}

func (cr *CustomRole) TableName() string {
	return "custom_roles"
}

type OrganizationMember struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:(gen_random_uuid())"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	OrganizationID uuid.UUID        `json:"organization_id" gorm:"type:uuid;not null;index"`
	UserID         uuid.UUID        `json:"user_id" gorm:"type:uuid;not null;index"`
	Role           OrganizationRole `json:"role" gorm:"type:varchar(50);not null;check:role IN ('owner','admin','member','billing','custom')"`
	CustomRoleID   *uuid.UUID       `json:"custom_role_id,omitempty" gorm:"type:uuid;index"`
	PublicMember   bool             `json:"public_member" gorm:"default:false"`

	// Relationships
	Organization Organization `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
	User         User         `json:"user,omitempty" gorm:"foreignKey:UserID"`
	CustomRole   *CustomRole  `json:"custom_role,omitempty" gorm:"foreignKey:CustomRoleID"`
}

func (om *OrganizationMember) TableName() string {
	return "organization_members"
}

type TeamPrivacy string

const (
	TeamPrivacyClosed TeamPrivacy = "closed"
	TeamPrivacySecret TeamPrivacy = "secret"
)

type Team struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:(gen_random_uuid())"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	OrganizationID uuid.UUID   `json:"organization_id" gorm:"type:uuid;not null;index"`
	Name           string      `json:"name" gorm:"not null;size:255"`
	Description    string      `json:"description" gorm:"type:text"`
	Privacy        TeamPrivacy `json:"privacy" gorm:"type:varchar(50);not null;check:privacy IN ('closed','secret')"`
	ParentTeamID   *uuid.UUID  `json:"parent_team_id,omitempty" gorm:"type:uuid;index"`

	// Relationships
	Organization Organization           `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
	Members      []TeamMember           `json:"members,omitempty" gorm:"foreignKey:TeamID"`
	ParentTeam   *Team                  `json:"parent_team,omitempty" gorm:"foreignKey:ParentTeamID"`
	ChildTeams   []Team                 `json:"child_teams,omitempty" gorm:"foreignKey:ParentTeamID"`
	Permissions  []RepositoryPermission `json:"permissions,omitempty" gorm:"foreignKey:SubjectID"`
}

func (t *Team) TableName() string {
	return "teams"
}

type TeamRole string

const (
	TeamRoleMaintainer TeamRole = "maintainer"
	TeamRoleMember     TeamRole = "member"
)

type TeamMember struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:(gen_random_uuid())"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	TeamID uuid.UUID `json:"team_id" gorm:"type:uuid;not null;index"`
	UserID uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	Role   TeamRole  `json:"role" gorm:"type:varchar(50);not null;check:role IN ('maintainer','member')"`

	// Relationships
	Team Team `json:"team,omitempty" gorm:"foreignKey:TeamID"`
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (tm *TeamMember) TableName() string {
	return "team_members"
}

// Repository Permission System
type SubjectType string

const (
	SubjectTypeUser SubjectType = "user"
	SubjectTypeTeam SubjectType = "team"
)

type RepositoryPermission struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:(gen_random_uuid())"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	RepositoryID uuid.UUID   `json:"repository_id" gorm:"type:uuid;not null;index"`
	SubjectID    uuid.UUID   `json:"subject_id" gorm:"type:uuid;not null;index"`
	SubjectType  SubjectType `json:"subject_type" gorm:"type:varchar(50);not null;check:subject_type IN ('user','team')"`
	Permission   Permission  `json:"permission" gorm:"type:varchar(50);not null;check:permission IN ('read','triage','write','maintain','admin')"`

	// Relationships
	Repository Repository `json:"repository,omitempty" gorm:"foreignKey:RepositoryID"`
}

func (rp *RepositoryPermission) TableName() string {
	return "repository_permissions"
}

// Organization Invitation System
type OrganizationInvitation struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:(gen_random_uuid())"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	OrganizationID uuid.UUID        `json:"organization_id" gorm:"type:uuid;not null;index"`
	InviterID      uuid.UUID        `json:"inviter_id" gorm:"type:uuid;not null;index"`
	Email          string           `json:"email" gorm:"not null;size:255;index"`
	Role           OrganizationRole `json:"role" gorm:"type:varchar(50);not null;check:role IN ('owner','admin','member','billing','custom')"`
	CustomRoleID   *uuid.UUID       `json:"custom_role_id,omitempty" gorm:"type:uuid;index"`
	Token          string           `json:"-" gorm:"uniqueIndex;not null;size:255"`
	ExpiresAt      time.Time        `json:"expires_at" gorm:"not null;index"`
	AcceptedAt     *time.Time       `json:"accepted_at,omitempty"`

	// Relationships
	Organization Organization `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
	Inviter      User         `json:"inviter,omitempty" gorm:"foreignKey:InviterID"`
}

func (oi *OrganizationInvitation) TableName() string {
	return "organization_invitations"
}

// Activity Logging System
type ActivityAction string

const (
	ActivityMemberAdded             ActivityAction = "member.added"
	ActivityMemberRemoved           ActivityAction = "member.removed"
	ActivityMemberRoleChanged       ActivityAction = "member.role_changed"
	ActivityMemberVisibilityChanged ActivityAction = "member.visibility_changed"
	ActivityTeamCreated             ActivityAction = "team.created"
	ActivityTeamDeleted             ActivityAction = "team.deleted"
	ActivityTeamUpdated             ActivityAction = "team.updated"
	ActivityRepositoryCreated       ActivityAction = "repository.created"
	ActivityRepositoryDeleted       ActivityAction = "repository.deleted"
	ActivityInvitationSent          ActivityAction = "invitation.sent"
	ActivityInvitationAccepted      ActivityAction = "invitation.accepted"
	ActivityPermissionGranted       ActivityAction = "permission.granted"
	ActivityPermissionRevoked       ActivityAction = "permission.revoked"
)

type OrganizationActivity struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:(gen_random_uuid())"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	OrganizationID uuid.UUID      `json:"organization_id" gorm:"type:uuid;not null;index"`
	ActorID        uuid.UUID      `json:"actor_id" gorm:"type:uuid;not null;index"`
	Action         ActivityAction `json:"action" gorm:"type:varchar(100);not null"`
	TargetType     string         `json:"target_type" gorm:"size:50"`
	TargetID       *uuid.UUID     `json:"target_id,omitempty" gorm:"type:uuid;index"`
	Metadata       string         `json:"metadata,omitempty" gorm:"type:jsonb"`

	// Relationships
	Organization Organization `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
	Actor        User         `json:"actor,omitempty" gorm:"foreignKey:ActorID"`
}

func (oa *OrganizationActivity) TableName() string {
	return "organization_activities"
}

// Organization Policy System
type PolicyType string

const (
	PolicyTypeRepositoryCreation PolicyType = "repository_creation"
	PolicyTypeMemberInvitation   PolicyType = "member_invitation"
	PolicyTypeBranchProtection   PolicyType = "branch_protection"
	PolicyTypeSecretManagement   PolicyType = "secret_management"
	PolicyTypeComplianceGDPR     PolicyType = "compliance_gdpr"
	PolicyTypeComplianceSOC2     PolicyType = "compliance_soc2"
	PolicyTypeIPRestriction      PolicyType = "ip_restriction"
	PolicyType2FAEnforcement     PolicyType = "2fa_enforcement"
	PolicyTypeSSO                PolicyType = "sso_enforcement"
)

type OrganizationPolicy struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:(gen_random_uuid())"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	OrganizationID uuid.UUID  `json:"organization_id" gorm:"type:uuid;not null;index"`
	PolicyType     PolicyType `json:"policy_type" gorm:"type:varchar(100);not null"`
	Name           string     `json:"name" gorm:"not null;size:255"`
	Description    string     `json:"description" gorm:"type:text"`
	Configuration  string     `json:"configuration" gorm:"type:jsonb"`
	Enabled        bool       `json:"enabled" gorm:"default:true"`
	Enforcement    string     `json:"enforcement" gorm:"type:varchar(50);default:warn"`

	// Relationships
	Organization Organization `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
}

func (op *OrganizationPolicy) TableName() string {
	return "organization_policies"
}

// Organization Templates System
type TemplateType string

const (
	TemplateTypeRepository TemplateType = "repository"
	TemplateTypeProject    TemplateType = "project"

	TemplateTypeTeam TemplateType = "team"
)

type OrganizationTemplate struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:(gen_random_uuid())"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	OrganizationID uuid.UUID    `json:"organization_id" gorm:"type:uuid;not null;index"`
	TemplateType   TemplateType `json:"template_type" gorm:"type:varchar(50);not null"`
	Name           string       `json:"name" gorm:"not null;size:255"`
	Description    string       `json:"description" gorm:"type:text"`
	Configuration  string       `json:"configuration" gorm:"type:jsonb"`
	IsDefault      bool         `json:"is_default" gorm:"default:false"`
	UsageCount     int          `json:"usage_count" gorm:"default:0"`

	// Relationships
	Organization Organization `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
}

func (ot *OrganizationTemplate) TableName() string {
	return "organization_templates"
}

// Organization Settings Enhancement
type OrganizationSettings struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:(gen_random_uuid())"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	OrganizationID uuid.UUID `json:"organization_id" gorm:"type:uuid;not null;uniqueIndex"`

	// Branding
	PrimaryColor   string `json:"primary_color" gorm:"size:7;default:'#1f2937'"`
	SecondaryColor string `json:"secondary_color" gorm:"size:7;default:'#6b7280'"`
	LogoURL        string `json:"logo_url" gorm:"type:text"`
	CustomCSS      string `json:"custom_css" gorm:"type:text"`

	// Security Settings
	RequireTwoFactor bool   `json:"require_two_factor" gorm:"default:false"`
	AllowedIPRanges  string `json:"allowed_ip_ranges" gorm:"type:jsonb"`
	SSOProvider      string `json:"sso_provider" gorm:"size:100"`
	SSOConfiguration string `json:"sso_configuration" gorm:"type:jsonb"`
	SessionTimeout   int    `json:"session_timeout" gorm:"default:86400"`

	// Repository Settings
	DefaultVisibility         string `json:"default_visibility" gorm:"size:20;default:'private'"`
	AllowPrivateRepos         bool   `json:"allow_private_repos" gorm:"default:true"`
	AllowInternalRepos        bool   `json:"allow_internal_repos" gorm:"default:true"`
	AllowForking              bool   `json:"allow_forking" gorm:"default:true"`
	AllowOutsideCollaborators bool   `json:"allow_outside_collaborators" gorm:"default:true"`

	// Billing and Usage
	BillingPlan     string     `json:"billing_plan" gorm:"size:50;default:'free'"`
	SeatCount       int        `json:"seat_count" gorm:"default:0"`
	StorageLimit    int64      `json:"storage_limit_gb" gorm:"default:1024"`
	BandwidthLimit  int64      `json:"bandwidth_limit_gb" gorm:"default:1024"`
	NextBillingDate *time.Time `json:"next_billing_date"`

	// Backup and Recovery
	BackupEnabled   bool   `json:"backup_enabled" gorm:"default:false"`
	BackupFrequency string `json:"backup_frequency" gorm:"size:20;default:'daily'"`
	RetentionDays   int    `json:"retention_days" gorm:"default:30"`

	// Relationships
	Organization Organization `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
}

func (os *OrganizationSettings) TableName() string {
	return "organization_settings"
}
