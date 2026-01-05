package repo

import (
	"core/app"
	"core/internal/domain_model/model"
)

func CreateExtraFee(extraFee *model.ExtraFee) error {
	tx := app.Database.DB.Create(extraFee)
	return tx.Error
}

func GetExtraFees(input Query) ([]model.ExtraFee, error) {
	var extraFees []model.ExtraFee
	tx := app.Database.DB.
		Where("procedure_id = ?", input.ProcedureID).
		Find(&extraFees)
	return extraFees, tx.Error
}

func GetExtraFeeByID(id int) (model.ExtraFee, error) {
	extraFee := model.ExtraFee{}
	tx := app.Database.DB.Where("id = ?", id).First(&extraFee)
	return extraFee, tx.Error
}

func UpdateExtraFee(extraFee *model.ExtraFee) error {
	tx := app.Database.DB.Save(extraFee)
	return tx.Error
}

func DeleteExtraFee(id int) error {
	tx := app.Database.DB.Delete(&model.ExtraFee{}, id)
	return tx.Error
}
