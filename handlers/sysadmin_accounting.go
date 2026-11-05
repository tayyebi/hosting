package handlers

import (
	"bytes"
	"html/template"
	"net/http"
	"path/filepath"
)

func accountingHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := getSession(r)

	innerTmpl, err := template.ParseFiles(filepath.Join("templates", "sysadmin", "accounting_content.html"))
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	var buf bytes.Buffer
	if err := innerTmpl.ExecuteTemplate(&buf, "sysadmin-accounting-content", nil); err != nil {
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
		ActiveNav: "accounting",
		PageTitle: "Accounting System",
		Content:   template.HTML(buf.String()),
	}
	if err := shellTmpl.ExecuteTemplate(w, "sysadmin-base", data); err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
	}
}

func RegisterAccountingRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/s/accounting/", requireRole(RoleSysAdmin, accountingHandler))
}
