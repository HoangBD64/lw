package domain

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type CreateContactFormRequest struct {
	Name        string          `json:"name,omitempty"`
	Email       *string         `json:"email,omitempty" validate:"required,email"`
	Phone       *string         `json:"phone,omitempty"`
	Description *datatypes.JSON `json:"description,omitempty" validate:"required"`
	UserID      *uuid.UUID      `json:"user_id,omitempty"`
}
