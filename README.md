# Mini Auth

Mini Auth is a small authentication service built with Go, Fiber v3, JWT, and bcrypt. It ships as an importable library: you call `miniauth.Init(cfg, app)` from your own Fiber application, and it connects to a database, applies a schema, seeds a default admin account, and mounts a set of auth/profile/admin routes onto your app.

## What this project covers

- User registration with password hashing (bcrypt)
- User login and JWT-based authentication (HS256, `golang-jwt/jwt/v5`)
- Protected profile retrieval and editing (name/email)
- Account deletion after password confirmation
- OTP-based password reset flow (in-memory OTP store)
- Role-based access control (`user` / `admin`) with admin-only user management endpoints
- A small example `patients` resource showing an authenticated CRUD-style endpoint
- MySQL and SQLite-backed persistence, selected and configured entirely by the caller via `Config`

## Project structure

- `config.go` — the `Config` struct callers fill in and pass to `Init`
- `init.go` — `miniauth.Init`: connects to the DB, runs the schema/migrations, seeds a default admin, wires up CORS, and registers routes
- `routes.go` — API route registration (`registerRoutes`)
- `db/db.go` — driver-agnostic `database/sql` connection setup (MySQL and SQLite drivers are blank-imported here)
- `handlers/` — request handlers for auth, profile, OTP, admin, and the example patients resource
- `middleware/authware.go` — JWT auth middleware (`JWTMiddleware`) and role-check middleware (`AdminMiddleware`)
- `middleware/checkJSON.go` — generic JSON body validator middleware
- `utils/` — password hashing (`bcrypt.go`), OTP generation/storage (`otp.go`, `store.go`), DB health check (`utils.go`)
- `utils/auth/jwt.go` — JWT generation (`GenerateSymmetricJWT`)
- `types/` — request and response payload definitions

This module has no `main.go` of its own — it's a library. The `main.go` shown under Usage below lives in the *consuming* application, not in this repo.

## Connecting it to a database

Mini Auth does not read any `DB_*` environment variables itself. **You** open the database connection details (driver name + DSN) and pass them in through `miniauth.Config`; `Init` does the rest (connect, ping, configure the pool, create tables if needed, seed a default admin).

### 1. Pick a driver

`db.Connect(driver, source)` just calls `sql.Open(driver, source)`, so `driver` must match a driver name registered with `database/sql`. Two are registered out of the box (via blank imports in `db/db.go`):

| `Config.DBDriver` | Backing package | Notes |
| --- | --- | --- |
| `"sqlite"` | `modernc.org/sqlite` | Pure Go, no CGO. **Not** `"sqlite3"` — that's the driver name used by CGO-based sqlite3 packages, which this project doesn't import. |
| `"mysql"` | `github.com/go-sql-driver/mysql` | Standard networked MySQL/MariaDB driver. |

