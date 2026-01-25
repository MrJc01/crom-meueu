package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"os"
)

func AuthAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Get Secret Hash from Env
		expectedHash := os.Getenv("ADMIN_SECRET_HASH")
		if expectedHash == "" {
			// If not set, strictly fail safe (or maybe log warning and deny all)
			http.Error(w, "Admin authentication not configured", http.StatusInternalServerError)
			return
		}

		// 2. Get Token from Header
		token := r.Header.Get("X-Admin-Token")
		if token == "" {
			http.Error(w, "Unauthorized: Missing Token", http.StatusUnauthorized)
			return
		}

		// 3. Hash the received token
		hash := sha256.Sum256([]byte(token))
		hashedToken := hex.EncodeToString(hash[:])

		// 4. Compare
		// Note: ConstantTimeCompare is better for security but for this MVP standard string compare is "okay"
		// provided we are careful. But let's stick to simple logic.
		if hashedToken != expectedHash {
			http.Error(w, "Unauthorized: Invalid Token", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
