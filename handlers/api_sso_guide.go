package handlers

import "net/http"

// ── GET /api/sso-guide ───────────────────────────────────────

type ssoProvider struct {
	Name     string `json:"name"`
	Protocol string `json:"protocol"`
	Status   string `json:"status"`
}

type ssoEndpoints struct {
	Authorization string `json:"authorization"`
	Token         string `json:"token"`
	UserInfo      string `json:"userinfo"`
	JWKS          string `json:"jwks"`
	SAMLMetadata  string `json:"saml_metadata"`
}

type ssoGuideData struct {
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Protocols   []string       `json:"protocols"`
	Endpoints   ssoEndpoints   `json:"endpoints"`
	Providers   []ssoProvider  `json:"providers"`
	Scopes      []string       `json:"scopes"`
	GrantTypes  []string       `json:"grant_types"`
	PKCERequired bool          `json:"pkce_required"`
	TokenFormat  string        `json:"token_format"`
	SigningAlgo  string        `json:"signing_algorithm"`
}

func apiSSOGuideHandler(w http.ResponseWriter, r *http.Request) {
	data := ssoGuideData{
		Title:       "SSO Developer Guide",
		Description: "Integrate your application with Platform's identity provider using industry-standard protocols.",
		Protocols:   []string{"OAuth 2.0", "OIDC", "SAML 2.0"},
		Endpoints: ssoEndpoints{
			Authorization: "https://auth.platform.io/oauth2/authorize",
			Token:         "https://auth.platform.io/oauth2/token",
			UserInfo:      "https://auth.platform.io/oauth2/userinfo",
			JWKS:          "https://auth.platform.io/.well-known/jwks.json",
			SAMLMetadata:  "https://auth.platform.io/saml/metadata",
		},
		Providers: []ssoProvider{
			{Name: "Google Workspace", Protocol: "OIDC", Status: "supported"},
			{Name: "Microsoft Entra ID", Protocol: "OIDC / SAML", Status: "supported"},
			{Name: "Okta", Protocol: "OIDC / SAML", Status: "supported"},
			{Name: "GitHub", Protocol: "OAuth 2.0", Status: "supported"},
			{Name: "GitLab", Protocol: "OAuth 2.0", Status: "supported"},
			{Name: "Custom SAML 2.0 IdP", Protocol: "SAML 2.0", Status: "supported"},
		},
		Scopes:       []string{"openid", "profile", "email", "groups", "offline_access"},
		GrantTypes:   []string{"authorization_code", "refresh_token", "client_credentials"},
		PKCERequired: true,
		TokenFormat:  "JWT",
		SigningAlgo:  "RS256",
	}
	apiOK(w, data)
}
