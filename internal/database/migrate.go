package database

import (
	"embed"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func RunMigrations(databaseURL string) error {
	source, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		log.Printf("[ERROR] migration source: %v", err)
		return fmt.Errorf("iofs: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", source, databaseURL)
	if err != nil {
		log.Printf("[ERROR] migration setup: %v", err)
		return fmt.Errorf("migrate.New: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Printf("[ERROR] migration up: %v", err)
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}
