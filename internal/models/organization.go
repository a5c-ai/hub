package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Organization struct {
	ID           uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
	
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
	Repositories []Repository         `json:"repositories,omitempty" gorm:"foreignKey:OwnerID;foreignKey:OwnerType"`
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
)

type OrganizationMember struct {
	ID             uuid.UUID        `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
	DeletedAt      gorm.DeletedAt   `json:"-" gorm:"index"`
	
	OrganizationID uuid.UUID        `json:"organization_id" gorm:"type:uuid;not null;index"`
	UserID         uuid.UUID        `json:"user_id" gorm:"type:uuid;not null;index"`
	Role           OrganizationRole `json:"role" gorm:"not null;size:50;check:role IN ('owner','admin','member','billing')"`

	// Relationships
	Organization Organization `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
	User         User         `json:"user,omitempty" gorm:"foreignKey:UserID"`
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
	ID             uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
	
	OrganizationID uuid.UUID   `json:"organization_id" gorm:"type:uuid;not null;index"`
	Name           string      `json:"name" gorm:"not null;size:255"`
	Description    string      `json:"description" gorm:"type:text"`
	Privacy        TeamPrivacy `json:"privacy" gorm:"not null;size:50;check:privacy IN ('closed','secret')"`

	// Relationships
	Organization Organization `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
	Members      []TeamMember `json:"members,omitempty" gorm:"foreignKey:TeamID"`
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
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	TeamID uuid.UUID `json:"team_id" gorm:"type:uuid;not null;index"`
	UserID uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	Role   TeamRole  `json:"role" gorm:"not null;size:50;check:role IN ('maintainer','member')"`

	// Relationships
	Team Team `json:"team,omitempty" gorm:"foreignKey:TeamID"`
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (tm *TeamMember) TableName() string {
	return "team_members"
}