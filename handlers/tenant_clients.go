package handlers

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"platform/db"
)

// ── Workspace & Client ────────────────────────────────────────

type Workspace struct {
	ID   string
	Name string
	DBID int64 // numeric DB primary key; 0 when not yet resolved
}

type Client struct {
	ID         string
	Name       string
	Email      string
	Company    string
	Status     string // "active" | "suspended"
	Workspaces []Workspace
}

func (c Client) HasDefaultWorkspace() bool {
	for _, w := range c.Workspaces {
		if strings.EqualFold(w.Name, "default") {
			return true
		}
	}
	return false
}

// ── Store ─────────────────────────────────────────────────────

type clientStoreType struct {
	mu      sync.RWMutex
	clients map[string][]Client // key: tenant slug
}

func mkWorkspaces(names ...string) []Workspace {
	ws := make([]Workspace, len(names))
	for i, n := range names {
		ws[i] = Workspace{ID: newID(), Name: n}
	}
	return ws
}

var clientStore = &clientStoreType{
	clients: map[string][]Client{
		"buyvm": {
			{ID: newID(), Name: "James Wright", Email: "james@acme.io", Company: "Acme Inc.", Status: "active", Workspaces: mkWorkspaces("default", "staging", "production")},
			{ID: newID(), Name: "Sara Nguyen", Email: "sara@techblue.com", Company: "TechBlue", Status: "active", Workspaces: mkWorkspaces("default")},
			{ID: newID(), Name: "Omar Khalid", Email: "omar@khalid.dev", Company: "", Status: "suspended", Workspaces: mkWorkspaces("default", "dev")},
			{ID: newID(), Name: "Lena Bauer", Email: "lena@bauerhosts.de", Company: "BauerHosts", Status: "active", Workspaces: mkWorkspaces("default", "eu-west")},
			{ID: newID(), Name: "Carlos Diaz", Email: "carlos@diaznet.mx", Company: "DiazNet", Status: "active", Workspaces: []Workspace{{ID: newID(), Name: "production"}}}, // missing default
		},
		"abrnoc": {
			{ID: newID(), Name: "Priya Sharma", Email: "priya@infranow.in", Company: "InfraNow", Status: "active", Workspaces: mkWorkspaces("default")},
			{ID: newID(), Name: "Tom Fischer", Email: "tom@fischer.eu", Company: "Fischer GmbH", Status: "active", Workspaces: mkWorkspaces("default", "backup")},
			{ID: newID(), Name: "Yuki Tanaka", Email: "yuki@tanaka.jp", Company: "", Status: "suspended", Workspaces: mkWorkspaces("default")},
		},
		"hosting": {
			{ID: newID(), Name: "Alice Monroe", Email: "alice@monroe.io", Company: "Monroe LLC", Status: "active", Workspaces: mkWorkspaces("default", "us-east", "eu-central")},
			{ID: newID(), Name: "Ben Okonkwo", Email: "ben@okonkwo.ng", Company: "Okonkwo Tech", Status: "active", Workspaces: mkWorkspaces("default")},
			{ID: newID(), Name: "Chloe Martin", Email: "chloe@martindev.fr", Company: "MartinDev", Status: "active", Workspaces: mkWorkspaces("default", "test")},
			{ID: newID(), Name: "Diego Vega", Email: "diego@vega.ar", Company: "", Status: "suspended", Workspaces: []Workspace{{ID: newID(), Name: "legacy"}}}, // missing default
			{ID: newID(), Name: "Elena Popov", Email: "elena@popov.ru", Company: "Popov Systems", Status: "active", Workspaces: mkWorkspaces("default")},
			{ID: newID(), Name: "Faisal Al-Amin", Email: "faisal@alamin.sa", Company: "Al-Amin Corp", Status: "active", Workspaces: mkWorkspaces("default", "prod", "staging")},
		},
	},
}

func (s *clientStoreType) list(slug string) []Client {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.clients[slug]
	out := make([]Client, len(src))
	copy(out, src)
	return out
}

func (s *clientStoreType) getByID(slug, id string) (Client, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, c := range s.clients[slug] {
		if c.ID == id {
			return c, true
		}
	}
	return Client{}, false
}

func (s *clientStoreType) add(slug string, c Client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[slug] = append(s.clients[slug], c)
}

func (s *clientStoreType) update(slug string, c Client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, existing := range s.clients[slug] {
		if existing.ID == c.ID {
			s.clients[slug][i] = c
			return
		}
	}
}

