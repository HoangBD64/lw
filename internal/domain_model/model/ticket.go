package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Ticket struct {
	ID              int             `json:"id" gorm:"primaryKey;autoIncrement"`
	Category        string          `json:"category,omitempty"`
	PreferSolution  string          `json:"prefer_solution,omitempty"`
	Status          string          `json:"status,omitempty" gorm:"index"`
	ResolutionType  string          `json:"resolution_type,omitempty"`
	TicketContents  []TicketContent `json:"ticket_contents,omitempty"`
	UserProcedureID *int            `json:"user_procedure_id" gorm:"index"`
	UserProcedure   *UserProcedure  `json:"user_procedure,omitempty"`
	ReadAt          *time.Time      `json:"read_at,omitempty" gorm:"index"`       // user read time
	AdminReadAt     *time.Time      `json:"admin_read_at,omitempty" gorm:"index"` // admin read time

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
	Content    string         `json:"content"`

	CreatedAt *time.Time      `json:"created_at,omitempty"`
	UpdatedAt *time.Time      `json:"updated_at,omitempty"`
	DeletedAt *gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

const ( //status ticket
	TICKET_OPEN         = "open"
	TICKET_INPROCESSING = "in_processing"
	TICKET_RESOLVED     = "resolved"
	TICKET_UNRESOLVED   = "unresolved"
)
