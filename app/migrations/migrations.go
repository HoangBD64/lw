package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

type migrationFunc = func() *gormigrate.Migration

type Migrator struct {
	core *gormigrate.Gormigrate
}

func PreTableMigrationsList() []*gormigrate.Migration {
	// Add your migration functions to this list
	list := []migrationFunc{}

	out := make([]*gormigrate.Migration, 0, len(list))
	for _, fn := range list {
		out = append(out, fn())
	}

	return out
}

func PostTableMigrationsList() []*gormigrate.Migration {
	// Add your migration functions to this list
	list := []migrationFunc{
		addTransactionIdIntAndBackfillNumberMigration,
		addUserProcedurePriorityAndDocumentPublishMigration,
		addUserProcedureSystemStatusBackfillMigration,
		addTicketWorkflowMigration,
	}

	out := make([]*gormigrate.Migration, 0, len(list))
	for _, fn := range list {
		out = append(out, fn())
	}

	return out
}

func New(db *gorm.DB, list []*gormigrate.Migration) *Migrator {
	core := gormigrate.New(db, gormigrate.DefaultOptions, list)

	return &Migrator{core}
}

func (m *Migrator) Migrate() error {
	return m.core.Migrate()
}
