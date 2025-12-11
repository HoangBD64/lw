package domain

import (
	"time"

	"github.com/google/uuid"
)

type DataUserResponse struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Avatar   *string   `json:"avatar,omitempty"`
	Email    *string   `json:"email,omitempty"`
	Phone    *string   `json:"phone,omitempty"`
	Gender   string    `json:"gender,omitempty"`
	Address  string    `json:"address,omitempty"`
	Status   string    `json:"status,omitempty"`
	Role     string    `json:"role,omitempty"`
	DOB      string    `json:"dob,omitempty"`
	SsoID    *string   `json:"sso_id"`
	Token    string    `json:"token"`
	UserType string    `json:"user_type"`
	IsOwner  bool      `json:"is_owner"`
}

type GetUsersResponse struct {
	ID        uuid.UUID  `json:"id,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
	Name      string     `json:"name"`
	Email     *string    `json:"email,omitempty"`
	Phone     *string    `json:"phone,omitempty"`
	Role      string     `json:"role"`
	Avatar    *string    `json:"avatar,omitempty"`
	SSO_ID    *string    `json:"sso_id,omitempty"`
}
