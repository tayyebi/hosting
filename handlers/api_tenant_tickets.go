package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
)


// ── GET /api/tenant/{slug}/tickets ──────────────────────────

type apiTicket struct {
	ID       string `json:"id"`
	Subject  string `json:"subject"`
	Client   string `json:"client"`
	Status   string `json:"status"`
	Priority string `json:"priority"`
	Created  string `json:"created"`
	Updated  string `json:"updated"`
}

type apiTicketsListResponse struct {
	Tickets []apiTicket `json:"tickets"`
	Total   int         `json:"total"`
	Limit   int         `json:"limit"`
	Offset  int         `json:"offset"`
}

func apiTenantTicketsListHandler(w http.ResponseWriter, r *http.Request) {
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
	status := r.URL.Query().Get("status")
	priority := r.URL.Query().Get("priority")

	limit, offset := parsePagination(r, 50, 200)
	tickets := filterTickets(ticketStore.list(slug), search, status, priority)
	total := len(tickets)
	if offset >= total {
		tickets = []Ticket{}
	} else {
		tickets = tickets[offset:]
		if len(tickets) > limit {
			tickets = tickets[:limit]
		}
	}

	out := make([]apiTicket, len(tickets))
	for i, t := range tickets {
		out[i] = apiTicket{
			ID: t.ID, Subject: t.Subject, Client: t.Client,
			Status: t.Status, Priority: t.Priority,
			Created: t.Created, Updated: t.Updated,
		}
	}

	apiOK(w, apiTicketsListResponse{Tickets: out, Total: total, Limit: limit, Offset: offset})
}

// ── POST /api/tenant/{slug}/tickets ─────────────────────────

type apiCreateTicketRequest struct {
	Subject  string `json:"subject"`
	Client   string `json:"client"`
	Priority string `json:"priority"`
}

func apiTenantTicketsCreateHandler(w http.ResponseWriter, r *http.Request) {
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

	var req apiCreateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}

	subject := strings.TrimSpace(req.Subject)
	client := strings.TrimSpace(req.Client)
	priority := strings.TrimSpace(req.Priority)

	if subject == "" || client == "" {
		apiError(w, http.StatusBadRequest, "Subject and client are required.")
		return
	}
	if priority == "" {
		priority = "medium"
	}

	t := Ticket{
		ID: newID(), Subject: subject, Client: client,
		Status: "open", Priority: priority,
		Created: "2026-04-17T00:00:00Z", Updated: "2026-04-17T00:00:00Z",
	}
	ticketStore.add(slug, t)

	apiCreated(w, apiTicket{
		ID: t.ID, Subject: t.Subject, Client: t.Client,
		Status: t.Status, Priority: t.Priority,
		Created: t.Created, Updated: t.Updated,
	})
}

// ── GET /api/tenant/{slug}/tickets/{id} ─────────────────────

func apiTenantTicketsGetHandler(w http.ResponseWriter, r *http.Request) {
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
	t, ok := ticketStore.getByID(slug, id)
	if !ok {
		apiError(w, http.StatusNotFound, "Ticket not found.")
		return
	}

	apiOK(w, apiTicket{
		ID: t.ID, Subject: t.Subject, Client: t.Client,
		Status: t.Status, Priority: t.Priority,
		Created: t.Created, Updated: t.Updated,
	})
}

// ── DELETE /api/tenant/{slug}/tickets/{id} ───────────────────

func apiTenantTicketsDeleteHandler(w http.ResponseWriter, r *http.Request) {
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
	if _, ok := ticketStore.getByID(slug, id); !ok {
		apiError(w, http.StatusNotFound, "Ticket not found.")
		return
	}

	ticketStore.delete(slug, id)
	apiOK(w, map[string]any{
		"ok":      true,
		"message": "Ticket deleted.",
	})
}
