package handlers

import (
	"context"
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
	failures int
	lockedAt time.Time
	lastAt   time.Time // time of most recent failure
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

// ResetLoginRateLimiter clears rate-limit buckets (used by tests).
func ResetLoginRateLimiter() {
	loginRateLimiter.mu.Lock()
	loginRateLimiter.entries = make(map[string]*loginEntry)
	loginRateLimiter.mu.Unlock()
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
			if now.Sub(e.lastAt) > loginWindowPeriod {
				delete(loginRateLimiter.entries, ip)
			}
		} else {
			if now.Sub(e.lockedAt) > loginLockoutPeriod*2 {
				delete(loginRateLimiter.entries, ip)
			}
		}
	}
}

type LoginPageData struct {
	Mode              string // "default" | "system" | "tenant"
	OriginDomain      string
	FormAction        string
	ForgotPasswordURL string
	RegisterURL       string
	TenantBrand       TenantLoginBrand
	Error             string
	Email             string
}

func isTenantSelfRegistrationOpen(ctx context.Context, t Tenant) bool {
	if Pools != nil && t.DBID > 0 {
		qCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		if rs, ok, err := db.GetRegistrationSettings(qCtx, Pools.Admin, t.DBID); err == nil && ok {
			return rs.Enabled
		}
	}
	return true
}

func buildLoginPageData(r *http.Request, explicitTenant *Tenant, email, errMsg string) LoginPageData {
	oc := originFromContext(r)

	switch oc.Mode {
	case OriginModeTenant:
		if oc.Tenant != nil {
			regURL := ""
			if isTenantSelfRegistrationOpen(r.Context(), *oc.Tenant) {
				regURL = "/register"
			}
			return LoginPageData{
				Mode:              "tenant",
				FormAction:        "/login",
				ForgotPasswordURL: "/forgot-password",
				RegisterURL:       regURL,
				TenantBrand:       resolveTenantLoginBrand(r.Context(), *oc.Tenant),
				Email:             email,
				Error:             errMsg,
			}
		}
	case OriginModeSystem:
		return LoginPageData{
			Mode:              "system",
			OriginDomain:      oc.MatchedDomain,
			FormAction:        "/login",
			ForgotPasswordURL: "/forgot-password",
			Email:             email,
			Error:             errMsg,
		}
	}

	if explicitTenant != nil {
		regURL := ""
		if isTenantSelfRegistrationOpen(r.Context(), *explicitTenant) {
			regURL = "/t/" + explicitTenant.Slug + "/register"
		}
		return LoginPageData{
			Mode:              "tenant",
			FormAction:        "/t/" + explicitTenant.Slug + "/login",
			ForgotPasswordURL: "/forgot-password?tenant=" + explicitTenant.Slug,
			RegisterURL:       regURL,
			TenantBrand:       resolveTenantLoginBrand(r.Context(), *explicitTenant),
			Email:             email,
			Error:             errMsg,
		}
	}

	return LoginPageData{
		Mode:              "default",
		FormAction:        "/login",
		ForgotPasswordURL: "/forgot-password",
		Email:             email,
		Error:             errMsg,
	}
}

func LoginPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	if user, ok := getSession(r); ok {
		redirectByUser(w, r, user)
		return
	}

	render(w, buildLoginPageData(r, nil, "", ""), "base.html", "login.html")
}

