package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Networking API — CDN: Static Asset Delivery
// networking/cdn/static
// ════════════════════════════════════════════════════════════════

type CDNStatic struct {
	ID          string
	Name        string
	Domain      string
	Origin      string
	Status      string
	Region      string
	CachePolicy string
	Bandwidth   string
	Requests    string
	HitRate     string
	SSL         bool
}

type cdnStaticStoreType struct {
	mu            sync.RWMutex
	distributions map[string][]CDNStatic
}

var cdnStaticStore = &cdnStaticStoreType{
	distributions: map[string][]CDNStatic{},
}

func (s *cdnStaticStoreType) list(slug string) []CDNStatic {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.distributions[slug]
	out := make([]CDNStatic, len(src))
	copy(out, src)
	return out
}

func (s *cdnStaticStoreType) getByID(slug, id string) (CDNStatic, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, d := range s.distributions[slug] {
		if d.ID == id {
			return d, true
		}
	}
	return CDNStatic{}, false
}

func (s *cdnStaticStoreType) add(slug string, d CDNStatic) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.distributions[slug] = append(s.distributions[slug], d)
}

func (s *cdnStaticStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.distributions[slug]
	for i, d := range list {
		if d.ID == id {
			s.distributions[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiCDNStatic struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Domain      string `json:"domain"`
	Origin      string `json:"origin"`
	Status      string `json:"status"`
	Region      string `json:"region"`
	CachePolicy string `json:"cache_policy"`
	Bandwidth   string `json:"bandwidth"`
	Requests    string `json:"requests"`
	HitRate     string `json:"hit_rate"`
	SSL         bool   `json:"ssl"`
}

type apiCDNStaticList struct {
	Distributions []apiCDNStatic `json:"distributions"`
	Total         int            `json:"total"`
}

func toAPICDNStatic(d CDNStatic) apiCDNStatic {
	return apiCDNStatic{
		ID: d.ID, Name: d.Name, Domain: d.Domain,
		Origin: d.Origin, Status: d.Status,
		Region: d.Region, CachePolicy: d.CachePolicy,
		Bandwidth: d.Bandwidth, Requests: d.Requests,
		HitRate: d.HitRate, SSL: d.SSL,
	}
}

func filterCDNStatic(all []CDNStatic, search, status string) []CDNStatic {
	var out []CDNStatic
	q := strings.ToLower(search)
	for _, d := range all {
		if status != "" && d.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(d.Name + " " + d.Domain + " " + d.Origin + " " + d.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, d)
	}
	return out
}

func apiWSNetCDNStaticListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	dists := filterCDNStatic(cdnStaticStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiCDNStatic, len(dists))
	for i, d := range dists {
		out[i] = toAPICDNStatic(d)
	}
	apiOK(w, apiCDNStaticList{Distributions: out, Total: len(out)})
}

type apiCreateCDNStaticRequest struct {
	Name        string `json:"name"`
	Domain      string `json:"domain"`
	Origin      string `json:"origin"`
	Region      string `json:"region"`
	CachePolicy string `json:"cache_policy"`
	SSL         bool   `json:"ssl"`
}

func apiWSNetCDNStaticCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateCDNStaticRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	domain := strings.TrimSpace(req.Domain)
	if domain == "" {
		apiError(w, http.StatusBadRequest, "Domain is required.")
		return
	}
	origin := strings.TrimSpace(req.Origin)
	if origin == "" {
		apiError(w, http.StatusBadRequest, "Origin is required.")
		return
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "global"
	}
	policy := strings.TrimSpace(req.CachePolicy)
	if policy == "" {
		policy = "standard"
	}
	dist := CDNStatic{
		ID: newID(), Name: name, Domain: domain,
		Origin: origin, Status: "deploying",
		Region: region, CachePolicy: policy,
		Bandwidth: "0 B/mo", Requests: "0/mo",
		HitRate: "—", SSL: req.SSL,
	}
	cdnStaticStore.add(slug, dist)
	apiCreated(w, toAPICDNStatic(dist))
}

func apiWSNetCDNStaticGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	dist, found := cdnStaticStore.getByID(slug, r.PathValue("did"))
	if !found {
		apiError(w, http.StatusNotFound, "CDN distribution not found.")
		return
	}
	apiOK(w, toAPICDNStatic(dist))
}

func apiWSNetCDNStaticDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := cdnStaticStore.getByID(slug, r.PathValue("did")); !found {
		apiError(w, http.StatusNotFound, "CDN distribution not found.")
		return
	}
	cdnStaticStore.delete(slug, r.PathValue("did"))
	apiOK(w, map[string]any{"ok": true, "message": "CDN distribution deleted."})
}

