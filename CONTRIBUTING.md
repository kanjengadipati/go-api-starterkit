# Contributing

Thanks for considering a contribution to `go-api-starterkit`.

## Before You Start

- Open an issue first for large changes, architectural changes, or new modules.
- Keep pull requests focused. Small, reviewable PRs are preferred over broad mixed changes.
- Make sure secrets, real credentials, and local environment files are never committed.

## Local Setup

1. Copy the local environment file:

```bash
cp .env.example .env
```

2. Run migrations and seed data:

```bash
go run ./cmd/migrate
go run ./cmd/seed
```

3. Start the app:

```bash
go run .
```

Or use Docker:

```bash
cp .env.docker.example .env.docker
make docker-up
```

## Development Guidelines

- Follow the existing modular structure under `modules/`.
- Keep API responses consistent with the project response envelope.
- Prefer clear, explicit dependency wiring over hidden globals.
- Add or update tests when changing behavior.
- Update README, OpenAPI, and Postman assets when public API behavior changes.

## Quality Checks

Run the baseline checks before opening a PR:

```bash
go test ./...
```

Optional repository integration tests can be run with:

```bash
TEST_DATABASE_URL="postgresql://..." go test ./tests -run 'TestPermissionRepository|TestAuditRepository' -v
```

## Pull Request Checklist

- Code compiles and tests pass
- No secrets are committed
- API/docs are updated when needed
- Migration and seed changes are included when schema changes
- PR description explains the problem, change, and impact
