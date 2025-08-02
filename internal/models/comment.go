package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Comment struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:(gen_random_uuid())"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	IssueID       *uuid.UUID `json:"issue_id" gorm:"type:uuid;index"`
	PullRequestID *uuid.UUID `json:"pull_request_id" gorm:"type:uuid;index"`
	UserID        *uuid.UUID `json:"user_id" gorm:"type:uuid;index"`
	Body          string     `json:"body" gorm:"not null;type:text"`

	// Relationships
	Issue       *Issue       `json:"issue,omitempty" gorm:"foreignKey:IssueID"`
	PullRequest *PullRequest `json:"pull_request,omitempty" gorm:"foreignKey:PullRequestID"`
	User        *User        `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (c *Comment) TableName() string {
	return "comments"
}
