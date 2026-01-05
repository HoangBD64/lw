package model

import (
	"time"

	"gorm.io/gorm"
)

type ExtraFee struct {
	ID            int            `gorm:"primaryKey,autoIncrement" json:"id"`
	OptionsPrices []OptionsPrice `json:"option_prices,omitempty" gorm:"type:jsonb;serializer:json"`
	RegionsPrices []RegionsPrice `json:"regions_price,omitempty" gorm:"type:jsonb;serializer:json"`
	ProcedureID   *int           `json:"procedure_id,omitempty"`

	CreatedAt *time.Time      `json:"created_at,omitempty"`
	UpdatedAt *time.Time      `json:"updated_at,omitempty"`
	DeletedAt *gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

type OptionsPrice struct {
	Name string  `json:"name"`
	Cost float64 `json:"cost"`
}

type RegionsPrice struct {
	Regions []string `json:"regions"`
	Cost    float64  `json:"cost"`
}
