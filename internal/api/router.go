package api

import (
	"encoding/json"
	"net/http"
	"os"

	"meueu/internal/api/handlers"
	"meueu/internal/api/middleware"
	"meueu/internal/core/identity"
)

// RouterConfig defines the dependencies for setting up the router.
type RouterConfig struct {
	PublishHandler *handlers.PublishHandler
	QueryHandler   *handlers.QueryHandler
	AdminHandler   *handlers.AdminHandler
	PeerHandler    *handlers.PeerHandler
	SyncHandler    *handlers.SyncHandler
	Middleware     *middleware.Container
	ServerIdentity identity.ServerIdentity
}

// NewRouter creates the fully-wired HTTP handler with all routes and middleware.
// This centralizes route registration and makes the endpoint tree visible at a glance.
func NewRouter(cfg RouterConfig) http.Handler {
	mux := http.NewServeMux()

	// ── Public API (v1) ──────────────────────────────────────────────
	// Publish: Full governance chain (RateLimiter → Banlist → ContentFilter → Whitelist)
	mux.Handle("/v1/publish", cfg.Middleware.StandardSecurity(http.HandlerFunc(cfg.PublishHandler.Handle)))

	// Query: Whitelist-only (if WHITELIST mode)
	mux.Handle("/v1/query", cfg.Middleware.QuerySecurity(http.HandlerFunc(cfg.QueryHandler.Handle)))

	// ── Discovery & Sync ─────────────────────────────────────────────
	mux.HandleFunc("/v1/peers", cfg.PeerHandler.HandleList)
	mux.HandleFunc("/v1/sync", cfg.SyncHandler.Handle) // GET historical Sync
	
	// Bulk Import (Restore Backup) goes through standard security (Rate Limiter etc)
	mux.Handle("/v1/import", cfg.Middleware.StandardSecurity(http.HandlerFunc(cfg.SyncHandler.HandleImport)))

	// ── Transparency (Required by License) ───────────────────────────
	mux.HandleFunc("/meta", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"server":        "Crom/Meueu Node",
			"version":       "2.0.0-sovereign",
			"source_url":    "https://github.com/your-username/crom-meueu",
			"mode":          os.Getenv("SERVER_MODE"),
			"network_id":    os.Getenv("NETWORK_ID"),
			"sync_enabled":  os.Getenv("SYNC_EXTERNAL_POSTS"),
			"server_pubkey": cfg.ServerIdentity.PubKeyHex,
		})
	})

	// ── Admin API (v1/admin/*) — Protected by AuthAdmin ──────────────
	adminMux := http.NewServeMux()
	adminMux.HandleFunc("/stats", cfg.AdminHandler.HandleStats)

	adminMux.HandleFunc("/whitelist", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			cfg.AdminHandler.HandleWhitelistList(w, r)
		case http.MethodPost:
			cfg.AdminHandler.HandleWhitelistAdd(w, r)
		case http.MethodDelete:
			cfg.AdminHandler.HandleWhitelistRemove(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	adminMux.HandleFunc("/banned_words", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			cfg.AdminHandler.HandleListBannedWords(w, r)
		case http.MethodPost:
			cfg.AdminHandler.HandleBanWord(w, r)
		case http.MethodDelete:
			cfg.AdminHandler.HandleUnbanWord(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	adminMux.HandleFunc("/banned_users", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			cfg.AdminHandler.HandleListBannedUsers(w, r)
		case http.MethodPost:
			cfg.AdminHandler.HandleBanUser(w, r)
		case http.MethodDelete:
			cfg.AdminHandler.HandleUnbanUser(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	adminMux.HandleFunc("/node", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			cfg.AdminHandler.HandleDeleteNode(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	adminMux.HandleFunc("/banned_hashes", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			cfg.AdminHandler.HandleListBannedHashes(w, r)
		case http.MethodPost:
			cfg.AdminHandler.HandleBanHash(w, r)
		case http.MethodDelete:
			cfg.AdminHandler.HandleUnbanHash(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Mount admin routes with auth — support both /admin/ and /v1/admin/ for backward compatibility
	adminWithAuth := middleware.AuthAdmin(adminMux)
	mux.Handle("/v1/admin/", http.StripPrefix("/v1/admin", adminWithAuth))
	mux.Handle("/admin/", http.StripPrefix("/admin", adminWithAuth))

	// ── Static Files (Frontend) ──────────────────────────────────────
	fs := http.FileServer(http.Dir("./frontend"))
	mux.Handle("/", fs)

	// ── Global Middleware (CORS) ─────────────────────────────────────
	return cfg.Middleware.CORS(mux)
}
