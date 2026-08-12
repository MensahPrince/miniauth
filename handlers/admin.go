package handlers

import (
	"github.com/MensahPrince/miniauth/db"
	"github.com/MensahPrince/miniauth/types"
	"github.com/MensahPrince/miniauth/utils"
	"github.com/gofiber/fiber/v3"
)

// AdminCreateUser allows an admin to create a new user with a specific role
func AdminCreateUser(c fiber.Ctx) error {
	var database = db.DB
	var req types.AdminCreateRequest

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid Request",
		})
	}

	if req.Role == "" {
		req.Role = "user"
	}

	hashedPassphrase, err := utils.BcryptHash(req.Password)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to Hash Password",
		})
	}

	_, err = database.Exec(
		"INSERT INTO users (name, email, password, role) VALUES (?,?,?,?)",
		req.Name,
		req.Email,
		hashedPassphrase,
		req.Role,
	)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	adminEmail := c.Locals("email").(string)
	_, _ = database.Exec(
		"INSERT INTO logs (user_email, action) VALUES (?, ?)",
		adminEmail,
		"Created user account: " + req.Email + " with role: " + req.Role,
	)

	return c.Status(201).JSON(fiber.Map{
		"message": "User created successfully",
		"name":    req.Name,
		"email":   req.Email,
		"role":    req.Role,
	})
}

// AdminRevokeUser allows an admin to delete a user account by ID
func AdminRevokeUser(c fiber.Ctx) error {
	database := db.DB
	id := c.Params("id")

	if id == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "User ID is required",
		})
	}

	var targetEmail string
	err := database.QueryRow("SELECT email FROM users WHERE id = ?", id).Scan(&targetEmail)
	if err != nil {
		targetEmail = "ID " + id
	}

	result, err := database.Exec("DELETE FROM users WHERE id = ?", id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to delete user",
		})
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return c.Status(404).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	adminEmail := c.Locals("email").(string)
	_, _ = database.Exec(
		"INSERT INTO logs (user_email, action) VALUES (?, ?)",
		adminEmail,
		"Deleted user account: " + targetEmail,
	)

	return c.Status(200).JSON(fiber.Map{
		"message": "User deleted successfully",
	})
}

// AdminResetPassword allows an admin to reset a user's password without OTP
func AdminResetPassword(c fiber.Ctx) error {
	database := db.DB
	id := c.Params("id")

	if id == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "User ID is required",
		})
	}

	var req struct {
		Password string `json:"password"`
	}

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid Request"})
	}

	hashed, err := utils.BcryptHash(req.Password)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to hash password"})
	}

	var targetEmail string
	queryErr := database.QueryRow("SELECT email FROM users WHERE id = ?", id).Scan(&targetEmail)
	if queryErr != nil {
		targetEmail = "ID " + id
	}

	result, err := database.Exec("UPDATE users SET password = ? WHERE id = ?", hashed, id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update password"})
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return c.Status(404).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	adminEmail := c.Locals("email").(string)
	_, _ = database.Exec(
		"INSERT INTO logs (user_email, action) VALUES (?, ?)",
		adminEmail,
		"Reset password for user: " + targetEmail,
	)

	return c.Status(200).JSON(fiber.Map{"message": "Password reset successfully"})
}
