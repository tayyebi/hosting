package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"platform/db"
)

// ── Registration settings response types ──────────────────────

type apiRegGeneral struct {
	Enabled       bool   `json:"enabled"`
	PageURL       string `json:"page_url"`
	RedirectAfter string `json:"redirect_after"`
	DefaultRole   string `json:"default_role"`
	AccessMode    string `json:"access_mode"`
}

type apiRegRules struct {
	DomainRestrict  bool     `json:"domain_restrict"`
	AllowedDomains  []string `json:"allowed_domains"`
	BlockDisposable bool     `json:"block_disposable"`
	MaxPerHourPerIP int      `json:"max_per_hour_per_ip"`
	MaxPerDay       int      `json:"max_per_day"`
	GeoRestrict     bool     `json:"geo_restrict"`
}

type apiRegField struct {
	Name     string `json:"name"`
	Required bool   `json:"required"`
	System   bool   `json:"system"`
	Enabled  bool   `json:"enabled"`
}

type apiRegVerification struct {
	EmailRequired    bool   `json:"email_required"`
	LinkExpiry       string `json:"link_expiry"`
	CaptchaEnabled   bool   `json:"captcha_enabled"`
	CaptchaProvider  string `json:"captcha_provider"`
	MinPasswordLen   int    `json:"min_password_length"`
	RequireUppercase bool   `json:"require_uppercase"`
	RequireNumber    bool   `json:"require_number"`
	RequireSpecial   bool   `json:"require_special"`
}

type apiRegNotifications struct {
	NotifyAdmins bool     `json:"notify_admins"`
	AdminEmails  []string `json:"admin_emails"`
	WelcomeEmail bool     `json:"welcome_email"`
	FromName     string   `json:"from_name"`
	ReplyTo      string   `json:"reply_to"`
}