func (s *clientStoreType) getWorkspace(slug, workspaceID string) (Client, Workspace, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, c := range s.clients[slug] {
		for _, w := range c.Workspaces {
			if w.ID == workspaceID {
				return c, w, true
			}
		}
	}
	return Client{}, Workspace{}, false
}

func (s *clientStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.clients[slug]
	for i, c := range list {
		if c.ID == id {
			s.clients[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

// ── DB conversion ─────────────────────────────────────────────

func dbClientToLocal(dc db.TenantClient) Client {
	ws := make([]Workspace, len(dc.Workspaces))
	for i, name := range dc.Workspaces {
		ws[i] = Workspace{ID: fmt.Sprintf("%d", dc.ID*100+int64(i)), Name: name}
	}
	return Client{
		ID: fmt.Sprintf("%d", dc.ID), Name: dc.Name, Email: dc.Email,
		Company: dc.Company, Status: dc.Status, Workspaces: ws,
	}
}

// ── Page data ─────────────────────────────────────────────────

type clientsInnerData struct {
	Clients      []Client
	TenantSlug   string
	Search       string
	StatusFilter string
	FormError    string
	EditClient   *Client
	ResetResult  string
	ResetEmail   string
}

func filterClients(all []Client, search, status string) []Client {
	var out []Client
	q := strings.ToLower(search)
	for _, c := range all {
		if status == "active" && c.Status != "active" {
			continue
		}
		if status == "suspended" && c.Status != "suspended" {
			continue
		}
		if q != "" {
			hay := strings.ToLower(c.Name + " " + c.Email + " " + c.Company)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, c)
	}
	return out
}

func renderClientsPage(w http.ResponseWriter, r *http.Request, tenant Tenant, user User, d clientsInnerData) {
	innerTmpl, err := template.ParseFiles(filepath.Join("templates", "tenant", "clients_content.html"))
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	var buf bytes.Buffer
	if err := innerTmpl.ExecuteTemplate(&buf, "clients-content", d); err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	shellPaths := []string{
		filepath.Join("templates", "tenant", "base.html"),
		filepath.Join("templates", "tenant", "shell.html"),
	}
	shellTmpl, err := template.ParseFiles(shellPaths...)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	data := TenantShellData{
		User: user, Tenant: tenant,
		Nav: tenantNav(tenant.Slug), NavGroups: tenantNavGroups(tenant.Slug),
		ActiveNav: "clients", PageTitle: "Clients",
		IsSysAdmin: user.Role == RoleSysAdmin,
		Content:    template.HTML(buf.String()),
	}
	if err := shellTmpl.ExecuteTemplate(w, "tenant-base", data); err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
	}
}

// ── Handlers ─────────────────────────────────────────────────

func clientsListHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := getSession(r)
	tenant, _ := tenantFromContext(r)
	search := r.URL.Query().Get("q")
	status := r.URL.Query().Get("status")

	var clients []Client
	if Pools != nil && tenant.DBID > 0 {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		dcs, err := db.ListTenantClients(ctx, Pools.Admin, tenant.DBID, search, status)
		if err != nil {
			log.Printf("clientsListHandler: ListTenantClients: %v", err)
			clients = filterClients(clientStore.list(tenant.Slug), search, status)
		} else {
			clients = make([]Client, len(dcs))
			for i, dc := range dcs {
				clients[i] = dbClientToLocal(dc)
			}
		}
	} else {
		clients = filterClients(clientStore.list(tenant.Slug), search, status)
	}

	renderClientsPage(w, r, tenant, user, clientsInnerData{
		Clients:      clients,
		TenantSlug:   tenant.Slug,
		Search:       search,
		StatusFilter: status,
	})
}

func clientsCreateHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := getSession(r)
	tenant, _ := tenantFromContext(r)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	email := strings.TrimSpace(r.FormValue("email"))
	company := strings.TrimSpace(r.FormValue("company"))
	if name == "" || email == "" {
		renderClientsPage(w, r, tenant, user, clientsInnerData{
			Clients: clientStore.list(tenant.Slug), TenantSlug: tenant.Slug,
			FormError: "Name and email are required.",
		})
		return
	}
	if len(name) > 100 || len(company) > 100 {
		renderClientsPage(w, r, tenant, user, clientsInnerData{
			Clients: clientStore.list(tenant.Slug), TenantSlug: tenant.Slug,
			FormError: "Name and company must be 100 characters or fewer.",
		})
		return
	}
	if len(email) > 254 {
		renderClientsPage(w, r, tenant, user, clientsInnerData{
			Clients: clientStore.list(tenant.Slug), TenantSlug: tenant.Slug,
			FormError: "Email must be 254 characters or fewer.",
		})
		return
	}
	if !isValidEmail(email) {
		renderClientsPage(w, r, tenant, user, clientsInnerData{
			Clients: clientStore.list(tenant.Slug), TenantSlug: tenant.Slug,
			FormError: "Invalid email address.",
		})
		return
	}

	if Pools != nil && tenant.DBID > 0 {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		if _, err := db.CreateTenantClient(ctx, Pools.Admin, tenant.DBID, name, email, company); err != nil {
			log.Printf("clientsCreateHandler: CreateTenantClient: %v", err)
			renderClientsPage(w, r, tenant, user, clientsInnerData{
				Clients: clientStore.list(tenant.Slug), TenantSlug: tenant.Slug,
				FormError: "Failed to create client: " + err.Error(),
			})
			return
		}
	} else {
		clientStore.add(tenant.Slug, Client{
			ID: newID(), Name: name, Email: email, Company: company,
			Status:     "active",
			Workspaces: mkWorkspaces("default"),
		})
	}
	http.Redirect(w, r, "/t/"+tenant.Slug+"/clients/", http.StatusSeeOther)
}

func clientsEditModalHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := getSession(r)
	tenant, _ := tenantFromContext(r)
	idStr := r.PathValue("id")

	var c Client
	if Pools != nil && tenant.DBID > 0 {
		clientID, err := strconv.ParseInt(idStr, 10, 64)
		if err == nil {
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()
			dc, dbErr := db.GetTenantClient(ctx, Pools.Admin, tenant.DBID, clientID)
			if dbErr != nil {
				log.Printf("clientsEditModalHandler: GetTenantClient: %v", dbErr)
				http.NotFound(w, r)
				return
			}
			c = dbClientToLocal(dc)
		} else {
			http.NotFound(w, r)
			return
		}
	} else {
		var ok bool
		c, ok = clientStore.getByID(tenant.Slug, idStr)
		if !ok {
			http.NotFound(w, r)
			return
		}
	}

	renderClientsPage(w, r, tenant, user, clientsInnerData{
		Clients: clientStore.list(tenant.Slug), TenantSlug: tenant.Slug,
		EditClient: &c,
	})
}

func clientsUpdateHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := getSession(r)
	tenant, _ := tenantFromContext(r)
	idStr := r.PathValue("id")

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	email := strings.TrimSpace(r.FormValue("email"))
	company := strings.TrimSpace(r.FormValue("company"))

	if name == "" || email == "" {
		renderClientsPage(w, r, tenant, user, clientsInnerData{
			Clients: clientStore.list(tenant.Slug), TenantSlug: tenant.Slug,
			FormError: "Name and email are required.",
		})
		return
	}
	if len(name) > 100 || len(company) > 100 {
		renderClientsPage(w, r, tenant, user, clientsInnerData{
			Clients: clientStore.list(tenant.Slug), TenantSlug: tenant.Slug,
			FormError: "Name and company must be 100 characters or fewer.",
		})
		return
	}
	if len(email) > 254 {
		renderClientsPage(w, r, tenant, user, clientsInnerData{
			Clients: clientStore.list(tenant.Slug), TenantSlug: tenant.Slug,
			FormError: "Email must be 254 characters or fewer.",
		})
		return
	}
	if !isValidEmail(email) {
		renderClientsPage(w, r, tenant, user, clientsInnerData{
			Clients: clientStore.list(tenant.Slug), TenantSlug: tenant.Slug,
			FormError: "Invalid email address.",
		})
		return
	}

	if Pools != nil && tenant.DBID > 0 {
		clientID, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		if err := db.UpdateTenantClient(ctx, Pools.Admin, tenant.DBID, clientID, name, email, company); err != nil {
			log.Printf("clientsUpdateHandler: UpdateTenantClient: %v", err)
			http.Error(w, "Update failed", http.StatusInternalServerError)
			return
		}
	} else {
		existing, ok := clientStore.getByID(tenant.Slug, idStr)
		if !ok {
			http.NotFound(w, r)
			return
		}
		existing.Name = name
		existing.Email = email
		existing.Company = company
		clientStore.update(tenant.Slug, existing)
	}
	http.Redirect(w, r, "/t/"+tenant.Slug+"/clients/", http.StatusSeeOther)
}

