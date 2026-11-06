package handlers

import (
	"html/template"
	"net/http"
	"path/filepath"
)

type NavItem struct {
	ID    string
	Label string
	Icon  string
	Path  string
}

var sysadminNav = []NavItem{
	{ID: "tenants", Label: "Tenants", Path: "/s/tenants/", Icon: "tenants"},
	{ID: "logs", Label: "Logs", Path: "/s/logs/", Icon: "logs"},
	{ID: "replicas", Label: "Replicas", Path: "/s/replicas/", Icon: "replicas"},
	{ID: "current-replica", Label: "Current Replica", Path: "/s/current-replica/", Icon: "current-replica"},
	{ID: "database", Label: "Database", Path: "/s/database/", Icon: "database"},
	{ID: "sso", Label: "SSO", Path: "/s/sso/", Icon: "sso"},
	{ID: "accounting", Label: "Accounting System", Path: "/s/accounting/", Icon: "accounting"},
	{ID: "jobs", Label: "Background Jobs", Path: "/s/jobs/", Icon: "jobs"},
}

type SysadminData struct {
	User      User
	Nav       []NavItem
	ActiveNav string
	PageTitle string
	Content   template.HTML // pre-rendered page content; empty = show placeholder
}

func sysadminPage(activeNav, pageTitle string) http.HandlerFunc {
	return requireRole(RoleSysAdmin, func(w http.ResponseWriter, r *http.Request) {
		user, _ := getSession(r)

		paths := []string{
			filepath.Join("templates", "sysadmin", "base.html"),
			filepath.Join("templates", "sysadmin", "shell.html"),
		}
		tmpl, err := template.ParseFiles(paths...)
		if err != nil {
			http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		data := SysadminData{
			User:      user,
			Nav:       sysadminNav,
			ActiveNav: activeNav,
			PageTitle: pageTitle,
		}
		if err := tmpl.ExecuteTemplate(w, "sysadmin-base", data); err != nil {
			http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
		}
	})
}

func RegisterSysadminRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/s/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/s/tenants/", http.StatusSeeOther)
	})
	mux.HandleFunc("/s/logs/", LogsPage)
	RegisterReplicaRoutes(mux)
	RegisterSysadminCurrentReplicaRoutes(mux)
	RegisterDatabaseRoutes(mux)
	RegisterTenantRoutes(mux)
	RegisterSSORoutes(mux)
	RegisterAccountingRoutes(mux)
	RegisterJobsRoutes(mux)
}
