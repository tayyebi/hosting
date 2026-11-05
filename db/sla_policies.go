package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SLAPolicy struct {
	ID           int64
	TenantID     int64
	Name         string
	Service      string
	Status       string
	UptimeTarget float64
	UptimeActual float64
	ResponseTime string
	CreditPolicy string
	PenaltyRate  string
	ReviewPeriod string
	LastReviewed *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func ListSLAPolicies(ctx context.Context, pool *pgxpool.Pool, tenantID int64, search, status string) ([]SLAPolicy, error) {
	args := []any{tenantID}
	conds := []string{"tenant_id = $1", "deleted_at IS NULL"}

	if search != "" {
		args = append(args, "%"+strings.ToLower(search)+"%")
		n := len(args)
		conds = append(conds, fmt.Sprintf(
			"(lower(name) LIKE $%d OR lower(service) LIKE $%d)", n, n))
	}
	if status != "" {
		args = append(args, status)
		conds = append(conds, fmt.Sprintf("status = $%d", len(args)))
	}

	q := fmt.Sprintf(`
		SELECT id, tenant_id, name, service, status, uptime_target, uptime_actual,
		       response_time, credit_policy, penalty_rate, review_period, last_reviewed,
		       created_at, updated_at
		FROM sla_policies
		WHERE %s
		ORDER BY created_at DESC`, strings.Join(conds, " AND "))

	rows, err := pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("db.ListSLAPolicies: %w", err)
	}
	defer rows.Close()

	var out []SLAPolicy
	for rows.Next() {
		var p SLAPolicy
		if err := rows.Scan(&p.ID, &p.TenantID, &p.Name, &p.Service, &p.Status,
			&p.UptimeTarget, &p.UptimeActual, &p.ResponseTime, &p.CreditPolicy,
			&p.PenaltyRate, &p.ReviewPeriod, &p.LastReviewed,
			&p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("db.ListSLAPolicies scan: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func CreateSLAPolicy(ctx context.Context, pool *pgxpool.Pool, tenantID int64, name, service string, uptimeTarget float64, responseTime, creditPolicy, penaltyRate, reviewPeriod string, lastReviewed *time.Time) (SLAPolicy, error) {
	var p SLAPolicy
	err := pool.QueryRow(ctx, `
		INSERT INTO sla_policies (tenant_id, name, service, uptime_target, response_time, credit_policy, penalty_rate, review_period, last_reviewed)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, tenant_id, name, service, status, uptime_target, uptime_actual,
		          response_time, credit_policy, penalty_rate, review_period, last_reviewed,
		          created_at, updated_at`,
		tenantID, name, service, uptimeTarget, responseTime, creditPolicy, penaltyRate, reviewPeriod, lastReviewed,
	).Scan(&p.ID, &p.TenantID, &p.Name, &p.Service, &p.Status,
		&p.UptimeTarget, &p.UptimeActual, &p.ResponseTime, &p.CreditPolicy,
		&p.PenaltyRate, &p.ReviewPeriod, &p.LastReviewed,
		&p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return SLAPolicy{}, fmt.Errorf("db.CreateSLAPolicy: %w", err)
	}
	return p, nil
}

func GetSLAPolicy(ctx context.Context, pool *pgxpool.Pool, tenantID, id int64) (SLAPolicy, error) {
	var p SLAPolicy
	err := pool.QueryRow(ctx, `
		SELECT id, tenant_id, name, service, status, uptime_target, uptime_actual,
		       response_time, credit_policy, penalty_rate, review_period, last_reviewed,
		       created_at, updated_at
		FROM sla_policies
		WHERE tenant_id = $1 AND id = $2 AND deleted_at IS NULL`,
		tenantID, id,
	).Scan(&p.ID, &p.TenantID, &p.Name, &p.Service, &p.Status,
		&p.UptimeTarget, &p.UptimeActual, &p.ResponseTime, &p.CreditPolicy,
		&p.PenaltyRate, &p.ReviewPeriod, &p.LastReviewed,
		&p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return SLAPolicy{}, fmt.Errorf("db.GetSLAPolicy: %w", err)
	}
	return p, nil
}

func DeleteSLAPolicy(ctx context.Context, pool *pgxpool.Pool, tenantID, id int64) error {
	_, err := pool.Exec(ctx, `
		UPDATE sla_policies SET deleted_at=now(), updated_at=now()
		WHERE tenant_id=$1 AND id=$2 AND deleted_at IS NULL`,
		tenantID, id)
	if err != nil {
		return fmt.Errorf("db.DeleteSLAPolicy: %w", err)
	}
	return nil
}
