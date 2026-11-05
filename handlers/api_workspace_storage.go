package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Storage API — mock REST endpoints for workspace leaf
// pages under the Storage category.
// ════════════════════════════════════════════════════════════════

// ═══════════════════════════════════════════════════════════════
// 1. Object Storage — Standard Tier   storage/object/standard
// ═══════════════════════════════════════════════════════════════

type ObjectBucket struct {
	ID         string
	Name       string
	Region     string
	Status     string
	Size       string
	Objects    string
	Versioning string
	Access     string
}

type objectBucketStoreType struct {
	mu      sync.RWMutex
	buckets map[string][]ObjectBucket
}

var objectBucketStore = &objectBucketStoreType{
	buckets: map[string][]ObjectBucket{},
}

func (s *objectBucketStoreType) list(slug string) []ObjectBucket {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.buckets[slug]
	out := make([]ObjectBucket, len(src))
	copy(out, src)
	return out
}

func (s *objectBucketStoreType) getByID(slug, id string) (ObjectBucket, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, b := range s.buckets[slug] {
		if b.ID == id {
			return b, true
		}
	}
	return ObjectBucket{}, false
}

func (s *objectBucketStoreType) add(slug string, b ObjectBucket) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.buckets[slug] = append(s.buckets[slug], b)
}

func (s *objectBucketStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.buckets[slug]
	for i, b := range list {
		if b.ID == id {
			s.buckets[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiObjectBucket struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Region     string `json:"region"`
	Status     string `json:"status"`
	Size       string `json:"size"`
	Objects    string `json:"objects"`
	Versioning string `json:"versioning"`
	Access     string `json:"access"`
}

type apiObjectBucketList struct {
	Buckets []apiObjectBucket `json:"buckets"`
	Total   int               `json:"total"`
}

func toAPIObjectBucket(b ObjectBucket) apiObjectBucket {
	return apiObjectBucket{
		ID: b.ID, Name: b.Name, Region: b.Region,
		Status: b.Status, Size: b.Size, Objects: b.Objects,
		Versioning: b.Versioning, Access: b.Access,
	}
}

func filterObjectBuckets(all []ObjectBucket, search, status string) []ObjectBucket {
	var out []ObjectBucket
	q := strings.ToLower(search)
	for _, b := range all {
		if status != "" && b.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(b.Name + " " + b.Region + " " + b.Access)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, b)
	}
	return out
}

func apiWSStorageObjectStandardListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	buckets := filterObjectBuckets(objectBucketStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiObjectBucket, len(buckets))
	for i, b := range buckets {
		out[i] = toAPIObjectBucket(b)
	}
	apiOK(w, apiObjectBucketList{Buckets: out, Total: len(out)})
}

type apiCreateObjectBucketRequest struct {
	Name       string `json:"name"`
	Region     string `json:"region"`
	Versioning string `json:"versioning"`
	Access     string `json:"access"`
}

func apiWSStorageObjectStandardCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateObjectBucketRequest
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
	versioning := strings.TrimSpace(req.Versioning)
	if versioning == "" {
		versioning = "disabled"
	}
	access := strings.TrimSpace(req.Access)
	if access == "" {
		access = "private"
	}
	bucket := ObjectBucket{
		ID: newID(), Name: name, Region: region,
		Status: "active", Size: "0 B", Objects: "0",
		Versioning: versioning, Access: access,
	}
	objectBucketStore.add(slug, bucket)
	apiCreated(w, toAPIObjectBucket(bucket))
}

func apiWSStorageObjectStandardGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	bucket, found := objectBucketStore.getByID(slug, r.PathValue("bid"))
	if !found {
		apiError(w, http.StatusNotFound, "Bucket not found.")
		return
	}
	apiOK(w, toAPIObjectBucket(bucket))
}

func apiWSStorageObjectStandardDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := objectBucketStore.getByID(slug, r.PathValue("bid")); !found {
		apiError(w, http.StatusNotFound, "Bucket not found.")
		return
	}
	objectBucketStore.delete(slug, r.PathValue("bid"))
	apiOK(w, map[string]any{"ok": true, "message": "Bucket deleted."})
}

