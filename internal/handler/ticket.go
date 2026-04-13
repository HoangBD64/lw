package handler

import (
	"context"
	"core/internal/domain_model/model"
	"core/internal/repo"
	"core/utils/base"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func normalizeTicketPriority(priority string) string {
	switch priority {
	case model.TICKET_PRIORITY_LOW, model.TICKET_PRIORITY_MEDIUM, model.TICKET_PRIORITY_HIGH:
		return priority
	default:
		return model.TICKET_PRIORITY_LOW
	}
}

func applyTicketMessageSLA(ticket *model.Ticket) {
	now := time.Now()
	ticket.LastMessageAt = &now
	ticket.ClosedAt = nil
	if ticket.Status == model.TICKET_RESOLVED || ticket.Status == model.TICKET_UNRESOLVED {
		ticket.ResolvedAt = nil
	}
}

func applyTicketStatusSLA(ticket *model.Ticket, status string) {
	now := time.Now()
	ticket.Status = status
	switch status {
	case model.TICKET_RESOLVED:
		ticket.ResolvedAt = &now
		ticket.ClosedAt = nil
	case model.TICKET_CLOSED:
		ticket.ClosedAt = &now
	default:
		ticket.ResolvedAt = nil
		if status != model.TICKET_CLOSED {
			ticket.ClosedAt = nil
		}
	}
}

func defaultTicketSubject(inputSubject string, up model.UserProcedure) string {
	if strings.TrimSpace(inputSubject) != "" {
		return strings.TrimSpace(inputSubject)
	}
	if strings.TrimSpace(up.TemplateName) != "" {
		return up.TemplateName
	}
	if strings.TrimSpace(up.ProcNumber) != "" {
		return up.ProcNumber
	}
	return fmt.Sprintf("Ticket for procedure %d", up.ID)
}

func createTicketNotification(ctx context.Context, title string, content string, ticket model.Ticket, userID *uuid.UUID, adminID *uuid.UUID) {
	metadata := model.NotiMetadata{Content: &content, UserProcedureID: ticket.UserProcedureID}
	_, err := repo.CreateNotification(ctx, model.Notification{
		UserID:   userID,
		AdminID:  adminID,
		Title:    title,
		Type:     model.NOTIFICATION_TICKET,
		Metadata: metadata,
	})
	if err != nil {
		return
	}
}

func appendTicketActivityForUser(user model.User, ticket model.Ticket, action string, description string, metadata any) {
	if ticket.UserProcedureID == nil {
		return
	}
	appendProcedureActivity(*ticket.UserProcedureID, userActivity(user, action, description, metadata))
}

func appendTicketActivityForAdmin(admin model.Admin, ticket model.Ticket, action string, description string, metadata any) {
	if ticket.UserProcedureID == nil {
		return
	}
	appendProcedureActivity(*ticket.UserProcedureID, adminActivity(admin, action, description, metadata))
}

func notifyTicketCreatedForProcedure(ctx context.Context, ticket model.Ticket) {
	content := fmt.Sprintf("Ticket %q was created for %s.", ticket.Subject, ticket.UserProcedure.ProcNumber)
	if ticket.UserProcedure != nil && ticket.UserProcedure.AdminID != nil {
		createTicketNotification(ctx, "Ticket created", content, ticket, nil, ticket.UserProcedure.AdminID)
	}
	if !ticket.IsInternal {
		createTicketNotification(ctx, "Ticket created", content, ticket, nil, nil)
	}
	if ticket.IsInternal && ticket.AdminID != nil {
		createTicketNotification(ctx, "Internal ticket created", content, ticket, nil, ticket.AdminID)
	}
}

func notifyTicketReply(ctx context.Context, ticket model.Ticket, replyByAdmin bool, senderAdminID *uuid.UUID, isInternal bool) {
	content := fmt.Sprintf("New message on ticket %q.", ticket.Subject)
	if replyByAdmin {
		if !isInternal && ticket.UserProcedure != nil && ticket.UserProcedure.UserID != nil {
			createTicketNotification(ctx, "Ticket replied", content, ticket, ticket.UserProcedure.UserID, nil)
			return
		}
		if ticket.IsInternal {
			if senderAdminID != nil && ticket.CreatorAdminID != nil && *ticket.CreatorAdminID == *senderAdminID {
				if ticket.AdminID != nil {
					createTicketNotification(ctx, "Internal ticket replied", content, ticket, nil, ticket.AdminID)
				}
				return
			}
			if ticket.CreatorAdminID != nil {
				createTicketNotification(ctx, "Internal ticket replied", content, ticket, nil, ticket.CreatorAdminID)
			}
		}
		return
	}
	if ticket.UserProcedure != nil && ticket.UserProcedure.AdminID != nil {
		createTicketNotification(ctx, "Ticket replied", content, ticket, nil, ticket.UserProcedure.AdminID)
	} else {
		createTicketNotification(ctx, "Ticket replied", content, ticket, nil, nil)
	}
}

func notifyTicketAssigned(ctx context.Context, ticket model.Ticket) {
	if ticket.AdminID == nil {
		return
	}
	content := fmt.Sprintf("Ticket %q has been assigned to you.", ticket.Subject)
	createTicketNotification(ctx, "Ticket assigned", content, ticket, nil, ticket.AdminID)
}

func notifyTicketStatusChanged(ctx context.Context, ticket model.Ticket) {
	content := fmt.Sprintf("Ticket %q status changed to %s.", ticket.Subject, ticket.Status)
	if !ticket.IsInternal && ticket.UserProcedure != nil && ticket.UserProcedure.UserID != nil {
		createTicketNotification(ctx, "Ticket updated", content, ticket, ticket.UserProcedure.UserID, nil)
	}
	if ticket.IsInternal && ticket.AdminID != nil {
		createTicketNotification(ctx, "Internal ticket updated", content, ticket, nil, ticket.AdminID)
	}
}

func adminCanAccessTicket(admin model.Admin, ticket model.Ticket) bool {
	if admin.IsAdmin() {
		return true
	}
	if !admin.IsLawyer() {
		return false
	}
	if ticket.IsInternal {
		if ticket.CreatorAdminID != nil && *ticket.CreatorAdminID == admin.ID {
			return true
		}
		return ticket.AdminID != nil && *ticket.AdminID == admin.ID
	}
	if ticket.AdminID != nil && *ticket.AdminID == admin.ID {
		return true
	}
	if ticket.UserProcedure == nil || ticket.UserProcedure.AdminID == nil {
		return false
	}
	return *ticket.UserProcedure.AdminID == admin.ID
}

func CreateTicket(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(model.User)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again"})
	}

	type request struct {
		UserProcedureID *int           `json:"user_procedure_id"`
		Category        string         `json:"category"`
		PreferSolution  string         `json:"prefer_solution"`
		Subject         string         `json:"subject"`
		Description     string         `json:"description"`
		Attachments     datatypes.JSON `json:"attachments"`
	}
	var input request

	if err := c.BodyParser(&input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Invalid request body"})
	}

	var up model.UserProcedure
	var err error
	ticketType := model.TICKET_TYPE_GENERAL
	if input.UserProcedureID != nil {
		up, err = repo.GetUserProcedureByID(*input.UserProcedureID)
		if err != nil {
			return c.JSON(base.BaseResponse{Status: false, Message: "Can't find user_procedure for this user", Error: err})
		}
		if up.UserID == nil || *up.UserID != user.ID {
			return c.JSON(base.BaseResponse{Status: false, Message: "Unauthorized"})
		}
		_ = repo.MarkUserProcedureHasTicket(*input.UserProcedureID, true)
		ticketType = model.TICKET_TYPE_PROCEDURE
	}
	new_ticket_content := model.TicketContent{
		Attachment: input.Attachments,
		UserID:     &user.ID,
		IsInternal: false,
		Content:    input.Description,
	}
	now := time.Now()
	new_ticket := model.Ticket{
		Status:          model.TICKET_OPEN,
		Category:        input.Category,
		PreferSolution:  input.PreferSolution,
		Type:            ticketType,
		Priority:        model.TICKET_PRIORITY_MEDIUM,
		Subject:         defaultTicketSubject(input.Subject, up),
		IsInternal:      false,
		TicketContents:  []model.TicketContent{new_ticket_content},
		UserProcedureID: input.UserProcedureID,
		UserID:          &user.ID,
		LastMessageAt:   &now,
	}

	if up.AdminID != nil && *up.AdminID != uuid.Nil {
		new_ticket.Status = model.TICKET_ASSIGNED
		new_ticket.AssignedAt = &now
		new_ticket.AdminID = up.AdminID
	}
	createdTicket, err := repo.CreateTicket(new_ticket)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't CreateTicket", Error: err})
	}
	createdTicket.UserProcedure = &up
	appendTicketActivityForUser(user, createdTicket, "ticket.created", fmt.Sprintf("Created ticket %s", createdTicket.Subject), map[string]any{"ticket_id": createdTicket.ID, "type": createdTicket.Type, "priority": createdTicket.Priority})
	notifyTicketCreatedForProcedure(c.UserContext(), createdTicket)
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
		_, _ = repo.UpdateTicket(ticket)
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
	if user.ID == uuid.Nil || ticket.UserID == nil || *ticket.UserID != user.ID {
		return c.JSON(base.BaseResponse{Status: false, Message: "Unauthorized"})
	}
	if ticket.Status == model.TICKET_RESOLVED {
		return c.JSON(base.BaseResponse{Status: false, Message: "ticket is in resolved status"})
	}

	ticket.AdminReadAt = nil
	applyTicketMessageSLA(&ticket)
	if _, err := repo.UpdateTicket(ticket); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "could not update ticket read time", Error: err})
	}

	new_ticket_content := model.TicketContent{
		Attachment: input.Attachments,
		TicketID:   &ticket.ID,
		UserID:     &user.ID,
		IsInternal: false,
		Content:    input.Description,
	}

	_, err = repo.CreateTicketContent(new_ticket_content)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't CreateTicketContent", Error: err})
	}
	appendTicketActivityForUser(user, ticket, "ticket.replied", fmt.Sprintf("Replied to ticket %s", ticket.Subject), map[string]any{"ticket_id": ticket.ID, "is_internal": false})
	notifyTicketReply(c.UserContext(), ticket, false, nil, false)
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
	}
	applyTicketStatusSLA(&ticket, input.Status)
	_, err = repo.UpdateTicket(ticket)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't UpdateStatusTicket", Error: err})
	}
	appendTicketActivityForUser(user, ticket, "ticket.status_changed", fmt.Sprintf("Changed ticket %s status to %s", ticket.Subject, ticket.Status), map[string]any{"ticket_id": ticket.ID, "status": ticket.Status})
	if ticket.UserProcedure != nil && ticket.UserProcedure.AdminID != nil {
		content := fmt.Sprintf("Ticket %q status changed to %s.", ticket.Subject, ticket.Status)
		createTicketNotification(c.UserContext(), "Ticket updated", content, ticket, nil, ticket.UserProcedure.AdminID)
	} else {
		content := fmt.Sprintf("Ticket %q status changed to %s.", ticket.Subject, ticket.Status)
		createTicketNotification(c.UserContext(), "Ticket updated", content, ticket, nil, nil)
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
		IsInternal  bool           `json:"is_internal"`
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
	if !adminCanAccessTicket(admin, ticket) {
		return c.JSON(base.BaseResponse{Status: false, Message: "Unauthorized"})
	}
	if ticket.Status == model.TICKET_RESOLVED {
		return c.JSON(base.BaseResponse{Status: false, Message: "ticket is in resolved status"})
	}

	if ticket.IsInternal || input.IsInternal {
		if ticket.CreatorAdminID != nil && *ticket.CreatorAdminID == admin.ID {
			ticket.AdminReadAt = nil
		} else {
			ticket.CreatorAdminReadAt = nil
		}
	} else {
		ticket.ReadAt = nil
		if ticket.FirstResponseAt == nil {
			now := time.Now()
			ticket.FirstResponseAt = &now
		}
	}
	applyTicketMessageSLA(&ticket)
	if _, err := repo.UpdateTicket(ticket); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "could not update ticket read time", Error: err})
	}

	new_ticket_content := model.TicketContent{
		Attachment: input.Attachments,
		TicketID:   &ticket.ID,
		AdminID:    &admin.ID,
		IsInternal: ticket.IsInternal || input.IsInternal,
		Content:    input.Description,
	}

	_, err = repo.CreateTicketContent(new_ticket_content)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't CreateTicketContent", Error: err})
	}
	appendTicketActivityForAdmin(admin, ticket, "ticket.replied", fmt.Sprintf("Replied to ticket %s", ticket.Subject), map[string]any{"ticket_id": ticket.ID, "is_internal": new_ticket_content.IsInternal})
	notifyTicketReply(c.UserContext(), ticket, true, &admin.ID, new_ticket_content.IsInternal)

	return c.JSON(base.BaseResponse{Status: true, Message: "Success"})
}

