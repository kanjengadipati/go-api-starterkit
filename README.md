# Go Auth App

A simple, modular authentication API built with Go and Gin.  
Supports user registration, login, JWT authentication, and protected routes.

## 🚀 Features

- User registration and login
- JWT-based authentication
- Password hashing with bcrypt
- Secure protected routes (e.g., `/profile`)
- Simple user roles (admin, user)
- Repository and model separation for easy testing/mocking
- Unit tests with mock repositories

## 🛠️ Getting Started

### Prerequisites

- Go 1.18+ ([Download](https://golang.org/dl/))
- Docker (optional, for containerization)

### Installation

Clone this repository:

```sh
git clone https://github.com/your-username/go-auth-app.git
cd go-auth-app
```

Install dependencies:

```sh
go mod tidy
```

### Configuration

- Update your database and secret configuration in `config/` as needed.

### Running the App

Start the API server:

```sh
go run main.go
```

- Server will start at `http://localhost:8080` by default.

## 📚 API Endpoints

- `POST /register` — Register a new user  
  **Body:**  
  ```json
  {
    "name": "Your Name",
    "email": "email@example.com",
    "password": "yourpassword"
  }
  ```

- `POST /login` — User login  
  **Body:**  
  ```json
  {
    "email": "email@example.com",
    "password": "yourpassword"
  }
  ```

- `GET /profile` — Fetch authenticated user's profile (requires JWT in Authorization header)

- `POST /logout` — Logout a user (handle JWT removal/invalidation as needed)

## 🧪 Running Tests

Run all tests with:

```sh
go test ./tests/...
```

- Includes unit tests for controllers and repository with mocks.

## 📂 Project Structure

```
.
├── controllers/         # HTTP route handlers
├── dto/                 # Data transfer objects
├── models/              # GORM data models
├── repositories/        # Repository implementations and interfaces
├── tests/               # Unit tests and repository mocks
├── config/              # Configuration (DB, JWT)
└── main.go              # Application entrypoint
```

## 🤝 Contributing

1. Fork this repository
2. Create a new branch (`git checkout -b feature/your-feature`)
3. Commit your changes (`git commit -am 'feat: add your feature'`)
4. Push to your fork (`git push origin feature/your-feature`)
5. Open a Pull Request

## 📄 License

MIT License

---