// ═══════════════════════════════════════════════════════════════
// 2. Object Storage — Archive / Cold Tier   storage/object/archive
// ═══════════════════════════════════════════════════════════════

type ArchiveVault struct {
	ID           string
	Name         string
	Region       string
	Status       string
	Size         string
	Archives     string
	Tier         string
	RetrievalSLA string
}

type archiveVaultStoreType struct {
	mu     sync.RWMutex
	vaults map[string][]ArchiveVault
}

var archiveVaultStore = &archiveVaultStoreType{
	vaults: map[string][]ArchiveVault{},
}

func (s *archiveVaultStoreType) list(slug string) []ArchiveVault {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.vaults[slug]
	out := make([]ArchiveVault, len(src))
	copy(out, src)
	return out
}

func (s *archiveVaultStoreType) getByID(slug, id string) (ArchiveVault, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, v := range s.vaults[slug] {
		if v.ID == id {
			return v, true
		}
	}
	return ArchiveVault{}, false
}

func (s *archiveVaultStoreType) add(slug string, v ArchiveVault) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.vaults[slug] = append(s.vaults[slug], v)
}

func (s *archiveVaultStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.vaults[slug]
	for i, v := range list {
		if v.ID == id {
			s.vaults[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiArchiveVault struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Region       string `json:"region"`
	Status       string `json:"status"`
	Size         string `json:"size"`
	Archives     string `json:"archives"`
	Tier         string `json:"tier"`
	RetrievalSLA string `json:"retrieval_sla"`
}

type apiArchiveVaultList struct {
	Vaults []apiArchiveVault `json:"vaults"`
	Total  int               `json:"total"`
}

func toAPIArchiveVault(v ArchiveVault) apiArchiveVault {
	return apiArchiveVault{
		ID: v.ID, Name: v.Name, Region: v.Region,
		Status: v.Status, Size: v.Size, Archives: v.Archives,
		Tier: v.Tier, RetrievalSLA: v.RetrievalSLA,
	}
}

func filterArchiveVaults(all []ArchiveVault, search, status string) []ArchiveVault {
	var out []ArchiveVault
	q := strings.ToLower(search)
	for _, v := range all {
		if status != "" && v.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(v.Name + " " + v.Region + " " + v.Tier)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, v)
	}
	return out
}

func apiWSStorageObjectArchiveListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	vaults := filterArchiveVaults(archiveVaultStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiArchiveVault, len(vaults))
	for i, v := range vaults {
		out[i] = toAPIArchiveVault(v)
	}
	apiOK(w, apiArchiveVaultList{Vaults: out, Total: len(out)})
}

type apiCreateArchiveVaultRequest struct {
	Name   string `json:"name"`
	Region string `json:"region"`
	Tier   string `json:"tier"`
}

func apiWSStorageObjectArchiveCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateArchiveVaultRequest
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
	tier := strings.TrimSpace(req.Tier)
	if tier == "" {
		tier = "glacier"
	}
	sla := "3-5 hours"
	switch tier {
	case "deep-archive":
		sla = "12 hours"
	case "glacier-ir":
		sla = "minutes"
	}
	vault := ArchiveVault{
		ID: newID(), Name: name, Region: region,
		Status: "active", Size: "0 B", Archives: "0",
		Tier: tier, RetrievalSLA: sla,
	}
	archiveVaultStore.add(slug, vault)
	apiCreated(w, toAPIArchiveVault(vault))
}

func apiWSStorageObjectArchiveGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	vault, found := archiveVaultStore.getByID(slug, r.PathValue("vid"))
	if !found {
		apiError(w, http.StatusNotFound, "Archive vault not found.")
		return
	}
	apiOK(w, toAPIArchiveVault(vault))
}

func apiWSStorageObjectArchiveDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := archiveVaultStore.getByID(slug, r.PathValue("vid")); !found {
		apiError(w, http.StatusNotFound, "Archive vault not found.")
		return
	}
	archiveVaultStore.delete(slug, r.PathValue("vid"))
	apiOK(w, map[string]any{"ok": true, "message": "Archive vault deleted."})
}

