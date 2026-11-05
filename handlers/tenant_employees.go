package handlers

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
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

// ── Permissions ──────────────────────────────────────────────

type PermissionGroup struct {
	Label string
	Items []PermissionItem
}

type PermissionItem struct {
	ID    string
	Label string
}

var permissionGroups = []PermissionGroup{
	{Label: "General", Items: []PermissionItem{
		{ID: "dashboard", Label: "Dashboard"},
	}},
	{Label: "Support", Items: []PermissionItem{
		{ID: "clients", Label: "Clients"},
		{ID: "tickets", Label: "Tickets"},
		{ID: "abuse-reports", Label: "Abuse Reports"},
		{ID: "refund-complaints", Label: "Refund Complaints"},
	}},
	{Label: "Infrastructure", Items: []PermissionItem{
		{ID: "nodes", Label: "Nodes"},
		{ID: "monitoring", Label: "Monitoring"},
		{ID: "incidents", Label: "Incidents"},
	}},
	{Label: "Finance", Items: []PermissionItem{
		{ID: "transactions", Label: "Transactions"},
		{ID: "pricing", Label: "Pricing"},
		{ID: "tax", Label: "Tax"},
		{ID: "sla", Label: "SLA"},
	}},
	{Label: "Business", Items: []PermissionItem{
		{ID: "employees", Label: "Employees"},
		{ID: "brand", Label: "Brand"},
		{ID: "localization", Label: "Localization"},
	}},
}

func allPermIDs() []string {
	var ids []string
	for _, g := range permissionGroups {
		for _, item := range g.Items {
			ids = append(ids, item.ID)
		}
	}
	return ids
}

func fullPerms() map[string]bool {
	m := map[string]bool{}
	for _, id := range allPermIDs() {
		m[id] = true
	}
	return m
}

func supportPerms() map[string]bool {
	m := map[string]bool{"dashboard": true, "clients": true, "tickets": true, "abuse-reports": true, "refund-complaints": true}
	return m
}

func financePerms() map[string]bool {
	m := map[string]bool{"dashboard": true, "transactions": true, "pricing": true, "tax": true, "sla": true}
	return m
}

func infraPerms() map[string]bool {
	m := map[string]bool{"dashboard": true, "nodes": true, "monitoring": true, "incidents": true}
	return m
}

// ── Employee store ────────────────────────────────────────────

type Employee struct {
	ID          string
	Name        string
	Email       string
	Role        string
	Active      bool
	Permissions map[string]bool
}

type employeeStoreType struct {
	mu        sync.RWMutex
	employees map[string][]Employee // key: tenant slug
}

var employeeStore = &employeeStoreType{
	employees: map[string][]Employee{
		"buyvm": {
			{ID: newID(), Name: "Alice Smith", Email: "alice@buyvm.net", Role: "Admin", Active: true, Permissions: fullPerms()},
			{ID: newID(), Name: "Bob Carter", Email: "bob@buyvm.net", Role: "Support Agent", Active: true, Permissions: supportPerms()},
			{ID: newID(), Name: "Carol White", Email: "carol@buyvm.net", Role: "Finance Manager", Active: false, Permissions: financePerms()},
		},
		"abrnoc": {
			{ID: newID(), Name: "David Lee", Email: "david@abrnoc.com", Role: "Admin", Active: true, Permissions: fullPerms()},
			{ID: newID(), Name: "Eva Müller", Email: "eva@abrnoc.com", Role: "Infrastructure Lead", Active: true, Permissions: infraPerms()},
		},
		"hosting": {
			{ID: newID(), Name: "Frank Hassan", Email: "frank@hosting.com", Role: "Admin", Active: true, Permissions: fullPerms()},
			{ID: newID(), Name: "Grace Kim", Email: "grace@hosting.com", Role: "Support Agent", Active: true, Permissions: supportPerms()},
			{ID: newID(), Name: "Henry Patel", Email: "henry@hosting.com", Role: "Finance Manager", Active: true, Permissions: financePerms()},
			{ID: newID(), Name: "Irene Russo", Email: "irene@hosting.com", Role: "Infrastructure Lead", Active: false, Permissions: infraPerms()},
		},
	},
}

func (s *employeeStoreType) list(slug string) []Employee {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.employees[slug]
	out := make([]Employee, len(src))
	copy(out, src)
	return out
}

func (s *employeeStoreType) getByID(slug, id string) (Employee, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, e := range s.employees[slug] {
		if e.ID == id {
			return e, true
		}
	}
	return Employee{}, false
}

