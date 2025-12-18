package repo

import (
	"context"
	"core/app"
	"core/internal/domain_model/model"
)

func GetAnalyticsTicket(ctx context.Context, query Query) (int64, error) {
	var count int64 = 0
	tx := app.Database.DB.Model(&model.Ticket{}).Where("created_at >= ? AND created_at <= ?", query.StartTime, query.EndTime)

	if err := tx.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

func GetAnalyticsTotalDeposit(ctx context.Context, query Query, type_transaction string) (float64, error) {
	var count float64 = 0

	tx := app.Database.DB.Model(&model.Transaction{}).
		Where("created_at >= ? AND created_at <= ?", query.StartTime, query.EndTime).
		Where("type = ? AND status = ?", type_transaction, model.COMPLETED)

	tx = tx.Select("COALESCE(Sum(amount), 0)").Scan(&count)
	if err := tx.Error; err != nil {
		return 0, err
	}

	return count, nil
}

func GetAnalyticsNewUser(ctx context.Context, query Query) (int64, error) {
	var count int64 = 0

	tx := app.Database.DB.Model(&model.User{}).Where(`"users"."created_at" >= ? AND "users"."created_at" <= ?`, query.StartTime, query.EndTime)

	if err := tx.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

func GetAnalyticsUnreadTicket(user model.User) (int64, error) {
	var count int64 = 0
	tx := app.Database.DB.Table("tickets").Where("read_at IS NULL").Count(&count)
	return count, tx.Error
}

func GetAnalyticsBilling(ctx context.Context, query Query, typeName string) (float64, error) {
	var count float64 = 0

	tx := app.Database.DB.Model(&model.Transaction{}).
		Where("type = ?", typeName).
		Where("created_at >= ? AND created_at <= ?", query.StartTime, query.EndTime).
		Select("COALESCE(Sum(amount), 0)").Scan(&count)

	if err := tx.Error; err != nil {
		return 0, err
	}
	return count, nil
}

// ============================
// User Procedure Analytics
// ============================

type TemplateUsage struct {
	ProcedureID int    `json:"procedure_id"`
	Name        string `json:"name"`
	Usage       int64  `json:"usage"`
}

func GetUserProcedureStatusCounts(ctx context.Context, q Query) (map[string]int64, error) {
	result := map[string]int64{
		model.USER_PROCEDURE_CREATED:   0,
		model.USER_PROCEDURE_PENDING:   0,
		model.USER_PROCEDURE_COMPLETED: 0,
	}

	type row struct {
		Status string
		Count  int64
	}
	var rows []row

	tx := app.Database.DB.Model(&model.UserProcedure{}).
		Select("status, COUNT(*) as count").
		Where("created_at >= ? AND created_at <= ?", q.StartTime, q.EndTime).
		Group("status").
		Scan(&rows)

	if err := tx.Error; err != nil {
		return nil, err
	}

	for _, r := range rows {
		switch r.Status {
		case model.USER_PROCEDURE_CREATED:
			result[model.USER_PROCEDURE_CREATED] = r.Count
		case model.USER_PROCEDURE_PENDING:
			result[model.USER_PROCEDURE_PENDING] = r.Count
		case model.USER_PROCEDURE_COMPLETED:
			result[model.USER_PROCEDURE_COMPLETED] = r.Count
		}
	}
	return result, nil
}

func GetTemplateUsages(ctx context.Context, q Query) ([]TemplateUsage, error) {
	var out []TemplateUsage
	tx := app.Database.DB.Table("user_procedures up").
		Select("up.procedure_id as procedure_id, p.name as name, COUNT(*) as usage").
		Joins("JOIN procedures p ON p.id = up.procedure_id").
		Where("up.created_at >= ? AND up.created_at <= ?", q.StartTime, q.EndTime).
		Group("up.procedure_id, p.name").
		Order("usage DESC").
		Scan(&out)
	if err := tx.Error; err != nil {
		return nil, err
	}
	return out, nil
}

func GetAnalyticsDeposit(ctx context.Context, query Query) (float64, error) {
	var count float64 = 0
	tx := app.Database.DB.Model(&model.Transaction{}).
		Where("type = ?", model.USER_DEPOSIT).
		Where("status = ?", model.COMPLETED).
		Where("created_at >= ? AND created_at <= ?", query.StartTime, query.EndTime).
		Select("COALESCE(Sum(amount), 0)").Scan(&count)

	if err := tx.Error; err != nil {
		return 0, err
	}
	return count, nil
}

func GetAnalyticsEfficiencyOrder(ctx context.Context, user model.User, query Query) (int64, error) {
	var count int64 = 0
	tx := app.Database.DB.Model(&model.UserProcedure{}).
		Where("created_at >= ? AND created_at <= ?", query.StartTime, query.EndTime).
		Count(&count)

	if err := tx.Error; err != nil {
		return 0, err
	}
	return count, nil
}

func GetAnalyticsEfficiencyTicket(ctx context.Context, user model.User, query Query) (int64, error) {
	var count int64 = 0
	tx := app.Database.DB.Model(&model.Ticket{}).Where("created_at >= ? AND created_at <= ?", query.StartTime, query.EndTime).Count(&count)
	if err := tx.Error; err != nil {
		return 0, err
	}
	return count, nil
}
