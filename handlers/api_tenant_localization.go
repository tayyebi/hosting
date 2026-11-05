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

// ── Localization settings response types ──────────────────────

type apiLocRegional struct {
	Language   string `json:"language"`
	Timezone   string `json:"timezone"`
	RTLEnabled bool   `json:"rtl_enabled"`
}

type apiLocFormats struct {
	DateFormat     string `json:"date_format"`
	TimeFormat     string `json:"time_format"`
	NumberFormat   string `json:"number_format"`
	FirstDayOfWeek string `json:"first_day_of_week"`
}

type apiLocCurrency struct {
	Currency       string `json:"currency"`
	SymbolPosition string `json:"symbol_position"`
	TaxInclusive   bool   `json:"tax_inclusive"`
	TaxBreakdown   bool   `json:"tax_breakdown"`
}

type apiLocAPILocale struct {
	AcceptLanguageFallback  string `json:"accept_language_fallback"`
	ContentLanguageHeader   string `json:"content_language_header"`
	IncludeXLocale          bool   `json:"include_x_locale"`
	HonourPerRequest        bool   `json:"honour_per_request"`
	FormatDatesPerLocale    bool   `json:"format_dates_per_locale"`
	FormatCurrencyPerLocale bool   `json:"format_currency_per_locale"`
}

type apiLocRoutingRule struct {
	GeoRegion     string `json:"geo_region"`
	LanguageMatch string `json:"language_match"`
	LocaleServed  string `json:"locale_served"`
	OriginPool    string `json:"origin_pool"`
	Currency      string `json:"currency"`
	LegalEntity   string `json:"legal_entity"`
}

type apiLocCDNStats struct {
	LocaleRoutesActive int    `json:"locale_routes_active"`
	GeoHitRatio        string `json:"geo_hit_ratio"`
	FallbackLocale     string `json:"fallback_locale"`
	ContentVariants    int    `json:"content_variants"`
}

type apiLocalizationResponse struct {
	Regional     apiLocRegional      `json:"regional"`
	Formats      apiLocFormats       `json:"formats"`
	Currency     apiLocCurrency      `json:"currency"`
	APILocale    apiLocAPILocale     `json:"api_locale"`
	CDNStats     apiLocCDNStats      `json:"cdn_stats"`
	RoutingRules []apiLocRoutingRule `json:"routing_rules"`
}

func defaultLocalizationResponse() apiLocalizationResponse {
	return apiLocalizationResponse{
		Regional: apiLocRegional{
			Language:   "English (United Kingdom)",
			Timezone:   "UTC+3:30 — Tehran",
			RTLEnabled: false,
		},
		Formats: apiLocFormats{
			DateFormat:     "DD/MM/YYYY",
			TimeFormat:     "24-hour",
			NumberFormat:   "1,234,567.89",
			FirstDayOfWeek: "Monday",
		},
		Currency: apiLocCurrency{
			Currency:       "USD",
			SymbolPosition: "before",
			TaxInclusive:   true,
			TaxBreakdown:   true,
		},
		APILocale: apiLocAPILocale{
			AcceptLanguageFallback:  "en-GB",
			ContentLanguageHeader:   "en",
			IncludeXLocale:          true,
			HonourPerRequest:        true,
			FormatDatesPerLocale:    false,
			FormatCurrencyPerLocale: false,
		},
		CDNStats: apiLocCDNStats{
			LocaleRoutesActive: 8,
			GeoHitRatio:        "96.1%",
			FallbackLocale:     "en-US",
			ContentVariants:    42,
		},
		RoutingRules: []apiLocRoutingRule{
			{GeoRegion: "EU-West (DE, AT, CH)", LanguageMatch: "de*", LocaleServed: "de-DE", OriginPool: "fra-1, ams-1", Currency: "EUR", LegalEntity: "GmbH"},
			{GeoRegion: "EU-South (ES, MX, AR)", LanguageMatch: "es*", LocaleServed: "es-ES", OriginPool: "mad-1, mia-1", Currency: "EUR / USD", LegalEntity: "S.L."},
			{GeoRegion: "EU-West (FR, BE, CA-QC)", LanguageMatch: "fr*", LocaleServed: "fr-FR", OriginPool: "par-1, mtl-1", Currency: "EUR / CAD", LegalEntity: "SAS"},
			{GeoRegion: "APAC (JP)", LanguageMatch: "ja*", LocaleServed: "ja-JP", OriginPool: "tyo-1, osa-1", Currency: "JPY", LegalEntity: "K.K."},
			{GeoRegion: "LATAM (BR)", LanguageMatch: "pt-BR*", LocaleServed: "pt-BR", OriginPool: "gru-1", Currency: "BRL", LegalEntity: "Ltda."},
			{GeoRegion: "MENA (AE, SA, EG)", LanguageMatch: "ar*", LocaleServed: "ar-AE", OriginPool: "auh-1, ruh-1", Currency: "AED", LegalEntity: "LLC (FZ)"},
			{GeoRegion: "APAC (KR)", LanguageMatch: "ko*", LocaleServed: "ko-KR", OriginPool: "icn-1", Currency: "KRW", LegalEntity: "유한회사"},
			{GeoRegion: "Default (all others)", LanguageMatch: "*", LocaleServed: "en-US", OriginPool: "iad-1, sjc-1", Currency: "USD", LegalEntity: "Inc."},
		},
	}
}

