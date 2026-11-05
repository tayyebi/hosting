package handlers

import (
	"bytes"
	"html/template"
	"net/http"
	"path/filepath"
)

func currentReplicaHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := getSession(r)

	innerTmpl, err := template.ParseFiles(filepath.Join("templates", "sysadmin", "current_replica_content.html"))
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	var buf bytes.Buffer
	if err := innerTmpl.ExecuteTemplate(&buf, "sysadmin-current-replica-content", nil); err != nil {
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
	data := SysadminData{
		User:      user,
		Nav:       sysadminNav,
		ActiveNav: "current-replica",
		PageTitle: "Current Replica",
		Content:   template.HTML(buf.String()),
	}
	if err := shellTmpl.ExecuteTemplate(w, "sysadmin-base", data); err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
	}
}

func RegisterSysadminCurrentReplicaRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/s/current-replica/", requireRole(RoleSysAdmin, currentReplicaHandler))
}
