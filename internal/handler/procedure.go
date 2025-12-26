package handler

import (
	"core/internal/common/templates"
	"core/internal/domain_model/domain"
	"core/internal/domain_model/model"
	"core/internal/notification"
	repo "core/internal/repo"
	"core/utils"
	"core/utils/base"
	"fmt"
	"log/slog"
	"slices"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/samber/lo"
)

func AdminCreateProcedure(c *fiber.Ctx) error {
	var input domain.ProcedureCreateRequest
	if err := c.BodyParser(&input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Review your input", Error: err})
	}
	if err := appUtils.Validate(c.Context(), &input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Validation failed", Error: err})
	}
	admin, ok := c.Locals("admin").(model.Admin)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again"})
	}
	totalCost := utils.MathRoundPrice(input.Cost)
	isActive := true
	p := model.Procedure{
		Name:        input.Name,
		Description: input.Description,
		AdminID:     &admin.ID,
		IsActive:    &isActive,
		DueDate:     input.DueDate,
		Cost:        &totalCost,
	}
	if input.IsActive != nil {
		p.IsActive = input.IsActive
	}
	if err := repo.AdminCreateProcedure(&p); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't Create Procedure", Error: err})
	}

	notification := model.Notification{
		Title:   "Procedure created",
		Type:    model.NOTIFICATION_PROCEDURE,
		Content: fmt.Sprintf("Procedure <strong>%s</strong> created by Admin. Please review the changes.", p.Name),
	}
	go repo.CreateNotification(c.Context(), notification)

	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: p})
}

func AdminUpdateProcedure(c *fiber.Ctx) error {
	var input domain.ProcedureUpdateRequest
	if err := c.BodyParser(&input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Review your input", Error: err})
	}
	id := c.Params("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Invalid id", Error: err})
	}
	p, err := repo.AdminGetProcedureByID(idInt)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Get Procedure Failed", Error: err})
	}
	if input.Name != "" {
		p.Name = input.Name
	}
	if input.Cost != nil {
		totalCost := utils.MathRoundPrice(*input.Cost)
		p.Cost = &totalCost
	}
	if input.IsActive != nil {
		p.IsActive = input.IsActive
	}
	p.Description = input.Description
	p.DueDate = input.DueDate
	if err := repo.AdminUpdateProcedure(&p); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't Update Procedure", Error: err})
	}

	notification := model.Notification{
		Title:   "Procedure updated",
		Type:    model.NOTIFICATION_PROCEDURE,
		Content: fmt.Sprintf("Procedure <strong>%s</strong> updated by Admin. Please review the changes.", p.Name),
	}
	go repo.CreateNotification(c.Context(), notification)
	return c.JSON(base.BaseResponse{Status: true, Message: "Success"})
}

func AdminGetProcedure(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Invalid id", Error: err})
	}
	p, err := repo.AdminGetProcedureByID(id)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Get Procedure Failed", Error: err})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: p})
}

func AdminGetProcedures(c *fiber.Ctx) error {
	var q repo.Query
	q.Parse(c)
	items, total, err := repo.AdminGetProcedures(q)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Get Procedures Failed", Error: err})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: fiber.Map{"items": items, "total": total}})
}

func AdminListUserProcedures(c *fiber.Ctx) error {
	var q repo.Query
	q.Parse(c)
	if admin, ok := c.Locals("admin").(model.Admin); ok && admin.IsLawyer() {
		q.AdminID = &admin.ID
	}
	items, total, err := repo.AdminListUserProcedures(q)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Get user procedures failed", Error: err})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: items, Total: &total})
}

func AdminDeleteProcedure(c *fiber.Ctx) error {
	var input domain.ProcedureDeleteRequest
	if err := c.BodyParser(&input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Invalid id", Error: err})
	}
	if err := repo.AdminDeleteProcedureByIDs(input.IDs); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't Delete Procedure", Error: err})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success"})
}

