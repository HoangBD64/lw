package domain

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type ProcedureCreateRequest struct {
	Name        string  `json:"name" validate:"required"`
	Description *string `json:"description,omitempty"`
	DueDate     *int    `json:"due_date,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
	Cost        float64 `json:"cost,omitempty" validate:"required,gte=0"`
	ExtraFeeID  *int    `json:"extra_fee_id,omitempty"`
}

type ProcedureUpdateRequest struct {
	Name        string   `json:"name"`
	Description *string  `json:"description,omitempty"`
	DueDate     *int     `json:"due_date,omitempty"`
	IsActive    *bool    `json:"is_active,omitempty"`
	Cost        *float64 `json:"cost,omitempty"`
	ExtraFeeID  *int     `json:"extra_fee_id,omitempty"`
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
	IsActive    *bool          `json:"is_active" validate:"required"`
	Description *string        `json:"description,omitempty"`
	Data        datatypes.JSON `json:"data"`
}

type AdminUpdateProcedureStepRequest struct {
	Name        string         `json:"name"`
	DueDate     *int           `json:"due_date"`
	AdminID     *uuid.UUID     `json:"admin_id"`
	Notes       string         `json:"notes"`
	Position    int            `json:"position"`
	IsActive    *bool          `json:"is_active"`
	Description *string        `json:"description,omitempty"`
	Data        datatypes.JSON `json:"data"`
}

type AdminDeleteProcedureStepRequest struct {
	ID int `json:"id" validate:"required,gt=0"`
}

type AdminBulkUpsertProcedureRequest struct {
	Procedure      AdminBulkUpsertProcedure       `json:"procedure" validate:"required"`
	ProcedureSteps []AdminBulkUpsertProcedureStep `json:"procedure_steps"`
	PruneMissing   bool                           `json:"prune_missing"` //remove steps not in payload
}

type AdminBulkUpsertProcedure struct {
	ID          *int       `json:"id"`
	AdminID     *uuid.UUID `json:"admin_id"`
	Name        string     `json:"name" validate:"required"`
	Description *string    `json:"description,omitempty"`
	DueDate     *int       `json:"due_date,omitempty"`
	IsActive    *bool      `json:"is_active,omitempty"`
	Cost        *float64   `json:"cost,omitempty" validate:"required,gte=0"`
	ExtraFeeID  *int       `json:"extra_fee_id,omitempty"`
}

type AdminBulkUpsertProcedureStep struct {
	ID          *int           `json:"id"`
	Name        string         `json:"name" validate:"required"`
	DueDate     *int           `json:"due_date"`
	AdminID     *uuid.UUID     `json:"admin_id"`
	Notes       string         `json:"notes"`
	Position    int            `json:"position"`
	IsActive    *bool          `json:"is_active"`
	Data        datatypes.JSON `json:"data"`
	Description *string        `json:"description,omitempty"`
}

type AssignUserProcedureRequest struct {
	UserProcedureID []int      `json:"user_procedure_id" validate:"required,gt=0"`
	AdminID         *uuid.UUID `json:"admin_id" validate:"required"`
}

type MarkReviewUserProcedureRequest struct {
	UserProcedureID []int `json:"user_procedure_id" validate:"required,gt=0"`
}

type AdminUpdateUserProcedureStatusRequest struct {
	UserProcedureID []int  `json:"user_procedure_id" validate:"required,gt=0"`
	Action          string `json:"action" validate:"required,oneof=hold resume mark_review"`
}

type UnassignUserProcedureRequest struct {
	UserProcedureID []int `json:"user_procedure_id" validate:"required,gt=0"`
}

type UserUpdateUserProcedureStepRequest struct {
	Data   datatypes.JSON `json:"data"`
	Status string         `json:"status"`
	// UserCompleted *bool          `json:"user_completed,omitempty"`
}

type AdminUpdateUserProcedureRequest struct {
	Status   string `json:"status"`
	Note     string `json:"note"`
	Priority string `json:"priority,omitempty"`
	IsLock   *bool  `json:"is_lock"`
}

type AdminAssignLawyerToProcedureStepRequest struct {
	AssignLawyerToProcedureSteps []AssignLawyerToProcedureStep `json:"assignments" validate:"required,gt=0"`
}

type AssignLawyerToProcedureStep struct {
	UserProcedureStepID int        `json:"user_procedure_step_id" validate:"required,gt=0"`
	AdminID             *uuid.UUID `json:"admin_id"`
}

type AdminUpdateUserProcedureStepStatusRequest struct {
	UserProcedureStepID int    `json:"user_procedure_step_id" validate:"required,gt=0"`
	Status              string `json:"status" validate:"required"`
}

type AdminUserUpdateProcedureStepRequest struct {
	FeedBack datatypes.JSON `json:"feedback"`
	Notes    string         `json:"notes"`
	Status   string         `json:"status"`
	// UserCompleted *bool          `json:"user_completed,omitempty"`
}

// AdminMigrateStatusRequest represents request to migrate status
type AdminMigrateStatusRequest struct {
	DryRun bool `json:"dry_run"` // If true, only return count without updating
}

// AdminMigrateStatusResponse represents response for status migration
type AdminMigrateStatusResponse struct {
	UserProceduresUpdated     int64  `json:"user_procedures_updated"`
	UserProcedureStepsUpdated int64  `json:"user_procedure_steps_updated"`
	Message                   string `json:"message"`
}
