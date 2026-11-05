package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"platform/db"
)

// ── GET /api/sysadmin/sso ──────────────────────────────────

type apiSSOGeneral struct {
	Enabled         bool   `json:"enabled"`
	ProviderName    string `json:"provider_name"`
	IssuerURL       string `json:"issuer_url"`
	SessionLifetime string `json:"session_lifetime"`
	OAuth2Enabled   bool   `json:"oauth2_enabled"`
	OIDCEnabled     bool   `json:"oidc_enabled"`
	SAMLEnabled     bool   `json:"saml_enabled"`
}

type apiSSOOAuth2 struct {
	AuthEndpoint    string `json:"auth_endpoint"`
	TokenEndpoint   string `json:"token_endpoint"`
	RevokeEndpoint  string `json:"revoke_endpoint"`
	AuthCode        bool   `json:"grant_auth_code"`
	PKCE            bool   `json:"grant_pkce"`
	ClientCreds     bool   `json:"grant_client_credentials"`
	RefreshToken    bool   `json:"grant_refresh_token"`
	DeviceCode      bool   `json:"grant_device_code"`
	AccessTokenTTL  string `json:"access_token_ttl"`
	RefreshTokenTTL string `json:"refresh_token_ttl"`
}

type apiSSOOIDC struct {
	DiscoveryURL string `json:"discovery_url"`
	JWKSURL      string `json:"jwks_url"`
	UserInfoURL  string `json:"userinfo_url"`
	SigningAlgo  string `json:"signing_algorithm"`
	IDTokenTTL   string `json:"id_token_ttl"`
	ClaimEmail   bool   `json:"claim_email"`
	ClaimRoles   bool   `json:"claim_roles"`
	ClaimTenant  bool   `json:"claim_tenant"`
}

type apiSSOSAML struct {
	Enabled     bool   `json:"enabled"`
	EntityID    string `json:"entity_id"`
	SSOURL      string `json:"sso_url"`
	SLOURL      string `json:"slo_url"`
	MetadataURL string `json:"metadata_url"`
	SigningAlgo string `json:"signing_algorithm"`
	SignAssert  bool   `json:"sign_assertions"`
	EncAssert   bool   `json:"encrypt_assertions"`
}

type apiSSOClient struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	ClientID string `json:"client_id"`
	Protocol string `json:"protocol"`
	Status   string `json:"status"`
}

type apiSSOToken struct {
	User    string `json:"user"`
	Client  string `json:"client"`
	Type    string `json:"type"`
	Issued  string `json:"issued"`
	Expires string `json:"expires"`
}

type apiSSOSecurity struct {
	BruteForce    bool     `json:"brute_force_enabled"`
	MaxAttempts   int      `json:"max_failed_attempts"`
	LockoutMin    int      `json:"lockout_duration_min"`
	NotifyLockout bool     `json:"notify_on_lockout"`
	RateToken     string   `json:"rate_token_endpoint"`
	RateAuth      string   `json:"rate_authorize_endpoint"`
	RateUserInfo  string   `json:"rate_userinfo_endpoint"`
	IPRestrict    bool     `json:"ip_restrict"`
	AllowedCIDRs  string   `json:"allowed_cidrs"`
	CORSOrigins   []string `json:"cors_origins"`
}

type apiSSOResponse struct {
	General  apiSSOGeneral  `json:"general"`
	OAuth2   apiSSOOAuth2   `json:"oauth2"`
	OIDC     apiSSOOIDC     `json:"oidc"`
	SAML     apiSSOSAML     `json:"saml"`
	Clients  []apiSSOClient `json:"clients"`
	Tokens   []apiSSOToken  `json:"tokens"`
	Security apiSSOSecurity `json:"security"`
}

