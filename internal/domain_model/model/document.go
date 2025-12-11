package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Document struct {
	ID          int             `gorm:"primaryKey,autoIncrement" json:"id"`
	AdminID     uuid.UUID       `gorm:"admin_id" json:"admin_id"`
	Admin       Admin           `json:"admin"`
	Name        string          `gorm:"name" json:"name"`
	UpdateBy    string          `gorm:"update_by" json:"update_by"`
	Description *string         `gorm:"index" json:"description,omitempty"`
	FileURL     string          `gorm:"file_url" json:"file_url"`
	FileType    string          `gorm:"file_type" json:"file_type,omitempty"`
	Size        int64           `gorm:"size" json:"size,omitempty"`
	CreatedAt   time.Time       `gorm:"created_at" json:"created_at"`
	UpdatedAt   time.Time       `gorm:"updated_at" json:"updated_at"`
	DeletedAt   *gorm.DeletedAt `gorm:"deleted_at,index" json:"deleted_at,omitempty"`
}
