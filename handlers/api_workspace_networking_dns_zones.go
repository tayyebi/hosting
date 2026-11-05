package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Networking API — DNS: Zones (Pro Feature)
// networking/dns/zones
// ════════════════════════════════════════════════════════════════

type ProDNSZone struct {
	ID            string
	Domain        string
	ZoneType      string
	NameServers   string
	RecordCount   int
	DNSSEC        bool
	DNSSECAlgo    string
	TTLDefault    int
	GeoRouting    bool
	ANAMEFlat     bool
	QueriesDay    int
	PropagationMs int
	Registrar     string
	ExpiresAt     string
	AutoRenew     bool
	Status        string
	Updated       string
}

type proDNSZoneStoreType struct {
	mu    sync.RWMutex
	items map[string][]ProDNSZone
}

var proDNSZoneStore = &proDNSZoneStoreType{
	items: map[string][]ProDNSZone{
		"buyvm": {
			{ID: newID(), Domain: "buyvm-demo.com", ZoneType: "Primary", NameServers: "ns1.platform.io, ns2.platform.io", RecordCount: 42, DNSSEC: true, DNSSECAlgo: "ECDSAP256SHA256", TTLDefault: 300, GeoRouting: true, ANAMEFlat: true, QueriesDay: 2400000, PropagationMs: 8, Registrar: "Platform Domains", ExpiresAt: "2027-03-15", AutoRenew: true, Status: "active", Updated: "2026-04-17"},
			{ID: newID(), Domain: "buyvm-cdn.net", ZoneType: "Primary", NameServers: "ns1.platform.io, ns2.platform.io", RecordCount: 18, DNSSEC: true, DNSSECAlgo: "ECDSAP256SHA256", TTLDefault: 60, GeoRouting: true, ANAMEFlat: false, QueriesDay: 8500000, PropagationMs: 5, Registrar: "Namecheap", ExpiresAt: "2027-01-20", AutoRenew: true, Status: "active", Updated: "2026-04-17"},
			{ID: newID(), Domain: "api.buyvm-demo.com", ZoneType: "Delegated", NameServers: "ns1.platform.io, ns2.platform.io", RecordCount: 8, DNSSEC: false, DNSSECAlgo: "—", TTLDefault: 300, GeoRouting: false, ANAMEFlat: false, QueriesDay: 650000, PropagationMs: 10, Registrar: "—", ExpiresAt: "—", AutoRenew: false, Status: "active", Updated: "2026-04-16"},
			{ID: newID(), Domain: "staging.buyvm-demo.com", ZoneType: "Delegated", NameServers: "ns1.platform.io, ns2.platform.io", RecordCount: 12, DNSSEC: false, DNSSECAlgo: "—", TTLDefault: 60, GeoRouting: false, ANAMEFlat: false, QueriesDay: 45000, PropagationMs: 12, Registrar: "—", ExpiresAt: "—", AutoRenew: false, Status: "active", Updated: "2026-04-14"},
			{ID: newID(), Domain: "old-brand.com", ZoneType: "Primary", NameServers: "ns1.platform.io, ns2.platform.io", RecordCount: 3, DNSSEC: false, DNSSECAlgo: "—", TTLDefault: 3600, GeoRouting: false, ANAMEFlat: false, QueriesDay: 1200, PropagationMs: 15, Registrar: "GoDaddy", ExpiresAt: "2026-06-01", AutoRenew: false, Status: "expiring_soon", Updated: "2026-04-10"},
		},
		"hosting": {
			{ID: newID(), Domain: "hosting-demo.com", ZoneType: "Primary", NameServers: "ns1.platform.io, ns2.platform.io, ns3.platform.io", RecordCount: 65, DNSSEC: true, DNSSECAlgo: "ECDSAP384SHA384", TTLDefault: 300, GeoRouting: true, ANAMEFlat: true, QueriesDay: 12000000, PropagationMs: 4, Registrar: "Platform Domains", ExpiresAt: "2028-01-10", AutoRenew: true, Status: "active", Updated: "2026-04-17"},
			{ID: newID(), Domain: "hosting.io", ZoneType: "Primary", NameServers: "ns1.platform.io, ns2.platform.io", RecordCount: 28, DNSSEC: true, DNSSECAlgo: "ECDSAP256SHA256", TTLDefault: 300, GeoRouting: true, ANAMEFlat: true, QueriesDay: 4200000, PropagationMs: 6, Registrar: "Cloudflare Registrar", ExpiresAt: "2027-09-05", AutoRenew: true, Status: "active", Updated: "2026-04-17"},
			{ID: newID(), Domain: "hosting-api.com", ZoneType: "Primary", NameServers: "ns1.platform.io, ns2.platform.io", RecordCount: 15, DNSSEC: true, DNSSECAlgo: "ECDSAP256SHA256", TTLDefault: 60, GeoRouting: false, ANAMEFlat: false, QueriesDay: 1800000, PropagationMs: 8, Registrar: "Platform Domains", ExpiresAt: "2027-06-22", AutoRenew: true, Status: "active", Updated: "2026-04-16"},
			{ID: newID(), Domain: "cz-internal.net", ZoneType: "Private", NameServers: "ns-internal.platform.io", RecordCount: 95, DNSSEC: false, DNSSECAlgo: "—", TTLDefault: 30, GeoRouting: false, ANAMEFlat: false, QueriesDay: 320000, PropagationMs: 2, Registrar: "—", ExpiresAt: "—", AutoRenew: false, Status: "active", Updated: "2026-04-17"},
		},
	},
}

