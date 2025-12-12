package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Document struct {
	ID           int             `gorm:"primaryKey,autoIncrement" json:"id"`
	AdminID      *uuid.UUID      `json:"admin_id,omitempty"`
	Admin        *Admin          `json:"admin,omitempty"`
	Name         string          `json:"name"`
	UpdateBy     string          `json:"update_by,omitempty"`
	Description  *string         `json:"description,omitempty"`
	FileURL      string          `json:"file_url"`
	FileType     string          `json:"file_type,omitempty"`
	DocumentType string          `json:"document_type,omitempty"`
	Size         int64           `json:"size,omitempty"`
	CreatedAt    *time.Time      `json:"created_at,omitempty"`
	UpdatedAt    *time.Time      `json:"updated_at,omitempty"`
	DeletedAt    *gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}
