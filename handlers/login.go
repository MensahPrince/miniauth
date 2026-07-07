package handlers

import (
	"github.com/MensahPrince/miniauth/db"
	"github.com/MensahPrince/miniauth/types"
	"github.com/MensahPrince/miniauth/utils"
	"github.com/MensahPrince/miniauth/utils/auth"
	"github.com/gofiber/fiber/v3"
)

func Login(c fiber.Ctx) error {
	database := db.DB

	var req types.USERLOGIN
	var user types.USEROBJECT

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid JSON request",
		})
	}

	err := database.QueryRow(
		"SELECT id, name, email, password FROM users WHERE email = ?",
		req.Email,
	).Scan(
		&user.Id,
		&user.Name,
		&user.Email,
		&user.Password,
	)

	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"message": "Invalid Credentials",
		})
	}

	if !utils.BcryptCompareHash(req.Password, user.Password) {
		return c.Status(400).JSON(fiber.Map{
			"message": "Invalid Credentials",
		})
	}

	token, err := auth.GenerateSymmetricJWT(user.Id, user.Email)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"success": true,
		"jwt":     token,
	})
}
