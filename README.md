# Go API Starterkit

[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://go.dev)
[![Gin](https://img.shields.io/badge/Gin-HTTP%20Framework-009688)](https://gin-gonic.com)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-Database-336791?logo=postgresql)](https://www.postgresql.org)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker)](https://www.docker.com)
[![Modular](https://img.shields.io/badge/Architecture-Modular-6f42c1)](https://github.com/kanjengadipati/go-api-starterkit)
[![AI Powered](https://img.shields.io/badge/AI-Audit%20Investigator-ff6b35?logo=ollama)](https://ollama.com)
[![License](https://img.shields.io/badge/License-MIT-green)](LICENSE)

> **Skip the boilerplate.** A modular, production-oriented Go REST API foundation with JWT authentication, social login, RBAC, audit trail, and an AI-powered log investigator — ready to clone, extend, and ship.

🔗 **Dashboard Demo:** [gokitdash.vercel.app](https://gokitdash.vercel.app/) &nbsp;|&nbsp; 📖 **API Docs:** [go-api-starterkit.onrender.com/docs](https://go-api-starterkit.onrender.com/docs)

---

## Overview

This project provides a complete authentication and authorization foundation built with Go, Gin, GORM, and PostgreSQL. It is structured around independent, domain-focused modules that can be extended or replaced without touching the rest of the app.

**Core features:**
- User registration and login
- Access token and refresh token flow
- Logout, profile, and session management endpoints
- Self profile update and password change
- Email verification, forgot password, and reset password
- Google, Facebook, and Apple social login
- Admin user management
- Audit trail for important auth and user actions
- Optional AI-powered audit log investigator (mock or Ollama)
- Permission-based authorization for admin actions
- Basic rate limiting with a swappable store abstraction and hardened security headers
- Request-scoped structured logging with request ID propagation
- Database migration and seeding
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
| AI | Ollama (optional) |
| Infrastructure | Docker, Nginx, Redis |

## Architecture

![Architecture](docs/architecture.svg)

### Go toolchain

`go.mod` declares **Go 1.25** because `go mod tidy` pulls that floor from transitive modules (for example `golang.org/x/oauth2`, `golang.org/x/sys`, and related `golang.org/x/*` releases used by Google APIs and the rest of the graph). Install a [supported Go toolchain](https://go.dev/dl/) that satisfies the `go` line before running `go test` or `go build`. If you must target an older Go release, you will need to fork and pin older `golang.org/x/*` (and possibly trim Google social-login dependencies)—that is not supported in this template out of the box.

---

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
REDIS_URL=
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
- Facebook and Apple social login are implemented, but production-hardening steps such as `appsecret_proof` for Facebook and stronger nonce handling for Apple should still be considered before public internet exposure.
- `AI_ENABLED=false` keeps the app fully usable without AI.
- `AI_PROVIDER` supports `mock`, `ollama`, `openai`, and `gemini`.
- `AI_BASE_URL` is optional for hosted providers and only required when `AI_PROVIDER=ollama`.
- `AUTO_RUN_MIGRATIONS` and `AUTO_RUN_SEEDS` are optional flags for startup-time initialization (keep `false` for local and Docker).
- `REDIS_URL` (optional) enables a shared Redis-backed rate limit store for auth endpoints; when unset, the app uses an in-memory limiter suitable for a single instance. Wiring happens in [`internal/appsetup/rate_limit_store.go`](internal/appsetup/rate_limit_store.go) so `internal/config` stays free of `middleware` imports.

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

For OpenAI-hosted AI:

```env
AI_ENABLED=true
AI_PROVIDER=openai
AI_MODEL=gpt-4.1-mini
AI_API_KEY=your_openai_api_key
AI_TIMEOUT_SECONDS=30
```

For Gemini free-tier style demos:

```env
AI_ENABLED=true
AI_PROVIDER=gemini
AI_MODEL=gemini-2.5-flash
AI_API_KEY=your_gemini_api_key
AI_TIMEOUT_SECONDS=30
```

Make sure Ollama is running and the model is available:

```bash
ollama serve
ollama pull qwen2.5:3b
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
- Creating or reusing an audit investigation is itself written into the audit log trail.

### Common Failures

| Error | Cause | Fix |
|---|---|---|
| `ai investigator is not enabled` | `AI_ENABLED` is still false | Set `AI_ENABLED=true` and restart |
| `ollama is unavailable` | Ollama is not running | Run `ollama serve` |
| `ollama model is not available` | Model not pulled | Run `ollama pull <model>` |
| `openai error: bad api key` | Invalid or missing OpenAI API key | Set `AI_API_KEY` to a valid OpenAI key |
| `gemini error: unsupported model` | Wrong Gemini model name or unavailable model | Use a supported Gemini model such as `gemini-2.5-flash` |
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
| POST | `/auth/admin/audit-logs/investigate` | Run AI investigation |
| GET | `/auth/admin/audit-logs/investigations` | List saved investigations |
| GET | `/auth/admin/audit-logs/investigations/:id` | Get investigation detail |
| GET | `/auth/admin/roles` | List roles |
| GET | `/auth/admin/permissions` | List permissions |
| GET | `/auth/admin/roles/:id/permissions` | Get role permissions |
| PUT | `/auth/admin/roles/:id/permissions` | Update role permissions |

### Health

| Method | Endpoint | Description |
|---|---|---|
| GET | `/health` | Health check |

---

## API Conventions

- Authenticated routes require `Authorization: Bearer <access_token>`
- Access tokens are short-lived (10 minutes) and carry a security version; admin routes reject tokens that no longer match the database after role or password changes
- Admin routes require an access token that belongs to an admin user
- Refresh tokens are only valid for `POST /auth/refresh`
- Success responses use the envelope: `status`, `message`, optional `data`, optional `meta`
- Error responses use the envelope: `status`, `message`, optional `errors`
- OpenAPI reference: [`docs/openapi.yaml`](docs/openapi.yaml)
- Swagger UI: served at `/docs`

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

The Docker setup includes an application container, PostgreSQL, Redis, Nginx gateway, and a `db-setup` container for migration and seed.

```bash
# Start full stack
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
make migrate-up           # run migrations
make migrate-down         # roll back one migration
make migrate-status       # show migration status
make migrate-create NAME=create_example_table
make seed                 # run seed data
make db-setup             # run migrations + seed
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
- Collection: [`go-api-starterkit.postman_collection.json`](postman/go-api-starterkit.postman_collection.json)
- Environment: [`go-api-starterkit.local.postman_environment.json`](postman/go-api-starterkit.local.postman_environment.json)

Recommended flow: Health → Register → Verify Email → Login → Profile → Update Profile → Change Password → Refresh Token → Logout → Admin endpoints → Audit Logs → AI Investigate

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

**TLS / HTTPS:** The Go process listens for plain HTTP (for example on `PORT` / 8080). Production setups should terminate TLS at a reverse proxy or load balancer (the included Docker stack uses Nginx for this). If clients can reach the app port directly without TLS (misconfigured firewall, bypassing the proxy), traffic will not be encrypted at the application layer—ensure only the TLS front end is exposed on the public internet.

```bash
go build -tags netgo -ldflags '-s -w' -o app ./cmd/api
./app
```

---

## Security Notes

- Never commit real secrets to the repository
- Use secret managers or platform-managed env vars for production deployments
- Rotate any third-party credentials that were ever exposed locally or in git history
- Use separate credentials for local, staging, and production environments
- Sensitive auth endpoints share one rate-limit store: set `REDIS_URL` for Redis (multi-instance friendly); otherwise an in-memory store is used (single instance only).
- The app sets baseline security headers including CSP, HSTS on HTTPS requests, `X-Content-Type-Options`, and `X-Frame-Options`
- Request IDs are propagated via the `X-Request-ID` header
- Trusted proxy handling is configurable through `TRUSTED_PROXIES`

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

Usually caused by a stale or incorrect token. Run Login again, confirm `access_token` exists in your Postman environment, and retry.

**`relation "users" does not exist`**

Migrations have not run yet. Run `go run ./cmd/migrate` or `make db-setup`.

---

## Roadmap Ideas

- Add database-backed integration tests
- Split readiness and liveness probes
- Add CI validation for migration and deployment smoke checks
- Further reduce infrastructure-specific behavior inside app startup

---

## Project Metadata

- License: [MIT](LICENSE)
- Contributing: [CONTRIBUTING.md](CONTRIBUTING.md)
- Security policy: [SECURITY.md](SECURITY.md)
- Code of conduct: [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)
- CI workflow: [ci.yml](.github/workflows/ci.yml)
