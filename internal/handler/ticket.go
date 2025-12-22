package handler

import (
	"core/internal/domain_model/model"
	"core/internal/repo"
	"core/utils/base"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

func CreateTicket(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(model.User)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again"})
	}

	type request struct {
		UserProcedureID int            `json:"user_procedure_id"`
		Category        string         `json:"category"`
		PreferSolution  string         `json:"prefer_solution"`
		Description     string         `json:"description"`
		Attachments     datatypes.JSON `json:"attachments"`
	}
	var input request

	if err := c.BodyParser(&input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Invalid request body"})
	}

	if input.UserProcedureID <= 0 {
		return c.JSON(base.BaseResponse{Status: false, Message: "user_procedure_id can't be empty"})
	}
	if _, err := repo.GetUserProcedureByID(input.UserProcedureID); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't find user_procedure for this user", Error: err})
	}
	new_ticket_content := model.TicketContent{
		Attachment: input.Attachments,
		UserID:     &user.ID,
		Content:    input.Description,
	}
	new_ticket := model.Ticket{
		Status:          model.TICKET_OPEN,
		Category:        input.Category,
		PreferSolution:  input.PreferSolution,
		TicketContents:  []model.TicketContent{new_ticket_content},
		UserProcedureID: &input.UserProcedureID,
	}
	_, err := repo.CreateTicket(new_ticket)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't CreateTicket", Error: err})
	}
	// Mark user procedure as having ticket
	_ = repo.MarkUserProcedureHasTicket(input.UserProcedureID, true)
	return c.JSON(base.BaseResponse{Status: true, Message: "Success"})
}

func GetTicketDetail(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(model.User)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again"})
	}
	var query repo.Query
	query.Parse(c)
	if query.TicketID == nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "id can't empty"})
	}
	ticket, err := repo.GetTicketDetail(*query.TicketID, user)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't GetTicketDetail", Error: err})
	}
	if ticket.ReadAt == nil {
		now := time.Now()
		ticket.ReadAt = &now
		repo.UpdateTicket(ticket)
	}

	for i := range ticket.TicketContents {
		ticket.TicketContents[i].AdminID = nil

		if ticket.TicketContents[i].Admin != nil {
			ticket.TicketContents[i].Admin.ID = uuid.Nil
		}
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: ticket})
}

func CreateTicketContent(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(model.User)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again"})
	}
	type request struct {
		TicketID    *int           `json:"ticket_id"`
		Description string         `json:"description"`
		Attachments datatypes.JSON `json:"attachments"`
	}
	var input request

	if err := c.BodyParser(&input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Invalid request body", Error: err})
	}
	if input.TicketID == nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "ticket_id can't empty"})
	}
	ticket, err := repo.GetTicketByID(*input.TicketID, user)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't GetTicketByID", Error: err})
	}
	if ticket.Status == model.TICKET_RESOLVED {
		return c.JSON(base.BaseResponse{Status: false, Message: "ticket is in resolved status"})
	}

	ticket.AdminReadAt = nil
	if _, err := repo.UpdateTicket(ticket); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "could not update ticket read time", Error: err})
	}

	new_ticket_content := model.TicketContent{
		Attachment: input.Attachments,
		TicketID:   &ticket.ID,
		UserID:     &user.ID,
		Content:    input.Description,
	}

	_, err = repo.CreateTicketContent(new_ticket_content)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't CreateTicketContent", Error: err})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success"})
}
func UpdateStatusTicket(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(model.User)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again"})
	}
	type request struct {
		TicketID *int   `json:"ticket_id"`
		Status   string `json:"status"`
	}
	var input request

	if err := c.BodyParser(&input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Invalid request body", Error: err})
	}
	if input.TicketID == nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "ticket_id can't empty"})
	}
	ticket, err := repo.GetTicketByID(*input.TicketID, user)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't GetTicketByID", Error: err})
	}
	if input.Status != model.TICKET_RESOLVED && input.Status != model.TICKET_UNRESOLVED {
		return c.JSON(base.BaseResponse{Status: false, Message: "can only change state to 'unresolved' or 'resolved'"})
	}

	if input.Status == ticket.Status {
		return c.JSON(base.BaseResponse{Status: false, Message: "ticket is already in this state"})
	} else {
		ticket.Status = input.Status
	}
	_, err = repo.UpdateTicket(ticket)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't UpdateStatusTicket", Error: err})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success"})
}

func GetTickets(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(model.User)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again"})
	}
	var query repo.Query
	query.Parse(c)
	query.UserID = &user.ID
	tickets, total, err := repo.GetTickets(query)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't GetTickets", Error: err})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: tickets, Total: &total})
}

func AdminCreateTicketContent(c *fiber.Ctx) error {
	admin, ok := c.Locals("admin").(model.Admin)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again"})
	}
	type request struct {
		TicketID    *int           `json:"ticket_id"`
		Description string         `json:"description"`
		Attachments datatypes.JSON `json:"attachments"`
	}
	var input request

	if err := c.BodyParser(&input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Invalid request body", Error: err})
	}
	if input.TicketID == nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "ticket_id can't empty"})
	}
	ticket, err := repo.AdminGetTicketByID(*input.TicketID)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't GetTicketByID", Error: err})
	}
	if ticket.Status == model.TICKET_RESOLVED {
		return c.JSON(base.BaseResponse{Status: false, Message: "ticket is in resolved status"})
	}

	ticket.ReadAt = nil
	if _, err := repo.UpdateTicket(ticket); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "could not update ticket read time", Error: err})
	}

	new_ticket_content := model.TicketContent{
		Attachment: input.Attachments,
		TicketID:   &ticket.ID,
		AdminID:    &admin.ID,
		Content:    input.Description,
	}

	_, err = repo.CreateTicketContent(new_ticket_content)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't CreateTicketContent", Error: err})
	}

	return c.JSON(base.BaseResponse{Status: true, Message: "Success"})
}

func AdminGetTicketDetail(c *fiber.Ctx) error {
	_, ok := c.Locals("admin").(model.Admin)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again"})
	}
	var query repo.Query
	query.Parse(c)
	if query.TicketID == nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "id can't empty"})
	}
	ticket, err := repo.AdminGetTicketDetail(*query.TicketID)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't GetTicketDetail", Error: err})
	}
	if ticket.AdminReadAt == nil {
		now := time.Now()
		ticket.AdminReadAt = &now
		repo.UpdateTicket(ticket)
	}

	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: ticket})
}

func AdminGetTicket(c *fiber.Ctx) error {
	var query repo.Query
	query.Parse(c)

	ticket, total, err := repo.AdminGetTickets(query)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't AdminGetTickets", Error: err})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: ticket, Total: &total})
}
