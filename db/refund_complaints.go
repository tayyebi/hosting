package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RefundComplaint struct {
	ID        int64
	TenantID  int64
	Client    string
	Amount    string
	Currency  string
	Reason    string
	Status    string
	InvoiceID string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func ListRefundComplaints(ctx context.Context, pool *pgxpool.Pool, tenantID int64, search, status string) ([]RefundComplaint, error) {
	args := []any{tenantID}
	conds := []string{"tenant_id = $1", "deleted_at IS NULL"}

	if search != "" {
		args = append(args, "%"+strings.ToLower(search)+"%")
		n := len(args)
		conds = append(conds, fmt.Sprintf(
			"(lower(client) LIKE $%d OR lower(reason) LIKE $%d OR lower(invoice_id) LIKE $%d)", n, n, n))
	}
	if status != "" {
		args = append(args, status)
		conds = append(conds, fmt.Sprintf("status = $%d", len(args)))
	}

	q := fmt.Sprintf(`
		SELECT id, tenant_id, client, amount::text, currency, reason, status, invoice_id, created_at, updated_at
		FROM refund_complaints
		WHERE %s
		ORDER BY created_at DESC`, strings.Join(conds, " AND "))

	rows, err := pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("db.ListRefundComplaints: %w", err)
	}
	defer rows.Close()

	var out []RefundComplaint
	for rows.Next() {
		var r RefundComplaint
		if err := rows.Scan(&r.ID, &r.TenantID, &r.Client, &r.Amount, &r.Currency,
			&r.Reason, &r.Status, &r.InvoiceID, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, fmt.Errorf("db.ListRefundComplaints scan: %w", err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func UpdateRefundComplaintStatus(ctx context.Context, pool *pgxpool.Pool, tenantID, id int64, status string) error {
	_, err := pool.Exec(ctx, `
		UPDATE refund_complaints SET status=$3, updated_at=now()
		WHERE tenant_id=$1 AND id=$2 AND deleted_at IS NULL`,
		tenantID, id, status)
	if err != nil {
		return fmt.Errorf("db.UpdateRefundComplaintStatus: %w", err)
	}
	return nil
}

func DeleteRefundComplaint(ctx context.Context, pool *pgxpool.Pool, tenantID, id int64) error {
	_, err := pool.Exec(ctx, `
		UPDATE refund_complaints SET deleted_at=now(), updated_at=now()
		WHERE tenant_id=$1 AND id=$2 AND deleted_at IS NULL`,
		tenantID, id)
	if err != nil {
		return fmt.Errorf("db.DeleteRefundComplaint: %w", err)
	}
	return nil
}

func CreateRefundComplaint(ctx context.Context, pool *pgxpool.Pool, tenantID int64, client, amount, currency, reason, invoiceID string) (RefundComplaint, error) {
	var r RefundComplaint
	err := pool.QueryRow(ctx, `
		INSERT INTO refund_complaints (tenant_id, client, amount, currency, reason, invoice_id)
		VALUES ($1, $2, $3::numeric, $4, $5, $6)
		RETURNING id, tenant_id, client, amount::text, currency, reason, status, invoice_id, created_at, updated_at`,
		tenantID, client, amount, currency, reason, invoiceID,
	).Scan(&r.ID, &r.TenantID, &r.Client, &r.Amount, &r.Currency,
		&r.Reason, &r.Status, &r.InvoiceID, &r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		return RefundComplaint{}, fmt.Errorf("db.CreateRefundComplaint: %w", err)
	}
	return r, nil
}
