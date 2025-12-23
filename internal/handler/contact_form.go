package handler

import (
	"core/internal/domain_model/domain"
	"core/internal/domain_model/model"
	"core/internal/repo"
	"core/utils/base"

	"github.com/gofiber/fiber/v2"
	"github.com/samber/lo"
)

func CreateContactForm(c *fiber.Ctx) error {
	var input domain.CreateContactFormRequest
	if err := c.BodyParser(&input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Invalid request body", Error: err})
	}

	if err := appUtils.Validate(c.Context(), &input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Validation failed", Error: err})
	}

	contactForm := model.ContactForm{
		Name:        input.Name,
		Email:       lo.FromPtr(input.Email),
		Phone:       lo.FromPtr(input.Phone),
		Description: lo.FromPtr(input.Description),
		UserID:      input.UserID,
	}
	_, err := repo.CreateContactForm(c.Context(), contactForm)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Failed to create contact form", Error: err})
	}

	return c.JSON(base.BaseResponse{Status: true, Message: "Success"})
}

func AdminGetContactForms(c *fiber.Ctx) error {
	admin := c.Locals("admin").(model.Admin)
	if !admin.IsAdmin() && !admin.IsOwner {
		return c.JSON(base.BaseResponse{Status: false, Message: "unauthorized"})
	}

	query := repo.Query{}
	query.Parse(c)
	resp, total, err := repo.GetContactForms(c.Context(), query)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Failed to get contact forms", Error: err})
	}

	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: resp, Total: &total})
}
