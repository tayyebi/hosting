package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Security API — Data Protection: Encryption in Transit
// security/data/enc-transit
// ════════════════════════════════════════════════════════════════

type EncTransitEndpoint struct {
	ID       string
	Name     string
	Protocol string
	Cipher   string
	Grade    string
	Expires  string
	Created  string
	Region   string
}

type encTransitStoreType struct {
	mu        sync.RWMutex
	endpoints map[string][]EncTransitEndpoint
}

var encTransitStore = &encTransitStoreType{
	endpoints: map[string][]EncTransitEndpoint{
		"buyvm": {
			{ID: newID(), Name: "api.acme.io", Protocol: "TLS 1.3", Cipher: "TLS_AES_256_GCM_SHA384", Grade: "A+", Expires: "89d", Created: "2025-08-12", Region: "eu-central-1"},
			{ID: newID(), Name: "acme.io", Protocol: "TLS 1.3", Cipher: "TLS_AES_256_GCM_SHA384", Grade: "A+", Expires: "89d", Created: "2025-08-12", Region: "eu-central-1"},
			{ID: newID(), Name: "internal-grpc.svc", Protocol: "mTLS 1.3", Cipher: "TLS_AES_128_GCM_SHA256", Grade: "A+", Expires: "auto-rotates", Created: "2025-09-01", Region: "eu-central-1"},
			{ID: newID(), Name: "admin.acme.io", Protocol: "TLS 1.3", Cipher: "TLS_CHACHA20_POLY1305", Grade: "A+", Expires: "82d", Created: "2025-10-15", Region: "eu-central-1"},
		},
		"hosting": {
			{ID: newID(), Name: "app.hosting.io", Protocol: "TLS 1.3", Cipher: "TLS_AES_256_GCM_SHA384", Grade: "A+", Expires: "142d", Created: "2025-10-01", Region: "ap-south-1"},
			{ID: newID(), Name: "api.hosting.io", Protocol: "TLS 1.3", Cipher: "TLS_AES_256_GCM_SHA384", Grade: "A+", Expires: "142d", Created: "2025-10-01", Region: "ap-south-1"},
			{ID: newID(), Name: "mesh.hosting.internal", Protocol: "mTLS 1.3", Cipher: "TLS_AES_128_GCM_SHA256", Grade: "A+", Expires: "auto-rotates", Created: "2025-10-15", Region: "ap-south-1"},
		},
	},
}

func (s *encTransitStoreType) list(slug string) []EncTransitEndpoint {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.endpoints[slug]
	out := make([]EncTransitEndpoint, len(src))
	copy(out, src)
	return out
}

func (s *encTransitStoreType) getByID(slug, id string) (EncTransitEndpoint, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, e := range s.endpoints[slug] {
		if e.ID == id {
			return e, true
		}
	}
	return EncTransitEndpoint{}, false
}

func (s *encTransitStoreType) add(slug string, e EncTransitEndpoint) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.endpoints[slug] = append(s.endpoints[slug], e)
}

func (s *encTransitStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.endpoints[slug]
	for i, e := range list {
		if e.ID == id {
			s.endpoints[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiEncTransitEndpoint struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Protocol string `json:"protocol"`
	Cipher   string `json:"cipher"`
	Grade    string `json:"grade"`
	Expires  string `json:"expires"`
	Created  string `json:"created"`
	Region   string `json:"region"`
}

type apiEncTransitEndpointList struct {
	Endpoints []apiEncTransitEndpoint `json:"endpoints"`
	Total     int                     `json:"total"`
}

func toAPIEncTransitEndpoint(e EncTransitEndpoint) apiEncTransitEndpoint {
	return apiEncTransitEndpoint{
		ID: e.ID, Name: e.Name, Protocol: e.Protocol,
		Cipher: e.Cipher, Grade: e.Grade,
		Expires: e.Expires, Created: e.Created,
		Region: e.Region,
	}
}

func filterEncTransitEndpoints(all []EncTransitEndpoint, search, protocol string) []EncTransitEndpoint {
	var out []EncTransitEndpoint
	q := strings.ToLower(search)
	for _, e := range all {
		if protocol != "" && !strings.EqualFold(e.Protocol, protocol) {
			continue
		}
		if q != "" {
			hay := strings.ToLower(e.Name + " " + e.Protocol + " " + e.Cipher + " " + e.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, e)
	}
	return out
}

func apiWSSecDataEncTransitListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	endpoints := filterEncTransitEndpoints(encTransitStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("protocol"))
	out := make([]apiEncTransitEndpoint, len(endpoints))
	for i, e := range endpoints {
		out[i] = toAPIEncTransitEndpoint(e)
	}
	apiOK(w, apiEncTransitEndpointList{Endpoints: out, Total: len(out)})
}

type apiCreateEncTransitEndpointRequest struct {
	Name     string `json:"name"`
	Protocol string `json:"protocol"`
	Cipher   string `json:"cipher"`
	Region   string `json:"region"`
}

func apiWSSecDataEncTransitCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateEncTransitEndpointRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	protocol := strings.TrimSpace(req.Protocol)
	if protocol == "" {
		protocol = "TLS 1.3"
	}
	cipher := strings.TrimSpace(req.Cipher)
	if cipher == "" {
		cipher = "TLS_AES_256_GCM_SHA384"
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "eu-central-1"
	}
	endpoint := EncTransitEndpoint{
		ID: newID(), Name: name, Protocol: protocol,
		Cipher: cipher, Grade: "A+",
		Expires: "365d", Created: "2026-04-17",
		Region: region,
	}
	encTransitStore.add(slug, endpoint)
	apiCreated(w, toAPIEncTransitEndpoint(endpoint))
}

func apiWSSecDataEncTransitGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	endpoint, found := encTransitStore.getByID(slug, r.PathValue("eid"))
	if !found {
		apiError(w, http.StatusNotFound, "TLS endpoint not found.")
		return
	}
	apiOK(w, toAPIEncTransitEndpoint(endpoint))
}

func apiWSSecDataEncTransitDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := encTransitStore.getByID(slug, r.PathValue("eid")); !found {
		apiError(w, http.StatusNotFound, "TLS endpoint not found.")
		return
	}
	encTransitStore.delete(slug, r.PathValue("eid"))
	apiOK(w, map[string]any{"ok": true, "message": "TLS endpoint removed."})
}
