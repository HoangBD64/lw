package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type ActivityLog struct {
	Action      string         `json:"action"`
	ActorType   string         `json:"actor_type,omitempty"`
	ActorID     string         `json:"actor_id,omitempty"`
	ActorName   string         `json:"actor_name,omitempty"`
	Description string         `json:"description"`
	Metadata    datatypes.JSON `json:"metadata,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
}

type Procedure struct {
	ID             int             `gorm:"primaryKey,autoIncrement" json:"id"`
	AdminID        *uuid.UUID      `json:"admin_id,omitempty"`
	Admin          *Admin          `json:"admin,omitempty"`
	Name           string          `json:"name"`
	UpdateBy       string          `json:"update_by,omitempty"`
	Description    *string         `json:"description,omitempty"`
	Cost           *float64        `json:"cost,omitempty"`
	ExtraFeeID     *int            `json:"extra_fee_id,omitempty"`
	ExtraFees      ExtraFee        `json:"extra_fees,omitempty"`
	ProcedureSteps []ProcedureStep `json:"procedure_steps,omitempty"`

	CreatedAt *time.Time      `json:"created_at,omitempty"`
	UpdatedAt *time.Time      `json:"updated_at,omitempty"`
	DeletedAt *gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	IsActive *bool `json:"is_active,omitempty" gorm:"default:true"`
	DueDate  *int  `json:"due_date,omitempty"`
}

type ProcedureStep struct {
	ID          int            `gorm:"primaryKey,autoIncrement" json:"id"`
	ProcedureID int            `json:"procedure_id"`
	Name        string         `json:"name"`
	Status      string         `json:"status,omitempty" gorm:"index"`
	DueDate     *int           `json:"due_date,omitempty"`
	AdminID     *uuid.UUID     `json:"admin_id,omitempty"`
	Admin       *Admin         `json:"admin,omitempty"`
	Notes       string         `json:"notes,omitempty"`
	Position    int            `json:"position,omitempty"`
	IsActive    *bool          `json:"is_active,omitempty" gorm:"default:true"`
	Data        datatypes.JSON `json:"data,omitempty"`
	Description *string        `json:"description,omitempty"`

	CreatedAt *time.Time      `json:"created_at,omitempty"`
	UpdatedAt *time.Time      `json:"updated_at,omitempty"`
	DeletedAt *gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

type UserProcedureStep struct {
	ID              int            `gorm:"primaryKey,autoIncrement" json:"id"`
	ProcedureStepID int            `json:"procedure_step_id"`
	UserProcedureID int            `json:"user_procedure_id"`
	UserProcedure   *UserProcedure `json:"user_procedure,omitempty"`
	Documents       []Document     `json:"documents,omitempty" gorm:"foreignKey:UserProcedureStepID"`
	Name            string         `json:"name"`
	DueDate         *int           `json:"due_date,omitempty"`
	AdminID         *uuid.UUID     `json:"admin_id,omitempty"`
	Admin           *Admin         `json:"admin,omitempty"`
	Notes           string         `json:"notes,omitempty"`
	Position        int            `json:"position,omitempty"`
	Status          string         `json:"status,omitempty" gorm:"index"`
	Data            datatypes.JSON `json:"data,omitempty"`
	FeedBack        datatypes.JSON `json:"feedback,omitempty"`
	Description     *string        `json:"description,omitempty"`
	UserUpdated     *bool          `json:"user_updated,omitempty"`
	// UserCompleted   *bool          `json:"user_completed,omitempty" gorm:"default:false"`

	CreatedAt *time.Time      `json:"created_at,omitempty"`
	UpdatedAt *time.Time      `json:"updated_at,omitempty"`
	DeletedAt *gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

type UserProcedure struct {
	ID                  int                 `gorm:"primaryKey,autoIncrement" json:"id"`
	UserID              *uuid.UUID          `json:"user_id,omitempty"`
	User                *User               `json:"user,omitempty"`
	ProcedureID         int                 `json:"procedure_id,omitempty"`
	Procedure           *Procedure          `json:"procedure,omitempty"`
	AdminID             *uuid.UUID          `json:"admin_id,omitempty"`
	Admin               *Admin              `json:"admin,omitempty"`
	TemplateName        string              `json:"template_name,omitempty"`
	CompanyCode         string              `json:"company_code,omitempty"`
	ProcNumber          string              `json:"proc_number,omitempty" gorm:"index"`
	State               *string             `json:"state,omitempty"`
	IsHaveTicket        *bool               `json:"is_have_ticket,omitempty"`
	IsSendNoti          *bool               `json:"is_send_noti,omitempty" gorm:"default:false"`
	Status              *string             `json:"status,omitempty" gorm:"index"`
	SystemStatus        *string             `json:"system_status,omitempty" gorm:"index"`
	Note                *string             `json:"note,omitempty"`
	CompanyName         *string             `json:"company_name,omitempty"`
	CompanyType         *string             `json:"company_type,omitempty"`
	DueDate             *int                `json:"due_date,omitempty"`
	PaymentAccount      *string             `json:"payment_account,omitempty"`
	UserProcedureSteps  []UserProcedureStep `json:"user_procedure_steps,omitempty"`
	Cost                *float64            `json:"cost,omitempty"`
	FulfillmentStatus   *string             `json:"fulfillment_status,omitempty"`
	Title               *string             `json:"title,omitempty"`
	Priority            *string             `json:"priority,omitempty" gorm:"default:low;index"`
	ActivityLogs        datatypes.JSON      `json:"activity_logs,omitempty" gorm:"type:jsonb"`
	FulfillmentCostPaid *time.Time          `json:"fulfillment_cost_paid,omitempty"`
	DeadlineAt          *time.Time          `json:"deadline_at,omitempty"`

	CreatedAt *time.Time      `json:"created_at,omitempty"`
	UpdatedAt *time.Time      `json:"updated_at,omitempty"`
	DeletedAt *gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

const (
	PROCEDURE_STEP_TODO = "todo"

	// USER_PROCEDURE_PENDING              = "pending"              // mới thanh toán xong
	USER_PROCEDURE_CREATED        = "created"        // mới thanh toán xong
	USER_PROCEDURE_SUBMITTED      = "submitted"      // đã nộp hồ sơ, đang chờ admin review
	USER_PROCEDURE_PENDING_REVIEW = "pending_review" // admin đã tiếp nhận, đang review hồ sơ để assign
	USER_PROCEDURE_IN_PROGRESS    = "in_progress"    // đã được assign, đang trong quá trình xử lý
	// USER_PROCEDURE_INFORMATION_REVIEW   = "information-review"   // admin đang xem thông tin hoặc mới dược assign
	USER_PROCEDURE_COMPLETED            = "completed"            // đã hoàn thành tất cả các bước
	USER_PROCEDURE_INFORMATION_REQUIRED = "information-required" // user cần bổ sung thông tin
	USER_PROCEDURE_CANCELLED            = "cancelled"            // thủ tục bị hủy
	USER_PROCEDURE_FILED                = "filed"                // thủ tục đã nộp
	USER_PROCEDURE_HOLD                 = "hold"                 // thủ tục bị tạm dừng

	// for system
	USER_PROCEDURE_SYSTEM_SUBMIT_REVIEW = "submit_review" // thủ tục cần admin review

	// USER_PROCEDURE_STEP_PENDING              = "pending"
	USER_PROCEDURE_STEP_CREATED = "created"
	// USER_PROCEDURE_STEP_INFORMATION_REVIEW   = "information-review"
	USER_PROCEDURE_STEP_PENDING_REVIEW       = "pending_review"
	USER_PROCEDURE_STEP_IN_PROGRESS          = "in_progress"
	USER_PROCEDURE_STEP_DONE                 = "done"
	USER_PROCEDURE_STEP_INFORMATION_REQUIRED = "information-required"
	USER_PROCEDURE_STEP_CANCELLED            = "cancelled"
	USER_PROCEDURE_STEP_SUBMITTED            = "submitted"

	FULFILLMENT_PAID = "paid"

	USER_PROCEDURE_PRIORITY_LOW    = "low"
	USER_PROCEDURE_PRIORITY_MEDIUM = "medium"
	USER_PROCEDURE_PRIORITY_HIGH   = "high"
)