func configToSSOResponse(cfg db.SSOConfig, clients []db.SSOClient, cors []db.SSOCORSOrigin) apiSSOResponse {
	sessionLifetime := "8 hours"
	switch cfg.SessionLifetimeH {
	case 1:
		sessionLifetime = "1 hour"
	case 24:
		sessionLifetime = "24 hours"
	case 168:
		sessionLifetime = "7 days"
	}

	accessTokenTTL := "1 hour"
	if cfg.AccessTokenTTLH != 1 {
		accessTokenTTL = strconv.Itoa(cfg.AccessTokenTTLH) + " hours"
	}
	refreshTokenTTL := strconv.Itoa(cfg.RefreshTokenTTLD) + " days"

	idTokenTTL := strconv.Itoa(cfg.IDTokenTTLH) + " hour"
	if cfg.IDTokenTTLH > 1 {
		idTokenTTL = strconv.Itoa(cfg.IDTokenTTLH) + " hours"
	}

	apiClients := make([]apiSSOClient, len(clients))
	for i, c := range clients {
		apiClients[i] = apiSSOClient{
			ID: c.ID, Name: c.Name, ClientID: c.ClientID,
			Protocol: c.Protocol, Status: c.Status,
		}
	}

	corsOrigins := make([]string, len(cors))
	for i, o := range cors {
		corsOrigins[i] = o.Origin
	}

	issuerURL := cfg.IssuerURL
	if issuerURL == "" {
		issuerURL = "https://platform.internal/sso"
	}

	return apiSSOResponse{
		General: apiSSOGeneral{
			Enabled: cfg.Enabled, ProviderName: cfg.ProviderName,
			IssuerURL: issuerURL, SessionLifetime: sessionLifetime,
			OAuth2Enabled: cfg.OAuth2Enabled, OIDCEnabled: cfg.OIDCEnabled, SAMLEnabled: cfg.SAMLEnabled,
		},
		OAuth2: apiSSOOAuth2{
			AuthEndpoint: "/sso/oauth2/authorize", TokenEndpoint: "/sso/oauth2/token",
			RevokeEndpoint: "/sso/oauth2/revoke",
			AuthCode: cfg.GrantAuthCode, PKCE: cfg.GrantPKCE,
			ClientCreds: cfg.GrantClientCreds, RefreshToken: cfg.GrantRefreshToken,
			DeviceCode: cfg.GrantDeviceCode,
			AccessTokenTTL: accessTokenTTL, RefreshTokenTTL: refreshTokenTTL,
		},
		OIDC: apiSSOOIDC{
			DiscoveryURL: "/sso/.well-known/openid-configuration",
			JWKSURL:      "/sso/.well-known/jwks.json",
			UserInfoURL:  "/sso/oidc/userinfo",
			SigningAlgo:  cfg.OIDCSigningAlgo, IDTokenTTL: idTokenTTL,
			ClaimEmail: cfg.ClaimEmail, ClaimRoles: cfg.ClaimRoles, ClaimTenant: cfg.ClaimTenant,
		},
		SAML: apiSSOSAML{
			Enabled: cfg.SAMLEnabled,
			EntityID: "https://platform.internal/sso/saml",
			SSOURL: "/sso/saml/sso", SLOURL: "/sso/saml/slo",
			MetadataURL: "/sso/saml/metadata", SigningAlgo: cfg.SAMLSigningAlgo,
			SignAssert: cfg.SAMLSignAssert, EncAssert: cfg.SAMLEncAssert,
		},
		Clients: apiClients,
		Tokens:  liveTokenSample(),
		Security: apiSSOSecurity{
			BruteForce: cfg.BruteForceEnabled, MaxAttempts: cfg.MaxFailedAttempts,
			LockoutMin: cfg.LockoutDurationM, NotifyLockout: cfg.NotifyOnLockout,
			RateToken: "100 req/min per client", RateAuth: "30 req/min per IP",
			RateUserInfo: "60 req/min per token",
			IPRestrict: cfg.IPRestrict, AllowedCIDRs: cfg.AllowedCIDRs,
			CORSOrigins: corsOrigins,
		},
	}
}

func liveTokenSample() []apiSSOToken {
	return []apiSSOToken{
		{User: "alice@example.com", Client: "Internal Dashboard", Type: "Access", Issued: "2 min ago", Expires: "58 min"},
		{User: "bob@example.com", Client: "CI Pipeline", Type: "Access", Issued: "14 min ago", Expires: "46 min"},
		{User: "alice@example.com", Client: "Internal Dashboard", Type: "Refresh", Issued: "2 min ago", Expires: "29 days"},
	}
}

func apiSysadminSSOHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok || user.Role != RoleSysAdmin {
		apiError(w, http.StatusForbidden, "Sysadmin access required.")
		return
	}

	if Pools == nil {
		apiOK(w, configToSSOResponse(db.DefaultSSOConfig(), nil, nil))
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	cfg, _ := db.GetSSOConfig(ctx, Pools.Admin)
	clients, _ := db.ListSSOClients(ctx, Pools.Admin)
	cors, _ := db.ListSSOCORSOrigins(ctx, Pools.Admin)

	apiOK(w, configToSSOResponse(cfg, clients, cors))
}

