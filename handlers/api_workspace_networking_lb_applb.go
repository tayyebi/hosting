package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Networking API — Load Balancer: Application LB (Pro Feature)
// networking/load-balancer/application-lb
// ════════════════════════════════════════════════════════════════

type ProAppLoadBalancer struct {
	ID             string
	Name           string
	Protocol       string
	ListenerPorts  string
	TargetCount    int
	HealthyTargets int
	RPS            int
	ActiveConns    int
	BandwidthGbps  float64
	TLSVersion     string
	CertDomain     string
	CertExpiry     string
	WAFEnabled     bool
	AccessLogs     bool
	StickySession  bool
	IdleTimeoutSec int
	Region         string
	CostPerMonth   float64
	Status         string
	Updated        string
}

type proAppLBStoreType struct {
	mu    sync.RWMutex
	items map[string][]ProAppLoadBalancer
}

var proAppLBStore = &proAppLBStoreType{
	items: map[string][]ProAppLoadBalancer{},
}

func (s *proAppLBStoreType) list(slug string) []ProAppLoadBalancer {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.items[slug]
	out := make([]ProAppLoadBalancer, len(src))
	copy(out, src)
	return out
}

func (s *proAppLBStoreType) getByID(slug, id string) (ProAppLoadBalancer, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, lb := range s.items[slug] {
		if lb.ID == id {
			return lb, true
		}
	}
	return ProAppLoadBalancer{}, false
}

func (s *proAppLBStoreType) add(slug string, lb ProAppLoadBalancer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[slug] = append(s.items[slug], lb)
}

func (s *proAppLBStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.items[slug]
	for i, lb := range list {
		if lb.ID == id {
			s.items[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiProAppLoadBalancer struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Protocol       string  `json:"protocol"`
	ListenerPorts  string  `json:"listener_ports"`
	TargetCount    int     `json:"target_count"`
	HealthyTargets int     `json:"healthy_targets"`
	RPS            int     `json:"rps"`
	ActiveConns    int     `json:"active_connections"`
	BandwidthGbps  float64 `json:"bandwidth_gbps"`
	TLSVersion     string  `json:"tls_version"`
	CertDomain     string  `json:"cert_domain"`
	CertExpiry     string  `json:"cert_expiry"`
	WAFEnabled     bool    `json:"waf_enabled"`
	AccessLogs     bool    `json:"access_logs"`
	StickySession  bool    `json:"sticky_session"`
	IdleTimeoutSec int     `json:"idle_timeout_sec"`
	Region         string  `json:"region"`
	CostPerMonth   float64 `json:"cost_per_month"`
	Status         string  `json:"status"`
	Updated        string  `json:"updated"`
}

type apiProAppLBList struct {
	LoadBalancers []apiProAppLoadBalancer `json:"load_balancers"`
	Total         int                     `json:"total"`
}

func toAPIProAppLB(lb ProAppLoadBalancer) apiProAppLoadBalancer {
	return apiProAppLoadBalancer{
		ID: lb.ID, Name: lb.Name, Protocol: lb.Protocol,
		ListenerPorts: lb.ListenerPorts, TargetCount: lb.TargetCount,
		HealthyTargets: lb.HealthyTargets, RPS: lb.RPS,
		ActiveConns: lb.ActiveConns, BandwidthGbps: lb.BandwidthGbps,
		TLSVersion: lb.TLSVersion, CertDomain: lb.CertDomain,
		CertExpiry: lb.CertExpiry, WAFEnabled: lb.WAFEnabled,
		AccessLogs: lb.AccessLogs, StickySession: lb.StickySession,
		IdleTimeoutSec: lb.IdleTimeoutSec, Region: lb.Region,
		CostPerMonth: lb.CostPerMonth, Status: lb.Status,
		Updated: lb.Updated,
	}
}

func filterProAppLBs(all []ProAppLoadBalancer, search, status string) []ProAppLoadBalancer {
	var out []ProAppLoadBalancer
	q := strings.ToLower(search)
	for _, lb := range all {
		if status != "" && lb.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(lb.Name + " " + lb.Protocol + " " + lb.CertDomain + " " + lb.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, lb)
	}
	return out
}

func apiWSNetLBAppLBListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	items := filterProAppLBs(proAppLBStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiProAppLoadBalancer, len(items))
	for i, lb := range items {
		out[i] = toAPIProAppLB(lb)
	}
	apiOK(w, apiProAppLBList{LoadBalancers: out, Total: len(out)})
}

type apiCreateProAppLBRequest struct {
	Name          string `json:"name"`
	Protocol      string `json:"protocol"`
	ListenerPorts string `json:"listener_ports"`
	WAFEnabled    bool   `json:"waf_enabled"`
	AccessLogs    bool   `json:"access_logs"`
	StickySession bool   `json:"sticky_session"`
	Region        string `json:"region"`
}

func apiWSNetLBAppLBCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateProAppLBRequest
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
		protocol = "HTTPS (TLS 1.3)"
	}
	ports := strings.TrimSpace(req.ListenerPorts)
	if ports == "" {
		ports = "443"
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "us-east-1"
	}
	lb := ProAppLoadBalancer{
		ID: newID(), Name: name, Protocol: protocol,
		ListenerPorts: ports, TargetCount: 0, HealthyTargets: 0,
		RPS: 0, ActiveConns: 0, BandwidthGbps: 0,
		TLSVersion: "TLS 1.3", CertDomain: "—", CertExpiry: "—",
		WAFEnabled: req.WAFEnabled, AccessLogs: req.AccessLogs,
		StickySession: req.StickySession, IdleTimeoutSec: 60,
		Region: region, CostPerMonth: 32.00,
		Status: "provisioning", Updated: "2026-04-18",
	}
	proAppLBStore.add(slug, lb)
	apiCreated(w, toAPIProAppLB(lb))
}

func apiWSNetLBAppLBGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	lb, found := proAppLBStore.getByID(slug, r.PathValue("lid"))
	if !found {
		apiError(w, http.StatusNotFound, "Application load balancer not found.")
		return
	}
	apiOK(w, toAPIProAppLB(lb))
}

func apiWSNetLBAppLBDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := proAppLBStore.getByID(slug, r.PathValue("lid")); !found {
		apiError(w, http.StatusNotFound, "Application load balancer not found.")
		return
	}
	proAppLBStore.delete(slug, r.PathValue("lid"))
	apiOK(w, map[string]any{"ok": true, "message": "Application load balancer deleted."})
}
