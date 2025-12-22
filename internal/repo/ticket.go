package repo

import (
	"core/app"
	"core/internal/domain_model/model"

	"gorm.io/gorm"
)

func CreateTicket(ticket model.Ticket) (model.Ticket, error) {
	tx := app.Database.DB.Model(&model.Ticket{}).Create(&ticket)
	return ticket, tx.Error
}

func GetTickets(query Query) ([]model.Ticket, int64, error) {
	var tickets []model.Ticket
	var count int64

	tx := app.Database.DB.Model(&model.Ticket{})
	sub := app.Database.DB.Model(&model.UserProcedure{}).Select("id").Where("user_id = ?", *query.UserID)
	tx = tx.Where("user_procedure_id IN (?)", sub)
	if query.UserProcedureID != nil {
		tx = tx.Where("user_procedure_id = ?", *query.UserProcedureID)
	}

	if query.TicketStatusRead != nil {
		if *query.TicketStatusRead {
			tx = tx.Where("read_at IS NOT NULL")
		} else {
			tx = tx.Where("read_at IS NULL")
		}
	}
	if query.TicketStatus != "" {
		tx = tx.Where("status = ?", query.TicketStatus)
	}
	if query.TicketCategory != "" {
		tx = tx.Where("category = ?", query.TicketCategory)
	}
	if query.TicketResolution != "" {
		tx = tx.Where("resolution_type = ?", query.TicketResolution)
	}
	if query.StartTime != nil && query.EndTime != nil {
		tx = tx.Where("created_at >= ? AND created_at <= ?", query.StartTime, query.EndTime)
	}

	tx = tx.Count(&count)
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	tx = tx.Scopes(pagination(query.Page, query.Limit)).
		Order("created_at DESC").
		Find(&tickets)

	return tickets, count, tx.Error
}

func GetTicketDetail(id int, user model.User) (model.Ticket, error) {
	var ticket model.Ticket
	tx := app.Database.DB.Model(&model.Ticket{}).Debug().Where("id = ?", id).
		Preload("TicketContents", func(tx *gorm.DB) *gorm.DB {
			tx = tx.Preload("User", func(tx *gorm.DB) *gorm.DB {
				tx = tx.Select("id", "name", "email", "avatar")
				return tx
			})
			tx = tx.Preload("Admin", func(tx *gorm.DB) *gorm.DB {
				tx = tx.Select("id", "name", "email", "avatar")
				return tx
			})
			return tx
		}).First(&ticket)
	return ticket, tx.Error
}
func GetTicketByID(id int, user model.User) (model.Ticket, error) {
	var ticket model.Ticket
	tx := app.Database.DB.Model(&model.Ticket{}).Where("id = ?", id).First(&ticket)
	return ticket, tx.Error
}

func CreateDocumentTicket(document_ticket []model.DocumentTicket) ([]model.DocumentTicket, error) {
	tx := app.Database.DB.Model(&model.DocumentTicket{}).Create(&document_ticket)
	return document_ticket, tx.Error
}

func CreateUserDocumentTicket(items []model.UserDocumentTicket) ([]model.UserDocumentTicket, error) {
	tx := app.Database.DB.Model(&model.UserDocumentTicket{}).Create(&items)
	return items, tx.Error
}

func CreateTicketContent(new model.TicketContent) (model.TicketContent, error) {
	tx := app.Database.DB.Model(&model.TicketContent{}).Create(&new)
	return new, tx.Error
}

func UpdateTicket(ticket model.Ticket) (model.Ticket, error) {
	tx := app.Database.DB.Model(&model.Ticket{}).Where("id = ?", ticket.ID).Select("status", "read_at", "admin_read_at").Updates(&ticket)
	return ticket, tx.Error
}

func AdminGetTicketByID(id int) (model.Ticket, error) {
	var ticket model.Ticket
	tx := app.Database.DB.Model(&model.Ticket{}).Where("id = ? ", id).First(&ticket)
	return ticket, tx.Error
}

func AdminGetTicketDetail(id int) (model.Ticket, error) {
	var ticket model.Ticket
	tx := app.Database.DB.Model(&model.Ticket{}).Debug().Where("id = ? ", id).Preload("TicketContents", func(tx *gorm.DB) *gorm.DB {
		tx = tx.Preload("User", func(tx *gorm.DB) *gorm.DB {
			tx = tx.Select("id", "name", "email", "avatar")
			return tx
		})
		tx = tx.Preload("Admin", func(tx *gorm.DB) *gorm.DB {
			tx = tx.Select("id", "name", "email", "avatar")
			return tx
		})
		return tx
	}).First(&ticket)
	return ticket, tx.Error
}

func AdminGetTickets(query Query) ([]model.Ticket, int64, error) {
	var tickets []model.Ticket
	var count int64

	tx := app.Database.DB.Debug().Model(&model.Ticket{})

	if query.UserID != nil {
		sub := app.Database.DB.Model(&model.UserProcedure{}).Select("id").Where("user_id = ?", *query.UserID)
		tx = tx.Where("user_procedure_id IN (?)", sub)
	}
	if query.UserProcedureID != nil {
		tx = tx.Where("user_procedure_id = ?", *query.UserProcedureID)
	}
	if query.TicketStatusRead != nil {
		if *query.TicketStatusRead {
			tx = tx.Where("admin_read_at IS NOT NULL")
		} else {
			tx = tx.Where("admin_read_at IS NULL")
		}
	}
	if query.TicketStatus != "" {
		tx = tx.Where("status = ?", query.TicketStatus)
	}
	if query.TicketCategory != "" {
		tx = tx.Where("category = ?", query.TicketCategory)
	}
	if query.TicketSubCategory != "" {
		tx = tx.Where("sub_category = ?", query.TicketSubCategory)
	}
	if query.TicketResolution != "" {
		tx = tx.Where("resolution_type = ?", query.TicketResolution)
	}
	if query.StartTime != nil && query.EndTime != nil {
		tx = tx.Where("created_at >= ? AND created_at <= ?", query.StartTime, query.EndTime)
	}

	tx = tx.Count(&count)
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	tx = tx.Scopes(pagination(query.Page, query.Limit)).
		Order("created_at DESC").
		Find(&tickets)

	return tickets, count, tx.Error

}