func AdminBulkUpsertProcedure(c *fiber.Ctx) error {
	var input domain.AdminBulkUpsertProcedureRequest
	if err := c.BodyParser(&input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Review your input", Error: err})
	}
	if err := appUtils.Validate(c.Context(), &input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Validation failed", Error: err})
	}
	totalCost := utils.MathRoundPrice(*input.Procedure.Cost)
	proc := model.Procedure{
		ID:          lo.FromPtr(input.Procedure.ID),
		Name:        input.Procedure.Name,
		Description: input.Procedure.Description,
		DueDate:     input.Procedure.DueDate,
		Cost:        &totalCost,
		AdminID:     input.Procedure.AdminID,
	}
	if input.Procedure.IsActive != nil {
		proc.IsActive = input.Procedure.IsActive
	}
	var steps []model.ProcedureStep
	for _, s := range input.ProcedureSteps {
		step := model.ProcedureStep{
			ID:          lo.FromPtr(s.ID),
			Name:        s.Name,
			DueDate:     s.DueDate,
			AdminID:     s.AdminID,
			Notes:       s.Notes,
			Position:    s.Position,
			Data:        s.Data,
			Description: s.Description,
			Status:      model.PROCEDURE_STEP_TODO,
		}
		if s.IsActive != nil {
			step.IsActive = s.IsActive
		}
		steps = append(steps, step)
	}
	savedProc, savedSteps, err := repo.AdminBulkUpsertProcedureWithSteps(proc, steps)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Bulk upsert failed", Error: err})
	}
	if input.PruneMissing {
		keepIDs := lo.FilterMap(savedSteps, func(item model.ProcedureStep, _ int) (int, bool) {
			return item.ID, item.ID != 0
		})
		if err := repo.PruneMissingProcedureSteps(savedProc.ID, keepIDs); err != nil {
			return c.JSON(base.BaseResponse{Status: false, Message: "Prune missing steps failed", Error: err})
		}
	}

	if len(savedSteps) != 0 {
		savedProc.ProcedureSteps = savedSteps
	}

	notification := model.Notification{
		Title:   "Procedure updated",
		Type:    model.NOTIFICATION_PROCEDURE,
		Content: fmt.Sprintf("Procedure <strong>%s</strong> updated by Admin. Please review the changes.", savedProc.Name),
	}
	go repo.CreateNotification(c.Context(), notification)
	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: savedProc})
}

func AdminCreateProcedureStep(c *fiber.Ctx) error {
	var input domain.AdminCreateProcedureStepRequest
	if err := c.BodyParser(&input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Review your input", Error: err})
	}
	err := appUtils.Validate(c.Context(), &input)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Validation failed", Error: err})
	}
	isActive := true
	step := model.ProcedureStep{
		ProcedureID: input.ProcedureID,
		Name:        input.Name,
		Status:      model.PROCEDURE_STEP_TODO,
		DueDate:     input.DueDate,
		AdminID:     input.AdminID,
		Notes:       input.Notes,
		Position:    input.Position,
		Data:        input.Data,
		Description: input.Description,
		IsActive:    lo.ToPtr(isActive),
	}
	if input.IsActive != nil {
		step.IsActive = input.IsActive
	}
	res, err := repo.AdminCreateProcedureStep(step)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't CreateProcedureStep", Error: err})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: res})
}

func AdminUpdateProcedureStep(c *fiber.Ctx) error {
	var input domain.AdminUpdateProcedureStepRequest
	if err := c.BodyParser(&input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Review your input", Error: err})
	}
	id := c.Params("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Invalid id", Error: err})
	}
	step, err := repo.AdminGetProcedureStepByID(idInt)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "", Error: err})
	}

	step.ID = idInt
	step.Name = input.Name
	step.DueDate = input.DueDate
	step.AdminID = input.AdminID
	step.Notes = input.Notes
	step.Position = input.Position
	step.IsActive = input.IsActive
	step.Data = input.Data
	step.Description = input.Description
	if err := repo.AdminUpdateProcedureStep(&step); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't UpdateProcedureStep", Error: err})
	}

	notification := model.Notification{
		Title:   "Procedure steps updated",
		Type:    model.NOTIFICATION_PROCEDURE,
		Content: fmt.Sprintf("Procedure step <strong>%s</strong> updated by Admin. Please review the changes.", step.Name),
	}
	go repo.CreateNotification(c.Context(), notification)
	return c.JSON(base.BaseResponse{Status: true, Message: "Success"})
}

