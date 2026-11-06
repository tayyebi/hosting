package db_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"platform/db"
)

func TestPostgresMigrationsAndOriginQueries(t *testing.T) {
	adminURL := os.Getenv("TEST_DATABASE_ADMIN_URL")
	appURL := os.Getenv("TEST_DATABASE_APP_URL")
	roURL := os.Getenv("TEST_DATABASE_READONLY_URL")
	if adminURL == "" {
		t.Skip("TEST_DATABASE_ADMIN_URL not set; skipping live Postgres integration test")
	}
	if appURL == "" {
		appURL = adminURL
	}
	if roURL == "" {
		roURL = adminURL
	}

	t.Setenv("DATABASE_URL_ADMIN", adminURL)
	t.Setenv("DATABASE_URL", appURL)
	t.Setenv("DATABASE_URL_READONLY", roURL)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pools, err := db.Open(ctx)
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	defer pools.Close()

	migFS := os.DirFS("../migrations")
	if err := db.Migrate(pools.Admin, migFS); err != nil {
		t.Fatalf("db.Migrate: %v", err)
	}
	if err := db.Seed(ctx, pools.Admin); err != nil {
		t.Fatalf("db.Seed: %v", err)
	}

	// 1. Ensure 0 domains are seeded by default after migration + seed.
	initialDomains, err := db.ListOriginDomains(ctx, pools.Admin)
	if err != nil {
		t.Fatalf("ListOriginDomains: %v", err)
	}
	if len(initialDomains) != 0 {
		t.Fatalf("expected 0 seeded system origin domains, got %d", len(initialDomains))
	}

	settings, err := db.GetOriginSettings(ctx, pools.Admin)
	if err != nil {
		t.Fatalf("GetOriginSettings: %v", err)
	}
	if settings.RedirectScheme != "https" || settings.RedirectStatusCode != 302 || !settings.TrustForwardedHeaders {
		t.Fatalf("unexpected default origin settings: %+v", settings)
	}

	// 2. Create two tenants with custom domains and distinct users.
	acme, err := db.CreateTenant(ctx, pools.Admin, "Acme Cloud", "acme", "cloud.acme.example")
	if err != nil {
		t.Fatalf("CreateTenant acme: %v", err)
	}
	beta, err := db.CreateTenant(ctx, pools.Admin, "Beta Cloud", "beta", "panel.beta.example")
	if err != nil {
		t.Fatalf("CreateTenant beta: %v", err)
	}

	if err := db.SetTenantOwner(ctx, pools.Admin, acme.ID, "Alice Owner", "alice@acme.example"); err != nil {
		t.Fatalf("SetTenantOwner acme: %v", err)
	}
	if _, err := db.SetUserPasswordByEmail(ctx, pools.Admin, "alice@acme.example", "Alice Owner", "user", "AlicePass123"); err != nil {
		t.Fatalf("SetUserPasswordByEmail alice: %v", err)
	}
	if err := db.SetTenantOwner(ctx, pools.Admin, beta.ID, "Bob Owner", "bob@beta.example"); err != nil {
		t.Fatalf("SetTenantOwner beta: %v", err)
	}
	if _, err := db.SetUserPasswordByEmail(ctx, pools.Admin, "bob@beta.example", "Bob Owner", "user", "BobPass123"); err != nil {
		t.Fatalf("SetUserPasswordByEmail bob: %v", err)
	}

	// 3. Verify GetTenantByDomain resolves portless domain even when request has :8021.
	matched, found, err := db.GetTenantByDomain(ctx, pools.Admin, "cloud.acme.example:8021", "cloud.acme.example")
	if err != nil || !found || matched.Slug != "acme" {
		t.Fatalf("GetTenantByDomain failed: found=%v slug=%q err=%v", found, matched.Slug, err)
	}

	// 4. Verify brand settings persistence via Admin pool (bypasses RLS for unauthenticated login page).
	if err := db.UpsertBrandSettings(ctx, pools.Admin, acme.ID, "Acme Enterprise", "Dedicated NVMe Cloud", "https://acme.example", "support@acme.example", ""); err != nil {
		t.Fatalf("UpsertBrandSettings: %v", err)
	}
	if err := db.UpsertBrandColors(ctx, pools.Admin, acme.ID, json.RawMessage(`[{"label":"Primary","hex":"#2563EB"}]`)); err != nil {
		t.Fatalf("UpsertBrandColors: %v", err)
	}
	bs, ok, err := db.GetBrandSettings(ctx, pools.Admin, acme.ID)
	if err != nil || !ok || bs.CompanyName != "Acme Enterprise" {
		t.Fatalf("GetBrandSettings: ok=%v bs=%+v err=%v", ok, bs, err)
	}

	// 5. Verify AddOriginDomain rejects collision with tenant domain, adds valid domains, and reorders cleanly.
	if _, err := db.AddOriginDomain(ctx, pools.Admin, "cloud.acme.example", "collide", true); err == nil {
		t.Fatal("expected error adding domain already assigned to tenant")
	}
	d1, err := db.AddOriginDomain(ctx, pools.Admin, "https://Sys1.Platform.Example/", "Primary", true)
	if err != nil {
		t.Fatalf("AddOriginDomain d1: %v", err)
	}
	d2, err := db.AddOriginDomain(ctx, pools.Admin, "sys2.platform.example", "Secondary", true)
	if err != nil {
		t.Fatalf("AddOriginDomain d2: %v", err)
	}
	if err := db.MoveOriginDomain(ctx, pools.Admin, d2.ID, "top"); err != nil {
		t.Fatalf("MoveOriginDomain top: %v", err)
	}
	enabled, err := db.ListEnabledOriginDomains(ctx, pools.Admin)
	if err != nil || len(enabled) != 2 || enabled[0].Domain != "sys2.platform.example" || enabled[1].Domain != "sys1.platform.example" {
		t.Fatalf("unexpected enabled order after MoveOriginDomain: %+v (err=%v)", enabled, err)
	}

	// 6. Verify AuthenticateSysadmin allows sysadmin and rejects tenant user with SysadminAuthForbiddenRole.
	bootEmail, _, bootPass := db.SysadminBootstrapCredentials()
	if _, status, err := db.AuthenticateSysadmin(ctx, pools.Admin, bootEmail, bootPass); err != nil || status != db.SysadminAuthOK {
		t.Fatalf("expected sysadmin auth OK, got status=%v err=%v", status, err)
	}
	if _, status, err := db.AuthenticateSysadmin(ctx, pools.Admin, "alice@acme.example", "AlicePass123"); err != nil || status != db.SysadminAuthForbiddenRole {
		t.Fatalf("expected SysadminAuthForbiddenRole for tenant user on system origin, got status=%v err=%v", status, err)
	}

	// 7. Verify AuthenticateTenantUser strictly isolates tenants and rejects non-member sysadmins.
	if u, ok, err := db.AuthenticateTenantUser(ctx, pools.Admin, acme.ID, acme.Slug, "alice@acme.example", "AlicePass123"); err != nil || !ok || u.TenantSlug != "acme" || u.TenantRole != "tenant" {
		t.Fatalf("expected Alice to authenticate for acme, got ok=%v u=%+v err=%v", ok, u, err)
	}
	if _, ok, err := db.AuthenticateTenantUser(ctx, pools.Admin, acme.ID, acme.Slug, "bob@beta.example", "BobPass123"); err != nil || ok {
		t.Fatalf("expected Bob (beta) to be rejected on acme, got ok=%v err=%v", ok, err)
	}
	if _, ok, err := db.AuthenticateTenantUser(ctx, pools.Admin, acme.ID, acme.Slug, bootEmail, bootPass); err != nil || ok {
		t.Fatalf("expected non-member sysadmin to be rejected on acme, got ok=%v err=%v", ok, err)
	}

	// 8. Verify session tenant scoping in LookupSession.
	aliceAuth, _, _ := db.AuthenticateTenantUser(ctx, pools.Admin, acme.ID, acme.Slug, "alice@acme.example", "AlicePass123")
	if err := db.CreateSession(ctx, pools.Admin, aliceAuth.ID, acme.ID, "tok-alice-acme", time.Hour, "127.0.0.1", "go-test"); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	rec, err := db.LookupSession(ctx, pools.Admin, "tok-alice-acme")
	if err != nil || rec.TenantSlug != "acme" || rec.TenantRole != "tenant" {
		t.Fatalf("LookupSession: rec=%+v err=%v", rec, err)
	}

	// Clean up test domains.
	_, _ = db.DeleteOriginDomain(ctx, pools.Admin, d1.ID)
	_, _ = db.DeleteOriginDomain(ctx, pools.Admin, d2.ID)
}
