package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
)

// ── GET /api/tenant/{slug}/abuse-reports ─────────────────────

type apiAbuseReport struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Source  string `json:"source"`
	Target  string `json:"target"`
	Status  string `json:"status"`
	Summary string `json:"summary"`
	Created string `json:"created"`
	Updated string `json:"updated"`
}

type apiAbuseReportsListResponse struct {
	Reports []apiAbuseReport `json:"reports"`
	Total   int              `json:"total"`
}

func apiTenantAbuseReportsListHandler(w http.ResponseWriter, r *http.Request) {
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
	abuseType := r.URL.Query().Get("type")

	reports := filterAbuseReports(abuseReportStore.list(slug), search, status, abuseType)

	out := make([]apiAbuseReport, len(reports))
	for i, rpt := range reports {
		out[i] = apiAbuseReport{
			ID: rpt.ID, Type: rpt.Type, Source: rpt.Source,
			Target: rpt.Target, Status: rpt.Status, Summary: rpt.Summary,
			Created: rpt.Created, Updated: rpt.Updated,
		}
	}

	apiOK(w, apiAbuseReportsListResponse{Reports: out, Total: len(out)})
}

// ── POST /api/tenant/{slug}/abuse-reports ────────────────────

type apiCreateAbuseReportRequest struct {
	Type    string `json:"type"`
	Source  string `json:"source"`
	Target  string `json:"target"`
	Summary string `json:"summary"`
}

func apiTenantAbuseReportsCreateHandler(w http.ResponseWriter, r *http.Request) {
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

	var req apiCreateAbuseReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}

	abuseType := strings.TrimSpace(req.Type)
	source := strings.TrimSpace(req.Source)
	target := strings.TrimSpace(req.Target)
	summary := strings.TrimSpace(req.Summary)

	if source == "" || target == "" || summary == "" {
		apiError(w, http.StatusBadRequest, "Source, target, and summary are required.")
		return
	}
	if abuseType == "" {
		abuseType = "other"
	}

	rpt := AbuseReport{
		ID: newID(), Type: abuseType, Source: source, Target: target,
		Status: "new", Summary: summary,
		Created: "2026-04-17T00:00:00Z", Updated: "2026-04-17T00:00:00Z",
	}
	abuseReportStore.add(slug, rpt)

	apiCreated(w, apiAbuseReport{
		ID: rpt.ID, Type: rpt.Type, Source: rpt.Source,
		Target: rpt.Target, Status: rpt.Status, Summary: rpt.Summary,
		Created: rpt.Created, Updated: rpt.Updated,
	})
}

// ── GET /api/tenant/{slug}/abuse-reports/{id} ────────────────

func apiTenantAbuseReportsGetHandler(w http.ResponseWriter, r *http.Request) {
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
	rpt, ok := abuseReportStore.getByID(slug, id)
	if !ok {
		apiError(w, http.StatusNotFound, "Abuse report not found.")
		return
	}

	apiOK(w, apiAbuseReport{
		ID: rpt.ID, Type: rpt.Type, Source: rpt.Source,
		Target: rpt.Target, Status: rpt.Status, Summary: rpt.Summary,
		Created: rpt.Created, Updated: rpt.Updated,
	})
}

// ── DELETE /api/tenant/{slug}/abuse-reports/{id} ─────────────

func apiTenantAbuseReportsDeleteHandler(w http.ResponseWriter, r *http.Request) {
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
	if _, ok := abuseReportStore.getByID(slug, id); !ok {
		apiError(w, http.StatusNotFound, "Abuse report not found.")
		return
	}

	abuseReportStore.delete(slug, id)
	apiOK(w, map[string]any{
		"ok":      true,
		"message": "Abuse report deleted.",
	})
}
