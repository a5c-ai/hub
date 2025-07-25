package services

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/a5c-ai/hub/internal/actions"
	"github.com/a5c-ai/hub/internal/models"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// WorkflowService handles workflow operations
type WorkflowService struct {
	db            *gorm.DB
	parser        *actions.WorkflowParser
	logger        *logrus.Logger
	jobExecutor   *JobExecutorService
	secretService *SecretService
}

// NewWorkflowService creates a new workflow service
func NewWorkflowService(db *gorm.DB, logger *logrus.Logger) *WorkflowService {
	return &WorkflowService{
		db:     db,
		parser: actions.NewWorkflowParser(),
		logger: logger,
	}
}

// SetJobExecutor sets the job executor (used to avoid circular dependencies)
func (s *WorkflowService) SetJobExecutor(jobExecutor *JobExecutorService) {
	s.jobExecutor = jobExecutor
}

// SetSecretService sets the secret service (used to avoid circular dependencies)
func (s *WorkflowService) SetSecretService(secretService *SecretService) {
	s.secretService = secretService
}

// CreateWorkflowRequest represents a request to create a workflow
type CreateWorkflowRequest struct {
	RepositoryID uuid.UUID `json:"repository_id" binding:"required"`
	Name         string    `json:"name" binding:"required"`
	Path         string    `json:"path" binding:"required"`
	Content      string    `json:"content" binding:"required"`
	Enabled      *bool     `json:"enabled,omitempty"`
}

// UpdateWorkflowRequest represents a request to update a workflow
type UpdateWorkflowRequest struct {
	Name    *string `json:"name,omitempty"`
	Content *string `json:"content,omitempty"`
	Enabled *bool   `json:"enabled,omitempty"`
}

// CreateWorkflowRunRequest represents a request to create a workflow run
type CreateWorkflowRunRequest struct {
	WorkflowID   uuid.UUID   `json:"workflow_id" binding:"required"`
	Event        string      `json:"event" binding:"required"`
	HeadSHA      string      `json:"head_sha" binding:"required"`
	HeadBranch   *string     `json:"head_branch,omitempty"`
	EventPayload interface{} `json:"event_payload,omitempty"`
	ActorID      *uuid.UUID  `json:"actor_id,omitempty"`
}

// ListWorkflowsRequest represents a request to list workflows
type ListWorkflowsRequest struct {
	RepositoryID uuid.UUID `json:"repository_id"`
	Enabled      *bool     `json:"enabled,omitempty"`
	Limit        int       `json:"limit,omitempty"`
	Offset       int       `json:"offset,omitempty"`
}

// ListWorkflowRunsRequest represents a request to list workflow runs
type ListWorkflowRunsRequest struct {
	RepositoryID uuid.UUID                     `json:"repository_id"`
	WorkflowID   *uuid.UUID                    `json:"workflow_id,omitempty"`
	Status       *models.WorkflowRunStatus     `json:"status,omitempty"`
	Conclusion   *models.WorkflowRunConclusion `json:"conclusion,omitempty"`
	Event        *string                       `json:"event,omitempty"`
	ActorID      *uuid.UUID                    `json:"actor_id,omitempty"`
	Limit        int                           `json:"limit,omitempty"`
	Offset       int                           `json:"offset,omitempty"`
}

// CreateWorkflow creates a new workflow
func (s *WorkflowService) CreateWorkflow(ctx context.Context, req CreateWorkflowRequest) (*models.Workflow, error) {
	// Validate the workflow YAML content
	workflowDef, err := s.parser.Parse(req.Content)
	if err != nil {
		return nil, fmt.Errorf("invalid workflow content: %w", err)
	}

	// Set default values
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	// Ensure path starts with .hub/workflows/
	if !strings.HasPrefix(req.Path, ".hub/workflows/") {
		req.Path = filepath.Join(".hub/workflows", req.Path)
	}

	// Ensure path ends with .yml or .yaml
	if !strings.HasSuffix(req.Path, ".yml") && !strings.HasSuffix(req.Path, ".yaml") {
		req.Path += ".yml"
	}

	workflow := &models.Workflow{
		RepositoryID: req.RepositoryID,
		Name:         workflowDef.Name, // Use name from YAML, not request
		Path:         req.Path,
		Content:      req.Content,
		Enabled:      enabled,
	}

	if err := s.db.WithContext(ctx).Create(workflow).Error; err != nil {
		return nil, fmt.Errorf("failed to create workflow: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"workflow_id":   workflow.ID,
		"repository_id": req.RepositoryID,
		"name":          workflow.Name,
		"path":          workflow.Path,
	}).Info("Created new workflow")

	return workflow, nil
}

