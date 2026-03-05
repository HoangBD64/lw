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
		items             []model.Document
		companyCodes      []string
		totalCompanyCodes int64
	)

	// Base query để lấy distinct company code
	companyQuery := app.Database.DB.Table("user_procedures").
		Select("DISTINCT company_code").
		Where("company_code IS NOT NULL AND company_code != ''")

	if q.AdminID != nil {
		companyQuery = companyQuery.Where("admin_id = ?", q.AdminID)
	}

	// Apply search filter cho company code query
	if q.Search != "" {
		like := "%" + q.Search + "%"
		companyQuery = companyQuery.Where(`
			user_procedures.proc_number ILIKE ? 
			OR user_procedures.company_code ILIKE ?
			OR EXISTS (
				SELECT 1 FROM documents d 
				WHERE d.user_procedure_id = user_procedures.id 
				AND (d.name ILIKE ? OR d.file_type ILIKE ?)
			)`, like, like, like, like)
	}

	// Count total company codes
	if err := companyQuery.Count(&totalCompanyCodes).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated company codes
	offset := (q.Page - 1) * q.Limit
	if err := companyQuery.
		Order("company_code ASC").
		Limit(q.Limit).
		Offset(offset).
		Pluck("company_code", &companyCodes).Error; err != nil {
		return nil, 0, err
	}

	if len(companyCodes) == 0 {
		return []model.Document{}, totalCompanyCodes, nil
	}

	// Get all documents for the selected company codes
	tx := app.Database.DB.Model(&model.Document{}).
		Preload("UserProcedure", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "proc_number", "company_code", "payment_account")
		}).
		Joins("JOIN user_procedures up ON documents.user_procedure_id = up.id").
		Where("up.company_code IN ?", companyCodes)

	if q.AdminID != nil {
		tx = tx.Where("up.admin_id = ?", q.AdminID)
	}
	if q.UserProcedureID != nil {
		tx = tx.Where("documents.user_procedure_id = ?", q.UserProcedureID)
	}
	if q.Type != "" {
		tx = tx.Where("documents.type = ?", q.Type)
	}
	if q.UserID != nil {
		tx = tx.Where("documents.user_id = ?", q.UserID)
	}
	if q.StartTime != nil && q.EndTime != nil {
		tx = tx.Where("documents.created_at >= ? AND documents.created_at <= ?", q.StartTime, q.EndTime)
	}

	// Order by company_code then by created_at
	tx = tx.Order("up.company_code ASC, documents.created_at DESC").Find(&items)
	return items, totalCompanyCodes, tx.Error
}

func UserGetDocuments(q Query) ([]model.Document, int64, error) {
	var (
		userDocuments     []model.Document
		companyCodes      []string
		totalCompanyCodes int64
	)

	companyQuery := app.Database.DB.Table("user_procedures").
		Select("DISTINCT company_code").
		Where("user_id = ? AND company_code IS NOT NULL AND company_code != ''", q.UserID)

	if q.Search != "" {
		like := "%" + q.Search + "%"
		companyQuery = companyQuery.Where(`
			user_procedures.proc_number ILIKE ? 
			OR user_procedures.company_code ILIKE ?
			OR EXISTS (
				SELECT 1 FROM documents d 
				WHERE d.user_procedure_id = user_procedures.id 
				AND (d.name ILIKE ? OR d.file_type ILIKE ?)
			)`, like, like, like, like)
	}

	// Count total company codes
	if err := companyQuery.Count(&totalCompanyCodes).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated company codes
	offset := (q.Page - 1) * q.Limit
	if err := companyQuery.
		Order("company_code ASC").
		Limit(q.Limit).
		Offset(offset).
		Pluck("company_code", &companyCodes).Error; err != nil {
		return nil, 0, err
	}

	if len(companyCodes) == 0 {
		return []model.Document{}, totalCompanyCodes, nil
	}

	// Get all documents for the selected company codes
	tx := app.Database.DB.Model(&model.Document{}).
		Preload("UserProcedure", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "proc_number", "company_code", "payment_account")
		}).
		Joins("JOIN user_procedures up ON documents.user_procedure_id = up.id").
		Where("up.user_id = ? AND up.company_code IN ?", q.UserID, companyCodes)

	if q.UserProcedureID != nil {
		tx = tx.Where("documents.user_procedure_id = ?", q.UserProcedureID)
	}
	if q.Type != "" {
		tx = tx.Where("documents.type = ?", q.Type)
	}
	if q.StartTime != nil && q.EndTime != nil {
		tx = tx.Where("documents.created_at >= ? AND documents.created_at <= ?", q.StartTime, q.EndTime)
	}

	tx = tx.Order("up.company_code ASC, documents.created_at DESC").Find(&userDocuments)
	return userDocuments, totalCompanyCodes, tx.Error
}

func UserCreateDocument(doc model.Document) (model.Document, error) {
	err := app.Database.DB.Create(&doc).Error
	return doc, err
}

func UserGetDocumentByID(id int, userID uuid.UUID) (model.Document, error) {
	var doc model.Document
	sub := app.Database.DB.
		Model(&model.UserProcedure{}).
		Select("id").
		Where("user_id = ?", userID)

	err := app.Database.DB.
		First(&doc, "id = ? AND user_procedure_id IN (?)", id, sub).Error
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

func UserDeleteDocumentByIDs(ids []int, userID uuid.UUID) error {
	return app.Database.DB.Delete(&model.Document{}, "id IN ? AND user_id = ?", ids, userID).Error
}
