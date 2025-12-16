package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Procedure struct {
	ID           int        `gorm:"primaryKey,autoIncrement" json:"id"`
	AdminID      *uuid.UUID `json:"admin_id,omitempty"`
	Admin        *Admin     `json:"admin,omitempty"`
	Name         string     `json:"name"`
	UpdateBy     string     `json:"update_by,omitempty"`
	Description  *string    `json:"description,omitempty"`
	Cost         *float64   `json:"cost"`

	CreatedAt *time.Time      `json:"created_at,omitempty"`
	UpdatedAt *time.Time      `json:"updated_at,omitempty"`
	DeletedAt *gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	IsActive       *bool           `json:"is_active" gorm:"index"`
	DueDate        *int            `json:"due_date"`
	ProcedureSteps []ProcedureStep `json:"procedure_steps,omitempty"`
}

type ProcedureStep struct {
	ID          int            `gorm:"primaryKey,autoIncrement" json:"id"`
	ProcedureID int            `json:"procedure_id"`
	Name        string         `json:"name"`
	Status      string         `json:"status" gorm:"index"`
	DueDate     *int           `json:"due_date"`
	AdminID     *uuid.UUID     `json:"admin_id,omitempty"`
	Admin       *Admin         `json:"admin,omitempty"`
	Notes       string         `json:"notes"`
	Position    int            `json:"position"`
	IsActive    *bool          `json:"is_active"`
	Data        datatypes.JSON `json:"data"`
	Description *string        `json:"description,omitempty"`

	CreatedAt *time.Time      `json:"created_at,omitempty"`
	UpdatedAt *time.Time      `json:"updated_at,omitempty"`
	DeletedAt *gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

type UserProcedureStep struct {
	ID              int            `gorm:"primaryKey,autoIncrement" json:"id"`
	UserProcedureID int            `json:"user_procedure_id"`
	ProcedureStepID int            `json:"procedure_step_id"`
	Name            string         `json:"name"`
	DueDate         *int           `json:"due_date"`
	AdminID         *uuid.UUID     `json:"admin_id,omitempty"`
	Admin           *Admin         `json:"admin,omitempty"`
	Notes           string         `json:"notes"`
	Position        int            `json:"position"`
	IsActive        *bool          `json:"is_active"`
	Status          string         `json:"status" gorm:"index"`
	Data            datatypes.JSON `json:"data"`
	Description     *string        `json:"description,omitempty"`

	CreatedAt *time.Time      `json:"created_at,omitempty"`
	UpdatedAt *time.Time      `json:"updated_at,omitempty"`
	DeletedAt *gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

type UserProcedure struct {
	ID                 int                 `gorm:"primaryKey,autoIncrement" json:"id"`
	UserID             *uuid.UUID          `json:"user_id,omitempty"`
	User               *User               `json:"user,omitempty"`
	ProcedureID        int                 `json:"procedure_id"`
	Procedure          *Procedure          `json:"procedure,omitempty"`
	AdminID            *uuid.UUID          `json:"admin_id,omitempty"`
	Admin              *Admin              `json:"admin,omitempty"`
	Name               string              `json:"name"`
	IsHaveTicket       *bool               `json:"is_have_ticket,omitempty"`
	Status             string              `json:"status" gorm:"index"`
	DueDate            *int                `json:"due_date"`
	UserProcedureSteps []UserProcedureStep `json:"user_procedure_steps,omitempty"`
	Cost               *float64            `json:"cost,omitempty"`

	CreatedAt *time.Time      `json:"created_at,omitempty"`
	UpdatedAt *time.Time      `json:"updated_at,omitempty"`
	DeletedAt *gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

const (
	PROCEDURE_STEP_TODO        = "todo"
	PROCEDURE_STEP_IN_PROGRESS = "in_progress"
	PROCEDURE_STEP_DONE        = "done"
	PROCEDURE_STEP_BLOCKED     = "blocked"

	PROCEDURE_CREATED = "created"
)
