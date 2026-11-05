package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Security API — IAM: Identity Platform (Pro Feature)
// security/iam/identity-platform
// ════════════════════════════════════════════════════════════════

type ProIdentityProvider struct {
	ID              string
	Name            string
	Protocol        string // "SAML 2.0", "OIDC", "LDAP", "SCIM"
	Issuer          string
	ClientID        string
	Domain          string
	UserCount       int
	GroupCount       int
	ActiveSessions  int
	MFAEnforced     bool
	MFAMethod       string // "TOTP", "WebAuthn", "SMS", "Push"
	PasswordPolicy  string // "strong", "standard", "custom"
	SessionTimeout  int    // minutes
	JITProvisioning bool
	Region          string
	CostPerMonth    float64
	Status          string // "active", "degraded", "configuring"
	Updated         string
}

type proIdentityStoreType struct {
	mu    sync.RWMutex
	items map[string][]ProIdentityProvider
}

var proIdentityStore = &proIdentityStoreType{
	items: map[string][]ProIdentityProvider{
		"buyvm": {
			{ID: newID(), Name: "corp-okta-sso", Protocol: "SAML 2.0", Issuer: "https://company.okta.com", ClientID: "0oa8k2x...", Domain: "company.io", UserCount: 2840, GroupCount: 64, ActiveSessions: 1420, MFAEnforced: true, MFAMethod: "WebAuthn", PasswordPolicy: "strong", SessionTimeout: 480, JITProvisioning: true, Region: "us-east-1", CostPerMonth: 284.00, Status: "active", Updated: "2026-04-18"},
			{ID: newID(), Name: "azure-ad-sync", Protocol: "OIDC", Issuer: "https://login.microsoftonline.com/tenant-id", ClientID: "app-id-12345", Domain: "company.onmicrosoft.com", UserCount: 1200, GroupCount: 38, ActiveSessions: 640, MFAEnforced: true, MFAMethod: "Push", PasswordPolicy: "strong", SessionTimeout: 360, JITProvisioning: true, Region: "us-east-1", CostPerMonth: 120.00, Status: "active", Updated: "2026-04-17"},
			{ID: newID(), Name: "ldap-legacy", Protocol: "LDAP", Issuer: "ldaps://ldap.internal.company.io:636", ClientID: "cn=admin,dc=company,dc=io", Domain: "internal.company.io", UserCount: 480, GroupCount: 12, ActiveSessions: 86, MFAEnforced: false, MFAMethod: "TOTP", PasswordPolicy: "standard", SessionTimeout: 120, JITProvisioning: false, Region: "us-east-1", CostPerMonth: 48.00, Status: "active", Updated: "2026-04-15"},
			{ID: newID(), Name: "scim-provisioner", Protocol: "SCIM", Issuer: "https://company.okta.com/scim/v2", ClientID: "scim-bearer-token", Domain: "company.io", UserCount: 2840, GroupCount: 64, ActiveSessions: 0, MFAEnforced: false, MFAMethod: "—", PasswordPolicy: "—", SessionTimeout: 0, JITProvisioning: true, Region: "us-east-1", CostPerMonth: 56.00, Status: "active", Updated: "2026-04-18"},
			{ID: newID(), Name: "partner-oidc", Protocol: "OIDC", Issuer: "https://auth.partner-corp.com", ClientID: "partner-client-9876", Domain: "partner-corp.com", UserCount: 320, GroupCount: 8, ActiveSessions: 42, MFAEnforced: true, MFAMethod: "TOTP", PasswordPolicy: "custom", SessionTimeout: 240, JITProvisioning: false, Region: "eu-central-1", CostPerMonth: 32.00, Status: "configuring", Updated: "2026-04-18"},
		},
		"hosting": {
			{ID: newID(), Name: "google-workspace", Protocol: "OIDC", Issuer: "https://accounts.google.com", ClientID: "gcp-client-id-abcd", Domain: "hosting.com", UserCount: 8400, GroupCount: 142, ActiveSessions: 4200, MFAEnforced: true, MFAMethod: "WebAuthn", PasswordPolicy: "strong", SessionTimeout: 720, JITProvisioning: true, Region: "eu-central-1", CostPerMonth: 840.00, Status: "active", Updated: "2026-04-18"},
			{ID: newID(), Name: "customer-sso-pool", Protocol: "SAML 2.0", Issuer: "https://sso.hosting.com/saml/metadata", ClientID: "customer-pool-001", Domain: "*.hosting.com", UserCount: 42000, GroupCount: 380, ActiveSessions: 18600, MFAEnforced: true, MFAMethod: "Push", PasswordPolicy: "strong", SessionTimeout: 480, JITProvisioning: true, Region: "eu-central-1", CostPerMonth: 4200.00, Status: "active", Updated: "2026-04-18"},
			{ID: newID(), Name: "contractor-ldap", Protocol: "LDAP", Issuer: "ldaps://ldap.hosting.com:636", ClientID: "cn=svc,dc=hosting,dc=com", Domain: "contractors.hosting.com", UserCount: 640, GroupCount: 24, ActiveSessions: 120, MFAEnforced: true, MFAMethod: "SMS", PasswordPolicy: "standard", SessionTimeout: 180, JITProvisioning: false, Region: "eu-central-1", CostPerMonth: 64.00, Status: "active", Updated: "2026-04-16"},
		},
	},
}

