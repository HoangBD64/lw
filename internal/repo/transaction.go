package repo

import (
	"context"
	"core/app"
	"core/internal/domain_model/model"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func CreateTransaction(ctx context.Context, newTransaction model.Transaction) (model.Transaction, error) {
	tx := app.Database.DB.Create(&newTransaction)

	if err := tx.Error; err != nil {
		return newTransaction, tx.Error
	}

	return newTransaction, nil
}

func UpdateTransaction(pay_his model.Transaction) (model.Transaction, error) {
	tx := app.Database.DB.Updates(&pay_his)
	return pay_his, tx.Error
}

func GetTransactionByOrderID(order_id string, user model.User) (model.Transaction, error) {
	var pay_his model.Transaction
	tx := app.Database.DB.Model(&model.Transaction{}).
		Preload("UserProcedure", func(tx *gorm.DB) *gorm.DB {
			return tx.Select("id", "fulfillment_status", "fulfillment_cost_paid")
		}).
		Where("order_payment_id = ? AND user_id = ?", order_id, user.ID).
		First(&pay_his)
	return pay_his, tx.Error
}

func GetTransaction(query Query, user model.User) ([]model.Transaction, int64, error) {
	var pays []model.Transaction
	tx := app.Database.DB.Model(&model.Transaction{}).Where("user_id = ?", user.ID)
	if query.Status != "" {
		tx = tx.Where("status = ? ", query.Status)
	}
	if query.Type != "" {
		if strings.Contains(query.Type, ",") {
			types := strings.Split(query.Type, ",")
			for i := range types {
				types[i] = strings.TrimSpace(types[i])
			}
			tx = tx.Where("type IN ?", types)
		} else {
			tx = tx.Where("type = ? ", query.Type)
		}
	}
	if query.Search != "" {
		tx = tx.Where("order_payment_id = ?", query.Search)
	}
	if query.StartTime != nil && query.EndTime != nil {
		tx = tx.Where("created_at >= ? AND created_at <= ?", query.StartTime, query.EndTime)
	}
	var count int64
	err := tx.Count(&count)
	err = tx.Offset(query.Limit * (query.Page - 1)).Limit(query.Limit).Order("created_at DESC").Find(&pays)
	return pays, count, err.Error
}

func GetTransactionBalance(ctx context.Context, user model.User) (float64, error) {
	var count float64
	tx := app.Database.DB.Model(&model.Transaction{}).Where("user_id = ? AND status = ?", user.ID, "completed").Select("COALESCE(Sum(amount), 0)")
	tx.Scan(&count)
	if tx.Error != nil {
		return count, tx.Error
	}
	return count, nil
}
func AdminGetTransactionBalance(store_id uuid.UUID) (float64, error) {
	var count float64
	tx := app.Database.DB.Model(&model.Transaction{}).Where("status = ?", "completed").Select("COALESCE(Sum(amount), 0)")
	tx.Scan(&count)
	return count, tx.Error
}

func GetTransactionByID(id uuid.UUID) (model.Transaction, error) {
	var pay_his model.Transaction
	tx := app.Database.DB.Model(&model.Transaction{}).Where("id = ? ", id).First(&pay_his)
	return pay_his, tx.Error
}
func GetTransactionDetails(id uuid.UUID, user model.User) (model.Transaction, error) {
	var pay_his model.Transaction
	tx := app.Database.DB.Model(&model.Transaction{}).
		Preload("UserProcedureTransactions").
		Preload("UserProcedureTransactions.UserProcedure", func(tx *gorm.DB) *gorm.DB {
			return tx.Select("id", "order_number", "fulfillment_status", "fulfillment_cost", "order_code")
		}).
		Where("id = ? ", id).First(&pay_his)
	return pay_his, tx.Error
}

func AdminGetTransaction(query Query) ([]model.Transaction, int64, error) {
	var pays []model.Transaction
	tx := app.Database.DB.Model(&model.Transaction{}).
		Preload("User", func(tx *gorm.DB) *gorm.DB { return tx.Select("id", "name", "email", "status") }).
		Preload("Admin", func(tx *gorm.DB) *gorm.DB { return tx.Select("id", "name", "email", "status", "role") })
	if query.Status != "" {
		tx = tx.Where("status = ? ", query.Status)
	}
	if query.UserID != nil {
		tx = tx.Where("user_id = ?", query.UserID)
	}
	if query.Type != "" {
		if strings.Contains(query.Type, ",") {
			types := strings.Split(query.Type, ",")
			for i := range types {
				types[i] = strings.TrimSpace(types[i])
			}
			tx = tx.Where("type IN ?", types)
		} else {
			tx = tx.Where("type = ? ", query.Type)
		}
	}
	if query.Search != "" {
		tx = tx.Where("order_payment_id = ?", query.Search)
	}
	if query.StartTime != nil && query.EndTime != nil {
		tx = tx.Where("created_at >= ? AND created_at <= ?", query.StartTime, query.EndTime)
	}
	var count int64
	err := tx.Count(&count)
	err = tx.Offset(query.Limit * (query.Page - 1)).Limit(query.Limit).Order("created_at DESC").Find(&pays)
	return pays, count, err.Error
}

func GetTransactionForExports(ctx context.Context, query Query) ([]model.Transaction, error) {
	var out []model.Transaction

	tx := app.Database.DB.Model(&model.Transaction{}).Where("user_id = ?", query.UserID).
		Preload("UserProcedureTransactions").
		Preload("UserProcedureTransactions.UserProcedure", func(tx *gorm.DB) *gorm.DB {
			return tx.Select("id", "order_number", "fulfillment_status", "fulfillment_cost", "order_code")
		})

	if query.StartTime != nil && query.EndTime != nil {
		tx = tx.Where("created_at BETWEEN ? AND ?", query.StartTime, query.EndTime)
	}

	tx = tx.Order("created_at DESC").Find(&out)

	if err := tx.Error; err != nil {
		return nil, err
	}

	return out, nil
}

func UpdateTransactionMetadata(ctx context.Context, id uuid.UUID, metadata datatypes.JSON) error {
	tx := app.Database.Model(&model.Transaction{}).Where("id = ?", id).Update("metadata", metadata)

	if err := tx.Error; err != nil {
		return err
	}

	return nil
}

func GetReportUserTopupAmount(ctx context.Context, startTime, endTime time.Time) (float64, error) {
	var amount float64

	tx := app.Database.DB.Model(&model.Transaction{}).
		Where("type = ? AND status = ? AND created_at BETWEEN ? AND ?", model.USER_DEPOSIT, model.COMPLETED, startTime, endTime).
		Select("COALESCE(Sum(amount), 0) as amount").Pluck("amount", &amount)

	if err := tx.Error; err != nil {
		return 0, err
	}

	return amount, nil
}

func GetAnalyticFulfillmentAmountTransaction(ctx context.Context, query Query) ([]model.Transaction, error) {
	var transactions []model.Transaction

	tx := app.Database.DB.
		Model(&model.Transaction{}).
		Where("transactions.type IN ? ", []string{model.USER_DEPOSIT}).
		Where("transactions.created_at BETWEEN ? AND ?", query.StartTime, query.EndTime).
		Where("transactions.status = ?", model.COMPLETED).
		Joins("Store", app.Database.DB.Select("id", "name", "admin_id"))

	if query.ID != nil {
		tx = tx.Where(`"Store"."admin_id" = ?`, query.ID)
	}

	if query.UserID != nil {
		tx = tx.Where("transactions.user_id = ?", query.UserID)
	}

	if err := tx.Find(&transactions).Error; err != nil {
		return nil, err
	}

	return transactions, nil
}
