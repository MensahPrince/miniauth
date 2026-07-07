# Mini Auth

Mini Auth is a small authentication service built with Go, Fiber, MySQL, JWT, and bcrypt. It is intended as a lightweight starter project for handling user registration, login, protected profile access, profile updates, account deletion, and password reset.

## What this project covers

- User registration with password hashing
- User login and JWT-based authentication
- Protected profile retrieval
- Profile editing for name and email
- Account deletion after password confirmation
- OTP-based password reset flow
- MySQL-backed persistence with environment-based configuration

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

Requires authentication. Expects an OTP and a new password.

Example body:

```json
{
  "otp": 1234,
  "password": "newpassword123"
}
```

### POST /delete

Requires authentication. Expects the current password for confirmation.

## Environment variables

Create a `.env` file with the following values:

- DB_HOST
- DB_USER
- DB_PASS
- DB_NAME
- JWT_KEY

## Running locally

1. Install Go and MySQL.
2. Create a `.env` file with the required values.
3. Run `go mod tidy`.
4. Start the server with `go run main.go`.
5. The app listens on port `3000` by default.

## Notes

- This project is a lightweight authentication starter and is not a full production-ready identity platform.
- OTP values are currently returned directly in the response for development convenience.
- The application expects a `users` table in the configured MySQL database.
