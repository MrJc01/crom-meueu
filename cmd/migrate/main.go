package main

import (
	"log"
	"os"

	"meueu/internal/storage/postgres"
)

func main() {
	log.Println("🛠️ Manual Migration Tool Starting...")

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://user:password@localhost:5432/meueu?sslmode=disable"
	}

	log.Printf("Target DB: %s", dbURL)

	if err := postgres.RunMigrations(dbURL); err != nil {
		log.Fatalf("❌ Migration Failed: %v", err)
	}

	log.Println("✅ Migrations Completed Successfully.")
}
