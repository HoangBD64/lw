package notification

import (
	"context"
	cmail "core/internal/common/mail"
	templates "core/internal/common/templates"
	"core/internal/domain_model/model"
	"core/internal/repo"
	"core/utils"
	"fmt"
	"log/slog"
	"strings"
)

func SendProcedureCreatedNoticeToAdmins(userProcedureID int) error {
	up, user, admins, err := loadProcedureMailContext(userProcedureID)
	if err != nil {
		return err
	}
	if len(admins) == 0 {
		return nil
	}

	data := baseInternalMailData(up, user)
	data.Badge = "New Procedure"
	data.Intro = fmt.Sprintf("Client %s has created procedure %s and it is ready for admin review.", user.Name, procedureLabel(up))
	subject := fmt.Sprintf("New procedure created: %s", procedureLabel(up))

	return sendMailToAdmins(admins, subject, templates.MailTemplateAdminProcedureNotice, data)
}

func SendSubmitReviewNoticeToAdmins(userProcedureID int, lawyerName string) error {
	up, user, admins, err := loadProcedureMailContext(userProcedureID)
	if err != nil {
		return err
	}
	if len(admins) == 0 {
		return nil
	}

	data := baseInternalMailData(up, user)
	data.Badge = "Review Submitted"
	data.AssignedLawyerName = lawyerName
	data.Intro = fmt.Sprintf("Lawyer %s has submitted procedure %s for admin review.", lawyerName, procedureLabel(up))
	subject := fmt.Sprintf("Procedure submitted for review: %s", procedureLabel(up))

	return sendMailToAdmins(admins, subject, templates.MailTemplateAdminProcedureNotice, data)
}

func SendProcedureAssignedNoticeToLawyer(userProcedureID int, lawyer model.Admin, assignedBy string) error {
	up, user, err := loadUserProcedureAndUser(userProcedureID)
	if err != nil {
		return err
	}

	data := baseInternalMailData(up, user)
	data.RecipientName = lawyer.Name
	data.Badge = "Assignment"
	data.AssignedLawyerName = lawyer.Name
	data.Intro = fmt.Sprintf("Procedure %s has been assigned to you by %s.", procedureLabel(up), assignedBy)
	subject := fmt.Sprintf("Procedure assigned to you: %s", procedureLabel(up))

	return sendMailToLawyer(lawyer, subject, templates.MailTemplateLawyerProcedureNotice, data)
}

func SendInfoRequiredResponseNoticeToLawyer(userProcedureID int, lawyer model.Admin, stepName string) error {
	up, user, err := loadUserProcedureAndUser(userProcedureID)
	if err != nil {
		return err
	}

	data := baseInternalMailData(up, user)
	data.RecipientName = lawyer.Name
	data.Badge = "Client Response"
	data.AssignedLawyerName = lawyer.Name
	data.ProcedureStatusLabel = readableProcedureStatus(up.Status)
	data.Intro = fmt.Sprintf("Client %s has updated the requested information for step %s in procedure %s.", user.Name, stepName, procedureLabel(up))
	subject := fmt.Sprintf("Client responded to info required: %s", procedureLabel(up))

	return sendMailToLawyer(lawyer, subject, templates.MailTemplateLawyerProcedureNotice, data)
}

func SendDeadlineAlertToLawyerAndAdmins(up model.UserProcedure, isOverdue bool) error {
	if up.AdminID == nil {
		return nil
	}

	user := model.User{}
	if up.UserID != nil {
		loadedUser, err := repo.GetUserByID(*up.UserID)
		if err != nil {
			return err
		}
		user = loadedUser
	}

	lawyer, err := repo.GetAdminByID(context.Background(), *up.AdminID)
	if err != nil {
		return err
	}
	admins, err := repo.GetActiveNonLawyerAdmins()
	if err != nil {
		return err
	}

	data := baseInternalMailData(up, user)
	data.RecipientName = lawyer.Name
	data.AssignedLawyerName = lawyer.Name

	var subject string
	if isOverdue {
		data.Badge = "Overdue"
		data.Intro = fmt.Sprintf("Procedure %s has passed its deadline and requires immediate attention.", procedureLabel(up))
		subject = fmt.Sprintf("Procedure overdue: %s", procedureLabel(up))
	} else {
		data.Badge = "Deadline in 48h"
		data.Intro = fmt.Sprintf("Procedure %s will reach its deadline within the next 48 hours.", procedureLabel(up))
		subject = fmt.Sprintf("Upcoming deadline in 48h: %s", procedureLabel(up))
	}

	return sendMailToLawyerWithCC(lawyer, admins, subject, templates.MailTemplateInternalDeadlineAlert, data)
}

