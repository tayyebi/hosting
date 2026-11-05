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
	monitors: map[string][]MLMonitor{
		"buyvm": {
			{ID: newID(), Model: "acme-recommender", Status: "healthy", DriftScore: "0.04", Confidence: "0.92", LatencyP99: "8 ms", Inferences: "4.2M", AlertsDay: 0, Owner: "growth-team", Region: "eu-central-1"},
			{ID: newID(), Model: "churn-pred", Status: "drift", DriftScore: "0.18", Confidence: "0.84", LatencyP99: "12 ms", Inferences: "1.8M", AlertsDay: 1, Owner: "data-team", Region: "eu-central-1"},
			{ID: newID(), Model: "llm-acme-v4", Status: "healthy", DriftScore: "—", Confidence: "0.91", LatencyP99: "842 ms", Inferences: "420K", AlertsDay: 0, Owner: "ml-team", Region: "eu-central-1"},
			{ID: newID(), Model: "vision-defect", Status: "healthy", DriftScore: "0.02", Confidence: "0.97", LatencyP99: "42 ms", Inferences: "6.4M", AlertsDay: 0, Owner: "cv-team", Region: "eu-central-1"},
		},
		"hosting": {
			{ID: newID(), Model: "fraud-classifier", Status: "healthy", DriftScore: "0.06", Confidence: "0.93", LatencyP99: "14 ms", Inferences: "2.8M", AlertsDay: 0, Owner: "security-team", Region: "ap-south-1"},
			{ID: newID(), Model: "price-optimizer", Status: "healthy", DriftScore: "0.08", Confidence: "0.89", LatencyP99: "22 ms", Inferences: "840K", AlertsDay: 0, Owner: "pricing-team", Region: "ap-south-1"},
			{ID: newID(), Model: "embedding-model", Status: "degraded", DriftScore: "0.14", Confidence: "0.86", LatencyP99: "180 ms", Inferences: "1.2M", AlertsDay: 2, Owner: "ml-team", Region: "ap-south-1"},
		},
	},
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
