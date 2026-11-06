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

// fallbackSysadminMu guards fallbackSysadminPassword and in-memory test accounts
// used ONLY when Pools == nil (unit test mode without PostgreSQL).
var (
	fallbackSysadminMu       sync.Mutex
	fallbackSysadminPassword string

	memUsersMu   sync.RWMutex
	memTestUsers []memTestUser

	memSessionsMu sync.RWMutex
	memSessions   = map[string]memSessionEntry{}
)

type memTestUser struct {
	ID         int64
	Email      string
	Name       string
	Password   string
	Role       Role
	TenantSlug string
}

type memSessionEntry struct {
	User      User
	ExpiresAt time.Time
}

// RegisterTestUser registers an in-memory user account for unit tests running with Pools == nil.
func RegisterTestUser(email, name, password string, role Role, tenantSlug string) {
	memUsersMu.Lock()
	defer memUsersMu.Unlock()
	id := int64(len(memTestUsers) + 100)
	memTestUsers = append(memTestUsers, memTestUser{
		ID:         id,
		Email:      strings.TrimSpace(email),
		Name:       name,
		Password:   password,
		Role:       role,
		TenantSlug: tenantSlug,
	})
}

// ResetTestAuthState clears in-memory test users and sessions (used in unit tests).
func ResetTestAuthState() {
	memUsersMu.Lock()
	memTestUsers = nil
	memUsersMu.Unlock()

	memSessionsMu.Lock()
	memSessions = map[string]memSessionEntry{}
	memSessionsMu.Unlock()

	fallbackSysadminMu.Lock()
	fallbackSysadminPassword = ""
	fallbackSysadminMu.Unlock()
}

// authenticateCredentials verifies user credentials in default (unrestricted) mode.
// When Pools != nil, PostgreSQL `users` is authoritative.
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
		} else if u.TenantRole == "tenant" {
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
	memUsersMu.RLock()
	for _, mu := range memTestUsers {
		if strings.EqualFold(mu.Email, email) && subtle.ConstantTimeCompare([]byte(mu.Password), []byte(password)) == 1 {
			memUsersMu.RUnlock()
			return User{
				ID:         mu.ID,
				Email:      mu.Email,
				Name:       mu.Name,
				Role:       mu.Role,
				TenantSlug: mu.TenantSlug,
			}, true
		}
	}
	memUsersMu.RUnlock()

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

// authenticateSysadminCredentials enforces that ONLY platform system administrators
// (RoleSysAdmin) may authenticate on a configured System Origin domain.
func authenticateSysadminCredentials(ctx context.Context, email, password string) (User, db.SysadminAuthStatus) {
	email = strings.TrimSpace(email)
	if email == "" || password == "" {
		return User{}, db.SysadminAuthInvalid
	}

	if Pools != nil {
		authCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		u, status, err := db.AuthenticateSysadmin(authCtx, Pools.Admin, email, password)
		if err != nil {
			log.Printf("auth: AuthenticateSysadmin error: %v", err)
			return User{}, db.SysadminAuthInvalid
		}
		if status != db.SysadminAuthOK {
			return User{ID: u.ID, Email: u.Email, Name: u.Name}, status
		}
		return User{
			ID:        u.ID,
			Email:     u.Email,
			Name:      u.Name,
			Role:      RoleSysAdmin,
			AvatarURL: u.AvatarURL,
		}, db.SysadminAuthOK
	}

	// Unit-test mode (Pools == nil)
	u, ok := authenticateCredentials(ctx, email, password)
	if !ok {
		return User{}, db.SysadminAuthInvalid
	}
	if u.Role != RoleSysAdmin {
		return u, db.SysadminAuthForbiddenRole
	}
	u.TenantSlug = ""
	return u, db.SysadminAuthOK
}

// authenticateTenantCredentials strictly scopes login to accounts belonging to `tenant`.
func authenticateTenantCredentials(ctx context.Context, tenant Tenant, email, password string) (User, bool) {
	email = strings.TrimSpace(email)
	if email == "" || password == "" || tenant.Slug == "" {
		return User{}, false
	}

	if Pools != nil {
		if tenant.DBID <= 0 {
			return User{}, false
		}
		authCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		u, ok, err := db.AuthenticateTenantUser(authCtx, Pools.Admin, tenant.DBID, tenant.Slug, email, password)
		if err != nil {
			log.Printf("auth: AuthenticateTenantUser(%s) error: %v", tenant.Slug, err)
			return User{}, false
		}
		if !ok {
			return User{}, false
		}
		role := RoleUser
		if u.TenantRole == "tenant" {
			role = RoleTenant
		}
		return User{
			ID:         u.ID,
			Email:      u.Email,
			Name:       u.Name,
			Role:       role,
			AvatarURL:  u.AvatarURL,
			TenantSlug: tenant.Slug,
		}, true
	}

	// Unit-test mode (Pools == nil): match only users belonging to tenant.Slug.
	memUsersMu.RLock()
	defer memUsersMu.RUnlock()
	for _, mu := range memTestUsers {
		if mu.TenantSlug == tenant.Slug &&
			strings.EqualFold(mu.Email, email) &&
			subtle.ConstantTimeCompare([]byte(mu.Password), []byte(password)) == 1 {
			role := mu.Role
			if role == RoleSysAdmin {
				role = RoleTenant
			}
			return User{
				ID:         mu.ID,
				Email:      mu.Email,
				Name:       mu.Name,
				Role:       role,
				TenantSlug: tenant.Slug,
			}, true
		}
	}
	return User{}, false
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
		memSessionsMu.Lock()
		memSessions[id] = memSessionEntry{
			User:      user,
			ExpiresAt: time.Now().Add(ttl),
		}
		memSessionsMu.Unlock()
	} else {
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
	if err != nil || c.Value == "" {
		return User{}, false
	}

	var u User
	if Pools == nil {
		memSessionsMu.RLock()
		entry, ok := memSessions[c.Value]
		memSessionsMu.RUnlock()
		if !ok || time.Now().After(entry.ExpiresAt) {
			return User{}, false
		}
		u = entry.User
	} else {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		rec, err := db.LookupSession(ctx, Pools.Admin, c.Value)
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				log.Printf("session: lookup: %v", err)
			}
			return User{}, false
		}
		if rec.TenantID > 0 && rec.TenantSlug == "" {
			// Tenant was deleted after session creation.
			return User{}, false
		}

		u = User{
			ID:         rec.UserID,
			Email:      rec.Email,
			Name:       rec.Name,
			TenantSlug: rec.TenantSlug,
		}
		if rec.GlobalRole == "sysadmin" && rec.TenantID == 0 {
			u.Role = RoleSysAdmin
		} else if rec.TenantRole == "tenant" {
			u.Role = RoleTenant
		} else {
			u.Role = RoleUser
		}
	}

	// Enforce origin-scoped session validity:
	oc := originFromContext(r)
	switch oc.Mode {
	case OriginModeSystem:
		if u.Role != RoleSysAdmin {
			return User{}, false
		}
	case OriginModeTenant:
		if oc.Tenant == nil || u.TenantSlug != oc.Tenant.Slug {
			return User{}, false
		}
		if u.Role == RoleSysAdmin {
			u.Role = RoleTenant
		}
	}

	return u, true
}

func destroySession(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("platform_session")
	if err == nil && c.Value != "" {
		if Pools != nil {
			if revErr := db.RevokeSession(r.Context(), Pools.Admin, c.Value); revErr != nil {
				log.Printf("session: revoke: %v", revErr)
			}
		} else {
			memSessionsMu.Lock()
			delete(memSessions, c.Value)
			memSessionsMu.Unlock()
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
