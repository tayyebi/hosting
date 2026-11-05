package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace AI API — APIs: Speech
// ai/apis/speech
// ════════════════════════════════════════════════════════════════

type SpeechAPI struct {
	ID         string
	Name       string
	Status     string
	Model      string
	Capability string
	Version    string
	Requests   string
	Latency    string
	Languages  string
	SampleRate string
	Region     string
	Endpoint   string
	APIKey     string
}

type speechAPIStoreType struct {
	mu   sync.RWMutex
	apis map[string][]SpeechAPI
}

var speechAPIStore = &speechAPIStoreType{
	apis: map[string][]SpeechAPI{},
}

func (s *speechAPIStoreType) list(slug string) []SpeechAPI {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.apis[slug]
	out := make([]SpeechAPI, len(src))
	copy(out, src)
	return out
}

func (s *speechAPIStoreType) getByID(slug, id string) (SpeechAPI, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, a := range s.apis[slug] {
		if a.ID == id {
			return a, true
		}
	}
	return SpeechAPI{}, false
}

func (s *speechAPIStoreType) add(slug string, a SpeechAPI) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.apis[slug] = append(s.apis[slug], a)
}

func (s *speechAPIStoreType) delete(slug, id string) {
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

type apiSpeechAPI struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	Model      string `json:"model"`
	Capability string `json:"capability"`
	Version    string `json:"version"`
	Requests   string `json:"requests"`
	Latency    string `json:"latency"`
	Languages  string `json:"languages"`
	SampleRate string `json:"sample_rate"`
	Region     string `json:"region"`
	Endpoint   string `json:"endpoint"`
	APIKey     string `json:"api_key"`
}

type apiSpeechAPIList struct {
	APIs  []apiSpeechAPI `json:"apis"`
	Total int            `json:"total"`
}

func toAPISpeechAPI(a SpeechAPI) apiSpeechAPI {
	return apiSpeechAPI{
		ID: a.ID, Name: a.Name, Status: a.Status,
		Model: a.Model, Capability: a.Capability,
		Version: a.Version, Requests: a.Requests,
		Latency: a.Latency, Languages: a.Languages,
		SampleRate: a.SampleRate, Region: a.Region,
		Endpoint: a.Endpoint, APIKey: a.APIKey,
	}
}

func filterSpeechAPIs(all []SpeechAPI, search, status string) []SpeechAPI {
	var out []SpeechAPI
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

func apiWSAIAPIsSpeechListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	apis := filterSpeechAPIs(speechAPIStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiSpeechAPI, len(apis))
	for i, a := range apis {
		out[i] = toAPISpeechAPI(a)
	}
	apiOK(w, apiSpeechAPIList{APIs: out, Total: len(out)})
}

type apiCreateSpeechAPIRequest struct {
	Name       string `json:"name"`
	Model      string `json:"model"`
	Capability string `json:"capability"`
	Region     string `json:"region"`
}

func apiWSAIAPIsSpeechCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateSpeechAPIRequest
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
		model = "whisper-large-v3"
	}
	capability := strings.TrimSpace(req.Capability)
	if capability == "" {
		capability = "stt"
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "eu-central-1"
	}
	api := SpeechAPI{
		ID: newID(), Name: name, Status: "provisioning",
		Model: model, Capability: capability,
		Version: "v1.0.0", Requests: "0/mo",
		Latency: "—", Languages: "en",
		SampleRate: "16kHz", Region: region,
		Endpoint: "speech-" + strings.ReplaceAll(strings.ToLower(name), " ", "-") + ".api." + region + ".example.com",
		APIKey:   "vk_new_****" + newID()[:4],
	}
	speechAPIStore.add(slug, api)
	apiCreated(w, toAPISpeechAPI(api))
}

func apiWSAIAPIsSpeechGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	api, found := speechAPIStore.getByID(slug, r.PathValue("aid"))
	if !found {
		apiError(w, http.StatusNotFound, "Speech API not found.")
		return
	}
	apiOK(w, toAPISpeechAPI(api))
}

func apiWSAIAPIsSpeechDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := speechAPIStore.getByID(slug, r.PathValue("aid")); !found {
		apiError(w, http.StatusNotFound, "Speech API not found.")
		return
	}
	speechAPIStore.delete(slug, r.PathValue("aid"))
	apiOK(w, map[string]any{"ok": true, "message": "Speech API deleted."})
}
