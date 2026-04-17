package main

import (
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"go-auth-app/config"
)

func getDatabaseURL() string {
	// Prefer DATABASE_URL, fallback to Makefile-style composition
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL != "" {
		return dbURL
	}

	// Compose from legacy variables
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	dbSSLMode := os.Getenv("DB_SSLMODE")
	if dbSSLMode == "" {
		dbSSLMode = "disable"
	}

	if dbUser == "" || dbPassword == "" || dbHost == "" || dbPort == "" || dbName == "" {
		log.Fatal("Missing database connection environment variables. Set DATABASE_URL or all of DB_USER, DB_PASSWORD, DB_HOST, DB_PORT, DB_NAME.")
	}

	return "postgres://" + dbUser + ":" + dbPassword + "@" + dbHost + ":" + dbPort + "/" + dbName + "?sslmode=" + dbSSLMode
}

func main() {
	config.LoadEnv() // Loads from .env by default

	dbURL := getDatabaseURL()

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
