package config

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"go-auth-app/models"
)

var DB *gorm.DB

func ConnectDB() {

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		GetEnv("DB_HOST", "localhost"),
		GetEnv("DB_USER", "postgres"),
		GetEnv("DB_PASSWORD", "postgres"),
		GetEnv("DB_NAME", "go_auth"),
		GetEnv("DB_PORT", "5432"),
		GetEnv("DB_SSLMODE", "disable"),
	)

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}

	log.Println("Database connected")

	database.AutoMigrate(&models.User{})

	DB = database
}
