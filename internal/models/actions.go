package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// WorkflowRunStatus represents the status of a workflow run
type WorkflowRunStatus string

const (
	WorkflowRunStatusQueued     WorkflowRunStatus = "queued"
	WorkflowRunStatusInProgress WorkflowRunStatus = "in_progress"
	WorkflowRunStatusCompleted  WorkflowRunStatus = "completed"
	WorkflowRunStatusCancelled  WorkflowRunStatus = "cancelled"
)

// WorkflowRunConclusion represents the conclusion of a workflow run
type WorkflowRunConclusion string

const (
	WorkflowRunConclusionSuccess   WorkflowRunConclusion = "success"
	WorkflowRunConclusionFailure   WorkflowRunConclusion = "failure"
	WorkflowRunConclusionCancelled WorkflowRunConclusion = "cancelled"
	WorkflowRunConclusionSkipped   WorkflowRunConclusion = "skipped"
)

// JobStatus represents the status of a job
type JobStatus string

const (
	JobStatusQueued     JobStatus = "queued"
	JobStatusInProgress JobStatus = "in_progress"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusCancelled  JobStatus = "cancelled"
)

// JobConclusion represents the conclusion of a job
type JobConclusion string

const (
	JobConclusionSuccess   JobConclusion = "success"
	JobConclusionFailure   JobConclusion = "failure"
	JobConclusionCancelled JobConclusion = "cancelled"
	JobConclusionSkipped   JobConclusion = "skipped"
)

// StepStatus represents the status of a step
type StepStatus string

const (
	StepStatusQueued     StepStatus = "queued"
	StepStatusInProgress StepStatus = "in_progress"
	StepStatusCompleted  StepStatus = "completed"
	StepStatusCancelled  StepStatus = "cancelled"
)

// StepConclusion represents the conclusion of a step
type StepConclusion string

const (
	StepConclusionSuccess   StepConclusion = "success"
	StepConclusionFailure   StepConclusion = "failure"
	StepConclusionCancelled StepConclusion = "cancelled"
	StepConclusionSkipped   StepConclusion = "skipped"
)

// RunnerStatus represents the status of a runner
type RunnerStatus string

const (
	RunnerStatusOnline  RunnerStatus = "online"
	RunnerStatusOffline RunnerStatus = "offline"
	RunnerStatusBusy    RunnerStatus = "busy"
)

// RunnerType represents the type of a runner
type RunnerType string

const (
	RunnerTypeKubernetes RunnerType = "kubernetes"
	RunnerTypeSelfHosted RunnerType = "self-hosted"
)

// Workflow represents a workflow definition
type Workflow struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	RepositoryID uuid.UUID  `json:"repository_id" gorm:"type:uuid;not null;index"`
	Repository   Repository `json:"repository" gorm:"foreignKey:RepositoryID"`

	Name    string `json:"name" gorm:"not null;size:255"`
	Path    string `json:"path" gorm:"not null;size:500"`     // .hub/workflows/ci.yml
	Content string `json:"content" gorm:"type:text;not null"` // YAML content
	Enabled bool   `json:"enabled" gorm:"default:true"`

	// Relationships
	WorkflowRuns []WorkflowRun `json:"workflow_runs,omitempty" gorm:"foreignKey:WorkflowID"`
}

// WorkflowRun represents a single execution of a workflow
type WorkflowRun struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:(gen_random_uuid())"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	WorkflowID   uuid.UUID  `json:"workflow_id" gorm:"type:uuid;not null;index"`
	Workflow     Workflow   `json:"workflow,omitempty" gorm:"foreignKey:WorkflowID"`
	RepositoryID uuid.UUID  `json:"repository_id" gorm:"type:uuid;not null;index"`
	Repository   Repository `json:"repository,omitempty" gorm:"foreignKey:RepositoryID"`

	Number     int                    `json:"number" gorm:"not null"` // Sequential run number per repo
	Status     WorkflowRunStatus      `json:"status" gorm:"type:varchar(50);not null"`
	Conclusion *WorkflowRunConclusion `json:"conclusion" gorm:"type:varchar(50)"`

	HeadSHA    string  `json:"head_sha" gorm:"size:40;not null"`
	HeadBranch *string `json:"head_branch" gorm:"size:255"`
	Event      string  `json:"event" gorm:"size:50;not null"` // push, pull_request, schedule, workflow_dispatch

	EventPayload interface{} `json:"event_payload" gorm:"type:jsonb"`
	ActorID      *uuid.UUID  `json:"actor_id" gorm:"type:uuid;index"`
	Actor        *User       `json:"actor,omitempty" gorm:"foreignKey:ActorID"`

	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`

	// Relationships
	Jobs      []Job      `json:"jobs,omitempty" gorm:"foreignKey:WorkflowRunID"`
	Artifacts []Artifact `json:"artifacts,omitempty" gorm:"foreignKey:WorkflowRunID"`
}

// Job represents a job within a workflow run
type Job struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:(gen_random_uuid())"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	WorkflowRunID uuid.UUID   `json:"workflow_run_id" gorm:"type:uuid;not null;index"`
	WorkflowRun   WorkflowRun `json:"workflow_run,omitempty" gorm:"foreignKey:WorkflowRunID"`

	Name       string         `json:"name" gorm:"not null;size:255"`
	Status     JobStatus      `json:"status" gorm:"type:varchar(50);not null"`
	Conclusion *JobConclusion `json:"conclusion" gorm:"type:varchar(50)"`

	RunnerID   *uuid.UUID `json:"runner_id" gorm:"type:uuid;index"`
	Runner     *Runner    `json:"runner,omitempty" gorm:"foreignKey:RunnerID"`
	RunnerName *string    `json:"runner_name" gorm:"size:255"`

	Needs       interface{} `json:"needs" gorm:"type:jsonb"`     // Array of job IDs this job depends on
	Strategy    interface{} `json:"strategy" gorm:"type:jsonb"`  // Matrix strategy configuration
	Environment *string     `json:"environment" gorm:"size:255"` // Target environment

	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`

	// Relationships
	Steps []Step `json:"steps,omitempty" gorm:"foreignKey:JobID"`
}