// ── PUT /api/sysadmin/sso ──────────────────────────────────

type apiSSOUpdateRequest struct {
	Enabled           *bool   `json:"enabled"`
	ProviderName      *string `json:"provider_name"`
	SessionLifetimeH  *int    `json:"session_lifetime_h"`
	OAuth2Enabled     *bool   `json:"oauth2_enabled"`
	OIDCEnabled       *bool   `json:"oidc_enabled"`
	SAMLEnabled       *bool   `json:"saml_enabled"`
	GrantAuthCode     *bool   `json:"grant_auth_code"`
	GrantPKCE         *bool   `json:"grant_pkce"`
	GrantClientCreds  *bool   `json:"grant_client_creds"`
	GrantRefreshToken *bool   `json:"grant_refresh_token"`
	GrantDeviceCode   *bool   `json:"grant_device_code"`
	AccessTokenTTLH   *int    `json:"access_token_ttl_h"`
	RefreshTokenTTLD  *int    `json:"refresh_token_ttl_d"`
	OIDCSigningAlgo   *string `json:"oidc_signing_algo"`
	IDTokenTTLH       *int    `json:"id_token_ttl_h"`
	ClaimEmail        *bool   `json:"claim_email"`
	ClaimRoles        *bool   `json:"claim_roles"`
	ClaimTenant       *bool   `json:"claim_tenant"`
	SAMLSigningAlgo   *string `json:"saml_signing_algo"`
	SAMLSignAssert    *bool   `json:"saml_sign_assert"`
	SAMLEncAssert     *bool   `json:"saml_enc_assert"`
	BruteForceEnabled *bool   `json:"brute_force_enabled"`
	MaxFailedAttempts *int    `json:"max_failed_attempts"`
	LockoutDurationM  *int    `json:"lockout_duration_m"`
	NotifyOnLockout   *bool   `json:"notify_on_lockout"`
	IPRestrict        *bool   `json:"ip_restrict"`
	AllowedCIDRs      *string `json:"allowed_cidrs"`
}

func apiSysadminSSOUpdateHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok || user.Role != RoleSysAdmin {
		apiError(w, http.StatusForbidden, "Sysadmin access required.")
		return
	}
	if Pools == nil {
		apiError(w, http.StatusServiceUnavailable, "Database not available.")
		return
	}

	var req apiSSOUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	cfg, err := db.GetSSOConfig(ctx, Pools.Admin)
	if err != nil {
		apiError(w, http.StatusInternalServerError, "Failed to load SSO config.")
		return
	}

	applyBool := func(dst *bool, src *bool) { if src != nil { *dst = *src } }
	applyInt  := func(dst *int, src *int)   { if src != nil { *dst = *src } }
	applyStr  := func(dst *string, src *string) { if src != nil && *src != "" { *dst = *src } }

	applyBool(&cfg.Enabled, req.Enabled)
	applyStr(&cfg.ProviderName, req.ProviderName)
	applyInt(&cfg.SessionLifetimeH, req.SessionLifetimeH)
	applyBool(&cfg.OAuth2Enabled, req.OAuth2Enabled)
	applyBool(&cfg.OIDCEnabled, req.OIDCEnabled)
	applyBool(&cfg.SAMLEnabled, req.SAMLEnabled)
	applyBool(&cfg.GrantAuthCode, req.GrantAuthCode)
	applyBool(&cfg.GrantPKCE, req.GrantPKCE)
	applyBool(&cfg.GrantClientCreds, req.GrantClientCreds)
	applyBool(&cfg.GrantRefreshToken, req.GrantRefreshToken)
	applyBool(&cfg.GrantDeviceCode, req.GrantDeviceCode)
	applyInt(&cfg.AccessTokenTTLH, req.AccessTokenTTLH)
	applyInt(&cfg.RefreshTokenTTLD, req.RefreshTokenTTLD)
	applyStr(&cfg.OIDCSigningAlgo, req.OIDCSigningAlgo)
	applyInt(&cfg.IDTokenTTLH, req.IDTokenTTLH)
	applyBool(&cfg.ClaimEmail, req.ClaimEmail)
	applyBool(&cfg.ClaimRoles, req.ClaimRoles)
	applyBool(&cfg.ClaimTenant, req.ClaimTenant)
	applyStr(&cfg.SAMLSigningAlgo, req.SAMLSigningAlgo)
	applyBool(&cfg.SAMLSignAssert, req.SAMLSignAssert)
	applyBool(&cfg.SAMLEncAssert, req.SAMLEncAssert)
	applyBool(&cfg.BruteForceEnabled, req.BruteForceEnabled)
	applyInt(&cfg.MaxFailedAttempts, req.MaxFailedAttempts)
	applyInt(&cfg.LockoutDurationM, req.LockoutDurationM)
	applyBool(&cfg.NotifyOnLockout, req.NotifyOnLockout)
	applyBool(&cfg.IPRestrict, req.IPRestrict)
	if req.AllowedCIDRs != nil {
		cfg.AllowedCIDRs = *req.AllowedCIDRs
	}

	if err := db.UpdateSSOConfig(ctx, Pools.Admin, cfg); err != nil {
		apiError(w, http.StatusInternalServerError, "Failed to save SSO config.")
		return
	}

	clients, _ := db.ListSSOClients(ctx, Pools.Admin)
	cors, _ := db.ListSSOCORSOrigins(ctx, Pools.Admin)
	apiOK(w, configToSSOResponse(cfg, clients, cors))
}

