package migrations

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type PasswordResetToken struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:(gen_random_uuid())"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
	Token     string    `gorm:"not null;uniqueIndex;size:255"`
	ExpiresAt time.Time `gorm:"not null"`
	Used      bool      `gorm:"default:false"`
	UsedAt    *time.Time
}

func (PasswordResetToken) TableName() string {
	return "password_reset_tokens"
}

func init() {
	registerMigration("013_password_reset_tokens", migrate013Up, migrate013Down)
}

func migrate013Up(db *gorm.DB) error {
	return db.AutoMigrate(&PasswordResetToken{})
}

func migrate013Down(db *gorm.DB) error {
	return db.Migrator().DropTable(&PasswordResetToken{})
}
