package handlers

import (
	"github.com/MensahPrince/miniauth/db"
	"github.com/MensahPrince/miniauth/types"
	"github.com/MensahPrince/miniauth/utils"
	"github.com/gofiber/fiber/v3"
)

func Register(c fiber.Ctx) error {

	var database = db.DB
	var req types.USER

	//Check for JSON body
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid Request",
		})
	}

	hashedPassphrase, err := utils.BcryptHash(req.Password)

	if err != nil {
		c.SendString("Failed to Hash Password")
	}

	//Write to Database
	_, err = database.Exec(
		"INSERT INTO users (name, email, password) VALUES (?,?,?)",
		req.Name,
		req.Email,
		hashedPassphrase,
	)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message":  "Success",
		"name":     req.Name,
		"email":    req.Email,
		"password": req.Password,
	})
}
