package handlers

import (
	"bytes"
	"context"
	"html/template"
	"net/http"
	"path/filepath"
	"time"

	"platform/db"
)

// SSOPageData carries live SSO config to the template.
type SSOPageData struct {
	SysadminData
	Config  db.SSOConfig
	Clients []db.SSOClient
	CORS    []db.SSOCORSOrigin
}

func ssoHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := getSession(r)

	data := SSOPageData{
		SysadminData: SysadminData{
			User:      user,
			Nav:       sysadminNav,
			ActiveNav: "sso",
			PageTitle: "SSO",
		},
		Config: db.DefaultSSOConfig(),
	}

	if Pools != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		if cfg, err := db.GetSSOConfig(ctx, Pools.Admin); err == nil {
			data.Config = cfg
		}
		if clients, err := db.ListSSOClients(ctx, Pools.Admin); err == nil {
			data.Clients = clients
		}
		if cors, err := db.ListSSOCORSOrigins(ctx, Pools.Admin); err == nil {
			data.CORS = cors
		}
	}

	innerTmpl, err := template.ParseFiles(filepath.Join("templates", "sysadmin", "sso_content.html"))
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	var buf bytes.Buffer
	if err := innerTmpl.ExecuteTemplate(&buf, "sysadmin-sso-content", data); err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	shellTmpl, err := template.ParseFiles(
		filepath.Join("templates", "sysadmin", "base.html"),
		filepath.Join("templates", "sysadmin", "shell.html"),
	)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	data.SysadminData.Content = template.HTML(buf.String())
	if err := shellTmpl.ExecuteTemplate(w, "sysadmin-base", data.SysadminData); err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
	}
}

func RegisterSSORoutes(mux *http.ServeMux) {
	mux.HandleFunc("/s/sso/", requireRole(RoleSysAdmin, ssoHandler))
}
