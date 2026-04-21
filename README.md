# Go API Starterkit

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
go run ./cmd/api
```

### Docker

```bash
cp .env.docker.example .env.docker
make docker-up
```

### Test

```bash
make test
# or
go test ./...
```

## Overview

This project provides:
- user registration and login
- access token and refresh token flow
- logout and profile endpoints
- self profile update and password change
- email verification
- forgot password and reset password
- Google, Facebook, and Apple social login
- admin user management
- audit trail for important auth and user actions
- permission-based authorization for admin actions
- basic auth endpoint rate limiting and security headers
- database migration and seeding
- local Docker workflow
- generic PostgreSQL-based deployment support

## Tech Stack

- Go
- Gin
- GORM
- PostgreSQL
- JWT
- SendGrid
- golang-migrate
- Docker

## Project Metadata

- License: [MIT](LICENSE)
- Contributing guide: [CONTRIBUTING.md](CONTRIBUTING.md)
- Security policy: [SECURITY.md](SECURITY.md)
- Code of conduct: [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)
- CI workflow: [ci.yml](.github/workflows/ci.yml)

## Project Structure

```text
.
├── cmd/              # API, migration, and seed entrypoints
├── docs/             # OpenAPI documentation
├── internal/         # application-only packages
│   ├── appsetup/     # app bootstrap and route registration
│   ├── config/       # env, db, and app config
│   ├── httpx/        # response helpers
│   ├── middleware/   # shared HTTP middleware
│   ├── modules/      # modular business domains
│   ├── seeds/        # seed logic
│   └── services/     # shared services (jwt, email)
├── migrations/       # SQL migrations
├── postman/          # manual API testing assets
├── tests/            # tests and mocks
└── main.go           # compatibility entrypoint
```

## Requirements

- Go
- PostgreSQL
- Docker and Docker Compose (or Colima), if you use the container workflow

## Environment Configuration

Copy one of the example files depending on your workflow:

- local development: [`.env.example`](.env.example)
- Docker: [`.env.docker.example`](.env.docker.example)

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
GOOGLE_CLIENT_ID=
FACEBOOK_APP_ID=
FACEBOOK_APP_SECRET=
APPLE_CLIENT_ID=
```

### Notes

- `DATABASE_URL` is the primary database connection setting.
- the app validates critical configuration at startup and exits early when required values are missing or incomplete.
- `APP_BASE_URL` is used for backend-generated links such as email verification.
- `FRONTEND_URL` is used for password reset links when you have a separate frontend.
- `GOOGLE_CLIENT_ID` is optional, but recommended so Google token validation checks the audience claim.
- `FACEBOOK_APP_ID` and `FACEBOOK_APP_SECRET` are required for Facebook social login.
- `APPLE_CLIENT_ID` is required for Sign in with Apple token validation.
- `AUTO_RUN_MIGRATIONS` and `AUTO_RUN_SEEDS` are optional flags if you intentionally want schema setup at app startup.

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
go run ./cmd/api
```

By default the API will be available at:

```text
http://localhost:8080
```

The app respects the `PORT` environment variable automatically.

Compatibility note:

```bash
go run .
```

still works, but `go run ./cmd/api` is now the recommended entrypoint.

## API Conventions

- Authenticated routes require:

```http
Authorization: Bearer <access_token>
```

- Admin routes require an access token that belongs to an admin user.
- Refresh tokens are only valid for `POST /auth/refresh`.
- API responses use a standard envelope:
  - success: `status`, `message`, optional `data`, optional `meta`
  - error: `status`, `message`, optional `errors`
- OpenAPI reference is available at [`docs/openapi.yaml`](docs/openapi.yaml)
- Swagger UI is served by the app at `/docs`

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
  "status": "success",
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
  "status": "success",
  "message": "Login success",
  "data": {
    "access_token": "<jwt>",
    "refresh_token": "<jwt>"
  }
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
  "status": "success",
  "message": "Profile fetched",
  "data": {
    "id": 1,
    "name": "Tester",
    "email": "tester@example.com",
    "role": "user"
  }
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
  }' | jq -r '.data.access_token')
```

### Profile

```bash
curl -X GET "$BASE_URL/auth/profile" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### Update Profile

```bash
curl -X PATCH "$BASE_URL/auth/profile" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Tester Updated"
  }'
```

### Change Password

```bash
curl -X PATCH "$BASE_URL/auth/change-password" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "current_password": "secret123",
    "new_password": "newsecret123"
  }'
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
  }' | jq -r '.data.refresh_token')
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

### Sessions

List active sessions:

```bash
curl -X GET "$BASE_URL/auth/sessions" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "X-Device-ID: web"
```

Revoke one session:

```bash
curl -X DELETE "$BASE_URL/auth/sessions/1" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

Revoke every session:

```bash
curl -X POST "$BASE_URL/auth/logout-all" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

Revoke every other session except the current device:

```bash
curl -X POST "$BASE_URL/auth/logout-others" \
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

Supported providers:
- `google`
- `facebook`
- `apple`

Notes:
- Google and Apple expect an ID token from the provider.
- Facebook expects a user access token, but the API accepts the same `token` field for consistency.
- The starterkit requires an email from the provider so it can map or create the local user safely.

```bash
curl -X POST "$BASE_URL/auth/social-login" \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "google",
    "token": "<provider-token>"
  }'
