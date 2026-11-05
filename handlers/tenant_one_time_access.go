package handlers

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"platform/db"
)

// ── One-Time Access store ────────────────────────────────────

type OneTimeAccess struct {
	ID        string
	Label     string
	Token     string
	Target    string // resource path or workspace ID
	GrantedTo string
	GrantedBy string
	Status    string // "active" | "used" | "expired" | "revoked"
	ExpiresAt string
	Created   string
}

type oneTimeAccessStoreType struct {
	mu     sync.RWMutex
	tokens map[string][]OneTimeAccess // key: tenant slug
}

var oneTimeAccessStore = &oneTimeAccessStoreType{
	tokens: map[string][]OneTimeAccess{
		"buyvm": {
			{ID: newID(), Label: "Emergency DB access for Acme", Token: "ota-a1b2c3d4", Target: "workspace:production", GrantedTo: "james@acme.io", GrantedBy: "alice@buyvm.net", Status: "used", ExpiresAt: "2026-04-16T23:59:00Z", Created: "2026-04-16T10:00:00Z"},
			{ID: newID(), Label: "Vendor audit access", Token: "ota-e5f6g7h8", Target: "workspace:staging", GrantedTo: "auditor@external.com", GrantedBy: "alice@buyvm.net", Status: "active", ExpiresAt: "2026-04-18T23:59:00Z", Created: "2026-04-17T08:00:00Z"},
			{ID: newID(), Label: "Temp support access", Token: "ota-i9j0k1l2", Target: "workspace:default", GrantedTo: "contractor@support.io", GrantedBy: "bob@buyvm.net", Status: "expired", ExpiresAt: "2026-04-14T12:00:00Z", Created: "2026-04-13T09:00:00Z"},
		},
		"abrnoc": {
			{ID: newID(), Label: "Partner integration test", Token: "ota-m3n4o5p6", Target: "workspace:backup", GrantedTo: "dev@partner.io", GrantedBy: "david@abrnoc.com", Status: "active", ExpiresAt: "2026-04-19T18:00:00Z", Created: "2026-04-17T06:00:00Z"},
		},
		"hosting": {
			{ID: newID(), Label: "Client data export", Token: "ota-q7r8s9t0", Target: "workspace:us-east", GrantedTo: "alice@monroe.io", GrantedBy: "frank@hosting.com", Status: "active", ExpiresAt: "2026-04-18T12:00:00Z", Created: "2026-04-17T09:00:00Z"},
			{ID: newID(), Label: "Security scan access", Token: "ota-u1v2w3x4", Target: "workspace:prod", GrantedTo: "scanner@secops.com", GrantedBy: "frank@hosting.com", Status: "revoked", ExpiresAt: "2026-04-20T00:00:00Z", Created: "2026-04-15T14:00:00Z"},
			{ID: newID(), Label: "Migration helper token", Token: "ota-y5z6a7b8", Target: "workspace:eu-central", GrantedTo: "chloe@martindev.fr", GrantedBy: "grace@hosting.com", Status: "used", ExpiresAt: "2026-04-16T20:00:00Z", Created: "2026-04-16T08:00:00Z"},
		},
	},
}

func (s *oneTimeAccessStoreType) list(slug string) []OneTimeAccess {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.tokens[slug]
	out := make([]OneTimeAccess, len(src))
	copy(out, src)
	return out
}

func (s *oneTimeAccessStoreType) getByID(slug, id string) (OneTimeAccess, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, t := range s.tokens[slug] {
		if t.ID == id {
			return t, true
		}
	}
	return OneTimeAccess{}, false
}

func (s *oneTimeAccessStoreType) add(slug string, t OneTimeAccess) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tokens[slug] = append(s.tokens[slug], t)
}

func (s *oneTimeAccessStoreType) revoke(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, t := range s.tokens[slug] {
		if t.ID == id {
			s.tokens[slug][i].Status = "revoked"
			return
		}
	}
}

