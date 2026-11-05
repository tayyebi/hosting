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
	ID          string
	Name        string
	Distro      string
	Version     string
	VCPUs       int
	MemoryGB    int
	DiskGB      int
	Region      string
	PublicIP    string
	PrivateIP   string
	SSHKey      string
	Status      string
	Uptime      string
	MonthlyCost string
	Updated     string
}

type linuxVMStoreType struct {
	mu    sync.RWMutex
	items map[string][]LinuxVM
}

var linuxVMStore = &linuxVMStoreType{
	items: map[string][]LinuxVM{},
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
