package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Storage API — Object Storage: S3-Compatible
// storage/object-storage/s3-compatible
// ════════════════════════════════════════════════════════════════

type S3Bucket struct {
	ID            string
	Name          string
	Region        string
	StorageClass  string
	SizeGB        float64
	ObjectCount   int64
	Versioning    bool
	Encryption    string
	AccessPolicy  string
	Endpoint      string
	CDNEnabled    bool
	LifecycleRule string
	Status        string
	Updated       string
}

type s3BucketStoreType struct {
	mu    sync.RWMutex
	items map[string][]S3Bucket
}

var s3BucketStore = &s3BucketStoreType{
	items: map[string][]S3Bucket{},
}

func (s *s3BucketStoreType) list(slug string) []S3Bucket {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.items[slug]
	out := make([]S3Bucket, len(src))
	copy(out, src)
	return out
}

func (s *s3BucketStoreType) getByID(slug, id string) (S3Bucket, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, b := range s.items[slug] {
		if b.ID == id {
			return b, true
		}
	}
	return S3Bucket{}, false
}

func (s *s3BucketStoreType) add(slug string, b S3Bucket) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[slug] = append(s.items[slug], b)
}

func (s *s3BucketStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.items[slug]
	for i, b := range list {
		if b.ID == id {
			s.items[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiS3Bucket struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Region        string  `json:"region"`
	StorageClass  string  `json:"storage_class"`
	SizeGB        float64 `json:"size_gb"`
	ObjectCount   int64   `json:"object_count"`
	Versioning    bool    `json:"versioning"`
	Encryption    string  `json:"encryption"`
	AccessPolicy  string  `json:"access_policy"`
	Endpoint      string  `json:"endpoint"`
	CDNEnabled    bool    `json:"cdn_enabled"`
	LifecycleRule string  `json:"lifecycle_rule"`
	Status        string  `json:"status"`
	Updated       string  `json:"updated"`
}

type apiS3BucketList struct {
	Buckets []apiS3Bucket `json:"buckets"`
	Total   int           `json:"total"`
}

func toAPIS3Bucket(b S3Bucket) apiS3Bucket {
	return apiS3Bucket{
		ID: b.ID, Name: b.Name, Region: b.Region,
		StorageClass: b.StorageClass, SizeGB: b.SizeGB,
		ObjectCount: b.ObjectCount, Versioning: b.Versioning,
		Encryption: b.Encryption, AccessPolicy: b.AccessPolicy,
		Endpoint: b.Endpoint, CDNEnabled: b.CDNEnabled,
		LifecycleRule: b.LifecycleRule, Status: b.Status,
		Updated: b.Updated,
	}
}

func filterS3Buckets(all []S3Bucket, search, status string) []S3Bucket {
	var out []S3Bucket
	q := strings.ToLower(search)
	for _, b := range all {
		if status != "" && b.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(b.Name + " " + b.Region + " " + b.StorageClass + " " + b.AccessPolicy)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, b)
	}
	return out
}

func apiWSStorageObjectS3ListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	items := filterS3Buckets(s3BucketStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiS3Bucket, len(items))
	for i, b := range items {
		out[i] = toAPIS3Bucket(b)
	}
	apiOK(w, apiS3BucketList{Buckets: out, Total: len(out)})
}

type apiCreateS3BucketRequest struct {
	Name         string `json:"name"`
	Region       string `json:"region"`
	StorageClass string `json:"storage_class"`
	Versioning   bool   `json:"versioning"`
	Encryption   string `json:"encryption"`
	AccessPolicy string `json:"access_policy"`
}

func apiWSStorageObjectS3CreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateS3BucketRequest
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
		region = "us-east-1"
	}
	storageClass := strings.TrimSpace(req.StorageClass)
	if storageClass == "" {
		storageClass = "Standard"
	}
	encryption := strings.TrimSpace(req.Encryption)
	if encryption == "" {
		encryption = "AES-256"
	}
	accessPolicy := strings.TrimSpace(req.AccessPolicy)
	if accessPolicy == "" {
		accessPolicy = "Private"
	}
	bucket := S3Bucket{
		ID: newID(), Name: name, Region: region,
		StorageClass: storageClass, SizeGB: 0, ObjectCount: 0,
		Versioning: req.Versioning, Encryption: encryption,
		AccessPolicy: accessPolicy,
		Endpoint:     "s3." + region + ".cloud",
		CDNEnabled:   false, LifecycleRule: "None",
		Status: "active", Updated: "2026-04-17",
	}
	s3BucketStore.add(slug, bucket)
	apiCreated(w, toAPIS3Bucket(bucket))
}

func apiWSStorageObjectS3GetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	bucket, found := s3BucketStore.getByID(slug, r.PathValue("bid"))
	if !found {
		apiError(w, http.StatusNotFound, "S3 bucket not found.")
		return
	}
	apiOK(w, toAPIS3Bucket(bucket))
}

func apiWSStorageObjectS3DeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := s3BucketStore.getByID(slug, r.PathValue("bid")); !found {
		apiError(w, http.StatusNotFound, "S3 bucket not found.")
		return
	}
	s3BucketStore.delete(slug, r.PathValue("bid"))
	apiOK(w, map[string]any{"ok": true, "message": "S3 bucket deleted."})
}
