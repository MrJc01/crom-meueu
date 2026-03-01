package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"meueu/internal/api"
	"meueu/internal/api/handlers"
	"meueu/internal/api/middleware"
	"meueu/internal/core/services"
	"meueu/internal/storage/postgres"

	"github.com/joho/godotenv"
)

func main() {
	log.Println("Starting Meueu (Crom Social Protocol) Node...")

	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Println("Note: No .env file found, relying on system environment variables.")
	}

	// 1. Context & Graceful Shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		log.Println("Shutting down...")
		cancel()
	}()

	// 2. Database
	log.Println("Connecting to PostgreSQL...")
	initCtx, initCancel := context.WithTimeout(ctx, 5*time.Second)
	defer initCancel()

	dbPool, err := postgres.NewConnection(initCtx)
	if err != nil {
		log.Printf("Failed to connect to database: %v\n", err)
		log.Println("Continuing without DB for now (Development Mode)...")
	} else {
		defer dbPool.Close()
		log.Println("Successfully connected to database.")
	}

	// Auto-Migrate
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		if err := postgres.RunMigrations(dbURL); err != nil {
			log.Printf("Migration warning: %v", err)
		}
	}

	// 3. Repositories
	nodeRepo := postgres.NewNodeRepository(dbPool)
	adminRepo := postgres.NewAdminRepository(dbPool)
	peerRepo := postgres.NewPeerRepository(dbPool)
	nonceRepo := postgres.NewNonceRepository(dbPool)

	// 4. Handlers
	publishHandler := handlers.NewPublishHandler(nodeRepo, nonceRepo)
	queryHandler := handlers.NewQueryHandler(nodeRepo)
	adminHandler := handlers.NewAdminHandler(adminRepo)
	peerHandler := handlers.NewPeerHandler(peerRepo)
	syncHandler := handlers.NewSyncHandler(nodeRepo)

	// 5. Middleware Container
	mw := middleware.NewContainer(adminRepo)

	// 6. Background Services (only with DB)
	if dbPool != nil {
		myURL := os.Getenv("SERVER_URL")
		if myURL == "" {
			myURL = "http://localhost:8080"
		}
		var seedNodes []string
		if s := os.Getenv("SEED_NODES"); s != "" {
			seedNodes = strings.Split(s, ",")
		}

		peerDiscovery := services.NewPeerDiscoveryService(peerRepo, nodeRepo, myURL, "local-pubkey", seedNodes)
		peerDiscovery.Start(ctx)

		// Nonce TTL Cleanup (every 1h, remove nonces > 24h)
		go func() {
			ticker := time.NewTicker(1 * time.Hour)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					deleted, err := nonceRepo.CleanupExpired(ctx, 24*time.Hour)
					if err != nil {
						log.Printf("[NonceTTL] Cleanup error: %v", err)
					} else if deleted > 0 {
						log.Printf("[NonceTTL] Cleaned up %d expired nonces", deleted)
					}
				}
			}
		}()
	} else {
		log.Println("[PeerDiscovery] Skipped — no database connection")
		log.Println("[NonceTTL] Skipped — no database connection")
	}

	// 7. Router (all routes registered in one place)
	handler := api.NewRouter(publishHandler, queryHandler, adminHandler, peerHandler, syncHandler, mw)

	// 8. Start Server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	go func() {
		log.Println("Server listening on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Goodbye.")
}
