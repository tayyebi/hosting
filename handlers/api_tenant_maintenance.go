package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
)

// ── GET /api/tenant/{slug}/maintenance ─────────────────────

type apiMaintenanceWindow struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Status      string `json:"status"`
	Services    string `json:"services"`
	ScheduledAt string `json:"scheduled_at"`
	Risk        string `json:"risk"`
	Rollback    string `json:"rollback"`
	Created     string `json:"created"`
	Updated     string `json:"updated"`
}

type apiMaintenanceListResponse struct {
	Windows []apiMaintenanceWindow `json:"windows"`
	Total   int                    `json:"total"`
}

func apiTenantMaintenanceListHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok {
		apiError(w, http.StatusUnauthorized, "Authentication required.")
		return
	}

	slug := r.PathValue("slug")
	if _, found := store.getBySlug(slug); !found {
		apiError(w, http.StatusNotFound, "Tenant not found.")
		return
	}
	if user.Role != RoleSysAdmin && (user.Role != RoleTenant || user.TenantSlug != slug) {
		apiError(w, http.StatusForbidden, "Access denied.")
		return
	}

	search := r.URL.Query().Get("q")
	status := r.URL.Query().Get("status")

	windows := filterMaintenanceWindows(maintenanceStore.list(slug), search, status)

	out := make([]apiMaintenanceWindow, len(windows))
	for i, mw := range windows {
		out[i] = apiMaintenanceWindow{
			ID: mw.ID, Title: mw.Title, Status: mw.Status,
			Services: mw.Affected, ScheduledAt: mw.ScheduledStart,
			Risk: mw.Risk, Rollback: mw.Rollback,
			Created: mw.Created, Updated: mw.Updated,
		}
	}

	apiOK(w, apiMaintenanceListResponse{Windows: out, Total: len(out)})
}

// ── POST /api/tenant/{slug}/maintenance ────────────────────

type apiCreateMaintenanceRequest struct {
	Title    string `json:"title"`
	Services string `json:"services"`
	Risk     string `json:"risk"`
	Rollback string `json:"rollback"`
}

func apiTenantMaintenanceCreateHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok {
		apiError(w, http.StatusUnauthorized, "Authentication required.")
		return
	}

	slug := r.PathValue("slug")
	if _, found := store.getBySlug(slug); !found {
		apiError(w, http.StatusNotFound, "Tenant not found.")
		return
	}
	if user.Role != RoleSysAdmin && (user.Role != RoleTenant || user.TenantSlug != slug) {
		apiError(w, http.StatusForbidden, "Access denied.")
		return
	}

	var req apiCreateMaintenanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}

	title := strings.TrimSpace(req.Title)
	services := strings.TrimSpace(req.Services)
	risk := strings.TrimSpace(req.Risk)
	rollback := strings.TrimSpace(req.Rollback)

	if title == "" {
		apiError(w, http.StatusBadRequest, "Title is required.")
		return
	}
	if risk == "" {
		risk = "low"
	}

	mw := MaintenanceWin{
		ID: newID(), Title: title, Status: "scheduled",
		Affected: services, Risk: risk, Rollback: rollback,
		Created: "2026-04-17T00:00:00Z", Updated: "2026-04-17T00:00:00Z",
	}
	maintenanceStore.add(slug, mw)

	apiCreated(w, apiMaintenanceWindow{
		ID: mw.ID, Title: mw.Title, Status: mw.Status,
		Services: mw.Affected, ScheduledAt: mw.ScheduledStart,
		Risk: mw.Risk, Rollback: mw.Rollback,
		Created: mw.Created, Updated: mw.Updated,
	})
}

// ── GET /api/tenant/{slug}/maintenance/{id} ────────────────

func apiTenantMaintenanceGetHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok {
		apiError(w, http.StatusUnauthorized, "Authentication required.")
		return
	}

	slug := r.PathValue("slug")
	if _, found := store.getBySlug(slug); !found {
		apiError(w, http.StatusNotFound, "Tenant not found.")
		return
	}
	if user.Role != RoleSysAdmin && (user.Role != RoleTenant || user.TenantSlug != slug) {
		apiError(w, http.StatusForbidden, "Access denied.")
		return
	}

	id := r.PathValue("id")
	mw, ok := maintenanceStore.getByID(slug, id)
	if !ok {
		apiError(w, http.StatusNotFound, "Maintenance window not found.")
		return
	}

	apiOK(w, apiMaintenanceWindow{
		ID: mw.ID, Title: mw.Title, Status: mw.Status,
		Services: mw.Affected, ScheduledAt: mw.ScheduledStart,
		Risk: mw.Risk, Rollback: mw.Rollback,
		Created: mw.Created, Updated: mw.Updated,
	})
}

// ── DELETE /api/tenant/{slug}/maintenance/{id} ─────────────

func apiTenantMaintenanceDeleteHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok {
		apiError(w, http.StatusUnauthorized, "Authentication required.")
		return
	}

	slug := r.PathValue("slug")
	if _, found := store.getBySlug(slug); !found {
		apiError(w, http.StatusNotFound, "Tenant not found.")
		return
	}
	if user.Role != RoleSysAdmin && (user.Role != RoleTenant || user.TenantSlug != slug) {
		apiError(w, http.StatusForbidden, "Access denied.")
		return
	}

	id := r.PathValue("id")
	if _, ok := maintenanceStore.getByID(slug, id); !ok {
		apiError(w, http.StatusNotFound, "Maintenance window not found.")
		return
	}

	maintenanceStore.delete(slug, id)
	apiOK(w, map[string]any{
		"ok":      true,
		"message": "Maintenance window deleted.",
	})
}
