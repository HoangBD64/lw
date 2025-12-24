package repo

import (
	"core/app"
	"core/internal/domain_model/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func AdminCreateProcedure(p *model.Procedure) error {
	return app.Database.DB.Create(p).Error
}

func AdminGetProcedureByID(id int) (model.Procedure, error) {
	var p model.Procedure
	tx := app.Database.DB.
		Preload("ProcedureSteps").
		First(&p, "id = ?", id)
	return p, tx.Error
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
	tx := app.Database.DB.Model(&model.Procedure{})
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

func GetUserProcedureStepsByUserProcedureAndStepIDs(userProcedureID int, stepIDs []int) ([]model.UserProcedureStep, error) {
	var ups []model.UserProcedureStep
	tx := app.Database.DB.Model(&model.UserProcedureStep{}).
		Where("user_procedure_id = ? AND id IN ?", userProcedureID, stepIDs).
		Find(&ups)
	return ups, tx.Error
}

func BulkAssignLawyerAdminToUserProcedureSteps(ups []model.UserProcedureStep) (int64, error) {
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

func GetUserProcedureDetail(userProcedureID int, userID uuid.UUID) ([]model.UserProcedure, error) {
	var ups []model.UserProcedure
	tx := app.Database.DB.Model(&model.UserProcedure{}).
		Preload("UserProcedureSteps").
		Where("id = ? AND user_id = ?", userProcedureID, userID).
		Find(&ups)
	return ups, tx.Error
}

func AdminGetUserProcedureByID(userProcedureID int) (model.UserProcedure, error) {
	var up model.UserProcedure
	err := app.Database.DB.Model(&model.UserProcedure{}).
		Where("id = ?", userProcedureID).
		First(&up).Error
	return up, err
}

func AdminListUserProcedures(q Query) ([]model.UserProcedure, int64, error) {
	var (
		items []model.UserProcedure
		count int64
	)
	tx := app.Database.DB.Model(&model.UserProcedure{}).
		Omit("Procedure", "UserProcedureSteps", "Admin").
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
	if q.AdminID != nil {
		tx = tx.Where("admin_id = ?", q.AdminID)
	}
	if q.ProcedureID != nil {
		tx = tx.Where("procedure_id = ?", q.ProcedureID)
	}
	if q.Search != "" {
		tx = tx.Where("name ILIKE ?", "%"+q.Search+"%")
	}
	if q.IsHaveTicket != nil {
		tx = tx.Where("is_have_ticket = ?", q.IsHaveTicket)
	}
	if q.IsAssigned != nil {
		if *q.IsAssigned {
			tx = tx.Where("admin_id IS NOT NULL")
		} else {
			tx = tx.Where("admin_id IS NULL AND status = ?", model.USER_PROCEDURE_PENDING)
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

func GetProcedureByID(id int) (model.Procedure, error) {
	var p model.Procedure
	tx := app.Database.DB.
		Preload("ProcedureSteps").
		Where("is_active = ?", true).
		First(&p, "id = ?", id)
	return p, tx.Error
}

func GetUserProcedureByProcedureID(userID uuid.UUID, procedureID int) (model.UserProcedure, error) {
	var up model.UserProcedure
	err := app.Database.DB.Model(&model.UserProcedure{}).
		Preload("UserProcedureSteps").
		Omit("Procedure").
		Where("user_id = ? AND procedure_id = ?", userID, procedureID).
		First(&up).Error
	return up, err
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

func AssignAdminToUserProcedures(userProcedureIDs []int, adminID uuid.UUID) (int64, error) {
	tx := app.Database.DB.Model(&model.UserProcedure{}).
		Where("id IN ? AND status = ?", userProcedureIDs, model.USER_PROCEDURE_PENDING).
		Updates(map[string]any{
			"admin_id": adminID,
			"status":   model.USER_PROCEDURE_PROCESSING,
		})

	tx = app.Database.DB.Model(&model.UserProcedureStep{}).
		Where("user_procedure_id IN ? AND status = ?", userProcedureIDs, model.USER_PROCEDURE_STEP_PENDING).
		Updates(map[string]any{
			"status": model.USER_PROCEDURE_STEP_PROCESSING,
		})
	return tx.RowsAffected, tx.Error
}

func UnassignAdminFromUserProcedure(userProcedureID int) error {
	return app.Database.DB.Model(&model.UserProcedure{}).
		Where("id = ?", userProcedureID).
		Updates(map[string]any{"admin_id": nil}).Error
}

func MarkUserProcedureHasTicket(userProcedureID int, has bool) error {
	return app.Database.DB.Model(&model.UserProcedure{}).
		Where("id = ?", userProcedureID).
		Updates(map[string]any{"is_have_ticket": has}).Error
}

func AdminGetUserProcedureDetailByID(userProcedureID int) (model.UserProcedure, error) {
	var up model.UserProcedure
	err := app.Database.DB.Model(&model.UserProcedure{}).
		Preload("UserProcedureSteps").
		Preload("Admin", func(dbx *gorm.DB) *gorm.DB {
			return dbx.Model(&model.Admin{}).Select("id", "name", "email", "role", "avatar", "status")
		}).
		Where("id = ?", userProcedureID).
		First(&up).Error
	return up, err
}

func CreateUserProcedure(userProcedure model.UserProcedure) (model.UserProcedure, error) {
	tx := app.Database.DB.Create(&userProcedure)
	return userProcedure, tx.Error
}

func UserCreateBulkProcedureStep(ups *[]model.UserProcedureStep) error {
	return app.Database.DB.Model(&model.UserProcedureStep{}).Create(ups).Error
}

func ListUserProceduresByUser(userID uuid.UUID, q Query) ([]model.UserProcedure, int64, error) {
	var (
		items []model.UserProcedure
		count int64
	)
	tx := app.Database.DB.Model(&model.UserProcedure{}).
		Preload("UserProcedureSteps", func(tx *gorm.DB) *gorm.DB { return tx.Select("id", "user_procedure_id", "status", "position", "name") }).
		Where("user_id = ?", userID)
	if q.Status != "" {
		tx = tx.Where("status = ?", q.Status)
	}
	if q.StartTime != nil && q.EndTime != nil {
		tx = tx.Where("created_at >= ? AND created_at <= ?", q.StartTime, q.EndTime)
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

func AdminUpdateUserProcedure(up *model.UserProcedure) error {
	return app.Database.DB.Model(&model.UserProcedure{}).
		Where("id = ?", up.ID).
		Updates(up).Error
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
