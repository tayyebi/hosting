package handlers

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"platform/db"
)

type accountingPageData struct {
	MRRDisplay         string
	SupplyCostDisplay  string
	GrossMarginDisplay string
	ActiveResources    int
	APILogs            []db.BillingAPILog
	Webhooks           []db.BillingWebhook
}

func formatCentsUSD(cents int64) string {
	sign := ""
	if cents < 0 {
		sign = "-"
		cents = -cents
	}
	return fmt.Sprintf("%s$%d.%02d", sign, cents/100, cents%100)
}

func accountingHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := getSession(r)

	var mrrCents, supplyCents int64
	var activeRes int
	var apiLogs []db.BillingAPILog
	var webhooks []db.BillingWebhook

	if Pools != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		var err error
		apiLogs, err = db.ListBillingLogs(ctx, Pools.Admin, 50)
		if err != nil {
			log.Printf("accounting: list billing api logs: %v", err)
		}
		webhooks, err = db.ListBillingWebhooks(ctx, Pools.Admin)
		if err != nil {
			log.Printf("accounting: list billing webhooks: %v", err)
		}

		// Live cost-centric accounting rollups from active resources & supply_hosts.
		_ = Pools.Admin.QueryRow(ctx, `
			SELECT COUNT(*)::int,
			       COALESCE(SUM(retail_price_cents), 0)
			FROM resources
			WHERE deleted_at IS NULL AND status IN ('active', 'provisioning')`,
		).Scan(&activeRes, &mrrCents)

		_ = Pools.Admin.QueryRow(ctx, `
			SELECT COALESCE(SUM(monthly_cost_cents), 0)
			FROM supply_hosts
			WHERE deleted_at IS NULL AND inventory_status IN ('online', 'maintenance')`,
		).Scan(&supplyCents)
	}

	pageData := accountingPageData{
		MRRDisplay:         formatCentsUSD(mrrCents),
		SupplyCostDisplay:  formatCentsUSD(supplyCents),
		GrossMarginDisplay: formatCentsUSD(mrrCents - supplyCents),
		ActiveResources:    activeRes,
		APILogs:            apiLogs,
		Webhooks:           webhooks,
	}

	innerTmpl, err := template.ParseFiles(
		filepath.Join("templates", "sysadmin", "accounting_content.html"),
	)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	var buf bytes.Buffer
	if err := innerTmpl.ExecuteTemplate(&buf, "sysadmin-accounting-content", pageData); err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	shellTmpl, err := template.ParseFiles(
		filepath.Join("templates", "sysadmin", "base.html"),
		filepath.Join("templates", "sysadmin", "shell.html"),
	)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	data := SysadminData{
		User:      user,
		Nav:       sysadminNav,
		ActiveNav: "accounting",
		PageTitle: "Accounting System",
		Content:   template.HTML(buf.String()),
	}
	if err := shellTmpl.ExecuteTemplate(w, "sysadmin-base", data); err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
	}
}

func RegisterAccountingRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/s/accounting/", requireRole(RoleSysAdmin, accountingHandler))
}
