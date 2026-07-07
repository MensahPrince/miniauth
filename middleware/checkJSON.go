package middleware

import (

	"github.com/MensahPrince/miniauth/types"
	"github.com/gofiber/fiber/v3"
)

func JsonValidator(c fiber.Ctx) error {
	var req types.USER

	//Check for JSON body
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid JSON Request",
		})
	}

	//Store parsed data for next handlers
	c.Locals("body", req)

	//continue to next handler
	return c.Next()
}
