package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/a5c-ai/hub/internal/models"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const (
	// Redis keys for job queue
	redisJobQueueKey           = "hub:job_queue"           // Priority queue for jobs
	redisJobQueuePendingKey    = "hub:job_queue:pending"   // List of pending jobs
	redisJobQueueProcessingKey = "hub:job_queue:processing" // Set of processing jobs
	redisJobDataKey            = "hub:job_data:"           // Hash prefix for job data
	redisJobStatusKey          = "hub:job_status:"         // Key prefix for job status
	redisJobMetricsKey         = "hub:job_metrics"         // Hash for queue metrics
)

// JobQueueService handles job queuing and scheduling
type JobQueueService struct {
	db           *gorm.DB
	redisService *RedisService
	logger       *logrus.Logger
}

// NewJobQueueService creates a new job queue service
func NewJobQueueService(db *gorm.DB, redisService *RedisService, logger *logrus.Logger) *JobQueueService {
	return &JobQueueService{
		db:           db,
		redisService: redisService,
		logger:       logger,
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

	// Use Redis if available, otherwise fallback to database
	if s.redisService.IsEnabled() {
		return s.enqueueJobRedis(ctx, &queueItem, job)
	} else {
		return s.enqueueJobDatabase(ctx, &queueItem, job)
	}
}

// enqueueJobRedis enqueues a job using Redis
func (s *JobQueueService) enqueueJobRedis(ctx context.Context, queueItem *JobQueueItem, job *models.Job) error {
	// Serialize job data
	queueItemData, err := json.Marshal(queueItem)
	if err != nil {
		return fmt.Errorf("failed to marshal queue item: %w", err)
	}

	// Use Redis transaction to ensure atomicity
	pipe := s.redisService.TxPipeline()

	// Add job to priority queue (sorted set) with priority as score
	pipe.ZAdd(ctx, redisJobQueueKey, redis.Z{
		Score:  float64(queueItem.Priority),
		Member: queueItem.JobID.String(),
	})

	// Store job data in hash
	jobDataKey := redisJobDataKey + queueItem.JobID.String()
	pipe.HSet(ctx, jobDataKey, map[string]interface{}{
		"data":           string(queueItemData),
		"status":         "pending",
		"queued_at":      queueItem.QueuedAt.Unix(),
		"workflow_run_id": queueItem.WorkflowRunID.String(),
		"priority":       queueItem.Priority,
		"runner_labels":  fmt.Sprintf("%v", queueItem.RunnerLabels),
	})

	// Set expiration for job data (24 hours)
	pipe.Expire(ctx, jobDataKey, 24*time.Hour)

	// Update metrics
	pipe.HIncrBy(ctx, redisJobMetricsKey, "total_queued", 1)
	pipe.HIncrBy(ctx, redisJobMetricsKey, "pending", 1)

	// Execute transaction
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to enqueue job in Redis: %w", err)
	}

	// Update job status in database
	if err := s.db.WithContext(ctx).Model(job).Updates(map[string]interface{}{
		"status":     models.JobStatusQueued,
		"updated_at": time.Now(),
	}).Error; err != nil {
		return fmt.Errorf("failed to update job status in database: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"job_id":          queueItem.JobID,
		"workflow_run_id": queueItem.WorkflowRunID,
		"priority":        queueItem.Priority,
		"runner_labels":   queueItem.RunnerLabels,
		"storage":         "redis",
	}).Info("Job enqueued")

	return nil
}

