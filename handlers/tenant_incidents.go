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

// ── Incident store ────────────────────────────────────────────

type TenantIncident struct {
	ID          string
	RefNum      string
	Title       string
	Severity    string // "critical" | "major" | "minor" | "low"
	Status      string // "investigating" | "identified" | "monitoring" | "resolved"
	Affected    string
	Description string
	StartedAt   string
	ResolvedAt  string
	Duration    string
	IsResolved  bool
	Created     string
	Updated     string
}

type incidentStoreType struct {
	mu        sync.RWMutex
	incidents map[string][]TenantIncident
}

var incidentStore = &incidentStoreType{
	incidents: map[string][]TenantIncident{},
}

func (s *incidentStoreType) list(slug string) []TenantIncident {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.incidents[slug]
	out := make([]TenantIncident, len(src))
	copy(out, src)
	return out
}

func (s *incidentStoreType) getByID(slug, id string) (TenantIncident, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, inc := range s.incidents[slug] {
		if inc.ID == id {
			return inc, true
		}
	}
	return TenantIncident{}, false
}

func (s *incidentStoreType) add(slug string, inc TenantIncident) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.incidents[slug] = append(s.incidents[slug], inc)
}

func (s *incidentStoreType) updateStatus(slug, id, status string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, inc := range s.incidents[slug] {
		if inc.ID == id {
			s.incidents[slug][i].Status = status
			s.incidents[slug][i].Updated = time.Now().Format(time.RFC3339)
			if status == "resolved" {
				s.incidents[slug][i].IsResolved = true
				s.incidents[slug][i].ResolvedAt = time.Now().Format(time.RFC3339)
			}
			return
		}
	}
}

func (s *incidentStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.incidents[slug]
	for i, inc := range list {
		if inc.ID == id {
			s.incidents[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

func filterIncidents(all []TenantIncident, search, severity, status string) []TenantIncident {
	var out []TenantIncident
	q := strings.ToLower(search)
	for _, inc := range all {
		if severity != "" && inc.Severity != severity {
			continue
		}
		if status != "" && inc.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(inc.Title + " " + inc.RefNum + " " + inc.Affected)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, inc)
	}
	return out
}

func dbIncidentToLocal(i db.Incident) TenantIncident {
	resolved := ""
	isResolved := false
	if i.ResolvedAt != nil {
		resolved = i.ResolvedAt.Format(time.RFC3339)
		isResolved = true
	}
	duration := "active"
	if isResolved && i.ResolvedAt != nil {
		d := i.ResolvedAt.Sub(i.StartedAt)
		if d.Hours() >= 1 {
			duration = fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
		} else {
			duration = fmt.Sprintf("%d min", int(d.Minutes()))
		}
	}
	return TenantIncident{
		ID:          fmt.Sprintf("%d", i.ID),
		RefNum:      i.RefNum,
		Title:       i.RefNum + " — " + i.Title,
		Severity:    i.Severity,
		Status:      i.Status,
		Affected:    i.Affected,
		Description: i.Description,
		StartedAt:   i.StartedAt.Format(time.RFC3339),
		ResolvedAt:  resolved,
		Duration:    duration,
		IsResolved:  isResolved,
		Created:     i.CreatedAt.Format(time.RFC3339),
		Updated:     i.UpdatedAt.Format(time.RFC3339),
	}
}

// ── Page data ─────────────────────────────────────────────────

type incidentsInnerData struct {
	Incidents      []TenantIncident
	TenantSlug     string
	Search         string
	SeverityFilter string
	StatusFilter   string
}

func renderIncidentsPage(w http.ResponseWriter, r *http.Request, tenant Tenant, user User, d incidentsInnerData) {
	innerTmpl, err := template.ParseFiles(filepath.Join("templates", "tenant", "incidents_content.html"))
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
		ActiveNav: "incidents", PageTitle: "Incidents",
		IsSysAdmin: user.Role == RoleSysAdmin,
		Content:    template.HTML(buf.String()),
	}
	if err := shellTmpl.ExecuteTemplate(w, "tenant-base", data); err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
	}
}

// ── Handlers ─────────────────────────────────────────────────

func incidentsHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := getSession(r)
	tenant, _ := tenantFromContext(r)
	q := r.URL.Query().Get("q")
	severity := r.URL.Query().Get("severity")
	status := r.URL.Query().Get("status")

	var incidents []TenantIncident
	if Pools != nil && tenant.DBID > 0 {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		dis, err := db.ListIncidents(ctx, Pools.Admin, tenant.DBID, q, severity, status)
		if err != nil {
			log.Printf("incidentsHandler: ListIncidents: %v", err)
			incidents = filterIncidents(incidentStore.list(tenant.Slug), q, severity, status)
		} else {
			incidents = make([]TenantIncident, len(dis))
			for i, di := range dis {
				incidents[i] = dbIncidentToLocal(di)
			}
		}
	} else {
		incidents = filterIncidents(incidentStore.list(tenant.Slug), q, severity, status)
	}

	renderIncidentsPage(w, r, tenant, user, incidentsInnerData{
		Incidents:      incidents,
		TenantSlug:     tenant.Slug,
		Search:         q,
		SeverityFilter: severity,
		StatusFilter:   status,
	})
}

func incidentStatusHandler(w http.ResponseWriter, r *http.Request) {
	tenant, _ := tenantFromContext(r)
	idStr := r.PathValue("id")
	newStatus := r.FormValue("status")

	if Pools != nil && tenant.DBID > 0 {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err == nil {
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()
			if err := db.UpdateIncidentStatus(ctx, Pools.Admin, tenant.DBID, id, newStatus); err != nil {
				log.Printf("incidentStatusHandler: %v", err)
			}
		}
	} else {
		incidentStore.updateStatus(tenant.Slug, idStr, newStatus)
	}
	http.Redirect(w, r, "/t/"+tenant.Slug+"/incidents/", http.StatusSeeOther)
}

func incidentDeleteHandler(w http.ResponseWriter, r *http.Request) {
	tenant, _ := tenantFromContext(r)
	idStr := r.PathValue("id")

	if Pools != nil && tenant.DBID > 0 {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err == nil {
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()
			if err := db.DeleteIncident(ctx, Pools.Admin, tenant.DBID, id); err != nil {
				log.Printf("incidentDeleteHandler: %v", err)
			}
		}
	} else {
		incidentStore.delete(tenant.Slug, idStr)
	}
	http.Redirect(w, r, "/t/"+tenant.Slug+"/incidents/", http.StatusSeeOther)
}

func RegisterTenantIncidentsRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /t/{slug}/incidents/", tenantMiddleware(incidentsHandler))
	mux.HandleFunc("POST /t/{slug}/incidents/{id}/status/", tenantMiddleware(incidentStatusHandler))
	mux.HandleFunc("POST /t/{slug}/incidents/{id}/delete/", tenantMiddleware(incidentDeleteHandler))
}
