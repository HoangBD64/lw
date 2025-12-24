package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NotificationPreference struct {
	ID           int        `gorm:"primaryKey,autoIncrement" json:"id"`
	UserID       *uuid.UUID `json:"user_id"`
	EmailEnabled bool       `json:"email_enabled"`

	CreatedAt *time.Time      `json:"created_at,omitempty"`
	UpdatedAt *time.Time      `json:"updated_at,omitempty"`
	DeletedAt *gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

type Notification struct {
	ID      int        `gorm:"primaryKey,autoIncrement" json:"id"`
	UserID  *uuid.UUID `json:"user_id"`
	AdminID *uuid.UUID `json:"admin_id"`
	Title   string     `json:"title"`
	Content string     `json:"content" gorm:"type:text"`
	Type    string     `json:"type"  gorm:"index"`
	ReadAt  *time.Time `json:"read_at"`

	CreatedAt *time.Time      `json:"created_at,omitempty"`
	UpdatedAt *time.Time      `json:"updated_at,omitempty"`
	DeletedAt *gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

const ( //notification type
	NOTIFICATION_PROCEDURE   = "procedure"
	NOTI_INVOICE_COMPLETED   = "invoice_completed"
	NOTI_PROCEDURE_ERROR     = "procedure_error"
	NOTI_PROCEDURE_COMPLETED = "procedure_completed"
)