func (s *oneTimeAccessStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.tokens[slug]
	for i, t := range list {
		if t.ID == id {
			s.tokens[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

func filterOneTimeAccess(all []OneTimeAccess, search, status string) []OneTimeAccess {
	var out []OneTimeAccess
	q := strings.ToLower(search)
	for _, t := range all {
		if status != "" && t.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(t.Label + " " + t.GrantedTo + " " + t.Target)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, t)
	}
	return out
}

func dbOTAToLocal(r db.OneTimeAccess) OneTimeAccess {
	return OneTimeAccess{
		ID:        fmt.Sprintf("%d", r.ID),
		Label:     r.Label,
		Token:     r.Token,
		Target:    r.Target,
		GrantedTo: r.GrantedTo,
		GrantedBy: r.GrantedBy,
		Status:    r.Status,
		ExpiresAt: r.ExpiresAt.Format(time.RFC3339),
		Created:   r.CreatedAt.Format(time.RFC3339),
	}
}

// ── Page data ─────────────────────────────────────────────────

type oneTimeAccessInnerData struct {
	Tokens       []OneTimeAccess
	TenantSlug   string
	Search       string
	StatusFilter string
}

func renderOneTimeAccessPage(w http.ResponseWriter, r *http.Request, tenant Tenant, user User, d oneTimeAccessInnerData) {
	innerTmpl, err := template.ParseFiles(filepath.Join("templates", "tenant", "one_time_access_content.html"))
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	var buf bytes.Buffer
	if err := innerTmpl.ExecuteTemplate(&buf, "tenant-page-content", d); err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	shellPaths := []string{
		filepath.Join("templates", "tenant", "base.html"),
		filepath.Join("templates", "tenant", "shell.html"),
	}
	shellTmpl, err := template.ParseFiles(shellPaths...)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	data := TenantShellData{
		User: user, Tenant: tenant,
		Nav: tenantNav(tenant.Slug), NavGroups: tenantNavGroups(tenant.Slug),
		ActiveNav: "one-time-access", PageTitle: "One-Time Access",
		IsSysAdmin: user.Role == RoleSysAdmin,
		Content:    template.HTML(buf.String()),
	}
	if err := shellTmpl.ExecuteTemplate(w, "tenant-base", data); err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
	}
}

// ── Handlers ─────────────────────────────────────────────────

func oneTimeAccessHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := getSession(r)
	tenant, _ := tenantFromContext(r)
	q := r.URL.Query().Get("q")
	status := r.URL.Query().Get("status")

	var tokens []OneTimeAccess
	if Pools != nil && tenant.DBID > 0 {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		dts, err := db.ListOneTimeAccess(ctx, Pools.Admin, tenant.DBID, q, status)
		if err != nil {
			log.Printf("oneTimeAccessHandler: ListOneTimeAccess: %v", err)
			tokens = filterOneTimeAccess(oneTimeAccessStore.list(tenant.Slug), q, status)
		} else {
			tokens = make([]OneTimeAccess, len(dts))
			for i, dt := range dts {
				tokens[i] = dbOTAToLocal(dt)
			}
		}
	} else {
		tokens = filterOneTimeAccess(oneTimeAccessStore.list(tenant.Slug), q, status)
	}

	renderOneTimeAccessPage(w, r, tenant, user, oneTimeAccessInnerData{
		Tokens:       tokens,
		TenantSlug:   tenant.Slug,
		Search:       q,
		StatusFilter: status,
	})
}

func oneTimeAccessRevokeHandler(w http.ResponseWriter, r *http.Request) {
	tenant, _ := tenantFromContext(r)
	idStr := r.PathValue("id")

	if Pools != nil && tenant.DBID > 0 {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err == nil {
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()
			if err := db.RevokeOneTimeAccess(ctx, Pools.Admin, tenant.DBID, id); err != nil {
				log.Printf("oneTimeAccessRevokeHandler: %v", err)
			}
		}
	} else {
		oneTimeAccessStore.revoke(tenant.Slug, idStr)
	}
	http.Redirect(w, r, "/t/"+tenant.Slug+"/one-time-access/", http.StatusSeeOther)
}

func oneTimeAccessDeleteHandler(w http.ResponseWriter, r *http.Request) {
	tenant, _ := tenantFromContext(r)
	idStr := r.PathValue("id")

	if Pools != nil && tenant.DBID > 0 {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err == nil {
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()
			if err := db.DeleteOneTimeAccess(ctx, Pools.Admin, tenant.DBID, id); err != nil {
				log.Printf("oneTimeAccessDeleteHandler: %v", err)
			}
		}
	} else {
		oneTimeAccessStore.delete(tenant.Slug, idStr)
	}
	http.Redirect(w, r, "/t/"+tenant.Slug+"/one-time-access/", http.StatusSeeOther)
}

func RegisterTenantOneTimeAccessRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /t/{slug}/one-time-access/", tenantMiddleware(oneTimeAccessHandler))
	mux.HandleFunc("POST /t/{slug}/one-time-access/{id}/revoke/", tenantMiddleware(oneTimeAccessRevokeHandler))
	mux.HandleFunc("POST /t/{slug}/one-time-access/{id}/delete/", tenantMiddleware(oneTimeAccessDeleteHandler))
}
