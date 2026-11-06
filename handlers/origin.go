package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"platform/db"
)

// OriginMode describes how an incoming HTTP request's Host was classified.
type OriginMode string

const (
	// OriginModeDefault is active when no system origin domains are enabled and
	// the request host does not match a tenant domain. Behavior is unchanged.
	OriginModeDefault OriginMode = "default"

	// OriginModeSystem is active when at least one system origin domain is
	// enabled and the request host matches an enabled system domain.
	OriginModeSystem OriginMode = "system"

	// OriginModeTenant is active when the request host matches an active
	// tenant's custom domain.
	OriginModeTenant OriginMode = "tenant"
)

// OriginContext carries the resolved origin state for the current HTTP request.
type OriginContext struct {
	Mode                OriginMode
	MatchedDomain       string
	PrimarySystemDomain string
	RequestHost         string
	Tenant              *Tenant
}

type originCtxKeyType struct{}

var originCtxKey = originCtxKeyType{}

func withOriginContext(r *http.Request, oc OriginContext) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), originCtxKey, oc))
}

func originFromContext(r *http.Request) OriginContext {
	if r == nil {
		return OriginContext{Mode: OriginModeDefault}
	}
	if v, ok := r.Context().Value(originCtxKey).(OriginContext); ok {
		return v
	}
	return OriginContext{Mode: OriginModeDefault}
}

// ── Origin Store (PostgreSQL + in-memory test store + fast cache) ────────────

type originSnapshot struct {
	settings       db.OriginSettings
	domains        []db.OriginDomain
	enabledDomains []db.OriginDomain
	fetchedAt      time.Time
}

type originStoreType struct {
	mu      sync.RWMutex
	memSeq  int64
	memSet  db.OriginSettings
	memDoms []db.OriginDomain

	cacheMu sync.RWMutex
	cache   *originSnapshot
}

var origins = &originStoreType{
	memSet: db.DefaultOriginSettings(),
}

const originCacheTTL = 3 * time.Second

func (o *originStoreType) invalidateCache() {
	o.cacheMu.Lock()
	o.cache = nil
	o.cacheMu.Unlock()
}

// ResetForTest clears in-memory origin state and caches (used by unit tests).
func (o *originStoreType) ResetForTest() {
	o.mu.Lock()
	o.memSeq = 0
	o.memSet = db.DefaultOriginSettings()
	o.memDoms = nil
	o.mu.Unlock()
	o.invalidateCache()
}

func (o *originStoreType) snapshot(ctx context.Context) originSnapshot {
	if Pools != nil {
		o.cacheMu.RLock()
		cached := o.cache
		if cached != nil && time.Since(cached.fetchedAt) < originCacheTTL {
			cp := *cached
			o.cacheMu.RUnlock()
			return cp
		}
		o.cacheMu.RUnlock()

		qCtx, cancel := context.WithTimeout(ctx, 4*time.Second)
		defer cancel()

		settings, sErr := db.GetOriginSettings(qCtx, Pools.Admin)
		if sErr != nil {
			log.Printf("origin: GetOriginSettings: %v", sErr)
			settings = db.DefaultOriginSettings()
		}
		domains, dErr := db.ListOriginDomains(qCtx, Pools.Admin)
		if dErr != nil {
			log.Printf("origin: ListOriginDomains: %v", dErr)
		}
		var enabled []db.OriginDomain
		for _, d := range domains {
			if d.Enabled {
				enabled = append(enabled, d)
			}
		}
		snap := originSnapshot{
			settings:       settings,
			domains:        domains,
			enabledDomains: enabled,
			fetchedAt:      time.Now(),
		}
		if sErr == nil && dErr == nil {
			o.cacheMu.Lock()
			o.cache = &snap
			o.cacheMu.Unlock()
		}
		return snap
	}

	o.mu.RLock()
	defer o.mu.RUnlock()
	doms := make([]db.OriginDomain, len(o.memDoms))
	copy(doms, o.memDoms)
	var enabled []db.OriginDomain
	for _, d := range doms {
		if d.Enabled {
			enabled = append(enabled, d)
		}
	}
	return originSnapshot{
		settings:       o.memSet,
		domains:        doms,
		enabledDomains: enabled,
		fetchedAt:      time.Now(),
	}
}

