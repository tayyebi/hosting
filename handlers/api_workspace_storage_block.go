package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Storage API — Block Storage
// ════════════════════════════════════════════════════════════════

// ═══════════════════════════════════════════════════════════════
// 1. Block Storage — SSD High Performance   storage/block/ssd
// ═══════════════════════════════════════════════════════════════

type BlockVolumeSSD struct {
	ID         string
	Name       string
	Region     string
	Status     string
	Size       string
	IOPS       string
	Throughput string
	AttachedTo string
	Encrypted  bool
}

type blockVolumeSSDStoreType struct {
	mu      sync.RWMutex
	volumes map[string][]BlockVolumeSSD
}

var blockVolumeSSDStore = &blockVolumeSSDStoreType{
	volumes: map[string][]BlockVolumeSSD{},
}

func (s *blockVolumeSSDStoreType) list(slug string) []BlockVolumeSSD {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.volumes[slug]
	out := make([]BlockVolumeSSD, len(src))
	copy(out, src)
	return out
}

func (s *blockVolumeSSDStoreType) getByID(slug, id string) (BlockVolumeSSD, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, v := range s.volumes[slug] {
		if v.ID == id {
			return v, true
		}
	}
	return BlockVolumeSSD{}, false
}

func (s *blockVolumeSSDStoreType) add(slug string, v BlockVolumeSSD) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.volumes[slug] = append(s.volumes[slug], v)
}

func (s *blockVolumeSSDStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.volumes[slug]
	for i, v := range list {
		if v.ID == id {
			s.volumes[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiBlockVolumeSSD struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Region     string `json:"region"`
	Status     string `json:"status"`
	Size       string `json:"size"`
	IOPS       string `json:"iops"`
	Throughput string `json:"throughput"`
	AttachedTo string `json:"attached_to"`
	Encrypted  bool   `json:"encrypted"`
}

type apiBlockVolumeSSDList struct {
	Volumes []apiBlockVolumeSSD `json:"volumes"`
	Total   int                 `json:"total"`
}

func toAPIBlockVolumeSSD(v BlockVolumeSSD) apiBlockVolumeSSD {
	return apiBlockVolumeSSD{
		ID: v.ID, Name: v.Name, Region: v.Region,
		Status: v.Status, Size: v.Size, IOPS: v.IOPS,
		Throughput: v.Throughput, AttachedTo: v.AttachedTo,
		Encrypted: v.Encrypted,
	}
}

func filterBlockVolumesSSD(all []BlockVolumeSSD, search, status string) []BlockVolumeSSD {
	var out []BlockVolumeSSD
	q := strings.ToLower(search)
	for _, v := range all {
		if status != "" && v.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(v.Name + " " + v.Region + " " + v.AttachedTo)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, v)
	}
	return out
}

func apiWSStorageBlockSSDListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	volumes := filterBlockVolumesSSD(blockVolumeSSDStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiBlockVolumeSSD, len(volumes))
	for i, v := range volumes {
		out[i] = toAPIBlockVolumeSSD(v)
	}
	apiOK(w, apiBlockVolumeSSDList{Volumes: out, Total: len(out)})
}

type apiCreateBlockVolumeSSDRequest struct {
	Name      string `json:"name"`
	Region    string `json:"region"`
	Size      string `json:"size"`
	IOPS      string `json:"iops"`
	Encrypted bool   `json:"encrypted"`
}

func apiWSStorageBlockSSDCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateBlockVolumeSSDRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "eu-central-1"
	}
	size := strings.TrimSpace(req.Size)
	if size == "" {
		size = "100 GB"
	}
	iops := strings.TrimSpace(req.IOPS)
	if iops == "" {
		iops = "3,000"
	}
	vol := BlockVolumeSSD{
		ID: newID(), Name: name, Region: region,
		Status: "available", Size: size, IOPS: iops,
		Throughput: "125 MB/s", AttachedTo: "",
		Encrypted: req.Encrypted,
	}
	blockVolumeSSDStore.add(slug, vol)
	apiCreated(w, toAPIBlockVolumeSSD(vol))
}

func apiWSStorageBlockSSDGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	vol, found := blockVolumeSSDStore.getByID(slug, r.PathValue("vid"))
	if !found {
		apiError(w, http.StatusNotFound, "SSD volume not found.")
		return
	}
	apiOK(w, toAPIBlockVolumeSSD(vol))
}

func apiWSStorageBlockSSDDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := blockVolumeSSDStore.getByID(slug, r.PathValue("vid")); !found {
		apiError(w, http.StatusNotFound, "SSD volume not found.")
		return
	}
	blockVolumeSSDStore.delete(slug, r.PathValue("vid"))
	apiOK(w, map[string]any{"ok": true, "message": "SSD volume deleted."})
}

