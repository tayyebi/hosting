package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ── Resource store ──────────────────────────────────────────

type Resource struct {
	ID     string
	Name   string
	Type   string // "vm" | "container" | "database" | "cdn" | "cache" | "worker" | "lb"
	Status string // "online" | "degraded" | "offline" | "maintenance"
	Region string
	CPU    string
	Memory string
	Created string
	Updated string
}

type resourceStoreType struct {
	mu        sync.RWMutex
	resources map[string][]Resource
}

var resourceStore = &resourceStoreType{
	resources: map[string][]Resource{
		"buyvm": {
			{ID: newID(), Name: "web-prod-01", Type: "vm", Status: "online", Region: "us-east", CPU: "38%", Memory: "61%", Created: "2026-01-10T08:00:00Z", Updated: "2026-04-17T14:00:00Z"},
			{ID: newID(), Name: "web-prod-02", Type: "vm", Status: "online", Region: "us-east", CPU: "41%", Memory: "58%", Created: "2026-01-10T08:00:00Z", Updated: "2026-04-17T14:00:00Z"},
			{ID: newID(), Name: "api-gateway", Type: "lb", Status: "degraded", Region: "us-east", CPU: "67%", Memory: "74%", Created: "2025-11-03T12:00:00Z", Updated: "2026-04-17T13:30:00Z"},
			{ID: newID(), Name: "db-primary", Type: "database", Status: "online", Region: "us-east", CPU: "22%", Memory: "81%", Created: "2025-09-15T10:00:00Z", Updated: "2026-04-17T14:00:00Z"},
			{ID: newID(), Name: "cache-redis-01", Type: "cache", Status: "online", Region: "us-east", CPU: "8%", Memory: "44%", Created: "2025-12-01T09:00:00Z", Updated: "2026-04-17T14:00:00Z"},
			{ID: newID(), Name: "worker-queue-1", Type: "worker", Status: "degraded", Region: "us-east", CPU: "88%", Memory: "69%", Created: "2026-02-20T11:00:00Z", Updated: "2026-04-17T13:45:00Z"},
			{ID: newID(), Name: "cdn-edge-us", Type: "cdn", Status: "online", Region: "us-east", CPU: "—", Memory: "—", Created: "2025-10-05T14:00:00Z", Updated: "2026-04-17T14:00:00Z"},
		},
		"abrnoc": {
			{ID: newID(), Name: "app-server-01", Type: "vm", Status: "online", Region: "eu-west", CPU: "32%", Memory: "55%", Created: "2026-02-01T10:00:00Z", Updated: "2026-04-17T14:00:00Z"},
			{ID: newID(), Name: "db-cluster-eu", Type: "database", Status: "online", Region: "eu-west", CPU: "18%", Memory: "72%", Created: "2026-01-15T08:00:00Z", Updated: "2026-04-17T14:00:00Z"},
			{ID: newID(), Name: "lb-eu-west", Type: "lb", Status: "online", Region: "eu-west", CPU: "12%", Memory: "30%", Created: "2026-03-01T09:00:00Z", Updated: "2026-04-17T14:00:00Z"},
		},
		"hosting": {
			{ID: newID(), Name: "k8s-node-01", Type: "container", Status: "online", Region: "ap-south", CPU: "45%", Memory: "62%", Created: "2026-03-10T06:00:00Z", Updated: "2026-04-17T14:00:00Z"},
			{ID: newID(), Name: "k8s-node-02", Type: "container", Status: "online", Region: "ap-south", CPU: "52%", Memory: "58%", Created: "2026-03-10T06:00:00Z", Updated: "2026-04-17T14:00:00Z"},
			{ID: newID(), Name: "db-mongo-prod", Type: "database", Status: "maintenance", Region: "ap-south", CPU: "0%", Memory: "0%", Created: "2025-11-20T12:00:00Z", Updated: "2026-04-16T22:00:00Z"},
			{ID: newID(), Name: "cdn-edge-ap", Type: "cdn", Status: "online", Region: "ap-south", CPU: "—", Memory: "—", Created: "2026-01-05T08:00:00Z", Updated: "2026-04-17T14:00:00Z"},
		},
	},
}

