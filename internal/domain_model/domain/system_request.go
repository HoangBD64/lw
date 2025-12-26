package domain

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type AdminUpdateSettingRequest struct {
	ID       uuid.UUID      `json:"id" validate:"required"`
	Resource string         `json:"resource"`
	Metadata datatypes.JSON `json:"metadata"`
	Contant  float64        `json:"contant"`
}

type TestSendMailRequest struct {
	Address string `json:"address" validate:"required,email"`
}
