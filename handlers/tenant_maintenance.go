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
	windows: map[string][]MaintenanceWin{},
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
