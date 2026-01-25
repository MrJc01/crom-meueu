package middleware

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"meueu/internal/storage/postgres"
)

type ContentFilterMiddleware struct {
	repo *postgres.AdminRepository
}

func NewContentFilterMiddleware(repo *postgres.AdminRepository) *ContentFilterMiddleware {
	return &ContentFilterMiddleware{repo: repo}
}

func (m *ContentFilterMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}

		// Get Banned Words
		// Optimization: In a real high-traffic app, this should be cached in memory
		// and updated via channel or polling. Checking DB on every POST is expensive-ish
		// but acceptable for MVP/Proof of Concept.
		words, err := m.repo.GetBannedWords(r.Context())
		if err != nil {
			// Fail open or fail closed?
			// Fail open: log err, allow post.
			// Fail closed: return 500.
			// Let's fail open for availability but log it (can't log easily without logger dependency).
			// We'll proceed.
			next.ServeHTTP(w, r)
			return
		}

		if len(words) == 0 {
			next.ServeHTTP(w, r)
			return
		}

		// Read Body
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusInternalServerError)
			return
		}

		// Restore Body
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// Check Content
		// We define "Content" as the raw JSON body for simplicity.
		// If a banned word is in a tag or attribute name, it blocks too.
		// This is a "Hammer" approach.
		bodyStr := string(bodyBytes)
		lowerBody := strings.ToLower(bodyStr)

		for _, word := range words {
			if strings.Contains(lowerBody, strings.ToLower(word)) {
				http.Error(w, "Content Rejected: Contains prohibited words.", http.StatusBadRequest)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
