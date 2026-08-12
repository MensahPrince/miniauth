// miniauth/routes.go
package miniauth

import (
	"github.com/MensahPrince/miniauth/handlers"
	"github.com/MensahPrince/miniauth/middleware"
	"github.com/gofiber/fiber/v3"
)

func registerRoutes(app *fiber.App) {
	app.Get("/", handlers.Base)
	app.Post("/register", handlers.Register)
	app.Post("/login", handlers.Login)
	app.Get("/profile", middleware.JWTMiddleware, handlers.FetchProfile)
	app.Post("/edit/:param", middleware.JWTMiddleware, handlers.EditHandler)
	app.Get("/request-otp", middleware.JWTMiddleware, handlers.RequestOTP)
	app.Post("/reset", middleware.JWTMiddleware, handlers.ResetPassword)
	app.Post("/delete", middleware.JWTMiddleware, handlers.DeleteAccount)

	// Patient routes — require valid JWT
	app.Post("/patients", middleware.JWTMiddleware, handlers.CreatePatient)
	app.Get("/patients", middleware.JWTMiddleware, handlers.GetPatients)

	// Admin routes — require both a valid JWT and admin role
	admin := app.Group("/admin", middleware.JWTMiddleware, middleware.AdminMiddleware)
	admin.Post("/users", handlers.AdminCreateUser)
	admin.Delete("/users/:id", handlers.AdminRevokeUser)
	admin.Post("/users/:id/reset", handlers.AdminResetPassword)
	admin.Get("/users", handlers.GetUsers)
	admin.Get("/logs", handlers.GetLogs)
}
