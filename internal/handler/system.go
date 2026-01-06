package handler

import (
	"context"
	"core/internal/common/templates"
	"core/internal/domain_model/domain"
	"core/internal/domain_model/model"
	"core/internal/repo"
	"core/internal/scheduler"
	"core/pkg/send_mail"
	"core/utils"
	"core/utils/base"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type PaymentInfoSetting struct {
	Type           string         `json:"type"`
	APIBase        string         `json:"api_base,omitempty"`
	ClientID       string         `json:"client_id,omitempty"`
	APIKey         string         `json:"api_key,omitempty"`
	ClientSecret   string         `json:"client_secret,omitempty"`
	IsActive       bool           `json:"is_active"`
	Metadata       datatypes.JSON `json:"metadata,omitempty"`
	TransactionFee float64        `json:"transaction_fee"`
}

func GetSystemSetting(c *fiber.Ctx) error {
	var query repo.Query
	query.Parse(c)

	admin, err := checkCurrentSystemAdmin(c)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "unauthorized", Error: err})
	}

	if err := isSystemSettingRoleAllowed(query.Type, admin); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "an error occurred", Error: err})
	}

	setting, err := repo.GetSystemSetting(context.TODO(), query.Type)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(base.BaseResponse{Status: false, Message: "failed to get data", Error: err})
		}

		newSetting := model.SystemSetting{Type: query.Type}

		createdSetting, err := repo.CreateSystemSetting(newSetting)
		if err != nil {
			return c.JSON(base.BaseResponse{Status: false, Message: "failed to create default setting", Error: err})
		}

		setting = createdSetting
	}

	if setting.Type == model.PAYMENT_INFO {
		data := []model.PaymentInfoPublic{}
		if err := json.Unmarshal(setting.Metadata, &data); err != nil {
			slog.Error("failed to unmarshal payment info data", "error", err)
		}

		jsonBytes, err := json.Marshal(data)
		if err != nil {
			slog.Error("failed to marshal payment info public data", "error", err)
		} else {
			setting.Metadata = jsonBytes
		}
	}

	if setting.Type == model.STORAGE_CREDENTIAL {
		credential := model.StorageCredential{}
		if err := json.Unmarshal(setting.Metadata, &credential); err != nil {
			slog.ErrorContext(context.TODO(), "failed to unmarshal storage credential data", "error", err)
		}

		// mask secret key
		credential.SecretKey = ""

		credentialJson, err := json.Marshal(credential)
		if err != nil {
			return c.JSON(base.BaseResponse{Status: false, Message: "could not process fetched data", Error: err})
		}

		setting.Metadata = credentialJson
	}

	if setting.Type == model.MAIL_SYSTEM {
		mailSetting := model.SystemSettingMail{}
		if err := json.Unmarshal(setting.Metadata, &mailSetting); err != nil {
			slog.ErrorContext(context.TODO(), "failed to unmarshal storage credential data", "error", err)
		}

		mailSetting.MAIL_PASS = ""
		mailSettingJson, err := json.Marshal(mailSetting)
		if err != nil {
			return c.JSON(base.BaseResponse{Status: false, Message: "could not process fetched data", Error: err})
		}

		setting.Metadata = mailSettingJson
	}

	if setting.Type == model.JOB_SCHEDULER {
		schedulerConfig := model.SchedulerConfig{}
		if err := json.Unmarshal(setting.Metadata, &schedulerConfig); err != nil {
			slog.ErrorContext(context.TODO(), "failed to unmarshal scheduler config data", "error", err)
		}
		// apply defaults when missing
		if schedulerConfig.DailyTime == "" {
			schedulerConfig.DailyTime = "09:00"
		}
		// default enabled true if metadata missing
		if len(setting.Metadata) == 0 {
			schedulerConfig.Enabled = true
		}
		schedulerConfigJson, err := json.Marshal(schedulerConfig)
		if err != nil {
			return c.JSON(base.BaseResponse{Status: false, Message: "could not process fetched data", Error: err})
		}
		setting.Metadata = schedulerConfigJson

		needUpdate := false

		if strings.TrimSpace(schedulerConfig.DailyTime) == "" {
			schedulerConfig.DailyTime = "09:00"
			needUpdate = true
		}

		if needUpdate {
			newMetadata, err := json.Marshal(schedulerConfig)
			if err != nil {
				return c.JSON(base.BaseResponse{Status: false, Message: "could not process fetched data", Error: err})
			}

			setting.Metadata = newMetadata

			updatedSetting, err := repo.UpdateSystemSetting(setting)
			if err != nil {
				return c.JSON(base.BaseResponse{Status: false, Message: "failed to update data", Error: err})
			}

			setting = updatedSetting
		}
	}

	return c.JSON(base.BaseResponse{Status: true, Message: "ok", Data: setting})
}

