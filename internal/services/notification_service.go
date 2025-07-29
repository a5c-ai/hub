package services

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// Notification represents a real-time notification message for a user.
type Notification struct {
	ID        uuid.UUID   `json:"id"`
	Type      string      `json:"type"`
	Payload   interface{} `json:"payload"`
	Timestamp time.Time   `json:"timestamp"`
}

// NotificationService manages subscriptions and broadcasts notifications to users.
type NotificationService interface {
	// Subscribe returns a channel to receive notifications for userID and a cancel function.
	Subscribe(userID uuid.UUID) (<-chan Notification, func())
	// Publish sends a notification to all subscribers of userID.
	Publish(userID uuid.UUID, notification Notification)
}

// notificationService is an in-memory NotificationService implementation.
type notificationService struct {
	mu          sync.RWMutex
	subscribers map[uuid.UUID]map[chan Notification]struct{}
}

// NewNotificationService creates a new in-memory NotificationService.
func NewNotificationService() NotificationService {
	return &notificationService{
		subscribers: make(map[uuid.UUID]map[chan Notification]struct{}),
	}
}

// Subscribe registers a channel for userID and returns the channel and a cancel func.
func (s *notificationService) Subscribe(userID uuid.UUID) (<-chan Notification, func()) {
	ch := make(chan Notification, 16)
	s.mu.Lock()
	subs, ok := s.subscribers[userID]
	if !ok {
		subs = make(map[chan Notification]struct{})
		s.subscribers[userID] = subs
	}
	subs[ch] = struct{}{}
	s.mu.Unlock()

	cancel := func() {
		s.mu.Lock()
		delete(s.subscribers[userID], ch)
		if len(s.subscribers[userID]) == 0 {
			delete(s.subscribers, userID)
		}
		s.mu.Unlock()
		close(ch)
	}
	return ch, cancel
}

// Publish broadcasts notification to all subscribers of userID.
func (s *notificationService) Publish(userID uuid.UUID, notification Notification) {
	s.mu.RLock()
	subs := s.subscribers[userID]
	for ch := range subs {
		select {
		case ch <- notification:
		default:
		}
	}
	s.mu.RUnlock()
}
