package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func addTransactionIdIntAndBackfillNumberMigration() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260306_add_transaction_id_int_backfill_number",
		Migrate: func(d *gorm.DB) error {
			// 1) add column id_int nếu chưa có
			if err := d.Exec(`
				ALTER TABLE transactions
				ADD COLUMN IF NOT EXISTS id_int BIGINT;
			`).Error; err != nil {
				return err
			}

			if err := d.Exec(`
				ALTER TABLE transactions
				ADD COLUMN IF NOT EXISTS transaction_number TEXT;
			`).Error; err != nil {
				return err
			}

			// 2) tạo sequence riêng (không phụ thuộc IDENTITY)
			if err := d.Exec(`
				CREATE SEQUENCE IF NOT EXISTS transactions_id_int_seq;
			`).Error; err != nil {
				return err
			}

			// 3) set default cho id_int = nextval(...)
			if err := d.Exec(`
				ALTER TABLE transactions
				ALTER COLUMN id_int SET DEFAULT nextval('transactions_id_int_seq');
			`).Error; err != nil {
				return err
			}

			// 4) backfill id_int cho các record cũ (atomic, safe concurrency trong transaction)
			if err := d.Exec(`
				UPDATE transactions
				SET id_int = nextval('transactions_id_int_seq')
				WHERE id_int IS NULL;
			`).Error; err != nil {
				return err
			}

			// 5) reset sequence = max(id_int) để insert mới tiếp tục tăng đúng
			if err := d.Exec(`
				SELECT setval(
					'transactions_id_int_seq',
					COALESCE((SELECT MAX(id_int) FROM transactions), 0)
				);
			`).Error; err != nil {
				return err
			}

			// 6) backfill transaction_number cho dữ liệu cũ (nếu null hoặc "")
			if err := d.Exec(`
				UPDATE transactions
				SET transaction_number = '#TS' || LPAD(id_int::text, 4, '0')
				WHERE (transaction_number IS NULL OR transaction_number = '')
				  AND id_int IS NOT NULL;
			`).Error; err != nil {
				return err
			}

			// 7) unique index để không trùng
			if err := d.Exec(`
				CREATE UNIQUE INDEX IF NOT EXISTS ux_transactions_id_int
				ON transactions(id_int);
			`).Error; err != nil {
				return err
			}

			if err := d.Exec(`
				CREATE UNIQUE INDEX IF NOT EXISTS ux_transactions_transaction_number
				ON transactions(transaction_number);
			`).Error; err != nil {
				return err
			}

			return nil
		},
	}
}
