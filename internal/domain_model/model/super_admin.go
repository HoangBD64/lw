package model

import (
	"time"

	"gorm.io/datatypes"
)

type SuperAdmin struct {
	Model       `gorm:"embedded"`
	Name        string         `json:"name"`
	Email       *string        `json:"email"`
	Phone       *string        `json:"phone"`
	Role        string         `json:"role" gorm:"default:super_admin"`
	Status      string         `json:"status" gorm:"default:inactive"`
	Avatar      *string        `json:"avatar" gorm:"size:2000"`
	SsoId       *string        `gorm:"index" json:"sso_id,omitempty"`
	IsOwner     bool           `json:"is_owner"`
	LastLogin   *time.Time     `json:"last_login"`
	Permissions datatypes.JSON `json:"permissions"`
}

type SuperAdminPermission struct {
	Admin     Permission `json:"admin"`
	Setting   Permission `json:"setting"`
	Package   Permission `json:"package"`
	Catalog   Permission `json:"catalog"`
	Dashboard Permission `json:"dashboard"`
}

const ( // system role
	SUPER_ADMIN = "super_admin"
)

func (SuperAdmin) AdminType() string {
	return SUPER_ADMIN
}

// both admin and super admin implement this
type SystemAdmin interface {
	AdminType() string
}
