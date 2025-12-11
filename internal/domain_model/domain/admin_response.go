package domain

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type DataAdminResponse struct {
	ID          uuid.UUID      `json:"id"`
	Name        string         `json:"name"`
	Avatar      *string        `json:"avatar,omitempty"`
	Email       *string        `json:"email,omitempty"`
	Phone       *string        `json:"phone,omitempty"`
	Gender      string         `json:"gender,omitempty"`
	Address     string         `json:"address,omitempty"`
	Role        string         `json:"role,omitempty"`
	DOB         string         `json:"dob,omitempty"`
	SsoID       *string        `json:"sso_id"`
	IsOwner     bool           `json:"is_owner"`
	Token       string         `json:"token"`
	Status      string         `json:"status"`
	UserType    string         `json:"user_type"`
	Permissions datatypes.JSON `json:"permissions,omitempty"`
}
