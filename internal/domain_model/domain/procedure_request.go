package domain

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type ProcedureCreateRequest struct {
	Name        string  `json:"name" validate:"required"`
	Description *string `json:"description,omitempty"`
	DueDate     *int    `json:"due_date,omitempty"`
}

type ProcedureUpdateRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	DueDate     *int    `json:"due_date,omitempty"`
}

type ProcedureDeleteRequest struct {
	IDs []int `json:"ids" validate:"required,gt=0"`
}

type AdminCreateProcedureStepRequest struct {
	ProcedureID int            `json:"procedure_id" validate:"required,gt=0"`
	Name        string         `json:"name" validate:"required"`
	DueDate     *int           `json:"due_date"`
	AdminID     *uuid.UUID     `json:"admin_id"`
	Notes       string         `json:"notes"`
	Position    int            `json:"position"`
	IsActive    bool           `json:"is_active" validate:"required"`
	Data        datatypes.JSON `json:"data"`
}

type AdminUpdateProcedureStepRequest struct {
	Name     string         `json:"name"`
	DueDate  *int           `json:"due_date"`
	AdminID  *uuid.UUID     `json:"admin_id"`
	Notes    string         `json:"notes"`
	Position int            `json:"position"`
	IsActive bool           `json:"is_active"`
	Data     datatypes.JSON `json:"data"`
}

type AdminDeleteProcedureStepRequest struct {
	ID int `json:"id" validate:"required,gt=0"`
}

type AdminListProcedureStepsRequest struct {
	ID int `json:"id" validate:"required,gt=0"` // procedure_id
}

type UserUpsertProcedureStepDataRequest struct {
	UserProcedureID int            `json:"user_procedure_id"`
	ProcedureStepID int            `json:"procedure_step_id"`
	Status          string         `json:"status"`
	Data            datatypes.JSON `json:"data"`
}

type UserListProcedureStepDataRequest struct {
	ID int `json:"id"` // user_procedure_id
}

type AdminListUserProcedureStepDataRequest struct {
	ID int `json:"id"`
}
