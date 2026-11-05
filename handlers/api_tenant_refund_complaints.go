package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
)

// ── GET /api/tenant/{slug}/refund-complaints ────────────────

type apiRefundComplaint struct {
	ID        string `json:"id"`
	Client    string `json:"client"`
	Amount    string `json:"amount"`
	Currency  string `json:"currency"`
	Reason    string `json:"reason"`
	Status    string `json:"status"`
	InvoiceID string `json:"invoice_id"`
	Created   string `json:"created"`
	Updated   string `json:"updated"`
}

type apiRefundComplaintsListResponse struct {
	Complaints []apiRefundComplaint `json:"complaints"`
	Total      int                  `json:"total"`
}

func apiTenantRefundComplaintsListHandler(w http.ResponseWriter, r *http.Request) {
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

	complaints := filterRefundComplaints(refundComplaintStore.list(slug), search, status)

	out := make([]apiRefundComplaint, len(complaints))
	for i, c := range complaints {
		out[i] = apiRefundComplaint{
			ID: c.ID, Client: c.Client, Amount: c.Amount,
			Currency: c.Currency, Reason: c.Reason, Status: c.Status,
			InvoiceID: c.InvoiceID, Created: c.Created, Updated: c.Updated,
		}
	}

	apiOK(w, apiRefundComplaintsListResponse{Complaints: out, Total: len(out)})
}

// ── POST /api/tenant/{slug}/refund-complaints ───────────────

type apiCreateRefundComplaintRequest struct {
	Client    string `json:"client"`
	Amount    string `json:"amount"`
	Currency  string `json:"currency"`
	Reason    string `json:"reason"`
	InvoiceID string `json:"invoice_id"`
}

func apiTenantRefundComplaintsCreateHandler(w http.ResponseWriter, r *http.Request) {
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

	var req apiCreateRefundComplaintRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}

	client := strings.TrimSpace(req.Client)
	amount := strings.TrimSpace(req.Amount)
	currency := strings.TrimSpace(req.Currency)
	reason := strings.TrimSpace(req.Reason)
	invoiceID := strings.TrimSpace(req.InvoiceID)

	if client == "" || amount == "" || reason == "" {
		apiError(w, http.StatusBadRequest, "Client, amount, and reason are required.")
		return
	}
	if currency == "" {
		currency = "USD"
	}

	c := RefundComplaint{
		ID: newID(), Client: client, Amount: amount, Currency: currency,
		Reason: reason, Status: "pending", InvoiceID: invoiceID,
		Created: "2026-04-17T00:00:00Z", Updated: "2026-04-17T00:00:00Z",
	}
	refundComplaintStore.add(slug, c)

	apiCreated(w, apiRefundComplaint{
		ID: c.ID, Client: c.Client, Amount: c.Amount,
		Currency: c.Currency, Reason: c.Reason, Status: c.Status,
		InvoiceID: c.InvoiceID, Created: c.Created, Updated: c.Updated,
	})
}

// ── GET /api/tenant/{slug}/refund-complaints/{id} ───────────

func apiTenantRefundComplaintsGetHandler(w http.ResponseWriter, r *http.Request) {
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
	c, ok := refundComplaintStore.getByID(slug, id)
	if !ok {
		apiError(w, http.StatusNotFound, "Refund complaint not found.")
		return
	}

	apiOK(w, apiRefundComplaint{
		ID: c.ID, Client: c.Client, Amount: c.Amount,
		Currency: c.Currency, Reason: c.Reason, Status: c.Status,
		InvoiceID: c.InvoiceID, Created: c.Created, Updated: c.Updated,
	})
}

// ── DELETE /api/tenant/{slug}/refund-complaints/{id} ────────

func apiTenantRefundComplaintsDeleteHandler(w http.ResponseWriter, r *http.Request) {
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
	if _, ok := refundComplaintStore.getByID(slug, id); !ok {
		apiError(w, http.StatusNotFound, "Refund complaint not found.")
		return
	}

	refundComplaintStore.delete(slug, id)
	apiOK(w, map[string]any{
		"ok":      true,
		"message": "Refund complaint deleted.",
	})
}
