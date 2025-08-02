package models

import (
	"time"

	"github.com/google/uuid"
)

// IssueLabel represents the many-to-many relationship between issues and labels
type IssueLabel struct {
	IssueID   uuid.UUID `json:"issue_id" gorm:"type:uuid;primaryKey"`
	LabelID   uuid.UUID `json:"label_id" gorm:"type:uuid;primaryKey"`
	CreatedAt time.Time `json:"created_at"`

	// Relationships
	Issue Issue `json:"issue,omitempty" gorm:"foreignKey:IssueID"`
	Label Label `json:"label,omitempty" gorm:"foreignKey:LabelID"`
}

func (il *IssueLabel) TableName() string {
	return "issue_labels"
}
