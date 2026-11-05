package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Networking API — Virtual Network: Peering & Transit
// networking/vnet/peering
// ════════════════════════════════════════════════════════════════

type NetPeering struct {
	ID          string
	Name        string
	LocalVPC    string
	RemoteVPC   string
	RemoteOwner string
	Status      string
	Direction   string
	Bandwidth   string
	Region      string
	Encrypted   bool
}

type netPeeringStoreType struct {
	mu        sync.RWMutex
	instances map[string][]NetPeering
}

var netPeeringStore = &netPeeringStoreType{
	instances: map[string][]NetPeering{
		"buyvm": {
			{ID: newID(), Name: "prod-to-data", LocalVPC: "prod-vpc", RemoteVPC: "data-vpc", RemoteOwner: "self", Status: "active", Direction: "bidirectional", Bandwidth: "10 Gbps", Region: "eu-central-1", Encrypted: true},
			{ID: newID(), Name: "prod-to-us", LocalVPC: "prod-vpc", RemoteVPC: "us-prod-vpc", RemoteOwner: "self", Status: "active", Direction: "bidirectional", Bandwidth: "5 Gbps", Region: "cross-region", Encrypted: true},
			{ID: newID(), Name: "staging-to-prod", LocalVPC: "staging-vpc", RemoteVPC: "prod-vpc", RemoteOwner: "self", Status: "active", Direction: "bidirectional", Bandwidth: "1 Gbps", Region: "eu-central-1", Encrypted: false},
			{ID: newID(), Name: "transit-partner-a", LocalVPC: "prod-vpc", RemoteVPC: "partner-vpc-001", RemoteOwner: "acct-9182736450", Status: "active", Direction: "requester", Bandwidth: "2 Gbps", Region: "eu-central-1", Encrypted: true},
		},
		"hosting": {
			{ID: newID(), Name: "main-to-backend", LocalVPC: "main-vpc", RemoteVPC: "backend-vpc", RemoteOwner: "self", Status: "active", Direction: "bidirectional", Bandwidth: "10 Gbps", Region: "ap-south-1", Encrypted: true},
			{ID: newID(), Name: "main-to-isolated", LocalVPC: "main-vpc", RemoteVPC: "isolated-vpc", RemoteOwner: "self", Status: "pending-acceptance", Direction: "requester", Bandwidth: "1 Gbps", Region: "ap-south-1", Encrypted: true},
		},
	},
}

func (s *netPeeringStoreType) list(slug string) []NetPeering {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.instances[slug]
	out := make([]NetPeering, len(src))
	copy(out, src)
	return out
}

func (s *netPeeringStoreType) getByID(slug, id string) (NetPeering, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, d := range s.instances[slug] {
		if d.ID == id {
			return d, true
		}
	}
	return NetPeering{}, false
}

func (s *netPeeringStoreType) add(slug string, d NetPeering) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.instances[slug] = append(s.instances[slug], d)
}

func (s *netPeeringStoreType) delete(slug, id string) {
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

type apiNetPeering struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	LocalVPC    string `json:"local_vpc"`
	RemoteVPC   string `json:"remote_vpc"`
	RemoteOwner string `json:"remote_owner"`
	Status      string `json:"status"`
	Direction   string `json:"direction"`
	Bandwidth   string `json:"bandwidth"`
	Region      string `json:"region"`
	Encrypted   bool   `json:"encrypted"`
}

type apiNetPeeringList struct {
	Peerings []apiNetPeering `json:"peerings"`
	Total    int             `json:"total"`
}

func toAPINetPeering(d NetPeering) apiNetPeering {
	return apiNetPeering{
		ID: d.ID, Name: d.Name, LocalVPC: d.LocalVPC,
		RemoteVPC: d.RemoteVPC, RemoteOwner: d.RemoteOwner,
		Status: d.Status, Direction: d.Direction,
		Bandwidth: d.Bandwidth, Region: d.Region,
		Encrypted: d.Encrypted,
	}
}

func filterNetPeering(all []NetPeering, search, status string) []NetPeering {
	var out []NetPeering
	q := strings.ToLower(search)
	for _, d := range all {
		if status != "" && d.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(d.Name + " " + d.LocalVPC + " " + d.RemoteVPC + " " + d.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, d)
	}
	return out
}

func apiWSNetPeeringListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	items := filterNetPeering(netPeeringStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiNetPeering, len(items))
	for i, d := range items {
		out[i] = toAPINetPeering(d)
	}
	apiOK(w, apiNetPeeringList{Peerings: out, Total: len(out)})
}

type apiCreateNetPeeringRequest struct {
	Name        string `json:"name"`
	LocalVPC    string `json:"local_vpc"`
	RemoteVPC   string `json:"remote_vpc"`
	RemoteOwner string `json:"remote_owner"`
	Direction   string `json:"direction"`
	Bandwidth   string `json:"bandwidth"`
	Region      string `json:"region"`
	Encrypted   bool   `json:"encrypted"`
}

func apiWSNetPeeringCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateNetPeeringRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	localVPC := strings.TrimSpace(req.LocalVPC)
	if localVPC == "" {
		apiError(w, http.StatusBadRequest, "Local VPC is required.")
		return
	}
	remoteVPC := strings.TrimSpace(req.RemoteVPC)
	if remoteVPC == "" {
		apiError(w, http.StatusBadRequest, "Remote VPC is required.")
		return
	}
	owner := strings.TrimSpace(req.RemoteOwner)
	if owner == "" {
		owner = "self"
	}
	direction := strings.TrimSpace(req.Direction)
	if direction == "" {
		direction = "bidirectional"
	}
	bandwidth := strings.TrimSpace(req.Bandwidth)
	if bandwidth == "" {
		bandwidth = "1 Gbps"
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "eu-central-1"
	}
	inst := NetPeering{
		ID: newID(), Name: name, LocalVPC: localVPC,
		RemoteVPC: remoteVPC, RemoteOwner: owner,
		Status: "pending-acceptance", Direction: direction,
		Bandwidth: bandwidth, Region: region,
		Encrypted: req.Encrypted,
	}
	netPeeringStore.add(slug, inst)
	apiCreated(w, toAPINetPeering(inst))
}

func apiWSNetPeeringGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	inst, found := netPeeringStore.getByID(slug, r.PathValue("pid"))
	if !found {
		apiError(w, http.StatusNotFound, "Peering not found.")
		return
	}
	apiOK(w, toAPINetPeering(inst))
}

func apiWSNetPeeringDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := netPeeringStore.getByID(slug, r.PathValue("pid")); !found {
		apiError(w, http.StatusNotFound, "Peering not found.")
		return
	}
	netPeeringStore.delete(slug, r.PathValue("pid"))
	apiOK(w, map[string]any{"ok": true, "message": "Peering deleted."})
}
