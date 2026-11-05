package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"platform/db"
)

// ── Pricing store ──────────────────────────────────────────

type PricingPlan struct {
	ID       string
	Name     string
	Tier     string // "free" | "starter" | "pro" | "enterprise"
	Status   string // "active" | "archived" | "draft"
	Price    string
	Currency string
	Interval string // "monthly" | "annually" | "one-time"
	Features string
	MaxUsers int
	Created  string
	Updated  string
}

type pricingStoreType struct {
	mu    sync.RWMutex
	plans map[string][]PricingPlan
}

var pricingStore = &pricingStoreType{
	plans: map[string][]PricingPlan{},
}

func (s *pricingStoreType) list(slug string) []PricingPlan {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.plans[slug]
	out := make([]PricingPlan, len(src))
	copy(out, src)
	return out
}

func (s *pricingStoreType) getByID(slug, id string) (PricingPlan, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, p := range s.plans[slug] {
		if p.ID == id {
			return p, true
		}
	}
	return PricingPlan{}, false
}

func (s *pricingStoreType) add(slug string, p PricingPlan) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.plans[slug] = append(s.plans[slug], p)
}

func (s *pricingStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.plans[slug]
	for i, p := range list {
		if p.ID == id {
			s.plans[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

func filterPricingPlans(all []PricingPlan, search, tier, status string) []PricingPlan {
	var out []PricingPlan
	q := strings.ToLower(search)
	for _, p := range all {
		if tier != "" && p.Tier != tier {
			continue
		}
		if status != "" && p.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(p.Name + " " + p.Features + " " + p.Price)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, p)
	}
	return out
}

func dbPricingPlanToLocal(p db.PricingPlan) PricingPlan {
	return PricingPlan{
		ID:       fmt.Sprintf("%d", p.ID),
		Name:     p.Name,
		Tier:     p.Tier,
		Status:   p.Status,
		Price:    fmt.Sprintf("%.2f", p.Price),
		Currency: p.Currency,
		Interval: p.Interval,
		Features: p.Features,
		MaxUsers: p.MaxUsers,
		Created:  p.CreatedAt.Format(time.RFC3339),
		Updated:  p.UpdatedAt.Format(time.RFC3339),
	}
}

// ── GET /api/tenant/{slug}/pricing ─────────────────────────

type apiPricingPlan struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Tier     string `json:"tier"`
	Status   string `json:"status"`
	Price    string `json:"price"`
	Currency string `json:"currency"`
	Interval string `json:"interval"`
	Features string `json:"features"`
	MaxUsers int    `json:"max_users"`
	Created  string `json:"created"`
	Updated  string `json:"updated"`
}

type apiPricingListResponse struct {
	Plans []apiPricingPlan `json:"plans"`
	Total int              `json:"total"`
}

func apiTenantPricingListHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok {
		apiError(w, http.StatusUnauthorized, "Authentication required.")
		return
	}

	slug := r.PathValue("slug")
	tenant, found := store.getBySlug(slug)
	if !found {
		apiError(w, http.StatusNotFound, "Tenant not found.")
		return
	}
	if user.Role != RoleSysAdmin && (user.Role != RoleTenant || user.TenantSlug != slug) {
		apiError(w, http.StatusForbidden, "Access denied.")
		return
	}

	search := r.URL.Query().Get("q")
	tier := r.URL.Query().Get("tier")
	status := r.URL.Query().Get("status")

	var plans []PricingPlan
	if Pools != nil && tenant.DBID > 0 {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		dps, err := db.ListPricingPlans(ctx, Pools.Admin, tenant.DBID, search, tier, status)
		if err != nil {
			log.Printf("apiTenantPricingListHandler: %v", err)
			plans = filterPricingPlans(pricingStore.list(slug), search, tier, status)
		} else {
			plans = make([]PricingPlan, len(dps))
			for i, dp := range dps {
				plans[i] = dbPricingPlanToLocal(dp)
			}
		}
	} else {
		plans = filterPricingPlans(pricingStore.list(slug), search, tier, status)
	}

	out := make([]apiPricingPlan, len(plans))
	for i, p := range plans {
		out[i] = apiPricingPlan{
			ID: p.ID, Name: p.Name, Tier: p.Tier,
			Status: p.Status, Price: p.Price, Currency: p.Currency,
			Interval: p.Interval, Features: p.Features,
			MaxUsers: p.MaxUsers,
			Created:  p.Created, Updated: p.Updated,
		}
	}

	apiOK(w, apiPricingListResponse{Plans: out, Total: len(out)})
}

// ── POST /api/tenant/{slug}/pricing ────────────────────────

type apiCreatePricingRequest struct {
	Name     string `json:"name"`
	Tier     string `json:"tier"`
	Price    string `json:"price"`
	Currency string `json:"currency"`
	Interval string `json:"interval"`
	Features string `json:"features"`
	MaxUsers int    `json:"max_users"`
}

func apiTenantPricingCreateHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok {
		apiError(w, http.StatusUnauthorized, "Authentication required.")
		return
	}

	slug := r.PathValue("slug")
	tenant, found := store.getBySlug(slug)
	if !found {
		apiError(w, http.StatusNotFound, "Tenant not found.")
		return
	}
	if user.Role != RoleSysAdmin && (user.Role != RoleTenant || user.TenantSlug != slug) {
		apiError(w, http.StatusForbidden, "Access denied.")
		return
	}

	var req apiCreatePricingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}

	name := strings.TrimSpace(req.Name)
	tier := strings.TrimSpace(req.Tier)
	price := strings.TrimSpace(req.Price)
	currency := strings.TrimSpace(req.Currency)
	interval := strings.TrimSpace(req.Interval)

	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	if tier == "" {
		tier = "starter"
	}
	if currency == "" {
		currency = "USD"
	}
	if interval == "" {
		interval = "monthly"
	}

	if Pools != nil && tenant.DBID > 0 {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		priceF, _ := strconv.ParseFloat(price, 64)
		dp, err := db.CreatePricingPlan(ctx, Pools.Admin, tenant.DBID, name, tier, currency, interval, strings.TrimSpace(req.Features), priceF, req.MaxUsers)
		if err != nil {
			log.Printf("apiTenantPricingCreateHandler: %v", err)
			apiError(w, http.StatusInternalServerError, "Failed to create pricing plan.")
			return
		}
		p := dbPricingPlanToLocal(dp)
		apiCreated(w, apiPricingPlan{
			ID: p.ID, Name: p.Name, Tier: p.Tier,
			Status: p.Status, Price: p.Price, Currency: p.Currency,
			Interval: p.Interval, Features: p.Features,
			MaxUsers: p.MaxUsers,
			Created:  p.Created, Updated: p.Updated,
		})
		return
	}

	p := PricingPlan{
		ID: newID(), Name: name, Tier: tier, Status: "draft",
		Price: price, Currency: currency, Interval: interval,
		Features: strings.TrimSpace(req.Features), MaxUsers: req.MaxUsers,
		Created: time.Now().Format(time.RFC3339), Updated: time.Now().Format(time.RFC3339),
	}
	pricingStore.add(slug, p)

	apiCreated(w, apiPricingPlan{
		ID: p.ID, Name: p.Name, Tier: p.Tier,
		Status: p.Status, Price: p.Price, Currency: p.Currency,
		Interval: p.Interval, Features: p.Features,
		MaxUsers: p.MaxUsers,
		Created:  p.Created, Updated: p.Updated,
	})
}

