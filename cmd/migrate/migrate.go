package main

import (
	"log"

	"go-api-starterkit/appsetup"
	"go-api-starterkit/config"
)

func main() {
	config.LoadEnv()
	appConfig := config.LoadAppConfig()
	if err := appsetup.RunMigrations(appConfig.DatabaseURL); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Migration success")
}