// GetWorkflow retrieves a workflow by ID
func (s *WorkflowService) GetWorkflow(ctx context.Context, id uuid.UUID) (*models.Workflow, error) {
	var workflow models.Workflow
	err := s.db.WithContext(ctx).
		Preload("Repository").
		First(&workflow, "id = ?", id).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow: %w", err)
	}

	return &workflow, nil
}

// GetWorkflowByPath retrieves a workflow by repository and path
func (s *WorkflowService) GetWorkflowByPath(ctx context.Context, repositoryID uuid.UUID, path string) (*models.Workflow, error) {
	var workflow models.Workflow
	err := s.db.WithContext(ctx).
		Preload("Repository").
		First(&workflow, "repository_id = ? AND path = ?", repositoryID, path).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow: %w", err)
	}

	return &workflow, nil
}

// ListWorkflows lists workflows for a repository
func (s *WorkflowService) ListWorkflows(ctx context.Context, req ListWorkflowsRequest) ([]models.Workflow, int64, error) {
	query := s.db.WithContext(ctx).Model(&models.Workflow{}).
		Where("repository_id = ?", req.RepositoryID)

	if req.Enabled != nil {
		query = query.Where("enabled = ?", *req.Enabled)
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count workflows: %w", err)
	}

	// Apply pagination
	if req.Limit > 0 {
		query = query.Limit(req.Limit)
	} else {
		query = query.Limit(50) // Default limit
	}

	if req.Offset > 0 {
		query = query.Offset(req.Offset)
	}

	var workflows []models.Workflow
	err := query.Preload("Repository").
		Order("created_at DESC").
		Find(&workflows).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list workflows: %w", err)
	}

	return workflows, total, nil
}

// UpdateWorkflow updates a workflow
func (s *WorkflowService) UpdateWorkflow(ctx context.Context, id uuid.UUID, req UpdateWorkflowRequest) (*models.Workflow, error) {
	workflow, err := s.GetWorkflow(ctx, id)
	if err != nil {
		return nil, err
	}

	updates := make(map[string]interface{})

	if req.Name != nil {
		updates["name"] = *req.Name
	}

	if req.Content != nil {
		// Validate the new workflow content
		workflowDef, err := s.parser.Parse(*req.Content)
		if err != nil {
			return nil, fmt.Errorf("invalid workflow content: %w", err)
		}
		updates["content"] = *req.Content
		updates["name"] = workflowDef.Name // Update name from YAML
	}

	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}

	if len(updates) > 0 {
		updates["updated_at"] = time.Now()
		if err := s.db.WithContext(ctx).Model(workflow).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("failed to update workflow: %w", err)
		}
	}

	// Reload workflow with updates
	return s.GetWorkflow(ctx, id)
}

