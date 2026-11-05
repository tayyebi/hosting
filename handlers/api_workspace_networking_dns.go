package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Networking API — DNS Management
// networking/dns/manage
// ════════════════════════════════════════════════════════════════

type DNSZone struct {
	ID        string
	Name      string
	Type      string
	Region    string
	Status    string
	Records   int
	TTL       string
	DNSSEC    string
	Nameservers string
}

type dnsZoneStoreType struct {
	mu    sync.RWMutex
	zones map[string][]DNSZone
}

var dnsZoneStore = &dnsZoneStoreType{
	zones: map[string][]DNSZone{
		"buyvm": {
			{ID: newID(), Name: "example.com", Type: "public", Region: "global", Status: "active", Records: 48, TTL: "300", DNSSEC: "enabled", Nameservers: "ns1.cloud.example.com, ns2.cloud.example.com"},
			{ID: newID(), Name: "api.example.com", Type: "public", Region: "global", Status: "active", Records: 12, TTL: "60", DNSSEC: "enabled", Nameservers: "ns1.cloud.example.com, ns2.cloud.example.com"},
			{ID: newID(), Name: "internal.corp", Type: "private", Region: "eu-central-1", Status: "active", Records: 86, TTL: "300", DNSSEC: "disabled", Nameservers: "ns-priv-1.eu-central-1.example.com"},
			{ID: newID(), Name: "staging.example.com", Type: "public", Region: "global", Status: "active", Records: 22, TTL: "300", DNSSEC: "disabled", Nameservers: "ns1.cloud.example.com, ns2.cloud.example.com"},
			{ID: newID(), Name: "legacy.oldsite.io", Type: "public", Region: "global", Status: "degraded", Records: 8, TTL: "3600", DNSSEC: "disabled", Nameservers: "ns1.cloud.example.com, ns2.cloud.example.com"},
		},
		"hosting": {
			{ID: newID(), Name: "hosting.app", Type: "public", Region: "global", Status: "active", Records: 34, TTL: "300", DNSSEC: "enabled", Nameservers: "ns1.cloud.example.com, ns2.cloud.example.com"},
			{ID: newID(), Name: "shop.hosting.app", Type: "public", Region: "global", Status: "active", Records: 14, TTL: "60", DNSSEC: "enabled", Nameservers: "ns1.cloud.example.com, ns2.cloud.example.com"},
			{ID: newID(), Name: "priv.hosting.internal", Type: "private", Region: "ap-south-1", Status: "active", Records: 42, TTL: "300", DNSSEC: "disabled", Nameservers: "ns-priv-1.ap-south-1.example.com"},
		},
	},
}

func (s *dnsZoneStoreType) list(slug string) []DNSZone {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.zones[slug]
	out := make([]DNSZone, len(src))
	copy(out, src)
	return out
}

func (s *dnsZoneStoreType) getByID(slug, id string) (DNSZone, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, z := range s.zones[slug] {
		if z.ID == id {
			return z, true
		}
	}
	return DNSZone{}, false
}

func (s *dnsZoneStoreType) add(slug string, z DNSZone) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.zones[slug] = append(s.zones[slug], z)
}

func (s *dnsZoneStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.zones[slug]
	for i, z := range list {
		if z.ID == id {
			s.zones[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiDNSZone struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Region      string `json:"region"`
	Status      string `json:"status"`
	Records     int    `json:"records"`
	TTL         string `json:"ttl"`
	DNSSEC      string `json:"dnssec"`
	Nameservers string `json:"nameservers"`
}

type apiDNSZoneList struct {
	Zones []apiDNSZone `json:"zones"`
	Total int          `json:"total"`
}

func toAPIDNSZone(z DNSZone) apiDNSZone {
	return apiDNSZone{
		ID: z.ID, Name: z.Name, Type: z.Type,
		Region: z.Region, Status: z.Status,
		Records: z.Records, TTL: z.TTL,
		DNSSEC: z.DNSSEC, Nameservers: z.Nameservers,
	}
}

func filterDNSZones(all []DNSZone, search, status string) []DNSZone {
	var out []DNSZone
	q := strings.ToLower(search)
	for _, z := range all {
		if status != "" && z.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(z.Name + " " + z.Type + " " + z.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, z)
	}
	return out
}

func apiWSNetDNSManageListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	zones := filterDNSZones(dnsZoneStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiDNSZone, len(zones))
	for i, z := range zones {
		out[i] = toAPIDNSZone(z)
	}
	apiOK(w, apiDNSZoneList{Zones: out, Total: len(out)})
}

type apiCreateDNSZoneRequest struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Region string `json:"region"`
	TTL    string `json:"ttl"`
	DNSSEC string `json:"dnssec"`
}

func apiWSNetDNSManageCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateDNSZoneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	zoneType := strings.TrimSpace(req.Type)
	if zoneType == "" {
		zoneType = "public"
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		if zoneType == "private" {
			region = "eu-central-1"
		} else {
			region = "global"
		}
	}
	ttl := strings.TrimSpace(req.TTL)
	if ttl == "" {
		ttl = "300"
	}
	dnssec := strings.TrimSpace(req.DNSSEC)
	if dnssec == "" {
		dnssec = "disabled"
	}
	ns := "ns1.cloud.example.com, ns2.cloud.example.com"
	if zoneType == "private" {
		ns = "ns-priv-1." + region + ".example.com"
	}
	zone := DNSZone{
		ID: newID(), Name: name, Type: zoneType,
		Region: region, Status: "active",
		Records: 0, TTL: ttl, DNSSEC: dnssec,
		Nameservers: ns,
	}
	dnsZoneStore.add(slug, zone)
	apiCreated(w, toAPIDNSZone(zone))
}

func apiWSNetDNSManageGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	zone, found := dnsZoneStore.getByID(slug, r.PathValue("zid"))
	if !found {
		apiError(w, http.StatusNotFound, "DNS zone not found.")
		return
	}
	apiOK(w, toAPIDNSZone(zone))
}

func apiWSNetDNSManageDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := dnsZoneStore.getByID(slug, r.PathValue("zid")); !found {
		apiError(w, http.StatusNotFound, "DNS zone not found.")
		return
	}
	dnsZoneStore.delete(slug, r.PathValue("zid"))
	apiOK(w, map[string]any{"ok": true, "message": "DNS zone deleted."})
}

// ════════════════════════════════════════════════════════════════
// Workspace Networking API — Domain Registration
// networking/dns/domains
// ════════════════════════════════════════════════════════════════

type DomainRegistration struct {
	ID         string
	Domain     string
	TLD        string
	Status     string
	Registrar  string
	ExpiresAt  string
	AutoRenew  bool
	Privacy    bool
	Nameservers string
	CreatedAt  string
}

type domainRegistrationStoreType struct {
	mu      sync.RWMutex
	domains map[string][]DomainRegistration
}

var domainRegistrationStore = &domainRegistrationStoreType{
	domains: map[string][]DomainRegistration{
		"buyvm": {
			{ID: newID(), Domain: "example.com", TLD: ".com", Status: "active", Registrar: "CloudRegistrar", ExpiresAt: "2027-03-15", AutoRenew: true, Privacy: true, Nameservers: "ns1.cloud.example.com, ns2.cloud.example.com", CreatedAt: "2020-03-15"},
			{ID: newID(), Domain: "example.io", TLD: ".io", Status: "active", Registrar: "CloudRegistrar", ExpiresAt: "2027-06-01", AutoRenew: true, Privacy: true, Nameservers: "ns1.cloud.example.com, ns2.cloud.example.com", CreatedAt: "2021-06-01"},
			{ID: newID(), Domain: "myapp.dev", TLD: ".dev", Status: "active", Registrar: "CloudRegistrar", ExpiresAt: "2026-11-20", AutoRenew: true, Privacy: true, Nameservers: "ns1.cloud.example.com, ns2.cloud.example.com", CreatedAt: "2022-11-20"},
			{ID: newID(), Domain: "oldsite.io", TLD: ".io", Status: "expiring", Registrar: "CloudRegistrar", ExpiresAt: "2026-05-01", AutoRenew: false, Privacy: false, Nameservers: "ns1.cloud.example.com, ns2.cloud.example.com", CreatedAt: "2019-05-01"},
		},
		"hosting": {
			{ID: newID(), Domain: "hosting.app", TLD: ".app", Status: "active", Registrar: "CloudRegistrar", ExpiresAt: "2028-01-10", AutoRenew: true, Privacy: true, Nameservers: "ns1.cloud.example.com, ns2.cloud.example.com", CreatedAt: "2022-01-10"},
			{ID: newID(), Domain: "hosting.com", TLD: ".com", Status: "active", Registrar: "CloudRegistrar", ExpiresAt: "2027-08-22", AutoRenew: true, Privacy: true, Nameservers: "ns1.cloud.example.com, ns2.cloud.example.com", CreatedAt: "2020-08-22"},
			{ID: newID(), Domain: "shop-hosting.store", TLD: ".store", Status: "active", Registrar: "CloudRegistrar", ExpiresAt: "2026-12-05", AutoRenew: true, Privacy: false, Nameservers: "ns1.cloud.example.com, ns2.cloud.example.com", CreatedAt: "2023-12-05"},
		},
	},
}

