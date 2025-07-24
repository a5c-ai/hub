package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	Username     string `json:"username" gorm:"uniqueIndex;not null"`
	Email        string `json:"email" gorm:"uniqueIndex;not null"`
	PasswordHash string `json:"-" gorm:"not null"`
	FullName     string `json:"full_name"`
	AvatarURL    string `json:"avatar_url"`
	IsActive     bool   `json:"is_active" gorm:"default:true"`
	IsAdmin      bool   `json:"is_admin" gorm:"default:false"`
	LastLoginAt  *time.Time `json:"last_login_at"`
}

func (u *User) TableName() string {
	return "users"
}