// ── GET /api/tenant/{slug}/localization ───────────────────────

func apiTenantLocalizationGetHandler(w http.ResponseWriter, r *http.Request) {
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

	resp := defaultLocalizationResponse()

	if Pools != nil && tenant.DBID > 0 {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		ls, exists, err := db.GetLocalizationSettings(ctx, Pools.Admin, tenant.DBID)
		if err != nil {
			log.Printf("apiTenantLocalizationGetHandler: %v", err)
		} else if exists {
			if ls.Language != "" {
				resp.Regional.Language = ls.Language
			}
			if ls.Timezone != "" {
				resp.Regional.Timezone = ls.Timezone
			}
			resp.Regional.RTLEnabled = ls.RTLEnabled
			if ls.DateFormat != "" {
				resp.Formats.DateFormat = ls.DateFormat
			}
			if ls.TimeFormat != "" {
				resp.Formats.TimeFormat = ls.TimeFormat
			}
			if ls.NumberFormat != "" {
				resp.Formats.NumberFormat = ls.NumberFormat
			}
			if ls.FirstDayOfWeek != "" {
				resp.Formats.FirstDayOfWeek = ls.FirstDayOfWeek
			}
			if ls.Currency != "" {
				resp.Currency.Currency = ls.Currency
			}
			if ls.SymbolPosition != "" {
				resp.Currency.SymbolPosition = ls.SymbolPosition
			}
			resp.Currency.TaxInclusive = ls.TaxInclusive
			resp.Currency.TaxBreakdown = ls.TaxBreakdown
		}
	}

	apiOK(w, resp)
}

// ── PUT /api/tenant/{slug}/localization ───────────────────────

type apiUpdateLocalizationRequest struct {
	Language   *string `json:"language"`
	Timezone   *string `json:"timezone"`
	RTLEnabled *bool   `json:"rtl_enabled"`
	DateFormat *string `json:"date_format"`
	TimeFormat *string `json:"time_format"`
	Currency   *string `json:"currency"`
}

func apiTenantLocalizationUpdateHandler(w http.ResponseWriter, r *http.Request) {
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

	var req apiUpdateLocalizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}

	if req.Language != nil && strings.TrimSpace(*req.Language) == "" {
		apiError(w, http.StatusBadRequest, "Language cannot be empty.")
		return
	}

	if Pools != nil && tenant.DBID > 0 {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		// Read existing to merge
		def := defaultLocalizationResponse()
		language := def.Regional.Language
		timezone := def.Regional.Timezone
		dateFormat := def.Formats.DateFormat
		timeFormat := def.Formats.TimeFormat
		currency := def.Currency.Currency
		rtlEnabled := def.Regional.RTLEnabled

		if ls, exists, err := db.GetLocalizationSettings(ctx, Pools.Admin, tenant.DBID); err == nil && exists {
			if ls.Language != "" {
				language = ls.Language
			}
			if ls.Timezone != "" {
				timezone = ls.Timezone
			}
			if ls.DateFormat != "" {
				dateFormat = ls.DateFormat
			}
			if ls.TimeFormat != "" {
				timeFormat = ls.TimeFormat
			}
			if ls.Currency != "" {
				currency = ls.Currency
			}
			rtlEnabled = ls.RTLEnabled
		}

		if req.Language != nil {
			language = strings.TrimSpace(*req.Language)
		}
		if req.Timezone != nil {
			timezone = *req.Timezone
		}
		if req.DateFormat != nil {
			dateFormat = *req.DateFormat
		}
		if req.TimeFormat != nil {
			timeFormat = *req.TimeFormat
		}
		if req.Currency != nil {
			currency = *req.Currency
		}
		if req.RTLEnabled != nil {
			rtlEnabled = *req.RTLEnabled
		}

		if err := db.UpsertLocalizationSettings(ctx, Pools.Admin, tenant.DBID, language, timezone, dateFormat, timeFormat, currency, rtlEnabled); err != nil {
			log.Printf("apiTenantLocalizationUpdateHandler: %v", err)
			apiError(w, http.StatusInternalServerError, "Failed to update localization settings.")
			return
		}
	}

	apiOK(w, map[string]any{
		"ok":      true,
		"message": "Localization settings updated.",
	})
}
