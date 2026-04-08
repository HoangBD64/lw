package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Ticket struct {
	ID                 int             `json:"id" gorm:"primaryKey;autoIncrement"`
	Category           string          `json:"category,omitempty"`
	TicketCode         string          `json:"ticket_code,omitempty"`
	PreferSolution     string          `json:"prefer_solution,omitempty"`
	Status             string          `json:"status,omitempty" gorm:"index"`
	ResolutionType     string          `json:"resolution_type,omitempty"`
	Type               string          `json:"type,omitempty" gorm:"index"`
	Priority           string          `json:"priority,omitempty" gorm:"default:low;index"`
	Subject            string          `json:"subject,omitempty"`
	InternalNote       string          `json:"internal_note,omitempty"`
	IsInternal         bool            `json:"is_internal,omitempty" gorm:"default:false;index"`
	TicketContents     []TicketContent `json:"ticket_contents,omitempty"`
	UserProcedureID    *int            `json:"user_procedure_id" gorm:"index"`
	UserProcedure      *UserProcedure  `json:"user_procedure,omitempty"`
	UserID             *uuid.UUID      `json:"user_id,omitempty" gorm:"index"`
	User               *User           `json:"user,omitempty"`
	AdminID            *uuid.UUID      `json:"admin_id,omitempty" gorm:"index"`
	Admin              *Admin          `json:"admin,omitempty"`
	CreatorAdminID     *uuid.UUID      `json:"creator_admin_id,omitempty" gorm:"index"`
	CreatorAdmin       *Admin          `json:"creator_admin,omitempty" gorm:"foreignKey:CreatorAdminID;references:ID"`
	ReadAt             *time.Time      `json:"read_at,omitempty" gorm:"index"`               // user read time
	AdminReadAt        *time.Time      `json:"admin_read_at,omitempty" gorm:"index"`         // assignee read time
	CreatorAdminReadAt *time.Time      `json:"creator_admin_read_at,omitempty" gorm:"index"` // internal creator read time
	LastMessageAt      *time.Time      `json:"last_message_at,omitempty" gorm:"index"`
	FirstResponseAt    *time.Time      `json:"first_response_at,omitempty"`
	ResolvedAt         *time.Time      `json:"resolved_at,omitempty"`
	AssignedAt         *time.Time      `json:"assigned_at,omitempty"`
	ClosedAt           *time.Time      `json:"closed_at,omitempty"`

	CreatedAt *time.Time      `json:"created_at,omitempty"`
	UpdatedAt *time.Time      `json:"updated_at,omitempty"`
	DeletedAt *gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}
type DocumentTicket struct {
	TicketID   uuid.UUID `json:"ticket_id" gorm:"type:uuid;index"`
	DocumentID int       `json:"document_id" gorm:"index"`
}

type UserDocumentTicket struct {
	TicketID       uuid.UUID `json:"ticket_id" gorm:"type:uuid;index;uniqueIndex:udt_unique"`
	UserDocumentID int       `json:"user_document_id" gorm:"index;uniqueIndex:udt_unique"`
}

type TicketContent struct {
	ID         int            `json:"id" gorm:"primaryKey;autoIncrement"`
	Attachment datatypes.JSON `json:"attachment"`
	TicketID   *int           `json:"ticket_id"`
	UserID     *uuid.UUID     `json:"user_id"`
	User       *User          `json:"user"`
	AdminID    *uuid.UUID     `json:"admin_id"`
	Admin      *Admin         `json:"admin"`
	IsInternal bool           `json:"is_internal" gorm:"default:false;index"`
	Content    string         `json:"content"`

	CreatedAt *time.Time      `json:"created_at,omitempty"`
	UpdatedAt *time.Time      `json:"updated_at,omitempty"`
	DeletedAt *gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

const ( //status ticket
	TICKET_OPEN         = "open"
	TICKET_ASSIGNED     = "assigned"
	TICKET_INPROCESSING = "in_processing"
	TICKET_RESOLVED     = "resolved"
	TICKET_UNRESOLVED   = "unresolved"
	TICKET_CLOSED       = "closed"
)

const (
	TICKET_TYPE_GENERAL   = "general"
	TICKET_TYPE_PROCEDURE = "procedure"
	TICKET_TYPE_INTERNAL  = "internal"
)

const (
	TICKET_PRIORITY_LOW    = "low"
	TICKET_PRIORITY_MEDIUM = "medium"
	TICKET_PRIORITY_HIGH   = "high"
)
