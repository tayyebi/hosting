package handlers

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"platform/db"
)

type Tenant struct {
	ID     string
	DBID   int64
	Name   string
	Slug   string
	Domain string // optional custom domain, e.g. "app.acme.com"
}

// TenantOwner holds the owner account info for a tenant.
type TenantOwner struct {
	Name  string
	Email string
}

type ownerStoreType struct {
	mu     sync.RWMutex
	owners map[string]*TenantOwner // keyed by tenant ID
}

var owners = &ownerStoreType{
	owners: map[string]*TenantOwner{},
}

func (o *ownerStoreType) get(tenantID string) (*TenantOwner, bool) {
	if Pools != nil {
		numID, err := strconv.ParseInt(tenantID, 10, 64)
		if err == nil && numID > 0 {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			row, ok, dbErr := db.GetTenantByID(ctx, Pools.Admin, numID)
			if dbErr == nil && ok && (row.OwnerEmail != "" || row.OwnerName != "") {
				return &TenantOwner{Name: row.OwnerName, Email: row.OwnerEmail}, true
			}
			return nil, false
		}
	}
	o.mu.RLock()
	defer o.mu.RUnlock()
	v, ok := o.owners[tenantID]
	return v, ok
}

func (o *ownerStoreType) set(tenantID string, owner TenantOwner) {
	if Pools != nil {
		numID, err := strconv.ParseInt(tenantID, 10, 64)
		if err == nil && numID > 0 {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if dbErr := db.SetTenantOwner(ctx, Pools.Admin, numID, owner.Name, owner.Email); dbErr != nil {
				log.Printf("tenants: SetTenantOwner(%d): %v", numID, dbErr)
			}
			return
		}
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	o.owners[tenantID] = &owner
}

func (o *ownerStoreType) all() map[string]*TenantOwner {
	cp := make(map[string]*TenantOwner)
	if Pools != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		rows, err := db.ListTenants(ctx, Pools.Admin)
		if err == nil {
			for _, r := range rows {
				if r.OwnerEmail != "" || r.OwnerName != "" {
					cp[strconv.FormatInt(r.ID, 10)] = &TenantOwner{
						Name:  r.OwnerName,
						Email: r.OwnerEmail,
					}
				}
			}
			return cp
		}
	}
	o.mu.RLock()
	defer o.mu.RUnlock()
	for k, v := range o.owners {
		cp[k] = v
	}
	return cp
}

