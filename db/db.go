package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"   // driver: "mysql"
	_ "github.com/mattn/go-sqlite3"      // driver: "sqlite3"
	// add more blank imports here as needed, e.g.:
	// _ "github.com/lib/pq"             // driver: "postgres"
)

var DB *sql.DB

// Connect opens a database connection for the given driver and source.
// driver must match a registered database/sql driver name, e.g. "sqlite3", "mysql", "postgres".
func Connect(driver, source string) error {
	var err error
	DB, err = sql.Open(driver, source)
	if err != nil {
		return fmt.Errorf("db: failed to open %s connection: %w", driver, err)
	}

	if err := DB.Ping(); err != nil {
		return fmt.Errorf("db: failed to ping %s database: %w", driver, err)
	}

	configurePool(driver)

	return nil
}

// configurePool sets connection pool limits appropriate to the driver.
// SQLite is file-based and only supports one writer at a time, so it
// needs a much smaller pool than a real client-server DB like MySQL/Postgres.
func configurePool(driver string) {
	switch driver {
	case "sqlite3":
		DB.SetMaxOpenConns(1)
		DB.SetConnMaxLifetime(time.Minute * 3)
	default: // mysql, postgres, etc. — normal networked DB pool sizing
		DB.SetMaxOpenConns(10)
		DB.SetMaxIdleConns(10)
		DB.SetConnMaxLifetime(time.Minute * 3)
	}
}