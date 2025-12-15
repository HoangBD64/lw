package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type NotificationPreference struct {
	Model
	UserID       *uuid.UUID `json:"user_id"`
	EmailEnabled bool       `json:"email_enabled"`
	InAppEnabled bool       `json:"in_app_enabled"`
	DigestDaily  bool       `json:"digest_daily"`
}

type Notification struct {
	Model
	UserID   *uuid.UUID     `json:"user_id"`
	Title    string         `json:"title"`
	Body     string         `json:"body"`
	Metadata datatypes.JSON `json:"metadata"`
	ReadAt   *time.Time     `json:"read_at"`
}