func (s *domainRegistrationStoreType) list(slug string) []DomainRegistration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.domains[slug]
	out := make([]DomainRegistration, len(src))
	copy(out, src)
	return out
}

func (s *domainRegistrationStoreType) getByID(slug, id string) (DomainRegistration, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, d := range s.domains[slug] {
		if d.ID == id {
			return d, true
		}
	}
	return DomainRegistration{}, false
}

func (s *domainRegistrationStoreType) add(slug string, d DomainRegistration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.domains[slug] = append(s.domains[slug], d)
}

func (s *domainRegistrationStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.domains[slug]
	for i, d := range list {
		if d.ID == id {
			s.domains[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiDomainRegistration struct {
	ID          string `json:"id"`
	Domain      string `json:"domain"`
	TLD         string `json:"tld"`
	Status      string `json:"status"`
	Registrar   string `json:"registrar"`
	ExpiresAt   string `json:"expires_at"`
	AutoRenew   bool   `json:"auto_renew"`
	Privacy     bool   `json:"privacy"`
	Nameservers string `json:"nameservers"`
	CreatedAt   string `json:"created_at"`
}

type apiDomainRegistrationList struct {
	Domains []apiDomainRegistration `json:"domains"`
	Total   int                     `json:"total"`
}

func toAPIDomainRegistration(d DomainRegistration) apiDomainRegistration {
	return apiDomainRegistration{
		ID: d.ID, Domain: d.Domain, TLD: d.TLD,
		Status: d.Status, Registrar: d.Registrar,
		ExpiresAt: d.ExpiresAt, AutoRenew: d.AutoRenew,
		Privacy: d.Privacy, Nameservers: d.Nameservers,
		CreatedAt: d.CreatedAt,
	}
}

func filterDomainRegistrations(all []DomainRegistration, search, status string) []DomainRegistration {
	var out []DomainRegistration
	q := strings.ToLower(search)
	for _, d := range all {
		if status != "" && d.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(d.Domain + " " + d.TLD + " " + d.Registrar)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, d)
	}
	return out
}

func apiWSNetDNSDomainsListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	domains := filterDomainRegistrations(domainRegistrationStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiDomainRegistration, len(domains))
	for i, d := range domains {
		out[i] = toAPIDomainRegistration(d)
	}
	apiOK(w, apiDomainRegistrationList{Domains: out, Total: len(out)})
}

type apiCreateDomainRegistrationRequest struct {
	Domain    string `json:"domain"`
	AutoRenew bool   `json:"auto_renew"`
	Privacy   bool   `json:"privacy"`
}

func apiWSNetDNSDomainsCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateDomainRegistrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	domain := strings.TrimSpace(req.Domain)
	if domain == "" {
		apiError(w, http.StatusBadRequest, "Domain is required.")
		return
	}
	tld := ".com"
	if idx := strings.LastIndex(domain, "."); idx >= 0 {
		tld = domain[idx:]
	}
	d := DomainRegistration{
		ID: newID(), Domain: domain, TLD: tld,
		Status: "pending", Registrar: "CloudRegistrar",
		ExpiresAt: "2027-04-17", AutoRenew: req.AutoRenew,
		Privacy: req.Privacy,
		Nameservers: "ns1.cloud.example.com, ns2.cloud.example.com",
		CreatedAt: "2026-04-17",
	}
	domainRegistrationStore.add(slug, d)
	apiCreated(w, toAPIDomainRegistration(d))
}

func apiWSNetDNSDomainsGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	d, found := domainRegistrationStore.getByID(slug, r.PathValue("did"))
	if !found {
		apiError(w, http.StatusNotFound, "Domain not found.")
		return
	}
	apiOK(w, toAPIDomainRegistration(d))
}

func apiWSNetDNSDomainsDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := domainRegistrationStore.getByID(slug, r.PathValue("did")); !found {
		apiError(w, http.StatusNotFound, "Domain not found.")
		return
	}
	domainRegistrationStore.delete(slug, r.PathValue("did"))
	apiOK(w, map[string]any{"ok": true, "message": "Domain registration deleted."})
}
