package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Compute API — Virtual Machines: Linux VMs
// compute/virtual-machines/linux-vms
// ════════════════════════════════════════════════════════════════

type LinuxVM struct {
	ID        string
	Name      string
	Distro    string
	Version   string
	VCPUs     int
	MemoryGB  int
	DiskGB    int
	Region    string
	PublicIP  string
	PrivateIP string
	SSHKey    string
	Status    string
	Uptime    string
	MonthlyCost string
	Updated   string
}

type linuxVMStoreType struct {
	mu    sync.RWMutex
	items map[string][]LinuxVM
}

var linuxVMStore = &linuxVMStoreType{
	items: map[string][]LinuxVM{
		"buyvm": {
			{ID: newID(), Name: "web-prod-01", Distro: "Ubuntu", Version: "24.04 LTS", VCPUs: 4, MemoryGB: 8, DiskGB: 160, Region: "us-east-1", PublicIP: "203.0.113.10", PrivateIP: "10.0.1.10", SSHKey: "ops-team-key", Status: "running", Uptime: "45d 12h", MonthlyCost: "$48.00", Updated: "2026-04-17"},
			{ID: newID(), Name: "db-replica-02", Distro: "Debian", Version: "12 Bookworm", VCPUs: 8, MemoryGB: 32, DiskGB: 500, Region: "us-east-1", PublicIP: "—", PrivateIP: "10.0.1.20", SSHKey: "db-admin-key", Status: "running", Uptime: "120d 3h", MonthlyCost: "$128.00", Updated: "2026-04-16"},
			{ID: newID(), Name: "ci-runner-03", Distro: "Rocky Linux", Version: "9.3", VCPUs: 2, MemoryGB: 4, DiskGB: 80, Region: "eu-west-1", PublicIP: "198.51.100.5", PrivateIP: "10.0.2.15", SSHKey: "ci-key", Status: "running", Uptime: "10d 8h", MonthlyCost: "$24.00", Updated: "2026-04-15"},
			{ID: newID(), Name: "staging-app", Distro: "Ubuntu", Version: "22.04 LTS", VCPUs: 2, MemoryGB: 4, DiskGB: 80, Region: "us-west-2", PublicIP: "203.0.113.55", PrivateIP: "10.0.3.10", SSHKey: "dev-key", Status: "stopped", Uptime: "—", MonthlyCost: "$6.00", Updated: "2026-04-14"},
			{ID: newID(), Name: "monitoring-node", Distro: "AlmaLinux", Version: "9.3", VCPUs: 4, MemoryGB: 16, DiskGB: 200, Region: "us-east-1", PublicIP: "203.0.113.99", PrivateIP: "10.0.1.50", SSHKey: "ops-team-key", Status: "running", Uptime: "60d 0h", MonthlyCost: "$72.00", Updated: "2026-04-17"},
		},
		"hosting": {
			{ID: newID(), Name: "api-gateway-01", Distro: "Ubuntu", Version: "24.04 LTS", VCPUs: 8, MemoryGB: 16, DiskGB: 200, Region: "eu-central-1", PublicIP: "198.51.100.20", PrivateIP: "10.1.0.10", SSHKey: "platform-key", Status: "running", Uptime: "90d 6h", MonthlyCost: "$96.00", Updated: "2026-04-17"},
			{ID: newID(), Name: "worker-pool-01", Distro: "Debian", Version: "12 Bookworm", VCPUs: 16, MemoryGB: 64, DiskGB: 1000, Region: "eu-central-1", PublicIP: "—", PrivateIP: "10.1.0.20", SSHKey: "platform-key", Status: "running", Uptime: "30d 18h", MonthlyCost: "$256.00", Updated: "2026-04-16"},
			{ID: newID(), Name: "bastion-host", Distro: "Alpine", Version: "3.19", VCPUs: 1, MemoryGB: 1, DiskGB: 20, Region: "eu-central-1", PublicIP: "198.51.100.30", PrivateIP: "10.1.0.5", SSHKey: "admin-key", Status: "running", Uptime: "180d 0h", MonthlyCost: "$5.00", Updated: "2026-04-10"},
			{ID: newID(), Name: "ml-training-gpu", Distro: "Ubuntu", Version: "22.04 LTS", VCPUs: 32, MemoryGB: 128, DiskGB: 2000, Region: "us-east-1", PublicIP: "203.0.113.200", PrivateIP: "10.2.0.10", SSHKey: "ml-team-key", Status: "stopped", Uptime: "—", MonthlyCost: "$45.00", Updated: "2026-04-12"},
		},
	},
}

func (s *linuxVMStoreType) list(slug string) []LinuxVM {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.items[slug]
	out := make([]LinuxVM, len(src))
	copy(out, src)
	return out
}