func AdminUpdateSetting(c *fiber.Ctx) error {
	admin, err := checkCurrentSystemAdmin(c)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "unauthorized", Error: err})
	}

	var input domain.AdminUpdateSettingRequest
	if err := c.BodyParser(&input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "bad request", Error: err})
	}

	if err := appUtils.Validate(context.TODO(), &input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "bad request", Error: err})
	}

	setting_found, err := repo.GetSystemSettingByID(input.ID)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Get GetSystemSettingByID Failed", Error: err})
	}

	if err := isSystemSettingRoleAllowed(setting_found.Type, admin); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "an error occurred", Error: err})
	}

	if setting_found.Type == model.PAYMENT_INFO {
		old_data := []PaymentInfoSetting{}
		if err := json.Unmarshal(setting_found.Metadata, &old_data); err != nil {
			slog.Debug("Error marshaling:", "error", err)
		}
		new_data := []PaymentInfoSetting{}
		if err := json.Unmarshal(input.Metadata, &new_data); err != nil {
			slog.Debug("Error marshaling:", "error", err)
		}
		old_data_update := []PaymentInfoSetting{}
		for _, new_setting := range new_data {
			for _, payment_info := range old_data {
				if new_setting.Type == payment_info.Type && (new_setting.Type == model.PAYMENT_PAYPAL || new_setting.Type == model.PAYMENT_AIRWALLEX) {
					if new_setting.APIKey == "" {
						new_setting.APIKey = payment_info.APIKey
					}
					if new_setting.ClientSecret == "" {
						new_setting.ClientSecret = payment_info.ClientSecret
					}
				}
			}

			old_data_update = append(old_data_update, new_setting)
		}
		setting_new, err := json.Marshal(old_data_update)
		if err != nil {
			return c.JSON(fiber.Map{"status": false, "message": "json.Marshal(old_data) Error", "error": err.Error()})
		}

		input.Metadata = setting_new
	}

	if setting_found.Type == model.STORAGE_CREDENTIAL {
		newData := model.StorageCredential{}
		if err := json.Unmarshal(input.Metadata, &newData); err != nil {
			return c.JSON(base.BaseResponse{Status: false, Message: "could not process requested storage config", Error: err})
		}

		oldData := model.StorageCredential{}
		if err := json.Unmarshal(setting_found.Metadata, &oldData); err != nil {
			slog.ErrorContext(context.TODO(), "failed to unmarshal existing storage credential data", "error", err)
		}

		// keep old access key and secret key if new ones are empty
		if newData.SecretKey == "" {
			newData.SecretKey = oldData.SecretKey
		}

		newSetting, err := json.Marshal(newData)
		if err != nil {
			return c.JSON(base.BaseResponse{Status: false, Message: "could not process requested data", Error: err})
		}

		input.Metadata = newSetting
	}

	if setting_found.Type == model.MAIL_SYSTEM {
		newData := model.SystemSettingMail{}
		if err := json.Unmarshal(input.Metadata, &newData); err != nil {
			return c.JSON(base.BaseResponse{Status: false, Message: "could not process requested storage config", Error: err})
		}

		oldData := model.SystemSettingMail{}
		if err := json.Unmarshal(setting_found.Metadata, &oldData); err != nil {
			slog.ErrorContext(context.TODO(), "failed to unmarshal existing storage credential data", "error", err)
		}

		if newData.MAIL_PASS == "" {
			newData.MAIL_PASS = oldData.MAIL_PASS
		}

		newSetting, err := json.Marshal(newData)
		if err != nil {
			return c.JSON(base.BaseResponse{Status: false, Message: "could not process requested data", Error: err})
		}

		input.Metadata = newSetting
	}

	if setting_found.Type == model.JOB_SCHEDULER {
		var oldConf model.SchedulerConfig
		var newConf model.SchedulerConfig

		if len(setting_found.Metadata) > 0 {
			if err := json.Unmarshal(setting_found.Metadata, &oldConf); err != nil {
				slog.Error("scheduler: failed to parse existing scheduler config", "error", err)
			}
		}

		if len(input.Metadata) > 0 {
			if err := json.Unmarshal(input.Metadata, &newConf); err != nil {
				slog.Error("scheduler: failed to parse updated scheduler config", "error", err)
				return c.JSON(base.BaseResponse{Status: false, Message: "invalid scheduler config", Error: err})
			}
		}

		if strings.TrimSpace(newConf.DailyTime) == "" {
			if strings.TrimSpace(oldConf.DailyTime) != "" {
				newConf.DailyTime = oldConf.DailyTime
			} else {
				newConf.DailyTime = "09:00"
			}
		}

		switch {
		case oldConf.Enabled && newConf.Enabled:
			if newConf.DailyTime != oldConf.DailyTime {
				if err := scheduler.RescheduleDaily(c.Context(), newConf.DailyTime); err != nil {
					slog.Error("scheduler: reschedule failed", "error", err)
				}
			}

		case !oldConf.Enabled && newConf.Enabled:
			if err := scheduler.Enable(c.Context(), newConf.DailyTime); err != nil {
				slog.Error("scheduler: enable failed", "error", err)
			}

		case oldConf.Enabled && !newConf.Enabled:
			scheduler.Disable()

		default:
		}
	}

	setting_found.Resource = input.Resource
	setting_found.Metadata = input.Metadata
	setting_found.Contant = input.Contant
	if _, err := repo.UpdateSystemSetting(setting_found); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "failed to update data", Error: err})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "ok"})
}

