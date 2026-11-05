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
	items: map[string][]ProServerlessFunc{
		"buyvm": {
			{ID: newID(), Name: "api-auth-handler", Runtime: "Node.js 20", Trigger: "HTTP POST", Route: "/auth/callback", MemoryMB: 256, TimeoutSec: 10, EnvVars: 5, Region: "us-east-1", Invocations24: 14500, AvgDurationMs: 35, ErrorRate: 0.1, LastDeployed: "2026-04-17 14:22", CodeSizeKB: 128, ColdStartMs: 8, Status: "active", Updated: "2026-04-17"},
			{ID: newID(), Name: "image-resizer", Runtime: "Python 3.12", Trigger: "Queue", Route: "—", MemoryMB: 512, TimeoutSec: 30, EnvVars: 3, Region: "us-east-1", Invocations24: 8200, AvgDurationMs: 180, ErrorRate: 0.3, LastDeployed: "2026-04-16 09:15", CodeSizeKB: 340, ColdStartMs: 45, Status: "active", Updated: "2026-04-16"},
			{ID: newID(), Name: "cron-cleanup", Runtime: "Go 1.22", Trigger: "Cron */15 * * * *", Route: "—", MemoryMB: 128, TimeoutSec: 60, EnvVars: 2, Region: "us-east-1", Invocations24: 96, AvgDurationMs: 420, ErrorRate: 0, LastDeployed: "2026-04-15 11:00", CodeSizeKB: 64, ColdStartMs: 4, Status: "active", Updated: "2026-04-15"},
			{ID: newID(), Name: "webhook-processor", Runtime: "Node.js 20", Trigger: "HTTP POST", Route: "/webhooks/stripe", MemoryMB: 256, TimeoutSec: 15, EnvVars: 8, Region: "us-east-1", Invocations24: 3100, AvgDurationMs: 55, ErrorRate: 0.05, LastDeployed: "2026-04-17 18:30", CodeSizeKB: 96, ColdStartMs: 7, Status: "active", Updated: "2026-04-17"},
			{ID: newID(), Name: "pdf-generator", Runtime: "Python 3.12", Trigger: "HTTP GET", Route: "/reports/pdf", MemoryMB: 1024, TimeoutSec: 60, EnvVars: 4, Region: "eu-west-1", Invocations24: 450, AvgDurationMs: 2200, ErrorRate: 1.2, LastDeployed: "2026-04-14 16:00", CodeSizeKB: 820, ColdStartMs: 120, Status: "active", Updated: "2026-04-14"},
			{ID: newID(), Name: "legacy-transform", Runtime: "Node.js 18", Trigger: "Queue", Route: "—", MemoryMB: 256, TimeoutSec: 30, EnvVars: 6, Region: "us-east-1", Invocations24: 0, AvgDurationMs: 0, ErrorRate: 0, LastDeployed: "2026-03-01 10:00", CodeSizeKB: 210, ColdStartMs: 0, Status: "inactive", Updated: "2026-03-01"},
		},
		"hosting": {
			{ID: newID(), Name: "edge-auth-verify", Runtime: "Rust", Trigger: "HTTP GET", Route: "/v1/verify", MemoryMB: 128, TimeoutSec: 5, EnvVars: 3, Region: "global-edge", Invocations24: 52000, AvgDurationMs: 8, ErrorRate: 0.01, LastDeployed: "2026-04-17 20:00", CodeSizeKB: 48, ColdStartMs: 2, Status: "active", Updated: "2026-04-17"},
			{ID: newID(), Name: "data-etl-pipeline", Runtime: "Python 3.12", Trigger: "Cron 0 */2 * * *", Route: "—", MemoryMB: 2048, TimeoutSec: 300, EnvVars: 12, Region: "us-west-2", Invocations24: 12, AvgDurationMs: 45000, ErrorRate: 0, LastDeployed: "2026-04-16 08:00", CodeSizeKB: 1500, ColdStartMs: 200, Status: "active", Updated: "2026-04-16"},
			{ID: newID(), Name: "notification-dispatch", Runtime: "Node.js 20", Trigger: "Queue", Route: "—", MemoryMB: 256, TimeoutSec: 15, EnvVars: 7, Region: "eu-central-1", Invocations24: 18400, AvgDurationMs: 22, ErrorRate: 0.05, LastDeployed: "2026-04-17 11:45", CodeSizeKB: 85, ColdStartMs: 6, Status: "active", Updated: "2026-04-17"},
			{ID: newID(), Name: "og-image-gen", Runtime: "Node.js 20", Trigger: "HTTP GET", Route: "/og/{slug}", MemoryMB: 512, TimeoutSec: 10, EnvVars: 2, Region: "us-west-2", Invocations24: 6700, AvgDurationMs: 95, ErrorRate: 0.2, LastDeployed: "2026-04-15 14:30", CodeSizeKB: 420, ColdStartMs: 15, Status: "active", Updated: "2026-04-15"},
			{ID: newID(), Name: "ml-feature-store", Runtime: "Python 3.12", Trigger: "HTTP POST", Route: "/features/compute", MemoryMB: 1024, TimeoutSec: 30, EnvVars: 9, Region: "us-west-2", Invocations24: 2300, AvgDurationMs: 140, ErrorRate: 0.08, LastDeployed: "2026-04-17 07:00", CodeSizeKB: 650, ColdStartMs: 80, Status: "active", Updated: "2026-04-17"},
		},
	},
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
