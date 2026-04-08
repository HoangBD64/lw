package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func addUserProcedurePriorityAndDocumentPublishMigration() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "1b8458fc-ff42-4496-b390-dd46a777c91e",
		Migrate: func(d *gorm.DB) error {
			statements := []string{
				`
				ALTER TABLE user_procedures
				ADD COLUMN IF NOT EXISTS priority TEXT DEFAULT 'medium';
				`,
				`
				UPDATE user_procedures
				SET priority = 'medium'
				WHERE priority IS NULL OR priority = '';
				`,
				`
				CREATE INDEX IF NOT EXISTS idx_user_procedures_priority
				ON user_procedures(priority);
				`,
				`
				ALTER TABLE user_procedures
				ADD COLUMN IF NOT EXISTS activity_logs JSONB DEFAULT '[]'::jsonb;
				`,
				`
				UPDATE user_procedures
				SET activity_logs = '[]'::jsonb
				WHERE activity_logs IS NULL;
				`,
				`
				ALTER TABLE documents
				ADD COLUMN IF NOT EXISTS is_publish BOOLEAN DEFAULT TRUE;
				`,
				`
				UPDATE documents
				SET is_publish = TRUE
				WHERE is_publish IS NULL;
				`,
				`
				ALTER TABLE documents
				ADD COLUMN IF NOT EXISTS user_procedure_step_id INTEGER;
				`,
				`
				CREATE INDEX IF NOT EXISTS idx_documents_is_publish
				ON documents(is_publish);
				`,
				`
				CREATE INDEX IF NOT EXISTS idx_documents_user_procedure_step_id
				ON documents(user_procedure_step_id);
				`,
			}

			for _, stmt := range statements {
				if err := d.Exec(stmt).Error; err != nil {
					return err
				}
			}

			return nil
		},
	}
}