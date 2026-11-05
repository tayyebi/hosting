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
	checks: map[string][]TenantMonitoringCheck{
		"buyvm": {
			{ID: newID(), Name: "web-prod-01", CheckType: "HTTP", Target: "https://web-prod-01.buyvm.net/health", Status: "up", Uptime30d: "100.00", LatencyMs: 22, LastChecked: "2026-04-17T00:00:00Z", Created: "2026-01-10T08:00:00Z", Updated: "2026-04-17T00:00:00Z"},
			{ID: newID(), Name: "web-prod-02", CheckType: "HTTP", Target: "https://web-prod-02.buyvm.net/health", Status: "up", Uptime30d: "100.00", LatencyMs: 19, LastChecked: "2026-04-17T00:00:00Z", Created: "2026-01-10T08:00:00Z", Updated: "2026-04-17T00:00:00Z"},
			{ID: newID(), Name: "api-gateway", CheckType: "TCP", Target: "185.22.153.1:443", Status: "degraded", Uptime30d: "99.72", LatencyMs: 148, LastChecked: "2026-04-17T00:00:00Z", Created: "2026-01-10T08:00:00Z", Updated: "2026-04-17T00:00:00Z"},
			{ID: newID(), Name: "db-primary", CheckType: "TCP", Target: "185.22.153.11:5432", Status: "up", Uptime30d: "100.00", LatencyMs: 4, LastChecked: "2026-04-17T00:00:00Z", Created: "2026-01-10T08:00:00Z", Updated: "2026-04-17T00:00:00Z"},
			{ID: newID(), Name: "cache-redis-01", CheckType: "TCP", Target: "185.22.153.12:6379", Status: "up", Uptime30d: "100.00", LatencyMs: 1, LastChecked: "2026-04-17T00:00:00Z", Created: "2026-01-10T08:00:00Z", Updated: "2026-04-17T00:00:00Z"},
			{ID: newID(), Name: "worker-queue-1", CheckType: "HTTP", Target: "http://worker-01.buyvm.net/status", Status: "degraded", Uptime30d: "98.82", LatencyMs: 0, LastChecked: "2026-04-16T22:00:00Z", Created: "2026-02-01T08:00:00Z", Updated: "2026-04-16T22:00:00Z"},
		},
		"abrnoc": {
			{ID: newID(), Name: "node-fra-01", CheckType: "ICMP", Target: "93.115.20.44", Status: "up", Uptime30d: "99.99", LatencyMs: 8, LastChecked: "2026-04-17T00:00:00Z", Created: "2026-01-05T08:00:00Z", Updated: "2026-04-17T00:00:00Z"},
			{ID: newID(), Name: "node-fra-02", CheckType: "ICMP", Target: "93.115.20.45", Status: "paused", Uptime30d: "99.50", LatencyMs: 0, LastChecked: "2026-04-16T20:00:00Z", Created: "2026-01-05T08:00:00Z", Updated: "2026-04-16T20:00:00Z"},
		},
		"hosting": {
			{ID: newID(), Name: "k8s-master-01", CheckType: "HTTP", Target: "https://k8s-api.hosting.io/healthz", Status: "up", Uptime30d: "99.95", LatencyMs: 15, LastChecked: "2026-04-17T00:00:00Z", Created: "2026-02-15T08:00:00Z", Updated: "2026-04-17T00:00:00Z"},
			{ID: newID(), Name: "k8s-worker-01", CheckType: "TCP", Target: "172.16.50.10:10250", Status: "up", Uptime30d: "99.95", LatencyMs: 12, LastChecked: "2026-04-17T00:00:00Z", Created: "2026-02-15T08:00:00Z", Updated: "2026-04-17T00:00:00Z"},
			{ID: newID(), Name: "k8s-worker-02", CheckType: "TCP", Target: "172.16.50.11:10250", Status: "down", Uptime30d: "97.40", LatencyMs: 0, LastChecked: "2026-04-16T22:00:00Z", Created: "2026-02-15T08:00:00Z", Updated: "2026-04-16T22:00:00Z"},
			{ID: newID(), Name: "cache-redis-01", CheckType: "TCP", Target: "172.16.50.50:6379", Status: "up", Uptime30d: "100.00", LatencyMs: 2, LastChecked: "2026-04-17T00:00:00Z", Created: "2026-03-01T08:00:00Z", Updated: "2026-04-17T00:00:00Z"},
		},
	},
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