// ── GET /api/tenant/{slug}/pricing/{id} ────────────────────

func apiTenantPricingGetHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok {
		apiError(w, http.StatusUnauthorized, "Authentication required.")
		return
	}

	slug := r.PathValue("slug")
	tenant, found := store.getBySlug(slug)
	if !found {
		apiError(w, http.StatusNotFound, "Tenant not found.")
		return
	}
	if user.Role != RoleSysAdmin && (user.Role != RoleTenant || user.TenantSlug != slug) {
		apiError(w, http.StatusForbidden, "Access denied.")
		return
	}

	idStr := r.PathValue("id")

	if Pools != nil && tenant.DBID > 0 {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err == nil {
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()
			dp, err := db.GetPricingPlan(ctx, Pools.Admin, tenant.DBID, id)
			if err != nil {
				apiError(w, http.StatusNotFound, "Pricing plan not found.")
				return
			}
			p := dbPricingPlanToLocal(dp)
			apiOK(w, apiPricingPlan{
				ID: p.ID, Name: p.Name, Tier: p.Tier,
				Status: p.Status, Price: p.Price, Currency: p.Currency,
				Interval: p.Interval, Features: p.Features,
				MaxUsers: p.MaxUsers,
				Created:  p.Created, Updated: p.Updated,
			})
			return
		}
	}

	p, ok := pricingStore.getByID(slug, idStr)
	if !ok {
		apiError(w, http.StatusNotFound, "Pricing plan not found.")
		return
	}

	apiOK(w, apiPricingPlan{
		ID: p.ID, Name: p.Name, Tier: p.Tier,
		Status: p.Status, Price: p.Price, Currency: p.Currency,
		Interval: p.Interval, Features: p.Features,
		MaxUsers: p.MaxUsers,
		Created:  p.Created, Updated: p.Updated,
	})
}

// ── DELETE /api/tenant/{slug}/pricing/{id} ──────────────────

func apiTenantPricingDeleteHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok {
		apiError(w, http.StatusUnauthorized, "Authentication required.")
		return
	}

	slug := r.PathValue("slug")
	tenant, found := store.getBySlug(slug)
	if !found {
		apiError(w, http.StatusNotFound, "Tenant not found.")
		return
	}
	if user.Role != RoleSysAdmin && (user.Role != RoleTenant || user.TenantSlug != slug) {
		apiError(w, http.StatusForbidden, "Access denied.")
		return
	}

	idStr := r.PathValue("id")

	if Pools != nil && tenant.DBID > 0 {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err == nil {
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()
			if err := db.DeletePricingPlan(ctx, Pools.Admin, tenant.DBID, id); err != nil {
				log.Printf("apiTenantPricingDeleteHandler: %v", err)
				apiError(w, http.StatusInternalServerError, "Failed to delete pricing plan.")
				return
			}
			apiOK(w, map[string]any{"ok": true, "message": "Pricing plan deleted."})
			return
		}
	}

	if _, ok := pricingStore.getByID(slug, idStr); !ok {
		apiError(w, http.StatusNotFound, "Pricing plan not found.")
		return
	}

	pricingStore.delete(slug, idStr)
	apiOK(w, map[string]any{
		"ok":      true,
		"message": "Pricing plan deleted.",
	})
}
