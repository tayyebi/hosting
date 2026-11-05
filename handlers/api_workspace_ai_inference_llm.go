package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace AI API — Inference: LLM Endpoints
// ai/inference/llm-endpoints
// ════════════════════════════════════════════════════════════════

type LLMEndpoint struct {
	ID              string
	Name            string
	Model           string
	ModelFormat     string
	GPUType         string
	GPUCount        int
	ContextWindow   int
	MaxTokens       int
	QuantBits       int
	ReplicasCurrent int
	ReplicasMax     int
	RPM             int
	AvgLatencyMs    int
	CostPerHour     float64
	Region          string
	ScaleToZero     bool
	Status          string
	Updated         string
}

type llmEndpointStoreType struct {
	mu    sync.RWMutex
	items map[string][]LLMEndpoint
}

var llmEndpointStore = &llmEndpointStoreType{
	items: map[string][]LLMEndpoint{},
}

func (s *llmEndpointStoreType) list(slug string) []LLMEndpoint {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.items[slug]
	out := make([]LLMEndpoint, len(src))
	copy(out, src)
	return out
}

func (s *llmEndpointStoreType) getByID(slug, id string) (LLMEndpoint, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, e := range s.items[slug] {
		if e.ID == id {
			return e, true
		}
	}
	return LLMEndpoint{}, false
}

func (s *llmEndpointStoreType) add(slug string, e LLMEndpoint) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[slug] = append(s.items[slug], e)
}

func (s *llmEndpointStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.items[slug]
	for i, e := range list {
		if e.ID == id {
			s.items[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiLLMEndpoint struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Model           string  `json:"model"`
	ModelFormat     string  `json:"model_format"`
	GPUType         string  `json:"gpu_type"`
	GPUCount        int     `json:"gpu_count"`
	ContextWindow   int     `json:"context_window"`
	MaxTokens       int     `json:"max_tokens"`
	QuantBits       int     `json:"quant_bits"`
	ReplicasCurrent int     `json:"replicas_current"`
	ReplicasMax     int     `json:"replicas_max"`
	RPM             int     `json:"rpm"`
	AvgLatencyMs    int     `json:"avg_latency_ms"`
	CostPerHour     float64 `json:"cost_per_hour"`
	Region          string  `json:"region"`
	ScaleToZero     bool    `json:"scale_to_zero"`
	Status          string  `json:"status"`
	Updated         string  `json:"updated"`
}

type apiLLMEndpointList struct {
	Endpoints []apiLLMEndpoint `json:"endpoints"`
	Total     int              `json:"total"`
}

func toAPILLMEndpoint(e LLMEndpoint) apiLLMEndpoint {
	return apiLLMEndpoint{
		ID: e.ID, Name: e.Name, Model: e.Model,
		ModelFormat: e.ModelFormat, GPUType: e.GPUType,
		GPUCount: e.GPUCount, ContextWindow: e.ContextWindow,
		MaxTokens: e.MaxTokens, QuantBits: e.QuantBits,
		ReplicasCurrent: e.ReplicasCurrent, ReplicasMax: e.ReplicasMax,
		RPM: e.RPM, AvgLatencyMs: e.AvgLatencyMs,
		CostPerHour: e.CostPerHour, Region: e.Region,
		ScaleToZero: e.ScaleToZero, Status: e.Status,
		Updated: e.Updated,
	}
}

func filterLLMEndpoints(all []LLMEndpoint, search, status string) []LLMEndpoint {
	var out []LLMEndpoint
	q := strings.ToLower(search)
	for _, e := range all {
		if status != "" && e.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(e.Name + " " + e.Model + " " + e.GPUType + " " + e.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, e)
	}
	return out
}

func apiWSAIInferenceLLMListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	items := filterLLMEndpoints(llmEndpointStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiLLMEndpoint, len(items))
	for i, e := range items {
		out[i] = toAPILLMEndpoint(e)
	}
	apiOK(w, apiLLMEndpointList{Endpoints: out, Total: len(out)})
}

type apiCreateLLMEndpointRequest struct {
	Name        string `json:"name"`
	Model       string `json:"model"`
	ModelFormat string `json:"model_format"`
	GPUType     string `json:"gpu_type"`
	GPUCount    int    `json:"gpu_count"`
	Region      string `json:"region"`
	ScaleToZero bool   `json:"scale_to_zero"`
}

func apiWSAIInferenceLLMCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateLLMEndpointRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	model := strings.TrimSpace(req.Model)
	if model == "" {
		apiError(w, http.StatusBadRequest, "Model is required.")
		return
	}
	modelFormat := strings.TrimSpace(req.ModelFormat)
	if modelFormat == "" {
		modelFormat = "SafeTensors"
	}
	gpuType := strings.TrimSpace(req.GPUType)
	if gpuType == "" {
		gpuType = "L4 24GB"
	}
	gpuCount := req.GPUCount
	if gpuCount <= 0 {
		gpuCount = 1
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "us-east-1"
	}
	ep := LLMEndpoint{
		ID: newID(), Name: name, Model: model,
		ModelFormat: modelFormat, GPUType: gpuType,
		GPUCount: gpuCount, ContextWindow: 4096, MaxTokens: 2048,
		QuantBits: 16, ReplicasCurrent: 0, ReplicasMax: gpuCount * 2,
		RPM: 0, AvgLatencyMs: 0, CostPerHour: 0,
		Region: region, ScaleToZero: req.ScaleToZero,
		Status: "provisioning", Updated: "2026-04-18",
	}
	llmEndpointStore.add(slug, ep)
	apiCreated(w, toAPILLMEndpoint(ep))
}

func apiWSAIInferenceLLMGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	ep, found := llmEndpointStore.getByID(slug, r.PathValue("eid"))
	if !found {
		apiError(w, http.StatusNotFound, "LLM endpoint not found.")
		return
	}
	apiOK(w, toAPILLMEndpoint(ep))
}

func apiWSAIInferenceLLMDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := llmEndpointStore.getByID(slug, r.PathValue("eid")); !found {
		apiError(w, http.StatusNotFound, "LLM endpoint not found.")
		return
	}
	llmEndpointStore.delete(slug, r.PathValue("eid"))
	apiOK(w, map[string]any{"ok": true, "message": "LLM endpoint deleted."})
}
