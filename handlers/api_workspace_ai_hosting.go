package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace AI API — Hosting: Inference Endpoints
// ai/hosting/inference
// ════════════════════════════════════════════════════════════════

type InferenceEndpoint struct {
	ID         string
	Name       string
	Status     string
	Model      string
	Framework  string
	Instance   string
	Replicas   int
	Latency    string
	RPS        string
	Region     string
	Endpoint   string
	AutoScale  bool
}

type inferenceEndpointStoreType struct {
	mu        sync.RWMutex
	endpoints map[string][]InferenceEndpoint
}

var inferenceEndpointStore = &inferenceEndpointStoreType{
	endpoints: map[string][]InferenceEndpoint{
		"buyvm": {
			{ID: newID(), Name: "chatbot-prod", Status: "running", Model: "ft:support-chatbot-v3", Framework: "vLLM", Instance: "gpu.a100.1x", Replicas: 3, Latency: "124ms p99", RPS: "420", Region: "eu-central-1", Endpoint: "chatbot-prod.inf.eu-central-1.example.com", AutoScale: true},
			{ID: newID(), Name: "product-classifier-api", Status: "running", Model: "ft:product-classifier", Framework: "TorchServe", Instance: "gpu.t4.1x", Replicas: 2, Latency: "18ms p99", RPS: "2,800", Region: "eu-central-1", Endpoint: "product-classifier.inf.eu-central-1.example.com", AutoScale: true},
			{ID: newID(), Name: "embedding-service", Status: "running", Model: "e5-large-v2", Framework: "Triton", Instance: "gpu.t4.1x", Replicas: 4, Latency: "8ms p99", RPS: "12,000", Region: "us-east-1", Endpoint: "embeddings.inf.us-east-1.example.com", AutoScale: true},
			{ID: newID(), Name: "image-gen-api", Status: "running", Model: "sdxl-1.0", Framework: "vLLM", Instance: "gpu.a100.1x", Replicas: 2, Latency: "3.2s p99", RPS: "45", Region: "eu-central-1", Endpoint: "image-gen.inf.eu-central-1.example.com", AutoScale: false},
			{ID: newID(), Name: "staging-llm", Status: "stopped", Model: "mistral-7b", Framework: "vLLM", Instance: "gpu.t4.1x", Replicas: 0, Latency: "—", RPS: "0", Region: "eu-central-1", Endpoint: "staging-llm.inf.eu-central-1.example.com", AutoScale: false},
		},
		"hosting": {
			{ID: newID(), Name: "intent-classifier", Status: "running", Model: "ft:customer-intent", Framework: "TorchServe", Instance: "gpu.t4.1x", Replicas: 2, Latency: "22ms p99", RPS: "1,800", Region: "ap-south-1", Endpoint: "intent.inf.ap-south-1.example.com", AutoScale: true},
			{ID: newID(), Name: "summarizer-api", Status: "running", Model: "ft:doc-summarizer", Framework: "vLLM", Instance: "gpu.a10g.1x", Replicas: 1, Latency: "840ms p99", RPS: "120", Region: "ap-south-1", Endpoint: "summarizer.inf.ap-south-1.example.com", AutoScale: true},
			{ID: newID(), Name: "ner-service", Status: "running", Model: "ft:ner-custom", Framework: "TorchServe", Instance: "cpu.c6i.2xl", Replicas: 3, Latency: "12ms p99", RPS: "4,200", Region: "ap-south-1", Endpoint: "ner.inf.ap-south-1.example.com", AutoScale: true},
		},
	},
}

func (s *inferenceEndpointStoreType) list(slug string) []InferenceEndpoint {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.endpoints[slug]
	out := make([]InferenceEndpoint, len(src))
	copy(out, src)
	return out
}

func (s *inferenceEndpointStoreType) getByID(slug, id string) (InferenceEndpoint, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, e := range s.endpoints[slug] {
		if e.ID == id {
			return e, true
		}
	}
	return InferenceEndpoint{}, false
}

func (s *inferenceEndpointStoreType) add(slug string, e InferenceEndpoint) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.endpoints[slug] = append(s.endpoints[slug], e)
}

