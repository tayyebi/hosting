package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Compute API — mock REST endpoints for workspace leaf
// pages under the Compute category.
// ════════════════════════════════════════════════════════════════

// ── helpers ─────────────────────────────────────────────────────

// wsAuth validates session, tenant, and workspace; returns slug, wsID, ok.
func wsAuth(w http.ResponseWriter, r *http.Request) (string, string, bool) {
	user, ok := getSession(r)
	if !ok {
		apiError(w, http.StatusUnauthorized, "Authentication required.")
		return "", "", false
	}
	slug := r.PathValue("slug")
	if _, found := store.getBySlug(slug); !found {
		apiError(w, http.StatusNotFound, "Tenant not found.")
		return "", "", false
	}
	if user.Role != RoleSysAdmin && (user.Role != RoleTenant || user.TenantSlug != slug) {
		apiError(w, http.StatusForbidden, "Access denied.")
		return "", "", false
	}
	wsID := r.PathValue("id")
	if _, _, ok := clientStore.getWorkspace(slug, wsID); !ok {
		apiError(w, http.StatusNotFound, "Workspace not found.")
		return "", "", false
	}
	return slug, wsID, true
}

// ═══════════════════════════════════════════════════════════════
// 1. General-purpose VMs   compute/vms/general
// ═══════════════════════════════════════════════════════════════

type GeneralVM struct {
	ID       string
	Name     string
	InstID   string
	Type     string
	Region   string
	Status   string
	CPUPct   int
	MemUsed  string
	MemTotal string
}

type generalVMStoreType struct {
	mu  sync.RWMutex
	vms map[string][]GeneralVM // key: tenant slug
}

var generalVMStore = &generalVMStoreType{
	vms: map[string][]GeneralVM{},
}

func (s *generalVMStoreType) list(slug string) []GeneralVM {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.vms[slug]
	out := make([]GeneralVM, len(src))
	copy(out, src)
	return out
}

func (s *generalVMStoreType) getByID(slug, id string) (GeneralVM, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, v := range s.vms[slug] {
		if v.ID == id {
			return v, true
		}
	}
	return GeneralVM{}, false
}

func (s *generalVMStoreType) add(slug string, v GeneralVM) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.vms[slug] = append(s.vms[slug], v)
}

func (s *generalVMStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.vms[slug]
	for i, v := range list {
		if v.ID == id {
			s.vms[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

// API types

type apiGeneralVM struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	InstID   string `json:"instance_id"`
	Type     string `json:"type"`
	Region   string `json:"region"`
	Status   string `json:"status"`
	CPUPct   int    `json:"cpu_percent"`
	MemUsed  string `json:"memory_used_gb"`
	MemTotal string `json:"memory_total_gb"`
}

type apiGeneralVMList struct {
	VMs   []apiGeneralVM `json:"vms"`
	Total int            `json:"total"`
}

func toAPIGeneralVM(v GeneralVM) apiGeneralVM {
	return apiGeneralVM{
		ID: v.ID, Name: v.Name, InstID: v.InstID,
		Type: v.Type, Region: v.Region, Status: v.Status,
		CPUPct: v.CPUPct, MemUsed: v.MemUsed, MemTotal: v.MemTotal,
	}
}

func filterGeneralVMs(all []GeneralVM, search, status string) []GeneralVM {
	var out []GeneralVM
	q := strings.ToLower(search)
	for _, v := range all {
		if status != "" && v.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(v.Name + " " + v.InstID + " " + v.Type + " " + v.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, v)
	}
	return out
}

func apiWSComputeVMsGeneralListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	vms := filterGeneralVMs(generalVMStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiGeneralVM, len(vms))
	for i, v := range vms {
		out[i] = toAPIGeneralVM(v)
	}
	apiOK(w, apiGeneralVMList{VMs: out, Total: len(out)})
}

type apiCreateGeneralVMRequest struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Region string `json:"region"`
}

func apiWSComputeVMsGeneralCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateGeneralVMRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	vmType := strings.TrimSpace(req.Type)
	if vmType == "" {
		vmType = "t3.medium"
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "eu-central-1"
	}
	vm := GeneralVM{
		ID: newID(), Name: name, InstID: "i-" + newID()[:12],
		Type: vmType, Region: region, Status: "running",
		CPUPct: 0, MemUsed: "0", MemTotal: "4",
	}
	generalVMStore.add(slug, vm)
	apiCreated(w, toAPIGeneralVM(vm))
}

func apiWSComputeVMsGeneralGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	vm, found := generalVMStore.getByID(slug, r.PathValue("vmid"))
	if !found {
		apiError(w, http.StatusNotFound, "VM not found.")
		return
	}
	apiOK(w, toAPIGeneralVM(vm))
}

func apiWSComputeVMsGeneralDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := generalVMStore.getByID(slug, r.PathValue("vmid")); !found {
		apiError(w, http.StatusNotFound, "VM not found.")
		return
	}
	generalVMStore.delete(slug, r.PathValue("vmid"))
	apiOK(w, map[string]any{"ok": true, "message": "VM deleted."})
}

