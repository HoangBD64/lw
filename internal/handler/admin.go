package handler

import (
	"core/app"
	"core/internal/domain_model/domain"
	"core/internal/domain_model/model"
	"core/internal/repo"
	"core/utils/base"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/casdoor/casdoor-go-sdk/casdoorsdk"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

func VerifyAdminSSO(c *fiber.Ctx) error {
	var input domain.VerifyAdminSSORequest
	if err := c.BodyParser(&input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Invalid request body", Error: err})
	}
	var claims *casdoorsdk.Claims
	token, err := app.SSO.GetOAuthToken(input.Code, input.State)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Error get oauth token", Error: err})
	}
	if token == nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Token is nil", Error: nil})
	}
	claims, err = app.SSO.ParseJwtToken(token.AccessToken)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Failed Parse Jwt Token", Error: err})
	}
	current_admin, err := repo.GetAdminBySsoID(claims.User.Id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			newAdmin := model.Admin{
				Name:   claims.User.DisplayName,
				Email:  &claims.User.Email,
				Phone:  &claims.User.Phone,
				Gender: claims.User.Gender,
				Avatar: &claims.User.Avatar,
				SSO_ID: &claims.User.Id,
			}

			if strings.TrimSpace(newAdmin.Name) == "" {
				newAdmin.Name = "Supervisor"
			}
			current_admin, err = repo.CreateAdmin(newAdmin)
			if err != nil {
				return c.JSON(base.BaseResponse{Status: false, Message: "Create Admin Failed", Error: err})
			}
		} else {
			return c.JSON(base.BaseResponse{Status: false, Message: "Can't GetAdminBySsoID", Error: err})
		}
	}
	claims_new := jwt.MapClaims{
		"admin_id": current_admin.ID,
		"name":     current_admin.Name,
		"sso_id":   current_admin.SSO_ID,
		"status":   true,
		"exp":      time.Now().Add(time.Hour * 200).Unix(),
	}
	token_new := jwt.NewWithClaims(jwt.SigningMethodHS256, claims_new)
	t, err := token_new.SignedString([]byte(app.Config("SECRET_ADMIN")))
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "token generation failed", Error: err})
	}
	adminReturn := domain.DataAdminResponse{
		ID:          current_admin.ID,
		Name:        current_admin.Name,
		Avatar:      current_admin.Avatar,
		Email:       current_admin.Email,
		Phone:       current_admin.Phone,
		DOB:         current_admin.DOB,
		Gender:      current_admin.Gender,
		SsoID:       current_admin.SSO_ID,
		IsOwner:     current_admin.IsOwner,
		Status:      current_admin.Status,
		Role:        current_admin.Role,
		Permissions: current_admin.Permissions,
		Token:       t,
		UserType:    "admin",
	}
	now := time.Now()
	current_admin.LastLogin = &now
	go repo.UpdateAdminLoginTime(current_admin)
	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: adminReturn})
}

func GetAdminDetail(c *fiber.Ctx) error {
	admin, ok := c.Locals("admin").(model.Admin)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again", Error: nil})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: admin, Error: nil})
}

func GetAdmin(c *fiber.Ctx) error {
	admin, ok := c.Locals("admin").(model.Admin)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again", Error: nil})
	}

	var query repo.Query
	query.ParseForAdmin(c, admin)

	admins, total, err := repo.GetAdmin(query)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't GetAdmin", Error: err})
	}

	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: admins, Total: &total})
}

func AdminUpdateAdminStatus(c *fiber.Ctx) error {
	admin, ok := c.Locals("admin").(model.Admin)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again", Error: nil})
	}
	if !admin.IsOwner {
		return c.JSON(base.BaseResponse{Status: false, Message: "only admin owner", Error: nil})
	}
	type Input struct {
		IDS    []uuid.UUID `json:"ids" validate:"required,gt=0"`
		Status string      `json:"status"`
	}
	var input Input
	if err := c.BodyParser(&input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Invalid request body", Error: err})
	}
	if err := appUtils.Validate(c.Context(), &input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "bad request", Error: err})
	}
	if input.Status != model.ACTIVE && input.Status != model.INACTIVE {
		return c.JSON(base.BaseResponse{Status: false, Message: "can only update status to 'active' or 'inactive'", Error: nil})
	}
	list_admin, err := repo.GetAdminByIds(input.IDS)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't GetAdminByIds", Error: err})
	}
	if len(list_admin) != len(input.IDS) {
		return c.JSON(base.BaseResponse{Status: false, Message: "Not enough users were found", Error: nil})
	}
	_, ok_check_owner := lo.Find(list_admin, func(admin model.Admin) bool {
		return admin.IsOwner
	})
	if ok_check_owner {
		return c.JSON(base.BaseResponse{Status: false, Message: "Cannot intervene to change the owner's status", Error: nil})
	}
	row_change, err := repo.UpdateAdminStatus(input.IDS, input.Status)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't UpdateAdminStatus", Error: err})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: fiber.Map{"total": row_change}})
}

// update admin permission
func AdminUpdatePermissions(c *fiber.Ctx) error {
	admin, ok := c.Locals("admin").(model.Admin)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again", Error: nil})
	}

	if !admin.IsOwner {
		return c.JSON(base.BaseResponse{Status: false, Message: "only admin owner", Error: nil})
	}

	type Input struct {
		ID          uuid.UUID             `json:"id" validate:"required"`
		Permissions model.AdminPermission `json:"permissions"`
	}
	var input Input
	if err := c.BodyParser(&input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Invalid request body", Error: err})
	}

	appUtils.Validate(c.Context(), &input)

	admin_found, err := repo.GetAdminByID(c.Context(), input.ID)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Admin not found", Error: err})
	}

	permision_json, err := json.Marshal(input.Permissions)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Invalid permissions format", Error: err})
	}

	admin_found.Permissions = permision_json

	admin_update, err := repo.UpdateAdminPermissions(c.Context(), admin_found)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't UpdateAdminPermissions", Error: err})
	}

	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: fiber.Map{"admin": admin_update}})
}

func GetAllAdmin(c *fiber.Ctx) error {
	admin, ok := c.Locals("admin").(model.Admin)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again", Error: nil})
	}
	var query repo.Query
	query.ParseForAdmin(c, admin)

	admins, err := repo.GetAllAdmin(c.Context(), query)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "failed to get data", Error: err})
	}

	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: admins})
}

func AdminUpdateRole(c *fiber.Ctx) error {
	admin, ok := c.Locals("admin").(model.Admin)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again", Error: nil})
	}
	if !admin.IsOwner {
		return c.JSON(base.BaseResponse{Status: false, Message: "only admin owner", Error: nil})
	}
	type Input struct {
		ID   *uuid.UUID `json:"id"`
		Role string     `json:"role"`
	}
	var input Input
	if err := c.BodyParser(&input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Invalid request body", Error: err})
	}

	if input.ID == nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "id can't empty", Error: nil})
	}

	admin_found, err := repo.GetAdminByID(c.Context(), *input.ID)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Admin not found", Error: err})
	}
	if !isValidAdminRole(input.Role) {
		return c.JSON(base.BaseResponse{Status: false, Message: "Invalid role. Must be 'admin' or 'cs'", Error: nil})
	}
	admin_found.Role = input.Role
	_, err = repo.UpdateAdminRole(c.Context(), admin_found)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't UpdateAdminRole", Error: err})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success"})
}

func isValidAdminRole(r string) bool {
	r = strings.ToLower(strings.TrimSpace(r))
	return r == model.RoleAdmin
}
