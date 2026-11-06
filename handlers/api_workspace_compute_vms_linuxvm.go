package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"platform/db"
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

func linuxVMFromDB(r db.LinuxVMDB) LinuxVM {
	st := r.VMStatus
	if st == "" {
		st = r.Status
	}
	return LinuxVM{
		ID:          strconv.FormatInt(r.ResourceID, 10),
		Name:        r.Name,
		Distro:      r.Distro,
		Version:     r.Version,
		VCPUs:       r.VCPUs,
		MemoryGB:    r.MemoryGB,
		DiskGB:      r.DiskGB,
		Region:      r.Region,
		PublicIP:    r.PublicIP,
		PrivateIP:   r.PrivateIP,
		SSHKey:      r.SSHKey,
		Status:      st,
		Uptime:      r.Uptime,
		MonthlyCost: r.MonthlyCost,
		Updated:     r.CreatedAt.Format("2006-01-02"),
	}
}

func loadWorkspaceLinuxVMs(ctx context.Context, slug, wsID string) []LinuxVM {
	if Pools != nil {
		_, wsDBID, err := resolveTenantAndWorkspaceDBIDs(ctx, slug, wsID)
		if err == nil && wsDBID > 0 {
			if rows, qErr := db.ListLinuxVMs(ctx, Pools.Admin, wsDBID); qErr == nil {
				out := make([]LinuxVM, len(rows))
				for i, r := range rows {
					out[i] = linuxVMFromDB(r)
				}
				return out
			}
		}
	}
	return linuxVMStore.list(slug)
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
	slug, wsID, ok := wsAuth(w, r)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	items := filterLinuxVMs(loadWorkspaceLinuxVMs(ctx, slug, wsID), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
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
	slug, wsID, ok := wsAuth(w, r)
	if !ok {
		return
	}
	user, _ := getSession(r)

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
		memGB = 2
	}
	diskGB := req.DiskGB
	if diskGB <= 0 {
		diskGB = 20
	}
	region := strings.TrimSpace(req.Region)
	sshKey := strings.TrimSpace(req.SSHKey)

	if Pools != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
		defer cancel()

		res, err := ProvisionCustomerVM(ctx, CustomerVMPurchaseInput{
			TenantSlug:      slug,
			WorkspaceIDStr:  wsID,
			ActorUserID:     user.ID,
			Name:            name,
			Distro:          distro,
			Version:         version,
			SSHPublicKey:    sshKey,
			VCPUs:           vcpus,
			MemoryGB:        memGB,
			DiskGB:          diskGB,
			PreferredRegion: region,
		})
		if err != nil {
			if errors.Is(err, db.ErrNoAssignableInventory) {
				apiError(w, http.StatusConflict, "No online provisioned supply host with available assignable IPs in tenant inventory. Onboard a Linux Nested VPS, run Auto-Provision, and add manual IPs in /t/"+slug+"/resources/.")
				return
			}
			apiError(w, http.StatusBadGateway, "VM provisioning failed: "+err.Error())
			return
		}

		vm := LinuxVM{
			ID:          strconv.FormatInt(res.VMResourceID, 10),
			Name:        name,
			Distro:      distro,
			Version:     version,
			VCPUs:       vcpus,
			MemoryGB:    memGB,
			DiskGB:      diskGB,
			Region:      res.RegionSlug,
			PublicIP:    res.AssignedIPv4,
			PrivateIP:   "",
			SSHKey:      sshKey,
			Status:      res.Status,
			Uptime:      "Just started",
			MonthlyCost: res.MonthlyCostDisplay,
			Updated:     time.Now().UTC().Format("2006-01-02"),
		}
		apiCreated(w, toAPILinuxVM(vm))
		return
	}

	vm := LinuxVM{
		ID: newID(), Name: name, Distro: distro,
		Version: version, VCPUs: vcpus, MemoryGB: memGB,
		DiskGB: diskGB, Region: region, PublicIP: "Pending",
		PrivateIP: "", SSHKey: sshKey, Status: "provisioning",
		Uptime: "—", MonthlyCost: "$0.00", Updated: time.Now().UTC().Format("2006-01-02"),
	}
	linuxVMStore.add(slug, vm)
	apiCreated(w, toAPILinuxVM(vm))
}

func apiWSComputeVMsLinuxGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, wsID, ok := wsAuth(w, r)
	if !ok {
		return
	}
	targetID := r.PathValue("vid")
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	for _, vm := range loadWorkspaceLinuxVMs(ctx, slug, wsID) {
		if vm.ID == targetID {
			apiOK(w, toAPILinuxVM(vm))
			return
		}
	}
	apiError(w, http.StatusNotFound, "Linux VM not found.")
}

func apiWSComputeVMsLinuxDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, wsID, ok := wsAuth(w, r)
	if !ok {
		return
	}
	vid := r.PathValue("vid")
	if Pools != nil {
		numID, err := strconv.ParseInt(vid, 10, 64)
		if err != nil || numID <= 0 {
			apiError(w, http.StatusNotFound, "Linux VM not found.")
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		found := false
		for _, vm := range loadWorkspaceLinuxVMs(ctx, slug, wsID) {
			if vm.ID == vid {
				found = true
				break
			}
		}
		if !found {
			apiError(w, http.StatusNotFound, "Linux VM not found.")
			return
		}
		if err := TerminateCustomerVM(ctx, numID); err != nil {
			apiError(w, http.StatusInternalServerError, err.Error())
			return
		}
		apiOK(w, map[string]any{"ok": true, "message": "Linux VM terminated and assignable IP returned to host pool."})
		return
	}

	if _, found := linuxVMStore.getByID(slug, vid); !found {
		apiError(w, http.StatusNotFound, "Linux VM not found.")
		return
	}
	linuxVMStore.delete(slug, vid)
	apiOK(w, map[string]any{"ok": true, "message": "Linux VM deleted."})
}
