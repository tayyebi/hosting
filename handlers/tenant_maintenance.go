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

// ── Maintenance Window store ───────────────────────────────────

type MaintenanceWin struct {
	ID             string
	RefNum         string
	Title          string
	Description    string
	Affected       string
	ScheduledStart string
	ScheduledEnd   string
	WindowLabel    string // e.g. "Nov 17 · 02:00 – 04:00"
	Risk           string // "low" | "medium" | "high" | "critical"
	Status         string // "scheduled" | "in-progress" | "completed" | "cancelled"
	Rollback       string
	Created        string
	Updated        string
}

type maintenanceStoreType struct {
	mu      sync.RWMutex
	windows map[string][]MaintenanceWin
}

var maintenanceStore = &maintenanceStoreType{
	windows: map[string][]MaintenanceWin{
		"buyvm": {
			{ID: newID(), RefNum: "CHG-0147", Title: "PostgreSQL 14 → 16 major upgrade", Description: "blue-green · logical replication cutover · approved by CAB 2026-04-13", Affected: "pg-primary-fra1 · pg-replica-fra1a", WindowLabel: "Nov 17 · 02:00 – 04:00", Risk: "high", Status: "scheduled", Rollback: "Promote old primary (≤ 8 min)", Created: "2026-04-13T00:00:00Z", Updated: "2026-04-13T00:00:00Z"},
			{ID: newID(), RefNum: "CHG-0149", Title: "CDN edge fleet rotation (32 PoPs)", Description: "canary → 10% → 50% → 100% · auto-rollback on 5xx > 0.5%", Affected: "edge-* (global)", WindowLabel: "Nov 20 · 00:00 – 01:00", Risk: "low", Status: "scheduled", Rollback: "Pin previous AMI (instant)", Created: "2026-04-14T00:00:00Z", Updated: "2026-04-14T00:00:00Z"},
			{ID: newID(), RefNum: "CHG-0151", Title: "API rate-limit config rollout (v2 token bucket)", Description: "config-only · feature-flagged · no restart required", Affected: "api-gateway · 14 pods", WindowLabel: "Nov 10 · 03:00 – 03:30", Risk: "low", Status: "completed", Rollback: "Flip flag to v1 (≤ 30s)", Created: "2026-04-10T03:00:00Z", Updated: "2026-04-10T03:32:00Z"},
			{ID: newID(), RefNum: "CHG-0144", Title: "TLS cert renewal (Let's Encrypt, 12 SANs)", Description: "zero-downtime reload · cert-manager automated", Affected: "dashboard.* · api.*", WindowLabel: "Nov 05 · 01:00 – 01:15", Risk: "low", Status: "completed", Rollback: "Previous cert pinned (auto)", Created: "2026-04-05T01:00:00Z", Updated: "2026-04-05T01:12:00Z"},
			{ID: newID(), RefNum: "CHG-0138", Title: "Ceph OSD rebalance (add 6 × 7.68 TB NVMe)", Description: "Cancelled — held for pg 14→16 completion, then re-plan", Affected: "ceph-osd-* (fra1)", WindowLabel: "Oct 28 · 02:00 – 05:00", Risk: "medium", Status: "cancelled", Rollback: "N/A (cancelled)", Created: "2026-04-01T00:00:00Z", Updated: "2026-04-08T00:00:00Z"},
		},
		"abrnoc": {
			{ID: newID(), RefNum: "CHG-0022", Title: "Network switch firmware upgrade", Description: "Staggered per-rack, < 5s failover per switch", Affected: "core-sw-01, core-sw-02", WindowLabel: "Nov 12 · 01:00 – 03:00", Risk: "medium", Status: "scheduled", Rollback: "Revert to previous firmware", Created: "2026-04-12T00:00:00Z", Updated: "2026-04-12T00:00:00Z"},
		},
		"hosting": {
			{ID: newID(), RefNum: "CHG-0088", Title: "Kubernetes control plane upgrade 1.28 → 1.30", Description: "Rolling upgrade, drain workers before upgrade", Affected: "k8s-master-01, k8s-worker-01", WindowLabel: "Nov 15 · 02:00 – 05:00", Risk: "high", Status: "scheduled", Rollback: "Downgrade via etcd snapshot", Created: "2026-04-15T00:00:00Z", Updated: "2026-04-15T00:00:00Z"},
		},
	},
}

func (s *maintenanceStoreType) list(slug string) []MaintenanceWin {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.windows[slug]
	out := make([]MaintenanceWin, len(src))
	copy(out, src)
	return out
}

func (s *maintenanceStoreType) getByID(slug, id string) (MaintenanceWin, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, w := range s.windows[slug] {
		if w.ID == id {
			return w, true
		}
	}
	return MaintenanceWin{}, false
}

func (s *maintenanceStoreType) add(slug string, w MaintenanceWin) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.windows[slug] = append(s.windows[slug], w)
}

func (s *maintenanceStoreType) updateStatus(slug, id, status string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, w := range s.windows[slug] {
		if w.ID == id {
			s.windows[slug][i].Status = status
			s.windows[slug][i].Updated = time.Now().Format(time.RFC3339)
			return
		}
	}
}

