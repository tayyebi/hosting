package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TenantTransaction struct {
	ID          int64
	TenantID    int64
	Reference   string
	Client      string
	ClientCo    string
	Description string
	TxnType     string
	Amount      float64
	Currency    string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func ListTenantTransactions(ctx context.Context, pool *pgxpool.Pool, tenantID int64, search, txnType, status string) ([]TenantTransaction, error) {
	args := []any{tenantID}
	conds := []string{"tenant_id = $1", "deleted_at IS NULL"}

	if search != "" {
		args = append(args, "%"+strings.ToLower(search)+"%")
		n := len(args)
		conds = append(conds, fmt.Sprintf(
			"(lower(reference) LIKE $%d OR lower(client) LIKE $%d OR lower(description) LIKE $%d)", n, n, n))
	}
	if txnType != "" {
		args = append(args, txnType)
		conds = append(conds, fmt.Sprintf("txn_type = $%d", len(args)))
	}
	if status != "" {
		args = append(args, status)
		conds = append(conds, fmt.Sprintf("status = $%d", len(args)))
	}

	q := fmt.Sprintf(`
		SELECT id, tenant_id, reference, client, client_co, description, txn_type, amount, currency, status, created_at, updated_at
		FROM tenant_transactions
		WHERE %s
		ORDER BY created_at DESC`, strings.Join(conds, " AND "))

	rows, err := pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("db.ListTenantTransactions: %w", err)
	}
	defer rows.Close()

	var out []TenantTransaction
	for rows.Next() {
		var t TenantTransaction
		if err := rows.Scan(&t.ID, &t.TenantID, &t.Reference, &t.Client, &t.ClientCo,
			&t.Description, &t.TxnType, &t.Amount, &t.Currency, &t.Status,
			&t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("db.ListTenantTransactions scan: %w", err)
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func UpdateTenantTransactionStatus(ctx context.Context, pool *pgxpool.Pool, tenantID, id int64, status string) error {
	_, err := pool.Exec(ctx, `
		UPDATE tenant_transactions SET status=$3, updated_at=now()
		WHERE tenant_id=$1 AND id=$2 AND deleted_at IS NULL`,
		tenantID, id, status)
	if err != nil {
		return fmt.Errorf("db.UpdateTenantTransactionStatus: %w", err)
	}
	return nil
}

func DeleteTenantTransaction(ctx context.Context, pool *pgxpool.Pool, tenantID, id int64) error {
	_, err := pool.Exec(ctx, `
		UPDATE tenant_transactions SET deleted_at=now(), updated_at=now()
		WHERE tenant_id=$1 AND id=$2 AND deleted_at IS NULL`,
		tenantID, id)
	if err != nil {
		return fmt.Errorf("db.DeleteTenantTransaction: %w", err)
	}
	return nil
}

func CreateTenantTransaction(ctx context.Context, pool *pgxpool.Pool, tenantID int64, reference, client, clientCo, description, txnType, currency string, amount float64) (TenantTransaction, error) {
	var t TenantTransaction
	err := pool.QueryRow(ctx, `
		INSERT INTO tenant_transactions (tenant_id, reference, client, client_co, description, txn_type, amount, currency)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, tenant_id, reference, client, client_co, description, txn_type, amount, currency, status, created_at, updated_at`,
		tenantID, reference, client, clientCo, description, txnType, amount, currency,
	).Scan(&t.ID, &t.TenantID, &t.Reference, &t.Client, &t.ClientCo, &t.Description,
		&t.TxnType, &t.Amount, &t.Currency, &t.Status, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return TenantTransaction{}, fmt.Errorf("db.CreateTenantTransaction: %w", err)
	}
	return t, nil
}
