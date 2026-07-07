package handlers

import (
	"github.com/MensahPrince/miniauth/db"
	"github.com/MensahPrince/miniauth/types"
	"github.com/MensahPrince/miniauth/utils"
	"github.com/gofiber/fiber/v3"
)

func ResetPassword(c fiber.Ctx) error {
	database := db.DB
	email := c.Locals("email").(string)

	var req types.REQUEST
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Invalid Request"})
	}

	if utils.OTPStore[email] != req.OTP {
		return c.Status(400).JSON(fiber.Map{"message": "Invalid OTP"})
	}

	hashed, err := utils.BcryptHash(req.Password)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to hash password"})
	}

	_, err = database.Exec("UPDATE users SET password = ? WHERE email = ?", hashed, email)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to update password"})
	}

	delete(utils.OTPStore, email)

	return c.Status(200).JSON(fiber.Map{"message": "success"})
}
