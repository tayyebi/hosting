package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Multi-Cloud & Hybrid API — Hybrid: Edge Computing
// multicloud/hybrid/edge
// ════════════════════════════════════════════════════════════════

type EdgeNode struct {
	ID          string
	Name        string
	Location    string
	Provider    string
	CPUCores    int
	MemoryGB    int
	Workloads   int
	Uptime      string
	Latency     string
	LastContact string
	Status      string
	Updated     string
}

type edgeNodeStoreType struct {
	mu    sync.RWMutex
	items map[string][]EdgeNode
}

var edgeNodeStore = &edgeNodeStoreType{
	items: map[string][]EdgeNode{
		"buyvm": {
			{ID: newID(), Name: "edge-nyc-pop1", Location: "New York, US", Provider: "Equinix Metal", CPUCores: 16, MemoryGB: 64, Workloads: 8, Uptime: "99.98%", Latency: "2 ms", LastContact: "2026-04-17 09:58", Status: "active", Updated: "2026-04-17"},
			{ID: newID(), Name: "edge-lon-pop1", Location: "London, UK", Provider: "Packet", CPUCores: 16, MemoryGB: 64, Workloads: 6, Uptime: "99.95%", Latency: "14 ms", LastContact: "2026-04-17 09:57", Status: "active", Updated: "2026-04-17"},
			{ID: newID(), Name: "edge-tok-pop1", Location: "Tokyo, JP", Provider: "Equinix Metal", CPUCores: 8, MemoryGB: 32, Workloads: 4, Uptime: "99.92%", Latency: "85 ms", LastContact: "2026-04-17 09:55", Status: "active", Updated: "2026-04-16"},
			{ID: newID(), Name: "edge-syd-pop1", Location: "Sydney, AU", Provider: "Vultr Bare Metal", CPUCores: 8, MemoryGB: 32, Workloads: 3, Uptime: "99.88%", Latency: "120 ms", LastContact: "2026-04-17 09:50", Status: "active", Updated: "2026-04-16"},
			{ID: newID(), Name: "edge-retail-store42", Location: "Dallas, US", Provider: "On-premises", CPUCores: 4, MemoryGB: 16, Workloads: 2, Uptime: "98.5%", Latency: "8 ms", LastContact: "2026-04-17 08:30", Status: "degraded", Updated: "2026-04-15"},
			{ID: newID(), Name: "edge-factory-floor", Location: "Detroit, US", Provider: "On-premises", CPUCores: 4, MemoryGB: 8, Workloads: 1, Uptime: "95.0%", Latency: "5 ms", LastContact: "2026-04-16 22:00", Status: "inactive", Updated: "2026-04-10"},
		},
		"hosting": {
			{ID: newID(), Name: "edge-ams-pop1", Location: "Amsterdam, NL", Provider: "Equinix Metal", CPUCores: 32, MemoryGB: 128, Workloads: 14, Uptime: "99.99%", Latency: "3 ms", LastContact: "2026-04-17 09:59", Status: "active", Updated: "2026-04-17"},
			{ID: newID(), Name: "edge-sgp-pop1", Location: "Singapore, SG", Provider: "Packet", CPUCores: 16, MemoryGB: 64, Workloads: 9, Uptime: "99.96%", Latency: "45 ms", LastContact: "2026-04-17 09:58", Status: "active", Updated: "2026-04-17"},
			{ID: newID(), Name: "edge-sao-pop1", Location: "São Paulo, BR", Provider: "Equinix Metal", CPUCores: 8, MemoryGB: 32, Workloads: 5, Uptime: "99.90%", Latency: "72 ms", LastContact: "2026-04-17 09:56", Status: "active", Updated: "2026-04-16"},
			{ID: newID(), Name: "edge-mum-pop1", Location: "Mumbai, IN", Provider: "Vultr Bare Metal", CPUCores: 8, MemoryGB: 32, Workloads: 4, Uptime: "99.85%", Latency: "95 ms", LastContact: "2026-04-17 09:54", Status: "active", Updated: "2026-04-15"},
		},
	},
}

