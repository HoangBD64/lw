package handler

import (
	"context"
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
	"time"

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
		ExtraFeeID:  input.ExtraFeeID,
	}
	if input.IsActive != nil {
		p.IsActive = input.IsActive
	}
	if err := repo.AdminCreateProcedure(&p); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't Create Procedure", Error: err})
	}

	metadata := model.NotiMetadata{
		Content: lo.ToPtr(fmt.Sprintf("Procedure template <strong>%s</strong> created by Admin. Please review the changes.", p.Name)),
	}

	notification := model.Notification{
		Title:    "Procedure created",
		Type:     model.NOTIFICATION_PROCEDURE,
		Metadata: metadata,
	}

	utils.RunAsync(3*time.Second, func(ctx context.Context) error {
		_, err := repo.CreateNotification(ctx, notification)
		return err
	})
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
	if input.ExtraFeeID != nil {
		p.ExtraFeeID = input.ExtraFeeID
	}
	if input.Description != nil {
		p.Description = input.Description
	}
	if input.DueDate != nil {
		p.DueDate = input.DueDate
	}
	if err := repo.AdminUpdateProcedure(&p); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't Update Procedure", Error: err})
	}

	metadata := model.NotiMetadata{
		Content: lo.ToPtr(fmt.Sprintf("Procedure template <strong>%s</strong> updated by Admin. Please review the changes.", p.Name)),
	}

	notification := model.Notification{
		Title:    "Procedure updated",
		Type:     model.NOTIFICATION_PROCEDURE,
		Metadata: metadata,
	}

	utils.RunAsync(3*time.Second, func(ctx context.Context) error {
		_, err := repo.CreateNotification(ctx, notification)
		return err
	})
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
		ExtraFeeID:  input.Procedure.ExtraFeeID,
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

	metadata := model.NotiMetadata{
		Content: lo.ToPtr(fmt.Sprintf("Procedure template <strong>%s</strong> updated by Admin. Please review the changes.", savedProc.Name)),
	}
	notification := model.Notification{
		Title:    "Procedure updated",
		Type:     model.NOTIFICATION_PROCEDURE,
		Metadata: metadata,
	}

	utils.RunAsync(3*time.Second, func(ctx context.Context) error {
		_, err := repo.CreateNotification(ctx, notification)
		return err
	})
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
	proc, err := repo.AdminGetProcedureByID(step.ProcedureID)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Get Procedure Failed", Error: err})
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

	metadata := model.NotiMetadata{
		UserProcedureID: nil,
		Content:         lo.ToPtr(fmt.Sprintf("Step <strong>%s</strong> in Procedure template <strong>%s</strong> has been updated.", step.Name, proc.Name)),
	}
	notification := model.Notification{
		Title:    "Procedure steps updated",
		Type:     model.NOTIFICATION_PROCEDURE,
		Metadata: metadata,
	}

	utils.RunAsync(3*time.Second, func(ctx context.Context) error {
		_, err := repo.CreateNotification(ctx, notification)
		return err
	})
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

	ups, err := repo.AdminGetUserProcedureStepsByUserProcedureAndStepIDs(userProcedureID, stepIDs)
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

	updated, err := repo.AdminBulkAssignToUserProcedureSteps(ups)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Assign lawyer failed", Error: err})
	}
	for _, step := range ups {
		appendProcedureActivity(
			step.UserProcedureID,
			adminActivity(
				admin, "user_procedure_step.assigned",
				fmt.Sprintf("Assigned lawyer to step %s", step.Name),
				js{"user_procedure_step_id": step.ID, "assigned_admin_id": step.AdminID},
			),
		)
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

func AdminGetUserProcedureActivityLogs(c *fiber.Ctx) error {
	admin, ok := c.Locals("admin").(model.Admin)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again"})
	}
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil || id == 0 {
		return c.JSON(base.BaseResponse{Status: false, Message: "Invalid id", Error: err})
	}

	up, err := repo.AdminGetUserProcedureByID(id)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Get user procedure failed", Error: err})
	}
	if admin.IsLawyer() && lo.FromPtr(up.AdminID) != admin.ID {
		return c.JSON(base.BaseResponse{Status: false, Message: "Not authorized to view this procedure"})
	}

	logs, err := repo.GetUserProcedureActivityLogs(id)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Get activity logs failed", Error: err})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: logs})
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
	if input.Priority != "" {
		allowPriority := []string{
			model.USER_PROCEDURE_PRIORITY_LOW,
			model.USER_PROCEDURE_PRIORITY_MEDIUM,
			model.USER_PROCEDURE_PRIORITY_HIGH,
		}
		if !slices.Contains(allowPriority, input.Priority) {
			return c.JSON(base.BaseResponse{Status: false, Message: "Invalid priority"})
		}
		up.Priority = &input.Priority
	}
	fmt.Println("============>2", input.Status)
	switch input.Status {
	case "":
		// do nothing
	case model.USER_PROCEDURE_SYSTEM_SUBMIT_REVIEW:
		fmt.Println("============>1")
		up.SystemStatus = &input.Status
	default:
		up.Status = &input.Status
	}
	up.Note = &input.Note
	if input.Status == model.USER_PROCEDURE_CANCELLED {
		// for cancelled procedure, update all pending steps to cancelled
		steps, err := repo.AdminGetAllStepStatusByUserProcedureID(up.ID)
		if err != nil {
			return c.JSON(base.BaseResponse{Status: false, Message: "Get user procedure steps failed", Error: err})
		}
		for i := range steps {
			steps[i].Status = model.USER_PROCEDURE_STEP_CANCELLED
		}
		if _, err := repo.AdminBulkUpdateUserProcedureSteps(steps); err != nil {
			return c.JSON(base.BaseResponse{Status: false, Message: "Update user procedure steps failed", Error: err})
		}

		utils.RunAsync(3*time.Second, func(ctx context.Context) error {
			return SendMailToUser([]int{up.ID}, model.NOTI_PROCEDURE_CANCELLED)
		})
	}
	if err := repo.AdminUpdateUserProcedure(&up); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Update user procedure failed", Error: err})
	}
	appendProcedureActivity(
		up.ID,
		adminActivity(
			admin, "user_procedure.updated",
			fmt.Sprintf("Updated procedure %s", strNonEmpty(up.ProcNumber, strconv.Itoa(up.ID))),
			js{"status": input.Status, "priority": input.Priority, "note": input.Note},
		),
	)
	if input.Status == model.USER_PROCEDURE_SYSTEM_SUBMIT_REVIEW && admin.IsLawyer() {
		utils.RunAsync(3*time.Second, func(ctx context.Context) error {
			return notification.SendSubmitReviewNoticeToAdmins(up.ID, admin.Name)
		})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success"})
}

