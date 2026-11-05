package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Compute API — Kubernetes: Managed K8s (Pro Feature)
// compute/kubernetes/managed-k8s
// ════════════════════════════════════════════════════════════════

type ManagedK8sCluster struct {
	ID              string
	Name            string
	K8sVersion      string
	NodePools       int
	TotalNodes      int
	VCPUs           int
	MemoryGB        int
	PodCount        int
	PodCapacity     int
	CNIPlugin       string
	IngressCtrl     string
	ServiceMesh     bool
	AutoScale       bool
	Region          string
	ControlPlaneHA  bool
	CostPerHour     float64
	Status          string
	Updated         string
}

type managedK8sStoreType struct {
	mu    sync.RWMutex
	items map[string][]ManagedK8sCluster
}

var managedK8sStore = &managedK8sStoreType{
	items: map[string][]ManagedK8sCluster{
		"buyvm": {
			{ID: newID(), Name: "prod-cluster-01", K8sVersion: "1.30.2", NodePools: 3, TotalNodes: 12, VCPUs: 96, MemoryGB: 384, PodCount: 284, PodCapacity: 660, CNIPlugin: "Cilium", IngressCtrl: "NGINX", ServiceMesh: true, AutoScale: true, Region: "us-east-1", ControlPlaneHA: true, CostPerHour: 18.40, Status: "running", Updated: "2026-04-18"},
			{ID: newID(), Name: "staging-cluster", K8sVersion: "1.30.2", NodePools: 1, TotalNodes: 3, VCPUs: 12, MemoryGB: 48, PodCount: 42, PodCapacity: 165, CNIPlugin: "Calico", IngressCtrl: "Traefik", ServiceMesh: false, AutoScale: false, Region: "us-east-1", ControlPlaneHA: false, CostPerHour: 3.60, Status: "running", Updated: "2026-04-17"},
			{ID: newID(), Name: "ml-gpu-cluster", K8sVersion: "1.29.6", NodePools: 2, TotalNodes: 4, VCPUs: 32, MemoryGB: 256, PodCount: 18, PodCapacity: 220, CNIPlugin: "Cilium", IngressCtrl: "—", ServiceMesh: false, AutoScale: true, Region: "us-west-2", ControlPlaneHA: true, CostPerHour: 12.80, Status: "running", Updated: "2026-04-16"},
			{ID: newID(), Name: "dev-sandbox", K8sVersion: "1.30.2", NodePools: 1, TotalNodes: 2, VCPUs: 8, MemoryGB: 16, PodCount: 0, PodCapacity: 110, CNIPlugin: "Flannel", IngressCtrl: "NGINX", ServiceMesh: false, AutoScale: false, Region: "eu-west-1", ControlPlaneHA: false, CostPerHour: 1.20, Status: "stopped", Updated: "2026-04-14"},
		},
		"hosting": {
			{ID: newID(), Name: "platform-prod", K8sVersion: "1.30.2", NodePools: 5, TotalNodes: 24, VCPUs: 192, MemoryGB: 768, PodCount: 620, PodCapacity: 1320, CNIPlugin: "Cilium", IngressCtrl: "Istio Gateway", ServiceMesh: true, AutoScale: true, Region: "eu-central-1", ControlPlaneHA: true, CostPerHour: 36.80, Status: "running", Updated: "2026-04-18"},
			{ID: newID(), Name: "edge-cluster-us", K8sVersion: "1.29.6", NodePools: 2, TotalNodes: 6, VCPUs: 24, MemoryGB: 96, PodCount: 88, PodCapacity: 330, CNIPlugin: "Calico", IngressCtrl: "NGINX", ServiceMesh: false, AutoScale: true, Region: "us-west-2", ControlPlaneHA: true, CostPerHour: 7.20, Status: "running", Updated: "2026-04-17"},
			{ID: newID(), Name: "data-pipeline", K8sVersion: "1.30.2", NodePools: 2, TotalNodes: 8, VCPUs: 64, MemoryGB: 256, PodCount: 142, PodCapacity: 440, CNIPlugin: "Cilium", IngressCtrl: "—", ServiceMesh: false, AutoScale: true, Region: "eu-central-1", ControlPlaneHA: true, CostPerHour: 14.40, Status: "running", Updated: "2026-04-17"},
			{ID: newID(), Name: "qa-ephemeral", K8sVersion: "1.30.2", NodePools: 1, TotalNodes: 0, VCPUs: 0, MemoryGB: 0, PodCount: 0, PodCapacity: 0, CNIPlugin: "Flannel", IngressCtrl: "NGINX", ServiceMesh: false, AutoScale: true, Region: "eu-central-1", ControlPlaneHA: false, CostPerHour: 0, Status: "scaled_to_zero", Updated: "2026-04-15"},
		},
	},
}

func (s *managedK8sStoreType) list(slug string) []ManagedK8sCluster {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.items[slug]
	out := make([]ManagedK8sCluster, len(src))
	copy(out, src)
	return out
}

func (s *managedK8sStoreType) getByID(slug, id string) (ManagedK8sCluster, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, c := range s.items[slug] {
		if c.ID == id {
			return c, true
		}
	}
	return ManagedK8sCluster{}, false
}

