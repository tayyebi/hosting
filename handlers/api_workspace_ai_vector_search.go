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
	items: map[string][]ProVectorCollection{
		"buyvm": {
			{ID: newID(), Name: "product-embeddings", Dimensions: 1536, VectorCount: 42000000, IndexType: "HNSW", IndexParams: "M=32, efConstruction=200", Metric: "cosine", QueriesPerSec: 1840, AvgLatencyMs: 2.4, Recall10: 97.8, StorageGB: 128.4, Replicas: 3, EmbedModel: "text-embedding-3-large", Region: "us-east-1", CostPerMonth: 420.00, Status: "serving", Updated: "2026-04-18"},
			{ID: newID(), Name: "support-articles", Dimensions: 768, VectorCount: 2400000, IndexType: "HNSW", IndexParams: "M=16, efConstruction=128", Metric: "cosine", QueriesPerSec: 620, AvgLatencyMs: 1.2, Recall10: 98.4, StorageGB: 4.8, Replicas: 2, EmbedModel: "text-embedding-3-small", Region: "us-east-1", CostPerMonth: 85.00, Status: "serving", Updated: "2026-04-17"},
			{ID: newID(), Name: "user-profiles", Dimensions: 1536, VectorCount: 18000000, IndexType: "IVF-PQ", IndexParams: "nlist=4096, m=64", Metric: "dotproduct", QueriesPerSec: 1200, AvgLatencyMs: 3.8, Recall10: 94.2, StorageGB: 42.6, Replicas: 2, EmbedModel: "text-embedding-3-large", Region: "us-east-1", CostPerMonth: 240.00, Status: "serving", Updated: "2026-04-16"},
			{ID: newID(), Name: "code-search", Dimensions: 1024, VectorCount: 64000000, IndexType: "HNSW", IndexParams: "M=48, efConstruction=256", Metric: "cosine", QueriesPerSec: 0, AvgLatencyMs: 0, Recall10: 0, StorageGB: 156.8, Replicas: 3, EmbedModel: "code-search-v2", Region: "us-east-1", CostPerMonth: 380.00, Status: "indexing", Updated: "2026-04-18"},
			{ID: newID(), Name: "image-features", Dimensions: 512, VectorCount: 1800000, IndexType: "HNSW", IndexParams: "M=16, efConstruction=128", Metric: "cosine", QueriesPerSec: 340, AvgLatencyMs: 0.8, Recall10: 99.1, StorageGB: 2.2, Replicas: 1, EmbedModel: "CLIP-ViT-B/32", Region: "us-east-1", CostPerMonth: 45.00, Status: "serving", Updated: "2026-04-15"},
		},
		"hosting": {
			{ID: newID(), Name: "catalog-semantic", Dimensions: 1536, VectorCount: 280000000, IndexType: "HNSW", IndexParams: "M=64, efConstruction=400", Metric: "cosine", QueriesPerSec: 8400, AvgLatencyMs: 3.2, Recall10: 96.8, StorageGB: 840.0, Replicas: 6, EmbedModel: "text-embedding-3-large", Region: "eu-central-1", CostPerMonth: 2400.00, Status: "serving", Updated: "2026-04-18"},
			{ID: newID(), Name: "rag-knowledge-base", Dimensions: 1536, VectorCount: 12000000, IndexType: "HNSW", IndexParams: "M=32, efConstruction=200", Metric: "cosine", QueriesPerSec: 2400, AvgLatencyMs: 1.8, Recall10: 98.2, StorageGB: 36.8, Replicas: 3, EmbedModel: "text-embedding-3-large", Region: "eu-central-1", CostPerMonth: 320.00, Status: "serving", Updated: "2026-04-17"},
			{ID: newID(), Name: "anomaly-detection", Dimensions: 256, VectorCount: 4200000, IndexType: "IVF-Flat", IndexParams: "nlist=2048", Metric: "euclidean", QueriesPerSec: 12000, AvgLatencyMs: 0.6, Recall10: 99.4, StorageGB: 2.8, Replicas: 2, EmbedModel: "custom-autoencoder-v4", Region: "eu-central-1", CostPerMonth: 95.00, Status: "serving", Updated: "2026-04-17"},
			{ID: newID(), Name: "multimodal-search", Dimensions: 768, VectorCount: 8600000, IndexType: "HNSW", IndexParams: "M=24, efConstruction=160", Metric: "cosine", QueriesPerSec: 1600, AvgLatencyMs: 2.1, Recall10: 97.5, StorageGB: 16.4, Replicas: 2, EmbedModel: "CLIP-ViT-L/14", Region: "us-west-2", CostPerMonth: 180.00, Status: "serving", Updated: "2026-04-16"},
		},
	},
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
		Metric: metric, QueriesPerSec: 0,
		AvgLatencyMs: 0, Recall10: 0,
		StorageGB: 0, Replicas: 1,
		EmbedModel: embedModel, Region: region,
		CostPerMonth: 45.00,
		Status: "creating", Updated: "2026-04-18",
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