func (s *maintenanceStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.windows[slug]
	for i, w := range list {
		if w.ID == id {
			s.windows[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

func filterMaintenanceWindows(all []MaintenanceWin, search, status string) []MaintenanceWin {
	var out []MaintenanceWin
	q := strings.ToLower(search)
	for _, w := range all {
		if status != "" && w.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(w.Title + " " + w.RefNum + " " + w.Affected)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, w)
	}
	return out
}

func dbMaintenanceWindowToLocal(w db.MaintenanceWindow) MaintenanceWin {
	windowLabel := fmt.Sprintf("%s – %s",
		w.ScheduledStart.Format("Jan 02 · 15:04"),
		w.ScheduledEnd.Format("15:04"))
	return MaintenanceWin{
		ID:             fmt.Sprintf("%d", w.ID),
		RefNum:         w.RefNum,
		Title:          w.Title,
		Description:    w.Description,
		Affected:       w.Affected,
		ScheduledStart: w.ScheduledStart.Format(time.RFC3339),
		ScheduledEnd:   w.ScheduledEnd.Format(time.RFC3339),
		WindowLabel:    windowLabel,
		Risk:           w.Risk,
		Status:         w.Status,
		Rollback:       w.Rollback,
		Created:        w.CreatedAt.Format(time.RFC3339),
		Updated:        w.UpdatedAt.Format(time.RFC3339),
	}
}

// ── Page data ─────────────────────────────────────────────────

type maintenanceInnerData struct {
	Windows      []MaintenanceWin
	TenantSlug   string
	Search       string
	StatusFilter string
}

func renderMaintenancePage(w http.ResponseWriter, r *http.Request, tenant Tenant, user User, d maintenanceInnerData) {
	innerTmpl, err := template.ParseFiles(filepath.Join("templates", "tenant", "maintenance_content.html"))
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
		ActiveNav: "maintenance", PageTitle: "Maintenance",
		IsSysAdmin: user.Role == RoleSysAdmin,
		Content:    template.HTML(buf.String()),
	}
	if err := shellTmpl.ExecuteTemplate(w, "tenant-base", data); err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
	}
}

// ── Handlers ─────────────────────────────────────────────────

func maintenanceHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := getSession(r)
	tenant, _ := tenantFromContext(r)
	q := r.URL.Query().Get("q")
	status := r.URL.Query().Get("status")

	var windows []MaintenanceWin
	if Pools != nil && tenant.DBID > 0 {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		dws, err := db.ListMaintenanceWindows(ctx, Pools.Admin, tenant.DBID, q, status)
		if err != nil {
			log.Printf("maintenanceHandler: ListMaintenanceWindows: %v", err)
			windows = filterMaintenanceWindows(maintenanceStore.list(tenant.Slug), q, status)
		} else {
			windows = make([]MaintenanceWin, len(dws))
			for i, dw := range dws {
				windows[i] = dbMaintenanceWindowToLocal(dw)
			}
		}
	} else {
		windows = filterMaintenanceWindows(maintenanceStore.list(tenant.Slug), q, status)
	}

	renderMaintenancePage(w, r, tenant, user, maintenanceInnerData{
		Windows:      windows,
		TenantSlug:   tenant.Slug,
		Search:       q,
		StatusFilter: status,
	})
}

func maintenanceWindowStatusHandler(w http.ResponseWriter, r *http.Request) {
	tenant, _ := tenantFromContext(r)
	idStr := r.PathValue("id")
	newStatus := r.FormValue("status")

	if Pools != nil && tenant.DBID > 0 {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err == nil {
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()
			if err := db.UpdateMaintenanceWindowStatus(ctx, Pools.Admin, tenant.DBID, id, newStatus); err != nil {
				log.Printf("maintenanceWindowStatusHandler: %v", err)
			}
		}
	} else {
		maintenanceStore.updateStatus(tenant.Slug, idStr, newStatus)
	}
	http.Redirect(w, r, "/t/"+tenant.Slug+"/maintenance/", http.StatusSeeOther)
}

func maintenanceWindowDeleteHandler(w http.ResponseWriter, r *http.Request) {
	tenant, _ := tenantFromContext(r)
	idStr := r.PathValue("id")

	if Pools != nil && tenant.DBID > 0 {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err == nil {
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()
			if err := db.DeleteMaintenanceWindow(ctx, Pools.Admin, tenant.DBID, id); err != nil {
				log.Printf("maintenanceWindowDeleteHandler: %v", err)
			}
		}
	} else {
		maintenanceStore.delete(tenant.Slug, idStr)
	}
	http.Redirect(w, r, "/t/"+tenant.Slug+"/maintenance/", http.StatusSeeOther)
}

func RegisterTenantMaintenanceRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /t/{slug}/maintenance/", tenantMiddleware(maintenanceHandler))
	mux.HandleFunc("POST /t/{slug}/maintenance/{id}/status/", tenantMiddleware(maintenanceWindowStatusHandler))
	mux.HandleFunc("POST /t/{slug}/maintenance/{id}/delete/", tenantMiddleware(maintenanceWindowDeleteHandler))
}
