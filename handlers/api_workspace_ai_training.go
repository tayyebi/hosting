package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace AI API — Training: Distributed Training
// ai/training/distributed
// ════════════════════════════════════════════════════════════════

type DistributedTraining struct {
	ID         string
	Name       string
	Status     string
	Framework  string
	Nodes      int
	GPUsPerNode int
	GPUType    string
	Dataset    string
	Epochs     int
	Progress   string
	Runtime    string
	Cost       string
	Region     string
}

type distributedTrainingStoreType struct {
	mu   sync.RWMutex
	jobs map[string][]DistributedTraining
}

var distributedTrainingStore = &distributedTrainingStoreType{
	jobs: map[string][]DistributedTraining{
		"buyvm": {
			{ID: newID(), Name: "llm-pretrain-7b", Status: "running", Framework: "PyTorch + DeepSpeed", Nodes: 8, GPUsPerNode: 8, GPUType: "A100 80GB", Dataset: "webtext-2024 (420 GB)", Epochs: 3, Progress: "67%", Runtime: "14d 6h", Cost: "$42,800", Region: "us-east-1"},
			{ID: newID(), Name: "vision-transformer-finetune", Status: "running", Framework: "PyTorch + FSDP", Nodes: 4, GPUsPerNode: 4, GPUType: "A100 40GB", Dataset: "imagenet-22k (256 GB)", Epochs: 20, Progress: "84%", Runtime: "3d 12h", Cost: "$6,200", Region: "eu-central-1"},
			{ID: newID(), Name: "speech-recognition-v3", Status: "completed", Framework: "JAX + Flax", Nodes: 16, GPUsPerNode: 8, GPUType: "H100", Dataset: "librispeech-960h (120 GB)", Epochs: 50, Progress: "100%", Runtime: "7d 0h", Cost: "$28,400", Region: "us-west-2"},
			{ID: newID(), Name: "recommendation-model", Status: "queued", Framework: "TensorFlow", Nodes: 2, GPUsPerNode: 4, GPUType: "A100 40GB", Dataset: "user-interactions (89 GB)", Epochs: 10, Progress: "0%", Runtime: "—", Cost: "$0", Region: "eu-central-1"},
		},
		"hosting": {
			{ID: newID(), Name: "nlp-multilingual", Status: "running", Framework: "PyTorch + DeepSpeed", Nodes: 4, GPUsPerNode: 8, GPUType: "A100 80GB", Dataset: "mc4-multilingual (1.2 TB)", Epochs: 5, Progress: "32%", Runtime: "8d 4h", Cost: "$18,600", Region: "ap-south-1"},
			{ID: newID(), Name: "fraud-detection-train", Status: "completed", Framework: "PyTorch", Nodes: 2, GPUsPerNode: 4, GPUType: "V100", Dataset: "transactions-2024 (45 GB)", Epochs: 30, Progress: "100%", Runtime: "1d 8h", Cost: "$1,200", Region: "ap-south-1"},
			{ID: newID(), Name: "image-segmentation-v2", Status: "running", Framework: "PyTorch + FSDP", Nodes: 4, GPUsPerNode: 4, GPUType: "A100 40GB", Dataset: "coco-2024 (38 GB)", Epochs: 100, Progress: "56%", Runtime: "2d 18h", Cost: "$3,800", Region: "eu-west-1"},
		},
	},
}

func (s *distributedTrainingStoreType) list(slug string) []DistributedTraining {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.jobs[slug]
	out := make([]DistributedTraining, len(src))
	copy(out, src)
	return out
}

func (s *distributedTrainingStoreType) getByID(slug, id string) (DistributedTraining, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, j := range s.jobs[slug] {
		if j.ID == id {
			return j, true
		}
	}
	return DistributedTraining{}, false
}

func (s *distributedTrainingStoreType) add(slug string, j DistributedTraining) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[slug] = append(s.jobs[slug], j)
}

