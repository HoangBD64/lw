package domain

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type VerifyUserSSORequest struct {
	Code  string `json:"code" validate:"required"`
	State string `json:"state"`
}

type UpdateUserInfoRequest struct {
	Name     string          `json:"name"`
	Phone    *string         `json:"phone"`
	Gender   string          `json:"gender,omitempty"`
	DOB      string          `json:"dob,omitempty"`
	Address  string          `json:"address,omitempty"`
	Avatar   *string         `json:"avatar,omitempty"`
	Metadata *datatypes.JSON `json:"metadata,omitempty"`
}

type AdminUpdateUserRequest struct {
	ID         uuid.UUID       `json:"id"`
	SystemNote *datatypes.JSON `json:"system_note"`
}

type LoginCustomerAccountRequest struct {
	ID uuid.UUID `json:"id" validate:"required"`
}

type AdminUpdateUserStatusRequest struct {
	IDS    []uuid.UUID `json:"ids" validate:"required,gt=0"`
	Status string      `json:"status" validate:"required,oneof=active inactive"`
}
