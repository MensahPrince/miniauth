package handlers

import (
	"github.com/MensahPrince/miniauth/db"
	"github.com/MensahPrince/miniauth/types"
	"github.com/gofiber/fiber/v3"
)

func FetchProfile(c fiber.Ctx) error {
	database := db.DB

	email := c.Locals("email").(string)
	var userobj types.USEROBJECT

	err := database.QueryRow("SELECT id, name, email FROM users WHERE email = ?",
		email,
	).Scan(
		&userobj.Id,
		&userobj.Name,
		&userobj.Email,
	)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"name":  userobj.Name,
		"email": userobj.Email,
	})
}
