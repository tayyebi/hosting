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

// ── SLA store ──────────────────────────────────────────────

type SLAPolicy struct {
	ID           string
	Name         string
	Service      string
	Status       string // "active" | "inactive" | "breached" | "at-risk"
	UptimeTarget string // percentage, e.g. "99.95"
	UptimeActual string
	ResponseTime string
	CreditPolicy string
	PenaltyRate  string
	ReviewPeriod string // "monthly" | "quarterly" | "annually"
	LastReviewed string
	Created      string
	Updated      string
}

type slaStoreType struct {
	mu       sync.RWMutex
	policies map[string][]SLAPolicy
}

var slaStore = &slaStoreType{
	policies: map[string][]SLAPolicy{
		"buyvm": {
			{ID: newID(), Name: "Infrastructure Uptime SLA", Service: "Compute & Networking", Status: "active", UptimeTarget: "99.99", UptimeActual: "99.97", ResponseTime: "< 100ms", CreditPolicy: "10x credit for downtime minutes exceeding SLA", PenaltyRate: "10%", ReviewPeriod: "monthly", LastReviewed: "2026-04-01", Created: "2025-01-01T00:00:00Z", Updated: "2026-04-01T00:00:00Z"},
			{ID: newID(), Name: "API Availability SLA", Service: "API Gateway", Status: "active", UptimeTarget: "99.95", UptimeActual: "99.92", ResponseTime: "< 200ms", CreditPolicy: "5% credit per 0.1% below target", PenaltyRate: "5%", ReviewPeriod: "monthly", LastReviewed: "2026-04-01", Created: "2025-01-01T00:00:00Z", Updated: "2026-04-01T00:00:00Z"},
			{ID: newID(), Name: "Database SLA", Service: "Managed PostgreSQL", Status: "at-risk", UptimeTarget: "99.99", UptimeActual: "99.96", ResponseTime: "< 50ms", CreditPolicy: "15x credit for downtime exceeding SLA", PenaltyRate: "15%", ReviewPeriod: "monthly", LastReviewed: "2026-04-01", Created: "2025-03-15T00:00:00Z", Updated: "2026-04-01T00:00:00Z"},
			{ID: newID(), Name: "Support Response SLA", Service: "Technical Support", Status: "active", UptimeTarget: "100.00", UptimeActual: "100.00", ResponseTime: "< 15min (critical), < 1h (high), < 4h (normal)", CreditPolicy: "Service credit for missed response times", PenaltyRate: "2%", ReviewPeriod: "quarterly", LastReviewed: "2026-04-01", Created: "2025-01-01T00:00:00Z", Updated: "2026-04-01T00:00:00Z"},
			{ID: newID(), Name: "CDN Performance SLA", Service: "Content Delivery", Status: "active", UptimeTarget: "99.99", UptimeActual: "99.99", ResponseTime: "< 50ms global p95", CreditPolicy: "5x credit for SLA breach", PenaltyRate: "5%", ReviewPeriod: "monthly", LastReviewed: "2026-04-01", Created: "2025-06-01T00:00:00Z", Updated: "2026-04-01T00:00:00Z"},
		},
		"abrnoc": {
			{ID: newID(), Name: "EU Cloud Availability SLA", Service: "EU Infrastructure", Status: "active", UptimeTarget: "99.95", UptimeActual: "99.98", ResponseTime: "< 150ms", CreditPolicy: "10% credit per hour of downtime beyond SLA", PenaltyRate: "10%", ReviewPeriod: "monthly", LastReviewed: "2026-04-01", Created: "2025-01-01T00:00:00Z", Updated: "2026-04-01T00:00:00Z"},
			{ID: newID(), Name: "DNS Resolution SLA", Service: "DNS Infrastructure", Status: "breached", UptimeTarget: "99.999", UptimeActual: "99.990", ResponseTime: "< 10ms", CreditPolicy: "20% credit for any breach in reporting period", PenaltyRate: "20%", ReviewPeriod: "monthly", LastReviewed: "2026-04-01", Created: "2025-01-01T00:00:00Z", Updated: "2026-04-10T00:00:00Z"},
			{ID: newID(), Name: "GDPR Data Processing SLA", Service: "Data Platform", Status: "active", UptimeTarget: "99.99", UptimeActual: "100.00", ResponseTime: "< 72h breach notification", CreditPolicy: "Contract penalty per GDPR guidelines", PenaltyRate: "N/A", ReviewPeriod: "quarterly", LastReviewed: "2026-03-01", Created: "2025-05-01T00:00:00Z", Updated: "2026-03-01T00:00:00Z"},
		},
		"hosting": {
			{ID: newID(), Name: "GPU Cluster Availability SLA", Service: "GPU Compute", Status: "active", UptimeTarget: "99.90", UptimeActual: "99.93", ResponseTime: "< 500ms job start", CreditPolicy: "2x credit for missed GPU hours", PenaltyRate: "8%", ReviewPeriod: "monthly", LastReviewed: "2026-04-01", Created: "2025-08-01T00:00:00Z", Updated: "2026-04-01T00:00:00Z"},
			{ID: newID(), Name: "K8s Control Plane SLA", Service: "Managed Kubernetes", Status: "active", UptimeTarget: "99.95", UptimeActual: "99.96", ResponseTime: "< 200ms API server", CreditPolicy: "10% credit per 0.05% below target", PenaltyRate: "10%", ReviewPeriod: "monthly", LastReviewed: "2026-04-01", Created: "2025-08-01T00:00:00Z", Updated: "2026-04-01T00:00:00Z"},
		},
	},
}

