package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"platform/db"

	"github.com/jackc/pgx/v5"
)

type Role string

const (
	RoleSysAdmin Role = "sysadmin"
	RoleTenant   Role = "tenant"
	RoleUser     Role = "user"
)

type User struct {
	ID         int64  // users.id in Postgres; 0 before a DB row exists
	Email      string
	Name       string
	Role       Role
	AvatarURL  string
	TenantSlug string // set for RoleTenant users; empty for sysadmin
}

// mockUsersMu guards mockUsers for concurrent reads/writes (e.g. password changes).
var mockUsersMu sync.Mutex

// Mock user credential table — Phase 1 still owns password checks here;
// DB persistence kicks in once a login succeeds. Phase 2 moves this to
// a users + auth_credentials table.
var mockUsers = map[string]struct {
	Password string
	User     User
}{
	"sysadmin@platform.io": {
		Password: "sysadmin123",
		User: User{
			Email: "sysadmin@platform.io",
			Name:  "Sys Admin",
			Role:  RoleSysAdmin,
		},
	},
}

// createSession issues a cookie and persists the session row in Postgres.
// The plaintext token lives only in the cookie; the DB stores its SHA-256
// hash so a DB leak cannot be replayed.
func createSession(w http.ResponseWriter, r *http.Request, user User, remember bool) {
	token := make([]byte, 24)
	_, _ = rand.Read(token)
	id := hex.EncodeToString(token)

	ttl := 8 * time.Hour
	if remember {
		ttl = 30 * 24 * time.Hour
	}

	if Pools == nil {
		// Phase 1 makes Postgres a hard dependency; this path is a
		// defensive no-op that surfaces misconfiguration loudly.
		log.Println("session: Pools is nil — Postgres not wired; refusing to issue session")
		http.Error(w, "database unavailable", http.StatusServiceUnavailable)
		return
	}

	// Ensure a users row exists so sessions.user_id FK holds. The seed
	// covers sysadmin@platform.io; any additional mockUsers are upserted
	// lazily here on first login.
	userID, err := db.UpsertUser(r.Context(), Pools.Admin, user.Email, user.Name, normaliseGlobalRole(user.Role))
	if err != nil {
		log.Printf("session: upsert user for %s: %v", user.Email, err)
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	user.ID = userID

	var tenantID int64
	if user.TenantSlug != "" {
		// Best-effort resolve: if the tenant is seeded, link the session
		// to it; otherwise leave 0 and the session is effectively sysadmin-scope.
		_ = Pools.Admin.QueryRow(r.Context(),
			`SELECT id FROM tenants WHERE slug = $1 AND deleted_at IS NULL`,
			user.TenantSlug).Scan(&tenantID)
	}

	ip := clientIP(r)
	ua := r.UserAgent()
	if err := db.CreateSession(r.Context(), Pools.Admin, userID, tenantID, id, ttl, ip, ua); err != nil {
		log.Printf("session: insert session for %s: %v", user.Email, err)
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "platform_session",
		Value:    id,
		Path:     "/",
		HttpOnly: true,
		Secure:   os.Getenv("PLATFORM_ENV") == "production",
		MaxAge:   int(ttl.Seconds()),
		SameSite: http.SameSiteStrictMode,
	})
}

// getSession resolves the cookie to a User by querying the sessions table.
// Returns (User{}, false) for anonymous, expired, or revoked cookies.
func getSession(r *http.Request) (User, bool) {
	c, err := r.Cookie("platform_session")
	if err != nil {
		return User{}, false
	}
	if Pools == nil {
		return User{}, false
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	rec, err := db.LookupSession(ctx, Pools.Admin, c.Value)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			log.Printf("session: lookup: %v", err)
		}
		return User{}, false
	}

	u := User{
		ID:         rec.UserID,
		Email:      rec.Email,
		Name:       rec.Name,
		TenantSlug: rec.TenantSlug,
	}
	switch rec.GlobalRole {
	case "sysadmin":
		u.Role = RoleSysAdmin
	default:
		// If the session has a tenant, treat the user as a tenant admin
		// for Phase 1 purposes (the UI doesn't yet distinguish between
		// tenant-owner and tenant-user in routing).
		if rec.TenantSlug != "" {
			u.Role = RoleTenant
		} else {
			u.Role = RoleUser
		}
	}
	return u, true
}

func destroySession(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("platform_session")
	if err == nil && Pools != nil {
		if revErr := db.RevokeSession(r.Context(), Pools.Admin, c.Value); revErr != nil {
			log.Printf("session: revoke: %v", revErr)
		}
	}
	http.SetCookie(w, &http.Cookie{
		Name:   "platform_session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
}

// requireRole is a middleware that checks the user's role.
func requireRole(role Role, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := getSession(r)
		if !ok || user.Role != role {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

func normaliseGlobalRole(r Role) string {
	if r == RoleSysAdmin {
		return "sysadmin"
	}
	return "user"
}

func clientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		// Take the first entry only.
		for i := 0; i < len(fwd); i++ {
			if fwd[i] == ',' {
				return fwd[:i]
			}
		}
		return fwd
	}
	// RemoteAddr is "ip:port" or "[ipv6]:port" — SplitHostPort strips the
	// brackets too so Postgres accepts it as inet.
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}
