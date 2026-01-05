package domain

import "core/internal/domain_model/model"

type AdminCreateExtraFeeRequest struct {
	OptionsPrices []model.OptionsPrice `json:"option_prices"`
	RegionsPrices []model.RegionsPrice `json:"regions_price" validate:"required"`
	ProcedureID   *int                 `json:"procedure_id" validate:"required"`
}

type AdminUpdateExtraFeeRequest struct {
	ID            int                  `json:"id" validate:"required"`
	OptionsPrices []model.OptionsPrice `json:"option_prices"`
	RegionsPrices []model.RegionsPrice `json:"regions_price" validate:"required"`
	ProcedureID   *int                 `json:"procedure_id" validate:"required"`
}