func AdminDeleteProcedureStep(c *fiber.Ctx) error {
	var q domain.AdminDeleteProcedureStepRequest
	if err := c.BodyParser(&q); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Review your input", Error: err})
	}
	err := appUtils.Validate(c.Context(), &q)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Validation failed", Error: err})
	}
	_, err = repo.AdminGetProcedureStepByID(q.ID)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Get Procedure Step Failed", Error: err})
	}

	if err := repo.AdminDeleteProcedureStep(q.ID); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't DeleteProcedureStep", Error: err})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success"})
}

func AdminAssignLawyerToProcedureStep(c *fiber.Ctx) error {
	admin, ok := c.Locals("admin").(model.Admin)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again"})
	}
	if !admin.IsAdmin() || !admin.IsOwner {
		return c.JSON(base.BaseResponse{Status: false, Message: "Not authorized to assign"})
	}

	idParam := c.Params("id")
	userProcedureID, err := strconv.Atoi(idParam)
	if err != nil || userProcedureID == 0 {
		return c.JSON(base.BaseResponse{Status: false, Message: "Invalid user procedure id", Error: err})
	}

	var input domain.AdminAssignLawyerToProcedureStepRequest
	if err := c.BodyParser(&input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Review your input", Error: err})
	}
	if err := appUtils.Validate(c.Context(), &input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Validation failed", Error: err})
	}

	adminIDs := make([]uuid.UUID, 0)
	stepIDs := make([]int, 0)
	assignMap := make(map[int]*uuid.UUID)
	for _, asn := range input.AssignLawyerToProcedureSteps {
		stepIDs = append(stepIDs, asn.UserProcedureStepID)
		adminIDs = append(adminIDs, *asn.AdminID)
		assignMap[asn.UserProcedureStepID] = asn.AdminID
	}

	if len(adminIDs) > 0 {
		admins, err := repo.GetAdminByIds(adminIDs)
		if err != nil {
			return c.JSON(base.BaseResponse{Status: false, Message: "Fetch admins failed", Error: err})
		}
		for _, a := range admins {
			if !a.IsLawyer() {
				return c.JSON(base.BaseResponse{Status: false, Message: "Assigned admin must be a lawyer"})
			}
		}
	}

	ups, err := repo.GetUserProcedureStepsByUserProcedureAndStepIDs(userProcedureID, stepIDs)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Fetch user procedure steps failed", Error: err})
	}
	if len(ups) != len(stepIDs) {
		return c.JSON(base.BaseResponse{Status: false, Message: "Some procedure steps not found in this user procedure"})
	}
	for i := range ups {
		if adminID, ok := assignMap[ups[i].ID]; ok {
			ups[i].AdminID = adminID
		}
	}

	updated, err := repo.BulkAssignLawyerAdminToUserProcedureSteps(ups)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Assign lawyer failed", Error: err})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: updated})
}

func AdminListProcedureSteps(c *fiber.Ctx) error {
	idInt := c.Params("id")
	id, err := strconv.Atoi(idInt)
	if err != nil || id == 0 {
		return c.JSON(base.BaseResponse{Status: false, Message: "Invalid id", Error: err})
	}
	steps, err := repo.AdminGetProcedureSteps(id)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't ListProcedureSteps", Error: err})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: steps})
}