// DeleteWorkflow deletes a workflow
func (s *WorkflowService) DeleteWorkflow(ctx context.Context, id uuid.UUID) error {
	result := s.db.WithContext(ctx).Delete(&models.Workflow{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete workflow: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("workflow not found")
	}

	s.logger.WithField("workflow_id", id).Info("Deleted workflow")
	return nil
}

// CreateWorkflowRun creates a new workflow run
func (s *WorkflowService) CreateWorkflowRun(ctx context.Context, req CreateWorkflowRunRequest) (*models.WorkflowRun, error) {
	// Get the workflow
	workflow, err := s.GetWorkflow(ctx, req.WorkflowID)
	if err != nil {
		return nil, fmt.Errorf("workflow not found: %w", err)
	}

	if !workflow.Enabled {
		return nil, fmt.Errorf("workflow is disabled")
	}

	// Parse workflow to validate triggers
	workflowDef, err := s.parser.Parse(workflow.Content)
	if err != nil {
		return nil, fmt.Errorf("invalid workflow content: %w", err)
	}

	// Get the next run number for this repository
	var maxRunNumber int
	s.db.WithContext(ctx).Model(&models.WorkflowRun{}).
		Where("repository_id = ?", workflow.RepositoryID).
		Select("COALESCE(MAX(number), 0)").
		Scan(&maxRunNumber)

	run := &models.WorkflowRun{
		WorkflowID:   req.WorkflowID,
		RepositoryID: workflow.RepositoryID,
		Number:       maxRunNumber + 1,
		Status:       models.WorkflowRunStatusQueued,
		HeadSHA:      req.HeadSHA,
		HeadBranch:   req.HeadBranch,
		Event:        req.Event,
		EventPayload: req.EventPayload,
		ActorID:      req.ActorID,
	}

	if err := s.db.WithContext(ctx).Create(run).Error; err != nil {
		return nil, fmt.Errorf("failed to create workflow run: %w", err)
	}

	// Create jobs for this workflow run
	if err := s.createJobsForRun(ctx, run, workflowDef); err != nil {
		// If job creation fails, delete the run
		s.db.WithContext(ctx).Delete(run)
		return nil, fmt.Errorf("failed to create jobs for workflow run: %w", err)
	}

	// Schedule ready jobs for execution
	if s.jobExecutor != nil {
		if err := s.jobExecutor.ScheduleWorkflowJobs(ctx, run.ID); err != nil {
			s.logger.WithError(err).WithField("workflow_run_id", run.ID).
				Error("Failed to schedule workflow jobs")
		}
	}

	s.logger.WithFields(logrus.Fields{
		"workflow_run_id": run.ID,
		"workflow_id":     req.WorkflowID,
		"repository_id":   workflow.RepositoryID,
		"event":           req.Event,
		"head_sha":        req.HeadSHA,
		"run_number":      run.Number,
	}).Info("Created new workflow run")

	return run, nil
}

// createJobsForRun creates jobs for a workflow run
func (s *WorkflowService) createJobsForRun(ctx context.Context, run *models.WorkflowRun, workflowDef *actions.WorkflowDefinition) error {
	for jobID, jobDef := range workflowDef.Jobs {
		job := &models.Job{
			WorkflowRunID: run.ID,
			Name:          jobDef.Name,
			Status:        models.JobStatusQueued,
			Needs:         jobDef.Needs,
			Strategy:      jobDef.Strategy,
			Environment:   nil, // TODO: Handle environment
		}

		if job.Name == "" {
			job.Name = jobID
		}

		if err := s.db.WithContext(ctx).Create(job).Error; err != nil {
			return fmt.Errorf("failed to create job '%s': %w", jobID, err)
		}

		// Create steps for this job
		for i, stepDef := range jobDef.Steps {
			step := &models.Step{
				JobID:      job.ID,
				Number:     i + 1,
				Name:       stepDef.Name,
				Action:     stepDef.Uses,
				WithParams: stepDef.With,
				Env:        stepDef.Env,
				Status:     models.StepStatusQueued,
			}

			if step.Name == "" {
				if stepDef.Uses != "" {
					step.Name = fmt.Sprintf("Run %s", stepDef.Uses)
				} else if stepDef.Run != "" {
					step.Name = "Run script"
				} else {
					step.Name = fmt.Sprintf("Step %d", i+1)
				}
			}

			if stepDef.Run != "" {
				step.Action = "shell"
				if step.WithParams == nil {
					step.WithParams = make(map[string]interface{})
				}
				step.WithParams.(map[string]interface{})["script"] = stepDef.Run
			}

			if err := s.db.WithContext(ctx).Create(step).Error; err != nil {
				return fmt.Errorf("failed to create step %d for job '%s': %w", i+1, jobID, err)
			}
		}
	}

	return nil
}

// GetWorkflowRun retrieves a workflow run by ID
func (s *WorkflowService) GetWorkflowRun(ctx context.Context, id uuid.UUID) (*models.WorkflowRun, error) {
	var run models.WorkflowRun
	err := s.db.WithContext(ctx).
		Preload("Workflow").
		Preload("Repository").
		Preload("Actor").
		Preload("Jobs").
		Preload("Jobs.Steps").
		Preload("Artifacts").
		First(&run, "id = ?", id).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow run: %w", err)
	}

	return &run, nil
}