func (s *resourceStoreType) list(slug string) []Resource {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.resources[slug]
	out := make([]Resource, len(src))
	copy(out, src)
	return out
}

func (s *resourceStoreType) getByID(slug, id string) (Resource, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, r := range s.resources[slug] {
		if r.ID == id {
			return r, true
		}
	}
	return Resource{}, false
}

func (s *resourceStoreType) add(slug string, r Resource) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resources[slug] = append(s.resources[slug], r)
}

func (s *resourceStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.resources[slug]
	for i, r := range list {
		if r.ID == id {
			s.resources[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

func filterResources(all []Resource, search, status, rtype string) []Resource {
	var out []Resource
	q := strings.ToLower(search)
	for _, r := range all {
		if status != "" && r.Status != status {
			continue
		}
		if rtype != "" && r.Type != rtype {
			continue
		}
		if q != "" {
			hay := strings.ToLower(r.Name + " " + r.Type + " " + r.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, r)
	}
	return out
}

// ── GET /api/tenant/{slug}/resources ────────────────────────

type apiResource struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Status  string `json:"status"`
	Region  string `json:"region"`
	CPU     string `json:"cpu"`
	Memory  string `json:"memory"`
	Created string `json:"created"`
	Updated string `json:"updated"`
}

type apiResourcesListResponse struct {
	Resources []apiResource `json:"resources"`
	Total     int           `json:"total"`
}

func apiTenantResourcesListHandler(w http.ResponseWriter, r *http.Request) {
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
	rtype := r.URL.Query().Get("type")

	resources := filterResources(resourceStore.list(slug), search, status, rtype)

	out := make([]apiResource, len(resources))
	for i, res := range resources {
		out[i] = apiResource{
			ID: res.ID, Name: res.Name, Type: res.Type,
			Status: res.Status, Region: res.Region,
			CPU: res.CPU, Memory: res.Memory,
			Created: res.Created, Updated: res.Updated,
		}
	}

	apiOK(w, apiResourcesListResponse{Resources: out, Total: len(out)})
}

// ── POST /api/tenant/{slug}/resources ───────────────────────

type apiCreateResourceRequest struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Region string `json:"region"`
}

func apiTenantResourcesCreateHandler(w http.ResponseWriter, r *http.Request) {
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

	var req apiCreateResourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}

	name := strings.TrimSpace(req.Name)
	rtype := strings.TrimSpace(req.Type)
	region := strings.TrimSpace(req.Region)

	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	if rtype == "" {
		rtype = "vm"
	}
	if region == "" {
		region = "us-east"
	}

	res := Resource{
		ID: newID(), Name: name, Type: rtype,
		Status: "online", Region: region,
		CPU: "0%", Memory: "0%",
		Created: "2026-04-17T00:00:00Z", Updated: "2026-04-17T00:00:00Z",
	}
	resourceStore.add(slug, res)

	apiCreated(w, apiResource{
		ID: res.ID, Name: res.Name, Type: res.Type,
		Status: res.Status, Region: res.Region,
		CPU: res.CPU, Memory: res.Memory,
		Created: res.Created, Updated: res.Updated,
	})
}

// ── GET /api/tenant/{slug}/resources/{id} ───────────────────

func apiTenantResourcesGetHandler(w http.ResponseWriter, r *http.Request) {
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
	res, ok := resourceStore.getByID(slug, id)
	if !ok {
		apiError(w, http.StatusNotFound, "Resource not found.")
		return
	}

	apiOK(w, apiResource{
		ID: res.ID, Name: res.Name, Type: res.Type,
		Status: res.Status, Region: res.Region,
		CPU: res.CPU, Memory: res.Memory,
		Created: res.Created, Updated: res.Updated,
	})
}

// ── DELETE /api/tenant/{slug}/resources/{id} ────────────────

func apiTenantResourcesDeleteHandler(w http.ResponseWriter, r *http.Request) {
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
	if _, ok := resourceStore.getByID(slug, id); !ok {
		apiError(w, http.StatusNotFound, "Resource not found.")
		return
	}

	resourceStore.delete(slug, id)
	apiOK(w, map[string]any{
		"ok":      true,
		"message": "Resource deleted.",
	})
}
