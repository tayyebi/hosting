package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Networking API — Load Balancer: Global
// networking/lb/global
// ════════════════════════════════════════════════════════════════

type NetLBGlobal struct {
	ID        string
	Name      string
	Protocol  string
	Anycast   string
	Regions   int
	Backends  int
	Status    string
	Routing   string
	Failover  string
	HealthChk string
	RPS       int
	Latency   string
}

type netLBGlobalStoreType struct {
	mu        sync.RWMutex
	instances map[string][]NetLBGlobal
}

var netLBGlobalStore = &netLBGlobalStoreType{
	instances: map[string][]NetLBGlobal{},
}

func (s *netLBGlobalStoreType) list(slug string) []NetLBGlobal {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.instances[slug]
	out := make([]NetLBGlobal, len(src))
	copy(out, src)
	return out
}

func (s *netLBGlobalStoreType) getByID(slug, id string) (NetLBGlobal, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, d := range s.instances[slug] {
		if d.ID == id {
			return d, true
		}
	}
	return NetLBGlobal{}, false
}

func (s *netLBGlobalStoreType) add(slug string, d NetLBGlobal) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.instances[slug] = append(s.instances[slug], d)
}

func (s *netLBGlobalStoreType) delete(slug, id string) {
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

type apiNetLBGlobal struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Protocol  string `json:"protocol"`
	Anycast   string `json:"anycast_ip"`
	Regions   int    `json:"regions"`
	Backends  int    `json:"backends"`
	Status    string `json:"status"`
	Routing   string `json:"routing"`
	Failover  string `json:"failover"`
	HealthChk string `json:"health_check"`
	RPS       int    `json:"rps"`
	Latency   string `json:"latency"`
}

type apiNetLBGlobalList struct {
	LoadBalancers []apiNetLBGlobal `json:"load_balancers"`
	Total         int              `json:"total"`
}

func toAPINetLBGlobal(d NetLBGlobal) apiNetLBGlobal {
	return apiNetLBGlobal{
		ID: d.ID, Name: d.Name, Protocol: d.Protocol,
		Anycast: d.Anycast, Regions: d.Regions,
		Backends: d.Backends, Status: d.Status,
		Routing: d.Routing, Failover: d.Failover,
		HealthChk: d.HealthChk, RPS: d.RPS,
		Latency: d.Latency,
	}
}

func filterNetLBGlobal(all []NetLBGlobal, search, status string) []NetLBGlobal {
	var out []NetLBGlobal
	q := strings.ToLower(search)
	for _, d := range all {
		if status != "" && d.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(d.Name + " " + d.Protocol + " " + d.Routing)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, d)
	}
	return out
}

func apiWSNetLBGlobalListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	items := filterNetLBGlobal(netLBGlobalStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiNetLBGlobal, len(items))
	for i, d := range items {
		out[i] = toAPINetLBGlobal(d)
	}
	apiOK(w, apiNetLBGlobalList{LoadBalancers: out, Total: len(out)})
}

type apiCreateNetLBGlobalRequest struct {
	Name      string `json:"name"`
	Protocol  string `json:"protocol"`
	Routing   string `json:"routing"`
	Failover  string `json:"failover"`
	HealthChk string `json:"health_check"`
}

func apiWSNetLBGlobalCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateNetLBGlobalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	protocol := strings.TrimSpace(req.Protocol)
	if protocol == "" {
		protocol = "HTTPS"
	}
	routing := strings.TrimSpace(req.Routing)
	if routing == "" {
		routing = "latency-based"
	}
	failover := strings.TrimSpace(req.Failover)
	if failover == "" {
		failover = "automatic"
	}
	hc := strings.TrimSpace(req.HealthChk)
	if hc == "" {
		hc = "/health"
	}
	inst := NetLBGlobal{
		ID: newID(), Name: name, Protocol: protocol,
		Anycast: "0.0.0.0", Regions: 0, Backends: 0,
		Status: "provisioning", Routing: routing,
		Failover: failover, HealthChk: hc,
		RPS: 0, Latency: "0ms",
	}
	netLBGlobalStore.add(slug, inst)
	apiCreated(w, toAPINetLBGlobal(inst))
}

func apiWSNetLBGlobalGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	inst, found := netLBGlobalStore.getByID(slug, r.PathValue("gid"))
	if !found {
		apiError(w, http.StatusNotFound, "Global load balancer not found.")
		return
	}
	apiOK(w, toAPINetLBGlobal(inst))
}

func apiWSNetLBGlobalDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := netLBGlobalStore.getByID(slug, r.PathValue("gid")); !found {
		apiError(w, http.StatusNotFound, "Global load balancer not found.")
		return
	}
	netLBGlobalStore.delete(slug, r.PathValue("gid"))
	apiOK(w, map[string]any{"ok": true, "message": "Global load balancer deleted."})
}
