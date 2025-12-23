package notification

import (
	"core/app"
	"core/internal/domain_model/model"
	"fmt"
)

type EmailSender interface {
	Send(to string, subject string, body string) error
}

type Service struct {
	email EmailSender
}

func NewService(email EmailSender) *Service { return &Service{email: email} }

func (s *Service) UpsertPreference(pref *model.NotificationPreference) error {
	return app.Database.DB.Model(&model.NotificationPreference{}).
		Where("user_id = ?", pref.UserID).
		Assign(pref).
		FirstOrCreate(pref).Error
}

func (s *Service) CreateNotifyInApp(n *model.Notification) error {
	return app.Database.DB.Model(&model.Notification{}).Create(n).Error
}

func (s *Service) SendEmailIfEnabled(user model.User, subject, body string) error {
	var pref model.NotificationPreference
	if err := app.Database.DB.Model(&model.NotificationPreference{}).Where("user_id = ?", user.ID).First(&pref).Error; err != nil {
		return nil
	}
	if !pref.EmailEnabled || s.email == nil {
		return nil
	}
	if user.Email == nil || *user.Email == "" {
		return fmt.Errorf("user email empty")
	}
	return s.email.Send(*user.Email, subject, body)
}