// ── POST /api/sysadmin/sso/clients ────────────────────────

type apiSSOClientCreateRequest struct {
	Name     string `json:"name"`
	Protocol string `json:"protocol"`
}

func apiSysadminSSOCreateClientHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok || user.Role != RoleSysAdmin {
		apiError(w, http.StatusForbidden, "Sysadmin access required.")
		return
	}
	if Pools == nil {
		apiError(w, http.StatusServiceUnavailable, "Database not available.")
		return
	}

	var req apiSSOClientCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		apiError(w, http.StatusBadRequest, "name is required.")
		return
	}
	if req.Protocol == "" {
		req.Protocol = "OAuth 2.0 / OIDC"
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	c, err := db.CreateSSOClient(ctx, Pools.Admin, req.Name, req.Protocol)
	if err != nil {
		apiError(w, http.StatusInternalServerError, "Failed to create client.")
		return
	}
	apiCreated(w, apiSSOClient{ID: c.ID, Name: c.Name, ClientID: c.ClientID, Protocol: c.Protocol, Status: c.Status})
}

// ── DELETE /api/sysadmin/sso/clients/{id} ─────────────────

func apiSysadminSSODeleteClientHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok || user.Role != RoleSysAdmin {
		apiError(w, http.StatusForbidden, "Sysadmin access required.")
		return
	}
	if Pools == nil {
		apiError(w, http.StatusServiceUnavailable, "Database not available.")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		apiError(w, http.StatusBadRequest, "Invalid client ID.")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := db.DeleteSSOClient(ctx, Pools.Admin, id); err != nil {
		apiError(w, http.StatusInternalServerError, "Failed to delete client.")
		return
	}
	apiOK(w, map[string]bool{"deleted": true})
}

// ── POST /api/sysadmin/sso/cors ───────────────────────────

type apiSSOCORSAddRequest struct {
	Origin string `json:"origin"`
}

func apiSysadminSSOAddCORSHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok || user.Role != RoleSysAdmin {
		apiError(w, http.StatusForbidden, "Sysadmin access required.")
		return
	}
	if Pools == nil {
		apiError(w, http.StatusServiceUnavailable, "Database not available.")
		return
	}

	var req apiSSOCORSAddRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Origin == "" {
		apiError(w, http.StatusBadRequest, "origin is required.")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := db.AddSSOCORSOrigin(ctx, Pools.Admin, req.Origin); err != nil {
		apiError(w, http.StatusInternalServerError, "Failed to add origin.")
		return
	}
	apiCreated(w, map[string]string{"origin": req.Origin})
}

// ── DELETE /api/sysadmin/sso/cors/{id} ────────────────────

func apiSysadminSSODeleteCORSHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok || user.Role != RoleSysAdmin {
		apiError(w, http.StatusForbidden, "Sysadmin access required.")
		return
	}
	if Pools == nil {
		apiError(w, http.StatusServiceUnavailable, "Database not available.")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		apiError(w, http.StatusBadRequest, "Invalid CORS ID.")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := db.DeleteSSOCORSOrigin(ctx, Pools.Admin, id); err != nil {
		apiError(w, http.StatusInternalServerError, "Failed to delete origin.")
		return
	}
	apiOK(w, map[string]bool{"deleted": true})
}
