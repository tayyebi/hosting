package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Storage API — Backup: Snapshot Management
// storage/backup/snapshots
// ════════════════════════════════════════════════════════════════

type Snapshot struct {
	ID         string
	Name       string
	Region     string
	Status     string
	SourceType string
	SourceName string
	Size       string
	CreatedAt  string
	Retention  string
	Encrypted  bool
}

type snapshotStoreType struct {
	mu        sync.RWMutex
	snapshots map[string][]Snapshot
}

var snapshotStore = &snapshotStoreType{
	snapshots: map[string][]Snapshot{
		"buyvm": {
			{ID: newID(), Name: "db-primary-daily-0416", Region: "eu-central-1", Status: "completed", SourceType: "block-ssd", SourceName: "db-primary-data", Size: "480 GB", CreatedAt: "2026-04-16T02:00:00Z", Retention: "30 days", Encrypted: true},
			{ID: newID(), Name: "db-primary-daily-0415", Region: "eu-central-1", Status: "completed", SourceType: "block-ssd", SourceName: "db-primary-data", Size: "478 GB", CreatedAt: "2026-04-15T02:00:00Z", Retention: "30 days", Encrypted: true},
			{ID: newID(), Name: "app-server-weekly-w15", Region: "us-east-1", Status: "completed", SourceType: "vm", SourceName: "vm-app-server-03", Size: "120 GB", CreatedAt: "2026-04-13T04:00:00Z", Retention: "90 days", Encrypted: true},
			{ID: newID(), Name: "elasticsearch-pre-upgrade", Region: "eu-central-1", Status: "completed", SourceType: "block-ssd", SourceName: "elasticsearch-hot", Size: "940 GB", CreatedAt: "2026-04-10T18:30:00Z", Retention: "180 days", Encrypted: true},
			{ID: newID(), Name: "media-vol-snapshot-q1", Region: "us-east-1", Status: "completed", SourceType: "block-hdd", SourceName: "media-ingest", Size: "7.2 TB", CreatedAt: "2026-03-31T00:00:00Z", Retention: "365 days", Encrypted: true},
			{ID: newID(), Name: "staging-scratch-snap", Region: "eu-central-1", Status: "in-progress", SourceType: "block-ssd", SourceName: "staging-scratch", Size: "—", CreatedAt: "2026-04-17T10:15:00Z", Retention: "7 days", Encrypted: false},
		},
		"hosting": {
			{ID: newID(), Name: "mongo-daily-0416", Region: "ap-south-1", Status: "completed", SourceType: "block-ssd", SourceName: "mongo-data", Size: "240 GB", CreatedAt: "2026-04-16T03:00:00Z", Retention: "30 days", Encrypted: true},
			{ID: newID(), Name: "redis-aof-snap-0414", Region: "ap-south-1", Status: "completed", SourceType: "block-ssd", SourceName: "redis-aof", Size: "92 GB", CreatedAt: "2026-04-14T03:00:00Z", Retention: "14 days", Encrypted: true},
			{ID: newID(), Name: "file-share-pre-migration", Region: "ap-south-1", Status: "completed", SourceType: "block-hdd", SourceName: "file-share-data", Size: "980 GB", CreatedAt: "2026-04-12T22:00:00Z", Retention: "60 days", Encrypted: false},
			{ID: newID(), Name: "dev-scratch-snap", Region: "ap-south-1", Status: "failed", SourceType: "block-hdd", SourceName: "dev-scratch", Size: "—", CreatedAt: "2026-04-16T06:00:00Z", Retention: "7 days", Encrypted: false},
		},
	},
}

func (s *snapshotStoreType) list(slug string) []Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.snapshots[slug]
	out := make([]Snapshot, len(src))
	copy(out, src)
	return out
}

func (s *snapshotStoreType) getByID(slug, id string) (Snapshot, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, sn := range s.snapshots[slug] {
		if sn.ID == id {
			return sn, true
		}
	}
	return Snapshot{}, false
}

func (s *snapshotStoreType) add(slug string, sn Snapshot) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.snapshots[slug] = append(s.snapshots[slug], sn)
}

func (s *snapshotStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.snapshots[slug]
	for i, sn := range list {
		if sn.ID == id {
			s.snapshots[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiSnapshot struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Region     string `json:"region"`
	Status     string `json:"status"`
	SourceType string `json:"source_type"`
	SourceName string `json:"source_name"`
	Size       string `json:"size"`
	CreatedAt  string `json:"created_at"`
	Retention  string `json:"retention"`
	Encrypted  bool   `json:"encrypted"`
}

type apiSnapshotList struct {
	Snapshots []apiSnapshot `json:"snapshots"`
	Total     int           `json:"total"`
}

func toAPISnapshot(sn Snapshot) apiSnapshot {
	return apiSnapshot{
		ID: sn.ID, Name: sn.Name, Region: sn.Region,
		Status: sn.Status, SourceType: sn.SourceType,
		SourceName: sn.SourceName, Size: sn.Size,
		CreatedAt: sn.CreatedAt, Retention: sn.Retention,
		Encrypted: sn.Encrypted,
	}
}

func filterSnapshots(all []Snapshot, search, status string) []Snapshot {
	var out []Snapshot
	q := strings.ToLower(search)
	for _, sn := range all {
		if status != "" && sn.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(sn.Name + " " + sn.Region + " " + sn.SourceType + " " + sn.SourceName)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, sn)
	}
	return out
}

func apiWSStorageBackupSnapshotsListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	snaps := filterSnapshots(snapshotStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiSnapshot, len(snaps))
	for i, sn := range snaps {
		out[i] = toAPISnapshot(sn)
	}
	apiOK(w, apiSnapshotList{Snapshots: out, Total: len(out)})
}

type apiCreateSnapshotRequest struct {
	Name       string `json:"name"`
	Region     string `json:"region"`
	SourceType string `json:"source_type"`
	SourceName string `json:"source_name"`
	Retention  string `json:"retention"`
	Encrypted  bool   `json:"encrypted"`
}

func apiWSStorageBackupSnapshotsCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateSnapshotRequest
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
	sourceType := strings.TrimSpace(req.SourceType)
	if sourceType == "" {
		sourceType = "block-ssd"
	}
	sourceName := strings.TrimSpace(req.SourceName)
	if sourceName == "" {
		apiError(w, http.StatusBadRequest, "Source name is required.")
		return
	}
	retention := strings.TrimSpace(req.Retention)
	if retention == "" {
		retention = "30 days"
	}
	snap := Snapshot{
		ID: newID(), Name: name, Region: region,
		Status: "in-progress", SourceType: sourceType,
		SourceName: sourceName, Size: "—",
		CreatedAt: "2026-04-17T00:00:00Z", Retention: retention,
		Encrypted: req.Encrypted,
	}
	snapshotStore.add(slug, snap)
	apiCreated(w, toAPISnapshot(snap))
}

func apiWSStorageBackupSnapshotsGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	snap, found := snapshotStore.getByID(slug, r.PathValue("snid"))
	if !found {
		apiError(w, http.StatusNotFound, "Snapshot not found.")
		return
	}
	apiOK(w, toAPISnapshot(snap))
}

func apiWSStorageBackupSnapshotsDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := snapshotStore.getByID(slug, r.PathValue("snid")); !found {
		apiError(w, http.StatusNotFound, "Snapshot not found.")
		return
	}
	snapshotStore.delete(slug, r.PathValue("snid"))
	apiOK(w, map[string]any{"ok": true, "message": "Snapshot deleted."})
}
