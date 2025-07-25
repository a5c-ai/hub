package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	Username         string     `json:"username" gorm:"uniqueIndex;not null;size:255"`
	Email            string     `json:"email" gorm:"uniqueIndex;not null;size:255"`
	PasswordHash     string     `json:"-" gorm:"not null;size:255"`
	FullName         string     `json:"full_name" gorm:"size:255"`
	AvatarURL        string     `json:"avatar_url" gorm:"type:text"`
	Bio              string     `json:"bio" gorm:"type:text"`
	Location         string     `json:"location" gorm:"size:255"`
	Website          string     `json:"website" gorm:"size:255"`
	Company          string     `json:"company" gorm:"size:255"`
	EmailVerified    bool       `json:"email_verified" gorm:"default:false"`
	TwoFactorEnabled bool       `json:"two_factor_enabled" gorm:"default:false"`
	IsActive         bool       `json:"is_active" gorm:"default:true"`
	IsAdmin          bool       `json:"is_admin" gorm:"default:false"`
	LastLoginAt      *time.Time `json:"last_login_at"`

	// Relationships
	SSHKeys                 []SSHKey                 `json:"ssh_keys,omitempty" gorm:"foreignKey:UserID"`
	OrganizationMembers     []OrganizationMember     `json:"organization_members,omitempty" gorm:"foreignKey:UserID"`
	TeamMembers             []TeamMember             `json:"team_members,omitempty" gorm:"foreignKey:UserID"`
	RepositoryCollaborators []RepositoryCollaborator `json:"repository_collaborators,omitempty" gorm:"foreignKey:UserID"`
	Stars                   []Star                   `json:"stars,omitempty" gorm:"foreignKey:UserID"`
	Issues                  []Issue                  `json:"issues,omitempty" gorm:"foreignKey:UserID"`
	Comments                []Comment                `json:"comments,omitempty" gorm:"foreignKey:UserID"`
}

func (u *User) TableName() string {
	return "users"
}

type SSHKey struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	UserID      uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index"`
	Title       string     `json:"title" gorm:"not null;size:255"`
	KeyData     string     `json:"key_data" gorm:"not null;type:text"`
	Fingerprint string     `json:"fingerprint" gorm:"uniqueIndex;not null;size:255"`
	LastUsedAt  *time.Time `json:"last_used_at"`

	// Relationships
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (s *SSHKey) TableName() string {
	return "ssh_keys"
}
