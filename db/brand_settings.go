package db

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BrandSettings struct {
	TenantID       int64
	CompanyName    string
	Tagline        string
	Website        string
	SupportEmail   string
	LogoURL        string
	Colors         json.RawMessage
	Social         json.RawMessage
	DevPortal      json.RawMessage
	DNSEntries     json.RawMessage
	EmailTemplates json.RawMessage
}

func GetBrandSettings(ctx context.Context, pool *pgxpool.Pool, tenantID int64) (BrandSettings, bool, error) {
	var b BrandSettings
	err := pool.QueryRow(ctx, `
		SELECT tenant_id, company_name, tagline, website, support_email, logo_url,
		       colors, social, dev_portal, dns_entries, email_templates
		FROM tenant_brand_settings
		WHERE tenant_id = $1`, tenantID,
	).Scan(&b.TenantID, &b.CompanyName, &b.Tagline, &b.Website, &b.SupportEmail, &b.LogoURL,
		&b.Colors, &b.Social, &b.DevPortal, &b.DNSEntries, &b.EmailTemplates)
	if err != nil {
		if err == pgx.ErrNoRows {
			return BrandSettings{}, false, nil
		}
		return BrandSettings{}, false, fmt.Errorf("db.GetBrandSettings: %w", err)
	}
	return b, true, nil
}

func UpsertBrandSettings(ctx context.Context, pool *pgxpool.Pool, tenantID int64, companyName, tagline, website, supportEmail, logoURL string) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO tenant_brand_settings (tenant_id, company_name, tagline, website, support_email, logo_url)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (tenant_id) DO UPDATE SET
		    company_name  = EXCLUDED.company_name,
		    tagline       = EXCLUDED.tagline,
		    website       = EXCLUDED.website,
		    support_email = EXCLUDED.support_email,
		    logo_url      = EXCLUDED.logo_url,
		    updated_at    = now()`,
		tenantID, companyName, tagline, website, supportEmail, logoURL)
	if err != nil {
		return fmt.Errorf("db.UpsertBrandSettings: %w", err)
	}
	return nil
}

func UpsertBrandColors(ctx context.Context, pool *pgxpool.Pool, tenantID int64, colors json.RawMessage) error {
	if len(colors) == 0 {
		colors = json.RawMessage(`[]`)
	}
	_, err := pool.Exec(ctx, `
		INSERT INTO tenant_brand_settings (tenant_id, colors)
		VALUES ($1, $2::jsonb)
		ON CONFLICT (tenant_id) DO UPDATE SET
		    colors     = EXCLUDED.colors,
		    updated_at = now()`,
		tenantID, string(colors))
	if err != nil {
		return fmt.Errorf("db.UpsertBrandColors: %w", err)
	}
	return nil
}
