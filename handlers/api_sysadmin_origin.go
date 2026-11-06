package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"platform/db"
)

type apiOriginDomain struct {
	ID        int64  `json:"id"`
	Domain    string `json:"domain"`
	Note      string `json:"note"`
	Position  int    `json:"position"`
	Enabled   bool   `json:"enabled"`
	IsPrimary bool   `json:"is_primary"`
}

type apiOriginTenantSummary struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Slug   string `json:"slug"`
	Domain string `json:"domain"`
}

type apiOriginOverviewResponse struct {
	Settings           db.OriginSettings        `json:"settings"`
	Domains            []apiOriginDomain        `json:"domains"`
	EnforcementActive  bool                     `json:"enforcement_active"`
	PrimaryDomain      string                   `json:"primary_domain,omitempty"`
	PrimaryRedirectURL string                   `json:"primary_redirect_url,omitempty"`
	TenantDomains      []apiOriginTenantSummary `json:"tenant_domains"`
}

// GET /api/sysadmin/origin
func apiSysadminOriginGetHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok || user.Role != RoleSysAdmin {
		apiError(w, http.StatusForbidden, "Sysadmin access required.")
		return
	}

	snap := origins.snapshot(r.Context())
	doms := make([]apiOriginDomain, 0, len(snap.domains))
	primaryFound := false
	var primaryDomain string
	for i, d := range snap.domains {
		isPrimary := false
		if d.Enabled && !primaryFound {
			isPrimary = true
			primaryFound = true
			primaryDomain = d.Domain
		}
		doms = append(doms, apiOriginDomain{
			ID:        d.ID,
			Domain:    d.Domain,
			Note:      d.Note,
			Position:  i + 1,
			Enabled:   d.Enabled,
			IsPrimary: isPrimary,
		})
	}

	var tenantSummaries []apiOriginTenantSummary
	for _, t := range store.list() {
		if strings.TrimSpace(t.Domain) != "" {
			tenantSummaries = append(tenantSummaries, apiOriginTenantSummary{
				ID:     t.ID,
				Name:   t.Name,
				Slug:   t.Slug,
				Domain: t.Domain,
			})
		}
	}
	if tenantSummaries == nil {
		tenantSummaries = []apiOriginTenantSummary{}
	}

	var redirectURL string
	if primaryFound {
		redirectURL = snap.settings.RedirectScheme + "://" + primaryDomain
	}

	apiOK(w, apiOriginOverviewResponse{
		Settings:           snap.settings,
		Domains:            doms,
		EnforcementActive:  primaryFound,
		PrimaryDomain:      primaryDomain,
		PrimaryRedirectURL: redirectURL,
		TenantDomains:      tenantSummaries,
	})
}

// PUT /api/sysadmin/origin
func apiSysadminOriginUpdateHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok || user.Role != RoleSysAdmin {
		apiError(w, http.StatusForbidden, "Sysadmin access required.")
		return
	}

	var req struct {
		RedirectScheme        *string `json:"redirect_scheme"`
		TrustForwardedHeaders *bool   `json:"trust_forwarded_headers"`
		RedirectStatusCode    *int    `json:"redirect_status_code"`
		OrderedDomainIDs      []int64 `json:"ordered_domain_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}

	cur := origins.getSettings(r.Context())
	if req.RedirectScheme != nil {
		cur.RedirectScheme = *req.RedirectScheme
	}
	if req.TrustForwardedHeaders != nil {
		cur.TrustForwardedHeaders = *req.TrustForwardedHeaders
	}
	if req.RedirectStatusCode != nil {
		cur.RedirectStatusCode = *req.RedirectStatusCode
	}

	updated, err := origins.updateSettings(r.Context(), cur)
	if err != nil {
		apiError(w, http.StatusBadRequest, err.Error())
		return
	}

	if len(req.OrderedDomainIDs) > 0 {
		if err := origins.reorderDomains(r.Context(), req.OrderedDomainIDs); err != nil {
			apiError(w, http.StatusInternalServerError, "Failed to reorder origin domains.")
			return
		}
	}

	if Pools != nil {
		go db.InsertAuditEvent(context.Background(), Pools.Admin, user.ID, string(user.Role),
			"sysadmin.origin.settings_updated", "info",
			map[string]any{
				"redirect_scheme":         updated.RedirectScheme,
				"redirect_status_code":    updated.RedirectStatusCode,
				"trust_forwarded_headers": updated.TrustForwardedHeaders,
			},
			"", r.Header.Get("X-Request-ID"), clientIP(r))
	}

	apiOK(w, map[string]any{
		"ok":       true,
		"settings": updated,
	})
}

// POST /api/sysadmin/origin/domains
func apiSysadminOriginAddDomainHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok || user.Role != RoleSysAdmin {
		apiError(w, http.StatusForbidden, "Sysadmin access required.")
		return
	}

	var req struct {
		Domain  string `json:"domain"`
		Note    string `json:"note"`
		Enabled *bool  `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	added, err := origins.addDomain(r.Context(), req.Domain, req.Note, enabled)
	if err != nil {
		if errors.Is(err, db.ErrInvalidDomain) {
			apiError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, db.ErrOriginDomainConflict) {
			apiError(w, http.StatusConflict, err.Error())
			return
		}
		apiError(w, http.StatusInternalServerError, "Failed to add origin domain.")
		return
	}

	if Pools != nil {
		go db.InsertAuditEvent(context.Background(), Pools.Admin, user.ID, string(user.Role),
			"sysadmin.origin.domain_added", "info",
			map[string]any{"id": added.ID, "domain": added.Domain, "enabled": added.Enabled},
			"", r.Header.Get("X-Request-ID"), clientIP(r))
	}

	apiCreated(w, added)
}

