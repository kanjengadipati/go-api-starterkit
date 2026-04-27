# Pleco
## Ship your product. Let Pleco handle the auth.

[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://go.dev)
[![Gin](https://img.shields.io/badge/Gin-HTTP%20Framework-009688)](https://gin-gonic.com)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-Database-336791?logo=postgresql)](https://www.postgresql.org)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker)](https://www.docker.com)
[![Modular](https://img.shields.io/badge/Architecture-Modular-6f42c1)](https://github.com/kanjengadipati/pleco-api)
[![AI Powered](https://img.shields.io/badge/AI-Audit%20Investigator-ff6b35?logo=ollama)](https://ollama.com)
[![License](https://img.shields.io/badge/License-MIT-green)](LICENSE)

> **Ship your product. Let Pleco handle the auth.** A modular, production-oriented Go REST API foundation with JWT authentication, social login, RBAC, per-device session management, audit trail, and an AI-powered log investigator — ready to clone, extend, and ship.

Intended for Go backend developers who want a solid, security-conscious auth foundation to build on — without reinventing JWT flows, email verification, social login, or audit logging from scratch.

🔗 **Dashboard Demo:** [gokitdash.vercel.app](https://pleco-console.vercel.app/) &nbsp;|&nbsp; 📖 **API Docs:** [pleco-api.onrender.com/docs](https://pleco-api.onrender.com/docs)

---

## Overview

This project provides a complete authentication and authorization foundation built with Go, Gin, GORM, and PostgreSQL. It is structured around independent, domain-focused modules that can be extended or replaced without touching the rest of the app.

**Core features:**
- User registration and login
- Access token and refresh token flow with token rotation
- Per-device session management — list, revoke, and logout individual sessions
- Self profile update and password change
- Email verification, forgot password, and reset password
- Google, Facebook, and Apple social login (server-side token validation)
- Admin user management
- Audit trail for all important auth and user actions
- Optional AI-powered audit log investigator (Ollama, OpenAI, Gemini, or mock)
- Role-based access control (RBAC) with fine-grained permission checks per route
- Per-route rate limiting with a swappable store abstraction
- Hardened security headers (CSP, HSTS, X-Frame-Options, X-Content-Type-Options)
- Request-scoped structured logging with request ID propagation
- Database migration and seeding via golang-migrate
- Local Docker workflow with Nginx, PostgreSQL, and Redis
- Generic PostgreSQL-based deployment support

---

## Tech Stack

| Layer | Technology |
|---|---|
| Language | Go 1.25+ |
| HTTP Framework | Gin |
| ORM | GORM |
| Database | PostgreSQL |
| Auth | JWT |
| Email | SendGrid |
| Migrations | golang-migrate |
| AI | Ollama / OpenAI / Gemini (optional) |
| Infrastructure | Docker, Nginx, Redis |

---


## Architecture

![Architecture](docs/architecture.svg)

___

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

---

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
│   │   ├── auth/
│   │   ├── user/
│   │   ├── role/
│   │   ├── permission/
│   │   ├── token/
│   │   └── social/
│   ├── seeds/        # seed logic
│   └── services/     # shared services (jwt, email, ai)
├── migrations/       # SQL migrations
├── postman/          # manual API testing assets
└── tests/            # tests and mocks
```

Each module owns its own handler, service, repository, and model — keeping domain logic isolated and easy to navigate.

---

## Environment Configuration

Copy one of the example files depending on your workflow:

- Local development: [`.env.example`](.env.example)
- Docker: [`.env.docker.example`](.env.docker.example)

### Common Variables

```env
DATABASE_URL=postgresql://postgres:password@localhost:5432/auth_db?sslmode=disable
TRUSTED_PROXIES=127.0.0.1,::1,10.0.0.0/8,172.16.0.0/12,192.168.0.0/16
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://127.0.0.1:3000
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
AI_ENABLED=false
AI_PROVIDER=mock
AI_MODEL=mock-model
AI_BASE_URL=
AI_API_KEY=
```

### Notes

- `DATABASE_URL` is the primary database connection setting.
- `TRUSTED_PROXIES` controls which proxy hops are trusted for forwarded client IP handling.
- The app validates critical configuration at startup and exits early when required values are missing.
- `APP_BASE_URL` is used for backend-generated links such as email verification.
- `FRONTEND_URL` is used for password reset links when you have a separate frontend.
- `GOOGLE_CLIENT_ID` is optional but recommended so Google token validation checks the audience claim.
- `FACEBOOK_APP_ID` and `FACEBOOK_APP_SECRET` are required for Facebook social login.
- `APPLE_CLIENT_ID` is required for Sign in with Apple token validation.
- `AI_ENABLED=false` keeps the app fully usable without AI.
- `AI_PROVIDER` supports `mock`, `ollama`, `openai`, and `gemini`.
- `AI_BASE_URL` is only required when `AI_PROVIDER=ollama`.
- `AUTO_RUN_MIGRATIONS` and `AUTO_RUN_SEEDS` are optional flags for startup-time initialization. Keep these `false` for local and Docker workflows — run migrations and seeds manually instead.

---

## AI Audit Log Investigator

This project includes an optional AI-powered audit log workflow for admin users.

**Capabilities:**
- Filter and inspect audit logs from admin endpoints
- Export matching audit logs as CSV
- Investigate a selected log window with AI
- Save and retrieve generated investigation results

**Investigation output is structured into:**
- `summary` — high-level description of what was detected
- `timeline` — ordered sequence of relevant events
- `suspicious_signals` — patterns or anomalies flagged by the model
- `recommendations` — suggested next steps

**Required admin permissions:**
- `audit.read` — list logs, export logs, read saved investigations
- `audit.investigate` — create a new AI investigation

### Setup

For quick local testing without a real model:

```env
AI_ENABLED=true
AI_PROVIDER=mock
AI_MODEL=mock-model
AI_TIMEOUT_SECONDS=30
```

For real local AI with Ollama:

```env
AI_ENABLED=true
AI_PROVIDER=ollama
AI_MODEL=qwen2.5:3b
AI_BASE_URL=http://localhost:11434
AI_TIMEOUT_SECONDS=30
```

Make sure Ollama is running and the model is pulled:

```bash
ollama serve
ollama pull qwen2.5:3b
```

For OpenAI:

```env
AI_ENABLED=true
AI_PROVIDER=openai
AI_MODEL=gpt-4.1-mini
AI_API_KEY=your_openai_api_key
AI_TIMEOUT_SECONDS=30
```

For Gemini:

```env
AI_ENABLED=true
AI_PROVIDER=gemini
AI_MODEL=gemini-2.5-flash
AI_API_KEY=your_gemini_api_key
AI_TIMEOUT_SECONDS=30
```

### Typical Admin Flow

1. Query audit logs with `GET /auth/admin/audit-logs`
2. Narrow the result with filters: `resource`, `status`, `actor_user_id`, `search`, `date_from`, `date_to`
3. Send the same filter scope to `POST /auth/admin/audit-logs/investigate`
4. Review the generated summary and recommendations
5. Re-open saved investigation history from `GET /auth/admin/audit-logs/investigations`

### Investigation Request

```json
POST /auth/admin/audit-logs/investigate

{
  "action": "login",
  "resource": "auth",
  "status": "failed",
  "actor_user_id": 1,
  "search": "invalid credentials",
  "date_from": "2026-04-20T00:00:00Z",
  "date_to": "2026-04-21T00:00:00Z",
  "limit": 50
}
```

### Investigation Response

```json
{
  "summary": "Multiple failed login attempts were clustered in a short time window.",
  "timeline": [
    "2026-04-20T08:00:00Z failed login attempt from 10.0.0.10",
    "2026-04-20T08:03:00Z repeated failure from the same IP"
  ],
  "suspicious_signals": [
    "high number of failed auth events from one IP",
    "repeated attempts against the same resource"
  ],
  "recommendations": [
    "review the source IP",
    "consider temporary blocking or tighter rate limiting"
  ]
}
```

### Notes

- Identical requests from the same admin over the same log snapshot are deduplicated and return the existing saved investigation.
- The server applies a hard cap to the investigation window to avoid overly large prompts.
- Larger windows are compressed into chunk summaries before being sent to the model.
- Creating or reusing an audit investigation is itself recorded in the audit log.

### Common Failures

| Error | Cause | Fix |
|---|---|---|
| `ai investigator is not enabled` | `AI_ENABLED` is still false | Set `AI_ENABLED=true` and restart |
| `ollama is unavailable` | Ollama is not running | Run `ollama serve` |
| `ollama model is not available` | Model not pulled | Run `ollama pull <model>` |
| `openai error: bad api key` | Invalid or missing OpenAI API key | Set `AI_API_KEY` to a valid OpenAI key |
| `gemini error: unsupported model` | Wrong Gemini model name | Use a supported model such as `gemini-2.5-flash` |
| `ai investigation timed out` | Model too slow | Increase `AI_TIMEOUT_SECONDS` or use a smaller model |

---

## Main Endpoints

### Auth

| Method | Endpoint | Description |
|---|---|---|
| POST | `/auth/register` | Register a new user |
| POST | `/auth/login` | Login and receive tokens |
| POST | `/auth/refresh` | Refresh access token |
| GET | `/auth/verify` | Verify email address |
| POST | `/auth/resend-verification` | Resend verification email |
| POST | `/auth/forgot-password` | Request password reset |
| POST | `/auth/reset-password` | Reset password with token |
| POST | `/auth/social-login` | Login via Google, Facebook, or Apple |
| GET | `/auth/profile` | Get current user profile |
| PATCH | `/auth/profile` | Update profile |
| PATCH | `/auth/change-password` | Change password |
| GET | `/auth/sessions` | List active sessions |
| POST | `/auth/logout` | Logout current session |
| POST | `/auth/logout-all` | Logout all sessions |
| POST | `/auth/logout-others` | Logout all other sessions |
| DELETE | `/auth/sessions/:id` | Revoke a specific session |

### Admin

| Method | Endpoint | Description |
|---|---|---|
| GET | `/auth/admin/users` | List users |
| GET | `/auth/admin/users/:id` | Get user by ID |
| POST | `/auth/admin/users` | Create user |
| PUT | `/auth/admin/users/:id` | Update user |
| DELETE | `/auth/admin/users/:id` | Delete user |
| GET | `/auth/admin/audit-logs` | List audit logs |
| GET | `/auth/admin/audit-logs/export` | Export audit logs as CSV |
| POST | `/auth/admin/audit-logs/investigations` | Run AI investigation |
| GET | `/auth/admin/audit-logs/investigations` | List saved investigations |
| GET | `/auth/admin/audit-logs/investigations/:id` | Get investigation detail |
| GET | `/auth/admin/roles` | List roles |
| GET | `/auth/admin/permissions` | List permissions |
| GET | `/auth/admin/roles/:id/permissions` | Get role permissions |
| PUT | `/auth/admin/roles/:id/permissions` | Update role permissions |

### Health

| Method | Endpoint | Description |
|---|---|---|
| GET | `/health` | Health check (legacy) |
| GET | `/health/live` | Liveness probe |
| GET | `/health/ready` | Readiness probe (checks DB) |

---

## API Conventions

- Authenticated routes require `Authorization: Bearer <access_token>`
- Admin routes require an access token that belongs to an admin user
- Refresh tokens are only valid for `POST /auth/refresh`
- Success responses use the envelope: `status`, `message`, optional `data`, optional `meta`
- Error responses use the envelope: `status`, `message`, optional `errors`
- OpenAPI reference: [`docs/openapi.yaml`](docs/openapi.yaml)
- Swagger UI: served at `/docs`

---

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

Response:

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

Response:

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

Response:

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

---

## cURL Examples

Set a base URL first:

```bash
BASE_URL=http://localhost:8080
```

### Health

```bash
curl -X GET "$BASE_URL/health"
curl -X GET "$BASE_URL/health/live"
curl -X GET "$BASE_URL/health/ready"
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

Store tokens for use in subsequent requests:

```bash
TOKENS=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -H "X-Device-ID: web" \
  -d '{
    "email": "tester@example.com",
    "password": "secret123"
  }')

ACCESS_TOKEN=$(echo $TOKENS | jq -r '.data.access_token')
REFRESH_TOKEN=$(echo $TOKENS | jq -r '.data.refresh_token')
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
  -d '{"name": "Tester Updated"}'
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

```bash
curl -X POST "$BASE_URL/auth/refresh" \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\": \"$REFRESH_TOKEN\"}"
```

### Verify Email

```bash
curl -X GET "$BASE_URL/auth/verify?token=<verify-token>"
```

### Resend Verification

```bash
curl -X POST "$BASE_URL/auth/resend-verification" \
  -H "Content-Type: application/json" \
  -d '{"email": "tester@example.com"}'
```

### Forgot Password

```bash
curl -X POST "$BASE_URL/auth/forgot-password" \
  -H "Content-Type: application/json" \
  -d '{"email": "tester@example.com"}'
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

### Social Login

Supported providers: `google`, `facebook`, `apple`.

- Google and Apple expect an ID token.
- Facebook expects a user access token.
- All three use the same `token` field for consistency.
- The starterkit requires an email from the provider to map or create a local user.

```bash
curl -X POST "$BASE_URL/auth/social-login" \
  -H "Content-Type: application/json" \
  -H "X-Device-ID: web" \
  -d '{
    "provider": "google",
    "token": "<provider-token>"
  }'
```

### Sessions

List active sessions:

```bash
curl -X GET "$BASE_URL/auth/sessions" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "X-Device-ID: web"
```

Revoke one session by ID:

```bash
curl -X DELETE "$BASE_URL/auth/sessions/1" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

Logout current session:

```bash
curl -X POST "$BASE_URL/auth/logout" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "X-Device-ID: web"
```

Logout all sessions:

```bash
curl -X POST "$BASE_URL/auth/logout-all" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

Logout every session except the current device:

```bash
curl -X POST "$BASE_URL/auth/logout-others" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "X-Device-ID: web"
```

### Admin: Users

```bash
# List users
curl -X GET "$BASE_URL/auth/admin/users?page=1&limit=10" \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# Get user by ID
curl -X GET "$BASE_URL/auth/admin/users/1" \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# Create user
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

# Update user
curl -X PUT "$BASE_URL/auth/admin/users/1" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Managed User Updated",
    "email": "managed@example.com",
    "role": "admin",
    "is_verified": true
  }'

# Delete user
curl -X DELETE "$BASE_URL/auth/admin/users/1" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### Admin: Roles and Permissions

```bash
# List roles
curl -X GET "$BASE_URL/auth/admin/roles" \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# List permissions
curl -X GET "$BASE_URL/auth/admin/permissions" \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# Get permissions for a role
curl -X GET "$BASE_URL/auth/admin/roles/2/permissions" \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# Update permissions for a role
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

### Admin: Audit Logs

```bash
# List audit logs with filters
curl -X GET "$BASE_URL/auth/admin/audit-logs?page=1&limit=10&resource=auth&status=failed&actor_user_id=1&search=invalid&date_from=2026-04-20T00:00:00Z&date_to=2026-04-21T00:00:00Z" \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# Export audit logs as CSV
curl -X GET "$BASE_URL/auth/admin/audit-logs/export?resource=auth&status=failed" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### Admin: AI Audit Investigation

```bash
# Run an AI investigation over a filtered log window
curl -X POST "$BASE_URL/auth/admin/audit-logs/investigate" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "action": "login",
    "resource": "auth",
    "status": "failed",
    "search": "invalid credentials",
    "date_from": "2026-04-20T00:00:00Z",
    "date_to": "2026-04-21T00:00:00Z",
    "limit": 50
  }'

# List saved investigations
curl -X GET "$BASE_URL/auth/admin/audit-logs/investigations" \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# Get a saved investigation by ID
curl -X GET "$BASE_URL/auth/admin/audit-logs/investigations/1" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

---

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

The API will be available at `http://localhost:8080`. The app respects the `PORT` environment variable automatically.

---

## Docker Workflow

The Docker setup includes an application container, PostgreSQL, Redis, Nginx gateway, and a `db-setup` container that handles migrations and seeding.

```bash
# Start the full stack
docker-compose --env-file .env.docker up --build

# Or use Makefile shortcuts
make docker-up
make docker-down
make docker-logs
make docker-rebuild
```

The gateway is exposed at `http://localhost`. The Nginx layer is optional — the app can run directly without it.

---

## Database Tasks

```bash
make migrate-up                              # run all pending migrations
make migrate-down                            # roll back one migration
make migrate-down-all                        # roll back all migrations
make migrate-status                          # show migration status
make migrate-create NAME=create_example_table  # create a new migration file
make migrate-force VERSION=1                 # force migration version
make migrate-drop CONFIRM=1                  # drop the schema (destructive)
make seed                                    # run seed data
make db-setup                               # run migrations + seed
```

---

## Testing

```bash
make test
# or
go test ./...
```

### Manual Testing with Postman

Included files:
- Collection: [`pleco-api.postman_collection.json`](postman/pleco-api.postman_collection.json)
- Environment: [`pleco-api.local.postman_environment.json`](postman/pleco-api.local.postman_environment.json)

Recommended flow:

1. `Health`
2. `Register`
3. `Verify Email` — or mark the user as verified directly in the database
4. `Login` — stores `access_token` and `refresh_token` automatically
5. `Profile`
6. `Update Profile`
7. `Change Password`
8. `Refresh Token`
9. `Sessions` — list and revoke individual sessions
10. `Logout`
11. Admin requests with an admin token
12. `Audit Logs` — filter by resource, status, date range
13. `AI Investigate` — run an investigation over a filtered log window

Notes:
- `Login` and `Refresh Token` update the Postman environment variables for `access_token` and `refresh_token` automatically.
- `Verify Email` and `Reset Password` require manual token input unless you automate email capture.
- Admin endpoints require an admin access token.
- The AI investigate endpoint requires `AI_ENABLED=true` in your environment.

---

## Makefile Reference

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

---

## Deployment Notes

This starterkit is designed to stay platform-agnostic.

**Recommended production approach:**
- Provide a PostgreSQL-compatible `DATABASE_URL`
- Run migrations before serving traffic
- Run seed data only when you intentionally need initial roles, permissions, or admin users
- Inject secrets through your deployment platform instead of committing real env files
- The app does not redirect HTTP to HTTPS — this should be handled at the gateway or load balancer layer. Do not expose port 8080 directly to the public internet without TLS termination in front of it.

Build and start:

```bash
go build -tags netgo -ldflags '-s -w' -o app ./cmd/api
./app
```

---

## Security Notes

- Never commit real secrets to the repository.
- Use secret managers or platform-managed env vars for production deployments.
- Rotate any third-party credentials that were ever exposed locally or in git history.
- Use separate credentials for local, staging, and production environments.
- The default rate limiter is in-memory and works correctly for a single instance. For multi-instance deployments, replace it with a shared backend such as Redis using the `RateLimitStore` interface.
- The app sets baseline security headers including CSP, HSTS (on HTTPS), `X-Content-Type-Options`, and `X-Frame-Options`.
- Request IDs are propagated via the `X-Request-ID` header for tracing across the gateway and backend.
- Trusted proxy handling is configurable through `TRUSTED_PROXIES` so client IP-based audit and rate limiting work correctly behind a gateway.
- Refresh tokens are rotated on every use — old tokens are invalidated immediately.
- Password reset tokens are invalidated if the user changes their password after the token was issued.

---

## Architecture Notes

- App bootstrap lives in [`internal/appsetup/`](internal/appsetup)
- Runtime configuration is centralized in [`internal/config/app.go`](internal/config/app.go)
- Auth service logic is split by use case under [`internal/modules/auth/`](internal/modules/auth)
- Repository constructors take explicit DB dependencies instead of relying on global DB state
- Admin routes use permission checks instead of role-only checks for finer authorization control
- The recommended Go entrypoint lives in [`cmd/api/`](cmd/api)

---

## Troubleshooting

**`invalid token` on `/auth/profile`**

Usually caused by a stale or incorrect token:
- A `refresh_token` was used instead of an `access_token`
- The access token has expired (15-minute window)
- The token was not stored correctly in Postman

Fix: run Login again, confirm `access_token` is set in your active Postman environment, and retry.

**`relation "users" does not exist`**

Migrations have not run yet. Fix:

```bash
go run ./cmd/migrate
# or
make db-setup
```

**`email not verified` on login**

The user was registered but the verification email was not clicked. Either:
- Check your inbox and click the link
- Resend with `POST /auth/resend-verification` and a JSON email body
- Or set `is_verified = true` directly in the database for development

**`ai investigator is not enabled` on `/auth/admin/audit-logs/investigate`**

`AI_ENABLED` is `false` in your environment. Set `AI_ENABLED=true`, choose a provider, and restart the server. For quick testing, use `AI_PROVIDER=mock`.

**Social login returns `email not available from <provider>`**

The provider did not return an email in the token payload. This can happen when:
- The user has not granted email permission on the provider side
- Apple Sign In is used for the first time and the email is hidden by Apple

Ensure email scope is requested in your frontend OAuth flow before calling this endpoint.

**`relation "audit_investigations" does not exist`**

The migration for the AI investigation feature has not run. Run all pending migrations:

```bash
go run ./cmd/migrate
```

---

## Roadmap Ideas

- Redis-backed rate limit store for multi-instance deployments
- Readiness and liveness probe endpoints
- Database-backed integration tests
- CI validation for migration smoke checks
- Refresh token family tracking to detect token theft

---

## Project Metadata

- License: [MIT](LICENSE)
- Contributing: [CONTRIBUTING.md](CONTRIBUTING.md)
- Security policy: [SECURITY.md](SECURITY.md)
- Code of conduct: [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)
- CI workflow: [ci.yml](.github/workflows/ci.yml)