func (s *distributedTrainingStoreType) delete(slug, id string) {
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

type apiDistributedTraining struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	Framework   string `json:"framework"`
	Nodes       int    `json:"nodes"`
	GPUsPerNode int    `json:"gpus_per_node"`
	GPUType     string `json:"gpu_type"`
	Dataset     string `json:"dataset"`
	Epochs      int    `json:"epochs"`
	Progress    string `json:"progress"`
	Runtime     string `json:"runtime"`
	Cost        string `json:"cost"`
	Region      string `json:"region"`
}

type apiDistributedTrainingList struct {
	Jobs  []apiDistributedTraining `json:"jobs"`
	Total int                      `json:"total"`
}

func toAPIDistributedTraining(j DistributedTraining) apiDistributedTraining {
	return apiDistributedTraining{
		ID: j.ID, Name: j.Name, Status: j.Status,
		Framework: j.Framework, Nodes: j.Nodes,
		GPUsPerNode: j.GPUsPerNode, GPUType: j.GPUType,
		Dataset: j.Dataset, Epochs: j.Epochs,
		Progress: j.Progress, Runtime: j.Runtime,
		Cost: j.Cost, Region: j.Region,
	}
}

func filterDistributedTraining(all []DistributedTraining, search, status string) []DistributedTraining {
	var out []DistributedTraining
	q := strings.ToLower(search)
	for _, j := range all {
		if status != "" && j.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(j.Name + " " + j.Framework + " " + j.GPUType + " " + j.Dataset + " " + j.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, j)
	}
	return out
}

func apiWSAITrainDistListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	jobs := filterDistributedTraining(distributedTrainingStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiDistributedTraining, len(jobs))
	for i, j := range jobs {
		out[i] = toAPIDistributedTraining(j)
	}
	apiOK(w, apiDistributedTrainingList{Jobs: out, Total: len(out)})
}

type apiCreateDistributedTrainingRequest struct {
	Name        string `json:"name"`
	Framework   string `json:"framework"`
	Nodes       int    `json:"nodes"`
	GPUsPerNode int    `json:"gpus_per_node"`
	GPUType     string `json:"gpu_type"`
	Dataset     string `json:"dataset"`
	Epochs      int    `json:"epochs"`
	Region      string `json:"region"`
}

func apiWSAITrainDistCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateDistributedTrainingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	framework := strings.TrimSpace(req.Framework)
	if framework == "" {
		framework = "PyTorch"
	}
	gpuType := strings.TrimSpace(req.GPUType)
	if gpuType == "" {
		gpuType = "A100 40GB"
	}
	dataset := strings.TrimSpace(req.Dataset)
	if dataset == "" {
		apiError(w, http.StatusBadRequest, "Dataset is required.")
		return
	}
	nodes := req.Nodes
	if nodes <= 0 {
		nodes = 2
	}
	gpusPerNode := req.GPUsPerNode
	if gpusPerNode <= 0 {
		gpusPerNode = 4
	}
	epochs := req.Epochs
	if epochs <= 0 {
		epochs = 10
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "eu-central-1"
	}
	job := DistributedTraining{
		ID: newID(), Name: name, Status: "queued",
		Framework: framework, Nodes: nodes,
		GPUsPerNode: gpusPerNode, GPUType: gpuType,
		Dataset: dataset, Epochs: epochs,
		Progress: "0%", Runtime: "—",
		Cost: "$0", Region: region,
	}
	distributedTrainingStore.add(slug, job)
	apiCreated(w, toAPIDistributedTraining(job))
}

func apiWSAITrainDistGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	job, found := distributedTrainingStore.getByID(slug, r.PathValue("jid"))
	if !found {
		apiError(w, http.StatusNotFound, "Distributed training job not found.")
		return
	}
	apiOK(w, toAPIDistributedTraining(job))
}

func apiWSAITrainDistDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := distributedTrainingStore.getByID(slug, r.PathValue("jid")); !found {
		apiError(w, http.StatusNotFound, "Distributed training job not found.")
		return
	}
	distributedTrainingStore.delete(slug, r.PathValue("jid"))
	apiOK(w, map[string]any{"ok": true, "message": "Distributed training job deleted."})
}

// ════════════════════════════════════════════════════════════════
// Workspace AI API — Training: AutoML
// ai/training/automl
// ════════════════════════════════════════════════════════════════

type AutoML struct {
	ID          string
	Name        string
	Status      string
	TaskType    string
	Dataset     string
	Target      string
	Trials      int
	BestModel   string
	BestScore   string
	Metric      string
	Runtime     string
	Cost        string
	Region      string
}

type autoMLStoreType struct {
	mu   sync.RWMutex
	runs map[string][]AutoML
}

var autoMLStore = &autoMLStoreType{
	runs: map[string][]AutoML{
		"buyvm": {
			{ID: newID(), Name: "churn-prediction", Status: "completed", TaskType: "classification", Dataset: "customer-events (2.4 GB)", Target: "churned", Trials: 200, BestModel: "XGBoost", BestScore: "0.942 AUC", Metric: "AUC", Runtime: "4h 22m", Cost: "$86", Region: "eu-central-1"},
			{ID: newID(), Name: "revenue-forecast", Status: "running", TaskType: "regression", Dataset: "sales-history (890 MB)", Target: "monthly_revenue", Trials: 150, BestModel: "LightGBM", BestScore: "0.89 R²", Metric: "R²", Runtime: "2h 10m", Cost: "$42", Region: "eu-central-1"},
			{ID: newID(), Name: "image-classifier-automl", Status: "completed", TaskType: "image-classification", Dataset: "product-images (12 GB)", Target: "category", Trials: 50, BestModel: "EfficientNet-B4", BestScore: "96.2% acc", Metric: "accuracy", Runtime: "8h 45m", Cost: "$340", Region: "us-east-1"},
			{ID: newID(), Name: "anomaly-detector", Status: "queued", TaskType: "anomaly-detection", Dataset: "server-metrics (560 MB)", Target: "is_anomaly", Trials: 100, BestModel: "—", BestScore: "—", Metric: "F1", Runtime: "—", Cost: "$0", Region: "eu-central-1"},
		},
		"hosting": {
			{ID: newID(), Name: "lead-scoring", Status: "completed", TaskType: "classification", Dataset: "crm-leads (1.1 GB)", Target: "converted", Trials: 180, BestModel: "CatBoost", BestScore: "0.918 AUC", Metric: "AUC", Runtime: "3h 48m", Cost: "$72", Region: "ap-south-1"},
			{ID: newID(), Name: "demand-forecast", Status: "running", TaskType: "time-series", Dataset: "order-history (2.8 GB)", Target: "daily_orders", Trials: 120, BestModel: "Prophet + LGBM", BestScore: "12.4 MAPE", Metric: "MAPE", Runtime: "1h 55m", Cost: "$38", Region: "ap-south-1"},
			{ID: newID(), Name: "text-classifier", Status: "completed", TaskType: "text-classification", Dataset: "support-tickets (340 MB)", Target: "category", Trials: 80, BestModel: "DistilBERT fine-tuned", BestScore: "93.8% acc", Metric: "accuracy", Runtime: "6h 12m", Cost: "$248", Region: "eu-west-1"},
		},
	},
}

func (s *autoMLStoreType) list(slug string) []AutoML {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.runs[slug]
	out := make([]AutoML, len(src))
	copy(out, src)
	return out
}

func (s *autoMLStoreType) getByID(slug, id string) (AutoML, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, r := range s.runs[slug] {
		if r.ID == id {
			return r, true
		}
	}
	return AutoML{}, false
}

func (s *autoMLStoreType) add(slug string, r AutoML) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.runs[slug] = append(s.runs[slug], r)
}

