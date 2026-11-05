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
	ID                string
	Name              string
	Provider          string  // "AWS", "Azure", "GCP", "Multi-Cloud"
	AccountID         string
	MonthlySpend      float64
	OptimizedSpend    float64
	SavingsPercent    float64
	RecommendationCount int
	ImplementedCount  int
	RightsizingAlerts int
	UnusedResources   int
	ReservedCoverage  float64 // percentage
	SpotUsage         float64 // percentage
	BudgetLimit       float64
	BudgetUsed        float64
	ForecastMonth     float64
	CostTrend         string  // "decreasing", "stable", "increasing"
	LastScan          string
	Region            string
	CostPerMonth      float64
	Status            string  // "active", "scanning", "paused"
	Updated           string
}

type proCostOptimizerStoreType struct {
	mu    sync.RWMutex
	items map[string][]ProCostOptimizer
}

var proCostOptimizerStore = &proCostOptimizerStoreType{
	items: map[string][]ProCostOptimizer{
		"buyvm": {
			{ID: newID(), Name: "aws-production", Provider: "AWS", AccountID: "123456789012", MonthlySpend: 48200.00, OptimizedSpend: 38560.00, SavingsPercent: 20.0, RecommendationCount: 34, ImplementedCount: 22, RightsizingAlerts: 8, UnusedResources: 6, ReservedCoverage: 68.4, SpotUsage: 24.2, BudgetLimit: 55000.00, BudgetUsed: 48200.00, ForecastMonth: 49800.00, CostTrend: "decreasing", LastScan: "2026-04-18T06:00:00Z", Region: "us-east-1", CostPerMonth: 120.00, Status: "active", Updated: "2026-04-18"},
			{ID: newID(), Name: "aws-staging", Provider: "AWS", AccountID: "987654321098", MonthlySpend: 4800.00, OptimizedSpend: 3600.00, SavingsPercent: 25.0, RecommendationCount: 12, ImplementedCount: 4, RightsizingAlerts: 4, UnusedResources: 8, ReservedCoverage: 0, SpotUsage: 42.0, BudgetLimit: 6000.00, BudgetUsed: 4800.00, ForecastMonth: 5200.00, CostTrend: "stable", LastScan: "2026-04-18T06:00:00Z", Region: "us-east-1", CostPerMonth: 45.00, Status: "active", Updated: "2026-04-18"},
			{ID: newID(), Name: "gcp-ml-workloads", Provider: "GCP", AccountID: "ml-project-42", MonthlySpend: 18400.00, OptimizedSpend: 14720.00, SavingsPercent: 20.0, RecommendationCount: 18, ImplementedCount: 12, RightsizingAlerts: 6, UnusedResources: 2, ReservedCoverage: 42.0, SpotUsage: 64.0, BudgetLimit: 20000.00, BudgetUsed: 18400.00, ForecastMonth: 19200.00, CostTrend: "increasing", LastScan: "2026-04-18T06:00:00Z", Region: "us-central1", CostPerMonth: 85.00, Status: "active", Updated: "2026-04-18"},
			{ID: newID(), Name: "azure-enterprise", Provider: "Azure", AccountID: "sub-ent-001", MonthlySpend: 12600.00, OptimizedSpend: 10080.00, SavingsPercent: 20.0, RecommendationCount: 24, ImplementedCount: 16, RightsizingAlerts: 4, UnusedResources: 4, ReservedCoverage: 56.0, SpotUsage: 12.0, BudgetLimit: 15000.00, BudgetUsed: 12600.00, ForecastMonth: 13200.00, CostTrend: "stable", LastScan: "2026-04-18T06:00:00Z", Region: "eastus", CostPerMonth: 65.00, Status: "active", Updated: "2026-04-18"},
		},
		"hosting": {
			{ID: newID(), Name: "multi-cloud-overview", Provider: "Multi-Cloud", AccountID: "consolidated-001", MonthlySpend: 284000.00, OptimizedSpend: 213000.00, SavingsPercent: 25.0, RecommendationCount: 128, ImplementedCount: 86, RightsizingAlerts: 24, UnusedResources: 18, ReservedCoverage: 72.0, SpotUsage: 28.0, BudgetLimit: 300000.00, BudgetUsed: 284000.00, ForecastMonth: 292000.00, CostTrend: "decreasing", LastScan: "2026-04-18T04:00:00Z", Region: "eu-central-1", CostPerMonth: 480.00, Status: "active", Updated: "2026-04-18"},
			{ID: newID(), Name: "aws-primary", Provider: "AWS", AccountID: "aws-hosting-prod", MonthlySpend: 186000.00, OptimizedSpend: 139500.00, SavingsPercent: 25.0, RecommendationCount: 64, ImplementedCount: 48, RightsizingAlerts: 12, UnusedResources: 8, ReservedCoverage: 78.0, SpotUsage: 32.0, BudgetLimit: 200000.00, BudgetUsed: 186000.00, ForecastMonth: 190000.00, CostTrend: "stable", LastScan: "2026-04-18T04:00:00Z", Region: "eu-central-1", CostPerMonth: 320.00, Status: "active", Updated: "2026-04-18"},
			{ID: newID(), Name: "gcp-analytics", Provider: "GCP", AccountID: "gcp-analytics-prod", MonthlySpend: 64000.00, OptimizedSpend: 48000.00, SavingsPercent: 25.0, RecommendationCount: 42, ImplementedCount: 28, RightsizingAlerts: 8, UnusedResources: 6, ReservedCoverage: 54.0, SpotUsage: 18.0, BudgetLimit: 70000.00, BudgetUsed: 64000.00, ForecastMonth: 68000.00, CostTrend: "increasing", LastScan: "2026-04-18T04:00:00Z", Region: "europe-west1", CostPerMonth: 180.00, Status: "active", Updated: "2026-04-18"},
			{ID: newID(), Name: "azure-backup-dr", Provider: "Azure", AccountID: "az-dr-sub-001", MonthlySpend: 34000.00, OptimizedSpend: 25500.00, SavingsPercent: 25.0, RecommendationCount: 22, ImplementedCount: 10, RightsizingAlerts: 4, UnusedResources: 4, ReservedCoverage: 48.0, SpotUsage: 0, BudgetLimit: 40000.00, BudgetUsed: 34000.00, ForecastMonth: 34000.00, CostTrend: "stable", LastScan: "2026-04-18T04:00:00Z", Region: "westeurope", CostPerMonth: 120.00, Status: "scanning", Updated: "2026-04-18"},
		},
	},
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
	Name         string  `json:"name"`
	Provider     string  `json:"provider"`
	AccountID    string  `json:"account_id"`
	BudgetLimit  float64 `json:"budget_limit"`
	Region       string  `json:"region"`
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
