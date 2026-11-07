package handlers

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"platform/db"
)

// TenantSSOPageData carries tenant-owned third-party provider settings to the
// tenant settings page. Provider configuration here controls which options are
// shown on public tenant auth pages; first-party SSO configuration stays in
// the sysadmin area.
type TenantSSOPageData struct {
	TenantShellData
	Providers []db.TenantSSOProvider
	Saved     bool
}

var tenantSSOMemorySettings = struct {
	sync.RWMutex
	bySlug map[string]map[string]bool
}{bySlug: make(map[string]map[string]bool)}

func resetTenantSSOSettingsForTest() {
	tenantSSOMemorySettings.Lock()
	tenantSSOMemorySettings.bySlug = make(map[string]map[string]bool)
	tenantSSOMemorySettings.Unlock()
}

func tenantDatabaseID(ctx context.Context, tenant Tenant) int64 {
	if tenant.DBID > 0 {
		return tenant.DBID
	}
	if Pools == nil || tenant.Slug == "" {
		return 0
	}
	id, err := db.LookupTenantIDBySlug(ctx, Pools.Admin, tenant.Slug)
	if err != nil {
		log.Printf("tenant SSO: resolve tenant %q: %v", tenant.Slug, err)
		return 0
	}
	return id
}

// tenantSSOSettings returns every supported provider, with enabled flags loaded
// from the tenant's row set. On a DB read error it safely leaves providers off
// so a failed settings read never exposes a provider that was not confirmed.
func tenantSSOSettings(ctx context.Context, tenant Tenant) []db.TenantSSOProvider {
	providers := db.SupportedTenantSSOProviders()
	if Pools == nil {
		tenantSSOMemorySettings.RLock()
		saved := tenantSSOMemorySettings.bySlug[tenant.Slug]
		for i := range providers {
			providers[i].Enabled = saved[providers[i].Key]
		}
		tenantSSOMemorySettings.RUnlock()
		return providers
	}

	tenantID := tenantDatabaseID(ctx, tenant)
	if tenantID == 0 {
		return providers
	}
	qCtx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()
	loaded, err := db.ListTenantSSOProviders(qCtx, Pools.Admin, tenantID)
	if err != nil {
		log.Printf("tenant SSO: load providers for %q: %v", tenant.Slug, err)
		return providers
	}
	return loaded
}

func enabledTenantSSOProviders(ctx context.Context, tenant Tenant) []db.TenantSSOProvider {
	providers := tenantSSOSettings(ctx, tenant)
	enabled := make([]db.TenantSSOProvider, 0, len(providers))
	for _, provider := range providers {
		if provider.Enabled {
			enabled = append(enabled, provider)
		}
	}
	return enabled
}

func saveTenantSSOSettings(ctx context.Context, tenant Tenant, form map[string][]string) error {
	providers := db.SupportedTenantSSOProviders()
	for i := range providers {
		for _, value := range form["provider_"+providers[i].Key] {
			if value == "on" {
				providers[i].Enabled = true
				break
			}
		}
	}

	if Pools == nil {
		saved := make(map[string]bool, len(providers))
		for _, provider := range providers {
			saved[provider.Key] = provider.Enabled
		}
		tenantSSOMemorySettings.Lock()
		tenantSSOMemorySettings.bySlug[tenant.Slug] = saved
		tenantSSOMemorySettings.Unlock()
		return nil
	}

	tenantID := tenantDatabaseID(ctx, tenant)
	if tenantID == 0 {
		return fmt.Errorf("tenant database id could not be resolved")
	}
	qCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	return db.SaveTenantSSOProviders(qCtx, Pools.Admin, tenantID, providers)
}

func tenantSSOSettingsPage(w http.ResponseWriter, r *http.Request) {
	user, _ := getSession(r)
	tenant, _ := tenantFromContext(r)

	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid provider settings.", http.StatusBadRequest)
			return
		}
		if err := saveTenantSSOSettings(r.Context(), tenant, r.Form); err != nil {
			log.Printf("tenant SSO: save providers for %q: %v", tenant.Slug, err)
			http.Error(w, "Could not save provider settings.", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/t/"+tenant.Slug+"/sso/?saved=1", http.StatusSeeOther)
		return
	}

	data := TenantSSOPageData{
		TenantShellData: TenantShellData{
			User:       user,
			Tenant:     tenant,
			Nav:        tenantNav(tenant.Slug),
			NavGroups:  tenantNavGroups(tenant.Slug),
			ActiveNav:  "sso",
			PageTitle:  "Sign-in & SSO",
			IsSysAdmin: user.Role == RoleSysAdmin,
		},
		Providers: tenantSSOSettings(r.Context(), tenant),
		Saved:     r.URL.Query().Get("saved") == "1",
	}

	innerTmpl, err := template.ParseFiles(filepath.Join("templates", "tenant", "sso_content.html"))
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	var content bytes.Buffer
	if err := innerTmpl.ExecuteTemplate(&content, "tenant-sso-content", data); err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	data.Content = template.HTML(content.String())

	shellTmpl, err := template.ParseFiles(
		filepath.Join("templates", "tenant", "base.html"),
		filepath.Join("templates", "tenant", "shell.html"),
	)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := shellTmpl.ExecuteTemplate(w, "tenant-base", data.TenantShellData); err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
	}
}

func RegisterTenantSSORoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /t/{slug}/sso/", tenantMiddleware(tenantSSOSettingsPage))
	mux.HandleFunc("POST /t/{slug}/sso/", tenantMiddleware(tenantSSOSettingsPage))
}
