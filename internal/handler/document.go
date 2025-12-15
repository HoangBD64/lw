package handler

import (
	"core/internal/domain_model/domain"
	"core/internal/domain_model/model"
	repo "core/internal/repo"
	"core/utils/base"
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func AdminCreateDocument(c *fiber.Ctx) error {
	var input domain.DocumentCreateRequest
	if err := c.BodyParser(&input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Review your input", Error: err})
	}
	admin, ok := c.Locals("admin").(model.Admin)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again"})
	}
	doc := model.Document{
		Name:         input.Name,
		Description:  &input.Description,
		FileURL:      input.FileURL,
		FileType:     input.FileType,
		DocumentType: input.DocumentType,
		Size:         input.Size,
		AdminID:      &admin.ID,
	}
	if err := repo.AdminCreateDocument(&doc); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't Create Document", Error: err})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success"})
}

func AdminUpdateDocument(c *fiber.Ctx) error {
	var input domain.DocumentUpdateRequest
	if err := c.BodyParser(&input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Review your input", Error: err})
	}
	id := c.Params("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Invalid id", Error: err})
	}
	doc, err := repo.AdminGetDocumentByID(idInt)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(base.BaseResponse{Status: false, Message: "Document not found", Error: nil})
		}
		return c.JSON(base.BaseResponse{Status: false, Message: "Get Document Failed", Error: err})
	}
	if input.Name != "" {
		doc.Name = input.Name
	}
	if input.Description != "" {
		doc.Description = &input.Description
	}
	if input.FileURL != "" {
		doc.FileURL = input.FileURL
	}
	if input.FileType != "" {
		doc.FileType = input.FileType
	}
	if input.DocumentType != "" {
		doc.DocumentType = input.DocumentType
	}
	if input.Size != 0 {
		doc.Size = input.Size
	}
	if err := repo.AdminUpdateDocument(&doc); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't Update Document", Error: err})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success"})
}

func AdminGetDocument(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Invalid id", Error: err})
	}
	doc, err := repo.AdminGetDocumentByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(base.BaseResponse{Status: false, Message: "Document not found", Error: nil})
		}
		return c.JSON(base.BaseResponse{Status: false, Message: "Get Document Failed", Error: err})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: doc})
}

func AdminGetDocuments(c *fiber.Ctx) error {
	var q repo.Query
	q.Parse(c)
	items, total, err := repo.AdminGetDocuments(q)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Get Documents Failed", Error: err})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: fiber.Map{"items": items, "total": total}})
}

func AdminDeleteDocument(c *fiber.Ctx) error {
	var input domain.DocumentDeleteRequest
	if err := c.BodyParser(&input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Invalid id", Error: err})
	}
	if err := repo.AdminDeleteDocumentByIDs(input.IDs); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't Delete Document", Error: err})
	}
	if err := repo.AdminDeleteUserDocumentByDocumentIDs(input.IDs); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't Delete Document", Error: err})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success"})
}

func AdminAssignDocument(c *fiber.Ctx) error {
	var input domain.DocumentAssignRequest
	if err := c.BodyParser(&input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Review your input", Error: err})
	}

	admin, ok := c.Locals("admin").(model.Admin)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again"})
	}

	var listAssign []model.UserDocument
	for _, assign := range input.Assignments {
		var documnent model.UserDocument
		userID, err := uuid.Parse(assign.UserID)
		if err != nil {
			return c.JSON(base.BaseResponse{
				Status:  false,
				Message: "Invalid user_id",
				Error:   err,
			})
		}
		documnent.UserID = &userID
		documnent.DocumentID = assign.DocumentID
		documnent.AdminID = &admin.ID
		now := time.Now()
		documnent.AssignAt = &now
		listAssign = append(listAssign, documnent)
	}

	if err := repo.AdminAssignDocumentToUser(listAssign); err != nil {
		return c.JSON(base.BaseResponse{
			Status:  false,
			Message: "Can't Assign Document",
			Error:   err,
		})
	}

	return c.JSON(base.BaseResponse{Status: true, Message: "Success"})
}

// func AdminUnassignDocument(c *fiber.Ctx) error {
// 	var input domain.DocumentUnassignRequest
// 	if err := c.BodyParser(&input); err != nil {
// 		return c.JSON(base.BaseResponse{Status: false, Message: "Review your input", Error: err})
// 	}

// 	var listUnassign []model.UserDocument
// 	for _, assign := range input.Unassignments {
// 		var documnent model.UserDocument
// 		userID, err := uuid.Parse(assign.UserID)
// 		if err != nil {
// 			return c.JSON(base.BaseResponse{
// 				Status:  false,
// 				Message: "Invalid user_id",
// 				Error:   err,
// 			})
// 		}
// 		documnent.UserID = userID
// 		documnent.DocumentID = assign.DocumentID
// 		listUnassign = append(listUnassign, documnent)
// 	}

// 	rows, err := repo.AdminUnassignDocumentFromUser(listUnassign)
// 	if err != nil {
// 		return c.JSON(base.BaseResponse{Status: false, Message: "Can't Unassign Document", Error: err})
// 	}
// 	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Total: &rows})
// }

func UserGetUserDocuments(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(model.User)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again", Error: nil})
	}

	var query repo.Query
	query.Parse(c)
	query.UserID = &user.ID

	items, total, err := repo.AdminGetDocuments(query)
	// items, total, err := repo.UserGetUserDocuments(query)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Failed to get user documents", Error: err})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: items, Total: &total})
}

// NOTE: User Create/Update Document (for uploading new document)
func UserCreateDocument(c *fiber.Ctx) error {
	var input domain.DocumentCreateRequest
	if err := c.BodyParser(&input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Review your input", Error: err})
	}
	doc := model.Document{
		Name:         input.Name,
		Description:  &input.Description,
		FileURL:      input.FileURL,
		FileType:     input.FileType,
		DocumentType: input.DocumentType,
		Size:         input.Size,
	}
	if err := repo.UserCreateDocument(&doc); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't Create Document", Error: err})
	}
	// if err := repo.UserCreateUserDocument(&doc); err != nil {
	// 	return c.JSON(base.BaseResponse{Status: false, Message: "Can't Create Document", Error: err})
	// }
	return c.JSON(base.BaseResponse{Status: true, Message: "Success"})
}

func UserUpdateDocument(c *fiber.Ctx) error {
	var input domain.DocumentUpdateRequest
	if err := c.BodyParser(&input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Review your input", Error: err})
	}
	id := c.Params("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Invalid id", Error: err})
	}
	doc, err := repo.UserGetDocumentByID(idInt)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(base.BaseResponse{Status: false, Message: "Document not found", Error: nil})
		}
		return c.JSON(base.BaseResponse{Status: false, Message: "Get Document Failed", Error: err})
	}
	if input.Name != "" {
		doc.Name = input.Name
	}
	if input.Description != "" {
		doc.Description = &input.Description
	}
	if input.FileURL != "" {
		doc.FileURL = input.FileURL
	}
	if input.FileType != "" {
		doc.FileType = input.FileType
	}
	if input.DocumentType != "" {
		doc.DocumentType = input.DocumentType
	}
	if input.Size != 0 {
		doc.Size = input.Size
	}
	if err := repo.UserUpdateDocument(&doc); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't Update Document", Error: err})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success"})
}

func UserDeleteDocument(c *fiber.Ctx) error {
	var input domain.DocumentDeleteRequest
	if err := c.BodyParser(&input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Invalid id", Error: err})
	}
	if err := repo.UserDeleteDocumentByIDs(input.IDs); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't Delete Document", Error: err})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success"})
}
