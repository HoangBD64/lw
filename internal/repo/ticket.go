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

	tx := app.Database.DB.Model(&model.Ticket{}).
		Where("user_id = ?", *query.UserID).
		Where("is_internal = ?", false).
		Preload("UserProcedure", func(tx *gorm.DB) *gorm.DB {
			tx = tx.Select("id", "proc_number", "template_name", "company_name")
			return tx
		})
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
	if query.Type != "" {
		tx = tx.Where("type = ?", query.Type)
	}
	if query.Search != "" {
		like := "%" + query.Search + "%"
		tx = tx.Where("category ILIKE ? OR subject ILIKE ? OR proc_number ILIKE ? OR template_name ILIKE ? OR company_name ILIKE ?", like, like, like, like, like)
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
	tx := app.Database.DB.Debug().Model(&model.Ticket{}).
		Where("id = ?", id).
		Where("user_id = ?", user.ID).
		Where("is_internal = ?", false).
		Preload("UserProcedure", func(tx *gorm.DB) *gorm.DB {
			tx = tx.Select("id", "user_id", "procedure_id", "admin_id", "proc_number", "template_name", "company_name", "status")
			return tx
		})
	tx = tx.
		Preload("TicketContents", func(tx *gorm.DB) *gorm.DB {
			tx = tx.Where("is_internal = ?", false)
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
	tx := app.Database.DB.Model(&model.Ticket{}).
		Where("id = ?", id).
		Where("user_id = ?", user.ID).
		Where("is_internal = ?", false).
		Preload("UserProcedure", func(tx *gorm.DB) *gorm.DB {
			tx = tx.Select("id", "user_id", "procedure_id", "admin_id", "proc_number", "template_name", "company_name", "status")
			return tx
		}).
		First(&ticket)
	return ticket, tx.Error
}

func CreateTicketContent(new model.TicketContent) (model.TicketContent, error) {
	tx := app.Database.DB.Model(&model.TicketContent{}).Create(&new)
	return new, tx.Error
}

func UpdateTicket(ticket model.Ticket) (model.Ticket, error) {
	tx := app.Database.DB.Model(&model.Ticket{}).Where("id = ?", ticket.ID).Updates(&ticket)
	return ticket, tx.Error
}

func AdminGetTicketByID(id int) (model.Ticket, error) {
	var ticket model.Ticket
	tx := app.Database.DB.Model(&model.Ticket{}).Preload("UserProcedure", func(tx *gorm.DB) *gorm.DB {
		tx = tx.Select("id", "user_id", "procedure_id", "admin_id", "proc_number", "template_name", "company_name", "status")
		return tx
	}).Where("id = ? ", id).First(&ticket)
	return ticket, tx.Error
}

func AdminGetTicketDetail(id int) (model.Ticket, error) {
	var ticket model.Ticket
	tx := app.Database.DB.Model(&model.Ticket{}).Where("id = ? ", id).Preload("UserProcedure", func(tx *gorm.DB) *gorm.DB {
		tx = tx.Select("id", "user_id", "procedure_id", "admin_id", "proc_number", "template_name", "company_name", "status")
		return tx
	})
	tx = tx.Preload("TicketContents", func(tx *gorm.DB) *gorm.DB {
		tx = tx.Preload("User", func(tx *gorm.DB) *gorm.DB {
			tx = tx.Select("id", "name", "email", "avatar")
			return tx
		})
		tx = tx.Preload("Admin", func(tx *gorm.DB) *gorm.DB {
			tx = tx.Select("id", "name", "email", "avatar", "role", "status")
			return tx
		})
		return tx
	}).First(&ticket)
	return ticket, tx.Error
}

func AdminGetTickets(query Query) ([]model.Ticket, int64, error) {
	var tickets []model.Ticket
	var count int64

	tx := app.Database.DB.Model(&model.Ticket{}).
		Preload("UserProcedure", func(tx *gorm.DB) *gorm.DB {
			tx = tx.Select("id", "user_id", "procedure_id", "admin_id", "proc_number", "template_name", "company_name", "status")
			return tx
		})
	if query.AdminID != nil {
		tx = tx.Where("admin_id = ? OR creator_admin_id = ?", *query.AdminID, *query.AdminID)
	}

	if query.UserID != nil {
		tx = tx.Where("user_id = ? ", *query.UserID)
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
	if query.TicketResolution != "" {
		tx = tx.Where("resolution_type = ?", query.TicketResolution)
	}
	if query.Type != "" {
		tx = tx.Where("type = ?", query.Type)
	}
	if query.Search != "" {
		tx = tx.Where("subject ILIKE ? OR ticket_code ILIKE ?", "%"+query.Search+"%", "%"+query.Search+"%")
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
