package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace AI API — Data: Data Catalog
// ai/data/catalog
// ════════════════════════════════════════════════════════════════

type CatalogAsset struct {
	ID      string
	Name    string
	Source  string
	Owner   string
	Type    string
	PII     string
	Tags    string
	Updated string
	Columns int
	Region  string
}

type catalogAssetStoreType struct {
	mu     sync.RWMutex
	assets map[string][]CatalogAsset
}

var catalogAssetStore = &catalogAssetStoreType{
	assets: map[string][]CatalogAsset{},
}

func (s *catalogAssetStoreType) list(slug string) []CatalogAsset {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.assets[slug]
	out := make([]CatalogAsset, len(src))
	copy(out, src)
	return out
}

func (s *catalogAssetStoreType) getByID(slug, id string) (CatalogAsset, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, a := range s.assets[slug] {
		if a.ID == id {
			return a, true
		}
	}
	return CatalogAsset{}, false
}

func (s *catalogAssetStoreType) add(slug string, a CatalogAsset) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.assets[slug] = append(s.assets[slug], a)
}

func (s *catalogAssetStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.assets[slug]
	for i, a := range list {
		if a.ID == id {
			s.assets[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiCatalogAsset struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Source  string `json:"source"`
	Owner   string `json:"owner"`
	Type    string `json:"type"`
	PII     string `json:"pii"`
	Tags    string `json:"tags"`
	Updated string `json:"updated"`
	Columns int    `json:"columns"`
	Region  string `json:"region"`
}

type apiCatalogAssetList struct {
	Assets []apiCatalogAsset `json:"assets"`
	Total  int               `json:"total"`
}

func toAPICatalogAsset(a CatalogAsset) apiCatalogAsset {
	return apiCatalogAsset{
		ID: a.ID, Name: a.Name, Source: a.Source,
		Owner: a.Owner, Type: a.Type, PII: a.PII,
		Tags: a.Tags, Updated: a.Updated,
		Columns: a.Columns, Region: a.Region,
	}
}

func filterCatalogAssets(all []CatalogAsset, search, pii string) []CatalogAsset {
	var out []CatalogAsset
	q := strings.ToLower(search)
	for _, a := range all {
		if pii != "" && a.PII != pii {
			continue
		}
		if q != "" {
			hay := strings.ToLower(a.Name + " " + a.Source + " " + a.Owner + " " + a.Type + " " + a.Tags + " " + a.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, a)
	}
	return out
}

func apiWSAIDataCatalogListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	assets := filterCatalogAssets(catalogAssetStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("pii"))
	out := make([]apiCatalogAsset, len(assets))
	for i, a := range assets {
		out[i] = toAPICatalogAsset(a)
	}
	apiOK(w, apiCatalogAssetList{Assets: out, Total: len(out)})
}

type apiCreateCatalogAssetRequest struct {
	Name   string `json:"name"`
	Source string `json:"source"`
	Owner  string `json:"owner"`
	Type   string `json:"type"`
	PII    string `json:"pii"`
	Tags   string `json:"tags"`
	Region string `json:"region"`
}

func apiWSAIDataCatalogCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateCatalogAssetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	source := strings.TrimSpace(req.Source)
	if source == "" {
		apiError(w, http.StatusBadRequest, "Source is required.")
		return
	}
	owner := strings.TrimSpace(req.Owner)
	if owner == "" {
		owner = "unassigned"
	}
	assetType := strings.TrimSpace(req.Type)
	if assetType == "" {
		assetType = "table"
	}
	pii := strings.TrimSpace(req.PII)
	if pii == "" {
		pii = "unknown"
	}
	tags := strings.TrimSpace(req.Tags)
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "eu-central-1"
	}
	asset := CatalogAsset{
		ID: newID(), Name: name, Source: source,
		Owner: owner, Type: assetType, PII: pii,
		Tags: tags, Updated: "just now",
		Columns: 0, Region: region,
	}
	catalogAssetStore.add(slug, asset)
	apiCreated(w, toAPICatalogAsset(asset))
}

func apiWSAIDataCatalogGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	asset, found := catalogAssetStore.getByID(slug, r.PathValue("aid"))
	if !found {
		apiError(w, http.StatusNotFound, "Catalog asset not found.")
		return
	}
	apiOK(w, toAPICatalogAsset(asset))
}

func apiWSAIDataCatalogDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := catalogAssetStore.getByID(slug, r.PathValue("aid")); !found {
		apiError(w, http.StatusNotFound, "Catalog asset not found.")
		return
	}
	catalogAssetStore.delete(slug, r.PathValue("aid"))
	apiOK(w, map[string]any{"ok": true, "message": "Catalog asset deleted."})
}
