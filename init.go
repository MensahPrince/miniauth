// mini_auth/init.go
package miniauth

import (
	"fmt"

	"github.com/MensahPrince/miniauth/db"
	"github.com/gofiber/fiber/v3"
)

var cfg Config

const schemaSQL = `
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    created_at TEXT DEFAULT (datetime('now'))
);
`

func Init(c Config, app *fiber.App) error {
	cfg = c

	if err := db.Connect(cfg.DBDriver, cfg.DBSource); err != nil {
		return fmt.Errorf("mini_auth: db connect failed: %w", err)
	}

	if _, err := db.DB.Exec(schemaSQL); err != nil {
		return fmt.Errorf("mini_auth: schema init failed: %w", err)
	}

	registerRoutes(app) // was SetupAuthRoutes — just rename/make internal
	return nil
}

func JWTKey() string {
	return cfg.JWTKey
}
