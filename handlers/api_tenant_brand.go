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

// ── Brand settings response types ─────────────────────────────

type apiBrandIdentity struct {
	CompanyName  string `json:"company_name"`
	Tagline      string `json:"tagline"`
	Website      string `json:"website"`
	SupportEmail string `json:"support_email"`
	LogoURL      string `json:"logo_url"`
}

type apiBrandColor struct {
	Label string `json:"label"`
	Hex   string `json:"hex"`
}

type apiBrandSocial struct {
	Twitter  string `json:"twitter"`
	LinkedIn string `json:"linkedin"`
	GitHub   string `json:"github"`
	Discord  string `json:"discord"`
}

type apiBrandDevPortal struct {
	APIDocsURL    string `json:"api_docs_url"`
	DevHubURL     string `json:"dev_hub_url"`
	SDKNamespace  string `json:"sdk_namespace"`
	ChangelogURL  string `json:"changelog_url"`
	APIBaseURL    string `json:"api_base_url"`
	WebhookPrefix string `json:"webhook_prefix"`
	ServerHeader  string `json:"server_header"`
}

type apiBrandDNSEntry struct {
	Domain string `json:"domain"`
	Type   string `json:"type"`
	Target string `json:"target"`
	Status string `json:"status"`
}

type apiBrandEmailTemplate struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type apiBrandResponse struct {
	Identity       apiBrandIdentity        `json:"identity"`
	Colors         []apiBrandColor         `json:"colors"`
	Social         apiBrandSocial          `json:"social"`
	DevPortal      apiBrandDevPortal       `json:"dev_portal"`
	DNS            []apiBrandDNSEntry      `json:"dns"`
	EmailTemplates []apiBrandEmailTemplate `json:"email_templates"`
}

func defaultBrandResponse(tenant Tenant) apiBrandResponse {
	domain := tenant.Domain
	if domain == "" {
		domain = "example.com"
	}
	return apiBrandResponse{
		Identity: apiBrandIdentity{
			CompanyName:  tenant.Name,
			Tagline:      "Reliable cloud infrastructure for everyone.",
			Website:      "https://" + domain,
			SupportEmail: "support@" + domain,
			LogoURL:      "",
		},
		Colors: []apiBrandColor{
			{Label: "Primary", Hex: "#6366F1"},
			{Label: "Background", Hex: "#0F1117"},
			{Label: "Surface", Hex: "#FFFFFF"},
			{Label: "Accent", Hex: "#22C55E"},
		},
		Social: apiBrandSocial{
			Twitter:  "https://twitter.com/" + tenant.Slug,
			LinkedIn: "https://linkedin.com/company/" + tenant.Slug,
			GitHub:   "",
			Discord:  "",
		},
		DevPortal: apiBrandDevPortal{
			APIDocsURL:    "https://docs." + domain,
			DevHubURL:     "https://developers." + domain,
			SDKNamespace:  "@" + tenant.Slug + "/sdk",
			ChangelogURL:  "",
			APIBaseURL:    "https://api." + domain,
			WebhookPrefix: "whsec_",
			ServerHeader:  tenant.Name + "-API",
		},
		DNS: []apiBrandDNSEntry{
			{Domain: "dashboard." + domain, Type: "CNAME", Target: "dash.platform.io", Status: "verified"},
			{Domain: "api." + domain, Type: "CNAME", Target: "api.platform.io", Status: "verified"},
			{Domain: "status." + domain, Type: "CNAME", Target: "status.platform.io", Status: "pending"},
			{Domain: "mail." + domain, Type: "SPF+DKIM+DMARC", Target: "sending domain", Status: "verified"},
		},
		EmailTemplates: []apiBrandEmailTemplate{
			{Name: "Welcome & onboarding", Description: "Sent on registration — includes API key and workspace URL"},
			{Name: "Invoice & payment receipt", Description: "Monthly charge confirmation — PDF attachment"},
			{Name: "Maintenance notification", Description: "Advance notice — start — completion — post-mortem"},
			{Name: "Payment failure & dunning", Description: "Card declined — retry notice — suspension warning"},
			{Name: "Security alerts", Description: "Password reset — new device login — API key rotation"},
		},
	}
}

// ── GET /api/tenant/{slug}/brand ──────────────────────────────

func apiTenantBrandGetHandler(w http.ResponseWriter, r *http.Request) {
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

	resp := defaultBrandResponse(tenant)

	if Pools != nil && tenant.DBID > 0 {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		bs, exists, err := db.GetBrandSettings(ctx, Pools.Admin, tenant.DBID)
		if err != nil {
			log.Printf("apiTenantBrandGetHandler: %v", err)
		} else if exists {
			if bs.CompanyName != "" {
				resp.Identity.CompanyName = bs.CompanyName
			}
			if bs.Tagline != "" {
				resp.Identity.Tagline = bs.Tagline
			}
			if bs.Website != "" {
				resp.Identity.Website = bs.Website
			}
			if bs.SupportEmail != "" {
				resp.Identity.SupportEmail = bs.SupportEmail
			}
			resp.Identity.LogoURL = bs.LogoURL
		}
	}

	apiOK(w, resp)
}

// ── PUT /api/tenant/{slug}/brand ──────────────────────────────

type apiUpdateBrandRequest struct {
	CompanyName  *string `json:"company_name"`
	Tagline      *string `json:"tagline"`
	Website      *string `json:"website"`
	SupportEmail *string `json:"support_email"`
	LogoURL      *string `json:"logo_url"`
}

func apiTenantBrandUpdateHandler(w http.ResponseWriter, r *http.Request) {
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

	var req apiUpdateBrandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}

	if req.CompanyName != nil && strings.TrimSpace(*req.CompanyName) == "" {
		apiError(w, http.StatusBadRequest, "Company name cannot be empty.")
		return
	}

	if Pools != nil && tenant.DBID > 0 {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		// Read existing to merge
		existing := defaultBrandResponse(tenant)
		if bs, exists, err := db.GetBrandSettings(ctx, Pools.Admin, tenant.DBID); err == nil && exists {
			if bs.CompanyName != "" {
				existing.Identity.CompanyName = bs.CompanyName
			}
			if bs.Tagline != "" {
				existing.Identity.Tagline = bs.Tagline
			}
			if bs.Website != "" {
				existing.Identity.Website = bs.Website
			}
			if bs.SupportEmail != "" {
				existing.Identity.SupportEmail = bs.SupportEmail
			}
			existing.Identity.LogoURL = bs.LogoURL
		}

		companyName := existing.Identity.CompanyName
		tagline := existing.Identity.Tagline
		website := existing.Identity.Website
		supportEmail := existing.Identity.SupportEmail
		logoURL := existing.Identity.LogoURL

		if req.CompanyName != nil {
			companyName = strings.TrimSpace(*req.CompanyName)
		}
		if req.Tagline != nil {
			tagline = *req.Tagline
		}
		if req.Website != nil {
			website = *req.Website
		}
		if req.SupportEmail != nil {
			supportEmail = *req.SupportEmail
		}
		if req.LogoURL != nil {
			logoURL = *req.LogoURL
		}

		if err := db.UpsertBrandSettings(ctx, Pools.Admin, tenant.DBID, companyName, tagline, website, supportEmail, logoURL); err != nil {
			log.Printf("apiTenantBrandUpdateHandler: %v", err)
			apiError(w, http.StatusInternalServerError, "Failed to update brand settings.")
			return
		}
	}

	apiOK(w, map[string]any{
		"ok":      true,
		"message": "Brand settings updated.",
	})
}