// TenantLoginPage serves GET /t/{slug}/login — a tenant-branded login page
// scoped strictly to accounts of that tenant.
func TenantLoginPage(w http.ResponseWriter, r *http.Request) {
	oc := originFromContext(r)
	if oc.Mode == OriginModeSystem {
		// On a System Origin domain, only system administrator login is permitted.
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	slug := r.PathValue("slug")
	tenant, found := store.getBySlug(slug)
	if !found {
		ServeError(w, r, http.StatusNotFound, "")
		return
	}

	if user, ok := getSession(r); ok && user.TenantSlug == tenant.Slug {
		redirectByUser(w, r, user)
		return
	}

	render(w, buildLoginPageData(r, &tenant, "", ""), "base.html", "login.html")
}

// TenantLoginHandler handles POST /t/{slug}/login — authenticating strictly
// against users of {slug}.
func TenantLoginHandler(w http.ResponseWriter, r *http.Request) {
	oc := originFromContext(r)
	if oc.Mode == OriginModeSystem {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	slug := r.PathValue("slug")
	tenant, found := store.getBySlug(slug)
	if !found {
		ServeError(w, r, http.StatusNotFound, "")
		return
	}

	handleScopedLoginPost(w, r, &tenant)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	handleScopedLoginPost(w, r, nil)
}

func handleScopedLoginPost(w http.ResponseWriter, r *http.Request, explicitTenant *Tenant) {
	ip := clientIP(r)
	reqID := w.Header().Get("X-Request-ID")
	email := r.FormValue("email")
	password := r.FormValue("password")

	if !checkLoginRate(ip) {
		if Pools != nil {
			go db.InsertAuditEvent(context.Background(), Pools.Admin, 0, "unknown", "auth.lockout", "warn", map[string]any{"email": email, "ip": ip}, "too many failed attempts", reqID, ip)
		}
		render(w, buildLoginPageData(r, explicitTenant, email, "Too many failed attempts. Please wait 1 minute before trying again."), "base.html", "login.html")
		return
	}

	oc := originFromContext(r)

	// 1. System Origin mode: ONLY platform system administrators may sign in.
	if oc.Mode == OriginModeSystem {
		sessionUser, status := authenticateSysadminCredentials(r.Context(), email, password)
		if status == db.SysadminAuthOK {
			clearLoginFailures(ip)
			remember := r.FormValue("remember") == "on"
			createSession(w, r, sessionUser, remember)
			if Pools != nil {
				go db.InsertAuditEvent(context.Background(), Pools.Admin, sessionUser.ID, string(sessionUser.Role),
					"auth.login.success", "info", map[string]any{"email": email, "origin": oc.MatchedDomain, "mode": "system"}, "", reqID, ip)
			}
			redirectByUser(w, r, sessionUser)
			return
		}

		justLocked := recordLoginFailure(ip)
		if status == db.SysadminAuthForbiddenRole {
			if Pools != nil {
				go db.InsertAuditEvent(context.Background(), Pools.Admin, sessionUser.ID, "user",
					"auth.login.denied", "warn", map[string]any{"email": email, "origin": oc.MatchedDomain},
					"non-sysadmin login rejected on system origin", reqID, ip)
			}
			render(w, buildLoginPageData(r, nil, email, "Access denied. Only platform system administrators can sign in on this domain."), "base.html", "login.html")
			return
		}
		if Pools != nil {
			go db.InsertAuditEvent(context.Background(), Pools.Admin, 0, "unknown", "auth.login.failure", "warn", map[string]any{"email": email, "origin": oc.MatchedDomain}, "invalid credentials", reqID, ip)
			if justLocked {
				go db.InsertAuditEvent(context.Background(), Pools.Admin, 0, "unknown", "auth.lockout", "warn", map[string]any{"email": email}, "account locked after repeated failures", reqID, ip)
			}
		}
		render(w, buildLoginPageData(r, nil, email, "Invalid email or password. Please try again."), "base.html", "login.html")
		return
	}

	// 2. Tenant Origin mode (custom domain or /t/{slug}/login): strictly scoped to that tenant's users.
	targetTenant := explicitTenant
	if oc.Mode == OriginModeTenant && oc.Tenant != nil {
		targetTenant = oc.Tenant
	}
	if targetTenant != nil {
		sessionUser, ok := authenticateTenantCredentials(r.Context(), *targetTenant, email, password)
		if ok {
			clearLoginFailures(ip)
			remember := r.FormValue("remember") == "on"
			createSession(w, r, sessionUser, remember)
			if Pools != nil {
				go db.InsertAuditEvent(context.Background(), Pools.Admin, sessionUser.ID, string(sessionUser.Role),
					"auth.login.success", "info", map[string]any{"email": email, "tenant": targetTenant.Slug, "mode": "tenant"}, "", reqID, ip)
			}
			redirectByUser(w, r, sessionUser)
			return
		}

		justLocked := recordLoginFailure(ip)
		if Pools != nil {
			go db.InsertAuditEvent(context.Background(), Pools.Admin, 0, "unknown",
				"auth.login.failure", "warn", map[string]any{"email": email, "tenant": targetTenant.Slug},
				"invalid or out-of-tenant credentials", reqID, ip)
			if justLocked {
				go db.InsertAuditEvent(context.Background(), Pools.Admin, 0, "unknown", "auth.lockout", "warn", map[string]any{"email": email}, "account locked after repeated failures", reqID, ip)
			}
		}
		render(w, buildLoginPageData(r, explicitTenant, email, "Invalid email or password. Please try again."), "base.html", "login.html")
		return
	}

	// 3. Default unrestricted mode (no system origin domains defined).
	sessionUser, ok := authenticateCredentials(r.Context(), email, password)
	if ok {
		clearLoginFailures(ip)
		remember := r.FormValue("remember") == "on"
		createSession(w, r, sessionUser, remember)
		if Pools != nil {
			go db.InsertAuditEvent(context.Background(), Pools.Admin, sessionUser.ID, string(sessionUser.Role), "auth.login.success", "info", map[string]any{"email": email}, "", reqID, ip)
		}
		redirectByUser(w, r, sessionUser)
		return
	}

	justLocked := recordLoginFailure(ip)
	if Pools != nil {
		go db.InsertAuditEvent(context.Background(), Pools.Admin, 0, "unknown", "auth.login.failure", "warn", map[string]any{"email": email}, "invalid credentials", reqID, ip)
		if justLocked {
			go db.InsertAuditEvent(context.Background(), Pools.Admin, 0, "unknown", "auth.lockout", "warn", map[string]any{"email": email}, "account locked after repeated failures", reqID, ip)
		}
	}
	render(w, buildLoginPageData(r, nil, email, "Invalid email or password. Please try again."), "base.html", "login.html")
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

// redirectByUser sends the user to their role-specific dashboard.
func redirectByUser(w http.ResponseWriter, r *http.Request, user User) {
	switch user.Role {
	case RoleSysAdmin:
		http.Redirect(w, r, "/s/", http.StatusSeeOther)
	case RoleTenant:
		if user.TenantSlug != "" {
			http.Redirect(w, r, "/t/"+user.TenantSlug+"/", http.StatusSeeOther)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	default:
		if user.TenantSlug != "" {
			http.Redirect(w, r, "/t/"+user.TenantSlug+"/", http.StatusSeeOther)
			return
		}
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
