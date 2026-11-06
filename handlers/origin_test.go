package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
)

func setupOriginTestMux(t *testing.T) http.Handler {
	t.Helper()

	origWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if strings.HasSuffix(origWD, "/handlers") || strings.HasSuffix(origWD, "\\handlers") {
		if err := os.Chdir(".."); err != nil {
			t.Fatalf("Chdir(..): %v", err)
		}
		t.Cleanup(func() { _ = os.Chdir(origWD) })
	}

	resetState := func() {
		origins.ResetForTest()
		ResetTestAuthState()
		ResetLoginRateLimiter()
		for _, ten := range store.list() {
			store.delete(ten.ID)
		}
		memBrandMu.Lock()
		memBrands = map[string]memTenantBrand{}
		memBrandMu.Unlock()
	}
	resetState()
	t.Cleanup(resetState)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", HealthzHandler)
	mux.HandleFunc("/login", LoginHandler)
	mux.HandleFunc("GET /t/{slug}/login", TenantLoginPage)
	mux.HandleFunc("POST /t/{slug}/login", TenantLoginHandler)
	mux.HandleFunc("/logout", LogoutHandler)
	mux.HandleFunc("/forgot-password", ForgotPasswordHandler)
	RegisterSysadminRoutes(mux)
	RegisterTenantDashboardRoutes(mux)
	RegisterTenantBrandRoutes(mux)
	RegisterTenantRegisterPageRoutes(mux)
	RegisterAPIRoutes(mux)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			ServeError(w, r, http.StatusNotFound, "")
			return
		}
		LoginPage(w, r)
	})

	return OriginMiddleware(mux)
}

func doFormPost(h http.Handler, host, path string, form url.Values, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "http://"+host+path, strings.NewReader(form.Encode()))
	req.Host = host
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func doGet(h http.Handler, host, path string, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "http://"+host+path, nil)
	req.Host = host
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func doJSON(h http.Handler, method, host, path string, body any, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	var rdr io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rdr = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, "http://"+host+path, rdr)
	req.Host = host
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func sessionCookieFrom(t *testing.T, rec *httptest.ResponseRecorder) *http.Cookie {
	t.Helper()
	for _, c := range rec.Result().Cookies() {
		if c.Name == "platform_session" && c.Value != "" {
			return c
		}
	}
	t.Fatalf("expected platform_session cookie in response (status=%d)", rec.Code)
	return nil
}

func TestDefaultModeWhenNoSystemOriginDefined(t *testing.T) {
	app := setupOriginTestMux(t)

	store.add(Tenant{ID: "t-acme", Name: "Acme Hosting", Slug: "acme", Domain: "cloud.acme.test"})
	RegisterTestUser("owner@acme.test", "Acme Owner", "AcmePass123", RoleTenant, "acme")

	// 1. Direct IP access with 0 system domains defined -> no redirect; renders standard Platform login.
	rec := doGet(app, "203.0.113.15:8021", "/")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK in default unrestricted mode, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Sign In — Platform") || !strings.Contains(body, "Welcome back") {
		t.Fatalf("expected standard Platform login page when no system domain is defined")
	}
	if strings.Contains(body, "Only system administrators can sign in") {
		t.Fatalf("did not expect system-only login banner when no system domain is defined")
	}

	// 2. Both sysadmin AND tenant users can sign in normally in default mode.
	sysLogin := doFormPost(app, "203.0.113.15:8021", "/login", url.Values{
		"email":    {"tayyebimohammadreza@gmail.com"},
		"password": {"sysadmin123"},
	})
	if sysLogin.Code != http.StatusSeeOther || sysLogin.Header().Get("Location") != "/s/" {
		t.Fatalf("expected sysadmin redirect to /s/, got status=%d loc=%q", sysLogin.Code, sysLogin.Header().Get("Location"))
	}

	tenLogin := doFormPost(app, "203.0.113.15:8021", "/login", url.Values{
		"email":    {"owner@acme.test"},
		"password": {"AcmePass123"},
	})
	if tenLogin.Code != http.StatusSeeOther || tenLogin.Header().Get("Location") != "/t/acme/" {
		t.Fatalf("expected tenant login redirect to /t/acme/ in default mode, got status=%d loc=%q", tenLogin.Code, tenLogin.Header().Get("Location"))
	}
}

