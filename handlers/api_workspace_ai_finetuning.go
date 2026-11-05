package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace AI API — Training: Fine-tuning
// ai/training/finetuning
// ════════════════════════════════════════════════════════════════

type FineTuningJob struct {
	ID           string
	Name         string
	Status       string
	BaseModel    string
	Method       string
	Dataset      string
	Epochs       int
	LearningRate string
	Progress     string
	Runtime      string
	Cost         string
	Region       string
	OutputModel  string
}

type fineTuningJobStoreType struct {
	mu   sync.RWMutex
	jobs map[string][]FineTuningJob
}

var fineTuningJobStore = &fineTuningJobStoreType{
	jobs: map[string][]FineTuningJob{},
}

func (s *fineTuningJobStoreType) list(slug string) []FineTuningJob {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.jobs[slug]
	out := make([]FineTuningJob, len(src))
	copy(out, src)
	return out
}

func (s *fineTuningJobStoreType) getByID(slug, id string) (FineTuningJob, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, j := range s.jobs[slug] {
		if j.ID == id {
			return j, true
		}
	}
	return FineTuningJob{}, false
}

func (s *fineTuningJobStoreType) add(slug string, j FineTuningJob) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[slug] = append(s.jobs[slug], j)
}

func (s *fineTuningJobStoreType) delete(slug, id string) {
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

type apiFineTuningJob struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Status       string `json:"status"`
	BaseModel    string `json:"base_model"`
	Method       string `json:"method"`
	Dataset      string `json:"dataset"`
	Epochs       int    `json:"epochs"`
	LearningRate string `json:"learning_rate"`
	Progress     string `json:"progress"`
	Runtime      string `json:"runtime"`
	Cost         string `json:"cost"`
	Region       string `json:"region"`
	OutputModel  string `json:"output_model"`
}

type apiFineTuningJobList struct {
	Jobs  []apiFineTuningJob `json:"jobs"`
	Total int                `json:"total"`
}

func toAPIFineTuningJob(j FineTuningJob) apiFineTuningJob {
	return apiFineTuningJob{
		ID: j.ID, Name: j.Name, Status: j.Status,
		BaseModel: j.BaseModel, Method: j.Method,
		Dataset: j.Dataset, Epochs: j.Epochs,
		LearningRate: j.LearningRate, Progress: j.Progress,
		Runtime: j.Runtime, Cost: j.Cost,
		Region: j.Region, OutputModel: j.OutputModel,
	}
}

func filterFineTuningJobs(all []FineTuningJob, search, status string) []FineTuningJob {
	var out []FineTuningJob
	q := strings.ToLower(search)
	for _, j := range all {
		if status != "" && j.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(j.Name + " " + j.BaseModel + " " + j.Method + " " + j.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, j)
	}
	return out
}

func apiWSAITrainFineTuneListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	jobs := filterFineTuningJobs(fineTuningJobStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiFineTuningJob, len(jobs))
	for i, j := range jobs {
		out[i] = toAPIFineTuningJob(j)
	}
	apiOK(w, apiFineTuningJobList{Jobs: out, Total: len(out)})
}

type apiCreateFineTuningJobRequest struct {
	Name         string `json:"name"`
	BaseModel    string `json:"base_model"`
	Method       string `json:"method"`
	Dataset      string `json:"dataset"`
	Epochs       int    `json:"epochs"`
	LearningRate string `json:"learning_rate"`
	Region       string `json:"region"`
}

func apiWSAITrainFineTuneCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateFineTuningJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	baseModel := strings.TrimSpace(req.BaseModel)
	if baseModel == "" {
		baseModel = "llama-3-8b"
	}
	method := strings.TrimSpace(req.Method)
	if method == "" {
		method = "LoRA"
	}
	dataset := strings.TrimSpace(req.Dataset)
	if dataset == "" {
		apiError(w, http.StatusBadRequest, "Dataset is required.")
		return
	}
	epochs := req.Epochs
	if epochs <= 0 {
		epochs = 3
	}
	lr := strings.TrimSpace(req.LearningRate)
	if lr == "" {
		lr = "2e-4"
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "eu-central-1"
	}
	job := FineTuningJob{
		ID: newID(), Name: name, Status: "queued",
		BaseModel: baseModel, Method: method,
		Dataset: dataset, Epochs: epochs,
		LearningRate: lr, Progress: "0%",
		Runtime: "—", Cost: "$0",
		Region: region, OutputModel: "",
	}
	fineTuningJobStore.add(slug, job)
	apiCreated(w, toAPIFineTuningJob(job))
}

func apiWSAITrainFineTuneGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	job, found := fineTuningJobStore.getByID(slug, r.PathValue("ftid"))
	if !found {
		apiError(w, http.StatusNotFound, "Fine-tuning job not found.")
		return
	}
	apiOK(w, toAPIFineTuningJob(job))
}

func apiWSAITrainFineTuneDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := fineTuningJobStore.getByID(slug, r.PathValue("ftid")); !found {
		apiError(w, http.StatusNotFound, "Fine-tuning job not found.")
		return
	}
	fineTuningJobStore.delete(slug, r.PathValue("ftid"))
	apiOK(w, map[string]any{"ok": true, "message": "Fine-tuning job deleted."})
}
