package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Application Services API — Media: Live Streaming
// apps/media/livestream
// ════════════════════════════════════════════════════════════════

type LiveStream struct {
	ID         string
	Name       string
	Protocol   string
	IngestURL  string
	Viewers    int
	Bitrate    string
	Resolution string
	Uptime     string
	Status     string
	Updated    string
}

type liveStreamStoreType struct {
	mu      sync.RWMutex
	streams map[string][]LiveStream
}

var liveStreamStore = &liveStreamStoreType{
	streams: map[string][]LiveStream{},
}

func (s *liveStreamStoreType) list(slug string) []LiveStream {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.streams[slug]
	out := make([]LiveStream, len(src))
	copy(out, src)
	return out
}

func (s *liveStreamStoreType) getByID(slug, id string) (LiveStream, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, ls := range s.streams[slug] {
		if ls.ID == id {
			return ls, true
		}
	}
	return LiveStream{}, false
}

func (s *liveStreamStoreType) add(slug string, ls LiveStream) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.streams[slug] = append(s.streams[slug], ls)
}

func (s *liveStreamStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.streams[slug]
	for i, ls := range list {
		if ls.ID == id {
			s.streams[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiLiveStream struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Protocol   string `json:"protocol"`
	IngestURL  string `json:"ingest_url"`
	Viewers    int    `json:"viewers"`
	Bitrate    string `json:"bitrate"`
	Resolution string `json:"resolution"`
	Uptime     string `json:"uptime"`
	Status     string `json:"status"`
	Updated    string `json:"updated"`
}

type apiLiveStreamList struct {
	Streams []apiLiveStream `json:"streams"`
	Total   int             `json:"total"`
}

func toAPILiveStream(ls LiveStream) apiLiveStream {
	return apiLiveStream{
		ID: ls.ID, Name: ls.Name, Protocol: ls.Protocol,
		IngestURL: ls.IngestURL, Viewers: ls.Viewers, Bitrate: ls.Bitrate,
		Resolution: ls.Resolution, Uptime: ls.Uptime,
		Status: ls.Status, Updated: ls.Updated,
	}
}

func filterLiveStreams(all []LiveStream, search, status string) []LiveStream {
	var out []LiveStream
	q := strings.ToLower(search)
	for _, ls := range all {
		if status != "" && ls.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(ls.Name + " " + ls.Protocol)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, ls)
	}
	return out
}

func apiWSAppsMediaLivestreamListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	streams := filterLiveStreams(liveStreamStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiLiveStream, len(streams))
	for i, ls := range streams {
		out[i] = toAPILiveStream(ls)
	}
	apiOK(w, apiLiveStreamList{Streams: out, Total: len(out)})
}

type apiCreateLiveStreamRequest struct {
	Name       string `json:"name"`
	Protocol   string `json:"protocol"`
	Resolution string `json:"resolution"`
	Bitrate    string `json:"bitrate"`
}

func apiWSAppsMediaLivestreamCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateLiveStreamRequest
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
		protocol = "RTMP"
	}
	resolution := strings.TrimSpace(req.Resolution)
	if resolution == "" {
		resolution = "1920x1080"
	}
	bitrate := strings.TrimSpace(req.Bitrate)
	if bitrate == "" {
		bitrate = "6 Mbps"
	}
	ls := LiveStream{
		ID: newID(), Name: name, Protocol: protocol,
		IngestURL: "rtmp://ingest.example.net/live/" + strings.ReplaceAll(strings.ToLower(name), " ", "-"),
		Viewers:   0, Bitrate: bitrate, Resolution: resolution,
		Uptime: "—", Status: "idle", Updated: "2026-04-17",
	}
	liveStreamStore.add(slug, ls)
	apiCreated(w, toAPILiveStream(ls))
}

func apiWSAppsMediaLivestreamGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	ls, found := liveStreamStore.getByID(slug, r.PathValue("lid"))
	if !found {
		apiError(w, http.StatusNotFound, "Live stream not found.")
		return
	}
	apiOK(w, toAPILiveStream(ls))
}

func apiWSAppsMediaLivestreamDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := liveStreamStore.getByID(slug, r.PathValue("lid")); !found {
		apiError(w, http.StatusNotFound, "Live stream not found.")
		return
	}
	liveStreamStore.delete(slug, r.PathValue("lid"))
	apiOK(w, map[string]any{"ok": true, "message": "Live stream deleted."})
}
