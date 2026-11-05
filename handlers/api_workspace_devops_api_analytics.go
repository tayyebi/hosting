package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace DevOps API — API Management: Analytics
// devops/api/analytics
// ════════════════════════════════════════════════════════════════

type APIAnalytics struct {
	ID         string
	Name       string
	Endpoint   string
	Method     string
	AvgLatency string
	P99Latency string
	ErrorRate  string
	Requests   string
	Status     string
	Updated    string
}

type apiAnalyticsStoreType struct {
	mu        sync.RWMutex
	analytics map[string][]APIAnalytics
}

var apiAnalyticsStore = &apiAnalyticsStoreType{
	analytics: map[string][]APIAnalytics{},
}

func (s *apiAnalyticsStoreType) list(slug string) []APIAnalytics {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.analytics[slug]
	out := make([]APIAnalytics, len(src))
	copy(out, src)
	return out
}

func (s *apiAnalyticsStoreType) getByID(slug, id string) (APIAnalytics, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, a := range s.analytics[slug] {
		if a.ID == id {
			return a, true
		}
	}
	return APIAnalytics{}, false
}

func (s *apiAnalyticsStoreType) add(slug string, a APIAnalytics) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.analytics[slug] = append(s.analytics[slug], a)
}

func (s *apiAnalyticsStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.analytics[slug]
	for i, a := range list {
		if a.ID == id {
			s.analytics[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiAPIAnalytics struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Endpoint   string `json:"endpoint"`
	Method     string `json:"method"`
	AvgLatency string `json:"avg_latency"`
	P99Latency string `json:"p99_latency"`
	ErrorRate  string `json:"error_rate"`
	Requests   string `json:"requests"`
	Status     string `json:"status"`
	Updated    string `json:"updated"`
}

type apiAPIAnalyticsList struct {
	Analytics []apiAPIAnalytics `json:"analytics"`
	Total     int               `json:"total"`
}

func toAPIAPIAnalytics(a APIAnalytics) apiAPIAnalytics {
	return apiAPIAnalytics{
		ID: a.ID, Name: a.Name, Endpoint: a.Endpoint,
		Method: a.Method, AvgLatency: a.AvgLatency, P99Latency: a.P99Latency,
		ErrorRate: a.ErrorRate, Requests: a.Requests, Status: a.Status,
		Updated: a.Updated,
	}
}

func filterAPIAnalytics(all []APIAnalytics, search, status string) []APIAnalytics {
	var out []APIAnalytics
	q := strings.ToLower(search)
	for _, a := range all {
		if status != "" && a.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(a.Name + " " + a.Endpoint + " " + a.Method)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, a)
	}
	return out
}

func apiWSDevOpsAPIAnalyticsListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	analytics := filterAPIAnalytics(apiAnalyticsStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiAPIAnalytics, len(analytics))
	for i, a := range analytics {
		out[i] = toAPIAPIAnalytics(a)
	}
	apiOK(w, apiAPIAnalyticsList{Analytics: out, Total: len(out)})
}

type apiCreateAPIAnalyticsRequest struct {
	Name     string `json:"name"`
	Endpoint string `json:"endpoint"`
	Method   string `json:"method"`
}

func apiWSDevOpsAPIAnalyticsCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateAPIAnalyticsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	endpoint := strings.TrimSpace(req.Endpoint)
	if endpoint == "" {
		endpoint = "/api/v1/endpoint"
	}
	method := strings.TrimSpace(req.Method)
	if method == "" {
		method = "GET"
	}
	analytics := APIAnalytics{
		ID: newID(), Name: name, Endpoint: endpoint,
		Method: method, AvgLatency: "0ms", P99Latency: "0ms",
		ErrorRate: "0.00%", Requests: "0/day", Status: "healthy",
		Updated: "2026-04-17",
	}
	apiAnalyticsStore.add(slug, analytics)
	apiCreated(w, toAPIAPIAnalytics(analytics))
}

func apiWSDevOpsAPIAnalyticsGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	a, found := apiAnalyticsStore.getByID(slug, r.PathValue("aid"))
	if !found {
		apiError(w, http.StatusNotFound, "API analytics entry not found.")
		return
	}
	apiOK(w, toAPIAPIAnalytics(a))
}

func apiWSDevOpsAPIAnalyticsDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := apiAnalyticsStore.getByID(slug, r.PathValue("aid")); !found {
		apiError(w, http.StatusNotFound, "API analytics entry not found.")
		return
	}
	apiAnalyticsStore.delete(slug, r.PathValue("aid"))
	apiOK(w, map[string]any{"ok": true, "message": "API analytics entry deleted."})
}
