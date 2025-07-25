package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/a5c-ai/hub/internal/models"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// JobQueueService handles job queuing and scheduling
type JobQueueService struct {
	db     *gorm.DB
	logger *logrus.Logger
	// TODO: Add Redis client when Redis is available
	// redis  *redis.Client
}

// NewJobQueueService creates a new job queue service
func NewJobQueueService(db *gorm.DB, logger *logrus.Logger) *JobQueueService {
	return &JobQueueService{
		db:     db,
		logger: logger,
	}
}

// JobQueueItem represents a job in the queue
type JobQueueItem struct {
	JobID        uuid.UUID           `json:"job_id"`
	WorkflowRunID uuid.UUID          `json:"workflow_run_id"`
	Priority     int                 `json:"priority"`
	RunnerLabels []string            `json:"runner_labels"`
	QueuedAt     time.Time           `json:"queued_at"`
	Timeout      time.Duration       `json:"timeout"`
	RetryCount   int                 `json:"retry_count"`
	MaxRetries   int                 `json:"max_retries"`
}

// EnqueueJob adds a job to the queue
func (s *JobQueueService) EnqueueJob(ctx context.Context, job *models.Job) error {
	// Calculate priority based on job properties
	priority := s.calculateJobPriority(job)

	// Extract runner labels from job requirements
	runnerLabels := s.extractRunnerLabels(job)

	queueItem := JobQueueItem{
		JobID:        job.ID,
		WorkflowRunID: job.WorkflowRunID,
		Priority:     priority,
		RunnerLabels: runnerLabels,
		QueuedAt:     time.Now(),
		Timeout:      30 * time.Minute, // Default timeout
		RetryCount:   0,
		MaxRetries:   3,
	}

	// For now, store queue items in database
	// TODO: Use Redis for better performance and persistence
	queueItemData, err := json.Marshal(queueItem)
	if err != nil {
		return fmt.Errorf("failed to marshal queue item: %w", err)
	}

	// Create a simple queue table entry
	if err := s.db.WithContext(ctx).Exec(`
		INSERT INTO job_queue (job_id, workflow_run_id, priority, data, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, job.ID, job.WorkflowRunID, priority, string(queueItemData), time.Now()).Error; err != nil {
		return fmt.Errorf("failed to enqueue job: %w", err)
	}

	// Update job status to queued
	if err := s.db.WithContext(ctx).Model(job).Updates(map[string]interface{}{
		"status":     models.JobStatusQueued,
		"updated_at": time.Now(),
	}).Error; err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"job_id":          job.ID,
		"workflow_run_id": job.WorkflowRunID,
		"priority":        priority,
		"runner_labels":   runnerLabels,
	}).Info("Job enqueued")

	return nil
}

// DequeueJob gets the next job from the queue for a specific runner
func (s *JobQueueService) DequeueJob(ctx context.Context, runnerLabels []string) (*JobQueueItem, error) {
	// For now, use database queries
	// TODO: Use Redis BLPOP or similar for better performance

	var queueEntry struct {
		ID   uuid.UUID `gorm:"column:id"`
		Data string    `gorm:"column:data"`
	}

	// Find the highest priority job that matches runner labels
	// This is a simplified implementation - in production, use proper Redis queues
	err := s.db.WithContext(ctx).Raw(`
		SELECT id, data FROM job_queue 
		WHERE status = 'pending'
		ORDER BY priority DESC, created_at ASC
		LIMIT 1
	`).Scan(&queueEntry).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // No jobs available
		}
		return nil, fmt.Errorf("failed to dequeue job: %w", err)
	}

	var queueItem JobQueueItem
	if err := json.Unmarshal([]byte(queueEntry.Data), &queueItem); err != nil {
		return nil, fmt.Errorf("failed to unmarshal queue item: %w", err)
	}

	// Check if runner can handle this job
	if !s.runnerCanHandleJob(runnerLabels, queueItem.RunnerLabels) {
		return nil, nil // No compatible jobs
	}

	// Mark queue entry as processing
	if err := s.db.WithContext(ctx).Exec(`
		UPDATE job_queue SET status = 'processing', updated_at = ? WHERE id = ?
	`, time.Now(), queueEntry.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to mark queue entry as processing: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"job_id":          queueItem.JobID,
		"workflow_run_id": queueItem.WorkflowRunID,
		"runner_labels":   runnerLabels,
	}).Info("Job dequeued")

	return &queueItem, nil
}

// CompleteJob marks a job as completed and removes it from the queue
func (s *JobQueueService) CompleteJob(ctx context.Context, jobID uuid.UUID, success bool, output string) error {
	// Update job status in database
	conclusion := models.JobConclusionSuccess
	if !success {
		conclusion = models.JobConclusionFailure
	}

	now := time.Now()
	if err := s.db.WithContext(ctx).Model(&models.Job{}).Where("id = ?", jobID).Updates(map[string]interface{}{
		"status":       models.JobStatusCompleted,
		"conclusion":   conclusion,
		"completed_at": &now,
		"updated_at":   now,
	}).Error; err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}

	// Remove from queue
	if err := s.db.WithContext(ctx).Exec(`
		DELETE FROM job_queue WHERE job_id = ?
	`, jobID).Error; err != nil {
		return fmt.Errorf("failed to remove job from queue: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"job_id": jobID,
		"success": success,
	}).Info("Job completed")

	return nil
}

// RetryJob re-queues a failed job for retry
func (s *JobQueueService) RetryJob(ctx context.Context, jobID uuid.UUID) error {
	// Get the job
	var job models.Job
	if err := s.db.WithContext(ctx).First(&job, "id = ?", jobID).Error; err != nil {
		return fmt.Errorf("job not found: %w", err)
	}

	// Reset job status
	if err := s.db.WithContext(ctx).Model(&job).Updates(map[string]interface{}{
		"status":       models.JobStatusQueued,
		"conclusion":   nil,
		"completed_at": nil,
		"updated_at":   time.Now(),
	}).Error; err != nil {
		return fmt.Errorf("failed to reset job status: %w", err)
	}

	// Re-enqueue the job
	return s.EnqueueJob(ctx, &job)
}

// GetQueueStatus returns the current queue status
func (s *JobQueueService) GetQueueStatus(ctx context.Context) (map[string]interface{}, error) {
	var stats struct {
		Pending    int64 `gorm:"column:pending"`
		Processing int64 `gorm:"column:processing"`
		Total      int64 `gorm:"column:total"`
	}

	err := s.db.WithContext(ctx).Raw(`
		SELECT 
			COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending,
			COUNT(CASE WHEN status = 'processing' THEN 1 END) as processing,
			COUNT(*) as total
		FROM job_queue
	`).Scan(&stats).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get queue status: %w", err)
	}

	return map[string]interface{}{
		"pending":    stats.Pending,
		"processing": stats.Processing,
		"total":      stats.Total,
		"timestamp":  time.Now(),
	}, nil
}

// ListQueuedJobs returns a list of jobs in the queue
func (s *JobQueueService) ListQueuedJobs(ctx context.Context, limit int) ([]JobQueueItem, error) {
	if limit <= 0 {
		limit = 50
	}

	var queueEntries []struct {
		Data string `gorm:"column:data"`
	}

	err := s.db.WithContext(ctx).Raw(`
		SELECT data FROM job_queue 
		WHERE status = 'pending'
		ORDER BY priority DESC, created_at ASC
		LIMIT ?
	`, limit).Scan(&queueEntries).Error

	if err != nil {
		return nil, fmt.Errorf("failed to list queued jobs: %w", err)
	}

	var jobs []JobQueueItem
	for _, entry := range queueEntries {
		var queueItem JobQueueItem
		if err := json.Unmarshal([]byte(entry.Data), &queueItem); err != nil {
			s.logger.WithError(err).Warn("Failed to unmarshal queue item")
			continue
		}
		jobs = append(jobs, queueItem)
	}

	return jobs, nil
}

// Helper methods

// calculateJobPriority calculates job priority based on various factors
func (s *JobQueueService) calculateJobPriority(job *models.Job) int {
	priority := 100 // Base priority

	// Increase priority for workflow dispatch events (manual triggers)
	if job.WorkflowRun.Event == "workflow_dispatch" {
		priority += 50
	}

	// Increase priority for main branch builds
	if job.WorkflowRun.HeadBranch != nil && *job.WorkflowRun.HeadBranch == "main" {
		priority += 30
	}

	// Increase priority for release builds
	if job.WorkflowRun.Event == "release" {
		priority += 40
	}

	// TODO: Add more priority rules based on organization settings, user tiers, etc.

	return priority
}

// extractRunnerLabels extracts required runner labels from job configuration
func (s *JobQueueService) extractRunnerLabels(job *models.Job) []string {
	// TODO: Parse job configuration to extract runs-on requirements
	// For now, return default labels
	return []string{"ubuntu-latest"}
}

// runnerCanHandleJob checks if a runner with given labels can handle a job
func (s *JobQueueService) runnerCanHandleJob(runnerLabels, requiredLabels []string) bool {
	// Simple implementation: runner must have all required labels
	labelMap := make(map[string]bool)
	for _, label := range runnerLabels {
		labelMap[label] = true
	}

	for _, required := range requiredLabels {
		if !labelMap[required] {
			return false
		}
	}

	return true
}

// CreateQueueTable creates the job queue table if it doesn't exist
func (s *JobQueueService) CreateQueueTable(ctx context.Context) error {
	return s.db.WithContext(ctx).Exec(`
		CREATE TABLE IF NOT EXISTS job_queue (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			job_id UUID NOT NULL,
			workflow_run_id UUID NOT NULL,
			priority INTEGER NOT NULL DEFAULT 100,
			status VARCHAR(50) NOT NULL DEFAULT 'pending',
			data JSONB NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			
			INDEX idx_job_queue_priority_created (priority DESC, created_at ASC),
			INDEX idx_job_queue_status (status),
			INDEX idx_job_queue_job_id (job_id)
		)
	`).Error
}