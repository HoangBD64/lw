package middleware

import (
	"core/app"
	"core/internal/domain_model/model"
	"core/internal/repo"
	"core/utils/base"
	"fmt"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	jwtware "github.com/gofiber/jwt/v3"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	slogfiber "github.com/samber/slog-fiber"
)

func AuthenAdmin(c *fiber.Ctx) error {
	return jwtware.New(jwtware.Config{
		SigningKey:     []byte(app.Config("SECRET_ADMIN")),
		ErrorHandler:   jwtError,
		SuccessHandler: jwtAdminSuccess,
		ContextKey:     "token",
	})(c)
}

func jwtAdminSuccess(c *fiber.Ctx) error {
	u := c.Locals("token").(*jwt.Token)
	claims := u.Claims.(jwt.MapClaims)
	adminId, ok := claims["admin_id"].(string)

	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(base.BaseResponse{Status: false, Message: "Invalid or expired JWT"})
	}

	admin_id, err := uuid.Parse(adminId)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(base.BaseResponse{Status: false, Message: "Invalid or expired JWT", Error: err})
	}

	admin, err := repo.GetAdminByID(c.Context(), admin_id)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(base.BaseResponse{Status: false, Message: "Can not found admin", Error: err})
	}

	if admin.Status != model.ACTIVE {
		return c.Status(fiber.StatusUnauthorized).JSON(base.BaseResponse{Status: false, Message: "Admin status inactive"})
	}
	c.Locals("admin", admin)
	return c.Next()
}

func Authen(c *fiber.Ctx) error {
	return jwtware.New(jwtware.Config{
		SigningKey:     []byte(app.Config("SECRET_USER")),
		ErrorHandler:   jwtError,
		SuccessHandler: jwtSuccess,
		ContextKey:     "token",
	})(c)
}
func jwtSuccess(c *fiber.Ctx) error {
	u := c.Locals("token").(*jwt.Token)
	claims := u.Claims.(jwt.MapClaims)
	userId, ok := claims["user_id"].(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": false, "message": "Invalid or expired JWT"})
	}
	user_id, err := uuid.Parse(userId)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": false, "message": "Invalid or expired JWT", "error": err.Error()})
	}
	user, err := repo.GetUserByID(user_id)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": false, "message": "Can not found user", "error": err.Error()})
	}
	if user.Status != model.ACTIVE {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": false, "message": "Account is not active"})
	}

	c.Locals("user", user)
	// add more data for this
	slogfiber.AddCustomAttributes(c, slog.String("user_id", fmt.Sprint(user.ID)))
	slogfiber.AddCustomAttributes(c, slog.String("user_name", user.Name))
	return c.Next()
}

func jwtError(c *fiber.Ctx, err error) error {
	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": false, "message": "Invalid or expired JWT", "error": err.Error()})
}
