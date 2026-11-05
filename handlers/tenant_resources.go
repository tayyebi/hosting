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

// ── Resource store ────────────────────────────────────────────

type TenantResource struct {
	ID           string
	Name         string
	ResourceType string
	Provider     string
	Region       string
	Status       string // "online"|"degraded"|"offline"|"maintenance"|"provisioning"
	IPAddress    string
	Created      string
	Updated      string
}

type tenantResourceStoreType struct {
	mu        sync.RWMutex
	resources map[string][]TenantResource
}

var tenantResourceStore = &tenantResourceStoreType{
	resources: map[string][]TenantResource{},
}

func (s *tenantResourceStoreType) list(slug string) []TenantResource {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.resources[slug]
	out := make([]TenantResource, len(src))
	copy(out, src)
	return out
}

func (s *tenantResourceStoreType) updateStatus(slug, id, status string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, r := range s.resources[slug] {
		if r.ID == id {
			s.resources[slug][i].Status = status
			s.resources[slug][i].Updated = time.Now().Format(time.RFC3339)
			return
		}
	}
}

func (s *tenantResourceStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.resources[slug]
	for i, r := range list {
		if r.ID == id {
			s.resources[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

func filterTenantResources(all []TenantResource, search, status string) []TenantResource {
	var out []TenantResource
	q := strings.ToLower(search)
	for _, r := range all {
		if status != "" && r.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(r.Name + " " + r.ResourceType + " " + r.IPAddress + " " + r.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, r)
	}
	return out
}

func dbResourceToLocal(r db.TenantResource) TenantResource {
	return TenantResource{
		ID:           fmt.Sprintf("%d", r.ID),
		Name:         r.Name,
		ResourceType: r.ResourceType,
		Provider:     r.Provider,
		Region:       r.Region,
		Status:       r.Status,
		IPAddress:    r.IPAddress,
		Created:      r.CreatedAt.Format(time.RFC3339),
		Updated:      r.UpdatedAt.Format(time.RFC3339),
	}
}

// ── Page data ─────────────────────────────────────────────────

type resourcesInnerData struct {
	Resources    []TenantResource
	TenantSlug   string
	Search       string
	StatusFilter string
}

func renderResourcesPage(w http.ResponseWriter, r *http.Request, tenant Tenant, user User, d resourcesInnerData) {
	innerTmpl, err := template.ParseFiles(filepath.Join("templates", "tenant", "resources_content.html"))
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
		ActiveNav: "resources", PageTitle: "Resources",
		IsSysAdmin: user.Role == RoleSysAdmin,
		Content:    template.HTML(buf.String()),
	}
	if err := shellTmpl.ExecuteTemplate(w, "tenant-base", data); err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
	}
}

// ── Handlers ─────────────────────────────────────────────────

func resourcesHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := getSession(r)
	tenant, _ := tenantFromContext(r)
	q := r.URL.Query().Get("q")
	status := r.URL.Query().Get("status")

	var resources []TenantResource
	if Pools != nil && tenant.DBID > 0 {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		drs, err := db.ListTenantResources(ctx, Pools.Admin, tenant.DBID, q, status)
		if err != nil {
			log.Printf("resourcesHandler: ListTenantResources: %v", err)
			resources = filterTenantResources(tenantResourceStore.list(tenant.Slug), q, status)
		} else {
			resources = make([]TenantResource, len(drs))
			for i, dr := range drs {
				resources[i] = dbResourceToLocal(dr)
			}
		}
	} else {
		resources = filterTenantResources(tenantResourceStore.list(tenant.Slug), q, status)
	}

	renderResourcesPage(w, r, tenant, user, resourcesInnerData{
		Resources:    resources,
		TenantSlug:   tenant.Slug,
		Search:       q,
		StatusFilter: status,
	})
}

func resourceStatusHandler(w http.ResponseWriter, r *http.Request) {
	tenant, _ := tenantFromContext(r)
	idStr := r.PathValue("id")
	newStatus := r.FormValue("status")

	if Pools != nil && tenant.DBID > 0 {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err == nil {
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()
			if err := db.UpdateTenantResourceStatus(ctx, Pools.Admin, tenant.DBID, id, newStatus); err != nil {
				log.Printf("resourceStatusHandler: %v", err)
			}
		}
	} else {
		tenantResourceStore.updateStatus(tenant.Slug, idStr, newStatus)
	}
	http.Redirect(w, r, "/t/"+tenant.Slug+"/resources/", http.StatusSeeOther)
}

func resourceDeleteHandler(w http.ResponseWriter, r *http.Request) {
	tenant, _ := tenantFromContext(r)
	idStr := r.PathValue("id")

	if Pools != nil && tenant.DBID > 0 {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err == nil {
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()
			if err := db.DeleteTenantResource(ctx, Pools.Admin, tenant.DBID, id); err != nil {
				log.Printf("resourceDeleteHandler: %v", err)
			}
		}
	} else {
		tenantResourceStore.delete(tenant.Slug, idStr)
	}
	http.Redirect(w, r, "/t/"+tenant.Slug+"/resources/", http.StatusSeeOther)
}

func RegisterTenantResourcesRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /t/{slug}/resources/", tenantMiddleware(resourcesHandler))
	mux.HandleFunc("POST /t/{slug}/resources/{id}/status/", tenantMiddleware(resourceStatusHandler))
	mux.HandleFunc("POST /t/{slug}/resources/{id}/delete/", tenantMiddleware(resourceDeleteHandler))
}
