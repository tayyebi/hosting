package handlers

import (
	"net/http"
	"path/filepath"
)

func RegisterTenantBrandRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /t/{slug}/brand/", tenantPage("brand", "Brand",
		filepath.Join("templates", "tenant", "brand_content.html"),
	))
}
