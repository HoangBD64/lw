package handler

import (
	"core/internal/domain_model/domain"
	"core/internal/domain_model/model"
	repo "core/internal/repo"
	"core/utils/base"
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func AdminCreateProcedure(c *fiber.Ctx) error {
	var input domain.ProcedureCreateRequest
	if err := c.BodyParser(&input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Review your input", Error: err})
	}
	admin, ok := c.Locals("admin").(model.Admin)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again"})
	}
	p := model.Procedure{
		Name:         input.Name,
		Description:  input.Description,
		IsHaveTicket: false,
		AdminID:      &admin.ID,
		Status:       model.PROCEDURE_CREATED,
		DueDate:      input.DueDate,
	}
	if err := repo.AdminCreateProcedure(&p); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't Create Procedure", Error: err})
	}
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(base.BaseResponse{Status: false, Message: "Procedure not found", Error: nil})
		}
		return c.JSON(base.BaseResponse{Status: false, Message: "Get Procedure Failed", Error: err})
	}
	if input.Name != "" {
		p.Name = input.Name
	}
	p.Description = input.Description
	p.DueDate = input.DueDate
	if err := repo.AdminUpdateProcedure(&p); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't Update Procedure", Error: err})
	}
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(base.BaseResponse{Status: false, Message: "Procedure not found", Error: nil})
		}
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

func AdminCreateProcedureStep(c *fiber.Ctx) error {
	var input domain.AdminCreateProcedureStepRequest
	if err := c.BodyParser(&input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Review your input", Error: err})
	}
	err := appUtils.Validate(c.Context(), &input)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Validation failed", Error: err})
	}
	step := model.ProcedureStep{
		ProcedureID: input.ProcedureID,
		Name:        input.Name,
		Status:      model.PROCEDURE_STEP_TODO,
		DueDate:     input.DueDate,
		AdminID:     input.AdminID,
		Notes:       input.Notes,
		Position:    input.Position,
		IsActive:    &input.IsActive,
		Data:        input.Data,
	}
	res, err := repo.CreateProcedureStep(step)
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
	step, err := repo.GetProcedureStepByID(idInt)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(base.BaseResponse{Status: false, Message: "Procedure step not found", Error: nil})
		}
		return c.JSON(base.BaseResponse{Status: false, Message: "Get Procedure Step Failed", Error: err})
	}

	step.ID = idInt
	step.Name = input.Name
	step.DueDate = input.DueDate
	step.AdminID = input.AdminID
	step.Notes = input.Notes
	step.Position = input.Position
	step.IsActive = &input.IsActive
	step.Data = input.Data
	if err := repo.UpdateProcedureStep(&step); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't UpdateProcedureStep", Error: err})
	}
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
	_, err = repo.GetProcedureStepByID(q.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(base.BaseResponse{Status: false, Message: "Procedure step not found", Error: nil})
		}
		return c.JSON(base.BaseResponse{Status: false, Message: "Get Procedure Step Failed", Error: err})
	}

	if err := repo.DeleteProcedureStep(q.ID); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't DeleteProcedureStep", Error: err})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success"})
}

func AdminListProcedureSteps(c *fiber.Ctx) error {
	idInt := c.Params("id")
	id, err := strconv.Atoi(idInt)
	if err != nil || id == 0 {
		return c.JSON(base.BaseResponse{Status: false, Message: "Invalid id", Error: err})
	}
	steps, err := repo.ListProcedureSteps(id)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't ListProcedureSteps", Error: err})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: steps})
}

func UserUpsertProcedureStepData(c *fiber.Ctx) error {
	var input domain.UserUpsertProcedureStepDataRequest
	if err := c.BodyParser(&input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Review your input", Error: err})
	}
	if input.UserProcedureID == 0 || input.ProcedureStepID == 0 {
		return c.JSON(base.BaseResponse{Status: false, Message: "user_procedure_id and procedure_step_id are required"})
	}
	ups := &model.UserProcedureStep{
		UserProcedureID: input.UserProcedureID,
		ProcedureStepID: input.ProcedureStepID,
		Status:          input.Status,
		Data:            input.Data,
	}
	if err := repo.UpsertUserProcedureStep(ups); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't upsert procedure step data", Error: err})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: ups})
}

