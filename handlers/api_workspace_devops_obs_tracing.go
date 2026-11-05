package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace DevOps API — Observability: Distributed Tracing
// devops/obs/tracing
// ════════════════════════════════════════════════════════════════

type TracingEndpoint struct {
	ID           string
	Name         string
	Backend      string
	Protocol     string
	SampleRate   string
	Services     int
	SpansPerSec  string
	Retention    string
	Status       string
	Updated      string
}

type tracingEndpointStoreType struct {
	mu        sync.RWMutex
	endpoints map[string][]TracingEndpoint
}

var tracingEndpointStore = &tracingEndpointStoreType{
	endpoints: map[string][]TracingEndpoint{
		"buyvm": {
			{ID: newID(), Name: "Production Tracing", Backend: "Jaeger", Protocol: "OpenTelemetry gRPC", SampleRate: "10%", Services: 48, SpansPerSec: "125,000", Retention: "14 days", Status: "active", Updated: "2026-04-16"},
			{ID: newID(), Name: "Staging Tracing", Backend: "Jaeger", Protocol: "OpenTelemetry gRPC", SampleRate: "100%", Services: 48, SpansPerSec: "8,200", Retention: "7 days", Status: "active", Updated: "2026-04-15"},
			{ID: newID(), Name: "Edge Service Mesh", Backend: "Zipkin", Protocol: "B3 propagation", SampleRate: "5%", Services: 12, SpansPerSec: "34,000", Retention: "7 days", Status: "active", Updated: "2026-04-12"},
			{ID: newID(), Name: "Payment Flow Tracing", Backend: "Tempo", Protocol: "OpenTelemetry HTTP", SampleRate: "100%", Services: 6, SpansPerSec: "4,500", Retention: "30 days", Status: "active", Updated: "2026-04-10"},
			{ID: newID(), Name: "Legacy X-Ray", Backend: "AWS X-Ray", Protocol: "X-Ray SDK", SampleRate: "1%", Services: 3, SpansPerSec: "1,200", Retention: "30 days", Status: "deprecated", Updated: "2025-08-15"},
		},
		"hosting": {
			{ID: newID(), Name: "Platform Tracing", Backend: "Tempo", Protocol: "OpenTelemetry gRPC", SampleRate: "15%", Services: 32, SpansPerSec: "85,000", Retention: "14 days", Status: "active", Updated: "2026-04-14"},
			{ID: newID(), Name: "API Gateway Traces", Backend: "Tempo", Protocol: "OpenTelemetry HTTP", SampleRate: "50%", Services: 8, SpansPerSec: "22,000", Retention: "7 days", Status: "active", Updated: "2026-04-09"},
			{ID: newID(), Name: "Database Query Tracing", Backend: "Jaeger", Protocol: "OpenTelemetry gRPC", SampleRate: "5%", Services: 4, SpansPerSec: "6,800", Retention: "7 days", Status: "active", Updated: "2026-04-03"},
		},
	},
}

func (s *tracingEndpointStoreType) list(slug string) []TracingEndpoint {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.endpoints[slug]
	out := make([]TracingEndpoint, len(src))
	copy(out, src)
	return out
}

func (s *tracingEndpointStoreType) getByID(slug, id string) (TracingEndpoint, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, e := range s.endpoints[slug] {
		if e.ID == id {
			return e, true
		}
	}
	return TracingEndpoint{}, false
}

func (s *tracingEndpointStoreType) add(slug string, e TracingEndpoint) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.endpoints[slug] = append(s.endpoints[slug], e)
}

func (s *tracingEndpointStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.endpoints[slug]
	for i, e := range list {
		if e.ID == id {
			s.endpoints[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiTracingEndpoint struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Backend     string `json:"backend"`
	Protocol    string `json:"protocol"`
	SampleRate  string `json:"sample_rate"`
	Services    int    `json:"services"`
	SpansPerSec string `json:"spans_per_sec"`
	Retention   string `json:"retention"`
	Status      string `json:"status"`
	Updated     string `json:"updated"`
}

type apiTracingEndpointList struct {
	Endpoints []apiTracingEndpoint `json:"endpoints"`
	Total     int                  `json:"total"`
}

func toAPITracingEndpoint(e TracingEndpoint) apiTracingEndpoint {
	return apiTracingEndpoint{
		ID: e.ID, Name: e.Name, Backend: e.Backend,
		Protocol: e.Protocol, SampleRate: e.SampleRate, Services: e.Services,
		SpansPerSec: e.SpansPerSec, Retention: e.Retention,
		Status: e.Status, Updated: e.Updated,
	}
}

func filterTracingEndpoints(all []TracingEndpoint, search, status string) []TracingEndpoint {
	var out []TracingEndpoint
	q := strings.ToLower(search)
	for _, e := range all {
		if status != "" && e.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(e.Name + " " + e.Backend + " " + e.Protocol)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, e)
	}
	return out
}

func apiWSDevOpsObsTracingListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	endpoints := filterTracingEndpoints(tracingEndpointStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiTracingEndpoint, len(endpoints))
	for i, e := range endpoints {
		out[i] = toAPITracingEndpoint(e)
	}
	apiOK(w, apiTracingEndpointList{Endpoints: out, Total: len(out)})
}

type apiCreateTracingEndpointRequest struct {
	Name       string `json:"name"`
	Backend    string `json:"backend"`
	Protocol   string `json:"protocol"`
	SampleRate string `json:"sample_rate"`
	Retention  string `json:"retention"`
}

func apiWSDevOpsObsTracingCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateTracingEndpointRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	backend := strings.TrimSpace(req.Backend)
	if backend == "" {
		backend = "Jaeger"
	}
	protocol := strings.TrimSpace(req.Protocol)
	if protocol == "" {
		protocol = "OpenTelemetry gRPC"
	}
	sampleRate := strings.TrimSpace(req.SampleRate)
	if sampleRate == "" {
		sampleRate = "10%"
	}
	retention := strings.TrimSpace(req.Retention)
	if retention == "" {
		retention = "14 days"
	}
	endpoint := TracingEndpoint{
		ID: newID(), Name: name, Backend: backend,
		Protocol: protocol, SampleRate: sampleRate, Services: 0,
		SpansPerSec: "0", Retention: retention,
		Status: "active", Updated: "2026-04-17",
	}
	tracingEndpointStore.add(slug, endpoint)
	apiCreated(w, toAPITracingEndpoint(endpoint))
}

func apiWSDevOpsObsTracingGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	endpoint, found := tracingEndpointStore.getByID(slug, r.PathValue("tid"))
	if !found {
		apiError(w, http.StatusNotFound, "Tracing endpoint not found.")
		return
	}
	apiOK(w, toAPITracingEndpoint(endpoint))
}

func apiWSDevOpsObsTracingDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := tracingEndpointStore.getByID(slug, r.PathValue("tid")); !found {
		apiError(w, http.StatusNotFound, "Tracing endpoint not found.")
		return
	}
	tracingEndpointStore.delete(slug, r.PathValue("tid"))
	apiOK(w, map[string]any{"ok": true, "message": "Tracing endpoint deleted."})
}
