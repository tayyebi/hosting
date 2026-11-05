package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
)

// ── GET /api/tenant/{slug}/incidents ────────────────────────

type apiIncident struct {
	ID               string `json:"id"`
	Title            string `json:"title"`
	Severity         string `json:"severity"`
	Status           string `json:"status"`
	AffectedServices string `json:"affected_services"`
	Duration         string `json:"duration"`
	Created          string `json:"created"`
	Updated          string `json:"updated"`
}

type apiIncidentsListResponse struct {
	Incidents []apiIncident `json:"incidents"`
	Total     int           `json:"total"`
}

func apiTenantIncidentsListHandler(w http.ResponseWriter, r *http.Request) {
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
	severity := r.URL.Query().Get("severity")
	status := r.URL.Query().Get("status")

	incidents := filterIncidents(incidentStore.list(slug), search, severity, status)

	out := make([]apiIncident, len(incidents))
	for i, inc := range incidents {
		out[i] = apiIncident{
			ID: inc.ID, Title: inc.Title, Severity: inc.Severity,
			Status: inc.Status, AffectedServices: inc.Affected,
			Duration: inc.Duration,
			Created: inc.Created, Updated: inc.Updated,
		}
	}

	apiOK(w, apiIncidentsListResponse{Incidents: out, Total: len(out)})
}

// ── POST /api/tenant/{slug}/incidents ───────────────────────

type apiCreateIncidentRequest struct {
	Title            string `json:"title"`
	Severity         string `json:"severity"`
	AffectedServices string `json:"affected_services"`
}

func apiTenantIncidentsCreateHandler(w http.ResponseWriter, r *http.Request) {
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

	var req apiCreateIncidentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}

	title := strings.TrimSpace(req.Title)
	severity := strings.TrimSpace(req.Severity)
	services := strings.TrimSpace(req.AffectedServices)

	if title == "" {
		apiError(w, http.StatusBadRequest, "Title is required.")
		return
	}
	if severity == "" {
		severity = "minor"
	}

	inc := TenantIncident{
		ID: newID(), Title: title, Severity: severity,
		Status: "investigating", Affected: services,
		Duration: "0m ongoing",
		Created: "2026-04-17T00:00:00Z", Updated: "2026-04-17T00:00:00Z",
	}
	incidentStore.add(slug, inc)

	apiCreated(w, apiIncident{
		ID: inc.ID, Title: inc.Title, Severity: inc.Severity,
		Status: inc.Status, AffectedServices: inc.Affected,
		Duration: inc.Duration,
		Created: inc.Created, Updated: inc.Updated,
	})
}

// ── GET /api/tenant/{slug}/incidents/{id} ───────────────────

func apiTenantIncidentsGetHandler(w http.ResponseWriter, r *http.Request) {
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
	inc, ok := incidentStore.getByID(slug, id)
	if !ok {
		apiError(w, http.StatusNotFound, "Incident not found.")
		return
	}

	apiOK(w, apiIncident{
		ID: inc.ID, Title: inc.Title, Severity: inc.Severity,
		Status: inc.Status, AffectedServices: inc.Affected,
		Duration: inc.Duration,
		Created: inc.Created, Updated: inc.Updated,
	})
}

// ── DELETE /api/tenant/{slug}/incidents/{id} ────────────────

func apiTenantIncidentsDeleteHandler(w http.ResponseWriter, r *http.Request) {
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
	if _, ok := incidentStore.getByID(slug, id); !ok {
		apiError(w, http.StatusNotFound, "Incident not found.")
		return
	}

	incidentStore.delete(slug, id)
	apiOK(w, map[string]any{
		"ok":      true,
		"message": "Incident deleted.",
	})
}
