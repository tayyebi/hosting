package handlers

import "net/http"

// ── GET /api/tenant/{slug}/dashboard ──────────────────────

type apiTenantNavItem struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Icon  string `json:"icon"`
	Path  string `json:"path"`
}

type apiTenantNavGroup struct {
	Label string             `json:"label"`
	Items []apiTenantNavItem `json:"items"`
}

type apiTenantKPI struct {
	Label string  `json:"label"`
	Value string  `json:"value"`
	Delta string  `json:"delta"`
	Trend string  `json:"trend"` // "up" | "down" | "flat"
}

type apiTenantDashboardResponse struct {
	Tenant    apiTenant           `json:"tenant"`
	Nav       []apiTenantNavItem  `json:"nav"`
	NavGroups []apiTenantNavGroup `json:"nav_groups"`
	KPIs      []apiTenantKPI      `json:"kpis"`
	Health    string              `json:"health"`
}

func apiTenantDashboardHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok {
		apiError(w, http.StatusUnauthorized, "Authentication required.")
		return
	}

	slug := r.PathValue("slug")
	t, found := store.getBySlug(slug)
	if !found {
		apiError(w, http.StatusNotFound, "Tenant not found.")
		return
	}

	if user.Role != RoleSysAdmin && (user.Role != RoleTenant || user.TenantSlug != slug) {
		apiError(w, http.StatusForbidden, "Access denied.")
		return
	}

	nav := tenantNav(slug)
	apiNav := make([]apiTenantNavItem, len(nav))
	for i, n := range nav {
		apiNav[i] = apiTenantNavItem{ID: n.ID, Label: n.Label, Icon: n.Icon, Path: n.Path}
	}

	groups := tenantNavGroups(slug)
	apiGroups := make([]apiTenantNavGroup, len(groups))
	for i, g := range groups {
		items := make([]apiTenantNavItem, len(g.Items))
		for j, it := range g.Items {
			items[j] = apiTenantNavItem{ID: it.ID, Label: it.Label, Icon: it.Icon, Path: it.Path}
		}
		apiGroups[i] = apiTenantNavGroup{Label: g.Label, Items: items}
	}

	clients := clientStore.list(slug)

	apiOK(w, apiTenantDashboardResponse{
		Tenant: apiTenant{
			ID:     t.ID,
			Name:   t.Name,
			Slug:   t.Slug,
			Domain: t.Domain,
		},
		Nav:       apiNav,
		NavGroups: apiGroups,
		KPIs: []apiTenantKPI{
			{Label: "Active Instances", Value: "127", Delta: "+12%", Trend: "up"},
			{Label: "Uptime", Value: "99.97%", Delta: "+0.02%", Trend: "up"},
			{Label: "Avg Latency", Value: "23 ms", Delta: "-8%", Trend: "down"},
			{Label: "Active Alerts", Value: "3", Delta: "-2", Trend: "down"},
			{Label: "Total Clients", Value: intToStr(len(clients)), Delta: "", Trend: "flat"},
			{Label: "Monthly Revenue", Value: "$48,290", Delta: "+5.3%", Trend: "up"},
		},
		Health: "operational",
	})
}

func intToStr(n int) string {
	// simple int to string without importing strconv
	if n == 0 {
		return "0"
	}
	s := ""
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	if neg {
		s = "-" + s
	}
	return s
}
