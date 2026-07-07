package miniauth

import (
	"time"
)

type Config struct {
	DBDriver  string        // "sqlite3" or "mysql"
	DBSource  string        // full DSN or file path, already assembled by caller
	JWTKey    string        // caller decides how to source this (env, file, etc.)
	JWTExpiry time.Duration // optional, default 24h if zero
}