func (o *originStoreType) getSettings(ctx context.Context) db.OriginSettings {
	return o.snapshot(ctx).settings
}

func (o *originStoreType) updateSettings(ctx context.Context, in db.OriginSettings) (db.OriginSettings, error) {
	defer o.invalidateCache()
	if Pools != nil {
		return db.UpdateOriginSettings(ctx, Pools.Admin, in)
	}
	validated, err := db.ValidateOriginSettings(in)
	if err != nil {
		return db.OriginSettings{}, err
	}
	validated.UpdatedAt = time.Now().UTC()
	o.mu.Lock()
	o.memSet = validated
	o.mu.Unlock()
	return validated, nil
}

func (o *originStoreType) listDomains(ctx context.Context) []db.OriginDomain {
	return o.snapshot(ctx).domains
}

func (o *originStoreType) getDomain(ctx context.Context, id int64) (db.OriginDomain, bool) {
	if Pools != nil {
		d, ok, err := db.GetOriginDomain(ctx, Pools.Admin, id)
		if err != nil {
			log.Printf("origin: GetOriginDomain(%d): %v", id, err)
			return db.OriginDomain{}, false
		}
		return d, ok
	}
	o.mu.RLock()
	defer o.mu.RUnlock()
	for _, d := range o.memDoms {
		if d.ID == id {
			return d, true
		}
	}
	return db.OriginDomain{}, false
}

func (o *originStoreType) domainExists(domain string, excludeID int64) bool {
	norm, err := db.NormalizeDomain(domain)
	if err != nil {
		norm = strings.ToLower(strings.TrimSpace(domain))
	}
	if norm == "" {
		return false
	}
	if Pools != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
		defer cancel()
		exists, dbErr := db.OriginDomainExists(ctx, Pools.Admin, norm, excludeID)
		if dbErr == nil {
			return exists
		}
	}
	o.mu.RLock()
	defer o.mu.RUnlock()
	for _, d := range o.memDoms {
		if strings.EqualFold(d.Domain, norm) && (excludeID == 0 || d.ID != excludeID) {
			return true
		}
	}
	return false
}

func (o *originStoreType) addDomain(ctx context.Context, rawDomain, note string, enabled bool) (db.OriginDomain, error) {
	defer o.invalidateCache()
	if Pools != nil {
		return db.AddOriginDomain(ctx, Pools.Admin, rawDomain, note, enabled)
	}

	norm, err := db.NormalizeDomain(rawDomain)
	if err != nil {
		return db.OriginDomain{}, err
	}
	note = strings.TrimSpace(note)
	if len(note) > 200 {
		note = note[:200]
	}

	if store.domainExists(norm, "") {
		return db.OriginDomain{}, fmt.Errorf("%w: %q is already assigned to a tenant", db.ErrOriginDomainConflict, norm)
	}

	o.mu.Lock()
	defer o.mu.Unlock()
	for _, existing := range o.memDoms {
		if strings.EqualFold(existing.Domain, norm) {
			return db.OriginDomain{}, fmt.Errorf("%w: %q is already in the system origin list", db.ErrOriginDomainConflict, norm)
		}
	}
	o.memSeq++
	now := time.Now().UTC()
	entry := db.OriginDomain{
		ID:        o.memSeq,
		Domain:    norm,
		Note:      note,
		Position:  len(o.memDoms) + 1,
		Enabled:   enabled,
		CreatedAt: now,
		UpdatedAt: now,
	}
	o.memDoms = append(o.memDoms, entry)
	return entry, nil
}

