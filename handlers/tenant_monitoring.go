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

// ── Monitoring Check store ───────────────────────────────────

type TenantMonitoringCheck struct {
	ID          string
	Name        string
	CheckType   string
	Target      string
	Status      string // "up"|"degraded"|"down"|"paused"
	Uptime30d   string
	LatencyMs   int
	LastChecked string
	Created     string
	Updated     string
}

type tenantMonitoringStoreType struct {
	mu     sync.RWMutex
	checks map[string][]TenantMonitoringCheck // key: tenant slug
}

var tenantMonitoringStore = &tenantMonitoringStoreType{
	checks: map[string][]TenantMonitoringCheck{},
}

func (s *tenantMonitoringStoreType) list(slug string) []TenantMonitoringCheck {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.checks[slug]
	out := make([]TenantMonitoringCheck, len(src))
	copy(out, src)
	return out
}

func (s *tenantMonitoringStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.checks[slug]
	for i, c := range list {
		if c.ID == id {
			s.checks[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

func filterTenantMonitoringChecks(all []TenantMonitoringCheck, search, status string) []TenantMonitoringCheck {
	var out []TenantMonitoringCheck
	q := strings.ToLower(search)
	for _, c := range all {
		if status != "" && c.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(c.Name + " " + c.CheckType + " " + c.Target)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, c)
	}
	return out
}

func dbMonitoringCheckToLocal(r db.MonitoringCheck) TenantMonitoringCheck {
	lastChecked := ""
	if r.LastChecked != nil {
		lastChecked = r.LastChecked.Format(time.RFC3339)
	}
	return TenantMonitoringCheck{
		ID:          fmt.Sprintf("%d", r.ID),
		Name:        r.Name,
		CheckType:   r.CheckType,
		Target:      r.Target,
		Status:      r.Status,
		Uptime30d:   r.Uptime30d,
		LatencyMs:   r.LatencyMs,
		LastChecked: lastChecked,
		Created:     r.CreatedAt.Format(time.RFC3339),
		Updated:     r.UpdatedAt.Format(time.RFC3339),
	}
}

// ── Page data ─────────────────────────────────────────────────

type monitoringInnerData struct {
	Checks       []TenantMonitoringCheck
	TenantSlug   string
	Search       string
	StatusFilter string
}

func renderMonitoringPage(w http.ResponseWriter, r *http.Request, tenant Tenant, user User, d monitoringInnerData) {
	innerTmpl, err := template.ParseFiles(filepath.Join("templates", "tenant", "monitoring_content.html"))
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
		ActiveNav: "monitoring", PageTitle: "Monitoring",
		IsSysAdmin: user.Role == RoleSysAdmin,
		Content:    template.HTML(buf.String()),
	}
	if err := shellTmpl.ExecuteTemplate(w, "tenant-base", data); err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
	}
}

// ── Handlers ─────────────────────────────────────────────────

func monitoringHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := getSession(r)
	tenant, _ := tenantFromContext(r)
	q := r.URL.Query().Get("q")
	status := r.URL.Query().Get("status")

	var checks []TenantMonitoringCheck
	if Pools != nil && tenant.DBID > 0 {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		dcs, err := db.ListMonitoringChecks(ctx, Pools.Admin, tenant.DBID, q, status)
		if err != nil {
			log.Printf("monitoringHandler: ListMonitoringChecks: %v", err)
			checks = filterTenantMonitoringChecks(tenantMonitoringStore.list(tenant.Slug), q, status)
		} else {
			checks = make([]TenantMonitoringCheck, len(dcs))
			for i, dc := range dcs {
				checks[i] = dbMonitoringCheckToLocal(dc)
			}
		}
	} else {
		checks = filterTenantMonitoringChecks(tenantMonitoringStore.list(tenant.Slug), q, status)
	}

	renderMonitoringPage(w, r, tenant, user, monitoringInnerData{
		Checks:       checks,
		TenantSlug:   tenant.Slug,
		Search:       q,
		StatusFilter: status,
	})
}

func monitoringCheckDeleteHandler(w http.ResponseWriter, r *http.Request) {
	tenant, _ := tenantFromContext(r)
	idStr := r.PathValue("id")

	if Pools != nil && tenant.DBID > 0 {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err == nil {
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()
			if err := db.DeleteMonitoringCheck(ctx, Pools.Admin, tenant.DBID, id); err != nil {
				log.Printf("monitoringCheckDeleteHandler: %v", err)
			}
		}
	} else {
		tenantMonitoringStore.delete(tenant.Slug, idStr)
	}
	http.Redirect(w, r, "/t/"+tenant.Slug+"/monitoring/", http.StatusSeeOther)
}

func RegisterTenantMonitoringRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /t/{slug}/monitoring/", tenantMiddleware(monitoringHandler))
	mux.HandleFunc("POST /t/{slug}/monitoring/{id}/delete/", tenantMiddleware(monitoringCheckDeleteHandler))
}
