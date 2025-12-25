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
		return c.JSON(fiber.Map{"status": false, "message": "Can't get list notification", "error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": true, "message": "Success", "notifications": notis, "total": total})
}

func GetNotificationDetail(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(model.User)
	if !ok {
		return c.JSON(fiber.Map{"status": false, "error": "Please login again"})
	}
	var query repo.Query
	query.Parse(c)
	if query.IdInt == nil {
		return c.JSON(fiber.Map{"status": false, "error": "id can't empty"})
	}
	noti, err := repo.GetNotificationDetails(*query.IdInt, user)
	if err != nil {
		return c.JSON(fiber.Map{"status": false, "message": "Can't GetNotificationDetails", "error": err.Error()})
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
		return c.JSON(fiber.Map{"status": false, "error": "Please login again"})
	}
	var query repo.Query
	query.Parse(c)
	total, err := repo.GetTotalNotificationType(query, user)
	if err != nil {
		return c.JSON(fiber.Map{"status": false, "message": "Can't GetTotalNotificationType", "error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": true, "message": "Success", "total": total})
}

func MarkReadAllNotification(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(model.User)
	if !ok {
		return c.JSON(fiber.Map{"status": false, "error": "Please login again"})
	}
	total, err := repo.UpdateReadAtAllNotification(user)
	if err != nil {
		return c.JSON(fiber.Map{"status": false, "message": "Can't UpdateReadAtAllNotification", "error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": true, "message": "Success", "total": total})
}

func AdminGetNotifications(c *fiber.Ctx) error {
	if _, ok := c.Locals("admin").(model.Admin); !ok {
		return c.JSON(fiber.Map{"status": false, "error": "Please login again"})
	}

	var query repo.Query
	query.Parse(c)

	notifications, total, err := repo.AdminGetNotifications(c.Context(), query)
	if err != nil {
		return c.JSON(fiber.Map{"status": false, "message": "Can't get list notification", "error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": true, "message": "Success", "data": notifications, "total": total})
}

func AdminGetTotalNotificationType(c *fiber.Ctx) error {
	if _, ok := c.Locals("admin").(model.Admin); !ok {
		return c.JSON(fiber.Map{"status": false, "error": "Please login again"})
	}

	var query repo.Query
	query.Parse(c)

	total, err := repo.AdminGetTotalNotificationType(c.Context(), query)
	if err != nil {
		return c.JSON(fiber.Map{"status": false, "message": "Can't GetTotalNotificationType", "error": err.Error()})
	}

	return c.JSON(fiber.Map{"status": true, "message": "Success", "total": total})
}

func AdminGetNotificationDetail(c *fiber.Ctx) error {
	if _, ok := c.Locals("admin").(model.Admin); !ok {
		return c.JSON(fiber.Map{"status": false, "error": "Please login again"})
	}

	var query repo.Query
	query.Parse(c)
	if query.IdInt == nil {
		return c.JSON(fiber.Map{"status": false, "error": "id can't empty"})
	}

	notification, err := repo.AdminGetNotificationDetails(c.Context(), *query.IdInt)
	if err != nil {
		return c.JSON(fiber.Map{"status": false, "message": "Can't GetNotificationDetails", "error": err.Error()})
	}

	if notification.ReadAt == nil {
		now := time.Now()
		notification.ReadAt = &now
		if _, err := repo.AdminUpdateNotification(c.Context(), notification); err != nil {
			slog.WarnContext(c.Context(), "failed to update notification", "error", err)
		}
	}
	return c.JSON(fiber.Map{"status": true, "message": "Success", "data": notification})
}

func AdminMarkReadAllNotification(c *fiber.Ctx) error {
	if _, ok := c.Locals("admin").(model.Admin); !ok {
		return c.JSON(fiber.Map{"status": false, "error": "Please login again"})
	}

	if _, err := repo.AminUpdateReadAtAllNotification(c.Context()); err != nil {
		return c.JSON(fiber.Map{"status": false, "message": "failed to update notification read time", "error": err.Error()})
	}

	return c.JSON(fiber.Map{"status": true, "message": "Success"})
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
