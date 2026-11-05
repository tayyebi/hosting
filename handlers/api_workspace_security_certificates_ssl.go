package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Security API — Certificates: SSL/TLS
// security/certificates/ssl-tls
// ════════════════════════════════════════════════════════════════

type SSLCertificate struct {
	ID            string
	Domain        string
	SAN           string
	Issuer        string
	Type          string
	Status        string
	AutoRenew     bool
	IssuedAt      string
	ExpiresAt     string
	KeyAlgorithm  string
	KeySize       int
	Protocol      string
	OCSPStapling  bool
	CTLogged      bool
	Updated       string
}

type sslCertStoreType struct {
	mu    sync.RWMutex
	items map[string][]SSLCertificate
}

var sslCertStore = &sslCertStoreType{
	items: map[string][]SSLCertificate{
		"buyvm": {
			{ID: newID(), Domain: "*.buyvm-demo.com", SAN: "buyvm-demo.com, *.buyvm-demo.com", Issuer: "Let's Encrypt R3", Type: "DV Wildcard", Status: "active", AutoRenew: true, IssuedAt: "2026-03-18", ExpiresAt: "2026-06-16", KeyAlgorithm: "ECDSA", KeySize: 256, Protocol: "TLS 1.3", OCSPStapling: true, CTLogged: true, Updated: "2026-04-17"},
			{ID: newID(), Domain: "api.buyvm-demo.com", SAN: "api.buyvm-demo.com", Issuer: "Let's Encrypt R3", Type: "DV", Status: "active", AutoRenew: true, IssuedAt: "2026-04-01", ExpiresAt: "2026-06-30", KeyAlgorithm: "ECDSA", KeySize: 256, Protocol: "TLS 1.3", OCSPStapling: true, CTLogged: true, Updated: "2026-04-17"},
			{ID: newID(), Domain: "cdn.buyvm-demo.com", SAN: "cdn.buyvm-demo.com, static.buyvm-demo.com", Issuer: "DigiCert G2", Type: "OV SAN", Status: "active", AutoRenew: true, IssuedAt: "2026-01-15", ExpiresAt: "2027-01-15", KeyAlgorithm: "RSA", KeySize: 2048, Protocol: "TLS 1.2", OCSPStapling: true, CTLogged: true, Updated: "2026-04-10"},
			{ID: newID(), Domain: "staging.buyvm-demo.com", SAN: "staging.buyvm-demo.com", Issuer: "Let's Encrypt R3", Type: "DV", Status: "expiring_soon", AutoRenew: false, IssuedAt: "2026-01-20", ExpiresAt: "2026-04-20", KeyAlgorithm: "ECDSA", KeySize: 256, Protocol: "TLS 1.3", OCSPStapling: true, CTLogged: true, Updated: "2026-04-15"},
			{ID: newID(), Domain: "legacy.buyvm-demo.com", SAN: "legacy.buyvm-demo.com", Issuer: "Comodo RSA CA", Type: "DV", Status: "expired", AutoRenew: false, IssuedAt: "2025-04-01", ExpiresAt: "2026-04-01", KeyAlgorithm: "RSA", KeySize: 2048, Protocol: "TLS 1.2", OCSPStapling: false, CTLogged: true, Updated: "2026-04-01"},
		},
		"hosting": {
			{ID: newID(), Domain: "*.hosting-demo.com", SAN: "hosting-demo.com, *.hosting-demo.com", Issuer: "Let's Encrypt R3", Type: "DV Wildcard", Status: "active", AutoRenew: true, IssuedAt: "2026-03-25", ExpiresAt: "2026-06-23", KeyAlgorithm: "ECDSA", KeySize: 256, Protocol: "TLS 1.3", OCSPStapling: true, CTLogged: true, Updated: "2026-04-17"},
			{ID: newID(), Domain: "portal.hosting-demo.com", SAN: "portal.hosting-demo.com, app.hosting-demo.com", Issuer: "DigiCert G2", Type: "OV SAN", Status: "active", AutoRenew: true, IssuedAt: "2026-02-10", ExpiresAt: "2027-02-10", KeyAlgorithm: "RSA", KeySize: 4096, Protocol: "TLS 1.3", OCSPStapling: true, CTLogged: true, Updated: "2026-04-16"},
			{ID: newID(), Domain: "api.hosting-demo.com", SAN: "api.hosting-demo.com, graphql.hosting-demo.com", Issuer: "Let's Encrypt R3", Type: "DV SAN", Status: "active", AutoRenew: true, IssuedAt: "2026-04-05", ExpiresAt: "2026-07-04", KeyAlgorithm: "ECDSA", KeySize: 256, Protocol: "TLS 1.3", OCSPStapling: true, CTLogged: true, Updated: "2026-04-17"},
			{ID: newID(), Domain: "payments.hosting-demo.com", SAN: "payments.hosting-demo.com", Issuer: "GlobalSign EV", Type: "EV", Status: "active", AutoRenew: true, IssuedAt: "2026-01-01", ExpiresAt: "2027-01-01", KeyAlgorithm: "RSA", KeySize: 4096, Protocol: "TLS 1.3", OCSPStapling: true, CTLogged: true, Updated: "2026-04-12"},
		},
	},
}

