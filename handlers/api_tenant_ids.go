package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
)

// ── GET /api/tenant/{slug}/ids ──────────────────────────────

type apiIDSEvent struct {
	ID        string `json:"id"`
	EventType string `json:"event_type"`
	Severity  string `json:"severity"`
	SourceIP  string `json:"source_ip"`
	Target    string `json:"target"`
	Status    string `json:"status"`
	RuleID    string `json:"rule_id"`
	Created   string `json:"created"`
	Updated   string `json:"updated"`
}

type apiIDSEventsListResponse struct {
	Events []apiIDSEvent `json:"events"`
	Total  int           `json:"total"`
}

func apiTenantIDSListHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok {
		apiError(w, http.StatusUnauthorized, "Authentication required.")
		return
	}

	slug := r.PathValue("slug")
	if _, found := store.getBySlug(slug); !found {
		apiError(w, http.StatusNotFound, "Tenant not found.")
		return
	}
	if user.Role != RoleSysAdmin && (user.Role != RoleTenant || user.TenantSlug != slug) {
		apiError(w, http.StatusForbidden, "Access denied.")
		return
	}

	search := r.URL.Query().Get("q")
	severity := r.URL.Query().Get("severity")
	status := r.URL.Query().Get("status")

	events := filterIDSEvents(idsEventStore.list(slug), search, severity, status)

	out := make([]apiIDSEvent, len(events))
	for i, e := range events {
		out[i] = apiIDSEvent{
			ID: e.ID, EventType: e.EventType, Severity: e.Severity,
			SourceIP: e.SourceIP, Target: e.Target,
			Status: e.Status, RuleID: e.RuleID,
			Created: e.Created, Updated: e.Updated,
		}
	}

	apiOK(w, apiIDSEventsListResponse{Events: out, Total: len(out)})
}

// ── POST /api/tenant/{slug}/ids ─────────────────────────────

type apiCreateIDSEventRequest struct {
	EventType string `json:"event_type"`
	Severity  string `json:"severity"`
	SourceIP  string `json:"source_ip"`
	Target    string `json:"target"`
	RuleID    string `json:"rule_id"`
}

func apiTenantIDSCreateHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok {
		apiError(w, http.StatusUnauthorized, "Authentication required.")
		return
	}

	slug := r.PathValue("slug")
	if _, found := store.getBySlug(slug); !found {
		apiError(w, http.StatusNotFound, "Tenant not found.")
		return
	}
	if user.Role != RoleSysAdmin && (user.Role != RoleTenant || user.TenantSlug != slug) {
		apiError(w, http.StatusForbidden, "Access denied.")
		return
	}

	var req apiCreateIDSEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}

	eventType := strings.TrimSpace(req.EventType)
	sourceIP := strings.TrimSpace(req.SourceIP)
	target := strings.TrimSpace(req.Target)
	severity := strings.TrimSpace(req.Severity)
	ruleID := strings.TrimSpace(req.RuleID)

	if eventType == "" || sourceIP == "" || target == "" {
		apiError(w, http.StatusBadRequest, "Event type, source IP, and target are required.")
		return
	}
	if severity == "" {
		severity = "medium"
	}
	if ruleID == "" {
		ruleID = "CUSTOM-001"
	}

	e := IDSEvent{
		ID: newID(), EventType: eventType, Severity: severity,
		SourceIP: sourceIP, Target: target,
		Status: "open", RuleID: ruleID,
		Created: "2026-04-17T00:00:00Z", Updated: "2026-04-17T00:00:00Z",
	}
	idsEventStore.add(slug, e)

	apiCreated(w, apiIDSEvent{
		ID: e.ID, EventType: e.EventType, Severity: e.Severity,
		SourceIP: e.SourceIP, Target: e.Target,
		Status: e.Status, RuleID: e.RuleID,
		Created: e.Created, Updated: e.Updated,
	})
}

// ── GET /api/tenant/{slug}/ids/{id} ─────────────────────────

func apiTenantIDSGetHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok {
		apiError(w, http.StatusUnauthorized, "Authentication required.")
		return
	}

	slug := r.PathValue("slug")
	if _, found := store.getBySlug(slug); !found {
		apiError(w, http.StatusNotFound, "Tenant not found.")
		return
	}
	if user.Role != RoleSysAdmin && (user.Role != RoleTenant || user.TenantSlug != slug) {
		apiError(w, http.StatusForbidden, "Access denied.")
		return
	}

	id := r.PathValue("id")
	e, ok := idsEventStore.getByID(slug, id)
	if !ok {
		apiError(w, http.StatusNotFound, "IDS event not found.")
		return
	}

	apiOK(w, apiIDSEvent{
		ID: e.ID, EventType: e.EventType, Severity: e.Severity,
		SourceIP: e.SourceIP, Target: e.Target,
		Status: e.Status, RuleID: e.RuleID,
		Created: e.Created, Updated: e.Updated,
	})
}

// ── DELETE /api/tenant/{slug}/ids/{id} ──────────────────────

func apiTenantIDSDeleteHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok {
		apiError(w, http.StatusUnauthorized, "Authentication required.")
		return
	}

	slug := r.PathValue("slug")
	if _, found := store.getBySlug(slug); !found {
		apiError(w, http.StatusNotFound, "Tenant not found.")
		return
	}
	if user.Role != RoleSysAdmin && (user.Role != RoleTenant || user.TenantSlug != slug) {
		apiError(w, http.StatusForbidden, "Access denied.")
		return
	}

	id := r.PathValue("id")
	if _, ok := idsEventStore.getByID(slug, id); !ok {
		apiError(w, http.StatusNotFound, "IDS event not found.")
		return
	}

	idsEventStore.delete(slug, id)
	apiOK(w, map[string]any{
		"ok":      true,
		"message": "IDS event deleted.",
	})
}
