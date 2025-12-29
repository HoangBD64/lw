package handler

import (
	"core/internal/domain_model/domain"
	"core/internal/domain_model/model"
	"core/internal/repo"
	"core/utils/base"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
)

func GetNotifications(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(model.User)
	if !ok {
		return c.JSON(fiber.Map{"status": false, "error": "Please login again"})
	}
	var query repo.Query
	query.Parse(c)
	notis, total, err := repo.GetNotifications(query, user)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't get notifications", Error: err})
	}
	return c.JSON(fiber.Map{"status": true, "message": "Success", "notifications": notis, "total": total})
}

func GetNotificationDetail(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(model.User)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again"})
	}
	var query repo.Query
	query.Parse(c)
	if query.IdInt == nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "id can't empty"})
	}
	noti, err := repo.GetNotificationDetails(*query.IdInt, user)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't GetNotificationDetails", Error: err})
	}
	if noti.ReadAt == nil {
		now := time.Now()
		noti.ReadAt = &now
		repo.UpdateNotification(noti, user)
	}
	return c.JSON(fiber.Map{"status": true, "message": "Success", "notification": noti})
}

func GetTotalNotificationType(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(model.User)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again"})
	}
	var query repo.Query
	query.Parse(c)
	total, err := repo.GetTotalNotificationType(query, user)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't GetTotalNotificationType", Error: err})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Total: &total})
}

func MarkReadAllNotification(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(model.User)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again"})
	}
	total, err := repo.UpdateReadAtAllNotification(user)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't UpdateReadAtAllNotification", Error: err})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Total: &total})
}

func AdminGetNotifications(c *fiber.Ctx) error {
	admin, ok := c.Locals("admin").(model.Admin)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again"})
	}
	isRoleAdmin := true
	if !admin.IsAdmin() {
		isRoleAdmin = false
	}

	var query repo.Query
	query.Parse(c)
	query.AdminID = &admin.ID
	notifications, total, err := repo.AdminGetNotifications(c.Context(), query, isRoleAdmin)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't get list notification", Error: err})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: notifications, Total: &total})
}

func AdminGetTotalNotificationType(c *fiber.Ctx) error {
	admin, ok := c.Locals("admin").(model.Admin)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again"})
	}

	var query repo.Query
	query.Parse(c)
	if admin.IsLawyer() {
		query.AdminID = &admin.ID
	}

	total, err := repo.AdminGetTotalNotificationType(c.Context(), query)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't GetTotalNotificationType", Error: err})
	}

	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Total: &total})
}

func AdminGetNotificationDetail(c *fiber.Ctx) error {
	if _, ok := c.Locals("admin").(model.Admin); !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again"})
	}

	var query repo.Query
	query.Parse(c)
	if query.IdInt == nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "id can't empty"})
	}

	notification, err := repo.AdminGetNotificationDetails(c.Context(), *query.IdInt)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't GetNotificationDetails", Error: err})
	}

	if notification.ReadAt == nil {
		now := time.Now()
		notification.ReadAt = &now
		if _, err := repo.AdminUpdateNotification(c.Context(), notification); err != nil {
			slog.WarnContext(c.Context(), "failed to update notification", "error", err)
		}
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: notification})
}

func AdminMarkReadAllNotification(c *fiber.Ctx) error {
	admin, ok := c.Locals("admin").(model.Admin)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again"})
	}
	isRoleAdmin := true
	if !admin.IsAdmin() {
		isRoleAdmin = false
	}

	if _, err := repo.AdminUpdateReadAtAllNotification(c.Context(), isRoleAdmin, admin.ID); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "failed to update notification read time", Error: err})
	}

	return c.JSON(base.BaseResponse{Status: true, Message: "Success"})
}

func UserUpsertNotificationPreference(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(model.User)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again", Error: nil})
	}

	var input domain.UserUpsertNotificationPreferenceRequest
	if err := c.BodyParser(&input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Review your input", Error: err})
	}

	pref, err := repo.UpsertNotificationPreferenceByUserID(user.ID, input.EmailEnabled)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't upsert notification preference", Error: err})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: pref})
}

func UserGetNotificationPreference(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(model.User)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again", Error: nil})
	}

	pref, err := repo.GetNotificationPreferenceByUserID(user.ID)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't get notification preference", Error: err})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: pref})
}
