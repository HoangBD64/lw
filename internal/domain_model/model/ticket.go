package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type Ticket struct {
	Model
	Category        string          `json:"category,omitempty"`
	SubCategory     string          `json:"sub_category,omitempty"`
	PreferSolution  string          `json:"prefer_solution,omitempty"`
	Status          string          `json:"status,omitempty" gorm:"index"`
	ResolutionType  string          `json:"resolution_type,omitempty"`
	TicketContents  []TicketContent `json:"ticket_contents,omitempty"`
	UserProcedureID *int            `json:"user_procedure_id" gorm:"index"`
	UserProcedure   *UserProcedure  `json:"user_procedure,omitempty"`
	ReadAt          *time.Time      `json:"read_at,omitempty" gorm:"index"`       // user read time
	AdminReadAt     *time.Time      `json:"admin_read_at,omitempty" gorm:"index"` // admin read time
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
	Model
	Attachment datatypes.JSON `json:"attachment"`
	TicketID   *uuid.UUID     `json:"ticket_id"`
	UserID     *uuid.UUID     `json:"user_id"`
	User       *User          `json:"user"`
	AdminID    *uuid.UUID     `json:"admin_id"`
	Admin      *Admin         `json:"admin"`
	Content    string         `json:"content"`
}

const ( //status ticket
	TICKET_OPEN         = "open"
	TICKET_INPROCESSING = "in_processing"
	TICKET_RESOLVED     = "resolved"
	TICKET_UNRESOLVED   = "unresolved"
)
