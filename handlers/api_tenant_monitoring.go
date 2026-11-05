package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ── Monitoring check store ──────────────────────────────────

type MonitoringCheck struct {
	ID       string
	Name     string
	Type     string // "http" | "tcp" | "icmp" | "dns"
	Endpoint string
	Status   string // "up" | "degraded" | "down"
	Latency  string
	Interval string
	Region   string
	Created  string
	Updated  string
}

type monitoringCheckStoreType struct {
	mu     sync.RWMutex
	checks map[string][]MonitoringCheck
}

var monitoringCheckStore = &monitoringCheckStoreType{
	checks: map[string][]MonitoringCheck{},
}

func (s *monitoringCheckStoreType) list(slug string) []MonitoringCheck {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.checks[slug]
	out := make([]MonitoringCheck, len(src))
	copy(out, src)
	return out
}

func (s *monitoringCheckStoreType) getByID(slug, id string) (MonitoringCheck, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, c := range s.checks[slug] {
		if c.ID == id {
			return c, true
		}
	}
	return MonitoringCheck{}, false
}

func (s *monitoringCheckStoreType) add(slug string, c MonitoringCheck) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.checks[slug] = append(s.checks[slug], c)
}

func (s *monitoringCheckStoreType) delete(slug, id string) {
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

func filterMonitoringChecks(all []MonitoringCheck, search, status, checkType string) []MonitoringCheck {
	var out []MonitoringCheck
	q := strings.ToLower(search)
	for _, c := range all {
		if status != "" && c.Status != status {
			continue
		}
		if checkType != "" && c.Type != checkType {
			continue
		}
		if q != "" {
			hay := strings.ToLower(c.Name + " " + c.Endpoint + " " + c.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, c)
	}
	return out
}

// ── GET /api/tenant/{slug}/monitoring ───────────────────────

type apiMonitoringCheck struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Endpoint string `json:"endpoint"`
	Status   string `json:"status"`
	Latency  string `json:"latency"`
	Interval string `json:"interval"`
	Region   string `json:"region"`
	Created  string `json:"created"`
	Updated  string `json:"updated"`
}

type apiMonitoringListResponse struct {
	Checks []apiMonitoringCheck `json:"checks"`
	Total  int                  `json:"total"`
}

func apiTenantMonitoringListHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok {
		apiError(w, http.StatusUnauthorized, "Authentication required.")
		return
	}

	slug := r.PathValue("slug")
	if _, found := store.getBySlug(slug); !found {
		apiError(w, http.StatusNotFound, "Tenant not found.")
		return
	}
	if user.Role != RoleSysAdmin && (user.Role != RoleTenant || user.TenantSlug != slug) {
		apiError(w, http.StatusForbidden, "Access denied.")
		return
	}

	search := r.URL.Query().Get("q")
	status := r.URL.Query().Get("status")
	checkType := r.URL.Query().Get("type")

	checks := filterMonitoringChecks(monitoringCheckStore.list(slug), search, status, checkType)

	out := make([]apiMonitoringCheck, len(checks))
	for i, c := range checks {
		out[i] = apiMonitoringCheck{
			ID: c.ID, Name: c.Name, Type: c.Type,
			Endpoint: c.Endpoint, Status: c.Status,
			Latency: c.Latency, Interval: c.Interval,
			Region:  c.Region,
			Created: c.Created, Updated: c.Updated,
		}
	}

	apiOK(w, apiMonitoringListResponse{Checks: out, Total: len(out)})
}

// ── POST /api/tenant/{slug}/monitoring ──────────────────────

type apiCreateMonitoringCheckRequest struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Endpoint string `json:"endpoint"`
	Interval string `json:"interval"`
	Region   string `json:"region"`
}

func apiTenantMonitoringCreateHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok {
		apiError(w, http.StatusUnauthorized, "Authentication required.")
		return
	}

	slug := r.PathValue("slug")
	if _, found := store.getBySlug(slug); !found {
		apiError(w, http.StatusNotFound, "Tenant not found.")
		return
	}
	if user.Role != RoleSysAdmin && (user.Role != RoleTenant || user.TenantSlug != slug) {
		apiError(w, http.StatusForbidden, "Access denied.")
		return
	}

	var req apiCreateMonitoringCheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}

	name := strings.TrimSpace(req.Name)
	endpoint := strings.TrimSpace(req.Endpoint)
	checkType := strings.TrimSpace(req.Type)
	interval := strings.TrimSpace(req.Interval)
	region := strings.TrimSpace(req.Region)

	if name == "" || endpoint == "" {
		apiError(w, http.StatusBadRequest, "Name and endpoint are required.")
		return
	}
	if checkType == "" {
		checkType = "http"
	}
	if interval == "" {
		interval = "60s"
	}
	if region == "" {
		region = "us-east"
	}

	c := MonitoringCheck{
		ID: newID(), Name: name, Type: checkType,
		Endpoint: endpoint, Status: "up",
		Latency: "—", Interval: interval, Region: region,
		Created: "2026-04-17T00:00:00Z", Updated: "2026-04-17T00:00:00Z",
	}
	monitoringCheckStore.add(slug, c)

	apiCreated(w, apiMonitoringCheck{
		ID: c.ID, Name: c.Name, Type: c.Type,
		Endpoint: c.Endpoint, Status: c.Status,
		Latency: c.Latency, Interval: c.Interval,
		Region:  c.Region,
		Created: c.Created, Updated: c.Updated,
	})
}

// ── GET /api/tenant/{slug}/monitoring/{id} ──────────────────

func apiTenantMonitoringGetHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok {
		apiError(w, http.StatusUnauthorized, "Authentication required.")
		return
	}

	slug := r.PathValue("slug")
	if _, found := store.getBySlug(slug); !found {
		apiError(w, http.StatusNotFound, "Tenant not found.")
		return
	}
	if user.Role != RoleSysAdmin && (user.Role != RoleTenant || user.TenantSlug != slug) {
		apiError(w, http.StatusForbidden, "Access denied.")
		return
	}

	id := r.PathValue("id")
	c, ok := monitoringCheckStore.getByID(slug, id)
	if !ok {
		apiError(w, http.StatusNotFound, "Monitoring check not found.")
		return
	}

	apiOK(w, apiMonitoringCheck{
		ID: c.ID, Name: c.Name, Type: c.Type,
		Endpoint: c.Endpoint, Status: c.Status,
		Latency: c.Latency, Interval: c.Interval,
		Region:  c.Region,
		Created: c.Created, Updated: c.Updated,
	})
}

// ── DELETE /api/tenant/{slug}/monitoring/{id} ───────────────

func apiTenantMonitoringDeleteHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok {
		apiError(w, http.StatusUnauthorized, "Authentication required.")
		return
	}

	slug := r.PathValue("slug")
	if _, found := store.getBySlug(slug); !found {
		apiError(w, http.StatusNotFound, "Tenant not found.")
		return
	}
	if user.Role != RoleSysAdmin && (user.Role != RoleTenant || user.TenantSlug != slug) {
		apiError(w, http.StatusForbidden, "Access denied.")
		return
	}

	id := r.PathValue("id")
	if _, ok := monitoringCheckStore.getByID(slug, id); !ok {
		apiError(w, http.StatusNotFound, "Monitoring check not found.")
		return
	}

	monitoringCheckStore.delete(slug, id)
	apiOK(w, map[string]any{
		"ok":      true,
		"message": "Monitoring check deleted.",
	})
}
