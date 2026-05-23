# Installation Guide

This guide covers the three main ways to run `pleco-api`:

- Local development with Go
- Local Docker Compose
- Production-style execution from GitHub release binaries

## Prerequisites

- Go 1.25 or newer
- PostgreSQL 15 or newer, or MySQL 8.4 or newer
- Redis if you want shared rate limiting and cache behavior
- Node.js 20 or newer if you want to run the bundled Postman suites with Newman

## 1. Clone and install

```bash
git clone https://github.com/kanjengadipati/pleco-api.git
cd pleco-api
go mod download
npm ci
```

`npm ci` is only needed for the Newman-based Postman suites.

## 2. Choose an environment file

Available examples:

- Local: `.env.example`
- Docker: `.env.docker.example`
- Production: `.env.production.example`

For local development:

```bash
cp .env.example .env
```

For Docker:

```bash
cp .env.docker.example .env.docker
```

## 3. Configure the minimum required settings

At minimum, confirm these values match your setup:

```env
DB_DRIVER=postgres
DATABASE_URL=postgresql://postgres:password@localhost:5432/auth_db?sslmode=disable
# MySQL alternative:
# DB_DRIVER=mysql
# DATABASE_URL=mysql://root:password@127.0.0.1:3306/auth_db?parseTime=true
JWT_SECRET=replace_with_a_strong_cryptographically_secure_secret_at_least_32_bytes_long
APP_BASE_URL=http://localhost:8080
FRONTEND_URL=http://localhost:3000
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://127.0.0.1:3000
ADMIN_EMAIL=admin@mail.com
ADMIN_PASSWORD=admin123
EMAIL_PROVIDER=disabled
WA_PROVIDER=disabled
```

Notes:

- `JWT_SECRET` should be at least 32 bytes and unique per environment.
- `CORS_ALLOWED_ORIGINS` must contain the exact browser origins that will call the API.
- `FRONTEND_URL` is used in flows like password reset links.
- `ADMIN_EMAIL` and `ADMIN_PASSWORD` are especially useful when seeding a fresh environment.
- For passwordless OTP, set `EMAIL_PROVIDER=resend` or `smtp` for email delivery, or `WA_PROVIDER=fonnte` with `FONNTE_TOKEN` for WhatsApp delivery.

## 4. Run locally with Go

Start your selected database first, then run:

```bash
go run ./cmd/migrate
go run ./cmd/seed
go run ./cmd/api
```

The API will start on `http://localhost:8080` unless you override `PORT`.

## 5. Run locally with Docker Compose

```bash
cp .env.docker.example .env.docker
make docker-up
```

This starts the selected local stack.

Docker selects the database stack from `DB_DRIVER` in `.env.docker`. Use `DB_DRIVER=postgres` for the default PostgreSQL/PgBouncer stack, or `DB_DRIVER=mysql` for the MySQL stack.

Useful companion commands:

```bash
make docker-logs
make docker-down
```

## 6. Run tests and smoke checks

Go tests:

```bash
make test
```

Postman smoke and negative suites:

```bash
make postman-all
```

These Newman suites expect the API to already be running on the local environment.

## 7. Build binaries locally

```bash
go build -o bin/api ./cmd/api
go build -o bin/migrate ./cmd/migrate
go build -o bin/seed ./cmd/seed
```

Then run them directly:

```bash
./bin/migrate
./bin/seed
./bin/api
```

## 8. Run from a GitHub release

Tagged releases publish binary bundles for Linux, macOS, and Windows.

Each release archive includes:

- `bin/api`
- `bin/migrate`
- `bin/seed`
- `README.md`
- `INSTALLATION.md`
- `TROUBLESHOOTING.md`
- example environment files

Example flow on Linux or macOS:

```bash
tar -xzf pleco-v0.1.0-linux-amd64.tar.gz
cd pleco-v0.1.0-linux-amd64
cp .env.production.example .env
./bin/migrate
./bin/api
```

On Windows, download the `.zip` release asset and run the `.exe` binaries from the `bin` folder.

## 9. Production checklist

- Use `.env.production.example` as the base
- Point `DATABASE_URL` at a managed or private PostgreSQL or MySQL instance
- Configure `REDIS_URL` for shared rate limiting and cache behavior
- Set a strong `JWT_SECRET`
- Restrict `CORS_ALLOWED_ORIGINS` to your real frontend origins
- Serve the API over HTTPS behind a trusted proxy or load balancer
- Keep `AUTO_RUN_MIGRATIONS=false` and `AUTO_RUN_SEEDS=false` for normal app startup

## 10. Verify the installation

- `GET /health` returns `200`
- `GET /docs` serves the API docs
- Register or seed an admin account successfully
- Login returns an access token and a `pleco_refresh_token` cookie
- Protected endpoints reject invalid or stale tokens as expected
