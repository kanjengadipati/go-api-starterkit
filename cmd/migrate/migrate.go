package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"go-auth-app/config"
)

func main() {
	// No need to call config.ConnectDB() here since we're not using GORM, only gathering env vars

	// Use GetEnv from config/db.go directly
	host := config.GetEnv("DB_HOST", "localhost")
	user := config.GetEnv("DB_USER", "postgres")
	password := config.GetEnv("DB_PASSWORD", "postgres")
	dbname := config.GetEnv("DB_NAME", "go_auth")
	port := config.GetEnv("DB_PORT", "5432")
	sslmode := config.GetEnv("DB_SSLMODE", "disable")

	// Properly encode special characters in user/password for URL compatibility
	userEscaped := strings.ReplaceAll(user, " ", "%20")
	passwordEscaped := strings.ReplaceAll(password, " ", "%20")

	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", userEscaped, passwordEscaped, host, port, dbname, sslmode)

	m, err := migrate.New(
		"file://migrations",
		dbURL,
	)
	if err != nil {
		log.Fatalf("Failed to create migrate instance: %v", err)
	}

	// .Up() returns migrate.ErrNoChange if already up to date
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Migration success")
}
