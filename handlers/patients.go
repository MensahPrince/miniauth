package handlers

import (
	"github.com/MensahPrince/miniauth/db"
	"github.com/gofiber/fiber/v3"
)

type PatientRequest struct {
	FirstName       string `json:"first_name"`
	Surname         string `json:"surname"`
	Phone           string `json:"phone"`
	Email           string `json:"email"`
	AppointmentDate string `json:"appointment_date"`
	Notes           string `json:"notes"`
}

func CreatePatient(c fiber.Ctx) error {
	database := db.DB
	creatorEmail := c.Locals("email").(string)

	var req PatientRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request payload",
		})
	}

	if req.FirstName == "" || req.Surname == "" || req.Phone == "" || req.AppointmentDate == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "First name, surname, phone, and appointment date are required",
		})
	}

	_, err := database.Exec(
		"INSERT INTO patients (first_name, surname, phone, email, appointment_date, notes, created_by) VALUES (?, ?, ?, ?, ?, ?, ?)",
		req.FirstName,
		req.Surname,
		req.Phone,
		req.Email,
		req.AppointmentDate,
		req.Notes,
		creatorEmail,
	)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create patient: " + err.Error(),
		})
	}

	// Insert audit log
	actionText := "Registered patient: " + req.FirstName + " " + req.Surname
	_, _ = database.Exec(
		"INSERT INTO logs (user_email, action) VALUES (?, ?)",
		creatorEmail,
		actionText,
	)

	return c.Status(201).JSON(fiber.Map{
		"message": "Patient registered successfully",
	})
}

func GetPatients(c fiber.Ctx) error {
	database := db.DB

	rows, err := database.Query("SELECT id, first_name, surname, phone, email, appointment_date, notes, created_by, created_at FROM patients ORDER BY appointment_date ASC")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to retrieve patients: " + err.Error(),
		})
	}
	defer rows.Close()

	var patients []fiber.Map
	for rows.Next() {
		var id int
		var firstName, surname, phone, email, appointmentDate, notes, createdBy, createdAt string
		err := rows.Scan(&id, &firstName, &surname, &phone, &email, &appointmentDate, &notes, &createdBy, &createdAt)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to scan patient: " + err.Error(),
			})
		}
		patients = append(patients, fiber.Map{
			"id":               id,
			"first_name":       firstName,
			"surname":          surname,
			"phone":            phone,
			"email":            email,
			"appointment_date": appointmentDate,
			"notes":            notes,
			"created_by":       createdBy,
			"created_at":       createdAt,
		})
	}

	return c.Status(200).JSON(patients)
}
