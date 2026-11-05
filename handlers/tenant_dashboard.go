package handlers

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"platform/db"
)

// TenantDashboardData is the data passed to the tenant dashboard template.
type TenantDashboardData struct {
	TenantShellData
	ClientCount    int
	EmployeeCount  int
	OpenTickets    int
	WorkspaceCount int
}

// tenantPage returns a handler for a tenant-scoped page.
// activeNav and pageTitle describe the current section.
// contentFiles are additional template files that define "tenant-page-content".
func tenantPage(activeNav, pageTitle string, contentFiles ...string) http.HandlerFunc {
	return tenantMiddleware(func(w http.ResponseWriter, r *http.Request) {
		user, _ := getSession(r)
		tenant, _ := tenantFromContext(r)

		paths := []string{
			filepath.Join("templates", "tenant", "base.html"),
			filepath.Join("templates", "tenant", "shell.html"),
		}
		paths = append(paths, contentFiles...)

		tmpl, err := template.ParseFiles(paths...)
		if err != nil {
			http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		data := TenantShellData{
			User:       user,
			Tenant:     tenant,
			Nav:        tenantNav(tenant.Slug),
			NavGroups:  tenantNavGroups(tenant.Slug),
			ActiveNav:  activeNav,
			PageTitle:  pageTitle,
			IsSysAdmin: user.Role == RoleSysAdmin,
		}
		if err := tmpl.ExecuteTemplate(w, "tenant-base", data); err != nil {
			http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
		}
	})
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := getSession(r)
	tenant, _ := tenantFromContext(r)

	var clientCount, employeeCount, openTickets, workspaceCount int

	if Pools != nil && tenant.DBID > 0 {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		if n, err := db.TenantClientCount(ctx, Pools.Admin, tenant.DBID); err == nil {
			clientCount = n
		} else {
			log.Printf("dashboardHandler: TenantClientCount: %v", err)
		}

		if n, err := db.TenantEmployeeCount(ctx, Pools.Admin, tenant.DBID); err == nil {
			employeeCount = n
		} else {
			log.Printf("dashboardHandler: TenantEmployeeCount: %v", err)
		}

		if n, err := db.OpenTicketCount(ctx, Pools.Admin, tenant.DBID); err == nil {
			openTickets = n
		} else {
			log.Printf("dashboardHandler: OpenTicketCount: %v", err)
		}

		if err := Pools.Admin.QueryRow(ctx,
			`SELECT count(*) FROM workspaces WHERE tenant_id=$1 AND deleted_at IS NULL`,
			tenant.DBID,
		).Scan(&workspaceCount); err != nil {
			log.Printf("dashboardHandler: workspace count: %v", err)
		}
	}

	paths := []string{
		filepath.Join("templates", "tenant", "base.html"),
		filepath.Join("templates", "tenant", "shell.html"),
		filepath.Join("templates", "tenant", "dashboard_content.html"),
	}
	tmpl, err := template.ParseFiles(paths...)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := TenantDashboardData{
		TenantShellData: TenantShellData{
			User:       user,
			Tenant:     tenant,
			Nav:        tenantNav(tenant.Slug),
			NavGroups:  tenantNavGroups(tenant.Slug),
			ActiveNav:  "dashboard",
			PageTitle:  "Dashboard",
			IsSysAdmin: user.Role == RoleSysAdmin,
		},
		ClientCount:    clientCount,
		EmployeeCount:  employeeCount,
		OpenTickets:    openTickets,
		WorkspaceCount: workspaceCount,
	}
	if err := tmpl.ExecuteTemplate(w, "tenant-base", data); err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
	}
}

func RegisterTenantDashboardRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /t/{slug}/", tenantMiddleware(dashboardHandler))
}
