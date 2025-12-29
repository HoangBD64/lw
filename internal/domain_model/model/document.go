package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Document struct {
	ID              int           `gorm:"primaryKey,autoIncrement" json:"id"`
	AdminID         *uuid.UUID    `json:"admin_id,omitempty"`
	Admin           *Admin        `json:"admin,omitempty"`
	UserID          *uuid.UUID    `json:"user_id,omitempty"`
	User            *User         `json:"user,omitempty"`
	Name            string        `json:"name"`
	UserProcedureID *int          `json:"user_procedure_id"`
	UserProcedure   UserProcedure `json:"user_procedure,omitempty"`
	UpdateBy        string        `json:"update_by,omitempty"`
	Description     string        `json:"description,omitempty"`
	FileURL         string        `json:"file_url"`
	FileType        string        `json:"file_type,omitempty"`
	DocumentType    string        `json:"document_type,omitempty"`
	Size            int64         `json:"size,omitempty"`
	Type            string        `json:"type"` // user_upload | admin_upload

	CreatedAt *time.Time      `json:"created_at,omitempty"`
	UpdatedAt *time.Time      `json:"updated_at,omitempty"`
	DeletedAt *gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

const (
	ADMIN_UPLOAD_DOCUMENT = "admin_upload"
	USER_UPLOAD_DOCUMENT  = "user_upload"
)
