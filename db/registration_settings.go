package db

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RegistrationSettings struct {
	TenantID         int64
	Enabled          bool
	RedirectAfter    string
	DefaultRole      string
	AccessMode       string
	DomainRestrict   bool
	AllowedDomains   []string
	BlockDisposable  bool
	MaxPerHourPerIP  int
	MaxPerDay        int
	GeoRestrict      bool
	EmailRequired    bool
	LinkExpiry       string
	CaptchaEnabled   bool
	CaptchaProvider  string
	MinPasswordLen   int
	RequireUppercase bool
	RequireNumber    bool
	RequireSpecial   bool
	NotifyAdmins     bool
	AdminEmails      []string
	WelcomeEmail     bool
	WelcomeFromName  string
	WelcomeReplyTo   string
}

func GetRegistrationSettings(ctx context.Context, pool *pgxpool.Pool, tenantID int64) (RegistrationSettings, bool, error) {
	var s RegistrationSettings
	var allowedDomains, adminEmails json.RawMessage
	err := pool.QueryRow(ctx, `
		SELECT tenant_id, enabled, redirect_after, default_role, access_mode,
		       domain_restrict, allowed_domains, block_disposable, max_per_hour_per_ip, max_per_day, geo_restrict,
		       email_required, link_expiry, captcha_enabled, captcha_provider, min_password_len,
		       require_uppercase, require_number, require_special,
		       notify_admins, admin_emails, welcome_email, welcome_from_name, welcome_reply_to
		FROM tenant_registration_settings WHERE tenant_id = $1`, tenantID,
	).Scan(&s.TenantID, &s.Enabled, &s.RedirectAfter, &s.DefaultRole, &s.AccessMode,
		&s.DomainRestrict, &allowedDomains, &s.BlockDisposable, &s.MaxPerHourPerIP, &s.MaxPerDay, &s.GeoRestrict,
		&s.EmailRequired, &s.LinkExpiry, &s.CaptchaEnabled, &s.CaptchaProvider, &s.MinPasswordLen,
		&s.RequireUppercase, &s.RequireNumber, &s.RequireSpecial,
		&s.NotifyAdmins, &adminEmails, &s.WelcomeEmail, &s.WelcomeFromName, &s.WelcomeReplyTo)
	if err != nil {
		if err == pgx.ErrNoRows {
			return RegistrationSettings{}, false, nil
		}
		return RegistrationSettings{}, false, fmt.Errorf("db.GetRegistrationSettings: %w", err)
	}
	_ = json.Unmarshal(allowedDomains, &s.AllowedDomains)
	_ = json.Unmarshal(adminEmails, &s.AdminEmails)
	return s, true, nil
}

func UpsertRegistrationSettings(ctx context.Context, pool *pgxpool.Pool, s RegistrationSettings) error {
	allowedDomains, _ := json.Marshal(s.AllowedDomains)
	adminEmails, _ := json.Marshal(s.AdminEmails)
	_, err := pool.Exec(ctx, `
		INSERT INTO tenant_registration_settings
		    (tenant_id, enabled, redirect_after, default_role, access_mode, domain_restrict,
		     allowed_domains, block_disposable, max_per_hour_per_ip, max_per_day, geo_restrict,
		     email_required, link_expiry, captcha_enabled, captcha_provider, min_password_len,
		     require_uppercase, require_number, require_special,
		     notify_admins, admin_emails, welcome_email, welcome_from_name, welcome_reply_to)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24)
		ON CONFLICT (tenant_id) DO UPDATE SET
		    enabled = EXCLUDED.enabled, redirect_after = EXCLUDED.redirect_after,
		    default_role = EXCLUDED.default_role, access_mode = EXCLUDED.access_mode,
		    domain_restrict = EXCLUDED.domain_restrict, allowed_domains = EXCLUDED.allowed_domains,
		    block_disposable = EXCLUDED.block_disposable, max_per_hour_per_ip = EXCLUDED.max_per_hour_per_ip,
		    max_per_day = EXCLUDED.max_per_day, geo_restrict = EXCLUDED.geo_restrict,
		    email_required = EXCLUDED.email_required, link_expiry = EXCLUDED.link_expiry,
		    captcha_enabled = EXCLUDED.captcha_enabled, captcha_provider = EXCLUDED.captcha_provider,
		    min_password_len = EXCLUDED.min_password_len, require_uppercase = EXCLUDED.require_uppercase,
		    require_number = EXCLUDED.require_number, require_special = EXCLUDED.require_special,
		    notify_admins = EXCLUDED.notify_admins, admin_emails = EXCLUDED.admin_emails,
		    welcome_email = EXCLUDED.welcome_email, welcome_from_name = EXCLUDED.welcome_from_name,
		    welcome_reply_to = EXCLUDED.welcome_reply_to, updated_at = now()`,
		s.TenantID, s.Enabled, s.RedirectAfter, s.DefaultRole, s.AccessMode, s.DomainRestrict,
		allowedDomains, s.BlockDisposable, s.MaxPerHourPerIP, s.MaxPerDay, s.GeoRestrict,
		s.EmailRequired, s.LinkExpiry, s.CaptchaEnabled, s.CaptchaProvider, s.MinPasswordLen,
		s.RequireUppercase, s.RequireNumber, s.RequireSpecial,
		s.NotifyAdmins, adminEmails, s.WelcomeEmail, s.WelcomeFromName, s.WelcomeReplyTo)
	if err != nil {
		return fmt.Errorf("db.UpsertRegistrationSettings: %w", err)
	}
	return nil
}
