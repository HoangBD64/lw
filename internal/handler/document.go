package handler

import (
	"core/internal/domain_model/domain"
	"core/internal/domain_model/model"
	repo "core/internal/repo"
	"core/utils/base"
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func AdminCreateDocument(c *fiber.Ctx) error {
	admin, ok := c.Locals("admin").(model.Admin)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again"})
	}
	if !admin.IsLawyer() {
		return c.JSON(base.BaseResponse{Status: false, Message: "Unauthorized"})
	}
	var input domain.DocumentCreateRequest
	if err := c.BodyParser(&input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Review your input", Error: err})
	}
	if input.UserProcedureID == nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "UserProcedureID is required"})
	}
	doc := model.Document{
		Name:            input.Name,
		Description:     input.Description,
		FileURL:         input.FileURL,
		FileType:        input.FileType,
		DocumentType:    input.DocumentType,
		Size:            input.Size,
		AdminID:         &admin.ID,
		Type:            model.ADMIN_UPLOAD_DOCUMENT,
		UserProcedureID: input.UserProcedureID,
	}
	if err := repo.AdminCreateDocument(&doc); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't Create Document", Error: err})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success"})
}

func AdminUpdateDocument(c *fiber.Ctx) error {
	admin, ok := c.Locals("admin").(model.Admin)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again"})
	}
	if !admin.IsLawyer() {
		return c.JSON(base.BaseResponse{Status: false, Message: "Unauthorized"})
	}

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
	doc.Name = input.Name
	doc.FileURL = input.FileURL
	doc.FileType = input.FileType
	doc.DocumentType = input.DocumentType
	doc.Size = input.Size
	doc.Description = input.Description
	if input.UserProcedureID != nil {
		doc.UserProcedureID = input.UserProcedureID
	}
	doc.UpdateBy = admin.Name

	if err := repo.AdminUpdateDocument(&doc, admin.ID); err != nil {
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
	admin, ok := c.Locals("admin").(model.Admin)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again"})
	}
	var q repo.Query
	q.Parse(c)
	if admin.IsLawyer() {
		q.AdminID = &admin.ID
	}
	items, total, err := repo.AdminGetDocuments(q)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Get Documents Failed", Error: err})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: fiber.Map{"items": items, "total": total}})
}

func AdminDeleteDocument(c *fiber.Ctx) error {
	// only lawyer
	admin, ok := c.Locals("admin").(model.Admin)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again"})
	}
	if !admin.IsLawyer() {
		return c.JSON(base.BaseResponse{Status: false, Message: "Unauthorized"})
	}

	var input domain.DocumentDeleteRequest
	if err := c.BodyParser(&input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Invalid id", Error: err})
	}
	total, err := repo.AdminDeleteDocumentByIDs(input.IDs)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't Delete Document", Error: err})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Total: &total})
}

func UserGetUserDocuments(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(model.User)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again", Error: nil})
	}

	var query repo.Query
	query.Parse(c)
	query.UserID = &user.ID

	items, total, err := repo.UserGetDocuments(query)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Failed to get user documents", Error: err})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: items, Total: &total})
}

func UserGetUserDocument(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(model.User)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again", Error: nil})
	}
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Invalid id", Error: err})
	}
	doc, err := repo.UserGetDocumentByID(id, user.ID)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Get Document Failed", Error: err})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: doc})
}

func UserCreateDocument(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(model.User)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again", Error: nil})
	}
	var input domain.DocumentCreateRequest
	if err := c.BodyParser(&input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Review your input", Error: err})
	}
	doc := model.Document{
		Name:            input.Name,
		Description:     input.Description,
		FileURL:         input.FileURL,
		FileType:        input.FileType,
		DocumentType:    input.DocumentType,
		Size:            input.Size,
		Type:            model.USER_UPLOAD_DOCUMENT,
		UserID:          &user.ID,
		UserProcedureID: input.UserProcedureID,
	}
	createdDoc, err := repo.UserCreateDocument(doc)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't Create Document", Error: err})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: createdDoc})
}

func UserUpdateDocument(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(model.User)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again", Error: nil})
	}
	var input domain.DocumentUpdateRequest
	if err := c.BodyParser(&input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Review your input", Error: err})
	}
	id := c.Params("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Invalid id", Error: err})
	}
	doc, err := repo.UserGetDocumentByIDForUpdate(idInt, user.ID)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Get Document Failed", Error: err})
	}
	doc.Name = input.Name
	doc.Description = input.Description
	doc.FileURL = input.FileURL
	doc.FileType = input.FileType
	doc.DocumentType = input.DocumentType
	doc.Size = input.Size
	if input.UserProcedureID != nil {
		doc.UserProcedureID = input.UserProcedureID
	}
	if err := repo.UserUpdateDocument(&doc, user.ID); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't Update Document", Error: err})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success"})
}
