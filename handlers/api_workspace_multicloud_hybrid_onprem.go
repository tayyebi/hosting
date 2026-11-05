package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Multi-Cloud & Hybrid API — Hybrid: On-premises Extension
// multicloud/hybrid/onprem
// ════════════════════════════════════════════════════════════════

type OnPremExtension struct {
	ID         string
	Name       string
	Location   string
	Connector  string
	Nodes      int
	CPUCores   int
	MemoryGB   int
	Latency    string
	SyncStatus string
	Status     string
	Updated    string
}

type onPremExtensionStoreType struct {
	mu    sync.RWMutex
	items map[string][]OnPremExtension
}

var onPremExtensionStore = &onPremExtensionStoreType{
	items: map[string][]OnPremExtension{},
}

func (s *onPremExtensionStoreType) list(slug string) []OnPremExtension {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.items[slug]
	out := make([]OnPremExtension, len(src))
	copy(out, src)
	return out
}

func (s *onPremExtensionStoreType) getByID(slug, id string) (OnPremExtension, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, o := range s.items[slug] {
		if o.ID == id {
			return o, true
		}
	}
	return OnPremExtension{}, false
}

func (s *onPremExtensionStoreType) add(slug string, o OnPremExtension) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[slug] = append(s.items[slug], o)
}

func (s *onPremExtensionStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.items[slug]
	for i, o := range list {
		if o.ID == id {
			s.items[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiOnPremExtension struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Location   string `json:"location"`
	Connector  string `json:"connector"`
	Nodes      int    `json:"nodes"`
	CPUCores   int    `json:"cpu_cores"`
	MemoryGB   int    `json:"memory_gb"`
	Latency    string `json:"latency"`
	SyncStatus string `json:"sync_status"`
	Status     string `json:"status"`
	Updated    string `json:"updated"`
}

type apiOnPremExtensionList struct {
	Extensions []apiOnPremExtension `json:"extensions"`
	Total      int                  `json:"total"`
}

func toAPIOnPremExtension(o OnPremExtension) apiOnPremExtension {
	return apiOnPremExtension{
		ID: o.ID, Name: o.Name, Location: o.Location,
		Connector: o.Connector, Nodes: o.Nodes, CPUCores: o.CPUCores,
		MemoryGB: o.MemoryGB, Latency: o.Latency, SyncStatus: o.SyncStatus,
		Status: o.Status, Updated: o.Updated,
	}
}

func filterOnPremExtensions(all []OnPremExtension, search, status string) []OnPremExtension {
	var out []OnPremExtension
	q := strings.ToLower(search)
	for _, o := range all {
		if status != "" && o.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(o.Name + " " + o.Location + " " + o.Connector)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, o)
	}
	return out
}

func apiWSMulticloudHybridOnpremListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	items := filterOnPremExtensions(onPremExtensionStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiOnPremExtension, len(items))
	for i, o := range items {
		out[i] = toAPIOnPremExtension(o)
	}
	apiOK(w, apiOnPremExtensionList{Extensions: out, Total: len(out)})
}

type apiCreateOnPremExtensionRequest struct {
	Name      string `json:"name"`
	Location  string `json:"location"`
	Connector string `json:"connector"`
	Nodes     int    `json:"nodes"`
	CPUCores  int    `json:"cpu_cores"`
	MemoryGB  int    `json:"memory_gb"`
}

func apiWSMulticloudHybridOnpremCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateOnPremExtensionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	location := strings.TrimSpace(req.Location)
	if location == "" {
		location = "Unknown"
	}
	connector := strings.TrimSpace(req.Connector)
	if connector == "" {
		connector = "Arc Agent v3.2"
	}
	nodes := req.Nodes
	if nodes <= 0 {
		nodes = 1
	}
	cpuCores := req.CPUCores
	if cpuCores <= 0 {
		cpuCores = 8
	}
	memoryGB := req.MemoryGB
	if memoryGB <= 0 {
		memoryGB = 32
	}
	ext := OnPremExtension{
		ID: newID(), Name: name, Location: location,
		Connector: connector, Nodes: nodes, CPUCores: cpuCores,
		MemoryGB: memoryGB, Latency: "—", SyncStatus: "Pending",
		Status: "active", Updated: "2026-04-17",
	}
	onPremExtensionStore.add(slug, ext)
	apiCreated(w, toAPIOnPremExtension(ext))
}

func apiWSMulticloudHybridOnpremGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	ext, found := onPremExtensionStore.getByID(slug, r.PathValue("oid"))
	if !found {
		apiError(w, http.StatusNotFound, "On-premises extension not found.")
		return
	}
	apiOK(w, toAPIOnPremExtension(ext))
}

func apiWSMulticloudHybridOnpremDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := onPremExtensionStore.getByID(slug, r.PathValue("oid")); !found {
		apiError(w, http.StatusNotFound, "On-premises extension not found.")
		return
	}
	onPremExtensionStore.delete(slug, r.PathValue("oid"))
	apiOK(w, map[string]any{"ok": true, "message": "On-premises extension deleted."})
}
