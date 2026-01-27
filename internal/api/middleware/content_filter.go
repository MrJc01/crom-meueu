package middleware

import (
	"bytes"
	"encoding/json"
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

		// Fail-open strategy: If DB is unreachable, we log error (optional) and allow traffic.
		words, err := m.repo.GetBannedWords(r.Context())
		if err != nil || len(words) == 0 {
			next.ServeHTTP(w, r)
			return
		}

		// Read Body (Limit to 1MB to match Publish Handler or use smaller limit for filter)
		bodyBytes, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1048576))
		if err != nil {
			http.Error(w, "Request body too large for analysis", http.StatusRequestEntityTooLarge)
			return
		}

		// Restore Body for the next handler
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// Parse JSON to check values only (avoiding keys or JSON structure overhead)
		var payload interface{}
		if err := json.Unmarshal(bodyBytes, &payload); err != nil {
			// If it's not valid JSON, we can't inspect deeply.
			// Depending on policy: Block or Pass?
			// Publish handler requires JSON. If this fails, Publish will likely fail too.
			// We pass it down to let the specific handler decide, or block here if we are strict.
			// Let's pass it down.
			next.ServeHTTP(w, r)
			return
		}

		// Recursive check
		if containsBannedWords(payload, words) {
			http.Error(w, "Content Rejected: Contains prohibited words.", http.StatusBadRequest)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// containsBannedWords recursively checks strings in a generic JSON structure
func containsBannedWords(data interface{}, banned []string) bool {
	switch v := data.(type) {
	case string:
		// Normalize: Lowercase for comparison
		norm := strings.ToLower(v)
		for _, word := range banned {
			if strings.Contains(norm, strings.ToLower(word)) {
				return true
			}
		}
	case map[string]interface{}:
		for _, val := range v {
			if containsBannedWords(val, banned) {
				return true
			}
		}
	case []interface{}:
		for _, val := range v {
			if containsBannedWords(val, banned) {
				return true
			}
		}
	}
	return false
}
