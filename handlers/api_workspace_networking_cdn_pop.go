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
	items: map[string][]CDNPoP{
		"buyvm": {
			{ID: newID(), Name: "nyc-edge-01", Location: "New York, US", Region: "us-east-1", Provider: "Tier-1 IX", CacheHitRate: 94.5, BandwidthTB: 12.8, RequestsM: 245.0, AvgLatencyMs: 8, Capacity: "100 Gbps", TLSVersion: "1.3", HTTP2: true, HTTP3: true, Status: "active", Updated: "2026-04-17"},
			{ID: newID(), Name: "ams-edge-01", Location: "Amsterdam, NL", Region: "eu-west-1", Provider: "AMS-IX", CacheHitRate: 92.1, BandwidthTB: 8.5, RequestsM: 180.0, AvgLatencyMs: 12, Capacity: "100 Gbps", TLSVersion: "1.3", HTTP2: true, HTTP3: true, Status: "active", Updated: "2026-04-17"},
			{ID: newID(), Name: "sin-edge-01", Location: "Singapore, SG", Region: "ap-southeast-1", Provider: "Equinix SG", CacheHitRate: 89.3, BandwidthTB: 4.2, RequestsM: 95.0, AvgLatencyMs: 15, Capacity: "40 Gbps", TLSVersion: "1.3", HTTP2: true, HTTP3: false, Status: "active", Updated: "2026-04-16"},
			{ID: newID(), Name: "lax-edge-01", Location: "Los Angeles, US", Region: "us-west-2", Provider: "CoreSite LA", CacheHitRate: 91.8, BandwidthTB: 6.1, RequestsM: 120.0, AvgLatencyMs: 10, Capacity: "40 Gbps", TLSVersion: "1.3", HTTP2: true, HTTP3: true, Status: "active", Updated: "2026-04-16"},
			{ID: newID(), Name: "fra-edge-01", Location: "Frankfurt, DE", Region: "eu-central-1", Provider: "DE-CIX", CacheHitRate: 93.7, BandwidthTB: 9.3, RequestsM: 210.0, AvgLatencyMs: 11, Capacity: "100 Gbps", TLSVersion: "1.3", HTTP2: true, HTTP3: true, Status: "active", Updated: "2026-04-17"},
			{ID: newID(), Name: "syd-edge-01", Location: "Sydney, AU", Region: "ap-southeast-2", Provider: "Equinix SY", CacheHitRate: 87.5, BandwidthTB: 2.1, RequestsM: 45.0, AvgLatencyMs: 22, Capacity: "10 Gbps", TLSVersion: "1.3", HTTP2: true, HTTP3: false, Status: "maintenance", Updated: "2026-04-15"},
		},
		"hosting": {
			{ID: newID(), Name: "lon-pop-01", Location: "London, UK", Region: "eu-west-2", Provider: "LINX", CacheHitRate: 95.2, BandwidthTB: 15.0, RequestsM: 320.0, AvgLatencyMs: 6, Capacity: "100 Gbps", TLSVersion: "1.3", HTTP2: true, HTTP3: true, Status: "active", Updated: "2026-04-17"},
			{ID: newID(), Name: "fra-pop-01", Location: "Frankfurt, DE", Region: "eu-central-1", Provider: "DE-CIX", CacheHitRate: 94.8, BandwidthTB: 18.5, RequestsM: 410.0, AvgLatencyMs: 7, Capacity: "200 Gbps", TLSVersion: "1.3", HTTP2: true, HTTP3: true, Status: "active", Updated: "2026-04-17"},
			{ID: newID(), Name: "iad-pop-01", Location: "Ashburn, US", Region: "us-east-1", Provider: "Equinix DC", CacheHitRate: 93.0, BandwidthTB: 22.0, RequestsM: 500.0, AvgLatencyMs: 9, Capacity: "200 Gbps", TLSVersion: "1.3", HTTP2: true, HTTP3: true, Status: "active", Updated: "2026-04-17"},
			{ID: newID(), Name: "nrt-pop-01", Location: "Tokyo, JP", Region: "ap-northeast-1", Provider: "JPIX", CacheHitRate: 90.5, BandwidthTB: 7.8, RequestsM: 165.0, AvgLatencyMs: 14, Capacity: "40 Gbps", TLSVersion: "1.3", HTTP2: true, HTTP3: true, Status: "active", Updated: "2026-04-16"},
			{ID: newID(), Name: "gru-pop-01", Location: "São Paulo, BR", Region: "sa-east-1", Provider: "IX.br", CacheHitRate: 86.2, BandwidthTB: 3.5, RequestsM: 78.0, AvgLatencyMs: 25, Capacity: "10 Gbps", TLSVersion: "1.3", HTTP2: true, HTTP3: false, Status: "active", Updated: "2026-04-15"},
		},
	},
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
