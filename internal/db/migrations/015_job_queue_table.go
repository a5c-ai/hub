package migrations

import (
	"gorm.io/gorm"
)

func init() {
	registerMigration("015_job_queue_table", migrate015Up, migrate015Down)
}

func migrate015Up(db *gorm.DB) error {
	// Create job queue table for job scheduling
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS job_queue (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			job_id UUID NOT NULL,
			workflow_run_id UUID NOT NULL,
			priority INTEGER NOT NULL DEFAULT 100,
			status VARCHAR(50) NOT NULL DEFAULT 'pending',
			data JSONB NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		-- Create indexes for performance
		CREATE INDEX IF NOT EXISTS idx_job_queue_priority_created 
		ON job_queue (priority DESC, created_at ASC);
		
		CREATE INDEX IF NOT EXISTS idx_job_queue_status 
		ON job_queue (status);
		
		CREATE INDEX IF NOT EXISTS idx_job_queue_job_id 
		ON job_queue (job_id);

		-- Add check constraint for status values
		ALTER TABLE job_queue 
		ADD CONSTRAINT check_job_queue_status 
		CHECK (status IN ('pending', 'processing', 'completed', 'failed'));
	`).Error
}

func migrate015Down(db *gorm.DB) error {
	return db.Exec("DROP TABLE IF EXISTS job_queue").Error
}
