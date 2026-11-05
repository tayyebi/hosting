package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace DevOps API — IaC: Config Management
// devops/iac/config
// ════════════════════════════════════════════════════════════════

type IaCConfig struct {
	ID       string
	Name     string
	Tool     string
	Scope    string
	Targets  string
	Status   string
	Version  string
	Updated  string
}

type iacConfigStoreType struct {
	mu      sync.RWMutex
	configs map[string][]IaCConfig
}

var iacConfigStore = &iacConfigStoreType{
	configs: map[string][]IaCConfig{
		"buyvm": {
			{ID: newID(), Name: "Web Server Fleet", Tool: "Ansible", Scope: "production", Targets: "48 hosts", Status: "active", Version: "2.4.1", Updated: "2026-04-15"},
			{ID: newID(), Name: "Database Params", Tool: "Puppet", Scope: "production", Targets: "12 hosts", Status: "active", Version: "1.8.0", Updated: "2026-04-10"},
			{ID: newID(), Name: "Load Balancer Rules", Tool: "Chef", Scope: "staging", Targets: "6 hosts", Status: "active", Version: "3.1.2", Updated: "2026-03-28"},
			{ID: newID(), Name: "Firewall Policies", Tool: "Ansible", Scope: "all", Targets: "120 hosts", Status: "active", Version: "1.2.0", Updated: "2026-04-12"},
			{ID: newID(), Name: "SSL Certificate Rotation", Tool: "SaltStack", Scope: "production", Targets: "96 hosts", Status: "active", Version: "1.0.3", Updated: "2026-04-01"},
			{ID: newID(), Name: "Legacy App Config", Tool: "Puppet", Scope: "staging", Targets: "8 hosts", Status: "deprecated", Version: "0.9.0", Updated: "2025-11-15"},
		},
		"hosting": {
			{ID: newID(), Name: "Kubernetes Node Config", Tool: "Ansible", Scope: "production", Targets: "32 hosts", Status: "active", Version: "2.1.0", Updated: "2026-04-08"},
			{ID: newID(), Name: "Nginx Configuration", Tool: "Chef", Scope: "production", Targets: "16 hosts", Status: "active", Version: "1.5.2", Updated: "2026-03-20"},
			{ID: newID(), Name: "Monitoring Agent Setup", Tool: "Ansible", Scope: "all", Targets: "64 hosts", Status: "active", Version: "1.3.0", Updated: "2026-04-05"},
			{ID: newID(), Name: "DNS Resolver Config", Tool: "SaltStack", Scope: "production", Targets: "24 hosts", Status: "active", Version: "1.0.1", Updated: "2026-02-28"},
		},
	},
}

func (s *iacConfigStoreType) list(slug string) []IaCConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.configs[slug]
	out := make([]IaCConfig, len(src))
	copy(out, src)
	return out
}

func (s *iacConfigStoreType) getByID(slug, id string) (IaCConfig, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, c := range s.configs[slug] {
		if c.ID == id {
			return c, true
		}
	}
	return IaCConfig{}, false
}

func (s *iacConfigStoreType) add(slug string, c IaCConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.configs[slug] = append(s.configs[slug], c)
}

func (s *iacConfigStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.configs[slug]
	for i, c := range list {
		if c.ID == id {
			s.configs[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiIaCConfig struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Tool    string `json:"tool"`
	Scope   string `json:"scope"`
	Targets string `json:"targets"`
	Status  string `json:"status"`
	Version string `json:"version"`
	Updated string `json:"updated"`
}

type apiIaCConfigList struct {
	Configs []apiIaCConfig `json:"configs"`
	Total   int            `json:"total"`
}

func toAPIIaCConfig(c IaCConfig) apiIaCConfig {
	return apiIaCConfig{
		ID: c.ID, Name: c.Name, Tool: c.Tool,
		Scope: c.Scope, Targets: c.Targets, Status: c.Status,
		Version: c.Version, Updated: c.Updated,
	}
}

func filterIaCConfigs(all []IaCConfig, search, status string) []IaCConfig {
	var out []IaCConfig
	q := strings.ToLower(search)
	for _, c := range all {
		if status != "" && c.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(c.Name + " " + c.Tool + " " + c.Scope)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, c)
	}
	return out
}

func apiWSDevOpsIaCConfigListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	configs := filterIaCConfigs(iacConfigStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiIaCConfig, len(configs))
	for i, c := range configs {
		out[i] = toAPIIaCConfig(c)
	}
	apiOK(w, apiIaCConfigList{Configs: out, Total: len(out)})
}

type apiCreateIaCConfigRequest struct {
	Name  string `json:"name"`
	Tool  string `json:"tool"`
	Scope string `json:"scope"`
}

func apiWSDevOpsIaCConfigCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateIaCConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	tool := strings.TrimSpace(req.Tool)
	if tool == "" {
		tool = "Ansible"
	}
	scope := strings.TrimSpace(req.Scope)
	if scope == "" {
		scope = "staging"
	}
	cfg := IaCConfig{
		ID: newID(), Name: name, Tool: tool,
		Scope: scope, Targets: "0 hosts", Status: "active",
		Version: "1.0.0", Updated: "2026-04-17",
	}
	iacConfigStore.add(slug, cfg)
	apiCreated(w, toAPIIaCConfig(cfg))
}

func apiWSDevOpsIaCConfigGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	cfg, found := iacConfigStore.getByID(slug, r.PathValue("cid"))
	if !found {
		apiError(w, http.StatusNotFound, "Config management entry not found.")
		return
	}
	apiOK(w, toAPIIaCConfig(cfg))
}

func apiWSDevOpsIaCConfigDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := iacConfigStore.getByID(slug, r.PathValue("cid")); !found {
		apiError(w, http.StatusNotFound, "Config management entry not found.")
		return
	}
	iacConfigStore.delete(slug, r.PathValue("cid"))
	apiOK(w, map[string]any{"ok": true, "message": "Config management entry deleted."})
}
