package handler

import (
	"context"
	"errors"
	"strings"
	"time"

	"core/app"
	"core/internal/domain_model/domain"
	"core/internal/domain_model/model"
	"core/internal/repo"
	"core/utils"
	"core/utils/base"

	"github.com/google/uuid"

	"github.com/casdoor/casdoor-go-sdk/casdoorsdk"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"gorm.io/gorm"
)

func VerifyUserSSO(c *fiber.Ctx) error {
	var input domain.VerifyUserSSORequest
	if err := c.BodyParser(&input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Invalid request body", Error: err})
	}
	var claims *casdoorsdk.Claims
	token, err := app.SSO.GetOAuthToken(input.Code, input.State)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Error get oauth token", Error: err})
	}
	if token == nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Failed after GetOauthToken", Error: nil})
	}
	claims, err = app.SSO.ParseJwtToken(token.AccessToken)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Failed Parse Jwt Token", Error: err})
	}
	current_user, err := repo.GetUserBySsoID(claims.User.Id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			newUser := model.User{
				Name:   claims.User.DisplayName,
				Email:  &claims.User.Email,
				Phone:  &claims.User.Phone,
				Gender: claims.User.Gender,
				Avatar: &claims.User.Avatar,
				SSO_ID: &claims.User.Id,
			}

			if strings.TrimSpace(newUser.Name) == "" {
				newUser.Name = "Visitor"
			}
			current_user, err = repo.CreateUser(newUser)
			if err != nil {
				return c.JSON(base.BaseResponse{Status: false, Message: "Can't create user", Error: err})
			}
			utils.RunAsync(3*time.Second, func(ctx context.Context) error {
				return PreDefaultSettingsUser(current_user)
			})
		} else {
			return c.JSON(base.BaseResponse{Status: false, Message: "Can't get user by sso id", Error: err})
		}
	} else {
		if current_user.Name == "Visitor" && claims.User.DisplayName != "" && current_user.Name != claims.User.DisplayName {
			current_user.Name = claims.User.DisplayName
			utils.RunAsync(3*time.Second, func(ctx context.Context) error {
				return repo.UpdateUserDisplayName(current_user)
			})
		}
	}
	claims_new := jwt.MapClaims{
		"user_id": current_user.ID,
		"name":    current_user.Name,
		"sso_id":  current_user.SSO_ID,
		"status":  true,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}
	token_new := jwt.NewWithClaims(jwt.SigningMethodHS256, claims_new)
	t, err := token_new.SignedString([]byte(app.Config("SECRET_USER")))
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "token generation failed", Error: err})
	}
	userReturn := domain.DataUserResponse{
		ID:       current_user.ID,
		Name:     current_user.Name,
		Avatar:   current_user.Avatar,
		Email:    current_user.Email,
		Phone:    current_user.Phone,
		Address:  current_user.Address,
		DOB:      current_user.DOB,
		Gender:   current_user.Gender,
		SsoID:    current_user.SSO_ID,
		Status:   current_user.Status,
		Token:    t,
		UserType: "user",
	}
	now := time.Now()
	current_user.LastLogin = &now
	utils.RunAsync(3*time.Second, func(ctx context.Context) error {
		return repo.UpdateUserLoginTime(current_user)
	})
	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: userReturn})
}

func PreDefaultSettingsUser(user model.User) error {
	// setup noti default
	emailEnable := true
	_, err := repo.UpsertNotificationPreferenceByUserID(user.ID, emailEnable)
	if err != nil {
		return err
	}

	return nil
}

func GetUserDetail(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(model.User)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again", Error: nil})
	}
	user.SystemNote = nil
	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: user})
}

func UpdateUserInfo(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(model.User)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again", Error: nil})
	}
	var input domain.UpdateUserInfoRequest
	if err := c.BodyParser(&input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Review your input", Error: err})
	}
	if input.Name != "" {
		user.Name = input.Name
	}
	if input.Phone != nil {
		user.Phone = input.Phone
	}
	if input.Avatar != nil {
		user.Avatar = input.Avatar
	}
	if input.Gender != "" {
		user.Gender = input.Gender
	}
	if input.DOB != "" {
		user.DOB = input.DOB
	}
	if input.Address != "" {
		user.Address = input.Address
	}
	if input.Metadata != nil {
		user.Metadata = *input.Metadata
	}
	user_updated, err := repo.UpdateUserInfo(user)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't UpdateUserInfo", Error: err})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: user_updated})
}

func AdminUpdateUser(c *fiber.Ctx) error {
	var input domain.AdminUpdateUserRequest
	if err := c.BodyParser(&input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Review your input", Error: err})
	}
	if input.ID == uuid.Nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "id can't empty", Error: nil})
	}

	user, err := repo.GetUserByID(input.ID)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't GetUser", Error: err})
	}

	if input.SystemNote != nil {
		user.SystemNote = *input.SystemNote
	}

	_, err = repo.AdminUpdateUser(c.Context(), user)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't Update User Fields", Error: err})
	}

	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Error: nil})
}

func AdminGetUser(c *fiber.Ctx) error {
	var query repo.Query
	query.Parse(c)

	users, total, err := repo.AdminGetUser(query)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't GetUser", Error: err})
	}

	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: fiber.Map{"users": users, "total": total}})
}

func AdminUpdateUserStatus(c *fiber.Ctx) error {
	var input domain.AdminUpdateUserStatusRequest
	if err := c.BodyParser(&input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Invalid request body", Error: err})
	}
	if err := appUtils.Validate(c.Context(), &input); err != nil {
		return c.JSON(base.BaseResponse{Message: "bad request", Error: err})
	}
	var query repo.Query
	_ = c.QueryParser(&query)
	query.IDS = input.IDS
	list_user, err := repo.GetUsersByAdmin(c.Context(), query)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't GetUsers", Error: err})
	}
	if len(list_user) != len(input.IDS) {
		return c.JSON(base.BaseResponse{Status: false, Message: "Not enough users were found", Error: nil})
	}
	row_change, err := repo.UpdateUserStatus(input.IDS, input.Status)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't UpdateUserStatus", Error: err})
	}

	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Total: &row_change})
}

func UserDashboard(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(model.User)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again"})
	}

	var query repo.Query
	query.Parse(c)

	counts, err := repo.CountUserProceduresByStatus(user.ID, query)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Failed to get dashboard", Error: err})
	}
	resp := domain.UserDashboardResponse{
		InformationRequired: counts[model.USER_PROCEDURE_INFORMATION_REQUIRED],
		Inprogress:          counts[model.USER_PROCEDURE_IN_PROGRESS],
		Created:             counts[model.USER_PROCEDURE_CREATED],
		Completed:           counts[model.USER_PROCEDURE_COMPLETED],
	}
	return c.JSON(base.BaseResponse{Status: true, Data: resp})
}
