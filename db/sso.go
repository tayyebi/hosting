package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SSOConfig mirrors the sso_config table (single-row).
type SSOConfig struct {
	ID                int64
	Enabled           bool
	ProviderName      string
	IssuerURL         string
	SessionLifetimeH  int
	OAuth2Enabled     bool
	OIDCEnabled       bool
	SAMLEnabled       bool
	GrantAuthCode     bool
	GrantPKCE         bool
	GrantClientCreds  bool
	GrantRefreshToken bool
	GrantDeviceCode   bool
	AccessTokenTTLH   int
	RefreshTokenTTLD  int
	OIDCSigningAlgo   string
	IDTokenTTLH       int
	ClaimEmail        bool
	ClaimRoles        bool
	ClaimTenant       bool
	SAMLSigningAlgo   string
	SAMLSignAssert    bool
	SAMLEncAssert     bool
	BruteForceEnabled bool
	MaxFailedAttempts int
	LockoutDurationM  int
	NotifyOnLockout   bool
	IPRestrict        bool
	AllowedCIDRs      string
	UpdatedAt         time.Time
}

// SSOClient is one row from sso_oauth_clients.
type SSOClient struct {
	ID        int64
	Name      string
	ClientID  string
	Protocol  string
	Status    string
	CreatedAt time.Time
}

// SSOCORSOrigin is one row from sso_cors_origins.
type SSOCORSOrigin struct {
	ID     int64
	Origin string
}

// DefaultSSOConfig returns sensible defaults when no row exists.
func DefaultSSOConfig() SSOConfig {
	return SSOConfig{
		Enabled:           true,
		ProviderName:      "Platform Identity",
		IssuerURL:         "https://platform.internal/sso",
		SessionLifetimeH:  8,
		OAuth2Enabled:     true,
		OIDCEnabled:       true,
		SAMLEnabled:       false,
		GrantAuthCode:     true,
		GrantPKCE:         true,
		GrantClientCreds:  true,
		GrantRefreshToken: true,
		GrantDeviceCode:   false,
		AccessTokenTTLH:   1,
		RefreshTokenTTLD:  30,
		OIDCSigningAlgo:   "RS256",
		IDTokenTTLH:       1,
		ClaimEmail:        true,
		ClaimRoles:        true,
		ClaimTenant:       false,
		SAMLSigningAlgo:   "RSA-SHA256",
		SAMLSignAssert:    true,
		SAMLEncAssert:     false,
		BruteForceEnabled: true,
		MaxFailedAttempts: 5,
		LockoutDurationM:  15,
		NotifyOnLockout:   true,
		IPRestrict:        false,
		AllowedCIDRs:      "",
	}
}

// GetSSOConfig returns the single sso_config row, seeding defaults if absent.
func GetSSOConfig(ctx context.Context, pool *pgxpool.Pool) (SSOConfig, error) {
	var c SSOConfig
	err := pool.QueryRow(ctx, `
		SELECT id, enabled, provider_name, issuer_url, session_lifetime_h,
		       oauth2_enabled, oidc_enabled, saml_enabled,
		       grant_auth_code, grant_pkce, grant_client_creds, grant_refresh_token, grant_device_code,
		       access_token_ttl_h, refresh_token_ttl_d,
		       oidc_signing_algo, id_token_ttl_h, claim_email, claim_roles, claim_tenant,
		       saml_signing_algo, saml_sign_assert, saml_enc_assert,
		       brute_force_enabled, max_failed_attempts, lockout_duration_m, notify_on_lockout,
		       ip_restrict, allowed_cidrs, updated_at
		FROM sso_config
		ORDER BY id LIMIT 1`,
	).Scan(
		&c.ID, &c.Enabled, &c.ProviderName, &c.IssuerURL, &c.SessionLifetimeH,
		&c.OAuth2Enabled, &c.OIDCEnabled, &c.SAMLEnabled,
		&c.GrantAuthCode, &c.GrantPKCE, &c.GrantClientCreds, &c.GrantRefreshToken, &c.GrantDeviceCode,
		&c.AccessTokenTTLH, &c.RefreshTokenTTLD,
		&c.OIDCSigningAlgo, &c.IDTokenTTLH, &c.ClaimEmail, &c.ClaimRoles, &c.ClaimTenant,
		&c.SAMLSigningAlgo, &c.SAMLSignAssert, &c.SAMLEncAssert,
		&c.BruteForceEnabled, &c.MaxFailedAttempts, &c.LockoutDurationM, &c.NotifyOnLockout,
		&c.IPRestrict, &c.AllowedCIDRs, &c.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		c = DefaultSSOConfig()
		_, seedErr := pool.Exec(ctx, `
			INSERT INTO sso_config (
				enabled, provider_name, issuer_url, session_lifetime_h,
				oauth2_enabled, oidc_enabled, saml_enabled,
				grant_auth_code, grant_pkce, grant_client_creds, grant_refresh_token, grant_device_code,
				access_token_ttl_h, refresh_token_ttl_d,
				oidc_signing_algo, id_token_ttl_h, claim_email, claim_roles, claim_tenant,
				saml_signing_algo, saml_sign_assert, saml_enc_assert,
				brute_force_enabled, max_failed_attempts, lockout_duration_m, notify_on_lockout,
				ip_restrict, allowed_cidrs
			) VALUES (
				$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28
			)`,
			c.Enabled, c.ProviderName, c.IssuerURL, c.SessionLifetimeH,
			c.OAuth2Enabled, c.OIDCEnabled, c.SAMLEnabled,
			c.GrantAuthCode, c.GrantPKCE, c.GrantClientCreds, c.GrantRefreshToken, c.GrantDeviceCode,
			c.AccessTokenTTLH, c.RefreshTokenTTLD,
			c.OIDCSigningAlgo, c.IDTokenTTLH, c.ClaimEmail, c.ClaimRoles, c.ClaimTenant,
			c.SAMLSigningAlgo, c.SAMLSignAssert, c.SAMLEncAssert,
			c.BruteForceEnabled, c.MaxFailedAttempts, c.LockoutDurationM, c.NotifyOnLockout,
			c.IPRestrict, c.AllowedCIDRs,
		)
		if seedErr != nil {
			return c, fmt.Errorf("db.GetSSOConfig seed: %w", seedErr)
		}
		return c, nil
	}
	if err != nil {
		return c, fmt.Errorf("db.GetSSOConfig: %w", err)
	}
	return c, nil
}

