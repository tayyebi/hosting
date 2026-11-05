package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Incident struct {
	ID          int64
	TenantID    int64
	RefNum      string
	Title       string
	Severity    string
	Status      string
	Affected    string
	Description string
	StartedAt   time.Time
	ResolvedAt  *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func ListIncidents(ctx context.Context, pool *pgxpool.Pool, tenantID int64, search, severity, status string) ([]Incident, error) {
	args := []any{tenantID}
	conds := []string{"tenant_id = $1", "deleted_at IS NULL"}

	if search != "" {
		args = append(args, "%"+strings.ToLower(search)+"%")
		n := len(args)
		conds = append(conds, fmt.Sprintf(
			"(lower(title) LIKE $%d OR lower(ref_num) LIKE $%d OR lower(affected) LIKE $%d)", n, n, n))
	}
	if severity != "" {
		args = append(args, severity)
		conds = append(conds, fmt.Sprintf("severity = $%d", len(args)))
	}
	if status != "" {
		args = append(args, status)
		conds = append(conds, fmt.Sprintf("status = $%d", len(args)))
	}

	q := fmt.Sprintf(`
		SELECT id, tenant_id, ref_num, title, severity, status, affected, description, started_at, resolved_at, created_at, updated_at
		FROM incidents
		WHERE %s
		ORDER BY created_at DESC`, strings.Join(conds, " AND "))

	rows, err := pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("db.ListIncidents: %w", err)
	}
	defer rows.Close()

	var out []Incident
	for rows.Next() {
		var i Incident
		if err := rows.Scan(&i.ID, &i.TenantID, &i.RefNum, &i.Title, &i.Severity,
			&i.Status, &i.Affected, &i.Description, &i.StartedAt, &i.ResolvedAt,
			&i.CreatedAt, &i.UpdatedAt); err != nil {
			return nil, fmt.Errorf("db.ListIncidents scan: %w", err)
		}
		out = append(out, i)
	}
	return out, rows.Err()
}

func UpdateIncidentStatus(ctx context.Context, pool *pgxpool.Pool, tenantID, id int64, status string) error {
	resolvedUpdate := ""
	if status == "resolved" {
		resolvedUpdate = ", resolved_at=now()"
	}
	_, err := pool.Exec(ctx, fmt.Sprintf(`
		UPDATE incidents SET status=$3, updated_at=now()%s
		WHERE tenant_id=$1 AND id=$2 AND deleted_at IS NULL`, resolvedUpdate),
		tenantID, id, status)
	if err != nil {
		return fmt.Errorf("db.UpdateIncidentStatus: %w", err)
	}
	return nil
}

func DeleteIncident(ctx context.Context, pool *pgxpool.Pool, tenantID, id int64) error {
	_, err := pool.Exec(ctx, `
		UPDATE incidents SET deleted_at=now(), updated_at=now()
		WHERE tenant_id=$1 AND id=$2 AND deleted_at IS NULL`,
		tenantID, id)
	if err != nil {
		return fmt.Errorf("db.DeleteIncident: %w", err)
	}
	return nil
}

func CreateIncident(ctx context.Context, pool *pgxpool.Pool, tenantID int64, refNum, title, severity, affected, description string) (Incident, error) {
	var i Incident
	err := pool.QueryRow(ctx, `
		INSERT INTO incidents (tenant_id, ref_num, title, severity, affected, description)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, tenant_id, ref_num, title, severity, status, affected, description, started_at, resolved_at, created_at, updated_at`,
		tenantID, refNum, title, severity, affected, description,
	).Scan(&i.ID, &i.TenantID, &i.RefNum, &i.Title, &i.Severity, &i.Status,
		&i.Affected, &i.Description, &i.StartedAt, &i.ResolvedAt, &i.CreatedAt, &i.UpdatedAt)
	if err != nil {
		return Incident{}, fmt.Errorf("db.CreateIncident: %w", err)
	}
	return i, nil
}
