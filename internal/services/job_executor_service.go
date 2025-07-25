package services

import (
	"context"
	"fmt"
	"time"

	"github.com/a5c-ai/hub/internal/models"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// JobExecutorService manages job execution workflow
type JobExecutorService struct {
	db              *gorm.DB
	jobQueueService *JobQueueService
	runnerService   *RunnerService
	logger          *logrus.Logger
}

// NewJobExecutorService creates a new job executor service
func NewJobExecutorService(db *gorm.DB, jobQueueService *JobQueueService, runnerService *RunnerService, logger *logrus.Logger) *JobExecutorService {
	return &JobExecutorService{
		db:              db,
		jobQueueService: jobQueueService,
		runnerService:   runnerService,
		logger:          logger,
	}
}

// ScheduleWorkflowJobs schedules all ready jobs from a workflow run
func (s *JobExecutorService) ScheduleWorkflowJobs(ctx context.Context, workflowRunID uuid.UUID) error {
	// Get all jobs for this workflow run
	var jobs []models.Job
	err := s.db.WithContext(ctx).
		Preload("WorkflowRun").
		Where("workflow_run_id = ?", workflowRunID).
		Find(&jobs).Error
	if err != nil {
		return fmt.Errorf("failed to get jobs for workflow run: %w", err)
	}

	// Find jobs that are ready to run (no pending dependencies)
	readyJobs := s.findReadyJobs(ctx, jobs)

	// Enqueue ready jobs
	for _, job := range readyJobs {
		if err := s.jobQueueService.EnqueueJob(ctx, &job); err != nil {
			s.logger.WithError(err).WithField("job_id", job.ID).
				Error("Failed to enqueue job")
			// Continue with other jobs
		}
	}

	return nil
}

// ProcessJobQueue continuously processes jobs from the queue
func (s *JobExecutorService) ProcessJobQueue(ctx context.Context) error {
	s.logger.Info("Starting job queue processor")

	ticker := time.NewTicker(5 * time.Second) // Check every 5 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Job queue processor stopped")
			return ctx.Err()
		case <-ticker.C:
			if err := s.processNextJob(ctx); err != nil {
				s.logger.WithError(err).Error("Failed to process job")
			}
		}
	}
}

// processNextJob processes the next available job in the queue
func (s *JobExecutorService) processNextJob(ctx context.Context) error {
	// Get list of online runners with their labels
	runners, _, err := s.runnerService.ListRunners(ctx, ListRunnersRequest{
		Status: &[]models.RunnerStatus{models.RunnerStatusOnline}[0],
		Limit:  100,
	})
	if err != nil {
		return fmt.Errorf("failed to get available runners: %w", err)
	}

	if len(runners) == 0 {
		return nil // No runners available
	}

	// Try to match jobs with available runners
	for _, runner := range runners {
		runnerLabels := s.extractLabelsFromRunner(&runner)
		
		// Try to dequeue a job for this runner
		queueItem, err := s.jobQueueService.DequeueJob(ctx, runnerLabels)
		if err != nil {
			return fmt.Errorf("failed to dequeue job: %w", err)
		}

		if queueItem == nil {
			continue // No compatible jobs for this runner
		}

		// Assign job to runner
		if err := s.runnerService.AssignJobToRunner(ctx, queueItem.JobID, runner.ID); err != nil {
			s.logger.WithError(err).WithFields(logrus.Fields{
				"job_id":    queueItem.JobID,
				"runner_id": runner.ID,
			}).Error("Failed to assign job to runner")
			
			// TODO: Re-queue the job
			continue
		}

		// Start job execution (in a separate goroutine for non-blocking)
		go func(jobID, runnerID uuid.UUID) {
			if err := s.executeJob(context.Background(), jobID, runnerID); err != nil {
				s.logger.WithError(err).WithFields(logrus.Fields{
					"job_id":    jobID,
					"runner_id": runnerID,
				}).Error("Job execution failed")
			}
		}(queueItem.JobID, runner.ID)

		// Only process one job per cycle to avoid overloading
		break
	}

	return nil
}

// executeJob executes a single job on a runner
func (s *JobExecutorService) executeJob(ctx context.Context, jobID, runnerID uuid.UUID) error {
	s.logger.WithFields(logrus.Fields{
		"job_id":    jobID,
		"runner_id": runnerID,
	}).Info("Starting job execution")

	// Get job details with all steps
	var job models.Job
	err := s.db.WithContext(ctx).
		Preload("WorkflowRun").
		Preload("Steps").
		First(&job, "id = ?", jobID).Error
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}

	defer func() {
		// Always release the runner when job completes
		if err := s.runnerService.ReleaseRunner(ctx, runnerID); err != nil {
			s.logger.WithError(err).WithField("runner_id", runnerID).
				Error("Failed to release runner")
		}
	}()

	// Execute steps sequentially
	success := true
	for _, step := range job.Steps {
		stepSuccess, err := s.executeStep(ctx, &step, runnerID)
		if err != nil {
			s.logger.WithError(err).WithFields(logrus.Fields{
				"job_id":  jobID,
				"step_id": step.ID,
			}).Error("Step execution failed")
			success = false
			break
		}

		if !stepSuccess {
			s.logger.WithFields(logrus.Fields{
				"job_id":  jobID,
				"step_id": step.ID,
			}).Info("Step failed")
			success = false
			break
		}
	}

	// Complete the job
	output := fmt.Sprintf("Job completed with %d steps", len(job.Steps))
	if err := s.jobQueueService.CompleteJob(ctx, jobID, success, output); err != nil {
		return fmt.Errorf("failed to complete job: %w", err)
	}

	// Check if this completes the workflow run
	if err := s.checkWorkflowRunCompletion(ctx, job.WorkflowRunID); err != nil {
		s.logger.WithError(err).WithField("workflow_run_id", job.WorkflowRunID).
			Error("Failed to check workflow run completion")
	}

	s.logger.WithFields(logrus.Fields{
		"job_id":    jobID,
		"runner_id": runnerID,
		"success":   success,
	}).Info("Job execution completed")

	return nil
}

