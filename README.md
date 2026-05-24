# Pleco
## Ship your product. Let Pleco handle the auth.

- Setup guide: [INSTALLATION.md](./INSTALLATION.md)
- Common issues: [TROUBLESHOOTING.md](./TROUBLESHOOTING.md)

<img src="https://pleco-console.vercel.app/logo.png" alt="Pleco Logo" width="120" />

[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://go.dev)
[![Gin](https://img.shields.io/badge/Gin-HTTP%20Framework-009688)](https://gin-gonic.com)
[![Database](https://img.shields.io/badge/Database-PostgreSQL%20%7C%20MySQL-336791)](https://github.com/kanjengadipati/pleco-api)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker)](https://www.docker.com)
[![Modular](https://img.shields.io/badge/Architecture-Modular-6f42c1)](https://github.com/kanjengadipati/pleco-api)
[![AI Powered](https://img.shields.io/badge/AI-Audit%20Investigator-ff6b35?logo=ollama)](https://ollama.com)
[![License](https://img.shields.io/badge/License-MIT-green)](LICENSE)

> **Ship your product. Let Pleco handle the auth.** A modular, production-oriented Go REST API foundation with JWT authentication, social login, RBAC, per-device session management, audit trail, and an AI-powered log investigator — ready to clone, extend, and ship.

Intended for Go backend developers who want a solid, security-conscious auth foundation to build on — without reinventing JWT flows, email verification, social login, or audit logging from scratch.

🔗 **Dashboard Demo:** [pleco-console.vercel.app](https://pleco-console.vercel.app/) &nbsp;|&nbsp; 📖 **API Docs:** [pleco-api.onrender.com/docs](https://pleco-api.onrender.com/docs)

---

## Overview

Pleco is a production-oriented authentication and authorization API foundation for Go applications. It gives you the core auth system most products need: JWT login, refresh token rotation, per-device sessions, email verification, password recovery, social login, RBAC, admin user management, audit logging, and optional AI-assisted audit investigation.

The codebase is organized around domain-focused modules, so auth, users, roles, permissions, tokens, social login, audit logs, and monitoring can evolve independently. Each module owns its handler, service, repository, and model, making the project easier to extend than a single large auth package.

Pleco is designed for teams who want a practical starting point for a real backend: secure defaults, PostgreSQL or MySQL migrations, Redis-backed rate limiting, structured logging, Docker Compose for local development, and deployment-friendly configuration.

**Authentication:**
- User registration and login
- Access token and HttpOnly refresh-cookie flow with token rotation
- Per-device session management - list, revoke, and logout individual sessions
- Self profile update and password change
- Email verification, forgot password, and reset password
- Google, Facebook, and Apple social login with server-side token validation

**Authorization and admin:**
- Role-based access control (RBAC) with fine-grained permission checks per route
- Admin user management
- Token invalidation after password resets, password changes, and role changes

**Security and operations:**
- Audit trail for important auth and user actions
- Optional AI-powered audit log investigator (Ollama, OpenAI, Gemini, or mock)
- Per-route rate limiting with a swappable store abstraction
- Hardened security headers (CSP, HSTS, X-Frame-Options, X-Content-Type-Options)
- Request-scoped structured logging with request ID propagation
- Database migration and seeding via golang-migrate
- Local Docker workflow with Nginx, PostgreSQL or MySQL, and Redis
- Generic PostgreSQL/MySQL deployment support

**Session revocation behavior:**
- Access tokens carry a token-version claim and protected routes reject stale tokens after revocation-sensitive events.
- Password changes and password resets revoke all stored refresh tokens for the user.
- Admin role changes revoke the target user's refresh tokens and invalidate previously issued access tokens.
- `POST /auth/logout-all` revokes every session, including the current access token for subsequent requests.

---

## Tech Stack

| Layer | Technology |
|---|---|
| Language | Go 1.25+ |
| HTTP Framework | Gin |
| ORM | GORM |
| Database | PostgreSQL or MySQL |
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

### Releases

Tagged releases publish prebuilt binary bundles for Linux, macOS, and Windows. Each archive includes the `api`, `migrate`, and `seed` executables together with the setup and troubleshooting docs so you can bootstrap a host without compiling Go on the target machine.

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
- Production: [`.env.production.example`](.env.production.example)

### Common Variables

```env
PORT=8080
DB_DRIVER=postgres
DATABASE_URL=postgresql://postgres:password@localhost:5432/auth_db?sslmode=disable
# MySQL alternative:
# DB_DRIVER=mysql
# DATABASE_URL=mysql://root:password@localhost:3306/auth_db?parseTime=true
TRUSTED_PROXIES=127.0.0.1,::1,10.0.0.0/8,172.16.0.0/12,192.168.0.0/16
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://127.0.0.1:3000
JWT_SECRET=replace-with-a-strong-secret
ACCESS_TOKEN_EXPIRY_MINUTES=15
APP_BASE_URL=http://localhost:8080
FRONTEND_URL=http://localhost:3000
ADMIN_EMAIL=admin@example.com
ADMIN_PASSWORD=supersecret
AUTO_RUN_MIGRATIONS=false
AUTO_RUN_SEEDS=false
DB_MAX_OPEN_CONNS=5
DB_MAX_IDLE_CONNS=2
DB_CONN_MAX_LIFETIME_MINUTES=30
EMAIL_PROVIDER=disabled
EMAIL_API_KEY=
EMAIL_API_BASE_URL=
EMAIL_FROM=
EMAIL_FROM_NAME=Go App
EMAIL_REPLY_TO=
EMAIL_TIMEOUT_SECONDS=15
EMAIL_SMTP_HOST=
EMAIL_SMTP_PORT=587
EMAIL_SMTP_USERNAME=
EMAIL_SMTP_PASSWORD=
EMAIL_SMTP_MODE=starttls
WA_PROVIDER=disabled
FONNTE_TOKEN=
FONNTE_BASE_URL=https://api.fonnte.com
WA_CLOUD_ACCESS_TOKEN=
WA_CLOUD_PHONE_NUMBER_ID=
WA_CLOUD_API_BASE_URL=https://graph.facebook.com
WA_CLOUD_API_VERSION=v20.0
WA_TIMEOUT_SECONDS=15
OTP_RATE_LIMIT_REQUESTS=5
OTP_RATE_LIMIT_WINDOW_SECONDS=3600
OTP_TARGET_COOLDOWN_SECONDS=60
OTP_TARGET_RATE_LIMIT_REQUESTS=5
OTP_TARGET_RATE_LIMIT_WINDOW_SECONDS=3600
REDIS_URL=
REDIS_HOST=
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
SOCIAL_ACTIVE_PROVIDERS=
SOCIAL_GOOGLE_CLIENT_ID=
SOCIAL_GOOGLE_CLIENT_SECRET=
SOCIAL_FACEBOOK_CLIENT_ID=
SOCIAL_FACEBOOK_CLIENT_SECRET=
SOCIAL_APPLE_CLIENT_ID=
SOCIAL_APPLE_CLIENT_SECRET=
AI_ENABLED=false
ENABLE_ERROR_OPTIMIZATION=false
AI_PROVIDER=mock
AI_MODEL=mock-model
AI_BASE_URL=
AI_API_KEY=
AI_TIMEOUT_SECONDS=30
```

### Notes

- `DATABASE_URL` is the primary database connection setting.
- `DB_DRIVER` supports `postgres` and `mysql`; if omitted, Pleco infers the driver from the `DATABASE_URL` scheme.
- `TRUSTED_PROXIES` controls which proxy hops are trusted for forwarded client IP handling.
- `CORS_ALLOWED_ORIGINS` should list explicit frontend origins for browser clients; credentialed refresh cookies cannot be used safely with wildcard CORS.
- The app validates critical configuration at startup and exits early when required values are missing.
- `APP_BASE_URL` is used for backend-generated links such as email verification.
- `FRONTEND_URL` is used for password reset links when you have a separate frontend.
- `EMAIL_PROVIDER` supports `disabled`, `smtp`, `sendgrid`, and `resend`.
- `smtp` is the most flexible option and works with any standard SMTP relay.
- `EMAIL_API_KEY`, `EMAIL_API_BASE_URL`, `EMAIL_FROM`, `EMAIL_FROM_NAME`, and `EMAIL_REPLY_TO` are the shared API-provider settings.
- `EMAIL_SMTP_HOST`, `EMAIL_SMTP_PORT`, `EMAIL_SMTP_USERNAME`, `EMAIL_SMTP_PASSWORD`, and `EMAIL_SMTP_MODE` are used when `EMAIL_PROVIDER=smtp`.
- `WA_PROVIDER` supports `disabled`, `fonnte`, and `whatsapp_cloud` for WhatsApp OTP delivery.
- `FONNTE_TOKEN`, `FONNTE_BASE_URL`, and `WA_TIMEOUT_SECONDS` configure Fonnte WhatsApp OTP delivery.
- `WA_CLOUD_ACCESS_TOKEN`, `WA_CLOUD_PHONE_NUMBER_ID`, `WA_CLOUD_API_BASE_URL`, and `WA_CLOUD_API_VERSION` configure Meta WhatsApp Cloud API delivery.
- `OTP_RATE_LIMIT_REQUESTS` and `OTP_RATE_LIMIT_WINDOW_SECONDS` configure the route-level `/auth/request-otp` limit. Defaults are 5 requests per 3600 seconds.
- `OTP_TARGET_COOLDOWN_SECONDS`, `OTP_TARGET_RATE_LIMIT_REQUESTS`, and `OTP_TARGET_RATE_LIMIT_WINDOW_SECONDS` configure per-target OTP limits. Defaults are a 60-second cooldown and 5 requests per 3600 seconds.
- `SOCIAL_ACTIVE_PROVIDERS` enables configured providers such as `google,facebook,apple`.
- `SOCIAL_GOOGLE_CLIENT_ID`, `SOCIAL_FACEBOOK_CLIENT_ID`, and `SOCIAL_APPLE_CLIENT_ID` are used for provider audience validation.
- `SOCIAL_FACEBOOK_CLIENT_SECRET` is required for Facebook token inspection when Facebook is active.
- `AI_ENABLED=false` keeps the app fully usable without AI.
- `ENABLE_ERROR_OPTIMIZATION` determines whether the API generates user-friendly, AI-enriched error responses (requires `AI_ENABLED=true`).
- `AI_PROVIDER` supports `mock`, `ollama`, `openai`, `gemini`, and `anthropic`.
- `AI_BASE_URL` is only required when `AI_PROVIDER=ollama`.
- `REDIS_URL` or `REDIS_HOST`/`REDIS_PORT` enables shared Redis-backed rate limiting and response caching. Without Redis, the app falls back to in-memory stores for local single-instance development.
- `AUTO_RUN_MIGRATIONS` and `AUTO_RUN_SEEDS` are optional flags for startup-time initialization. Keep these `false` for local and Docker workflows — run migrations and seeds manually instead.

---

## AI-Powered Error Optimization

This project features an intelligent error handling system that converts generic backend errors (e.g., "invalid credentials") into user-friendly, actionable feedback dynamically. 

When `ENABLE_ERROR_OPTIMIZATION=true` and an error occurs:
1. The error is intercepted and mapped to an internal code (e.g., `AUTH_INVALID_CREDENTIALS`).
2. The AI model generates a friendly message, additional context, and actionable suggestions (like "Reset Password").
3. The response is cached in Redis to guarantee ultra-fast responses for subsequent identical errors.
4. If AI is unavailable, the system safely falls back to standard generic messages.

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

For Claude via Anthropic:

```env
AI_ENABLED=true
AI_PROVIDER=anthropic
AI_MODEL=claude-sonnet-4-5
AI_API_KEY=your_anthropic_api_key
AI_TIMEOUT_SECONDS=30
```

### Typical Admin Flow

1. Query audit logs with `GET /auth/admin/audit-logs`
2. Narrow the result with filters: `resource`, `status`, `actor_user_id`, `search`, `date_from`, `date_to`
3. Send the same filter scope to `POST /auth/admin/audit-logs/investigations`
4. Review the generated summary and recommendations
5. Re-open saved investigation history from `GET /auth/admin/audit-logs/investigations`

### Investigation Request

```json
POST /auth/admin/audit-logs/investigations

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
  "status": "success",
  "message": "Audit investigation completed",
  "data": {
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
  },
  "meta": {
    "investigation_id": 12,
    "reused_existing": false,
    "log_count": 42,
    "limit": 50,
    "resource": "auth",
    "action": "login",
    "status": "failed"
  }
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
| `anthropic error: bad api key` | Invalid or missing Anthropic API key | Set `AI_API_KEY` to a valid Anthropic key |
| `ai investigation timed out` | Model too slow | Increase `AI_TIMEOUT_SECONDS` or use a smaller model |

---

## Main Endpoints

### Auth

| Method | Endpoint | Description |
|---|---|---|
| POST | `/auth/register` | Register a new user |
| POST | `/auth/login` | Login with email/password, receive an access token, and set the refresh cookie |
| POST | `/auth/request-otp` | Request passwordless OTP via WhatsApp or email |
| POST | `/auth/verify-otp` | Verify OTP, auto-register when needed, and create a session |
| POST | `/auth/refresh` | Refresh access token using the refresh cookie |
| GET | `/auth/verify` | Verify email address |
| POST | `/auth/resend-verification` | Resend verification email |
| POST | `/auth/forgot-password` | Request password reset |
| POST | `/auth/reset-password` | Reset password with token |
| POST | `/auth/social-login` | Login via Google, Facebook, or Apple |
| GET | `/auth/profile` | Get current user profile |
| GET | `/auth/social/:provider/account` | Get linked social account |
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
| GET | `/auth/admin/users/:id/permissions` | Get user permissions |
| POST | `/auth/admin/users` | Create user |
| PUT | `/auth/admin/users/:id` | Update user |
| DELETE | `/auth/admin/users/:id` | Delete user |
| GET | `/auth/admin/audit-logs` | List audit logs |
| GET | `/auth/admin/audit-logs/export` | Export audit logs as CSV |
| POST | `/auth/admin/audit-logs/investigations` | Run AI investigation |
| GET | `/auth/admin/audit-logs/investigations` | List saved investigations |
| GET | `/auth/admin/audit-logs/investigations/:id` | Get investigation detail |
| GET | `/auth/admin/roles` | List roles |
| GET | `/auth/admin/roles/:id` | Get role by ID |
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
- Refresh tokens are issued as the `pleco_refresh_token` HttpOnly cookie and are only valid for `POST /auth/refresh`
- Device identity is issued as the `pleco_device_id` HttpOnly cookie and reused by session, trusted-device, and logout flows
- Browser clients must send credentials/cookies when calling login, refresh, logout, or logout-others
- Access tokens must include the server-issued token-version claim
- After password reset, password change, role change, or `logout-all`, previously issued tokens can start returning `401` immediately
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

{
  "email": "tester@example.com",
  "password": "secret123"
}
```

Response:

The response also sets HttpOnly cookies for refresh and device identity: `pleco_refresh_token=...` and `pleco_device_id=...`.

```json
{
  "status": "success",
  "message": "Login success",
  "data": {
    "access_token": "<jwt>"
  }
}
```

### Passwordless Login

Pleco supports passwordless login through OTP and trusted-device magic links. The first step validates an email address or WhatsApp number without sending anything. The second step lets the backend decide the delivery: trusted devices receive a magic link by email, while other attempts receive an OTP through the selected channel.

Check identity:

```http
POST /auth/passwordless/check
Content-Type: application/json

{
  "channel": "whatsapp",
  "target": "+628123456789"
}
```

Start delivery:

```http
POST /auth/passwordless/start
Content-Type: application/json

{
  "channel": "whatsapp",
  "target": "+628123456789"
}
```

If `next_step` is `otp`, verify the OTP:

```http
POST /auth/verify-otp
Content-Type: application/json

{
  "channel": "whatsapp",
  "target": "+628123456789",
  "otp": "123456",
  "device_name": "Chrome macOS",
  "trusted_device": true
}
```

If `next_step` is `magic_link`, open the emailed link and verify it through `POST /auth/magic-link/verify`. Magic links are stored as hashed, single-use DB tokens with expiry and are rejected after first use. Successful verification returns an access token and sets the `pleco_refresh_token` and `pleco_device_id` HttpOnly cookies. If an OTP target does not exist yet, Pleco auto-creates a user and marks `phone_verified` or `email_verified` based on the channel. OTP codes are hashed, expire after 5 minutes, are consumed after successful verification, and are guarded by cooldown, hourly target limits, max attempts, and audit logs.

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
    "role": "user",
    "permissions": [
      "user.read"
    ]
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
  -d '{
    "email": "tester@example.com",
    "password": "secret123"
  }'
```

Store the access token and refresh cookie for use in subsequent requests:

```bash
COOKIE_JAR=/tmp/pleco-cookies.txt

TOKENS=$(curl -s -c "$COOKIE_JAR" -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "tester@example.com",
    "password": "secret123"
  }')

ACCESS_TOKEN=$(echo $TOKENS | jq -r '.data.access_token')
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

After a successful password change, existing refresh tokens are revoked. Log in again before calling `POST /auth/refresh` or other authenticated flows with an old session.

### Refresh Token

```bash
curl -X POST "$BASE_URL/auth/refresh" \
  -H "Content-Type: application/json" \
  -b "$COOKIE_JAR" \
  -c "$COOKIE_JAR" \
  -d '{}'
```

`POST /auth/refresh` also accepts the legacy JSON `refresh_token` body for non-browser clients, but browser and dashboard clients should rely on the HttpOnly cookie.
The API also manages `pleco_device_id` as an HttpOnly cookie, so browser clients do not need to create or send a device id.

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

After a successful password reset, existing refresh tokens are revoked and older access tokens can be rejected immediately on protected routes.

### Social Login

Supported providers: `google`, `facebook`, `apple`.

- Google and Apple expect an ID token.
- Facebook expects a user access token.
- All three use the same `token` field for consistency.
- The starterkit requires an email from the provider to map or create a local user.

```bash
curl -X POST "$BASE_URL/auth/social-login" \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "google",
    "token": "<provider-token>"
  }'
```

Fetch a linked social account:

```bash
curl -X GET "$BASE_URL/auth/social/google/account" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### Sessions

List active sessions:

```bash
curl -X GET "$BASE_URL/auth/sessions" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -b "$COOKIE_JAR"
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
  -b "$COOKIE_JAR"
```

Logout all sessions:

```bash
curl -X POST "$BASE_URL/auth/logout-all" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

After `logout-all`, the current access token should be treated as expired for the rest of the session and the user should log in again.

Logout every session except the current device:

```bash
curl -X POST "$BASE_URL/auth/logout-others" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -b "$COOKIE_JAR" \
  -c "$COOKIE_JAR"
```

### Admin: Users

```bash
# List users
curl -X GET "$BASE_URL/auth/admin/users?page=1&limit=10" \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# Get user by ID
curl -X GET "$BASE_URL/auth/admin/users/1" \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# Get user permissions
curl -X GET "$BASE_URL/auth/admin/users/1/permissions" \
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

# Get role by ID
curl -X GET "$BASE_URL/auth/admin/roles/2" \
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
curl -X POST "$BASE_URL/auth/admin/audit-logs/investigations" \
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

Update the values to match your selected local database setup. Use `DB_DRIVER=postgres` with a PostgreSQL URL, or `DB_DRIVER=mysql` with a MySQL URL.

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

The Docker setup is intended for local development. It includes an application container, Redis, Nginx gateway, and a `db-setup` container that handles migrations and seeding. The default `DB_DRIVER=postgres` stack includes PostgreSQL and PgBouncer; `DB_DRIVER=mysql` uses MySQL and the MySQL migration set.

```bash
# Start the selected stack from .env.docker
make docker-up
make docker-down
make docker-logs
make docker-rebuild
```

The gateway is exposed at `http://localhost`. The Nginx layer is optional - the app can run directly without it.

Set `DB_DRIVER=postgres` in `.env.docker` to use `docker-compose.yaml`, or `DB_DRIVER=mysql` to use `docker-compose.mysql.yaml`. The Makefile selects the compose file automatically. PgBouncer is included only in the PostgreSQL Docker stack as a production-like pooling layer; API traffic uses PgBouncer at `pgbouncer:5432`, while the `db-setup` container connects directly to Postgres for migrations and seed data.

### Production Deployment

For production, deploy only the API image and connect it to managed/private infrastructure:

- Managed PostgreSQL or MySQL via `DATABASE_URL`.
- Managed Redis via `REDIS_URL` so rate limits and cache entries are shared across replicas.
- A platform load balancer or reverse proxy for TLS and public traffic.
- `AUTO_RUN_MIGRATIONS=false` and `AUTO_RUN_SEEDS=false` on the long-running API service.

Use [`docker-compose.prod.example.yaml`](docker-compose.prod.example.yaml) as a minimal production-style reference. It intentionally does not include local database containers, PgBouncer, Redis, or Nginx. Run migrations once per deploy as a release job or CI/CD step before starting or rolling the API replicas.

Do not publish database ports such as PostgreSQL `5432` or MySQL `3306` in production. For database administration, use your provider console, VPN, SSH tunnel, bastion host, or IP-restricted external endpoint.

---

## Response Caching

Pleco caches hot auth/admin reads with Redis when `REDIS_URL` or `REDIS_HOST` is configured. If Redis is unavailable, it falls back to an in-memory cache suitable for local single-instance development.

| Endpoint / Path | Cache Key | TTL |
|---|---:|---:|
| `GET /auth/admin/users/:id/permissions` | `user:permissions:{userID}` | 10 minutes |
| `GET /auth/profile` | `user:profile:{userID}` | 5 minutes |
| `GET /auth/admin/roles` | `roles` | 20 minutes |
| `GET /auth/admin/roles/:id` | `role:{roleID}` | 15 minutes |
| `GET /auth/admin/roles/:id/permissions` | `role:{roleID}:permissions` | 15 minutes |
| `GET /auth/admin/users/:id` | `user:detail:{userID}` | 5 minutes |
| `GET /auth/social/:provider/account` | `social:account:{userID}:{provider}` | 15 minutes |

Permission middleware also caches role permission checks for 10 minutes using `role:permission:{role}:{permission}`. User and role writes invalidate the related cached profile, detail, permission, and role entries so authorization-sensitive changes are refreshed promptly.

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

The Makefile uses `migrations/` for PostgreSQL and `migrations/mysql/` when `DB_DRIVER=mysql`.

---

## Testing

```bash
make test
# or
go test ./...
```

### Manual Testing with Postman

Included files:
- Collection: [`pleco.postman_collection.json`](postman/pleco.postman_collection.json)
- Docker collection: [`pleco.docker.postman_collection.json`](postman/pleco.docker.postman_collection.json)
- Environment: [`pleco.local.postman_environment.json`](postman/pleco.local.postman_environment.json)

Use the Docker collection when the Compose stack is running through the Nginx gateway at `http://localhost`. Use the regular collection when the API is running directly at `http://localhost:8080`.

Recommended flow:

1. `Health`
2. `Register`
3. `Verify Email` — or mark the user as verified directly in the database
4. `Login` — stores `access_token` and the `pleco_refresh_token` / `pleco_device_id` cookies automatically
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
- `Login` and `Refresh Token` update the Postman environment variable for `access_token`; refresh and device cookies are managed by Postman's cookie jar as `pleco_refresh_token` and `pleco_device_id`.
- `Logout Other Sessions` also rotates the refresh cookie and stores a fresh `access_token`.
- `Verify Email` and `Reset Password` require manual token input unless you automate email capture.
- Admin endpoints require an admin access token.
- The AI investigate endpoint requires `AI_ENABLED=true` in your environment.

### Automated API Checks with Newman

Install the Newman dependency once:

```bash
npm install
```

Run the local collection against the checked-in environment file:

```bash
npm run postman:local
# or
make postman-test
```

This uses:
- `postman/pleco.smoke.postman_collection.json`
- `postman/pleco.local.postman_environment.json`

The collection expects the API to already be running at the `base_url` configured in the environment file.

For the full manual collection, including flows that need verification/reset tokens or intentionally mutate credentials:

```bash
npm run postman:manual
```

Run the negative test suite for validation/authz/error handling:

```bash
npm run postman:negative
# or
make postman-negative
```

Run both positive smoke checks and negative checks in one command:

```bash
npm run postman:all
# or
make postman-all
```

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
- Provide a PostgreSQL or MySQL `DATABASE_URL`
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
- The default rate limiter is in-memory for local single-instance development. When Redis is configured, Pleco switches to a shared Redis-backed store suitable for multi-instance deployments.
- The app sets baseline security headers including CSP, HSTS (on HTTPS), `X-Content-Type-Options`, and `X-Frame-Options`.
- Request IDs are propagated via the `X-Request-ID` header for tracing across the gateway and backend.
- Trusted proxy handling is configurable through `TRUSTED_PROXIES` so client IP-based audit and rate limiting work correctly behind a gateway.
- Device identity is managed by the API through the `pleco_device_id` HttpOnly cookie. Browser clients should not create or persist their own device id in JavaScript.
- Refresh tokens are stored in the `pleco_refresh_token` HttpOnly cookie and rotated on every use. Pleco keeps refresh-token family metadata so reuse of an already-rotated token can revoke the whole family and invalidate prior access tokens.
- Passwordless magic links are stored as hashed, single-use database tokens with expiry; verification consumes the token in a transaction so reused links are rejected.
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
- A refresh cookie value was used as a bearer token instead of `data.access_token`
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

**`ai investigator is not enabled` on `/auth/admin/audit-logs/investigations`**

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

- Audit anomaly scoring tuned with real-world operator feedback

---

## Project Metadata

- License: [MIT](LICENSE)
- Contributing: [CONTRIBUTING.md](CONTRIBUTING.md)
- Security policy: [SECURITY.md](SECURITY.md)
- Code of conduct: [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)
- CI workflow: [ci.yml](.github/workflows/ci.yml)

## 📧 Contact & Support

- **Email:** theplecodev@gmail.com
- **GitHub Issues:** [Report bugs](https://github.com/pleco-dev/pleco-api/issues)
- **LinkedIn:** [@heriheriyadi](https://linkedin.com/in/heriheriyadi)

For security vulnerabilities, please email directly. See [`SECURITY.md`](SECURITY.md).

## Monitoring & Observability

Pleco includes optional monitoring with AI-powered error analysis capabilities.

### Basic Monitoring

Monitor errors with Sentry or Datadog. It automatically captures 5xx errors and supports standard metrics.

```env
# Using Sentry
MONITORING_PROVIDER=sentry
SENTRY_DSN=https://key@sentry.io/project
```

```env
# Using Datadog
MONITORING_PROVIDER=datadog
DATADOG_API_KEY=your_key
```

```bash
go run ./cmd/api
```

### AI-Powered Error Analysis

When enabled, AI analyzes error patterns, stores root causes in the database, and samples errors to save AI provider costs (analyzing 1 in N errors).

```env
MONITORING_PROVIDER=sentry
SENTRY_DSN=...
AI_MONITORING_ENABLED=true
AI_MONITORING_ERROR_THRESHOLD=5
AI_PROVIDER=ollama
AI_MODEL=qwen2.5:3b
AI_BASE_URL=http://localhost:11434
```
