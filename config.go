package miniauth

import (
	"time"
)

type Config struct {
	DBDriver   string        // "sqlite3" or "mysql"
	DBSource   string        // full DSN or file path, already assembled by caller
	JWTKey     string        // caller decides how to source this (env, file, etc.)
	JWTExpiry  time.Duration // optional, default 24h if zero
	SchemaPath string        // Optional: Path to a .sql file to run on startup
	SchemaSQL  string        // Optional: Raw SQL string to run on startup
}
