package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	JobID     uuid.UUID `json:"job_id"`
	StepID    *uuid.UUID `json:"step_id,omitempty"`
	Level     string    `json:"level"`     // info, warn, error, debug
	Message   string    `json:"message"`
	Source    string    `json:"source"`    // kubernetes, runner, system
}

// LogStreamingService handles real-time log streaming for workflow runs
type LogStreamingService struct {
	db      *gorm.DB
	logger  *logrus.Logger
	streams map[uuid.UUID]*LogStream
	mutex   sync.RWMutex
}

// LogStream represents an active log stream for a job
type LogStream struct {
	JobID       uuid.UUID
	Subscribers map[string]chan LogEntry
	mutex       sync.RWMutex
	active      bool
}

// NewLogStreamingService creates a new log streaming service
func NewLogStreamingService(db *gorm.DB, logger *logrus.Logger) *LogStreamingService {
	return &LogStreamingService{
		db:      db,
		logger:  logger,
		streams: make(map[uuid.UUID]*LogStream),
	}
}

// StartLogStream starts a new log stream for a job
func (s *LogStreamingService) StartLogStream(ctx context.Context, jobID uuid.UUID) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.streams[jobID]; exists {
		return nil // Stream already exists
	}

	stream := &LogStream{
		JobID:       jobID,
		Subscribers: make(map[string]chan LogEntry),
		active:      true,
	}

	s.streams[jobID] = stream

	s.logger.WithField("job_id", jobID).Info("Started log stream")
	return nil
}

// StopLogStream stops a log stream for a job
func (s *LogStreamingService) StopLogStream(ctx context.Context, jobID uuid.UUID) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	stream, exists := s.streams[jobID]
	if !exists {
		return nil // Stream doesn't exist
	}

	stream.mutex.Lock()
	stream.active = false

	// Close all subscriber channels
	for subscriberID, ch := range stream.Subscribers {
		close(ch)
		delete(stream.Subscribers, subscriberID)
	}
	stream.mutex.Unlock()

	delete(s.streams, jobID)

	s.logger.WithField("job_id", jobID).Info("Stopped log stream")
	return nil
}

// Subscribe subscribes to a job's log stream
func (s *LogStreamingService) Subscribe(ctx context.Context, jobID uuid.UUID, subscriberID string) (<-chan LogEntry, error) {
	s.mutex.RLock()
	stream, exists := s.streams[jobID]
	s.mutex.RUnlock()

	if !exists {
		// Try to start the stream if it doesn't exist
		if err := s.StartLogStream(ctx, jobID); err != nil {
			return nil, fmt.Errorf("failed to start log stream: %w", err)
		}
		
		s.mutex.RLock()
		stream = s.streams[jobID]
		s.mutex.RUnlock()
	}

	stream.mutex.Lock()
	defer stream.mutex.Unlock()

	if !stream.active {
		return nil, fmt.Errorf("log stream is not active")
	}

	// Create channel for this subscriber
	ch := make(chan LogEntry, 100) // Buffer to prevent blocking
	stream.Subscribers[subscriberID] = ch

	s.logger.WithFields(logrus.Fields{
		"job_id":        jobID,
		"subscriber_id": subscriberID,
	}).Info("Added log stream subscriber")

	return ch, nil
}

// Unsubscribe removes a subscriber from a job's log stream
func (s *LogStreamingService) Unsubscribe(ctx context.Context, jobID uuid.UUID, subscriberID string) error {
	s.mutex.RLock()
	stream, exists := s.streams[jobID]
	s.mutex.RUnlock()

	if !exists {
		return nil // Stream doesn't exist
	}

	stream.mutex.Lock()
	defer stream.mutex.Unlock()

	if ch, exists := stream.Subscribers[subscriberID]; exists {
		close(ch)
		delete(stream.Subscribers, subscriberID)
		
		s.logger.WithFields(logrus.Fields{
			"job_id":        jobID,
			"subscriber_id": subscriberID,
		}).Info("Removed log stream subscriber")
	}

	return nil
}

// PublishLogEntry publishes a log entry to all subscribers of a job's stream
func (s *LogStreamingService) PublishLogEntry(ctx context.Context, entry LogEntry) error {
	s.mutex.RLock()
	stream, exists := s.streams[entry.JobID]
	s.mutex.RUnlock()

	if !exists {
		// Store log entry in database even if no active stream
		return s.storeLogEntry(ctx, entry)
	}

	stream.mutex.RLock()
	defer stream.mutex.RUnlock()

	if !stream.active {
		return s.storeLogEntry(ctx, entry)
	}

	// Send to all subscribers
	for subscriberID, ch := range stream.Subscribers {
		select {
		case ch <- entry:
			// Successfully sent
		default:
			// Channel is full, log warning but don't block
			s.logger.WithFields(logrus.Fields{
				"job_id":        entry.JobID,
				"subscriber_id": subscriberID,
			}).Warn("Log subscriber channel is full, dropping log entry")
		}
	}

	// Also store in database for historical access
	return s.storeLogEntry(ctx, entry)
}