func (s *proIdentityStoreType) list(slug string) []ProIdentityProvider {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.items[slug]
	out := make([]ProIdentityProvider, len(src))
	copy(out, src)
	return out
}

func (s *proIdentityStoreType) getByID(slug, id string) (ProIdentityProvider, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, p := range s.items[slug] {
		if p.ID == id {
			return p, true
		}
	}
	return ProIdentityProvider{}, false
}

func (s *proIdentityStoreType) add(slug string, p ProIdentityProvider) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[slug] = append(s.items[slug], p)
}

func (s *proIdentityStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.items[slug]
	for i, p := range list {
		if p.ID == id {
			s.items[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiProIdentityProvider struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Protocol        string  `json:"protocol"`
	Issuer          string  `json:"issuer"`
	ClientID        string  `json:"client_id"`
	Domain          string  `json:"domain"`
	UserCount       int     `json:"user_count"`
	GroupCount       int     `json:"group_count"`
	ActiveSessions  int     `json:"active_sessions"`
	MFAEnforced     bool    `json:"mfa_enforced"`
	MFAMethod       string  `json:"mfa_method"`
	PasswordPolicy  string  `json:"password_policy"`
	SessionTimeout  int     `json:"session_timeout_min"`
	JITProvisioning bool    `json:"jit_provisioning"`
	Region          string  `json:"region"`
	CostPerMonth    float64 `json:"cost_per_month"`
	Status          string  `json:"status"`
	Updated         string  `json:"updated"`
}

type apiProIdentityProviderList struct {
	Providers []apiProIdentityProvider `json:"providers"`
	Total     int                      `json:"total"`
}

func toAPIProIdentity(p ProIdentityProvider) apiProIdentityProvider {
	return apiProIdentityProvider{
		ID: p.ID, Name: p.Name, Protocol: p.Protocol,
		Issuer: p.Issuer, ClientID: p.ClientID, Domain: p.Domain,
		UserCount: p.UserCount, GroupCount: p.GroupCount,
		ActiveSessions: p.ActiveSessions, MFAEnforced: p.MFAEnforced,
		MFAMethod: p.MFAMethod, PasswordPolicy: p.PasswordPolicy,
		SessionTimeout: p.SessionTimeout, JITProvisioning: p.JITProvisioning,
		Region: p.Region, CostPerMonth: p.CostPerMonth,
		Status: p.Status, Updated: p.Updated,
	}
}

func filterProIdentity(all []ProIdentityProvider, search, status string) []ProIdentityProvider {
	var out []ProIdentityProvider
	q := strings.ToLower(search)
	for _, p := range all {
		if status != "" && p.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(p.Name + " " + p.Protocol + " " + p.Domain + " " + p.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, p)
	}
	return out
}

func apiWSSecIAMIdentityListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	items := filterProIdentity(proIdentityStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiProIdentityProvider, len(items))
	for i, p := range items {
		out[i] = toAPIProIdentity(p)
	}
	apiOK(w, apiProIdentityProviderList{Providers: out, Total: len(out)})
}

type apiCreateProIdentityRequest struct {
	Name            string `json:"name"`
	Protocol        string `json:"protocol"`
	Issuer          string `json:"issuer"`
	ClientID        string `json:"client_id"`
	Domain          string `json:"domain"`
	MFAEnforced     bool   `json:"mfa_enforced"`
	MFAMethod       string `json:"mfa_method"`
	JITProvisioning bool   `json:"jit_provisioning"`
	Region          string `json:"region"`
}

func apiWSSecIAMIdentityCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateProIdentityRequest
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
	issuer := strings.TrimSpace(req.Issuer)
	if issuer == "" {
		issuer = "https://auth.example.com"
	}
	clientID := strings.TrimSpace(req.ClientID)
	if clientID == "" {
		clientID = "client-" + newID()
	}
	domain := strings.TrimSpace(req.Domain)
	if domain == "" {
		domain = "example.com"
	}
	mfaMethod := strings.TrimSpace(req.MFAMethod)
	if mfaMethod == "" {
		mfaMethod = "TOTP"
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "us-east-1"
	}
	prov := ProIdentityProvider{
		ID: newID(), Name: name, Protocol: protocol,
		Issuer: issuer, ClientID: clientID, Domain: domain,
		UserCount: 0, GroupCount: 0, ActiveSessions: 0,
		MFAEnforced: req.MFAEnforced, MFAMethod: mfaMethod,
		PasswordPolicy: "standard", SessionTimeout: 480,
		JITProvisioning: req.JITProvisioning, Region: region,
		CostPerMonth: 0.00, Status: "configuring", Updated: "2026-04-18",
	}
	proIdentityStore.add(slug, prov)
	apiCreated(w, toAPIProIdentity(prov))
}

func apiWSSecIAMIdentityGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	p, found := proIdentityStore.getByID(slug, r.PathValue("pid"))
	if !found {
		apiError(w, http.StatusNotFound, "Identity provider not found.")
		return
	}
	apiOK(w, toAPIProIdentity(p))
}

func apiWSSecIAMIdentityDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := proIdentityStore.getByID(slug, r.PathValue("pid")); !found {
		apiError(w, http.StatusNotFound, "Identity provider not found.")
		return
	}
	proIdentityStore.delete(slug, r.PathValue("pid"))
	apiOK(w, map[string]any{"ok": true, "message": "Identity provider deleted."})
}
