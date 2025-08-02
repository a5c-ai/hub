package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IssueState string

const (
	IssueStateOpen   IssueState = "open"
	IssueStateClosed IssueState = "closed"
)

type Issue struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:(gen_random_uuid())"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	RepositoryID uuid.UUID  `json:"repository_id" gorm:"type:uuid;not null;index"`
	Number       int        `json:"number" gorm:"not null"`
	Title        string     `json:"title" gorm:"not null;size:255"`
	Body         string     `json:"body" gorm:"type:text"`
	UserID       *uuid.UUID `json:"user_id" gorm:"type:uuid;index"`
	State        IssueState `json:"state" gorm:"type:varchar(50);not null;check:state IN ('open','closed')"`
	ClosedAt     *time.Time `json:"closed_at"`
	ClosedByID   *uuid.UUID `json:"closed_by_id" gorm:"type:uuid;index"`

	// Relationships
	Repository Repository `json:"repository,omitempty" gorm:"foreignKey:RepositoryID"`
	User       *User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	ClosedBy   *User      `json:"closed_by,omitempty" gorm:"foreignKey:ClosedByID"`
	Comments   []Comment  `json:"comments,omitempty" gorm:"foreignKey:IssueID"`
	Labels     []Label    `json:"labels,omitempty" gorm:"many2many:issue_labels"`
}

func (i *Issue) TableName() string {
	return "issues"
}