func loadProcedureMailContext(userProcedureID int) (model.UserProcedure, model.User, []model.Admin, error) {
	up, user, err := loadUserProcedureAndUser(userProcedureID)
	if err != nil {
		return model.UserProcedure{}, model.User{}, nil, err
	}
	admins, err := repo.GetActiveNonLawyerAdmins()
	if err != nil {
		return model.UserProcedure{}, model.User{}, nil, err
	}
	return up, user, admins, nil
}

func loadUserProcedureAndUser(userProcedureID int) (model.UserProcedure, model.User, error) {
	up, err := repo.GetUserProcedureByID(userProcedureID)
	if err != nil {
		return model.UserProcedure{}, model.User{}, err
	}
	user := model.User{}
	if up.UserID != nil {
		user, err = repo.GetUserByID(*up.UserID)
		if err != nil {
			return model.UserProcedure{}, model.User{}, err
		}
	}
	return up, user, nil
}

func baseInternalMailData(up model.UserProcedure, user model.User) templates.InternalProcedureMailData {
	data := templates.InternalProcedureMailData{
		ProcedureName:        up.TemplateName,
		CompanyName:          strings.TrimSpace(valueOrEmpty(up.CompanyName)),
		ClientName:           strings.TrimSpace(user.Name),
		UserProcedureNumber:  procedureLabel(up),
		WebsiteURL:           strings.TrimSpace(utils.GetWebBaseURL()),
		ProcedureStatusLabel: readableProcedureStatus(up.Status),
	}
	if up.DeadlineAt != nil {
		data.DueDate = up.DeadlineAt.Format("2006-01-02 15:04")
	}
	return data
}

func sendMailToAdmins(admins []model.Admin, subject, tmpl string, data templates.InternalProcedureMailData) error {
	var firstErr error
	for _, admin := range admins {
		email := strings.TrimSpace(valueOrEmpty(admin.Email))
		if email == "" {
			continue
		}
		payload := data
		payload.RecipientName = admin.Name
		body, err := templates.RenderMailTemplate(tmpl, payload)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			slog.Error("failed to render admin mail template", "template", tmpl, "error", err)
			continue
		}
		if err := cmail.SendSystemMail(context.Background(), cmail.MailContent{Subject: subject, Content: body, Address: email}); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			slog.Error("failed to send admin mail", "admin_id", admin.ID, "error", err)
		}
	}
	return firstErr
}

func sendMailToLawyer(lawyer model.Admin, subject, tmpl string, data templates.InternalProcedureMailData) error {
	email := strings.TrimSpace(valueOrEmpty(lawyer.Email))
	if email == "" {
		return nil
	}
	body, err := templates.RenderMailTemplate(tmpl, data)
	if err != nil {
		return err
	}
	return cmail.SendSystemMail(context.Background(), cmail.MailContent{Subject: subject, Content: body, Address: email})
}

func sendMailToLawyerWithCC(lawyer model.Admin, ccAdmins []model.Admin, subject, tmpl string, data templates.InternalProcedureMailData) error {
	email := strings.TrimSpace(valueOrEmpty(lawyer.Email))
	if email == "" {
		return nil
	}
	body, err := templates.RenderMailTemplate(tmpl, data)
	if err != nil {
		return err
	}
	cc := uniqueAdminEmails(ccAdmins, map[string]struct{}{strings.ToLower(email): {}})
	return cmail.SendSystemMailWithCC(context.Background(), cmail.MailContent{Subject: subject, Content: body, Address: email}, cc)
}

func uniqueAdminEmails(admins []model.Admin, skip map[string]struct{}) []string {
	cc := make([]string, 0, len(admins))
	seen := map[string]struct{}{}
	for k, v := range skip {
		seen[k] = v
	}
	for _, admin := range admins {
		email := strings.ToLower(strings.TrimSpace(valueOrEmpty(admin.Email)))
		if email == "" {
			continue
		}
		if _, ok := seen[email]; ok {
			continue
		}
		seen[email] = struct{}{}
		cc = append(cc, email)
	}
	return cc
}

func procedureLabel(up model.UserProcedure) string {
	if strings.TrimSpace(up.ProcNumber) != "" {
		return up.ProcNumber
	}
	return fmt.Sprintf("#%d", up.ID)
}

func readableProcedureStatus(status *string) string {
	if status == nil || strings.TrimSpace(*status) == "" {
		return ""
	}
	replacer := strings.NewReplacer("-", " ", "_", " ")
	return strings.Title(replacer.Replace(strings.TrimSpace(*status)))
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