func (s *inferenceEndpointStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.endpoints[slug]
	for i, e := range list {
		if e.ID == id {
			s.endpoints[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiInferenceEndpoint struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	Model     string `json:"model"`
	Framework string `json:"framework"`
	Instance  string `json:"instance"`
	Replicas  int    `json:"replicas"`
	Latency   string `json:"latency"`
	RPS       string `json:"rps"`
	Region    string `json:"region"`
	Endpoint  string `json:"endpoint"`
	AutoScale bool   `json:"auto_scale"`
}

type apiInferenceEndpointList struct {
	Endpoints []apiInferenceEndpoint `json:"endpoints"`
	Total     int                    `json:"total"`
}

func toAPIInferenceEndpoint(e InferenceEndpoint) apiInferenceEndpoint {
	return apiInferenceEndpoint{
		ID: e.ID, Name: e.Name, Status: e.Status,
		Model: e.Model, Framework: e.Framework,
		Instance: e.Instance, Replicas: e.Replicas,
		Latency: e.Latency, RPS: e.RPS,
		Region: e.Region, Endpoint: e.Endpoint,
		AutoScale: e.AutoScale,
	}
}

func filterInferenceEndpoints(all []InferenceEndpoint, search, status string) []InferenceEndpoint {
	var out []InferenceEndpoint
	q := strings.ToLower(search)
	for _, e := range all {
		if status != "" && e.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(e.Name + " " + e.Model + " " + e.Framework + " " + e.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, e)
	}
	return out
}

func apiWSAIHostInferenceListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	eps := filterInferenceEndpoints(inferenceEndpointStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiInferenceEndpoint, len(eps))
	for i, e := range eps {
		out[i] = toAPIInferenceEndpoint(e)
	}
	apiOK(w, apiInferenceEndpointList{Endpoints: out, Total: len(out)})
}

type apiCreateInferenceEndpointRequest struct {
	Name      string `json:"name"`
	Model     string `json:"model"`
	Framework string `json:"framework"`
	Instance  string `json:"instance"`
	Replicas  int    `json:"replicas"`
	Region    string `json:"region"`
	AutoScale bool   `json:"auto_scale"`
}

func apiWSAIHostInferenceCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateInferenceEndpointRequest
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
		apiError(w, http.StatusBadRequest, "Model is required.")
		return
	}
	framework := strings.TrimSpace(req.Framework)
	if framework == "" {
		framework = "vLLM"
	}
	instance := strings.TrimSpace(req.Instance)
	if instance == "" {
		instance = "gpu.t4.1x"
	}
	replicas := req.Replicas
	if replicas <= 0 {
		replicas = 1
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "eu-central-1"
	}
	ep := InferenceEndpoint{
		ID: newID(), Name: name, Status: "provisioning",
		Model: model, Framework: framework,
		Instance: instance, Replicas: replicas,
		Latency: "—", RPS: "0",
		Region: region,
		Endpoint: name + ".inf." + region + ".example.com",
		AutoScale: req.AutoScale,
	}
	inferenceEndpointStore.add(slug, ep)
	apiCreated(w, toAPIInferenceEndpoint(ep))
}

func apiWSAIHostInferenceGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	ep, found := inferenceEndpointStore.getByID(slug, r.PathValue("eid"))
	if !found {
		apiError(w, http.StatusNotFound, "Inference endpoint not found.")
		return
	}
	apiOK(w, toAPIInferenceEndpoint(ep))
}

func apiWSAIHostInferenceDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := inferenceEndpointStore.getByID(slug, r.PathValue("eid")); !found {
		apiError(w, http.StatusNotFound, "Inference endpoint not found.")
		return
	}
	inferenceEndpointStore.delete(slug, r.PathValue("eid"))
	apiOK(w, map[string]any{"ok": true, "message": "Inference endpoint deleted."})
}

// ════════════════════════════════════════════════════════════════
// Workspace AI API — Hosting: Batch Inference
// ai/hosting/batch
// ════════════════════════════════════════════════════════════════

type BatchInference struct {
	ID        string
	Name      string
	Status    string
	Model     string
	InputPath string
	OutputPath string
	Records   string
	Processed string
	Progress  string
	Runtime   string
	Cost      string
	Region    string
}

type batchInferenceStoreType struct {
	mu   sync.RWMutex
	jobs map[string][]BatchInference
}

