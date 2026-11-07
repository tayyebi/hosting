package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TenantSSOProvider is the tenant-level third-party provider setting exposed on
// that tenant's public sign-in and registration pages. It deliberately contains
// no client credentials; the platform's first-party identity provider is
// configured separately through the system-level SSO settings.
type TenantSSOProvider struct {
	Key      string `json:"key"`
	Name     string `json:"name"`
	Protocol string `json:"protocol"`
	Enabled  bool   `json:"enabled"`
}

// SupportedTenantSSOProviders returns the providers that a tenant may enable.
// Returning a new slice keeps callers from mutating the shared catalog.
func SupportedTenantSSOProviders() []TenantSSOProvider {
	return []TenantSSOProvider{
		{Key: "google", Name: "Google", Protocol: "OpenID Connect"},
		{Key: "microsoft", Name: "Microsoft Entra ID", Protocol: "OpenID Connect"},
		{Key: "github", Name: "GitHub", Protocol: "OAuth 2.0"},
		{Key: "okta", Name: "Okta", Protocol: "OpenID Connect"},
		{Key: "custom_oidc", Name: "Custom OIDC", Protocol: "OpenID Connect"},
		{Key: "custom_saml", Name: "Custom SAML 2.0", Protocol: "SAML 2.0"},
	}
}

// ListTenantSSOProviders loads one tenant's provider flags and merges them into
// the supported catalog. Providers with no saved row remain disabled.
func ListTenantSSOProviders(ctx context.Context, pool *pgxpool.Pool, tenantID int64) ([]TenantSSOProvider, error) {
	providers := SupportedTenantSSOProviders()
	if tenantID <= 0 {
		return providers, fmt.Errorf("db.ListTenantSSOProviders: tenant id must be positive")
	}

	rows, err := pool.Query(ctx, `
		SELECT provider_key, enabled
		FROM tenant_sso_providers
		WHERE tenant_id = $1
		ORDER BY provider_key`, tenantID)
	if err != nil {
		return providers, fmt.Errorf("db.ListTenantSSOProviders: %w", err)
	}
	defer rows.Close()

	enabled := make(map[string]bool, len(providers))
	for rows.Next() {
		var key string
		var isEnabled bool
		if err := rows.Scan(&key, &isEnabled); err != nil {
			return providers, fmt.Errorf("db.ListTenantSSOProviders scan: %w", err)
		}
		enabled[key] = isEnabled
	}
	if err := rows.Err(); err != nil {
		return providers, fmt.Errorf("db.ListTenantSSOProviders rows: %w", err)
	}

	for i := range providers {
		providers[i].Enabled = enabled[providers[i].Key]
	}
	return providers, nil
}

// SaveTenantSSOProviders replaces the tenant's provider availability flags.
// Provider names and protocols always come from the server-side allowlist.
func SaveTenantSSOProviders(ctx context.Context, pool *pgxpool.Pool, tenantID int64, settings []TenantSSOProvider) error {
	if tenantID <= 0 {
		return fmt.Errorf("db.SaveTenantSSOProviders: tenant id must be positive")
	}

	catalog := SupportedTenantSSOProviders()
	allowed := make(map[string]bool, len(catalog))
	for _, provider := range catalog {
		allowed[provider.Key] = true
	}
	requested := make(map[string]bool, len(settings))
	for _, setting := range settings {
		if !allowed[setting.Key] {
			return fmt.Errorf("db.SaveTenantSSOProviders: unsupported provider %q", setting.Key)
		}
		if _, exists := requested[setting.Key]; exists {
			return fmt.Errorf("db.SaveTenantSSOProviders: duplicate provider %q", setting.Key)
		}
		requested[setting.Key] = setting.Enabled
	}

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("db.SaveTenantSSOProviders begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	for _, provider := range catalog {
		_, err := tx.Exec(ctx, `
			INSERT INTO tenant_sso_providers (tenant_id, provider_key, enabled)
			VALUES ($1, $2, $3)
			ON CONFLICT (tenant_id, provider_key) DO UPDATE
			SET enabled = EXCLUDED.enabled, updated_at = now()`,
			tenantID, provider.Key, requested[provider.Key])
		if err != nil {
			return fmt.Errorf("db.SaveTenantSSOProviders upsert %q: %w", provider.Key, err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("db.SaveTenantSSOProviders commit: %w", err)
	}
	return nil
}
