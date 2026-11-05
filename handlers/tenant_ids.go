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

// ── IDS Event store ────────────────────────────────────────────

type IDSEvent struct {
	ID        string
	Severity  string // "critical" | "high" | "medium" | "low" | "info"
	EventType string
	SourceIP  string
	Target    string
	Protocol  string
	Action    string
	Status    string // "open" | "reviewing" | "acknowledged" | "mitigated" | "suppressed"
	RuleID    string
	Details   string
	TimeLabel string
	Created   string
	Updated   string
}

type idsEventStoreType struct {
	mu     sync.RWMutex
	events map[string][]IDSEvent
}

var idsEventStore = &idsEventStoreType{
	events: map[string][]IDSEvent{},
}

func (s *idsEventStoreType) list(slug string) []IDSEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.events[slug]
	out := make([]IDSEvent, len(src))
	copy(out, src)
	return out
}

func (s *idsEventStoreType) getByID(slug, id string) (IDSEvent, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, e := range s.events[slug] {
		if e.ID == id {
			return e, true
		}
	}
	return IDSEvent{}, false
}

func (s *idsEventStoreType) add(slug string, e IDSEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events[slug] = append(s.events[slug], e)
}

func (s *idsEventStoreType) updateStatus(slug, id, status string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, e := range s.events[slug] {
		if e.ID == id {
			s.events[slug][i].Status = status
			s.events[slug][i].Updated = time.Now().Format(time.RFC3339)
			return
		}
	}
}

func (s *idsEventStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.events[slug]
	for i, e := range list {
		if e.ID == id {
			s.events[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

func filterIDSEvents(all []IDSEvent, search, severity, status string) []IDSEvent {
	var out []IDSEvent
	q := strings.ToLower(search)
	for _, e := range all {
		if severity != "" && e.Severity != severity {
			continue
		}
		if status != "" && e.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(e.EventType + " " + e.SourceIP + " " + e.Target + " " + e.RuleID)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, e)
	}
	return out
}

func dbIDSEventToLocal(e db.IDSEvent) IDSEvent {
	return IDSEvent{
		ID:        fmt.Sprintf("%d", e.ID),
		Severity:  e.Severity,
		EventType: e.EventType,
		SourceIP:  e.SourceIP,
		Target:    e.Target,
		Protocol:  e.Protocol,
		Action:    e.Action,
		Status:    e.Status,
		RuleID:    e.RuleID,
		Details:   e.Details,
		TimeLabel: e.CreatedAt.Format("15:04:05"),
		Created:   e.CreatedAt.Format(time.RFC3339),
		Updated:   e.UpdatedAt.Format(time.RFC3339),
	}
}

// ── Page data ─────────────────────────────────────────────────

type idsInnerData struct {
	Events         []IDSEvent
	TenantSlug     string
	Search         string
	SeverityFilter string
	StatusFilter   string
}

func renderIDSPage(w http.ResponseWriter, r *http.Request, tenant Tenant, user User, d idsInnerData) {
	funcMap := template.FuncMap{
		"title": func(s string) string {
			if len(s) == 0 {
				return s
			}
			return strings.ToUpper(s[:1]) + s[1:]
		},
	}
	innerTmpl, err := template.New("ids_content.html").Funcs(funcMap).ParseFiles(filepath.Join("templates", "tenant", "ids_content.html"))
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
		ActiveNav: "ids", PageTitle: "Intrusion Detection",
		IsSysAdmin: user.Role == RoleSysAdmin,
		Content:    template.HTML(buf.String()),
	}
	if err := shellTmpl.ExecuteTemplate(w, "tenant-base", data); err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
	}
}

// ── Handlers ─────────────────────────────────────────────────

func idsHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := getSession(r)
	tenant, _ := tenantFromContext(r)
	q := r.URL.Query().Get("q")
	severity := r.URL.Query().Get("severity")
	status := r.URL.Query().Get("status")

	var events []IDSEvent
	if Pools != nil && tenant.DBID > 0 {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		des, err := db.ListIDSEvents(ctx, Pools.Admin, tenant.DBID, q, severity, status)
		if err != nil {
			log.Printf("idsHandler: ListIDSEvents: %v", err)
			events = filterIDSEvents(idsEventStore.list(tenant.Slug), q, severity, status)
		} else {
			events = make([]IDSEvent, len(des))
			for i, de := range des {
				events[i] = dbIDSEventToLocal(de)
			}
		}
	} else {
		events = filterIDSEvents(idsEventStore.list(tenant.Slug), q, severity, status)
	}

	renderIDSPage(w, r, tenant, user, idsInnerData{
		Events:         events,
		TenantSlug:     tenant.Slug,
		Search:         q,
		SeverityFilter: severity,
		StatusFilter:   status,
	})
}

func idsEventStatusHandler(w http.ResponseWriter, r *http.Request) {
	tenant, _ := tenantFromContext(r)
	idStr := r.PathValue("id")
	newStatus := r.FormValue("status")

	if Pools != nil && tenant.DBID > 0 {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err == nil {
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()
			if err := db.UpdateIDSEventStatus(ctx, Pools.Admin, tenant.DBID, id, newStatus); err != nil {
				log.Printf("idsEventStatusHandler: %v", err)
			}
		}
	} else {
		idsEventStore.updateStatus(tenant.Slug, idStr, newStatus)
	}
	http.Redirect(w, r, "/t/"+tenant.Slug+"/ids/", http.StatusSeeOther)
}

func idsEventDeleteHandler(w http.ResponseWriter, r *http.Request) {
	tenant, _ := tenantFromContext(r)
	idStr := r.PathValue("id")

	if Pools != nil && tenant.DBID > 0 {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err == nil {
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()
			if err := db.DeleteIDSEvent(ctx, Pools.Admin, tenant.DBID, id); err != nil {
				log.Printf("idsEventDeleteHandler: %v", err)
			}
		}
	} else {
		idsEventStore.delete(tenant.Slug, idStr)
	}
	http.Redirect(w, r, "/t/"+tenant.Slug+"/ids/", http.StatusSeeOther)
}

func idsBlockIPHandler(w http.ResponseWriter, r *http.Request) {
	tenant, _ := tenantFromContext(r)
	ip := r.FormValue("ip")
	if ip != "" {
		log.Printf("idsBlockIPHandler: tenant=%s blocking IP=%s duration=%s", tenant.Slug, ip, r.FormValue("duration"))
	}
	http.Redirect(w, r, "/t/"+tenant.Slug+"/ids/", http.StatusSeeOther)
}

func RegisterTenantIDSRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /t/{slug}/ids/", tenantMiddleware(idsHandler))
	mux.HandleFunc("POST /t/{slug}/ids/{id}/status/", tenantMiddleware(idsEventStatusHandler))
	mux.HandleFunc("POST /t/{slug}/ids/{id}/delete/", tenantMiddleware(idsEventDeleteHandler))
	mux.HandleFunc("POST /t/{slug}/ids/block", tenantMiddleware(idsBlockIPHandler))
}
