package handler

import (
	"core/internal/domain_model/model"
	"core/internal/repo"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

func CreateTicket(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(model.User)
	if !ok {
		return c.JSON(fiber.Map{"status": false, "error": "Please login again"})
	}

	type request struct {
		UserProcedureID int            `json:"user_procedure_id"`
		Category        string         `json:"category"`
		SubCategory     string         `json:"sub_category"`
		PreferSolution  string         `json:"prefer_solution"`
		Description     string         `json:"description"`
		Attachments     datatypes.JSON `json:"attachments"`
	}
	var input request

	if err := c.BodyParser(&input); err != nil {
		return c.JSON(fiber.Map{"status": false, "error": "Invalid request body", "err": err.Error()})
	}

	if input.UserProcedureID <= 0 {
		return c.JSON(fiber.Map{"status": false, "error": "user_procedure_id can't be empty"})
	}
	if _, err := repo.GetUserProcedureByID(user.ID, input.UserProcedureID); err != nil {
		return c.JSON(fiber.Map{"status": false, "message": "Can't find user_procedure for this user", "error": err.Error()})
	}
	new_ticket_content := model.TicketContent{
		Attachment: input.Attachments,
		UserID:     &user.ID,
		Content:    input.Description,
	}
	new_ticket := model.Ticket{
		Status:          model.TICKET_OPEN,
		Category:        input.Category,
		SubCategory:     input.SubCategory,
		PreferSolution:  input.PreferSolution,
		TicketContents:  []model.TicketContent{new_ticket_content},
		UserProcedureID: &input.UserProcedureID,
	}
	_, err1 := repo.CreateTicket(new_ticket)
	if err1 != nil {
		return c.JSON(fiber.Map{"status": false, "message": "Can't CreateTicket", "error": err1.Error()})
	}
	return c.JSON(fiber.Map{"status": true, "message": "Success"})
}

func GetTicketDetail(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(model.User)
	if !ok {
		return c.JSON(fiber.Map{"status": false, "error": "Please login again"})
	}
	var query repo.Query
	query.Parse(c)
	if query.ID == nil {
		return c.JSON(fiber.Map{"status": false, "message": "id can't empty"})
	}
	ticket, err := repo.GetTicketDetail(*query.ID, user)
	if err != nil {
		return c.JSON(fiber.Map{"status": false, "message": "Can't GetTicketDetail", "error": err.Error()})
	}
	if ticket.ReadAt == nil {
		now := time.Now()
		ticket.ReadAt = &now
		repo.UpdateTicket(ticket)
	}

	// hide admin_id
	for i := range ticket.TicketContents {
		ticket.TicketContents[i].AdminID = nil

		if ticket.TicketContents[i].Admin != nil {
			ticket.TicketContents[i].Admin.ID = uuid.Nil
		}
	}
	return c.JSON(fiber.Map{"status": true, "message": "Success", "ticket": ticket})
}

func CreateTicketContent(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(model.User)
	if !ok {
		return c.JSON(fiber.Map{"status": false, "error": "Please login again"})
	}
	type request struct {
		TicketID    *uuid.UUID     `json:"ticket_id"`
		Description string         `json:"description"`
		Attachments datatypes.JSON `json:"attachments"`
	}
	var input request

	if err := c.BodyParser(&input); err != nil {
		return c.JSON(fiber.Map{"status": false, "error": "Invalid request body", "err": err.Error()})
	}
	if input.TicketID == nil {
		return c.JSON(fiber.Map{"status": false, "error": "ticket_id can't empty"})
	}
	ticket, err := repo.GetTicketByID(*input.TicketID, user)
	if err != nil {
		return c.JSON(fiber.Map{"status": false, "message": "Can't GetTicketByID", "error": err.Error()})
	}
	if ticket.Status == model.TICKET_RESOLVED {
		return c.JSON(fiber.Map{"status": false, "message": "ticket is in resolved status"})
	}

	ticket.AdminReadAt = nil
	if _, err := repo.UpdateTicket(ticket); err != nil {
		return c.JSON(fiber.Map{"status": false, "message": "could not update ticket read time", "error": err.Error()})
	}

	new_ticket_content := model.TicketContent{
		Attachment: input.Attachments,
		TicketID:   &ticket.ID,
		UserID:     &user.ID,
		Content:    input.Description,
	}

	_, err1 := repo.CreateTicketContent(new_ticket_content)
	if err1 != nil {
		return c.JSON(fiber.Map{"status": false, "message": "Can't CreateTicketContent", "error": err1.Error()})
	}
	return c.JSON(fiber.Map{"status": true, "message": "Success"})
}
func UpdateStatusTicket(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(model.User)
	if !ok {
		return c.JSON(fiber.Map{"status": false, "error": "Please login again"})
	}
	type request struct {
		ID     *uuid.UUID `json:"id"`
		Status string     `json:"status"`
	}
	var input request

	if err := c.BodyParser(&input); err != nil {
		return c.JSON(fiber.Map{"status": false, "error": "Invalid request body", "err": err.Error()})
	}
	if input.ID == nil {
		return c.JSON(fiber.Map{"status": false, "error": "id can't empty"})
	}
	ticket, err := repo.GetTicketByID(*input.ID, user)
	if err != nil {
		return c.JSON(fiber.Map{"status": false, "message": "Can't GetTicketByID", "error": err.Error()})
	}
	if input.Status != model.TICKET_RESOLVED && input.Status != model.TICKET_UNRESOLVED {
		return c.JSON(fiber.Map{"status": false, "message": "can only change state to 'unresolved' or 'resolved'"})
	}

	if input.Status == ticket.Status {
		return c.JSON(fiber.Map{"status": false, "message": "ticket is already in this state"})
	} else {
		ticket.Status = input.Status
	}
	_, err1 := repo.UpdateTicket(ticket)
	if err1 != nil {
		return c.JSON(fiber.Map{"status": false, "message": "Can't UpdateStatusTicket", "error": err1.Error()})
	}
	return c.JSON(fiber.Map{"status": true, "message": "Success"})
}

func GetTicketProcedure(c *fiber.Ctx) error {
	var query repo.Query
	query.Parse(c)
	if query.UserProcedureID == nil {
		return c.JSON(fiber.Map{"status": false, "message": "user_procedure_id can't be empty"})
	}
	ticket, total, err := repo.GetTicketDocument(query)
	if err != nil {
		return c.JSON(fiber.Map{"status": false, "message": "Can't GetTickets", "error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": true, "message": "Success", "tickets": ticket, "total": total})
}
func GetTickets(c *fiber.Ctx) error {
	var query repo.Query
	query.Parse(c)
	ticket, total, err := repo.GetTicket(query)
	if err != nil {
		return c.JSON(fiber.Map{"status": false, "message": "Can't GetTickets", "error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": true, "message": "Success", "tickets": ticket, "total": total})
}

// List tickets by multiple user_document_ids
func GetTicketsByUserDocuments(c *fiber.Ctx) error {
	type req struct {
		UserProcedureID *int `json:"user_procedure_id"`
	}
	var body req
	if err := c.BodyParser(&body); err != nil {
		return c.JSON(fiber.Map{"status": false, "error": "Invalid request body", "err": err.Error()})
	}
	var q repo.Query
	q.Parse(c)
	if body.UserProcedureID != nil {
		q.UserProcedureID = body.UserProcedureID
	}
	tickets, total, err := repo.GetTicketDocument(q)
	if err != nil {
		return c.JSON(fiber.Map{"status": false, "message": "Can't GetTicketsByUserDocumentIDs", "error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": true, "message": "Success", "tickets": tickets, "total": total})
}

func AdminCreateTicketContent(c *fiber.Ctx) error {
	admin, ok := c.Locals("admin").(model.Admin)
	if !ok {
		return c.JSON(fiber.Map{"status": false, "error": "Please login again"})
	}
	type request struct {
		TicketID    *uuid.UUID     `json:"ticket_id"`
		Description string         `json:"description"`
		Attachments datatypes.JSON `json:"attachments"`
	}
	var input request

	if err := c.BodyParser(&input); err != nil {
		return c.JSON(fiber.Map{"status": false, "error": "Invalid request body", "err": err.Error()})
	}
	if input.TicketID == nil {
		return c.JSON(fiber.Map{"status": false, "error": "ticket_id can't empty"})
	}
	ticket, err := repo.AdminGetTicketByID(*input.TicketID)
	if err != nil {
		return c.JSON(fiber.Map{"status": false, "message": "Can't GetTicketByID", "error": err.Error()})
	}
	if ticket.Status == model.TICKET_RESOLVED {
		return c.JSON(fiber.Map{"status": false, "message": "ticket is in resolved status"})
	}

	ticket.ReadAt = nil
	if _, err := repo.UpdateTicket(ticket); err != nil {
		return c.JSON(fiber.Map{"status": false, "message": "could not update ticket read time", "error": err.Error()})
	}

	new_ticket_content := model.TicketContent{
		Attachment: input.Attachments,
		TicketID:   &ticket.ID,
		AdminID:    &admin.ID,
		Content:    input.Description,
	}

	_, err1 := repo.CreateTicketContent(new_ticket_content)
	if err1 != nil {
		return c.JSON(fiber.Map{"status": false, "message": "Can't CreateTicketContent", "error": err1.Error()})
	}

	return c.JSON(fiber.Map{"status": true, "message": "Success"})
}

func AdminGetTicketDetail(c *fiber.Ctx) error {
	_, ok := c.Locals("admin").(model.Admin)
	if !ok {
		return c.JSON(fiber.Map{"status": false, "error": "Please login again"})
	}
	var query repo.Query
	query.Parse(c)
	if query.ID == nil {
		return c.JSON(fiber.Map{"status": false, "message": "id can't empty"})
	}
	ticket, err := repo.AdminGetTicketDetail(*query.ID)
	if err != nil {
		return c.JSON(fiber.Map{"status": false, "message": "Can't GetTicketDetail", "error": err.Error()})
	}
	if ticket.AdminReadAt == nil {
		now := time.Now()
		ticket.AdminReadAt = &now
		repo.UpdateTicket(ticket)
	}

	return c.JSON(fiber.Map{"status": true, "message": "Success", "ticket": ticket})
}

func AdminGetTicket(c *fiber.Ctx) error {
	var query repo.Query
	c.QueryParser(&query)

	ticket, total, err := repo.AdminGetTickets(query)
	if err != nil {
		return c.JSON(fiber.Map{"status": false, "message": "Can't AdminGetTickets", "error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": true, "message": "Success", "tickets": ticket, "total": total})
}
