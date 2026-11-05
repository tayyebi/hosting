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
	checks: map[string][]MonitoringCheck{
		"buyvm": {
			{ID: newID(), Name: "buyvm.io", Type: "http", Endpoint: "https://buyvm.io", Status: "up", Latency: "34ms", Interval: "30s", Region: "us-east", Created: "2026-01-15T08:00:00Z", Updated: "2026-04-17T14:38:00Z"},
			{ID: newID(), Name: "api.buyvm.io/health", Type: "http", Endpoint: "https://api.buyvm.io/health", Status: "up", Latency: "22ms", Interval: "30s", Region: "us-east", Created: "2026-01-15T08:00:00Z", Updated: "2026-04-17T14:38:00Z"},
			{ID: newID(), Name: "api.buyvm.io/graphql", Type: "http", Endpoint: "https://api.buyvm.io/graphql", Status: "degraded", Latency: "148ms", Interval: "30s", Region: "us-east", Created: "2026-02-10T10:00:00Z", Updated: "2026-04-17T14:38:00Z"},
			{ID: newID(), Name: "auth.buyvm.io", Type: "http", Endpoint: "https://auth.buyvm.io", Status: "up", Latency: "41ms", Interval: "60s", Region: "eu-west", Created: "2026-01-20T09:00:00Z", Updated: "2026-04-17T14:38:00Z"},
			{ID: newID(), Name: "cdn.buyvm.io", Type: "http", Endpoint: "https://cdn.buyvm.io", Status: "up", Latency: "11ms", Interval: "30s", Region: "global", Created: "2026-03-01T12:00:00Z", Updated: "2026-04-17T14:38:00Z"},
			{ID: newID(), Name: "db-pg-prod:5432", Type: "tcp", Endpoint: "tcp://db-pg-prod.buyvm.io:5432", Status: "up", Latency: "4ms", Interval: "15s", Region: "us-east", Created: "2025-12-01T08:00:00Z", Updated: "2026-04-17T14:38:00Z"},
		},
		"abrnoc": {
			{ID: newID(), Name: "abrnoc.io", Type: "http", Endpoint: "https://abrnoc.io", Status: "up", Latency: "28ms", Interval: "30s", Region: "eu-west", Created: "2026-02-01T10:00:00Z", Updated: "2026-04-17T14:38:00Z"},
			{ID: newID(), Name: "api.abrnoc.io/health", Type: "http", Endpoint: "https://api.abrnoc.io/health", Status: "up", Latency: "19ms", Interval: "30s", Region: "eu-west", Created: "2026-02-01T10:00:00Z", Updated: "2026-04-17T14:38:00Z"},
		},
		"hosting": {
			{ID: newID(), Name: "hosting.com", Type: "http", Endpoint: "https://hosting.com", Status: "up", Latency: "42ms", Interval: "60s", Region: "us-east", Created: "2026-03-05T14:00:00Z", Updated: "2026-04-17T14:38:00Z"},
			{ID: newID(), Name: "cdn-edge-fra", Type: "icmp", Endpoint: "icmp://cdn-edge-fra.hosting.com", Status: "up", Latency: "8ms", Interval: "10s", Region: "eu-west", Created: "2026-03-10T08:00:00Z", Updated: "2026-04-17T14:38:00Z"},
			{ID: newID(), Name: "dns.hosting.com", Type: "dns", Endpoint: "dns://ns1.hosting.com", Status: "up", Latency: "6ms", Interval: "60s", Region: "global", Created: "2026-01-01T00:00:00Z", Updated: "2026-04-17T14:38:00Z"},
		},
	},
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
			Region: c.Region,
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
		Region: c.Region,
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
		Region: c.Region,
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