func (s *slaStoreType) list(slug string) []SLAPolicy {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.policies[slug]
	out := make([]SLAPolicy, len(src))
	copy(out, src)
	return out
}

func (s *slaStoreType) getByID(slug, id string) (SLAPolicy, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, p := range s.policies[slug] {
		if p.ID == id {
			return p, true
		}
	}
	return SLAPolicy{}, false
}

func (s *slaStoreType) add(slug string, p SLAPolicy) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.policies[slug] = append(s.policies[slug], p)
}

func (s *slaStoreType) delete(slug, id string) {
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

func filterSLAPolicies(all []SLAPolicy, search, status string) []SLAPolicy {
	var out []SLAPolicy
	q := strings.ToLower(search)
	for _, p := range all {
		if status != "" && p.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(p.Name + " " + p.Service + " " + p.UptimeTarget)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, p)
	}
	return out
}

func dbSLAPolicyToLocal(p db.SLAPolicy) SLAPolicy {
	lastReviewed := ""
	if p.LastReviewed != nil {
		lastReviewed = p.LastReviewed.Format("2006-01-02")
	}
	return SLAPolicy{
		ID:           fmt.Sprintf("%d", p.ID),
		Name:         p.Name,
		Service:      p.Service,
		Status:       p.Status,
		UptimeTarget: fmt.Sprintf("%.2f", p.UptimeTarget),
		UptimeActual: fmt.Sprintf("%.2f", p.UptimeActual),
		ResponseTime: p.ResponseTime,
		CreditPolicy: p.CreditPolicy,
		PenaltyRate:  p.PenaltyRate,
		ReviewPeriod: p.ReviewPeriod,
		LastReviewed: lastReviewed,
		Created:      p.CreatedAt.Format(time.RFC3339),
		Updated:      p.UpdatedAt.Format(time.RFC3339),
	}
}

// ── GET /api/tenant/{slug}/sla ─────────────────────────────

type apiSLAPolicy struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Service      string `json:"service"`
	Status       string `json:"status"`
	UptimeTarget string `json:"uptime_target"`
	UptimeActual string `json:"uptime_actual"`
	ResponseTime string `json:"response_time"`
	CreditPolicy string `json:"credit_policy"`
	PenaltyRate  string `json:"penalty_rate"`
	ReviewPeriod string `json:"review_period"`
	LastReviewed string `json:"last_reviewed"`
	Created      string `json:"created"`
	Updated      string `json:"updated"`
}

