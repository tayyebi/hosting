package db

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LocalizationSettings struct {
	TenantID       int64
	Language       string
	Timezone       string
	RTLEnabled     bool
	DateFormat     string
	TimeFormat     string
	NumberFormat   string
	FirstDayOfWeek string
	Currency       string
	SymbolPosition string
	TaxInclusive   bool
	TaxBreakdown   bool
	RoutingRules   json.RawMessage
	APILocale      json.RawMessage
	CDNStats       json.RawMessage
}

func GetLocalizationSettings(ctx context.Context, pool *pgxpool.Pool, tenantID int64) (LocalizationSettings, bool, error) {
	var s LocalizationSettings
	err := pool.QueryRow(ctx, `
		SELECT tenant_id, language, timezone, rtl_enabled, date_format, time_format,
		       number_format, first_day_of_week, currency, symbol_position,
		       tax_inclusive, tax_breakdown, routing_rules, api_locale, cdn_stats
		FROM tenant_localization_settings
		WHERE tenant_id = $1`, tenantID,
	).Scan(&s.TenantID, &s.Language, &s.Timezone, &s.RTLEnabled,
		&s.DateFormat, &s.TimeFormat, &s.NumberFormat, &s.FirstDayOfWeek,
		&s.Currency, &s.SymbolPosition, &s.TaxInclusive, &s.TaxBreakdown,
		&s.RoutingRules, &s.APILocale, &s.CDNStats)
	if err != nil {
		if err == pgx.ErrNoRows {
			return LocalizationSettings{}, false, nil
		}
		return LocalizationSettings{}, false, fmt.Errorf("db.GetLocalizationSettings: %w", err)
	}
	return s, true, nil
}

func UpsertLocalizationSettings(ctx context.Context, pool *pgxpool.Pool, tenantID int64, language, timezone, dateFormat, timeFormat, currency string, rtlEnabled bool) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO tenant_localization_settings (tenant_id, language, timezone, date_format, time_format, currency, rtl_enabled)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (tenant_id) DO UPDATE SET
		    language    = EXCLUDED.language,
		    timezone    = EXCLUDED.timezone,
		    date_format = EXCLUDED.date_format,
		    time_format = EXCLUDED.time_format,
		    currency    = EXCLUDED.currency,
		    rtl_enabled = EXCLUDED.rtl_enabled,
		    updated_at  = now()`,
		tenantID, language, timezone, dateFormat, timeFormat, currency, rtlEnabled)
	if err != nil {
		return fmt.Errorf("db.UpsertLocalizationSettings: %w", err)
	}
	return nil
}
