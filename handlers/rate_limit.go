package handlers

import (
	"net/http"
	"sync"
	"time"
)

// ipRateLimiter is a generic sliding-window per-IP rate limiter.
type ipRateLimiter struct {
	mu      sync.Mutex
	entries map[string]*rlEntry
	max     int
	window  time.Duration
}

type rlEntry struct {
	count     int
	windowEnd time.Time
}

func newIPRateLimiter(max int, window time.Duration) *ipRateLimiter {
	rl := &ipRateLimiter{
		entries: make(map[string]*rlEntry),
		max:     max,
		window:  window,
	}
	go rl.pruneLoop()
	return rl
}

// allow returns true if the request is within the rate limit.
func (rl *ipRateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	e, ok := rl.entries[ip]
	if !ok || time.Now().After(e.windowEnd) {
		rl.entries[ip] = &rlEntry{count: 1, windowEnd: time.Now().Add(rl.window)}
		return true
	}
	if e.count >= rl.max {
		return false
	}
	e.count++
	return true
}

// prune removes expired entries to prevent unbounded memory growth.
func (rl *ipRateLimiter) prune() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	for ip, e := range rl.entries {
		if now.After(e.windowEnd) {
			delete(rl.entries, ip)
		}
	}
}

func (rl *ipRateLimiter) pruneLoop() {
	for range time.Tick(5 * time.Minute) {
		rl.prune()
	}
}

// apiRateLimiter caps /api/* at 100 requests per minute per IP.
var apiRateLimiter = newIPRateLimiter(100, time.Minute)

// userIDRateLimiter is a sliding-window rate limiter keyed by user ID (as string).
type userIDRateLimiter struct {
	mu      sync.Mutex
	entries map[string]*rlEntry
	max     int
	window  time.Duration
}

func newUserIDRateLimiter(max int, window time.Duration) *userIDRateLimiter {
	rl := &userIDRateLimiter{
		entries: make(map[string]*rlEntry),
		max:     max,
		window:  window,
	}
	go func() {
		for range time.Tick(5 * time.Minute) {
			rl.mu.Lock()
			now := time.Now()
			for k, e := range rl.entries {
				if now.After(e.windowEnd) {
					delete(rl.entries, k)
				}
			}
			rl.mu.Unlock()
		}
	}()
	return rl
}

func (rl *userIDRateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	e, ok := rl.entries[key]
	if !ok || time.Now().After(e.windowEnd) {
		rl.entries[key] = &rlEntry{count: 1, windowEnd: time.Now().Add(rl.window)}
		return true
	}
	if e.count >= rl.max {
		return false
	}
	e.count++
	return true
}

// passwordChangeRateLimiter allows 5 password changes per user ID per hour.
var passwordChangeRateLimiter = newUserIDRateLimiter(5, time.Hour)

// apiKeyCreationRateLimiter allows 20 API key creation attempts per user ID per hour,
// preventing rapid-fire revoke+create cycles even when under the 10-key cap.
var apiKeyCreationRateLimiter = newUserIDRateLimiter(20, time.Hour)

// registerRateLimiter caps tenant self-registration at 10 attempts per IP per hour
// to slow down bot-driven account creation.
var registerRateLimiter = newIPRateLimiter(10, time.Hour)

// sessionListRateLimiter caps GET /api/user/sessions at 30 requests per minute
// per user ID to prevent rapid session enumeration.
var sessionListRateLimiter = newUserIDRateLimiter(30, time.Minute)

// workspaceCreateRateLimiter allows at most 5 workspace creations per user per hour,
// preventing resource-exhaustion via repeated create-delete cycling.
var workspaceCreateRateLimiter = newUserIDRateLimiter(5, time.Hour)

// APIRateLimit is a middleware that enforces per-IP rate limits on /api/* routes.
func APIRateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Path) >= 4 && r.URL.Path[:4] == "/api" {
			ip := clientIP(r)
			if !apiRateLimiter.allow(ip) {
				w.Header().Set("Retry-After", "60")
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error":"Too Many Requests","message":"Rate limit exceeded. Retry after 60 seconds.","code":429}`))
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