// ════════════════════════════════════════════════════════════════
// Workspace Networking API — CDN: Video Streaming
// networking/cdn/video
// ════════════════════════════════════════════════════════════════

type CDNVideo struct {
	ID          string
	Name        string
	Domain      string
	Origin      string
	Status      string
	Region      string
	Protocol    string
	Bandwidth   string
	Viewers     string
	Transcoding string
	DRM         bool
}

type cdnVideoStoreType struct {
	mu      sync.RWMutex
	streams map[string][]CDNVideo
}

var cdnVideoStore = &cdnVideoStoreType{
	streams: map[string][]CDNVideo{},
}

func (s *cdnVideoStoreType) list(slug string) []CDNVideo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.streams[slug]
	out := make([]CDNVideo, len(src))
	copy(out, src)
	return out
}

func (s *cdnVideoStoreType) getByID(slug, id string) (CDNVideo, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, v := range s.streams[slug] {
		if v.ID == id {
			return v, true
		}
	}
	return CDNVideo{}, false
}

func (s *cdnVideoStoreType) add(slug string, v CDNVideo) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.streams[slug] = append(s.streams[slug], v)
}

func (s *cdnVideoStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.streams[slug]
	for i, v := range list {
		if v.ID == id {
			s.streams[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiCDNVideo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Domain      string `json:"domain"`
	Origin      string `json:"origin"`
	Status      string `json:"status"`
	Region      string `json:"region"`
	Protocol    string `json:"protocol"`
	Bandwidth   string `json:"bandwidth"`
	Viewers     string `json:"viewers"`
	Transcoding string `json:"transcoding"`
	DRM         bool   `json:"drm"`
}

type apiCDNVideoList struct {
	Streams []apiCDNVideo `json:"streams"`
	Total   int           `json:"total"`
}

func toAPICDNVideo(v CDNVideo) apiCDNVideo {
	return apiCDNVideo{
		ID: v.ID, Name: v.Name, Domain: v.Domain,
		Origin: v.Origin, Status: v.Status,
		Region: v.Region, Protocol: v.Protocol,
		Bandwidth: v.Bandwidth, Viewers: v.Viewers,
		Transcoding: v.Transcoding, DRM: v.DRM,
	}
}

func filterCDNVideo(all []CDNVideo, search, status string) []CDNVideo {
	var out []CDNVideo
	q := strings.ToLower(search)
	for _, v := range all {
		if status != "" && v.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(v.Name + " " + v.Domain + " " + v.Protocol + " " + v.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, v)
	}
	return out
}

func apiWSNetCDNVideoListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	streams := filterCDNVideo(cdnVideoStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiCDNVideo, len(streams))
	for i, v := range streams {
		out[i] = toAPICDNVideo(v)
	}
	apiOK(w, apiCDNVideoList{Streams: out, Total: len(out)})
}

type apiCreateCDNVideoRequest struct {
	Name        string `json:"name"`
	Domain      string `json:"domain"`
	Origin      string `json:"origin"`
	Region      string `json:"region"`
	Protocol    string `json:"protocol"`
	Transcoding string `json:"transcoding"`
	DRM         bool   `json:"drm"`
}

func apiWSNetCDNVideoCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateCDNVideoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	domain := strings.TrimSpace(req.Domain)
	if domain == "" {
		apiError(w, http.StatusBadRequest, "Domain is required.")
		return
	}
	origin := strings.TrimSpace(req.Origin)
	if origin == "" {
		apiError(w, http.StatusBadRequest, "Origin is required.")
		return
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "global"
	}
	protocol := strings.TrimSpace(req.Protocol)
	if protocol == "" {
		protocol = "HLS"
	}
	transcoding := strings.TrimSpace(req.Transcoding)
	if transcoding == "" {
		transcoding = "adaptive"
	}
	stream := CDNVideo{
		ID: newID(), Name: name, Domain: domain,
		Origin: origin, Status: "provisioning",
		Region: region, Protocol: protocol,
		Bandwidth: "0 B/mo", Viewers: "0/mo",
		Transcoding: transcoding, DRM: req.DRM,
	}
	cdnVideoStore.add(slug, stream)
	apiCreated(w, toAPICDNVideo(stream))
}

func apiWSNetCDNVideoGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	stream, found := cdnVideoStore.getByID(slug, r.PathValue("vid"))
	if !found {
		apiError(w, http.StatusNotFound, "Video streaming endpoint not found.")
		return
	}
	apiOK(w, toAPICDNVideo(stream))
}

func apiWSNetCDNVideoDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := cdnVideoStore.getByID(slug, r.PathValue("vid")); !found {
		apiError(w, http.StatusNotFound, "Video streaming endpoint not found.")
		return
	}
	cdnVideoStore.delete(slug, r.PathValue("vid"))
	apiOK(w, map[string]any{"ok": true, "message": "Video streaming endpoint deleted."})
}

// ════════════════════════════════════════════════════════════════
// Workspace Networking API — CDN: Edge Caching
// networking/cdn/edge
// ════════════════════════════════════════════════════════════════

type CDNEdge struct {
	ID        string
	Name      string
	Region    string
	Status    string
	POPs      int
	Rules     int
	CacheSize string
	HitRate   string
	Bandwidth string
	Purges    string
	WAF       bool
}

type cdnEdgeStoreType struct {
	mu    sync.RWMutex
	edges map[string][]CDNEdge
}

var cdnEdgeStore = &cdnEdgeStoreType{
	edges: map[string][]CDNEdge{},
}

func (s *cdnEdgeStoreType) list(slug string) []CDNEdge {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.edges[slug]
	out := make([]CDNEdge, len(src))
	copy(out, src)
	return out
}

func (s *cdnEdgeStoreType) getByID(slug, id string) (CDNEdge, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, e := range s.edges[slug] {
		if e.ID == id {
			return e, true
		}
	}
	return CDNEdge{}, false
}

func (s *cdnEdgeStoreType) add(slug string, e CDNEdge) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.edges[slug] = append(s.edges[slug], e)
}

func (s *cdnEdgeStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.edges[slug]
	for i, e := range list {
		if e.ID == id {
			s.edges[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiCDNEdge struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Region    string `json:"region"`
	Status    string `json:"status"`
	POPs      int    `json:"pops"`
	Rules     int    `json:"rules"`
	CacheSize string `json:"cache_size"`
	HitRate   string `json:"hit_rate"`
	Bandwidth string `json:"bandwidth"`
	Purges    string `json:"purges"`
	WAF       bool   `json:"waf"`
}

type apiCDNEdgeList struct {
	Edges []apiCDNEdge `json:"edges"`
	Total int          `json:"total"`
}

func toAPICDNEdge(e CDNEdge) apiCDNEdge {
	return apiCDNEdge{
		ID: e.ID, Name: e.Name, Region: e.Region,
		Status: e.Status, POPs: e.POPs, Rules: e.Rules,
		CacheSize: e.CacheSize, HitRate: e.HitRate,
		Bandwidth: e.Bandwidth, Purges: e.Purges,
		WAF: e.WAF,
	}
}

func filterCDNEdge(all []CDNEdge, search, status string) []CDNEdge {
	var out []CDNEdge
	q := strings.ToLower(search)
	for _, e := range all {
		if status != "" && e.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(e.Name + " " + e.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, e)
	}
	return out
}

func apiWSNetCDNEdgeListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	edges := filterCDNEdge(cdnEdgeStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiCDNEdge, len(edges))
	for i, e := range edges {
		out[i] = toAPICDNEdge(e)
	}
	apiOK(w, apiCDNEdgeList{Edges: out, Total: len(out)})
}

type apiCreateCDNEdgeRequest struct {
	Name   string `json:"name"`
	Region string `json:"region"`
	WAF    bool   `json:"waf"`
}

func apiWSNetCDNEdgeCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateCDNEdgeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "global"
	}
	pops := 84
	switch region {
	case "eu":
		pops = 28
	case "us":
		pops = 32
	case "apac":
		pops = 18
	}
	edge := CDNEdge{
		ID: newID(), Name: name, Region: region,
		Status: "provisioning", POPs: pops, Rules: 0,
		CacheSize: "0 B", HitRate: "—",
		Bandwidth: "0 B/mo", Purges: "0/day",
		WAF: req.WAF,
	}
	cdnEdgeStore.add(slug, edge)
	apiCreated(w, toAPICDNEdge(edge))
}

func apiWSNetCDNEdgeGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	edge, found := cdnEdgeStore.getByID(slug, r.PathValue("eid"))
	if !found {
		apiError(w, http.StatusNotFound, "Edge cache config not found.")
		return
	}
	apiOK(w, toAPICDNEdge(edge))
}

func apiWSNetCDNEdgeDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := cdnEdgeStore.getByID(slug, r.PathValue("eid")); !found {
		apiError(w, http.StatusNotFound, "Edge cache config not found.")
		return
	}
	cdnEdgeStore.delete(slug, r.PathValue("eid"))
	apiOK(w, map[string]any{"ok": true, "message": "Edge cache config deleted."})
}
