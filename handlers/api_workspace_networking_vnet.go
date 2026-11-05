package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Networking API — Virtual Network: VPC / Subnets
// networking/vnet/vpc
// ════════════════════════════════════════════════════════════════

type NetVPC struct {
	ID        string
	Name      string
	CIDR      string
	Region    string
	Status    string
	Subnets   int
	Instances int
	Gateway   string
	DNS       string
	Peerings  int
	Encrypted bool
}

type netVPCStoreType struct {
	mu        sync.RWMutex
	instances map[string][]NetVPC
}

var netVPCStore = &netVPCStoreType{
	instances: map[string][]NetVPC{
		"buyvm": {
			{ID: newID(), Name: "prod-vpc", CIDR: "10.0.0.0/16", Region: "eu-central-1", Status: "active", Subnets: 8, Instances: 142, Gateway: "igw-0a1b2c3d4e5f", DNS: "enabled", Peerings: 3, Encrypted: true},
			{ID: newID(), Name: "staging-vpc", CIDR: "10.1.0.0/16", Region: "eu-central-1", Status: "active", Subnets: 4, Instances: 28, Gateway: "igw-1b2c3d4e5f6a", DNS: "enabled", Peerings: 1, Encrypted: true},
			{ID: newID(), Name: "data-vpc", CIDR: "10.2.0.0/16", Region: "eu-central-1", Status: "active", Subnets: 6, Instances: 64, Gateway: "igw-2c3d4e5f6a7b", DNS: "enabled", Peerings: 2, Encrypted: true},
			{ID: newID(), Name: "us-prod-vpc", CIDR: "172.16.0.0/16", Region: "us-east-1", Status: "active", Subnets: 6, Instances: 86, Gateway: "igw-3d4e5f6a7b8c", DNS: "enabled", Peerings: 2, Encrypted: true},
			{ID: newID(), Name: "dev-vpc", CIDR: "10.99.0.0/16", Region: "eu-central-1", Status: "active", Subnets: 2, Instances: 8, Gateway: "igw-4e5f6a7b8c9d", DNS: "enabled", Peerings: 0, Encrypted: false},
		},
		"hosting": {
			{ID: newID(), Name: "main-vpc", CIDR: "10.0.0.0/16", Region: "ap-south-1", Status: "active", Subnets: 6, Instances: 94, Gateway: "igw-5f6a7b8c9d0e", DNS: "enabled", Peerings: 2, Encrypted: true},
			{ID: newID(), Name: "backend-vpc", CIDR: "10.1.0.0/16", Region: "ap-south-1", Status: "active", Subnets: 4, Instances: 52, Gateway: "igw-6a7b8c9d0e1f", DNS: "enabled", Peerings: 1, Encrypted: true},
			{ID: newID(), Name: "isolated-vpc", CIDR: "192.168.0.0/16", Region: "ap-south-1", Status: "active", Subnets: 2, Instances: 12, Gateway: "none", DNS: "disabled", Peerings: 0, Encrypted: true},
		},
	},
}

func (s *netVPCStoreType) list(slug string) []NetVPC {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.instances[slug]
	out := make([]NetVPC, len(src))
	copy(out, src)
	return out
}

func (s *netVPCStoreType) getByID(slug, id string) (NetVPC, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, d := range s.instances[slug] {
		if d.ID == id {
			return d, true
		}
	}
	return NetVPC{}, false
}

func (s *netVPCStoreType) add(slug string, d NetVPC) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.instances[slug] = append(s.instances[slug], d)
}

func (s *netVPCStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.instances[slug]
	for i, d := range list {
		if d.ID == id {
			s.instances[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiNetVPC struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CIDR      string `json:"cidr"`
	Region    string `json:"region"`
	Status    string `json:"status"`
	Subnets   int    `json:"subnets"`
	Instances int    `json:"instances"`
	Gateway   string `json:"gateway"`
	DNS       string `json:"dns"`
	Peerings  int    `json:"peerings"`
	Encrypted bool   `json:"encrypted"`
}

type apiNetVPCList struct {
	VPCs  []apiNetVPC `json:"vpcs"`
	Total int         `json:"total"`
}

func toAPINetVPC(d NetVPC) apiNetVPC {
	return apiNetVPC{
		ID: d.ID, Name: d.Name, CIDR: d.CIDR,
		Region: d.Region, Status: d.Status,
		Subnets: d.Subnets, Instances: d.Instances,
		Gateway: d.Gateway, DNS: d.DNS,
		Peerings: d.Peerings, Encrypted: d.Encrypted,
	}
}

func filterNetVPC(all []NetVPC, search, status string) []NetVPC {
	var out []NetVPC
	q := strings.ToLower(search)
	for _, d := range all {
		if status != "" && d.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(d.Name + " " + d.CIDR + " " + d.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, d)
	}
	return out
}

func apiWSNetVPCListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	vpcs := filterNetVPC(netVPCStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiNetVPC, len(vpcs))
	for i, d := range vpcs {
		out[i] = toAPINetVPC(d)
	}
	apiOK(w, apiNetVPCList{VPCs: out, Total: len(out)})
}

type apiCreateNetVPCRequest struct {
	Name      string `json:"name"`
	CIDR      string `json:"cidr"`
	Region    string `json:"region"`
	DNS       string `json:"dns"`
	Encrypted bool   `json:"encrypted"`
}

func apiWSNetVPCCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateNetVPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	cidr := strings.TrimSpace(req.CIDR)
	if cidr == "" {
		cidr = "10.0.0.0/16"
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "eu-central-1"
	}
	dns := strings.TrimSpace(req.DNS)
	if dns == "" {
		dns = "enabled"
	}
	inst := NetVPC{
		ID: newID(), Name: name, CIDR: cidr,
		Region: region, Status: "provisioning",
		Subnets: 0, Instances: 0,
		Gateway: "igw-" + newID()[:12],
		DNS: dns, Peerings: 0,
		Encrypted: req.Encrypted,
	}
	netVPCStore.add(slug, inst)
	apiCreated(w, toAPINetVPC(inst))
}

func apiWSNetVPCGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	inst, found := netVPCStore.getByID(slug, r.PathValue("vid"))
	if !found {
		apiError(w, http.StatusNotFound, "VPC not found.")
		return
	}
	apiOK(w, toAPINetVPC(inst))
}

func apiWSNetVPCDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := netVPCStore.getByID(slug, r.PathValue("vid")); !found {
		apiError(w, http.StatusNotFound, "VPC not found.")
		return
	}
	netVPCStore.delete(slug, r.PathValue("vid"))
	apiOK(w, map[string]any{"ok": true, "message": "VPC deleted."})
}
