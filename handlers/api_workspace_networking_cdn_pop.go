package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Networking API — CDN: PoP Management
// networking/cdn/pop-management
// ════════════════════════════════════════════════════════════════

type CDNPoP struct {
	ID           string
	Name         string
	Location     string
	Region       string
	Provider     string
	CacheHitRate float64
	BandwidthTB  float64
	RequestsM    float64
	AvgLatencyMs int
	Capacity     string
	TLSVersion   string
	HTTP2        bool
	HTTP3        bool
	Status       string
	Updated      string
}

type cdnPopStoreType struct {
	mu    sync.RWMutex
	items map[string][]CDNPoP
}

var cdnPopStore = &cdnPopStoreType{
	items: map[string][]CDNPoP{},
}

func (s *cdnPopStoreType) list(slug string) []CDNPoP {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.items[slug]
	out := make([]CDNPoP, len(src))
	copy(out, src)
	return out
}

func (s *cdnPopStoreType) getByID(slug, id string) (CDNPoP, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, p := range s.items[slug] {
		if p.ID == id {
			return p, true
		}
	}
	return CDNPoP{}, false
}

func (s *cdnPopStoreType) add(slug string, p CDNPoP) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[slug] = append(s.items[slug], p)
}

func (s *cdnPopStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.items[slug]
	for i, p := range list {
		if p.ID == id {
			s.items[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiCDNPoP struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Location     string  `json:"location"`
	Region       string  `json:"region"`
	Provider     string  `json:"provider"`
	CacheHitRate float64 `json:"cache_hit_rate"`
	BandwidthTB  float64 `json:"bandwidth_tb"`
	RequestsM    float64 `json:"requests_millions"`
	AvgLatencyMs int     `json:"avg_latency_ms"`
	Capacity     string  `json:"capacity"`
	TLSVersion   string  `json:"tls_version"`
	HTTP2        bool    `json:"http2"`
	HTTP3        bool    `json:"http3"`
	Status       string  `json:"status"`
	Updated      string  `json:"updated"`
}

type apiCDNPoPList struct {
	PoPs  []apiCDNPoP `json:"pops"`
	Total int         `json:"total"`
}

func toAPICDNPoP(p CDNPoP) apiCDNPoP {
	return apiCDNPoP{
		ID: p.ID, Name: p.Name, Location: p.Location,
		Region: p.Region, Provider: p.Provider,
		CacheHitRate: p.CacheHitRate, BandwidthTB: p.BandwidthTB,
		RequestsM: p.RequestsM, AvgLatencyMs: p.AvgLatencyMs,
		Capacity: p.Capacity, TLSVersion: p.TLSVersion,
		HTTP2: p.HTTP2, HTTP3: p.HTTP3,
		Status: p.Status, Updated: p.Updated,
	}
}

func filterCDNPoPs(all []CDNPoP, search, status string) []CDNPoP {
	var out []CDNPoP
	q := strings.ToLower(search)
	for _, p := range all {
		if status != "" && p.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(p.Name + " " + p.Location + " " + p.Region + " " + p.Provider)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, p)
	}
	return out
}

func apiWSNetCDNPoPListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	items := filterCDNPoPs(cdnPopStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiCDNPoP, len(items))
	for i, p := range items {
		out[i] = toAPICDNPoP(p)
	}
	apiOK(w, apiCDNPoPList{PoPs: out, Total: len(out)})
}

type apiCreateCDNPoPRequest struct {
	Name     string `json:"name"`
	Location string `json:"location"`
	Region   string `json:"region"`
	Provider string `json:"provider"`
	Capacity string `json:"capacity"`
}

func apiWSNetCDNPoPCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateCDNPoPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	location := strings.TrimSpace(req.Location)
	if location == "" {
		apiError(w, http.StatusBadRequest, "Location is required.")
		return
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "us-east-1"
	}
	provider := strings.TrimSpace(req.Provider)
	if provider == "" {
		provider = "Standard IX"
	}
	capacity := strings.TrimSpace(req.Capacity)
	if capacity == "" {
		capacity = "10 Gbps"
	}
	pop := CDNPoP{
		ID: newID(), Name: name, Location: location,
		Region: region, Provider: provider,
		CacheHitRate: 0, BandwidthTB: 0, RequestsM: 0,
		AvgLatencyMs: 0, Capacity: capacity,
		TLSVersion: "1.3", HTTP2: true, HTTP3: false,
		Status: "provisioning", Updated: "2026-04-17",
	}
	cdnPopStore.add(slug, pop)
	apiCreated(w, toAPICDNPoP(pop))
}

func apiWSNetCDNPoPGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	pop, found := cdnPopStore.getByID(slug, r.PathValue("pid"))
	if !found {
		apiError(w, http.StatusNotFound, "CDN PoP not found.")
		return
	}
	apiOK(w, toAPICDNPoP(pop))
}

func apiWSNetCDNPoPDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := cdnPopStore.getByID(slug, r.PathValue("pid")); !found {
		apiError(w, http.StatusNotFound, "CDN PoP not found.")
		return
	}
	cdnPopStore.delete(slug, r.PathValue("pid"))
	apiOK(w, map[string]any{"ok": true, "message": "CDN PoP deleted."})
}