func AdminTestSendMail(c *fiber.Ctx) error {
	var input domain.TestSendMailRequest
	if err := c.BodyParser(&input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "bad request", Error: err})
	}

	if err := appUtils.Validate(context.TODO(), &input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "bad request", Error: err})
	}

	setting, err := repo.GetSystemSetting(context.TODO(), model.MAIL_SYSTEM)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(base.BaseResponse{Status: false, Message: "failed to get data", Error: err})
		}
	}

	mail := model.SystemSettingMail{}
	if err := json.Unmarshal(setting.Metadata, &mail); err != nil {
		slog.ErrorContext(context.TODO(), "failed to unmarshal existing storage credential data", "error", err)
	}

	subject := "LegalSwift Mail Test"
	content := "<p>This is a test email to verify mail settings.</p>"

	switch input.Option {
	case "deadline_reminder":
		subject = "Reminder: Upcoming Deadline<<LegalSwift Mail Test>>"
		fake := templates.DeadlineReminderData{
			UserName:            "Jane Doe",
			ProcedureName:       "Company Registration",
			CompanyName:         "ACME LLC",
			DueDate:             "2026-01-15",
			DaysLeft:            2,
			UserProcedureNumber: "#LW0001",
			URL:                 utils.GetWebBaseURL(),
		}
		if html, err := templates.RenderMailTemplate(templates.MailTemplateDeadlineReminder, fake); err == nil {
			content = html
		} else {
			return c.JSON(base.BaseResponse{Status: false, Message: "failed to render template", Error: err})
		}
	case "invoice_completed":
		subject = "Payment Confirmed: Company Registration<<LegalSwift Mail Test>>"
		fake := templates.DeadlineReminderData{
			UserName:            "Jane Doe",
			ProcedureName:       "Company Registration",
			CompanyName:         "ACME LLC",
			Cost:                "$299",
			UserProcedureNumber: "#LW0001",
			URL:                 utils.GetWebBaseURL(),
		}
		if html, err := templates.RenderMailTemplate(templates.MailTemplateInvoiceCompleted, fake); err == nil {
			content = html
		} else {
			return c.JSON(base.BaseResponse{Status: false, Message: "failed to render template", Error: err})
		}
	case "procedure_cancelled":
		subject = "Procedure Cancelled: Company Registration<<LegalSwift Mail Test>>"
		fake := templates.DeadlineReminderData{
			UserName:            "Jane Doe",
			ProcedureName:       "Company Registration",
			CompanyName:         "ACME LLC",
			Status:              "Cancelled",
			UserProcedureNumber: "#LW0001",
			URL:                 utils.GetWebBaseURL(),
		}
		if html, err := templates.RenderMailTemplate(templates.MailTemplateProcedureCancelled, fake); err == nil {
			content = html
		} else {
			return c.JSON(base.BaseResponse{Status: false, Message: "failed to render template", Error: err})
		}
	case "procedure_completed":
		subject = "Procedure Completed: Company Registration<<LegalSwift Mail Test>>"
		fake := templates.DeadlineReminderData{
			UserName:            "Jane Doe",
			ProcedureName:       "Company Registration",
			CompanyName:         "ACME LLC",
			Status:              "Completed",
			UserProcedureNumber: "#LW0001",
			URL:                 utils.GetWebBaseURL(),
		}
		if html, err := templates.RenderMailTemplate(templates.MailTemplateProcedureCompleted, fake); err == nil {
			content = html
		} else {
			return c.JSON(base.BaseResponse{Status: false, Message: "failed to render template", Error: err})
		}
	case "procedure_error":
		subject = "Procedure Issue Detected: Company Registration<<LegalSwift Mail Test>>"
		fake := templates.DeadlineReminderData{
			UserName:            "Jane Doe",
			ProcedureName:       "Company Registration",
			CompanyName:         "ACME LLC",
			Status:              "Needs Attention",
			UserProcedureNumber: "#LW0001",
			URL:                 utils.GetWebBaseURL(),
		}
		if html, err := templates.RenderMailTemplate(templates.MailTemplateProcedureError, fake); err == nil {
			content = html
		} else {
			return c.JSON(base.BaseResponse{Status: false, Message: "failed to render template", Error: err})
		}
	default:
		// test mail procedure template
		subject = "Procedure Completed: Company Registration<<LegalSwift Mail Test>>"
		fake := templates.DeadlineReminderData{
			UserName:            "Jane Doe",
			ProcedureName:       "Company Registration",
			CompanyName:         "ACME LLC",
			Status:              "Completed",
			UserProcedureNumber: "#LW0001",
			URL:                 utils.GetWebBaseURL(),
		}
		if html, err := templates.RenderMailTemplate(templates.MailTemplateProcedureCompleted, fake); err == nil {
			content = html
		} else {
			return c.JSON(base.BaseResponse{Status: false, Message: "failed to render template", Error: err})
		}
	}

	mailInput := send_mail.InputMail{
		Subject: subject,
		Content: content,
		Mail:    input.Address,
	}

	port, _ := strconv.Atoi(mail.MAIL_PORT)
	mailConfig := send_mail.Config{
		PORT:        port,
		SENDER_MAIL: mail.MAIL_SENDER,
		ACCOUNT:     mail.MAIL_ACCOUNT,
		PASS:        mail.MAIL_PASS,
		HOST:        mail.MAIL_HOST,
		HeaderName:  mail.MAIL_HEADER,
	}

	if err := send_mail.SendMail(mailInput, mailConfig); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "failed to send test mail", Error: err})
	}

	return c.JSON(base.BaseResponse{Status: true, Message: "test mail sent successfully"})
}

func checkCurrentSystemAdmin(c *fiber.Ctx) (model.SystemAdmin, error) {
	if admin, ok := c.Locals("admin").(model.Admin); ok {
		if admin.IsLawyer() {
			return nil, errors.New("unauthorized access")
		}
		return admin, nil
	}
	return nil, errors.New("unauthorized access")
}

func isSystemSettingRoleAllowed(typeName string, sysAdmin model.SystemAdmin) error {
	dualAccess := "both"

	src := map[string]string{
		// NOTE: TEMPORARY
		// model.STORAGE_CREDENTIAL: model.SUPER_ADMIN,
		model.STORAGE_CREDENTIAL: model.RoleAdmin,
		model.PAYMENT_INFO:       model.RoleAdmin,
		model.MAIL_SYSTEM:        model.RoleAdmin,
		model.JOB_SCHEDULER:      model.RoleAdmin,
	}

	perm, exists := src[typeName]
	if !exists {
		return fmt.Errorf("setting type not recognized")
	}

	// allow both admin and super admin
	if perm == dualAccess {
		return nil
	}

	if perm != sysAdmin.AdminType() {
		return fmt.Errorf("insufficient permissions to access this setting")
	}

	return nil
}
