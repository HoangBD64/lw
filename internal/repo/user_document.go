package repo

import (
	"core/app"
	"core/internal/domain_model/model"

	"github.com/google/uuid"
)

func AdminAssignDocumentToUser(listAssign []model.UserDocument) error {
	tx := app.Database.DB.Create(&listAssign)
	return tx.Error
}

func AdminDeleteUserDocumentByDocumentIDs(ids []int) error {
	return app.Database.DB.Delete(&model.UserDocument{}, "document_id IN ?", ids).Error
}

// func AdminUnassignDocumentFromUser(listUnassign []model.UserDocument) (int64, error) {
// 	tx := app.Database.DB.Where("user_id IN ? AND document_id IN ?", userID, documentID).Delete(&model.UserDocument{})
// 	return tx.RowsAffected, tx.Error
// }

func ExistsUserDocument(userID uuid.UUID, documentID int) (bool, error) {
	var count int64
	tx := app.Database.DB.Model(&model.UserDocument{}).Where("user_id = ? AND document_id = ?", userID, documentID).Count(&count)
	return count > 0, tx.Error
}
