# Go Auth App

![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)
![Gin](https://img.shields.io/badge/Gin-HTTP%20Framework-009688)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-Database-336791?logo=postgresql)
![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker)

Authentication API built with Go, Gin, GORM, PostgreSQL, and JWT.

This repository uses a modular structure centered around:
- `auth`
- `user`
- `role`
- `permission`
- `token`
- `social`

## Quickstart

### Local

```bash
cp .env.example .env
go run ./cmd/migrate
go run ./cmd/seed
go run .
```

### Docker

```bash
cp .env.docker.example .env.docker
make docker-up
```

### Test

```bash
go test ./...
```

## Overview

This project provides:
- user registration and login
- access token and refresh token flow
- logout and profile endpoints
- email verification
- forgot password and reset password
- Google social login
- admin user listing and deletion
- database migration and seeding
- local Docker workflow
- Render deployment support

## Tech Stack

- Go
- Gin
- GORM
- PostgreSQL
- JWT
- SendGrid
- golang-migrate
- Docker

## Project Structure

```text
.
├── appsetup/         # app bootstrap and route registration
├── cmd/              # migration and seed entrypoints
├── config/           # env, db, and app config
├── migrations/       # SQL migrations
├── modules/          # modular business domains
├── postman/          # manual API testing assets
├── routes/           # compatibility route entrypoint
├── seeds/            # seed logic
├── services/         # shared services (jwt, email)
├── tests/            # tests and mocks
└── utils/            # helpers
```

## Requirements

- Go
- PostgreSQL
- Docker and Docker Compose, if you use the container workflow

## Environment Configuration

Copy one of the example files depending on your workflow:

- local development: [`.env.example`](/Users/meilanasapta/Code/go-auth-app/.env.example#L1)
- Docker: [`.env.docker.example`](/Users/meilanasapta/Code/go-auth-app/.env.docker.example#L1)
- Render: [`.env.render.example`](/Users/meilanasapta/Code/go-auth-app/.env.render.example#L1)

### Common Variables

```env
DATABASE_URL=postgresql://postgres:password@localhost:5432/auth_db?sslmode=disable
JWT_SECRET=replace-with-a-strong-secret
APP_BASE_URL=http://localhost:8080
FRONTEND_URL=http://localhost:3000
ADMIN_EMAIL=admin@example.com
ADMIN_PASSWORD=supersecret
SENDGRID_API_KEY=
SENDGRID_EMAIL=
```

### Notes

- `DATABASE_URL` is the primary database connection setting.
- `APP_BASE_URL` is used for backend-generated links such as email verification.
- `FRONTEND_URL` is used for password reset links when you have a separate frontend.
- `AUTO_RUN_MIGRATIONS` and `AUTO_RUN_SEEDS` are intended mainly for hosted deployment flows such as Render.

For local development and Docker, keep:

```env
AUTO_RUN_MIGRATIONS=false
AUTO_RUN_SEEDS=false
```

unless you intentionally want startup-time initialization.

## Local Development

### 1. Configure environment

```bash
cp .env.example .env
```

Update the values to match your local PostgreSQL setup.

### 2. Run migrations

```bash
go run ./cmd/migrate
```

### 3. Run seed data

```bash
go run ./cmd/seed
```

### 4. Start the application

```bash
go run .
```

By default the API will be available at:

```text
http://localhost:8080
```

The app respects the `PORT` environment variable automatically.

## API Conventions

- Authenticated routes require:

```http
Authorization: Bearer <access_token>
```

- Admin routes require an access token that belongs to an admin user.
- Refresh tokens are only valid for `POST /auth/refresh`.
- Most auth endpoints return either:
  - a JSON object with tokens
  - or a JSON message/error payload

## Example Requests

### Register

```http
POST /auth/register
Content-Type: application/json

{
  "name": "Tester",
  "email": "tester@example.com",
  "password": "secret123"
}
```

Example response:

```json
{
  "message": "User registered"
}
```

### Login

```http
POST /auth/login
Content-Type: application/json
X-Device-ID: web

{
  "email": "tester@example.com",
  "password": "secret123"
}
```

Example response:

```json
{
  "access_token": "<jwt>",
  "refresh_token": "<jwt>"
}
```

### Profile

```http
GET /auth/profile
Authorization: Bearer <access_token>
```

Example response:

```json
{
  "id": 1,
  "name": "Tester",
  "email": "tester@example.com",
  "role": "user"
}
```

## cURL Examples

Set a base URL first:

```bash
BASE_URL=http://localhost:8080
```

### Health

```bash
curl -X GET "$BASE_URL/health"
```

### Register

```bash
curl -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Tester",
    "email": "tester@example.com",
    "password": "secret123"
  }'
```

### Login

```bash
curl -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -H "X-Device-ID: web" \
  -d '{
    "email": "tester@example.com",
    "password": "secret123"
  }'
```

If you want to store the access token quickly in shell:

```bash
ACCESS_TOKEN=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -H "X-Device-ID: web" \
  -d '{
    "email": "tester@example.com",
    "password": "secret123"
  }' | jq -r '.access_token')
```

### Profile

```bash
curl -X GET "$BASE_URL/auth/profile" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### Refresh Token

Store the refresh token:

```bash
REFRESH_TOKEN=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -H "X-Device-ID: web" \
  -d '{
    "email": "tester@example.com",
    "password": "secret123"
  }' | jq -r '.refresh_token')
```

Then refresh:

```bash
curl -X POST "$BASE_URL/auth/refresh" \
  -H "Content-Type: application/json" \
  -d "{
    \"refresh_token\": \"$REFRESH_TOKEN\"
  }"
```

### Logout

```bash
curl -X POST "$BASE_URL/auth/logout" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "X-Device-ID: web"
```

### Forgot Password

```bash
curl -X POST "$BASE_URL/auth/forgot-password" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "tester@example.com"
  }'
```

### Reset Password

```bash
curl -X POST "$BASE_URL/auth/reset-password" \
  -H "Content-Type: application/json" \
  -d '{
    "token": "<reset-token>",
    "new_password": "newsecret123"
  }'