func (s *edgeNodeStoreType) list(slug string) []EdgeNode {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.items[slug]
	out := make([]EdgeNode, len(src))
	copy(out, src)
	return out
}

func (s *edgeNodeStoreType) getByID(slug, id string) (EdgeNode, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, e := range s.items[slug] {
		if e.ID == id {
			return e, true
		}
	}
	return EdgeNode{}, false
}

func (s *edgeNodeStoreType) add(slug string, e EdgeNode) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[slug] = append(s.items[slug], e)
}

func (s *edgeNodeStoreType) delete(slug, id string) {
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

type apiEdgeNode struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Location    string `json:"location"`
	Provider    string `json:"provider"`
	CPUCores    int    `json:"cpu_cores"`
	MemoryGB    int    `json:"memory_gb"`
	Workloads   int    `json:"workloads"`
	Uptime      string `json:"uptime"`
	Latency     string `json:"latency"`
	LastContact string `json:"last_contact"`
	Status      string `json:"status"`
	Updated     string `json:"updated"`
}

type apiEdgeNodeList struct {
	Nodes []apiEdgeNode `json:"nodes"`
	Total int           `json:"total"`
}

func toAPIEdgeNode(e EdgeNode) apiEdgeNode {
	return apiEdgeNode{
		ID: e.ID, Name: e.Name, Location: e.Location,
		Provider: e.Provider, CPUCores: e.CPUCores, MemoryGB: e.MemoryGB,
		Workloads: e.Workloads, Uptime: e.Uptime, Latency: e.Latency,
		LastContact: e.LastContact, Status: e.Status, Updated: e.Updated,
	}
}

func filterEdgeNodes(all []EdgeNode, search, status string) []EdgeNode {
	var out []EdgeNode
	q := strings.ToLower(search)
	for _, e := range all {
		if status != "" && e.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(e.Name + " " + e.Location + " " + e.Provider)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, e)
	}
	return out
}

func apiWSMulticloudHybridEdgeListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	items := filterEdgeNodes(edgeNodeStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiEdgeNode, len(items))
	for i, e := range items {
		out[i] = toAPIEdgeNode(e)
	}
	apiOK(w, apiEdgeNodeList{Nodes: out, Total: len(out)})
}

type apiCreateEdgeNodeRequest struct {
	Name     string `json:"name"`
	Location string `json:"location"`
	Provider string `json:"provider"`
	CPUCores int    `json:"cpu_cores"`
	MemoryGB int    `json:"memory_gb"`
}

func apiWSMulticloudHybridEdgeCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateEdgeNodeRequest
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
	provider := strings.TrimSpace(req.Provider)
	if provider == "" {
		provider = "On-premises"
	}
	cpuCores := req.CPUCores
	if cpuCores <= 0 {
		cpuCores = 4
	}
	memoryGB := req.MemoryGB
	if memoryGB <= 0 {
		memoryGB = 16
	}
	node := EdgeNode{
		ID: newID(), Name: name, Location: location,
		Provider: provider, CPUCores: cpuCores, MemoryGB: memoryGB,
		Workloads: 0, Uptime: "—", Latency: "—",
		LastContact: "—", Status: "active", Updated: "2026-04-17",
	}
	edgeNodeStore.add(slug, node)
	apiCreated(w, toAPIEdgeNode(node))
}

func apiWSMulticloudHybridEdgeGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	node, found := edgeNodeStore.getByID(slug, r.PathValue("eid"))
	if !found {
		apiError(w, http.StatusNotFound, "Edge node not found.")
		return
	}
	apiOK(w, toAPIEdgeNode(node))
}

func apiWSMulticloudHybridEdgeDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := edgeNodeStore.getByID(slug, r.PathValue("eid")); !found {
		apiError(w, http.StatusNotFound, "Edge node not found.")
		return
	}
	edgeNodeStore.delete(slug, r.PathValue("eid"))
	apiOK(w, map[string]any{"ok": true, "message": "Edge node deleted."})
}
