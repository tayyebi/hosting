package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace AI API — MLOps: Monitoring & Drift Detection
// ai/mlops/monitoring
// ════════════════════════════════════════════════════════════════

type MLMonitor struct {
	ID         string
	Model      string
	Status     string
	DriftScore string
	Confidence string
	LatencyP99 string
	Inferences string
	AlertsDay  int
	Owner      string
	Region     string
}

type mlMonitorStoreType struct {
	mu       sync.RWMutex
	monitors map[string][]MLMonitor
}

var mlMonitorStore = &mlMonitorStoreType{
	monitors: map[string][]MLMonitor{},
}

func (s *mlMonitorStoreType) list(slug string) []MLMonitor {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.monitors[slug]
	out := make([]MLMonitor, len(src))
	copy(out, src)
	return out
}

func (s *mlMonitorStoreType) getByID(slug, id string) (MLMonitor, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, m := range s.monitors[slug] {
		if m.ID == id {
			return m, true
		}
	}
	return MLMonitor{}, false
}

func (s *mlMonitorStoreType) add(slug string, m MLMonitor) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.monitors[slug] = append(s.monitors[slug], m)
}

func (s *mlMonitorStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.monitors[slug]
	for i, m := range list {
		if m.ID == id {
			s.monitors[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiMLMonitor struct {
	ID         string `json:"id"`
	Model      string `json:"model"`
	Status     string `json:"status"`
	DriftScore string `json:"drift_score"`
	Confidence string `json:"confidence"`
	LatencyP99 string `json:"latency_p99"`
	Inferences string `json:"inferences"`
	AlertsDay  int    `json:"alerts_24h"`
	Owner      string `json:"owner"`
	Region     string `json:"region"`
}

type apiMLMonitorList struct {
	Monitors []apiMLMonitor `json:"monitors"`
	Total    int            `json:"total"`
}

func toAPIMLMonitor(m MLMonitor) apiMLMonitor {
	return apiMLMonitor{
		ID: m.ID, Model: m.Model, Status: m.Status,
		DriftScore: m.DriftScore, Confidence: m.Confidence,
		LatencyP99: m.LatencyP99, Inferences: m.Inferences,
		AlertsDay: m.AlertsDay, Owner: m.Owner,
		Region: m.Region,
	}
}

func filterMLMonitors(all []MLMonitor, search, status string) []MLMonitor {
	var out []MLMonitor
	q := strings.ToLower(search)
	for _, m := range all {
		if status != "" && m.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(m.Model + " " + m.Owner + " " + m.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, m)
	}
	return out
}

func apiWSAIMLOpsMonitoringListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	monitors := filterMLMonitors(mlMonitorStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiMLMonitor, len(monitors))
	for i, m := range monitors {
		out[i] = toAPIMLMonitor(m)
	}
	apiOK(w, apiMLMonitorList{Monitors: out, Total: len(out)})
}

type apiCreateMLMonitorRequest struct {
	Model  string `json:"model"`
	Owner  string `json:"owner"`
	Region string `json:"region"`
}

func apiWSAIMLOpsMonitoringCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateMLMonitorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	model := strings.TrimSpace(req.Model)
	if model == "" {
		apiError(w, http.StatusBadRequest, "Model name is required.")
		return
	}
	owner := strings.TrimSpace(req.Owner)
	if owner == "" {
		owner = "unassigned"
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "eu-central-1"
	}
	monitor := MLMonitor{
		ID: newID(), Model: model, Status: "initializing",
		DriftScore: "—", Confidence: "—",
		LatencyP99: "—", Inferences: "0",
		AlertsDay: 0, Owner: owner,
		Region: region,
	}
	mlMonitorStore.add(slug, monitor)
	apiCreated(w, toAPIMLMonitor(monitor))
}

func apiWSAIMLOpsMonitoringGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	monitor, found := mlMonitorStore.getByID(slug, r.PathValue("mid"))
	if !found {
		apiError(w, http.StatusNotFound, "Monitor not found.")
		return
	}
	apiOK(w, toAPIMLMonitor(monitor))
}

func apiWSAIMLOpsMonitoringDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := mlMonitorStore.getByID(slug, r.PathValue("mid")); !found {
		apiError(w, http.StatusNotFound, "Monitor not found.")
		return
	}
	mlMonitorStore.delete(slug, r.PathValue("mid"))
	apiOK(w, map[string]any{"ok": true, "message": "Monitor deleted."})
}
