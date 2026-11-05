package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AbuseReport struct {
	ID        int64
	TenantID  int64
	Type      string
	Source    string
	Target    string
	Status    string
	Summary   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func ListAbuseReports(ctx context.Context, pool *pgxpool.Pool, tenantID int64, search, status, abuseType string) ([]AbuseReport, error) {
	args := []any{tenantID}
	conds := []string{"tenant_id = $1", "deleted_at IS NULL"}

	if search != "" {
		args = append(args, "%"+strings.ToLower(search)+"%")
		n := len(args)
		conds = append(conds, fmt.Sprintf(
			"(lower(summary) LIKE $%d OR lower(source) LIKE $%d OR lower(target) LIKE $%d)", n, n, n))
	}
	if status != "" {
		args = append(args, status)
		conds = append(conds, fmt.Sprintf("status = $%d", len(args)))
	}
	if abuseType != "" {
		args = append(args, abuseType)
		conds = append(conds, fmt.Sprintf("type = $%d", len(args)))
	}

	q := fmt.Sprintf(`
		SELECT id, tenant_id, type, source, target, status, summary, created_at, updated_at
		FROM abuse_reports
		WHERE %s
		ORDER BY created_at DESC`, strings.Join(conds, " AND "))

	rows, err := pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("db.ListAbuseReports: %w", err)
	}
	defer rows.Close()

	var out []AbuseReport
	for rows.Next() {
		var r AbuseReport
		if err := rows.Scan(&r.ID, &r.TenantID, &r.Type, &r.Source, &r.Target,
			&r.Status, &r.Summary, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, fmt.Errorf("db.ListAbuseReports scan: %w", err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func UpdateAbuseReportStatus(ctx context.Context, pool *pgxpool.Pool, tenantID, id int64, status string) error {
	_, err := pool.Exec(ctx, `
		UPDATE abuse_reports SET status=$3, updated_at=now()
		WHERE tenant_id=$1 AND id=$2 AND deleted_at IS NULL`,
		tenantID, id, status)
	if err != nil {
		return fmt.Errorf("db.UpdateAbuseReportStatus: %w", err)
	}
	return nil
}

func DeleteAbuseReport(ctx context.Context, pool *pgxpool.Pool, tenantID, id int64) error {
	_, err := pool.Exec(ctx, `
		UPDATE abuse_reports SET deleted_at=now(), updated_at=now()
		WHERE tenant_id=$1 AND id=$2 AND deleted_at IS NULL`,
		tenantID, id)
	if err != nil {
		return fmt.Errorf("db.DeleteAbuseReport: %w", err)
	}
	return nil
}

func CreateAbuseReport(ctx context.Context, pool *pgxpool.Pool, tenantID int64, abuseType, source, target, summary string) (AbuseReport, error) {
	var r AbuseReport
	err := pool.QueryRow(ctx, `
		INSERT INTO abuse_reports (tenant_id, type, source, target, summary)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, tenant_id, type, source, target, status, summary, created_at, updated_at`,
		tenantID, abuseType, source, target, summary,
	).Scan(&r.ID, &r.TenantID, &r.Type, &r.Source, &r.Target,
		&r.Status, &r.Summary, &r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		return AbuseReport{}, fmt.Errorf("db.CreateAbuseReport: %w", err)
	}
	return r, nil
}
