package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Apps API — Serverless: Functions (Pro Feature)
// apps/serverless/functions
// ════════════════════════════════════════════════════════════════

type ProServerlessFunc struct {
	ID            string
	Name          string
	Runtime       string
	Trigger       string
	Route         string
	MemoryMB      int
	TimeoutSec    int
	EnvVars       int
	Region        string
	Invocations24 int
	AvgDurationMs int
	ErrorRate     float64
	LastDeployed  string
	CodeSizeKB    int
	ColdStartMs   int
	Status        string
	Updated       string
}

type proServerlessFuncStoreType struct {
	mu    sync.RWMutex
	items map[string][]ProServerlessFunc
}

var proServerlessFuncStore = &proServerlessFuncStoreType{
	items: map[string][]ProServerlessFunc{},
}

func (s *proServerlessFuncStoreType) list(slug string) []ProServerlessFunc {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.items[slug]
	out := make([]ProServerlessFunc, len(src))
	copy(out, src)
	return out
}

func (s *proServerlessFuncStoreType) getByID(slug, id string) (ProServerlessFunc, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, f := range s.items[slug] {
		if f.ID == id {
			return f, true
		}
	}
	return ProServerlessFunc{}, false
}

func (s *proServerlessFuncStoreType) add(slug string, f ProServerlessFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[slug] = append(s.items[slug], f)
}

func (s *proServerlessFuncStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.items[slug]
	for i, f := range list {
		if f.ID == id {
			s.items[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiProServerlessFunc struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Runtime       string  `json:"runtime"`
	Trigger       string  `json:"trigger"`
	Route         string  `json:"route"`
	MemoryMB      int     `json:"memory_mb"`
	TimeoutSec    int     `json:"timeout_sec"`
	EnvVars       int     `json:"env_vars"`
	Region        string  `json:"region"`
	Invocations24 int     `json:"invocations_24h"`
	AvgDurationMs int     `json:"avg_duration_ms"`
	ErrorRate     float64 `json:"error_rate"`
	LastDeployed  string  `json:"last_deployed"`
	CodeSizeKB    int     `json:"code_size_kb"`
	ColdStartMs   int     `json:"cold_start_ms"`
	Status        string  `json:"status"`
	Updated       string  `json:"updated"`
}

type apiProServerlessFuncList struct {
	Functions []apiProServerlessFunc `json:"functions"`
	Total     int                    `json:"total"`
}

func toAPIProServerlessFunc(f ProServerlessFunc) apiProServerlessFunc {
	return apiProServerlessFunc{
		ID: f.ID, Name: f.Name, Runtime: f.Runtime,
		Trigger: f.Trigger, Route: f.Route,
		MemoryMB: f.MemoryMB, TimeoutSec: f.TimeoutSec,
		EnvVars: f.EnvVars, Region: f.Region,
		Invocations24: f.Invocations24, AvgDurationMs: f.AvgDurationMs,
		ErrorRate: f.ErrorRate, LastDeployed: f.LastDeployed,
		CodeSizeKB: f.CodeSizeKB, ColdStartMs: f.ColdStartMs,
		Status: f.Status, Updated: f.Updated,
	}
}

func filterProServerlessFuncs(all []ProServerlessFunc, search, status string) []ProServerlessFunc {
	var out []ProServerlessFunc
	q := strings.ToLower(search)
	for _, f := range all {
		if status != "" && f.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(f.Name + " " + f.Runtime + " " + f.Trigger + " " + f.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, f)
	}
	return out
}

func apiWSAppsServerlessFuncListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	items := filterProServerlessFuncs(proServerlessFuncStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiProServerlessFunc, len(items))
	for i, f := range items {
		out[i] = toAPIProServerlessFunc(f)
	}
	apiOK(w, apiProServerlessFuncList{Functions: out, Total: len(out)})
}

type apiCreateProServerlessFuncRequest struct {
	Name       string `json:"name"`
	Runtime    string `json:"runtime"`
	Trigger    string `json:"trigger"`
	Route      string `json:"route"`
	MemoryMB   int    `json:"memory_mb"`
	TimeoutSec int    `json:"timeout_sec"`
	Region     string `json:"region"`
}

func apiWSAppsServerlessFuncCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateProServerlessFuncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	runtime := strings.TrimSpace(req.Runtime)
	if runtime == "" {
		runtime = "Node.js 20"
	}
	trigger := strings.TrimSpace(req.Trigger)
	if trigger == "" {
		trigger = "HTTP GET"
	}
	route := strings.TrimSpace(req.Route)
	if route == "" {
		route = "/" + name
	}
	memoryMB := req.MemoryMB
	if memoryMB <= 0 {
		memoryMB = 256
	}
	timeoutSec := req.TimeoutSec
	if timeoutSec <= 0 {
		timeoutSec = 10
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "us-east-1"
	}
	fn := ProServerlessFunc{
		ID: newID(), Name: name, Runtime: runtime,
		Trigger: trigger, Route: route,
		MemoryMB: memoryMB, TimeoutSec: timeoutSec,
		EnvVars: 0, Region: region,
		Invocations24: 0, AvgDurationMs: 0, ErrorRate: 0,
		LastDeployed: "—", CodeSizeKB: 0, ColdStartMs: 0,
		Status: "deploying", Updated: "2026-04-18",
	}
	proServerlessFuncStore.add(slug, fn)
	apiCreated(w, toAPIProServerlessFunc(fn))
}

func apiWSAppsServerlessFuncGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	fn, found := proServerlessFuncStore.getByID(slug, r.PathValue("fid"))
	if !found {
		apiError(w, http.StatusNotFound, "Serverless function not found.")
		return
	}
	apiOK(w, toAPIProServerlessFunc(fn))
}

func apiWSAppsServerlessFuncDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := proServerlessFuncStore.getByID(slug, r.PathValue("fid")); !found {
		apiError(w, http.StatusNotFound, "Serverless function not found.")
		return
	}
	proServerlessFuncStore.delete(slug, r.PathValue("fid"))
	apiOK(w, map[string]any{"ok": true, "message": "Serverless function deleted."})
}
