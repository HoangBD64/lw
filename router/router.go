package router

import (
	"core/app"
	mainapp "core/app"
	"core/internal/handler"
	"core/internal/middleware"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	rev "github.com/gofiber/fiber/v2/middleware/recover"
	slogfiber "github.com/samber/slog-fiber"
)

func Setup() {
	app.InitLoggers()

	app := fiber.New(fiber.Config{
		BodyLimit: 200 * 1024 * 1024,
		// Override default error handler
		EnableTrustedProxyCheck: true,
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			var e *fiber.Error
			if errors.As(err, &e) {
				return ctx.Status(e.Code).JSON(fiber.Map{"status": false, "message": e.Message})
			}

			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": false, "message": fiber.ErrInternalServerError.Message})
		},
	})

	app.Use(cors.New())
	app.Use(rev.New(rev.Config{EnableStackTrace: true}))

	config := slogfiber.Config{
		WithRequestBody:    true,
		WithSpanID:         true,
		WithTraceID:        true,
		WithResponseBody:   true,
		WithRequestHeader:  true,
		WithResponseHeader: false,
	}
	if !mainapp.IsInProduction() {
		config.WithRequestBody = false
		config.WithResponseBody = false
	}

	logger := slog.Default().With("app", "legaswift")
	app.Use(slogfiber.NewWithConfig(logger, config))

	port := mainapp.Config("WEB_PORT")
	if len(port) == 0 {
		port = "3535"
	}
	setupRouter(app)

	go func() {
		if err := app.Listen(":" + port); err != nil {
			slog.Error("Can't start server", "error", err)
		}
	}()

	// graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	_ = <-c
	fmt.Println("Gracefully shutting down...")
	app.Shutdown()
}

func setupRouter(fiber_app *fiber.App) {
	fiber_app.Get("/status.json", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": true})
	})

	api_user := fiber_app.Group("/api/user")

	api_user.Post("/login_sso.json", handler.VerifyUserSSO)
	api_user.Use(middleware.Authen)

	// user info
	api_user.Get("/info/detail.json", handler.GetUserDetail)
	api_user.Patch("/info.json", handler.UpdateUserInfo)

	// documents
	// api_user.Get("/documents/:id.json", handler.GetDocumentByID)
	api_user.Get("/documents.json", handler.UserGetUserDocuments)
	// api_user.Get("/documents.json", handler.AdminGetDocuments)
	api_user.Post("/documents.json", handler.UserCreateDocument)
	api_user.Patch("/documents/:id.json", handler.UserUpdateDocument)

	//s3
	api_user.Put("/s3.json", middleware.Author(""), handler.GetUploadUrl)

	// admin
	api_admin := fiber_app.Group("/api/admin")
	api_admin.Post("/login_sso.json", handler.VerifyAdminSSO)

	api_admin.Use(middleware.AuthenAdmin)

	// permissions
	api_admin.Patch("/permissions.json", middleware.AuthorizeAdmin("admin", "manage"), handler.AdminUpdatePermissions)

	//s3
	api_admin.Put("/s3.json", handler.AdminGetUploadUrl)

	// users
	api_admin.Get("/user.json", middleware.AuthorizeAdmin("user", "view"), handler.AdminGetUser)
	api_admin.Patch("/user.json", middleware.AuthorizeAdmin("user", "manage"), handler.AdminUpdateUser)
	api_admin.Patch("/user/status.json", middleware.AuthorizeAdmin("user", "manage"), handler.AdminUpdateUserStatus)

	// settings
	api_admin.Get("/setting.json", middleware.AuthorizeAdmin("setting", "view"), handler.GetSystemSetting)
	api_admin.Patch("/setting.json", middleware.AuthorizeAdmin("setting", "manage"), handler.AdminUpdateSetting)

	// documents
	api_admin.Get("/documents/:id.json", middleware.AuthorizeAdmin("document", "view"), handler.AdminGetDocument)
	api_admin.Get("/documents.json", middleware.AuthorizeAdmin("document", "view"), handler.AdminGetDocuments)
	api_admin.Post("/documents.json", middleware.AuthorizeAdmin("document", "manage"), handler.AdminCreateDocument)
	api_admin.Patch("/documents/:id.json", middleware.AuthorizeAdmin("document", "manage"), handler.AdminUpdateDocument)
	api_admin.Delete("/documents.json", middleware.AuthorizeAdmin("document", "manage"), handler.AdminDeleteDocument)

	// document assignments
	api_admin.Post("/documents/assign.json", middleware.AuthorizeAdmin("document", "manage"), handler.AdminAssignDocument)
	// api_admin.Delete("/documents/assign.json", middleware.AuthorizeAdmin("document", "manage"), handler.AdminUnassignDocument)
}

func Recover() {
	if err := recover(); err != nil {
		slog.Error("System encountered a FATAL PANIC", "error", err, "out", "lark", "stack", fmt.Sprintf("\n%s", debug.Stack()))
	}
}
