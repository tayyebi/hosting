package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Multi-Cloud & Hybrid API — Management: Policy Governance
// multicloud/mgmt/policy
// ════════════════════════════════════════════════════════════════

type GovernancePolicy struct {
	ID           string
	Name         string
	Category     string
	Scope        string
	Enforcement  string
	Compliant    int
	NonCompliant int
	Exemptions   int
	Severity     string
	LastEval     string
	Status       string
	Updated      string
}

type governancePolicyStoreType struct {
	mu    sync.RWMutex
	items map[string][]GovernancePolicy
}

var governancePolicyStore = &governancePolicyStoreType{
	items: map[string][]GovernancePolicy{
		"buyvm": {
			{ID: newID(), Name: "enforce-encryption-at-rest", Category: "Security", Scope: "All accounts", Enforcement: "Deny", Compliant: 142, NonCompliant: 3, Exemptions: 1, Severity: "High", LastEval: "2026-04-17 09:00", Status: "active", Updated: "2026-04-17"},
			{ID: newID(), Name: "require-tagging-standards", Category: "Cost Management", Scope: "Production", Enforcement: "Audit", Compliant: 890, NonCompliant: 45, Exemptions: 12, Severity: "Medium", LastEval: "2026-04-17 09:00", Status: "active", Updated: "2026-04-17"},
			{ID: newID(), Name: "restrict-public-endpoints", Category: "Networking", Scope: "All accounts", Enforcement: "Deny", Compliant: 78, NonCompliant: 0, Exemptions: 2, Severity: "Critical", LastEval: "2026-04-17 09:00", Status: "active", Updated: "2026-04-16"},
			{ID: newID(), Name: "allowed-vm-sizes", Category: "Compute", Scope: "Dev / Staging", Enforcement: "Deny", Compliant: 56, NonCompliant: 2, Exemptions: 0, Severity: "Low", LastEval: "2026-04-17 09:00", Status: "active", Updated: "2026-04-16"},
			{ID: newID(), Name: "data-residency-eu", Category: "Compliance", Scope: "EU workloads", Enforcement: "Deny", Compliant: 210, NonCompliant: 0, Exemptions: 0, Severity: "Critical", LastEval: "2026-04-17 09:00", Status: "active", Updated: "2026-04-15"},
			{ID: newID(), Name: "deprecated-api-versions", Category: "Governance", Scope: "All accounts", Enforcement: "Audit", Compliant: 320, NonCompliant: 18, Exemptions: 5, Severity: "Medium", LastEval: "2026-04-17 09:00", Status: "active", Updated: "2026-04-12"},
		},
		"hosting": {
			{ID: newID(), Name: "multi-cloud-rbac-sync", Category: "Identity", Scope: "All providers", Enforcement: "Deny", Compliant: 500, NonCompliant: 8, Exemptions: 3, Severity: "High", LastEval: "2026-04-17 09:00", Status: "active", Updated: "2026-04-17"},
			{ID: newID(), Name: "budget-guardrails", Category: "Cost Management", Scope: "All accounts", Enforcement: "Deny", Compliant: 42, NonCompliant: 1, Exemptions: 0, Severity: "High", LastEval: "2026-04-17 09:00", Status: "active", Updated: "2026-04-17"},
			{ID: newID(), Name: "require-mfa-admin", Category: "Security", Scope: "Admin roles", Enforcement: "Deny", Compliant: 28, NonCompliant: 0, Exemptions: 0, Severity: "Critical", LastEval: "2026-04-17 09:00", Status: "active", Updated: "2026-04-16"},
			{ID: newID(), Name: "network-segmentation", Category: "Networking", Scope: "Production", Enforcement: "Audit", Compliant: 115, NonCompliant: 7, Exemptions: 2, Severity: "Medium", LastEval: "2026-04-17 09:00", Status: "active", Updated: "2026-04-15"},
			{ID: newID(), Name: "container-image-signing", Category: "Security", Scope: "Kubernetes", Enforcement: "Deny", Compliant: 65, NonCompliant: 0, Exemptions: 1, Severity: "High", LastEval: "2026-04-17 09:00", Status: "active", Updated: "2026-04-14"},
		},
	},
}

