package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Storage API — Volumes: Block Storage (Pro Feature)
// storage/volumes/block-storage
// ════════════════════════════════════════════════════════════════

type ProBlockVolume struct {
	ID           string
	Name         string
	SizeGB       int
	UsedGB       int
	VolumeType   string
	IOPS         int
	ThroughputMB int
	AttachedTo   string
	MountPoint   string
	Encrypted    bool
	SnapshotID   string
	Region       string
	AZ           string
	CostPerMonth float64
	Status       string
	Updated      string
}

type proBlockVolumeStoreType struct {
	mu    sync.RWMutex
	items map[string][]ProBlockVolume
}

var proBlockVolumeStore = &proBlockVolumeStoreType{
	items: map[string][]ProBlockVolume{},
}

func (s *proBlockVolumeStoreType) list(slug string) []ProBlockVolume {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.items[slug]
	out := make([]ProBlockVolume, len(src))
	copy(out, src)
	return out
}

func (s *proBlockVolumeStoreType) getByID(slug, id string) (ProBlockVolume, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, v := range s.items[slug] {
		if v.ID == id {
			return v, true
		}
	}
	return ProBlockVolume{}, false
}

func (s *proBlockVolumeStoreType) add(slug string, v ProBlockVolume) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[slug] = append(s.items[slug], v)
}

func (s *proBlockVolumeStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.items[slug]
	for i, v := range list {
		if v.ID == id {
			s.items[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiProBlockVolume struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	SizeGB       int     `json:"size_gb"`
	UsedGB       int     `json:"used_gb"`
	VolumeType   string  `json:"volume_type"`
	IOPS         int     `json:"iops"`
	ThroughputMB int     `json:"throughput_mb"`
	AttachedTo   string  `json:"attached_to"`
	MountPoint   string  `json:"mount_point"`
	Encrypted    bool    `json:"encrypted"`
	SnapshotID   string  `json:"snapshot_id"`
	Region       string  `json:"region"`
	AZ           string  `json:"availability_zone"`
	CostPerMonth float64 `json:"cost_per_month"`
	Status       string  `json:"status"`
	Updated      string  `json:"updated"`
}

type apiProBlockVolumeList struct {
	Volumes []apiProBlockVolume `json:"volumes"`
	Total   int                 `json:"total"`
}

func toAPIProBlockVolume(v ProBlockVolume) apiProBlockVolume {
	return apiProBlockVolume{
		ID: v.ID, Name: v.Name, SizeGB: v.SizeGB,
		UsedGB: v.UsedGB, VolumeType: v.VolumeType,
		IOPS: v.IOPS, ThroughputMB: v.ThroughputMB,
		AttachedTo: v.AttachedTo, MountPoint: v.MountPoint,
		Encrypted: v.Encrypted, SnapshotID: v.SnapshotID,
		Region: v.Region, AZ: v.AZ,
		CostPerMonth: v.CostPerMonth, Status: v.Status,
		Updated: v.Updated,
	}
}

func filterProBlockVolumes(all []ProBlockVolume, search, status string) []ProBlockVolume {
	var out []ProBlockVolume
	q := strings.ToLower(search)
	for _, v := range all {
		if status != "" && v.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(v.Name + " " + v.VolumeType + " " + v.AttachedTo + " " + v.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, v)
	}
	return out
}

func apiWSStorageBlockVolListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	items := filterProBlockVolumes(proBlockVolumeStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiProBlockVolume, len(items))
	for i, v := range items {
		out[i] = toAPIProBlockVolume(v)
	}
	apiOK(w, apiProBlockVolumeList{Volumes: out, Total: len(out)})
}

type apiCreateProBlockVolumeRequest struct {
	Name       string `json:"name"`
	SizeGB     int    `json:"size_gb"`
	VolumeType string `json:"volume_type"`
	IOPS       int    `json:"iops"`
	Encrypted  bool   `json:"encrypted"`
	Region     string `json:"region"`
	AZ         string `json:"availability_zone"`
}

func apiWSStorageBlockVolCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateProBlockVolumeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	sizeGB := req.SizeGB
	if sizeGB <= 0 {
		sizeGB = 100
	}
	volType := strings.TrimSpace(req.VolumeType)
	if volType == "" {
		volType = "SSD (gp3)"
	}
	iops := req.IOPS
	if iops <= 0 {
		iops = 3000
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "us-east-1"
	}
	az := strings.TrimSpace(req.AZ)
	if az == "" {
		az = region + "a"
	}
	vol := ProBlockVolume{
		ID: newID(), Name: name, SizeGB: sizeGB, UsedGB: 0,
		VolumeType: volType, IOPS: iops, ThroughputMB: 125,
		AttachedTo: "—", MountPoint: "—",
		Encrypted: req.Encrypted, SnapshotID: "—",
		Region: region, AZ: az,
		CostPerMonth: float64(sizeGB) * 0.08,
		Status:       "available", Updated: "2026-04-18",
	}
	proBlockVolumeStore.add(slug, vol)
	apiCreated(w, toAPIProBlockVolume(vol))
}

func apiWSStorageBlockVolGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	v, found := proBlockVolumeStore.getByID(slug, r.PathValue("vid"))
	if !found {
		apiError(w, http.StatusNotFound, "Block volume not found.")
		return
	}
	apiOK(w, toAPIProBlockVolume(v))
}

func apiWSStorageBlockVolDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := proBlockVolumeStore.getByID(slug, r.PathValue("vid")); !found {
		apiError(w, http.StatusNotFound, "Block volume not found.")
		return
	}
	proBlockVolumeStore.delete(slug, r.PathValue("vid"))
	apiOK(w, map[string]any{"ok": true, "message": "Block volume deleted."})
}