func UserGetProcedureStepData(c *fiber.Ctx) error {
	id := c.Params("id")
	procedureID, err := strconv.Atoi(id)
	if err != nil || procedureID == 0 {
		return c.JSON(base.BaseResponse{Status: false, Message: "Invalid procedure ID", Error: err})
	}
	ups, err := repo.GetUserProcedureStep(procedureID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(base.BaseResponse{Status: false, Message: "Not found"})
		}
		return c.JSON(base.BaseResponse{Status: false, Message: "Get failed", Error: err})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: ups})
}

func UserListProcedureStepData(c *fiber.Ctx) error {
	var q domain.UserListProcedureStepDataRequest
	if err := c.QueryParser(&q); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Invalid query", Error: err})
	}
	if q.ID == 0 {
		return c.JSON(base.BaseResponse{Status: false, Message: "id (user_procedure_id) is required"})
	}
	items, err := repo.ListUserProcedureStepsByUserProcedure(q.ID)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "List failed", Error: err})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: items})
}

func AdminListUserProcedureStepData(c *fiber.Ctx) error {
	var q domain.AdminListUserProcedureStepDataRequest
	if err := c.QueryParser(&q); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Invalid query", Error: err})
	}
	if q.ID == 0 {
		return c.JSON(base.BaseResponse{Status: false, Message: "id (user_procedure_id) is required"})
	}
	items, err := repo.ListUserProcedureStepsByUserProcedure(q.ID)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "List failed", Error: err})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: items})
}

func UserGetProcedure(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(model.User)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again", Error: nil})
	}
	id := c.Params("id")
	procedureID, err := strconv.Atoi(id)
	if err != nil || procedureID == 0 {
		return c.JSON(base.BaseResponse{Status: false, Message: "Invalid procedure ID", Error: err})
	}
	userProcedure, err := repo.GetUserProcedureByID(user.ID, procedureID)
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		procedure, err := repo.GetProcedureByID(procedureID)
		if err != nil {
			return c.JSON(base.BaseResponse{Status: false, Message: "Get procedure failed", Error: err})
		}
		userProcedure := model.UserProcedure{
			UserID:       &user.ID,
			ProcedureID:  procedureID,
			Status:       model.PROCEDURE_CREATED,
			IsHaveTicket: false,
			Name:         procedure.Name,
			DueDate:      procedure.DueDate,
		}
		procedureSteps, err := repo.ListProcedureSteps(procedureID)
		if err != nil {
			return c.JSON(base.BaseResponse{Status: false, Message: "Get procedure steps failed", Error: err})
		}
		userProcedure, err = repo.CreateUserProcedure(userProcedure)
		if err != nil {
			return c.JSON(base.BaseResponse{Status: false, Message: "Create user procedure failed", Error: err})
		}
		userProcedureSteps := MappingProcedureStepsToUserProcedureSteps(procedureSteps, userProcedure.ID)
		for i := range userProcedureSteps {
			userProcedureSteps[i].UserProcedureID = userProcedure.ID
		}
		err = repo.UserCreateBulkProcedureStep(&userProcedureSteps)
		if err != nil {
			return c.JSON(base.BaseResponse{Status: false, Message: "Create user procedure steps failed", Error: err})
		}
		userProcedure.UserProcedureSteps = userProcedureSteps
		return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: userProcedure})
	} else if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Get procedure failed", Error: err})
	}

	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: userProcedure})
}

func MappingProcedureStepsToUserProcedureSteps(procedureSteps []model.ProcedureStep, userProcedureID int) []model.UserProcedureStep {
	var userProcedureSteps []model.UserProcedureStep
	for _, step := range procedureSteps {
		userProcedureStep := model.UserProcedureStep{
			UserProcedureID: userProcedureID,
			ProcedureStepID: step.ID,
			Status:          model.PROCEDURE_STEP_TODO,
			Data:            step.Data,
			Name:            step.Name,
			DueDate:         step.DueDate,
			AdminID:         step.AdminID,
			Notes:           step.Notes,
			Position:        step.Position,
			IsActive:        step.IsActive,
		}
		userProcedureSteps = append(userProcedureSteps, userProcedureStep)
	}
	return userProcedureSteps
}
