package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"platform/db"
)

// ── GET /api/sysadmin/accounting ───────────────────────────

type apiAcctKPI struct {
	Label string `json:"label"`
	Value string `json:"value"`
	Delta string `json:"delta"`
	Up    bool   `json:"up"`
}

type apiAcctLogEntry struct {
	Status   int    `json:"status"`
	Method   string `json:"method"`
	Path     string `json:"path"`
	Tenant   string `json:"tenant"`
	Latency  string `json:"latency"`
	Time     string `json:"time"`
	Request  string `json:"request"`
	Response string `json:"response"`
}

type apiAcctWebhook struct {
	Endpoint     string `json:"endpoint"`
	Events       string `json:"events"`
	Status       string `json:"status"`
	DeliveryRate string `json:"delivery_rate"`
}

type apiAcctConnection struct {
	URL       string `json:"url"`
	Timeout   string `json:"timeout"`
	Retry     bool   `json:"retry"`
	LogAll    bool   `json:"log_all"`
	Connected bool   `json:"connected"`
	LastCall  string `json:"last_call"`
}

type apiAccountingResponse struct {
	KPIs       []apiAcctKPI      `json:"kpis"`
	Logs       []apiAcctLogEntry `json:"logs"`
	Webhooks   []apiAcctWebhook  `json:"webhooks"`
	Connection apiAcctConnection `json:"connection"`
}

func apiSysadminAccountingHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok || user.Role != RoleSysAdmin {
		apiError(w, http.StatusForbidden, "Sysadmin access required.")
		return
	}

	if Pools == nil {
		apiOK(w, apiAccountingResponse{
			Connection: apiAcctConnection{URL: "(not configured)", Timeout: "10 seconds"},
		})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Billing settings (connection info)
	settings, _ := db.GetBillingSettings(ctx, Pools.Admin)
	kpiStats, _ := db.BillingKPIs(ctx, Pools.Admin)

	conn := apiAcctConnection{
		URL:       settings.URL,
		Timeout:   fmt.Sprintf("%d seconds", settings.TimeoutS),
		Retry:     settings.Retry,
		LogAll:    settings.LogAll,
		Connected: settings.Enabled,
		LastCall:  kpiStats.LastCallAge,
	}

	// KPIs from DB stats
	successRate := "—"
	if kpiStats.TotalCalls > 0 {
		successRate = fmt.Sprintf("%.1f%%", float64(kpiStats.SuccessCalls)/float64(kpiStats.TotalCalls)*100)
	}
	kpis := []apiAcctKPI{
		{Label: "Total API Calls", Value: fmt.Sprintf("%d", kpiStats.TotalCalls), Delta: fmt.Sprintf("%d errors", kpiStats.ErrorCalls), Up: kpiStats.ErrorCalls == 0},
		{Label: "Success Rate", Value: successRate, Delta: fmt.Sprintf("%d failed calls", kpiStats.ErrorCalls), Up: kpiStats.ErrorCalls == 0},
		{Label: "Avg Latency", Value: fmt.Sprintf("%d ms", kpiStats.AvgLatencyMS), Delta: "last 7 days", Up: kpiStats.AvgLatencyMS < 500},
		{Label: "Last Call", Value: kpiStats.LastCallAge, Delta: "billing API", Up: true},
	}

	// API logs from DB
	dbLogs, _ := db.ListBillingLogs(ctx, Pools.Admin, 20)
	logs := make([]apiAcctLogEntry, 0, len(dbLogs))
	for _, l := range dbLogs {
		age := time.Since(l.CreatedAt)
		var ageStr string
		switch {
		case age < time.Minute:
			ageStr = fmt.Sprintf("%d sec ago", int(age.Seconds()))
		case age < time.Hour:
			ageStr = fmt.Sprintf("%d min ago", int(age.Minutes()))
		default:
			ageStr = fmt.Sprintf("%d hr ago", int(age.Hours()))
		}
		logs = append(logs, apiAcctLogEntry{
			Status:   l.StatusCode,
			Method:   l.Method,
			Path:     l.Path,
			Tenant:   l.TenantRef,
			Latency:  fmt.Sprintf("%d ms", l.LatencyMS),
			Time:     ageStr,
			Request:  l.RequestBody,
			Response: l.ResponseBody,
		})
	}

	// Webhooks from DB
	dbWH, _ := db.ListBillingWebhooks(ctx, Pools.Admin)
	webhooks := make([]apiAcctWebhook, 0, len(dbWH))
	for _, wh := range dbWH {
		webhooks = append(webhooks, apiAcctWebhook{
			Endpoint:     wh.Endpoint,
			Events:       wh.Events,
			Status:       wh.Status,
			DeliveryRate: wh.DeliveryRate(),
		})
	}

	apiOK(w, apiAccountingResponse{
		KPIs:       kpis,
		Logs:       logs,
		Webhooks:   webhooks,
		Connection: conn,
	})
}
