package repo

import (
	"context"
	"core/app"
	"core/internal/domain_model/model"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func GetUserBySsoID(id string) (model.User, error) {
	user := model.User{}
	tx := app.Database.DB.Where("sso_id = ?", id).First(&user)
	return user, tx.Error
}

func CreateUser(user model.User) (model.User, error) {
	tx := app.Database.DB.Model(&model.User{}).Create(&user)
	return user, tx.Error
}

func GetUserByID(id uuid.UUID) (model.User, error) {
	user := model.User{}
	tx := app.Database.DB.Where("id = ?", id).First(&user)
	return user, tx.Error
}

func AdminGetUser(query Query) ([]model.User, int64, error) {
	var pa []model.User
	tx := app.Database.DB.Model(&model.User{})
	if query.Search != "" {
		tx = tx.Where("name ILIKE ? OR  email ILIKE ? ", "%"+query.Search+"%", "%"+query.Search+"%")
	}
	if query.Status != "" {
		tx = tx.Where("status = ? ", query.Status)
	}
	if query.StartTime != nil && query.EndTime != nil {
		tx = tx.Where("created_at >= ? AND created_at <= ?", query.StartTime, query.EndTime)
	}
	var count int64
	err := tx.Count(&count)
	err = tx.Offset(query.Limit * (query.Page - 1)).Limit(query.Limit).Order("created_at DESC").Find(&pa)
	return pa, count, err.Error
}

func GetUsersByAdmin(ctx context.Context, query Query) ([]model.User, error) {
	var pa []model.User
	tx := app.Database.DB.Model(&model.User{})
	if len(query.IDS) > 0 {
		tx = tx.Where("id IN ?", query.IDS)
	}
	err := tx.Find(&pa)

	return pa, err.Error
}

func UpdateUserInfo(user model.User) (model.User, error) {
	result := app.Database.DB.Select("name", "phone", "avatar", "gender", "dob", "address").Updates(&user)
	return user, result.Error
}

func AdminUpdateUser(ctx context.Context, user model.User) (model.User, error) {
	tx := app.Database.DB.Where("id = ?", user.ID).Updates(&user)
	if err := tx.Error; err != nil {
		return model.User{}, err
	}
	if tx.RowsAffected == 0 {
		return model.User{}, gorm.ErrRecordNotFound
	}

	return user, nil
}

func UpdateUserLoginTime(user model.User) error {
	result := app.Database.DB.Select("last_login").Updates(&user)
	return result.Error
}

func UpdateUserStatus(ids []uuid.UUID, status string) (int64, error) {
	tx := app.Database.DB.Model(&model.User{}).Where("id IN ?", ids).Select("Status").Updates(&model.User{
		Status: status,
	})
	return tx.RowsAffected, tx.Error
}

func GetUser(query Query, user model.User) ([]model.User, int64, error) {
	var users []model.User
	tx := app.Database.DB.Model(&model.User{}).Where("id IN ?", query.IDS).Select("id", "created_at", "updated_at", "name", "email", "sso_id", "avatar").Preload("StoreUser", func(tx *gorm.DB) *gorm.DB {
		tx = tx.Select("user_id", "role")
		return tx
	})
	if query.Search != "" {
		tx = tx.Where("name ILIKE ? OR  email ILIKE ? ", "%"+query.Search+"%", "%"+query.Search+"%")
	}
	var count int64
	err := tx.Count(&count)
	err = tx.Offset(query.Limit * (query.Page - 1)).Limit(query.Limit).Order("created_at DESC").Find(&users)
	return users, count, err.Error
}

func GetUserByEmail(email string) (model.User, error) {
	user := model.User{}
	tx := app.Database.DB.Where("email = ?", email).First(&user)
	return user, tx.Error
}

func UpdateUserDisplayName(user model.User) error {
	result := app.Database.DB.Select("name").Updates(&user)
	return result.Error
}

func GetReportUserCount(ctx context.Context, startTime, endTime time.Time) (int, error) {
	var count int64

	tx := app.Database.DB.WithContext(ctx).Model(&model.User{}).Where("created_at BETWEEN ? AND ?", startTime, endTime).Count(&count)

	if err := tx.Error; err != nil {
		return 0, err
	}

	return int(count), tx.Error
}
