package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"meueu/internal/api/handlers"
	"meueu/internal/api/middleware"
	"meueu/internal/storage/postgres"
)

func main() {
	log.Println("Starting Meueu (Crom Social Protocol) Node...")

	// 1. Setup Context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		log.Println("Shutting down...")
		cancel()
	}()

	// 2. Connect to Database
	log.Println("Connecting to PostgreSQL...")
	// Note: We use a short timeout for the initial connection check
	initCtx, initCancel := context.WithTimeout(ctx, 5*time.Second)
	defer initCancel()

	dbPool, err := postgres.NewConnection(initCtx)
	if err != nil {
		log.Printf("Failed to connect to database: %v\n", err)
		log.Println("Continuing without DB for now (Development Mode)...")
		// In production, we might want to os.Exit(1) here
	} else {
		defer dbPool.Close()
		log.Println("Successfully connected to database.")
	}

	// 3. Setup core components
	nodeRepo := postgres.NewNodeRepository(dbPool)
	publishHandler := handlers.NewPublishHandler(nodeRepo)
	queryHandler := handlers.NewQueryHandler(nodeRepo)

	// 4. Setup Routes & Middleware
	mux := http.NewServeMux()

	// API Routes (v1)
	mux.HandleFunc("/v1/publish", publishHandler.Handle)
	mux.HandleFunc("/v1/query", queryHandler.Handle)

	// Static Files (Frontend)
	fs := http.FileServer(http.Dir("./frontend"))
	mux.Handle("/", fs)

	// Wrap with Middleware (CORS)
	handler := middleware.CORS(mux)

	// 5. Start Server
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

	// Keep alive until context is cancelled
	<-ctx.Done()

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Goodbye.")
}
