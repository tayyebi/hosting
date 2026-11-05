package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Multi-Cloud API — Cost: Cloud Optimizer (Pro Feature)
// multicloud/cost/cloud-optimizer
// ════════════════════════════════════════════════════════════════

type ProCostOptimizer struct {
	ID                  string
	Name                string
	Provider            string // "AWS", "Azure", "GCP", "Multi-Cloud"
	AccountID           string
	MonthlySpend        float64
	OptimizedSpend      float64
	SavingsPercent      float64
	RecommendationCount int
	ImplementedCount    int
	RightsizingAlerts   int
	UnusedResources     int
	ReservedCoverage    float64 // percentage
	SpotUsage           float64 // percentage
	BudgetLimit         float64
	BudgetUsed          float64
	ForecastMonth       float64
	CostTrend           string // "decreasing", "stable", "increasing"
	LastScan            string
	Region              string
	CostPerMonth        float64
	Status              string // "active", "scanning", "paused"
	Updated             string
}

type proCostOptimizerStoreType struct {
	mu    sync.RWMutex
	items map[string][]ProCostOptimizer
}

var proCostOptimizerStore = &proCostOptimizerStoreType{
	items: map[string][]ProCostOptimizer{},
}

func (s *proCostOptimizerStoreType) list(slug string) []ProCostOptimizer {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.items[slug]
	out := make([]ProCostOptimizer, len(src))
	copy(out, src)
	return out
}

func (s *proCostOptimizerStoreType) getByID(slug, id string) (ProCostOptimizer, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, o := range s.items[slug] {
		if o.ID == id {
			return o, true
		}
	}
	return ProCostOptimizer{}, false
}

func (s *proCostOptimizerStoreType) add(slug string, o ProCostOptimizer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[slug] = append(s.items[slug], o)
}

func (s *proCostOptimizerStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.items[slug]
	for i, o := range list {
		if o.ID == id {
			s.items[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiProCostOptimizer struct {
	ID                  string  `json:"id"`
	Name                string  `json:"name"`
	Provider            string  `json:"provider"`
	AccountID           string  `json:"account_id"`
	MonthlySpend        float64 `json:"monthly_spend"`
	OptimizedSpend      float64 `json:"optimized_spend"`
	SavingsPercent      float64 `json:"savings_percent"`
	RecommendationCount int     `json:"recommendation_count"`
	ImplementedCount    int     `json:"implemented_count"`
	RightsizingAlerts   int     `json:"rightsizing_alerts"`
	UnusedResources     int     `json:"unused_resources"`
	ReservedCoverage    float64 `json:"reserved_coverage_pct"`
	SpotUsage           float64 `json:"spot_usage_pct"`
	BudgetLimit         float64 `json:"budget_limit"`
	BudgetUsed          float64 `json:"budget_used"`
	ForecastMonth       float64 `json:"forecast_month"`
	CostTrend           string  `json:"cost_trend"`
	LastScan            string  `json:"last_scan"`
	Region              string  `json:"region"`
	CostPerMonth        float64 `json:"cost_per_month"`
	Status              string  `json:"status"`
	Updated             string  `json:"updated"`
}

type apiProCostOptimizerList struct {
	Optimizers []apiProCostOptimizer `json:"optimizers"`
	Total      int                   `json:"total"`
}

func toAPIProCostOptimizer(o ProCostOptimizer) apiProCostOptimizer {
	return apiProCostOptimizer{
		ID: o.ID, Name: o.Name, Provider: o.Provider,
		AccountID: o.AccountID, MonthlySpend: o.MonthlySpend,
		OptimizedSpend: o.OptimizedSpend, SavingsPercent: o.SavingsPercent,
		RecommendationCount: o.RecommendationCount, ImplementedCount: o.ImplementedCount,
		RightsizingAlerts: o.RightsizingAlerts, UnusedResources: o.UnusedResources,
		ReservedCoverage: o.ReservedCoverage, SpotUsage: o.SpotUsage,
		BudgetLimit: o.BudgetLimit, BudgetUsed: o.BudgetUsed,
		ForecastMonth: o.ForecastMonth, CostTrend: o.CostTrend,
		LastScan: o.LastScan, Region: o.Region,
		CostPerMonth: o.CostPerMonth, Status: o.Status, Updated: o.Updated,
	}
}

func filterProCostOptimizers(all []ProCostOptimizer, search, status string) []ProCostOptimizer {
	var out []ProCostOptimizer
	q := strings.ToLower(search)
	for _, o := range all {
		if status != "" && o.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(o.Name + " " + o.Provider + " " + o.AccountID + " " + o.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, o)
	}
	return out
}

func apiWSMulticloudCostOptimizerListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	items := filterProCostOptimizers(proCostOptimizerStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiProCostOptimizer, len(items))
	for i, o := range items {
		out[i] = toAPIProCostOptimizer(o)
	}
	apiOK(w, apiProCostOptimizerList{Optimizers: out, Total: len(out)})
}

type apiCreateProCostOptimizerRequest struct {
	Name        string  `json:"name"`
	Provider    string  `json:"provider"`
	AccountID   string  `json:"account_id"`
	BudgetLimit float64 `json:"budget_limit"`
	Region      string  `json:"region"`
}

func apiWSMulticloudCostOptimizerCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateProCostOptimizerRequest
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
	accountID := strings.TrimSpace(req.AccountID)
	if accountID == "" {
		accountID = "account-" + newID()
	}
	budgetLimit := req.BudgetLimit
	if budgetLimit <= 0 {
		budgetLimit = 10000.00
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "us-east-1"
	}
	opt := ProCostOptimizer{
		ID: newID(), Name: name, Provider: provider,
		AccountID: accountID, MonthlySpend: 0, OptimizedSpend: 0,
		SavingsPercent: 0, RecommendationCount: 0, ImplementedCount: 0,
		RightsizingAlerts: 0, UnusedResources: 0,
		ReservedCoverage: 0, SpotUsage: 0,
		BudgetLimit: budgetLimit, BudgetUsed: 0,
		ForecastMonth: 0, CostTrend: "stable",
		LastScan: "—", Region: region,
		CostPerMonth: 45.00, Status: "scanning", Updated: "2026-04-18",
	}
	proCostOptimizerStore.add(slug, opt)
	apiCreated(w, toAPIProCostOptimizer(opt))
}

func apiWSMulticloudCostOptimizerGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	o, found := proCostOptimizerStore.getByID(slug, r.PathValue("oid"))
	if !found {
		apiError(w, http.StatusNotFound, "Cost optimizer not found.")
		return
	}
	apiOK(w, toAPIProCostOptimizer(o))
}

func apiWSMulticloudCostOptimizerDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := proCostOptimizerStore.getByID(slug, r.PathValue("oid")); !found {
		apiError(w, http.StatusNotFound, "Cost optimizer not found.")
		return
	}
	proCostOptimizerStore.delete(slug, r.PathValue("oid"))
	apiOK(w, map[string]any{"ok": true, "message": "Cost optimizer deleted."})
}
