package main

import (
	"log"

	"go-auth-app/config"
	"go-auth-app/seeds"
)

func main() {
	// Load env (WAJIB)
	config.LoadEnv()

	// Init DB (WAJIB)
	config.ConnectDB()

	// Run seeder
	seeds.SeedRoles(config.DB)
	seeds.SeedPermissions(config.DB)
	seeds.SeedAdmin(config.DB)

	log.Println("Seeding done 🚀")
}
