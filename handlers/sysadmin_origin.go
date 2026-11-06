package handlers

import (
	"bytes"
	"context"
	"errors"
	"html/template"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	"platform/db"
)

type originDomainView struct {
	ID        int64
	Domain    string
	Note      string
	Position  int
	Enabled   bool
	IsPrimary bool
	IsFirst   bool
	IsLast    bool
}

type originPageInnerData struct {
	Settings           db.OriginSettings
	Domains            []originDomainView
	EnforcementActive  bool
	PrimaryDomain      string
	PrimaryRedirectURL string
	CurrentHost        string
	CurrentHostIsValid bool
	CurrentHostInList  bool
	TenantDomains      []Tenant
	FlashSuccess       string
	FlashError         string
}

func renderSysadminOriginPage(w http.ResponseWriter, r *http.Request, flashErr, flashOK string) {
	user, _ := getSession(r)

	snap := origins.snapshot(r.Context())
	fullHost, bareHost := effectiveRequestHost(r, snap.settings.TrustForwardedHeaders)
	displayHost := fullHost
	if displayHost == "" {
		displayHost = bareHost
	}

	var views []originDomainView
	primaryFound := false
	var primaryDomain string
	inList := false

	for i, d := range snap.domains {
		isPrimary := false
		if d.Enabled && !primaryFound {
			isPrimary = true
			primaryFound = true
			primaryDomain = d.Domain
		}
		if db.HostMatchesConfiguredDomain(fullHost, bareHost, d.Domain) {
			inList = true
		}
		views = append(views, originDomainView{
			ID:        d.ID,
			Domain:    d.Domain,
			Note:      d.Note,
			Position:  i + 1,
			Enabled:   d.Enabled,
			IsPrimary: isPrimary,
			IsFirst:   i == 0,
			IsLast:    i == len(snap.domains)-1,
		})
	}

	var tenantDomains []Tenant
	for _, t := range store.list() {
		if strings.TrimSpace(t.Domain) != "" {
			tenantDomains = append(tenantDomains, t)
		}
	}

	_, normErr := db.NormalizeDomain(displayHost)
	currentValid := normErr == nil && displayHost != ""

	var primaryRedirectURL string
	if primaryFound {
		primaryRedirectURL = snap.settings.RedirectScheme + "://" + primaryDomain
	}

	if flashOK == "" {
		flashOK = strings.TrimSpace(r.URL.Query().Get("ok"))
	}
	if flashErr == "" {
		flashErr = strings.TrimSpace(r.URL.Query().Get("err"))
	}

	data := originPageInnerData{
		Settings:           snap.settings,
		Domains:            views,
		EnforcementActive:  primaryFound,
		PrimaryDomain:      primaryDomain,
		PrimaryRedirectURL: primaryRedirectURL,
		CurrentHost:        displayHost,
		CurrentHostIsValid: currentValid,
		CurrentHostInList:  inList,
		TenantDomains:      tenantDomains,
		FlashSuccess:       flashOK,
		FlashError:         flashErr,
	}

	innerTmpl, err := template.ParseFiles(filepath.Join("templates", "sysadmin", "origin_content.html"))
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	var buf bytes.Buffer
	if err := innerTmpl.ExecuteTemplate(&buf, "origin-content", data); err != nil {
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
		ActiveNav: "origin",
		PageTitle: "Origin",
		Content:   template.HTML(buf.String()),
	}
	if err := shellTmpl.ExecuteTemplate(w, "sysadmin-base", shellData); err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
	}
}

func redirectOriginWithFlash(w http.ResponseWriter, r *http.Request, okMsg, errMsg string) {
	u := "/s/origin/"
	vals := url.Values{}
	if okMsg != "" {
		vals.Set("ok", okMsg)
	}
	if errMsg != "" {
		vals.Set("err", errMsg)
	}
	if enc := vals.Encode(); enc != "" {
		u += "?" + enc
	}
	http.Redirect(w, r, u, http.StatusSeeOther)
}

func originPageHandler(w http.ResponseWriter, r *http.Request) {
	renderSysadminOriginPage(w, r, "", "")
}

func originAddDomainHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := getSession(r)
	if err := r.ParseForm(); err != nil {
		renderSysadminOriginPage(w, r, "Invalid form submission.", "")
		return
	}
	rawDomain := strings.TrimSpace(r.FormValue("domain"))
	note := strings.TrimSpace(r.FormValue("note"))
	enabled := r.FormValue("enabled") != "false" && r.FormValue("enabled") != "0"

	added, err := origins.addDomain(r.Context(), rawDomain, note, enabled)
	if err != nil {
		msg := err.Error()
		if errors.Is(err, db.ErrInvalidDomain) {
			msg = "Please enter a valid domain name (for example, system.example.com)."
		} else if errors.Is(err, db.ErrOriginDomainConflict) {
			msg = err.Error()
		}
		renderSysadminOriginPage(w, r, msg, "")
		return
	}

	if Pools != nil {
		go db.InsertAuditEvent(context.Background(), Pools.Admin, user.ID, string(user.Role),
			"sysadmin.origin.domain_added", "info",
			map[string]any{"id": added.ID, "domain": added.Domain, "enabled": added.Enabled},
			"", r.Header.Get("X-Request-ID"), clientIP(r))
	}

	redirectOriginWithFlash(w, r, "Added system origin domain \""+added.Domain+"\".", "")
}

func originMoveDomainHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := getSession(r)
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		redirectOriginWithFlash(w, r, "", "Invalid domain ID.")
		return
	}
	if err := r.ParseForm(); err != nil {
		redirectOriginWithFlash(w, r, "", "Invalid form submission.")
		return
	}
	direction := strings.ToLower(strings.TrimSpace(r.FormValue("direction")))
	if err := origins.moveDomain(r.Context(), id, direction); err != nil {
		redirectOriginWithFlash(w, r, "", "Unable to reorder domain.")
		return
	}

	if Pools != nil {
		go db.InsertAuditEvent(context.Background(), Pools.Admin, user.ID, string(user.Role),
			"sysadmin.origin.domain_reordered", "info",
			map[string]any{"id": id, "direction": direction},
			"", r.Header.Get("X-Request-ID"), clientIP(r))
	}

	redirectOriginWithFlash(w, r, "Updated domain priority order.", "")
}

func originToggleDomainHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := getSession(r)
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		redirectOriginWithFlash(w, r, "", "Invalid domain ID.")
		return
	}
	existing, ok := origins.getDomain(r.Context(), id)
	if !ok {
		redirectOriginWithFlash(w, r, "", "Origin domain not found.")
		return
	}
	updated, ok, err := origins.updateDomain(r.Context(), id, existing.Note, !existing.Enabled)
	if err != nil || !ok {
		redirectOriginWithFlash(w, r, "", "Failed to update domain status.")
		return
	}

	if Pools != nil {
		go db.InsertAuditEvent(context.Background(), Pools.Admin, user.ID, string(user.Role),
			"sysadmin.origin.domain_toggled", "info",
			map[string]any{"id": id, "domain": updated.Domain, "enabled": updated.Enabled},
			"", r.Header.Get("X-Request-ID"), clientIP(r))
	}

	stateWord := "enabled"
	if !updated.Enabled {
		stateWord = "disabled"
	}
	redirectOriginWithFlash(w, r, "Domain \""+updated.Domain+"\" "+stateWord+".", "")
}

func originDeleteDomainHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := getSession(r)
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		redirectOriginWithFlash(w, r, "", "Invalid domain ID.")
		return
	}
	existing, _ := origins.getDomain(r.Context(), id)
	deleted, err := origins.deleteDomain(r.Context(), id)
	if err != nil || !deleted {
		redirectOriginWithFlash(w, r, "", "Origin domain not found.")
		return
	}

	if Pools != nil {
		go db.InsertAuditEvent(context.Background(), Pools.Admin, user.ID, string(user.Role),
			"sysadmin.origin.domain_deleted", "warn",
			map[string]any{"id": id, "domain": existing.Domain},
			"", r.Header.Get("X-Request-ID"), clientIP(r))
	}

	redirectOriginWithFlash(w, r, "Removed system origin domain \""+existing.Domain+"\".", "")
}

func originSaveSettingsHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := getSession(r)
	if err := r.ParseForm(); err != nil {
		renderSysadminOriginPage(w, r, "Invalid form submission.", "")
		return
	}
	scheme := strings.ToLower(strings.TrimSpace(r.FormValue("redirect_scheme")))
	code, _ := strconv.Atoi(strings.TrimSpace(r.FormValue("redirect_status_code")))
	trustFwd := r.FormValue("trust_forwarded_headers") == "on" || r.FormValue("trust_forwarded_headers") == "true" || r.FormValue("trust_forwarded_headers") == "1"

	saved, err := origins.updateSettings(r.Context(), db.OriginSettings{
		RedirectScheme:        scheme,
		RedirectStatusCode:    code,
		TrustForwardedHeaders: trustFwd,
	})
	if err != nil {
		renderSysadminOriginPage(w, r, err.Error(), "")
		return
	}

	if Pools != nil {
		go db.InsertAuditEvent(context.Background(), Pools.Admin, user.ID, string(user.Role),
			"sysadmin.origin.settings_updated", "info",
			map[string]any{
				"redirect_scheme":         saved.RedirectScheme,
				"redirect_status_code":    saved.RedirectStatusCode,
				"trust_forwarded_headers": saved.TrustForwardedHeaders,
			},
			"", r.Header.Get("X-Request-ID"), clientIP(r))
	}

	redirectOriginWithFlash(w, r, "Origin routing settings saved.", "")
}

func RegisterOriginRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /s/origin/", requireRole(RoleSysAdmin, originPageHandler))
	mux.HandleFunc("POST /s/origin/domains/", requireRole(RoleSysAdmin, originAddDomainHandler))
	mux.HandleFunc("POST /s/origin/domains/{id}/move/", requireRole(RoleSysAdmin, originMoveDomainHandler))
	mux.HandleFunc("POST /s/origin/domains/{id}/toggle/", requireRole(RoleSysAdmin, originToggleDomainHandler))
	mux.HandleFunc("POST /s/origin/domains/{id}/delete/", requireRole(RoleSysAdmin, originDeleteDomainHandler))
	mux.HandleFunc("POST /s/origin/settings/", requireRole(RoleSysAdmin, originSaveSettingsHandler))
}
