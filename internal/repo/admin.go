package repo

import (
	"context"
	"core/app"
	"core/internal/domain_model/model"

	"github.com/google/uuid"
)

func GetAdminBySsoID(id string) (model.Admin, error) {
	admin := model.Admin{}
	tx := app.Database.DB.Where("sso_id = ?", id).First(&admin)
	return admin, tx.Error
}

func CreateAdmin(admin model.Admin) (model.Admin, error) {
	var count int64
	app.Database.DB.Model(&model.Admin{}).Count(&count)
	if count < 2 {
		admin.Status = model.ACTIVE
		admin.IsOwner = true
	}
	tx := app.Database.DB.Model(&model.Admin{}).Create(&admin)
	return admin, tx.Error
}

func GetAdminByID(ctx context.Context, id uuid.UUID) (model.Admin, error) {
	admin := model.Admin{}
	tx := app.Database.DB.Where("id = ?", id).First(&admin)
	return admin, tx.Error
}

func GetAdmin(query Query) ([]model.Admin, int64, error) {
	var admins []model.Admin
	tx := app.Database.DB.Model(&model.Admin{}).Select("id", "created_at", "updated_at", "name", "email", "sso_id", "avatar", "role", "status", "last_login", "is_owner", "permissions")
	if query.Search != "" {
		tx = tx.Where("name ILIKE ? OR  email ILIKE ? ", "%"+query.Search+"%", "%"+query.Search+"%")
	}
	if query.Status != "" {
		tx = tx.Where("status = ?", query.Status)
	}
	if query.Role != "" {
		tx = tx.Where("role = ?", query.Role)
	}
	if query.AdminID != nil {
		tx = tx.Where("id = ?", query.AdminID)
	}
	if query.StartTime != nil && query.EndTime != nil {
		tx = tx.Where("created_at >= ? AND created_at <= ?", query.StartTime, query.EndTime)
	}

	var count int64
	err := tx.Count(&count)
	err = tx.Offset(query.Limit * (query.Page - 1)).Limit(query.Limit).Order("created_at DESC").Find(&admins)
	return admins, count, err.Error
}

func GetAdminByIds(ids []uuid.UUID) ([]model.Admin, error) {
	admins := []model.Admin{}
	tx := app.Database.DB.Where("id IN ?", ids).Find(&admins)
	return admins, tx.Error
}

func UpdateAdminLoginTime(admin model.Admin) error {
	result := app.Database.DB.Select("last_login").Updates(&admin)
	return result.Error
}

func UpdateAdminStatus(ids []uuid.UUID, status string) (int64, error) {
	tx := app.Database.DB.Model(&model.Admin{}).Where("id IN ?", ids).Select("Status").Updates(&model.Admin{
		Status: status,
	})
	return tx.RowsAffected, tx.Error
}

func UpdateAdminPermissions(ctx context.Context, admin model.Admin) (model.Admin, error) {
	tx := app.Database.DB.Select("permissions").Updates(&admin)
	if err := tx.Error; err != nil {
		return model.Admin{}, err
	}
	return admin, tx.Error
}

func UpdateAdminRole(ctx context.Context, admin model.Admin) (model.Admin, error) {
	tx := app.Database.DB.Model(&model.Admin{}).Where("id = ?", admin.ID).Update("role", admin.Role)
	if tx.Error != nil {
		return admin, tx.Error
	}

	return admin, nil
}

func GetAllAdmin(ctx context.Context, query Query) ([]model.Admin, error) {
	var admins []model.Admin

	selectColumns := []string{
		"id",
		"created_at",
		"updated_at",
		"name",
		"email",
		"sso_id",
		"avatar",
		"role",
		"status",
		"last_login",
		"is_owner",
		"permissions",
	}

	tx := app.Database.DB.Model(&model.Admin{}).Select(selectColumns)
	if query.AdminID != nil {
		tx = tx.Where("id = ?", query.AdminID)
	}
	if query.Search != "" {
		tx = tx.Where("name ILIKE ? OR  email ILIKE ? ", "%"+query.Search+"%", "%"+query.Search+"%")
	}
	if query.Status != "" {
		tx = tx.Where("status = ?", query.Status)
	}
	if query.Role != "" {
		tx = tx.Where("role = ?", query.Role)
	}
	if query.AdminID != nil {
		tx = tx.Where("id = ?", query.AdminID)
	}
	if query.StartTime != nil && query.EndTime != nil {
		tx = tx.Where("created_at >= ? AND created_at <= ?", query.StartTime, query.EndTime)
	}

	var count int64
	tx = tx.Count(&count)
	tx = tx.Offset(query.Limit * (query.Page - 1)).Limit(query.Limit).Order("created_at DESC").Find(&admins)

	if err := tx.Error; err != nil {
		return nil, err
	}

	return admins, nil
}
