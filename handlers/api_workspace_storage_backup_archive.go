package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Storage API — Backup: Long-term Archiving
// storage/backup/archive
// ════════════════════════════════════════════════════════════════

type ArchiveEntry struct {
	ID          string
	Name        string
	Region      string
	Status      string
	SourceType  string
	SourceName  string
	Size        string
	ArchivedAt  string
	ExpiresAt   string
	StorageClass string
	Encrypted   bool
}

type archiveStoreType struct {
	mu       sync.RWMutex
	archives map[string][]ArchiveEntry
}

var archiveStore = &archiveStoreType{
	archives: map[string][]ArchiveEntry{
		"buyvm": {
			{ID: newID(), Name: "db-yearly-2025", Region: "eu-central-1", Status: "stored", SourceType: "snapshot", SourceName: "db-primary-annual-2025", Size: "1.2 TB", ArchivedAt: "2026-01-01T00:00:00Z", ExpiresAt: "2033-01-01T00:00:00Z", StorageClass: "glacier-deep", Encrypted: true},
			{ID: newID(), Name: "compliance-logs-q4-2025", Region: "eu-central-1", Status: "stored", SourceType: "object-store", SourceName: "audit-logs-bucket", Size: "340 GB", ArchivedAt: "2026-01-15T00:00:00Z", ExpiresAt: "2033-01-15T00:00:00Z", StorageClass: "glacier-deep", Encrypted: true},
			{ID: newID(), Name: "media-archive-2025", Region: "us-east-1", Status: "stored", SourceType: "block-hdd", SourceName: "media-ingest", Size: "14.8 TB", ArchivedAt: "2026-02-01T00:00:00Z", ExpiresAt: "2031-02-01T00:00:00Z", StorageClass: "glacier", Encrypted: true},
			{ID: newID(), Name: "app-config-backup-q1", Region: "eu-central-1", Status: "restoring", SourceType: "snapshot", SourceName: "app-config-snap-q1", Size: "8.5 GB", ArchivedAt: "2026-04-01T00:00:00Z", ExpiresAt: "2029-04-01T00:00:00Z", StorageClass: "glacier", Encrypted: false},
			{ID: newID(), Name: "legal-docs-2024", Region: "eu-central-1", Status: "stored", SourceType: "object-store", SourceName: "legal-documents", Size: "52 GB", ArchivedAt: "2025-06-01T00:00:00Z", ExpiresAt: "2035-06-01T00:00:00Z", StorageClass: "glacier-deep", Encrypted: true},
		},
		"hosting": {
			{ID: newID(), Name: "mongo-yearly-2025", Region: "ap-south-1", Status: "stored", SourceType: "snapshot", SourceName: "mongo-data-annual", Size: "680 GB", ArchivedAt: "2026-01-02T00:00:00Z", ExpiresAt: "2033-01-02T00:00:00Z", StorageClass: "glacier-deep", Encrypted: true},
			{ID: newID(), Name: "user-uploads-archive-2025", Region: "ap-south-1", Status: "stored", SourceType: "object-store", SourceName: "user-uploads-bucket", Size: "4.2 TB", ArchivedAt: "2026-01-10T00:00:00Z", ExpiresAt: "2031-01-10T00:00:00Z", StorageClass: "glacier", Encrypted: true},
			{ID: newID(), Name: "analytics-raw-2024", Region: "ap-south-1", Status: "expired", SourceType: "object-store", SourceName: "analytics-raw-bucket", Size: "2.1 TB", ArchivedAt: "2024-12-01T00:00:00Z", ExpiresAt: "2025-12-01T00:00:00Z", StorageClass: "glacier", Encrypted: false},
		},
	},
}

func (s *archiveStoreType) list(slug string) []ArchiveEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.archives[slug]
	out := make([]ArchiveEntry, len(src))
	copy(out, src)
	return out
}

