# Mini Auth

Mini Auth is a small authentication service built with Go, Fiber, MySQL, JWT, and bcrypt. It is intended as a lightweight starter project for handling user registration, login, protected profile access, profile updates, account deletion, and password reset.

## What this project covers

- User registration with password hashing
- User login and JWT-based authentication
- Protected profile retrieval
- Profile editing for name and email
- Account deletion after password confirmation
- OTP-based password reset flow
- MySQL and SQLite-backed persistence with environment-based configuration

## Project structure

- main.go — application entry point and server bootstrap
- handlers/ — request handlers for auth, profile, OTP, and account management
- middleware/authware.go — JWT authentication middleware
- routes/auth_routes.go — API route registration
- db/db.go — MySQL connection setup
- utils/ — password hashing, OTP generation, and supporting helpers
- types/ — request and response payload definitions

## API endpoints

### GET /

Returns a basic health/status response.

### POST /register

Creates a new user account.

Example body:

```json
{
  "name": "Jane Doe",
  "email": "jane@example.com",
  "password": "password123"
}
```

The password is hashed before being stored.

### POST /login

Authenticates a user and returns a JWT.

Example body:

```json
{
  "email": "jane@example.com",
  "password": "password123"
}
```

### GET /profile

Requires a Bearer token in the Authorization header.

### POST /edit/:param

Requires authentication. Use `name` or `email` as the route parameter to update the corresponding field.

### GET /request-otp

Requires authentication. Returns a one-time password value for development/testing purposes.

### POST /reset

Requires authentication. Expects an OTP and a new password. This currently acts more like an authenticated "Change Password" flow.

Example body:

```json
{
  "otp": 1234,
  "password": "newpassword123"
}
```

### POST /delete

Requires authentication. Expects the current password for confirmation.

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

Because `miniauth` uses `os.Getenv` internally for JWT generation and validation, you can either ensure that the `JWT_KEY` environment variable is set before running your application, or simply pass it via the `Config` struct on initialization (which will automatically set the environment variable for you).

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
		DBDriver: "sqlite3", // or "mysql"
		DBSource: "file:auth.db?cache=shared&mode=rwc", // DSN for your DB
		JWTKey:   "super-secret-key",
		// SchemaPath: "./schema.sql", // Optional: Provide a custom schema file
		// SchemaSQL:  "CREATE TABLE ...", // Optional: Provide a raw SQL string
	}

	// 4. Initialize miniauth (this connects to the DB, runs migrations, and mounts auth routes)
	if err := miniauth.Init(cfg, app); err != nil {
		log.Fatalf("failed to initialize miniauth: %v", err)
	}

	// 5. Optionally, use the JWTMiddleware to protect your own application routes
	app.Get("/my-protected-route", middleware.JWTMiddleware, func(c fiber.Ctx) error {
		userID := c.Locals("user_id")
		return c.JSON(fiber.Map{
			"message": "Hello, authenticated user!",
			"user_id": userID,
		})
	})

	// 6. Start the server
	log.Fatal(app.Listen(":3000"))
}
```

## Notes

- This project is a lightweight authentication starter and is not a full production-ready identity platform.
- OTP values are currently returned directly in the response for development convenience.
- The application expects a `users` table in the configured database (MySQL or SQLite). Default schemas are applied automatically on init unless overridden via `SchemaPath` or `SchemaSQL`.

## Future Roadmap / Missing Features

To make this a fully complete production-grade authentication system, the following features are missing and would be logical next steps:

1. **Unauthenticated Forgot Password Flow**: Currently, requesting an OTP and resetting the password require a valid JWT. A true "forgot password" flow requires endpoints that accept an email, send an OTP, and allow resetting the password without being logged in.
2. **Refresh Tokens**: Issuing a short-lived Access Token and a long-lived Refresh Token to maintain user sessions securely.
3. **Logout / Token Revocation**: A mechanism to invalidate JWTs on the server-side before they naturally expire (e.g., using a token blacklist or Redis).
4. **Email Verification**: Requiring users to verify their email address via a link or OTP immediately after registration.
5. **Rate Limiting**: Protecting authentication and OTP endpoints against brute-force attacks.
6. **OAuth2 / Social Logins**: Support for Google, GitHub, Apple, etc.
7. **Role-Based Access Control (RBAC)**: Implementing user roles (e.g., Admin vs User) and permissions.
