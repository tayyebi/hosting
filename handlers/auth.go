package handlers

import (
	"context"
	"crypto/subtle"
	"html/template"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"platform/db"
)

// loginRateLimiter tracks failed login attempts per IP.
var loginRateLimiter = struct {
	mu      sync.Mutex
	entries map[string]*loginEntry
}{entries: make(map[string]*loginEntry)}

type loginEntry struct {
	failures  int
	lockedAt  time.Time
	lastAt    time.Time // time of most recent failure
}

const (
	loginMaxFailures   = 5
	loginLockoutPeriod = time.Minute
	loginWindowPeriod  = 5 * time.Minute
)

func init() {
	go func() {
		for range time.Tick(5 * time.Minute) {
			pruneLoginEntries()
		}
	}()
}

func checkLoginRate(ip string) bool {
	loginRateLimiter.mu.Lock()
	defer loginRateLimiter.mu.Unlock()
	e, ok := loginRateLimiter.entries[ip]
	if !ok {
		return true
	}
	if !e.lockedAt.IsZero() {
		if time.Since(e.lockedAt) >= loginLockoutPeriod {
			delete(loginRateLimiter.entries, ip)
			return true
		}
		return false
	}
	return true
}

// recordLoginFailure increments the failure counter and returns true if this
// failure triggered a new lockout (i.e., the threshold was just crossed).
func recordLoginFailure(ip string) bool {
	loginRateLimiter.mu.Lock()
	defer loginRateLimiter.mu.Unlock()
	e, ok := loginRateLimiter.entries[ip]
	if !ok {
		e = &loginEntry{}
		loginRateLimiter.entries[ip] = e
	}
	e.failures++
	e.lastAt = time.Now()
	if e.failures == loginMaxFailures {
		e.lockedAt = time.Now()
		return true
	}
	return false
}

func clearLoginFailures(ip string) {
	loginRateLimiter.mu.Lock()
	defer loginRateLimiter.mu.Unlock()
	delete(loginRateLimiter.entries, ip)
}

// pruneLoginEntries removes stale entries to prevent unbounded memory growth.
func pruneLoginEntries() {
	loginRateLimiter.mu.Lock()
	defer loginRateLimiter.mu.Unlock()
	now := time.Now()
	for ip, e := range loginRateLimiter.entries {
		if e.lockedAt.IsZero() {
			// Not locked: prune if no failure in the last window period.
			if now.Sub(e.lastAt) > loginWindowPeriod {
				delete(loginRateLimiter.entries, ip)
			}
		} else {
			// Locked: prune if the lockout period has long expired.
			if now.Sub(e.lockedAt) > loginLockoutPeriod*2 {
				delete(loginRateLimiter.entries, ip)
			}
		}
	}
}

type LoginPageData struct {
	Error string
	Email string
}

func LoginPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Already logged in? Redirect to the right dashboard.
	if user, ok := getSession(r); ok {
		redirectByRole(w, r, user.Role)
		return
	}

	render(w, LoginPageData{}, "base.html", "login.html")
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	ip := clientIP(r)

	reqID := w.Header().Get("X-Request-ID")

	if !checkLoginRate(ip) {
		if Pools != nil {
			go db.InsertAuditEvent(context.Background(), Pools.Admin, 0, "unknown", "auth.lockout", "warn", map[string]any{"email": r.FormValue("email"), "ip": ip}, "too many failed attempts", reqID, ip)
		}
		render(w, LoginPageData{
			Error: "Too many failed attempts. Please wait 1 minute before trying again.",
			Email: r.FormValue("email"),
		}, "base.html", "login.html")
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	record, ok := mockUsers[email]
	if ok && subtle.ConstantTimeCompare([]byte(record.Password), []byte(password)) == 1 {
		clearLoginFailures(ip)
		remember := r.FormValue("remember") == "on"
		createSession(w, r, record.User, remember)
		if Pools != nil {
			go db.InsertAuditEvent(context.Background(), Pools.Admin, record.User.ID, string(record.User.Role), "auth.login.success", "info", map[string]any{"email": email}, "", reqID, ip)
		}
		redirectByRole(w, r, record.User.Role)
		return
	}

	justLocked := recordLoginFailure(ip)
	if Pools != nil {
		go db.InsertAuditEvent(context.Background(), Pools.Admin, 0, "unknown", "auth.login.failure", "warn", map[string]any{"email": email}, "invalid credentials", reqID, ip)
		if justLocked {
			go db.InsertAuditEvent(context.Background(), Pools.Admin, 0, "unknown", "auth.lockout", "warn", map[string]any{"email": email}, "account locked after repeated failures", reqID, ip)
		}
	}
	render(w, LoginPageData{
		Error: "Invalid email or password. Please try again.",
		Email: email,
	}, "base.html", "login.html")
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	ip := clientIP(r)
	reqID := w.Header().Get("X-Request-ID")
	if user, ok := getSession(r); ok && Pools != nil {
		go db.InsertAuditEvent(context.Background(), Pools.Admin, user.ID, string(user.Role), "auth.logout", "info", map[string]any{"email": user.Email}, "", reqID, ip)
	}
	destroySession(w, r)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// redirectByRole sends the user to their role-specific dashboard.
func redirectByRole(w http.ResponseWriter, r *http.Request, role Role) {
	switch role {
	case RoleSysAdmin:
		http.Redirect(w, r, "/s/", http.StatusSeeOther)
	default:
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

// render parses and executes template files relative to the templates/ dir.
func render(w http.ResponseWriter, data any, files ...string) {
	paths := make([]string, len(files))
	for i, f := range files {
		paths[i] = filepath.Join("templates", f)
	}
	tmpl, err := template.ParseFiles(paths...)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
	}
}
