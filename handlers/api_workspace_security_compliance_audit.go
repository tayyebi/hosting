package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Security API — Compliance: Audit Logging
// security/compliance/audit
// ════════════════════════════════════════════════════════════════

type AuditLog struct {
	ID        string
	Name      string
	Source    string
	Retention string
	Format    string
	Status    string
	Events7d  string
	Updated   string
}

type auditLogStoreType struct {
	mu   sync.RWMutex
	logs map[string][]AuditLog
}

var auditLogStore = &auditLogStoreType{
	logs: map[string][]AuditLog{
		"buyvm": {
			{ID: newID(), Name: "API Access Log", Source: "API Gateway", Retention: "90 days", Format: "JSON", Status: "active", Events7d: "1.2M", Updated: "2026-01-15"},
			{ID: newID(), Name: "IAM Activity Trail", Source: "IAM Service", Retention: "365 days", Format: "JSON", Status: "active", Events7d: "84k", Updated: "2026-01-20"},
			{ID: newID(), Name: "Network Flow Logs", Source: "VPC", Retention: "30 days", Format: "CEF", Status: "active", Events7d: "18.5M", Updated: "2026-02-01"},
			{ID: newID(), Name: "Database Audit Trail", Source: "RDS / Managed DB", Retention: "180 days", Format: "JSON", Status: "active", Events7d: "456k", Updated: "2025-12-10"},
			{ID: newID(), Name: "Console Login Events", Source: "Auth Service", Retention: "365 days", Format: "JSON", Status: "active", Events7d: "1.2k", Updated: "2026-01-25"},
		},
		"hosting": {
			{ID: newID(), Name: "API Access Log", Source: "API Gateway", Retention: "60 days", Format: "JSON", Status: "active", Events7d: "890k", Updated: "2026-01-18"},
			{ID: newID(), Name: "Admin Activity Trail", Source: "IAM Service", Retention: "365 days", Format: "JSON", Status: "active", Events7d: "12k", Updated: "2026-02-05"},
			{ID: newID(), Name: "Storage Access Logs", Source: "Object Storage", Retention: "90 days", Format: "JSON", Status: "active", Events7d: "2.4M", Updated: "2026-01-30"},
		},
	},
}

func (s *auditLogStoreType) list(slug string) []AuditLog {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.logs[slug]
	out := make([]AuditLog, len(src))
	copy(out, src)
	return out
}

func (s *auditLogStoreType) getByID(slug, id string) (AuditLog, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, l := range s.logs[slug] {
		if l.ID == id {
			return l, true
		}
	}
	return AuditLog{}, false
}

func (s *auditLogStoreType) add(slug string, l AuditLog) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logs[slug] = append(s.logs[slug], l)
}

func (s *auditLogStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.logs[slug]
	for i, l := range list {
		if l.ID == id {
			s.logs[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiAuditLog struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Source    string `json:"source"`
	Retention string `json:"retention"`
	Format    string `json:"format"`
	Status    string `json:"status"`
	Events7d  string `json:"events_7d"`
	Updated   string `json:"updated"`
}

type apiAuditLogList struct {
	Logs  []apiAuditLog `json:"logs"`
	Total int           `json:"total"`
}

func toAPIAuditLog(l AuditLog) apiAuditLog {
	return apiAuditLog{
		ID: l.ID, Name: l.Name, Source: l.Source,
		Retention: l.Retention, Format: l.Format, Status: l.Status,
		Events7d: l.Events7d, Updated: l.Updated,
	}
}

func filterAuditLogs(all []AuditLog, search, status string) []AuditLog {
	var out []AuditLog
	q := strings.ToLower(search)
	for _, l := range all {
		if status != "" && l.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(l.Name + " " + l.Source + " " + l.Format)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, l)
	}
	return out
}

func apiWSSecCompAuditListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	logs := filterAuditLogs(auditLogStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiAuditLog, len(logs))
	for i, l := range logs {
		out[i] = toAPIAuditLog(l)
	}
	apiOK(w, apiAuditLogList{Logs: out, Total: len(out)})
}

type apiCreateAuditLogRequest struct {
	Name      string `json:"name"`
	Source    string `json:"source"`
	Retention string `json:"retention"`
	Format    string `json:"format"`
}

func apiWSSecCompAuditCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateAuditLogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	source := strings.TrimSpace(req.Source)
	if source == "" {
		source = "Custom"
	}
	retention := strings.TrimSpace(req.Retention)
	if retention == "" {
		retention = "90 days"
	}
	format := strings.TrimSpace(req.Format)
	if format == "" {
		format = "JSON"
	}
	logEntry := AuditLog{
		ID: newID(), Name: name, Source: source,
		Retention: retention, Format: format, Status: "active",
		Events7d: "0", Updated: "2026-04-17",
	}
	auditLogStore.add(slug, logEntry)
	apiCreated(w, toAPIAuditLog(logEntry))
}

func apiWSSecCompAuditGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	logEntry, found := auditLogStore.getByID(slug, r.PathValue("aid"))
	if !found {
		apiError(w, http.StatusNotFound, "Audit log not found.")
		return
	}
	apiOK(w, toAPIAuditLog(logEntry))
}

func apiWSSecCompAuditDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := auditLogStore.getByID(slug, r.PathValue("aid")); !found {
		apiError(w, http.StatusNotFound, "Audit log not found.")
		return
	}
	auditLogStore.delete(slug, r.PathValue("aid"))
	apiOK(w, map[string]any{"ok": true, "message": "Audit log deleted."})
}
