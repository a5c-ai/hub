package migrations

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type Session struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:(gen_random_uuid())"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	UserID       uuid.UUID `gorm:"type:uuid;not null;index"`
	RefreshToken string    `gorm:"not null;uniqueIndex;size:255"`
	ExpiresAt    time.Time `gorm:"not null"`
	IPAddress    string    `gorm:"size:45"`
	UserAgent    string    `gorm:"size:255"`
	IsActive     bool      `gorm:"default:true"`
	LastUsedAt   time.Time
}

func (Session) TableName() string {
	return "sessions"
}

func init() {
	registerMigration("010_sessions", migrate010Up, migrate010Down)
}

func migrate010Up(db *gorm.DB) error {
	return db.AutoMigrate(&Session{})
}

func migrate010Down(db *gorm.DB) error {
	return db.Migrator().DropTable(&Session{})
}
