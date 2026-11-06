package handlers

import (
	"html/template"
	"net/http"
	"path/filepath"
)

type TenantBrandPageData struct {
	TenantShellData
	Brand TenantLoginBrand
}

func tenantBrandPageHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := getSession(r)
	tenant, _ := tenantFromContext(r)
	brand := resolveTenantLoginBrand(r.Context(), tenant)

	paths := []string{
		filepath.Join("templates", "tenant", "base.html"),
		filepath.Join("templates", "tenant", "shell.html"),
		filepath.Join("templates", "tenant", "brand_content.html"),
	}
	tmpl, err := template.ParseFiles(paths...)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := TenantBrandPageData{
		TenantShellData: TenantShellData{
			User:       user,
			Tenant:     tenant,
			Nav:        tenantNav(tenant.Slug),
			NavGroups:  tenantNavGroups(tenant.Slug),
			ActiveNav:  "brand",
			PageTitle:  "Brand",
			IsSysAdmin: user.Role == RoleSysAdmin,
		},
		Brand: brand,
	}
	if err := tmpl.ExecuteTemplate(w, "tenant-base", data); err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
	}
}

func RegisterTenantBrandRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /t/{slug}/brand/", tenantMiddleware(tenantBrandPageHandler))
}
