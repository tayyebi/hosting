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

	"golang.org/x/crypto/bcrypt"
)

type RegisterPageData struct {
	Tenant Tenant
	Error  string
}

func tenantRegisterPage(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	tenant, found := store.getBySlug(slug)
	if !found {
		http.NotFound(w, r)
		return
	}

	paths := []string{
		filepath.Join("templates", "base.html"),
		filepath.Join("templates", "tenant", "register.html"),
	}
	tmpl, err := template.ParseFiles(paths...)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	data := RegisterPageData{Tenant: tenant}
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
	}
}

func tenantRegisterSubmit(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	tenant, found := store.getBySlug(slug)
	if !found {
		http.NotFound(w, r)
		return
	}

	renderErr := func(msg string) {
		paths := []string{
			filepath.Join("templates", "base.html"),
			filepath.Join("templates", "tenant", "register.html"),
		}
		tmpl, err := template.ParseFiles(paths...)
		if err != nil {
			http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if err := tmpl.ExecuteTemplate(w, "base", RegisterPageData{Tenant: tenant, Error: msg}); err != nil {
			http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
		}
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
		renderErr("Registration is temporarily unavailable. Please try again later.")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()

	// Resolve tenant DB ID
	tenantID := tenant.DBID
	if tenantID == 0 {
		id, err := db.LookupTenantIDBySlug(ctx, Pools.Admin, slug)
		if err != nil {
			log.Printf("tenantRegisterSubmit: resolve tenant %s: %v", slug, err)
			renderErr("Registration unavailable — tenant not found.")
			return
		}
		tenantID = id
	}

	// Check if already registered for this tenant
	exists, err := db.IsEmailRegistered(ctx, Pools.Admin, email, tenantID)
	if err != nil {
		log.Printf("tenantRegisterSubmit: email check: %v", err)
	}
	if exists {
		renderErr("An account with that email already exists.")
		return
	}

	// Hash password with bcrypt
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("tenantRegisterSubmit: bcrypt: %v", err)
		renderErr("Registration failed. Please try again.")
		return
	}

	// Create user + tenant membership in DB
	userID, err := db.RegisterUser(ctx, Pools.Admin, email, name, string(hash), tenantID)
	if err != nil {
		log.Printf("tenantRegisterSubmit: register user: %v", err)
		renderErr("Registration failed. Please try again.")
		return
	}

	// Issue a session for the new user
	newUser := User{
		ID:         userID,
		Email:      email,
		Name:       name,
		Role:       RoleTenant,
		TenantSlug: slug,
	}
	createSession(w, r, newUser, false)

	http.Redirect(w, r, "/t/"+slug+"/", http.StatusSeeOther)
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
	mux.HandleFunc("GET /t/{slug}/register", tenantRegisterPage)
	mux.HandleFunc("POST /t/{slug}/register", tenantRegisterSubmit)
}