func TestSystemOriginDomainsRedirectAndSysadminOnlyAuth(t *testing.T) {
	app := setupOriginTestMux(t)

	store.add(Tenant{ID: "t-acme", Name: "Acme Hosting", Slug: "acme", Domain: "cloud.acme.test"})
	RegisterTestUser("owner@acme.test", "Acme Owner", "AcmePass123", RoleTenant, "acme")

	// Log in as sysadmin initially while no domain is defined, then configure system origin domains via /s/origin/.
	sysLogin := doFormPost(app, "127.0.0.1:8021", "/login", url.Values{
		"email":    {"tayyebimohammadreza@gmail.com"},
		"password": {"sysadmin123"},
	})
	sysCookie := sessionCookieFrom(t, sysLogin)

	// Verify /s/origin/ renders empty default state with zero pre-seeded domains.
	originPage := doGet(app, "127.0.0.1:8021", "/s/origin/", sysCookie)
	if originPage.Code != http.StatusOK {
		t.Fatalf("GET /s/origin/ status=%d", originPage.Code)
	}
	if !strings.Contains(originPage.Body.String(), "Default Unrestricted Mode") {
		t.Fatalf("expected Default Unrestricted Mode banner when 0 domains exist")
	}

	// Add two system domains via REST API: sys1.platform.test (#1) and sys2.platform.test (#2).
	add1 := doJSON(app, http.MethodPost, "127.0.0.1:8021", "/api/sysadmin/origin/domains", map[string]any{
		"domain": "sys1.platform.test",
		"note":   "Primary EU Control Plane",
	}, sysCookie)
	if add1.Code != http.StatusCreated {
		t.Fatalf("expected 201 adding sys1, got %d: %s", add1.Code, add1.Body.String())
	}

	// Now sys1.platform.test is active, so add2 must be called via sys1.platform.test!
	add2 := doJSON(app, http.MethodPost, "sys1.platform.test", "/api/sysadmin/origin/domains", map[string]any{
		"domain": "sys2.platform.test",
		"note":   "Secondary US Control Plane",
	}, sysCookie)
	if add2.Code != http.StatusCreated {
		t.Fatalf("expected 201 adding sys2, got %d: %s", add2.Code, add2.Body.String())
	}

	// Conflict check: cannot add a domain already owned by a tenant!
	conflict := doJSON(app, http.MethodPost, "sys1.platform.test", "/api/sysadmin/origin/domains", map[string]any{
		"domain": "cloud.acme.test",
	}, sysCookie)
	if conflict.Code != http.StatusConflict {
		t.Fatalf("expected 409 Conflict when adding tenant domain as system origin, got %d", conflict.Code)
	}

	// 1. Direct IP access redirects to the FIRST domain in the list (sys1.platform.test), preserving path & query.
	direct := doGet(app, "203.0.113.99:8021", "/s/origin/?tab=domains")
	if direct.Code != http.StatusFound {
		t.Fatalf("expected 302 redirect on direct IP access, got %d", direct.Code)
	}
	if got := direct.Header().Get("Location"); got != "https://sys1.platform.test/s/origin/?tab=domains" {
		t.Fatalf("unexpected redirect Location %q", got)
	}

	// 2. /healthz on loopback/IP is exempt from redirect so container healthchecks succeed.
	hz := doGet(app, "127.0.0.1:8021", "/healthz")
	if hz.Code != http.StatusOK {
		t.Fatalf("expected /healthz 200 OK on 127.0.0.1, got %d", hz.Code)
	}

	// 3. Reorder sys2.platform.test to #1 ("Make #1") -> direct IP now redirects to sys2.platform.test!
	doms := origins.listDomains(context.Background())
	if len(doms) != 2 {
		t.Fatalf("expected 2 domains, got %d", len(doms))
	}
	moveRes := doFormPost(app, "sys1.platform.test", "/s/origin/domains/"+strconv.FormatInt(doms[1].ID, 10)+"/move/", url.Values{
		"direction": {"top"},
	}, sysCookie)
	if moveRes.Code != http.StatusSeeOther {
		t.Fatalf("expected 303 on move top, got %d", moveRes.Code)
	}
	directAfterMove := doGet(app, "203.0.113.99:8021", "/")
	if got := directAfterMove.Header().Get("Location"); got != "https://sys2.platform.test/" {
		t.Fatalf("expected redirect to new #1 domain https://sys2.platform.test/, got %q", got)
	}

	// 4. Both sys1.platform.test and sys2.platform.test render the Platform System login page.
	for _, host := range []string{"sys1.platform.test", "sys2.platform.test"} {
		lp := doGet(app, host, "/")
		if lp.Code != http.StatusOK {
			t.Fatalf("GET / on %s status=%d", host, lp.Code)
		}
		html := lp.Body.String()
		if !strings.Contains(html, "Platform System") || !strings.Contains(html, "System Administrator Access") {
			t.Fatalf("expected login page on %s to identify itself as Platform System", host)
		}
	}

	// 5. Tenant user CANNOT sign in on a System Origin domain (both HTML form and JSON API).
	tenOnSys := doFormPost(app, "sys1.platform.test", "/login", url.Values{
		"email":    {"owner@acme.test"},
		"password": {"AcmePass123"},
	})
	if tenOnSys.Code != http.StatusOK || !strings.Contains(tenOnSys.Body.String(), "Only platform system administrators can sign in") {
		t.Fatalf("expected tenant login on system domain to be rejected with sysadmin-only error, status=%d", tenOnSys.Code)
	}

	tenAPIOnSys := doJSON(app, http.MethodPost, "sys1.platform.test", "/api/login", map[string]any{
		"email":    "owner@acme.test",
		"password": "AcmePass123",
	})
	if tenAPIOnSys.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for non-sysadmin /api/login on system domain, got %d", tenAPIOnSys.Code)
	}

	// 6. Sysadmin CAN sign in on a System Origin domain.
	sysOnSys := doFormPost(app, "sys1.platform.test", "/login", url.Values{
		"email":    {"tayyebimohammadreza@gmail.com"},
		"password": {"sysadmin123"},
	})
	if sysOnSys.Code != http.StatusSeeOther || sysOnSys.Header().Get("Location") != "/s/" {
		t.Fatalf("expected sysadmin login on system domain to redirect to /s/, got status=%d loc=%q", sysOnSys.Code, sysOnSys.Header().Get("Location"))
	}
}

