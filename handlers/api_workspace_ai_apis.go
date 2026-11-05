package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace AI API — APIs: Vision
// ai/apis/vision
// ════════════════════════════════════════════════════════════════

type VisionAPI struct {
	ID         string
	Name       string
	Status     string
	Model      string
	Capability string
	Version    string
	Requests   string
	Latency    string
	Accuracy   string
	Region     string
	Endpoint   string
	APIKey     string
}

type visionAPIStoreType struct {
	mu   sync.RWMutex
	apis map[string][]VisionAPI
}

var visionAPIStore = &visionAPIStoreType{
	apis: map[string][]VisionAPI{},
}

func (s *visionAPIStoreType) list(slug string) []VisionAPI {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.apis[slug]
	out := make([]VisionAPI, len(src))
	copy(out, src)
	return out
}

func (s *visionAPIStoreType) getByID(slug, id string) (VisionAPI, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, a := range s.apis[slug] {
		if a.ID == id {
			return a, true
		}
	}
	return VisionAPI{}, false
}

func (s *visionAPIStoreType) add(slug string, a VisionAPI) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.apis[slug] = append(s.apis[slug], a)
}

func (s *visionAPIStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.apis[slug]
	for i, a := range list {
		if a.ID == id {
			s.apis[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiVisionAPI struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	Model      string `json:"model"`
	Capability string `json:"capability"`
	Version    string `json:"version"`
	Requests   string `json:"requests"`
	Latency    string `json:"latency"`
	Accuracy   string `json:"accuracy"`
	Region     string `json:"region"`
	Endpoint   string `json:"endpoint"`
	APIKey     string `json:"api_key"`
}

type apiVisionAPIList struct {
	APIs  []apiVisionAPI `json:"apis"`
	Total int            `json:"total"`
}

func toAPIVisionAPI(a VisionAPI) apiVisionAPI {
	return apiVisionAPI{
		ID: a.ID, Name: a.Name, Status: a.Status,
		Model: a.Model, Capability: a.Capability,
		Version: a.Version, Requests: a.Requests,
		Latency: a.Latency, Accuracy: a.Accuracy,
		Region: a.Region, Endpoint: a.Endpoint,
		APIKey: a.APIKey,
	}
}

func filterVisionAPIs(all []VisionAPI, search, status string) []VisionAPI {
	var out []VisionAPI
	q := strings.ToLower(search)
	for _, a := range all {
		if status != "" && a.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(a.Name + " " + a.Model + " " + a.Capability + " " + a.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, a)
	}
	return out
}

func apiWSAIAPIsVisionListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	apis := filterVisionAPIs(visionAPIStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiVisionAPI, len(apis))
	for i, a := range apis {
		out[i] = toAPIVisionAPI(a)
	}
	apiOK(w, apiVisionAPIList{APIs: out, Total: len(out)})
}

type apiCreateVisionAPIRequest struct {
	Name       string `json:"name"`
	Model      string `json:"model"`
	Capability string `json:"capability"`
	Region     string `json:"region"`
}

func apiWSAIAPIsVisionCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateVisionAPIRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	model := strings.TrimSpace(req.Model)
	if model == "" {
		model = "yolov8-m"
	}
	capability := strings.TrimSpace(req.Capability)
	if capability == "" {
		capability = "object-detection"
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "eu-central-1"
	}
	api := VisionAPI{
		ID: newID(), Name: name, Status: "provisioning",
		Model: model, Capability: capability,
		Version: "v1.0.0", Requests: "0/mo",
		Latency: "—", Accuracy: "—",
		Region:   region,
		Endpoint: "vision-" + strings.ReplaceAll(strings.ToLower(name), " ", "-") + ".api." + region + ".example.com",
		APIKey:   "vk_new_****" + newID()[:4],
	}
	visionAPIStore.add(slug, api)
	apiCreated(w, toAPIVisionAPI(api))
}

func apiWSAIAPIsVisionGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	api, found := visionAPIStore.getByID(slug, r.PathValue("aid"))
	if !found {
		apiError(w, http.StatusNotFound, "Vision API not found.")
		return
	}
	apiOK(w, toAPIVisionAPI(api))
}

func apiWSAIAPIsVisionDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := visionAPIStore.getByID(slug, r.PathValue("aid")); !found {
		apiError(w, http.StatusNotFound, "Vision API not found.")
		return
	}
	visionAPIStore.delete(slug, r.PathValue("aid"))
	apiOK(w, map[string]any{"ok": true, "message": "Vision API deleted."})
}
