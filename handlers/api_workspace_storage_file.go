package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Storage API — File Storage
// ════════════════════════════════════════════════════════════════

// ═══════════════════════════════════════════════════════════════
// 1. File Storage — NFS / SMB Shares   storage/file/nfs
// ═══════════════════════════════════════════════════════════════

type NFSShare struct {
	ID         string
	Name       string
	Region     string
	Status     string
	Protocol   string
	Size       string
	Used       string
	MountPoint string
	AccessCIDR string
}

type nfsShareStoreType struct {
	mu     sync.RWMutex
	shares map[string][]NFSShare
}

var nfsShareStore = &nfsShareStoreType{
	shares: map[string][]NFSShare{
		"buyvm": {
			{ID: newID(), Name: "home-dirs", Region: "eu-central-1", Status: "active", Protocol: "NFSv4", Size: "2 TB", Used: "1.4 TB", MountPoint: "/mnt/home", AccessCIDR: "10.0.0.0/16"},
			{ID: newID(), Name: "shared-assets", Region: "eu-central-1", Status: "active", Protocol: "SMB3", Size: "500 GB", Used: "312 GB", MountPoint: "\\\\fs01\\assets", AccessCIDR: "10.0.0.0/16"},
			{ID: newID(), Name: "build-artifacts", Region: "us-east-1", Status: "active", Protocol: "NFSv4", Size: "1 TB", Used: "680 GB", MountPoint: "/mnt/builds", AccessCIDR: "10.1.0.0/16"},
			{ID: newID(), Name: "media-render-farm", Region: "eu-central-1", Status: "active", Protocol: "NFSv4", Size: "10 TB", Used: "7.8 TB", MountPoint: "/mnt/renders", AccessCIDR: "10.0.0.0/8"},
			{ID: newID(), Name: "legacy-smb-share", Region: "eu-central-1", Status: "degraded", Protocol: "SMB2", Size: "200 GB", Used: "198 GB", MountPoint: "\\\\fs02\\legacy", AccessCIDR: "10.0.1.0/24"},
		},
		"hosting": {
			{ID: newID(), Name: "dev-workspace", Region: "ap-south-1", Status: "active", Protocol: "NFSv4", Size: "500 GB", Used: "120 GB", MountPoint: "/mnt/devws", AccessCIDR: "172.16.0.0/12"},
			{ID: newID(), Name: "windows-profiles", Region: "ap-south-1", Status: "active", Protocol: "SMB3", Size: "200 GB", Used: "85 GB", MountPoint: "\\\\fs01\\profiles", AccessCIDR: "172.16.0.0/12"},
			{ID: newID(), Name: "data-staging", Region: "ap-south-1", Status: "active", Protocol: "NFSv4", Size: "1 TB", Used: "340 GB", MountPoint: "/mnt/staging", AccessCIDR: "172.16.0.0/12"},
		},
	},
}

func (s *nfsShareStoreType) list(slug string) []NFSShare {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.shares[slug]
	out := make([]NFSShare, len(src))
	copy(out, src)
	return out
}

func (s *nfsShareStoreType) getByID(slug, id string) (NFSShare, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, sh := range s.shares[slug] {
		if sh.ID == id {
			return sh, true
		}
	}
	return NFSShare{}, false
}

func (s *nfsShareStoreType) add(slug string, sh NFSShare) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.shares[slug] = append(s.shares[slug], sh)
}

