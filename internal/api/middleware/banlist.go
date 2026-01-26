package middleware

import (
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
		// Only check banlist for POST/PUT methods where we accept data
		if r.Method != http.MethodPost && r.Method != http.MethodPut {
			next.ServeHTTP(w, r)
			return
		}

		// Security Optimization: Read Author Key from Header to avoid reading Body (DoS mitigation)
		authorPubkey := r.Header.Get("X-MeuEu-Author")

		if authorPubkey == "" {
			// Strict Mode: Require the header.
			// This forces clients to upgrade but protects against large body attacks.
			http.Error(w, "Missing X-MeuEu-Author header", http.StatusBadRequest)
			return
		}

		// Basic format validation (hex string of 32 bytes = 64 chars)
		if len(authorPubkey) != 64 {
			http.Error(w, "Invalid X-MeuEu-Author header format", http.StatusBadRequest)
			return
		}

		// Check Ban with Repository
		isBanned, err := m.repo.IsBannedUser(r.Context(), authorPubkey)
		if err != nil {
			// Fail open or closed? Security says fail closed usually, but availability says fail open if DB is down.
			// For banlist, if DB is down, maybe we shouldn't block everyone?
			// But for strict security, we return 500.
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if isBanned {
			http.Error(w, "Access Denied: You are banned from this server.", http.StatusForbidden)
			return
		}

		// Security: Limit Request Body Size to 1MB to prevent memory exhaustion DoS
		// This applies to the subsequent handlers that read the body (like PublishHandler)
		r.Body = http.MaxBytesReader(w, r.Body, 1048576) // 1MB

		next.ServeHTTP(w, r)
	})
}
