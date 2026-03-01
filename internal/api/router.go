package api

import (
	"encoding/json"
	"net/http"
	"os"

	"meueu/internal/api/handlers"
	"meueu/internal/api/middleware"
)

// NewRouter creates the fully-wired HTTP handler with all routes and middleware.
// This centralizes route registration and makes the endpoint tree visible at a glance.
func NewRouter(
	publishHandler *handlers.PublishHandler,
	queryHandler *handlers.QueryHandler,
	adminHandler *handlers.AdminHandler,
	peerHandler *handlers.PeerHandler,
	syncHandler *handlers.SyncHandler,
	mw *middleware.Container,
) http.Handler {
	mux := http.NewServeMux()

	// ── Public API (v1) ──────────────────────────────────────────────
	// Publish: Full governance chain (RateLimiter → Banlist → ContentFilter → Whitelist)
	mux.Handle("/v1/publish", mw.StandardSecurity(http.HandlerFunc(publishHandler.Handle)))

	// Query: Whitelist-only (if WHITELIST mode)
	mux.Handle("/v1/query", mw.QuerySecurity(http.HandlerFunc(queryHandler.Handle)))

	// ── Discovery & Sync ─────────────────────────────────────────────
	mux.HandleFunc("/v1/peers", peerHandler.HandleList)
	mux.HandleFunc("/v1/sync", syncHandler.Handle)

	// ── Transparency (Required by License) ───────────────────────────
	mux.HandleFunc("/meta", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"server":       "Crom/Meueu Node",
			"version":      "2.0.0-sovereign",
			"source_url":   "https://github.com/your-username/crom-meueu",
			"mode":         os.Getenv("SERVER_MODE"),
			"network_id":   os.Getenv("NETWORK_ID"),
			"sync_enabled": os.Getenv("SYNC_EXTERNAL_POSTS"),
		})
	})

	// ── Admin API (v1/admin/*) — Protected by AuthAdmin ──────────────
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

	adminMux.HandleFunc("/node", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			adminHandler.HandleDeleteNode(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	adminMux.HandleFunc("/banned_hashes", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			adminHandler.HandleListBannedHashes(w, r)
		case http.MethodPost:
			adminHandler.HandleBanHash(w, r)
		case http.MethodDelete:
			adminHandler.HandleUnbanHash(w, r)
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
	return mw.CORS(mux)
}
