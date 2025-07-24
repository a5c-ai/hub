package migrations

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type SecurityEvent struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
	
	UserID      *uuid.UUID `gorm:"type:uuid;index"`
	EventType   string     `gorm:"not null;size:50;index"`
	IPAddress   string     `gorm:"size:45;index"`
	UserAgent   string     `gorm:"size:255"`
	Details     string     `gorm:"type:text"`
	Severity    string     `gorm:"size:20;default:'info'"`
}

func (SecurityEvent) TableName() string {
	return "security_events"
}

type LoginAttempt struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
	
	UserID      *uuid.UUID `gorm:"type:uuid;index"`
	Email       string     `gorm:"not null;size:255;index"`
	IPAddress   string     `gorm:"not null;size:45;index"`
	Success     bool       `gorm:"not null;index"`
	UserAgent   string     `gorm:"size:255"`
	FailReason  string     `gorm:"size:255"`
}

func (LoginAttempt) TableName() string {
	return "login_attempts"
}

func init() {
	registerMigration("007_security_tables", migrate007Up, migrate007Down)
}

func migrate007Up(db *gorm.DB) error {
	err := db.AutoMigrate(&SecurityEvent{})
	if err != nil {
		return err
	}
	return db.AutoMigrate(&LoginAttempt{})
}

func migrate007Down(db *gorm.DB) error {
	err := db.Migrator().DropTable(&LoginAttempt{})
	if err != nil {
		return err
	}
	return db.Migrator().DropTable(&SecurityEvent{})
}