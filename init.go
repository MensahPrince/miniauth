// mini_auth/init.go
package miniauth

import (
	"fmt"
	"os"
	"strings"

	"github.com/MensahPrince/miniauth/db"
	"github.com/MensahPrince/miniauth/utils"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

var cfg Config

func Init(c Config, app *fiber.App) error {
	cfg = c

	if cfg.JWTKey != "" {
		os.Setenv("JWT_KEY", cfg.JWTKey)
	}

	if err := db.Connect(cfg.DBDriver, cfg.DBSource); err != nil {
		return fmt.Errorf("mini_auth: db connect failed: %w", err)
	}

	// mini_auth ships no schema of its own — the caller supplies the SQL to
	// run (via SchemaPath or SchemaSQL) and Init just migrates it in.
	var schemaToRun string
	if cfg.SchemaPath != "" {
		b, err := os.ReadFile(cfg.SchemaPath)
		if err != nil {
			return fmt.Errorf("mini_auth: failed to read schema file %s: %w", cfg.SchemaPath, err)
		}
		schemaToRun = string(b)
	} else if cfg.SchemaSQL != "" {
		schemaToRun = cfg.SchemaSQL
	}

	if schemaToRun != "" {
		if err := execStatements(schemaToRun); err != nil {
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

	// Run any caller-supplied SQL after the core schema/migration/seed steps
	// (e.g. app-specific tables, extra indexes, seed data).
	if cfg.PostInitSQL != "" {
		if err := execStatements(cfg.PostInitSQL); err != nil {
			return fmt.Errorf("mini_auth: post-init SQL failed: %w", err)
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

// execStatements splits a ";"-separated SQL script into individual statements
// and executes them one at a time, since the sql.DB drivers used here don't
// support multiple statements per Exec call unless the caller opts into
// driver-specific multi-statement DSN flags.
func execStatements(script string) error {
	for _, stmt := range strings.Split(script, ";") {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := db.DB.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}
