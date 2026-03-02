package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"sync"
	"time"
)

// RateLimiter implements a simple in-memory token bucket rate limiter.
// It limits requests by both IP address and public key.
type RateLimiter struct {
	ipBuckets     sync.Map // map[string]*bucket
	pubKeyBuckets sync.Map // map[string]*bucket
	ipRate        int      // max requests per window
	pubKeyRate    int      // max requests per window
	window        time.Duration
}

type bucket struct {
	tokens    int
	lastReset time.Time
	mu        sync.Mutex
}

// NewRateLimiter creates a new rate limiter.
// ipRate: max requests per window per IP (e.g., 100)
// pubKeyRate: max requests per window per PubKey (e.g., 30)
// window: time window for rate limiting (e.g., 1 minute)
func NewRateLimiter(ipRate, pubKeyRate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		ipRate:     ipRate,
		pubKeyRate: pubKeyRate,
		window:     window,
	}

	// Periodic cleanup of stale buckets (every 5 minutes)
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			now := time.Now()
			rl.ipBuckets.Range(func(key, value interface{}) bool {
				b := value.(*bucket)
				b.mu.Lock()
				if now.Sub(b.lastReset) > 10*time.Minute {
					rl.ipBuckets.Delete(key)
				}
				b.mu.Unlock()
				return true
			})
			rl.pubKeyBuckets.Range(func(key, value interface{}) bool {
				b := value.(*bucket)
				b.mu.Lock()
				if now.Sub(b.lastReset) > 10*time.Minute {
					rl.pubKeyBuckets.Delete(key)
				}
				b.mu.Unlock()
				return true
			})
		}
	}()

	return rl
}

func (rl *RateLimiter) allow(buckets *sync.Map, key string, maxTokens int) bool {
	now := time.Now()

	val, _ := buckets.LoadOrStore(key, &bucket{
		tokens:    maxTokens,
		lastReset: now,
	})
	b := val.(*bucket)

	b.mu.Lock()
	defer b.mu.Unlock()

	// Reset tokens if window has passed
	if now.Sub(b.lastReset) >= rl.window {
		b.tokens = maxTokens
		b.lastReset = now
	}

	if b.tokens <= 0 {
		return false
	}

	b.tokens--
	return true
}

// Middleware returns an HTTP middleware that enforces rate limits.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only rate-limit write operations
		if r.Method != http.MethodPost && r.Method != http.MethodPut {
			next.ServeHTTP(w, r)
			return
		}

		// 1. Rate limit by IP
		ip := extractIP(r)
		if !rl.allow(&rl.ipBuckets, ip, rl.ipRate) {
			http.Error(w, "Rate limit exceeded (IP). Try again later.", http.StatusTooManyRequests)
			return
		}

		// 2. Rate limit by PubKey via JSON Body Inspection
		pubKey := r.Header.Get("X-MeuEu-Author")

		// If publishing, read the body to extract author_pubkey
		if r.Body != nil && (r.URL.Path == "/v1/publish" || pubKey == "") {
			bodyBytes, err := io.ReadAll(r.Body)
			if err == nil && len(bodyBytes) > 0 {
				// Important: Restore the body buffer for downstream handlers
				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

				var partial struct {
					AuthorPubkey string `json:"author_pubkey"`
				}
				if json.Unmarshal(bodyBytes, &partial) == nil && partial.AuthorPubkey != "" {
					pubKey = partial.AuthorPubkey
				}
			}
		}

		if pubKey != "" {
			if !rl.allow(&rl.pubKeyBuckets, pubKey, rl.pubKeyRate) {
				http.Error(w, "Rate limit exceeded for pubkey", http.StatusTooManyRequests)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func extractIP(r *http.Request) string {
	// Check X-Forwarded-For first (behind reverse proxy)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// Take the first IP in the chain
		if idx := len(forwarded); idx > 0 {
			parts := splitFirst(forwarded, ',')
			return parts
		}
	}

	// Check X-Real-IP
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fallback to RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func splitFirst(s string, sep byte) string {
	for i := 0; i < len(s); i++ {
		if s[i] == sep {
			return s[:i]
		}
	}
	return s
}
