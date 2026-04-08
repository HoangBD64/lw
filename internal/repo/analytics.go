package repo

import (
	"context"
	"core/app"
	"core/internal/domain_model/domain"
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
		Where("type = ? AND status = ?", type_transaction, model.TRANSACTION_COMPLETED)

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
		model.USER_PROCEDURE_CREATED:              0,
		model.USER_PROCEDURE_IN_PROGRESS:          0,
		model.USER_PROCEDURE_INFORMATION_REQUIRED: 0,
		model.USER_PROCEDURE_COMPLETED:            0,
		model.USER_PROCEDURE_PENDING_REVIEW:       0,
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
		case model.USER_PROCEDURE_STEP_CREATED:
			result[model.USER_PROCEDURE_STEP_CREATED] = r.Count
		case model.USER_PROCEDURE_IN_PROGRESS:
			result[model.USER_PROCEDURE_IN_PROGRESS] = r.Count
		case model.USER_PROCEDURE_INFORMATION_REQUIRED:
			result[model.USER_PROCEDURE_INFORMATION_REQUIRED] = r.Count
		case model.USER_PROCEDURE_COMPLETED:
			result[model.USER_PROCEDURE_COMPLETED] = r.Count
		case model.USER_PROCEDURE_PENDING_REVIEW:
			result[model.USER_PROCEDURE_PENDING_REVIEW] = r.Count
		}
	}
	return result, nil
}

func GetUserProcedureStepStatusCounts(ctx context.Context, q Query) (map[string]int64, error) {
	result := map[string]int64{
		model.USER_PROCEDURE_STEP_CREATED:              0,
		model.USER_PROCEDURE_STEP_IN_PROGRESS:          0,
		model.USER_PROCEDURE_STEP_INFORMATION_REQUIRED: 0,
		model.USER_PROCEDURE_STEP_DONE:                 0,
	}

	type row struct {
		Status string
		Count  int64
	}
	var rows []row

	tx := app.Database.DB.Model(&model.UserProcedureStep{}).
		Select("status, COUNT(*) as count").
		Where("created_at >= ? AND created_at <= ?", q.StartTime, q.EndTime).
		Group("status").
		Scan(&rows)

	if err := tx.Error; err != nil {
		return nil, err
	}

	for _, r := range rows {
		switch r.Status {
		case model.USER_PROCEDURE_STEP_CREATED:
			result[model.USER_PROCEDURE_STEP_CREATED] = r.Count
		case model.USER_PROCEDURE_STEP_IN_PROGRESS:
			result[model.USER_PROCEDURE_STEP_IN_PROGRESS] = r.Count
		case model.USER_PROCEDURE_STEP_INFORMATION_REQUIRED:
			result[model.USER_PROCEDURE_STEP_INFORMATION_REQUIRED] = r.Count
		case model.USER_PROCEDURE_STEP_DONE:
			result[model.USER_PROCEDURE_STEP_DONE] = r.Count
		}
	}
	return result, nil
}

func GetTemplateUsages(ctx context.Context, q Query) ([]TemplateUsage, error) {
	var out []TemplateUsage
	tx := app.Database.DB.Debug().
		Model(&model.UserProcedure{}).
		Select("user_procedures.procedure_id as procedure_id, p.name as name, COUNT(*) as usage").
		Joins("JOIN procedures p ON p.id = user_procedures.procedure_id").
		Where("p.deleted_at IS NULL").
		Where("user_procedures.created_at >= ? AND user_procedures.created_at <= ?", q.StartTime, q.EndTime).
		Group("user_procedures.procedure_id, p.name").
		Order("usage DESC").
		Scan(&out)

	if err := tx.Error; err != nil {
		return nil, err
	}
	return out, nil
}

func GetAnalyticsIsUsingService(ctx context.Context, q Query) (float64, error) {
	var count float64 = 0

	tx := app.Database.DB.Model(&model.UserProcedure{}).
		Where("created_at >= ? AND created_at <= ? AND status NOT IN ?", q.StartTime, q.EndTime, []string{model.USER_PROCEDURE_CANCELLED, model.USER_PROCEDURE_COMPLETED}).
		Select("COALESCE(Sum(CASE WHEN user_id IS NOT NULL THEN 1 ELSE 0 END) / NULLIF(COUNT(*), 0), 0)").Scan(&count)

	if err := tx.Error; err != nil {
		return 0, err
	}
	return count, nil
}