// enqueueJobDatabase enqueues a job using database (fallback)
func (s *JobQueueService) enqueueJobDatabase(ctx context.Context, queueItem *JobQueueItem, job *models.Job) error {
	queueItemData, err := json.Marshal(queueItem)
	if err != nil {
		return fmt.Errorf("failed to marshal queue item: %w", err)
	}

	// Create a simple queue table entry
	if err := s.db.WithContext(ctx).Exec(`
		INSERT INTO job_queue (job_id, workflow_run_id, priority, data, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, job.ID, job.WorkflowRunID, queueItem.Priority, string(queueItemData), time.Now()).Error; err != nil {
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
		"job_id":          queueItem.JobID,
		"workflow_run_id": queueItem.WorkflowRunID,
		"priority":        queueItem.Priority,
		"runner_labels":   queueItem.RunnerLabels,
		"storage":         "database",
	}).Info("Job enqueued")

	return nil
}

// DequeueJob gets the next job from the queue for a specific runner
func (s *JobQueueService) DequeueJob(ctx context.Context, runnerLabels []string) (*JobQueueItem, error) {
	// Use Redis if available, otherwise fallback to database
	if s.redisService.IsEnabled() {
		return s.dequeueJobRedis(ctx, runnerLabels)
	} else {
		return s.dequeueJobDatabase(ctx, runnerLabels)
	}
}

// dequeueJobRedis dequeues a job using Redis
func (s *JobQueueService) dequeueJobRedis(ctx context.Context, runnerLabels []string) (*JobQueueItem, error) {
	for {
		// Get the highest priority job from the sorted set
		results, err := s.redisService.ZPopMax(ctx, redisJobQueueKey, 1)
		if err != nil {
			if err == redis.Nil {
				return nil, nil // No jobs available
			}
			return nil, fmt.Errorf("failed to pop job from Redis queue: %w", err)
		}

		if len(results) == 0 {
			return nil, nil // No jobs available
		}

		jobIDStr := results[0].Member.(string)
		if _, err := uuid.Parse(jobIDStr); err != nil {
			s.logger.WithError(err).WithField("job_id", jobIDStr).Warn("Invalid job ID in queue")
			continue
		}

		// Get job data
		jobDataKey := redisJobDataKey + jobIDStr
		jobData, err := s.redisService.HGetAll(ctx, jobDataKey)
		if err != nil {
			s.logger.WithError(err).WithField("job_id", jobIDStr).Warn("Failed to get job data")
			continue
		}

		if len(jobData) == 0 {
			s.logger.WithField("job_id", jobIDStr).Warn("Job data not found")
			continue
		}

		// Check if job is still pending
		if jobData["status"] != "pending" {
			s.logger.WithField("job_id", jobIDStr).Warn("Job is not in pending state")
			continue
		}

		// Deserialize job data
		var queueItem JobQueueItem
		if err := json.Unmarshal([]byte(jobData["data"]), &queueItem); err != nil {
			s.logger.WithError(err).WithField("job_id", jobIDStr).Warn("Failed to unmarshal job data")
			continue
		}

		// Check if runner can handle this job
		if !s.runnerCanHandleJob(runnerLabels, queueItem.RunnerLabels) {
			// Re-queue the job (put it back in the queue)
			if err := s.redisService.ZAdd(ctx, redisJobQueueKey, redis.Z{
				Score:  float64(queueItem.Priority),
				Member: jobIDStr,
			}); err != nil {
				s.logger.WithError(err).WithField("job_id", jobIDStr).Error("Failed to re-queue incompatible job")
			}
			return nil, nil // No compatible jobs for this runner
		}

		// Mark job as processing using Redis transaction
		pipe := s.redisService.TxPipeline()
		pipe.HSet(ctx, jobDataKey, "status", "processing")
		pipe.HSet(ctx, jobDataKey, "processing_started_at", time.Now().Unix())
		pipe.HIncrBy(ctx, redisJobMetricsKey, "pending", -1)
		pipe.HIncrBy(ctx, redisJobMetricsKey, "processing", 1)

		_, err = pipe.Exec(ctx)
		if err != nil {
			s.logger.WithError(err).WithField("job_id", jobIDStr).Error("Failed to mark job as processing")
			// Re-queue the job
			if err := s.redisService.ZAdd(ctx, redisJobQueueKey, redis.Z{
				Score:  float64(queueItem.Priority),
				Member: jobIDStr,
			}); err != nil {
				s.logger.WithError(err).WithField("job_id", jobIDStr).Error("Failed to re-queue job after processing failure")
			}
			continue
		}

		s.logger.WithFields(logrus.Fields{
			"job_id":          queueItem.JobID,
			"workflow_run_id": queueItem.WorkflowRunID,
			"runner_labels":   runnerLabels,
			"priority":        queueItem.Priority,
			"storage":         "redis",
		}).Info("Job dequeued")

		return &queueItem, nil
	}
}

// dequeueJobDatabase dequeues a job using database (fallback)
func (s *JobQueueService) dequeueJobDatabase(ctx context.Context, runnerLabels []string) (*JobQueueItem, error) {
	var queueEntry struct {
		ID   uuid.UUID `gorm:"column:id"`
		Data string    `gorm:"column:data"`
	}

	// Find the highest priority job that matches runner labels
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

	// Check if we actually got a result
	if queueEntry.Data == "" {
		return nil, nil // No jobs available
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
		"storage":         "database",
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

	// Remove from queue (Redis or database)
	if s.redisService.IsEnabled() {
		return s.completeJobRedis(ctx, jobID, success, output)
	} else {
		return s.completeJobDatabase(ctx, jobID, success, output)
	}
}

// completeJobRedis completes a job using Redis
func (s *JobQueueService) completeJobRedis(ctx context.Context, jobID uuid.UUID, success bool, output string) error {
	jobIDStr := jobID.String()
	jobDataKey := redisJobDataKey + jobIDStr

	// Use Redis transaction
	pipe := s.redisService.TxPipeline()

	// Update job status to completed
	pipe.HSet(ctx, jobDataKey, map[string]interface{}{
		"status":       "completed",
		"success":      success,
		"output":       output,
		"completed_at": time.Now().Unix(),
	})

	// Update metrics
	pipe.HIncrBy(ctx, redisJobMetricsKey, "processing", -1)
	if success {
		pipe.HIncrBy(ctx, redisJobMetricsKey, "completed", 1)
	} else {
		pipe.HIncrBy(ctx, redisJobMetricsKey, "failed", 1)
	}

	// Set shorter expiration for completed job data (1 hour)
	pipe.Expire(ctx, jobDataKey, 1*time.Hour)

	// Execute transaction
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to complete job in Redis: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"job_id":  jobID,
		"success": success,
		"storage": "redis",
	}).Info("Job completed")

	return nil
}

// completeJobDatabase completes a job using database (fallback)
func (s *JobQueueService) completeJobDatabase(ctx context.Context, jobID uuid.UUID, success bool, output string) error {
	// Remove from queue
	if err := s.db.WithContext(ctx).Exec(`
		DELETE FROM job_queue WHERE job_id = ?
	`, jobID).Error; err != nil {
		return fmt.Errorf("failed to remove job from queue: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"job_id":  jobID,
		"success": success,
		"storage": "database",
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
	if s.redisService.IsEnabled() {
		return s.getQueueStatusRedis(ctx)
	} else {
		return s.getQueueStatusDatabase(ctx)
	}
}

// getQueueStatusRedis gets queue status from Redis
func (s *JobQueueService) getQueueStatusRedis(ctx context.Context) (map[string]interface{}, error) {
	// Get metrics from Redis
	metrics, err := s.redisService.HGetAll(ctx, redisJobMetricsKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get queue metrics from Redis: %w", err)
	}

	// Get current queue size (pending jobs)
	queueSize, err := s.redisService.ZCard(ctx, redisJobQueueKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get queue size from Redis: %w", err)
	}

	// Parse metrics
	pending := queueSize
	processing, _ := strconv.ParseInt(metrics["processing"], 10, 64)
	completed, _ := strconv.ParseInt(metrics["completed"], 10, 64)
	failed, _ := strconv.ParseInt(metrics["failed"], 10, 64)
	totalQueued, _ := strconv.ParseInt(metrics["total_queued"], 10, 64)

	return map[string]interface{}{
		"pending":      pending,
		"processing":   processing,
		"completed":    completed,
		"failed":       failed,
		"total_queued": totalQueued,
		"total":        pending + processing,
		"timestamp":    time.Now(),
		"storage":      "redis",
	}, nil
}

// getQueueStatusDatabase gets queue status from database (fallback)
func (s *JobQueueService) getQueueStatusDatabase(ctx context.Context) (map[string]interface{}, error) {
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
		"storage":    "database",
	}, nil
}

// ListQueuedJobs returns a list of jobs in the queue
func (s *JobQueueService) ListQueuedJobs(ctx context.Context, limit int) ([]JobQueueItem, error) {
	if limit <= 0 {
		limit = 50
	}

	if s.redisService.IsEnabled() {
		return s.listQueuedJobsRedis(ctx, limit)
	} else {
		return s.listQueuedJobsDatabase(ctx, limit)
	}
}

// listQueuedJobsRedis lists queued jobs from Redis
func (s *JobQueueService) listQueuedJobsRedis(ctx context.Context, limit int) ([]JobQueueItem, error) {
	// Get jobs from the priority queue (highest priority first)
	jobIDs, err := s.redisService.ZRangeWithScores(ctx, redisJobQueueKey, -int64(limit), -1)
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs from Redis queue: %w", err)
	}

	var jobs []JobQueueItem
	for _, item := range jobIDs {
		jobIDStr := item.Member.(string)
		jobDataKey := redisJobDataKey + jobIDStr

		// Get job data
		jobData, err := s.redisService.HGetAll(ctx, jobDataKey)
		if err != nil {
			s.logger.WithError(err).WithField("job_id", jobIDStr).Warn("Failed to get job data")
			continue
		}

		if len(jobData) == 0 || jobData["status"] != "pending" {
			continue
		}

		// Deserialize job data
		var queueItem JobQueueItem
		if err := json.Unmarshal([]byte(jobData["data"]), &queueItem); err != nil {
			s.logger.WithError(err).WithField("job_id", jobIDStr).Warn("Failed to unmarshal job data")
			continue
		}

		jobs = append(jobs, queueItem)
	}

	return jobs, nil
}

// listQueuedJobsDatabase lists queued jobs from database (fallback)
func (s *JobQueueService) listQueuedJobsDatabase(ctx context.Context, limit int) ([]JobQueueItem, error) {
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

// RecoverStuckJobs recovers jobs that have been processing for too long
func (s *JobQueueService) RecoverStuckJobs(ctx context.Context, timeout time.Duration) error {
	if s.redisService.IsEnabled() {
		return s.recoverStuckJobsRedis(ctx, timeout)
	} else {
		return s.recoverStuckJobsDatabase(ctx, timeout)
	}
}

// recoverStuckJobsRedis recovers stuck jobs from Redis
func (s *JobQueueService) recoverStuckJobsRedis(ctx context.Context, timeout time.Duration) error {
	// This would typically involve scanning for jobs that have been in "processing" state
	// for longer than the timeout period and re-queuing them
	// For now, we'll implement a basic version
	s.logger.Info("Redis job recovery mechanism not fully implemented yet")
	return nil
}

// recoverStuckJobsDatabase recovers stuck jobs from database
func (s *JobQueueService) recoverStuckJobsDatabase(ctx context.Context, timeout time.Duration) error {
	cutoff := time.Now().Add(-timeout)
	
	// Find jobs that have been processing for too long
	var stuckJobs []struct {
		JobID uuid.UUID `gorm:"column:job_id"`
		Data  string    `gorm:"column:data"`
	}

	err := s.db.WithContext(ctx).Raw(`
		SELECT job_id, data FROM job_queue 
		WHERE status = 'processing' AND updated_at < ?
	`, cutoff).Scan(&stuckJobs).Error

	if err != nil {
		return fmt.Errorf("failed to find stuck jobs: %w", err)
	}

	for _, stuckJob := range stuckJobs {
		// Reset job status to pending
		if err := s.db.WithContext(ctx).Exec(`
			UPDATE job_queue SET status = 'pending', updated_at = ? WHERE job_id = ?
		`, time.Now(), stuckJob.JobID).Error; err != nil {
			s.logger.WithError(err).WithField("job_id", stuckJob.JobID).Error("Failed to recover stuck job")
			continue
		}

		s.logger.WithField("job_id", stuckJob.JobID).Info("Recovered stuck job")
	}

	return nil
}

// CleanupCompletedJobs removes old completed job data
func (s *JobQueueService) CleanupCompletedJobs(ctx context.Context, maxAge time.Duration) error {
	if s.redisService.IsEnabled() {
		return s.cleanupCompletedJobsRedis(ctx, maxAge)
	} else {
		return s.cleanupCompletedJobsDatabase(ctx, maxAge)
	}
}

// cleanupCompletedJobsRedis cleans up old completed jobs from Redis
func (s *JobQueueService) cleanupCompletedJobsRedis(ctx context.Context, maxAge time.Duration) error {
	// Redis TTL handles most cleanup automatically, but we could implement
	// additional cleanup logic here if needed
	s.logger.Info("Redis cleanup relies on TTL, no additional cleanup needed")
	return nil
}

// cleanupCompletedJobsDatabase cleans up old completed jobs from database
func (s *JobQueueService) cleanupCompletedJobsDatabase(ctx context.Context, maxAge time.Duration) error {
	cutoff := time.Now().Add(-maxAge)

	result := s.db.WithContext(ctx).Exec(`
		DELETE FROM job_queue 
		WHERE status IN ('completed', 'failed') AND updated_at < ?
	`, cutoff)

	if result.Error != nil {
		return fmt.Errorf("failed to cleanup completed jobs: %w", result.Error)
	}

	if result.RowsAffected > 0 {
		s.logger.WithField("rows_affected", result.RowsAffected).Info("Cleaned up old completed jobs")
	}

	return nil
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