package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

type migrationFunc = func() *gormigrate.Migration

type Migrator struct {
	core *gormigrate.Gormigrate
}

func migrationsList() []*gormigrate.Migration {
	// Add your migration functions to this list
	list := []migrationFunc{
		addTransactionIdIntAndBackfillNumberMigration,
	}

	out := make([]*gormigrate.Migration, 0, len(list))
	for _, fn := range list {
		out = append(out, fn())
	}

	return out
}

func New(db *gorm.DB) *Migrator {
	core := gormigrate.New(db, gormigrate.DefaultOptions, migrationsList())

	return &Migrator{core}
}

func (m *Migrator) Migrate() error {
	return m.core.Migrate()
}
