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

func AdminDeleteDocumentByID(id int) error {
	return app.Database.DB.Delete(&model.Document{}, "id = ?", id).Error
}

func AdminUpdateDocument(doc *model.Document) error {
	return app.Database.DB.Model(&model.Document{}).Where("id = ?", doc.ID).Updates(doc).Error
}

func AdminGetDocuments(q Query) ([]model.Document, int64) {
	var (
		items []model.Document
		count int64
	)
	db := app.Database.DB.Model(&model.Document{})
	if q.Search != "" {
		db = db.Where("name ILIKE ? OR file_type ILIKE ?", "%"+q.Search+"%", "%"+q.Search+"%")
	}
	if q.Status != "" {
		db = db.Where("status = ?", q.Status)
	}
	if q.AdminID != nil {
		db = db.Where("admin_id = ?", q.AdminID)
	}
	tx := db.Count(&count)
	if tx.Error != nil {
		return nil, 0
	}
	tx = db.Scopes(pagination(q.Page, q.Limit)).Order("created_at DESC").Find(&items)
	return items, count
}