func AdminGetDetailUserProcedures(c *fiber.Ctx) error {
	admin, ok := c.Locals("admin").(model.Admin)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again"})
	}
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil || id == 0 {
		return c.JSON(base.BaseResponse{Status: false, Message: "Invalid id", Error: err})
	}

	up, err := repo.AdminGetUserProcedureDetailByID(id)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Get user procedure failed", Error: err})
	}
	if admin.IsLawyer() && (up.AdminID == nil || *up.AdminID != admin.ID) {
		return c.JSON(base.BaseResponse{Status: false, Message: "Not authorized to view this procedure"})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: up})
}

func AdminUpdateUserProcedure(c *fiber.Ctx) error {
	admin, ok := c.Locals("admin").(model.Admin)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again"})
	}
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil || id == 0 {
		return c.JSON(base.BaseResponse{Status: false, Message: "Invalid id", Error: err})
	}
	var input domain.AdminUpdateUserProcedureRequest
	if err := c.BodyParser(&input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Review your input", Error: err})
	}
	err = appUtils.Validate(c.Context(), &input)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Validation failed", Error: err})
	}
	up, err := repo.AdminGetUserProcedureByID(id)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Get user procedure failed", Error: err})
	}
	if admin.IsLawyer() {
		if lo.FromPtr(up.AdminID) != admin.ID {
			return c.JSON(base.BaseResponse{Status: false, Message: "Not authorized to manage this procedure"})
		}
	}
	up.Status = &input.Status
	up.Note = &input.Note
	if err := repo.AdminUpdateUserProcedure(&up); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Update user procedure failed", Error: err})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success"})
}

func AdminAssignUserProcedure(c *fiber.Ctx) error {
	currentAdmin, ok := c.Locals("admin").(model.Admin)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again"})
	}
	if !currentAdmin.IsAdmin() && !currentAdmin.IsOwner {
		return c.JSON(base.BaseResponse{Status: false, Message: "Not authorized to assign"})
	}

	var input domain.AssignUserProcedureRequest
	if err := c.BodyParser(&input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Review your input", Error: err})
	}
	if err := appUtils.Validate(c.Context(), &input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Validation failed", Error: err})
	}
	if len(input.UserProcedureID) == 0 {
		return c.JSON(base.BaseResponse{Status: false, Message: "user_procedure_id is required"})
	}

	adminAssign, er := repo.GetAdminByID(c.Context(), lo.FromPtr(input.AdminID))
	if er != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Fetch admin failed", Error: er})
	}
	if !adminAssign.IsLawyer() {
		return c.JSON(base.BaseResponse{Status: false, Message: "Assigned admin must be a lawyer"})
	}

	affected, err := repo.AssignAdminToUserProcedures(input.UserProcedureID, adminAssign.ID)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Assign failed", Error: err})
	}

	ids := make([]string, len(input.UserProcedureID))
	for i, id := range input.UserProcedureID {
		ids[i] = strconv.Itoa(id)
	}
	notification := model.Notification{
		AdminID: &adminAssign.ID,
		Title:   "Procedure assigned",
		Type:    model.NOTIFICATION_PROCEDURE,
		Content: fmt.Sprintf("Procedures <strong>%s</strong> have been assigned to you by Admin.", strings.Join(ids, ", ")),
	}
	go repo.CreateNotification(c.Context(), notification)

	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Total: &affected})
}

