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

		// Fail-open strategy
		words, err := m.repo.GetBannedWords(r.Context())
		if err != nil || len(words) == 0 {
			next.ServeHTTP(w, r)
			return
		}

		// Read Body (Limit 1MB)
		bodyBytes, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1048576))
		if err != nil {
			http.Error(w, "Request body too large for analysis", http.StatusRequestEntityTooLarge)
			return
		}

		// Restore Body
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// Parse JSON to normalize escapes
		// Fixes "Evasão de Filtro": \u0061 becomes 'a' automatically.
		var payload interface{}
		if err := json.Unmarshal(bodyBytes, &payload); err != nil {
			// If JSON is invalid, pass valid requests down (let handler handle it)
			// But since we are protection layer, maybe we should swallow?
			// Policy: Pass it.
			next.ServeHTTP(w, r)
			return
		}

		if containsBannedWords(payload, words) {
			http.Error(w, "Content Rejected: Contains prohibited words.", http.StatusBadRequest)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// containsBannedWords recursively checks strings in a generic JSON structure.
// json.Unmarshal automatically decodes escapes like \u0061 -> 'a'.
func containsBannedWords(data interface{}, banned []string) bool {
	switch v := data.(type) {
	case string:
		// Normalize: Lowercase for comparison
		norm := strings.ToLower(v)
		for _, word := range banned {
			// Basic substring check on normalized text
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
