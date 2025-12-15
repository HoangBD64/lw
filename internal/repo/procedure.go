package repo

import (
	"core/app"
	"core/internal/domain_model/model"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func AdminCreateProcedure(p *model.Procedure) error {
	return app.Database.DB.Create(p).Error
}

func AdminGetProcedureByID(id int) (model.Procedure, error) {
	var p model.Procedure
	err := app.Database.DB.First(&p, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return p, err
	}
	return p, err
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
	tx = tx.Count(&count)
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	tx = tx.Scopes(pagination(q.Page, q.Limit)).Order("created_at DESC").Find(&items)
	return items, count, tx.Error
}

func CreateProcedureStep(step model.ProcedureStep) (model.ProcedureStep, error) {
	tx := app.Database.DB.Model(&model.ProcedureStep{}).Create(&step)
	return step, tx.Error
}

func UpdateProcedureStep(step *model.ProcedureStep) error {
	return app.Database.DB.Model(&model.ProcedureStep{}).Where("id = ?", step.ID).Updates(step).Error
}

func DeleteProcedureStep(id int) error {
	return app.Database.DB.Delete(&model.ProcedureStep{}, "id = ?", id).Error
}

func ListProcedureSteps(procedureID int) ([]model.ProcedureStep, error) {
	var steps []model.ProcedureStep
	tx := app.Database.DB.Model(&model.ProcedureStep{}).
		Where("procedure_id = ?", procedureID).
		Order("position ASC").
		Find(&steps)
	return steps, tx.Error
}

func GetProcedureStepByID(id int) (model.ProcedureStep, error) {
	var step model.ProcedureStep
	tx := app.Database.DB.Model(&model.ProcedureStep{}).
		Where("id = ?", id).First(&step)
	return step, tx.Error
}

func UpsertUserProcedureStep(ups *model.UserProcedureStep) error {
	var existing model.UserProcedureStep
	err := app.Database.DB.Model(&model.UserProcedureStep{}).
		Where("user_procedure_id = ? AND procedure_step_id = ?", ups.UserProcedureID, ups.ProcedureStepID).
		First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return app.Database.DB.Create(ups).Error
	}
	if err != nil {
		return err
	}
	ups.ID = existing.ID
	return app.Database.DB.Model(&model.UserProcedureStep{}).Where("id = ?", existing.ID).Updates(ups).Error
}

func GetUserProcedureStep(userProcedureID int) ([]model.UserProcedureStep, error) {
	var ups []model.UserProcedureStep
	tx := app.Database.DB.Model(&model.UserProcedureStep{}).
		Where("user_procedure_id = ?", userProcedureID).
		Find(&ups)
	return ups, tx.Error
}

func ListUserProcedureStepsByUserProcedure(userProcedureID int) ([]model.UserProcedureStep, error) {
	var items []model.UserProcedureStep
	tx := app.Database.DB.Model(&model.UserProcedureStep{}).
		Where("user_procedure_id = ?", userProcedureID).
		Order("created_at DESC").
		Find(&items)
	return items, tx.Error
}

func GetProcedureByID(id int) (model.Procedure, error) {
	var p model.Procedure
	err := app.Database.DB.First(&p, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return p, err
	}
	return p, err
}

func GetUserProcedureByID(userID uuid.UUID, procedureID int) (model.UserProcedure, error) {
	var up model.UserProcedure
	err := app.Database.DB.Model(&model.UserProcedure{}).
		Preload("UserProcedureSteps").
		Omit("Procedure").
		Where("user_id = ? AND procedure_id = ?", userID, procedureID).
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