type apiRegFunnelStep struct {
	Label      string  `json:"label"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

type apiRegKPIs struct {
	SignUps7d       int     `json:"signups_7d"`
	VerifyRate      float64 `json:"verify_activate_rate"`
	MedianKYCMin    int     `json:"median_kyc_minutes"`
	HighRiskPending int     `json:"high_risk_pending"`
}

type apiRegistrationResponse struct {
	General       apiRegGeneral       `json:"general"`
	Rules         apiRegRules         `json:"rules"`
	Fields        []apiRegField       `json:"fields"`
	Verification  apiRegVerification  `json:"verification"`
	Notifications apiRegNotifications `json:"notifications"`
	KPIs          apiRegKPIs          `json:"kpis"`
	Funnel        []apiRegFunnelStep  `json:"funnel"`
}

func defaultRegistrationResponse(slug string) apiRegistrationResponse {
	return apiRegistrationResponse{
		General: apiRegGeneral{
			Enabled:       true,
			PageURL:       "https://" + slug + ".platform.io/register",
			RedirectAfter: "/dashboard",
			DefaultRole:   "Client",
			AccessMode:    "open",
		},
		Rules: apiRegRules{
			DomainRestrict:  false,
			AllowedDomains:  []string{"acme.com", "acme.io"},
			BlockDisposable: true,
			MaxPerHourPerIP: 5,
			MaxPerDay:       200,
			GeoRestrict:     false,
		},
		Fields: []apiRegField{
			{Name: "Email address", Required: true, System: true, Enabled: true},
			{Name: "Password", Required: true, System: true, Enabled: true},
			{Name: "Full name", Required: false, System: false, Enabled: true},
			{Name: "Company name", Required: false, System: false, Enabled: false},
			{Name: "Phone number", Required: false, System: false, Enabled: false},
			{Name: "Country", Required: false, System: false, Enabled: false},
		},
		Verification: apiRegVerification{
			EmailRequired:    true,
			LinkExpiry:       "24 hours",
			CaptchaEnabled:   true,
			CaptchaProvider:  "hCaptcha",
			MinPasswordLen:   8,
			RequireUppercase: true,
			RequireNumber:    true,
			RequireSpecial:   false,
		},
		Notifications: apiRegNotifications{
			NotifyAdmins: true,
			AdminEmails:  []string{"admin@acme.com", "ops@acme.com"},
			WelcomeEmail: true,
			FromName:     "Acme Platform",
			ReplyTo:      "no-reply@acme.com",
		},
		KPIs: apiRegKPIs{
			SignUps7d:       218,
			VerifyRate:      63.8,
			MedianKYCMin:    9,
			HighRiskPending: 4,
		},
		Funnel: []apiRegFunnelStep{
			{Label: "Form submitted", Count: 218, Percentage: 100.0},
			{Label: "Email verified", Count: 201, Percentage: 92.2},
			{Label: "Phone OTP passed", Count: 178, Percentage: 88.6},
			{Label: "KYC approved", Count: 158, Percentage: 88.8},
			{Label: "Billing & TOS accepted", Count: 139, Percentage: 87.9},
		},
	}
}

// ── GET /api/tenant/{slug}/registration ───────────────────────

func apiTenantRegistrationGetHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok {
		apiError(w, http.StatusUnauthorized, "Authentication required.")
		return
	}

	slug := r.PathValue("slug")
	tenant, found := store.getBySlug(slug)
	if !found {
		apiError(w, http.StatusNotFound, "Tenant not found.")
		return
	}
	if user.Role != RoleSysAdmin && (user.Role != RoleTenant || user.TenantSlug != slug) {
		apiError(w, http.StatusForbidden, "Access denied.")
		return
	}

	resp := defaultRegistrationResponse(slug)

	if Pools != nil && tenant.DBID > 0 {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		rs, exists, err := db.GetRegistrationSettings(ctx, Pools.Admin, tenant.DBID)
		if err != nil {
			log.Printf("apiTenantRegistrationGetHandler: %v", err)
		} else if exists {
			resp.General.Enabled = rs.Enabled
			resp.General.RedirectAfter = rs.RedirectAfter
			resp.General.DefaultRole = rs.DefaultRole
			resp.General.AccessMode = rs.AccessMode
			resp.Rules.DomainRestrict = rs.DomainRestrict
			if len(rs.AllowedDomains) > 0 {
				resp.Rules.AllowedDomains = rs.AllowedDomains
			}
			resp.Rules.BlockDisposable = rs.BlockDisposable
			resp.Rules.MaxPerHourPerIP = rs.MaxPerHourPerIP
			resp.Rules.MaxPerDay = rs.MaxPerDay
			resp.Rules.GeoRestrict = rs.GeoRestrict
			resp.Verification.EmailRequired = rs.EmailRequired
			if rs.LinkExpiry != "" {
				resp.Verification.LinkExpiry = rs.LinkExpiry
			}
			resp.Verification.CaptchaEnabled = rs.CaptchaEnabled
			if rs.CaptchaProvider != "" {
				resp.Verification.CaptchaProvider = rs.CaptchaProvider
			}
			if rs.MinPasswordLen > 0 {
				resp.Verification.MinPasswordLen = rs.MinPasswordLen
			}
			resp.Verification.RequireUppercase = rs.RequireUppercase
			resp.Verification.RequireNumber = rs.RequireNumber
			resp.Verification.RequireSpecial = rs.RequireSpecial
			resp.Notifications.NotifyAdmins = rs.NotifyAdmins
			if len(rs.AdminEmails) > 0 {
				resp.Notifications.AdminEmails = rs.AdminEmails
			}
			resp.Notifications.WelcomeEmail = rs.WelcomeEmail
			if rs.WelcomeFromName != "" {
				resp.Notifications.FromName = rs.WelcomeFromName
			}
			if rs.WelcomeReplyTo != "" {
				resp.Notifications.ReplyTo = rs.WelcomeReplyTo
			}
		}
	}

	apiOK(w, resp)
}

// ── PUT /api/tenant/{slug}/registration ───────────────────────

type apiUpdateRegistrationRequest struct {
	Enabled          *bool    `json:"enabled"`
	AccessMode       *string  `json:"access_mode"`
	RedirectAfter    *string  `json:"redirect_after"`
	DefaultRole      *string  `json:"default_role"`
	DomainRestrict   *bool    `json:"domain_restrict"`
	AllowedDomains   []string `json:"allowed_domains"`
	BlockDisposable  *bool    `json:"block_disposable"`
	MaxPerHourPerIP  *int     `json:"max_per_hour_per_ip"`
	MaxPerDay        *int     `json:"max_per_day"`
	GeoRestrict      *bool    `json:"geo_restrict"`
	EmailRequired    *bool    `json:"email_required"`
	LinkExpiry       *string  `json:"link_expiry"`
	CaptchaEnabled   *bool    `json:"captcha_enabled"`
	CaptchaProvider  *string  `json:"captcha_provider"`
	MinPasswordLen   *int     `json:"min_password_length"`
	RequireUppercase *bool    `json:"require_uppercase"`
	RequireNumber    *bool    `json:"require_number"`
	RequireSpecial   *bool    `json:"require_special"`
	NotifyAdmins     *bool    `json:"notify_admins"`
	AdminEmails      []string `json:"admin_emails"`
	WelcomeEmail     *bool    `json:"welcome_email"`
	WelcomeFromName  *string  `json:"from_name"`
	WelcomeReplyTo   *string  `json:"reply_to"`
}

func apiTenantRegistrationUpdateHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok {
		apiError(w, http.StatusUnauthorized, "Authentication required.")
		return
	}

	slug := r.PathValue("slug")
	tenant, found := store.getBySlug(slug)
	if !found {
		apiError(w, http.StatusNotFound, "Tenant not found.")
		return
	}
	if user.Role != RoleSysAdmin && (user.Role != RoleTenant || user.TenantSlug != slug) {
		apiError(w, http.StatusForbidden, "Access denied.")
		return
	}

	var req apiUpdateRegistrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}

	if req.AccessMode != nil {
		mode := strings.TrimSpace(*req.AccessMode)
		if mode != "open" && mode != "invite" && mode != "approval" {
			apiError(w, http.StatusBadRequest, "Access mode must be open, invite, or approval.")
			return
		}
	}

	if req.RedirectAfter != nil {
		ra := strings.TrimSpace(*req.RedirectAfter)
		if ra != "" && (!strings.HasPrefix(ra, "/") || strings.HasPrefix(ra, "//") || strings.Contains(ra, "://")) {
			apiError(w, http.StatusBadRequest, "redirect_after must be a relative path starting with /.")
			return
		}
	}

	if Pools != nil && tenant.DBID > 0 {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		// Load existing to merge
		existing, _, _ := db.GetRegistrationSettings(ctx, Pools.Admin, tenant.DBID)
		existing.TenantID = tenant.DBID

		if req.Enabled != nil {
			existing.Enabled = *req.Enabled
		}
		if req.AccessMode != nil {
			existing.AccessMode = *req.AccessMode
		}
		if req.RedirectAfter != nil {
			existing.RedirectAfter = *req.RedirectAfter
		}
		if req.DefaultRole != nil {
			existing.DefaultRole = *req.DefaultRole
		}
		if req.DomainRestrict != nil {
			existing.DomainRestrict = *req.DomainRestrict
		}
		if req.AllowedDomains != nil {
			existing.AllowedDomains = req.AllowedDomains
		}
		if req.BlockDisposable != nil {
			existing.BlockDisposable = *req.BlockDisposable
		}
		if req.MaxPerHourPerIP != nil {
			existing.MaxPerHourPerIP = *req.MaxPerHourPerIP
		}
		if req.MaxPerDay != nil {
			existing.MaxPerDay = *req.MaxPerDay
		}
		if req.GeoRestrict != nil {
			existing.GeoRestrict = *req.GeoRestrict
		}
		if req.EmailRequired != nil {
			existing.EmailRequired = *req.EmailRequired
		}
		if req.LinkExpiry != nil {
			existing.LinkExpiry = *req.LinkExpiry
		}
		if req.CaptchaEnabled != nil {
			existing.CaptchaEnabled = *req.CaptchaEnabled
		}
		if req.CaptchaProvider != nil {
			existing.CaptchaProvider = *req.CaptchaProvider
		}
		if req.MinPasswordLen != nil {
			existing.MinPasswordLen = *req.MinPasswordLen
		}
		if req.RequireUppercase != nil {
			existing.RequireUppercase = *req.RequireUppercase
		}
		if req.RequireNumber != nil {
			existing.RequireNumber = *req.RequireNumber
		}
		if req.RequireSpecial != nil {
			existing.RequireSpecial = *req.RequireSpecial
		}
		if req.NotifyAdmins != nil {
			existing.NotifyAdmins = *req.NotifyAdmins
		}
		if req.AdminEmails != nil {
			existing.AdminEmails = req.AdminEmails
		}
		if req.WelcomeEmail != nil {
			existing.WelcomeEmail = *req.WelcomeEmail
		}
		if req.WelcomeFromName != nil {
			existing.WelcomeFromName = *req.WelcomeFromName
		}
		if req.WelcomeReplyTo != nil {
			existing.WelcomeReplyTo = *req.WelcomeReplyTo
		}

		if err := db.UpsertRegistrationSettings(ctx, Pools.Admin, existing); err != nil {
			log.Printf("apiTenantRegistrationUpdateHandler: %v", err)
			apiError(w, http.StatusInternalServerError, "Failed to save settings.")
			return
		}
	}

	apiOK(w, map[string]any{
		"ok":      true,
		"message": "Registration settings updated.",
	})
}
