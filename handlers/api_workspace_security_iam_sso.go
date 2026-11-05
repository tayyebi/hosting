package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Security API — IAM: SSO / Federation
// security/iam/sso
// ════════════════════════════════════════════════════════════════

type SSOProvider struct {
	ID       string
	Name     string
	Protocol string
	Status   string
	Users    int
	Domains  string
	MFA      string
	LastUsed string
	Created  string
	Region   string
}

type ssoProviderStoreType struct {
	mu        sync.RWMutex
	providers map[string][]SSOProvider
}

var ssoProviderStore = &ssoProviderStoreType{
	providers: map[string][]SSOProvider{},
}

func (s *ssoProviderStoreType) list(slug string) []SSOProvider {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.providers[slug]
	out := make([]SSOProvider, len(src))
	copy(out, src)
	return out
}

func (s *ssoProviderStoreType) getByID(slug, id string) (SSOProvider, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, p := range s.providers[slug] {
		if p.ID == id {
			return p, true
		}
	}
	return SSOProvider{}, false
}

func (s *ssoProviderStoreType) add(slug string, p SSOProvider) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.providers[slug] = append(s.providers[slug], p)
}

func (s *ssoProviderStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.providers[slug]
	for i, p := range list {
		if p.ID == id {
			s.providers[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiSSOProvider struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Protocol string `json:"protocol"`
	Status   string `json:"status"`
	Users    int    `json:"users"`
	Domains  string `json:"domains"`
	MFA      string `json:"mfa"`
	LastUsed string `json:"last_used"`
	Created  string `json:"created"`
	Region   string `json:"region"`
}

type apiSSOProviderList struct {
	Providers []apiSSOProvider `json:"providers"`
	Total     int              `json:"total"`
}

func toAPISSOProvider(p SSOProvider) apiSSOProvider {
	return apiSSOProvider{
		ID: p.ID, Name: p.Name, Protocol: p.Protocol,
		Status: p.Status, Users: p.Users,
		Domains: p.Domains, MFA: p.MFA,
		LastUsed: p.LastUsed, Created: p.Created,
		Region: p.Region,
	}
}

func filterSSOProviders(all []SSOProvider, search, status string) []SSOProvider {
	var out []SSOProvider
	q := strings.ToLower(search)
	for _, p := range all {
		if status != "" && p.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(p.Name + " " + p.Protocol + " " + p.Domains + " " + p.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, p)
	}
	return out
}

func apiWSSecIAMSSOListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	providers := filterSSOProviders(ssoProviderStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiSSOProvider, len(providers))
	for i, p := range providers {
		out[i] = toAPISSOProvider(p)
	}
	apiOK(w, apiSSOProviderList{Providers: out, Total: len(out)})
}

type apiCreateSSOProviderRequest struct {
	Name     string `json:"name"`
	Protocol string `json:"protocol"`
	Domains  string `json:"domains"`
	Region   string `json:"region"`
}

func apiWSSecIAMSSOCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateSSOProviderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	protocol := strings.TrimSpace(req.Protocol)
	if protocol == "" {
		protocol = "OIDC"
	}
	domains := strings.TrimSpace(req.Domains)
	if domains == "" {
		apiError(w, http.StatusBadRequest, "Domains is required.")
		return
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "eu-central-1"
	}
	provider := SSOProvider{
		ID: newID(), Name: name, Protocol: protocol,
		Status: "configuring", Users: 0,
		Domains: domains, MFA: "optional",
		LastUsed: "—", Created: "2026-04-17",
		Region: region,
	}
	ssoProviderStore.add(slug, provider)
	apiCreated(w, toAPISSOProvider(provider))
}

func apiWSSecIAMSSOGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	provider, found := ssoProviderStore.getByID(slug, r.PathValue("pid"))
	if !found {
		apiError(w, http.StatusNotFound, "SSO provider not found.")
		return
	}
	apiOK(w, toAPISSOProvider(provider))
}

func apiWSSecIAMSSODeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := ssoProviderStore.getByID(slug, r.PathValue("pid")); !found {
		apiError(w, http.StatusNotFound, "SSO provider not found.")
		return
	}
	ssoProviderStore.delete(slug, r.PathValue("pid"))
	apiOK(w, map[string]any{"ok": true, "message": "SSO provider deleted."})
}