// ═══════════════════════════════════════════════════════════════
// 2. GPU / accelerated VMs   compute/vms/gpu
// ═══════════════════════════════════════════════════════════════

type GPUVM struct {
	ID        string
	Name      string
	InstID    string
	GPU       string
	Type      string
	Region    string
	Status    string
	GPUPct    int
	VRAMUsed  string
	VRAMTotal string
}

type gpuVMStoreType struct {
	mu  sync.RWMutex
	vms map[string][]GPUVM
}

var gpuVMStore = &gpuVMStoreType{
	vms: map[string][]GPUVM{},
}

func (s *gpuVMStoreType) list(slug string) []GPUVM {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.vms[slug]
	out := make([]GPUVM, len(src))
	copy(out, src)
	return out
}

func (s *gpuVMStoreType) getByID(slug, id string) (GPUVM, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, v := range s.vms[slug] {
		if v.ID == id {
			return v, true
		}
	}
	return GPUVM{}, false
}

func (s *gpuVMStoreType) add(slug string, v GPUVM) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.vms[slug] = append(s.vms[slug], v)
}

func (s *gpuVMStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.vms[slug]
	for i, v := range list {
		if v.ID == id {
			s.vms[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiGPUVM struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	InstID    string `json:"instance_id"`
	GPU       string `json:"gpu"`
	Type      string `json:"type"`
	Region    string `json:"region"`
	Status    string `json:"status"`
	GPUPct    int    `json:"gpu_utilization_percent"`
	VRAMUsed  string `json:"vram_used_gb"`
	VRAMTotal string `json:"vram_total_gb"`
}

type apiGPUVMList struct {
	VMs   []apiGPUVM `json:"vms"`
	Total int        `json:"total"`
}

func toAPIGPUVM(v GPUVM) apiGPUVM {
	return apiGPUVM{
		ID: v.ID, Name: v.Name, InstID: v.InstID,
		GPU: v.GPU, Type: v.Type, Region: v.Region,
		Status: v.Status, GPUPct: v.GPUPct,
		VRAMUsed: v.VRAMUsed, VRAMTotal: v.VRAMTotal,
	}
}

func filterGPUVMs(all []GPUVM, search, status string) []GPUVM {
	var out []GPUVM
	q := strings.ToLower(search)
	for _, v := range all {
		if status != "" && v.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(v.Name + " " + v.GPU + " " + v.Type + " " + v.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, v)
	}
	return out
}

func apiWSComputeVMsGPUListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	vms := filterGPUVMs(gpuVMStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiGPUVM, len(vms))
	for i, v := range vms {
		out[i] = toAPIGPUVM(v)
	}
	apiOK(w, apiGPUVMList{VMs: out, Total: len(out)})
}

type apiCreateGPUVMRequest struct {
	Name   string `json:"name"`
	GPU    string `json:"gpu"`
	Type   string `json:"type"`
	Region string `json:"region"`
}

func apiWSComputeVMsGPUCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateGPUVMRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	gpu := strings.TrimSpace(req.GPU)
	if gpu == "" {
		gpu = "1x T4 16GB"
	}
	vmType := strings.TrimSpace(req.Type)
	if vmType == "" {
		vmType = "g4.large"
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "us-east-1"
	}
	vm := GPUVM{
		ID: newID(), Name: name, InstID: "i-" + newID()[:12],
		GPU: gpu, Type: vmType, Region: region,
		Status: "running", GPUPct: 0, VRAMUsed: "0", VRAMTotal: "16",
	}
	gpuVMStore.add(slug, vm)
	apiCreated(w, toAPIGPUVM(vm))
}

func apiWSComputeVMsGPUGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	vm, found := gpuVMStore.getByID(slug, r.PathValue("vmid"))
	if !found {
		apiError(w, http.StatusNotFound, "GPU VM not found.")
		return
	}
	apiOK(w, toAPIGPUVM(vm))
}

func apiWSComputeVMsGPUDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := gpuVMStore.getByID(slug, r.PathValue("vmid")); !found {
		apiError(w, http.StatusNotFound, "GPU VM not found.")
		return
	}
	gpuVMStore.delete(slug, r.PathValue("vmid"))
	apiOK(w, map[string]any{"ok": true, "message": "GPU VM deleted."})
}

// ═══════════════════════════════════════════════════════════════
// 3. Spot / preemptible VMs   compute/vms/spot
// ═══════════════════════════════════════════════════════════════

type SpotVM struct {
	ID      string
	Name    string
	InstID  string
	Pool    string
	Bid     string
	Region  string
	Status  string
	UpSince string
}

type spotVMStoreType struct {
	mu  sync.RWMutex
	vms map[string][]SpotVM
}

var spotVMStore = &spotVMStoreType{
	vms: map[string][]SpotVM{},
}

func (s *spotVMStoreType) list(slug string) []SpotVM {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.vms[slug]
	out := make([]SpotVM, len(src))
	copy(out, src)
	return out
}

func (s *spotVMStoreType) getByID(slug, id string) (SpotVM, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, v := range s.vms[slug] {
		if v.ID == id {
			return v, true
		}
	}
	return SpotVM{}, false
}

