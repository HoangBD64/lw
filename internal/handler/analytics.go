package handler

import (
	"core/internal/domain_model/model"
	"core/internal/repo"
	base "core/utils/base"

	"github.com/gofiber/fiber/v2"
)

func AdminGetSystemStats(c *fiber.Ctx) error {
	_, ok := c.Locals("admin").(model.Admin)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again as admin"})
	}

	var q repo.Query
	q.Parse(c)
	if q.StartTime == nil || q.EndTime == nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "must specify time range"})
	}

	statusCounts, err := repo.GetUserProcedureStatusCounts(c.Context(), q)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "failed to get status counts", Error: err})
	}

	templateUsages, err := repo.GetTemplateUsages(c.Context(), q)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "failed to get template usages", Error: err})
	}

	newUsers, err := repo.GetAnalyticsNewUser(c.Context(), q)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "failed to get new users", Error: err})
	}

	data := fiber.Map{
		"user_procedure_status": statusCounts,
		"template_usage":        templateUsages,
		"new_users":             newUsers,
	}

	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: data})
}
