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
	snapshots: map[string][]Snapshot{},
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
