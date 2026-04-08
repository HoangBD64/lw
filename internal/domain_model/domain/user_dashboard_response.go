package domain

import (
	"time"

	"github.com/google/uuid"
)

type UserDashboardResponse struct {
	InformationRequired int64 `json:"information_required"`
	Inprogress          int64 `json:"in_progress"`
	Created             int64 `json:"created"`
	Completed           int64 `json:"completed"`
}

type AdminDashboardAssignedProcedure struct {
	ID           int        `json:"id"`
	ProcNumber   string     `json:"proc_number,omitempty"`
	TemplateName string     `json:"template_name,omitempty"`
	CompanyName  *string    `json:"company_name,omitempty"`
	Status       *string    `json:"status,omitempty"`
	Priority     *string    `json:"priority,omitempty"`
	DeadlineAt   *time.Time `json:"deadline_at,omitempty"`
	AdminID      *uuid.UUID `json:"admin_id,omitempty"`
	AdminName    string     `json:"admin_name,omitempty"`
	UserID       *uuid.UUID `json:"user_id,omitempty"`
	UserName     string     `json:"user_name,omitempty"`
}

type AdminDashboardLawyerWorkload struct {
	AdminID             uuid.UUID `json:"admin_id"`
	AdminName           string    `json:"admin_name"`
	ActiveProcedures    int64     `json:"active_procedures"`
	CompletedProcedures int64     `json:"completed_procedures"`
}

type AdminDashboardUserWorkload struct {
	UserID              uuid.UUID `json:"user_id"`
	UserName            string    `json:"user_name"`
	ActiveProcedures    int64     `json:"active_procedures"`
	CompletedProcedures int64     `json:"completed_procedures"`
}

type AdminDashboardResponse struct {
	AssignedProcedures []AdminDashboardAssignedProcedure `json:"assigned_procedures"`
	LawyerWorkloads    []AdminDashboardLawyerWorkload    `json:"lawyer_workloads"`
	UserWorkloads      []AdminDashboardUserWorkload      `json:"user_workloads"`
	AssignedTotal      int64                             `json:"assigned_total"`
}