func AdminCreateInternalTicket(c *fiber.Ctx) error {
	admin, ok := c.Locals("admin").(model.Admin)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again"})
	}
	type request struct {
		UserProcedureID int            `json:"user_procedure_id"`
		AdminID         *uuid.UUID     `json:"admin_id"`
		Category        string         `json:"category"`
		PreferSolution  string         `json:"prefer_solution"`
		Subject         string         `json:"subject"`
		Priority        string         `json:"priority"`
		Description     string         `json:"description"`
		Attachments     datatypes.JSON `json:"attachments"`
	}
	var input request

	if err := c.BodyParser(&input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Invalid request body", Error: err})
	}
	if input.UserProcedureID <= 0 {
		return c.JSON(base.BaseResponse{Status: false, Message: "user_procedure_id can't be empty"})
	}
	if input.AdminID == nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "admin_id can't be empty"})
	}
	up, err := repo.AdminGetUserProcedureByID(input.UserProcedureID)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't find user procedure", Error: err})
	}
	if !admin.IsAdmin() && (up.AdminID == nil || *up.AdminID != admin.ID) {
		return c.JSON(base.BaseResponse{Status: false, Message: "Unauthorized"})
	}
	now := time.Now()
	newTicket := model.Ticket{
		Status:             model.TICKET_ASSIGNED,
		Category:           input.Category,
		PreferSolution:     input.PreferSolution,
		Type:               model.TICKET_TYPE_INTERNAL,
		Priority:           normalizeTicketPriority(input.Priority),
		Subject:            defaultTicketSubject(input.Subject, up),
		IsInternal:         true,
		UserProcedureID:    &input.UserProcedureID,
		AdminID:            input.AdminID,
		CreatorAdminID:     &admin.ID,
		AssignedAt:         &now,
		LastMessageAt:      &now,
		CreatorAdminReadAt: &now,
		TicketContents: []model.TicketContent{{
			Attachment: input.Attachments,
			AdminID:    &admin.ID,
			IsInternal: true,
			Content:    input.Description,
		}},
	}
	createdTicket, err := repo.CreateTicket(newTicket)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't CreateTicket", Error: err})
	}
	createdTicket.UserProcedure = &up
	_ = repo.MarkUserProcedureHasTicket(input.UserProcedureID, true)
	appendTicketActivityForAdmin(admin, createdTicket, "ticket.internal_created", fmt.Sprintf("Created internal ticket %s", createdTicket.Subject), map[string]any{"ticket_id": createdTicket.ID, "admin_id": input.AdminID, "priority": createdTicket.Priority})
	notifyTicketCreatedForProcedure(c.UserContext(), createdTicket)
	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: createdTicket})
}

