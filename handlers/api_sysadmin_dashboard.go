package handlers

import "net/http"

// ── GET /api/sysadmin/dashboard ─────────────────────────────

type apiSysadminNavItem struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Icon  string `json:"icon"`
	Path  string `json:"path"`
}

type apiSysadminDashboardResponse struct {
	Title    string               `json:"title"`
	Nav      []apiSysadminNavItem `json:"nav"`
	Tenants  int                  `json:"tenant_count"`
	Users    int                  `json:"user_count"`
	Replicas int                  `json:"replica_count"`
	Sessions int                  `json:"active_sessions"`
	Health   string               `json:"health"`
}

func apiSysadminDashboardHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok || user.Role != RoleSysAdmin {
		apiError(w, http.StatusForbidden, "Sysadmin access required.")
		return
	}

	nav := make([]apiSysadminNavItem, len(sysadminNav))
	for i, n := range sysadminNav {
		nav[i] = apiSysadminNavItem{
			ID:    n.ID,
			Label: n.Label,
			Icon:  n.Icon,
			Path:  n.Path,
		}
	}

	apiOK(w, apiSysadminDashboardResponse{
		Title:    "Sysadmin Dashboard",
		Nav:      nav,
		Tenants:  len(store.list()),
		Users:    1842,
		Replicas: 84,
		Sessions: 3291,
		Health:   "operational",
	})
}