func (s *autoMLStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.runs[slug]
	for i, r := range list {
		if r.ID == id {
			s.runs[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiAutoML struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	TaskType  string `json:"task_type"`
	Dataset   string `json:"dataset"`
	Target    string `json:"target"`
	Trials    int    `json:"trials"`
	BestModel string `json:"best_model"`
	BestScore string `json:"best_score"`
	Metric    string `json:"metric"`
	Runtime   string `json:"runtime"`
	Cost      string `json:"cost"`
	Region    string `json:"region"`
}

type apiAutoMLList struct {
	Runs  []apiAutoML `json:"runs"`
	Total int         `json:"total"`
}

func toAPIAutoML(r AutoML) apiAutoML {
	return apiAutoML{
		ID: r.ID, Name: r.Name, Status: r.Status,
		TaskType: r.TaskType, Dataset: r.Dataset,
		Target: r.Target, Trials: r.Trials,
		BestModel: r.BestModel, BestScore: r.BestScore,
		Metric: r.Metric, Runtime: r.Runtime,
		Cost: r.Cost, Region: r.Region,
	}
}

func filterAutoML(all []AutoML, search, status string) []AutoML {
	var out []AutoML
	q := strings.ToLower(search)
	for _, r := range all {
		if status != "" && r.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(r.Name + " " + r.TaskType + " " + r.BestModel + " " + r.Dataset + " " + r.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, r)
	}
	return out
}

func apiWSAITrainAutoMLListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	runs := filterAutoML(autoMLStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiAutoML, len(runs))
	for i, run := range runs {
		out[i] = toAPIAutoML(run)
	}
	apiOK(w, apiAutoMLList{Runs: out, Total: len(out)})
}

type apiCreateAutoMLRequest struct {
	Name     string `json:"name"`
	TaskType string `json:"task_type"`
	Dataset  string `json:"dataset"`
	Target   string `json:"target"`
	Trials   int    `json:"trials"`
	Metric   string `json:"metric"`
	Region   string `json:"region"`
}

func apiWSAITrainAutoMLCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateAutoMLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	taskType := strings.TrimSpace(req.TaskType)
	if taskType == "" {
		taskType = "classification"
	}
	dataset := strings.TrimSpace(req.Dataset)
	if dataset == "" {
		apiError(w, http.StatusBadRequest, "Dataset is required.")
		return
	}
	target := strings.TrimSpace(req.Target)
	if target == "" {
		apiError(w, http.StatusBadRequest, "Target column is required.")
		return
	}
	trials := req.Trials
	if trials <= 0 {
		trials = 100
	}
	metric := strings.TrimSpace(req.Metric)
	if metric == "" {
		switch taskType {
		case "regression":
			metric = "R²"
		case "time-series":
			metric = "MAPE"
		default:
			metric = "AUC"
		}
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "eu-central-1"
	}
	run := AutoML{
		ID: newID(), Name: name, Status: "queued",
		TaskType: taskType, Dataset: dataset,
		Target: target, Trials: trials,
		BestModel: "—", BestScore: "—",
		Metric: metric, Runtime: "—",
		Cost: "$0", Region: region,
	}
	autoMLStore.add(slug, run)
	apiCreated(w, toAPIAutoML(run))
}

func apiWSAITrainAutoMLGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	run, found := autoMLStore.getByID(slug, r.PathValue("rid"))
	if !found {
		apiError(w, http.StatusNotFound, "AutoML run not found.")
		return
	}
	apiOK(w, toAPIAutoML(run))
}

func apiWSAITrainAutoMLDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := autoMLStore.getByID(slug, r.PathValue("rid")); !found {
		apiError(w, http.StatusNotFound, "AutoML run not found.")
		return
	}
	autoMLStore.delete(slug, r.PathValue("rid"))
	apiOK(w, map[string]any{"ok": true, "message": "AutoML run deleted."})
}
