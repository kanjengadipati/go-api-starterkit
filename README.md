# Go Auth App

Authentication API built with Go, Gin, GORM, PostgreSQL, and JWT.

This project now uses a modular structure under [`modules/`](/Users/meilanasapta/Code/go-auth-app/modules), with the main modules:
- `auth`
- `user`
- `role`
- `permission`
- `token`
- `social`

## Features

- user registration
- login with access token and refresh token
- logout
- profile endpoint
- email verification
- resend verification email
- forgot password and reset password
- social login
- admin user listing and deletion
- database migration and seeding
- Docker-based local development

## Tech Stack

- Go
- Gin
- GORM
- PostgreSQL
- JWT
- SendGrid
- golang-migrate

## Project Structure

```text
.
├── cmd/
│   ├── migrate/
│   └── seed/
├── config/
├── migrations/
├── modules/
│   ├── auth/
│   ├── permission/
│   ├── role/
│   ├── social/
│   ├── token/
│   └── user/
├── routes/
├── seeds/
├── services/
└── utils/
```

## Environment Variables

Copy `.env.example` to `.env`, then fill in the values.

Common variables used by this project:

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=auth_db
DB_SSLMODE=disable

JWT_SECRET=your-secret

ADMIN_EMAIL=admin@example.com
ADMIN_PASSWORD=supersecret
```

If you use the email flow, make sure the SendGrid variables are also configured based on [`services/email_service.go`](/Users/meilanasapta/Code/go-auth-app/services/email_service.go#L1).

## Running the Application

Start the app:

```bash
go run .
```

The server runs on:

```text
http://localhost:8080
```

## Running with Docker

This project is Docker-ready and includes:
- application container
- PostgreSQL
- Redis
- Nginx gateway
- automatic database setup container
- basic Nginx rate limiting for `/health` and general API traffic

Start the full stack:

```bash
docker-compose --env-file .env.docker up --build
```

Or use the Makefile shortcuts:

```bash
make docker-up
make docker-down
make docker-logs
make docker-rebuild
```

By default, the gateway is exposed on:

```text
http://localhost
```

## Database Migration

Run migrations with:

```bash
go run ./cmd/migrate
```

Migration files are stored in [`migrations/`](/Users/meilanasapta/Code/go-auth-app/migrations).

## Database Seeding

Run the seeder with:

```bash
go run ./cmd/seed
```

Seeder logic is defined in [`seeds/seed.go`](/Users/meilanasapta/Code/go-auth-app/seeds/seed.go#L1).

## Database Setup Shortcut

To run migration and seed in one step:

```bash
make db-setup
```

This target runs:
- `make migrate-up`
- `make seed`

The Docker setup also uses this shortcut internally through the `db-setup` service in [`docker-compose.yaml`](/Users/meilanasapta/Code/go-auth-app/docker-compose.yaml#L1).

## Running Tests

```bash
go test ./...
```

## Main Endpoints

Auth:
- `POST /auth/register`
- `POST /auth/login`
- `POST /auth/refresh`
- `GET /auth/verify`
- `GET /auth/resend-verification`
- `POST /auth/forgot-password`
- `POST /auth/reset-password`
- `POST /auth/social-login`
- `GET /auth/profile`
- `POST /auth/logout`

Admin:
- `GET /auth/admin/users`
- `DELETE /auth/admin/users/:id`

Health:
- `GET /health`

## Notes

- The main route wiring is assembled in [`main.go`](/Users/meilanasapta/Code/go-auth-app/main.go#L1) and [`routes/route.go`](/Users/meilanasapta/Code/go-auth-app/routes/route.go#L1).
- JWT helpers live in [`services/jwt_service.go`](/Users/meilanasapta/Code/go-auth-app/services/jwt_service.go#L1).
- Shared utilities live in [`utils/`](/Users/meilanasapta/Code/go-auth-app/utils).

## Status

The modular migration is active, and the current repository state passes the test suite.

## TODO

- Use Docker secrets, CI secrets, or platform-managed secrets for serious deployments.
  Examples: GitHub Actions secrets, Railway, Render, Fly.io, or VPS environment injection.
- Separate development, Docker, and production environment configuration more explicitly.
- Rotate any real third-party credentials that were ever stored in local env files.
- Replace placeholder/local secret handling with a proper secret management workflow.
- Consider splitting `/health` and `/ready` for clearer liveness vs readiness checks.
- Consider stronger edge protection for production, such as WAF, CDN, or upstream rate limiting.
