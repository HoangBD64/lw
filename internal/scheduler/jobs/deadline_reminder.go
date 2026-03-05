package jobs

import (
	"context"
	templates "core/internal/common/templates"
	"core/internal/domain_model/model"
	"core/internal/notification"
	"core/internal/repo"
	"core/utils"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/samber/lo"
)

type DeadlineReminderJob struct {
	notify *notification.Service
}

func NewDeadlineReminderJob(service *notification.Service) *DeadlineReminderJob {
	return &DeadlineReminderJob{notify: service}
}

func (j *DeadlineReminderJob) Run(ctx context.Context) error {
	ups, err := repo.GetAllUserProceduresDueInDays()
	if err != nil {
		return err
	}
	upsIDs := make([]int, 0, len(ups))
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
		if up.ProcNumber != "" {
			procNumber = up.ProcNumber
		}
		title := fmt.Sprintf("Upcoming deadline in 2 days: %s", procNumber)

		websiteURL := strings.TrimSpace(utils.GetWebBaseURL())
		data := templates.DeadlineReminderData{
			UserName:            user.Name,
			ProcedureName:       up.TemplateName,
			CompanyName:         lo.FromPtr(up.CompanyName),
			DueDate:             deadlineStr,
			DaysLeft:            2,
			UserProcedureNumber: procNumber,
			URL:                 websiteURL,
		}

		bodyHTML, tmplErr := templates.RenderMailTemplate(templates.MailTemplateDeadlineReminder, data)
		if tmplErr != nil {
			slog.Debug("deadline job: render template failed, using plain text", "error", tmplErr)
			bodyHTML = fmt.Sprintf("Your procedure %s is due on %s.", procNumber, deadlineStr)
		}

		userContent := fmt.Sprintf("Your procedure <strong>%s</strong> is due on <strong>%s</strong>.", procNumber, deadlineStr)
		adminContent := fmt.Sprintf("Procedure <strong>%s</strong> is due on <strong>%s</strong>.", procNumber, deadlineStr)
		// user app noti
		notis := []model.Notification{}
		userNoti := model.Notification{
			UserID: up.UserID,
			Title:  title,
			Metadata: model.NotiMetadata{
				UserProcedureID: &up.ID,
				Content:         &userContent,
			},
			Type: model.NOTI_DEADLINE_REMINDER,
		}
		notis = append(notis, userNoti)

		// admin app noti
		var adminID *uuid.UUID = up.AdminID
		adminNoti := model.Notification{
			AdminID: adminID,
			Title:   title,
			Metadata: model.NotiMetadata{
				UserProcedureID: &up.ID,
				Content:         &adminContent,
			},
			Type: model.NOTI_DEADLINE_REMINDER,
		}
		notis = append(notis, adminNoti)

		if err := j.notify.CreateBatchNotifyInApp(notis); err != nil {
			slog.Error("deadline job: failed to create admin notification", "up_id", up.ID, "error", err)
		}

		// send mail
		if err := j.notify.SendEmailIfEnabled(user, title, bodyHTML); err != nil {
			slog.Debug("deadline job: email skipped", "reason", err)
		}
		upsIDs = append(upsIDs, up.ID)
	}
	// Mark as sent to avoid duplicates
	if err := repo.MarkReminderSent(upsIDs); err != nil {
		slog.Error("deadline job: failed to mark reminder sent", "up_ids", upsIDs, "error", err)
	}

	return nil
}
