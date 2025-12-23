package repo

import (
	"context"
	"core/app"
	"core/internal/domain_model/model"
)

func CreateContactForm(ctx context.Context, contactForm model.ContactForm) (model.ContactForm, error) {
	if err := app.Database.DB.Create(&contactForm).Error; err != nil {
		return model.ContactForm{}, err
	}

	return contactForm, nil
}

func GetContactForms(ctx context.Context, q Query) ([]model.ContactForm, int64, error) {
	var (
		items []model.ContactForm
		count int64
	)
	tx := app.Database.DB.Model(&model.ContactForm{})
	if q.Search != "" {
		tx = tx.Where("name ILIKE ? OR email ILIKE ? OR phone ILIKE ? OR description::text ILIKE ?", "%"+q.Search+"%", "%"+q.Search+"%", "%"+q.Search+"%", "%"+q.Search+"%")
	}
	if q.UserID != nil {
		tx = tx.Where("user_id = ?", q.UserID)
	}
	if q.StartTime != nil && q.EndTime != nil {
		tx = tx.Where("created_at BETWEEN ? AND ?", q.StartTime, q.EndTime)
	}

	if err := tx.Order("created_at DESC").Find(&items).Error; err != nil {
		return items, 0, err
	}
	tx = tx.Count(&count)
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	tx = tx.Scopes(pagination(q.Page, q.Limit)).Order("created_at DESC").Find(&items)
	return items, count, nil
}
