# Mini Auth

Mini Auth is a small authentication service built with Go, Fiber v3, JWT, and bcrypt. It ships as an importable library: you call `miniauth.Init(cfg, app)` from your own Fiber application, and it connects to a database, migrates in the schema SQL *you* provide, seeds a default admin account, and mounts a set of auth/profile/admin routes onto your app.

Mini Auth does not ship any schema of its own — it only migrates whatever SQL you attach via `Config.SchemaPath`/`SchemaSQL` (and, optionally, `PostInitSQL`). This keeps the library from baking in opinions about your tables; you own the schema, `Init` just applies it. See **Required schema** below for the tables/columns the built-in routes expect.

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
- `init.go` — `miniauth.Init`: connects to the DB, migrates in the caller-supplied schema, seeds a default admin, wires up CORS, and registers routes
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

Mini Auth does not read any `DB_*` environment variables itself. **You** open the database connection details (driver name + DSN) and pass them in through `miniauth.Config`; `Init` does the rest (connect, ping, configure the pool, migrate in the schema SQL you attach, seed a default admin).

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
    DBDriver:   "sqlite", // or "mysql"
    DBSource:   "file:auth.db?cache=shared&mode=rwc",
    JWTKey:     os.Getenv("JWT_KEY"),
    SchemaPath: "./schema.sql", // your schema — see "Required schema" below
}

