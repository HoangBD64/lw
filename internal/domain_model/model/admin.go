package model

import (
	"time"

	"gorm.io/datatypes"
)

type Admin struct {
	Model       `gorm:"embedded"`
	Name        string         `json:"name"`
	Email       *string        `gorm:"Index" json:"email,omitempty"`
	Phone       *string        `gorm:"Index" json:"phone,omitempty"`
	Role        string         `json:"role,omitempty" gorm:"default:admin"`
	Status      string         `json:"status,omitempty" gorm:"default:inactive"`
	Gender      string         `json:"gender,omitempty"`
	DOB         string         `json:"dob,omitempty"`
	Avatar      *string        `json:"avatar,omitempty" gorm:"size:2000"`
	SSO_ID      *string        `gorm:"Index" json:"sso_id,omitempty"`
	IsOwner     bool           `json:"is_owner"`
	LastLogin   *time.Time     `json:"last_login,omitempty"`
	Permissions datatypes.JSON `json:"permissions,omitempty"`
}

type AdminPermission struct {
	Admin      Permission `json:"admin"`
	User       Permission `json:"user"`
	Setting    Permission `json:"setting"`
	Document   Permission `json:"document"`
	Procedures Permission `json:"procedures"`
	Assignment Permission `json:"assignment"`
	Ticket     Permission `json:"ticket"`
}

type Permission struct {
	View   bool `json:"view"`
	Manage bool `json:"manage"`
}

const (
	RoleAdmin  = "admin"
	RoleLawyer = "lawyer"
)

func (a Admin) hasRole(role string) bool {
	return a.Role == role
}

func (a Admin) IsAdmin() bool  { return a.hasRole(RoleAdmin) }
func (a Admin) IsLawyer() bool { return a.hasRole(RoleLawyer) }

func (Admin) AdminType() string {
	return RoleAdmin
}