func AdminMarkReviewUserProcedure(c *fiber.Ctx) error {
	admin, ok := c.Locals("admin").(model.Admin)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again"})
	}
	if !admin.IsAdmin() {
		return c.JSON(base.BaseResponse{Status: false, Message: "Not authorized to assign"})
	}

	var input domain.MarkReviewUserProcedureRequest
	if err := c.BodyParser(&input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Review your input", Error: err})
	}
	if err := appUtils.Validate(c.Context(), &input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Validation failed", Error: err})
	}
	allowStatuses := []string{
		model.USER_PROCEDURE_SUBMITTED,
	}
	upList, err := repo.AdminGetUserProcedureByIDs(input.UserProcedureID)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Fetch user procedures failed", Error: err})
	}
	for _, up := range upList {
		if up.Status == nil || !slices.Contains(allowStatuses, lo.FromPtr(up.Status)) {
			return c.JSON(base.BaseResponse{Status: false, Message: fmt.Sprintf("Only procedures with status %s can be marked review", strings.Join(allowStatuses, ", "))})
		}
	}
	if err := repo.AdminMarkReviewUserProcedures(upList); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Update user procedure failed", Error: err})
	}
	for _, up := range upList {
		appendProcedureActivity(
			up.ID,
			adminActivity(
				admin, "user_procedure.mark_review",
				fmt.Sprintf("Marked procedure %s for review", strNonEmpty(up.ProcNumber, strconv.Itoa(up.ID))),
				js{"status": model.USER_PROCEDURE_SYSTEM_SUBMIT_REVIEW},
			),
		)
	}

	return c.JSON(base.BaseResponse{Status: true, Message: "Success"})
}

