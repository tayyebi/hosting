package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type StatusComponent struct {
	ID          int64
	TenantID    int64
	Name        string
	GroupName   string
	Description string
	Status      string
	Uptime90d   float64
	OrderNum    int
	Visible     bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func ListStatusComponents(ctx context.Context, pool *pgxpool.Pool, tenantID int64, search string) ([]StatusComponent, error) {
	args := []any{tenantID}
	conds := []string{"tenant_id = $1", "deleted_at IS NULL"}

	if search != "" {
		args = append(args, "%"+strings.ToLower(search)+"%")
		n := len(args)
		conds = append(conds, fmt.Sprintf(
			"(lower(name) LIKE $%d OR lower(group_name) LIKE $%d)", n, n))
	}

	q := fmt.Sprintf(`
		SELECT id, tenant_id, name, group_name, description, status, uptime_90d, order_num, visible, created_at, updated_at
		FROM status_components
		WHERE %s
		ORDER BY order_num ASC, name ASC`, strings.Join(conds, " AND "))

	rows, err := pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("db.ListStatusComponents: %w", err)
	}
	defer rows.Close()

	var out []StatusComponent
	for rows.Next() {
		var c StatusComponent
		if err := rows.Scan(&c.ID, &c.TenantID, &c.Name, &c.GroupName, &c.Description,
			&c.Status, &c.Uptime90d, &c.OrderNum, &c.Visible, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("db.ListStatusComponents scan: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func UpdateStatusComponentStatus(ctx context.Context, pool *pgxpool.Pool, tenantID, id int64, status string) error {
	_, err := pool.Exec(ctx, `
		UPDATE status_components SET status=$3, updated_at=now()
		WHERE tenant_id=$1 AND id=$2 AND deleted_at IS NULL`,
		tenantID, id, status)
	if err != nil {
		return fmt.Errorf("db.UpdateStatusComponentStatus: %w", err)
	}
	return nil
}

func DeleteStatusComponent(ctx context.Context, pool *pgxpool.Pool, tenantID, id int64) error {
	_, err := pool.Exec(ctx, `
		UPDATE status_components SET deleted_at=now(), updated_at=now()
		WHERE tenant_id=$1 AND id=$2 AND deleted_at IS NULL`,
		tenantID, id)
	if err != nil {
		return fmt.Errorf("db.DeleteStatusComponent: %w", err)
	}
	return nil
}

func CreateStatusComponent(ctx context.Context, pool *pgxpool.Pool, tenantID int64, name, groupName, description string) (StatusComponent, error) {
	var c StatusComponent
	err := pool.QueryRow(ctx, `
		INSERT INTO status_components (tenant_id, name, group_name, description)
		VALUES ($1, $2, $3, $4)
		RETURNING id, tenant_id, name, group_name, description, status, uptime_90d, order_num, visible, created_at, updated_at`,
		tenantID, name, groupName, description,
	).Scan(&c.ID, &c.TenantID, &c.Name, &c.GroupName, &c.Description,
		&c.Status, &c.Uptime90d, &c.OrderNum, &c.Visible, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return StatusComponent{}, fmt.Errorf("db.CreateStatusComponent: %w", err)
	}
	return c, nil
}
