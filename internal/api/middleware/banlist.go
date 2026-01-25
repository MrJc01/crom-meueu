package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"meueu/internal/storage/postgres"
)

type BanlistMiddleware struct {
	repo *postgres.AdminRepository
}

func NewBanlistMiddleware(repo *postgres.AdminRepository) *BanlistMiddleware {
	return &BanlistMiddleware{repo: repo}
}

func (m *BanlistMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusInternalServerError)
			return
		}
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		var req struct {
			AuthorPubkey string `json:"author_pubkey"`
		}
		if err := json.Unmarshal(bodyBytes, &req); err != nil {
			// Malformed JSON (or different endpoint structure). Skip check or fail?
			// Since this is a global security middleware, better to be permissive IF it's not a node publish?
			// But this is attached to /publish.
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if req.AuthorPubkey == "" {
			http.Error(w, "Missing author_pubkey", http.StatusBadRequest)
			return
		}

		// Check Ban with Repository
		isBanned, err := m.repo.IsBannedUser(r.Context(), req.AuthorPubkey)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if isBanned {
			http.Error(w, "Access Denied: You are banned from this server.", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
