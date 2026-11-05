package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Application Services API — Media: Video Encoding
// apps/media/encoding
// ════════════════════════════════════════════════════════════════

type EncodingJob struct {
	ID         string
	Name       string
	InputFile  string
	OutputFmt  string
	Resolution string
	Bitrate    string
	Progress   string
	Duration   string
	Status     string
	Updated    string
}

type encodingJobStoreType struct {
	mu   sync.RWMutex
	jobs map[string][]EncodingJob
}

var encodingJobStore = &encodingJobStoreType{
	jobs: map[string][]EncodingJob{},
}

func (s *encodingJobStoreType) list(slug string) []EncodingJob {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.jobs[slug]
	out := make([]EncodingJob, len(src))
	copy(out, src)
	return out
}

func (s *encodingJobStoreType) getByID(slug, id string) (EncodingJob, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, j := range s.jobs[slug] {
		if j.ID == id {
			return j, true
		}
	}
	return EncodingJob{}, false
}

func (s *encodingJobStoreType) add(slug string, j EncodingJob) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[slug] = append(s.jobs[slug], j)
}

func (s *encodingJobStoreType) delete(slug, id string) {
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

type apiEncodingJob struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	InputFile  string `json:"input_file"`
	OutputFmt  string `json:"output_format"`
	Resolution string `json:"resolution"`
	Bitrate    string `json:"bitrate"`
	Progress   string `json:"progress"`
	Duration   string `json:"duration"`
	Status     string `json:"status"`
	Updated    string `json:"updated"`
}

type apiEncodingJobList struct {
	Jobs  []apiEncodingJob `json:"jobs"`
	Total int              `json:"total"`
}

func toAPIEncodingJob(j EncodingJob) apiEncodingJob {
	return apiEncodingJob{
		ID: j.ID, Name: j.Name, InputFile: j.InputFile,
		OutputFmt: j.OutputFmt, Resolution: j.Resolution, Bitrate: j.Bitrate,
		Progress: j.Progress, Duration: j.Duration,
		Status: j.Status, Updated: j.Updated,
	}
}

func filterEncodingJobs(all []EncodingJob, search, status string) []EncodingJob {
	var out []EncodingJob
	q := strings.ToLower(search)
	for _, j := range all {
		if status != "" && j.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(j.Name + " " + j.InputFile + " " + j.OutputFmt)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, j)
	}
	return out
}

func apiWSAppsMediaEncodingListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	jobs := filterEncodingJobs(encodingJobStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiEncodingJob, len(jobs))
	for i, j := range jobs {
		out[i] = toAPIEncodingJob(j)
	}
	apiOK(w, apiEncodingJobList{Jobs: out, Total: len(out)})
}

type apiCreateEncodingJobRequest struct {
	Name       string `json:"name"`
	InputFile  string `json:"input_file"`
	OutputFmt  string `json:"output_format"`
	Resolution string `json:"resolution"`
	Bitrate    string `json:"bitrate"`
}

func apiWSAppsMediaEncodingCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateEncodingJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	inputFile := strings.TrimSpace(req.InputFile)
	if inputFile == "" {
		apiError(w, http.StatusBadRequest, "Input file is required.")
		return
	}
	outputFmt := strings.TrimSpace(req.OutputFmt)
	if outputFmt == "" {
		outputFmt = "H.264 / MP4"
	}
	resolution := strings.TrimSpace(req.Resolution)
	if resolution == "" {
		resolution = "1920x1080"
	}
	bitrate := strings.TrimSpace(req.Bitrate)
	if bitrate == "" {
		bitrate = "6 Mbps"
	}
	job := EncodingJob{
		ID: newID(), Name: name, InputFile: inputFile,
		OutputFmt: outputFmt, Resolution: resolution, Bitrate: bitrate,
		Progress: "0%", Duration: "—",
		Status: "queued", Updated: "2026-04-17",
	}
	encodingJobStore.add(slug, job)
	apiCreated(w, toAPIEncodingJob(job))
}

func apiWSAppsMediaEncodingGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	job, found := encodingJobStore.getByID(slug, r.PathValue("eid"))
	if !found {
		apiError(w, http.StatusNotFound, "Encoding job not found.")
		return
	}
	apiOK(w, toAPIEncodingJob(job))
}

func apiWSAppsMediaEncodingDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := encodingJobStore.getByID(slug, r.PathValue("eid")); !found {
		apiError(w, http.StatusNotFound, "Encoding job not found.")
		return
	}
	encodingJobStore.delete(slug, r.PathValue("eid"))
	apiOK(w, map[string]any{"ok": true, "message": "Encoding job deleted."})
}