func clientsDeleteHandler(w http.ResponseWriter, r *http.Request) {
	tenant, _ := tenantFromContext(r)
	idStr := r.PathValue("id")

	if Pools != nil && tenant.DBID > 0 {
		clientID, err := strconv.ParseInt(idStr, 10, 64)
		if err == nil {
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()
			if err := db.DeleteTenantClient(ctx, Pools.Admin, tenant.DBID, clientID); err != nil {
				log.Printf("clientsDeleteHandler: DeleteTenantClient: %v", err)
			}
		}
	} else {
		clientStore.delete(tenant.Slug, idStr)
	}
	http.Redirect(w, r, "/t/"+tenant.Slug+"/clients/", http.StatusSeeOther)
}

func clientsSuspendHandler(w http.ResponseWriter, r *http.Request) {
	tenant, _ := tenantFromContext(r)
	idStr := r.PathValue("id")

	if Pools != nil && tenant.DBID > 0 {
		clientID, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		dc, err := db.GetTenantClient(ctx, Pools.Admin, tenant.DBID, clientID)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		newStatus := "suspended"
		if dc.Status == "suspended" {
			newStatus = "active"
		}
		if err := db.SetTenantClientStatus(ctx, Pools.Admin, tenant.DBID, clientID, newStatus); err != nil {
			log.Printf("clientsSuspendHandler: SetTenantClientStatus: %v", err)
		}
	} else {
		c, ok := clientStore.getByID(tenant.Slug, idStr)
		if !ok {
			http.NotFound(w, r)
			return
		}
		if c.Status == "active" {
			c.Status = "suspended"
		} else {
			c.Status = "active"
		}
		clientStore.update(tenant.Slug, c)
	}
	http.Redirect(w, r, "/t/"+tenant.Slug+"/clients/", http.StatusSeeOther)
}

func clientsResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := getSession(r)
	tenant, _ := tenantFromContext(r)
	idStr := r.PathValue("id")
	c, ok := clientStore.getByID(tenant.Slug, idStr)
	if !ok {
		// If not in mock store and DB is available, fabricate a minimal record for display.
		if Pools != nil && tenant.DBID > 0 {
			clientID, err := strconv.ParseInt(idStr, 10, 64)
			if err == nil {
				ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
				defer cancel()
				dc, dbErr := db.GetTenantClient(ctx, Pools.Admin, tenant.DBID, clientID)
				if dbErr == nil {
					c = dbClientToLocal(dc)
					ok = true
				}
			}
		}
		if !ok {
			http.NotFound(w, r)
			return
		}
	}
	renderClientsPage(w, r, tenant, user, clientsInnerData{
		Clients: clientStore.list(tenant.Slug), TenantSlug: tenant.Slug,
		ResetResult: "Password reset email sent.", ResetEmail: c.Email,
	})
}

func clientsCreateDefaultWorkspaceHandler(w http.ResponseWriter, r *http.Request) {
	tenant, _ := tenantFromContext(r)
	idStr := r.PathValue("id")

	if Pools != nil && tenant.DBID > 0 {
		clientID, err := strconv.ParseInt(idStr, 10, 64)
		if err == nil {
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()
			if err := db.CreateClientWorkspace(ctx, Pools.Admin, clientID, "default"); err != nil {
				log.Printf("clientsCreateDefaultWorkspaceHandler: CreateClientWorkspace: %v", err)
			}
		}
	} else {
		c, ok := clientStore.getByID(tenant.Slug, idStr)
		if !ok {
			http.NotFound(w, r)
			return
		}
		c.Workspaces = append(c.Workspaces, Workspace{ID: newID(), Name: "default"})
		clientStore.update(tenant.Slug, c)
	}
	http.Redirect(w, r, "/t/"+tenant.Slug+"/clients/", http.StatusSeeOther)
}

func RegisterTenantClientRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /t/{slug}/clients/", tenantMiddleware(clientsListHandler))
	mux.HandleFunc("POST /t/{slug}/clients/", tenantMiddleware(clientsCreateHandler))
	mux.HandleFunc("GET /t/{slug}/clients/{id}/edit/", tenantMiddleware(clientsEditModalHandler))
	mux.HandleFunc("POST /t/{slug}/clients/{id}/edit/", tenantMiddleware(clientsUpdateHandler))
	mux.HandleFunc("POST /t/{slug}/clients/{id}/delete/", tenantMiddleware(clientsDeleteHandler))
	mux.HandleFunc("POST /t/{slug}/clients/{id}/suspend/", tenantMiddleware(clientsSuspendHandler))
	mux.HandleFunc("POST /t/{slug}/clients/{id}/reset-password/", tenantMiddleware(clientsResetPasswordHandler))
	mux.HandleFunc("POST /t/{slug}/clients/{id}/create-default-workspace/", tenantMiddleware(clientsCreateDefaultWorkspaceHandler))
}
