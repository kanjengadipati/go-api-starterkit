package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// LoadEnv loads environment variables from a .env file, if present.
func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or failed to load .env")
	}
}

// GetEnv returns the environment variable value for the given key,
// or the provided fallback if the variable is not set or empty.
func GetEnv(key, fallback string) string {
	if val, exists := os.LookupEnv(key); exists && val != "" {
		return val
	}
	return fallback
}
