package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Security API — Network Security: Intrusion Detection
// security/network/ids
// ════════════════════════════════════════════════════════════════

type IDSRule struct {
	ID        string
	Name      string
	Category  string
	Severity  string
	Action    string
	Status    string
	Alerts7d  string
	Updated   string
}

type idsRuleStoreType struct {
	mu    sync.RWMutex
	rules map[string][]IDSRule
}

var idsRuleStore = &idsRuleStoreType{
	rules: map[string][]IDSRule{
		"buyvm": {
			{ID: newID(), Name: "Port Scan Detection", Category: "Reconnaissance", Severity: "medium", Action: "alert", Status: "active", Alerts7d: "84", Updated: "2026-01-12"},
			{ID: newID(), Name: "Brute Force SSH", Category: "Credential Access", Severity: "high", Action: "alert+block", Status: "active", Alerts7d: "23", Updated: "2026-01-15"},
			{ID: newID(), Name: "Lateral Movement Detector", Category: "Lateral Movement", Severity: "critical", Action: "alert+block", Status: "active", Alerts7d: "2", Updated: "2026-02-01"},
			{ID: newID(), Name: "DNS Tunneling", Category: "Exfiltration", Severity: "high", Action: "alert", Status: "active", Alerts7d: "7", Updated: "2025-12-20"},
			{ID: newID(), Name: "Malware C2 Beacon", Category: "Command & Control", Severity: "critical", Action: "alert+block", Status: "active", Alerts7d: "0", Updated: "2026-01-28"},
		},
		"hosting": {
			{ID: newID(), Name: "Port Scan Detection", Category: "Reconnaissance", Severity: "medium", Action: "alert", Status: "active", Alerts7d: "42", Updated: "2026-01-10"},
			{ID: newID(), Name: "SQL Injection Attempt", Category: "Initial Access", Severity: "high", Action: "alert+block", Status: "active", Alerts7d: "156", Updated: "2026-02-05"},
			{ID: newID(), Name: "Privilege Escalation", Category: "Privilege Escalation", Severity: "critical", Action: "alert+block", Status: "active", Alerts7d: "3", Updated: "2026-01-22"},
		},
	},
}

func (s *idsRuleStoreType) list(slug string) []IDSRule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.rules[slug]
	out := make([]IDSRule, len(src))
	copy(out, src)
	return out
}

func (s *idsRuleStoreType) getByID(slug, id string) (IDSRule, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, r := range s.rules[slug] {
		if r.ID == id {
			return r, true
		}
	}
	return IDSRule{}, false
}

func (s *idsRuleStoreType) add(slug string, r IDSRule) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rules[slug] = append(s.rules[slug], r)
}

func (s *idsRuleStoreType) delete(slug, id string) {
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

type apiIDSRule struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
	Severity string `json:"severity"`
	Action   string `json:"action"`
	Status   string `json:"status"`
	Alerts7d string `json:"alerts_7d"`
	Updated  string `json:"updated"`
}

type apiIDSRuleList struct {
	Rules []apiIDSRule `json:"rules"`
	Total int          `json:"total"`
}

func toAPIIDSRule(r IDSRule) apiIDSRule {
	return apiIDSRule{
		ID: r.ID, Name: r.Name, Category: r.Category,
		Severity: r.Severity, Action: r.Action, Status: r.Status,
		Alerts7d: r.Alerts7d, Updated: r.Updated,
	}
}

func filterIDSRules(all []IDSRule, search, status string) []IDSRule {
	var out []IDSRule
	q := strings.ToLower(search)
	for _, r := range all {
		if status != "" && r.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(r.Name + " " + r.Category + " " + r.Severity)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, r)
	}
	return out
}

func apiWSSecNetIDSListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	rules := filterIDSRules(idsRuleStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiIDSRule, len(rules))
	for i, rl := range rules {
		out[i] = toAPIIDSRule(rl)
	}
	apiOK(w, apiIDSRuleList{Rules: out, Total: len(out)})
}

type apiCreateIDSRuleRequest struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Severity string `json:"severity"`
	Action   string `json:"action"`
}

func apiWSSecNetIDSCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateIDSRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	category := strings.TrimSpace(req.Category)
	if category == "" {
		category = "General"
	}
	severity := strings.TrimSpace(req.Severity)
	if severity == "" {
		severity = "medium"
	}
	action := strings.TrimSpace(req.Action)
	if action == "" {
		action = "alert"
	}
	rule := IDSRule{
		ID: newID(), Name: name, Category: category,
		Severity: severity, Action: action, Status: "active",
		Alerts7d: "0", Updated: "2026-04-17",
	}
	idsRuleStore.add(slug, rule)
	apiCreated(w, toAPIIDSRule(rule))
}

func apiWSSecNetIDSGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	rule, found := idsRuleStore.getByID(slug, r.PathValue("iid"))
	if !found {
		apiError(w, http.StatusNotFound, "IDS rule not found.")
		return
	}
	apiOK(w, toAPIIDSRule(rule))
}

func apiWSSecNetIDSDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := idsRuleStore.getByID(slug, r.PathValue("iid")); !found {
		apiError(w, http.StatusNotFound, "IDS rule not found.")
		return
	}
	idsRuleStore.delete(slug, r.PathValue("iid"))
	apiOK(w, map[string]any{"ok": true, "message": "IDS rule deleted."})
}
