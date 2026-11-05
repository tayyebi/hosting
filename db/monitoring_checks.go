package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type MonitoringCheck struct {
	ID          int64
	TenantID    int64
	Name        string
	CheckType   string
	Target      string
	Status      string
	Uptime30d   string
	LatencyMs   int
	LastChecked *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func ListMonitoringChecks(ctx context.Context, pool *pgxpool.Pool, tenantID int64, search, status string) ([]MonitoringCheck, error) {
	args := []any{tenantID}
	conds := []string{"tenant_id = $1", "deleted_at IS NULL"}

	if search != "" {
		args = append(args, "%"+strings.ToLower(search)+"%")
		n := len(args)
		conds = append(conds, fmt.Sprintf(
			"(lower(name) LIKE $%d OR lower(target) LIKE $%d)", n, n))
	}
	if status != "" {
		args = append(args, status)
		conds = append(conds, fmt.Sprintf("status = $%d", len(args)))
	}

	q := fmt.Sprintf(`
		SELECT id, tenant_id, name, check_type, target, status, uptime_30d::text, latency_ms, last_checked, created_at, updated_at
		FROM monitoring_checks
		WHERE %s
		ORDER BY name ASC`, strings.Join(conds, " AND "))

	rows, err := pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("db.ListMonitoringChecks: %w", err)
	}
	defer rows.Close()

	var out []MonitoringCheck
	for rows.Next() {
		var r MonitoringCheck
		if err := rows.Scan(&r.ID, &r.TenantID, &r.Name, &r.CheckType, &r.Target,
			&r.Status, &r.Uptime30d, &r.LatencyMs, &r.LastChecked,
			&r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, fmt.Errorf("db.ListMonitoringChecks scan: %w", err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func DeleteMonitoringCheck(ctx context.Context, pool *pgxpool.Pool, tenantID, id int64) error {
	_, err := pool.Exec(ctx, `
		UPDATE monitoring_checks SET deleted_at=now(), updated_at=now()
		WHERE tenant_id=$1 AND id=$2 AND deleted_at IS NULL`,
		tenantID, id)
	if err != nil {
		return fmt.Errorf("db.DeleteMonitoringCheck: %w", err)
	}
	return nil
}

func CreateMonitoringCheck(ctx context.Context, pool *pgxpool.Pool, tenantID int64, name, checkType, target string) (MonitoringCheck, error) {
	var r MonitoringCheck
	err := pool.QueryRow(ctx, `
		INSERT INTO monitoring_checks (tenant_id, name, check_type, target)
		VALUES ($1, $2, $3, $4)
		RETURNING id, tenant_id, name, check_type, target, status, uptime_30d::text, latency_ms, last_checked, created_at, updated_at`,
		tenantID, name, checkType, target,
	).Scan(&r.ID, &r.TenantID, &r.Name, &r.CheckType, &r.Target,
		&r.Status, &r.Uptime30d, &r.LatencyMs, &r.LastChecked,
		&r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		return MonitoringCheck{}, fmt.Errorf("db.CreateMonitoringCheck: %w", err)
	}
	return r, nil
}
