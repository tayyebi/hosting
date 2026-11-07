package handlers

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"platform/db"
)

type RegisterPageData struct {
	Tenant       Tenant
	Brand        TenantLoginBrand
	SSOProviders []db.TenantSSOProvider
	FormAction   string
	SignInURL    string
	Error        string
}

func resolveRegisterTenant(r *http.Request) (Tenant, string, string, bool) {
	oc := originFromContext(r)
	if oc.Mode == OriginModeTenant && oc.Tenant != nil {
		return *oc.Tenant, "/register", "/", true
	}
	slug := r.PathValue("slug")
	if slug == "" {
		return Tenant{}, "", "", false
	}
	tenant, found := store.getBySlug(slug)
	if !found {
		return Tenant{}, "", "", false
	}
	return tenant, "/t/" + tenant.Slug + "/register", "/t/" + tenant.Slug + "/login", true
}

func renderTenantRegisterPage(w http.ResponseWriter, r *http.Request, tenant Tenant, formAction, signInURL, errMsg string) {
	paths := []string{
		filepath.Join("templates", "base.html"),
		filepath.Join("templates", "tenant", "register.html"),
	}
	tmpl, err := template.ParseFiles(paths...)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	brand := resolveTenantLoginBrand(r.Context(), tenant)
	if strings.TrimSpace(brand.Name) != "" {
		tenant.Name = brand.Name
	}
	data := RegisterPageData{
		Tenant:       tenant,
		Brand:        brand,
		SSOProviders: enabledTenantSSOProviders(r.Context(), tenant),
		FormAction:   formAction,
		SignInURL:    signInURL,
		Error:        errMsg,
	}
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
	}
}

func tenantRegisterPage(w http.ResponseWriter, r *http.Request) {
	if originFromContext(r).Mode == OriginModeSystem {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	tenant, formAction, signInURL, ok := resolveRegisterTenant(r)
	if !ok {
		ServeError(w, r, http.StatusNotFound, "")
		return
	}
	renderTenantRegisterPage(w, r, tenant, formAction, signInURL, "")
}

func tenantRegisterSubmit(w http.ResponseWriter, r *http.Request) {
	if originFromContext(r).Mode == OriginModeSystem {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	tenant, formAction, signInURL, ok := resolveRegisterTenant(r)
	if !ok {
		ServeError(w, r, http.StatusNotFound, "")
		return
	}

	renderErr := func(msg string) {
		renderTenantRegisterPage(w, r, tenant, formAction, signInURL, msg)
	}

	name := strings.TrimSpace(r.FormValue("name"))
	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")
	confirm := r.FormValue("confirm")

	if email == "" || password == "" {
		renderErr("Email and password are required.")
		return
	}
	if password != confirm {
		renderErr("Passwords do not match.")
		return
	}
	if len(password) < 8 {
		renderErr("Password must be at least 8 characters.")
		return
	}
	if !hasUppercase(password) || !hasDigit(password) {
		renderErr("Password must contain at least one uppercase letter and one number.")
		return
	}

	if Pools == nil {
		RegisterTestUser(email, name, password, RoleUser, tenant.Slug)
		newUser := User{
			Email:      email,
			Name:       name,
			Role:       RoleUser,
			TenantSlug: tenant.Slug,
		}
		createSession(w, r, newUser, false)
		http.Redirect(w, r, "/t/"+tenant.Slug+"/", http.StatusSeeOther)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()

	tenantID := tenant.DBID
	if tenantID == 0 {
		id, err := db.LookupTenantIDBySlug(ctx, Pools.Admin, tenant.Slug)
		if err != nil {
			log.Printf("tenantRegisterSubmit: resolve tenant %s: %v", tenant.Slug, err)
			renderErr("Registration unavailable — tenant not found.")
			return
		}
		tenantID = id
	}

	exists, err := db.IsEmailRegistered(ctx, Pools.Admin, email, tenantID)
	if err != nil {
		log.Printf("tenantRegisterSubmit: email check: %v", err)
	}
	if exists {
		renderErr("An account with that email already exists.")
		return
	}

	// Hash the password with the same PBKDF2 scheme used everywhere else so
	// sign-in and password changes stay consistent for self-registered users.
	hash, err := db.HashPassword(password)
	if err != nil {
		log.Printf("tenantRegisterSubmit: hash password: %v", err)
		renderErr("Registration failed. Please try again.")
		return
	}

	// Create user + tenant membership in DB
	userID, err := db.RegisterUser(ctx, Pools.Admin, email, name, hash, tenantID)
	if err != nil {
		log.Printf("tenantRegisterSubmit: register user: %v", err)
		renderErr("Registration failed. Please try again.")
		return
	}

	newUser := User{
		ID:         userID,
		Email:      email,
		Name:       name,
		Role:       RoleUser,
		TenantSlug: tenant.Slug,
	}
	createSession(w, r, newUser, false)

	http.Redirect(w, r, "/t/"+tenant.Slug+"/", http.StatusSeeOther)
}

func hasUppercase(s string) bool {
	for _, r := range s {
		if unicode.IsUpper(r) {
			return true
		}
	}
	return false
}

func hasDigit(s string) bool {
	for _, r := range s {
		if unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

func RegisterTenantRegisterPageRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /register", tenantRegisterPage)
	mux.HandleFunc("POST /register", tenantRegisterSubmit)
	mux.HandleFunc("GET /t/{slug}/register", tenantRegisterPage)
	mux.HandleFunc("POST /t/{slug}/register", tenantRegisterSubmit)
}
