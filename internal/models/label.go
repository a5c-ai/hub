package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Label struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:(gen_random_uuid())"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	RepositoryID uuid.UUID `json:"repository_id" gorm:"type:uuid;not null;index"`
	Name         string    `json:"name" gorm:"not null;size:255"`
	Description  string    `json:"description" gorm:"type:text"`
	Color        string    `json:"color" gorm:"not null;size:7;default:'#6b7280'"` // Hex color code
	IsDefault    bool      `json:"is_default" gorm:"default:false"`

	// Relationships
	Repository Repository `json:"repository,omitempty" gorm:"foreignKey:RepositoryID"`
}

func (l *Label) TableName() string {
	return "labels"
}
