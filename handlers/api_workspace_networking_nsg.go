package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Networking API — Virtual Network: Network Security Groups
// networking/vnet/nsg
// ════════════════════════════════════════════════════════════════

type NetNSG struct {
	ID             string
	Name           string
	VPC            string
	Region         string
	Status         string
	InboundRules   int
	OutboundRules  int
	AssociatedNICs int
	DefaultAction  string
	Logging        bool
}

type netNSGStoreType struct {
	mu        sync.RWMutex
	instances map[string][]NetNSG
}

var netNSGStore = &netNSGStoreType{
	instances: map[string][]NetNSG{
		"buyvm": {
			{ID: newID(), Name: "prod-web-nsg", VPC: "prod-vpc", Region: "eu-central-1", Status: "active", InboundRules: 12, OutboundRules: 5, AssociatedNICs: 28, DefaultAction: "deny", Logging: true},
			{ID: newID(), Name: "prod-db-nsg", VPC: "prod-vpc", Region: "eu-central-1", Status: "active", InboundRules: 4, OutboundRules: 2, AssociatedNICs: 8, DefaultAction: "deny", Logging: true},
			{ID: newID(), Name: "staging-nsg", VPC: "staging-vpc", Region: "eu-central-1", Status: "active", InboundRules: 8, OutboundRules: 8, AssociatedNICs: 14, DefaultAction: "deny", Logging: false},
			{ID: newID(), Name: "data-nsg", VPC: "data-vpc", Region: "eu-central-1", Status: "active", InboundRules: 6, OutboundRules: 3, AssociatedNICs: 18, DefaultAction: "deny", Logging: true},
			{ID: newID(), Name: "dev-open-nsg", VPC: "dev-vpc", Region: "eu-central-1", Status: "active", InboundRules: 2, OutboundRules: 1, AssociatedNICs: 4, DefaultAction: "allow", Logging: false},
		},
		"hosting": {
			{ID: newID(), Name: "main-nsg", VPC: "main-vpc", Region: "ap-south-1", Status: "active", InboundRules: 10, OutboundRules: 6, AssociatedNICs: 32, DefaultAction: "deny", Logging: true},
			{ID: newID(), Name: "backend-nsg", VPC: "backend-vpc", Region: "ap-south-1", Status: "active", InboundRules: 5, OutboundRules: 3, AssociatedNICs: 16, DefaultAction: "deny", Logging: true},
			{ID: newID(), Name: "isolated-nsg", VPC: "isolated-vpc", Region: "ap-south-1", Status: "active", InboundRules: 1, OutboundRules: 1, AssociatedNICs: 6, DefaultAction: "deny", Logging: true},
		},
	},
}

func (s *netNSGStoreType) list(slug string) []NetNSG {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.instances[slug]
	out := make([]NetNSG, len(src))
	copy(out, src)
	return out
}

func (s *netNSGStoreType) getByID(slug, id string) (NetNSG, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, d := range s.instances[slug] {
		if d.ID == id {
			return d, true
		}
	}
	return NetNSG{}, false
}

func (s *netNSGStoreType) add(slug string, d NetNSG) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.instances[slug] = append(s.instances[slug], d)
}

func (s *netNSGStoreType) delete(slug, id string) {
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

type apiNetNSG struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	VPC            string `json:"vpc"`
	Region         string `json:"region"`
	Status         string `json:"status"`
	InboundRules   int    `json:"inbound_rules"`
	OutboundRules  int    `json:"outbound_rules"`
	AssociatedNICs int    `json:"associated_nics"`
	DefaultAction  string `json:"default_action"`
	Logging        bool   `json:"logging"`
}

type apiNetNSGList struct {
	NSGs  []apiNetNSG `json:"nsgs"`
	Total int         `json:"total"`
}

func toAPINetNSG(d NetNSG) apiNetNSG {
	return apiNetNSG{
		ID: d.ID, Name: d.Name, VPC: d.VPC,
		Region: d.Region, Status: d.Status,
		InboundRules: d.InboundRules, OutboundRules: d.OutboundRules,
		AssociatedNICs: d.AssociatedNICs, DefaultAction: d.DefaultAction,
		Logging: d.Logging,
	}
}

func filterNetNSG(all []NetNSG, search, status string) []NetNSG {
	var out []NetNSG
	q := strings.ToLower(search)
	for _, d := range all {
		if status != "" && d.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(d.Name + " " + d.VPC + " " + d.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, d)
	}
	return out
}

func apiWSNetNSGListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	items := filterNetNSG(netNSGStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiNetNSG, len(items))
	for i, d := range items {
		out[i] = toAPINetNSG(d)
	}
	apiOK(w, apiNetNSGList{NSGs: out, Total: len(out)})
}

type apiCreateNetNSGRequest struct {
	Name          string `json:"name"`
	VPC           string `json:"vpc"`
	Region        string `json:"region"`
	DefaultAction string `json:"default_action"`
	Logging       bool   `json:"logging"`
}

func apiWSNetNSGCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateNetNSGRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	vpc := strings.TrimSpace(req.VPC)
	if vpc == "" {
		apiError(w, http.StatusBadRequest, "VPC is required.")
		return
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "eu-central-1"
	}
	action := strings.TrimSpace(req.DefaultAction)
	if action == "" {
		action = "deny"
	}
	inst := NetNSG{
		ID: newID(), Name: name, VPC: vpc,
		Region: region, Status: "active",
		InboundRules: 0, OutboundRules: 0,
		AssociatedNICs: 0, DefaultAction: action,
		Logging: req.Logging,
	}
	netNSGStore.add(slug, inst)
	apiCreated(w, toAPINetNSG(inst))
}

func apiWSNetNSGGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	inst, found := netNSGStore.getByID(slug, r.PathValue("nid"))
	if !found {
		apiError(w, http.StatusNotFound, "NSG not found.")
		return
	}
	apiOK(w, toAPINetNSG(inst))
}

func apiWSNetNSGDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := netNSGStore.getByID(slug, r.PathValue("nid")); !found {
		apiError(w, http.StatusNotFound, "NSG not found.")
		return
	}
	netNSGStore.delete(slug, r.PathValue("nid"))
	apiOK(w, map[string]any{"ok": true, "message": "NSG deleted."})
}