func (s *employeeStoreType) add(slug string, e Employee) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.employees[slug] = append(s.employees[slug], e)
}

func (s *employeeStoreType) update(slug string, e Employee) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, existing := range s.employees[slug] {
		if existing.ID == e.ID {
			s.employees[slug][i] = e
			return true
		}
	}
	return false
}

func (s *employeeStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.employees[slug]
	for i, e := range list {
		if e.ID == id {
			s.employees[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

// ── DB conversion ─────────────────────────────────────────────

func dbEmployeeToLocal(de db.TenantEmployee) Employee {
	return Employee{
		ID:          fmt.Sprintf("%d", de.MembershipID),
		Name:        de.Name,
		Email:       de.Email,
		Role:        de.Title,
		Active:      de.Active,
		Permissions: de.Permissions,
	}
}

// ── Page data ─────────────────────────────────────────────────

type employeesInnerData struct {
	Employees    []Employee
	TenantSlug   string
	FormError    string
	EditEmployee *Employee
	PermEmployee *Employee
	PermGroups   []PermissionGroup
	ResetResult  string
}

func renderEmployeesPage(w http.ResponseWriter, r *http.Request, tenant Tenant, user User, d employeesInnerData) {
	innerTmpl, err := template.ParseFiles(filepath.Join("templates", "tenant", "employees_content.html"))
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	var buf bytes.Buffer
	if err := innerTmpl.ExecuteTemplate(&buf, "employees-content", d); err != nil {
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
		User:       user,
		Tenant:     tenant,
		Nav:        tenantNav(tenant.Slug),
		NavGroups:  tenantNavGroups(tenant.Slug),
		ActiveNav:  "employees",
		PageTitle:  "Employees",
		IsSysAdmin: user.Role == RoleSysAdmin,
		Content:    template.HTML(buf.String()),
	}
	if err := shellTmpl.ExecuteTemplate(w, "tenant-base", data); err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
	}
}

// ── Handlers ─────────────────────────────────────────────────

func employeesListHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := getSession(r)
	tenant, _ := tenantFromContext(r)

	var employees []Employee
	if Pools != nil && tenant.DBID > 0 {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		des, err := db.ListTenantEmployees(ctx, Pools.Admin, tenant.DBID)
		if err != nil {
			log.Printf("employeesListHandler: ListTenantEmployees: %v", err)
			employees = employeeStore.list(tenant.Slug)
		} else {
			employees = make([]Employee, len(des))
			for i, de := range des {
				employees[i] = dbEmployeeToLocal(de)
			}
		}
	} else {
		employees = employeeStore.list(tenant.Slug)
	}

	renderEmployeesPage(w, r, tenant, user, employeesInnerData{
		Employees:  employees,
		TenantSlug: tenant.Slug,
		PermGroups: permissionGroups,
	})
}

func employeesCreateHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := getSession(r)
	tenant, _ := tenantFromContext(r)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	email := strings.TrimSpace(r.FormValue("email"))
	role := strings.TrimSpace(r.FormValue("role"))
	if name == "" || email == "" {
		renderEmployeesPage(w, r, tenant, user, employeesInnerData{
			Employees: employeeStore.list(tenant.Slug), TenantSlug: tenant.Slug,
			FormError: "Name and email are required.", PermGroups: permissionGroups,
		})
		return
	}
	if len(name) > 100 || len(role) > 100 {
		renderEmployeesPage(w, r, tenant, user, employeesInnerData{
			Employees: employeeStore.list(tenant.Slug), TenantSlug: tenant.Slug,
			FormError: "Name and role must be 100 characters or fewer.", PermGroups: permissionGroups,
		})
		return
	}
	if len(email) > 254 {
		renderEmployeesPage(w, r, tenant, user, employeesInnerData{
			Employees: employeeStore.list(tenant.Slug), TenantSlug: tenant.Slug,
			FormError: "Email must be 254 characters or fewer.", PermGroups: permissionGroups,
		})
		return
	}
	if !isValidEmail(email) {
		renderEmployeesPage(w, r, tenant, user, employeesInnerData{
			Employees: employeeStore.list(tenant.Slug), TenantSlug: tenant.Slug,
			FormError: "Invalid email address.", PermGroups: permissionGroups,
		})
		return
	}

	if Pools != nil && tenant.DBID > 0 {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		if _, err := db.CreateTenantEmployee(ctx, Pools.Admin, tenant.DBID, name, email, role); err != nil {
			log.Printf("employeesCreateHandler: CreateTenantEmployee: %v", err)
			renderEmployeesPage(w, r, tenant, user, employeesInnerData{
				Employees: employeeStore.list(tenant.Slug), TenantSlug: tenant.Slug,
				FormError: "Failed to create employee: " + err.Error(), PermGroups: permissionGroups,
			})
			return
		}
	} else {
		employeeStore.add(tenant.Slug, Employee{
			ID: newID(), Name: name, Email: email, Role: role, Active: true,
			Permissions: map[string]bool{"dashboard": true},
		})
	}
	http.Redirect(w, r, "/t/"+tenant.Slug+"/employees/", http.StatusSeeOther)
}

func employeesEditModalHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := getSession(r)
	tenant, _ := tenantFromContext(r)
	idStr := r.PathValue("id")

	var e Employee
	if Pools != nil && tenant.DBID > 0 {
		membershipID, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		de, dbErr := db.GetTenantEmployee(ctx, Pools.Admin, tenant.DBID, membershipID)
		if dbErr != nil {
			log.Printf("employeesEditModalHandler: GetTenantEmployee: %v", dbErr)
			http.NotFound(w, r)
			return
		}
		e = dbEmployeeToLocal(de)
	} else {
		var ok bool
		e, ok = employeeStore.getByID(tenant.Slug, idStr)
		if !ok {
			http.NotFound(w, r)
			return
		}
	}

	renderEmployeesPage(w, r, tenant, user, employeesInnerData{
		Employees: employeeStore.list(tenant.Slug), TenantSlug: tenant.Slug,
		EditEmployee: &e, PermGroups: permissionGroups,
	})
}

func employeesUpdateHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := getSession(r)
	tenant, _ := tenantFromContext(r)
	idStr := r.PathValue("id")
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	email := strings.TrimSpace(r.FormValue("email"))
	role := strings.TrimSpace(r.FormValue("role"))
	active := r.FormValue("active") == "on"

	if name == "" || email == "" {
		renderEmployeesPage(w, r, tenant, user, employeesInnerData{
			Employees: employeeStore.list(tenant.Slug), TenantSlug: tenant.Slug,
			FormError: "Name and email are required.", PermGroups: permissionGroups,
		})
		return
	}
	if len(name) > 100 || len(role) > 100 {
		renderEmployeesPage(w, r, tenant, user, employeesInnerData{
			Employees: employeeStore.list(tenant.Slug), TenantSlug: tenant.Slug,
			FormError: "Name and role must be 100 characters or fewer.", PermGroups: permissionGroups,
		})
		return
	}
	if len(email) > 254 {
		renderEmployeesPage(w, r, tenant, user, employeesInnerData{
			Employees: employeeStore.list(tenant.Slug), TenantSlug: tenant.Slug,
			FormError: "Email must be 254 characters or fewer.", PermGroups: permissionGroups,
		})
		return
	}
	if !isValidEmail(email) {
		renderEmployeesPage(w, r, tenant, user, employeesInnerData{
			Employees: employeeStore.list(tenant.Slug), TenantSlug: tenant.Slug,
			FormError: "Invalid email address.", PermGroups: permissionGroups,
		})
		return
	}

	if Pools != nil && tenant.DBID > 0 {
		membershipID, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		if err := db.UpdateTenantEmployee(ctx, Pools.Admin, tenant.DBID, membershipID, name, email, role, active); err != nil {
			log.Printf("employeesUpdateHandler: UpdateTenantEmployee: %v", err)
			http.Error(w, "Update failed", http.StatusInternalServerError)
			return
		}
	} else {
		existing, ok := employeeStore.getByID(tenant.Slug, idStr)
		if !ok {
			http.NotFound(w, r)
			return
		}
		existing.Name = name
		existing.Email = email
		existing.Role = role
		existing.Active = active
		employeeStore.update(tenant.Slug, existing)
	}
	http.Redirect(w, r, "/t/"+tenant.Slug+"/employees/", http.StatusSeeOther)
}

func employeesDeleteHandler(w http.ResponseWriter, r *http.Request) {
	tenant, _ := tenantFromContext(r)
	idStr := r.PathValue("id")

	if Pools != nil && tenant.DBID > 0 {
		membershipID, err := strconv.ParseInt(idStr, 10, 64)
		if err == nil {
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()
			if err := db.DeleteTenantEmployee(ctx, Pools.Admin, tenant.DBID, membershipID); err != nil {
				log.Printf("employeesDeleteHandler: DeleteTenantEmployee: %v", err)
			}
		}
	} else {
		employeeStore.delete(tenant.Slug, idStr)
	}
	http.Redirect(w, r, "/t/"+tenant.Slug+"/employees/", http.StatusSeeOther)
}

func employeesPermsModalHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := getSession(r)
	tenant, _ := tenantFromContext(r)
	idStr := r.PathValue("id")

	var e Employee
	if Pools != nil && tenant.DBID > 0 {
		membershipID, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		de, dbErr := db.GetTenantEmployee(ctx, Pools.Admin, tenant.DBID, membershipID)
		if dbErr != nil {
			log.Printf("employeesPermsModalHandler: GetTenantEmployee: %v", dbErr)
			http.NotFound(w, r)
			return
		}
		e = dbEmployeeToLocal(de)
	} else {
		var ok bool
		e, ok = employeeStore.getByID(tenant.Slug, idStr)
		if !ok {
			http.NotFound(w, r)
			return
		}
	}

	renderEmployeesPage(w, r, tenant, user, employeesInnerData{
		Employees: employeeStore.list(tenant.Slug), TenantSlug: tenant.Slug,
		PermEmployee: &e, PermGroups: permissionGroups,
	})
}

func employeesPermsUpdateHandler(w http.ResponseWriter, r *http.Request) {
	tenant, _ := tenantFromContext(r)
	idStr := r.PathValue("id")
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	perms := map[string]bool{}
	for _, pid := range allPermIDs() {
		perms[pid] = r.FormValue("perm_"+pid) == "on"
	}

	if Pools != nil && tenant.DBID > 0 {
		membershipID, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		if err := db.UpdateTenantEmployeePermissions(ctx, Pools.Admin, tenant.DBID, membershipID, perms); err != nil {
			log.Printf("employeesPermsUpdateHandler: UpdateTenantEmployeePermissions: %v", err)
		}
	} else {
		e, ok := employeeStore.getByID(tenant.Slug, idStr)
		if !ok {
			http.NotFound(w, r)
			return
		}
		e.Permissions = perms
		employeeStore.update(tenant.Slug, e)
	}
	http.Redirect(w, r, "/t/"+tenant.Slug+"/employees/", http.StatusSeeOther)
}

func employeesResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := getSession(r)
	tenant, _ := tenantFromContext(r)
	id := r.PathValue("id")
	if _, ok := employeeStore.getByID(tenant.Slug, id); !ok {
		// If not in mock store, check DB exists (just to return 404 for totally unknown IDs).
		if Pools != nil && tenant.DBID > 0 {
			membershipID, err := strconv.ParseInt(id, 10, 64)
			if err == nil {
				ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
				defer cancel()
				if _, dbErr := db.GetTenantEmployee(ctx, Pools.Admin, tenant.DBID, membershipID); dbErr != nil {
					http.NotFound(w, r)
					return
				}
			} else {
				http.NotFound(w, r)
				return
			}
		} else {
			http.NotFound(w, r)
			return
		}
	}
	b := make([]byte, 6)
	rand.Read(b)
	tmpPw := "Tmp-" + hex.EncodeToString(b)
	renderEmployeesPage(w, r, tenant, user, employeesInnerData{
		Employees: employeeStore.list(tenant.Slug), TenantSlug: tenant.Slug,
		PermGroups: permissionGroups, ResetResult: tmpPw,
	})
}

func RegisterTenantEmployeeRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /t/{slug}/employees/", tenantMiddleware(employeesListHandler))
	mux.HandleFunc("POST /t/{slug}/employees/", tenantMiddleware(employeesCreateHandler))
	mux.HandleFunc("GET /t/{slug}/employees/{id}/edit/", tenantMiddleware(employeesEditModalHandler))
	mux.HandleFunc("POST /t/{slug}/employees/{id}/edit/", tenantMiddleware(employeesUpdateHandler))
	mux.HandleFunc("POST /t/{slug}/employees/{id}/delete/", tenantMiddleware(employeesDeleteHandler))
	mux.HandleFunc("GET /t/{slug}/employees/{id}/permissions/", tenantMiddleware(employeesPermsModalHandler))
	mux.HandleFunc("POST /t/{slug}/employees/{id}/permissions/", tenantMiddleware(employeesPermsUpdateHandler))
	mux.HandleFunc("POST /t/{slug}/employees/{id}/reset-password/", tenantMiddleware(employeesResetPasswordHandler))
}
