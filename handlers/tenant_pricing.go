package handlers

import (
	"net/http"
	"path/filepath"
)

func RegisterTenantPricingRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /t/{slug}/pricing/", tenantPage("pricing", "Pricing",
		filepath.Join("templates", "tenant", "pricing_content.html"),
	))
}