// ListWorkflowRuns lists workflow runs
func (s *WorkflowService) ListWorkflowRuns(ctx context.Context, req ListWorkflowRunsRequest) ([]models.WorkflowRun, int64, error) {
	query := s.db.WithContext(ctx).Model(&models.WorkflowRun{}).
		Where("repository_id = ?", req.RepositoryID)

	if req.WorkflowID != nil {
		query = query.Where("workflow_id = ?", *req.WorkflowID)
	}

	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	if req.Conclusion != nil {
		query = query.Where("conclusion = ?", *req.Conclusion)
	}

	if req.Event != nil {
		query = query.Where("event = ?", *req.Event)
	}

	if req.ActorID != nil {
		query = query.Where("actor_id = ?", *req.ActorID)
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count workflow runs: %w", err)
	}

	// Apply pagination
	if req.Limit > 0 {
		query = query.Limit(req.Limit)
	} else {
		query = query.Limit(50) // Default limit
	}

	if req.Offset > 0 {
		query = query.Offset(req.Offset)
	}

	var runs []models.WorkflowRun
	err := query.Preload("Workflow").
		Preload("Repository").
		Preload("Actor").
		Order("created_at DESC").
		Find(&runs).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list workflow runs: %w", err)
	}

	return runs, total, nil
}

// CancelWorkflowRun cancels a workflow run
func (s *WorkflowService) CancelWorkflowRun(ctx context.Context, id uuid.UUID) error {
	run, err := s.GetWorkflowRun(ctx, id)
	if err != nil {
		return err
	}

	if run.Status == models.WorkflowRunStatusCompleted {
		return fmt.Errorf("cannot cancel completed workflow run")
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":       models.WorkflowRunStatusCancelled,
		"conclusion":   models.WorkflowRunConclusionCancelled,
		"completed_at": &now,
		"updated_at":   now,
	}

	if err := s.db.WithContext(ctx).Model(run).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to cancel workflow run: %w", err)
	}

	// Cancel all running jobs
	jobUpdates := map[string]interface{}{
		"status":       models.JobStatusCancelled,
		"conclusion":   models.JobConclusionCancelled,
		"completed_at": &now,
		"updated_at":   now,
	}

	if err := s.db.WithContext(ctx).Model(&models.Job{}).
		Where("workflow_run_id = ? AND status IN (?)", id, []models.JobStatus{
			models.JobStatusQueued,
			models.JobStatusInProgress,
		}).Updates(jobUpdates).Error; err != nil {
		return fmt.Errorf("failed to cancel jobs: %w", err)
	}

	// Cancel all running steps
	stepUpdates := map[string]interface{}{
		"status":       models.StepStatusCancelled,
		"conclusion":   models.StepConclusionCancelled,
		"completed_at": &now,
		"updated_at":   now,
	}

	if err := s.db.WithContext(ctx).Model(&models.Step{}).
		Where("job_id IN (SELECT id FROM jobs WHERE workflow_run_id = ?) AND status IN (?)", 
			id, []models.StepStatus{
				models.StepStatusQueued,
				models.StepStatusInProgress,
			}).Updates(stepUpdates).Error; err != nil {
		return fmt.Errorf("failed to cancel steps: %w", err)
	}

	s.logger.WithField("workflow_run_id", id).Info("Cancelled workflow run")
	return nil
}

