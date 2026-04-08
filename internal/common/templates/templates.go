package templates

import (
	"bytes"
	"embed"
	"html/template"
)

//go:embed *.html
var mailFS embed.FS

func TemplateFuncs() template.FuncMap {
	return template.FuncMap{}
}

// RenderMailTemplate renders an embedded mail template by filename with provided data.
// tmplName must be one of the embedded files (e.g., "mail_procedure_error.html").
func RenderMailTemplate(tmplName string, data any) (string, error) {
	t, err := template.New(tmplName).Funcs(TemplateFuncs()).ParseFS(mailFS, tmplName)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

const (
	MailTemplateProcedureError            = "mail_procedure_error.html"
	MailTemplateDeadlineReminder          = "mail_deadline_reminder.html"
	MailTemplateProcedureCancelled        = "mail_procedure_cancelled.html"
	MailTemplateInvoiceCompleted          = "mail_invoice_completed.html"
	MailTemplateProcedureCompleted        = "mail_procedure_completed.html"
	MailTemplateUserRegistrationCompleted = "mail_user_registration_completed.html"
	MailTemplateAdminProcedureNotice      = "mail_admin_procedure_notice.html"
	MailTemplateLawyerProcedureNotice     = "mail_lawyer_procedure_notice.html"
	MailTemplateInternalDeadlineAlert     = "mail_internal_deadline_alert.html"
)

type DeadlineReminderData struct {
	UserName            string
	ProcedureName       string
	CompanyName         string
	DueDate             string
	DaysLeft            int
	UserProcedureNumber string
	URL                 string
	Cost                string
	Status              string
	DueDays             int
}

type InternalProcedureMailData struct {
	RecipientName        string
	Badge                string
	Intro                string
	ProcedureName        string
	CompanyName          string
	ClientName           string
	DueDate              string
	UserProcedureNumber  string
	AssignedLawyerName   string
	WebsiteURL           string
	ProcedureStatusLabel string
}
