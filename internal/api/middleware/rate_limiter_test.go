package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestRateLimiter_Allow(t *testing.T) {
	rl := NewRateLimiter(5, 5, 1*time.Second)
	key := "192.168.1.1"

	// Should allow 5 requests
	for i := 0; i < 5; i++ {
		if !rl.allow(&rl.ipBuckets, key, rl.ipRate) {
			t.Errorf("Request %d should have been allowed", i+1)
		}
	}

	// 6th should be blocked
	if rl.allow(&rl.ipBuckets, key, rl.ipRate) {
		t.Errorf("6th request should have been rate limited")
	}

	// Wait window + buffer
	time.Sleep(1200 * time.Millisecond)

	// Should allow again
	if !rl.allow(&rl.ipBuckets, key, rl.ipRate) {
		t.Errorf("Request after window reset should have been allowed")
	}
}

func TestRateLimiter_ConcurrentAccess(t *testing.T) {
	rl := NewRateLimiter(50, 50, 1*time.Second)
	key := "test_author_pk"

	var wg sync.WaitGroup
	allowedCount := 0
	var mu sync.Mutex

	// Fire 100 concurrent requests, but bucket holds only 50
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if rl.allow(&rl.pubKeyBuckets, key, rl.pubKeyRate) {
				mu.Lock()
				allowedCount++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	if allowedCount != 50 {
		t.Errorf("Expected exactly 50 allowed concurrent requests, got %d", allowedCount)
	}
}

func TestRateLimiter_ExtractIP(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	
	// Test standard
	req.RemoteAddr = "203.0.113.1:8080"
	if ip := extractIP(req); ip != "203.0.113.1" {
		t.Errorf("Expected 203.0.113.1, got %s", ip)
	}

	// Test X-Forwarded-For
	req.Header.Set("X-Forwarded-For", "198.51.100.1, 10.0.0.1")
	if ip := extractIP(req); ip != "198.51.100.1" {
		t.Errorf("Expected 198.51.100.1, got %s", ip)
	}

	// Test X-Real-IP
	req.Header.Del("X-Forwarded-For")
	req.Header.Set("X-Real-IP", "100.64.0.1")
	if ip := extractIP(req); ip != "100.64.0.1" {
		t.Errorf("Expected 100.64.0.1, got %s", ip)
	}
}

func TestRateLimiter_Middleware(t *testing.T) {
	// 2 requests per IP, 1 per PubKey
	rl := NewRateLimiter(2, 1, 1*time.Minute)
	
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// --- Test PubKey limit (Author) ---
	reqAuthor, _ := http.NewRequest("POST", "/", nil)
	reqAuthor.RemoteAddr = "1.1.1.1:123"
	reqAuthor.Header.Set("X-MeuEu-Author", "pubkey_AAA")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, reqAuthor) // 1st OK
	if rr.Code != http.StatusOK {
		t.Errorf("First POST should be 200 OK")
	}

	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, reqAuthor) // 2nd Rate Limited by PubKey
	if rr2.Code != http.StatusTooManyRequests {
		t.Errorf("Second POST with same PubKey should be Rate Limited (429), got %d", rr2.Code)
	}
}