func GetAnalyticsDeposit(ctx context.Context, query Query) (float64, error) {
	var count float64 = 0
	tx := app.Database.DB.Model(&model.Transaction{}).
		Where("type = ?", model.USER_DEPOSIT).
		Where("status = ?", model.TRANSACTION_COMPLETED).
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

func GetAdminDashboardAssignedProcedures(admin model.Admin, q Query) ([]domain.AdminDashboardAssignedProcedure, int64, error) {
	var (
		items []domain.AdminDashboardAssignedProcedure
		total int64
	)

	query := app.Database.DB.Model(&model.UserProcedure{}).
		Select(`
			user_procedures.id,
			user_procedures.proc_number,
			user_procedures.template_name,
			user_procedures.company_name,
			user_procedures.status,
			user_procedures.priority,
			user_procedures.deadline_at,
			user_procedures.admin_id,
			admins.name as admin_name,
			user_procedures.user_id,
			users.name as user_name
		`).
		Joins("LEFT JOIN admins ON admins.id = user_procedures.admin_id").
		Joins("LEFT JOIN users ON users.id = user_procedures.user_id").
		Where("user_procedures.admin_id IS NOT NULL")

	if admin.IsLawyer() {
		query = query.Where("user_procedures.admin_id = ?", admin.ID)
	}
	if q.Status != "" {
		query = query.Where("user_procedures.status = ?", q.Status)
	}
	if q.Search != "" {
		like := "%" + q.Search + "%"
		query = query.Where("user_procedures.proc_number ILIKE ? OR user_procedures.template_name ILIKE ? OR user_procedures.company_name ILIKE ? OR users.name ILIKE ?", like, like, like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderBy := `user_procedures.deadline_at ASC NULLS LAST,
		CASE user_procedures.priority
			WHEN 'high' THEN 3
			WHEN 'medium' THEN 2
			WHEN 'low' THEN 1
			ELSE 0
		END DESC,
		user_procedures.updated_at DESC`

	err := query.Scopes(pagination(q.Page, q.Limit)).Order(orderBy).Scan(&items).Error
	return items, total, err
}

func GetAdminDashboardLawyerWorkloads(admin model.Admin) ([]domain.AdminDashboardLawyerWorkload, error) {
	var items []domain.AdminDashboardLawyerWorkload
	trackedStatuses := []string{
		model.USER_PROCEDURE_CREATED,
		model.USER_PROCEDURE_SUBMITTED,
		model.USER_PROCEDURE_PENDING_REVIEW,
		model.USER_PROCEDURE_IN_PROGRESS,
		model.USER_PROCEDURE_INFORMATION_REQUIRED,
		model.USER_PROCEDURE_FILED,
		model.USER_PROCEDURE_COMPLETED,
	}

	query := app.Database.DB.Model(&model.Admin{}).
		Select(`
		admins.id as admin_id,
		admins.name as admin_name,
		COALESCE(SUM(CASE WHEN user_procedures.id IS NOT NULL AND user_procedures.status <> ? THEN 1 ELSE 0 END), 0) as active_procedures,
		COALESCE(SUM(CASE WHEN user_procedures.id IS NOT NULL AND user_procedures.status = ? THEN 1 ELSE 0 END), 0) as completed_procedures
	`, model.USER_PROCEDURE_COMPLETED, model.USER_PROCEDURE_COMPLETED).
		Joins(`
		LEFT JOIN user_procedures 
		ON user_procedures.admin_id = admins.id 
		AND user_procedures.status IN ? 
		AND user_procedures.deleted_at IS NULL
	`, trackedStatuses).
		Where("admins.role = ?", model.RoleLawyer).
		Group("admins.id, admins.name").
		Order("admins.name ASC")

	if admin.IsLawyer() {
		query = query.Where("admins.id = ?", admin.ID)
	}

	return items, query.Scan(&items).Error
}

func GetAdminDashboardUserWorkloads() ([]domain.AdminDashboardUserWorkload, error) {
	var items []domain.AdminDashboardUserWorkload
	trackedStatuses := []string{
		model.USER_PROCEDURE_CREATED,
		model.USER_PROCEDURE_SUBMITTED,
		model.USER_PROCEDURE_PENDING_REVIEW,
		model.USER_PROCEDURE_IN_PROGRESS,
		model.USER_PROCEDURE_INFORMATION_REQUIRED,
		model.USER_PROCEDURE_FILED,
		model.USER_PROCEDURE_COMPLETED,
	}

	query := app.Database.DB.Model(&model.User{}).
		Select(`
		users.id as user_id,
		users.name as user_name,
		COALESCE(SUM(CASE WHEN user_procedures.id IS NOT NULL AND user_procedures.status <> ? THEN 1 ELSE 0 END), 0) as active_procedures,
		COALESCE(SUM(CASE WHEN user_procedures.id IS NOT NULL AND user_procedures.status = ? THEN 1 ELSE 0 END), 0) as completed_procedures
	`, model.USER_PROCEDURE_COMPLETED, model.USER_PROCEDURE_COMPLETED).
		Joins(`
		LEFT JOIN user_procedures 
		ON user_procedures.user_id = users.id 
		AND user_procedures.status IN ? 
		AND user_procedures.deleted_at IS NULL
	`, trackedStatuses).
		Group("users.id, users.name").
		Order("users.name ASC")

	return items, query.Scan(&items).Error
}