// UpdateSSOConfig persists partial updates to sso_config.
func UpdateSSOConfig(ctx context.Context, pool *pgxpool.Pool, c SSOConfig) error {
	_, err := pool.Exec(ctx, `
		UPDATE sso_config SET
			enabled=$1, provider_name=$2, session_lifetime_h=$3,
			oauth2_enabled=$4, oidc_enabled=$5, saml_enabled=$6,
			grant_auth_code=$7, grant_pkce=$8, grant_client_creds=$9,
			grant_refresh_token=$10, grant_device_code=$11,
			access_token_ttl_h=$12, refresh_token_ttl_d=$13,
			oidc_signing_algo=$14, id_token_ttl_h=$15,
			claim_email=$16, claim_roles=$17, claim_tenant=$18,
			saml_signing_algo=$19, saml_sign_assert=$20, saml_enc_assert=$21,
			brute_force_enabled=$22, max_failed_attempts=$23, lockout_duration_m=$24,
			notify_on_lockout=$25, ip_restrict=$26, allowed_cidrs=$27,
			updated_at=now()
		WHERE id=$28`,
		c.Enabled, c.ProviderName, c.SessionLifetimeH,
		c.OAuth2Enabled, c.OIDCEnabled, c.SAMLEnabled,
		c.GrantAuthCode, c.GrantPKCE, c.GrantClientCreds,
		c.GrantRefreshToken, c.GrantDeviceCode,
		c.AccessTokenTTLH, c.RefreshTokenTTLD,
		c.OIDCSigningAlgo, c.IDTokenTTLH,
		c.ClaimEmail, c.ClaimRoles, c.ClaimTenant,
		c.SAMLSigningAlgo, c.SAMLSignAssert, c.SAMLEncAssert,
		c.BruteForceEnabled, c.MaxFailedAttempts, c.LockoutDurationM,
		c.NotifyOnLockout, c.IPRestrict, c.AllowedCIDRs,
		c.ID,
	)
	return err
}

// ListSSOClients returns all registered OAuth/SAML clients.
func ListSSOClients(ctx context.Context, pool *pgxpool.Pool) ([]SSOClient, error) {
	rows, err := pool.Query(ctx, `
		SELECT id, name, client_id, protocol, status, created_at
		FROM sso_oauth_clients
		ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("db.ListSSOClients: %w", err)
	}
	defer rows.Close()
	var out []SSOClient
	for rows.Next() {
		var c SSOClient
		if err := rows.Scan(&c.ID, &c.Name, &c.ClientID, &c.Protocol, &c.Status, &c.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// CreateSSOClient inserts a new OAuth client and returns the created row.
func CreateSSOClient(ctx context.Context, pool *pgxpool.Pool, name, protocol string) (SSOClient, error) {
	var c SSOClient
	err := pool.QueryRow(ctx, `
		INSERT INTO sso_oauth_clients (name, protocol)
		VALUES ($1, $2)
		RETURNING id, name, client_id, protocol, status, created_at`,
		name, protocol,
	).Scan(&c.ID, &c.Name, &c.ClientID, &c.Protocol, &c.Status, &c.CreatedAt)
	if err != nil {
		return c, fmt.Errorf("db.CreateSSOClient: %w", err)
	}
	return c, nil
}

// DeleteSSOClient removes a client by ID.
func DeleteSSOClient(ctx context.Context, pool *pgxpool.Pool, id int64) error {
	_, err := pool.Exec(ctx, `DELETE FROM sso_oauth_clients WHERE id=$1`, id)
	return err
}

// ListSSOCORSOrigins returns all allowed CORS origins.
func ListSSOCORSOrigins(ctx context.Context, pool *pgxpool.Pool) ([]SSOCORSOrigin, error) {
	rows, err := pool.Query(ctx, `SELECT id, origin FROM sso_cors_origins ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("db.ListSSOCORSOrigins: %w", err)
	}
	defer rows.Close()
	var out []SSOCORSOrigin
	for rows.Next() {
		var o SSOCORSOrigin
		if err := rows.Scan(&o.ID, &o.Origin); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

// AddSSOCORSOrigin inserts an origin.
func AddSSOCORSOrigin(ctx context.Context, pool *pgxpool.Pool, origin string) error {
	_, err := pool.Exec(ctx, `INSERT INTO sso_cors_origins (origin) VALUES ($1) ON CONFLICT (origin) DO NOTHING`, origin)
	return err
}

// DeleteSSOCORSOrigin removes an origin by ID.
func DeleteSSOCORSOrigin(ctx context.Context, pool *pgxpool.Pool, id int64) error {
	_, err := pool.Exec(ctx, `DELETE FROM sso_cors_origins WHERE id=$1`, id)
	return err
}
