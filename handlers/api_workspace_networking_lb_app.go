package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Networking API — Load Balancer: Application (Layer 7)
// networking/lb/app
// ════════════════════════════════════════════════════════════════

type NetLBApp struct {
	ID          string
	Name        string
	Scheme      string
	Protocol    string
	Port        int
	Region      string
	Status      string
	Targets     int
	HealthCheck string
	SSLARN      string
	WAF         bool
	RPS         int
}

type netLBAppStoreType struct {
	mu        sync.RWMutex
	instances map[string][]NetLBApp
}

var netLBAppStore = &netLBAppStoreType{
	instances: map[string][]NetLBApp{},
}

func (s *netLBAppStoreType) list(slug string) []NetLBApp {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.instances[slug]
	out := make([]NetLBApp, len(src))
	copy(out, src)
	return out
}

func (s *netLBAppStoreType) getByID(slug, id string) (NetLBApp, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, d := range s.instances[slug] {
		if d.ID == id {
			return d, true
		}
	}
	return NetLBApp{}, false
}

func (s *netLBAppStoreType) add(slug string, d NetLBApp) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.instances[slug] = append(s.instances[slug], d)
}

func (s *netLBAppStoreType) delete(slug, id string) {
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

type apiNetLBApp struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Scheme      string `json:"scheme"`
	Protocol    string `json:"protocol"`
	Port        int    `json:"port"`
	Region      string `json:"region"`
	Status      string `json:"status"`
	Targets     int    `json:"targets"`
	HealthCheck string `json:"health_check"`
	SSLARN      string `json:"ssl_arn"`
	WAF         bool   `json:"waf"`
	RPS         int    `json:"rps"`
}

type apiNetLBAppList struct {
	LoadBalancers []apiNetLBApp `json:"load_balancers"`
	Total         int           `json:"total"`
}

func toAPINetLBApp(d NetLBApp) apiNetLBApp {
	return apiNetLBApp{
		ID: d.ID, Name: d.Name, Scheme: d.Scheme,
		Protocol: d.Protocol, Port: d.Port,
		Region: d.Region, Status: d.Status,
		Targets: d.Targets, HealthCheck: d.HealthCheck,
		SSLARN: d.SSLARN, WAF: d.WAF, RPS: d.RPS,
	}
}

func filterNetLBApp(all []NetLBApp, search, status string) []NetLBApp {
	var out []NetLBApp
	q := strings.ToLower(search)
	for _, d := range all {
		if status != "" && d.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(d.Name + " " + d.Scheme + " " + d.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, d)
	}
	return out
}

func apiWSNetLBAppListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	items := filterNetLBApp(netLBAppStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiNetLBApp, len(items))
	for i, d := range items {
		out[i] = toAPINetLBApp(d)
	}
	apiOK(w, apiNetLBAppList{LoadBalancers: out, Total: len(out)})
}

type apiCreateNetLBAppRequest struct {
	Name        string `json:"name"`
	Scheme      string `json:"scheme"`
	Protocol    string `json:"protocol"`
	Port        int    `json:"port"`
	Region      string `json:"region"`
	HealthCheck string `json:"health_check"`
	SSLARN      string `json:"ssl_arn"`
	WAF         bool   `json:"waf"`
}

func apiWSNetLBAppCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateNetLBAppRequest
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
		protocol = "HTTPS"
	}
	port := req.Port
	if port == 0 {
		port = 443
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "eu-central-1"
	}
	hc := strings.TrimSpace(req.HealthCheck)
	if hc == "" {
		hc = "/health"
	}
	inst := NetLBApp{
		ID: newID(), Name: name, Scheme: scheme,
		Protocol: protocol, Port: port,
		Region: region, Status: "provisioning",
		Targets: 0, HealthCheck: hc,
		SSLARN: strings.TrimSpace(req.SSLARN),
		WAF:    req.WAF, RPS: 0,
	}
	netLBAppStore.add(slug, inst)
	apiCreated(w, toAPINetLBApp(inst))
}

func apiWSNetLBAppGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	inst, found := netLBAppStore.getByID(slug, r.PathValue("lid"))
	if !found {
		apiError(w, http.StatusNotFound, "Application load balancer not found.")
		return
	}
	apiOK(w, toAPINetLBApp(inst))
}

func apiWSNetLBAppDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := netLBAppStore.getByID(slug, r.PathValue("lid")); !found {
		apiError(w, http.StatusNotFound, "Application load balancer not found.")
		return
	}
	netLBAppStore.delete(slug, r.PathValue("lid"))
	apiOK(w, map[string]any{"ok": true, "message": "Application load balancer deleted."})
}