func AdminUpdateUserProcedureStatus(c *fiber.Ctx) error {
	admin, ok := c.Locals("admin").(model.Admin)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again"})
	}

	var input domain.AdminUpdateUserProcedureStatusRequest
	if err := c.BodyParser(&input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Review your input", Error: err})
	}
	if err := appUtils.Validate(c.Context(), &input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Validation failed", Error: err})
	}
	upList, err := repo.AdminGetUserProcedureByIDs(input.UserProcedureID)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Fetch user procedures failed", Error: err})
	}

	switch input.Action {
	case "hold":
		allowStatuses := []string{
			model.USER_PROCEDURE_IN_PROGRESS,
		}
		for _, up := range upList {
			if up.Status == nil || !slices.Contains(allowStatuses, lo.FromPtr(up.Status)) {
				return c.JSON(base.BaseResponse{Status: false, Message: fmt.Sprintf("Only procedures with status %s can be hold", strings.Join(allowStatuses, ", "))})
			}
		}
		if err := repo.AdminHoldUserProcedures(upList); err != nil {
			return c.JSON(base.BaseResponse{Status: false, Message: "Update user procedure failed", Error: err})
		}
	case "resume":
		allowStatuses := []string{
			model.USER_PROCEDURE_HOLD,
		}
		for _, up := range upList {
			if up.Status == nil || !slices.Contains(allowStatuses, lo.FromPtr(up.Status)) {
				return c.JSON(base.BaseResponse{Status: false, Message: fmt.Sprintf("Only procedures with status %s can be resumed", strings.Join(allowStatuses, ", "))})
			}
		}
		if err := repo.AdminResumeUserProcedures(upList); err != nil {
			return c.JSON(base.BaseResponse{Status: false, Message: "Update user procedure failed", Error: err})
		}

	case "mark_review":
		if !admin.IsAdmin() {
			return c.JSON(base.BaseResponse{Status: false, Message: "Not authorized to assign"})
		}
		allowStatuses := []string{
			model.USER_PROCEDURE_SUBMITTED,
		}
		for _, up := range upList {
			if up.Status == nil || !slices.Contains(allowStatuses, lo.FromPtr(up.Status)) {
				return c.JSON(base.BaseResponse{Status: false, Message: fmt.Sprintf("Only procedures with status %s can be marked review", strings.Join(allowStatuses, ", "))})
			}
		}
		if err := repo.AdminMarkReviewUserProcedures(upList); err != nil {
			return c.JSON(base.BaseResponse{Status: false, Message: "Update user procedure failed", Error: err})
		}

	}
	for _, up := range upList {
		appendProcedureActivity(
			up.ID,
			adminActivity(
				admin, "user_procedure.status_action",
				fmt.Sprintf("Updated procedure %s status to %s", strNonEmpty(up.ProcNumber, strconv.Itoa(up.ID)), input.Action),
				js{"action": input.Action},
			),
		)
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

	allowAssignStatuses := []string{
		model.USER_PROCEDURE_SUBMITTED,
		model.USER_PROCEDURE_PENDING_REVIEW,
	}
	up, err := repo.AdminGetUserProcedureByIDs(input.UserProcedureID)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Fetch user procedures failed", Error: err})
	}
	for _, item := range up {
		if item.Status == nil || !slices.Contains(allowAssignStatuses, lo.FromPtr(item.Status)) {
			return c.JSON(base.BaseResponse{Status: false, Message: fmt.Sprintf("Only procedures with status %s can be assigned", strings.Join(allowAssignStatuses, ", "))})
		}
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

	notifications := make([]model.Notification, 0, len(input.UserProcedureID))
	itemNumber := ""
	for _, upItem := range up {
		if upItem.ProcNumber != "" {
			itemNumber = upItem.ProcNumber
		} else {
			itemNumber = strconv.Itoa(upItem.ID)
		}

		metadata := model.NotiMetadata{
			UserProcedureID: lo.ToPtr(upItem.ID),
			Content:         lo.ToPtr(fmt.Sprintf("Procedure <strong>%s</strong> has been assigned to you by Admin.", itemNumber)),
		}
		notification := model.Notification{
			AdminID:  &adminAssign.ID,
			Title:    "Procedure assigned",
			Type:     model.NOTIFICATION_PROCEDURE,
			Metadata: metadata,
		}
		notifications = append(notifications, notification)
	}

	utils.RunAsync(3*time.Second, func(ctx context.Context) error {
		_, err := repo.BulkCreateNotification(ctx, notifications)
		return err
	})
	utils.RunAsync(3*time.Second, func(ctx context.Context) error {
		var firstErr error
		for _, upID := range input.UserProcedureID {
			if err := notification.SendProcedureAssignedNoticeToLawyer(upID, adminAssign, currentAdmin.Name); err != nil {
				if firstErr == nil {
					firstErr = err
				}
			}
		}
		return firstErr
	})
	for _, upID := range input.UserProcedureID {
		appendProcedureActivity(
			upID,
			adminActivity(
				currentAdmin,
				"user_procedure.assigned",
				fmt.Sprintf("Assigned procedure to lawyer %s", adminAssign.Name),
				js{"assigned_admin_id": adminAssign.ID, "assigned_admin_name": adminAssign.Name},
			),
		)
	}
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

		var subject, tmpl, contentNoti string
		switch mailType {
		case model.NOTI_PROCEDURE_ERROR:
			subject = fmt.Sprintf("Procedure Issue Detected: %s", up.TemplateName)
			tmpl = templates.MailTemplateProcedureError
			contentNoti = fmt.Sprintf("Procedure <strong>%s</strong> has an error. Please review the changes.", up.ProcNumber)

		case model.NOTI_PROCEDURE_COMPLETED:
			subject = fmt.Sprintf("Procedure Completed: %s", up.TemplateName)
			tmpl = templates.MailTemplateProcedureCompleted
			contentNoti = fmt.Sprintf("Procedure <strong>%s</strong> has been completed.", up.ProcNumber)

		case model.NOTI_INVOICE_COMPLETED:
			subject = fmt.Sprintf("Payment Confirmed: %s", up.TemplateName)
			tmpl = templates.MailTemplateInvoiceCompleted
			contentNoti = fmt.Sprintf("Payment for Procedure <strong>%s</strong> has been confirmed.", up.ProcNumber)

		case model.NOTI_PROCEDURE_CANCELLED:
			subject = fmt.Sprintf("Procedure Cancelled: %s", up.TemplateName)
			tmpl = templates.MailTemplateProcedureCancelled
			contentNoti = fmt.Sprintf("Your procedure <strong>%s</strong> has been cancelled.", up.ProcNumber)

		case model.NOTI_USER_REGISTRATION_COMPLETED:
			subject = fmt.Sprintf("Registration Completed: %s", up.TemplateName)
			tmpl = templates.MailTemplateUserRegistrationCompleted
			contentNoti = fmt.Sprintf("Registration for procedure <strong>%s</strong> has been completed successfully.", up.ProcNumber)

		default:
			slog.Error("Unknown mail type", "mail_type", mailType)
			continue
		}

		dueDays := 0
		if up.DueDate != nil {
			dueDays = *up.DueDate
		}
		// get cost from transaction if exists, otherwise use cost from procedure template
		cost := 0.0
		tran, _ := repo.GetLatestTransactionByUserProcedureID(up.ID)
		if tran != nil && tran.AmountDeposit != 0 {
			cost = tran.AmountDeposit
		} else if up.Cost != nil {
			cost = *up.Cost
		}
		status := ""
		if up.Status != nil {
			status = *up.Status
		}
		websiteURL := strings.TrimSpace(utils.GetWebBaseURL())

		data := templates.DeadlineReminderData{
			UserName:            user.Name,
			ProcedureName:       up.TemplateName,
			CompanyName:         lo.FromPtr(up.CompanyName),
			Status:              status,
			DueDays:             dueDays,
			Cost:                fmt.Sprintf("$%.2f", cost),
			UserProcedureNumber: up.ProcNumber,
			URL:                 websiteURL,
		}

		body, err := templates.RenderMailTemplate(tmpl, data)
		if err != nil {
			slog.Error("Failed to render mail template", "template", tmpl, "error", err)
			continue
		}

		metadata := model.NotiMetadata{
			UserProcedureID: lo.ToPtr(up.ID),
			Content:         lo.ToPtr(contentNoti),
		}
		err = service.CreateNotifyInApp(&model.Notification{
			UserID:   &user.ID,
			Title:    subject,
			Metadata: metadata,
			Type:     model.NOTIFICATION_PROCEDURE})
		if err != nil {
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

	ids := make([]string, 0, len(input.UserProcedureID))
	for _, upID := range input.UserProcedureID {
		upItem, _ := repo.GetUserProcedureByID(upID)
		if upItem.ProcNumber != "" {
			ids = append(ids, upItem.ProcNumber)
		} else {
			ids = append(ids, strconv.Itoa(upID))
		}
	}

	metadata := model.NotiMetadata{
		Content: lo.ToPtr(fmt.Sprintf("Procedures %s have been unassigned by Admin.", strings.Join(ids, ", "))),
	}
	notification := model.Notification{
		Title:    "Procedure unassigned",
		Type:     model.NOTIFICATION_PROCEDURE,
		Metadata: metadata,
	}

	utils.RunAsync(3*time.Second, func(ctx context.Context) error {
		_, err := repo.CreateNotification(ctx, notification)
		return err
	})
	for _, upID := range input.UserProcedureID {
		appendProcedureActivity(
			upID,
			adminActivity(
				admin,
				"user_procedure.unassigned",
				"Unassigned lawyer from procedure",
				nil,
			),
		)
	}
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
	appendProcedureActivity(
		ups.UserProcedureID,
		adminActivity(
			admin,
			"user_procedure_step.status_updated",
			fmt.Sprintf("Updated step %s status to %s", ups.Name, input.Status),
			js{"user_procedure_step_id": ups.ID, "status": input.Status},
		),
	)
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

	if admin.IsLawyer() && lo.FromPtr(userProcedure.AdminID) != admin.ID {
		return c.JSON(base.BaseResponse{Status: false, Message: "Not authorized to manage this procedure"})
	}
	ups.Notes = input.Notes
	ups.FeedBack = input.FeedBack
	if input.Status == model.USER_PROCEDURE_STEP_INFORMATION_REQUIRED {
		UserUpdatedFalse := false
		ups.UserUpdated = &UserUpdatedFalse

		status := model.USER_PROCEDURE_INFORMATION_REQUIRED
		userProcedure.Status = &status
		userProcedure.SystemStatus = &status
		utils.RunAsync(3*time.Second, func(ctx context.Context) error {
			return repo.AdminUpdateUserProcedure(&userProcedure)
		})
		if ups.Status != model.USER_PROCEDURE_STEP_INFORMATION_REQUIRED {
			utils.RunAsync(3*time.Second, func(ctx context.Context) error {
				return SendMailToUser([]int{ups.UserProcedureID}, model.NOTI_PROCEDURE_ERROR)
			})
		}
	}
	ups.Status = input.Status
	// ups.UserCompleted = input.UserCompleted

	if err := repo.UpdateUserProcedureStep(&ups); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Update user procedure step failed", Error: err})
	}
	appendProcedureActivity(
		ups.UserProcedureID,
		adminActivity(
			admin,
			"user_procedure_step.updated",
			fmt.Sprintf("Updated step %s", ups.Name),
			js{"user_procedure_step_id": ups.ID, "status": input.Status, "notes": input.Notes},
		),
	)

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
				if s.Status == model.USER_PROCEDURE_STEP_INFORMATION_REQUIRED {
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
				up.SystemStatus = &status
			} else if anyError {
				status := model.USER_PROCEDURE_INFORMATION_REQUIRED
				up.Status = &status
				up.SystemStatus = &status
			} else {
				status := model.USER_PROCEDURE_IN_PROGRESS
				up.Status = &status
				up.SystemStatus = &status
			}

			if err := repo.AdminUpdateUserProcedure(&up); err != nil {
				slog.Error("Failed to update user procedure status", "user_procedure_id", userProcedureID, "error", err)
			}

			if allDone {
				utils.RunAsync(3*time.Second, func(ctx context.Context) error {
					return SendMailToUser([]int{userProcedureID}, model.NOTI_PROCEDURE_COMPLETED)
				})
			}
		}(ups.UserProcedureID)
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success"})
}

