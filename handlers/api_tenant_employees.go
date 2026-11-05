package handlers

import (
	"encoding/json"
	"net/http"
	"net/mail"
	"strings"
)

// ── GET /api/tenant/{slug}/employees ────────────────────────

type apiEmployeePermissions struct {
	ID      string `json:"id"`
	Label   string `json:"label"`
	Granted bool   `json:"granted"`
}

type apiEmployee struct {
	ID          string                   `json:"id"`
	Name        string                   `json:"name"`
	Email       string                   `json:"email"`
	Role        string                   `json:"role"`
	Active      bool                     `json:"active"`
	Permissions []apiEmployeePermissions `json:"permissions"`
}

type apiEmployeesListResponse struct {
	Employees []apiEmployee `json:"employees"`
	Total     int           `json:"total"`
	Limit     int           `json:"limit"`
	Offset    int           `json:"offset"`
}

func employeeToAPI(e Employee) apiEmployee {
	perms := make([]apiEmployeePermissions, 0, len(allPermIDs()))
	for _, g := range permissionGroups {
		for _, item := range g.Items {
			perms = append(perms, apiEmployeePermissions{
				ID:      item.ID,
				Label:   item.Label,
				Granted: e.Permissions[item.ID],
			})
		}
	}
	return apiEmployee{
		ID:          e.ID,
		Name:        e.Name,
		Email:       e.Email,
		Role:        e.Role,
		Active:      e.Active,
		Permissions: perms,
	}
}

func apiTenantEmployeesListHandler(w http.ResponseWriter, r *http.Request) {
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
	role := r.URL.Query().Get("role")
	status := r.URL.Query().Get("status")

	employees := employeeStore.list(slug)

	// Filter
	var filtered []Employee
	q := strings.ToLower(search)
	for _, e := range employees {
		if status == "active" && !e.Active {
			continue
		}
		if status == "inactive" && e.Active {
			continue
		}
		if role != "" && !strings.EqualFold(e.Role, role) {
			continue
		}
		if q != "" {
			hay := strings.ToLower(e.Name + " " + e.Email + " " + e.Role)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		filtered = append(filtered, e)
	}

	limit, offset := parsePagination(r, 50, 200)
	total := len(filtered)
	if offset >= total {
		filtered = []Employee{}
	} else {
		filtered = filtered[offset:]
		if len(filtered) > limit {
			filtered = filtered[:limit]
		}
	}

	out := make([]apiEmployee, len(filtered))
	for i, e := range filtered {
		out[i] = employeeToAPI(e)
	}

	apiOK(w, apiEmployeesListResponse{Employees: out, Total: total, Limit: limit, Offset: offset})
}

// ── POST /api/tenant/{slug}/employees ───────────────────────

type apiCreateEmployeeRequest struct {
	Name string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

func apiTenantEmployeesCreateHandler(w http.ResponseWriter, r *http.Request) {
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

	var req apiCreateEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}

	name := strings.TrimSpace(req.Name)
	email := strings.TrimSpace(req.Email)
	role := strings.TrimSpace(req.Role)

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
	if len(role) > 100 {
		fieldErrs["role"] = "Role must be 100 characters or fewer."
	}
	if len(fieldErrs) > 0 {
		apiValidationError(w, fieldErrs)
		return
	}

	e := Employee{
		ID: newID(), Name: name, Email: email, Role: role,
		Active:      true,
		Permissions: map[string]bool{"dashboard": true},
	}
	employeeStore.add(slug, e)

	apiCreated(w, employeeToAPI(e))
}

// ── GET /api/tenant/{slug}/employees/{id} ───────────────────

func apiTenantEmployeesGetHandler(w http.ResponseWriter, r *http.Request) {
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
	e, ok := employeeStore.getByID(slug, id)
	if !ok {
		apiError(w, http.StatusNotFound, "Employee not found.")
		return
	}

	apiOK(w, employeeToAPI(e))
}

// ── DELETE /api/tenant/{slug}/employees/{id} ────────────────

func apiTenantEmployeesDeleteHandler(w http.ResponseWriter, r *http.Request) {
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
	if _, ok := employeeStore.getByID(slug, id); !ok {
		apiError(w, http.StatusNotFound, "Employee not found.")
		return
	}

	employeeStore.delete(slug, id)
	apiOK(w, map[string]any{
		"ok":      true,
		"message": "Employee deleted.",
	})
}
