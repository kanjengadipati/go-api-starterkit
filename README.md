# Go Auth App

A robust, modular authentication API built with Go and Gin.  
Features secure user registration, login, JWT authentication, refresh logic, logout, simple role-based access control (RBAC), and now **email verification** for account activation.  
Designed for easy customization, strong testing, and rapid startup.

## 🚀 Features

- User registration and login endpoints
- Email verification: users receive a verification link after registering
- Secure password hashing with bcrypt
- JWT-based authentication with support for access and refresh tokens
- Auto-rotation & invalidation of used/expired refresh tokens
- Server-side logout: deletes/invalidate refresh tokens (not just JWT expiry)
- Middleware-protected endpoints (e.g., `/profile`, `/users`)
- Simple, extensible role-based guards (admin/user)
- Clean architecture: repositories, models, controllers, services
- Extensive unit tests in `/tests` (controllers, services, repositories)
- Mocks for business logic/testing
- `.env` support for configuration

## 🏁 Quickstart

### Prerequisites

- Go 1.18 or newer ([get Go](https://golang.org/dl/))
- Docker (optional, for Postgres database)

### Installation

```sh
git clone https://github.com/your-username/go-auth-app.git
cd go-auth-app
go mod tidy
cp .env.example .env    # Copy and edit .env for your DB/email/JWT config
```

### Configuration

Configure your `.env` file:
- `DATABASE_URL` — Postgres connection string
- `JWT_SECRET` — your secret for JWT signing
- `PORT` — API server port (default: 8080)
- `EMAIL_HOST`, `EMAIL_PORT`, `EMAIL_USER`, `EMAIL_PASS`, `EMAIL_FROM` — for verification emails

### Running the Server

```sh
go run main.go
```

API available at: [http://localhost:8080](http://localhost:8080)

## 📚 API Endpoints Overview

- `POST /register` — Register a new user  
  **Body:**
  ```json
  {
    "name": "Alice Smith",
    "email": "alice@email.com",
    "password": "supersecure"
  }
  ```
  - On success: User receives a verification email.

- `GET /verify-email?token=...` — Verify user’s email  
  - User clicks the link sent via email to activate the account.

- `POST /login` — Authenticate and get tokens (only for verified users)  
  **Body:**
  ```json
  {
    "email": "alice@email.com",
    "password": "supersecure"
  }
  ```
  **Success Response:**
  ```json
  {
    "access_token": "JWT...",
    "refresh_token": "..."
  }
  ```

- `POST /refresh-token` — Get new tokens using a valid refresh token  
  **Body:**
  ```json
  {
    "refresh_token": "existing_refresh_token"
  }
  ```
  - Returns new `access_token` and `refresh_token`.
  - Invalidates the previous refresh token for security (rotation).

- `POST /logout` — Log out and invalidate a refresh token  
  - **Requires:**  
    - `Authorization: Bearer <access_token>`
    - **Body:**
    ```json
    {
      "refresh_token": "the_token_to_invalidate"
    }
    ```

- `GET /profile` — Get current user’s profile  
  - **Requires:** valid `Authorization: Bearer <access_token>`

- `GET /users` — List all users *(admin only)*  
  - **Requires:** admin's authorization header

## 🧪 Running Tests

```sh
go test ./tests/...
```
- Dedicated unit and isolation tests for controllers, services, repositories using mocks.
- Tests cover registration, login, failed login, token lifecycles, role guards, email verification logic, and more.

## 📂 Project Structure

```
.
├── controllers/         # HTTP handlers for API endpoints
├── dto/                 # Request/response data (DTOs)
├── models/              # GORM models (User, RefreshToken, etc.)
├── repositories/        # Interface & DB logic
├── services/            # Business logic (authentication, user, email, etc.)
├── middleware/          # JWT & RBAC middleware
├── config/              # Env, DB, JWT, and email settings
├── tests/               # Unit & mock tests
└── main.go              # Application entry
```

## 🤝 Contributing

1. Fork this repo
2. Create a feature branch (`git checkout -b feat/your-feature`)
3. Commit and push your changes
4. Open a Pull Request!

## 📄 License

MIT License

---