func TestTenantCustomDomainBrandingAndStrictTenantScopedAuth(t *testing.T) {
	app := setupOriginTestMux(t)

	// Configure a system origin domain too, to verify tenant domains still resolve directly and don't redirect to system!
	_, err := origins.addDomain(context.Background(), "control.platform.test", "System", true)
	if err != nil {
		t.Fatalf("addDomain: %v", err)
	}

	store.add(Tenant{ID: "t-acme", Name: "Acme Cloud", Slug: "acme", Domain: "portal.acme.test"})
	store.add(Tenant{ID: "t-beta", Name: "Beta Hosting", Slug: "beta", Domain: "panel.beta.test"})

	// Customize Acme's brand identity & colors.
	setMemTenantBrand("acme", memTenantBrand{
		CompanyName:  "Acme Cloud Enterprise",
		Tagline:      "High-frequency bare metal & NVMe compute.",
		Website:      "https://acme.test",
		SupportEmail: "help@acme.test",
		Colors:       json.RawMessage(`[{"label":"Primary","hex":"#0EA5E9"},{"label":"Background","hex":"#0B132B"}]`),
	})

	RegisterTestUser("alice@acme.test", "Alice Acme", "AcmeUserPass1", RoleTenant, "acme")
	RegisterTestUser("bob@beta.test", "Bob Beta", "BetaUserPass1", RoleTenant, "beta")

	// 1. Accessing portal.acme.test renders Acme's brand — NOT "Platform" or "Platform, Inc."!
	acmeLogin := doGet(app, "portal.acme.test", "/")
	if acmeLogin.Code != http.StatusOK {
		t.Fatalf("GET / on portal.acme.test status=%d", acmeLogin.Code)
	}
	acmeHTML := acmeLogin.Body.String()
	for _, want := range []string{
		"Sign In — Acme Cloud Enterprise",
		"Sign in to Acme Cloud Enterprise",
		"High-frequency bare metal &amp; NVMe compute.",
		"help@acme.test",
		"#0EA5E9",
		"#0B132B",
	} {
		if !strings.Contains(acmeHTML, want) {
			t.Fatalf("expected Acme login page to contain %q", want)
		}
	}
	if strings.Contains(acmeHTML, "Platform, Inc.") || strings.Contains(acmeHTML, "Sign In — Platform") || strings.Contains(acmeHTML, "Platform System") {
		t.Fatalf("tenant login page leaked Platform branding")
	}

	// 2. Forgot-password on portal.acme.test also uses Acme's brand.
	acmeFP := doGet(app, "portal.acme.test", "/forgot-password")
	if acmeFP.Code != http.StatusOK || !strings.Contains(acmeFP.Body.String(), "Reset Password — Acme Cloud Enterprise") {
		t.Fatalf("expected forgot-password on tenant domain to carry tenant brand")
	}

	// 3. Cross-tenant user (Bob from Beta) CANNOT log in on Acme's domain.
	bobOnAcme := doFormPost(app, "portal.acme.test", "/login", url.Values{
		"email":    {"bob@beta.test"},
		"password": {"BetaUserPass1"},
	})
	if bobOnAcme.Code != http.StatusOK || len(bobOnAcme.Result().Cookies()) > 0 {
		t.Fatalf("expected Beta user login on Acme domain to fail, status=%d", bobOnAcme.Code)
	}

	bobAPIOnAcme := doJSON(app, http.MethodPost, "portal.acme.test", "/api/login", map[string]any{
		"email":    "bob@beta.test",
		"password": "BetaUserPass1",
	})
	if bobAPIOnAcme.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for Beta user /api/login on Acme domain, got %d", bobAPIOnAcme.Code)
	}

	// 4. Platform Sysadmin (not a member of Acme) CANNOT log in on Acme's domain.
	sysOnAcme := doFormPost(app, "portal.acme.test", "/login", url.Values{
		"email":    {"tayyebimohammadreza@gmail.com"},
		"password": {"sysadmin123"},
	})
	if sysOnAcme.Code != http.StatusOK || len(sysOnAcme.Result().Cookies()) > 0 {
		t.Fatalf("expected non-member sysadmin login on Acme domain to fail, status=%d", sysOnAcme.Code)
	}

	// 5. Alice (Acme user) CAN log in on Acme's domain and is redirected to /t/acme/.
	aliceOnAcme := doFormPost(app, "portal.acme.test", "/login", url.Values{
		"email":    {"alice@acme.test"},
		"password": {"AcmeUserPass1"},
	})
	if aliceOnAcme.Code != http.StatusSeeOther || aliceOnAcme.Header().Get("Location") != "/t/acme/" {
		t.Fatalf("expected Alice login on Acme domain to redirect to /t/acme/, got status=%d loc=%q", aliceOnAcme.Code, aliceOnAcme.Header().Get("Location"))
	}
	aliceCookie := sessionCookieFrom(t, aliceOnAcme)

	// 6. Alice's Acme session cookie is rejected if replayed against Beta's domain OR the System Origin domain!
	replayOnBeta := doGet(app, "panel.beta.test", "/t/beta/", aliceCookie)
	if replayOnBeta.Code != http.StatusSeeOther || replayOnBeta.Header().Get("Location") != "/" {
		t.Fatalf("expected Acme cookie on Beta domain to be unauthenticated redirect to /, got status=%d loc=%q", replayOnBeta.Code, replayOnBeta.Header().Get("Location"))
	}
	replayOnSys := doGet(app, "control.platform.test", "/s/origin/", aliceCookie)
	if replayOnSys.Code != http.StatusSeeOther || replayOnSys.Header().Get("Location") != "/" {
		t.Fatalf("expected Acme cookie on system domain to be rejected, got status=%d loc=%q", replayOnSys.Code, replayOnSys.Header().Get("Location"))
	}

	// 7. Sysadmin paths (/s/*, /api/sysadmin/*) and other tenants' paths (/t/beta/*) are 404 on Acme's domain.
	for _, blockedPath := range []string{"/s/origin/", "/api/sysadmin/origin", "/t/beta/", "/api/tenant/beta/dashboard"} {
		res := doGet(app, "portal.acme.test", blockedPath, aliceCookie)
		if res.Code != http.StatusNotFound {
			t.Fatalf("expected 404 for %s on portal.acme.test, got %d", blockedPath, res.Code)
		}
	}
}