// ═══════════════════════════════════════════════════════════════
// 2. Block Storage — HDD Standard   storage/block/hdd
// ═══════════════════════════════════════════════════════════════

type BlockVolumeHDD struct {
	ID         string
	Name       string
	Region     string
	Status     string
	Size       string
	IOPS       string
	Throughput string
	AttachedTo string
	Encrypted  bool
}

type blockVolumeHDDStoreType struct {
	mu      sync.RWMutex
	volumes map[string][]BlockVolumeHDD
}

var blockVolumeHDDStore = &blockVolumeHDDStoreType{
	volumes: map[string][]BlockVolumeHDD{},
}

func (s *blockVolumeHDDStoreType) list(slug string) []BlockVolumeHDD {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.volumes[slug]
	out := make([]BlockVolumeHDD, len(src))
	copy(out, src)
	return out
}

func (s *blockVolumeHDDStoreType) getByID(slug, id string) (BlockVolumeHDD, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, v := range s.volumes[slug] {
		if v.ID == id {
			return v, true
		}
	}
	return BlockVolumeHDD{}, false
}

func (s *blockVolumeHDDStoreType) add(slug string, v BlockVolumeHDD) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.volumes[slug] = append(s.volumes[slug], v)
}

func (s *blockVolumeHDDStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.volumes[slug]
	for i, v := range list {
		if v.ID == id {
			s.volumes[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiBlockVolumeHDD struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Region     string `json:"region"`
	Status     string `json:"status"`
	Size       string `json:"size"`
	IOPS       string `json:"iops"`
	Throughput string `json:"throughput"`
	AttachedTo string `json:"attached_to"`
	Encrypted  bool   `json:"encrypted"`
}

type apiBlockVolumeHDDList struct {
	Volumes []apiBlockVolumeHDD `json:"volumes"`
	Total   int                 `json:"total"`
}

func toAPIBlockVolumeHDD(v BlockVolumeHDD) apiBlockVolumeHDD {
	return apiBlockVolumeHDD{
		ID: v.ID, Name: v.Name, Region: v.Region,
		Status: v.Status, Size: v.Size, IOPS: v.IOPS,
		Throughput: v.Throughput, AttachedTo: v.AttachedTo,
		Encrypted: v.Encrypted,
	}
}

func filterBlockVolumesHDD(all []BlockVolumeHDD, search, status string) []BlockVolumeHDD {
	var out []BlockVolumeHDD
	q := strings.ToLower(search)
	for _, v := range all {
		if status != "" && v.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(v.Name + " " + v.Region + " " + v.AttachedTo)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, v)
	}
	return out
}

func apiWSStorageBlockHDDListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	volumes := filterBlockVolumesHDD(blockVolumeHDDStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiBlockVolumeHDD, len(volumes))
	for i, v := range volumes {
		out[i] = toAPIBlockVolumeHDD(v)
	}
	apiOK(w, apiBlockVolumeHDDList{Volumes: out, Total: len(out)})
}

type apiCreateBlockVolumeHDDRequest struct {
	Name      string `json:"name"`
	Region    string `json:"region"`
	Size      string `json:"size"`
	Encrypted bool   `json:"encrypted"`
}

func apiWSStorageBlockHDDCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateBlockVolumeHDDRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "eu-central-1"
	}
	size := strings.TrimSpace(req.Size)
	if size == "" {
		size = "1 TB"
	}
	vol := BlockVolumeHDD{
		ID: newID(), Name: name, Region: region,
		Status: "available", Size: size, IOPS: "500",
		Throughput: "60 MB/s", AttachedTo: "",
		Encrypted: req.Encrypted,
	}
	blockVolumeHDDStore.add(slug, vol)
	apiCreated(w, toAPIBlockVolumeHDD(vol))
}

func apiWSStorageBlockHDDGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	vol, found := blockVolumeHDDStore.getByID(slug, r.PathValue("vid"))
	if !found {
		apiError(w, http.StatusNotFound, "HDD volume not found.")
		return
	}
	apiOK(w, toAPIBlockVolumeHDD(vol))
}

func apiWSStorageBlockHDDDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := blockVolumeHDDStore.getByID(slug, r.PathValue("vid")); !found {
		apiError(w, http.StatusNotFound, "HDD volume not found.")
		return
	}
	blockVolumeHDDStore.delete(slug, r.PathValue("vid"))
	apiOK(w, map[string]any{"ok": true, "message": "HDD volume deleted."})
}
