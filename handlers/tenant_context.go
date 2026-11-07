package handlers

import (
	"context"
	"html/template"
	"net/http"
	"regexp"
	"time"

	"platform/db"
)

var validTenantSlug = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,61}[a-z0-9]$`)

type contextKey string

const tenantCtxKey contextKey = "tenant"

// tenantFromContext returns the Tenant injected by tenantMiddleware.
func tenantFromContext(r *http.Request) (Tenant, bool) {
	t, ok := r.Context().Value(tenantCtxKey).(Tenant)
	return t, ok
}

// tenantMiddleware resolves the {slug} path value, looks the tenant up,
// and injects it into the request context. Returns 404 for unknown slugs.
// It also enforces that the caller is authenticated; sysadmins may enter
// any tenant, tenant-role users may only enter their own (by slug match).
func tenantMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")

		user, ok := getSession(r)
		if !ok {
			oc := originFromContext(r)
			if oc.Mode == OriginModeDefault && slug != "" {
				if _, found := store.getBySlug(slug); found {
					http.Redirect(w, r, "/t/"+slug+"/login", http.StatusSeeOther)
					return
				}
			}
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		if !validTenantSlug.MatchString(slug) {
			http.NotFound(w, r)
			return
		}
		t, found := store.getBySlug(slug)
		if !found {
			http.NotFound(w, r)
			return
		}

		// Sysadmin can enter any tenant.
		// Tenant-role users can only enter their own space.
		if user.Role != RoleSysAdmin {
			if user.Role != RoleTenant || user.TenantSlug != slug {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
		}

		// Resolve numeric DB ID for the tenant (used by DB-backed handlers).
		if Pools != nil && t.DBID == 0 {
			ctx2, cancel2 := context.WithTimeout(r.Context(), 3*time.Second)
			defer cancel2()
			if dbID, err := db.LookupTenantIDBySlug(ctx2, Pools.Admin, slug); err == nil {
				t.DBID = dbID
			}
		}

		ctx := context.WithValue(r.Context(), tenantCtxKey, t)
		next(w, r.WithContext(ctx))
	}
}

// TenantNavItem is a nav entry scoped to a tenant (paths include /t/{slug}).
type TenantNavItem struct {
	ID    string
	Label string
	Icon  string
	Path  string
}

// TenantNavGroup is a labelled group of nav items in the tenant sidebar.
type TenantNavGroup struct {
	Label string
	Items []TenantNavItem
}

// TenantShellData is the data passed to the tenant shell template.
type TenantShellData struct {
	User       User
	Tenant     Tenant
	Nav        []TenantNavItem
	NavGroups  []TenantNavGroup
	ActiveNav  string
	PageTitle  string
	IsSysAdmin bool
	Content    template.HTML
}

func tenantNav(slug string) []TenantNavItem {
	base := "/t/" + slug
	return []TenantNavItem{
		{ID: "dashboard", Label: "Dashboard", Icon: "dashboard", Path: base + "/"},
	}
}

func tenantNavGroups(slug string) []TenantNavGroup {
	base := "/t/" + slug
	return []TenantNavGroup{
		{
			Label: "Support",
			Items: []TenantNavItem{
				{ID: "clients", Label: "Clients", Icon: "clients", Path: base + "/clients/"},
				{ID: "tickets", Label: "Tickets", Icon: "tickets", Path: base + "/tickets/"},
				{ID: "abuse-reports", Label: "Abuse Reports", Icon: "abuse", Path: base + "/abuse-reports/"},
				{ID: "refund-complaints", Label: "Refund Complaints", Icon: "refunds", Path: base + "/refund-complaints/"},
				{ID: "one-time-access", Label: "One-Time Access", Icon: "one-time-access", Path: base + "/one-time-access/"},
			},
		},
		{
			Label: "Infrastructure",
			Items: []TenantNavItem{
				{ID: "resources", Label: "Resources", Icon: "resources", Path: base + "/resources/"},
				{ID: "monitoring", Label: "Monitoring", Icon: "monitoring", Path: base + "/monitoring/"},
				{ID: "ids", Label: "Intrusion Detection", Icon: "ids", Path: base + "/ids/"},
				{ID: "incidents", Label: "Incidents", Icon: "incidents", Path: base + "/incidents/"},
				{ID: "status-page", Label: "Public Status Page", Icon: "status-page", Path: base + "/status-page/"},
				{ID: "maintenance", Label: "Maintenance", Icon: "maintenance", Path: base + "/maintenance/"},
			},
		},
		{
			Label: "Finance",
			Items: []TenantNavItem{
				{ID: "transactions", Label: "Transactions", Icon: "transactions", Path: base + "/transactions/"},
				{ID: "pricing", Label: "Pricing", Icon: "pricing", Path: base + "/pricing/"},
				{ID: "tax", Label: "Tax", Icon: "tax", Path: base + "/tax/"},
				{ID: "sla", Label: "SLA", Icon: "sla", Path: base + "/sla/"},
			},
		},
		{
			Label: "Business",
			Items: []TenantNavItem{
				{ID: "employees", Label: "Employees", Icon: "employees", Path: base + "/employees/"},
				{ID: "brand", Label: "Brand", Icon: "brand", Path: base + "/brand/"},
				{ID: "localization", Label: "Localization", Icon: "localization", Path: base + "/localization/"},
				{ID: "registration", Label: "Registration", Icon: "registration", Path: base + "/registration/"},
				{ID: "sso", Label: "Sign-in & SSO", Icon: "sso", Path: base + "/sso/"},
			},
		},
	}
}
