package handlers

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"

	"platform/db"
)

type Role string

const (
	RoleSysAdmin Role = "sysadmin"
	RoleTenant   Role = "tenant"
	RoleUser     Role = "user"
)

type User struct {
	ID         int64 // database user id (0 in legacy fallback)
	Email      string
	Name       string
	Role       Role
	AvatarURL  string
	TenantSlug string // non-empty when the user belongs to a specific tenant
}

type SessionUser = User

// fallbackSysadminMu guards fallbackSysadminPassword when Pools == nil (e.g. unit tests without DB).
var fallbackSysadminMu sync.Mutex
var fallbackSysadminPassword string

// authenticateCredentials verifies user credentials against PostgreSQL `users`
// when Pools is configured. Only when Pools == nil (isolated unit test mode)
// does it fall back to the bootstrap System Super Administrator credentials.
func authenticateCredentials(ctx context.Context, email, password string) (User, bool) {
	email = strings.TrimSpace(email)
	if email == "" || password == "" {
		return User{}, false
	}

	if Pools != nil {
		authCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		u, ok, err := db.AuthenticateUser(authCtx, Pools.Admin, email, password)
		if err != nil {
			log.Printf("auth: AuthenticateUser error: %v", err)
			return User{}, false
		}
		if !ok {
			return User{}, false
		}
		role := RoleUser
		if u.GlobalRole == "sysadmin" {
			role = RoleSysAdmin
		} else if u.TenantRole == "tenant" || u.TenantSlug != "" {
			role = RoleTenant
		}
		return User{
			ID:         u.ID,
			Email:      u.Email,
			Name:       u.Name,
			Role:       role,
			AvatarURL:  u.AvatarURL,
			TenantSlug: u.TenantSlug,
		}, true
	}

	// Fallback ONLY when running without a database pool.
	bootEmail, bootName, bootPass := db.SysadminBootstrapCredentials()
	fallbackSysadminMu.Lock()
	if fallbackSysadminPassword == "" {
		fallbackSysadminPassword = bootPass
	}
	expected := fallbackSysadminPassword
	fallbackSysadminMu.Unlock()

	if !strings.EqualFold(email, bootEmail) {
		return User{}, false
	}
	if subtle.ConstantTimeCompare([]byte(expected), []byte(password)) != 1 {
		return User{}, false
	}
	return User{
		Email: bootEmail,
		Name:  bootName,
		Role:  RoleSysAdmin,
	}, true
}

func changeAccountPassword(ctx context.Context, user User, currentPw, newPw string) (bool, error) {
	if Pools != nil && user.ID > 0 {
		return db.ChangeUserPassword(ctx, Pools.Admin, user.ID, currentPw, newPw)
	}
	fallbackSysadminMu.Lock()
	defer fallbackSysadminMu.Unlock()
	bootEmail, _, bootPass := db.SysadminBootstrapCredentials()
	if fallbackSysadminPassword == "" {
		fallbackSysadminPassword = bootPass
	}
	if !strings.EqualFold(user.Email, bootEmail) ||
		subtle.ConstantTimeCompare([]byte(fallbackSysadminPassword), []byte(currentPw)) != 1 {
		return false, nil
	}
	fallbackSysadminPassword = newPw
	return true, nil
}

func createSession(w http.ResponseWriter, r *http.Request, user User, remember bool) {
	token := make([]byte, 24)
	_, _ = rand.Read(token)
	id := hex.EncodeToString(token)

	ttl := 8 * time.Hour
	if remember {
		ttl = 30 * 24 * time.Hour
	}

	if Pools == nil {
		log.Println("session: Pools is nil — Postgres not wired; refusing to issue session")
		http.Error(w, "database unavailable", http.StatusServiceUnavailable)
		return
	}

	userID := user.ID
	if userID == 0 {
		var err error
		userID, err = db.UpsertUser(r.Context(), Pools.Admin, user.Email, user.Name, normaliseGlobalRole(user.Role))
		if err != nil {
			log.Printf("session: upsert user for %s: %v", user.Email, err)
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
	}
	user.ID = userID

	var tenantID int64
	if user.TenantSlug != "" {
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
		for i := 0; i < len(fwd); i++ {
			if fwd[i] == ',' {
				return fwd[:i]
			}
		}
		return fwd
	}
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}
