package notification

import (
	"context"
	cmail "core/internal/common/mail"
)

type SystemEmailSender struct{}

func (SystemEmailSender) Send(to string, subject string, body string) error {
	return cmail.SendSystemMail(context.Background(), cmail.MailContent{Subject: subject, Content: body, Address: to})
}
