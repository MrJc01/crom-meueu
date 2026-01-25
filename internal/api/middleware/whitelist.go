package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"

	"meueu/internal/storage/postgres"
)

type WhitelistMiddleware struct {
	repo *postgres.AdminRepository
}

func NewWhitelistMiddleware(repo *postgres.AdminRepository) *WhitelistMiddleware {
	return &WhitelistMiddleware{repo: repo}
}

func (m *WhitelistMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only check on POST (Write operations)
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}

		// Check Server Mode
		mode := os.Getenv("SERVER_MODE")
		if mode != "WHITELIST" {
			next.ServeHTTP(w, r)
			return
		}

		// Read Body
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusInternalServerError)
			return
		}

		// Restore Body for next handler
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// Parse just what we need
		var req struct {
			AuthorPubkey string `json:"author_pubkey"`
		}
		if err := json.Unmarshal(bodyBytes, &req); err != nil {
			// If JSON is invalid, let the next handler (PublishHandler) deal with it or fail here.
			// Failing here is safer for security middleware.
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if req.AuthorPubkey == "" {
			http.Error(w, "Missing author_pubkey", http.StatusBadRequest)
			return
		}

		// Check Whitelist
		whitelisted, err := m.repo.IsWhitelisted(r.Context(), req.AuthorPubkey)
		if err != nil {
			// Log error?
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if !whitelisted {
			http.Error(w, "Access Denied: You are not whitelisted on this server.", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
