package handlers

import (
	"net/http"
	"path/filepath"
)

func RegisterTenantTaxRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /t/{slug}/tax/", tenantPage("tax", "Tax",
		filepath.Join("templates", "tenant", "tax_content.html"),
	))
}
