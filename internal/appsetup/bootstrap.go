package appsetup

import (
	"log"

	"pleco-api/internal/config"
	"pleco-api/internal/seeds"
	migrationFiles "pleco-api/migrations"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
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

	sourceDriver, err := iofs.New(migrationFiles.Files, ".")
	if err != nil {
		return err
	}

	m, err := migrate.NewWithSourceInstance("iofs", sourceDriver, dbURL)
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