```

### Verify Email

```bash
curl -X GET "$BASE_URL/auth/verify?token=<verify-token>"
```

### Resend Verification

```bash
curl -X POST "$BASE_URL/auth/resend-verification" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "tester@example.com"
  }'
```

### Social Login

```bash
curl -X POST "$BASE_URL/auth/social-login" \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "google",
    "id_token": "<google-id-token>"
  }'
```

### Admin: Get Users

```bash
curl -X GET "$BASE_URL/auth/admin/users?page=1&limit=10" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### Admin: Delete User

```bash
curl -X DELETE "$BASE_URL/auth/admin/users/1" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

## Docker Workflow

This repository includes:
- application container
- PostgreSQL
- Redis
- Nginx gateway
- `db-setup` container for migration and seed
- basic Nginx rate limiting

### Start the full stack

```bash
docker-compose --env-file .env.docker up --build
```

### Or use the Makefile shortcuts

```bash
make docker-up
make docker-down
make docker-logs
make docker-rebuild
```

By default the gateway is exposed at:

```text
http://localhost
```

## Database Tasks

### Run migrations

```bash
go run ./cmd/migrate
```

### Run seed data

```bash
go run ./cmd/seed
```

### Run both with Makefile

```bash
make db-setup
```

Useful Makefile commands:

```bash
make migrate-up
make migrate-down
make migrate-status
make migrate-create NAME=create_example_table
make seed
```

`make migrate-*` uses `DATABASE_URL` when it is present.

## Deploying to Render with Neon

Recommended Render setup:
- Service type: `Web Service`
- Runtime: `Go`
- Build command: `go build -tags netgo -ldflags '-s -w' -o app .`
- Start command: `./app`
- Health check path: `/health`

Recommended environment variables:

```env
DATABASE_URL=postgresql://<user>:<password>@<your-neon-host>/<db>?sslmode=require
AUTO_RUN_MIGRATIONS=true
AUTO_RUN_SEEDS=true
JWT_SECRET=replace-with-a-long-random-secret
APP_BASE_URL=https://<your-render-service>.onrender.com
FRONTEND_URL=https://<your-frontend-domain>
SENDGRID_API_KEY=...
SENDGRID_EMAIL=...
ADMIN_EMAIL=admin@example.com
ADMIN_PASSWORD=supersecret
```

Notes:
- Prefer Neon direct connection strings first.
- Use `sslmode=require` with Neon.
- Startup auto-run for migrations and seeds is supported through the app bootstrap.
- A Render Blueprint starter is included in [`render.yaml`](/Users/meilanasapta/Code/go-auth-app/render.yaml#L1).

## Makefile Shortcuts

```bash
make migrate-up
make migrate-down
make migrate-down-all
make migrate-status
make migrate-create NAME=create_example_table
make migrate-force VERSION=1
make migrate-drop
make seed
make db-setup
make docker-up
make docker-down
make docker-logs
make docker-rebuild
```

## Testing

### Automated tests

```bash
go test ./...
```

### Manual testing with Postman

Files included:
- collection: [`go-auth-app.postman_collection.json`](/Users/meilanasapta/Code/go-auth-app/postman/go-auth-app.postman_collection.json#L1)
- local environment: [`go-auth-app.local.postman_environment.json`](/Users/meilanasapta/Code/go-auth-app/postman/go-auth-app.local.postman_environment.json#L1)

Recommended manual flow:
1. `Health`
2. `Register`
3. `Verify Email` or mark the user as verified in the database
4. `Login`
5. `Profile`
6. `Refresh Token`
7. `Logout`

Notes:
- `Login` and `Refresh Token` update the Postman environment variables for `access_token` and `refresh_token`.
- `Verify Email` and `Reset Password` require manual token input unless you automate email capture.
- admin endpoints require an admin access token.

## Troubleshooting

### `invalid token` on `/auth/profile`

Usually caused by one of these:
- `access_token` was not stored correctly in Postman
- a `refresh_token` was used instead of an `access_token`
- the token has expired

Recommended check:
1. run `Login`
2. confirm `access_token` exists in the active Postman environment
3. retry `Profile`

### `relation "users" does not exist`

This means migrations have not run yet.

Use one of:

```bash
go run ./cmd/migrate
```

or:

```bash
make db-setup
```

For Render, make sure:

```env
AUTO_RUN_MIGRATIONS=true
AUTO_RUN_SEEDS=true
```

## Main Endpoints

### Auth

- `POST /auth/register`
- `POST /auth/login`
- `POST /auth/refresh`
- `GET /auth/verify`
- `POST /auth/resend-verification`
- `POST /auth/forgot-password`
- `POST /auth/reset-password`
- `POST /auth/social-login`
- `GET /auth/profile`
- `POST /auth/logout`

### Admin

- `GET /auth/admin/users`
- `DELETE /auth/admin/users/:id`

### Health

- `GET /health`

## Current Architecture Notes

- app bootstrap lives in [`appsetup/`](/Users/meilanasapta/Code/go-auth-app/appsetup)
- runtime configuration is centralized in [`config/app.go`](/Users/meilanasapta/Code/go-auth-app/config/app.go#L1)
- auth service logic is split by use case under [`modules/auth/`](/Users/meilanasapta/Code/go-auth-app/modules/auth)
- repository constructors now take explicit DB dependencies instead of relying on global DB state

## Security Notes

- Never commit real secrets to the repository.
- Use secret managers or platform-managed env vars for production deployments.
- Rotate any third-party credentials that were ever exposed locally or in git history.
- Use separate credentials for local, staging, and production environments.

## Roadmap Ideas

- standardize all API responses under one response envelope
- add database-backed integration tests
- split readiness and liveness probes
- further reduce infrastructure-specific behavior inside app startup
- add CI validation for migration and deployment smoke checks

## Status

The codebase has gone through:
- modularization cleanup
- auth hardening
- bootstrap/config standardization
- local, Docker, and Render deployment preparation
- Postman manual testing setup

The current repository state passes the test suite.
