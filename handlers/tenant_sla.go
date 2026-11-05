package handlers

import (
	"net/http"
	"path/filepath"
)

func RegisterTenantSLARoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /t/{slug}/sla/", tenantPage("sla", "SLA",
		filepath.Join("templates", "tenant", "sla_content.html"),
	))
}