if err := miniauth.Init(cfg, app); err != nil {
    log.Fatalf("failed to initialize miniauth: %v", err)
}
```

`Init` will:

1. Open the connection with `db.Connect` and `Ping` it.
2. Set pool limits appropriate to the driver — SQLite is capped at a single open connection (`SetMaxOpenConns(1)`) because it only supports one writer at a time; MySQL gets a normal pool (10 open/idle conns).
3. Migrate in your schema. `Init` ships no schema of its own — set `SchemaPath` (path to a `.sql` file) or `SchemaSQL` (a raw SQL string) on `Config` to supply it (see **Required schema** below for what the built-in routes expect). The SQL is split on `;` and executed statement-by-statement, so multi-statement scripts work regardless of driver-specific multi-statement DSN flags.
4. Run a best-effort `ALTER TABLE users ADD COLUMN role ...` so pre-existing databases from older versions of this schema pick up the `role` column (the error is ignored if the column already exists).
5. Seed a default admin user (see **Default admin account** below) if the `users` table is empty.
6. Run `Config.PostInitSQL`, if set — a raw SQL string executed after the schema/migration/seed steps above, for app-specific tables, indexes, or seed data (see **Attaching additional custom SQL** below).
7. Enable permissive CORS (`AllowOrigins: ["*"]`).
8. Register all routes from `routes.go` onto your Fiber app.

### Required schema

`Init` does not create any tables on its own — if you skip `SchemaPath`/`SchemaSQL` entirely, `schemaToRun` stays empty and no migration runs at all. The built-in handlers assume the following tables/columns exist by the time `Init` finishes, so your `SchemaPath`/`SchemaSQL` needs to create them:

- `users` — `id`, `name`, `email` (unique), `password` (bcrypt hash), `role` (defaults to `"user"`), `created_at`. Required by registration, login, profile, admin user management, and the default-admin seed step.
- `logs` — `id`, `user_email`, `action`, `created_at`. Required by registration, admin actions, patient creation, and the seed step's audit entry.
- `patients` — `id`, `first_name`, `surname`, `phone`, `email`, `appointment_date`, `notes`, `created_by`, `created_at`. Only required if you mount/use the example `POST/GET /patients` routes.

Example `schema.sql` (SQLite) that satisfies all of the above — see the tutorials below for the MySQL equivalent:

```sql
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    role TEXT DEFAULT 'user',
    created_at TEXT DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE TABLE IF NOT EXISTS logs (
    id INTEGER PRIMARY KEY,
    user_email TEXT NOT NULL,
    action TEXT NOT NULL,
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
```

If the `users` table doesn't exist by the time `Init` reaches the seed step, that step's `SELECT COUNT(*) FROM users` simply errors and is skipped — `Init` still returns successfully, but no admin account is seeded and the auth routes will fail against a missing table until you supply a schema that creates it.

### Default admin account

**Important:** if the `users` table is empty after the schema runs, `Init` automatically inserts a default admin:

- email: `admin@watchdog.local`
- password: `admin123`
- role: `admin`

Change this password (via `POST /admin/users/:id/reset` or directly in the database) immediately in any environment that isn't purely local/dev, since the credentials are hard-coded in `init.go`.

### Attaching additional custom SQL

`SchemaPath`/`SchemaSQL` is where your core schema (including the required tables above) belongs. If you also want extra tables, indexes, or seed rows applied *after* that core schema, the `role`-column migration, and the default-admin seed have all run, set `Config.PostInitSQL` to a second raw SQL string:

```go
cfg := miniauth.Config{
    DBDriver: "sqlite",
    DBSource: "file:auth.db?cache=shared&mode=rwc",
    JWTKey:   os.Getenv("JWT_KEY"),
    PostInitSQL: `
        CREATE TABLE IF NOT EXISTS links (
            id INTEGER PRIMARY KEY,
            short_code TEXT NOT NULL UNIQUE,
            target_url TEXT NOT NULL,
            created_by TEXT,
            created_at TEXT DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
        );
    `,
}
```

Like `SchemaSQL`/`SchemaPath`, `PostInitSQL` is split on `;` and executed one statement at a time, so multiple `CREATE TABLE`/`INSERT`/etc. statements in one string work without needing a driver-specific multi-statement DSN flag.

### Writing your own queries against your own tables

Mini Auth does not ship a query builder, repository layer, or generic "run arbitrary SQL" utility for your app's own tables — it only executes the fixed queries its own handlers need (`users`, `logs`, and `patients` if you use those routes). If your app has additional tables beyond what `PostInitSQL` creates, you write and run those queries yourself, using the exact same connection Mini Auth opened:

```go
import "github.com/MensahPrince/miniauth/db"

rows, err := db.DB.Query("SELECT * FROM links WHERE created_by = ?", email)
```

`db.DB` is an exported `*sql.DB` (plain `database/sql`, no ORM), so there's no second connection pool to manage — after `miniauth.Init(cfg, app)` runs, `db.DB` is ready for both Mini Auth's handlers and your own code to use.

This split is deliberate and also a security boundary: every table name Mini Auth's handlers touch is a literal string in its Go source, never derived from a request body, query param, or header, and every value passed to those queries is bound via `?` placeholders rather than concatenated into the SQL text. That means a client calling the HTTP API can never make Mini Auth write to a table of its choosing — the only way SQL text is attacker-shaped at all is if *you* build `SchemaSQL`/`SchemaPath`/`PostInitSQL` dynamically from untrusted input at startup, which you shouldn't do. Keep the same discipline (fixed table names, parameterized values) in whatever query utility you write for your own tables.

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

3. Create `schema.sql` (see **Required schema** above for why each column is there):

   ```sql
   CREATE TABLE IF NOT EXISTS users (
       id INTEGER PRIMARY KEY,
       name TEXT NOT NULL,
       email TEXT NOT NULL UNIQUE,
       password TEXT NOT NULL,
       role TEXT DEFAULT 'user',
       created_at TEXT DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
   );

   CREATE TABLE IF NOT EXISTS logs (
       id INTEGER PRIMARY KEY,
       user_email TEXT NOT NULL,
       action TEXT NOT NULL,
       created_at TEXT DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
   );
   ```

4. Create `main.go`, pointing `SchemaPath` at that file:

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
           DBDriver:   "sqlite",
           DBSource:   "file:auth.db?cache=shared&mode=rwc",
           JWTKey:     os.Getenv("JWT_KEY"),
           SchemaPath: "./schema.sql",
       }

       if err := miniauth.Init(cfg, app); err != nil {
           log.Fatalf("failed to initialize miniauth: %v", err)
       }

       log.Fatal(app.Listen(":3000"))
   }
   ```

5. Run it:

   ```bash
   go run main.go
   ```

   On first run this creates `auth.db` in the current directory, migrates in `schema.sql` (`users`, `logs`), and seeds the default admin account (`admin@watchdog.local` / `admin123`).

6. Verify the connection came up:

   ```bash
   curl http://localhost:3000/
   ```

   The response includes the output of `utils.CheckDB()` — `"Connected"` means `db.DB` opened and pinged successfully.

7. Log in as the seeded admin to confirm the DB round-trips real queries:

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

3. Create `schema.sql` (MySQL equivalent of the SQLite one above):

   ```sql
   CREATE TABLE IF NOT EXISTS users (
       id INT AUTO_INCREMENT PRIMARY KEY,
       name VARCHAR(255) NOT NULL,
       email VARCHAR(255) NOT NULL UNIQUE,
       password VARCHAR(255) NOT NULL,
       role VARCHAR(20) DEFAULT 'user',
       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
   );

   CREATE TABLE IF NOT EXISTS logs (
       id INT AUTO_INCREMENT PRIMARY KEY,
       user_email VARCHAR(100) NOT NULL,
       action VARCHAR(255) NOT NULL,
       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
   );
   ```

4. Build the DSN from those variables and pass `DBDriver: "mysql"` plus `SchemaPath` in `Config`:

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
           DBDriver:   "mysql",
           DBSource:   dsn,
           JWTKey:     os.Getenv("JWT_KEY"),
           SchemaPath: "./schema.sql",
       }

       if err := miniauth.Init(cfg, app); err != nil {
           log.Fatalf("failed to initialize miniauth: %v", err)
       }

       log.Fatal(app.Listen(":3000"))
   }
   ```

5. Run it and verify, same as the SQLite tutorial:

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
- **`no such table: users` / `Table 'x.users' doesn't exist`** — `Init` doesn't create this table on its own; you must supply a `SchemaPath` or `SchemaSQL` that creates it (see **Required schema** above).

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

Require authentication (any logged-in user, not just admins). Create or list example patient records; each row records `created_by` (the creator's email) and is logged to the `logs` table. Requires a `patients` table (see **Required schema** above) — these routes are always mounted, so include `patients` in your schema if you use them.

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
		DBDriver:   "sqlite", // or "mysql" — see "Connecting it to a database" above
		DBSource:   "file:auth.db?cache=shared&mode=rwc", // DSN for your DB
		JWTKey:     os.Getenv("JWT_KEY"),
		SchemaPath: "./schema.sql", // Required: mini_auth ships no schema of its own — see "Required schema" above
		// SchemaSQL:   "CREATE TABLE ...", // Alternative to SchemaPath: a raw SQL string instead of a file
		// PostInitSQL: "CREATE TABLE ...", // Optional: extra SQL run after the schema/migration/seed steps
	}

	// 3. Initialize miniauth (this connects to the DB, migrates in your schema, and mounts auth routes)
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
- `Init` migrates in whatever SQL you attach and nothing more — it ships no default schema. You must supply `users`, `logs`, and (if you use those routes) `patients` via `SchemaPath`/`SchemaSQL` (see **Required schema** above); extra app-specific tables can be layered on top via `PostInitSQL` (see **Attaching additional custom SQL** above).
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
