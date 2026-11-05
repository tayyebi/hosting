package handlers

import (
	"encoding/json"
	"net/http"
	"net/mail"
	"strings"
)

// ── GET /api/tenant/{slug}/clients ────────────────────────

type apiClientWorkspace struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type apiClient struct {
	ID         string               `json:"id"`
	Name       string               `json:"name"`
	Email      string               `json:"email"`
	Company    string               `json:"company,omitempty"`
	Status     string               `json:"status"`
	Workspaces []apiClientWorkspace `json:"workspaces"`
}

type apiClientsListResponse struct {
	Clients []apiClient `json:"clients"`
	Total   int         `json:"total"`
	Limit   int         `json:"limit"`
	Offset  int         `json:"offset"`
}

func apiTenantClientsListHandler(w http.ResponseWriter, r *http.Request) {
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
	limit, offset := parsePagination(r, 50, 200)

	clients := filterClients(clientStore.list(slug), search, status)
	total := len(clients)
	if offset >= total {
		clients = []Client{}
	} else {
		clients = clients[offset:]
		if len(clients) > limit {
			clients = clients[:limit]
		}
	}

	out := make([]apiClient, len(clients))
	for i, c := range clients {
		ws := make([]apiClientWorkspace, len(c.Workspaces))
		for j, w := range c.Workspaces {
			ws[j] = apiClientWorkspace{ID: w.ID, Name: w.Name}
		}
		out[i] = apiClient{
			ID: c.ID, Name: c.Name, Email: c.Email,
			Company: c.Company, Status: c.Status, Workspaces: ws,
		}
	}

	apiOK(w, apiClientsListResponse{Clients: out, Total: total, Limit: limit, Offset: offset})
}

// ── POST /api/tenant/{slug}/clients ───────────────────────

type apiCreateClientRequest struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Company string `json:"company"`
}

func apiTenantClientsCreateHandler(w http.ResponseWriter, r *http.Request) {
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

	var req apiCreateClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}

	name := strings.TrimSpace(req.Name)
	email := strings.TrimSpace(req.Email)
	company := strings.TrimSpace(req.Company)

	fieldErrs := map[string]string{}
	if name == "" {
		fieldErrs["name"] = "Name is required."
	} else if len(name) > 100 {
		fieldErrs["name"] = "Name must be 100 characters or fewer."
	}
	if email == "" {
		fieldErrs["email"] = "Email is required."
	} else if len(email) > 254 {
		fieldErrs["email"] = "Email must be 254 characters or fewer."
	} else if _, err := mail.ParseAddress(email); err != nil {
		fieldErrs["email"] = "Email address is not valid."
	}
	if len(company) > 100 {
		fieldErrs["company"] = "Company must be 100 characters or fewer."
	}
	if len(fieldErrs) > 0 {
		apiValidationError(w, fieldErrs)
		return
	}

	c := Client{
		ID: newID(), Name: name, Email: email, Company: company,
		Status:     "active",
		Workspaces: mkWorkspaces("default"),
	}
	clientStore.add(slug, c)

	ws := make([]apiClientWorkspace, len(c.Workspaces))
	for j, w := range c.Workspaces {
		ws[j] = apiClientWorkspace{ID: w.ID, Name: w.Name}
	}

	apiCreated(w, apiClient{
		ID: c.ID, Name: c.Name, Email: c.Email,
		Company: c.Company, Status: c.Status, Workspaces: ws,
	})
}

// ── GET /api/tenant/{slug}/clients/{id} ───────────────────

func apiTenantClientsGetHandler(w http.ResponseWriter, r *http.Request) {
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
	c, ok := clientStore.getByID(slug, id)
	if !ok {
		apiError(w, http.StatusNotFound, "Client not found.")
		return
	}

	ws := make([]apiClientWorkspace, len(c.Workspaces))
	for j, w := range c.Workspaces {
		ws[j] = apiClientWorkspace{ID: w.ID, Name: w.Name}
	}

	apiOK(w, apiClient{
		ID: c.ID, Name: c.Name, Email: c.Email,
		Company: c.Company, Status: c.Status, Workspaces: ws,
	})
}

// ── DELETE /api/tenant/{slug}/clients/{id} ────────────────

func apiTenantClientsDeleteHandler(w http.ResponseWriter, r *http.Request) {
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
	if _, ok := clientStore.getByID(slug, id); !ok {
		apiError(w, http.StatusNotFound, "Client not found.")
		return
	}

	clientStore.delete(slug, id)
	apiOK(w, map[string]any{
		"ok":      true,
		"message": "Client deleted.",
	})
}
