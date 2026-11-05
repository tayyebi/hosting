package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
)

// ── Register form API types ───────────────────────────────────

type apiRegisterFormField struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Placeholder string `json:"placeholder"`
	Required    bool   `json:"required"`
}

type apiRegisterFormResponse struct {
	TenantName string                 `json:"tenant_name"`
	TenantSlug string                 `json:"tenant_slug"`
	Enabled    bool                   `json:"enabled"`
	Fields     []apiRegisterFormField `json:"fields"`
	Captcha    string                 `json:"captcha_provider"`
	TOSUrl     string                 `json:"tos_url"`
	PrivacyUrl string                 `json:"privacy_url"`
}

// ── GET /api/tenant/{slug}/register ───────────────────────────

func apiTenantRegisterGetHandler(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	tenant, found := store.getBySlug(slug)
	if !found {
		apiError(w, http.StatusNotFound, "Tenant not found.")
		return
	}

	domain := tenant.Domain
	if domain == "" {
		domain = "example.com"
	}

	resp := apiRegisterFormResponse{
		TenantName: tenant.Name,
		TenantSlug: tenant.Slug,
		Enabled:    true,
		Fields: []apiRegisterFormField{
			{Name: "email", Type: "email", Placeholder: "you@example.com", Required: true},
			{Name: "password", Type: "password", Placeholder: "Create a password", Required: true},
			{Name: "full_name", Type: "text", Placeholder: "Full name", Required: false},
		},
		Captcha:    "hCaptcha",
		TOSUrl:     "https://" + domain + "/terms",
		PrivacyUrl: "https://" + domain + "/privacy",
	}

	apiOK(w, resp)
}

// ── POST /api/tenant/{slug}/register ──────────────────────────

type apiRegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
}

type apiRegisterResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
	UserID  string `json:"user_id"`
}

func apiTenantRegisterPostHandler(w http.ResponseWriter, r *http.Request) {
	ip := clientIP(r)
	if !registerRateLimiter.allow(ip) {
		w.Header().Set("Retry-After", "3600")
		apiError(w, http.StatusTooManyRequests, "Too many registration attempts. Try again in 1 hour.")
		return
	}

	slug := r.PathValue("slug")
	tenant, found := store.getBySlug(slug)
	if !found {
		apiError(w, http.StatusNotFound, "Tenant not found.")
		return
	}
	_ = tenant

	var req apiRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}

	email := strings.TrimSpace(req.Email)
	password := strings.TrimSpace(req.Password)

	if email == "" {
		apiError(w, http.StatusBadRequest, "Email is required.")
		return
	}
	if !isValidEmail(email) {
		apiError(w, http.StatusBadRequest, "Invalid email address.")
		return
	}
	if password == "" {
		apiError(w, http.StatusBadRequest, "Password is required.")
		return
	}
	if len(password) < 8 {
		apiError(w, http.StatusBadRequest, "Password must be at least 8 characters.")
		return
	}
	if !hasUppercase(password) || !hasDigit(password) {
		apiError(w, http.StatusBadRequest, "Password must contain at least one uppercase letter and one number.")
		return
	}

	apiCreated(w, apiRegisterResponse{
		OK:      true,
		Message: "Registration successful. Please check your email to verify your account.",
		UserID:  newID(),
	})
}
