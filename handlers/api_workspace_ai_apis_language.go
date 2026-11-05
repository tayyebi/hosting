package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace AI API — APIs: Language/NLP
// ai/apis/language
// ════════════════════════════════════════════════════════════════

type LanguageAPI struct {
	ID         string
	Name       string
	Status     string
	Model      string
	Capability string
	Version    string
	Requests   string
	Latency    string
	Languages  string
	Region     string
	Endpoint   string
	APIKey     string
}

type languageAPIStoreType struct {
	mu   sync.RWMutex
	apis map[string][]LanguageAPI
}

var languageAPIStore = &languageAPIStoreType{
	apis: map[string][]LanguageAPI{},
}

func (s *languageAPIStoreType) list(slug string) []LanguageAPI {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.apis[slug]
	out := make([]LanguageAPI, len(src))
	copy(out, src)
	return out
}

func (s *languageAPIStoreType) getByID(slug, id string) (LanguageAPI, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, a := range s.apis[slug] {
		if a.ID == id {
			return a, true
		}
	}
	return LanguageAPI{}, false
}

func (s *languageAPIStoreType) add(slug string, a LanguageAPI) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.apis[slug] = append(s.apis[slug], a)
}

func (s *languageAPIStoreType) delete(slug, id string) {
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

type apiLanguageAPI struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	Model      string `json:"model"`
	Capability string `json:"capability"`
	Version    string `json:"version"`
	Requests   string `json:"requests"`
	Latency    string `json:"latency"`
	Languages  string `json:"languages"`
	Region     string `json:"region"`
	Endpoint   string `json:"endpoint"`
	APIKey     string `json:"api_key"`
}

type apiLanguageAPIList struct {
	APIs  []apiLanguageAPI `json:"apis"`
	Total int              `json:"total"`
}

func toAPILanguageAPI(a LanguageAPI) apiLanguageAPI {
	return apiLanguageAPI{
		ID: a.ID, Name: a.Name, Status: a.Status,
		Model: a.Model, Capability: a.Capability,
		Version: a.Version, Requests: a.Requests,
		Latency: a.Latency, Languages: a.Languages,
		Region: a.Region, Endpoint: a.Endpoint,
		APIKey: a.APIKey,
	}
}

func filterLanguageAPIs(all []LanguageAPI, search, status string) []LanguageAPI {
	var out []LanguageAPI
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

func apiWSAIAPIsLanguageListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	apis := filterLanguageAPIs(languageAPIStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiLanguageAPI, len(apis))
	for i, a := range apis {
		out[i] = toAPILanguageAPI(a)
	}
	apiOK(w, apiLanguageAPIList{APIs: out, Total: len(out)})
}

type apiCreateLanguageAPIRequest struct {
	Name       string `json:"name"`
	Model      string `json:"model"`
	Capability string `json:"capability"`
	Region     string `json:"region"`
}

func apiWSAIAPIsLanguageCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateLanguageAPIRequest
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
		model = "roberta-base"
	}
	capability := strings.TrimSpace(req.Capability)
	if capability == "" {
		capability = "sentiment"
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "eu-central-1"
	}
	api := LanguageAPI{
		ID: newID(), Name: name, Status: "provisioning",
		Model: model, Capability: capability,
		Version: "v1.0.0", Requests: "0/mo",
		Latency: "—", Languages: "en",
		Region:   region,
		Endpoint: "nlp-" + strings.ReplaceAll(strings.ToLower(name), " ", "-") + ".api." + region + ".example.com",
		APIKey:   "vk_new_****" + newID()[:4],
	}
	languageAPIStore.add(slug, api)
	apiCreated(w, toAPILanguageAPI(api))
}

func apiWSAIAPIsLanguageGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	api, found := languageAPIStore.getByID(slug, r.PathValue("aid"))
	if !found {
		apiError(w, http.StatusNotFound, "Language API not found.")
		return
	}
	apiOK(w, toAPILanguageAPI(api))
}

func apiWSAIAPIsLanguageDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := languageAPIStore.getByID(slug, r.PathValue("aid")); !found {
		apiError(w, http.StatusNotFound, "Language API not found.")
		return
	}
	languageAPIStore.delete(slug, r.PathValue("aid"))
	apiOK(w, map[string]any{"ok": true, "message": "Language API deleted."})
}