type apiSLAListResponse struct {
	Policies []apiSLAPolicy `json:"policies"`
	Total    int            `json:"total"`
}

func apiTenantSLAListHandler(w http.ResponseWriter, r *http.Request) {
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
	status := r.URL.Query().Get("status")

	var policies []SLAPolicy
	if Pools != nil && tenant.DBID > 0 {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		dps, err := db.ListSLAPolicies(ctx, Pools.Admin, tenant.DBID, search, status)
		if err != nil {
			log.Printf("apiTenantSLAListHandler: %v", err)
			policies = filterSLAPolicies(slaStore.list(slug), search, status)
		} else {
			policies = make([]SLAPolicy, len(dps))
			for i, dp := range dps {
				policies[i] = dbSLAPolicyToLocal(dp)
			}
		}
	} else {
		policies = filterSLAPolicies(slaStore.list(slug), search, status)
	}

	out := make([]apiSLAPolicy, len(policies))
	for i, p := range policies {
		out[i] = apiSLAPolicy{
			ID: p.ID, Name: p.Name, Service: p.Service,
			Status: p.Status, UptimeTarget: p.UptimeTarget,
			UptimeActual: p.UptimeActual, ResponseTime: p.ResponseTime,
			CreditPolicy: p.CreditPolicy, PenaltyRate: p.PenaltyRate,
			ReviewPeriod: p.ReviewPeriod, LastReviewed: p.LastReviewed,
			Created: p.Created, Updated: p.Updated,
		}
	}

	apiOK(w, apiSLAListResponse{Policies: out, Total: len(out)})
}

// ── POST /api/tenant/{slug}/sla ────────────────────────────

type apiCreateSLARequest struct {
	Name         string `json:"name"`
	Service      string `json:"service"`
	UptimeTarget string `json:"uptime_target"`
	ResponseTime string `json:"response_time"`
	CreditPolicy string `json:"credit_policy"`
	PenaltyRate  string `json:"penalty_rate"`
	ReviewPeriod string `json:"review_period"`
}

func apiTenantSLACreateHandler(w http.ResponseWriter, r *http.Request) {
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

	var req apiCreateSLARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}

	name := strings.TrimSpace(req.Name)
	service := strings.TrimSpace(req.Service)
	uptimeTarget := strings.TrimSpace(req.UptimeTarget)
	reviewPeriod := strings.TrimSpace(req.ReviewPeriod)

	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	if uptimeTarget == "" {
		uptimeTarget = "99.95"
	}
	if reviewPeriod == "" {
		reviewPeriod = "monthly"
	}

	if Pools != nil && tenant.DBID > 0 {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		uptimeF, _ := strconv.ParseFloat(uptimeTarget, 64)
		now := time.Now()
		dp, err := db.CreateSLAPolicy(ctx, Pools.Admin, tenant.DBID, name, service, uptimeF,
			strings.TrimSpace(req.ResponseTime), strings.TrimSpace(req.CreditPolicy),
			strings.TrimSpace(req.PenaltyRate), reviewPeriod, &now)
		if err != nil {
			log.Printf("apiTenantSLACreateHandler: %v", err)
			apiError(w, http.StatusInternalServerError, "Failed to create SLA policy.")
			return
		}
		p := dbSLAPolicyToLocal(dp)
		apiCreated(w, apiSLAPolicy{
			ID: p.ID, Name: p.Name, Service: p.Service,
			Status: p.Status, UptimeTarget: p.UptimeTarget,
			UptimeActual: p.UptimeActual, ResponseTime: p.ResponseTime,
			CreditPolicy: p.CreditPolicy, PenaltyRate: p.PenaltyRate,
			ReviewPeriod: p.ReviewPeriod, LastReviewed: p.LastReviewed,
			Created: p.Created, Updated: p.Updated,
		})
		return
	}

	p := SLAPolicy{
		ID: newID(), Name: name, Service: service,
		Status: "active", UptimeTarget: uptimeTarget, UptimeActual: "100.00",
		ResponseTime: strings.TrimSpace(req.ResponseTime),
		CreditPolicy: strings.TrimSpace(req.CreditPolicy),
		PenaltyRate:  strings.TrimSpace(req.PenaltyRate),
		ReviewPeriod: reviewPeriod, LastReviewed: time.Now().Format("2006-01-02"),
		Created: time.Now().Format(time.RFC3339), Updated: time.Now().Format(time.RFC3339),
	}
	slaStore.add(slug, p)

	apiCreated(w, apiSLAPolicy{
		ID: p.ID, Name: p.Name, Service: p.Service,
		Status: p.Status, UptimeTarget: p.UptimeTarget,
		UptimeActual: p.UptimeActual, ResponseTime: p.ResponseTime,
		CreditPolicy: p.CreditPolicy, PenaltyRate: p.PenaltyRate,
		ReviewPeriod: p.ReviewPeriod, LastReviewed: p.LastReviewed,
		Created: p.Created, Updated: p.Updated,
	})
}

