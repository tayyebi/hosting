package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BillingSettings is the single-row config from billing_settings.
type BillingSettings struct {
	ID        int64
	URL       string
	TimeoutS  int
	Retry     bool
	LogAll    bool
	Enabled   bool
	UpdatedAt time.Time
}

// BillingAPILog is one row from billing_api_logs.
type BillingAPILog struct {
	ID           int64
	Method       string
	Path         string
	TenantRef    string
	StatusCode   int
	LatencyMS    int
	RequestBody  string
	ResponseBody string
	CreatedAt    time.Time
}

// BillingWebhook is one row from billing_webhooks.
type BillingWebhook struct {
	ID            int64
	Endpoint      string
	Events        string
	Status        string
	DeliveryOK    int
	DeliveryTotal int
	CreatedAt     time.Time
}

// DeliveryRate returns a formatted delivery rate string, e.g. "99.8%".
func (w BillingWebhook) DeliveryRate() string {
	if w.DeliveryTotal == 0 {
		return "—"
	}
	pct := float64(w.DeliveryOK) / float64(w.DeliveryTotal) * 100
	return fmt.Sprintf("%.1f%%", pct)
}

// GetBillingSettings returns the single billing_settings row, or a sensible
// default when none exists.
func GetBillingSettings(ctx context.Context, pool *pgxpool.Pool) (BillingSettings, error) {
	var s BillingSettings
	err := pool.QueryRow(ctx, `
		SELECT id, url, timeout_s, retry, log_all, enabled, updated_at
		FROM billing_settings
		ORDER BY id LIMIT 1`).Scan(
		&s.ID, &s.URL, &s.TimeoutS, &s.Retry, &s.LogAll, &s.Enabled, &s.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return BillingSettings{URL: "(not configured)", TimeoutS: 10, Retry: true}, nil
	}
	if err != nil {
		return s, fmt.Errorf("db.GetBillingSettings: %w", err)
	}
	return s, nil
}

// ListBillingLogs returns the most recent billing API log entries.
func ListBillingLogs(ctx context.Context, pool *pgxpool.Pool, limit int) ([]BillingAPILog, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := pool.Query(ctx, `
		SELECT id, method, path, tenant_ref, status_code, latency_ms,
		       COALESCE(request_body,''), COALESCE(response_body,''), created_at
		FROM billing_api_logs
		ORDER BY created_at DESC
		LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("db.ListBillingLogs: %w", err)
	}
	defer rows.Close()
	var out []BillingAPILog
	for rows.Next() {
		var l BillingAPILog
		if err := rows.Scan(
			&l.ID, &l.Method, &l.Path, &l.TenantRef, &l.StatusCode,
			&l.LatencyMS, &l.RequestBody, &l.ResponseBody, &l.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

// ListBillingWebhooks returns all billing webhook rows.
func ListBillingWebhooks(ctx context.Context, pool *pgxpool.Pool) ([]BillingWebhook, error) {
	rows, err := pool.Query(ctx, `
		SELECT id, endpoint, events, status, delivery_ok, delivery_total, created_at
		FROM billing_webhooks
		ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("db.ListBillingWebhooks: %w", err)
	}
	defer rows.Close()
	var out []BillingWebhook
	for rows.Next() {
		var w BillingWebhook
		if err := rows.Scan(
			&w.ID, &w.Endpoint, &w.Events, &w.Status,
			&w.DeliveryOK, &w.DeliveryTotal, &w.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

// BillingKPIStats aggregates quick KPIs from billing_api_logs.
type BillingKPIStats struct {
	TotalCalls    int
	SuccessCalls  int
	ErrorCalls    int
	AvgLatencyMS  int
	LastCallAge   string // human-readable "N seconds ago"
	LastCallTime  *time.Time
}

// BillingKPIs aggregates key stats from billing_api_logs.
func BillingKPIs(ctx context.Context, pool *pgxpool.Pool) (BillingKPIStats, error) {
	var s BillingKPIStats
	err := pool.QueryRow(ctx, `
		SELECT
			count(*)::int,
			count(*) FILTER (WHERE status_code < 400)::int,
			count(*) FILTER (WHERE status_code >= 400)::int,
			COALESCE(avg(latency_ms)::int, 0),
			max(created_at)
		FROM billing_api_logs`).Scan(
		&s.TotalCalls, &s.SuccessCalls, &s.ErrorCalls, &s.AvgLatencyMS, &s.LastCallTime,
	)
	if err != nil {
		return s, fmt.Errorf("db.BillingKPIs: %w", err)
	}
	if s.LastCallTime != nil {
		age := time.Since(*s.LastCallTime)
		switch {
		case age < time.Minute:
			s.LastCallAge = fmt.Sprintf("%d seconds ago", int(age.Seconds()))
		case age < time.Hour:
			s.LastCallAge = fmt.Sprintf("%d minutes ago", int(age.Minutes()))
		default:
			s.LastCallAge = fmt.Sprintf("%d hours ago", int(age.Hours()))
		}
	} else {
		s.LastCallAge = "never"
	}
	return s, nil
}
