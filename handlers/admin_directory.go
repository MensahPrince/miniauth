package handlers

import (
	"github.com/MensahPrince/miniauth/db"
	"github.com/gofiber/fiber/v3"
)

func GetUsers(c fiber.Ctx) error {
	database := db.DB

	rows, err := database.Query("SELECT id, name, email, role, created_at FROM users ORDER BY id DESC")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to retrieve users: " + err.Error(),
		})
	}
	defer rows.Close()

	var users []fiber.Map
	for rows.Next() {
		var id int
		var name, email, role, createdAt string
		err := rows.Scan(&id, &name, &email, &role, &createdAt)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to scan user: " + err.Error(),
			})
		}
		users = append(users, fiber.Map{
			"id":         id,
			"name":       name,
			"email":      email,
			"role":       role,
			"created_at": createdAt,
		})
	}

	return c.Status(200).JSON(users)
}

func GetLogs(c fiber.Ctx) error {
	database := db.DB

	rows, err := database.Query("SELECT id, user_email, action, created_at FROM logs ORDER BY id DESC")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to retrieve logs: " + err.Error(),
		})
	}
	defer rows.Close()

	var logs []fiber.Map
	for rows.Next() {
		var id int
		var userEmail, action, createdAt string
		err := rows.Scan(&id, &userEmail, &action, &createdAt)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to scan log: " + err.Error(),
			})
		}
		logs = append(logs, fiber.Map{
			"id":         id,
			"user_email": userEmail,
			"action":     action,
			"created_at": createdAt,
		})
	}

	return c.Status(200).JSON(logs)
}
