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
	ID               string
	Name             string
	Protocol         string
	ListenerPorts    string
	TargetCount      int
	HealthyTargets   int
	RPS              int
	ActiveConns      int
	BandwidthGbps    float64
	TLSVersion       string
	CertDomain       string
	CertExpiry       string
	WAFEnabled       bool
	AccessLogs       bool
	StickySession    bool
	IdleTimeoutSec   int
	Region           string
	CostPerMonth     float64
	Status           string
	Updated          string
}

type proAppLBStoreType struct {
	mu    sync.RWMutex
	items map[string][]ProAppLoadBalancer
}

var proAppLBStore = &proAppLBStoreType{
	items: map[string][]ProAppLoadBalancer{
		"buyvm": {
			{ID: newID(), Name: "api-gateway-prod", Protocol: "HTTPS (TLS 1.3)", ListenerPorts: "443, 8443", TargetCount: 6, HealthyTargets: 6, RPS: 8420, ActiveConns: 4200, BandwidthGbps: 2.4, TLSVersion: "TLS 1.3", CertDomain: "*.api.company.io", CertExpiry: "2027-01-15", WAFEnabled: true, AccessLogs: true, StickySession: false, IdleTimeoutSec: 60, Region: "us-east-1", CostPerMonth: 85.00, Status: "active", Updated: "2026-04-18"},
			{ID: newID(), Name: "web-frontend-prod", Protocol: "HTTPS (TLS 1.3)", ListenerPorts: "443", TargetCount: 4, HealthyTargets: 4, RPS: 6100, ActiveConns: 12400, BandwidthGbps: 4.8, TLSVersion: "TLS 1.3", CertDomain: "*.company.io", CertExpiry: "2026-11-20", WAFEnabled: true, AccessLogs: true, StickySession: true, IdleTimeoutSec: 120, Region: "us-east-1", CostPerMonth: 72.00, Status: "active", Updated: "2026-04-18"},
			{ID: newID(), Name: "grpc-services", Protocol: "gRPC + TLS", ListenerPorts: "443", TargetCount: 8, HealthyTargets: 8, RPS: 2800, ActiveConns: 1600, BandwidthGbps: 0.8, TLSVersion: "TLS 1.3", CertDomain: "grpc.company.io", CertExpiry: "2027-03-10", WAFEnabled: false, AccessLogs: true, StickySession: false, IdleTimeoutSec: 300, Region: "us-east-1", CostPerMonth: 65.00, Status: "active", Updated: "2026-04-17"},
			{ID: newID(), Name: "websocket-lb", Protocol: "WSS", ListenerPorts: "443, 8080", TargetCount: 3, HealthyTargets: 3, RPS: 640, ActiveConns: 8200, BandwidthGbps: 0.4, TLSVersion: "TLS 1.3", CertDomain: "ws.company.io", CertExpiry: "2026-07-18", WAFEnabled: false, AccessLogs: false, StickySession: true, IdleTimeoutSec: 3600, Region: "us-east-1", CostPerMonth: 45.00, Status: "active", Updated: "2026-04-16"},
			{ID: newID(), Name: "staging-lb", Protocol: "HTTPS", ListenerPorts: "443", TargetCount: 4, HealthyTargets: 2, RPS: 120, ActiveConns: 42, BandwidthGbps: 0.01, TLSVersion: "TLS 1.2", CertDomain: "staging.company.io", CertExpiry: "2026-04-01", WAFEnabled: false, AccessLogs: false, StickySession: false, IdleTimeoutSec: 60, Region: "eu-west-1", CostPerMonth: 32.00, Status: "degraded", Updated: "2026-04-15"},
		},
		"hosting": {
			{ID: newID(), Name: "platform-ingress", Protocol: "HTTPS (TLS 1.3)", ListenerPorts: "443, 8443, 9443", TargetCount: 12, HealthyTargets: 12, RPS: 24600, ActiveConns: 18400, BandwidthGbps: 8.2, TLSVersion: "TLS 1.3", CertDomain: "*.platform.hosting.io", CertExpiry: "2027-02-28", WAFEnabled: true, AccessLogs: true, StickySession: false, IdleTimeoutSec: 60, Region: "eu-central-1", CostPerMonth: 180.00, Status: "active", Updated: "2026-04-18"},
			{ID: newID(), Name: "api-v2-gateway", Protocol: "HTTPS (TLS 1.3)", ListenerPorts: "443", TargetCount: 8, HealthyTargets: 8, RPS: 14200, ActiveConns: 6800, BandwidthGbps: 3.6, TLSVersion: "TLS 1.3", CertDomain: "*.api.hosting.io", CertExpiry: "2027-01-05", WAFEnabled: true, AccessLogs: true, StickySession: false, IdleTimeoutSec: 60, Region: "eu-central-1", CostPerMonth: 120.00, Status: "active", Updated: "2026-04-18"},
			{ID: newID(), Name: "internal-services", Protocol: "HTTP/2", ListenerPorts: "8080, 8081", TargetCount: 16, HealthyTargets: 15, RPS: 8400, ActiveConns: 3200, BandwidthGbps: 1.2, TLSVersion: "mTLS", CertDomain: "internal.hosting.local", CertExpiry: "2026-12-31", WAFEnabled: false, AccessLogs: true, StickySession: false, IdleTimeoutSec: 30, Region: "eu-central-1", CostPerMonth: 95.00, Status: "active", Updated: "2026-04-17"},
			{ID: newID(), Name: "edge-us-west", Protocol: "HTTPS (TLS 1.3)", ListenerPorts: "443", TargetCount: 4, HealthyTargets: 4, RPS: 3600, ActiveConns: 2400, BandwidthGbps: 1.4, TLSVersion: "TLS 1.3", CertDomain: "us.hosting.io", CertExpiry: "2027-04-10", WAFEnabled: true, AccessLogs: true, StickySession: false, IdleTimeoutSec: 60, Region: "us-west-2", CostPerMonth: 65.00, Status: "active", Updated: "2026-04-16"},
		},
	},
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
	Total         int                    `json:"total"`
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