func (s *proDNSZoneStoreType) list(slug string) []ProDNSZone {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.items[slug]
	out := make([]ProDNSZone, len(src))
	copy(out, src)
	return out
}

func (s *proDNSZoneStoreType) getByID(slug, id string) (ProDNSZone, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, z := range s.items[slug] {
		if z.ID == id {
			return z, true
		}
	}
	return ProDNSZone{}, false
}

func (s *proDNSZoneStoreType) add(slug string, z ProDNSZone) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[slug] = append(s.items[slug], z)
}

func (s *proDNSZoneStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.items[slug]
	for i, z := range list {
		if z.ID == id {
			s.items[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiProDNSZone struct {
	ID            string `json:"id"`
	Domain        string `json:"domain"`
	ZoneType      string `json:"zone_type"`
	NameServers   string `json:"name_servers"`
	RecordCount   int    `json:"record_count"`
	DNSSEC        bool   `json:"dnssec"`
	DNSSECAlgo    string `json:"dnssec_algorithm"`
	TTLDefault    int    `json:"ttl_default"`
	GeoRouting    bool   `json:"geo_routing"`
	ANAMEFlat     bool   `json:"aname_flattening"`
	QueriesDay    int    `json:"queries_per_day"`
	PropagationMs int    `json:"propagation_ms"`
	Registrar     string `json:"registrar"`
	ExpiresAt     string `json:"expires_at"`
	AutoRenew     bool   `json:"auto_renew"`
	Status        string `json:"status"`
	Updated       string `json:"updated"`
}

type apiProDNSZoneList struct {
	Zones []apiProDNSZone `json:"zones"`
	Total int             `json:"total"`
}

func toAPIProDNSZone(z ProDNSZone) apiProDNSZone {
	return apiProDNSZone{
		ID: z.ID, Domain: z.Domain, ZoneType: z.ZoneType,
		NameServers: z.NameServers, RecordCount: z.RecordCount,
		DNSSEC: z.DNSSEC, DNSSECAlgo: z.DNSSECAlgo,
		TTLDefault: z.TTLDefault, GeoRouting: z.GeoRouting,
		ANAMEFlat: z.ANAMEFlat, QueriesDay: z.QueriesDay,
		PropagationMs: z.PropagationMs, Registrar: z.Registrar,
		ExpiresAt: z.ExpiresAt, AutoRenew: z.AutoRenew,
		Status: z.Status, Updated: z.Updated,
	}
}

func filterProDNSZones(all []ProDNSZone, search, status string) []ProDNSZone {
	var out []ProDNSZone
	q := strings.ToLower(search)
	for _, z := range all {
		if status != "" && z.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(z.Domain + " " + z.ZoneType + " " + z.Registrar)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, z)
	}
	return out
}

func apiWSNetDNSZonesListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	items := filterProDNSZones(proDNSZoneStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiProDNSZone, len(items))
	for i, z := range items {
		out[i] = toAPIProDNSZone(z)
	}
	apiOK(w, apiProDNSZoneList{Zones: out, Total: len(out)})
}

type apiCreateProDNSZoneRequest struct {
	Domain   string `json:"domain"`
	ZoneType string `json:"zone_type"`
	DNSSEC   bool   `json:"dnssec"`
	TTL      int    `json:"ttl"`
}

func apiWSNetDNSZonesCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateProDNSZoneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	domain := strings.TrimSpace(req.Domain)
	if domain == "" {
		apiError(w, http.StatusBadRequest, "Domain is required.")
		return
	}
	zoneType := strings.TrimSpace(req.ZoneType)
	if zoneType == "" {
		zoneType = "Primary"
	}
	ttl := req.TTL
	if ttl <= 0 {
		ttl = 300
	}
	dnssecAlgo := "—"
	if req.DNSSEC {
		dnssecAlgo = "ECDSAP256SHA256"
	}
	zone := ProDNSZone{
		ID: newID(), Domain: domain, ZoneType: zoneType,
		NameServers: "ns1.platform.io, ns2.platform.io",
		RecordCount: 0, DNSSEC: req.DNSSEC, DNSSECAlgo: dnssecAlgo,
		TTLDefault: ttl, GeoRouting: false, ANAMEFlat: false,
		QueriesDay: 0, PropagationMs: 0, Registrar: "—",
		ExpiresAt: "—", AutoRenew: false,
		Status: "pending_ns", Updated: "2026-04-18",
	}
	proDNSZoneStore.add(slug, zone)
	apiCreated(w, toAPIProDNSZone(zone))
}

func apiWSNetDNSZonesGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	zone, found := proDNSZoneStore.getByID(slug, r.PathValue("zid"))
	if !found {
		apiError(w, http.StatusNotFound, "DNS zone not found.")
		return
	}
	apiOK(w, toAPIProDNSZone(zone))
}

func apiWSNetDNSZonesDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := proDNSZoneStore.getByID(slug, r.PathValue("zid")); !found {
		apiError(w, http.StatusNotFound, "DNS zone not found.")
		return
	}
	proDNSZoneStore.delete(slug, r.PathValue("zid"))
	apiOK(w, map[string]any{"ok": true, "message": "DNS zone deleted."})
}