func (s *nfsShareStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.shares[slug]
	for i, sh := range list {
		if sh.ID == id {
			s.shares[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiNFSShare struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Region     string `json:"region"`
	Status     string `json:"status"`
	Protocol   string `json:"protocol"`
	Size       string `json:"size"`
	Used       string `json:"used"`
	MountPoint string `json:"mount_point"`
	AccessCIDR string `json:"access_cidr"`
}

type apiNFSShareList struct {
	Shares []apiNFSShare `json:"shares"`
	Total  int           `json:"total"`
}

func toAPINFSShare(sh NFSShare) apiNFSShare {
	return apiNFSShare{
		ID: sh.ID, Name: sh.Name, Region: sh.Region,
		Status: sh.Status, Protocol: sh.Protocol, Size: sh.Size,
		Used: sh.Used, MountPoint: sh.MountPoint, AccessCIDR: sh.AccessCIDR,
	}
}

func filterNFSShares(all []NFSShare, search, status string) []NFSShare {
	var out []NFSShare
	q := strings.ToLower(search)
	for _, sh := range all {
		if status != "" && sh.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(sh.Name + " " + sh.Region + " " + sh.Protocol + " " + sh.MountPoint)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, sh)
	}
	return out
}

func apiWSStorageFileNFSListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	shares := filterNFSShares(nfsShareStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiNFSShare, len(shares))
	for i, sh := range shares {
		out[i] = toAPINFSShare(sh)
	}
	apiOK(w, apiNFSShareList{Shares: out, Total: len(out)})
}

type apiCreateNFSShareRequest struct {
	Name       string `json:"name"`
	Region     string `json:"region"`
	Protocol   string `json:"protocol"`
	Size       string `json:"size"`
	AccessCIDR string `json:"access_cidr"`
}

func apiWSStorageFileNFSCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateNFSShareRequest
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
	protocol := strings.TrimSpace(req.Protocol)
	if protocol == "" {
		protocol = "NFSv4"
	}
	size := strings.TrimSpace(req.Size)
	if size == "" {
		size = "500 GB"
	}
	cidr := strings.TrimSpace(req.AccessCIDR)
	if cidr == "" {
		cidr = "10.0.0.0/16"
	}
	mountPoint := "/mnt/" + strings.ReplaceAll(strings.ToLower(name), " ", "-")
	if protocol == "SMB2" || protocol == "SMB3" {
		mountPoint = "\\\\fs01\\" + strings.ReplaceAll(strings.ToLower(name), " ", "-")
	}
	share := NFSShare{
		ID: newID(), Name: name, Region: region,
		Status: "active", Protocol: protocol, Size: size,
		Used: "0 B", MountPoint: mountPoint, AccessCIDR: cidr,
	}
	nfsShareStore.add(slug, share)
	apiCreated(w, toAPINFSShare(share))
}

func apiWSStorageFileNFSGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	share, found := nfsShareStore.getByID(slug, r.PathValue("sid"))
	if !found {
		apiError(w, http.StatusNotFound, "NFS share not found.")
		return
	}
	apiOK(w, toAPINFSShare(share))
}

func apiWSStorageFileNFSDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := nfsShareStore.getByID(slug, r.PathValue("sid")); !found {
		apiError(w, http.StatusNotFound, "NFS share not found.")
		return
	}
	nfsShareStore.delete(slug, r.PathValue("sid"))
	apiOK(w, map[string]any{"ok": true, "message": "NFS share deleted."})
}

// ═══════════════════════════════════════════════════════════════
// 2. File Storage — Distributed File Systems   storage/file/distributed
// ═══════════════════════════════════════════════════════════════

type DistributedFS struct {
	ID         string
	Name       string
	Region     string
	Status     string
	Type       string
	Capacity   string
	Used       string
	Nodes      int
	Replication int
	MountPoint string
}

type distributedFSStoreType struct {
	mu          sync.RWMutex
	filesystems map[string][]DistributedFS
}

var distributedFSStore = &distributedFSStoreType{
	filesystems: map[string][]DistributedFS{
		"buyvm": {
			{ID: newID(), Name: "hadoop-cluster-main", Region: "eu-central-1", Status: "active", Type: "HDFS", Capacity: "100 TB", Used: "72 TB", Nodes: 12, Replication: 3, MountPoint: "hdfs://namenode:8020/"},
			{ID: newID(), Name: "ceph-prod", Region: "eu-central-1", Status: "active", Type: "CephFS", Capacity: "50 TB", Used: "31 TB", Nodes: 8, Replication: 3, MountPoint: "/mnt/cephfs"},
			{ID: newID(), Name: "gluster-media", Region: "us-east-1", Status: "active", Type: "GlusterFS", Capacity: "20 TB", Used: "14.2 TB", Nodes: 6, Replication: 2, MountPoint: "/mnt/gluster/media"},
			{ID: newID(), Name: "lustre-hpc", Region: "eu-central-1", Status: "active", Type: "Lustre", Capacity: "200 TB", Used: "145 TB", Nodes: 24, Replication: 2, MountPoint: "/mnt/lustre"},
		},
		"hosting": {
			{ID: newID(), Name: "ceph-shared", Region: "ap-south-1", Status: "active", Type: "CephFS", Capacity: "10 TB", Used: "4.8 TB", Nodes: 4, Replication: 3, MountPoint: "/mnt/cephfs"},
			{ID: newID(), Name: "minio-datalake", Region: "ap-south-1", Status: "active", Type: "MinIO", Capacity: "30 TB", Used: "18 TB", Nodes: 6, Replication: 2, MountPoint: "s3://minio.internal/"},
			{ID: newID(), Name: "gluster-dev", Region: "ap-south-1", Status: "degraded", Type: "GlusterFS", Capacity: "5 TB", Used: "4.1 TB", Nodes: 3, Replication: 2, MountPoint: "/mnt/gluster/dev"},
		},
	},
}