func SendMailToUser(UserProcedureIDs []int, mailType string) error {
	service := notification.NewService(notification.SystemEmailSender{})
	for _, upID := range UserProcedureIDs {
		up, err := repo.GetUserProcedureByID(upID)
		if err != nil || up.UserID == nil {
			if err != nil {
				slog.Error("Failed to fetch user procedure", "user_procedure_id", upID, "error", err)
			}
			continue
		}
		user, err := repo.GetUserByID(*up.UserID)
		if err != nil {
			slog.Error("Failed to fetch user for notification", "user_id", *up.UserID, "error", err)
			continue
		}

		var subject, tmpl string
		switch mailType {
		case model.NOTI_PROCEDURE_ERROR:
			subject = fmt.Sprintf("Procedure Issue Detected: %s", up.TemplateName)
			tmpl = "mail_procedure_error.html"
		case model.NOTI_PROCEDURE_COMPLETED:
			subject = fmt.Sprintf("Procedure Completed: %s", up.TemplateName)
			tmpl = "mail_procedure_completed.html"
		case model.NOTI_INVOICE_COMPLETED:
			subject = fmt.Sprintf("Payment Confirmed: %s", up.TemplateName)
			tmpl = "mail_invoice_completed.html"
		default:
			slog.Error("Unknown mail type", "mail_type", mailType)
			continue
		}

		dueDays := 0
		if up.DueDate != nil {
			dueDays = *up.DueDate
		}
		cost := 0.0
		if up.Cost != nil {
			cost = *up.Cost
		}
		status := ""
		if up.Status != nil {
			status = *up.Status
		}
		data := map[string]any{
			"UserName":        user.Name,
			"ProcedureName":   up.TemplateName,
			"CompanyName":  lo.FromPtr(up.CompanyName),
			"Status":          status,
			"DueDays":         dueDays,
			"Cost":            fmt.Sprintf("$%.2f", cost),
			"UserProcedureID": up.ID,
		}

		body, err := templates.RenderMailTemplate(tmpl, data)
		if err != nil {
			slog.Error("Failed to render mail template", "template", tmpl, "error", err)
			continue
		}

		if err := service.CreateNotifyInApp(&model.Notification{UserID: &user.ID, Title: subject, Content: fmt.Sprintf("%s - %s", up.TemplateName, mailType), Type: model.NOTIFICATION_PROCEDURE}); err != nil {
			slog.Error("Failed to create in-app notification", "user_id", user.ID, "error", err)
		}

		if err := service.SendEmailIfEnabled(user, subject, body); err != nil {
			slog.Error("Failed to send email notification", "user_id", user.ID, "error", err)
		}
	}
	return nil
}

func AdminUnassignUserProcedure(c *fiber.Ctx) error {
	admin, ok := c.Locals("admin").(model.Admin)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again"})
	}
	if !admin.IsAdmin() && !admin.IsOwner {
		return c.JSON(base.BaseResponse{Status: false, Message: "Not authorized to unassign"})
	}
	var input domain.UnassignUserProcedureRequest
	if err := c.BodyParser(&input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Review your input", Error: err})
	}
	total, err := repo.UnassignAdminFromUserProcedure(input.UserProcedureID)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Unassign failed", Error: err})
	}

	ids := make([]string, len(input.UserProcedureID))
	for i, id := range input.UserProcedureID {
		ids[i] = strconv.Itoa(id)
	}

	notification := model.Notification{
		Title:   "Procedure unassigned",
		Type:    model.NOTIFICATION_PROCEDURE,
		Content: fmt.Sprintf("Procedures <strong>%s</strong> have been unassigned by Admin.", strings.Join(ids, ", ")),
	}
	go repo.CreateNotification(c.Context(), notification)

	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Total: &total})
}

func AdminUpdateUserProcedureStepStatus(c *fiber.Ctx) error {
	admin, ok := c.Locals("admin").(model.Admin)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again", Error: nil})
	}
	var input domain.AdminUpdateUserProcedureStepStatusRequest
	if err := c.BodyParser(&input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Review your input", Error: err})
	}
	if err := appUtils.Validate(c.Context(), &input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Validation failed", Error: err})
	}
	ups, err := repo.GetUserProcedureStepByID(input.UserProcedureStepID)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Get user procedure step failed", Error: err})
	}
	userProcedure, err := repo.GetUserProcedureByID(ups.UserProcedureID)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Get user procedure failed", Error: err})
	}
	if !admin.IsLawyer() {
		return c.JSON(base.BaseResponse{Status: false, Message: "Not authorized to manage this procedure"})
	} else if lo.FromPtr(userProcedure.AdminID) != admin.ID {
		return c.JSON(base.BaseResponse{Status: false, Message: "Not authorized to manage this procedure"})
	}
	ups.Status = input.Status
	if err := repo.UpdateUserProcedureStep(&ups); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Update user procedure step failed", Error: err})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success"})
}

