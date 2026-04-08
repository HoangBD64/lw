package mail

import (
	"context"
	"core/internal/domain_model/model"
	repo "core/internal/repo"
	"core/pkg/send_mail"
	"encoding/json"
	"log/slog"
	"net/mail"
	"strconv"
)

type MailContent struct {
	Subject string `json:"subject"`
	Content string `json:"content"`
	Address string `json:"address"`
}

func SendSystemMail(ctx context.Context, mailContent MailContent) error {
	setting, err := repo.GetSystemSetting(ctx, model.MAIL_SYSTEM)
	if err != nil {
		return err
	}

	mailSetting := model.SystemSettingMail{}
	if err := json.Unmarshal(setting.Metadata, &mailSetting); err != nil {
		slog.Debug("failed to unmarshal mail system setting", "error", err)
	}

	if _, err := mail.ParseAddress(mailSetting.MAIL_SENDER); err != nil {
		slog.Info("mail system setting is invalid", "mail", mailSetting.MAIL_SENDER, "error", err)
		return err
	}

	port, err := strconv.Atoi(mailSetting.MAIL_PORT)
	if err != nil {
		return err
	}

	mailInput := send_mail.InputMail{
		Subject: mailContent.Subject,
		Content: mailContent.Content,
		Mail:    mailContent.Address,
	}

	config := send_mail.Config{
		PORT:        port,
		SENDER_MAIL: mailSetting.MAIL_SENDER,
		ACCOUNT:     mailSetting.MAIL_ACCOUNT,
		PASS:        mailSetting.MAIL_PASS,
		HOST:        mailSetting.MAIL_HOST,
		HeaderName:  mailSetting.MAIL_HEADER,
	}
	if err := send_mail.SendMail(mailInput, config); err != nil {
		return err
	}
	return nil
}

func SendSystemMailWithCC(ctx context.Context, mailContent MailContent, cc []string) error {
	setting, err := repo.GetSystemSetting(ctx, model.MAIL_SYSTEM)
	if err != nil {
		return err
	}

	mailSetting := model.SystemSettingMail{}
	if err := json.Unmarshal(setting.Metadata, &mailSetting); err != nil {
		slog.Debug("failed to unmarshal mail system setting", "error", err)
	}

	if _, err := mail.ParseAddress(mailSetting.MAIL_SENDER); err != nil {
		slog.Info("mail system setting is invalid", "mail", mailSetting.MAIL_SENDER, "error", err)
		return err
	}

	port, err := strconv.Atoi(mailSetting.MAIL_PORT)
	if err != nil {
		return err
	}

	mailInput := send_mail.InputMail{
		Subject: mailContent.Subject,
		Content: mailContent.Content,
		Mail:    mailContent.Address,
		CC:      cc,
	}

	config := send_mail.Config{
		PORT:        port,
		SENDER_MAIL: mailSetting.MAIL_SENDER,
		ACCOUNT:     mailSetting.MAIL_ACCOUNT,
		PASS:        mailSetting.MAIL_PASS,
		HOST:        mailSetting.MAIL_HOST,
		HeaderName:  mailSetting.MAIL_HEADER,
	}
	if err := send_mail.SendMail(mailInput, config); err != nil {
		return err
	}
	return nil
}
