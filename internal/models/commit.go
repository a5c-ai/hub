package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Commit represents a Git commit in the database
type Commit struct {
	ID           uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
	
	RepositoryID uuid.UUID `json:"repository_id" gorm:"type:uuid;not null;index"`
	SHA          string    `json:"sha" gorm:"not null;size:40;uniqueIndex:idx_repo_sha"`
	Message      string    `json:"message" gorm:"type:text"`
	AuthorName   string    `json:"author_name" gorm:"not null;size:255"`
	AuthorEmail  string    `json:"author_email" gorm:"not null;size:255"`
	AuthorDate   time.Time `json:"author_date" gorm:"not null"`
	CommitterName  string  `json:"committer_name" gorm:"not null;size:255"`
	CommitterEmail string  `json:"committer_email" gorm:"not null;size:255"`
	CommitterDate  time.Time `json:"committer_date" gorm:"not null"`
	TreeSHA      string    `json:"tree_sha" gorm:"not null;size:40"`
	ParentSHA    string    `json:"parent_sha" gorm:"size:40"` // For merge commits, we'll store the first parent
	
	// Statistics
	Additions int `json:"additions" gorm:"default:0"`
	Deletions int `json:"deletions" gorm:"default:0"`
	Changes   int `json:"changes" gorm:"default:0"`
	
	// Relationships
	Repository Repository `json:"repository,omitempty" gorm:"foreignKey:RepositoryID"`
}

func (c *Commit) TableName() string {
	return "commits"
}

// CommitFile represents a file changed in a commit
type CommitFile struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	CommitID    uuid.UUID `json:"commit_id" gorm:"type:uuid;not null;index"`
	Path        string    `json:"path" gorm:"not null;size:500"`
	PreviousPath string   `json:"previous_path" gorm:"size:500"`
	Status      string    `json:"status" gorm:"not null;size:20"` // added, modified, deleted, renamed
	Additions   int       `json:"additions" gorm:"default:0"`
	Deletions   int       `json:"deletions" gorm:"default:0"`
	Changes     int       `json:"changes" gorm:"default:0"`
	
	// Relationships
	Commit Commit `json:"commit,omitempty" gorm:"foreignKey:CommitID"`
}

func (cf *CommitFile) TableName() string {
	return "commit_files"
}

// Tag represents a Git tag
type Tag struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	RepositoryID uuid.UUID  `json:"repository_id" gorm:"type:uuid;not null;index"`
	Name         string     `json:"name" gorm:"not null;size:255"`
	SHA          string     `json:"sha" gorm:"not null;size:40"`
	Message      string     `json:"message" gorm:"type:text"`
	TaggerName   string     `json:"tagger_name" gorm:"size:255"`
	TaggerEmail  string     `json:"tagger_email" gorm:"size:255"`
	TaggerDate   *time.Time `json:"tagger_date"`
	
	// Relationships
	Repository Repository `json:"repository,omitempty" gorm:"foreignKey:RepositoryID"`
}

func (t *Tag) TableName() string {
	return "tags"
}

// GitRef represents a Git reference (branch or tag)
type GitRef struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	RepositoryID uuid.UUID `json:"repository_id" gorm:"type:uuid;not null;index"`
	Name         string    `json:"name" gorm:"not null;size:255"` // refs/heads/main, refs/tags/v1.0.0
	SHA          string    `json:"sha" gorm:"not null;size:40"`
	Type         string    `json:"type" gorm:"not null;size:20"` // branch, tag
	
	// Relationships
	Repository Repository `json:"repository,omitempty" gorm:"foreignKey:RepositoryID"`
}

func (gr *GitRef) TableName() string {
	return "git_refs"
}

// RepositoryHook represents a repository webhook
type RepositoryHook struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	RepositoryID uuid.UUID `json:"repository_id" gorm:"type:uuid;not null;index"`
	Name         string    `json:"name" gorm:"not null;size:255"`
	URL          string    `json:"url" gorm:"not null;size:500"`
	Secret       string    `json:"secret" gorm:"size:255"` // Webhook secret for verification
	Events       string    `json:"events" gorm:"type:json"` // JSON array of events
	Active       bool      `json:"active" gorm:"default:true"`
	InsecureSSL  bool      `json:"insecure_ssl" gorm:"default:false"`
	ContentType  string    `json:"content_type" gorm:"default:'application/json';size:50"`
	
	// Relationships
	Repository Repository `json:"repository,omitempty" gorm:"foreignKey:RepositoryID"`
}

func (rh *RepositoryHook) TableName() string {
	return "repository_hooks"
}

// RepositoryClone represents a clone operation for analytics
type RepositoryClone struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	RepositoryID uuid.UUID  `json:"repository_id" gorm:"type:uuid;not null;index"`
	UserID       *uuid.UUID `json:"user_id" gorm:"type:uuid;index"` // Null for anonymous clones
	IPAddress    string     `json:"ip_address" gorm:"size:45"` // IPv4 or IPv6
	UserAgent    string     `json:"user_agent" gorm:"type:text"`
	Protocol     string     `json:"protocol" gorm:"size:10"` // http, https, ssh, git
	
	// Relationships
	Repository Repository `json:"repository,omitempty" gorm:"foreignKey:RepositoryID"`
	User       *User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (rc *RepositoryClone) TableName() string {
	return "repository_clones"
}

// RepositoryView represents a repository view for analytics
type RepositoryView struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	RepositoryID uuid.UUID  `json:"repository_id" gorm:"type:uuid;not null;index"`
	UserID       *uuid.UUID `json:"user_id" gorm:"type:uuid;index"` // Null for anonymous views
	IPAddress    string     `json:"ip_address" gorm:"size:45"`
	UserAgent    string     `json:"user_agent" gorm:"type:text"`
	Path         string     `json:"path" gorm:"size:500"` // Path within repository
	
	// Relationships
	Repository Repository `json:"repository,omitempty" gorm:"foreignKey:RepositoryID"`
	User       *User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (rv *RepositoryView) TableName() string {
	return "repository_views"
}