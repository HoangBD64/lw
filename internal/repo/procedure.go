package repo

import (
	"encoding/json"
	"fmt"
	"time"

	"core/app"
	"core/internal/domain_model/domain"
	"core/internal/domain_model/model"

	"github.com/google/uuid"
	"github.com/samber/lo"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func AdminCreateProcedure(p *model.Procedure) error {
	return app.Database.DB.Create(p).Error
}

func AdminGetProcedureByID(id int) (model.Procedure, error) {
	var p model.Procedure
	tx := app.Database.DB.
		Preload("ProcedureSteps").
		Preload("ExtraFees").
		First(&p, "id = ?", id)
	return p, tx.Error
}

func AdminGetProceduresByIDs(ids []int) ([]model.Procedure, error) {
	var procedures []model.Procedure
	tx := app.Database.DB.Model(&model.Procedure{}).
		Preload("ProcedureSteps", func(tx *gorm.DB) *gorm.DB {
			return tx.Select("id", "procedure_id", "position", "status")
		}).
		Where("id IN ?", ids).
		Find(&procedures)

	return procedures, tx.Error
}

func AdminUpdateProcedure(p *model.Procedure) error {
	return app.Database.DB.Model(&model.Procedure{}).Where("id = ?", p.ID).Updates(p).Error
}

func AdminDeleteProcedureByIDs(ids []int) error {
	return app.Database.DB.Delete(&model.Procedure{}, "id IN ?", ids).Error
}

func AdminGetProcedures(q Query) ([]model.Procedure, int64, error) {
	var (
		items []model.Procedure
		count int64
	)
	tx := app.Database.DB.Model(&model.Procedure{}).
		Preload("ProcedureSteps", func(tx *gorm.DB) *gorm.DB {
			return tx.Select("id", "procedure_id", "position", "name").Order("position ASC")
		})
	if q.Search != "" {
		tx = tx.Where("name ILIKE ?", "%"+q.Search+"%")
	}
	if q.Status != "" {
		tx = tx.Where("status = ?", q.Status)
	}
	if q.AdminID != nil {
		tx = tx.Where("admin_id = ?", q.AdminID)
	}
	if q.StartTime != nil && q.EndTime != nil {
		tx = tx.Where("created_at >= ? AND created_at <= ?", q.StartTime, q.EndTime)
	}
	tx = tx.Count(&count)
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	tx = tx.Scopes(pagination(q.Page, q.Limit)).Order("created_at DESC").Find(&items)
	return items, count, tx.Error
}

func AdminGetProcedureSteps(procedureID int) ([]model.ProcedureStep, error) {
	var steps []model.ProcedureStep
	tx := app.Database.DB.Model(&model.ProcedureStep{}).
		Where("procedure_id = ?", procedureID).
		Order("position ASC").
		Find(&steps)
	return steps, tx.Error
}

func AdminCreateProcedureStep(step model.ProcedureStep) (model.ProcedureStep, error) {
	tx := app.Database.DB.Model(&model.ProcedureStep{}).Create(&step)
	return step, tx.Error
}

func AdminUpdateProcedureStep(step *model.ProcedureStep) error {
	return app.Database.DB.Model(&model.ProcedureStep{}).Where("id = ?", step.ID).Updates(step).Error
}

func AdminDeleteProcedureStep(id int) error {
	return app.Database.DB.Delete(&model.ProcedureStep{}, "id = ?", id).Error
}

func AdminGetProcedureStepByID(id int) (model.ProcedureStep, error) {
	var step model.ProcedureStep
	tx := app.Database.DB.Model(&model.ProcedureStep{}).
		Where("id = ?", id).First(&step)
	return step, tx.Error
}

func AdminGetUserProcedureStepsByUserProcedureAndStepIDs(userProcedureID int, stepIDs []int) ([]model.UserProcedureStep, error) {
	var ups []model.UserProcedureStep
	tx := app.Database.DB.Model(&model.UserProcedureStep{}).
		Where("user_procedure_id = ? AND id IN ?", userProcedureID, stepIDs).
		Find(&ups)
	return ups, tx.Error
}

func AdminBulkAssignToUserProcedureSteps(ups []model.UserProcedureStep) (int64, error) {
	err := app.Database.DB.Transaction(func(tx *gorm.DB) error {
		for _, u := range ups {
			if err := tx.Model(&model.UserProcedureStep{}).
				Where("id = ?", u.ID).
				Updates(map[string]any{"admin_id": u.AdminID}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return int64(len(ups)), nil
}

func AdminGetUserProcedureByID(userProcedureID int) (model.UserProcedure, error) {
	var up model.UserProcedure
	err := app.Database.DB.Model(&model.UserProcedure{}).
		Where("id = ?", userProcedureID).
		First(&up).Error
	return up, err
}

func AdminGetUserProcedureByIDs(userProcedureIDs []int) ([]model.UserProcedure, error) {
	var ups []model.UserProcedure
	err := app.Database.DB.Model(&model.UserProcedure{}).
		Where("id IN ?", userProcedureIDs).
		Find(&ups).Error
	return ups, err
}

func AdminListUserProcedures(q Query) ([]model.UserProcedure, int64, error) {
	var (
		items []model.UserProcedure
		count int64
	)
	tx := app.Database.DB.Model(&model.UserProcedure{}).
		Omit("Procedure", "Admin").
		Preload("UserProcedureSteps", func(tx *gorm.DB) *gorm.DB {
			return tx.Select("id", "user_procedure_id", "position", "status").Order("position ASC")
		}).
		Preload("User", func(dbx *gorm.DB) *gorm.DB {
			return dbx.Model(&model.User{}).Select("id", "name", "email", "phone", "role", "avatar", "status")
		}).
		Preload("Admin", func(dbx *gorm.DB) *gorm.DB {
			return dbx.Model(&model.Admin{}).Select("id", "name", "email", "avatar", "status")
		})
	if q.UserID != nil {
		tx = tx.Where("user_id = ?", q.UserID)
	}
	if q.StartTime != nil && q.EndTime != nil {
		tx = tx.Where("created_at >= ? AND created_at <= ?", q.StartTime, q.EndTime)
	}
	if q.Status != "" {
		tx = tx.Where("status = ?", q.Status)
	}
	if q.SystemStatus != "" {
		tx = tx.Where("system_status = ?", q.SystemStatus)
	}
	if q.AdminID != nil {
		tx = tx.Where("admin_id = ?", q.AdminID)
	}
	if q.ProcedureID != nil {
		tx = tx.Where("procedure_id = ?", q.ProcedureID)
	}
	if q.Search != "" {
		tx = tx.Where("template_name ILIKE ? OR company_name ILIKE ? OR proc_number ILIKE ? OR company_code ILIKE ?", "%"+q.Search+"%", "%"+q.Search+"%", "%"+q.Search+"%", "%"+q.Search+"%")
	}
	if q.IsHaveTicket != nil {
		tx = tx.Where("is_have_ticket = ?", q.IsHaveTicket)
	}
	if q.IsAssigned != nil {
		if *q.IsAssigned {
			tx = tx.Where("admin_id IS NOT NULL")
		} else {
			tx = tx.Where("admin_id IS NULL AND status = ?", model.USER_PROCEDURE_CREATED)
		}
	}
	tx = tx.Count(&count)
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	tx = tx.Scopes(pagination(q.Page, q.Limit)).
		Order("created_at DESC").
		Find(&items)
	return items, count, tx.Error
}

func AssignAdminToUserProcedures(userProcedureIDs []int, adminID uuid.UUID) (int64, error) {
	// if status is PENDING, change to PROCESSING, else keep current status
	allowAssignStatuses := []string{
		model.USER_PROCEDURE_SUBMITTED,
		model.USER_PROCEDURE_PENDING_REVIEW,
	}

	tx := app.Database.DB.Model(&model.UserProcedure{}).
		Where("id IN ? AND status IN ?",
			userProcedureIDs,
			allowAssignStatuses,
		).
		Updates(map[string]any{
			"admin_id":      adminID,
			"status":        model.USER_PROCEDURE_IN_PROGRESS,
			"system_status": model.USER_PROCEDURE_IN_PROGRESS,
		})

	_ = app.Database.DB.Model(&model.UserProcedureStep{}).
		Where("user_procedure_id IN ? AND status IN ?",
			userProcedureIDs,
			allowAssignStatuses,
		).
		Update("status", model.USER_PROCEDURE_STEP_IN_PROGRESS)

	return tx.RowsAffected, tx.Error
}

func UnassignAdminFromUserProcedure(userProcedureIDs []int) (int64, error) {
	tx := app.Database.DB.Model(&model.UserProcedure{}).
		Where("id IN ? AND status NOT IN ? AND admin_id IS NOT NULL", userProcedureIDs, []string{model.USER_PROCEDURE_COMPLETED}).
		Updates(map[string]any{"admin_id": nil})

	return tx.RowsAffected, tx.Error
}

func AdminGetUserProcedureSteps(query Query) ([]model.UserProcedureStep, int64, error) {
	var (
		steps []model.UserProcedureStep
		count int64
	)
	tx := app.Database.DB.Model(&model.UserProcedureStep{}).
		Preload("UserProcedure", func(dbx *gorm.DB) *gorm.DB {
			return dbx.Select("id", "proc_number", "template_name", "company_name", "user_id", "admin_id")
		}).
		Omit("Data")
	if query.AdminID != nil {
		sub := app.Database.DB.
			Model(&model.UserProcedure{}).
			Select("id").
			Where("admin_id = ?", *query.AdminID)
		tx = tx.Where("user_procedure_id IN (?)", sub)
	}
	if query.Status != "" {
		tx = tx.Where("status = ?", query.Status)
	}
	if query.Search != "" {
		like := "%" + query.Search + "%"
		tx = tx.Where(`
		user_procedure_steps.name ILIKE ?
		OR EXISTS (
			SELECT 1
			FROM user_procedures up
			WHERE up.id = user_procedure_steps.user_procedure_id
			  AND (up.proc_number ILIKE ? OR up.template_name ILIKE ? OR up.company_name ILIKE ?)
		)`, like, like, like, like)
	}
	if query.StartTime != nil && query.EndTime != nil {
		tx = tx.Where("created_at >= ? AND created_at <= ?", query.StartTime, query.EndTime)
	}
	tx = tx.Count(&count)
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	tx = tx.Scopes(pagination(query.Page, query.Limit)).
		Order("created_at DESC").
		Find(&steps)
	return steps, count, tx.Error
}

func AdminGetUserProcedureDetailByID(userProcedureID int) (model.UserProcedure, error) {
	var up model.UserProcedure
	err := app.Database.DB.Model(&model.UserProcedure{}).
		Preload("UserProcedureSteps", func(dbx *gorm.DB) *gorm.DB {
			return dbx.
				Order("position ASC").
				Preload("Documents", func(docTx *gorm.DB) *gorm.DB {
					return docTx.Order("created_at DESC")
				})
		}).
		Preload("Admin", func(dbx *gorm.DB) *gorm.DB {
			return dbx.Model(&model.Admin{}).Select("id", "name", "email", "role", "avatar", "status")
		}).
		Preload("User", func(dbx *gorm.DB) *gorm.DB {
			return dbx.Model(&model.User{}).Select("id", "name", "email", "phone", "avatar", "status")
		}).
		Where("id = ?", userProcedureID).
		First(&up).Error
	return up, err
}

func AdminUpdateUserProcedure(up *model.UserProcedure) error {
	return app.Database.DB.Model(&model.UserProcedure{}).
		Where("id = ?", up.ID).
		Updates(up).Error
}

func AdminMarkReviewUserProcedures(ups []model.UserProcedure) error {
	ids := lo.Map(ups, func(up model.UserProcedure, _ int) int { return up.ID })
	err := app.Database.DB.Debug().Model(&model.UserProcedure{}).
		Where("id IN ?", ids).
		Updates(map[string]interface{}{
			"status":        model.USER_PROCEDURE_PENDING_REVIEW,
			"system_status": model.USER_PROCEDURE_PENDING_REVIEW,
		}).Error
	return err
}

func AdminHoldUserProcedures(ups []model.UserProcedure) error {
	ids := lo.Map(ups, func(up model.UserProcedure, _ int) int { return up.ID })
	err := app.Database.DB.Debug().Model(&model.UserProcedure{}).
		Where("id IN ?", ids).
		Updates(map[string]interface{}{
			"system_status": model.USER_PROCEDURE_HOLD,
			"status":        model.USER_PROCEDURE_HOLD,
		}).Error
	return err
}

func AdminResumeUserProcedures(ups []model.UserProcedure) error {
	ids := lo.Map(ups, func(up model.UserProcedure, _ int) int { return up.ID })
	err := app.Database.DB.Debug().Model(&model.UserProcedure{}).
		Where("id IN ?", ids).
		Updates(map[string]interface{}{
			"system_status": model.USER_PROCEDURE_IN_PROGRESS,
			"status":        model.USER_PROCEDURE_IN_PROGRESS,
		}).Error
	return err
}

func GetUserProcedureActivityLogs(userProcedureID int) ([]model.ActivityLog, error) {
	var up model.UserProcedure
	if err := app.Database.DB.Model(&model.UserProcedure{}).
		Select("id", "activity_logs").
		Where("id = ?", userProcedureID).
		First(&up).Error; err != nil {
		return nil, err
	}
	if len(up.ActivityLogs) == 0 {
		return []model.ActivityLog{}, nil
	}

	var logs []model.ActivityLog
	if err := json.Unmarshal(up.ActivityLogs, &logs); err != nil {
		return nil, err
	}
	return logs, nil
}

func AppendUserProcedureActivity(userProcedureID int, entry model.ActivityLog) error {
	logs, err := GetUserProcedureActivityLogs(userProcedureID)
	if err != nil {
		return err
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}
	logs = append(logs, entry)
	payload, err := json.Marshal(logs)
	if err != nil {
		return err
	}
	return app.Database.DB.Model(&model.UserProcedure{}).
		Where("id = ?", userProcedureID).
		Update("activity_logs", datatypes.JSON(payload)).Error
}

func AdminBulkUpsertProcedureWithSteps(proc model.Procedure, steps []model.ProcedureStep) (model.Procedure, []model.ProcedureStep, error) {
	var savedProc model.Procedure
	var savedSteps []model.ProcedureStep
	isCreateSteps := false
	err := app.Database.DB.Transaction(func(tx *gorm.DB) error {
		if proc.ID == 0 {
			if err := tx.Create(&proc).Error; err != nil {
				return err
			}
			isCreateSteps = true
		} else {
			if err := tx.Model(&model.Procedure{}).Where("id = ?", proc.ID).Updates(&proc).Error; err != nil {
				return err
			}
		}
		savedProc = proc

		if isCreateSteps {
			// only create steps after
			for i := range steps {
				steps[i].ProcedureID = savedProc.ID
				if err := tx.Model(&model.ProcedureStep{}).Create(&steps[i]).Error; err != nil {
					return err
				}
				savedSteps = append(savedSteps, steps[i])
			}
			return nil
		} else {
			// upsert steps
			for _, step := range steps {
				if step.ID == 0 {
					// create
					step.ProcedureID = savedProc.ID
					if err := tx.Model(&model.ProcedureStep{}).Create(&step).Error; err != nil {
						return err
					}
					savedSteps = append(savedSteps, step)
				} else {
					// update
					if err := tx.Model(&model.ProcedureStep{}).Where("id = ? AND procedure_id = ?", step.ID, savedProc.ID).Updates(&step).Error; err != nil {
						return err
					}
					savedSteps = append(savedSteps, step)
				}
			}
		}
		return nil
	})
	return savedProc, savedSteps, err
}

func PruneMissingProcedureSteps(procedureID int, keepIDs []int) error {
	return app.Database.DB.Model(&model.ProcedureStep{}).
		Where("procedure_id = ? AND id NOT IN ?", procedureID, keepIDs).
		Delete(&model.ProcedureStep{}).Error
}

func AdminGetAllStepStatusByUserProcedureID(userProcedureID int) ([]model.UserProcedureStep, error) {
	var steps []model.UserProcedureStep
	tx := app.Database.DB.Model(&model.UserProcedureStep{}).
		Where("user_procedure_id = ?", userProcedureID).
		Select("id", "status").
		Order("position ASC").
		Find(&steps)
	return steps, tx.Error
}

func AdminBulkUpdateUserProcedureSteps(steps []model.UserProcedureStep) (int64, error) {
	err := app.Database.DB.Transaction(func(tx *gorm.DB) error {
		for _, step := range steps {
			if err := tx.Model(&model.UserProcedureStep{}).
				Where("id = ?", step.ID).
				Updates(step).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return int64(len(steps)), nil
}

func UserGetProcedureSteps(procedureID int) ([]model.ProcedureStep, error) {
	var steps []model.ProcedureStep
	tx := app.Database.DB.Model(&model.ProcedureStep{}).
		Where("procedure_id = ?", procedureID).
		Order("position ASC").
		Find(&steps)
	return steps, tx.Error
}

func UserGetProcedures(q Query) ([]model.Procedure, int64, error) {
	var (
		items []model.Procedure
		count int64
	)
	tx := app.Database.DB.Model(&model.Procedure{}).
		Preload("ProcedureSteps", func(tx *gorm.DB) *gorm.DB {
			return tx.Select("id", "procedure_id", "position", "name").Order("position ASC")
		}).
		Where("is_active = ?", true)
	if q.Search != "" {
		tx = tx.Where("name ILIKE ?", "%"+q.Search+"%")
	}
	if q.Status != "" {
		tx = tx.Where("status = ?", q.Status)
	}
	if q.AdminID != nil {
		tx = tx.Where("admin_id = ?", q.AdminID)
	}
	tx = tx.Count(&count)
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	tx = tx.Scopes(pagination(q.Page, q.Limit)).Order("created_at DESC").Find(&items)
	return items, count, tx.Error
}

func UserGetUserProcedureDetail(userProcedureID int, userID uuid.UUID) ([]model.UserProcedure, error) {
	var ups []model.UserProcedure
	tx := app.Database.DB.Model(&model.UserProcedure{}).
		Preload("UserProcedureSteps", func(dbx *gorm.DB) *gorm.DB {
			return dbx.
				Order("position ASC").
				Preload("Documents", func(docTx *gorm.DB) *gorm.DB {
					return docTx.Where("COALESCE(is_publish, false) = ?", true).Order("created_at DESC")
				})
		}).
		Omit("SystemStatus").
		Where("id = ? AND user_id = ?", userProcedureID, userID).
		Find(&ups)
	return ups, tx.Error
}

func UserGetProcedureByID(id int) (model.Procedure, error) {
	var p model.Procedure
	tx := app.Database.DB.
		Preload("ProcedureSteps").
		Preload("ExtraFees").
		Omit("SystemStatus").
		Where("is_active = ?", true).
		First(&p, "id = ?", id)
	return p, tx.Error
}

func GetUserProcedureByID(userProcedureID int) (model.UserProcedure, error) {
	var up model.UserProcedure
	err := app.Database.DB.Model(&model.UserProcedure{}).
		Preload("UserProcedureSteps").
		Omit("Procedure").
		Where("id = ?", userProcedureID).
		First(&up).Error
	return up, err
}

func MarkUserProcedureHasTicket(userProcedureID int, has bool) error {
	return app.Database.DB.Model(&model.UserProcedure{}).
		Where("id = ?", userProcedureID).
		Updates(map[string]any{"is_have_ticket": has}).Error
}

func CreateUserProcedure(userProcedure model.UserProcedure) (model.UserProcedure, error) {
	if err := app.Database.DB.Create(&userProcedure).Error; err != nil {
		return userProcedure, err
	}
	userProcedure.ProcNumber = fmt.Sprintf("#LW%04d", userProcedure.ID)

	if userProcedure.PaymentAccount == nil || lo.FromPtr(userProcedure.PaymentAccount) == "" {
		userProcedure.CompanyCode = fmt.Sprintf("#CC%04d", userProcedure.ID)
	}

	if err := app.Database.DB.Model(&model.UserProcedure{}).
		Where("id = ?", userProcedure.ID).
		Updates(map[string]interface{}{
			"proc_number":  userProcedure.ProcNumber,
			"company_code": userProcedure.CompanyCode,
		}).Error; err != nil {
		return userProcedure, err
	}

	return userProcedure, nil
}

func UserCreateBulkProcedureStep(ups *[]model.UserProcedureStep) error {
	return app.Database.DB.Model(&model.UserProcedureStep{}).Create(ups).Error
}

func UserGetUserProceduresByUser(userID uuid.UUID, q Query) ([]model.UserProcedure, int64, error) {
	var (
		items []model.UserProcedure
		count int64
	)
	tx := app.Database.DB.Model(&model.UserProcedure{}).
		Preload("UserProcedureSteps", func(tx *gorm.DB) *gorm.DB { return tx.Select("id", "user_procedure_id", "status", "position", "name") }).
		Omit("SystemStatus").
		Where("user_id = ?", userID)
	if q.Status != "" {
		tx = tx.Where("status = ?", q.Status)
	}
	if q.StartTime != nil && q.EndTime != nil {
		tx = tx.Where("created_at >= ? AND created_at <= ?", q.StartTime, q.EndTime)
	}
	if q.Search != "" {
		tx = tx.Where("template_name ILIKE ? OR company_name ILIKE ? OR proc_number ILIKE ?", "%"+q.Search+"%", "%"+q.Search+"%", "%"+q.Search+"%")
	}
	if q.TemplateName != "" {
		tx = tx.Where("template_name = ?", q.TemplateName)
	}
	tx = tx.Count(&count)
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	tx = tx.Scopes(pagination(q.Page, q.Limit)).
		Order("created_at DESC").
		Find(&items)
	return items, count, tx.Error
}

func CreateUserProcedureStep(step *model.UserProcedureStep) error {
	return app.Database.DB.Model(&model.UserProcedureStep{}).Create(step).Error
}

func GetUserProcedureStepByID(userProcedureStepID int) (model.UserProcedureStep, error) {
	var ups model.UserProcedureStep
	err := app.Database.DB.Model(&model.UserProcedureStep{}).
		Where("id = ?", userProcedureStepID).
		First(&ups).Error
	return ups, err
}

func UpdateUserProcedureStep(ups *model.UserProcedureStep) error {
	return app.Database.DB.Model(&model.UserProcedureStep{}).
		Where("id = ?", ups.ID).
		Updates(ups).Error
}

func UserGetUserProcedureSteps(q Query, userID uuid.UUID) ([]model.UserProcedureStep, int64, error) {
	var (
		steps []model.UserProcedureStep
		count int64
	)
	tx := app.Database.DB.Model(&model.UserProcedureStep{}).
		Preload("UserProcedure", func(tx *gorm.DB) *gorm.DB { return tx.Select("id", "proc_number", "template_name", "company_name") }).
		Joins("JOIN user_procedures up ON up.id = user_procedure_steps.user_procedure_id").
		Where("up.user_id = ?", userID)
	if q.Status != "" {
		tx = tx.Where("user_procedure_steps.status = ?", q.Status)
	}
	if q.Search != "" {
		like := "%" + q.Search + "%"
		tx = tx.Where("user_procedure_steps.name ILIKE ? OR up.proc_number ILIKE ? OR up.template_name ILIKE ? OR up.company_name ILIKE ?",
			like, like, like, like)
	}
	if q.StartTime != nil && q.EndTime != nil {
		tx = tx.Where("user_procedure_steps.created_at >= ? AND user_procedure_steps.created_at <= ?", q.StartTime, q.EndTime)
	}
	tx = tx.Count(&count)
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	tx = tx.Scopes(pagination(q.Page, q.Limit)).
		Order("user_procedure_steps.created_at DESC").
		Find(&steps)
	return steps, count, tx.Error
}

// for cron job
func GetAllUserProceduresDueInDays() ([]model.UserProcedure, error) {
	now := time.Now()
	target := now.Add(48 * time.Hour)
	startOfDay := time.Date(target.Year(), target.Month(), target.Day(), 0, 0, 0, 0, target.Location())
	endOfDay := startOfDay.Add(24 * time.Hour).Add(-time.Nanosecond)

	var ups []model.UserProcedure
	tx := app.Database.DB.Model(&model.UserProcedure{}).
		Where("is_send_noti = ?", false).
		Where("status NOT IN ?", []string{model.USER_PROCEDURE_COMPLETED}).
		Where("deadline_at >= ? AND deadline_at <= ?", startOfDay, endOfDay).
		Find(&ups)
	if err := tx.Error; err != nil {
		return nil, err
	}
	return ups, nil
}

func MarkReminderSent(userProcedureID []int) error {
	return app.Database.DB.Transaction(func(tx *gorm.DB) error {
		return tx.Model(&model.UserProcedure{}).
			Where("id IN ?", userProcedureID).
			Updates(map[string]any{"is_send_noti": true}).Error
	})
}

func AdminMigrateStatus(dryRun bool) (domain.AdminMigrateStatusResponse, error) {
	var response domain.AdminMigrateStatusResponse

	if dryRun {
		var userProcCount int64
		err := app.Database.DB.Model(&model.UserProcedure{}).
			Where("status IN ?", []string{"processing", "error"}).
			Count(&userProcCount).Error
		if err != nil {
			return response, err
		}

		var userProcStepCount int64
		err = app.Database.DB.Model(&model.UserProcedureStep{}).
			Where("status IN ?", []string{"processing", "error"}).
			Count(&userProcStepCount).Error
		if err != nil {
			return response, err
		}

		response.UserProceduresUpdated = userProcCount
		response.UserProcedureStepsUpdated = userProcStepCount
		response.Message = fmt.Sprintf("DRY RUN: Would update %d user procedures and %d user procedure steps", userProcCount, userProcStepCount)
		return response, nil
	}

	tx := app.Database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Update UserProcedures: information-review -> in-progress
	userProcResult1 := tx.Model(&model.UserProcedure{}).
		Where("status = ?", "information-review").
		Update("status", model.USER_PROCEDURE_IN_PROGRESS)
	if userProcResult1.Error != nil {
		tx.Rollback()
		return response, userProcResult1.Error
	}

	// Update UserProcedures: error -> information-required
	userProcResult2 := tx.Model(&model.UserProcedure{}).
		Where("status = ?", "error").
		Update("status", model.USER_PROCEDURE_INFORMATION_REQUIRED)
	if userProcResult2.Error != nil {
		tx.Rollback()
		return response, userProcResult2.Error
	}

	// Update UserProcedures: pending -> created
	userProcResult3 := tx.Model(&model.UserProcedure{}).
		Where("status = ?", "pending").
		Update("status", model.USER_PROCEDURE_CREATED)
	if userProcResult3.Error != nil {
		tx.Rollback()
		return response, userProcResult3.Error
	}

	// Update UserProcedureSteps: information-review -> in-progress
	userStepResult1 := tx.Model(&model.UserProcedureStep{}).
		Where("status = ?", "information-review").
		Update("status", model.USER_PROCEDURE_STEP_IN_PROGRESS)
	if userStepResult1.Error != nil {
		tx.Rollback()
		return response, userStepResult1.Error
	}

	// Update UserProcedureSteps: error -> information-required
	userStepResult2 := tx.Model(&model.UserProcedureStep{}).
		Where("status = ?", "error").
		Update("status", model.USER_PROCEDURE_STEP_INFORMATION_REQUIRED)
	if userStepResult2.Error != nil {
		tx.Rollback()
		return response, userStepResult2.Error
	}

	// Update UserProcedureSteps: pending -> created
	userStepResult3 := tx.Model(&model.UserProcedureStep{}).
		Where("status = ?", "pending").
		Update("status", model.USER_PROCEDURE_STEP_CREATED)
	if userStepResult3.Error != nil {
		tx.Rollback()
		return response, userStepResult3.Error
	}

	if err := tx.Commit().Error; err != nil {
		return response, err
	}

	response.UserProceduresUpdated = userProcResult1.RowsAffected + userProcResult2.RowsAffected
	response.UserProcedureStepsUpdated = userStepResult1.RowsAffected + userStepResult2.RowsAffected
	response.Message = fmt.Sprintf("Successfully migrated %d user procedures and %d user procedure steps",
		response.UserProceduresUpdated, response.UserProcedureStepsUpdated)

	return response, nil
}

func CountUserProceduresByStatus(userID uuid.UUID, query Query) (map[string]int64, error) {
	statuses := []string{
		model.USER_PROCEDURE_INFORMATION_REQUIRED,
		model.USER_PROCEDURE_IN_PROGRESS,
		model.USER_PROCEDURE_CREATED,
		model.USER_PROCEDURE_COMPLETED,
	}

	result := map[string]int64{
		model.USER_PROCEDURE_INFORMATION_REQUIRED: 0,
		model.USER_PROCEDURE_IN_PROGRESS:          0,
		model.USER_PROCEDURE_CREATED:              0,
		model.USER_PROCEDURE_COMPLETED:            0,
	}

	type row struct {
		Status string
		Count  int64
	}

	var rows []row
	tx := app.Database.DB.Model(&model.UserProcedure{}).
		Select("status, COUNT(*) as count").
		Where("user_id = ? AND status IN ?", userID, statuses)

	if query.StartTime != nil && query.EndTime != nil {
		tx = tx.Where("created_at >= ? AND created_at <= ?", query.StartTime, query.EndTime)
	}
	tx = tx.Group("status").Scan(&rows)
	if tx.Error != nil {
		return nil, tx.Error
	}

	for _, r := range rows {
		result[r.Status] = r.Count
	}

	return result, nil
}
