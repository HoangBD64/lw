package repo

import (
	"context"
	"core/app"
	"core/internal/domain_model/model"

	"github.com/google/uuid"
)

func GetSettingSystemByType(type_setting string) (model.SystemSetting, error) {
	setting := model.SystemSetting{}
	tx := app.Database.DB.Where("type = ?", type_setting).First(&setting)
	if tx.Error != nil {
		return setting, tx.Error
	}

	return setting, nil
}

func GetSystemSetting(ctx context.Context, type_system string) (model.SystemSetting, error) {
	var setting model.SystemSetting

	tx := app.Database.DB.Model(&model.SystemSetting{}).Where("type = ? ", type_system).First(&setting)
	if tx.Error != nil {
		return setting, tx.Error
	}
	return setting, nil
}

func CreateSystemSetting(setting model.SystemSetting) (model.SystemSetting, error) {
	tx := app.Database.DB.Create(&setting)
	return setting, tx.Error
}

func GetSystemSettingByID(id uuid.UUID) (model.SystemSetting, error) {
	var setting model.SystemSetting
	tx := app.Database.DB.Model(&model.SystemSetting{}).Where("id = ? ", id)
	err := tx.First(&setting)
	return setting, err.Error
}

func UpdateSystemSetting(setting model.SystemSetting) (model.SystemSetting, error) {
	tx := app.Database.DB.Updates(&setting)
	return setting, tx.Error
}
