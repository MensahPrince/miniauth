package auth

import (
	"os"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

//A simple JWT Generator generates a jwt token with empty data. {} . As such it is necessary to include claims: These are metadata about a session. or basically metadata to anything that the jwt token is generated for. In this case, jwt is defined below.
// So I used NewWithClaims()

func GenerateSymmetricJWT(userId int, email string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userId,
		"email":   email,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}

	jwtKey := os.Getenv("JWT_KEY")
	if jwtKey == "" {
		return "", fiber.NewError(fiber.StatusInternalServerError, "JWT_KEY not set")
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	return token.SignedString([]byte(jwtKey)) // correct
}
