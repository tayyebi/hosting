package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace DevOps API — Monitoring: Metrics
// devops/monitoring/metrics
// ════════════════════════════════════════════════════════════════

type MonitoringAlert struct {
	ID             string
	Name           string
	Metric         string
	Condition      string
	Threshold      float64
	Window         string
	Severity       string
	NotifyChannels string
	EvalInterval   string
	State          string
	LastTriggered  string
	Silenced       bool
	Status         string
	Updated        string
}

type monitoringAlertStoreType struct {
	mu    sync.RWMutex
	items map[string][]MonitoringAlert
}

var monitoringAlertStore = &monitoringAlertStoreType{
	items: map[string][]MonitoringAlert{},
}

func (s *monitoringAlertStoreType) list(slug string) []MonitoringAlert {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.items[slug]
	out := make([]MonitoringAlert, len(src))
	copy(out, src)
	return out
}

func (s *monitoringAlertStoreType) getByID(slug, id string) (MonitoringAlert, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, a := range s.items[slug] {
		if a.ID == id {
			return a, true
		}
	}
	return MonitoringAlert{}, false
}

func (s *monitoringAlertStoreType) add(slug string, a MonitoringAlert) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[slug] = append(s.items[slug], a)
}

func (s *monitoringAlertStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.items[slug]
	for i, a := range list {
		if a.ID == id {
			s.items[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiMonitoringAlert struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Metric         string  `json:"metric"`
	Condition      string  `json:"condition"`
	Threshold      float64 `json:"threshold"`
	Window         string  `json:"window"`
	Severity       string  `json:"severity"`
	NotifyChannels string  `json:"notify_channels"`
	EvalInterval   string  `json:"eval_interval"`
	State          string  `json:"state"`
	LastTriggered  string  `json:"last_triggered"`
	Silenced       bool    `json:"silenced"`
	Status         string  `json:"status"`
	Updated        string  `json:"updated"`
}

type apiMonitoringAlertList struct {
	Alerts []apiMonitoringAlert `json:"alerts"`
	Total  int                  `json:"total"`
}

func toAPIMonitoringAlert(a MonitoringAlert) apiMonitoringAlert {
	return apiMonitoringAlert{
		ID: a.ID, Name: a.Name, Metric: a.Metric,
		Condition: a.Condition, Threshold: a.Threshold,
		Window: a.Window, Severity: a.Severity,
		NotifyChannels: a.NotifyChannels, EvalInterval: a.EvalInterval,
		State: a.State, LastTriggered: a.LastTriggered,
		Silenced: a.Silenced, Status: a.Status, Updated: a.Updated,
	}
}

func filterMonitoringAlerts(all []MonitoringAlert, search, status string) []MonitoringAlert {
	var out []MonitoringAlert
	q := strings.ToLower(search)
	for _, a := range all {
		if status != "" && a.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(a.Name + " " + a.Metric + " " + a.Severity + " " + a.State)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, a)
	}
	return out
}

func apiWSDevOpsMonMetricsListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	items := filterMonitoringAlerts(monitoringAlertStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiMonitoringAlert, len(items))
	for i, a := range items {
		out[i] = toAPIMonitoringAlert(a)
	}
	apiOK(w, apiMonitoringAlertList{Alerts: out, Total: len(out)})
}

type apiCreateMonitoringAlertRequest struct {
	Name           string  `json:"name"`
	Metric         string  `json:"metric"`
	Condition      string  `json:"condition"`
	Threshold      float64 `json:"threshold"`
	Window         string  `json:"window"`
	Severity       string  `json:"severity"`
	NotifyChannels string  `json:"notify_channels"`
}

func apiWSDevOpsMonMetricsCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateMonitoringAlertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	metric := strings.TrimSpace(req.Metric)
	if metric == "" {
		apiError(w, http.StatusBadRequest, "Metric is required.")
		return
	}
	condition := strings.TrimSpace(req.Condition)
	if condition == "" {
		condition = ">"
	}
	window := strings.TrimSpace(req.Window)
	if window == "" {
		window = "5m"
	}
	severity := strings.TrimSpace(req.Severity)
	if severity == "" {
		severity = "warning"
	}
	channels := strings.TrimSpace(req.NotifyChannels)
	if channels == "" {
		channels = "Slack #alerts"
	}
	alert := MonitoringAlert{
		ID: newID(), Name: name, Metric: metric,
		Condition: condition, Threshold: req.Threshold,
		Window: window, Severity: severity,
		NotifyChannels: channels, EvalInterval: "30s",
		State: "pending", LastTriggered: "—",
		Silenced: false, Status: "active", Updated: "2026-04-18",
	}
	monitoringAlertStore.add(slug, alert)
	apiCreated(w, toAPIMonitoringAlert(alert))
}

func apiWSDevOpsMonMetricsGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	alert, found := monitoringAlertStore.getByID(slug, r.PathValue("aid"))
	if !found {
		apiError(w, http.StatusNotFound, "Alert rule not found.")
		return
	}
	apiOK(w, toAPIMonitoringAlert(alert))
}

func apiWSDevOpsMonMetricsDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := monitoringAlertStore.getByID(slug, r.PathValue("aid")); !found {
		apiError(w, http.StatusNotFound, "Alert rule not found.")
		return
	}
	monitoringAlertStore.delete(slug, r.PathValue("aid"))
	apiOK(w, map[string]any{"ok": true, "message": "Alert rule deleted."})
}