To support another database (e.g. Postgres), blank-import its driver in `db/db.go` (there's a commented example for `lib/pq`) and pass that driver's registered name as `DBDriver`.

### 2. Build a DSN (`Config.DBSource`)

- **SQLite** — a file path or `file:` DSN, e.g.:
  ```go
  DBSource: "file:auth.db?cache=shared&mode=rwc"
  ```
- **MySQL** — a standard `go-sql-driver/mysql` DSN, e.g.:
  ```go
  DBSource: fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?parseTime=true", dbUser, dbPass, dbHost, dbName)
  ```
  `.env.local.example` lists `DB_USER`, `DB_PASS`, `DB_HOST`, and `DB_NAME` as a suggested set of variables for *your* app to read (with `os.Getenv`) and assemble into this DSN yourself — Mini Auth never reads them directly.

### 3. Pass both to `Config` and call `Init`

```go
cfg := miniauth.Config{
    DBDriver: "sqlite", // or "mysql"
    DBSource: "file:auth.db?cache=shared&mode=rwc",
    JWTKey:   os.Getenv("JWT_KEY"),
}

if err := miniauth.Init(cfg, app); err != nil {
    log.Fatalf("failed to initialize miniauth: %v", err)
}
```

`Init` will:

1. Open the connection with `db.Connect` and `Ping` it.
2. Set pool limits appropriate to the driver — SQLite is capped at a single open connection (`SetMaxOpenConns(1)`) because it only supports one writer at a time; MySQL gets a normal pool (10 open/idle conns).
3. Run a schema. By default it uses a built-in schema matching the chosen driver (see below); you can override this with `SchemaPath` (path to a `.sql` file) or `SchemaSQL` (a raw SQL string) on `Config`.
4. Run a best-effort `ALTER TABLE users ADD COLUMN role ...` so pre-existing databases from older versions of this schema pick up the `role` column (the error is ignored if the column already exists).
5. Seed a default admin user (see **Default admin account** below) if the `users` table is empty.
6. Enable permissive CORS (`AllowOrigins: ["*"]`).
7. Register all routes from `routes.go` onto your Fiber app.

### Default schema

If you don't supply `SchemaPath`/`SchemaSQL`, `Init` creates three tables (`users`, `patients`, `logs`) using either `sqliteSchema` or `mysqlSchema` from `init.go`, matched to `Config.DBDriver`:

- `users` — `id`, `name`, `email` (unique), `password` (bcrypt hash), `role` (defaults to `"user"`), `created_at`
- `patients` — a small example resource (`first_name`, `surname`, `phone`, `email`, `appointment_date`, `notes`, `created_by`, `created_at`) exercised by `POST/GET /patients`
- `logs` — a simple audit trail (`user_email`, `action`, `created_at`) written to by registration, admin actions, and patient creation

### Default admin account

**Important:** if the `users` table is empty after the schema runs, `Init` automatically inserts a default admin:

- email: `admin@watchdog.local`
- password: `admin123`
- role: `admin`

Change this password (via `POST /admin/users/:id/reset` or directly in the database) immediately in any environment that isn't purely local/dev, since the credentials are hard-coded in `init.go`.

## Tutorials

Two full walkthroughs, from an empty directory to a running server with a working database connection.

### Tutorial: SQLite quickstart (no external services)

This is the fastest way to try Mini Auth locally — SQLite needs no server, just a file on disk.

1. Create a new project and pull in the dependencies:

   ```bash
   mkdir myapp && cd myapp
   go mod init myapp
   go get github.com/MensahPrince/miniauth
   go get github.com/gofiber/fiber/v3
   ```

2. Set a JWT signing key (any random string works for local dev):

   ```bash
   export JWT_KEY="dev-secret-change-me"
   ```

3. Create `main.go`:

   ```go
   package main

   import (
       "log"
       "os"

       "github.com/MensahPrince/miniauth"
       "github.com/gofiber/fiber/v3"
   )

   func main() {
       app := fiber.New()

       cfg := miniauth.Config{
           DBDriver: "sqlite",
           DBSource: "file:auth.db?cache=shared&mode=rwc",
           JWTKey:   os.Getenv("JWT_KEY"),
       }

       if err := miniauth.Init(cfg, app); err != nil {
           log.Fatalf("failed to initialize miniauth: %v", err)
       }

       log.Fatal(app.Listen(":3000"))
   }
   ```

4. Run it:

   ```bash
   go run main.go
   ```

   On first run this creates `auth.db` in the current directory, applies the default schema (`users`, `patients`, `logs`), and seeds the default admin account (`admin@watchdog.local` / `admin123`).

5. Verify the connection came up:

   ```bash
   curl http://localhost:3000/
   ```

   The response includes the output of `utils.CheckDB()` — `"Connected"` means `db.DB` opened and pinged successfully.

6. Log in as the seeded admin to confirm the DB round-trips real queries:

   ```bash
   curl -X POST http://localhost:3000/login \
     -H "Content-Type: application/json" \
     -d '{"email":"admin@watchdog.local","password":"admin123"}'
   ```

   A JWT in the response means `Init` connected, created tables, and inserted/read rows successfully.

### Tutorial: MySQL setup

Use this when you want a real client-server database instead of a local file.

1. Start a MySQL instance (skip this step if you already have one). This example uses Docker:

   ```bash
   docker run --name miniauth-mysql \
     -e MYSQL_ROOT_PASSWORD=rootpass \
     -e MYSQL_DATABASE=miniauth \
     -e MYSQL_USER=miniauth \
     -e MYSQL_PASSWORD=miniauthpass \
     -p 3306:3306 \
     -d mysql:8
   ```

   Give it a few seconds to finish initializing before continuing.

2. In your project directory, copy `.env.local.example` (from this repo) to `.env.local`, or just export the variables directly, matching what you started MySQL with:

   ```bash
   export DB_USER=miniauth
   export DB_PASS=miniauthpass
   export DB_HOST=127.0.0.1
   export DB_NAME=miniauth
   export JWT_KEY="dev-secret-change-me"
   ```

3. Build the DSN from those variables and pass `DBDriver: "mysql"` in `Config`:

   ```go
   package main

   import (
       "fmt"
       "log"
       "os"

       "github.com/MensahPrince/miniauth"
       "github.com/gofiber/fiber/v3"
   )

   func main() {
       app := fiber.New()

       dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?parseTime=true",
           os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("DB_HOST"), os.Getenv("DB_NAME"))

       cfg := miniauth.Config{
           DBDriver: "mysql",
           DBSource: dsn,
           JWTKey:   os.Getenv("JWT_KEY"),
       }

       if err := miniauth.Init(cfg, app); err != nil {
           log.Fatalf("failed to initialize miniauth: %v", err)
       }

       log.Fatal(app.Listen(":3000"))
   }
   ```

4. Run it and verify, same as the SQLite tutorial:

   ```bash
   go run main.go
   curl http://localhost:3000/
   curl -X POST http://localhost:3000/login \
     -H "Content-Type: application/json" \
     -d '{"email":"admin@watchdog.local","password":"admin123"}'
   ```

### Troubleshooting

- **`db: failed to open sqlite connection` / `unknown driver`** — the driver name in `Config.DBDriver` must exactly match `"sqlite"` or `"mysql"` (not `"sqlite3"` — see the driver table above). A typo here fails before any DSN is even attempted.
- **`db: failed to ping mysql database`** — usually means MySQL isn't reachable yet at `DB_HOST:3306`, or the credentials/database name in the DSN don't match what the server was started with. If you used the Docker command above, confirm the container is up with `docker ps` and that you gave it a few seconds to finish initializing.
- **SQLite "database is locked"** — expected under concurrent writers; `db/db.go` already caps SQLite to `SetMaxOpenConns(1)` to avoid this, but if you're running multiple processes against the same `auth.db` file, switch to MySQL.
- **`GET /` reports `"Database Connection Failed"`** — `miniauth.Init` returned an error before this point would normally be reached, so this really means `db.DB` was never set; check the error returned by `Init` in your own logs rather than relying on this endpoint alone.
- **Login with the default admin fails** — the seed step only runs once, when the `users` table is empty. If you've already got rows in `users` (e.g. reusing an old `auth.db` or MySQL database), the seed is skipped and `admin@watchdog.local` may not exist in that database.

## API endpoints

### GET /

Returns a basic health/status response, including whether the DB connection is up (`utils.CheckDB`).

### POST /register

Creates a new user account with role `user` (there is no public way to self-register as `admin` — use the `/admin/users` endpoint below for that).

Example body:

```json
{
  "name": "Jane Doe",
  "email": "jane@example.com",
  "password": "password123"
}
```

The password is hashed (bcrypt) before being stored. Note: the response currently echoes the plaintext password back in the JSON body — avoid logging or exposing this response in production.

### POST /login

Authenticates a user and returns a JWT containing `user_id`, `email`, and `role` claims (24h expiry).

Example body:

```json
{
  "email": "jane@example.com",
  "password": "password123"
}
```

### GET /profile

Requires `Authorization: Bearer <jwt>`. Returns the caller's `name` and `email`.

### POST /edit/:param

Requires authentication. Use `name` or `email` as the route parameter to update the corresponding field for the logged-in user.

### GET /request-otp

Requires authentication. Generates a 4-digit OTP, stores it in-memory (`utils.OTPStore`, keyed by email), and returns it directly in the response for development/testing purposes (in production this would be sent via email/SMS instead).

### POST /reset

Requires authentication. Expects an OTP (previously obtained from `/request-otp`) and a new password. This currently acts more like an authenticated "change password" flow than a true unauthenticated "forgot password" flow.

Example body:

```json
{
  "otp": 1234,
  "password": "newpassword123"
}
```

### POST /delete

Requires authentication. Expects the current password for confirmation, then deletes the account.

### POST /patients, GET /patients

Require authentication (any logged-in user, not just admins). Create or list example patient records; each row records `created_by` (the creator's email) and is logged to the `logs` table.

### Admin routes (require a valid JWT **and** `role: admin`)

Mounted under `/admin` with `JWTMiddleware` + `AdminMiddleware`:

| Method & path | Description |
| --- | --- |
| `POST /admin/users` | Create a user with an explicit `role` (defaults to `user` if omitted) |
| `DELETE /admin/users/:id` | Delete a user by ID |
| `POST /admin/users/:id/reset` | Reset a user's password without needing an OTP |
| `GET /admin/users` | List all users |
| `GET /admin/logs` | List the audit log |

A JWT's `role` claim is only trusted from tokens Mini Auth itself issued at login — there is no separate re-check against the database on each request, so revoking/changing a user's role takes effect only after their existing token expires or they log in again.

## Installation

To install `miniauth` in your own project, use `go get`:

```bash
go get github.com/MensahPrince/miniauth
```

You will also need Fiber v3:

```bash
go get github.com/gofiber/fiber/v3
```

## Usage

Because `miniauth` uses `os.Getenv("JWT_KEY")` internally (in the JWT middleware and generator), you can either ensure that the `JWT_KEY` environment variable is set before running your application, or simply pass it via the `Config` struct on initialization — `Init` will call `os.Setenv("JWT_KEY", cfg.JWTKey)` for you if `cfg.JWTKey` is non-empty.

Here is an example of how to use `miniauth` in your application:

```go
package main

import (
	"log"
	"os"

	"github.com/MensahPrince/miniauth"
	"github.com/MensahPrince/miniauth/middleware"
	"github.com/gofiber/fiber/v3"
)

func main() {
	// 1. Initialize a new Fiber app
	app := fiber.New()

	// 2. Configure the authentication module
	cfg := miniauth.Config{
		DBDriver: "sqlite", // or "mysql" — see "Connecting it to a database" above
		DBSource: "file:auth.db?cache=shared&mode=rwc", // DSN for your DB
		JWTKey:   os.Getenv("JWT_KEY"),
		// SchemaPath: "./schema.sql", // Optional: Provide a custom schema file
		// SchemaSQL:  "CREATE TABLE ...", // Optional: Provide a raw SQL string
	}

	// 3. Initialize miniauth (this connects to the DB, runs migrations, and mounts auth routes)
	if err := miniauth.Init(cfg, app); err != nil {
		log.Fatalf("failed to initialize miniauth: %v", err)
	}

	// 4. Optionally, use the JWTMiddleware (and AdminMiddleware) to protect your own application routes
	app.Get("/my-protected-route", middleware.JWTMiddleware, func(c fiber.Ctx) error {
		userID := c.Locals("user_id")
		return c.JSON(fiber.Map{
			"message": "Hello, authenticated user!",
			"user_id": userID,
		})
	})

	// 5. Start the server
	log.Fatal(app.Listen(":3000"))
}
```

For MySQL, building `DBSource` from environment variables (matching the shape in `.env.local.example`) typically looks like:

```go
dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?parseTime=true",
	os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("DB_HOST"), os.Getenv("DB_NAME"))

cfg := miniauth.Config{
	DBDriver: "mysql",
	DBSource: dsn,
	JWTKey:   os.Getenv("JWT_KEY"),
}
```

## Notes

- This project is a lightweight authentication starter and is not a full production-ready identity platform.
- OTP values are currently returned directly in the response for development convenience.
- The application expects `users`, `patients`, and `logs` tables in the configured database. Default schemas are applied automatically on init unless overridden via `SchemaPath` or `SchemaSQL`.
- A default admin account (`admin@watchdog.local` / `admin123`) is seeded automatically the first time `Init` runs against an empty `users` table — see **Default admin account** above.
- CORS is currently enabled for all origins (`AllowOrigins: ["*"]`); tighten this before deploying anywhere reachable from the public internet.

## Future Roadmap / Missing Features

To make this a fully complete production-grade authentication system, the following features are missing and would be logical next steps:

1. **Unauthenticated Forgot Password Flow**: Currently, requesting an OTP and resetting the password require a valid JWT. A true "forgot password" flow requires endpoints that accept an email, send an OTP, and allow resetting the password without being logged in.
2. **Refresh Tokens**: Issuing a short-lived Access Token and a long-lived Refresh Token to maintain user sessions securely.
3. **Logout / Token Revocation**: A mechanism to invalidate JWTs on the server-side before they naturally expire (e.g., using a token blacklist or Redis).
4. **Email Verification**: Requiring users to verify their email address via a link or OTP immediately after registration.
5. **Rate Limiting**: Protecting authentication and OTP endpoints against brute-force attacks.
6. **OAuth2 / Social Logins**: Support for Google, GitHub, Apple, etc.
7. **Role-Based Access Control (RBAC) refinements**: Roles exist and are enforced (`user`/`admin`), but there's no re-check of the DB on each request, no `admin`-vs-`user` self-service role changes, and no support for finer-grained permissions.
8. **Configurable JWT expiry**: `Config.JWTExpiry` is declared but not yet wired into `GenerateSymmetricJWT`, which currently hard-codes a 24h expiry.
