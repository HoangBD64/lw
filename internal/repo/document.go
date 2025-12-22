package repo

import (
	"core/app"
	"core/internal/domain_model/model"
	"errors"

	"github.com/google/uuid"
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

func AdminDeleteDocumentByIDs(ids []int) (int64, error) {
	result := app.Database.DB.
		Delete(&model.Document{}, "id IN ? AND type = ?", ids, model.ADMIN_UPLOAD_DOCUMENT)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func AdminUpdateDocument(doc *model.Document, adminID uuid.UUID) error {
	return app.Database.DB.Model(&model.Document{}).Where("id = ? AND admin_id = ?", doc.ID, adminID).Updates(doc).Error
}

func AdminGetDocuments(q Query) ([]model.Document, int64, error) {
	var (
		items []model.Document
		count int64
	)
	tx := app.Database.DB.Model(&model.Document{})
	if q.Search != "" {
		tx = tx.Where("name ILIKE ? OR file_type ILIKE ?", "%"+q.Search+"%", "%"+q.Search+"%")
	}
	if q.AdminID != nil {
		tx = tx.Where("admin_id = ?", q.AdminID)
	}
	if q.Type != "" {
		tx = tx.Where("type = ?", q.Type)
	}
	if q.UserID != nil {
		tx = tx.Where("user_id = ?", q.UserID)
	}
	if q.UserProcedureID != nil {
		tx = tx.Where("user_procedure_id = ?", q.UserProcedureID)
	}
	if q.StartTime != nil && q.EndTime != nil {
		tx = tx.Where("created_at >= ? AND created_at <= ?", q.StartTime, q.EndTime)
	}
	tx = tx.Count(&count)
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	tx = tx.Scopes(pagination(q.Page, q.Limit)).Order("created_at DESC").Find(&items)
	return items, count, nil
}

// base functions for user
// func UserGetDocuments(q Query) ([]model.Document, int64, error) {
// 	var (
// 		userDocuments []model.Document
// 		count         int64
// 	)
// 	sub := app.Database.DB.
// 		Model(&model.UserProcedure{}).
// 		Select("id").
// 		Where("user_id = ?", q.UserID)

// 	tx := app.Database.DB.Model(&model.Document{}).
// 		Where("user_procedure_id IN (?)", sub)

// 	if q.Search != "" {
// 		tx = tx.Where("name ILIKE ? OR file_type ILIKE ?", "%"+q.Search+"%", "%"+q.Search+"%")
// 	}
// 	if q.Type != "" {
// 		tx = tx.Where("type = ?", q.Type)
// 	}
// 	if q.UserProcedureID != nil {
// 		tx = tx.Where("user_procedure_id = ?", *q.UserProcedureID)
// 	}
// 	if q.StartTime != nil && q.EndTime != nil {
// 		tx = tx.Where("created_at >= ? AND created_at <= ?", q.StartTime, q.EndTime)
// 	}
// 	tx = tx.Count(&count)
// 	if tx.Error != nil {
// 		return nil, 0, tx.Error
// 	}
// 	tx = tx.Scopes(pagination(q.Page, q.Limit)).Order("created_at DESC").Find(&userDocuments)
// 	return userDocuments, count, tx.Error
// }

func UserGetDocuments(q Query) ([]model.Document, int64, error) {
	var (
		userDocuments []model.Document
		count         int64
	)

	tx := app.Database.DB.Model(&model.Document{}).
		Where("user_id = ?", q.UserID)

	if q.Search != "" {
		tx = tx.Where("name ILIKE ? OR file_type ILIKE ?", "%"+q.Search+"%", "%"+q.Search+"%")
	}
	if q.Type != "" {
		tx = tx.Where("type = ?", q.Type)
	}
	if q.UserProcedureID != nil {
		tx = tx.Where("user_procedure_id = ?", *q.UserProcedureID)
	}
	if q.StartTime != nil && q.EndTime != nil {
		tx = tx.Where("created_at >= ? AND created_at <= ?", q.StartTime, q.EndTime)
	}
	tx = tx.Count(&count)
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	tx = tx.Scopes(pagination(q.Page, q.Limit)).Order("created_at DESC").Find(&userDocuments)
	return userDocuments, count, tx.Error
}

func UserCreateDocument(doc model.Document) (model.Document, error) {
	err := app.Database.DB.Create(&doc).Error
	return doc, err
}

func UserGetDocumentByID(id int) (model.Document, error) {
	var doc model.Document
	err := app.Database.DB.First(&doc, "id = ?", id).Error
	return doc, err
}

func UserGetDocumentByIDForUpdate(id int, userID uuid.UUID) (model.Document, error) {
	var doc model.Document
	err := app.Database.DB.First(&doc, "id = ? AND user_id = ?", id, userID).Error
	return doc, err
}

func UserUpdateDocument(doc *model.Document, userID uuid.UUID) error {
	return app.Database.DB.Debug().Model(&model.Document{}).Where("id = ? AND user_id = ?", doc.ID, userID).Updates(doc).Error
}

func UserDeleteDocumentByIDs(ids []int) error {
	return app.Database.DB.Delete(&model.Document{}, "id IN ?", ids).Error
}
