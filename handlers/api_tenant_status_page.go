package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
)

// ── GET /api/tenant/{slug}/status-page ──────────────────────

type apiStatusPageComponent struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Group   string `json:"group"`
	Status  string `json:"status"`
	Uptime  string `json:"uptime"`
	Visible bool   `json:"visible"`
	Created string `json:"created"`
	Updated string `json:"updated"`
}

type apiStatusPageListResponse struct {
	Components []apiStatusPageComponent `json:"components"`
	Total      int                      `json:"total"`
}

func apiTenantStatusPageListHandler(w http.ResponseWriter, r *http.Request) {
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

	components := filterStatusComponents(statusComponentStore.list(slug), search)

	out := make([]apiStatusPageComponent, len(components))
	for i, c := range components {
		out[i] = apiStatusPageComponent{
			ID: c.ID, Name: c.Name, Group: c.GroupName,
			Status: c.Status, Uptime: c.Uptime90d + "%",
			Visible: c.Visible,
			Created: c.Created, Updated: c.Updated,
		}
	}

	apiOK(w, apiStatusPageListResponse{Components: out, Total: len(out)})
}

// ── POST /api/tenant/{slug}/status-page ─────────────────────

type apiCreateStatusPageComponentRequest struct {
	Name   string `json:"name"`
	Group  string `json:"group"`
	Status string `json:"status"`
}

func apiTenantStatusPageCreateHandler(w http.ResponseWriter, r *http.Request) {
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

	var req apiCreateStatusPageComponentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}

	name := strings.TrimSpace(req.Name)
	group := strings.TrimSpace(req.Group)
	status := strings.TrimSpace(req.Status)

	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	if group == "" {
		group = "General"
	}
	if status == "" {
		status = "operational"
	}

	c := StatusPageComponent{
		ID: newID(), Name: name, GroupName: group,
		Status: status, Uptime90d: "100.00", Visible: true,
		Created: "2026-04-17T00:00:00Z", Updated: "2026-04-17T00:00:00Z",
	}
	statusComponentStore.add(slug, c)

	apiCreated(w, apiStatusPageComponent{
		ID: c.ID, Name: c.Name, Group: c.GroupName,
		Status: c.Status, Uptime: c.Uptime90d + "%",
		Visible: c.Visible,
		Created: c.Created, Updated: c.Updated,
	})
}

// ── GET /api/tenant/{slug}/status-page/{id} ─────────────────

func apiTenantStatusPageGetHandler(w http.ResponseWriter, r *http.Request) {
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
	c, ok := statusComponentStore.getByID(slug, id)
	if !ok {
		apiError(w, http.StatusNotFound, "Status page component not found.")
		return
	}

	apiOK(w, apiStatusPageComponent{
		ID: c.ID, Name: c.Name, Group: c.GroupName,
		Status: c.Status, Uptime: c.Uptime90d + "%",
		Visible: c.Visible,
		Created: c.Created, Updated: c.Updated,
	})
}

// ── DELETE /api/tenant/{slug}/status-page/{id} ──────────────

func apiTenantStatusPageDeleteHandler(w http.ResponseWriter, r *http.Request) {
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
	if _, ok := statusComponentStore.getByID(slug, id); !ok {
		apiError(w, http.StatusNotFound, "Status page component not found.")
		return
	}

	statusComponentStore.delete(slug, id)
	apiOK(w, map[string]any{
		"ok":      true,
		"message": "Status page component deleted.",
	})
}