func (o *originStoreType) updateDomain(ctx context.Context, id int64, note string, enabled bool) (db.OriginDomain, bool, error) {
	defer o.invalidateCache()
	if Pools != nil {
		return db.UpdateOriginDomain(ctx, Pools.Admin, id, note, enabled)
	}
	note = strings.TrimSpace(note)
	if len(note) > 200 {
		note = note[:200]
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	for i := range o.memDoms {
		if o.memDoms[i].ID == id {
			o.memDoms[i].Note = note
			o.memDoms[i].Enabled = enabled
			o.memDoms[i].UpdatedAt = time.Now().UTC()
			return o.memDoms[i], true, nil
		}
	}
	return db.OriginDomain{}, false, nil
}

func (o *originStoreType) moveDomain(ctx context.Context, id int64, direction string) error {
	defer o.invalidateCache()
	if Pools != nil {
		return db.MoveOriginDomain(ctx, Pools.Admin, id, direction)
	}
	direction = strings.ToLower(strings.TrimSpace(direction))
	if direction != "top" && direction != "up" && direction != "down" {
		return fmt.Errorf("invalid move direction %q", direction)
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	idx := -1
	for i, d := range o.memDoms {
		if d.ID == id {
			idx = i
			break
		}
	}
	if idx == -1 {
		return errors.New("domain not found")
	}
	switch direction {
	case "top":
		if idx > 0 {
			item := o.memDoms[idx]
			copy(o.memDoms[1:idx+1], o.memDoms[0:idx])
			o.memDoms[0] = item
		}
	case "up":
		if idx > 0 {
			o.memDoms[idx-1], o.memDoms[idx] = o.memDoms[idx], o.memDoms[idx-1]
		}
	case "down":
		if idx < len(o.memDoms)-1 {
			o.memDoms[idx+1], o.memDoms[idx] = o.memDoms[idx], o.memDoms[idx+1]
		}
	}
	for i := range o.memDoms {
		o.memDoms[i].Position = i + 1
	}
	return nil
}

func (o *originStoreType) reorderDomains(ctx context.Context, orderedIDs []int64) error {
	defer o.invalidateCache()
	if Pools != nil {
		return db.ReorderOriginDomains(ctx, Pools.Admin, orderedIDs)
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	byID := make(map[int64]db.OriginDomain, len(o.memDoms))
	for _, d := range o.memDoms {
		byID[d.ID] = d
	}
	seen := make(map[int64]bool, len(o.memDoms))
	next := make([]db.OriginDomain, 0, len(o.memDoms))
	for _, id := range orderedIDs {
		if d, ok := byID[id]; ok && !seen[id] {
			seen[id] = true
			next = append(next, d)
		}
	}
	for _, d := range o.memDoms {
		if !seen[d.ID] {
			seen[d.ID] = true
			next = append(next, d)
		}
	}
	for i := range next {
		next[i].Position = i + 1
	}
	o.memDoms = next
	return nil
}

func (o *originStoreType) deleteDomain(ctx context.Context, id int64) (bool, error) {
	defer o.invalidateCache()
	if Pools != nil {
		return db.DeleteOriginDomain(ctx, Pools.Admin, id)
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	for i, d := range o.memDoms {
		if d.ID == id {
			o.memDoms = append(o.memDoms[:i], o.memDoms[i+1:]...)
			for j := range o.memDoms {
				o.memDoms[j].Position = j + 1
			}
			return true, nil
		}
	}
	return false, nil
}

// ── HTTP Origin Resolution Middleware ────────────────────────────────────────

func effectiveRequestHost(r *http.Request, trustForwarded bool) (fullHost string, bareHost string) {
	raw := r.Host
	if trustForwarded {
		if xfh := strings.TrimSpace(r.Header.Get("X-Forwarded-Host")); xfh != "" {
			raw = xfh
		}
	}
	return db.SplitHostAndPort(raw)
}

// OriginMiddleware resolves the incoming request's Host against tenant custom
// domains and configured system origin domains:
//
//  1. /healthz is always served without redirection so container probes work.
//  2. If the host matches an active tenant's custom domain:
//     - Sysadmin routes (/s/*, /api/sysadmin/*) and other tenants' routes are blocked.
//     - The request runs in OriginModeTenant (tenant-branded login & tenant-scoped auth).
//  3. Otherwise, if no system origin domains are enabled:
//     - The request runs in OriginModeDefault (unchanged default behavior).
//  4. Otherwise (one or more system origin domains are enabled):
//     - If the host matches an enabled system origin domain, the request runs in
//       OriginModeSystem (Platform System login & sysadmin-only authentication).
//     - Otherwise (direct IP or unlisted host), the client is redirected to the
//       first enabled domain in the system origin list.
func OriginMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" {
			next.ServeHTTP(w, r)
			return
		}

		snap := origins.snapshot(r.Context())
		fullHost, bareHost := effectiveRequestHost(r, snap.settings.TrustForwardedHeaders)

		// 1. Check whether the incoming host belongs to a tenant custom domain.
		if bareHost != "" {
			if tenant, ok := store.getByDomain(fullHost, bareHost); ok {
				if isBlockedOnTenantOrigin(r.URL.Path, tenant.Slug) {
					serveOriginNotFound(w, r)
					return
				}
				var primarySys string
				if len(snap.enabledDomains) > 0 {
					primarySys = snap.enabledDomains[0].Domain
				}
				tCopy := tenant
				oc := OriginContext{
					Mode:                OriginModeTenant,
					MatchedDomain:       tenant.Domain,
					PrimarySystemDomain: primarySys,
					RequestHost:         fullHost,
					Tenant:              &tCopy,
				}
				next.ServeHTTP(w, withOriginContext(r, oc))
				return
			}
		}

		// 2. If no system origin domains are enabled, behave exactly as default.
		if len(snap.enabledDomains) == 0 {
			oc := OriginContext{
				Mode:        OriginModeDefault,
				RequestHost: fullHost,
			}
			next.ServeHTTP(w, withOriginContext(r, oc))
			return
		}

		// 3. System origin domains are configured: check if request host is allowed.
		primary := snap.enabledDomains[0].Domain
		for _, d := range snap.enabledDomains {
			if db.HostMatchesConfiguredDomain(fullHost, bareHost, d.Domain) {
				oc := OriginContext{
					Mode:                OriginModeSystem,
					MatchedDomain:       d.Domain,
					PrimarySystemDomain: primary,
					RequestHost:         fullHost,
				}
				next.ServeHTTP(w, withOriginContext(r, oc))
				return
			}
		}

		// 4. Direct access or unrecognized host -> redirect to the first domain in the list.
		scheme := snap.settings.RedirectScheme
		if scheme != "http" && scheme != "https" {
			scheme = "https"
		}
		code := snap.settings.RedirectStatusCode
		switch code {
		case 301, 302, 307, 308:
		default:
			code = http.StatusFound
		}
		uri := r.URL.RequestURI()
		if uri == "" || uri[0] != '/' {
			uri = "/"
		}
		target := scheme + "://" + primary + uri
		http.Redirect(w, r, target, code)
	})
}