func (s *distributedFSStoreType) list(slug string) []DistributedFS {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.filesystems[slug]
	out := make([]DistributedFS, len(src))
	copy(out, src)
	return out
}

func (s *distributedFSStoreType) getByID(slug, id string) (DistributedFS, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, fs := range s.filesystems[slug] {
		if fs.ID == id {
			return fs, true
		}
	}
	return DistributedFS{}, false
}

func (s *distributedFSStoreType) add(slug string, fs DistributedFS) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.filesystems[slug] = append(s.filesystems[slug], fs)
}

func (s *distributedFSStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.filesystems[slug]
	for i, fs := range list {
		if fs.ID == id {
			s.filesystems[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiDistributedFS struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Region      string `json:"region"`
	Status      string `json:"status"`
	Type        string `json:"type"`
	Capacity    string `json:"capacity"`
	Used        string `json:"used"`
	Nodes       int    `json:"nodes"`
	Replication int    `json:"replication"`
	MountPoint  string `json:"mount_point"`
}

type apiDistributedFSList struct {
	Filesystems []apiDistributedFS `json:"filesystems"`
	Total       int                `json:"total"`
}

func toAPIDistributedFS(fs DistributedFS) apiDistributedFS {
	return apiDistributedFS{
		ID: fs.ID, Name: fs.Name, Region: fs.Region,
		Status: fs.Status, Type: fs.Type, Capacity: fs.Capacity,
		Used: fs.Used, Nodes: fs.Nodes, Replication: fs.Replication,
		MountPoint: fs.MountPoint,
	}
}

func filterDistributedFS(all []DistributedFS, search, status string) []DistributedFS {
	var out []DistributedFS
	q := strings.ToLower(search)
	for _, fs := range all {
		if status != "" && fs.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(fs.Name + " " + fs.Region + " " + fs.Type)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, fs)
	}
	return out
}

func apiWSStorageFileDistributedListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	fsList := filterDistributedFS(distributedFSStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiDistributedFS, len(fsList))
	for i, fs := range fsList {
		out[i] = toAPIDistributedFS(fs)
	}
	apiOK(w, apiDistributedFSList{Filesystems: out, Total: len(out)})
}

type apiCreateDistributedFSRequest struct {
	Name        string `json:"name"`
	Region      string `json:"region"`
	Type        string `json:"type"`
	Capacity    string `json:"capacity"`
	Nodes       int    `json:"nodes"`
	Replication int    `json:"replication"`
}

func apiWSStorageFileDistributedCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateDistributedFSRequest
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
	fsType := strings.TrimSpace(req.Type)
	if fsType == "" {
		fsType = "CephFS"
	}
	capacity := strings.TrimSpace(req.Capacity)
	if capacity == "" {
		capacity = "10 TB"
	}
	nodes := req.Nodes
	if nodes <= 0 {
		nodes = 3
	}
	replication := req.Replication
	if replication <= 0 {
		replication = 2
	}
	mountPoint := "/mnt/" + strings.ReplaceAll(strings.ToLower(name), " ", "-")
	if fsType == "HDFS" {
		mountPoint = "hdfs://namenode:8020/" + strings.ReplaceAll(strings.ToLower(name), " ", "-")
	} else if fsType == "MinIO" {
		mountPoint = "s3://minio.internal/" + strings.ReplaceAll(strings.ToLower(name), " ", "-")
	}
	fs := DistributedFS{
		ID: newID(), Name: name, Region: region,
		Status: "provisioning", Type: fsType, Capacity: capacity,
		Used: "0 B", Nodes: nodes, Replication: replication,
		MountPoint: mountPoint,
	}
	distributedFSStore.add(slug, fs)
	apiCreated(w, toAPIDistributedFS(fs))
}

func apiWSStorageFileDistributedGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	fs, found := distributedFSStore.getByID(slug, r.PathValue("did"))
	if !found {
		apiError(w, http.StatusNotFound, "Distributed filesystem not found.")
		return
	}
	apiOK(w, toAPIDistributedFS(fs))
}

func apiWSStorageFileDistributedDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := distributedFSStore.getByID(slug, r.PathValue("did")); !found {
		apiError(w, http.StatusNotFound, "Distributed filesystem not found.")
		return
	}
	distributedFSStore.delete(slug, r.PathValue("did"))
	apiOK(w, map[string]any{"ok": true, "message": "Distributed filesystem deleted."})
}