var batchInferenceStore = &batchInferenceStoreType{
	jobs: map[string][]BatchInference{
		"buyvm": {
			{ID: newID(), Name: "product-embeddings-full", Status: "completed", Model: "e5-large-v2", InputPath: "s3://ml-data/products.jsonl", OutputPath: "s3://ml-data/embeddings/", Records: "890,000", Processed: "890,000", Progress: "100%", Runtime: "2h 14m", Cost: "$42", Region: "eu-central-1"},
			{ID: newID(), Name: "sentiment-bulk-2026q1", Status: "running", Model: "ft:sentiment-analyzer-v2", InputPath: "s3://ml-data/reviews-q1.jsonl", OutputPath: "s3://ml-data/sentiment-q1/", Records: "1,240,000", Processed: "820,000", Progress: "66%", Runtime: "3h 40m", Cost: "$86", Region: "eu-central-1"},
			{ID: newID(), Name: "image-classification-batch", Status: "completed", Model: "resnet-50", InputPath: "s3://ml-data/images/", OutputPath: "s3://ml-data/labels/", Records: "2,400,000", Processed: "2,400,000", Progress: "100%", Runtime: "5h 20m", Cost: "$124", Region: "us-east-1"},
			{ID: newID(), Name: "doc-summarization-weekly", Status: "queued", Model: "ft:doc-summarizer", InputPath: "s3://ml-data/docs-week16.jsonl", OutputPath: "s3://ml-data/summaries-w16/", Records: "18,000", Processed: "0", Progress: "0%", Runtime: "—", Cost: "$0", Region: "eu-central-1"},
		},
		"hosting": {
			{ID: newID(), Name: "intent-classification-export", Status: "completed", Model: "ft:customer-intent", InputPath: "s3://ml-data/chats-march.jsonl", OutputPath: "s3://ml-data/intents-march/", Records: "340,000", Processed: "340,000", Progress: "100%", Runtime: "48m", Cost: "$18", Region: "ap-south-1"},
			{ID: newID(), Name: "ner-bulk-extract", Status: "running", Model: "ft:ner-custom", InputPath: "s3://ml-data/contracts.jsonl", OutputPath: "s3://ml-data/entities/", Records: "42,000", Processed: "28,000", Progress: "67%", Runtime: "1h 12m", Cost: "$8", Region: "ap-south-1"},
		},
	},
}

func (s *batchInferenceStoreType) list(slug string) []BatchInference {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.jobs[slug]
	out := make([]BatchInference, len(src))
	copy(out, src)
	return out
}

func (s *batchInferenceStoreType) getByID(slug, id string) (BatchInference, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, j := range s.jobs[slug] {
		if j.ID == id {
			return j, true
		}
	}
	return BatchInference{}, false
}

func (s *batchInferenceStoreType) add(slug string, j BatchInference) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[slug] = append(s.jobs[slug], j)
}

func (s *batchInferenceStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.jobs[slug]
	for i, j := range list {
		if j.ID == id {
			s.jobs[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiBatchInference struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	Model      string `json:"model"`
	InputPath  string `json:"input_path"`
	OutputPath string `json:"output_path"`
	Records    string `json:"records"`
	Processed  string `json:"processed"`
	Progress   string `json:"progress"`
	Runtime    string `json:"runtime"`
	Cost       string `json:"cost"`
	Region     string `json:"region"`
}

type apiBatchInferenceList struct {
	Jobs  []apiBatchInference `json:"jobs"`
	Total int                 `json:"total"`
}

func toAPIBatchInference(j BatchInference) apiBatchInference {
	return apiBatchInference{
		ID: j.ID, Name: j.Name, Status: j.Status,
		Model: j.Model, InputPath: j.InputPath,
		OutputPath: j.OutputPath, Records: j.Records,
		Processed: j.Processed, Progress: j.Progress,
		Runtime: j.Runtime, Cost: j.Cost, Region: j.Region,
	}
}

func filterBatchInference(all []BatchInference, search, status string) []BatchInference {
	var out []BatchInference
	q := strings.ToLower(search)
	for _, j := range all {
		if status != "" && j.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(j.Name + " " + j.Model + " " + j.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, j)
	}
	return out
}

func apiWSAIHostBatchListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	jobs := filterBatchInference(batchInferenceStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiBatchInference, len(jobs))
	for i, j := range jobs {
		out[i] = toAPIBatchInference(j)
	}
	apiOK(w, apiBatchInferenceList{Jobs: out, Total: len(out)})
}

type apiCreateBatchInferenceRequest struct {
	Name       string `json:"name"`
	Model      string `json:"model"`
	InputPath  string `json:"input_path"`
	OutputPath string `json:"output_path"`
	Region     string `json:"region"`
}

func apiWSAIHostBatchCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateBatchInferenceRequest
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
		apiError(w, http.StatusBadRequest, "Model is required.")
		return
	}
	inputPath := strings.TrimSpace(req.InputPath)
	if inputPath == "" {
		apiError(w, http.StatusBadRequest, "Input path is required.")
		return
	}
	outputPath := strings.TrimSpace(req.OutputPath)
	if outputPath == "" {
		outputPath = "s3://ml-data/output/" + name + "/"
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "eu-central-1"
	}
	job := BatchInference{
		ID: newID(), Name: name, Status: "queued",
		Model: model, InputPath: inputPath,
		OutputPath: outputPath, Records: "—",
		Processed: "0", Progress: "0%",
		Runtime: "—", Cost: "$0", Region: region,
	}
	batchInferenceStore.add(slug, job)
	apiCreated(w, toAPIBatchInference(job))
}

func apiWSAIHostBatchGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	job, found := batchInferenceStore.getByID(slug, r.PathValue("bid"))
	if !found {
		apiError(w, http.StatusNotFound, "Batch inference job not found.")
		return
	}
	apiOK(w, toAPIBatchInference(job))
}

func apiWSAIHostBatchDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := batchInferenceStore.getByID(slug, r.PathValue("bid")); !found {
		apiError(w, http.StatusNotFound, "Batch inference job not found.")
		return
	}
	batchInferenceStore.delete(slug, r.PathValue("bid"))
	apiOK(w, map[string]any{"ok": true, "message": "Batch inference job deleted."})
}

// ════════════════════════════════════════════════════════════════
// Workspace AI API — Hosting: Edge Deployment
// ai/hosting/edge
// ════════════════════════════════════════════════════════════════

type EdgeDeployment struct {
	ID        string
	Name      string
	Status    string
	Model     string
	Runtime   string
	Target    string
	Devices   int
	Version   string
	Latency   string
	Region    string
}

type edgeDeploymentStoreType struct {
	mu          sync.RWMutex
	deployments map[string][]EdgeDeployment
}

var edgeDeploymentStore = &edgeDeploymentStoreType{
	deployments: map[string][]EdgeDeployment{
		"buyvm": {
			{ID: newID(), Name: "anomaly-detector-factory", Status: "deployed", Model: "anomaly-det-v2-quantized", Runtime: "ONNX Runtime", Target: "Jetson Orin", Devices: 24, Version: "v2.1.0", Latency: "4ms", Region: "eu-central-1"},
			{ID: newID(), Name: "quality-inspector", Status: "deployed", Model: "yolov8-custom-int8", Runtime: "TensorRT", Target: "Jetson AGX", Devices: 8, Version: "v1.4.0", Latency: "12ms", Region: "eu-central-1"},
			{ID: newID(), Name: "voice-assistant-edge", Status: "deployed", Model: "whisper-tiny-quantized", Runtime: "ONNX Runtime", Target: "Coral Edge TPU", Devices: 120, Version: "v3.0.1", Latency: "180ms", Region: "global"},
			{ID: newID(), Name: "retail-counter", Status: "updating", Model: "mobilenet-v3-int8", Runtime: "TFLite", Target: "Raspberry Pi 5", Devices: 42, Version: "v1.1.0 → v1.2.0", Latency: "45ms", Region: "us-east-1"},
		},
		"hosting": {
			{ID: newID(), Name: "pos-product-scan", Status: "deployed", Model: "efficientnet-lite-int8", Runtime: "TFLite", Target: "Android Tablet", Devices: 180, Version: "v2.0.0", Latency: "28ms", Region: "ap-south-1"},
			{ID: newID(), Name: "occupancy-sensor", Status: "deployed", Model: "person-det-tiny", Runtime: "ONNX Runtime", Target: "Jetson Nano", Devices: 35, Version: "v1.3.0", Latency: "8ms", Region: "ap-south-1"},
			{ID: newID(), Name: "language-id-edge", Status: "deployed", Model: "lid-micro-quantized", Runtime: "TFLite", Target: "ESP32-S3", Devices: 500, Version: "v1.0.0", Latency: "15ms", Region: "global"},
		},
	},
}

