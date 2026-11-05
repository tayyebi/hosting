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

// ── Tax store ──────────────────────────────────────────────

type TaxRule struct {
	ID         string
	Name       string
	Region     string
	Type       string // "vat" | "sales_tax" | "gst" | "withholding" | "exempt"
	Status     string // "active" | "inactive" | "pending"
	Rate       string // percentage, e.g. "21.00"
	AppliesTo  string // "all" | "hosting" | "domains" | "support" | "hardware"
	ValidFrom  string
	ValidUntil string
	Created    string
	Updated    string
}

type taxStoreType struct {
	mu    sync.RWMutex
	rules map[string][]TaxRule
}

var taxStore = &taxStoreType{
	rules: map[string][]TaxRule{
		"buyvm": {
			{ID: newID(), Name: "US Federal — No Sales Tax", Region: "US", Type: "exempt", Status: "active", Rate: "0.00", AppliesTo: "all", ValidFrom: "2025-01-01", ValidUntil: "", Created: "2025-01-01T00:00:00Z", Updated: "2025-01-01T00:00:00Z"},
			{ID: newID(), Name: "New York State Sales Tax", Region: "US-NY", Type: "sales_tax", Status: "active", Rate: "8.00", AppliesTo: "hardware", ValidFrom: "2025-01-01", ValidUntil: "", Created: "2025-01-01T00:00:00Z", Updated: "2025-06-15T00:00:00Z"},
			{ID: newID(), Name: "EU VAT — Standard", Region: "EU", Type: "vat", Status: "active", Rate: "21.00", AppliesTo: "all", ValidFrom: "2025-01-01", ValidUntil: "", Created: "2025-01-01T00:00:00Z", Updated: "2026-01-01T00:00:00Z"},
			{ID: newID(), Name: "EU VAT — Reverse Charge (B2B)", Region: "EU", Type: "vat", Status: "active", Rate: "0.00", AppliesTo: "all", ValidFrom: "2025-01-01", ValidUntil: "", Created: "2025-01-01T00:00:00Z", Updated: "2025-01-01T00:00:00Z"},
			{ID: newID(), Name: "Canada GST", Region: "CA", Type: "gst", Status: "active", Rate: "5.00", AppliesTo: "all", ValidFrom: "2025-03-01", ValidUntil: "", Created: "2025-03-01T00:00:00Z", Updated: "2025-03-01T00:00:00Z"},
		},
		"abrnoc": {
			{ID: newID(), Name: "Germany VAT (Umsatzsteuer)", Region: "DE", Type: "vat", Status: "active", Rate: "19.00", AppliesTo: "all", ValidFrom: "2025-01-01", ValidUntil: "", Created: "2025-01-01T00:00:00Z", Updated: "2025-01-01T00:00:00Z"},
			{ID: newID(), Name: "France VAT (TVA)", Region: "FR", Type: "vat", Status: "active", Rate: "20.00", AppliesTo: "all", ValidFrom: "2025-01-01", ValidUntil: "", Created: "2025-01-01T00:00:00Z", Updated: "2025-01-01T00:00:00Z"},
			{ID: newID(), Name: "Netherlands VAT (BTW)", Region: "NL", Type: "vat", Status: "active", Rate: "21.00", AppliesTo: "all", ValidFrom: "2025-01-01", ValidUntil: "", Created: "2025-01-01T00:00:00Z", Updated: "2025-01-01T00:00:00Z"},
			{ID: newID(), Name: "EU Reverse Charge B2B", Region: "EU", Type: "vat", Status: "active", Rate: "0.00", AppliesTo: "all", ValidFrom: "2025-01-01", ValidUntil: "", Created: "2025-01-01T00:00:00Z", Updated: "2025-01-01T00:00:00Z"},
		},
		"hosting": {
			{ID: newID(), Name: "US — Digital Services (No Tax)", Region: "US", Type: "exempt", Status: "active", Rate: "0.00", AppliesTo: "all", ValidFrom: "2025-01-01", ValidUntil: "", Created: "2025-01-01T00:00:00Z", Updated: "2025-01-01T00:00:00Z"},
			{ID: newID(), Name: "India GST — Cloud Services", Region: "IN", Type: "gst", Status: "active", Rate: "18.00", AppliesTo: "all", ValidFrom: "2025-04-01", ValidUntil: "", Created: "2025-04-01T00:00:00Z", Updated: "2025-04-01T00:00:00Z"},
			{ID: newID(), Name: "India Withholding Tax (TDS)", Region: "IN", Type: "withholding", Status: "active", Rate: "2.00", AppliesTo: "all", ValidFrom: "2025-04-01", ValidUntil: "", Created: "2025-04-01T00:00:00Z", Updated: "2025-04-01T00:00:00Z"},
		},
	},
}