func AdminUpdateTicket(c *fiber.Ctx) error {
	admin, ok := c.Locals("admin").(model.Admin)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again"})
	}
	type request struct {
		TicketID       *int       `json:"ticket_id"`
		AdminID        *uuid.UUID `json:"admin_id"`
		Status         string     `json:"status"`
		Priority       string     `json:"priority"`
		Subject        string     `json:"subject"`
		InternalNote   string     `json:"internal_note"`
	}
	var input request

	if err := c.BodyParser(&input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Invalid request body", Error: err})
	}
	if input.TicketID == nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "ticket_id can't be empty"})
	}
	ticket, err := repo.AdminGetTicketByID(*input.TicketID)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't GetTicketByID", Error: err})
	}
	if !adminCanAccessTicket(admin, ticket) {
		return c.JSON(base.BaseResponse{Status: false, Message: "Unauthorized"})
	}
	if input.AdminID != nil {
		if !admin.IsAdmin() {
			return c.JSON(base.BaseResponse{Status: false, Message: "Only admin can assign ticket"})
		}
		ticket.AdminID = input.AdminID
		now := time.Now()
		ticket.AssignedAt = &now
		ticket.AdminReadAt = nil
		if ticket.Status == model.TICKET_OPEN || ticket.Status == model.TICKET_INPROCESSING {
			ticket.Status = model.TICKET_ASSIGNED
		}
	}
	if input.Status != "" {
		switch input.Status {
		case model.TICKET_OPEN, model.TICKET_ASSIGNED, model.TICKET_INPROCESSING, model.TICKET_RESOLVED, model.TICKET_UNRESOLVED, model.TICKET_CLOSED:
			applyTicketStatusSLA(&ticket, input.Status)
		default:
			return c.JSON(base.BaseResponse{Status: false, Message: "Invalid ticket status"})
		}
	}
	if input.Priority != "" {
		ticket.Priority = normalizeTicketPriority(input.Priority)
	}
	if strings.TrimSpace(input.Subject) != "" {
		ticket.Subject = strings.TrimSpace(input.Subject)
	}
	if input.InternalNote != "" {
		ticket.InternalNote = input.InternalNote
	}
	if _, err := repo.UpdateTicket(ticket); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't UpdateTicket", Error: err})
	}
	activityAction := "ticket.updated"
	activityDescription := fmt.Sprintf("Updated ticket %s", ticket.Subject)
	if input.AdminID != nil {
		activityAction = "ticket.assigned"
		activityDescription = fmt.Sprintf("Assigned ticket %s", ticket.Subject)
	}
	if input.Status != "" {
		activityAction = "ticket.status_changed"
		activityDescription = fmt.Sprintf("Changed ticket %s status to %s", ticket.Subject, ticket.Status)
	}
	appendTicketActivityForAdmin(admin, ticket, activityAction, activityDescription, map[string]any{"ticket_id": ticket.ID, "status": ticket.Status, "priority": ticket.Priority, "admin_id": ticket.AdminID})
	if input.AdminID != nil {
		notifyTicketAssigned(c.UserContext(), ticket)
	}
	if input.Status != "" {
		notifyTicketStatusChanged(c.UserContext(), ticket)
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: ticket})
}

