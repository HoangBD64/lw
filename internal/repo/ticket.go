package repo

import (
	"core/app"
	"core/internal/domain_model/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func CreateTicket(ticket model.Ticket) (model.Ticket, error) {
	tx := app.Database.DB.Model(&model.Ticket{}).Create(&ticket)
	return ticket, tx.Error
}

// Removed: order-based ticket queries (system uses documents only)

func GetTicketDocument(query Query) ([]model.Ticket, int64, error) {
	var tickets []model.Ticket
	var count int64

	tx := app.Database.DB.Model(&model.Ticket{}).
		Joins("JOIN user_document_tickets ON tickets.id = user_document_tickets.ticket_id").
		Where("user_document_tickets.user_document_id = ?", query.DocumentID)

	if query.TicketStatusRead != nil {
		if *query.TicketStatusRead {
			tx = tx.Where("tickets.read_at IS NOT NULL")
		} else {
			tx = tx.Where("tickets.read_at IS NULL")
		}
	}
	if query.TicketStatus != "" {
		tx = tx.Where("tickets.status = ?", query.TicketStatus)
	}
	if query.TicketCategory != "" {
		tx = tx.Where("tickets.category = ?", query.TicketCategory)
	}
	if query.TicketSubCategory != "" {
		tx = tx.Where("tickets.sub_category = ?", query.TicketSubCategory)
	}
	if query.TicketResolution != "" {
		tx = tx.Where("tickets.resolution_type = ?", query.TicketResolution)
	}

	if err := tx.Distinct().Count(&count).Error; err != nil {
		return nil, 0, err
	}

	err := tx.Distinct().
		Offset(query.Limit * (query.Page - 1)).
		Limit(query.Limit).
		Order("tickets.created_at DESC").
		Find(&tickets).Error

	return tickets, count, err
}

// Get tickets linked to any of the provided user_document_ids
func GetTicketsByUserDocumentIDs(ids []int, q Query) ([]model.Ticket, int64, error) {
	var tickets []model.Ticket
	var count int64

	if len(ids) == 0 {
		return []model.Ticket{}, 0, nil
	}

	tx := app.Database.DB.Model(&model.Ticket{}).
		Joins("JOIN user_document_tickets ON tickets.id = user_document_tickets.ticket_id").
		Where("user_document_tickets.user_document_id IN ?", ids)

	if q.TicketStatusRead != nil {
		if *q.TicketStatusRead {
			tx = tx.Where("tickets.read_at IS NOT NULL")
		} else {
			tx = tx.Where("tickets.read_at IS NULL")
		}
	}
	if q.TicketStatus != "" {
		tx = tx.Where("tickets.status = ?", q.TicketStatus)
	}
	if q.TicketCategory != "" {
		tx = tx.Where("tickets.category = ?", q.TicketCategory)
	}
	if q.TicketSubCategory != "" {
		tx = tx.Where("tickets.sub_category = ?", q.TicketSubCategory)
	}
	if q.TicketResolution != "" {
		tx = tx.Where("tickets.resolution_type = ?", q.TicketResolution)
	}

	if err := tx.Distinct().Count(&count).Error; err != nil {
		return nil, 0, err
	}

	err := tx.Distinct().
		Offset(q.Limit * (q.Page - 1)).
		Limit(q.Limit).
		Order("tickets.created_at DESC").
		Find(&tickets).Error

	return tickets, count, err
}

func GetTicket(query Query) ([]model.Ticket, int64, error) {
	var tickets []model.Ticket
	var count int64

	tx := app.Database.DB.Model(&model.Ticket{}).
		Preload("Documents", func(tx *gorm.DB) *gorm.DB {
			return tx.Select("id", "name", "file_type", "file_url", "size", "description")
		}).
		Preload("UserDocuments", func(tx *gorm.DB) *gorm.DB { return tx.Select("id", "document_id") })

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
	if query.TicketSubCategory != "" {
		tx = tx.Where("sub_category = ?", query.TicketSubCategory)
	}
	if query.TicketResolution != "" {
		tx = tx.Where("resolution_type = ?", query.TicketResolution)
	}
	if query.StartTime != nil && query.EndTime != nil {
		tx = tx.Where("created_at >= ? AND created_at <= ?", query.StartTime, query.EndTime)
	}

	if err := tx.Count(&count).Error; err != nil {
		return nil, 0, err
	}
	err := tx.Offset(query.Limit * (query.Page - 1)).Limit(query.Limit).Order("created_at DESC").Find(&tickets).Error

	return tickets, count, err
}

func GetTicketDetail(id uuid.UUID, user model.User) (model.Ticket, error) {
	var ticket model.Ticket
	tx := app.Database.DB.Model(&model.Ticket{}).Debug().Where("id = ?", id).
		Preload("Documents", func(tx *gorm.DB) *gorm.DB {
			return tx.Select("id", "name", "file_type", "file_url", "size", "description")
		}).
		Preload("UserDocuments", func(tx *gorm.DB) *gorm.DB { return tx.Select("id", "document_id") }).
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
func GetTicketByID(id uuid.UUID, user model.User) (model.Ticket, error) {
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

// Removed: order ID aggregation (documents only)

func CreateTicketContent(new model.TicketContent) (model.TicketContent, error) {
	tx := app.Database.DB.Model(&model.TicketContent{}).Create(&new)
	return new, tx.Error
}

func UpdateTicket(ticket model.Ticket) (model.Ticket, error) {
	tx := app.Database.DB.Model(&model.Ticket{}).Where("id = ?", ticket.ID).Select("status", "read_at", "admin_read_at").Updates(&ticket)
	return ticket, tx.Error
}

func AdminGetTicketByID(id uuid.UUID) (model.Ticket, error) {
	var ticket model.Ticket
	tx := app.Database.DB.Model(&model.Ticket{}).Where("id = ? ", id).First(&ticket)
	return ticket, tx.Error
}

func AdminGetTicketDetail(id uuid.UUID) (model.Ticket, error) {
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

	tx := app.Database.DB.Model(&model.Ticket{}).Preload("Orders", func(tx *gorm.DB) *gorm.DB {
		tx = tx.Select("id", "order_number", "is_have_ticket", "order_code")
		return tx
	}).Preload("Store", func(tx *gorm.DB) *gorm.DB {
		return tx.Select("id", "name", "status")
	})
	if query.TicketStatusRead != nil {
		if *query.TicketStatusRead {
			tx = tx.Where("tickets.admin_read_at IS NOT NULL")
		} else {
			tx = tx.Where("tickets.admin_read_at IS NULL")
		}
	}
	if query.TicketStatus != "" {
		tx = tx.Where("tickets.status = ?", query.TicketStatus)
	}
	if query.TicketCategory != "" {
		tx = tx.Where("tickets.category = ?", query.TicketCategory)
	}
	if query.TicketSubCategory != "" {
		tx = tx.Where("tickets.sub_category = ?", query.TicketSubCategory)
	}
	if query.TicketResolution != "" {
		tx = tx.Where("tickets.resolution_type = ?", query.TicketResolution)
	}
	if query.StartTime != nil && query.EndTime != nil {
		tx = tx.Where("tickets.created_at >= ? AND tickets.created_at <= ?", query.StartTime, query.EndTime)
	}
	if err := tx.Distinct().Count(&count).Error; err != nil {
		return nil, 0, err
	}

	err := tx.Distinct().
		Offset(query.Limit * (query.Page - 1)).
		Limit(query.Limit).
		Order("tickets.created_at DESC").
		Find(&tickets).Error

	return tickets, count, err

}