func (s *spotVMStoreType) add(slug string, v SpotVM) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.vms[slug] = append(s.vms[slug], v)
}

func (s *spotVMStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.vms[slug]
	for i, v := range list {
		if v.ID == id {
			s.vms[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiSpotVM struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	InstID  string `json:"instance_id"`
	Pool    string `json:"pool"`
	Bid     string `json:"bid"`
	Region  string `json:"region"`
	Status  string `json:"status"`
	UpSince string `json:"up_since"`
}

type apiSpotVMList struct {
	VMs   []apiSpotVM `json:"vms"`
	Total int         `json:"total"`
}

func toAPISpotVM(v SpotVM) apiSpotVM {
	return apiSpotVM{
		ID: v.ID, Name: v.Name, InstID: v.InstID,
		Pool: v.Pool, Bid: v.Bid, Region: v.Region,
		Status: v.Status, UpSince: v.UpSince,
	}
}

func filterSpotVMs(all []SpotVM, search, status string) []SpotVM {
	var out []SpotVM
	q := strings.ToLower(search)
	for _, v := range all {
		if status != "" && v.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(v.Name + " " + v.Pool + " " + v.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, v)
	}
	return out
}

func apiWSComputeVMsSpotListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	vms := filterSpotVMs(spotVMStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiSpotVM, len(vms))
	for i, v := range vms {
		out[i] = toAPISpotVM(v)
	}
	apiOK(w, apiSpotVMList{VMs: out, Total: len(out)})
}

type apiCreateSpotVMRequest struct {
	Name   string `json:"name"`
	Pool   string `json:"pool"`
	Bid    string `json:"bid"`
	Region string `json:"region"`
}

func apiWSComputeVMsSpotCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateSpotVMRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	pool := strings.TrimSpace(req.Pool)
	if pool == "" {
		pool = "c5.large pool"
	}
	bid := strings.TrimSpace(req.Bid)
	if bid == "" {
		bid = "$0.018"
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "us-east-1a"
	}
	vm := SpotVM{
		ID: newID(), Name: name, InstID: "i-" + newID()[:12],
		Pool: pool, Bid: bid, Region: region,
		Status: "running", UpSince: "0m",
	}
	spotVMStore.add(slug, vm)
	apiCreated(w, toAPISpotVM(vm))
}

func apiWSComputeVMsSpotGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	vm, found := spotVMStore.getByID(slug, r.PathValue("vmid"))
	if !found {
		apiError(w, http.StatusNotFound, "Spot VM not found.")
		return
	}
	apiOK(w, toAPISpotVM(vm))
}

func apiWSComputeVMsSpotDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := spotVMStore.getByID(slug, r.PathValue("vmid")); !found {
		apiError(w, http.StatusNotFound, "Spot VM not found.")
		return
	}
	spotVMStore.delete(slug, r.PathValue("vmid"))
	apiOK(w, map[string]any{"ok": true, "message": "Spot VM deleted."})
}

// ═══════════════════════════════════════════════════════════════
// 4. Managed Kubernetes   compute/containers/kubernetes
// ═══════════════════════════════════════════════════════════════

type K8sCluster struct {
	ID      string
	Name    string
	ClustID string
	Version string
	Nodes   string
	Region  string
	Status  string
	Pods    string
}

type k8sClusterStoreType struct {
	mu       sync.RWMutex
	clusters map[string][]K8sCluster
}

var k8sClusterStore = &k8sClusterStoreType{
	clusters: map[string][]K8sCluster{},
}

func (s *k8sClusterStoreType) list(slug string) []K8sCluster {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.clusters[slug]
	out := make([]K8sCluster, len(src))
	copy(out, src)
	return out
}

func (s *k8sClusterStoreType) getByID(slug, id string) (K8sCluster, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, c := range s.clusters[slug] {
		if c.ID == id {
			return c, true
		}
	}
	return K8sCluster{}, false
}

func (s *k8sClusterStoreType) add(slug string, c K8sCluster) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clusters[slug] = append(s.clusters[slug], c)
}

