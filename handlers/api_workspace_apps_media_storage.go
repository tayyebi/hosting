package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Application Services API — Media: Media Storage
// apps/media/storage
// ════════════════════════════════════════════════════════════════

type MediaAsset struct {
	ID        string
	Name      string
	Type      string
	Size      string
	Format    string
	CDN       string
	Downloads string
	Bucket    string
	Status    string
	Updated   string
}

type mediaAssetStoreType struct {
	mu     sync.RWMutex
	assets map[string][]MediaAsset
}

var mediaAssetStore = &mediaAssetStoreType{
	assets: map[string][]MediaAsset{
		"buyvm": {
			{ID: newID(), Name: "hero-banner-2026.mp4", Type: "Video", Size: "245 MB", Format: "H.264 / MP4", CDN: "Enabled", Downloads: "12,400", Bucket: "media-production", Status: "active", Updated: "2026-04-17"},
			{ID: newID(), Name: "product-shots.zip", Type: "Image archive", Size: "1.2 GB", Format: "PNG / ZIP", CDN: "Enabled", Downloads: "3,200", Bucket: "media-production", Status: "active", Updated: "2026-04-16"},
			{ID: newID(), Name: "podcast-ep42.mp3", Type: "Audio", Size: "85 MB", Format: "MP3 320kbps", CDN: "Enabled", Downloads: "8,750", Bucket: "media-production", Status: "active", Updated: "2026-04-15"},
			{ID: newID(), Name: "webinar-apr-2026.mp4", Type: "Video", Size: "1.8 GB", Format: "H.265 / MP4", CDN: "Disabled", Downloads: "560", Bucket: "media-archive", Status: "archived", Updated: "2026-04-14"},
			{ID: newID(), Name: "brand-guidelines.pdf", Type: "Document", Size: "12 MB", Format: "PDF", CDN: "Enabled", Downloads: "1,100", Bucket: "media-production", Status: "active", Updated: "2026-04-10"},
			{ID: newID(), Name: "old-promo-2024.mov", Type: "Video", Size: "4.5 GB", Format: "ProRes / MOV", CDN: "Disabled", Downloads: "23", Bucket: "media-archive", Status: "archived", Updated: "2024-12-01"},
		},
		"hosting": {
			{ID: newID(), Name: "platform-tour.mp4", Type: "Video", Size: "320 MB", Format: "H.264 / MP4", CDN: "Enabled", Downloads: "6,800", Bucket: "hosting-media", Status: "active", Updated: "2026-04-17"},
			{ID: newID(), Name: "api-docs-diagrams.svg", Type: "Image", Size: "2.4 MB", Format: "SVG", CDN: "Enabled", Downloads: "15,200", Bucket: "hosting-media", Status: "active", Updated: "2026-04-16"},
			{ID: newID(), Name: "customer-stories.mp4", Type: "Video", Size: "890 MB", Format: "H.264 / MP4", CDN: "Enabled", Downloads: "2,100", Bucket: "hosting-media", Status: "active", Updated: "2026-04-13"},
			{ID: newID(), Name: "conference-recordings-2025.tar", Type: "Video archive", Size: "12.5 GB", Format: "TAR (mixed)", CDN: "Disabled", Downloads: "45", Bucket: "hosting-archive", Status: "archived", Updated: "2025-11-30"},
		},
	},
}

func (s *mediaAssetStoreType) list(slug string) []MediaAsset {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.assets[slug]
	out := make([]MediaAsset, len(src))
	copy(out, src)
	return out
}

func (s *mediaAssetStoreType) getByID(slug, id string) (MediaAsset, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, a := range s.assets[slug] {
		if a.ID == id {
			return a, true
		}
	}
	return MediaAsset{}, false
}

func (s *mediaAssetStoreType) add(slug string, a MediaAsset) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.assets[slug] = append(s.assets[slug], a)
}

func (s *mediaAssetStoreType) delete(slug, id string) {
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

type apiMediaAsset struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Size      string `json:"size"`
	Format    string `json:"format"`
	CDN       string `json:"cdn"`
	Downloads string `json:"downloads"`
	Bucket    string `json:"bucket"`
	Status    string `json:"status"`
	Updated   string `json:"updated"`
}

type apiMediaAssetList struct {
	Assets []apiMediaAsset `json:"assets"`
	Total  int             `json:"total"`
}

func toAPIMediaAsset(a MediaAsset) apiMediaAsset {
	return apiMediaAsset{
		ID: a.ID, Name: a.Name, Type: a.Type,
		Size: a.Size, Format: a.Format, CDN: a.CDN,
		Downloads: a.Downloads, Bucket: a.Bucket,
		Status: a.Status, Updated: a.Updated,
	}
}

func filterMediaAssets(all []MediaAsset, search, status string) []MediaAsset {
	var out []MediaAsset
	q := strings.ToLower(search)
	for _, a := range all {
		if status != "" && a.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(a.Name + " " + a.Type + " " + a.Format + " " + a.Bucket)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, a)
	}
	return out
}

func apiWSAppsMediaStorageListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	assets := filterMediaAssets(mediaAssetStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiMediaAsset, len(assets))
	for i, a := range assets {
		out[i] = toAPIMediaAsset(a)
	}
	apiOK(w, apiMediaAssetList{Assets: out, Total: len(out)})
}

type apiCreateMediaAssetRequest struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Size   string `json:"size"`
	Format string `json:"format"`
	Bucket string `json:"bucket"`
}

func apiWSAppsMediaStorageCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateMediaAssetRequest
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
		typ = "Video"
	}
	size := strings.TrimSpace(req.Size)
	if size == "" {
		size = "0 B"
	}
	format := strings.TrimSpace(req.Format)
	if format == "" {
		format = "Unknown"
	}
	bucket := strings.TrimSpace(req.Bucket)
	if bucket == "" {
		bucket = "media-default"
	}
	asset := MediaAsset{
		ID: newID(), Name: name, Type: typ,
		Size: size, Format: format, CDN: "Disabled",
		Downloads: "0", Bucket: bucket,
		Status: "active", Updated: "2026-04-17",
	}
	mediaAssetStore.add(slug, asset)
	apiCreated(w, toAPIMediaAsset(asset))
}

func apiWSAppsMediaStorageGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	asset, found := mediaAssetStore.getByID(slug, r.PathValue("mid"))
	if !found {
		apiError(w, http.StatusNotFound, "Media asset not found.")
		return
	}
	apiOK(w, toAPIMediaAsset(asset))
}

func apiWSAppsMediaStorageDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := mediaAssetStore.getByID(slug, r.PathValue("mid")); !found {
		apiError(w, http.StatusNotFound, "Media asset not found.")
		return
	}
	mediaAssetStore.delete(slug, r.PathValue("mid"))
	apiOK(w, map[string]any{"ok": true, "message": "Media asset deleted."})
}
