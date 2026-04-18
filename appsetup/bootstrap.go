package appsetup

import (
	"log"

	"go-api-starterkit/config"
	"go-api-starterkit/seeds"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gorm.io/gorm"
)

func RunStartupTasks(cfg config.AppConfig, db *gorm.DB) {
	if cfg.AutoRunMigrations {
		if err := RunMigrations(cfg.DatabaseURL); err != nil {
			log.Fatalf("❌ startup migrations failed: %v", err)
		}
		log.Println("✅ startup migrations completed")
	}

	if cfg.AutoRunSeeds {
		RunSeeds(db, cfg)
		log.Println("✅ startup seeds completed")
	}
}

func RunMigrations(dbURL string) error {
	if dbURL == "" {
		log.Fatal("❌ DATABASE_URL is not set")
	}

	m, err := migrate.New("file://migrations", dbURL)
	if err != nil {
		return err
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}

func RunSeeds(db *gorm.DB, cfg config.AppConfig) {
	if db == nil {
		log.Fatal("❌ DB is not initialized before seeding")
	}

	seeds.SeedRoles(db)
	seeds.SeedPermissions(db)
	seeds.SeedRolePermissions(db)
	seeds.SeedAdmin(db, cfg)
}