func AdminListUserProcedureSteps(c *fiber.Ctx) error {
	admin, ok := c.Locals("admin").(model.Admin)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again"})
	}
	var query repo.Query
	query.Parse(c)
	if admin.IsLawyer() {
		query.AdminID = &admin.ID
	}
	steps, total, err := repo.AdminGetUserProcedureSteps(query)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't ListUserProcedureSteps", Error: err})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: steps, Total: &total})
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

	ups, err := repo.UserGetUserProcedureDetail(procedureID, user.ID)
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
	proc, err := repo.UserGetProcedureByID(procedureID)
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
			Status:          model.USER_PROCEDURE_STEP_CREATED,
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
	items, total, err := repo.UserGetUserProceduresByUser(user.ID, q)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Get procedures failed", Error: err})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: items, Total: &total})
}

func UserUpdateUserProcedureStep(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(model.User)
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
		model.USER_PROCEDURE_STEP_CREATED,
		model.USER_PROCEDURE_STEP_INFORMATION_REQUIRED,
		model.USER_PROCEDURE_STEP_SUBMITTED,
		model.USER_PROCEDURE_STEP_IN_PROGRESS,
	}
	if !slices.Contains(allowStatus, ups.Status) {
		return c.JSON(base.BaseResponse{Status: false, Message: "Status not allowed"})
	}
	up, err := repo.GetUserProcedureByID(ups.UserProcedureID)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Get user procedure failed", Error: err})
	}
	ups.Data = input.Data
	isCheckUserProcedureStatus := false

	// ưu tiên info required hơn submitted
	if ups.Status == model.USER_PROCEDURE_STEP_INFORMATION_REQUIRED {
		UserUpdatedFalse := false
		ups.UserUpdated = &UserUpdatedFalse
		ups.Status = model.USER_PROCEDURE_STEP_IN_PROGRESS
		isCheckUserProcedureStatus = true
	} else if input.Status == model.USER_PROCEDURE_STEP_SUBMITTED {
		ups.Status = input.Status
		status := model.USER_PROCEDURE_SUBMITTED
		up.Status = &status
		up.SystemStatus = &status
		if err := repo.AdminUpdateUserProcedure(&up); err != nil {
			return c.JSON(base.BaseResponse{Status: false, Message: "Update user procedure failed", Error: err})
		}
	}
	// ups.UserCompleted = input.UserCompleted
	if err := repo.UpdateUserProcedureStep(&ups); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Update user procedure step failed", Error: err})
	}
	appendProcedureActivity(
		ups.UserProcedureID,
		userActivity(
			user,
			"user_procedure_step.updated",
			fmt.Sprintf("User updated step %s", ups.Name),
			fiber.Map{"user_procedure_step_id": ups.ID},
		),
	)

	// TODO: AFTER CHECK MAIL
	if input.Status == model.USER_PROCEDURE_STEP_SUBMITTED {
		utils.RunAsync(3*time.Second, func(ctx context.Context) error {
			return SendMailToUser([]int{ups.UserProcedureID}, model.NOTI_USER_REGISTRATION_COMPLETED)
		})
	}

	if up.ProcNumber == "" {
		up.ProcNumber = fmt.Sprintf("LW%04d", up.ID)
	}

	metadata := model.NotiMetadata{
		UserProcedureID: lo.ToPtr(ups.UserProcedureID),
		Content:         lo.ToPtr(fmt.Sprintf("Step %s in procedure %s has been updated.", ups.Name, up.ProcNumber)),
	}
	notification := model.Notification{
		AdminID:  up.AdminID,
		Title:    "Review Required",
		Type:     model.NOTIFICATION_PROCEDURE,
		Metadata: metadata,
	}
	utils.RunAsync(3*time.Second, func(ctx context.Context) error {
		_, err := repo.CreateNotification(ctx, notification)
		return err
	})

	if isCheckUserProcedureStatus {
		userProcedure, err := repo.GetUserProcedureByID(ups.UserProcedureID)
		if err != nil {
			return c.JSON(base.BaseResponse{Status: false, Message: "Get user procedure failed", Error: err})
		}
		for _, step := range userProcedure.UserProcedureSteps {
			if step.ID == ups.ID {
				continue
			}
			if step.Status == model.USER_PROCEDURE_STEP_INFORMATION_REQUIRED {
				return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: ups})
			}
		}
		status := model.USER_PROCEDURE_IN_PROGRESS
		userProcedure.Status = &status
		userProcedure.SystemStatus = &status
		utils.RunAsync(3*time.Second, func(ctx context.Context) error {
			return repo.AdminUpdateUserProcedure(&userProcedure)
		})
	}

	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: ups})
}

