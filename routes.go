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
}