func isBlockedOnTenantOrigin(path, tenantSlug string) bool {
	if path == "/s" || strings.HasPrefix(path, "/s/") {
		return true
	}
	if path == "/api/sysadmin" || strings.HasPrefix(path, "/api/sysadmin/") {
		return true
	}
	if strings.HasPrefix(path, "/t/") {
		rest := strings.TrimPrefix(path, "/t/")
		seg := rest
		if idx := strings.IndexByte(rest, '/'); idx >= 0 {
			seg = rest[:idx]
		}
		if seg != "" && seg != tenantSlug {
			return true
		}
	}
	if strings.HasPrefix(path, "/api/tenant/") {
		rest := strings.TrimPrefix(path, "/api/tenant/")
		seg := rest
		if idx := strings.IndexByte(rest, '/'); idx >= 0 {
			seg = rest[:idx]
		}
		if seg != "" && seg != tenantSlug {
			return true
		}
	}
	return false
}

func serveOriginNotFound(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/") {
		apiError(w, http.StatusNotFound, "Not found.")
		return
	}
	ServeError(w, r, http.StatusNotFound, "")
}

// ── Tenant Brand Store & Login Branding Resolver ─────────────────────────────

type memTenantBrand struct {
	CompanyName  string
	Tagline      string
	Website      string
	SupportEmail string
	LogoURL      string
	Colors       json.RawMessage
}

var (
	memBrandMu sync.RWMutex
	memBrands  = map[string]memTenantBrand{} // keyed by tenant slug
)

func setMemTenantBrand(slug string, b memTenantBrand) {
	memBrandMu.Lock()
	memBrands[slug] = b
	memBrandMu.Unlock()
}

func getMemTenantBrand(slug string) (memTenantBrand, bool) {
	memBrandMu.RLock()
	b, ok := memBrands[slug]
	memBrandMu.RUnlock()
	return b, ok
}