func (s *k8sClusterStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.clusters[slug]
	for i, c := range list {
		if c.ID == id {
			s.clusters[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiK8sCluster struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	ClustID string `json:"cluster_id"`
	Version string `json:"version"`
	Nodes   string `json:"nodes"`
	Region  string `json:"region"`
	Status  string `json:"status"`
	Pods    string `json:"pods"`
}

type apiK8sClusterList struct {
	Clusters []apiK8sCluster `json:"clusters"`
	Total    int             `json:"total"`
}

func toAPIK8sCluster(c K8sCluster) apiK8sCluster {
	return apiK8sCluster{
		ID: c.ID, Name: c.Name, ClustID: c.ClustID,
		Version: c.Version, Nodes: c.Nodes, Region: c.Region,
		Status: c.Status, Pods: c.Pods,
	}
}

func filterK8sClusters(all []K8sCluster, search, status string) []K8sCluster {
	var out []K8sCluster
	q := strings.ToLower(search)
	for _, c := range all {
		if status != "" && c.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(c.Name + " " + c.ClustID + " " + c.Version + " " + c.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, c)
	}
	return out
}

func apiWSComputeK8sListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	clusters := filterK8sClusters(k8sClusterStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiK8sCluster, len(clusters))
	for i, c := range clusters {
		out[i] = toAPIK8sCluster(c)
	}
	apiOK(w, apiK8sClusterList{Clusters: out, Total: len(out)})
}

type apiCreateK8sClusterRequest struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Nodes   string `json:"nodes"`
	Region  string `json:"region"`
}

func apiWSComputeK8sCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateK8sClusterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	version := strings.TrimSpace(req.Version)
	if version == "" {
		version = "v1.29.4"
	}
	nodes := strings.TrimSpace(req.Nodes)
	if nodes == "" {
		nodes = "3"
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "eu-central-1"
	}
	cluster := K8sCluster{
		ID: newID(), Name: name, ClustID: "eks-" + newID()[:6],
		Version: version, Nodes: nodes, Region: region,
		Status: "provisioning", Pods: "0",
	}
	k8sClusterStore.add(slug, cluster)
	apiCreated(w, toAPIK8sCluster(cluster))
}

func apiWSComputeK8sGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	c, found := k8sClusterStore.getByID(slug, r.PathValue("cid"))
	if !found {
		apiError(w, http.StatusNotFound, "Kubernetes cluster not found.")
		return
	}
	apiOK(w, toAPIK8sCluster(c))
}

func apiWSComputeK8sDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := k8sClusterStore.getByID(slug, r.PathValue("cid")); !found {
		apiError(w, http.StatusNotFound, "Kubernetes cluster not found.")
		return
	}
	k8sClusterStore.delete(slug, r.PathValue("cid"))
	apiOK(w, map[string]any{"ok": true, "message": "Kubernetes cluster deleted."})
}

// ═══════════════════════════════════════════════════════════════
// 5. Container Registry   compute/containers/registry
// ═══════════════════════════════════════════════════════════════

type ContainerRepo struct {
	ID       string
	Name     string
	Tags     int
	Size     string
	LastScan string
	Vulns    string
	Status   string
}

type containerRepoStoreType struct {
	mu    sync.RWMutex
	repos map[string][]ContainerRepo
}

var containerRepoStore = &containerRepoStoreType{
	repos: map[string][]ContainerRepo{},
}

func (s *containerRepoStoreType) list(slug string) []ContainerRepo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.repos[slug]
	out := make([]ContainerRepo, len(src))
	copy(out, src)
	return out
}

func (s *containerRepoStoreType) getByID(slug, id string) (ContainerRepo, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, r := range s.repos[slug] {
		if r.ID == id {
			return r, true
		}
	}
	return ContainerRepo{}, false
}

func (s *containerRepoStoreType) add(slug string, repo ContainerRepo) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.repos[slug] = append(s.repos[slug], repo)
}

func (s *containerRepoStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.repos[slug]
	for i, repo := range list {
		if repo.ID == id {
			s.repos[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiContainerRepo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Tags     int    `json:"tags"`
	Size     string `json:"size"`
	LastScan string `json:"last_scan"`
	Vulns    string `json:"vulnerabilities"`
	Status   string `json:"status"`
}

type apiContainerRepoList struct {
	Repositories []apiContainerRepo `json:"repositories"`
	Total        int                `json:"total"`
}

func toAPIContainerRepo(repo ContainerRepo) apiContainerRepo {
	return apiContainerRepo{
		ID: repo.ID, Name: repo.Name, Tags: repo.Tags,
		Size: repo.Size, LastScan: repo.LastScan,
		Vulns: repo.Vulns, Status: repo.Status,
	}
}

func filterContainerRepos(all []ContainerRepo, search, status string) []ContainerRepo {
	var out []ContainerRepo
	q := strings.ToLower(search)
	for _, repo := range all {
		if status != "" && repo.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(repo.Name + " " + repo.Vulns)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, repo)
	}
	return out
}

func apiWSComputeRegistryListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	repos := filterContainerRepos(containerRepoStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiContainerRepo, len(repos))
	for i, repo := range repos {
		out[i] = toAPIContainerRepo(repo)
	}
	apiOK(w, apiContainerRepoList{Repositories: out, Total: len(out)})
}

type apiCreateContainerRepoRequest struct {
	Name string `json:"name"`
}

func apiWSComputeRegistryCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateContainerRepoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Repository name is required.")
		return
	}
	repo := ContainerRepo{
		ID: newID(), Name: name, Tags: 0,
		Size: "0 MB", LastScan: "never",
		Vulns: "not scanned", Status: "clean",
	}
	containerRepoStore.add(slug, repo)
	apiCreated(w, toAPIContainerRepo(repo))
}

func apiWSComputeRegistryGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	repo, found := containerRepoStore.getByID(slug, r.PathValue("rid"))
	if !found {
		apiError(w, http.StatusNotFound, "Repository not found.")
		return
	}
	apiOK(w, toAPIContainerRepo(repo))
}

func apiWSComputeRegistryDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := containerRepoStore.getByID(slug, r.PathValue("rid")); !found {
		apiError(w, http.StatusNotFound, "Repository not found.")
		return
	}
	containerRepoStore.delete(slug, r.PathValue("rid"))
	apiOK(w, map[string]any{"ok": true, "message": "Repository deleted."})
}

// ═══════════════════════════════════════════════════════════════
// 6. Container Instances   compute/containers/instances
// ═══════════════════════════════════════════════════════════════

type ContainerInstance struct {
	ID     string
	Name   string
	Image  string
	CPUs   string
	Memory string
	Region string
	Status string
	IP     string
}

type containerInstanceStoreType struct {
	mu        sync.RWMutex
	instances map[string][]ContainerInstance
}

var containerInstanceStore = &containerInstanceStoreType{
	instances: map[string][]ContainerInstance{},
}

func (s *containerInstanceStoreType) list(slug string) []ContainerInstance {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.instances[slug]
	out := make([]ContainerInstance, len(src))
	copy(out, src)
	return out
}

func (s *containerInstanceStoreType) getByID(slug, id string) (ContainerInstance, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, c := range s.instances[slug] {
		if c.ID == id {
			return c, true
		}
	}
	return ContainerInstance{}, false
}

func (s *containerInstanceStoreType) add(slug string, c ContainerInstance) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.instances[slug] = append(s.instances[slug], c)
}

func (s *containerInstanceStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.instances[slug]
	for i, c := range list {
		if c.ID == id {
			s.instances[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiContainerInstance struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Image  string `json:"image"`
	CPUs   string `json:"cpus"`
	Memory string `json:"memory"`
	Region string `json:"region"`
	Status string `json:"status"`
	IP     string `json:"ip"`
}

type apiContainerInstanceList struct {
	Instances []apiContainerInstance `json:"instances"`
	Total     int                    `json:"total"`
}

func toAPIContainerInstance(c ContainerInstance) apiContainerInstance {
	return apiContainerInstance{
		ID: c.ID, Name: c.Name, Image: c.Image,
		CPUs: c.CPUs, Memory: c.Memory, Region: c.Region,
		Status: c.Status, IP: c.IP,
	}
}

func filterContainerInstances(all []ContainerInstance, search, status string) []ContainerInstance {
	var out []ContainerInstance
	q := strings.ToLower(search)
	for _, c := range all {
		if status != "" && c.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(c.Name + " " + c.Image + " " + c.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, c)
	}
	return out
}

func apiWSComputeContainerInstancesListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	instances := filterContainerInstances(containerInstanceStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiContainerInstance, len(instances))
	for i, c := range instances {
		out[i] = toAPIContainerInstance(c)
	}
	apiOK(w, apiContainerInstanceList{Instances: out, Total: len(out)})
}

type apiCreateContainerInstanceRequest struct {
	Name   string `json:"name"`
	Image  string `json:"image"`
	CPUs   string `json:"cpus"`
	Memory string `json:"memory"`
	Region string `json:"region"`
}

func apiWSComputeContainerInstancesCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateContainerInstanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	image := strings.TrimSpace(req.Image)
	if image == "" {
		apiError(w, http.StatusBadRequest, "Image is required.")
		return
	}
	cpus := strings.TrimSpace(req.CPUs)
	if cpus == "" {
		cpus = "1"
	}
	memory := strings.TrimSpace(req.Memory)
	if memory == "" {
		memory = "512 MB"
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "eu-central-1"
	}
	ci := ContainerInstance{
		ID: newID(), Name: name, Image: image,
		CPUs: cpus, Memory: memory, Region: region,
		Status: "creating", IP: "",
	}
	containerInstanceStore.add(slug, ci)
	apiCreated(w, toAPIContainerInstance(ci))
}

func apiWSComputeContainerInstancesGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	ci, found := containerInstanceStore.getByID(slug, r.PathValue("ciid"))
	if !found {
		apiError(w, http.StatusNotFound, "Container instance not found.")
		return
	}
	apiOK(w, toAPIContainerInstance(ci))
}

func apiWSComputeContainerInstancesDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := containerInstanceStore.getByID(slug, r.PathValue("ciid")); !found {
		apiError(w, http.StatusNotFound, "Container instance not found.")
		return
	}
	containerInstanceStore.delete(slug, r.PathValue("ciid"))
	apiOK(w, map[string]any{"ok": true, "message": "Container instance deleted."})
}

// ═══════════════════════════════════════════════════════════════
// 7. Web App Hosting   compute/hosting/web
// ═══════════════════════════════════════════════════════════════

type WebApp struct {
	ID       string
	Name     string
	Runtime  string
	Domain   string
	Region   string
	Status   string
	Replicas int
	Memory   string
}

type webAppStoreType struct {
	mu   sync.RWMutex
	apps map[string][]WebApp
}

var webAppStore = &webAppStoreType{
	apps: map[string][]WebApp{},
}

func (s *webAppStoreType) list(slug string) []WebApp {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.apps[slug]
	out := make([]WebApp, len(src))
	copy(out, src)
	return out
}

