package repo

import (
	"context"
	"core/app"
	"core/internal/domain_model/model"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func CreateNotification(ctx context.Context, noti model.Notification) (model.Notification, error) {
	tx := app.Database.DB.Create(&noti)
	if err := tx.Error; err != nil {
		return noti, tx.Error
	}
	return noti, tx.Error
}

func GetNotifications(query Query, user model.User) ([]model.Notification, int64, error) {
	var notis []model.Notification
	tx := app.Database.DB.Model(&model.Notification{}).Where("user_id = ?", user.ID)
	if query.Search != "" {
		tx = tx.Where("name ILIKE ? ", "%"+query.Search+"%")
	}
	if query.Type != "" {
		tx = tx.Where("type = ?", query.Type)
	}
	var count int64
	err := tx.Count(&count)
	err = tx.Offset(query.Limit * (query.Page - 1)).Limit(query.Limit).Order("created_at DESC").Find(&notis)
	return notis, count, err.Error
}

func GetNotificationDetails(id uuid.UUID, user model.User) (model.Notification, error) {
	var noti model.Notification
	tx := app.Database.DB.Model(&model.Notification{}).Where("user_id = ?", user.ID).Where("id = ?", id).First(&noti)
	return noti, tx.Error
}

func DeleteNotification(ids []uuid.UUID, user model.User) (int64, error) {
	result := app.Database.DB.Where("user_id = ?", user.ID).Delete(&model.Notification{}, ids)
	return result.RowsAffected, result.Error
}

func UpdateNotification(noti model.Notification, user model.User) (model.Notification, error) {
	tx := app.Database.DB.Where("user_id = ?", user.ID).Updates(&noti)
	return noti, tx.Error
}

func GetTotalNotificationType(query Query, user model.User) (int64, error) {
	tx := app.Database.DB.Model(&model.Notification{}).Where("user_id = ?", user.ID)
	if query.Type != "" {
		tx = tx.Where("type = ?", query.Type)
	}
	switch query.Status {
	case "read":
		tx = tx.Where("read_at IS NOT NULL")
	case "unread":
		tx = tx.Where("read_at IS NULL")
	}
	var count int64
	err := tx.Count(&count)
	return count, err.Error
}

func UpdateReadAtAllNotification(user model.User) (int64, error) {
	now := time.Now()
	tx := app.Database.DB.Model(&model.Notification{}).Where("user_id = ?", user.ID).Where("read_at IS NULL").Select("ReadAt").Updates(&model.Notification{
		ReadAt: &now,
	})
	return tx.RowsAffected, tx.Error
}

func AdminGetNotifications(ctx context.Context, query Query) ([]model.Notification, int, error) {
	var out []model.Notification
	var count int64

	tx := app.Database.DB.Model(&model.Notification{}).Where("user_id IS NULL")

	if query.Type != "" {
		tx = tx.Where("type = ?", query.Type)
	}

	if err := tx.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	if err := tx.Offset(query.Limit * (query.Page - 1)).Limit(query.Limit).Order("created_at DESC").Find(&out).Error; err != nil {
		return nil, 0, err
	}

	return out, int(count), nil
}

func AdminGetNotificationDetails(ctx context.Context, id uuid.UUID) (model.Notification, error) {
	var notification model.Notification

	tx := app.Database.DB.Model(&model.Notification{}).Where("user_id IS NULL").Where("id = ?", id).First(&notification)
	if err := tx.Error; err != nil {
		return model.Notification{}, err
	}

	return notification, nil
}

func AminUpdateReadAtAllNotification(ctx context.Context) (int, error) {
	now := time.Now()
	tx := app.Database.DB.Model(&model.Notification{}).Where("user_id IS NULL").Where("read_at IS NULL").Select("ReadAt").Updates(&model.Notification{ReadAt: &now})

	if err := tx.Error; err != nil {
		return 0, err
	}

	return int(tx.RowsAffected), nil
}

func AdminGetTotalNotificationType(ctx context.Context, query Query) (int, error) {
	var count int64
	tx := app.Database.DB.Model(&model.Notification{}).Where("user_id IS NULL")

	if query.Type != "" {
		tx = tx.Where("type = ?", query.Type)
	}

	switch query.Status {
	case "read":
		tx = tx.Where("read_at IS NOT NULL")
	case "unread":
		tx = tx.Where("read_at IS NULL")
	}

	if err := tx.Count(&count).Error; err != nil {
		return 0, err
	}

	return int(count), nil
}

func AdminUpdateNotification(ctx context.Context, noti model.Notification) (model.Notification, error) {
	if err := app.Database.DB.Where("user_id IS NULL").Updates(&noti).Error; err != nil {
		return model.Notification{}, err
	}

	return noti, nil
}

func GetNotificationPreferenceByUserID(userID uuid.UUID) (model.NotificationPreference, error) {
	var pref model.NotificationPreference
	tx := app.Database.DB.Model(&model.NotificationPreference{}).Where("user_id = ?", userID).First(&pref)
	return pref, tx.Error
}

func UpsertNotificationPreferenceByUserID(userID uuid.UUID, emailEnabled bool) (model.NotificationPreference, error) {
	var out model.NotificationPreference
	err := app.Database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.NotificationPreference{}).Where("user_id = ?", userID).First(&out).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				out = model.NotificationPreference{UserID: &userID, EmailEnabled: emailEnabled}
				if err := tx.Create(&out).Error; err != nil {
					return err
				}
				return nil
			}
			return err
		}
		out.EmailEnabled = emailEnabled
		if err := tx.Model(&out).Select("email_enabled").Updates(&out).Error; err != nil {
			return err
		}
		return nil
	})
	return out, err
}