// Step represents a single step within a job
type Step struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:(gen_random_uuid())"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	JobID uuid.UUID `json:"job_id" gorm:"type:uuid;not null;index"`
	Job   Job       `json:"job,omitempty" gorm:"foreignKey:JobID"`

	Number int    `json:"number" gorm:"not null"`
	Name   string `json:"name" gorm:"not null;size:255"`
	Action string `json:"action" gorm:"size:500"` // action@version or script

	WithParams interface{} `json:"with_params" gorm:"type:jsonb"` // Action inputs
	Env        interface{} `json:"env" gorm:"type:jsonb"`         // Environment variables

	Status     StepStatus      `json:"status" gorm:"type:varchar(50);not null"`
	Conclusion *StepConclusion `json:"conclusion" gorm:"type:varchar(50)"`
	Output     *string         `json:"output" gorm:"type:text"` // Step output logs

	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

// Runner represents a workflow runner
type Runner struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:(gen_random_uuid())"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	Name   string       `json:"name" gorm:"not null;size:255"`
	Labels interface{}  `json:"labels" gorm:"type:jsonb;not null"` // ['ubuntu-latest', 'self-hosted']
	Status RunnerStatus `json:"status" gorm:"type:varchar(50);not null"`
	Type   RunnerType   `json:"type" gorm:"type:varchar(50);not null"`

	Version      *string `json:"version" gorm:"size:50"`
	OS           *string `json:"os" gorm:"size:50"`
	Architecture *string `json:"architecture" gorm:"size:50"`

	RepositoryID   *uuid.UUID    `json:"repository_id" gorm:"type:uuid;index"` // null for organization runners
	Repository     *Repository   `json:"repository,omitempty" gorm:"foreignKey:RepositoryID"`
	OrganizationID *uuid.UUID    `json:"organization_id" gorm:"type:uuid;index"` // null for repo runners
	Organization   *Organization `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`

	LastSeenAt *time.Time `json:"last_seen_at"`

	// Relationships
	Jobs []Job `json:"jobs,omitempty" gorm:"foreignKey:RunnerID"`
}

// Artifact represents a build artifact
type Artifact struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:(gen_random_uuid())"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	WorkflowRunID uuid.UUID   `json:"workflow_run_id" gorm:"type:uuid;not null;index"`
	WorkflowRun   WorkflowRun `json:"workflow_run,omitempty" gorm:"foreignKey:WorkflowRunID"`

	Name      string     `json:"name" gorm:"not null;size:255"`
	Path      string     `json:"path" gorm:"not null;size:1000"` // Storage path
	SizeBytes int64      `json:"size_bytes" gorm:"not null"`
	Expired   bool       `json:"expired" gorm:"default:false"`
	ExpiresAt *time.Time `json:"expires_at"`
}

// Secret represents an encrypted secret
type Secret struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:(gen_random_uuid())"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	Name           string `json:"name" gorm:"not null;size:255"`
	EncryptedValue string `json:"-" gorm:"type:text;not null"` // Never expose in JSON

	RepositoryID   *uuid.UUID    `json:"repository_id" gorm:"type:uuid;index"`
	Repository     *Repository   `json:"repository,omitempty" gorm:"foreignKey:RepositoryID"`
	OrganizationID *uuid.UUID    `json:"organization_id" gorm:"type:uuid;index"`
	Organization   *Organization `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
	Environment    *string       `json:"environment" gorm:"size:255"` // Environment-specific secrets
}

// TableName methods for custom table names if needed
func (Workflow) TableName() string    { return "workflows" }
func (WorkflowRun) TableName() string { return "workflow_runs" }
func (Job) TableName() string         { return "jobs" }
func (Step) TableName() string        { return "steps" }
func (Runner) TableName() string      { return "runners" }
func (Artifact) TableName() string    { return "artifacts" }
func (Secret) TableName() string      { return "secrets" }

// BeforeCreate hook will set a new UUID if ID is empty to support SQLite in-memory databases
func (a *Artifact) BeforeCreate(tx *gorm.DB) (err error) {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return
}
