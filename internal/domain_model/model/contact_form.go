package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type ContactForm struct {
	ID          int            `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID      *uuid.UUID     `json:"user_id"`
	Name        string         `json:"name" gorm:"index"`
	Email       string         `json:"email"`
	Phone       string         `json:"phone"`
	Description datatypes.JSON `json:"description"`

	CreatedAt *time.Time      `json:"created_at,omitempty"`
	UpdatedAt *time.Time      `json:"updated_at,omitempty"`
	DeletedAt *gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}
