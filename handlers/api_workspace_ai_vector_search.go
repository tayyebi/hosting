package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace AI API — Vector Search: Embeddings (Pro Feature)
// ai/vector-search/embeddings
// ════════════════════════════════════════════════════════════════

type ProVectorCollection struct {
	ID            string
	Name          string
	Dimensions    int
	VectorCount   int64
	IndexType     string
	IndexParams   string
	Metric        string
	QueriesPerSec int
	AvgLatencyMs  float64
	Recall10      float64
	StorageGB     float64
	Replicas      int
	EmbedModel    string
	Region        string
	CostPerMonth  float64
	Status        string
	Updated       string
}

type proVectorCollStoreType struct {
	mu    sync.RWMutex
	items map[string][]ProVectorCollection
}

var proVectorCollStore = &proVectorCollStoreType{
	items: map[string][]ProVectorCollection{},
}

func (s *proVectorCollStoreType) list(slug string) []ProVectorCollection {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.items[slug]
	out := make([]ProVectorCollection, len(src))
	copy(out, src)
	return out
}

func (s *proVectorCollStoreType) getByID(slug, id string) (ProVectorCollection, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, c := range s.items[slug] {
		if c.ID == id {
			return c, true
		}
	}
	return ProVectorCollection{}, false
}

func (s *proVectorCollStoreType) add(slug string, c ProVectorCollection) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[slug] = append(s.items[slug], c)
}

func (s *proVectorCollStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.items[slug]
	for i, c := range list {
		if c.ID == id {
			s.items[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiProVectorCollection struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Dimensions    int     `json:"dimensions"`
	VectorCount   int64   `json:"vector_count"`
	IndexType     string  `json:"index_type"`
	IndexParams   string  `json:"index_params"`
	Metric        string  `json:"metric"`
	QueriesPerSec int     `json:"queries_per_sec"`
	AvgLatencyMs  float64 `json:"avg_latency_ms"`
	Recall10      float64 `json:"recall_at_10"`
	StorageGB     float64 `json:"storage_gb"`
	Replicas      int     `json:"replicas"`
	EmbedModel    string  `json:"embedding_model"`
	Region        string  `json:"region"`
	CostPerMonth  float64 `json:"cost_per_month"`
	Status        string  `json:"status"`
	Updated       string  `json:"updated"`
}

type apiProVectorCollList struct {
	Collections []apiProVectorCollection `json:"collections"`
	Total       int                      `json:"total"`
}

func toAPIProVectorColl(c ProVectorCollection) apiProVectorCollection {
	return apiProVectorCollection{
		ID: c.ID, Name: c.Name, Dimensions: c.Dimensions,
		VectorCount: c.VectorCount, IndexType: c.IndexType,
		IndexParams: c.IndexParams, Metric: c.Metric,
		QueriesPerSec: c.QueriesPerSec, AvgLatencyMs: c.AvgLatencyMs,
		Recall10: c.Recall10, StorageGB: c.StorageGB,
		Replicas: c.Replicas, EmbedModel: c.EmbedModel,
		Region: c.Region, CostPerMonth: c.CostPerMonth,
		Status: c.Status, Updated: c.Updated,
	}
}

func filterProVectorColls(all []ProVectorCollection, search, status string) []ProVectorCollection {
	var out []ProVectorCollection
	q := strings.ToLower(search)
	for _, c := range all {
		if status != "" && c.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(c.Name + " " + c.IndexType + " " + c.EmbedModel + " " + c.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, c)
	}
	return out
}

func apiWSAIVectorSearchListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	items := filterProVectorColls(proVectorCollStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiProVectorCollection, len(items))
	for i, c := range items {
		out[i] = toAPIProVectorColl(c)
	}
	apiOK(w, apiProVectorCollList{Collections: out, Total: len(out)})
}

type apiCreateProVectorCollRequest struct {
	Name       string `json:"name"`
	Dimensions int    `json:"dimensions"`
	IndexType  string `json:"index_type"`
	Metric     string `json:"metric"`
	EmbedModel string `json:"embedding_model"`
	Region     string `json:"region"`
}

func apiWSAIVectorSearchCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateProVectorCollRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	dims := req.Dimensions
	if dims <= 0 {
		dims = 1536
	}
	indexType := strings.TrimSpace(req.IndexType)
	if indexType == "" {
		indexType = "HNSW"
	}
	metric := strings.TrimSpace(req.Metric)
	if metric == "" {
		metric = "cosine"
	}
	embedModel := strings.TrimSpace(req.EmbedModel)
	if embedModel == "" {
		embedModel = "text-embedding-3-large"
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "us-east-1"
	}
	coll := ProVectorCollection{
		ID: newID(), Name: name, Dimensions: dims,
		VectorCount: 0, IndexType: indexType,
		IndexParams: "M=32, efConstruction=200",
		Metric:      metric, QueriesPerSec: 0,
		AvgLatencyMs: 0, Recall10: 0,
		StorageGB: 0, Replicas: 1,
		EmbedModel: embedModel, Region: region,
		CostPerMonth: 45.00,
		Status:       "creating", Updated: "2026-04-18",
	}
	proVectorCollStore.add(slug, coll)
	apiCreated(w, toAPIProVectorColl(coll))
}

func apiWSAIVectorSearchGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	c, found := proVectorCollStore.getByID(slug, r.PathValue("cid"))
	if !found {
		apiError(w, http.StatusNotFound, "Vector collection not found.")
		return
	}
	apiOK(w, toAPIProVectorColl(c))
}

func apiWSAIVectorSearchDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := proVectorCollStore.getByID(slug, r.PathValue("cid")); !found {
		apiError(w, http.StatusNotFound, "Vector collection not found.")
		return
	}
	proVectorCollStore.delete(slug, r.PathValue("cid"))
	apiOK(w, map[string]any{"ok": true, "message": "Vector collection deleted."})
}
