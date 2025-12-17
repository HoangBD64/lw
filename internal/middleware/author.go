package middleware

import (
	"core/internal/domain_model/model"
	"encoding/json"

	"github.com/gofiber/fiber/v2"
)

func Author(role string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return RoutePermission(c, role)
	}
}

func RoutePermission(c *fiber.Ctx, role string) error {
	user, ok := c.Locals("user").(model.User)
	if !ok {
		return c.JSON(fiber.Map{"status": false, "error": "Please login again"})
	}
	if role == "" {
		return c.Next()
	}
	if user.Role == role {
		return c.Next()
	}
	return c.JSON(fiber.Map{"status": false, "message": "Actions not allow"})
}

func AuthorizeAdmin(module string, action string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		admin, ok := c.Locals("admin").(model.Admin)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"status": false,
				"error":  "Please login again",
			})
		}
		if admin.IsOwner {
			return c.Next()
		}
		admin_permissions := model.AdminPermission{}
		if err := json.Unmarshal(admin.Permissions, &admin_permissions); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  false,
				"message": "Failed to parse admin permissions",
				"error":   err.Error(),
			})
		}

		var allowed bool

		switch module {
		case "admin":
			allowed = checkPermission(admin_permissions.Admin, action)
		case "user":
			allowed = checkPermission(admin_permissions.User, action)
		case "setting":
			allowed = checkPermission(admin_permissions.Setting, action)
		case "document":
			allowed = checkPermission(admin_permissions.Document, action)
		case "procedures":
			allowed = checkPermission(admin_permissions.Procedures, action)
		case "assignment":
			allowed = checkPermission(admin_permissions.Assignment, action)
		case "ticket":
			allowed = checkPermission(admin_permissions.Ticket, action)
		default:
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"status":     false,
				"message":    "Invalid module",
				"error_code": "INVALID_AUTHORIZATION",
			})
		}
		if !allowed {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"status":     false,
				"message":    "Not authorized to perform this action",
				"error_code": "INVALID_AUTHORIZATION",
			})
		}

		return c.Next()
	}
}

func checkPermission(p model.Permission, action string) bool {
	switch action {
	case "view":
		return p.View
	case "manage":
		return p.Manage
	default:
		return false
	}
}
