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
	ID        string
	Name      string
	Status    string
	Model     string
	Capability string
	Version   string
	Requests  string
	Latency   string
	Accuracy  string
	Region    string
	Endpoint  string
	APIKey    string
}

type visionAPIStoreType struct {
	mu   sync.RWMutex
	apis map[string][]VisionAPI
}

var visionAPIStore = &visionAPIStoreType{
	apis: map[string][]VisionAPI{
		"buyvm": {
			{ID: newID(), Name: "object-detection-v3", Status: "active", Model: "yolov8-x", Capability: "object-detection", Version: "v3.2.0", Requests: "2.4M/mo", Latency: "42ms p99", Accuracy: "94.2% mAP", Region: "eu-central-1", Endpoint: "vision-objdet.api.eu-central-1.example.com", APIKey: "vk_prod_****8a2f"},
			{ID: newID(), Name: "image-classification", Status: "active", Model: "efficientnet-v2-l", Capability: "classification", Version: "v2.0.0", Requests: "1.8M/mo", Latency: "18ms p99", Accuracy: "96.8% top-5", Region: "eu-central-1", Endpoint: "vision-classify.api.eu-central-1.example.com", APIKey: "vk_prod_****c4d1"},
			{ID: newID(), Name: "ocr-document", Status: "active", Model: "trocr-large", Capability: "ocr", Version: "v1.5.0", Requests: "480K/mo", Latency: "320ms p99", Accuracy: "99.1% CER", Region: "us-east-1", Endpoint: "vision-ocr.api.us-east-1.example.com", APIKey: "vk_prod_****e7b3"},
			{ID: newID(), Name: "face-analysis", Status: "active", Model: "arcface-r100", Capability: "face-detection", Version: "v2.1.0", Requests: "920K/mo", Latency: "56ms p99", Accuracy: "99.8% LFW", Region: "eu-central-1", Endpoint: "vision-face.api.eu-central-1.example.com", APIKey: "vk_prod_****1f9e"},
			{ID: newID(), Name: "segmentation-beta", Status: "beta", Model: "sam-2-large", Capability: "segmentation", Version: "v0.9.0", Requests: "12K/mo", Latency: "180ms p99", Accuracy: "91.4% IoU", Region: "eu-central-1", Endpoint: "vision-seg.api.eu-central-1.example.com", APIKey: "vk_beta_****4d2a"},
		},
		"hosting": {
			{ID: newID(), Name: "product-recognition", Status: "active", Model: "resnet-101-custom", Capability: "classification", Version: "v1.2.0", Requests: "640K/mo", Latency: "24ms p99", Accuracy: "97.4% top-5", Region: "ap-south-1", Endpoint: "vision-product.api.ap-south-1.example.com", APIKey: "vk_prod_****a8c2"},
			{ID: newID(), Name: "barcode-scanner", Status: "active", Model: "barcode-det-v2", Capability: "ocr", Version: "v2.0.0", Requests: "1.2M/mo", Latency: "12ms p99", Accuracy: "99.9%", Region: "ap-south-1", Endpoint: "vision-barcode.api.ap-south-1.example.com", APIKey: "vk_prod_****b5e7"},
			{ID: newID(), Name: "defect-detection", Status: "active", Model: "yolov8-custom", Capability: "object-detection", Version: "v1.0.0", Requests: "280K/mo", Latency: "38ms p99", Accuracy: "92.6% mAP", Region: "ap-south-1", Endpoint: "vision-defect.api.ap-south-1.example.com", APIKey: "vk_prod_****f3a1"},
		},
	},
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
		Region: region,
		Endpoint: "vision-" + strings.ReplaceAll(strings.ToLower(name), " ", "-") + ".api." + region + ".example.com",
		APIKey: "vk_new_****" + newID()[:4],
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
