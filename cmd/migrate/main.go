package main

import (
	"flag"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var direction string
	flag.StringVar(&direction, "direction", "up", "migration direction (up or down)")
	flag.Parse()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://user:password@localhost:5432/meueu?sslmode=disable"
	}

	m, err := migrate.New(
		"file://db/migrations",
		dbURL,
	)
	if err != nil {
		log.Fatalf("Migration initialization failed: %v", err)
	}

	log.Printf("Running migrations: %s...", direction)

	if direction == "up" {
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Migration UP failed: %v", err)
		}
	} else if direction == "down" {
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Migration DOWN failed: %v", err)
		}
	} else {
		log.Fatalf("Invalid direction: %s", direction)
	}

	log.Println("Migration completed successfully.")
}
