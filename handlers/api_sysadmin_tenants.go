package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"platform/db"
)

var domainRE = regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)

// parsePagination reads ?limit=N&offset=M from a request, capping limit at maxLimit.
func parsePagination(r *http.Request, defaultLimit, maxLimit int) (limit, offset int) {
	limit = defaultLimit
	offset = 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			if n > maxLimit {
				n = maxLimit
			}
			limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}
	return
}

// ── GET /api/sysadmin/tenants ───────────────────────────────

type apiTenant struct {
	ID     string          `json:"id"`
	Name   string          `json:"name"`
	Slug   string          `json:"slug"`
	Domain string          `json:"domain,omitempty"`
	Owner  *apiTenantOwner `json:"owner,omitempty"`
}

type apiTenantOwner struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type apiTenantsListResponse struct {
	Tenants []apiTenant `json:"tenants"`
	Total   int         `json:"total"`
	Limit   int         `json:"limit"`
	Offset  int         `json:"offset"`
}

func apiSysadminTenantsListHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok || user.Role != RoleSysAdmin {
		apiError(w, http.StatusForbidden, "Sysadmin access required.")
		return
	}

	limit, offset := parsePagination(r, 50, 100)

	tenants := store.list()
	total := len(tenants)
	if offset >= total {
		tenants = []Tenant{}
	} else {
		tenants = tenants[offset:]
		if len(tenants) > limit {
			tenants = tenants[:limit]
		}
	}

	allOwners := owners.all()
	out := make([]apiTenant, len(tenants))
	for i, t := range tenants {
		at := apiTenant{
			ID:     t.ID,
			Name:   t.Name,
			Slug:   t.Slug,
			Domain: t.Domain,
		}
		if o, ok := allOwners[t.ID]; ok {
			at.Owner = &apiTenantOwner{Name: o.Name, Email: o.Email}
		}
		out[i] = at
	}

	apiOK(w, apiTenantsListResponse{
		Tenants: out,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
	})
}

// ── POST /api/sysadmin/tenants ──────────────────────────────

type apiCreateTenantRequest struct {
	Name   string `json:"name"`
	Slug   string `json:"slug"`
	Domain string `json:"domain"`
}

func apiSysadminTenantsCreateHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok || user.Role != RoleSysAdmin {
		apiError(w, http.StatusForbidden, "Sysadmin access required.")
		return
	}

	var req apiCreateTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}

	name := strings.TrimSpace(req.Name)
	slug := strings.TrimSpace(req.Slug)
	domain := strings.TrimSpace(req.Domain)

	if name == "" || slug == "" {
		apiError(w, http.StatusBadRequest, "Name and slug are required.")
		return
	}
	if len(name) > 100 {
		apiError(w, http.StatusBadRequest, "Tenant name must not exceed 100 characters.")
		return
	}
	if domain != "" {
		norm, err := db.NormalizeDomain(domain)
		if err != nil {
			apiError(w, http.StatusBadRequest, "Domain must be a valid hostname (e.g. example.com).")
			return
		}
		domain = norm
		if store.domainExists(domain, "") {
			apiError(w, http.StatusConflict, "Domain \""+domain+"\" is already assigned to another tenant.")
			return
		}
		if origins.domainExists(domain, 0) {
			apiError(w, http.StatusConflict, "Domain \""+domain+"\" is already registered as a System Origin domain.")
			return
		}
	}
	if store.slugExists(slug, "") {
		apiError(w, http.StatusConflict, "Slug \""+slug+"\" is already in use.")
		return
	}

	t := Tenant{ID: newID(), Name: name, Slug: slug, Domain: domain}
	t = store.add(t)
	origins.invalidateCache()

	go db.InsertAuditEvent(context.Background(), Pools.Admin, user.ID, string(user.Role),
		"tenant.created", "info", map[string]any{"name": name, "slug": slug, "domain": domain},
		"", r.Header.Get("X-Request-ID"), clientIP(r))

	apiCreated(w, apiTenant{
		ID:     t.ID,
		Name:   t.Name,
		Slug:   t.Slug,
		Domain: t.Domain,
	})
}

// ── GET /api/sysadmin/tenants/{id} ──────────────────────────

func apiSysadminTenantsGetHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok || user.Role != RoleSysAdmin {
		apiError(w, http.StatusForbidden, "Sysadmin access required.")
		return
	}

	id := r.PathValue("id")
	t, ok := store.get(id)
	if !ok {
		apiError(w, http.StatusNotFound, "Tenant not found.")
		return
	}

	at := apiTenant{
		ID:     t.ID,
		Name:   t.Name,
		Slug:   t.Slug,
		Domain: t.Domain,
	}
	if o, exists := owners.get(t.ID); exists {
		at.Owner = &apiTenantOwner{Name: o.Name, Email: o.Email}
	}

	apiOK(w, at)
}

// ── PATCH /api/sysadmin/tenants/{id} ───────────────────────

type apiUpdateTenantRequest struct {
	Name   string `json:"name"`
	Domain string `json:"domain"`
}

func apiSysadminTenantsUpdateHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok || user.Role != RoleSysAdmin {
		apiError(w, http.StatusForbidden, "Sysadmin access required.")
		return
	}

	id := r.PathValue("id")
	t, ok := store.get(id)
	if !ok {
		apiError(w, http.StatusNotFound, "Tenant not found.")
		return
	}

	var req apiUpdateTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = t.Name
	}
	if len(name) > 100 {
		apiError(w, http.StatusBadRequest, "Tenant name must not exceed 100 characters.")
		return
	}

	domain := strings.TrimSpace(req.Domain)
	if domain != "" {
		norm, err := db.NormalizeDomain(domain)
		if err != nil {
			apiError(w, http.StatusBadRequest, "Domain must be a valid hostname (e.g. example.com).")
			return
		}
		domain = norm
		if store.domainExists(domain, id) {
			apiError(w, http.StatusConflict, "Domain \""+domain+"\" is already assigned to another tenant.")
			return
		}
		if origins.domainExists(domain, 0) {
			apiError(w, http.StatusConflict, "Domain \""+domain+"\" is already registered as a System Origin domain.")
			return
		}
	}

	t.Name = name
	t.Domain = domain
	store.update(t)
	origins.invalidateCache()

	go db.InsertAuditEvent(context.Background(), Pools.Admin, user.ID, string(user.Role),
		"tenant.updated", "info", map[string]any{"tenant_id": id, "slug": t.Slug, "name": name, "domain": domain},
		"", r.Header.Get("X-Request-ID"), clientIP(r))

	apiOK(w, apiTenant{
		ID:     t.ID,
		Name:   t.Name,
		Slug:   t.Slug,
		Domain: t.Domain,
	})
}

// ── DELETE /api/sysadmin/tenants/{id} ───────────────────────

func apiSysadminTenantsDeleteHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok || user.Role != RoleSysAdmin {
		apiError(w, http.StatusForbidden, "Sysadmin access required.")
		return
	}

	id := r.PathValue("id")
	t, ok := store.get(id)
	if !ok {
		apiError(w, http.StatusNotFound, "Tenant not found.")
		return
	}

	store.delete(id)
	origins.invalidateCache()
	go db.InsertAuditEvent(context.Background(), Pools.Admin, user.ID, string(user.Role),
		"tenant.deleted", "warn", map[string]any{"tenant_id": id, "slug": t.Slug, "name": t.Name},
		"", r.Header.Get("X-Request-ID"), clientIP(r))

	apiOK(w, map[string]any{
		"ok":      true,
		"message": "Tenant deleted.",
	})
}
