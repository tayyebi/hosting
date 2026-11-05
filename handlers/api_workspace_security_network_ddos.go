package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Security API — Network Security: DDoS Protection
// security/network/ddos
// ════════════════════════════════════════════════════════════════

type DDoSProfile struct {
	ID          string
	Name        string
	Tier        string
	Mode        string
	Threshold   string
	Status      string
	Mitigated7d string
	Updated     string
}

type ddosProfileStoreType struct {
	mu       sync.RWMutex
	profiles map[string][]DDoSProfile
}

var ddosProfileStore = &ddosProfileStoreType{
	profiles: map[string][]DDoSProfile{
		"buyvm": {
			{ID: newID(), Name: "Edge Shield – Production", Tier: "Advanced", Mode: "always-on", Threshold: "10 Gbps", Status: "active", Mitigated7d: "3 events", Updated: "2026-01-15"},
			{ID: newID(), Name: "Edge Shield – Staging", Tier: "Standard", Mode: "on-demand", Threshold: "5 Gbps", Status: "active", Mitigated7d: "0 events", Updated: "2025-12-20"},
			{ID: newID(), Name: "L7 Rate Limiter", Tier: "Advanced", Mode: "always-on", Threshold: "50k rps", Status: "active", Mitigated7d: "12 events", Updated: "2026-02-01"},
			{ID: newID(), Name: "UDP Flood Guard", Tier: "Standard", Mode: "always-on", Threshold: "2 Gbps", Status: "active", Mitigated7d: "1 event", Updated: "2025-11-10"},
		},
		"hosting": {
			{ID: newID(), Name: "Global Shield", Tier: "Advanced", Mode: "always-on", Threshold: "20 Gbps", Status: "active", Mitigated7d: "5 events", Updated: "2026-01-20"},
			{ID: newID(), Name: "API Gateway Protection", Tier: "Standard", Mode: "always-on", Threshold: "25k rps", Status: "active", Mitigated7d: "2 events", Updated: "2026-02-10"},
			{ID: newID(), Name: "Legacy Endpoint Shield", Tier: "Basic", Mode: "on-demand", Threshold: "1 Gbps", Status: "disabled", Mitigated7d: "0 events", Updated: "2025-09-05"},
		},
	},
}

func (s *ddosProfileStoreType) list(slug string) []DDoSProfile {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.profiles[slug]
	out := make([]DDoSProfile, len(src))
	copy(out, src)
	return out
}

func (s *ddosProfileStoreType) getByID(slug, id string) (DDoSProfile, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, p := range s.profiles[slug] {
		if p.ID == id {
			return p, true
		}
	}
	return DDoSProfile{}, false
}

func (s *ddosProfileStoreType) add(slug string, p DDoSProfile) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.profiles[slug] = append(s.profiles[slug], p)
}

func (s *ddosProfileStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.profiles[slug]
	for i, p := range list {
		if p.ID == id {
			s.profiles[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiDDoSProfile struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Tier        string `json:"tier"`
	Mode        string `json:"mode"`
	Threshold   string `json:"threshold"`
	Status      string `json:"status"`
	Mitigated7d string `json:"mitigated_7d"`
	Updated     string `json:"updated"`
}

type apiDDoSProfileList struct {
	Profiles []apiDDoSProfile `json:"profiles"`
	Total    int              `json:"total"`
}

func toAPIDDoSProfile(p DDoSProfile) apiDDoSProfile {
	return apiDDoSProfile{
		ID: p.ID, Name: p.Name, Tier: p.Tier,
		Mode: p.Mode, Threshold: p.Threshold, Status: p.Status,
		Mitigated7d: p.Mitigated7d, Updated: p.Updated,
	}
}

func filterDDoSProfiles(all []DDoSProfile, search, status string) []DDoSProfile {
	var out []DDoSProfile
	q := strings.ToLower(search)
	for _, p := range all {
		if status != "" && p.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(p.Name + " " + p.Tier + " " + p.Mode)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, p)
	}
	return out
}

func apiWSSecNetDDoSListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	profiles := filterDDoSProfiles(ddosProfileStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiDDoSProfile, len(profiles))
	for i, p := range profiles {
		out[i] = toAPIDDoSProfile(p)
	}
	apiOK(w, apiDDoSProfileList{Profiles: out, Total: len(out)})
}

type apiCreateDDoSProfileRequest struct {
	Name      string `json:"name"`
	Tier      string `json:"tier"`
	Mode      string `json:"mode"`
	Threshold string `json:"threshold"`
}

func apiWSSecNetDDoSCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateDDoSProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	tier := strings.TrimSpace(req.Tier)
	if tier == "" {
		tier = "Standard"
	}
	mode := strings.TrimSpace(req.Mode)
	if mode == "" {
		mode = "always-on"
	}
	threshold := strings.TrimSpace(req.Threshold)
	if threshold == "" {
		threshold = "5 Gbps"
	}
	profile := DDoSProfile{
		ID: newID(), Name: name, Tier: tier,
		Mode: mode, Threshold: threshold, Status: "active",
		Mitigated7d: "0 events", Updated: "2026-04-17",
	}
	ddosProfileStore.add(slug, profile)
	apiCreated(w, toAPIDDoSProfile(profile))
}

func apiWSSecNetDDoSGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	profile, found := ddosProfileStore.getByID(slug, r.PathValue("did"))
	if !found {
		apiError(w, http.StatusNotFound, "DDoS profile not found.")
		return
	}
	apiOK(w, toAPIDDoSProfile(profile))
}

func apiWSSecNetDDoSDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := ddosProfileStore.getByID(slug, r.PathValue("did")); !found {
		apiError(w, http.StatusNotFound, "DDoS profile not found.")
		return
	}
	ddosProfileStore.delete(slug, r.PathValue("did"))
	apiOK(w, map[string]any{"ok": true, "message": "DDoS profile deleted."})
}
