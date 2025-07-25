package migrations

import (
	"github.com/a5c-ai/hub/internal/models"
	"gorm.io/gorm"
)

func init() {
	registerMigration("017_stars_table", migrate017Up, migrate017Down)
}

func migrate017Up(db *gorm.DB) error {
	// Create stars table using AutoMigrate
	if err := db.AutoMigrate(&models.Star{}); err != nil {
		return err
	}

	// Create unique constraint for user_id and repository_id
	if err := db.Exec(`
		ALTER TABLE stars ADD CONSTRAINT unique_user_repository_star UNIQUE (user_id, repository_id);
	`).Error; err != nil {
		// Ignore error if constraint already exists
	}

	// Create function to update stars_count on repositories table
	if err := db.Exec(`
		CREATE OR REPLACE FUNCTION update_repository_stars_count()
		RETURNS TRIGGER AS $$
		BEGIN
			IF TG_OP = 'INSERT' THEN
				UPDATE repositories SET stars_count = stars_count + 1 WHERE id = NEW.repository_id;
			ELSIF TG_OP = 'DELETE' THEN
				UPDATE repositories SET stars_count = stars_count - 1 WHERE id = OLD.repository_id;
			END IF;
			RETURN NULL;
		END;
		$$ LANGUAGE plpgsql;
	`).Error; err != nil {
		return err
	}

	// Create triggers to automatically update stars_count
	if err := db.Exec(`
		DROP TRIGGER IF EXISTS trigger_update_stars_count_insert ON stars;
		DROP TRIGGER IF EXISTS trigger_update_stars_count_delete ON stars;
		
		CREATE TRIGGER trigger_update_stars_count_insert
		AFTER INSERT ON stars
		FOR EACH ROW EXECUTE FUNCTION update_repository_stars_count();

		CREATE TRIGGER trigger_update_stars_count_delete
		AFTER DELETE ON stars
		FOR EACH ROW EXECUTE FUNCTION update_repository_stars_count();
	`).Error; err != nil {
		return err
	}

	return nil
}

func migrate017Down(db *gorm.DB) error {
	// Drop triggers
	if err := db.Exec(`
		DROP TRIGGER IF EXISTS trigger_update_stars_count_insert ON stars;
		DROP TRIGGER IF EXISTS trigger_update_stars_count_delete ON stars;
	`).Error; err != nil {
		return err
	}

	// Drop function
	if err := db.Exec(`
		DROP FUNCTION IF EXISTS update_repository_stars_count();
	`).Error; err != nil {
		return err
	}

	// Drop table
	if err := db.Migrator().DropTable(&models.Star{}); err != nil {
		return err
	}

	return nil
}
