package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Multi-Cloud & Hybrid API — Management: Cost Management
// multicloud/mgmt/cost
// ════════════════════════════════════════════════════════════════

type CostProfile struct {
	ID              string
	Name            string
	Provider        string
	Account         string
	MonthlyBudget   string
	CurrentSpend    string
	Forecast        string
	SavingsFound    string
	Alerts          int
	OptimizationPct string
	Status          string
	Updated         string
}

type costProfileStoreType struct {
	mu    sync.RWMutex
	items map[string][]CostProfile
}

var costProfileStore = &costProfileStoreType{
	items: map[string][]CostProfile{},
}

func (s *costProfileStoreType) list(slug string) []CostProfile {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.items[slug]
	out := make([]CostProfile, len(src))
	copy(out, src)
	return out
}

func (s *costProfileStoreType) getByID(slug, id string) (CostProfile, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, c := range s.items[slug] {
		if c.ID == id {
			return c, true
		}
	}
	return CostProfile{}, false
}

func (s *costProfileStoreType) add(slug string, c CostProfile) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[slug] = append(s.items[slug], c)
}

func (s *costProfileStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.items[slug]
	for i, c := range list {
		if c.ID == id {
			s.items[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiCostProfile struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Provider        string `json:"provider"`
	Account         string `json:"account"`
	MonthlyBudget   string `json:"monthly_budget"`
	CurrentSpend    string `json:"current_spend"`
	Forecast        string `json:"forecast"`
	SavingsFound    string `json:"savings_found"`
	Alerts          int    `json:"alerts"`
	OptimizationPct string `json:"optimization_pct"`
	Status          string `json:"status"`
	Updated         string `json:"updated"`
}

type apiCostProfileList struct {
	Profiles []apiCostProfile `json:"profiles"`
	Total    int              `json:"total"`
}

func toAPICostProfile(c CostProfile) apiCostProfile {
	return apiCostProfile{
		ID: c.ID, Name: c.Name, Provider: c.Provider,
		Account: c.Account, MonthlyBudget: c.MonthlyBudget,
		CurrentSpend: c.CurrentSpend, Forecast: c.Forecast,
		SavingsFound: c.SavingsFound, Alerts: c.Alerts,
		OptimizationPct: c.OptimizationPct, Status: c.Status,
		Updated: c.Updated,
	}
}

func filterCostProfiles(all []CostProfile, search, status string) []CostProfile {
	var out []CostProfile
	q := strings.ToLower(search)
	for _, c := range all {
		if status != "" && c.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(c.Name + " " + c.Provider + " " + c.Account)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, c)
	}
	return out
}

func apiWSMulticloudMgmtCostListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	items := filterCostProfiles(costProfileStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiCostProfile, len(items))
	for i, c := range items {
		out[i] = toAPICostProfile(c)
	}
	apiOK(w, apiCostProfileList{Profiles: out, Total: len(out)})
}

type apiCreateCostProfileRequest struct {
	Name          string `json:"name"`
	Provider      string `json:"provider"`
	Account       string `json:"account"`
	MonthlyBudget string `json:"monthly_budget"`
}

func apiWSMulticloudMgmtCostCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateCostProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	provider := strings.TrimSpace(req.Provider)
	if provider == "" {
		provider = "AWS"
	}
	account := strings.TrimSpace(req.Account)
	if account == "" {
		account = "default"
	}
	budget := strings.TrimSpace(req.MonthlyBudget)
	if budget == "" {
		budget = "$0"
	}
	cp := CostProfile{
		ID: newID(), Name: name, Provider: provider,
		Account: account, MonthlyBudget: budget, CurrentSpend: "$0",
		Forecast: "—", SavingsFound: "$0", Alerts: 0,
		OptimizationPct: "—", Status: "active", Updated: "2026-04-17",
	}
	costProfileStore.add(slug, cp)
	apiCreated(w, toAPICostProfile(cp))
}

func apiWSMulticloudMgmtCostGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	cp, found := costProfileStore.getByID(slug, r.PathValue("pid"))
	if !found {
		apiError(w, http.StatusNotFound, "Cost profile not found.")
		return
	}
	apiOK(w, toAPICostProfile(cp))
}

func apiWSMulticloudMgmtCostDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := costProfileStore.getByID(slug, r.PathValue("pid")); !found {
		apiError(w, http.StatusNotFound, "Cost profile not found.")
		return
	}
	costProfileStore.delete(slug, r.PathValue("pid"))
	apiOK(w, map[string]any{"ok": true, "message": "Cost profile deleted."})
}
