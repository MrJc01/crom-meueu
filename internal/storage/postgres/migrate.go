package postgres

import (
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(dbURL string) error {
	if dbURL == "" {
		dbURL = "postgres://user:password@localhost:5432/meueu?sslmode=disable"
	}

	m, err := migrate.New(
		"file://db/migrations",
		dbURL,
	)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	log.Println("Database migrations applied successfully!")
	return nil
}