func UserGetUserProcedureSteps(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(model.User)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again", Error: nil})
	}
	var query repo.Query
	query.Parse(c)
	ups, total, err := repo.UserGetUserProcedureSteps(query, user.ID)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Get user procedure steps failed", Error: err})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: ups, Total: &total})
}

func AdminMigrateStatus(c *fiber.Ctx) error {
	admin, ok := c.Locals("admin").(model.Admin)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again"})
	}

	// Only allow admin or owner to perform migration
	if !admin.IsAdmin() && !admin.IsOwner {
		return c.JSON(base.BaseResponse{Status: false, Message: "Not authorized to perform migration"})
	}

	var input domain.AdminMigrateStatusRequest
	if err := c.BodyParser(&input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Review your input", Error: err})
	}

	result, err := repo.AdminMigrateStatus(input.DryRun)
	if err != nil {
		slog.Error("Migration failed", "error", err, "admin_id", admin.ID, "dry_run", input.DryRun)
		return c.JSON(base.BaseResponse{Status: false, Message: "Migration failed", Error: err})
	}

	slog.Info("Status migration completed",
		"admin_id", admin.ID,
		"dry_run", input.DryRun,
		"user_procedures_updated", result.UserProceduresUpdated,
		"user_procedure_steps_updated", result.UserProcedureStepsUpdated)

	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: result})
}