func AdminUpdateUserProcedureStep(c *fiber.Ctx) error {
	admin, ok := c.Locals("admin").(model.Admin)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again", Error: nil})
	}
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil || id == 0 {
		return c.JSON(base.BaseResponse{Status: false, Message: "Invalid id", Error: err})
	}
	var input domain.AdminUserUpdateProcedureStepRequest
	if err := c.BodyParser(&input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Review your input", Error: err})
	}
	ups, err := repo.GetUserProcedureStepByID(id)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Get user procedure step failed", Error: err})
	}
	userProcedure, err := repo.GetUserProcedureByID(ups.UserProcedureID)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Get user procedure failed", Error: err})
	}

	if !admin.IsLawyer() {
		return c.JSON(base.BaseResponse{Status: false, Message: "Not authorized to manage this procedure"})
	} else if lo.FromPtr(userProcedure.AdminID) != admin.ID {
		return c.JSON(base.BaseResponse{Status: false, Message: "Not authorized to manage this procedure"})
	}
	ups.Notes = input.Notes
	ups.FeedBack = input.FeedBack
	if input.Status == model.USER_PROCEDURE_STEP_ERROR {
		UserUpdatedFalse := false
		ups.UserUpdated = &UserUpdatedFalse

		status := model.USER_PROCEDURE_ERROR
		userProcedure.Status = &status
		go repo.AdminUpdateUserProcedure(&userProcedure)
		go SendMailToUser([]int{ups.UserProcedureID}, model.NOTI_PROCEDURE_ERROR)
		notification := model.Notification{
			UserID:  userProcedure.UserID,
			Title:   "Procedure step error",
			Type:    model.NOTIFICATION_PROCEDURE,
			Content: fmt.Sprintf("Procedure step <strong>%s</strong> of procedure <strong>%d</strong> has an error. Please review the changes.", ups.Name, userProcedure.ID),
		}
		go repo.CreateNotification(c.Context(), notification)
	}
	ups.Status = input.Status
	ups.UserCompleted = input.UserCompleted

	if err := repo.UpdateUserProcedureStep(&ups); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Update user procedure step failed", Error: err})
	}

	if input.Status == model.USER_PROCEDURE_STEP_DONE {
		go func(userProcedureID int) {
			steps, err := repo.AdminGetAllStepStatusByUserProcedureID(userProcedureID)
			if err != nil {
				slog.Error("Failed to fetch user procedure steps", "user_procedure_id", userProcedureID, "error", err)
				return
			}
			allDone := true
			anyError := false

			for _, s := range steps {
				if s.Status != model.USER_PROCEDURE_STEP_DONE {
					allDone = false
				}
				if s.Status == model.USER_PROCEDURE_STEP_ERROR {
					anyError = true
				}
			}

			up, err := repo.GetUserProcedureByID(userProcedureID)
			if err != nil {
				slog.Error("Failed to fetch user procedure", "user_procedure_id", userProcedureID, "error", err)
				return
			}

			if allDone {
				status := model.USER_PROCEDURE_COMPLETED
				up.Status = &status
				go SendMailToUser([]int{userProcedureID}, model.NOTI_PROCEDURE_COMPLETED)
				notification := model.Notification{
					UserID:  up.UserID,
					Title:   "Procedure completed",
					Type:    model.NOTIFICATION_PROCEDURE,
					Content: fmt.Sprintf("Congratulations! Your procedure <strong>%s</strong> has been completed.", up.TemplateName),
				}
				go repo.CreateNotification(c.Context(), notification)
			} else if anyError {
				status := model.USER_PROCEDURE_ERROR
				up.Status = &status
			} else {
				status := model.USER_PROCEDURE_PROCESSING
				up.Status = &status
			}

			if err := repo.AdminUpdateUserProcedure(&up); err != nil {
				slog.Error("Failed to update user procedure status", "user_procedure_id", userProcedureID, "error", err)
			}
		}(ups.UserProcedureID)
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success"})
}

