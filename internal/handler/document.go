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
	var input domain.DocumentCreateRequest
	if err := c.BodyParser(&input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Review your input", Error: err})
	}
	doc := model.Document{
		Name:        input.Name,
		Description: input.Description,
		FileURL:     input.FileURL,
		FileType:    input.FileType,
		Size:        input.Size,
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
	if input.Description != nil {
		doc.Description = input.Description
	}
	if input.FileURL != "" {
		doc.FileURL = input.FileURL
	}
	if input.FileType != nil {
		doc.FileType = *input.FileType
	}
	if input.Size != nil {
		doc.Size = *input.Size
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
	items, total := repo.AdminGetDocuments(q)
	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: fiber.Map{"items": items, "total": total}})
}

func AdminDeleteDocument(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Invalid id", Error: err})
	}
	if err := repo.AdminDeleteDocumentByID(id); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't Delete Document", Error: err})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success"})
}