// ── GET /api/tenant/{slug}/sla/{id} ────────────────────────

func apiTenantSLAGetHandler(w http.ResponseWriter, r *http.Request) {
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
			dp, err := db.GetSLAPolicy(ctx, Pools.Admin, tenant.DBID, id)
			if err != nil {
				apiError(w, http.StatusNotFound, "SLA policy not found.")
				return
			}
			p := dbSLAPolicyToLocal(dp)
			apiOK(w, apiSLAPolicy{
				ID: p.ID, Name: p.Name, Service: p.Service,
				Status: p.Status, UptimeTarget: p.UptimeTarget,
				UptimeActual: p.UptimeActual, ResponseTime: p.ResponseTime,
				CreditPolicy: p.CreditPolicy, PenaltyRate: p.PenaltyRate,
				ReviewPeriod: p.ReviewPeriod, LastReviewed: p.LastReviewed,
				Created: p.Created, Updated: p.Updated,
			})
			return
		}
	}

	p, ok := slaStore.getByID(slug, idStr)
	if !ok {
		apiError(w, http.StatusNotFound, "SLA policy not found.")
		return
	}

	apiOK(w, apiSLAPolicy{
		ID: p.ID, Name: p.Name, Service: p.Service,
		Status: p.Status, UptimeTarget: p.UptimeTarget,
		UptimeActual: p.UptimeActual, ResponseTime: p.ResponseTime,
		CreditPolicy: p.CreditPolicy, PenaltyRate: p.PenaltyRate,
		ReviewPeriod: p.ReviewPeriod, LastReviewed: p.LastReviewed,
		Created: p.Created, Updated: p.Updated,
	})
}

// ── DELETE /api/tenant/{slug}/sla/{id} ─────────────────────

func apiTenantSLADeleteHandler(w http.ResponseWriter, r *http.Request) {
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
			if err := db.DeleteSLAPolicy(ctx, Pools.Admin, tenant.DBID, id); err != nil {
				log.Printf("apiTenantSLADeleteHandler: %v", err)
				apiError(w, http.StatusInternalServerError, "Failed to delete SLA policy.")
				return
			}
			apiOK(w, map[string]any{"ok": true, "message": "SLA policy deleted."})
			return
		}
	}

	if _, ok := slaStore.getByID(slug, idStr); !ok {
		apiError(w, http.StatusNotFound, "SLA policy not found.")
		return
	}

	slaStore.delete(slug, idStr)
	apiOK(w, map[string]any{
		"ok":      true,
		"message": "SLA policy deleted.",
	})
}
