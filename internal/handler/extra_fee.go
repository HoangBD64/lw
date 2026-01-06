package handler

import (
	"core/internal/domain_model/domain"
	"core/internal/domain_model/model"
	"core/internal/repo"
	"core/utils/base"
	"strconv"
	"strings"

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

	for i, rp := range extraFee.RegionsPrices {
		for j, region := range rp.Regions {
			extraFee.RegionsPrices[i].Regions[j] = strings.ToLower(strings.TrimSpace(region))
			if extraFee.RegionsPrices[i].Regions[j] == "row" {
				extraFee.RegionsPrices[i].Regions[j] = "ROW"
			}
		}
	}

	input := repo.Query{
		ProcedureID: req.ProcedureID,
	}
	extraFees, err := repo.GetExtraFees(input)
	if len(extraFees) > 0 {
		return c.JSON(base.BaseResponse{Status: false, Message: "Extra fee for this procedure already exists"})
	}

	if duplicated, regions := checkRegionDuplicate(req.RegionsPrices); duplicated {
		return c.JSON(base.BaseResponse{
			Status:  false,
			Message: "Duplicate regions found",
			Data:    regions,
		})
	}

	if err := repo.CreateExtraFee(&extraFee); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Failed to create extra fee", Error: err})
	}

	return c.JSON(base.BaseResponse{Status: true, Message: "Extra fee created successfully", Data: extraFee})
}

func checkRegionDuplicate(prices []model.RegionsPrice) (bool, []string) {
	seen := make(map[string]int)
	duplicates := make(map[string]struct{})

	for _, price := range prices {
		for _, region := range price.Regions {
			seen[region]++
			if seen[region] > 1 {
				duplicates[region] = struct{}{}
			}
		}
	}

	if len(duplicates) == 0 {
		return false, nil
	}

	result := make([]string, 0, len(duplicates))
	for region := range duplicates {
		result = append(result, region)
	}

	return true, result
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

	for i, rp := range extraFee.RegionsPrices {
		for j, region := range rp.Regions {
			extraFee.RegionsPrices[i].Regions[j] = strings.ToLower(strings.TrimSpace(region))
			if extraFee.RegionsPrices[i].Regions[j] == "row" {
				extraFee.RegionsPrices[i].Regions[j] = "ROW"
			}
		}
	}

	if duplicated, regions := checkRegionDuplicate(extraFee.RegionsPrices); duplicated {
		return c.JSON(base.BaseResponse{
			Status:  false,
			Message: "Duplicate regions found",
			Data:    regions,
		})
	}
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
