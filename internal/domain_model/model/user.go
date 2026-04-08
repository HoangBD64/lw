package model

import (
	"time"

	"gorm.io/datatypes"
)

type User struct {
	Model      `gorm:"embedded"`
	Name       string         `json:"name"`
	Email      *string        `gorm:"index" json:"email,omitempty"`
	Phone      *string        `gorm:"index" json:"phone,omitempty"`
	Role       string         `json:"role,omitempty"`
	Status     string         `json:"status,omitempty" gorm:"default:active;index"`
	Gender     string         `json:"gender,omitempty"`
	DOB        string         `json:"dob,omitempty"`
	Address    string         `json:"address,omitempty"`
	Avatar     *string        `json:"avatar,omitempty" gorm:"type:text"`
	SSO_ID     *string        `gorm:"Index" json:"sso_id,omitempty"`
	LastLogin  *time.Time     `json:"last_login,omitempty"`
	IsAccept   bool           `json:"is_accept,omitempty" gorm:"default:false"`
	Referral   string         `json:"ref,omitempty"`
	SystemNote datatypes.JSON `json:"system_note,omitempty"`
	Metadata   datatypes.JSON `json:"metadata,omitempty"`
}

const ( //status
	ACTIVE   = "active"
	INACTIVE = "inactive"
)
