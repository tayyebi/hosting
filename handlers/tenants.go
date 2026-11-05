package handlers

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"html/template"
	"net/http"
	"net/mail"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

var slugRe = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,61}[a-z0-9]$`)

type Tenant struct {
	ID     string
	Name   string
	Slug   string
	Domain string
	DBID   int64
}

type tenantsStore struct {
	mu      sync.RWMutex
	tenants []Tenant
}

var buyVMID = newID()
var abrNOCID = newID()
var hostingID = newID()

var store = &tenantsStore{
	tenants: []Tenant{
		{ID: buyVMID, Name: "BuyVM", Slug: "buyvm", Domain: "buyvm.net"},
		{ID: abrNOCID, Name: "abrNOC", Slug: "abrnoc", Domain: ""},
		{ID: hostingID, Name: "Hosting", Slug: "hosting", Domain: "new.hosting.com"},
	},
}

// ── Tenant owner store ────────────────────────────────────────

type TenantOwner struct {
	Name  string
	Email string
}

type ownerStoreType struct {
	mu     sync.RWMutex
	owners map[string]*TenantOwner // key: tenant ID
}

var owners = &ownerStoreType{
	owners: map[string]*TenantOwner{
		buyVMID:  {Name: "BuyVM Admin", Email: "admin@buyvm.net"},
		hostingID: {Name: "Hosting Admin", Email: "admin@new.hosting.com"},
	},
}

func (o *ownerStoreType) get(tenantID string) (*TenantOwner, bool) {
	o.mu.RLock()
	defer o.mu.RUnlock()
	owner, ok := o.owners[tenantID]
	return owner, ok
}

func (o *ownerStoreType) set(tenantID string, owner TenantOwner) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.owners[tenantID] = &owner
}

func (o *ownerStoreType) all() map[string]*TenantOwner {
	o.mu.RLock()
	defer o.mu.RUnlock()
	out := make(map[string]*TenantOwner, len(o.owners))
	for k, v := range o.owners {
		cp := *v
		out[k] = &cp
	}
	return out
}

func newID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func isValidEmail(s string) bool {
	_, err := mail.ParseAddress(s)
	return err == nil && strings.Contains(s, "@")
}

func (s *tenantsStore) list() []Tenant {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Tenant, len(s.tenants))
	copy(out, s.tenants)
	return out
}

func (s *tenantsStore) add(t Tenant) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tenants = append(s.tenants, t)
}

func (s *tenantsStore) update(t Tenant) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, existing := range s.tenants {
		if existing.ID == t.ID {
			s.tenants[i] = t
			return true
		}
	}
	return false
}

func (s *tenantsStore) delete(id string) {
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
	Tenants    []Tenant
	Owners     map[string]*TenantOwner
	FormError  string
	EditTenant *Tenant
	// OwnerModal: tenant being targeted for add-owner action
	OwnerTenant *Tenant
	// ResetResult: temp password shown after reset
	ResetResult string
}

// renderTenantsPage is a helper used by all tenant handlers to re-render the page.
func renderTenantsPage(w http.ResponseWriter, r *http.Request, formErr string, editTenant *Tenant, extra ...interface{}) {
	user, _ := getSession(r)

	data := tenantsInnerData{
		Tenants:   store.list(),
		Owners:    owners.all(),
		FormError: formErr,
		EditTenant: editTenant,
	}
	// extra[0] = *Tenant (ownerTenant), extra[1] = string (resetResult)
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

	// Step 1: render the tenant-specific inner content to a buffer.
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

	// Step 2: render the shell with the pre-built content.
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

// tenantsOwnerModalHandler shows the add-owner modal for a tenant.
func tenantsOwnerModalHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	t, ok := store.get(id)
	if !ok {
		http.NotFound(w, r)
		return
	}
	renderTenantsPage(w, r, "", nil, &t)
}

// tenantsAddOwnerHandler creates or replaces the owner for a tenant.
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

// tenantsResetOwnerPasswordHandler generates a mock temporary password.
func tenantsResetOwnerPasswordHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, ok := store.get(id); !ok {
		http.NotFound(w, r)
		return
	}
	// Generate a mock temporary password (8 random hex chars).
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	tmpPw := "Tmp-" + hex.EncodeToString(b)
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
