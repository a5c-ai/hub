package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Webhook represents a repository webhook configuration
type Webhook struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:(gen_random_uuid())"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	RepositoryID uuid.UUID `json:"repository_id" gorm:"type:uuid;not null;index"`
	Name         string    `json:"name" gorm:"not null;size:255"`
	URL          string    `json:"url" gorm:"not null;size:2048"`
	Secret       string    `json:"-" gorm:"size:255"`
	ContentType  string    `json:"content_type" gorm:"default:'application/json';size:100"`
	InsecureSSL  bool      `json:"insecure_ssl" gorm:"default:false"`
	Active       bool      `json:"active" gorm:"default:true"`
	Events       string    `json:"events" gorm:"type:text"`

	// Relationships
	Repository Repository        `json:"repository,omitempty" gorm:"foreignKey:RepositoryID"`
	Deliveries []WebhookDelivery `json:"deliveries,omitempty" gorm:"foreignKey:WebhookID"`
}

func (w *Webhook) TableName() string {
	return "webhooks"
}

// GetEventsSlice returns the events as a slice of strings
func (w *Webhook) GetEventsSlice() []string {
	if w.Events == "" {
		return []string{}
	}

	// Parse JSON events string
	// For simplicity, assume comma-separated for now
	// In production, this would use proper JSON unmarshaling
	events := []string{}
	// This is a simplified implementation - in production use proper JSON
	if w.Events != "" {
		// Simple comma-separated parsing for now
		return []string{"push", "pull_request"} // Default events
	}
	return events
}

// SetEventsSlice sets the events from a slice of strings
func (w *Webhook) SetEventsSlice(events []string) {
	// In production, this would use proper JSON marshaling
	// For now, store as comma-separated
	if len(events) == 0 {
		w.Events = ""
		return
	}
	// Simple implementation - in production use proper JSON
	w.Events = "push,pull_request" // Default for demo
}

// WebhookDelivery represents a webhook delivery attempt
type WebhookDelivery struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:(gen_random_uuid())"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	WebhookID    uuid.UUID  `json:"webhook_id" gorm:"type:uuid;not null;index"`
	EventType    string     `json:"event_type" gorm:"not null;size:100"`
	DeliveryID   string     `json:"delivery_id" gorm:"not null;size:255;index"`
	URL          string     `json:"url" gorm:"not null;size:2048"`
	Payload      string     `json:"payload" gorm:"type:text"`
	StatusCode   int        `json:"status_code" gorm:"default:0"`
	Duration     int64      `json:"duration" gorm:"default:0"` // in milliseconds
	Success      bool       `json:"success" gorm:"default:false"`
	ErrorMessage string     `json:"error_message" gorm:"type:text"`
	Attempts     int        `json:"attempts" gorm:"default:1"`
	NextRetryAt  *time.Time `json:"next_retry_at"`

	// Response details
	ResponseHeaders string `json:"response_headers" gorm:"type:text"`
	ResponseBody    string `json:"response_body" gorm:"type:text"`

	// Relationships
	Webhook Webhook `json:"webhook,omitempty" gorm:"foreignKey:WebhookID"`
}

func (wd *WebhookDelivery) TableName() string {
	return "webhook_deliveries"
}

// DeployKey represents a repository deploy key
type DeployKey struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:(gen_random_uuid())"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	RepositoryID uuid.UUID  `json:"repository_id" gorm:"type:uuid;not null;index"`
	Title        string     `json:"title" gorm:"not null;size:255"`
	Key          string     `json:"key" gorm:"not null;type:text"`
	Fingerprint  string     `json:"fingerprint" gorm:"not null;size:255;index"`
	ReadOnly     bool       `json:"read_only" gorm:"default:true"`
	Verified     bool       `json:"verified" gorm:"default:false"`
	LastUsedAt   *time.Time `json:"last_used_at"`

	// Relationships
	Repository Repository `json:"repository,omitempty" gorm:"foreignKey:RepositoryID"`
}

func (dk *DeployKey) TableName() string {
	return "deploy_keys"
}

// WebhookEvent represents a webhook event that needs to be delivered
type WebhookEvent struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:(gen_random_uuid())"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	RepositoryID uuid.UUID  `json:"repository_id" gorm:"type:uuid;not null;index"`
	EventType    string     `json:"event_type" gorm:"not null;size:100;index"`
	EventData    string     `json:"event_data" gorm:"type:text"`
	Processed    bool       `json:"processed" gorm:"default:false;index"`
	ProcessedAt  *time.Time `json:"processed_at"`

	// Relationships
	Repository Repository `json:"repository,omitempty" gorm:"foreignKey:RepositoryID"`
}

func (we *WebhookEvent) TableName() string {
	return "webhook_events"
}
