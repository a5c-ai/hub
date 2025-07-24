package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OwnerType string

const (
	OwnerTypeUser         OwnerType = "user"
	OwnerTypeOrganization OwnerType = "organization"
)

type Visibility string

const (
	VisibilityPublic   Visibility = "public"
	VisibilityPrivate  Visibility = "private"
	VisibilityInternal Visibility = "internal"
)

type Repository struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	
	OwnerID               uuid.UUID  `json:"owner_id" gorm:"type:uuid;not null;index"`
	OwnerType             OwnerType  `json:"owner_type" gorm:"not null;size:50;check:owner_type IN ('user','organization')"`
	Name                  string     `json:"name" gorm:"not null;size:255"`
	Description           string     `json:"description" gorm:"type:text"`
	DefaultBranch         string     `json:"default_branch" gorm:"default:'main';size:255"`
	Visibility            Visibility `json:"visibility" gorm:"not null;size:50;check:visibility IN ('public','private','internal')"`
	IsFork                bool       `json:"is_fork" gorm:"default:false"`
	ParentID              *uuid.UUID `json:"parent_id" gorm:"type:uuid;index"`
	IsTemplate            bool       `json:"is_template" gorm:"default:false"`
	IsArchived            bool       `json:"is_archived" gorm:"default:false"`
	HasIssues             bool       `json:"has_issues" gorm:"default:true"`
	HasProjects           bool       `json:"has_projects" gorm:"default:true"`
	HasWiki               bool       `json:"has_wiki" gorm:"default:true"`
	HasDownloads          bool       `json:"has_downloads" gorm:"default:true"`
	AllowMergeCommit      bool       `json:"allow_merge_commit" gorm:"default:true"`
	AllowSquashMerge      bool       `json:"allow_squash_merge" gorm:"default:true"`
	AllowRebaseMerge      bool       `json:"allow_rebase_merge" gorm:"default:true"`
	DeleteBranchOnMerge   bool       `json:"delete_branch_on_merge" gorm:"default:false"`
	SizeKB                int64      `json:"size_kb" gorm:"default:0"`
	StarsCount            int        `json:"stars_count" gorm:"default:0"`
	ForksCount            int        `json:"forks_count" gorm:"default:0"`
	WatchersCount         int        `json:"watchers_count" gorm:"default:0"`
	PushedAt              *time.Time `json:"pushed_at"`

	// Relationships
	Parent                 *Repository               `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Forks                  []Repository              `json:"forks,omitempty" gorm:"foreignKey:ParentID"`
	Collaborators          []RepositoryCollaborator  `json:"collaborators,omitempty" gorm:"foreignKey:RepositoryID"`
	Branches               []Branch                  `json:"branches,omitempty" gorm:"foreignKey:RepositoryID"`
	BranchProtectionRules  []BranchProtectionRule    `json:"branch_protection_rules,omitempty" gorm:"foreignKey:RepositoryID"`
	Releases               []Release                 `json:"releases,omitempty" gorm:"foreignKey:RepositoryID"`
	Issues                 []Issue                   `json:"issues,omitempty" gorm:"foreignKey:RepositoryID"`
}

func (r *Repository) TableName() string {
	return "repositories"
}

type Permission string

const (
	PermissionRead     Permission = "read"
	PermissionTriage   Permission = "triage"
	PermissionWrite    Permission = "write"
	PermissionMaintain Permission = "maintain"
	PermissionAdmin    Permission = "admin"
)

type RepositoryCollaborator struct {
	ID           uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
	
	RepositoryID uuid.UUID  `json:"repository_id" gorm:"type:uuid;not null;index"`
	UserID       uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index"`
	Permission   Permission `json:"permission" gorm:"not null;size:50;check:permission IN ('read','triage','write','maintain','admin')"`

	// Relationships
	Repository Repository `json:"repository,omitempty" gorm:"foreignKey:RepositoryID"`
	User       User       `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (rc *RepositoryCollaborator) TableName() string {
	return "repository_collaborators"
}

type Branch struct {
	ID           uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
	
	RepositoryID uuid.UUID `json:"repository_id" gorm:"type:uuid;not null;index"`
	Name         string    `json:"name" gorm:"not null;size:255"`
	SHA          string    `json:"sha" gorm:"not null;size:40"`
	IsProtected  bool      `json:"is_protected" gorm:"default:false"`
	IsDefault    bool      `json:"is_default" gorm:"default:false"`

	// Relationships
	Repository Repository `json:"repository,omitempty" gorm:"foreignKey:RepositoryID"`
}

func (b *Branch) TableName() string {
	return "branches"
}

type BranchProtectionRule struct {
	ID                         uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt                  time.Time      `json:"created_at"`
	UpdatedAt                  time.Time      `json:"updated_at"`
	DeletedAt                  gorm.DeletedAt `json:"-" gorm:"index"`
	
	RepositoryID               uuid.UUID `json:"repository_id" gorm:"type:uuid;not null;index"`
	Pattern                    string    `json:"pattern" gorm:"not null;size:255"`
	RequiredStatusChecks       string    `json:"required_status_checks" gorm:"type:json"`
	EnforceAdmins              bool      `json:"enforce_admins" gorm:"default:false"`
	RequiredPullRequestReviews string    `json:"required_pull_request_reviews" gorm:"type:json"`
	Restrictions               string    `json:"restrictions" gorm:"type:json"`

	// Relationships
	Repository Repository `json:"repository,omitempty" gorm:"foreignKey:RepositoryID"`
}

func (bpr *BranchProtectionRule) TableName() string {
	return "branch_protection_rules"
}

type Release struct {
	ID              uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`
	
	RepositoryID    uuid.UUID  `json:"repository_id" gorm:"type:uuid;not null;index"`
	UserID          *uuid.UUID `json:"user_id" gorm:"type:uuid;index"`
	TagName         string     `json:"tag_name" gorm:"not null;size:255"`
	TargetCommitish string     `json:"target_commitish" gorm:"not null;size:255"`
	Name            string     `json:"name" gorm:"size:255"`
	Body            string     `json:"body" gorm:"type:text"`
	Draft           bool       `json:"draft" gorm:"default:false"`
	Prerelease      bool       `json:"prerelease" gorm:"default:false"`
	PublishedAt     *time.Time `json:"published_at"`

	// Relationships
	Repository Repository `json:"repository,omitempty" gorm:"foreignKey:RepositoryID"`
	User       *User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (r *Release) TableName() string {
	return "releases"
}