// slugRe validates a tenant slug: lowercase alphanumeric + hyphens, 3-63 chars,
// cannot start or end with a hyphen.
var slugRe = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,61}[a-z0-9]$`)

func isValidEmail(s string) bool {
	at := strings.IndexByte(s, '@')
	if at <= 0 || at >= len(s)-3 {
		return false
	}
	domain := s[at+1:]
	dot := strings.LastIndexByte(domain, '.')
	return dot > 0 && dot < len(domain)-1 && !strings.ContainsAny(s, " \t\r\n")
}

type tenantsStore struct {
	mu      sync.RWMutex
	tenants []Tenant
}

var store = &tenantsStore{
	tenants: nil, // No dummy tenants; backed by PostgreSQL `tenants` when Pools != nil.
}

func newID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func tenantFromRow(r db.TenantRow) Tenant {
	return Tenant{
		ID:     strconv.FormatInt(r.ID, 10),
		DBID:   r.ID,
		Name:   r.Name,
		Slug:   r.Slug,
		Domain: r.Domain,
	}
}

func (s *tenantsStore) list() []Tenant {
	if Pools != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		rows, err := db.ListTenants(ctx, Pools.Admin)
		if err != nil {
			log.Printf("tenants: db.ListTenants: %v", err)
			return nil
		}
		out := make([]Tenant, 0, len(rows))
		for _, r := range rows {
			out = append(out, tenantFromRow(r))
		}
		return out
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	cp := make([]Tenant, len(s.tenants))
	copy(cp, s.tenants)
	return cp
}

func (s *tenantsStore) add(t Tenant) Tenant {
	if Pools != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		row, err := db.CreateTenant(ctx, Pools.Admin, t.Name, t.Slug, t.Domain)
		if err != nil {
			log.Printf("tenants: db.CreateTenant: %v", err)
			return t
		}
		return tenantFromRow(row)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if t.ID == "" {
		t.ID = newID()
	}
	s.tenants = append(s.tenants, t)
	return t
}

func (s *tenantsStore) update(updated Tenant) {
	if Pools != nil {
		numID, err := strconv.ParseInt(updated.ID, 10, 64)
		if err == nil && numID > 0 {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if _, _, dbErr := db.UpdateTenant(ctx, Pools.Admin, numID, updated.Name, updated.Slug, updated.Domain); dbErr != nil {
				log.Printf("tenants: db.UpdateTenant(%d): %v", numID, dbErr)
			}
			return
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, t := range s.tenants {
		if t.ID == updated.ID {
			s.tenants[i] = updated
			return
		}
	}
}

func (s *tenantsStore) delete(id string) {
	if Pools != nil {
		numID, err := strconv.ParseInt(id, 10, 64)
		if err == nil && numID > 0 {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if dbErr := db.DeleteTenant(ctx, Pools.Admin, numID); dbErr != nil {
				log.Printf("tenants: db.DeleteTenant(%d): %v", numID, dbErr)
			}
			return
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, t := range s.tenants {
		if t.ID == id {
			s.tenants = append(s.tenants[:i], s.tenants[i+1:]...)
			return
		}
	}
}

func (s *tenantsStore) get(id string) (Tenant, bool) {
	if Pools != nil {
		numID, err := strconv.ParseInt(id, 10, 64)
		if err == nil && numID > 0 {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			row, ok, dbErr := db.GetTenantByID(ctx, Pools.Admin, numID)
			if dbErr == nil && ok {
				return tenantFromRow(row), true
			}
		}
		// Allow lookup by slug if non-numeric ID passed
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		row, ok, dbErr := db.GetTenantBySlug(ctx, Pools.Admin, id)
		if dbErr == nil && ok {
			return tenantFromRow(row), true
		}
		return Tenant{}, false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, t := range s.tenants {
		if t.ID == id {
			return t, true
		}
	}
	return Tenant{}, false
}

func (s *tenantsStore) getBySlug(slug string) (Tenant, bool) {
	if Pools != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		row, ok, err := db.GetTenantBySlug(ctx, Pools.Admin, slug)
		if err == nil && ok {
			return tenantFromRow(row), true
		}
		return Tenant{}, false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, t := range s.tenants {
		if t.Slug == slug {
			return t, true
		}
	}
	return Tenant{}, false
}

// slugExists returns true if another tenant already uses the given slug.
func (s *tenantsStore) slugExists(slug, excludeID string) bool {
	if Pools != nil {
		var exNum int64
		if excludeID != "" {
			exNum, _ = strconv.ParseInt(excludeID, 10, 64)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		exists, err := db.TenantSlugExists(ctx, Pools.Admin, slug, exNum)
		if err == nil {
			return exists
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, t := range s.tenants {
		if t.Slug == slug && t.ID != excludeID {
			return true
		}
	}
	return false
}

// ── page data ────────────────────────────────────────────────

type tenantsInnerData struct {
	Tenants     []Tenant
	Owners      map[string]*TenantOwner
	FormError   string
	EditTenant  *Tenant
	OwnerTenant *Tenant
	ResetResult string
}

// renderTenantsPage is a helper used by all tenant handlers to re-render the page.
func renderTenantsPage(w http.ResponseWriter, r *http.Request, formErr string, editTenant *Tenant, extra ...interface{}) {
	user, _ := getSession(r)

	data := tenantsInnerData{
		Tenants:    store.list(),
		Owners:     owners.all(),
		FormError:  formErr,
		EditTenant: editTenant,
	}
	if len(extra) > 0 {
		if t, ok := extra[0].(*Tenant); ok {
			data.OwnerTenant = t
		}
	}
	if len(extra) > 1 {
		if s, ok := extra[1].(string); ok {
			data.ResetResult = s
		}
	}

	innerTmpl, err := template.ParseFiles(filepath.Join("templates", "sysadmin", "tenants.html"))
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	var buf bytes.Buffer
	if err := innerTmpl.ExecuteTemplate(&buf, "tenants-content", data); err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	shellPaths := []string{
		filepath.Join("templates", "sysadmin", "base.html"),
		filepath.Join("templates", "sysadmin", "shell.html"),
	}
	shellTmpl, err := template.ParseFiles(shellPaths...)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	shellData := SysadminData{
		User:      user,
		Nav:       sysadminNav,
		ActiveNav: "tenants",
		PageTitle: "Tenants",
		Content:   template.HTML(buf.String()),
	}
	if err := shellTmpl.ExecuteTemplate(w, "sysadmin-base", shellData); err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
	}
}

// ── handlers ─────────────────────────────────────────────────

func tenantsListHandler(w http.ResponseWriter, r *http.Request) {
	renderTenantsPage(w, r, "", nil)
}

func tenantsCreateHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	slug := strings.TrimSpace(r.FormValue("slug"))
	domain := strings.TrimSpace(r.FormValue("domain"))

	if name == "" || slug == "" {
		renderTenantsPage(w, r, "Name and slug are required.", nil)
		return
	}
	if len(name) > 100 {
		renderTenantsPage(w, r, "Name must be 100 characters or fewer.", nil)
		return
	}
	if !slugRe.MatchString(slug) {
		renderTenantsPage(w, r, "Slug must be 3–63 lowercase alphanumeric characters or hyphens, and cannot start or end with a hyphen.", nil)
		return
	}
	if store.slugExists(slug, "") {
		renderTenantsPage(w, r, "Slug \""+slug+"\" is already in use.", nil)
		return
	}
	store.add(Tenant{ID: newID(), Name: name, Slug: slug, Domain: domain})
	http.Redirect(w, r, "/s/tenants/", http.StatusSeeOther)
}

func tenantsEditHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	t, ok := store.get(id)
	if !ok {
		http.NotFound(w, r)
		return
	}
	renderTenantsPage(w, r, "", &t)
}

func tenantsUpdateHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, ok := store.get(id); !ok {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	slug := strings.TrimSpace(r.FormValue("slug"))
	domain := strings.TrimSpace(r.FormValue("domain"))

	if name == "" || slug == "" {
		t, _ := store.get(id)
		renderTenantsPage(w, r, "Name and slug are required.", &t)
		return
	}
	if len(name) > 100 {
		t, _ := store.get(id)
		renderTenantsPage(w, r, "Name must be 100 characters or fewer.", &t)
		return
	}
	if !slugRe.MatchString(slug) {
		t, _ := store.get(id)
		renderTenantsPage(w, r, "Slug must be 3–63 lowercase alphanumeric characters or hyphens, and cannot start or end with a hyphen.", &t)
		return
	}
	if store.slugExists(slug, id) {
		t, _ := store.get(id)
		renderTenantsPage(w, r, "Slug \""+slug+"\" is already in use.", &t)
		return
	}
	store.update(Tenant{ID: id, Name: name, Slug: slug, Domain: domain})
	http.Redirect(w, r, "/s/tenants/", http.StatusSeeOther)
}

func tenantsDeleteHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	store.delete(id)
	http.Redirect(w, r, "/s/tenants/", http.StatusSeeOther)
}

func tenantsOwnerModalHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	t, ok := store.get(id)
	if !ok {
		http.NotFound(w, r)
		return
	}
	renderTenantsPage(w, r, "", nil, &t)
}

func tenantsAddOwnerHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, ok := store.get(id); !ok {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(r.FormValue("owner_name"))
	email := strings.TrimSpace(r.FormValue("owner_email"))
	if name == "" || email == "" {
		t, _ := store.get(id)
		renderTenantsPage(w, r, "Owner name and email are required.", nil, &t)
		return
	}
	owners.set(id, TenantOwner{Name: name, Email: email})
	http.Redirect(w, r, "/s/tenants/", http.StatusSeeOther)
}

func tenantsResetOwnerPasswordHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, ok := store.get(id); !ok {
		http.NotFound(w, r)
		return
	}
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	tmpPw := "Tmp1" + hex.EncodeToString(b)
	if o, hasOwner := owners.get(id); hasOwner && o.Email != "" && Pools != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		if _, err := db.SetUserPasswordByEmail(ctx, Pools.Admin, o.Email, o.Name, "user", tmpPw); err != nil {
			log.Printf("tenants: reset owner password: %v", err)
		}
	}
	renderTenantsPage(w, r, "", nil, nil, tmpPw)
}

func RegisterTenantRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /s/tenants/", requireRole(RoleSysAdmin, tenantsListHandler))
	mux.HandleFunc("POST /s/tenants/", requireRole(RoleSysAdmin, tenantsCreateHandler))
	mux.HandleFunc("GET /s/tenants/{id}/edit/", requireRole(RoleSysAdmin, tenantsEditHandler))
	mux.HandleFunc("POST /s/tenants/{id}/edit/", requireRole(RoleSysAdmin, tenantsUpdateHandler))
	mux.HandleFunc("POST /s/tenants/{id}/delete/", requireRole(RoleSysAdmin, tenantsDeleteHandler))
	mux.HandleFunc("GET /s/tenants/{id}/owner/", requireRole(RoleSysAdmin, tenantsOwnerModalHandler))
	mux.HandleFunc("POST /s/tenants/{id}/owner/", requireRole(RoleSysAdmin, tenantsAddOwnerHandler))
	mux.HandleFunc("POST /s/tenants/{id}/owner/reset/", requireRole(RoleSysAdmin, tenantsResetOwnerPasswordHandler))
}
