package handlers

import (
	"net/http"
	"path/filepath"
)

func RegisterTenantRegistrationRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /t/{slug}/registration/", tenantPage("registration", "Registration",
		filepath.Join("templates", "tenant", "registration_content.html"),
	))
}
