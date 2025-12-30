package jobs

import (
	"context"
	"core/app"
	"core/internal/domain_model/model"
	"core/internal/notification"
	"core/internal/repo"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

type DeadlineReminderJob struct {
	notify *notification.Service
}

func NewDeadlineReminderJob(service *notification.Service) *DeadlineReminderJob {
	return &DeadlineReminderJob{notify: service}
}

func (j *DeadlineReminderJob) Run(ctx context.Context) error {
	now := time.Now()
	target := now.Add(48 * time.Hour)
	startOfDay := time.Date(target.Year(), target.Month(), target.Day(), 0, 0, 0, 0, target.Location())
	endOfDay := startOfDay.Add(24 * time.Hour).Add(-time.Nanosecond)

	var ups []model.UserProcedure
	tx := app.Database.DB.WithContext(ctx).Model(&model.UserProcedure{}).
		Where("deadline_at >= ? AND deadline_at <= ?", startOfDay, endOfDay).
		Find(&ups)
	if err := tx.Error; err != nil {
		return err
	}

	for _, up := range ups {
		if up.UserID == nil {
			continue
		}
		user, err := repo.GetUserByID(*up.UserID)
		if err != nil {
			slog.Error("deadline job: failed to get user", "user_id", up.UserID, "error", err)
			continue
		}

		if up.DeadlineAt == nil {
			continue
		}
		deadlineStr := up.DeadlineAt.Format("2006-01-02")
		procNumber := ""
		if up.ProcNumber != nil {
			procNumber = *up.ProcNumber
		}
		title := fmt.Sprintf("Upcoming deadline in 2 days: %s", procNumber)
		content := fmt.Sprintf("Your procedure %s is due on %s.", procNumber, deadlineStr)

		md := model.NotiMetadata{UserProcedureID: &up.ID, Content: &content}
		mdJSON, _ := json.Marshal(md)
		userNoti := model.Notification{
			UserID:   up.UserID,
			Title:    title,
			Metadata: mdJSON,
			Type:     model.NOTI_DEADLINE_REMINDER,
		}
		if err := j.notify.CreateNotifyInApp(&userNoti); err != nil {
			slog.Error("deadline job: failed to create user notification", "up_id", up.ID, "error", err)
		}

		// Admin notification: targeted if AdminID exists, otherwise broadcast to admin channel
		var adminID *uuid.UUID = up.AdminID
		adminNoti := model.Notification{
			AdminID:  adminID,
			Title:    title,
			Metadata: mdJSON,
			Type:     model.NOTI_DEADLINE_REMINDER,
		}
		if err := j.notify.CreateNotifyInApp(&adminNoti); err != nil {
			slog.Error("deadline job: failed to create admin notification", "up_id", up.ID, "error", err)
		}

		if err := j.notify.SendEmailIfEnabled(user, title, content); err != nil {
			slog.Debug("deadline job: email skipped", "reason", err)
		}

		// Mark as sent to avoid duplicates
		// if err := markReminderSent(up.ID); err != nil {
		// 	slog.Error("deadline job: failed to mark reminder sent", "up_id", up.ID, "error", err)
		// }
	}

	return nil
}

// func markReminderSent(userProcedureID int) error {
// 	return app.Database.DB.Transaction(func(tx *gorm.DB) error {
// 		return tx.Model(&model.UserProcedure{}).
// 			Where("id = ?", userProcedureID).
// 			Updates(map[string]any{"is_send_noti": true}).Error
// 	})
// }
