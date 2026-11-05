package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Security API — IAM: Role-based Access Control
// security/iam/rbac
// ════════════════════════════════════════════════════════════════

type RBACRole struct {
	ID          string
	Name        string
	Type        string
	Users       int
	Permissions string
	Status      string
	Created     string
	Region      string
}

type rbacRoleStoreType struct {
	mu    sync.RWMutex
	roles map[string][]RBACRole
}

var rbacRoleStore = &rbacRoleStoreType{
	roles: map[string][]RBACRole{
		"buyvm": {
			{ID: newID(), Name: "Admin", Type: "Built-in", Users: 4, Permissions: "all (284)", Status: "active", Created: "2025-08-12", Region: "eu-central-1"},
			{ID: newID(), Name: "DevOps Engineer", Type: "Custom", Users: 14, Permissions: "compute, networking, observability", Status: "active", Created: "2025-09-01", Region: "eu-central-1"},
			{ID: newID(), Name: "Billing Viewer", Type: "Custom", Users: 6, Permissions: "billing.read only", Status: "active", Created: "2025-10-15", Region: "eu-central-1"},
			{ID: newID(), Name: "Support Agent", Type: "Custom", Users: 8, Permissions: "tickets.* + clients.read", Status: "active", Created: "2025-11-01", Region: "eu-central-1"},
			{ID: newID(), Name: "Read-only Auditor", Type: "Custom", Users: 2, Permissions: "*.read across all", Status: "active", Created: "2026-01-10", Region: "eu-central-1"},
		},
		"hosting": {
			{ID: newID(), Name: "Admin", Type: "Built-in", Users: 2, Permissions: "all (284)", Status: "active", Created: "2025-10-01", Region: "ap-south-1"},
			{ID: newID(), Name: "Developer", Type: "Custom", Users: 18, Permissions: "compute, databases, devops", Status: "active", Created: "2025-10-15", Region: "ap-south-1"},
			{ID: newID(), Name: "Viewer", Type: "Built-in", Users: 10, Permissions: "*.read", Status: "active", Created: "2025-10-01", Region: "ap-south-1"},
		},
	},
}

func (s *rbacRoleStoreType) list(slug string) []RBACRole {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.roles[slug]
	out := make([]RBACRole, len(src))
	copy(out, src)
	return out
}

func (s *rbacRoleStoreType) getByID(slug, id string) (RBACRole, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, r := range s.roles[slug] {
		if r.ID == id {
			return r, true
		}
	}
	return RBACRole{}, false
}

func (s *rbacRoleStoreType) add(slug string, role RBACRole) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.roles[slug] = append(s.roles[slug], role)
}

func (s *rbacRoleStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.roles[slug]
	for i, r := range list {
		if r.ID == id {
			s.roles[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiRBACRole struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Users       int    `json:"users"`
	Permissions string `json:"permissions"`
	Status      string `json:"status"`
	Created     string `json:"created"`
	Region      string `json:"region"`
}

type apiRBACRoleList struct {
	Roles []apiRBACRole `json:"roles"`
	Total int           `json:"total"`
}

func toAPIRBACRole(r RBACRole) apiRBACRole {
	return apiRBACRole{
		ID: r.ID, Name: r.Name, Type: r.Type,
		Users: r.Users, Permissions: r.Permissions,
		Status: r.Status, Created: r.Created,
		Region: r.Region,
	}
}

func filterRBACRoles(all []RBACRole, search, status string) []RBACRole {
	var out []RBACRole
	q := strings.ToLower(search)
	for _, r := range all {
		if status != "" && r.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(r.Name + " " + r.Type + " " + r.Permissions + " " + r.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, r)
	}
	return out
}

func apiWSSecIAMRBACListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	roles := filterRBACRoles(rbacRoleStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiRBACRole, len(roles))
	for i, role := range roles {
		out[i] = toAPIRBACRole(role)
	}
	apiOK(w, apiRBACRoleList{Roles: out, Total: len(out)})
}

type apiCreateRBACRoleRequest struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Permissions string `json:"permissions"`
	Region      string `json:"region"`
}

func apiWSSecIAMRBACCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateRBACRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	typ := strings.TrimSpace(req.Type)
	if typ == "" {
		typ = "Custom"
	}
	permissions := strings.TrimSpace(req.Permissions)
	if permissions == "" {
		apiError(w, http.StatusBadRequest, "Permissions is required.")
		return
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "eu-central-1"
	}
	role := RBACRole{
		ID: newID(), Name: name, Type: typ,
		Users: 0, Permissions: permissions,
		Status: "active", Created: "2026-04-17",
		Region: region,
	}
	rbacRoleStore.add(slug, role)
	apiCreated(w, toAPIRBACRole(role))
}

func apiWSSecIAMRBACGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	role, found := rbacRoleStore.getByID(slug, r.PathValue("rid"))
	if !found {
		apiError(w, http.StatusNotFound, "RBAC role not found.")
		return
	}
	apiOK(w, toAPIRBACRole(role))
}

func apiWSSecIAMRBACDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := rbacRoleStore.getByID(slug, r.PathValue("rid")); !found {
		apiError(w, http.StatusNotFound, "RBAC role not found.")
		return
	}
	rbacRoleStore.delete(slug, r.PathValue("rid"))
	apiOK(w, map[string]any{"ok": true, "message": "RBAC role deleted."})
}