func (s *managedK8sStoreType) add(slug string, c ManagedK8sCluster) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[slug] = append(s.items[slug], c)
}

func (s *managedK8sStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.items[slug]
	for i, c := range list {
		if c.ID == id {
			s.items[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiManagedK8sCluster struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	K8sVersion     string  `json:"k8s_version"`
	NodePools      int     `json:"node_pools"`
	TotalNodes     int     `json:"total_nodes"`
	VCPUs          int     `json:"vcpus"`
	MemoryGB       int     `json:"memory_gb"`
	PodCount       int     `json:"pod_count"`
	PodCapacity    int     `json:"pod_capacity"`
	CNIPlugin      string  `json:"cni_plugin"`
	IngressCtrl    string  `json:"ingress_controller"`
	ServiceMesh    bool    `json:"service_mesh"`
	AutoScale      bool    `json:"auto_scale"`
	Region         string  `json:"region"`
	ControlPlaneHA bool    `json:"control_plane_ha"`
	CostPerHour    float64 `json:"cost_per_hour"`
	Status         string  `json:"status"`
	Updated        string  `json:"updated"`
}

type apiManagedK8sList struct {
	Clusters []apiManagedK8sCluster `json:"clusters"`
	Total    int                    `json:"total"`
}

func toAPIManagedK8s(c ManagedK8sCluster) apiManagedK8sCluster {
	return apiManagedK8sCluster{
		ID: c.ID, Name: c.Name, K8sVersion: c.K8sVersion,
		NodePools: c.NodePools, TotalNodes: c.TotalNodes,
		VCPUs: c.VCPUs, MemoryGB: c.MemoryGB,
		PodCount: c.PodCount, PodCapacity: c.PodCapacity,
		CNIPlugin: c.CNIPlugin, IngressCtrl: c.IngressCtrl,
		ServiceMesh: c.ServiceMesh, AutoScale: c.AutoScale,
		Region: c.Region, ControlPlaneHA: c.ControlPlaneHA,
		CostPerHour: c.CostPerHour, Status: c.Status,
		Updated: c.Updated,
	}
}

func filterManagedK8s(all []ManagedK8sCluster, search, status string) []ManagedK8sCluster {
	var out []ManagedK8sCluster
	q := strings.ToLower(search)
	for _, c := range all {
		if status != "" && c.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(c.Name + " " + c.K8sVersion + " " + c.CNIPlugin + " " + c.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, c)
	}
	return out
}

func apiWSComputeProK8sListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	items := filterManagedK8s(managedK8sStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiManagedK8sCluster, len(items))
	for i, c := range items {
		out[i] = toAPIManagedK8s(c)
	}
	apiOK(w, apiManagedK8sList{Clusters: out, Total: len(out)})
}

type apiCreateManagedK8sRequest struct {
	Name       string `json:"name"`
	K8sVersion string `json:"k8s_version"`
	NodePools  int    `json:"node_pools"`
	TotalNodes int    `json:"total_nodes"`
	CNIPlugin  string `json:"cni_plugin"`
	Region     string `json:"region"`
	AutoScale  bool   `json:"auto_scale"`
	HA         bool   `json:"control_plane_ha"`
}

func apiWSComputeProK8sCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateManagedK8sRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	version := strings.TrimSpace(req.K8sVersion)
	if version == "" {
		version = "1.30.2"
	}
	pools := req.NodePools
	if pools <= 0 {
		pools = 1
	}
	nodes := req.TotalNodes
	if nodes <= 0 {
		nodes = 3
	}
	cni := strings.TrimSpace(req.CNIPlugin)
	if cni == "" {
		cni = "Cilium"
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "us-east-1"
	}
	cluster := ManagedK8sCluster{
		ID: newID(), Name: name, K8sVersion: version,
		NodePools: pools, TotalNodes: nodes,
		VCPUs: nodes * 4, MemoryGB: nodes * 16,
		PodCount: 0, PodCapacity: nodes * 55,
		CNIPlugin: cni, IngressCtrl: "NGINX",
		ServiceMesh: false, AutoScale: req.AutoScale,
		Region: region, ControlPlaneHA: req.HA,
		CostPerHour: float64(nodes) * 1.20,
		Status: "provisioning", Updated: "2026-04-18",
	}
	managedK8sStore.add(slug, cluster)
	apiCreated(w, toAPIManagedK8s(cluster))
}

func apiWSComputeProK8sGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	c, found := managedK8sStore.getByID(slug, r.PathValue("cid"))
	if !found {
		apiError(w, http.StatusNotFound, "Kubernetes cluster not found.")
		return
	}
	apiOK(w, toAPIManagedK8s(c))
}

func apiWSComputeProK8sDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := managedK8sStore.getByID(slug, r.PathValue("cid")); !found {
		apiError(w, http.StatusNotFound, "Kubernetes cluster not found.")
		return
	}
	managedK8sStore.delete(slug, r.PathValue("cid"))
	apiOK(w, map[string]any{"ok": true, "message": "Kubernetes cluster deleted."})
}
