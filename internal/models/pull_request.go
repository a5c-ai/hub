package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PullRequestState string

const (
	PullRequestStateOpen   PullRequestState = "open"
	PullRequestStateClosed PullRequestState = "closed"
	PullRequestStateMerged PullRequestState = "merged"
)

type PullRequest struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:(gen_random_uuid())"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	RepositoryID     uuid.UUID        `json:"repository_id" gorm:"type:uuid;not null;index"`
	Number           int              `json:"number" gorm:"not null"`
	Title            string           `json:"title" gorm:"not null;size:255"`
	Body             string           `json:"body" gorm:"type:text"`
	UserID           *uuid.UUID       `json:"user_id" gorm:"type:uuid;index"`
	HeadRepositoryID *uuid.UUID       `json:"head_repository_id" gorm:"type:uuid;index"`
	HeadBranch       string           `json:"head_branch" gorm:"not null;size:255"`
	BaseBranch       string           `json:"base_branch" gorm:"not null;size:255"`
	State            PullRequestState `json:"state" gorm:"type:varchar(50);not null;check:state IN ('open','closed','merged')"`
	Draft            bool             `json:"draft" gorm:"default:false"`
	Merged           bool             `json:"merged" gorm:"default:false"`
	MergedAt         *time.Time       `json:"merged_at"`
	MergedByID       *uuid.UUID       `json:"merged_by_id" gorm:"type:uuid;index"`
	ClosedAt         *time.Time       `json:"closed_at"`

	// Relationships
	Repository     Repository  `json:"repository,omitempty" gorm:"foreignKey:RepositoryID"`
	User           *User       `json:"user,omitempty" gorm:"foreignKey:UserID"`
	HeadRepository *Repository `json:"head_repository,omitempty" gorm:"foreignKey:HeadRepositoryID"`
	MergedBy       *User       `json:"merged_by,omitempty" gorm:"foreignKey:MergedByID"`
}

func (pr *PullRequest) TableName() string {
	return "pull_requests"
}
