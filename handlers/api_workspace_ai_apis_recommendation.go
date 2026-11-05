package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace AI API — APIs: Recommendation
// ai/apis/recommendation
// ════════════════════════════════════════════════════════════════

type RecommendationAPI struct {
	ID        string
	Name      string
	Status    string
	Model     string
	Algorithm string
	Version   string
	Requests  string
	Latency   string
	Catalog   string
	Region    string
	Endpoint  string
	APIKey    string
}

type recommendationAPIStoreType struct {
	mu   sync.RWMutex
	apis map[string][]RecommendationAPI
}

var recommendationAPIStore = &recommendationAPIStoreType{
	apis: map[string][]RecommendationAPI{
		"buyvm": {
			{ID: newID(), Name: "product-recommender-v2", Status: "active", Model: "deepfm-v2", Algorithm: "hybrid", Version: "v2.0.0", Requests: "4.8M/mo", Latency: "18ms p99", Catalog: "2.4M products", Region: "eu-central-1", Endpoint: "recommend.api.eu-central-1.example.com", APIKey: "vk_prod_****6e1a"},
			{ID: newID(), Name: "content-personalization", Status: "active", Model: "two-tower-bert", Algorithm: "content-based", Version: "v1.5.0", Requests: "2.1M/mo", Latency: "24ms p99", Catalog: "180K articles", Region: "eu-central-1", Endpoint: "personalize.api.eu-central-1.example.com", APIKey: "vk_prod_****b3f9"},
			{ID: newID(), Name: "similar-items", Status: "active", Model: "item2vec-custom", Algorithm: "collaborative", Version: "v1.2.0", Requests: "1.6M/mo", Latency: "12ms p99", Catalog: "2.4M products", Region: "eu-central-1", Endpoint: "similar.api.eu-central-1.example.com", APIKey: "vk_prod_****d8c4"},
			{ID: newID(), Name: "contextual-offers", Status: "active", Model: "bandit-contextual", Algorithm: "contextual", Version: "v1.0.0", Requests: "920K/mo", Latency: "8ms p99", Catalog: "45K offers", Region: "us-east-1", Endpoint: "offers.api.us-east-1.example.com", APIKey: "vk_prod_****a2e7"},
		},
		"hosting": {
			{ID: newID(), Name: "shop-recommender", Status: "active", Model: "als-spark", Algorithm: "collaborative", Version: "v1.0.0", Requests: "680K/mo", Latency: "22ms p99", Catalog: "320K products", Region: "ap-south-1", Endpoint: "shop-recommend.api.ap-south-1.example.com", APIKey: "vk_prod_****f5b8"},
			{ID: newID(), Name: "news-feed", Status: "active", Model: "dssm-custom", Algorithm: "content-based", Version: "v1.1.0", Requests: "1.2M/mo", Latency: "16ms p99", Catalog: "50K articles", Region: "ap-south-1", Endpoint: "news-feed.api.ap-south-1.example.com", APIKey: "vk_prod_****c7d3"},
			{ID: newID(), Name: "cross-sell-engine", Status: "active", Model: "wide-and-deep", Algorithm: "hybrid", Version: "v1.0.0", Requests: "380K/mo", Latency: "28ms p99", Catalog: "320K products", Region: "ap-south-1", Endpoint: "cross-sell.api.ap-south-1.example.com", APIKey: "vk_prod_****e1a6"},
		},
	},
}

func (s *recommendationAPIStoreType) list(slug string) []RecommendationAPI {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.apis[slug]
	out := make([]RecommendationAPI, len(src))
	copy(out, src)
	return out
}

func (s *recommendationAPIStoreType) getByID(slug, id string) (RecommendationAPI, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, a := range s.apis[slug] {
		if a.ID == id {
			return a, true
		}
	}
	return RecommendationAPI{}, false
}

func (s *recommendationAPIStoreType) add(slug string, a RecommendationAPI) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.apis[slug] = append(s.apis[slug], a)
}

func (s *recommendationAPIStoreType) delete(slug, id string) {
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

type apiRecommendationAPI struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	Model     string `json:"model"`
	Algorithm string `json:"algorithm"`
	Version   string `json:"version"`
	Requests  string `json:"requests"`
	Latency   string `json:"latency"`
	Catalog   string `json:"catalog"`
	Region    string `json:"region"`
	Endpoint  string `json:"endpoint"`
	APIKey    string `json:"api_key"`
}

type apiRecommendationAPIList struct {
	APIs  []apiRecommendationAPI `json:"apis"`
	Total int                    `json:"total"`
}

func toAPIRecommendationAPI(a RecommendationAPI) apiRecommendationAPI {
	return apiRecommendationAPI{
		ID: a.ID, Name: a.Name, Status: a.Status,
		Model: a.Model, Algorithm: a.Algorithm,
		Version: a.Version, Requests: a.Requests,
		Latency: a.Latency, Catalog: a.Catalog,
		Region: a.Region, Endpoint: a.Endpoint,
		APIKey: a.APIKey,
	}
}

func filterRecommendationAPIs(all []RecommendationAPI, search, status string) []RecommendationAPI {
	var out []RecommendationAPI
	q := strings.ToLower(search)
	for _, a := range all {
		if status != "" && a.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(a.Name + " " + a.Model + " " + a.Algorithm + " " + a.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, a)
	}
	return out
}

func apiWSAIAPIsRecommendListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	apis := filterRecommendationAPIs(recommendationAPIStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiRecommendationAPI, len(apis))
	for i, a := range apis {
		out[i] = toAPIRecommendationAPI(a)
	}
	apiOK(w, apiRecommendationAPIList{APIs: out, Total: len(out)})
}

type apiCreateRecommendationAPIRequest struct {
	Name      string `json:"name"`
	Model     string `json:"model"`
	Algorithm string `json:"algorithm"`
	Region    string `json:"region"`
}

func apiWSAIAPIsRecommendCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateRecommendationAPIRequest
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
		model = "deepfm-v2"
	}
	algorithm := strings.TrimSpace(req.Algorithm)
	if algorithm == "" {
		algorithm = "hybrid"
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "eu-central-1"
	}
	api := RecommendationAPI{
		ID: newID(), Name: name, Status: "provisioning",
		Model: model, Algorithm: algorithm,
		Version: "v1.0.0", Requests: "0/mo",
		Latency: "—", Catalog: "0 items",
		Region: region,
		Endpoint: "recommend-" + strings.ReplaceAll(strings.ToLower(name), " ", "-") + ".api." + region + ".example.com",
		APIKey: "vk_new_****" + newID()[:4],
	}
	recommendationAPIStore.add(slug, api)
	apiCreated(w, toAPIRecommendationAPI(api))
}

func apiWSAIAPIsRecommendGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	api, found := recommendationAPIStore.getByID(slug, r.PathValue("aid"))
	if !found {
		apiError(w, http.StatusNotFound, "Recommendation API not found.")
		return
	}
	apiOK(w, toAPIRecommendationAPI(api))
}

func apiWSAIAPIsRecommendDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := recommendationAPIStore.getByID(slug, r.PathValue("aid")); !found {
		apiError(w, http.StatusNotFound, "Recommendation API not found.")
		return
	}
	recommendationAPIStore.delete(slug, r.PathValue("aid"))
	apiOK(w, map[string]any{"ok": true, "message": "Recommendation API deleted."})
}
