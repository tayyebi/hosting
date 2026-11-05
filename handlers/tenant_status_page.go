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

// ── Status Component store ────────────────────────────────────

type StatusPageComponent struct {
	ID          string
	Name        string
	GroupName   string
	Description string
	Status      string // "operational" | "degraded" | "partial-outage" | "major-outage" | "maintenance"
	Uptime90d   string
	OrderNum    int
	Visible     bool
	Created     string
	Updated     string
}

type statusComponentStoreType struct {
	mu         sync.RWMutex
	components map[string][]StatusPageComponent
}

var statusComponentStore = &statusComponentStoreType{
	components: map[string][]StatusPageComponent{},
}

func (s *statusComponentStoreType) list(slug string) []StatusPageComponent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.components[slug]
	out := make([]StatusPageComponent, len(src))
	copy(out, src)
	return out
}

func (s *statusComponentStoreType) getByID(slug, id string) (StatusPageComponent, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, c := range s.components[slug] {
		if c.ID == id {
			return c, true
		}
	}
	return StatusPageComponent{}, false
}

func (s *statusComponentStoreType) add(slug string, c StatusPageComponent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.components[slug] = append(s.components[slug], c)
}

func (s *statusComponentStoreType) updateStatus(slug, id, status string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, c := range s.components[slug] {
		if c.ID == id {
			s.components[slug][i].Status = status
			s.components[slug][i].Updated = time.Now().Format(time.RFC3339)
			return
		}
	}
}

func (s *statusComponentStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.components[slug]
	for i, c := range list {
		if c.ID == id {
			s.components[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

func filterStatusComponents(all []StatusPageComponent, search string) []StatusPageComponent {
	if search == "" {
		return all
	}
	var out []StatusPageComponent
	q := strings.ToLower(search)
	for _, c := range all {
		if strings.Contains(strings.ToLower(c.Name+" "+c.GroupName), q) {
			out = append(out, c)
		}
	}
	return out
}

func dbStatusComponentToLocal(c db.StatusComponent) StatusPageComponent {
	return StatusPageComponent{
		ID:          fmt.Sprintf("%d", c.ID),
		Name:        c.Name,
		GroupName:   c.GroupName,
		Description: c.Description,
		Status:      c.Status,
		Uptime90d:   fmt.Sprintf("%.2f", c.Uptime90d),
		OrderNum:    c.OrderNum,
		Visible:     c.Visible,
		Created:     c.CreatedAt.Format(time.RFC3339),
		Updated:     c.UpdatedAt.Format(time.RFC3339),
	}
}

// ── Page data ─────────────────────────────────────────────────

type statusPageInnerData struct {
	Components []StatusPageComponent
	TenantSlug string
	Search     string
}

func renderStatusPage(w http.ResponseWriter, r *http.Request, tenant Tenant, user User, d statusPageInnerData) {
	innerTmpl, err := template.ParseFiles(filepath.Join("templates", "tenant", "status_page_content.html"))
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
		ActiveNav: "status-page", PageTitle: "Status Page",
		IsSysAdmin: user.Role == RoleSysAdmin,
		Content:    template.HTML(buf.String()),
	}
	if err := shellTmpl.ExecuteTemplate(w, "tenant-base", data); err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
	}
}

// ── Handlers ─────────────────────────────────────────────────

func statusPageHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := getSession(r)
	tenant, _ := tenantFromContext(r)
	q := r.URL.Query().Get("q")

	var components []StatusPageComponent
	if Pools != nil && tenant.DBID > 0 {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		dcs, err := db.ListStatusComponents(ctx, Pools.Admin, tenant.DBID, q)
		if err != nil {
			log.Printf("statusPageHandler: ListStatusComponents: %v", err)
			components = filterStatusComponents(statusComponentStore.list(tenant.Slug), q)
		} else {
			components = make([]StatusPageComponent, len(dcs))
			for i, dc := range dcs {
				components[i] = dbStatusComponentToLocal(dc)
			}
		}
	} else {
		components = filterStatusComponents(statusComponentStore.list(tenant.Slug), q)
	}

	renderStatusPage(w, r, tenant, user, statusPageInnerData{
		Components: components,
		TenantSlug: tenant.Slug,
		Search:     q,
	})
}

func statusComponentStatusHandler(w http.ResponseWriter, r *http.Request) {
	tenant, _ := tenantFromContext(r)
	idStr := r.PathValue("id")
	newStatus := r.FormValue("status")

	if Pools != nil && tenant.DBID > 0 {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err == nil {
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()
			if err := db.UpdateStatusComponentStatus(ctx, Pools.Admin, tenant.DBID, id, newStatus); err != nil {
				log.Printf("statusComponentStatusHandler: %v", err)
			}
		}
	} else {
		statusComponentStore.updateStatus(tenant.Slug, idStr, newStatus)
	}
	http.Redirect(w, r, "/t/"+tenant.Slug+"/status-page/", http.StatusSeeOther)
}

func statusComponentDeleteHandler(w http.ResponseWriter, r *http.Request) {
	tenant, _ := tenantFromContext(r)
	idStr := r.PathValue("id")

	if Pools != nil && tenant.DBID > 0 {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err == nil {
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()
			if err := db.DeleteStatusComponent(ctx, Pools.Admin, tenant.DBID, id); err != nil {
				log.Printf("statusComponentDeleteHandler: %v", err)
			}
		}
	} else {
		statusComponentStore.delete(tenant.Slug, idStr)
	}
	http.Redirect(w, r, "/t/"+tenant.Slug+"/status-page/", http.StatusSeeOther)
}

func RegisterTenantStatusPageRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /t/{slug}/status-page/", tenantMiddleware(statusPageHandler))
	mux.HandleFunc("POST /t/{slug}/status-page/{id}/status/", tenantMiddleware(statusComponentStatusHandler))
	mux.HandleFunc("POST /t/{slug}/status-page/{id}/delete/", tenantMiddleware(statusComponentDeleteHandler))
}
