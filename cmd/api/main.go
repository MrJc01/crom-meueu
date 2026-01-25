package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"meueu/internal/api/handlers"
	"meueu/internal/api/middleware"
	"meueu/internal/storage/postgres"

	"github.com/joho/godotenv"
)

func main() {
	log.Println("Starting Meueu (Crom Social Protocol) Node...")

	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Println("Note: No .env file found, relying on system environment variables.")
	}

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
	// Run Migrations (Auto-Schema)
	dbURL := os.Getenv("DATABASE_URL")
	if err := postgres.RunMigrations(dbURL); err != nil {
		log.Printf("Migration warning: %v", err)
		// Don't die, maybe just connection issue or already locked
	}

	nodeRepo := postgres.NewNodeRepository(dbPool)
	adminRepo := postgres.NewAdminRepository(dbPool)

	publishHandler := handlers.NewPublishHandler(nodeRepo)
	queryHandler := handlers.NewQueryHandler(nodeRepo)
	adminHandler := handlers.NewAdminHandler(adminRepo)

	// Middlewares
	whitelistMiddleware := middleware.NewWhitelistMiddleware(adminRepo)
	contentFilterMiddleware := middleware.NewContentFilterMiddleware(adminRepo)
	banlistMiddleware := middleware.NewBanlistMiddleware(adminRepo)

	// 4. Setup Routes & Middleware
	mux := http.NewServeMux()

	// Public API Routes (v1)
	// Apply Governance Middlewares to Publish Endpoint
	// Chain: Banlist -> ContentFilter -> Whitelist -> Publish
	// Note: We wrap the handler specifically
	publishChain := banlistMiddleware.Middleware(
		contentFilterMiddleware.Middleware(
			whitelistMiddleware.Middleware(http.HandlerFunc(publishHandler.Handle)),
		),
	)

	mux.Handle("/v1/publish", publishChain)

	// Apply Whitelist to Query as well (if strict mode enabled)
	// QueryHandler is read-only usually, but user requested "Rotas Públicas... middleware CheckWhitelist"
	// ContentFilter is typically for Write, so we skip it for Query.
	queryChain := whitelistMiddleware.Middleware(http.HandlerFunc(queryHandler.Handle))
	mux.Handle("/v1/query", queryChain)

	// Transparency Route (Required by License)
	mux.HandleFunc("/meta", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"server":     "Crom/Meueu Node",
			"version":    "0.2.0-governance",
			"source_url": "https://github.com/your-username/crom-meueu", // Should be configured via env realistically
			"mode":       os.Getenv("SERVER_MODE"),
		})
	})

	// Admin API Routes (Protected)
	adminMux := http.NewServeMux()
	adminMux.HandleFunc("/stats", adminHandler.HandleStats)
	adminMux.HandleFunc("/whitelist", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			adminHandler.HandleWhitelistList(w, r)
		case http.MethodPost:
			adminHandler.HandleWhitelistAdd(w, r)
		case http.MethodDelete:
			adminHandler.HandleWhitelistRemove(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	adminMux.HandleFunc("/banned_words", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			adminHandler.HandleListBannedWords(w, r)
		case http.MethodPost:
			adminHandler.HandleBanWord(w, r)
		case http.MethodDelete:
			adminHandler.HandleUnbanWord(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// User Bans
	adminMux.HandleFunc("/banned_users", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			adminHandler.HandleListBannedUsers(w, r)
		case http.MethodPost:
			adminHandler.HandleBanUser(w, r)
		case http.MethodDelete:
			adminHandler.HandleUnbanUser(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Content Deletion
	adminMux.HandleFunc("/node", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			adminHandler.HandleDeleteNode(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Mount Admin Routes with Auth Middleware
	mux.Handle("/admin/", http.StripPrefix("/admin", middleware.AuthAdmin(adminMux)))

	// Static Files (Frontend)
	fs := http.FileServer(http.Dir("./frontend"))
	mux.Handle("/", fs)

	// Wrap with Global Middleware (CORS)
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
