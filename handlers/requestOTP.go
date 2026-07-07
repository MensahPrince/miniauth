package handlers

import (
	"github.com/MensahPrince/miniauth/utils"
	"github.com/gofiber/fiber/v3"
)

func RequestOTP(c fiber.Ctx) error {
	email := c.Locals("email").(string)
	otp := utils.GenerateOTP(c)
	utils.OTPStore[email] = otp

	return c.Status(200).JSON(fiber.Map{
		"otp":     otp,
		"message": "dev only, will be sent via email/sms in production",
	})
}
