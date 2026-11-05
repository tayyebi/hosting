package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Security API — Compliance: Policy Management
// security/compliance/policy
// ════════════════════════════════════════════════════════════════

type CompliancePolicy struct {
	ID          string
	Name        string
	Framework   string
	Scope       string
	Enforcement string
	Status      string
	Violations  string
	Updated     string
}

type compliancePolicyStoreType struct {
	mu       sync.RWMutex
	policies map[string][]CompliancePolicy
}

var compliancePolicyStore = &compliancePolicyStoreType{
	policies: map[string][]CompliancePolicy{},
}

func (s *compliancePolicyStoreType) list(slug string) []CompliancePolicy {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.policies[slug]
	out := make([]CompliancePolicy, len(src))
	copy(out, src)
	return out
}

func (s *compliancePolicyStoreType) getByID(slug, id string) (CompliancePolicy, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, p := range s.policies[slug] {
		if p.ID == id {
			return p, true
		}
	}
	return CompliancePolicy{}, false
}

func (s *compliancePolicyStoreType) add(slug string, p CompliancePolicy) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.policies[slug] = append(s.policies[slug], p)
}

func (s *compliancePolicyStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.policies[slug]
	for i, p := range list {
		if p.ID == id {
			s.policies[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiCompliancePolicy struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Framework   string `json:"framework"`
	Scope       string `json:"scope"`
	Enforcement string `json:"enforcement"`
	Status      string `json:"status"`
	Violations  string `json:"violations"`
	Updated     string `json:"updated"`
}

type apiCompliancePolicyList struct {
	Policies []apiCompliancePolicy `json:"policies"`
	Total    int                   `json:"total"`
}

func toAPICompliancePolicy(p CompliancePolicy) apiCompliancePolicy {
	return apiCompliancePolicy{
		ID: p.ID, Name: p.Name, Framework: p.Framework,
		Scope: p.Scope, Enforcement: p.Enforcement, Status: p.Status,
		Violations: p.Violations, Updated: p.Updated,
	}
}

func filterCompliancePolicies(all []CompliancePolicy, search, status string) []CompliancePolicy {
	var out []CompliancePolicy
	q := strings.ToLower(search)
	for _, p := range all {
		if status != "" && p.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(p.Name + " " + p.Framework + " " + p.Scope)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, p)
	}
	return out
}

func apiWSSecCompPolicyListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	policies := filterCompliancePolicies(compliancePolicyStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiCompliancePolicy, len(policies))
	for i, p := range policies {
		out[i] = toAPICompliancePolicy(p)
	}
	apiOK(w, apiCompliancePolicyList{Policies: out, Total: len(out)})
}

type apiCreateCompliancePolicyRequest struct {
	Name        string `json:"name"`
	Framework   string `json:"framework"`
	Scope       string `json:"scope"`
	Enforcement string `json:"enforcement"`
}

func apiWSSecCompPolicyCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateCompliancePolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	framework := strings.TrimSpace(req.Framework)
	if framework == "" {
		framework = "Internal"
	}
	scope := strings.TrimSpace(req.Scope)
	if scope == "" {
		scope = "All resources"
	}
	enforcement := strings.TrimSpace(req.Enforcement)
	if enforcement == "" {
		enforcement = "warn"
	}
	policy := CompliancePolicy{
		ID: newID(), Name: name, Framework: framework,
		Scope: scope, Enforcement: enforcement, Status: "active",
		Violations: "0", Updated: "2026-04-17",
	}
	compliancePolicyStore.add(slug, policy)
	apiCreated(w, toAPICompliancePolicy(policy))
}

func apiWSSecCompPolicyGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	policy, found := compliancePolicyStore.getByID(slug, r.PathValue("pid"))
	if !found {
		apiError(w, http.StatusNotFound, "Policy not found.")
		return
	}
	apiOK(w, toAPICompliancePolicy(policy))
}

func apiWSSecCompPolicyDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := compliancePolicyStore.getByID(slug, r.PathValue("pid")); !found {
		apiError(w, http.StatusNotFound, "Policy not found.")
		return
	}
	compliancePolicyStore.delete(slug, r.PathValue("pid"))
	apiOK(w, map[string]any{"ok": true, "message": "Policy deleted."})
}