func UserGetProcedures(c *fiber.Ctx) error {
	var q repo.Query
	q.Parse(c)
	items, total, err := repo.UserGetProcedures(q)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Get Procedures Failed", Error: err})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: fiber.Map{"items": items, "total": total}})
}

func UserGetProcedureDetail(c *fiber.Ctx) error {
	id := c.Params("id")
	procedureID, err := strconv.Atoi(id)
	user, ok := c.Locals("user").(model.User)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again", Error: nil})
	}
	if err != nil || procedureID == 0 {
		return c.JSON(base.BaseResponse{Status: false, Message: "Invalid procedure ID", Error: err})
	}

	ups, err := repo.GetUserProcedureDetail(procedureID, user.ID)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Get failed", Error: err})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: ups})
}

func UserGetProcedure(c *fiber.Ctx) error {
	if _, ok := c.Locals("user").(model.User); !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again", Error: nil})
	}
	id := c.Params("id")
	procedureID, err := strconv.Atoi(id)
	if err != nil || procedureID == 0 {
		return c.JSON(base.BaseResponse{Status: false, Message: "Invalid procedure ID", Error: err})
	}
	proc, err := repo.GetProcedureByID(procedureID)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Get procedure failed", Error: err})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: proc})
}

func MappingProcedureStepsToUserProcedureSteps(procedureSteps []model.ProcedureStep, userProcedureID int) []model.UserProcedureStep {
	var userProcedureSteps []model.UserProcedureStep
	for _, step := range procedureSteps {
		userProcedureStep := model.UserProcedureStep{
			UserProcedureID: userProcedureID,
			ProcedureStepID: step.ID,
			Status:          model.USER_PROCEDURE_STEP_PENDING,
			Data:            step.Data,
			Name:            step.Name,
			DueDate:         step.DueDate,
			Notes:           step.Notes,
			Position:        step.Position,
			Description:     step.Description,
		}
		userProcedureSteps = append(userProcedureSteps, userProcedureStep)
	}
	return userProcedureSteps
}

func UserListUserProcedures(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(model.User)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again", Error: nil})
	}
	var q repo.Query
	q.Parse(c)
	items, total, err := repo.ListUserProceduresByUser(user.ID, q)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Get procedures failed", Error: err})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: items, Total: &total})
}

// NOTE: TEMPORARY
// func UserCreateProcedure(c *fiber.Ctx) error {
// 	user, ok := c.Locals("user").(model.User)
// 	if !ok {
// 		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again", Error: nil})
// 	}
// 	var input domain.UserCreateProcedureRequest
// 	if err := c.BodyParser(&input); err != nil {
// 		return c.JSON(base.BaseResponse{Status: false, Message: "Review your input", Error: err})
// 	}
// 	if input.ProcedureID <= 0 {
// 		return c.JSON(base.BaseResponse{Status: false, Message: "procedure_id is required"})
// 	}
// 	procedure, err := repo.GetProcedureByID(input.ProcedureID)
// 	if err != nil {
// 		return c.JSON(base.BaseResponse{Status: false, Message: "Get procedure failed", Error: err})
// 	}
// 	totalCost := lo.FromPtr(procedure.Cost)
// 	isHaveTicket := false
// 	status := model.USER_PROCEDURE_CREATED
// 	userProcedure := model.UserProcedure{
// 		UserID:       &user.ID,
// 		ProcedureID:  input.ProcedureID,
// 		Title:        &input.Title,
// 		Status:       &status,
// 		IsHaveTicket: &isHaveTicket,
// 		TemplateName: procedure.Name,
// 		DueDate:      procedure.DueDate,
// 		Cost:         &totalCost,
// 	}
// 	baseSteps, err := repo.UserGetProcedureSteps(input.ProcedureID)
// 	if err != nil {
// 		return c.JSON(base.BaseResponse{Status: false, Message: "Get procedure steps failed", Error: err})
// 	}
// 	userProcedure, err = repo.CreateUserProcedure(userProcedure)
// 	if err != nil {
// 		return c.JSON(base.BaseResponse{Status: false, Message: "Create user procedure failed", Error: err})
// 	}
// 	ups := MappingProcedureStepsToUserProcedureSteps(baseSteps, userProcedure.ID)
// 	for i := range ups {
// 		ups[i].UserProcedureID = userProcedure.ID
// 	}
// 	if err := repo.UserCreateBulkProcedureStep(&ups); err != nil {
// 		return c.JSON(base.BaseResponse{Status: false, Message: "Create user procedure steps failed", Error: err})
// 	}

