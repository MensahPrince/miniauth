package handlers

import (
	"github.com/MensahPrince/miniauth/db"
	"github.com/MensahPrince/miniauth/types"
	"github.com/gofiber/fiber/v3"
)

func EditHandler(c fiber.Ctx) error {

	database := db.DB
	email := c.Locals("email").(string)

	var req types.USERDATA
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid Request",
		})
	}

	switch c.Params("param") {
	case "name":
		_, err := database.Exec("UPDATE users SET name = ? WHERE email = ?",
			req.Name,
			email,
		)

		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"message": err.Error(),
			})
		}

	case "email":
		_, err := database.Exec("UPDATE users SET email = ? WHERE email = ?",
			req.Email,
			email,
		)

		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"message": err.Error(),
			})
		}

	default:
		return c.Status(400).JSON(fiber.Map{
			"message": "Invalid Param",
		})

	}

	return c.Status(200).JSON(fiber.Map{
		"message": "success",
		"name":    req.Name + req.Email,
	})

}