// ═══════════════════════════════════════════════════════════════
// 3. Object Storage — Data Lake   storage/object/datalake
// ═══════════════════════════════════════════════════════════════

type DataLakeZone struct {
	ID       string
	Name     string
	Region   string
	Status   string
	Size     string
	Tables   int
	Format   string
	Catalog  string
	LastSync string
}

type dataLakeZoneStoreType struct {
	mu    sync.RWMutex
	zones map[string][]DataLakeZone
}

var dataLakeZoneStore = &dataLakeZoneStoreType{
	zones: map[string][]DataLakeZone{},
}

func (s *dataLakeZoneStoreType) list(slug string) []DataLakeZone {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.zones[slug]
	out := make([]DataLakeZone, len(src))
	copy(out, src)
	return out
}

func (s *dataLakeZoneStoreType) getByID(slug, id string) (DataLakeZone, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, z := range s.zones[slug] {
		if z.ID == id {
			return z, true
		}
	}
	return DataLakeZone{}, false
}

func (s *dataLakeZoneStoreType) add(slug string, z DataLakeZone) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.zones[slug] = append(s.zones[slug], z)
}

func (s *dataLakeZoneStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.zones[slug]
	for i, z := range list {
		if z.ID == id {
			s.zones[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiDataLakeZone struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Region   string `json:"region"`
	Status   string `json:"status"`
	Size     string `json:"size"`
	Tables   int    `json:"tables"`
	Format   string `json:"format"`
	Catalog  string `json:"catalog"`
	LastSync string `json:"last_sync"`
}

type apiDataLakeZoneList struct {
	Zones []apiDataLakeZone `json:"zones"`
	Total int               `json:"total"`
}

func toAPIDataLakeZone(z DataLakeZone) apiDataLakeZone {
	return apiDataLakeZone{
		ID: z.ID, Name: z.Name, Region: z.Region,
		Status: z.Status, Size: z.Size, Tables: z.Tables,
		Format: z.Format, Catalog: z.Catalog, LastSync: z.LastSync,
	}
}

func filterDataLakeZones(all []DataLakeZone, search, status string) []DataLakeZone {
	var out []DataLakeZone
	q := strings.ToLower(search)
	for _, z := range all {
		if status != "" && z.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(z.Name + " " + z.Region + " " + z.Format + " " + z.Catalog)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, z)
	}
	return out
}

func apiWSStorageObjectDatalakeListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	zones := filterDataLakeZones(dataLakeZoneStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiDataLakeZone, len(zones))
	for i, z := range zones {
		out[i] = toAPIDataLakeZone(z)
	}
	apiOK(w, apiDataLakeZoneList{Zones: out, Total: len(out)})
}

type apiCreateDataLakeZoneRequest struct {
	Name    string `json:"name"`
	Region  string `json:"region"`
	Format  string `json:"format"`
	Catalog string `json:"catalog"`
}

func apiWSStorageObjectDatalakeCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateDataLakeZoneRequest
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
	format := strings.TrimSpace(req.Format)
	if format == "" {
		format = "Parquet"
	}
	catalog := strings.TrimSpace(req.Catalog)
	if catalog == "" {
		catalog = "Glue"
	}
	zone := DataLakeZone{
		ID: newID(), Name: name, Region: region,
		Status: "active", Size: "0 B", Tables: 0,
		Format: format, Catalog: catalog, LastSync: "never",
	}
	dataLakeZoneStore.add(slug, zone)
	apiCreated(w, toAPIDataLakeZone(zone))
}

func apiWSStorageObjectDatalakeGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	zone, found := dataLakeZoneStore.getByID(slug, r.PathValue("zid"))
	if !found {
		apiError(w, http.StatusNotFound, "Data lake zone not found.")
		return
	}
	apiOK(w, toAPIDataLakeZone(zone))
}

func apiWSStorageObjectDatalakeDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := dataLakeZoneStore.getByID(slug, r.PathValue("zid")); !found {
		apiError(w, http.StatusNotFound, "Data lake zone not found.")
		return
	}
	dataLakeZoneStore.delete(slug, r.PathValue("zid"))
	apiOK(w, map[string]any{"ok": true, "message": "Data lake zone deleted."})
}