// TenantLoginBrand holds sanitized branding fields for rendering a tenant's
// white-labeled login and password-recovery pages.
type TenantLoginBrand struct {
	TenantID        string
	TenantSlug      string
	TenantDomain    string
	Name            string
	Initial         string
	Tagline         string
	Website         string
	WebsiteLabel    string
	SupportEmail    string
	LogoURL         string
	PrimaryColor    string
	BackgroundColor string
	AccentColor     string
	LightPanel      bool
}

var safeHexColorRE = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

func sanitizeHexColor(raw, fallback string) string {
	s := strings.TrimSpace(raw)
	if safeHexColorRE.MatchString(s) {
		return s
	}
	return fallback
}

func firstInitial(name string) string {
	s := strings.TrimSpace(name)
	if s == "" {
		return "T"
	}
	r, _ := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError {
		return "T"
	}
	return string(unicode.ToUpper(r))
}

func resolveTenantLoginBrand(ctx context.Context, t Tenant) TenantLoginBrand {
	out := TenantLoginBrand{
		TenantID:        t.ID,
		TenantSlug:      t.Slug,
		TenantDomain:    t.Domain,
		Name:            t.Name,
		Initial:         firstInitial(t.Name),
		Tagline:         "Manage your cloud resources, deployments, and services in one place.",
		PrimaryColor:    "#1d1d1f",
		BackgroundColor: "#1d1d1f",
		AccentColor:     "#0071e3",
	}
	if t.Domain != "" {
		out.Website = "https://" + t.Domain
		out.WebsiteLabel = t.Domain
		out.SupportEmail = "support@" + t.Domain
	}

	var companyName, tagline, website, supportEmail, logoURL string
	var colorsJSON json.RawMessage

	if Pools != nil && t.DBID > 0 {
		qCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		if bs, ok, err := db.GetBrandSettings(qCtx, Pools.Admin, t.DBID); err == nil && ok {
			companyName = bs.CompanyName
			tagline = bs.Tagline
			website = bs.Website
			supportEmail = bs.SupportEmail
			logoURL = bs.LogoURL
			colorsJSON = bs.Colors
		}
	} else if mb, ok := getMemTenantBrand(t.Slug); ok {
		companyName = mb.CompanyName
		tagline = mb.Tagline
		website = mb.Website
		supportEmail = mb.SupportEmail
		logoURL = mb.LogoURL
		colorsJSON = mb.Colors
	}

	if strings.TrimSpace(companyName) != "" {
		out.Name = strings.TrimSpace(companyName)
		out.Initial = firstInitial(out.Name)
	}
	if strings.TrimSpace(tagline) != "" {
		out.Tagline = strings.TrimSpace(tagline)
	}
	if strings.TrimSpace(website) != "" {
		w := strings.TrimSpace(website)
		if !strings.HasPrefix(strings.ToLower(w), "http://") && !strings.HasPrefix(strings.ToLower(w), "https://") {
			w = "https://" + w
		}
		out.Website = w
		lbl := strings.TrimPrefix(strings.TrimPrefix(w, "https://"), "http://")
		lbl = strings.TrimSuffix(lbl, "/")
		out.WebsiteLabel = lbl
	}
	if strings.TrimSpace(supportEmail) != "" {
		out.SupportEmail = strings.TrimSpace(supportEmail)
	}
	if strings.TrimSpace(logoURL) != "" {
		u := strings.TrimSpace(logoURL)
		lower := strings.ToLower(u)
		if strings.HasPrefix(lower, "https://") || strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "/") || strings.HasPrefix(lower, "data:image/") {
			out.LogoURL = u
		}
	}

	if len(colorsJSON) > 2 {
		var swatches []struct {
			Label string `json:"label"`
			Hex   string `json:"hex"`
		}
		if json.Unmarshal(colorsJSON, &swatches) == nil {
			for _, sw := range swatches {
				lbl := strings.ToLower(strings.TrimSpace(sw.Label))
				switch lbl {
				case "primary":
					out.PrimaryColor = sanitizeHexColor(sw.Hex, out.PrimaryColor)
					// If no dark background override is set, tint the brand panel with Primary if dark enough,
					// or use Primary as button color.
				case "background":
					if hex := sanitizeHexColor(sw.Hex, ""); hex != "" {
						out.BackgroundColor = hex
					}
				case "accent":
					out.AccentColor = sanitizeHexColor(sw.Hex, out.AccentColor)
				}
			}
		}
	}

	return out
}