func (s *governancePolicyStoreType) list(slug string) []GovernancePolicy {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.items[slug]
	out := make([]GovernancePolicy, len(src))
	copy(out, src)
	return out
}

func (s *governancePolicyStoreType) getByID(slug, id string) (GovernancePolicy, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, p := range s.items[slug] {
		if p.ID == id {
			return p, true
		}
	}
	return GovernancePolicy{}, false
}

func (s *governancePolicyStoreType) add(slug string, p GovernancePolicy) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[slug] = append(s.items[slug], p)
}

func (s *governancePolicyStoreType) delete(slug, id string) {
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

type apiGovernancePolicy struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Category     string `json:"category"`
	Scope        string `json:"scope"`
	Enforcement  string `json:"enforcement"`
	Compliant    int    `json:"compliant"`
	NonCompliant int    `json:"non_compliant"`
	Exemptions   int    `json:"exemptions"`
	Severity     string `json:"severity"`
	LastEval     string `json:"last_eval"`
	Status       string `json:"status"`
	Updated      string `json:"updated"`
}

type apiGovernancePolicyList struct {
	Policies []apiGovernancePolicy `json:"policies"`
	Total    int                   `json:"total"`
}

func toAPIGovernancePolicy(p GovernancePolicy) apiGovernancePolicy {
	return apiGovernancePolicy{
		ID: p.ID, Name: p.Name, Category: p.Category,
		Scope: p.Scope, Enforcement: p.Enforcement,
		Compliant: p.Compliant, NonCompliant: p.NonCompliant,
		Exemptions: p.Exemptions, Severity: p.Severity,
		LastEval: p.LastEval, Status: p.Status, Updated: p.Updated,
	}
}

func filterGovernancePolicies(all []GovernancePolicy, search, status string) []GovernancePolicy {
	var out []GovernancePolicy
	q := strings.ToLower(search)
	for _, p := range all {
		if status != "" && p.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(p.Name + " " + p.Category + " " + p.Scope + " " + p.Severity)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, p)
	}
	return out
}

func apiWSMulticloudMgmtPolicyListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	items := filterGovernancePolicies(governancePolicyStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiGovernancePolicy, len(items))
	for i, p := range items {
		out[i] = toAPIGovernancePolicy(p)
	}
	apiOK(w, apiGovernancePolicyList{Policies: out, Total: len(out)})
}

type apiCreateGovernancePolicyRequest struct {
	Name        string `json:"name"`
	Category    string `json:"category"`
	Scope       string `json:"scope"`
	Enforcement string `json:"enforcement"`
	Severity    string `json:"severity"`
}

func apiWSMulticloudMgmtPolicyCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateGovernancePolicyRequest
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
		category = "Governance"
	}
	scope := strings.TrimSpace(req.Scope)
	if scope == "" {
		scope = "All accounts"
	}
	enforcement := strings.TrimSpace(req.Enforcement)
	if enforcement == "" {
		enforcement = "Audit"
	}
	severity := strings.TrimSpace(req.Severity)
	if severity == "" {
		severity = "Medium"
	}
	policy := GovernancePolicy{
		ID: newID(), Name: name, Category: category,
		Scope: scope, Enforcement: enforcement,
		Compliant: 0, NonCompliant: 0, Exemptions: 0,
		Severity: severity, LastEval: "—",
		Status: "active", Updated: "2026-04-17",
	}
	governancePolicyStore.add(slug, policy)
	apiCreated(w, toAPIGovernancePolicy(policy))
}

func apiWSMulticloudMgmtPolicyGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	policy, found := governancePolicyStore.getByID(slug, r.PathValue("pid"))
	if !found {
		apiError(w, http.StatusNotFound, "Governance policy not found.")
		return
	}
	apiOK(w, toAPIGovernancePolicy(policy))
}

func apiWSMulticloudMgmtPolicyDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := governancePolicyStore.getByID(slug, r.PathValue("pid")); !found {
		apiError(w, http.StatusNotFound, "Governance policy not found.")
		return
	}
	governancePolicyStore.delete(slug, r.PathValue("pid"))
	apiOK(w, map[string]any{"ok": true, "message": "Governance policy deleted."})
}
