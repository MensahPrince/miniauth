package utils

import (
	"math/rand/v2"

	"github.com/gofiber/fiber/v3"
)

func GenerateOTP(c fiber.Ctx) int {
	min, max := 1000, 9999
	rangeInt := min + rand.IntN(max-min)

	return rangeInt
}
