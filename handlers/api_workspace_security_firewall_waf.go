package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Security API — Firewall: Cloud WAF (Pro Feature)
// security/firewall/cloud-waf
// ════════════════════════════════════════════════════════════════

type ProCloudWAFPolicy struct {
	ID               string
	Name             string
	Mode             string // "blocking", "detection", "learning"
	RulesetVersion   string
	CustomRules      int
	ManagedRules     int
	ProtectedDomains string
	RequestsBlocked  int64
	RequestsAllowed  int64
	ThreatScore      float64
	OWASPTop10       bool
	BotProtection    bool
	RateLimiting     bool
	RateLimit        int    // requests per second
	GeoBlocking      string // "none", "CN,RU,KP", etc.
	IPWhitelist      int
	IPBlacklist      int
	Region           string
	CostPerMonth     float64
	Status           string // "active", "learning", "disabled"
	Updated          string
}

type proCloudWAFStoreType struct {
	mu    sync.RWMutex
	items map[string][]ProCloudWAFPolicy
}

var proCloudWAFStore = &proCloudWAFStoreType{
	items: map[string][]ProCloudWAFPolicy{},
}

func (s *proCloudWAFStoreType) list(slug string) []ProCloudWAFPolicy {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.items[slug]
	out := make([]ProCloudWAFPolicy, len(src))
	copy(out, src)
	return out
}

func (s *proCloudWAFStoreType) getByID(slug, id string) (ProCloudWAFPolicy, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, p := range s.items[slug] {
		if p.ID == id {
			return p, true
		}
	}
	return ProCloudWAFPolicy{}, false
}

func (s *proCloudWAFStoreType) add(slug string, p ProCloudWAFPolicy) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[slug] = append(s.items[slug], p)
}

func (s *proCloudWAFStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.items[slug]
	for i, p := range list {
		if p.ID == id {
			s.items[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiProCloudWAFPolicy struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	Mode             string  `json:"mode"`
	RulesetVersion   string  `json:"ruleset_version"`
	CustomRules      int     `json:"custom_rules"`
	ManagedRules     int     `json:"managed_rules"`
	ProtectedDomains string  `json:"protected_domains"`
	RequestsBlocked  int64   `json:"requests_blocked"`
	RequestsAllowed  int64   `json:"requests_allowed"`
	ThreatScore      float64 `json:"threat_score"`
	OWASPTop10       bool    `json:"owasp_top10"`
	BotProtection    bool    `json:"bot_protection"`
	RateLimiting     bool    `json:"rate_limiting"`
	RateLimit        int     `json:"rate_limit_rps"`
	GeoBlocking      string  `json:"geo_blocking"`
	IPWhitelist      int     `json:"ip_whitelist_count"`
	IPBlacklist      int     `json:"ip_blacklist_count"`
	Region           string  `json:"region"`
	CostPerMonth     float64 `json:"cost_per_month"`
	Status           string  `json:"status"`
	Updated          string  `json:"updated"`
}

type apiProCloudWAFPolicyList struct {
	Policies []apiProCloudWAFPolicy `json:"policies"`
	Total    int                    `json:"total"`
}

func toAPIProCloudWAF(p ProCloudWAFPolicy) apiProCloudWAFPolicy {
	return apiProCloudWAFPolicy{
		ID: p.ID, Name: p.Name, Mode: p.Mode,
		RulesetVersion: p.RulesetVersion, CustomRules: p.CustomRules,
		ManagedRules: p.ManagedRules, ProtectedDomains: p.ProtectedDomains,
		RequestsBlocked: p.RequestsBlocked, RequestsAllowed: p.RequestsAllowed,
		ThreatScore: p.ThreatScore, OWASPTop10: p.OWASPTop10,
		BotProtection: p.BotProtection, RateLimiting: p.RateLimiting,
		RateLimit: p.RateLimit, GeoBlocking: p.GeoBlocking,
		IPWhitelist: p.IPWhitelist, IPBlacklist: p.IPBlacklist,
		Region: p.Region, CostPerMonth: p.CostPerMonth,
		Status: p.Status, Updated: p.Updated,
	}
}

func filterProCloudWAF(all []ProCloudWAFPolicy, search, status string) []ProCloudWAFPolicy {
	var out []ProCloudWAFPolicy
	q := strings.ToLower(search)
	for _, p := range all {
		if status != "" && p.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(p.Name + " " + p.Mode + " " + p.ProtectedDomains + " " + p.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, p)
	}
	return out
}

func apiWSSecFirewallWAFListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	items := filterProCloudWAF(proCloudWAFStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiProCloudWAFPolicy, len(items))
	for i, p := range items {
		out[i] = toAPIProCloudWAF(p)
	}
	apiOK(w, apiProCloudWAFPolicyList{Policies: out, Total: len(out)})
}

type apiCreateProCloudWAFRequest struct {
	Name             string `json:"name"`
	Mode             string `json:"mode"`
	ProtectedDomains string `json:"protected_domains"`
	OWASPTop10       bool   `json:"owasp_top10"`
	BotProtection    bool   `json:"bot_protection"`
	RateLimiting     bool   `json:"rate_limiting"`
	RateLimit        int    `json:"rate_limit_rps"`
	GeoBlocking      string `json:"geo_blocking"`
	Region           string `json:"region"`
}

func apiWSSecFirewallWAFCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateProCloudWAFRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	mode := strings.TrimSpace(req.Mode)
	if mode == "" {
		mode = "detection"
	}
	domains := strings.TrimSpace(req.ProtectedDomains)
	if domains == "" {
		domains = "*.example.com"
	}
	geoBlock := strings.TrimSpace(req.GeoBlocking)
	if geoBlock == "" {
		geoBlock = "none"
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "us-east-1"
	}
	rateLimit := req.RateLimit
	if req.RateLimiting && rateLimit <= 0 {
		rateLimit = 1000
	}
	pol := ProCloudWAFPolicy{
		ID: newID(), Name: name, Mode: mode,
		RulesetVersion: "CRS 4.2.0", CustomRules: 0, ManagedRules: 142,
		ProtectedDomains: domains, RequestsBlocked: 0, RequestsAllowed: 0,
		ThreatScore: 0, OWASPTop10: req.OWASPTop10,
		BotProtection: req.BotProtection, RateLimiting: req.RateLimiting,
		RateLimit: rateLimit, GeoBlocking: geoBlock,
		IPWhitelist: 0, IPBlacklist: 0,
		Region: region, CostPerMonth: 45.00,
		Status: "learning", Updated: "2026-04-18",
	}
	proCloudWAFStore.add(slug, pol)
	apiCreated(w, toAPIProCloudWAF(pol))
}

func apiWSSecFirewallWAFGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	p, found := proCloudWAFStore.getByID(slug, r.PathValue("wid"))
	if !found {
		apiError(w, http.StatusNotFound, "WAF policy not found.")
		return
	}
	apiOK(w, toAPIProCloudWAF(p))
}

func apiWSSecFirewallWAFDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := proCloudWAFStore.getByID(slug, r.PathValue("wid")); !found {
		apiError(w, http.StatusNotFound, "WAF policy not found.")
		return
	}
	proCloudWAFStore.delete(slug, r.PathValue("wid"))
	apiOK(w, map[string]any{"ok": true, "message": "WAF policy deleted."})
}
