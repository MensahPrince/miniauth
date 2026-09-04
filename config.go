package miniauth

import (
	"time"
)

type Config struct {
	DBDriver    string        // "sqlite3" or "mysql"
	DBSource    string        // full DSN or file path, already assembled by caller
	JWTKey      string        // caller decides how to source this (env, file, etc.)
	JWTExpiry   time.Duration // optional, default 24h if zero
	SchemaPath  string        // Path to a .sql file to migrate in on startup. mini_auth ships no schema of its own — see README for the tables/columns the built-in routes require.
	SchemaSQL   string        // Raw SQL string to migrate in on startup (alternative to SchemaPath)
	PostInitSQL string        // Optional: Extra raw SQL executed after SchemaPath/SchemaSQL and the seed step (e.g. additional app-specific tables)
}
