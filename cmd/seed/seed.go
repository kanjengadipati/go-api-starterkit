package main

import (
	"log"

	"pleco-api/internal/appsetup"
	"pleco-api/internal/config"
)

func main() {
	// Load env (WAJIB)
	config.LoadEnv()
	appConfig := config.LoadAppConfig()

	// Init DB (WAJIB)
	db := config.ConnectDB(appConfig.DatabaseURL)
	log.Println("Start seeding...")
	appsetup.RunSeeds(db, appConfig)

	log.Println("Seeding done 🚀")
}
