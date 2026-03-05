package router

import (
	"context"
	"core/app"
	mainapp "core/app"
	"core/internal/handler"
	"core/internal/middleware"
	"core/internal/scheduler"
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
	scheduler.Setup(context.TODO())

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

	// ---------user---------
	api_user := fiber_app.Group("/api/user")

	// api upblic
	api_user.Get("/payment/info.json", handler.GetPaymentInfo)
	api_user.Post("/contact-forms.json", handler.CreateContactForm)
	api_user.Post("/login_sso.json", handler.VerifyUserSSO)
	api_user.Use(middleware.Authen)

	// user info
	api_user.Get("/info/detail.json", middleware.Author(""), handler.GetUserDetail)
	api_user.Patch("/info.json", middleware.Author(""), handler.UpdateUserInfo)

	//s3
	api_user.Put("/s3.json", middleware.Author(""), handler.GetUploadUrl)

	// documents
	api_user.Get("/documents.json", middleware.Author(""), handler.UserGetUserDocuments)
	api_user.Get("/documents/:id.json", middleware.Author(""), handler.UserGetUserDocument)
	api_user.Post("/documents.json", middleware.Author(""), handler.UserCreateDocument)
	api_user.Patch("/documents/:id.json", middleware.Author(""), handler.UserUpdateDocument)

	// notification preferences
	api_user.Put("/notification-preferences.json", middleware.Author(""), handler.UserUpsertNotificationPreference)
	api_user.Get("/notification-preferences.json", middleware.Author(""), handler.UserGetNotificationPreference)
	// notification
	api_user.Get("/notification.json", middleware.Author(""), handler.GetNotifications)
	api_user.Get("/notification/detail.json", middleware.Author(""), handler.GetNotificationDetail)
	api_user.Get("/notification/total.json", middleware.Author(""), handler.GetTotalNotificationType)
	api_user.Get("/notification/mark_read_all.json", middleware.Author(""), handler.MarkReadAllNotification)

	// procedure
	api_user.Get("/procedures-templates.json", middleware.Author(""), handler.UserGetProcedures)
	api_user.Get("/procedures-templates/:id.json", middleware.Author(""), handler.UserGetProcedure)
	api_user.Get("/procedures.json", middleware.Author(""), handler.UserListUserProcedures)
	api_user.Get("/procedures/:id/steps.json", middleware.Author(""), handler.UserGetProcedureDetail)
	// api_user.Post("/procedures.json", middleware.Author(""), handler.UserCreateProcedure)
	// procedure step
	api_user.Get("/procedures-steps.json", middleware.Author(""), handler.UserGetUserProcedureSteps)
	api_user.Patch("/procedures/steps/:id.json", middleware.Author(""), handler.UserUpdateUserProcedureStep)

	// tickets
	api_user.Get("/tickets.json", middleware.Author(""), handler.GetTickets)
	api_user.Post("/tickets.json", middleware.Author(""), handler.CreateTicket)
	api_user.Get("/tickets/detail.json", middleware.Author(""), handler.GetTicketDetail)
	api_user.Post("/tickets/content.json", middleware.Author(""), handler.CreateTicketContent)
	api_user.Patch("/tickets/status.json", middleware.Author(""), handler.UpdateStatusTicket)

	//payment
	// setup all info in checkout page: client id, token , ..
	// api_user.Get("/payment/info.json", middleware.Author(""), handler.GetPaymentInfo)
	// setup all thing needed in server: calculate pricing, offer, ...
	api_user.Post("/payment/init.json", middleware.Author(""), handler.InitPayment)
	// finish transaction: update needed in server
	api_user.Post("/payment/finish.json", middleware.Author(""), handler.FinishPayment)

	// dashboard
	api_user.Get("/dashboard.json", middleware.Author(""), handler.UserDashboard)

	// payment
	api_user.Get("/transaction.json", middleware.Author(""), handler.GetTransactions)
	api_user.Get("/transaction/detail.json", middleware.Author(""), handler.GetTransactionDetails)
	api_user.Get("/transaction/balance.json", middleware.Author(""), handler.GetUserBalance)
	api_user.Get("/transaction/analytics.json", middleware.Author(""), handler.GetBillingAnalytics)
	api_user.Get("/transaction/export.json", middleware.Author(""), handler.ExportTransactions)

	// ---------admin---------
	api_admin := fiber_app.Group("/api/admin")
	api_admin.Post("/login_sso.json", handler.VerifyAdminSSO)
	api_admin.Use(middleware.AuthenAdmin)

	// admin info
	api_admin.Get("/admin.json", middleware.AuthorizeAdmin("admin", "view"), handler.GetAdmin)
	api_admin.Get("/admins.json", middleware.AuthorizeAdmin("admin", "view"), handler.GetAllAdmin)
	api_admin.Patch("/admin/status.json", middleware.AuthorizeAdmin("admin", "manage"), handler.AdminUpdateAdminStatus)
	api_admin.Patch("/admin/role.json", middleware.AuthorizeAdmin("admin", "manage"), handler.AdminUpdateRole)
	api_admin.Patch("admin/permissions.json", middleware.AuthorizeAdmin("admin", "manage"), handler.AdminUpdatePermissions)

	//s3
	api_admin.Put("/s3.json", handler.AdminGetUploadUrl)

	// users
	api_admin.Get("/user.json", middleware.AuthorizeAdmin("user", "view"), handler.AdminGetUser)
	api_admin.Patch("/user.json", middleware.AuthorizeAdmin("user", "manage"), handler.AdminUpdateUser)
	api_admin.Patch("/user/status.json", middleware.AuthorizeAdmin("user", "manage"), handler.AdminUpdateUserStatus)

	// settings
	api_admin.Get("/setting.json", middleware.AuthorizeAdmin("setting", "view"), handler.GetSystemSetting)
	api_admin.Patch("/setting.json", middleware.AuthorizeAdmin("setting", "manage"), handler.AdminUpdateSetting)
	// test send mail
	api_admin.Post("/setting/test-mail.json", middleware.AuthorizeAdmin("setting", "manage"), handler.AdminTestSendMail)

	// documents
	api_admin.Get("/documents/:id.json", middleware.AuthorizeAdmin("document", "view"), handler.AdminGetDocument)
	api_admin.Get("/documents.json", middleware.AuthorizeAdmin("document", "view"), handler.AdminGetDocuments)
	api_admin.Post("/documents.json", middleware.AuthorizeAdmin("document", "manage"), handler.AdminCreateDocument)
	api_admin.Patch("/documents/:id.json", middleware.AuthorizeAdmin("document", "manage"), handler.AdminUpdateDocument)
	api_admin.Delete("/documents.json", middleware.AuthorizeAdmin("document", "manage"), handler.AdminDeleteDocument)

	// procedures
	api_admin.Get("/procedures/:id.json", middleware.AuthorizeAdmin("procedures", "view"), handler.AdminGetProcedure)
	api_admin.Get("/procedures.json", middleware.AuthorizeAdmin("procedures", "view"), handler.AdminGetProcedures)
	api_admin.Post("/procedures.json", middleware.AuthorizeAdmin("procedures", "manage"), handler.AdminCreateProcedure)
	api_admin.Patch("/procedures/:id.json", middleware.AuthorizeAdmin("procedures", "manage"), handler.AdminUpdateProcedure)
	api_admin.Delete("/procedures.json", middleware.AuthorizeAdmin("procedures", "manage"), handler.AdminDeleteProcedure)
	api_admin.Post("/procedures/bulk.json", middleware.AuthorizeAdmin("procedures", "manage"), handler.AdminBulkUpsertProcedure)

	// procedure steps
	api_admin.Get("/procedure/:id/steps.json", middleware.AuthorizeAdmin("procedures", "view"), handler.AdminListProcedureSteps)
	api_admin.Post("/procedure/steps.json", middleware.AuthorizeAdmin("procedures", "manage"), handler.AdminCreateProcedureStep)
	api_admin.Patch("/procedure/steps/:id.json", middleware.AuthorizeAdmin("procedures", "manage"), handler.AdminUpdateProcedureStep)
	api_admin.Delete("/procedure/steps.json", middleware.AuthorizeAdmin("procedures", "manage"), handler.AdminDeleteProcedureStep)

	// user procedure
	api_admin.Get("/user-procedures.json", middleware.AuthorizeAdmin("procedures", "view"), handler.AdminListUserProcedures)
	api_admin.Get("/user-procedures/:id.json", middleware.AuthorizeAdmin("procedures", "view"), handler.AdminGetDetailUserProcedures)
	api_admin.Patch("/user-procedures/:id.json", middleware.AuthorizeAdmin("procedures", "manage"), handler.AdminUpdateUserProcedure)
	api_admin.Patch("/user-procedures/admin/assign.json", middleware.AuthorizeAdmin("assignment", "manage"), handler.AdminAssignUserProcedure)
	api_admin.Patch("/user-procedures/admin/unassign.json", middleware.AuthorizeAdmin("assignment", "manage"), handler.AdminUnassignUserProcedure)

	// user procedure step
	api_admin.Get("/user-procedures-steps.json", middleware.AuthorizeAdmin("procedures", "view"), handler.AdminListUserProcedureSteps)
	api_admin.Patch("/user-procedures/:id/steps/assign.json", middleware.AuthorizeAdmin("assignment", "manage"), handler.AdminAssignLawyerToProcedureStep)
	api_admin.Patch("/user-procedures/steps/status.json", middleware.AuthorizeAdmin("procedures", "manage"), handler.AdminUpdateUserProcedureStepStatus)
	api_admin.Patch("/user-procedures/steps/:id.json", middleware.AuthorizeAdmin("assignment", "manage"), handler.AdminUpdateUserProcedureStep)

	// migration
	api_admin.Post("/procedures/migrate-status.json", middleware.AuthorizeAdmin("procedures", "manage"), handler.AdminMigrateStatus)

	// tickets
	api_admin.Get("/tickets/detail.json", middleware.AuthorizeAdmin("ticket", "view"), handler.AdminGetTicketDetail)
	api_admin.Get("/tickets.json", middleware.AuthorizeAdmin("ticket", "view"), handler.AdminGetTicket)
	api_admin.Post("/tickets/content.json", middleware.AuthorizeAdmin("ticket", "manage"), handler.AdminCreateTicketContent)

	// transaction
	api_admin.Get("/transaction.json", middleware.AuthorizeAdmin("transaction", "view"), handler.AdminGetTransaction)

	// analytics overview
	api_admin.Get("/analytics.json", middleware.AuthorizeAdmin("procedures", "view"), handler.AdminGetSystemStats)

	// contact forms
	api_admin.Get("/contact-forms.json", handler.AdminGetContactForms)

	// notification
	api_admin.Get("/notification.json", handler.AdminGetNotifications)
	api_admin.Get("/notification/detail.json", handler.AdminGetNotificationDetail)
	api_admin.Get("/notification/total.json", handler.AdminGetTotalNotificationType)
	api_admin.Get("/notification/mark_read_all.json", handler.AdminMarkReadAllNotification)

	// extra fee
	api_admin.Get("/extra-fees.json", middleware.AuthorizeAdmin("procedures", "view"), handler.AdminGetExtraFees)
	api_admin.Get("/extra-fees/:id.json", middleware.AuthorizeAdmin("procedures", "view"), handler.AdminGetExtraFeeDetail)
	api_admin.Post("/extra-fees.json", middleware.AuthorizeAdmin("procedures", "manage"), handler.AdminCreateExtraFee)
	api_admin.Patch("/extra-fees.json", middleware.AuthorizeAdmin("procedures", "manage"), handler.AdminUpdateExtraFee)
	api_admin.Delete("/extra-fees/:id.json", middleware.AuthorizeAdmin("procedures", "manage"), handler.AdminDeleteExtraFee)
}

func Recover() {
	if err := recover(); err != nil {
		slog.Error("System encountered a FATAL PANIC", "error", err, "out", "lark", "stack", fmt.Sprintf("\n%s", debug.Stack()))
	}
}
