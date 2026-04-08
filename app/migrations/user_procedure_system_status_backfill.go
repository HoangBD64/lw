package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func addUserProcedureSystemStatusBackfillMigration() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260403_backfill_user_procedure_system_status",
		Migrate: func(d *gorm.DB) error {
			statements := []string{
				`
				ALTER TABLE user_procedures
				ADD COLUMN IF NOT EXISTS system_status TEXT;
				`,
				`
				UPDATE user_procedures
				SET system_status = status
				WHERE status IS NOT NULL;
				`,
				`
				CREATE INDEX IF NOT EXISTS idx_user_procedures_system_status
				ON user_procedures(system_status);
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