func AdminGetTicketDetail(c *fiber.Ctx) error {
	admin, ok := c.Locals("admin").(model.Admin)
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(base.BaseResponse{Status: false, Message: "Ticket not found"})
		}
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't GetTicketDetail", Error: err})
	}
	if !adminCanAccessTicket(admin, ticket) {
		return c.JSON(base.BaseResponse{Status: false, Message: "Unauthorized"})
	}
	if ticket.IsInternal && ticket.CreatorAdminID != nil && *ticket.CreatorAdminID == admin.ID {
		if ticket.CreatorAdminReadAt == nil {
			now := time.Now()
			ticket.CreatorAdminReadAt = &now
			_, _ = repo.UpdateTicket(ticket)
		}
	} else if ticket.AdminReadAt == nil {
		now := time.Now()
		ticket.AdminReadAt = &now
		_, _ = repo.UpdateTicket(ticket)
	}

	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: ticket})
}

func AdminGetTicket(c *fiber.Ctx) error {
	admin, ok := c.Locals("admin").(model.Admin)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again"})
	}
	var query repo.Query
	query.Parse(c)
	if admin.IsLawyer() {
		query.AdminID = &admin.ID
	}

	ticket, total, err := repo.AdminGetTickets(query)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't AdminGetTickets", Error: err})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: ticket, Total: &total})
}
