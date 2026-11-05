package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Security API — Data Protection: Encryption at Rest
// security/data/enc-rest
// ════════════════════════════════════════════════════════════════

type EncAtRestResource struct {
	ID        string
	Name      string
	Type      string
	Algorithm string
	Key       string
	Status    string
	Created   string
	Region    string
}

type encAtRestStoreType struct {
	mu        sync.RWMutex
	resources map[string][]EncAtRestResource
}

var encAtRestStore = &encAtRestStoreType{
	resources: map[string][]EncAtRestResource{
		"buyvm": {
			{ID: newID(), Name: "vol-data-db-primary", Type: "EBS volume", Algorithm: "AES-256-XTS", Key: "alias/db-master", Status: "encrypted", Created: "2025-08-12", Region: "eu-central-1"},
			{ID: newID(), Name: "acme-user-uploads", Type: "S3 bucket", Algorithm: "AES-256 (SSE-KMS)", Key: "alias/uploads-key", Status: "encrypted", Created: "2025-09-01", Region: "eu-central-1"},
			{ID: newID(), Name: "prod-main", Type: "RDS Postgres", Algorithm: "AES-256", Key: "alias/rds-default", Status: "encrypted", Created: "2025-10-15", Region: "eu-central-1"},
		},
		"hosting": {
			{ID: newID(), Name: "vol-app-data", Type: "EBS volume", Algorithm: "AES-256-XTS", Key: "alias/app-master", Status: "encrypted", Created: "2025-10-01", Region: "ap-south-1"},
			{ID: newID(), Name: "hosting-assets", Type: "S3 bucket", Algorithm: "AES-256 (SSE-KMS)", Key: "alias/assets-key", Status: "encrypted", Created: "2025-10-15", Region: "ap-south-1"},
			{ID: newID(), Name: "analytics-db", Type: "RDS MySQL", Algorithm: "AES-256", Key: "alias/rds-analytics", Status: "encrypted", Created: "2025-11-20", Region: "ap-south-1"},
		},
	},
}

func (s *encAtRestStoreType) list(slug string) []EncAtRestResource {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.resources[slug]
	out := make([]EncAtRestResource, len(src))
	copy(out, src)
	return out
}

func (s *encAtRestStoreType) getByID(slug, id string) (EncAtRestResource, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, r := range s.resources[slug] {
		if r.ID == id {
			return r, true
		}
	}
	return EncAtRestResource{}, false
}

func (s *encAtRestStoreType) add(slug string, r EncAtRestResource) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resources[slug] = append(s.resources[slug], r)
}

func (s *encAtRestStoreType) delete(slug, id string) {
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

type apiEncAtRestResource struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Algorithm string `json:"algorithm"`
	Key       string `json:"key"`
	Status    string `json:"status"`
	Created   string `json:"created"`
	Region    string `json:"region"`
}

type apiEncAtRestResourceList struct {
	Resources []apiEncAtRestResource `json:"resources"`
	Total     int                    `json:"total"`
}

func toAPIEncAtRestResource(r EncAtRestResource) apiEncAtRestResource {
	return apiEncAtRestResource{
		ID: r.ID, Name: r.Name, Type: r.Type,
		Algorithm: r.Algorithm, Key: r.Key,
		Status: r.Status, Created: r.Created,
		Region: r.Region,
	}
}

func filterEncAtRestResources(all []EncAtRestResource, search, status string) []EncAtRestResource {
	var out []EncAtRestResource
	q := strings.ToLower(search)
	for _, r := range all {
		if status != "" && r.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(r.Name + " " + r.Type + " " + r.Algorithm + " " + r.Key + " " + r.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, r)
	}
	return out
}

func apiWSSecDataEncRestListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	resources := filterEncAtRestResources(encAtRestStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiEncAtRestResource, len(resources))
	for i, res := range resources {
		out[i] = toAPIEncAtRestResource(res)
	}
	apiOK(w, apiEncAtRestResourceList{Resources: out, Total: len(out)})
}

type apiCreateEncAtRestResourceRequest struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Algorithm string `json:"algorithm"`
	Key       string `json:"key"`
	Region    string `json:"region"`
}

func apiWSSecDataEncRestCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateEncAtRestResourceRequest
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
		apiError(w, http.StatusBadRequest, "Type is required.")
		return
	}
	algorithm := strings.TrimSpace(req.Algorithm)
	if algorithm == "" {
		algorithm = "AES-256"
	}
	key := strings.TrimSpace(req.Key)
	if key == "" {
		apiError(w, http.StatusBadRequest, "Key is required.")
		return
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "eu-central-1"
	}
	resource := EncAtRestResource{
		ID: newID(), Name: name, Type: typ,
		Algorithm: algorithm, Key: key,
		Status: "encrypted", Created: "2026-04-17",
		Region: region,
	}
	encAtRestStore.add(slug, resource)
	apiCreated(w, toAPIEncAtRestResource(resource))
}

func apiWSSecDataEncRestGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	resource, found := encAtRestStore.getByID(slug, r.PathValue("rid"))
	if !found {
		apiError(w, http.StatusNotFound, "Encrypted resource not found.")
		return
	}
	apiOK(w, toAPIEncAtRestResource(resource))
}

func apiWSSecDataEncRestDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := encAtRestStore.getByID(slug, r.PathValue("rid")); !found {
		apiError(w, http.StatusNotFound, "Encrypted resource not found.")
		return
	}
	encAtRestStore.delete(slug, r.PathValue("rid"))
	apiOK(w, map[string]any{"ok": true, "message": "Encrypted resource removed."})
}
