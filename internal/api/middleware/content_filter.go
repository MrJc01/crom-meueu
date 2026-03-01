package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
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
			// If JSON is invalid, pass it down (let handler handle it)
			next.ServeHTTP(w, r)
			return
		}

		// --- Check 1: Banned Words ---
		// Fail-open strategy: if DB error, allow through
		words, err := m.repo.GetBannedWords(r.Context())
		if err == nil && len(words) > 0 {
			if containsBannedWords(payload, words) {
				http.Error(w, "Content Rejected: Contains prohibited words.", http.StatusBadRequest)
				return
			}
		}

		// --- Check 2: Banned Hashes (SHA256 of signature) ---
		// Extract signature from the parsed JSON and check against banned hashes
		if payloadMap, ok := payload.(map[string]interface{}); ok {
			if sig, ok := payloadMap["signature"].(string); ok && sig != "" {
				sigHash := sha256.Sum256([]byte(sig))
				sigHashHex := hex.EncodeToString(sigHash[:])

				isBanned, err := m.repo.IsBannedHash(r.Context(), sigHashHex)
				if err == nil && isBanned {
					http.Error(w, "Content Rejected: This content has been blocked by hash.", http.StatusForbidden)
					return
				}
			}
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
