package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Security API — Data Protection: Key Management
// security/data/kms
// ════════════════════════════════════════════════════════════════

type KMSKey struct {
	ID        string
	Alias     string
	Algorithm string
	Rotation  string
	Status    string
	Usage24h  string
	Created   string
	Region    string
}

type kmsKeyStoreType struct {
	mu   sync.RWMutex
	keys map[string][]KMSKey
}

var kmsKeyStore = &kmsKeyStoreType{
	keys: map[string][]KMSKey{},
}

func (s *kmsKeyStoreType) list(slug string) []KMSKey {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.keys[slug]
	out := make([]KMSKey, len(src))
	copy(out, src)
	return out
}

func (s *kmsKeyStoreType) getByID(slug, id string) (KMSKey, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, k := range s.keys[slug] {
		if k.ID == id {
			return k, true
		}
	}
	return KMSKey{}, false
}

func (s *kmsKeyStoreType) add(slug string, k KMSKey) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.keys[slug] = append(s.keys[slug], k)
}

func (s *kmsKeyStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.keys[slug]
	for i, k := range list {
		if k.ID == id {
			s.keys[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiKMSKey struct {
	ID        string `json:"id"`
	Alias     string `json:"alias"`
	Algorithm string `json:"algorithm"`
	Rotation  string `json:"rotation"`
	Status    string `json:"status"`
	Usage24h  string `json:"usage_24h"`
	Created   string `json:"created"`
	Region    string `json:"region"`
}

type apiKMSKeyList struct {
	Keys  []apiKMSKey `json:"keys"`
	Total int         `json:"total"`
}

func toAPIKMSKey(k KMSKey) apiKMSKey {
	return apiKMSKey{
		ID: k.ID, Alias: k.Alias, Algorithm: k.Algorithm,
		Rotation: k.Rotation, Status: k.Status,
		Usage24h: k.Usage24h, Created: k.Created,
		Region: k.Region,
	}
}

func filterKMSKeys(all []KMSKey, search, status string) []KMSKey {
	var out []KMSKey
	q := strings.ToLower(search)
	for _, k := range all {
		if status != "" && k.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(k.Alias + " " + k.Algorithm + " " + k.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, k)
	}
	return out
}

func apiWSSecDataKMSListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	keys := filterKMSKeys(kmsKeyStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiKMSKey, len(keys))
	for i, k := range keys {
		out[i] = toAPIKMSKey(k)
	}
	apiOK(w, apiKMSKeyList{Keys: out, Total: len(out)})
}

type apiCreateKMSKeyRequest struct {
	Alias     string `json:"alias"`
	Algorithm string `json:"algorithm"`
	Rotation  string `json:"rotation"`
	Region    string `json:"region"`
}

func apiWSSecDataKMSCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateKMSKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	alias := strings.TrimSpace(req.Alias)
	if alias == "" {
		apiError(w, http.StatusBadRequest, "Alias is required.")
		return
	}
	algorithm := strings.TrimSpace(req.Algorithm)
	if algorithm == "" {
		algorithm = "AES-256 symmetric"
	}
	rotation := strings.TrimSpace(req.Rotation)
	if rotation == "" {
		rotation = "Auto · 90d"
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "eu-central-1"
	}
	key := KMSKey{
		ID: newID(), Alias: alias, Algorithm: algorithm,
		Rotation: rotation, Status: "active",
		Usage24h: "0 ops", Created: "2026-04-17",
		Region: region,
	}
	kmsKeyStore.add(slug, key)
	apiCreated(w, toAPIKMSKey(key))
}

func apiWSSecDataKMSGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	key, found := kmsKeyStore.getByID(slug, r.PathValue("kid"))
	if !found {
		apiError(w, http.StatusNotFound, "KMS key not found.")
		return
	}
	apiOK(w, toAPIKMSKey(key))
}

func apiWSSecDataKMSDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := kmsKeyStore.getByID(slug, r.PathValue("kid")); !found {
		apiError(w, http.StatusNotFound, "KMS key not found.")
		return
	}
	kmsKeyStore.delete(slug, r.PathValue("kid"))
	apiOK(w, map[string]any{"ok": true, "message": "KMS key deleted."})
}