func (s *linuxVMStoreType) getByID(slug, id string) (LinuxVM, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, vm := range s.items[slug] {
		if vm.ID == id {
			return vm, true
		}
	}
	return LinuxVM{}, false
}

func (s *linuxVMStoreType) add(slug string, vm LinuxVM) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[slug] = append(s.items[slug], vm)
}

func (s *linuxVMStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.items[slug]
	for i, vm := range list {
		if vm.ID == id {
			s.items[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiLinuxVM struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Distro      string `json:"distro"`
	Version     string `json:"version"`
	VCPUs       int    `json:"vcpus"`
	MemoryGB    int    `json:"memory_gb"`
	DiskGB      int    `json:"disk_gb"`
	Region      string `json:"region"`
	PublicIP    string `json:"public_ip"`
	PrivateIP   string `json:"private_ip"`
	SSHKey      string `json:"ssh_key"`
	Status      string `json:"status"`
	Uptime      string `json:"uptime"`
	MonthlyCost string `json:"monthly_cost"`
	Updated     string `json:"updated"`
}

type apiLinuxVMList struct {
	VMs   []apiLinuxVM `json:"vms"`
	Total int          `json:"total"`
}

func toAPILinuxVM(v LinuxVM) apiLinuxVM {
	return apiLinuxVM{
		ID: v.ID, Name: v.Name, Distro: v.Distro,
		Version: v.Version, VCPUs: v.VCPUs, MemoryGB: v.MemoryGB,
		DiskGB: v.DiskGB, Region: v.Region, PublicIP: v.PublicIP,
		PrivateIP: v.PrivateIP, SSHKey: v.SSHKey, Status: v.Status,
		Uptime: v.Uptime, MonthlyCost: v.MonthlyCost, Updated: v.Updated,
	}
}

func filterLinuxVMs(all []LinuxVM, search, status string) []LinuxVM {
	var out []LinuxVM
	q := strings.ToLower(search)
	for _, v := range all {
		if status != "" && v.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(v.Name + " " + v.Distro + " " + v.Version + " " + v.Region + " " + v.PublicIP)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, v)
	}
	return out
}

func apiWSComputeVMsLinuxListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	items := filterLinuxVMs(linuxVMStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiLinuxVM, len(items))
	for i, v := range items {
		out[i] = toAPILinuxVM(v)
	}
	apiOK(w, apiLinuxVMList{VMs: out, Total: len(out)})
}

type apiCreateLinuxVMRequest struct {
	Name     string `json:"name"`
	Distro   string `json:"distro"`
	Version  string `json:"version"`
	VCPUs    int    `json:"vcpus"`
	MemoryGB int    `json:"memory_gb"`
	DiskGB   int    `json:"disk_gb"`
	Region   string `json:"region"`
	SSHKey   string `json:"ssh_key"`
}

func apiWSComputeVMsLinuxCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateLinuxVMRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	distro := strings.TrimSpace(req.Distro)
	if distro == "" {
		distro = "Ubuntu"
	}
	version := strings.TrimSpace(req.Version)
	if version == "" {
		version = "24.04 LTS"
	}
	vcpus := req.VCPUs
	if vcpus <= 0 {
		vcpus = 1
	}
	memGB := req.MemoryGB
	if memGB <= 0 {
		memGB = 1
	}
	diskGB := req.DiskGB
	if diskGB <= 0 {
		diskGB = 20
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "us-east-1"
	}
	sshKey := strings.TrimSpace(req.SSHKey)
	if sshKey == "" {
		sshKey = "default-key"
	}
	vm := LinuxVM{
		ID: newID(), Name: name, Distro: distro,
		Version: version, VCPUs: vcpus, MemoryGB: memGB,
		DiskGB: diskGB, Region: region, PublicIP: "Pending",
		PrivateIP: "Pending", SSHKey: sshKey, Status: "provisioning",
		Uptime: "—", MonthlyCost: "Calculating...", Updated: "2026-04-17",
	}
	linuxVMStore.add(slug, vm)
	apiCreated(w, toAPILinuxVM(vm))
}

func apiWSComputeVMsLinuxGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	vm, found := linuxVMStore.getByID(slug, r.PathValue("vid"))
	if !found {
		apiError(w, http.StatusNotFound, "Linux VM not found.")
		return
	}
	apiOK(w, toAPILinuxVM(vm))
}

func apiWSComputeVMsLinuxDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := linuxVMStore.getByID(slug, r.PathValue("vid")); !found {
		apiError(w, http.StatusNotFound, "Linux VM not found.")
		return
	}
	linuxVMStore.delete(slug, r.PathValue("vid"))
	apiOK(w, map[string]any{"ok": true, "message": "Linux VM deleted."})
}
