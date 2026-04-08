package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func addTicketWorkflowMigration() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260403_add_ticket_workflow_fields_001",
		Migrate: func(d *gorm.DB) error {
			statements := []string{
				`
				ALTER TABLE tickets
				ADD COLUMN IF NOT EXISTS admin_id UUID;
				`,
				`
				ALTER TABLE tickets
				ADD COLUMN IF NOT EXISTS creator_admin_id UUID;
				`,
				`
				ALTER TABLE tickets
				ADD COLUMN IF NOT EXISTS creator_admin_read_at TIMESTAMPTZ;
				`,
				`
				ALTER TABLE tickets
				ADD COLUMN IF NOT EXISTS type TEXT DEFAULT 'procedure';
				`,
				`
				ALTER TABLE tickets
				ADD COLUMN IF NOT EXISTS priority TEXT DEFAULT 'low';
				`,
				`
				ALTER TABLE tickets
				ADD COLUMN IF NOT EXISTS subject TEXT;
				`,
				`
				ALTER TABLE tickets
				ADD COLUMN IF NOT EXISTS internal_note TEXT;
				`,
				`
				ALTER TABLE tickets
				ADD COLUMN IF NOT EXISTS is_internal BOOLEAN DEFAULT FALSE;
				`,
				`
				ALTER TABLE tickets
				ADD COLUMN IF NOT EXISTS last_message_at TIMESTAMPTZ;
				`,
				`
				ALTER TABLE tickets
				ADD COLUMN IF NOT EXISTS first_response_at TIMESTAMPTZ;
				`,
				`
				ALTER TABLE tickets
				ADD COLUMN IF NOT EXISTS resolved_at TIMESTAMPTZ;
				`,
				`
				ALTER TABLE tickets
				ADD COLUMN IF NOT EXISTS assigned_at TIMESTAMPTZ;
				`,
				`
				ALTER TABLE tickets
				ADD COLUMN IF NOT EXISTS closed_at TIMESTAMPTZ;
				`,
				`
				UPDATE tickets
				SET type = 'procedure'
				WHERE type IS NULL OR type = '';
				`,
				`
				UPDATE tickets
				SET priority = 'low'
				WHERE priority IS NULL OR priority = '';
				`,
				`
				UPDATE tickets
				SET subject = COALESCE(NULLIF(subject, ''), NULLIF(category, ''), 'Ticket #' || id::text)
				WHERE subject IS NULL OR subject = '';
				`,
				`
				UPDATE tickets
				SET last_message_at = COALESCE(last_message_at, updated_at, created_at)
				WHERE last_message_at IS NULL;
				`,
				`
				ALTER TABLE ticket_contents
				ADD COLUMN IF NOT EXISTS is_internal BOOLEAN DEFAULT FALSE;
				`,
				`
				UPDATE ticket_contents
				SET is_internal = FALSE
				WHERE is_internal IS NULL;
				`,
				`
				CREATE INDEX IF NOT EXISTS idx_tickets_admin_id ON tickets(admin_id);
				`,
				`
				CREATE INDEX IF NOT EXISTS idx_tickets_creator_admin_id ON tickets(creator_admin_id);
				`,
				`
				CREATE INDEX IF NOT EXISTS idx_tickets_type ON tickets(type);
				`,
				`
				CREATE INDEX IF NOT EXISTS idx_tickets_priority ON tickets(priority);
				`,
				`
				CREATE INDEX IF NOT EXISTS idx_tickets_is_internal ON tickets(is_internal);
				`,
				`
				CREATE INDEX IF NOT EXISTS idx_tickets_creator_admin_read_at ON tickets(creator_admin_read_at);
				`,
				`
				CREATE INDEX IF NOT EXISTS idx_tickets_last_message_at ON tickets(last_message_at);
				`,
				`
				CREATE INDEX IF NOT EXISTS idx_ticket_contents_is_internal ON ticket_contents(is_internal);
				`,
				`
				UPDATE tickets
				SET ticket_code = '#TK' || LPAD(id::text, 4, '0')
				WHERE ticket_code IS NULL OR ticket_code = '';
				`,
				`
				CREATE OR REPLACE FUNCTION set_ticket_code()
				RETURNS TRIGGER AS $$
				BEGIN
					IF NEW.ticket_code IS NULL OR NEW.ticket_code = '' THEN
						NEW.ticket_code := '#TK' || LPAD(NEW.id::text, 4, '0');
					END IF;
					RETURN NEW;
				END;
				$$ LANGUAGE plpgsql;
				`,
				`
				DROP TRIGGER IF EXISTS trg_set_ticket_code ON tickets;
				`,
				`
				CREATE TRIGGER trg_set_ticket_code
				BEFORE INSERT ON tickets
				FOR EACH ROW
				EXECUTE FUNCTION set_ticket_code();
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