func (s *edgeDeploymentStoreType) list(slug string) []EdgeDeployment {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.deployments[slug]
	out := make([]EdgeDeployment, len(src))
	copy(out, src)
	return out
}

func (s *edgeDeploymentStoreType) getByID(slug, id string) (EdgeDeployment, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, d := range s.deployments[slug] {
		if d.ID == id {
			return d, true
		}
	}
	return EdgeDeployment{}, false
}

func (s *edgeDeploymentStoreType) add(slug string, d EdgeDeployment) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.deployments[slug] = append(s.deployments[slug], d)
}

func (s *edgeDeploymentStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.deployments[slug]
	for i, d := range list {
		if d.ID == id {
			s.deployments[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiEdgeDeployment struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Status  string `json:"status"`
	Model   string `json:"model"`
	Runtime string `json:"runtime"`
	Target  string `json:"target"`
	Devices int    `json:"devices"`
	Version string `json:"version"`
	Latency string `json:"latency"`
	Region  string `json:"region"`
}

type apiEdgeDeploymentList struct {
	Deployments []apiEdgeDeployment `json:"deployments"`
	Total       int                 `json:"total"`
}

func toAPIEdgeDeployment(d EdgeDeployment) apiEdgeDeployment {
	return apiEdgeDeployment{
		ID: d.ID, Name: d.Name, Status: d.Status,
		Model: d.Model, Runtime: d.Runtime,
		Target: d.Target, Devices: d.Devices,
		Version: d.Version, Latency: d.Latency,
		Region: d.Region,
	}
}

func filterEdgeDeployments(all []EdgeDeployment, search, status string) []EdgeDeployment {
	var out []EdgeDeployment
	q := strings.ToLower(search)
	for _, d := range all {
		if status != "" && d.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(d.Name + " " + d.Model + " " + d.Target + " " + d.Runtime + " " + d.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, d)
	}
	return out
}

func apiWSAIHostEdgeListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	deps := filterEdgeDeployments(edgeDeploymentStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiEdgeDeployment, len(deps))
	for i, d := range deps {
		out[i] = toAPIEdgeDeployment(d)
	}
	apiOK(w, apiEdgeDeploymentList{Deployments: out, Total: len(out)})
}

type apiCreateEdgeDeploymentRequest struct {
	Name    string `json:"name"`
	Model   string `json:"model"`
	Runtime string `json:"runtime"`
	Target  string `json:"target"`
	Region  string `json:"region"`
}

func apiWSAIHostEdgeCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateEdgeDeploymentRequest
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
		apiError(w, http.StatusBadRequest, "Model is required.")
		return
	}
	runtime := strings.TrimSpace(req.Runtime)
	if runtime == "" {
		runtime = "ONNX Runtime"
	}
	target := strings.TrimSpace(req.Target)
	if target == "" {
		target = "Jetson Nano"
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "eu-central-1"
	}
	dep := EdgeDeployment{
		ID: newID(), Name: name, Status: "provisioning",
		Model: model, Runtime: runtime,
		Target: target, Devices: 0,
		Version: "v0.1.0", Latency: "—",
		Region: region,
	}
	edgeDeploymentStore.add(slug, dep)
	apiCreated(w, toAPIEdgeDeployment(dep))
}

func apiWSAIHostEdgeGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	dep, found := edgeDeploymentStore.getByID(slug, r.PathValue("did"))
	if !found {
		apiError(w, http.StatusNotFound, "Edge deployment not found.")
		return
	}
	apiOK(w, toAPIEdgeDeployment(dep))
}

func apiWSAIHostEdgeDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := edgeDeploymentStore.getByID(slug, r.PathValue("did")); !found {
		apiError(w, http.StatusNotFound, "Edge deployment not found.")
		return
	}
	edgeDeploymentStore.delete(slug, r.PathValue("did"))
	apiOK(w, map[string]any{"ok": true, "message": "Edge deployment deleted."})
}
