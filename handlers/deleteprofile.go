package handlers

import (
	"github.com/MensahPrince/miniauth/db"
	"github.com/MensahPrince/miniauth/types"
	"github.com/MensahPrince/miniauth/utils"
	"github.com/gofiber/fiber/v3"
)

func DeleteAccount(c fiber.Ctx) error {

	database := db.DB

	var req types.USER
	var user types.USERLOGIN

	var email = c.Locals("email")

	if err := c.Bind().Body(&user); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"message": "Invalid Request",
		})
	}

	err := database.QueryRow("SELECT password FROM users WHERE email = ?",
		email).Scan(
		&req.Password,
	)
	if err != nil {
		return c.JSON(fiber.Map{
			"error": "Not found",
		})
	}

	//Compare Passwords

	match := utils.BcryptCompareHash(user.Password, req.Password) // stored hash, plain input
	if !match {
		return c.Status(400).JSON(fiber.Map{
			"error": "Wrong credentials",
		})
	}

	_, err = database.Exec("DELETE FROM users WHERE email = ?",
		email)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Not found",
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"message": "success",
	})
}
