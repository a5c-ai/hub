package migrations

import (
	"github.com/a5c-ai/hub/internal/models"
	"gorm.io/gorm"
)

func init() {
	registerMigration("014_actions_tables", migrate014Up, migrate014Down)
}

func migrate014Up(db *gorm.DB) error {
	// Create Actions tables in proper order due to foreign key dependencies
	if err := db.AutoMigrate(
		&models.Workflow{},
		&models.WorkflowRun{},
		&models.Runner{},
		&models.Job{},
		&models.Step{},
		&models.Artifact{},
		&models.Secret{},
	); err != nil {
		return err
	}

	// Create indexes for performance
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_workflow_runs_repository_status 
		ON workflow_runs(repository_id, status);
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_jobs_workflow_run_status 
		ON jobs(workflow_run_id, status);
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_runners_status_labels 
		ON runners USING gin(labels);
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_artifacts_workflow_run 
		ON artifacts(workflow_run_id);
	`).Error; err != nil {
		return err
	}

	// Full-text search index for workflow content
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_workflows_content 
		ON workflows USING gin(to_tsvector('english', content));
	`).Error; err != nil {
		return err
	}

	// Unique constraint for workflow path per repository
	if err := db.Exec(`
		ALTER TABLE workflows 
		ADD CONSTRAINT unique_workflow_path_per_repo 
		UNIQUE (repository_id, path);
	`).Error; err != nil {
		return err
	}

	// Unique constraint for workflow run number per repository
	if err := db.Exec(`
		ALTER TABLE workflow_runs 
		ADD CONSTRAINT unique_run_number_per_repo 
		UNIQUE (repository_id, number);
	`).Error; err != nil {
		return err
	}

	// Unique constraint for secret name per scope
	if err := db.Exec(`
		ALTER TABLE secrets 
		ADD CONSTRAINT unique_secret_per_scope 
		UNIQUE (name, repository_id, organization_id, environment);
	`).Error; err != nil {
		return err
	}

	// Check constraints for enum values
	if err := db.Exec(`
		ALTER TABLE workflow_runs 
		ADD CONSTRAINT check_workflow_run_status 
		CHECK (status IN ('queued', 'in_progress', 'completed', 'cancelled'));
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		ALTER TABLE workflow_runs 
		ADD CONSTRAINT check_workflow_run_conclusion 
		CHECK (conclusion IN ('success', 'failure', 'cancelled', 'skipped') OR conclusion IS NULL);
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		ALTER TABLE jobs 
		ADD CONSTRAINT check_job_status 
		CHECK (status IN ('queued', 'in_progress', 'completed', 'cancelled'));
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		ALTER TABLE jobs 
		ADD CONSTRAINT check_job_conclusion 
		CHECK (conclusion IN ('success', 'failure', 'cancelled', 'skipped') OR conclusion IS NULL);
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		ALTER TABLE steps 
		ADD CONSTRAINT check_step_status 
		CHECK (status IN ('queued', 'in_progress', 'completed', 'cancelled'));
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		ALTER TABLE steps 
		ADD CONSTRAINT check_step_conclusion 
		CHECK (conclusion IN ('success', 'failure', 'cancelled', 'skipped') OR conclusion IS NULL);
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		ALTER TABLE runners 
		ADD CONSTRAINT check_runner_status 
		CHECK (status IN ('online', 'offline', 'busy'));
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		ALTER TABLE runners 
		ADD CONSTRAINT check_runner_type 
		CHECK (type IN ('kubernetes', 'self-hosted'));
	`).Error; err != nil {
		return err
	}

	return nil
}

func migrate014Down(db *gorm.DB) error {
	// Drop tables in reverse order
	return db.Migrator().DropTable(
		&models.Secret{},
		&models.Artifact{},
		&models.Step{},
		&models.Job{},
		&models.Runner{},
		&models.WorkflowRun{},
		&models.Workflow{},
	)
}