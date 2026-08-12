package middleware

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

func JWTMiddleware(c fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return fiber.NewError(fiber.StatusUnauthorized, "missing or invalid Authorization header")
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	jwtKey := os.Getenv("JWT_KEY")

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		// Enforce the signing method matches what you used
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fiber.NewError(fiber.StatusUnauthorized, "unexpected signing method")
		}
		return []byte(jwtKey), nil
	})

	if err != nil || !token.Valid {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid or expired token")
	}

	// Inject claims into context for downstream handlers
	claims := token.Claims.(jwt.MapClaims)
	c.Locals("user_id", claims["user_id"])
	c.Locals("email", claims["email"])
	if role, ok := claims["role"]; ok {
		c.Locals("role", role)
	}

	return c.Next()
}

func AdminMiddleware(c fiber.Ctx) error {
	role := c.Locals("role")
	if role == nil || role.(string) != "admin" {
		return fiber.NewError(fiber.StatusForbidden, "requires admin privileges")
	}
	return c.Next()
}