func (s *webAppStoreType) getByID(slug, id string) (WebApp, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, a := range s.apps[slug] {
		if a.ID == id {
			return a, true
		}
	}
	return WebApp{}, false
}

func (s *webAppStoreType) add(slug string, a WebApp) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.apps[slug] = append(s.apps[slug], a)
}

func (s *webAppStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.apps[slug]
	for i, a := range list {
		if a.ID == id {
			s.apps[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiWebApp struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Runtime  string `json:"runtime"`
	Domain   string `json:"domain"`
	Region   string `json:"region"`
	Status   string `json:"status"`
	Replicas int    `json:"replicas"`
	Memory   string `json:"memory"`
}

type apiWebAppList struct {
	Apps  []apiWebApp `json:"apps"`
	Total int         `json:"total"`
}

func toAPIWebApp(a WebApp) apiWebApp {
	return apiWebApp{
		ID: a.ID, Name: a.Name, Runtime: a.Runtime,
		Domain: a.Domain, Region: a.Region, Status: a.Status,
		Replicas: a.Replicas, Memory: a.Memory,
	}
}

func filterWebApps(all []WebApp, search, status string) []WebApp {
	var out []WebApp
	q := strings.ToLower(search)
	for _, a := range all {
		if status != "" && a.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(a.Name + " " + a.Runtime + " " + a.Domain + " " + a.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, a)
	}
	return out
}

func apiWSComputeHostingWebListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	apps := filterWebApps(webAppStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiWebApp, len(apps))
	for i, a := range apps {
		out[i] = toAPIWebApp(a)
	}
	apiOK(w, apiWebAppList{Apps: out, Total: len(out)})
}

type apiCreateWebAppRequest struct {
	Name    string `json:"name"`
	Runtime string `json:"runtime"`
	Domain  string `json:"domain"`
	Region  string `json:"region"`
}

func apiWSComputeHostingWebCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateWebAppRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	runtime := strings.TrimSpace(req.Runtime)
	if runtime == "" {
		runtime = "Node.js 20"
	}
	domain := strings.TrimSpace(req.Domain)
	if domain == "" {
		domain = name + ".app"
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "eu-central-1"
	}
	app := WebApp{
		ID: newID(), Name: name, Runtime: runtime,
		Domain: domain, Region: region,
		Status: "deploying", Replicas: 1, Memory: "256 MB",
	}
	webAppStore.add(slug, app)
	apiCreated(w, toAPIWebApp(app))
}

func apiWSComputeHostingWebGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	app, found := webAppStore.getByID(slug, r.PathValue("waid"))
	if !found {
		apiError(w, http.StatusNotFound, "Web app not found.")
		return
	}
	apiOK(w, toAPIWebApp(app))
}

func apiWSComputeHostingWebDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := webAppStore.getByID(slug, r.PathValue("waid")); !found {
		apiError(w, http.StatusNotFound, "Web app not found.")
		return
	}
	webAppStore.delete(slug, r.PathValue("waid"))
	apiOK(w, map[string]any{"ok": true, "message": "Web app deleted."})
}

// ═══════════════════════════════════════════════════════════════
// 8. Static Site Hosting   compute/hosting/static
// ═══════════════════════════════════════════════════════════════

type StaticSite struct {
	ID         string
	Name       string
	Framework  string
	Domain     string
	Region     string
	Status     string
	Size       string
	LastDeploy string
}

type staticSiteStoreType struct {
	mu    sync.RWMutex
	sites map[string][]StaticSite
}

var staticSiteStore = &staticSiteStoreType{
	sites: map[string][]StaticSite{},
}

func (s *staticSiteStoreType) list(slug string) []StaticSite {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.sites[slug]
	out := make([]StaticSite, len(src))
	copy(out, src)
	return out
}

func (s *staticSiteStoreType) getByID(slug, id string) (StaticSite, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, site := range s.sites[slug] {
		if site.ID == id {
			return site, true
		}
	}
	return StaticSite{}, false
}

func (s *staticSiteStoreType) add(slug string, site StaticSite) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sites[slug] = append(s.sites[slug], site)
}

