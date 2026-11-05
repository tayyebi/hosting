package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Security API — Network Security: Web Application Firewall
// security/network/waf
// ════════════════════════════════════════════════════════════════

type WAFRule struct {
	ID        string
	Name      string
	RuleSet   string
	Action    string
	Priority  int
	Status    string
	Matches7d string
	Updated   string
}

type wafRuleStoreType struct {
	mu    sync.RWMutex
	rules map[string][]WAFRule
}

var wafRuleStore = &wafRuleStoreType{
	rules: map[string][]WAFRule{
		"buyvm": {
			{ID: newID(), Name: "OWASP Core Rule Set", RuleSet: "OWASP CRS 3.3", Action: "block", Priority: 1, Status: "active", Matches7d: "12.4k", Updated: "2025-11-20"},
			{ID: newID(), Name: "SQL Injection Filter", RuleSet: "Custom", Action: "block", Priority: 2, Status: "active", Matches7d: "3.8k", Updated: "2025-12-01"},
			{ID: newID(), Name: "XSS Protection", RuleSet: "Custom", Action: "block", Priority: 3, Status: "active", Matches7d: "2.1k", Updated: "2025-12-05"},
			{ID: newID(), Name: "Rate Limiter", RuleSet: "Custom", Action: "throttle", Priority: 4, Status: "active", Matches7d: "45.2k", Updated: "2026-01-10"},
			{ID: newID(), Name: "Geo-Block List", RuleSet: "Custom", Action: "block", Priority: 10, Status: "disabled", Matches7d: "0", Updated: "2025-10-15"},
		},
		"hosting": {
			{ID: newID(), Name: "OWASP Core Rule Set", RuleSet: "OWASP CRS 3.3", Action: "block", Priority: 1, Status: "active", Matches7d: "8.6k", Updated: "2025-11-18"},
			{ID: newID(), Name: "Bot Protection", RuleSet: "Managed", Action: "challenge", Priority: 2, Status: "active", Matches7d: "22.1k", Updated: "2026-01-05"},
			{ID: newID(), Name: "API Abuse Prevention", RuleSet: "Custom", Action: "throttle", Priority: 3, Status: "active", Matches7d: "5.4k", Updated: "2026-02-01"},
		},
	},
}

func (s *wafRuleStoreType) list(slug string) []WAFRule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.rules[slug]
	out := make([]WAFRule, len(src))
	copy(out, src)
	return out
}

func (s *wafRuleStoreType) getByID(slug, id string) (WAFRule, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, r := range s.rules[slug] {
		if r.ID == id {
			return r, true
		}
	}
	return WAFRule{}, false
}

func (s *wafRuleStoreType) add(slug string, r WAFRule) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rules[slug] = append(s.rules[slug], r)
}

func (s *wafRuleStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.rules[slug]
	for i, r := range list {
		if r.ID == id {
			s.rules[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiWAFRule struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	RuleSet   string `json:"rule_set"`
	Action    string `json:"action"`
	Priority  int    `json:"priority"`
	Status    string `json:"status"`
	Matches7d string `json:"matches_7d"`
	Updated   string `json:"updated"`
}

type apiWAFRuleList struct {
	Rules []apiWAFRule `json:"rules"`
	Total int          `json:"total"`
}

func toAPIWAFRule(r WAFRule) apiWAFRule {
	return apiWAFRule{
		ID: r.ID, Name: r.Name, RuleSet: r.RuleSet,
		Action: r.Action, Priority: r.Priority, Status: r.Status,
		Matches7d: r.Matches7d, Updated: r.Updated,
	}
}

func filterWAFRules(all []WAFRule, search, status string) []WAFRule {
	var out []WAFRule
	q := strings.ToLower(search)
	for _, r := range all {
		if status != "" && r.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(r.Name + " " + r.RuleSet + " " + r.Action)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, r)
	}
	return out
}

func apiWSSecNetWAFListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	rules := filterWAFRules(wafRuleStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiWAFRule, len(rules))
	for i, rl := range rules {
		out[i] = toAPIWAFRule(rl)
	}
	apiOK(w, apiWAFRuleList{Rules: out, Total: len(out)})
}

type apiCreateWAFRuleRequest struct {
	Name     string `json:"name"`
	RuleSet  string `json:"rule_set"`
	Action   string `json:"action"`
	Priority int    `json:"priority"`
}

func apiWSSecNetWAFCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateWAFRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	ruleSet := strings.TrimSpace(req.RuleSet)
	if ruleSet == "" {
		ruleSet = "Custom"
	}
	action := strings.TrimSpace(req.Action)
	if action == "" {
		action = "block"
	}
	priority := req.Priority
	if priority == 0 {
		priority = 100
	}
	rule := WAFRule{
		ID: newID(), Name: name, RuleSet: ruleSet,
		Action: action, Priority: priority, Status: "active",
		Matches7d: "0", Updated: "2026-04-17",
	}
	wafRuleStore.add(slug, rule)
	apiCreated(w, toAPIWAFRule(rule))
}

func apiWSSecNetWAFGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	rule, found := wafRuleStore.getByID(slug, r.PathValue("rid"))
	if !found {
		apiError(w, http.StatusNotFound, "WAF rule not found.")
		return
	}
	apiOK(w, toAPIWAFRule(rule))
}

func apiWSSecNetWAFDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := wafRuleStore.getByID(slug, r.PathValue("rid")); !found {
		apiError(w, http.StatusNotFound, "WAF rule not found.")
		return
	}
	wafRuleStore.delete(slug, r.PathValue("rid"))
	apiOK(w, map[string]any{"ok": true, "message": "WAF rule deleted."})
}