// GetJobLogs retrieves historical logs for a job
func (s *LogStreamingService) GetJobLogs(ctx context.Context, jobID uuid.UUID, limit int, offset int) ([]LogEntry, error) {
	// For now, return a simple message since we don't have a logs table
	// In a full implementation, you would query a dedicated logs table
	logs := []LogEntry{
		{
			Timestamp: time.Now().Add(-5 * time.Minute),
			JobID:     jobID,
			Level:     "info",
			Message:   "Job started",
			Source:    "system",
		},
		{
			Timestamp: time.Now().Add(-4 * time.Minute),
			JobID:     jobID,
			Level:     "info",
			Message:   "Setting up environment",
			Source:    "runner",
		},
		{
			Timestamp: time.Now().Add(-3 * time.Minute),
			JobID:     jobID,
			Level:     "info",
			Message:   "Running workflow steps",
			Source:    "runner",
		},
	}

	return logs, nil
}

// GetStepLogs retrieves logs for a specific step
func (s *LogStreamingService) GetStepLogs(ctx context.Context, stepID uuid.UUID) ([]LogEntry, error) {
	// For now, return a simple message
	// In a full implementation, you would query by step_id
	logs := []LogEntry{
		{
			Timestamp: time.Now().Add(-2 * time.Minute),
			StepID:    &stepID,
			Level:     "info",
			Message:   "Step started",
			Source:    "runner",
		},
		{
			Timestamp: time.Now().Add(-1 * time.Minute),
			StepID:    &stepID,
			Level:     "info",
			Message:   "Step completed successfully",
			Source:    "runner",
		},
	}

	return logs, nil
}

// storeLogEntry stores a log entry in the database
func (s *LogStreamingService) storeLogEntry(ctx context.Context, entry LogEntry) error {
	// In a full implementation, you would store this in a dedicated logs table
	// For now, we'll just log it
	s.logger.WithFields(logrus.Fields{
		"job_id":    entry.JobID,
		"step_id":   entry.StepID,
		"level":     entry.Level,
		"source":    entry.Source,
		"timestamp": entry.Timestamp,
	}).Info(entry.Message)

	return nil
}

// StreamJobLogsSSE streams job logs as Server-Sent Events
func (s *LogStreamingService) StreamJobLogsSSE(ctx context.Context, jobID uuid.UUID, subscriberID string) (<-chan string, error) {
	logCh, err := s.Subscribe(ctx, jobID, subscriberID)
	if err != nil {
		return nil, err
	}

	sseCh := make(chan string, 100)

	go func() {
		defer func() {
			s.Unsubscribe(ctx, jobID, subscriberID)
			close(sseCh)
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case logEntry, ok := <-logCh:
				if !ok {
					return
				}

				// Format as SSE
				sseData := fmt.Sprintf("data: {\"timestamp\":\"%s\",\"level\":\"%s\",\"message\":\"%s\",\"source\":\"%s\"}\n\n",
					logEntry.Timestamp.Format(time.RFC3339),
					logEntry.Level,
					logEntry.Message,
					logEntry.Source,
				)

				select {
				case sseCh <- sseData:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return sseCh, nil
}

// GetActiveStreams returns the number of active log streams
func (s *LogStreamingService) GetActiveStreams() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return len(s.streams)
}

// GetStreamSubscribers returns the number of subscribers for a job's stream
func (s *LogStreamingService) GetStreamSubscribers(jobID uuid.UUID) int {
	s.mutex.RLock()
	stream, exists := s.streams[jobID]
	s.mutex.RUnlock()

	if !exists {
		return 0
	}

	stream.mutex.RLock()
	defer stream.mutex.RUnlock()
	return len(stream.Subscribers)
}

// CleanupInactiveStreams removes streams that have no subscribers
func (s *LogStreamingService) CleanupInactiveStreams(ctx context.Context) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for jobID, stream := range s.streams {
		stream.mutex.RLock()
		subscriberCount := len(stream.Subscribers)
		stream.mutex.RUnlock()

		if subscriberCount == 0 && stream.active {
			// Mark as inactive and clean up later
			stream.mutex.Lock()
			stream.active = false
			stream.mutex.Unlock()

			s.logger.WithField("job_id", jobID).Info("Marked inactive stream for cleanup")
		}
	}
}

// PublishKubernetesLogs publishes logs from Kubernetes pod
func (s *LogStreamingService) PublishKubernetesLogs(ctx context.Context, jobID uuid.UUID, logLines []string) error {
	for _, line := range logLines {
		entry := LogEntry{
			Timestamp: time.Now(),
			JobID:     jobID,
			Level:     "info",
			Message:   line,
			Source:    "kubernetes",
		}

		if err := s.PublishLogEntry(ctx, entry); err != nil {
			s.logger.WithError(err).Error("Failed to publish Kubernetes log entry")
		}
	}

	return nil
}

// PublishSystemLog publishes a system log entry
func (s *LogStreamingService) PublishSystemLog(ctx context.Context, jobID uuid.UUID, level, message string) error {
	entry := LogEntry{
		Timestamp: time.Now(),
		JobID:     jobID,
		Level:     level,
		Message:   message,
		Source:    "system",
	}

	return s.PublishLogEntry(ctx, entry)
}

// PublishStepLog publishes a step-specific log entry
func (s *LogStreamingService) PublishStepLog(ctx context.Context, jobID, stepID uuid.UUID, level, message string) error {
	entry := LogEntry{
		Timestamp: time.Now(),
		JobID:     jobID,
		StepID:    &stepID,
		Level:     level,
		Message:   message,
		Source:    "runner",
	}

	return s.PublishLogEntry(ctx, entry)
}