// executeStep executes a single step
func (s *JobExecutorService) executeStep(ctx context.Context, step *models.Step, runnerID uuid.UUID) (bool, error) {
	s.logger.WithFields(logrus.Fields{
		"step_id":   step.ID,
		"step_name": step.Name,
		"runner_id": runnerID,
	}).Info("Executing step")

	// Update step status to in_progress
	now := time.Now()
	if err := s.db.WithContext(ctx).Model(step).Updates(map[string]interface{}{
		"status":     models.StepStatusInProgress,
		"started_at": &now,
		"updated_at": now,
	}).Error; err != nil {
		return false, fmt.Errorf("failed to update step status: %w", err)
	}

	// Simulate step execution (in a real implementation, this would communicate with the runner)
	stepSuccess := s.simulateStepExecution(step)
	
	// Update step with completion status
	completedAt := time.Now()
	conclusion := models.StepConclusionSuccess
	output := "Step completed successfully"
	
	if !stepSuccess {
		conclusion = models.StepConclusionFailure
		output = "Step failed"
	}

	if err := s.db.WithContext(ctx).Model(step).Updates(map[string]interface{}{
		"status":       models.StepStatusCompleted,
		"conclusion":   conclusion,
		"output":       output,
		"completed_at": &completedAt,
		"updated_at":   completedAt,
	}).Error; err != nil {
		return false, fmt.Errorf("failed to update step completion: %w", err)
	}

	return stepSuccess, nil
}

// simulateStepExecution simulates step execution (placeholder for real implementation)
func (s *JobExecutorService) simulateStepExecution(step *models.Step) bool {
	// In a real implementation, this would:
	// 1. Send step details to the runner
	// 2. Monitor step execution
	// 3. Collect logs and artifacts
	// 4. Return actual success/failure status

	// For now, simulate success for most steps
	if step.Action == "actions/checkout" || step.Action == "shell" {
		return true
	}

	// Simulate occasional failures for demonstration
	return step.Number != 3 // Make step 3 fail for testing
}

// findReadyJobs finds jobs that have all their dependencies completed
func (s *JobExecutorService) findReadyJobs(ctx context.Context, jobs []models.Job) []models.Job {
	var readyJobs []models.Job
	
	for _, job := range jobs {
		if job.Status != models.JobStatusQueued {
			continue // Job already processed or in progress
		}

		// Check if job dependencies are satisfied
		if s.areDependenciesSatisfied(job, jobs) {
			readyJobs = append(readyJobs, job)
		}
	}

	return readyJobs
}

// areDependenciesSatisfied checks if all job dependencies are completed successfully
func (s *JobExecutorService) areDependenciesSatisfied(job models.Job, allJobs []models.Job) bool {
	// Extract needs from job (this is a simplified implementation)
	if job.Needs == nil {
		return true // No dependencies
	}

	// In a real implementation, parse the needs field properly
	// For now, assume no dependencies or all are satisfied
	return true
}

// checkWorkflowRunCompletion checks if a workflow run is complete and updates its status
func (s *JobExecutorService) checkWorkflowRunCompletion(ctx context.Context, workflowRunID uuid.UUID) error {
	// Get all jobs for this workflow run
	var jobs []models.Job
	err := s.db.WithContext(ctx).
		Where("workflow_run_id = ?", workflowRunID).
		Find(&jobs).Error
	if err != nil {
		return fmt.Errorf("failed to get jobs for workflow run: %w", err)
	}

	// Check if all jobs are completed
	allCompleted := true
	allSuccessful := true
	
	for _, job := range jobs {
		if job.Status != models.JobStatusCompleted {
			allCompleted = false
			break
		}
		
		if job.Conclusion != nil && *job.Conclusion != models.JobConclusionSuccess {
			allSuccessful = false
		}
	}

	if !allCompleted {
		return nil // Workflow run still in progress
	}

	// Update workflow run status
	conclusion := models.WorkflowRunConclusionSuccess
	if !allSuccessful {
		conclusion = models.WorkflowRunConclusionFailure
	}

	completedAt := time.Now()
	if err := s.db.WithContext(ctx).Model(&models.WorkflowRun{}).
		Where("id = ?", workflowRunID).
		Updates(map[string]interface{}{
			"status":       models.WorkflowRunStatusCompleted,
			"conclusion":   conclusion,
			"completed_at": &completedAt,
			"updated_at":   completedAt,
		}).Error; err != nil {
		return fmt.Errorf("failed to update workflow run completion: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"workflow_run_id": workflowRunID,
		"conclusion":      conclusion,
		"jobs_count":      len(jobs),
	}).Info("Workflow run completed")

	return nil
}

// extractLabelsFromRunner extracts labels from a runner model
func (s *JobExecutorService) extractLabelsFromRunner(runner *models.Runner) []string {
	if runner.Labels == nil {
		return []string{}
	}

	// Convert interface{} to []string
	labels := []string{}
	switch v := runner.Labels.(type) {
	case []string:
		labels = v
	case []interface{}:
		for _, label := range v {
			if labelStr, ok := label.(string); ok {
				labels = append(labels, labelStr)
			}
		}
	}

	return labels
}