func (s *sslCertStoreType) list(slug string) []SSLCertificate {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.items[slug]
	out := make([]SSLCertificate, len(src))
	copy(out, src)
	return out
}

func (s *sslCertStoreType) getByID(slug, id string) (SSLCertificate, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, c := range s.items[slug] {
		if c.ID == id {
			return c, true
		}
	}
	return SSLCertificate{}, false
}

func (s *sslCertStoreType) add(slug string, c SSLCertificate) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[slug] = append(s.items[slug], c)
}

func (s *sslCertStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.items[slug]
	for i, c := range list {
		if c.ID == id {
			s.items[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiSSLCertificate struct {
	ID            string `json:"id"`
	Domain        string `json:"domain"`
	SAN           string `json:"san"`
	Issuer        string `json:"issuer"`
	Type          string `json:"type"`
	Status        string `json:"status"`
	AutoRenew     bool   `json:"auto_renew"`
	IssuedAt      string `json:"issued_at"`
	ExpiresAt     string `json:"expires_at"`
	KeyAlgorithm  string `json:"key_algorithm"`
	KeySize       int    `json:"key_size"`
	Protocol      string `json:"protocol"`
	OCSPStapling  bool   `json:"ocsp_stapling"`
	CTLogged      bool   `json:"ct_logged"`
	Updated       string `json:"updated"`
}

type apiSSLCertificateList struct {
	Certificates []apiSSLCertificate `json:"certificates"`
	Total        int                 `json:"total"`
}

func toAPISSLCertificate(c SSLCertificate) apiSSLCertificate {
	return apiSSLCertificate{
		ID: c.ID, Domain: c.Domain, SAN: c.SAN,
		Issuer: c.Issuer, Type: c.Type, Status: c.Status,
		AutoRenew: c.AutoRenew, IssuedAt: c.IssuedAt,
		ExpiresAt: c.ExpiresAt, KeyAlgorithm: c.KeyAlgorithm,
		KeySize: c.KeySize, Protocol: c.Protocol,
		OCSPStapling: c.OCSPStapling, CTLogged: c.CTLogged,
		Updated: c.Updated,
	}
}

func filterSSLCertificates(all []SSLCertificate, search, status string) []SSLCertificate {
	var out []SSLCertificate
	q := strings.ToLower(search)
	for _, c := range all {
		if status != "" && c.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(c.Domain + " " + c.SAN + " " + c.Issuer + " " + c.Type)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, c)
	}
	return out
}

func apiWSSecCertSSLListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	items := filterSSLCertificates(sslCertStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiSSLCertificate, len(items))
	for i, c := range items {
		out[i] = toAPISSLCertificate(c)
	}
	apiOK(w, apiSSLCertificateList{Certificates: out, Total: len(out)})
}

type apiCreateSSLCertRequest struct {
	Domain       string `json:"domain"`
	Type         string `json:"type"`
	KeyAlgorithm string `json:"key_algorithm"`
	AutoRenew    bool   `json:"auto_renew"`
}

func apiWSSecCertSSLCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateSSLCertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	domain := strings.TrimSpace(req.Domain)
	if domain == "" {
		apiError(w, http.StatusBadRequest, "Domain is required.")
		return
	}
	certType := strings.TrimSpace(req.Type)
	if certType == "" {
		certType = "DV"
	}
	keyAlg := strings.TrimSpace(req.KeyAlgorithm)
	if keyAlg == "" {
		keyAlg = "ECDSA"
	}
	keySize := 256
	if keyAlg == "RSA" {
		keySize = 2048
	}
	cert := SSLCertificate{
		ID: newID(), Domain: domain, SAN: domain,
		Issuer: "Let's Encrypt R3", Type: certType,
		Status: "pending_validation", AutoRenew: req.AutoRenew,
		IssuedAt: "—", ExpiresAt: "—",
		KeyAlgorithm: keyAlg, KeySize: keySize,
		Protocol: "TLS 1.3", OCSPStapling: true,
		CTLogged: true, Updated: "2026-04-18",
	}
	sslCertStore.add(slug, cert)
	apiCreated(w, toAPISSLCertificate(cert))
}

func apiWSSecCertSSLGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	cert, found := sslCertStore.getByID(slug, r.PathValue("cid"))
	if !found {
		apiError(w, http.StatusNotFound, "SSL certificate not found.")
		return
	}
	apiOK(w, toAPISSLCertificate(cert))
}

func apiWSSecCertSSLDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := sslCertStore.getByID(slug, r.PathValue("cid")); !found {
		apiError(w, http.StatusNotFound, "SSL certificate not found.")
		return
	}
	sslCertStore.delete(slug, r.PathValue("cid"))
	apiOK(w, map[string]any{"ok": true, "message": "SSL certificate deleted."})
}
