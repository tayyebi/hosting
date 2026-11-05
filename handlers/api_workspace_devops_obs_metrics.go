package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace DevOps API — Observability: Metrics & Dashboards
// devops/obs/metrics
// ════════════════════════════════════════════════════════════════

type MetricDashboard struct {
	ID          string
	Name        string
	Type        string
	Source      string
	Panels      int
	RefreshRate string
	Retention   string
	Status      string
	Updated     string
}

type metricDashboardStoreType struct {
	mu         sync.RWMutex
	dashboards map[string][]MetricDashboard
}

var metricDashboardStore = &metricDashboardStoreType{
	dashboards: map[string][]MetricDashboard{},
}

func (s *metricDashboardStoreType) list(slug string) []MetricDashboard {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.dashboards[slug]
	out := make([]MetricDashboard, len(src))
	copy(out, src)
	return out
}

func (s *metricDashboardStoreType) getByID(slug, id string) (MetricDashboard, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, d := range s.dashboards[slug] {
		if d.ID == id {
			return d, true
		}
	}
	return MetricDashboard{}, false
}

func (s *metricDashboardStoreType) add(slug string, d MetricDashboard) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dashboards[slug] = append(s.dashboards[slug], d)
}

func (s *metricDashboardStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.dashboards[slug]
	for i, d := range list {
		if d.ID == id {
			s.dashboards[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiMetricDashboard struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Source      string `json:"source"`
	Panels      int    `json:"panels"`
	RefreshRate string `json:"refresh_rate"`
	Retention   string `json:"retention"`
	Status      string `json:"status"`
	Updated     string `json:"updated"`
}

type apiMetricDashboardList struct {
	Dashboards []apiMetricDashboard `json:"dashboards"`
	Total      int                  `json:"total"`
}

func toAPIMetricDashboard(d MetricDashboard) apiMetricDashboard {
	return apiMetricDashboard{
		ID: d.ID, Name: d.Name, Type: d.Type,
		Source: d.Source, Panels: d.Panels, RefreshRate: d.RefreshRate,
		Retention: d.Retention, Status: d.Status, Updated: d.Updated,
	}
}

func filterMetricDashboards(all []MetricDashboard, search, status string) []MetricDashboard {
	var out []MetricDashboard
	q := strings.ToLower(search)
	for _, d := range all {
		if status != "" && d.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(d.Name + " " + d.Type + " " + d.Source)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, d)
	}
	return out
}

func apiWSDevOpsObsMetricsListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	dashboards := filterMetricDashboards(metricDashboardStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiMetricDashboard, len(dashboards))
	for i, d := range dashboards {
		out[i] = toAPIMetricDashboard(d)
	}
	apiOK(w, apiMetricDashboardList{Dashboards: out, Total: len(out)})
}

type apiCreateMetricDashboardRequest struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Source      string `json:"source"`
	RefreshRate string `json:"refresh_rate"`
	Retention   string `json:"retention"`
}

func apiWSDevOpsObsMetricsCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateMetricDashboardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	typ := strings.TrimSpace(req.Type)
	if typ == "" {
		typ = "System metrics"
	}
	source := strings.TrimSpace(req.Source)
	if source == "" {
		source = "Prometheus"
	}
	refreshRate := strings.TrimSpace(req.RefreshRate)
	if refreshRate == "" {
		refreshRate = "30s"
	}
	retention := strings.TrimSpace(req.Retention)
	if retention == "" {
		retention = "30 days"
	}
	dashboard := MetricDashboard{
		ID: newID(), Name: name, Type: typ,
		Source: source, Panels: 0, RefreshRate: refreshRate,
		Retention: retention, Status: "active", Updated: "2026-04-17",
	}
	metricDashboardStore.add(slug, dashboard)
	apiCreated(w, toAPIMetricDashboard(dashboard))
}

func apiWSDevOpsObsMetricsGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	dashboard, found := metricDashboardStore.getByID(slug, r.PathValue("mid"))
	if !found {
		apiError(w, http.StatusNotFound, "Metric dashboard not found.")
		return
	}
	apiOK(w, toAPIMetricDashboard(dashboard))
}

func apiWSDevOpsObsMetricsDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := metricDashboardStore.getByID(slug, r.PathValue("mid")); !found {
		apiError(w, http.StatusNotFound, "Metric dashboard not found.")
		return
	}
	metricDashboardStore.delete(slug, r.PathValue("mid"))
	apiOK(w, map[string]any{"ok": true, "message": "Metric dashboard deleted."})
}