```

### Admin: Get Users

```bash
curl -X GET "$BASE_URL/auth/admin/users?page=1&limit=10" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### Admin: Get User By ID

```bash
curl -X GET "$BASE_URL/auth/admin/users/1" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### Admin: Create User

```bash
curl -X POST "$BASE_URL/auth/admin/users" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Managed User",
    "email": "managed@example.com",
    "password": "secret123",
    "role": "user",
    "is_verified": true
  }'
```

### Admin: Update User

```bash
curl -X PUT "$BASE_URL/auth/admin/users/1" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Managed User Updated",
    "email": "managed@example.com",
    "role": "admin",
    "is_verified": true
  }'
```

### Admin: Delete User

```bash
curl -X DELETE "$BASE_URL/auth/admin/users/1" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### Admin: Get Audit Logs

```bash
curl -X GET "$BASE_URL/auth/admin/audit-logs?page=1&limit=10&resource=user" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### Admin: Get Roles

```bash
curl -X GET "$BASE_URL/auth/admin/roles" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### Admin: Get Permissions

```bash
curl -X GET "$BASE_URL/auth/admin/permissions" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### Admin: Get Role Permissions

```bash
curl -X GET "$BASE_URL/auth/admin/roles/2/permissions" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### Admin: Update Role Permissions

```bash
curl -X PUT "$BASE_URL/auth/admin/roles/2/permissions" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "permissions": [
      "user.read_all",
      "user.read",
      "permission.read",
      "role.read",
      "role.update_permissions"
    ]
  }'
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
make help
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
make help
make migrate-up
make migrate-down
make migrate-status
make migrate-create NAME=create_example_table
make seed
make test
make fmt
```

`make migrate-*` uses `DATABASE_URL` when it is present.
`make docker-*` uses `DOCKER_ENV_FILE=.env.docker` by default.

## Deployment Notes

This starterkit is designed to stay platform-agnostic.

Recommended production approach:
- provide a PostgreSQL-compatible `DATABASE_URL`
- run migrations before serving traffic
- run seed data only when you intentionally need initial roles, permissions, or admin users
- inject secrets through your deployment platform instead of committing real env files

Example build and start commands for generic Go platforms:

```bash
go build -tags netgo -ldflags '-s -w' -o app .
./app
```

## Makefile Shortcuts

```bash
make help
make fmt
make test
make check
make migrate-up
make migrate-down
make migrate-down-all
make migrate-status
make migrate-create NAME=create_example_table
make migrate-force VERSION=1
make migrate-drop CONFIRM=1
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
make test
# or
go test ./...
```

### Manual testing with Postman

Files included:
- collection: [`go-api-starterkit.postman_collection.json`](postman/go-api-starterkit.postman_collection.json)
- local environment: [`go-api-starterkit.local.postman_environment.json`](postman/go-api-starterkit.local.postman_environment.json)

Recommended manual flow:
1. `Health`
2. `Register`
3. `Verify Email` or mark the user as verified in the database
4. `Login`
5. `Profile`
6. `Update Profile`
7. `Change Password`
8. `Refresh Token`
9. `Logout`
10. `Admin` requests with an admin token
11. `Get Audit Logs`

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
- `GET /auth/sessions`
- `PATCH /auth/profile`
- `PATCH /auth/change-password`
- `POST /auth/logout`
- `POST /auth/logout-all`
- `POST /auth/logout-others`
- `DELETE /auth/sessions/:id`

### Admin

- `GET /auth/admin/users`
- `GET /auth/admin/users/:id`
- `POST /auth/admin/users`
- `PUT /auth/admin/users/:id`
- `DELETE /auth/admin/users/:id`
- `GET /auth/admin/audit-logs`
- `GET /auth/admin/roles`
- `GET /auth/admin/permissions`
- `GET /auth/admin/roles/:id/permissions`
- `PUT /auth/admin/roles/:id/permissions`

### Health

- `GET /health`

## Current Architecture Notes

- app bootstrap lives in [`internal/appsetup/`](internal/appsetup)
- runtime configuration is centralized in [`internal/config/app.go`](internal/config/app.go)
- auth service logic is split by use case under [`internal/modules/auth/`](internal/modules/auth)
- repository constructors now take explicit DB dependencies instead of relying on global DB state
- admin routes now use permission checks instead of role-only checks for finer authorization control
- the recommended Go entrypoint now lives in [`cmd/api/`](cmd/api)

## Security Notes

- Never commit real secrets to the repository.
- Use secret managers or platform-managed env vars for production deployments.
- Rotate any third-party credentials that were ever exposed locally or in git history.
- Use separate credentials for local, staging, and production environments.
- Sensitive auth endpoints include basic in-memory rate limiting to reduce brute-force and spam attempts.
- The app sets lightweight security headers such as `X-Content-Type-Options` and `X-Frame-Options`.

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
- audit trail foundation for auth and user management
- local and Docker development preparation
- Postman manual testing setup

The current repository state passes the test suite.
