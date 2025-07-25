package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RepositoryLanguage represents programming language statistics for a repository
type RepositoryLanguage struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	RepositoryID uuid.UUID `json:"repository_id" gorm:"type:uuid;not null;index"`
	Language     string    `json:"language" gorm:"not null;size:100"`
	Bytes        int64     `json:"bytes" gorm:"not null;default:0"`
	Percentage   float64   `json:"percentage" gorm:"not null;default:0"`

	// Relationships
	Repository Repository `json:"repository,omitempty" gorm:"foreignKey:RepositoryID"`
}

func (rl *RepositoryLanguage) TableName() string {
	return "repository_languages"
}

// RepositoryStatistics represents comprehensive statistics for a repository
type RepositoryStatistics struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	RepositoryID     uuid.UUID  `json:"repository_id" gorm:"type:uuid;not null;uniqueIndex"`
	SizeBytes        int64      `json:"size_bytes" gorm:"not null;default:0"`
	CommitCount      int        `json:"commit_count" gorm:"not null;default:0"`
	BranchCount      int        `json:"branch_count" gorm:"not null;default:0"`
	TagCount         int        `json:"tag_count" gorm:"not null;default:0"`
	Contributors     int        `json:"contributors" gorm:"not null;default:0"`
	LastActivity     *time.Time `json:"last_activity"`
	PrimaryLanguage  string     `json:"primary_language" gorm:"size:100"`
	LanguageCount    int        `json:"language_count" gorm:"not null;default:0"`

	// Relationships
	Repository Repository            `json:"repository,omitempty" gorm:"foreignKey:RepositoryID"`
	Languages  []RepositoryLanguage  `json:"languages,omitempty" gorm:"foreignKey:RepositoryID"`
}

func (rs *RepositoryStatistics) TableName() string {
	return "repository_statistics"
}

// RepositoryTemplate represents a repository template
type RepositoryTemplate struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	RepositoryID   uuid.UUID `json:"repository_id" gorm:"type:uuid;not null;uniqueIndex"`
	Name           string    `json:"name" gorm:"not null;size:255"`
	Description    string    `json:"description" gorm:"type:text"`
	Category       string    `json:"category" gorm:"size:100"`
	Tags           string    `json:"tags" gorm:"type:json"` // JSON array of tags
	UsageCount     int       `json:"usage_count" gorm:"not null;default:0"`
	IsFeatured     bool      `json:"is_featured" gorm:"default:false"`
	IsPublic       bool      `json:"is_public" gorm:"default:true"`

	// Relationships
	Repository Repository `json:"repository,omitempty" gorm:"foreignKey:RepositoryID"`
}

func (rt *RepositoryTemplate) TableName() string {
	return "repository_templates"
}

// GitHook represents a Git hook configuration for a repository
type GitHook struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	RepositoryID uuid.UUID `json:"repository_id" gorm:"type:uuid;not null;index"`
	HookType     string    `json:"hook_type" gorm:"not null;size:50"` // pre-receive, post-receive, update, etc.
	IsEnabled    bool      `json:"is_enabled" gorm:"default:true"`
	Script       string    `json:"script" gorm:"type:text"` // Hook script content
	Language     string    `json:"language" gorm:"size:20;default:'bash'"` // bash, python, etc.
	Order        int       `json:"order" gorm:"default:0"` // Execution order for multiple hooks

	// Relationships
	Repository Repository `json:"repository,omitempty" gorm:"foreignKey:RepositoryID"`
}

func (gh *GitHook) TableName() string {
	return "git_hooks"
}

// RepositoryImport represents an import operation for a repository
type RepositoryImport struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	RepositoryID   uuid.UUID `json:"repository_id" gorm:"type:uuid;not null;index"`
	SourceType     string    `json:"source_type" gorm:"not null;size:50"` // github, gitlab, bitbucket
	SourceURL      string    `json:"source_url" gorm:"not null;size:500"`
	Status         string    `json:"status" gorm:"not null;size:50;default:'pending'"` // pending, in_progress, completed, failed
	Progress       int       `json:"progress" gorm:"default:0"` // 0-100
	ErrorMessage   string    `json:"error_message" gorm:"type:text"`
	ImportedAt     *time.Time `json:"imported_at"`
	TotalCommits   int       `json:"total_commits" gorm:"default:0"`
	ImportedCommits int      `json:"imported_commits" gorm:"default:0"`
	TotalBranches  int       `json:"total_branches" gorm:"default:0"`
	ImportedBranches int     `json:"imported_branches" gorm:"default:0"`

	// Relationships
	Repository Repository `json:"repository,omitempty" gorm:"foreignKey:RepositoryID"`
}

func (ri *RepositoryImport) TableName() string {
	return "repository_imports"
}