// PATCH /api/sysadmin/origin/domains/{id}
func apiSysadminOriginPatchDomainHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok || user.Role != RoleSysAdmin {
		apiError(w, http.StatusForbidden, "Sysadmin access required.")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		apiError(w, http.StatusBadRequest, "Invalid domain ID.")
		return
	}

	existing, exists := origins.getDomain(r.Context(), id)
	if !exists {
		apiError(w, http.StatusNotFound, "Origin domain not found.")
		return
	}

	var req struct {
		Note    *string `json:"note"`
		Enabled *bool   `json:"enabled"`
		Move    *string `json:"move"` // "top", "up", "down"
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}

	note := existing.Note
	if req.Note != nil {
		note = *req.Note
	}
	enabled := existing.Enabled
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	if req.Note != nil || req.Enabled != nil {
		updated, ok, err := origins.updateDomain(r.Context(), id, note, enabled)
		if err != nil || !ok {
			apiError(w, http.StatusInternalServerError, "Failed to update origin domain.")
			return
		}
		existing = updated
	}

	if req.Move != nil && strings.TrimSpace(*req.Move) != "" {
		if err := origins.moveDomain(r.Context(), id, *req.Move); err != nil {
			apiError(w, http.StatusBadRequest, err.Error())
			return
		}
		if refreshed, ok := origins.getDomain(r.Context(), id); ok {
			existing = refreshed
		}
	}

	if Pools != nil {
		go db.InsertAuditEvent(context.Background(), Pools.Admin, user.ID, string(user.Role),
			"sysadmin.origin.domain_updated", "info",
			map[string]any{"id": id, "domain": existing.Domain, "enabled": existing.Enabled},
			"", r.Header.Get("X-Request-ID"), clientIP(r))
	}

	apiOK(w, existing)
}

// DELETE /api/sysadmin/origin/domains/{id}
func apiSysadminOriginDeleteDomainHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok || user.Role != RoleSysAdmin {
		apiError(w, http.StatusForbidden, "Sysadmin access required.")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		apiError(w, http.StatusBadRequest, "Invalid domain ID.")
		return
	}

	existing, _ := origins.getDomain(r.Context(), id)
	deleted, err := origins.deleteDomain(r.Context(), id)
	if err != nil {
		apiError(w, http.StatusInternalServerError, "Failed to delete origin domain.")
		return
	}
	if !deleted {
		apiError(w, http.StatusNotFound, "Origin domain not found.")
		return
	}

	if Pools != nil {
		go db.InsertAuditEvent(context.Background(), Pools.Admin, user.ID, string(user.Role),
			"sysadmin.origin.domain_deleted", "warn",
			map[string]any{"id": id, "domain": existing.Domain},
			"", r.Header.Get("X-Request-ID"), clientIP(r))
	}

	apiOK(w, map[string]any{
		"ok":      true,
		"message": "Origin domain deleted.",
	})
}