func (s *archiveStoreType) getByID(slug, id string) (ArchiveEntry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, a := range s.archives[slug] {
		if a.ID == id {
			return a, true
		}
	}
	return ArchiveEntry{}, false
}

func (s *archiveStoreType) add(slug string, a ArchiveEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.archives[slug] = append(s.archives[slug], a)
}

func (s *archiveStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.archives[slug]
	for i, a := range list {
		if a.ID == id {
			s.archives[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiArchiveEntry struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Region       string `json:"region"`
	Status       string `json:"status"`
	SourceType   string `json:"source_type"`
	SourceName   string `json:"source_name"`
	Size         string `json:"size"`
	ArchivedAt   string `json:"archived_at"`
	ExpiresAt    string `json:"expires_at"`
	StorageClass string `json:"storage_class"`
	Encrypted    bool   `json:"encrypted"`
}

type apiArchiveEntryList struct {
	Archives []apiArchiveEntry `json:"archives"`
	Total    int               `json:"total"`
}

func toAPIArchiveEntry(a ArchiveEntry) apiArchiveEntry {
	return apiArchiveEntry{
		ID: a.ID, Name: a.Name, Region: a.Region,
		Status: a.Status, SourceType: a.SourceType,
		SourceName: a.SourceName, Size: a.Size,
		ArchivedAt: a.ArchivedAt, ExpiresAt: a.ExpiresAt,
		StorageClass: a.StorageClass, Encrypted: a.Encrypted,
	}
}

func filterArchiveEntries(all []ArchiveEntry, search, status string) []ArchiveEntry {
	var out []ArchiveEntry
	q := strings.ToLower(search)
	for _, a := range all {
		if status != "" && a.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(a.Name + " " + a.Region + " " + a.SourceType + " " + a.SourceName + " " + a.StorageClass)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, a)
	}
	return out
}

func apiWSStorageBackupArchiveListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	archives := filterArchiveEntries(archiveStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiArchiveEntry, len(archives))
	for i, a := range archives {
		out[i] = toAPIArchiveEntry(a)
	}
	apiOK(w, apiArchiveEntryList{Archives: out, Total: len(out)})
}

type apiCreateArchiveEntryRequest struct {
	Name         string `json:"name"`
	Region       string `json:"region"`
	SourceType   string `json:"source_type"`
	SourceName   string `json:"source_name"`
	Retention    string `json:"retention"`
	StorageClass string `json:"storage_class"`
	Encrypted    bool   `json:"encrypted"`
}

func apiWSStorageBackupArchiveCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateArchiveEntryRequest
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
		sourceType = "snapshot"
	}
	sourceName := strings.TrimSpace(req.SourceName)
	if sourceName == "" {
		apiError(w, http.StatusBadRequest, "Source name is required.")
		return
	}
	storageClass := strings.TrimSpace(req.StorageClass)
	if storageClass == "" {
		storageClass = "glacier"
	}
	entry := ArchiveEntry{
		ID: newID(), Name: name, Region: region,
		Status: "archiving", SourceType: sourceType,
		SourceName: sourceName, Size: "—",
		ArchivedAt: "2026-04-17T00:00:00Z", ExpiresAt: "2033-04-17T00:00:00Z",
		StorageClass: storageClass, Encrypted: req.Encrypted,
	}
	archiveStore.add(slug, entry)
	apiCreated(w, toAPIArchiveEntry(entry))
}

func apiWSStorageBackupArchiveGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	entry, found := archiveStore.getByID(slug, r.PathValue("aid"))
	if !found {
		apiError(w, http.StatusNotFound, "Archive entry not found.")
		return
	}
	apiOK(w, toAPIArchiveEntry(entry))
}

func apiWSStorageBackupArchiveDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := archiveStore.getByID(slug, r.PathValue("aid")); !found {
		apiError(w, http.StatusNotFound, "Archive entry not found.")
		return
	}
	archiveStore.delete(slug, r.PathValue("aid"))
	apiOK(w, map[string]any{"ok": true, "message": "Archive entry deleted."})
}
