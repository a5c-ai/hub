package migrations

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type EmailVerificationToken struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index"`
	Token     string     `gorm:"not null;uniqueIndex;size:255"`
	ExpiresAt time.Time  `gorm:"not null"`
	Used      bool       `gorm:"default:false"`
	UsedAt    *time.Time
}

func (EmailVerificationToken) TableName() string {
	return "email_verification_tokens"
}

func init() {
	registerMigration("012_email_verification_tokens", migrate012Up, migrate012Down)
}

func migrate012Up(db *gorm.DB) error {
	return db.AutoMigrate(&EmailVerificationToken{})
}

func migrate012Down(db *gorm.DB) error {
	return db.Migrator().DropTable(&EmailVerificationToken{})
}