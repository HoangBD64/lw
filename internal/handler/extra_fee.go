package handler

import (
	"core/internal/domain_model/domain"
	"core/internal/domain_model/model"
	"core/internal/repo"
	"core/utils/base"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func AdminCreateExtraFee(c *fiber.Ctx) error {
	var req domain.AdminCreateExtraFeeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Invalid request", Error: err})
	}
	err := appUtils.Validate(c.Context(), &req)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Validation failed", Error: err})
	}

	extraFee := model.ExtraFee{
		OptionsPrices: req.OptionsPrices,
		RegionsPrices: req.RegionsPrices,
		ProcedureID:   req.ProcedureID,
	}

	if err := repo.CreateExtraFee(&extraFee); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Failed to create extra fee", Error: err})
	}

	return c.JSON(base.BaseResponse{Status: true, Message: "Extra fee created successfully", Data: extraFee})
}

func AdminGetExtraFees(c *fiber.Ctx) error {
	input := repo.Query{}
	input.Parse(c)
	if input.ProcedureID == nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Invalid procedure_id"})
	}
	extraFees, err := repo.GetExtraFees(input)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Failed to get extra fees", Error: err})
	}

	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: extraFees})
}

func AdminGetExtraFeeDetail(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Invalid ID", Error: err})
	}
	extraFee, err := repo.GetExtraFeeByID(id)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Failed to get extra fee detail", Error: err})
	}

	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: extraFee})
}
func AdminUpdateExtraFee(c *fiber.Ctx) error {
	var req domain.AdminUpdateExtraFeeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Invalid request", Error: err})
	}

	extraFee, err := repo.GetExtraFeeByID(req.ID)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Failed to get extra fee", Error: err})
	}

	extraFee.OptionsPrices = req.OptionsPrices
	extraFee.RegionsPrices = req.RegionsPrices

	if err := repo.UpdateExtraFee(&extraFee); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Failed to update extra fee", Error: err})
	}

	return c.JSON(base.BaseResponse{Status: true, Message: "Extra fee updated successfully", Data: extraFee})
}

func AdminDeleteExtraFee(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Invalid ID", Error: err})
	}

	if err := repo.DeleteExtraFee(id); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Failed to delete extra fee", Error: err})
	}

	return c.JSON(base.BaseResponse{Status: true, Message: "Extra fee deleted successfully"})
}
