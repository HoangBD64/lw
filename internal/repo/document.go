package repo

import (
	"core/app"
	"core/internal/domain_model/model"
	"errors"

	"gorm.io/gorm"
)

func AdminCreateDocument(doc *model.Document) error {
	return app.Database.DB.Create(doc).Error
}

func AdminGetDocumentByID(id int) (model.Document, error) {
	var doc model.Document
	err := app.Database.DB.First(&doc, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return doc, err
	}
	return doc, err
}

func AdminDeleteDocumentByIDs(ids []int) error {
	return app.Database.DB.Delete(&model.Document{}, "id IN ?", ids).Error
}

func AdminUpdateDocument(doc *model.Document) error {
	return app.Database.DB.Model(&model.Document{}).Where("id = ?", doc.ID).Updates(doc).Error
}

func AdminGetDocuments(q Query) ([]model.Document, int64) {
	var (
		items []model.Document
		count int64
	)
	tx := app.Database.DB.Model(&model.Document{})
	if q.Search != "" {
		tx = tx.Where("name ILIKE ? OR file_type ILIKE ?", "%"+q.Search+"%", "%"+q.Search+"%")
	}
	if q.Status != "" {
		tx = tx.Where("status = ?", q.Status)
	}
	if q.AdminID != nil {
		tx = tx.Where("admin_id = ?", q.AdminID)
	}
	tx = tx.Count(&count)
	if tx.Error != nil {
		return nil, 0
	}
	tx = tx.Scopes(pagination(q.Page, q.Limit)).Order("created_at DESC").Find(&items)
	return items, count
}

func UserGetUserDocuments(q Query) ([]model.UserDocument, int64, error) {
	var (
		userDocuments []model.UserDocument
		count         int64
	)
	tx := app.Database.DB.Model(&model.UserDocument{}).
		Preload("Document", func(tx *gorm.DB) *gorm.DB { return tx.Select("id", "name", "file_type", "file_url", "size", "description")}).
		Select("id", "document_id").
		Where("user_id = ?", q.UserID)
	tx = tx.Count(&count)
	if tx.Error != nil {
		return nil, 0, nil
	}
	tx = tx.Scopes(pagination(q.Page, q.Limit)).Order("created_at DESC").Find(&userDocuments)
	return userDocuments, count, tx.Error
}

func UserCreateDocument(doc *model.Document) error {
	return app.Database.DB.Create(doc).Error
}

func UserGetDocumentByID(id int) (model.Document, error) {
	var doc model.Document
	err := app.Database.DB.First(&doc, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return doc, err
	}
	return doc, err
}

func UserUpdateDocument(doc *model.Document) error {
	return app.Database.DB.Model(&model.Document{}).Where("id = ?", doc.ID).Updates(doc).Error
}

func UserDeleteDocumentByIDs(ids []int) error {
	return app.Database.DB.Delete(&model.Document{}, "id IN ?", ids).Error
}