// 	if len(input.ProcedureSteps) > 0 {
// 		idxByProcStep := make(map[int]int, len(ups))
// 		for i := range ups {
// 			idxByProcStep[ups[i].ProcedureStepID] = i
// 		}
// 		for _, ov := range input.ProcedureSteps {
// 			if i, ok := idxByProcStep[ov.ProcedureStepID]; ok {
// 				ups[i].Data = ov.Data
// 				if err := repo.UpdateUserProcedureStep(&ups[i]); err != nil {
// 					return c.JSON(base.BaseResponse{Status: false, Message: "Apply step override failed", Error: err})
// 				}
// 			}
// 		}
// 	}
// 	userProcedure.UserProcedureSteps = ups
// 	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: userProcedure})
// }

func UserUpdateUserProcedureStep(c *fiber.Ctx) error {
	_, ok := c.Locals("user").(model.User)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again", Error: nil})
	}
	id := c.Params("id")
	indInt, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Invalid ID"})
	}
	var input domain.UserUpdateUserProcedureStepRequest
	if err := c.BodyParser(&input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Review your input", Error: err})
	}
	ups, err := repo.GetUserProcedureStepByID(indInt)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Get user procedure step failed", Error: err})
	}

	allowStatus := []string{
		model.USER_PROCEDURE_STEP_PENDING,
		model.USER_PROCEDURE_STEP_ERROR,
	}
	if !slices.Contains(allowStatus, ups.Status) && !lo.FromPtr(input.UserCompleted) {
		return c.JSON(base.BaseResponse{Status: false, Message: "Status not allowed"})
	}
	up, err := repo.GetUserProcedureByID(ups.UserProcedureID)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Get user procedure failed", Error: err})
	}
	ups.Data = input.Data
	isCheckUserProcedureStatus := false
	if ups.Status == model.USER_PROCEDURE_STEP_ERROR {
		UserUpdatedFalse := false
		ups.UserUpdated = &UserUpdatedFalse
		ups.Status = model.USER_PROCEDURE_STEP_PROCESSING
		isCheckUserProcedureStatus = true
	}
	ups.UserCompleted = input.UserCompleted
	if err := repo.UpdateUserProcedureStep(&ups); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Update user procedure step failed", Error: err})
	}

	notification := model.Notification{
		AdminID: up.AdminID,
		Title:   "Review Required",
		Content: fmt.Sprintf("Step <strong>%s</strong> in procedure <strong>%d</strong> has been updated. Please review the changes.", ups.Name, up.ID),
		Type:    model.NOTIFICATION_PROCEDURE,
	}
	go repo.CreateNotification(c.Context(), notification)

	if isCheckUserProcedureStatus {
		userProcedure, err := repo.GetUserProcedureByID(ups.UserProcedureID)
		if err != nil {
			return c.JSON(base.BaseResponse{Status: false, Message: "Get user procedure failed", Error: err})
		}
		for _, step := range userProcedure.UserProcedureSteps {
			if step.ID == ups.ID {
				continue
			}
			if step.Status == model.USER_PROCEDURE_STEP_ERROR {
				return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: ups})
			}
		}
		status := model.USER_PROCEDURE_PROCESSING
		userProcedure.Status = &status
		go repo.AdminUpdateUserProcedure(&userProcedure)
	}

	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: ups})
}
