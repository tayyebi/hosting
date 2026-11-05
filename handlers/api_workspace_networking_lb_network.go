package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Networking API — Load Balancer: Network (Layer 4)
// networking/lb/network
// ════════════════════════════════════════════════════════════════

type NetLBNetwork struct {
	ID         string
	Name       string
	Scheme     string
	Protocol   string
	Port       int
	Region     string
	Status     string
	Targets    int
	CrossZone  bool
	StaticIPs  int
	Throughput string
	PPS        int
}

type netLBNetworkStoreType struct {
	mu        sync.RWMutex
	instances map[string][]NetLBNetwork
}

var netLBNetworkStore = &netLBNetworkStoreType{
	instances: map[string][]NetLBNetwork{
		"buyvm": {
			{ID: newID(), Name: "prod-tcp-nlb", Scheme: "internet-facing", Protocol: "TCP", Port: 443, Region: "eu-central-1", Status: "active", Targets: 8, CrossZone: true, StaticIPs: 3, Throughput: "10 Gbps", PPS: 1200000},
			{ID: newID(), Name: "prod-udp-nlb", Scheme: "internal", Protocol: "UDP", Port: 5060, Region: "eu-central-1", Status: "active", Targets: 4, CrossZone: false, StaticIPs: 1, Throughput: "5 Gbps", PPS: 480000},
			{ID: newID(), Name: "us-tcp-nlb", Scheme: "internet-facing", Protocol: "TCP", Port: 443, Region: "us-east-1", Status: "active", Targets: 6, CrossZone: true, StaticIPs: 2, Throughput: "10 Gbps", PPS: 960000},
		},
		"hosting": {
			{ID: newID(), Name: "main-nlb", Scheme: "internet-facing", Protocol: "TCP", Port: 443, Region: "ap-south-1", Status: "active", Targets: 10, CrossZone: true, StaticIPs: 3, Throughput: "20 Gbps", PPS: 2400000},
			{ID: newID(), Name: "backend-nlb", Scheme: "internal", Protocol: "TCP", Port: 9090, Region: "ap-south-1", Status: "active", Targets: 6, CrossZone: false, StaticIPs: 1, Throughput: "5 Gbps", PPS: 600000},
		},
	},
}

func (s *netLBNetworkStoreType) list(slug string) []NetLBNetwork {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.instances[slug]
	out := make([]NetLBNetwork, len(src))
	copy(out, src)
	return out
}

func (s *netLBNetworkStoreType) getByID(slug, id string) (NetLBNetwork, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, d := range s.instances[slug] {
		if d.ID == id {
			return d, true
		}
	}
	return NetLBNetwork{}, false
}

func (s *netLBNetworkStoreType) add(slug string, d NetLBNetwork) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.instances[slug] = append(s.instances[slug], d)
}

func (s *netLBNetworkStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.instances[slug]
	for i, d := range list {
		if d.ID == id {
			s.instances[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiNetLBNetwork struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Scheme     string `json:"scheme"`
	Protocol   string `json:"protocol"`
	Port       int    `json:"port"`
	Region     string `json:"region"`
	Status     string `json:"status"`
	Targets    int    `json:"targets"`
	CrossZone  bool   `json:"cross_zone"`
	StaticIPs  int    `json:"static_ips"`
	Throughput string `json:"throughput"`
	PPS        int    `json:"pps"`
}

type apiNetLBNetworkList struct {
	LoadBalancers []apiNetLBNetwork `json:"load_balancers"`
	Total         int               `json:"total"`
}

func toAPINetLBNetwork(d NetLBNetwork) apiNetLBNetwork {
	return apiNetLBNetwork{
		ID: d.ID, Name: d.Name, Scheme: d.Scheme,
		Protocol: d.Protocol, Port: d.Port,
		Region: d.Region, Status: d.Status,
		Targets: d.Targets, CrossZone: d.CrossZone,
		StaticIPs: d.StaticIPs, Throughput: d.Throughput,
		PPS: d.PPS,
	}
}

func filterNetLBNetwork(all []NetLBNetwork, search, status string) []NetLBNetwork {
	var out []NetLBNetwork
	q := strings.ToLower(search)
	for _, d := range all {
		if status != "" && d.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(d.Name + " " + d.Scheme + " " + d.Protocol + " " + d.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, d)
	}
	return out
}

func apiWSNetLBNetworkListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	items := filterNetLBNetwork(netLBNetworkStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiNetLBNetwork, len(items))
	for i, d := range items {
		out[i] = toAPINetLBNetwork(d)
	}
	apiOK(w, apiNetLBNetworkList{LoadBalancers: out, Total: len(out)})
}

type apiCreateNetLBNetworkRequest struct {
	Name      string `json:"name"`
	Scheme    string `json:"scheme"`
	Protocol  string `json:"protocol"`
	Port      int    `json:"port"`
	Region    string `json:"region"`
	CrossZone bool   `json:"cross_zone"`
}

func apiWSNetLBNetworkCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateNetLBNetworkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	scheme := strings.TrimSpace(req.Scheme)
	if scheme == "" {
		scheme = "internet-facing"
	}
	protocol := strings.TrimSpace(req.Protocol)
	if protocol == "" {
		protocol = "TCP"
	}
	port := req.Port
	if port == 0 {
		port = 443
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "eu-central-1"
	}
	inst := NetLBNetwork{
		ID: newID(), Name: name, Scheme: scheme,
		Protocol: protocol, Port: port,
		Region: region, Status: "provisioning",
		Targets: 0, CrossZone: req.CrossZone,
		StaticIPs: 1, Throughput: "0 Gbps", PPS: 0,
	}
	netLBNetworkStore.add(slug, inst)
	apiCreated(w, toAPINetLBNetwork(inst))
}

func apiWSNetLBNetworkGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	inst, found := netLBNetworkStore.getByID(slug, r.PathValue("lid"))
	if !found {
		apiError(w, http.StatusNotFound, "Network load balancer not found.")
		return
	}
	apiOK(w, toAPINetLBNetwork(inst))
}

func apiWSNetLBNetworkDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := netLBNetworkStore.getByID(slug, r.PathValue("lid")); !found {
		apiError(w, http.StatusNotFound, "Network load balancer not found.")
		return
	}
	netLBNetworkStore.delete(slug, r.PathValue("lid"))
	apiOK(w, map[string]any{"ok": true, "message": "Network load balancer deleted."})
}
