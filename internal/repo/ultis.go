package repo

import (
	"core/internal/domain_model/model"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Query struct {
	UserID       *uuid.UUID `query:"user_id" json:"user_id,omitempty"`
	ID           *uuid.UUID `query:"id"  json:"id,omitempty"`
	IDS          []uuid.UUID
	UserIDs      []uuid.UUID
	AdminID      *uuid.UUID `query:"admin_id" json:"admin_id"`
	ProcedureID  *int       `query:"procedure_id" json:"procedure_id,omitempty"`
	PublicID     *uuid.UUID `query:"public_id" json:"public_id"`
	Role         string     `query:"role"  json:"role,omitempty"`
	Limit        int        `query:"limit" json:"limit,omitempty"`
	Page         int        `query:"page" json:"page,omitempty"`
	Name         string     `query:"name" json:"name,omitempty"`
	Email        string     `query:"email" json:"email,omitempty"`
	Search       string     `query:"search" json:"search,omitempty"`
	IsHaveTicket *bool      `query:"is_have_ticket" json:"is_have_ticket,omitempty"`
	IsAssigned   *bool      `query:"is_assigned" json:"is_assigned,omitempty"`
	//
	StartTime *time.Time `query:"start_time" json:"start_time,omitempty"`
	EndTime   *time.Time `query:"end_time" json:"end_time,omitempty"`
	Status    string     `query:"status" json:"status,omitempty"`
	Type      string     `query:"type" json:"type,omitempty"`
	Sort      string     `query:"sort" json:"sort,omitempty"`
	Filter    string     `query:"filter" json:"filter"`
	Hidden    *bool      `query:"hidden"`

	// Ticket/document filters
	DocumentID        *int   `query:"document_id" json:"document_id,omitempty"`
	UserProcedureID   *int   `query:"user_procedure_id" json:"user_procedure_id,omitempty"`
	TicketStatusRead  *bool  `query:"ticket_status_read" json:"ticket_status_read,omitempty"`
	TicketStatus      string `query:"ticket_status" json:"ticket_status,omitempty"`
	TicketCategory    string `query:"ticket_category" json:"ticket_category,omitempty"`
	TicketSubCategory string `query:"ticket_sub_category" json:"ticket_sub_category,omitempty"`
	TicketResolution  string `query:"ticket_resolution" json:"ticket_resolution,omitempty"`
}

func (query *Query) Parse(c *fiber.Ctx) {
	if err := c.QueryParser(query); err != nil {
		query.Limit = 10
		query.Page = 1
	}
	if query.Limit <= 0 {
		query.Limit = 10
	}
	if query.Page <= 0 {
		query.Page = 1
	}
}

func (q *Query) ParseForAdmin(c *fiber.Ctx, admin model.Admin) error {
	q.Parse(c)
	if admin.IsLawyer() {
		q.AdminID = &admin.ID
	}
	return nil
}

type SystemSettingMail struct {
	SENDER_MAIL string `json:"sender_mail"`
	ACCOUNT     string `json:"mail_account"`
	PASS        string `json:"mail_pass"`
	HeaderName  string `json:"header"`
	PORT        string `json:"port"`
	HOST        string `json:"host"`
}

// gorm pagination scope
func pagination(page, pageSize int) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB { return db.Offset(pageSize * (page - 1)).Limit(pageSize) }
}

// preload with select arguments
func populate(model string, fields ...string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if len(fields) > 0 {
			return db.Preload(model, func(tx *gorm.DB) *gorm.DB { return tx.Select(fields) })
		}
		return db.Preload(model)
	}
}
