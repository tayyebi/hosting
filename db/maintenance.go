package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type MaintenanceWindow struct {
	ID             int64
	TenantID       int64
	RefNum         string
	Title          string
	Description    string
	Affected       string
	ScheduledStart time.Time
	ScheduledEnd   time.Time
	Risk           string
	Status         string
	Rollback       string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func ListMaintenanceWindows(ctx context.Context, pool *pgxpool.Pool, tenantID int64, search, status string) ([]MaintenanceWindow, error) {
	args := []any{tenantID}
	conds := []string{"tenant_id = $1", "deleted_at IS NULL"}

	if search != "" {
		args = append(args, "%"+strings.ToLower(search)+"%")
		n := len(args)
		conds = append(conds, fmt.Sprintf(
			"(lower(title) LIKE $%d OR lower(ref_num) LIKE $%d OR lower(affected) LIKE $%d)", n, n, n))
	}
	if status != "" {
		args = append(args, status)
		conds = append(conds, fmt.Sprintf("status = $%d", len(args)))
	}

	q := fmt.Sprintf(`
		SELECT id, tenant_id, ref_num, title, description, affected, scheduled_start, scheduled_end, risk, status, rollback, created_at, updated_at
		FROM maintenance_windows
		WHERE %s
		ORDER BY scheduled_start DESC`, strings.Join(conds, " AND "))

	rows, err := pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("db.ListMaintenanceWindows: %w", err)
	}
	defer rows.Close()

	var out []MaintenanceWindow
	for rows.Next() {
		var w MaintenanceWindow
		if err := rows.Scan(&w.ID, &w.TenantID, &w.RefNum, &w.Title, &w.Description,
			&w.Affected, &w.ScheduledStart, &w.ScheduledEnd, &w.Risk, &w.Status,
			&w.Rollback, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, fmt.Errorf("db.ListMaintenanceWindows scan: %w", err)
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

func UpdateMaintenanceWindowStatus(ctx context.Context, pool *pgxpool.Pool, tenantID, id int64, status string) error {
	_, err := pool.Exec(ctx, `
		UPDATE maintenance_windows SET status=$3, updated_at=now()
		WHERE tenant_id=$1 AND id=$2 AND deleted_at IS NULL`,
		tenantID, id, status)
	if err != nil {
		return fmt.Errorf("db.UpdateMaintenanceWindowStatus: %w", err)
	}
	return nil
}

func DeleteMaintenanceWindow(ctx context.Context, pool *pgxpool.Pool, tenantID, id int64) error {
	_, err := pool.Exec(ctx, `
		UPDATE maintenance_windows SET deleted_at=now(), updated_at=now()
		WHERE tenant_id=$1 AND id=$2 AND deleted_at IS NULL`,
		tenantID, id)
	if err != nil {
		return fmt.Errorf("db.DeleteMaintenanceWindow: %w", err)
	}
	return nil
}

func CreateMaintenanceWindow(ctx context.Context, pool *pgxpool.Pool, tenantID int64, refNum, title, description, affected, risk, rollback string, scheduledStart, scheduledEnd time.Time) (MaintenanceWindow, error) {
	var w MaintenanceWindow
	err := pool.QueryRow(ctx, `
		INSERT INTO maintenance_windows (tenant_id, ref_num, title, description, affected, risk, rollback, scheduled_start, scheduled_end)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, tenant_id, ref_num, title, description, affected, scheduled_start, scheduled_end, risk, status, rollback, created_at, updated_at`,
		tenantID, refNum, title, description, affected, risk, rollback, scheduledStart, scheduledEnd,
	).Scan(&w.ID, &w.TenantID, &w.RefNum, &w.Title, &w.Description, &w.Affected,
		&w.ScheduledStart, &w.ScheduledEnd, &w.Risk, &w.Status, &w.Rollback, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		return MaintenanceWindow{}, fmt.Errorf("db.CreateMaintenanceWindow: %w", err)
	}
	return w, nil
}