func (s *taxStoreType) list(slug string) []TaxRule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.rules[slug]
	out := make([]TaxRule, len(src))
	copy(out, src)
	return out
}

func (s *taxStoreType) getByID(slug, id string) (TaxRule, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, t := range s.rules[slug] {
		if t.ID == id {
			return t, true
		}
	}
	return TaxRule{}, false
}

func (s *taxStoreType) add(slug string, t TaxRule) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rules[slug] = append(s.rules[slug], t)
}

func (s *taxStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.rules[slug]
	for i, t := range list {
		if t.ID == id {
			s.rules[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

func filterTaxRules(all []TaxRule, search, taxType, status string) []TaxRule {
	var out []TaxRule
	q := strings.ToLower(search)
	for _, t := range all {
		if taxType != "" && t.Type != taxType {
			continue
		}
		if status != "" && t.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(t.Name + " " + t.Region + " " + t.AppliesTo + " " + t.Rate)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, t)
	}
	return out
}

func dbTaxRuleToLocal(t db.TaxRule) TaxRule {
	validFrom := ""
	if t.ValidFrom != nil {
		validFrom = t.ValidFrom.Format("2006-01-02")
	}
	validUntil := ""
	if t.ValidUntil != nil {
		validUntil = t.ValidUntil.Format("2006-01-02")
	}
	return TaxRule{
		ID:         fmt.Sprintf("%d", t.ID),
		Name:       t.Name,
		Region:     t.Region,
		Type:       t.TaxType,
		Status:     t.Status,
		Rate:       fmt.Sprintf("%.2f", t.Rate),
		AppliesTo:  t.AppliesTo,
		ValidFrom:  validFrom,
		ValidUntil: validUntil,
		Created:    t.CreatedAt.Format(time.RFC3339),
		Updated:    t.UpdatedAt.Format(time.RFC3339),
	}
}

// ── GET /api/tenant/{slug}/tax ─────────────────────────────

type apiTaxRule struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Region     string `json:"region"`
	Type       string `json:"type"`
	Status     string `json:"status"`
	Rate       string `json:"rate"`
	AppliesTo  string `json:"applies_to"`
	ValidFrom  string `json:"valid_from"`
	ValidUntil string `json:"valid_until"`
	Created    string `json:"created"`
	Updated    string `json:"updated"`
}

type apiTaxListResponse struct {
	Rules []apiTaxRule `json:"rules"`
	Total int          `json:"total"`
}

func apiTenantTaxListHandler(w http.ResponseWriter, r *http.Request) {
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
	taxType := r.URL.Query().Get("type")
	status := r.URL.Query().Get("status")

	var rules []TaxRule
	if Pools != nil && tenant.DBID > 0 {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		drs, err := db.ListTaxRules(ctx, Pools.Admin, tenant.DBID, search, taxType, status)
		if err != nil {
			log.Printf("apiTenantTaxListHandler: %v", err)
			rules = filterTaxRules(taxStore.list(slug), search, taxType, status)
		} else {
			rules = make([]TaxRule, len(drs))
			for i, dr := range drs {
				rules[i] = dbTaxRuleToLocal(dr)
			}
		}
	} else {
		rules = filterTaxRules(taxStore.list(slug), search, taxType, status)
	}

	out := make([]apiTaxRule, len(rules))
	for i, t := range rules {
		out[i] = apiTaxRule{
			ID: t.ID, Name: t.Name, Region: t.Region,
			Type: t.Type, Status: t.Status, Rate: t.Rate,
			AppliesTo:  t.AppliesTo,
			ValidFrom:  t.ValidFrom, ValidUntil: t.ValidUntil,
			Created: t.Created, Updated: t.Updated,
		}
	}

	apiOK(w, apiTaxListResponse{Rules: out, Total: len(out)})
}

// ── POST /api/tenant/{slug}/tax ────────────────────────────

type apiCreateTaxRequest struct {
	Name       string `json:"name"`
	Region     string `json:"region"`
	Type       string `json:"type"`
	Rate       string `json:"rate"`
	AppliesTo  string `json:"applies_to"`
	ValidFrom  string `json:"valid_from"`
	ValidUntil string `json:"valid_until"`
}

func apiTenantTaxCreateHandler(w http.ResponseWriter, r *http.Request) {
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

	var req apiCreateTaxRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}

	name := strings.TrimSpace(req.Name)
	region := strings.TrimSpace(req.Region)
	taxType := strings.TrimSpace(req.Type)
	rate := strings.TrimSpace(req.Rate)
	appliesTo := strings.TrimSpace(req.AppliesTo)

	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	if taxType == "" {
		taxType = "vat"
	}
	if appliesTo == "" {
		appliesTo = "all"
	}

	if Pools != nil && tenant.DBID > 0 {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		rateF, _ := strconv.ParseFloat(rate, 64)
		var validFrom, validUntil *time.Time
		if vf := strings.TrimSpace(req.ValidFrom); vf != "" {
			if t, err := time.Parse("2006-01-02", vf); err == nil {
				validFrom = &t
			}
		}
		if vu := strings.TrimSpace(req.ValidUntil); vu != "" {
			if t, err := time.Parse("2006-01-02", vu); err == nil {
				validUntil = &t
			}
		}
		dr, err := db.CreateTaxRule(ctx, Pools.Admin, tenant.DBID, name, region, taxType, appliesTo, rateF, validFrom, validUntil)
		if err != nil {
			log.Printf("apiTenantTaxCreateHandler: %v", err)
			apiError(w, http.StatusInternalServerError, "Failed to create tax rule.")
			return
		}
		t := dbTaxRuleToLocal(dr)
		apiCreated(w, apiTaxRule{
			ID: t.ID, Name: t.Name, Region: t.Region,
			Type: t.Type, Status: t.Status, Rate: t.Rate,
			AppliesTo:  t.AppliesTo,
			ValidFrom:  t.ValidFrom, ValidUntil: t.ValidUntil,
			Created: t.Created, Updated: t.Updated,
		})
		return
	}

	t := TaxRule{
		ID: newID(), Name: name, Region: region,
		Type: taxType, Status: "active", Rate: rate,
		AppliesTo:  appliesTo,
		ValidFrom:  strings.TrimSpace(req.ValidFrom), ValidUntil: strings.TrimSpace(req.ValidUntil),
		Created: time.Now().Format(time.RFC3339), Updated: time.Now().Format(time.RFC3339),
	}
	taxStore.add(slug, t)

	apiCreated(w, apiTaxRule{
		ID: t.ID, Name: t.Name, Region: t.Region,
		Type: t.Type, Status: t.Status, Rate: t.Rate,
		AppliesTo:  t.AppliesTo,
		ValidFrom:  t.ValidFrom, ValidUntil: t.ValidUntil,
		Created: t.Created, Updated: t.Updated,
	})
}

// ── GET /api/tenant/{slug}/tax/{id} ────────────────────────

func apiTenantTaxGetHandler(w http.ResponseWriter, r *http.Request) {
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
			dr, err := db.GetTaxRule(ctx, Pools.Admin, tenant.DBID, id)
			if err != nil {
				apiError(w, http.StatusNotFound, "Tax rule not found.")
				return
			}
			t := dbTaxRuleToLocal(dr)
			apiOK(w, apiTaxRule{
				ID: t.ID, Name: t.Name, Region: t.Region,
				Type: t.Type, Status: t.Status, Rate: t.Rate,
				AppliesTo:  t.AppliesTo,
				ValidFrom:  t.ValidFrom, ValidUntil: t.ValidUntil,
				Created: t.Created, Updated: t.Updated,
			})
			return
		}
	}

	t, ok := taxStore.getByID(slug, idStr)
	if !ok {
		apiError(w, http.StatusNotFound, "Tax rule not found.")
		return
	}

	apiOK(w, apiTaxRule{
		ID: t.ID, Name: t.Name, Region: t.Region,
		Type: t.Type, Status: t.Status, Rate: t.Rate,
		AppliesTo:  t.AppliesTo,
		ValidFrom:  t.ValidFrom, ValidUntil: t.ValidUntil,
		Created: t.Created, Updated: t.Updated,
	})
}

// ── DELETE /api/tenant/{slug}/tax/{id} ─────────────────────

func apiTenantTaxDeleteHandler(w http.ResponseWriter, r *http.Request) {
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
			if err := db.DeleteTaxRule(ctx, Pools.Admin, tenant.DBID, id); err != nil {
				log.Printf("apiTenantTaxDeleteHandler: %v", err)
				apiError(w, http.StatusInternalServerError, "Failed to delete tax rule.")
				return
			}
			apiOK(w, map[string]any{"ok": true, "message": "Tax rule deleted."})
			return
		}
	}

	if _, ok := taxStore.getByID(slug, idStr); !ok {
		apiError(w, http.StatusNotFound, "Tax rule not found.")
		return
	}

	taxStore.delete(slug, idStr)
	apiOK(w, map[string]any{
		"ok":      true,
		"message": "Tax rule deleted.",
	})
}
