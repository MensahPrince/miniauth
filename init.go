// mini_auth/init.go
package miniauth

import (
	"fmt"
	"os"

	"github.com/MensahPrince/miniauth/db"
	"github.com/MensahPrince/miniauth/utils"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

var cfg Config

const sqliteSchema = `
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    role TEXT DEFAULT 'user',
    created_at TEXT DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE TABLE IF NOT EXISTS patients (
    id INTEGER PRIMARY KEY,
    first_name TEXT NOT NULL,
    surname TEXT NOT NULL,
    phone TEXT NOT NULL,
    email TEXT,
    appointment_date TEXT NOT NULL,
    notes TEXT,
    created_by TEXT,
    created_at TEXT DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE TABLE IF NOT EXISTS logs (
    id INTEGER PRIMARY KEY,
    user_email TEXT NOT NULL,
    action TEXT NOT NULL,
    created_at TEXT DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);
`

const mysqlSchema = `
CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(20) DEFAULT 'user',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS patients (
    id INT AUTO_INCREMENT PRIMARY KEY,
    first_name VARCHAR(100) NOT NULL,
    surname VARCHAR(100) NOT NULL,
    phone VARCHAR(50) NOT NULL,
    email VARCHAR(100),
    appointment_date VARCHAR(50) NOT NULL,
    notes TEXT,
    created_by VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS logs (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_email VARCHAR(100) NOT NULL,
    action VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`

func Init(c Config, app *fiber.App) error {
	cfg = c

	if cfg.JWTKey != "" {
		os.Setenv("JWT_KEY", cfg.JWTKey)
	}

	if err := db.Connect(cfg.DBDriver, cfg.DBSource); err != nil {
		return fmt.Errorf("mini_auth: db connect failed: %w", err)
	}

	var schemaToRun string
	if cfg.SchemaPath != "" {
		b, err := os.ReadFile(cfg.SchemaPath)
		if err != nil {
			return fmt.Errorf("mini_auth: failed to read schema file %s: %w", cfg.SchemaPath, err)
		}
		schemaToRun = string(b)
	} else if cfg.SchemaSQL != "" {
		schemaToRun = cfg.SchemaSQL
	} else {
		switch cfg.DBDriver {
		case "sqlite":
			schemaToRun = sqliteSchema
		case "mysql":
			schemaToRun = mysqlSchema
		}
	}

	if schemaToRun != "" {
		if _, err := db.DB.Exec(schemaToRun); err != nil {
			return fmt.Errorf("mini_auth: schema init failed: %w", err)
		}
	}

	// Auto-migration: try to add the role column in case it's an old database.
	// We ignore the error because it will fail harmlessly if the column already exists.
	if cfg.DBDriver == "sqlite" {
		db.DB.Exec("ALTER TABLE users ADD COLUMN role TEXT DEFAULT 'user'")
	} else if cfg.DBDriver == "mysql" {
		db.DB.Exec("ALTER TABLE users ADD COLUMN role VARCHAR(20) DEFAULT 'user'")
	}

	// Seed a default admin if table is empty
	var count int
	err := db.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err == nil && count == 0 {
		hashed, err := utils.BcryptHash("admin123")
		if err == nil {
			_, _ = db.DB.Exec(
				"INSERT INTO users (name, email, password, role) VALUES (?, ?, ?, ?)",
				"Default Admin",
				"admin@watchdog.local",
				hashed,
				"admin",
			)
			_, _ = db.DB.Exec(
				"INSERT INTO logs (user_email, action) VALUES (?, ?)",
				"system",
				"Seeded default admin account admin@watchdog.local",
			)
		}
	}

	// Enable CORS for frontend requests
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	}))

	registerRoutes(app) // was SetupAuthRoutes — just rename/make internal
	return nil
}

func JWTKey() string {
	return cfg.JWTKey
}