func (s *staticSiteStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.sites[slug]
	for i, site := range list {
		if site.ID == id {
			s.sites[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiStaticSite struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Framework  string `json:"framework"`
	Domain     string `json:"domain"`
	Region     string `json:"region"`
	Status     string `json:"status"`
	Size       string `json:"size"`
	LastDeploy string `json:"last_deploy"`
}

type apiStaticSiteList struct {
	Sites []apiStaticSite `json:"sites"`
	Total int             `json:"total"`
}

func toAPIStaticSite(s StaticSite) apiStaticSite {
	return apiStaticSite{
		ID: s.ID, Name: s.Name, Framework: s.Framework,
		Domain: s.Domain, Region: s.Region, Status: s.Status,
		Size: s.Size, LastDeploy: s.LastDeploy,
	}
}

func filterStaticSites(all []StaticSite, search, status string) []StaticSite {
	var out []StaticSite
	q := strings.ToLower(search)
	for _, site := range all {
		if status != "" && site.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(site.Name + " " + site.Framework + " " + site.Domain + " " + site.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, site)
	}
	return out
}

func apiWSComputeHostingStaticListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	sites := filterStaticSites(staticSiteStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiStaticSite, len(sites))
	for i, site := range sites {
		out[i] = toAPIStaticSite(site)
	}
	apiOK(w, apiStaticSiteList{Sites: out, Total: len(out)})
}

type apiCreateStaticSiteRequest struct {
	Name      string `json:"name"`
	Framework string `json:"framework"`
	Domain    string `json:"domain"`
	Region    string `json:"region"`
}

func apiWSComputeHostingStaticCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateStaticSiteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	framework := strings.TrimSpace(req.Framework)
	if framework == "" {
		framework = "HTML/CSS"
	}
	domain := strings.TrimSpace(req.Domain)
	if domain == "" {
		domain = name + ".static.app"
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "eu-central-1"
	}
	site := StaticSite{
		ID: newID(), Name: name, Framework: framework,
		Domain: domain, Region: region,
		Status: "building", Size: "0 MB", LastDeploy: "never",
	}
	staticSiteStore.add(slug, site)
	apiCreated(w, toAPIStaticSite(site))
}

func apiWSComputeHostingStaticGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	site, found := staticSiteStore.getByID(slug, r.PathValue("ssid"))
	if !found {
		apiError(w, http.StatusNotFound, "Static site not found.")
		return
	}
	apiOK(w, toAPIStaticSite(site))
}

func apiWSComputeHostingStaticDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := staticSiteStore.getByID(slug, r.PathValue("ssid")); !found {
		apiError(w, http.StatusNotFound, "Static site not found.")
		return
	}
	staticSiteStore.delete(slug, r.PathValue("ssid"))
	apiOK(w, map[string]any{"ok": true, "message": "Static site deleted."})
}

// ═══════════════════════════════════════════════════════════════
// 9. Functions as a Service   compute/serverless/functions
// ═══════════════════════════════════════════════════════════════

type ServerlessFunc struct {
	ID          string
	Name        string
	Runtime     string
	Trigger     string
	Region      string
	Status      string
	Invocations string
	AvgLatency  string
	Memory      string
}

type serverlessFuncStoreType struct {
	mu    sync.RWMutex
	funcs map[string][]ServerlessFunc
}

var serverlessFuncStore = &serverlessFuncStoreType{
	funcs: map[string][]ServerlessFunc{},
}

func (s *serverlessFuncStoreType) list(slug string) []ServerlessFunc {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.funcs[slug]
	out := make([]ServerlessFunc, len(src))
	copy(out, src)
	return out
}

func (s *serverlessFuncStoreType) getByID(slug, id string) (ServerlessFunc, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, f := range s.funcs[slug] {
		if f.ID == id {
			return f, true
		}
	}
	return ServerlessFunc{}, false
}

func (s *serverlessFuncStoreType) add(slug string, f ServerlessFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.funcs[slug] = append(s.funcs[slug], f)
}

func (s *serverlessFuncStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.funcs[slug]
	for i, f := range list {
		if f.ID == id {
			s.funcs[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiServerlessFunc struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Runtime     string `json:"runtime"`
	Trigger     string `json:"trigger"`
	Region      string `json:"region"`
	Status      string `json:"status"`
	Invocations string `json:"invocations"`
	AvgLatency  string `json:"avg_latency"`
	Memory      string `json:"memory"`
}

type apiServerlessFuncList struct {
	Functions []apiServerlessFunc `json:"functions"`
	Total     int                 `json:"total"`
}

func toAPIServerlessFunc(f ServerlessFunc) apiServerlessFunc {
	return apiServerlessFunc{
		ID: f.ID, Name: f.Name, Runtime: f.Runtime,
		Trigger: f.Trigger, Region: f.Region, Status: f.Status,
		Invocations: f.Invocations, AvgLatency: f.AvgLatency, Memory: f.Memory,
	}
}

func filterServerlessFuncs(all []ServerlessFunc, search, status string) []ServerlessFunc {
	var out []ServerlessFunc
	q := strings.ToLower(search)
	for _, f := range all {
		if status != "" && f.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(f.Name + " " + f.Runtime + " " + f.Trigger + " " + f.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, f)
	}
	return out
}

func apiWSComputeServerlessFunctionsListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	funcs := filterServerlessFuncs(serverlessFuncStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiServerlessFunc, len(funcs))
	for i, f := range funcs {
		out[i] = toAPIServerlessFunc(f)
	}
	apiOK(w, apiServerlessFuncList{Functions: out, Total: len(out)})
}

type apiCreateServerlessFuncRequest struct {
	Name    string `json:"name"`
	Runtime string `json:"runtime"`
	Trigger string `json:"trigger"`
	Region  string `json:"region"`
	Memory  string `json:"memory"`
}

func apiWSComputeServerlessFunctionsCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateServerlessFuncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	runtime := strings.TrimSpace(req.Runtime)
	if runtime == "" {
		runtime = "Node.js 20"
	}
	trigger := strings.TrimSpace(req.Trigger)
	if trigger == "" {
		trigger = "HTTP GET"
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "eu-central-1"
	}
	memory := strings.TrimSpace(req.Memory)
	if memory == "" {
		memory = "128 MB"
	}
	fn := ServerlessFunc{
		ID: newID(), Name: name, Runtime: runtime,
		Trigger: trigger, Region: region,
		Status: "active", Invocations: "0", AvgLatency: "—", Memory: memory,
	}
	serverlessFuncStore.add(slug, fn)
	apiCreated(w, toAPIServerlessFunc(fn))
}

func apiWSComputeServerlessFunctionsGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	fn, found := serverlessFuncStore.getByID(slug, r.PathValue("fid"))
	if !found {
		apiError(w, http.StatusNotFound, "Function not found.")
		return
	}
	apiOK(w, toAPIServerlessFunc(fn))
}

func apiWSComputeServerlessFunctionsDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := serverlessFuncStore.getByID(slug, r.PathValue("fid")); !found {
		apiError(w, http.StatusNotFound, "Function not found.")
		return
	}
	serverlessFuncStore.delete(slug, r.PathValue("fid"))
	apiOK(w, map[string]any{"ok": true, "message": "Function deleted."})
}

// ═══════════════════════════════════════════════════════════════
// 10. Edge Functions   compute/serverless/edge
// ═══════════════════════════════════════════════════════════════

type EdgeFunc struct {
	ID         string
	Name       string
	Route      string
	Regions    string
	Status     string
	Requests   string
	P50Latency string
	Size       string
}

type edgeFuncStoreType struct {
	mu    sync.RWMutex
	funcs map[string][]EdgeFunc
}

var edgeFuncStore = &edgeFuncStoreType{
	funcs: map[string][]EdgeFunc{},
}

func (s *edgeFuncStoreType) list(slug string) []EdgeFunc {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.funcs[slug]
	out := make([]EdgeFunc, len(src))
	copy(out, src)
	return out
}

func (s *edgeFuncStoreType) getByID(slug, id string) (EdgeFunc, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, f := range s.funcs[slug] {
		if f.ID == id {
			return f, true
		}
	}
	return EdgeFunc{}, false
}

func (s *edgeFuncStoreType) add(slug string, f EdgeFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.funcs[slug] = append(s.funcs[slug], f)
}

func (s *edgeFuncStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.funcs[slug]
	for i, f := range list {
		if f.ID == id {
			s.funcs[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiEdgeFunc struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Route      string `json:"route"`
	Regions    string `json:"regions"`
	Status     string `json:"status"`
	Requests   string `json:"requests"`
	P50Latency string `json:"p50_latency"`
	Size       string `json:"size"`
}

type apiEdgeFuncList struct {
	Functions []apiEdgeFunc `json:"functions"`
	Total     int           `json:"total"`
}

func toAPIEdgeFunc(f EdgeFunc) apiEdgeFunc {
	return apiEdgeFunc{
		ID: f.ID, Name: f.Name, Route: f.Route,
		Regions: f.Regions, Status: f.Status,
		Requests: f.Requests, P50Latency: f.P50Latency, Size: f.Size,
	}
}

func filterEdgeFuncs(all []EdgeFunc, search, status string) []EdgeFunc {
	var out []EdgeFunc
	q := strings.ToLower(search)
	for _, f := range all {
		if status != "" && f.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(f.Name + " " + f.Route + " " + f.Regions)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, f)
	}
	return out
}

func apiWSComputeServerlessEdgeListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	funcs := filterEdgeFuncs(edgeFuncStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiEdgeFunc, len(funcs))
	for i, f := range funcs {
		out[i] = toAPIEdgeFunc(f)
	}
	apiOK(w, apiEdgeFuncList{Functions: out, Total: len(out)})
}

type apiCreateEdgeFuncRequest struct {
	Name    string `json:"name"`
	Route   string `json:"route"`
	Regions string `json:"regions"`
}

func apiWSComputeServerlessEdgeCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateEdgeFuncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	route := strings.TrimSpace(req.Route)
	if route == "" {
		route = "/*"
	}
	regions := strings.TrimSpace(req.Regions)
	if regions == "" {
		regions = "all (42 PoPs)"
	}
	fn := EdgeFunc{
		ID: newID(), Name: name, Route: route,
		Regions: regions, Status: "active",
		Requests: "0", P50Latency: "—", Size: "0 KB",
	}
	edgeFuncStore.add(slug, fn)
	apiCreated(w, toAPIEdgeFunc(fn))
}

func apiWSComputeServerlessEdgeGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	fn, found := edgeFuncStore.getByID(slug, r.PathValue("eid"))
	if !found {
		apiError(w, http.StatusNotFound, "Edge function not found.")
		return
	}
	apiOK(w, toAPIEdgeFunc(fn))
}

func apiWSComputeServerlessEdgeDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := edgeFuncStore.getByID(slug, r.PathValue("eid")); !found {
		apiError(w, http.StatusNotFound, "Edge function not found.")
		return
	}
	edgeFuncStore.delete(slug, r.PathValue("eid"))
	apiOK(w, map[string]any{"ok": true, "message": "Edge function deleted."})
}
