package handler

import (
	"core/internal/domain_model/domain"
	"core/internal/domain_model/model"
	repo "core/internal/repo"
	"core/utils/base"
	"errors"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

func AdminCreateDocument(c *fiber.Ctx) error {
	admin, ok := c.Locals("admin").(model.Admin)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again"})
	}
	var input domain.DocumentCreateRequest
	if err := c.BodyParser(&input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Review your input", Error: err})
	}
	if input.UserProcedureStepID != nil {
		step, err := repo.GetUserProcedureStepByID(*input.UserProcedureStepID)
		if err != nil {
			return c.JSON(base.BaseResponse{Status: false, Message: "Invalid user procedure step", Error: err})
		}
		if input.UserProcedureID == nil {
			input.UserProcedureID = &step.UserProcedureID
		}
	}
	if input.UserProcedureID == nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "UserProcedureID is required"})
	}
	if admin.IsLawyer() {
		up, err := repo.AdminGetUserProcedureByID(*input.UserProcedureID)
		if err != nil {
			return c.JSON(base.BaseResponse{Status: false, Message: "Get user procedure failed", Error: err})
		}
		if lo.FromPtr(up.AdminID) != admin.ID {
			return c.JSON(base.BaseResponse{Status: false, Message: "Unauthorized"})
		}
	}
	doc := model.Document{
		Name:                input.Name,
		Description:         input.Description,
		FileURL:             input.FileURL,
		FileType:            input.FileType,
		DocumentType:        input.DocumentType,
		Size:                input.Size,
		AdminID:             &admin.ID,
		Type:                model.ADMIN_UPLOAD_DOCUMENT,
		UserProcedureID:     input.UserProcedureID,
		UserProcedureStepID: input.UserProcedureStepID,
		IsPublish:           input.IsPublish,
	}
	if err := repo.AdminCreateDocument(&doc); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't Create Document", Error: err})
	}
	appendProcedureActivity(
		lo.FromPtr(input.UserProcedureID),
		adminActivity(
			admin,
			"document.created",
			fmt.Sprintf("Created document %s", doc.Name),
			js{"document_id": doc.ID, "is_publish": lo.FromPtr(input.IsPublish), "user_procedure_step_id": input.UserProcedureStepID},
		),
	)
	return c.JSON(base.BaseResponse{Status: true, Message: "Success"})
}

func AdminUpdateDocument(c *fiber.Ctx) error {
	admin, ok := c.Locals("admin").(model.Admin)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again"})
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
	if admin.IsLawyer() {
		if doc.UserProcedureID == nil {
			return c.JSON(base.BaseResponse{Status: false, Message: "Unauthorized"})
		}
		up, err := repo.AdminGetUserProcedureByID(*doc.UserProcedureID)
		if err != nil {
			return c.JSON(base.BaseResponse{Status: false, Message: "Get user procedure failed", Error: err})
		}
		if lo.FromPtr(up.AdminID) != admin.ID {
			return c.JSON(base.BaseResponse{Status: false, Message: "Unauthorized"})
		}
	}
	if input.UserProcedureStepID != nil {
		step, err := repo.GetUserProcedureStepByID(*input.UserProcedureStepID)
		if err != nil {
			return c.JSON(base.BaseResponse{Status: false, Message: "Invalid user procedure step", Error: err})
		}
		if input.UserProcedureID == nil {
			input.UserProcedureID = &step.UserProcedureID
		}
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
	if input.UserProcedureStepID != nil {
		doc.UserProcedureStepID = input.UserProcedureStepID
	}
	if input.IsPublish != nil {
		doc.IsPublish = input.IsPublish
	}
	doc.UpdateBy = admin.Name

	if err := repo.AdminUpdateDocument(&doc); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't Update Document", Error: err})
	}
	if doc.UserProcedureID != nil {
		appendProcedureActivity(
			lo.FromPtr(doc.UserProcedureID),
			adminActivity(
				admin,
				"document.updated",
				fmt.Sprintf("Updated document %s", doc.Name),
				js{"document_id": doc.ID, "is_publish": lo.FromPtr(doc.IsPublish), "user_procedure_step_id": doc.UserProcedureStepID},
			),
		)
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success"})
}

func AdminGetDocument(c *fiber.Ctx) error {
	admin, ok := c.Locals("admin").(model.Admin)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again"})
	}
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
	if admin.IsLawyer() && (doc.UserProcedure.AdminID == nil || *doc.UserProcedure.AdminID != admin.ID) {
		return c.JSON(base.BaseResponse{Status: false, Message: "Unauthorized"})
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
	admin, ok := c.Locals("admin").(model.Admin)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again"})
	}

	var input domain.DocumentDeleteRequest
	if err := c.BodyParser(&input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Invalid id", Error: err})
	}
	if admin.IsLawyer() {
		for _, id := range input.IDs {
			doc, err := repo.AdminGetDocumentByID(id)
			if err != nil {
				return c.JSON(base.BaseResponse{Status: false, Message: "Get Document Failed", Error: err})
			}
			if doc.UserProcedure.AdminID == nil || *doc.UserProcedure.AdminID != admin.ID {
				return c.JSON(base.BaseResponse{Status: false, Message: "Unauthorized"})
			}
		}
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
		Name:                input.Name,
		Description:         input.Description,
		FileURL:             input.FileURL,
		FileType:            input.FileType,
		DocumentType:        input.DocumentType,
		Size:                input.Size,
		Type:                model.USER_UPLOAD_DOCUMENT,
		UserID:              &user.ID,
		UserProcedureID:     input.UserProcedureID,
		UserProcedureStepID: input.UserProcedureStepID,
		IsPublish:           lo.ToPtr(true),
	}
	createdDoc, err := repo.UserCreateDocument(doc)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't Create Document", Error: err})
	}
	if input.UserProcedureID != nil {
		appendProcedureActivity(
			lo.FromPtr(input.UserProcedureID),
			userActivity(
				user,
				"document.created",
				fmt.Sprintf("Uploaded document %s", createdDoc.Name),
				js{"document_id": createdDoc.ID, "user_procedure_step_id": input.UserProcedureStepID},
			),
		)
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
	if input.UserProcedureStepID != nil {
		doc.UserProcedureStepID = input.UserProcedureStepID
	}
	if err := repo.UserUpdateDocument(&doc, user.ID); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't Update Document", Error: err})
	}
	if doc.UserProcedureID != nil {
		appendProcedureActivity(
			lo.FromPtr(doc.UserProcedureID),
			userActivity(
				user, "document.updated",
				fmt.Sprintf("Updated document %s", doc.Name),
				js{"document_id": doc.ID, "user_procedure_step_id": doc.UserProcedureStepID},
			),
		)
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success"})
}