// CheckWorkflowTrigger checks if a workflow should be triggered for the given event and context
func (s *WorkflowService) CheckWorkflowTrigger(ctx context.Context, repositoryID uuid.UUID, event string, triggerCtx actions.TriggerContext) ([]models.Workflow, error) {
	// Get all enabled workflows for the repository
	workflows, _, err := s.ListWorkflows(ctx, ListWorkflowsRequest{
		RepositoryID: repositoryID,
		Enabled:      &[]bool{true}[0],
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get workflows: %w", err)
	}

	var triggeredWorkflows []models.Workflow
	for _, workflow := range workflows {
		workflowDef, err := s.parser.Parse(workflow.Content)
		if err != nil {
			s.logger.WithError(err).WithField("workflow_id", workflow.ID).
				Warn("Failed to parse workflow for trigger check")
			continue
		}

		if s.parser.ShouldTrigger(workflowDef, event, triggerCtx) {
			triggeredWorkflows = append(triggeredWorkflows, workflow)
		}
	}

	return triggeredWorkflows, nil
}

// Secret Management Methods (delegating to SecretService)

// ListSecrets lists secrets for a repository
func (s *WorkflowService) ListSecrets(ctx context.Context, repositoryID uuid.UUID) ([]models.Secret, error) {
	if s.secretService == nil {
		return nil, fmt.Errorf("secret service not initialized")
	}
	return s.secretService.ListSecrets(ctx, &repositoryID, nil)
}

// CreateSecret creates a new secret for a repository
func (s *WorkflowService) CreateSecret(ctx context.Context, repositoryID uuid.UUID, name, value string, environment *string) (*models.Secret, error) {
	if s.secretService == nil {
		return nil, fmt.Errorf("secret service not initialized")
	}
	return s.secretService.CreateSecret(ctx, &repositoryID, nil, name, value, environment)
}

// UpdateSecret updates a secret's value
func (s *WorkflowService) UpdateSecret(ctx context.Context, repositoryID uuid.UUID, secretID uuid.UUID, value string, environment *string) (*models.Secret, error) {
	if s.secretService == nil {
		return nil, fmt.Errorf("secret service not initialized")
	}
	
	// Verify the secret belongs to this repository
	secret, err := s.secretService.GetSecret(ctx, secretID)
	if err != nil {
		return nil, err
	}
	
	if secret.RepositoryID == nil || *secret.RepositoryID != repositoryID {
		return nil, fmt.Errorf("secret not found in repository")
	}
	
	return s.secretService.UpdateSecret(ctx, secretID, value, environment)
}

// DeleteSecret deletes a secret
func (s *WorkflowService) DeleteSecret(ctx context.Context, repositoryID uuid.UUID, secretID uuid.UUID) error {
	if s.secretService == nil {
		return fmt.Errorf("secret service not initialized")
	}
	
	// Verify the secret belongs to this repository
	secret, err := s.secretService.GetSecret(ctx, secretID)
	if err != nil {
		return err
	}
	
	if secret.RepositoryID == nil || *secret.RepositoryID != repositoryID {
		return fmt.Errorf("secret not found in repository")
	}
	
	return s.secretService.DeleteSecret(ctx, secretID)
}

// Job Log Methods

// GetJobLogs retrieves logs for a job
func (s *WorkflowService) GetJobLogs(ctx context.Context, jobID uuid.UUID) (string, error) {
	var job models.Job
	err := s.db.WithContext(ctx).
		Preload("Steps").
		First(&job, "id = ?", jobID).Error
	if err != nil {
		return "", fmt.Errorf("failed to get job: %w", err)
	}

	// Combine all step outputs into a single log
	var logs strings.Builder
	for _, step := range job.Steps {
		logs.WriteString(fmt.Sprintf("==== Step %d: %s ====\n", step.Number, step.Name))
		if step.Output != nil {
			logs.WriteString(*step.Output)
		} else {
			logs.WriteString("No output available\n")
		}
		logs.WriteString("\n")
	}

	return logs.String(), nil
}