package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserDocument struct {
	ID         int        `gorm:"primaryKey,autoIncrement" json:"id"`
	UserID     *uuid.UUID `json:"user_id,omitempty"`
	User       *User      `json:"user,omitempty"`
	DocumentID int        `json:"document_id"`
	Document   Document   `json:"document,omitempty"`

	// The admin who assigned the document to the user
	AdminID  *uuid.UUID `json:"admin_id,omitempty"`
	Admin    *Admin     `json:"admin,omitempty"`
	AssignAt *time.Time `json:"assign_at,omitempty"`

	CreatedAt *time.Time      `json:"created_at,omitempty"`
	UpdatedAt *time.Time      `json:"updated_at,omitempty"`
	DeletedAt *gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}
