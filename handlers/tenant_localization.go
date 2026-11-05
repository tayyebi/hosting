package handlers

import (
	"net/http"
	"path/filepath"
)

func RegisterTenantLocalizationRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /t/{slug}/localization/", tenantPage("localization", "Localization",
		filepath.Join("templates", "tenant", "localization_content.html"),
	))
}
