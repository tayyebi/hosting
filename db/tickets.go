package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Ticket is one row from the tickets table.
type Ticket struct {
	ID          int64
	TenantID    int64
	Subject     string
	ClientEmail string
	Status      string
	Priority    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ListTickets returns non-deleted tickets for a tenant with optional filters.
func ListTickets(ctx context.Context, pool *pgxpool.Pool, tenantID int64, search, status, priority string) ([]Ticket, error) {
	args := []any{tenantID}
	conds := []string{"tenant_id = $1", "deleted_at IS NULL"}

	if search != "" {
		args = append(args, "%"+strings.ToLower(search)+"%")
		n := len(args)
		conds = append(conds, fmt.Sprintf(
			"(lower(subject) LIKE $%d OR lower(client_email) LIKE $%d)", n, n))
	}
	if status != "" {
		args = append(args, status)
		conds = append(conds, fmt.Sprintf("status = $%d", len(args)))
	}
	if priority != "" {
		args = append(args, priority)
		conds = append(conds, fmt.Sprintf("priority = $%d", len(args)))
	}

	q := fmt.Sprintf(`
		SELECT id, tenant_id, subject, client_email, status, priority, created_at, updated_at
		FROM tickets
		WHERE %s
		ORDER BY created_at DESC`, strings.Join(conds, " AND "))

	rows, err := pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("db.ListTickets: %w", err)
	}
	defer rows.Close()

	var out []Ticket
	for rows.Next() {
		var t Ticket
		if err := rows.Scan(&t.ID, &t.TenantID, &t.Subject, &t.ClientEmail,
			&t.Status, &t.Priority, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("db.ListTickets scan: %w", err)
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// CreateTicket inserts a new ticket and returns the created row.
func CreateTicket(ctx context.Context, pool *pgxpool.Pool, tenantID int64, subject, clientEmail, priority string) (Ticket, error) {
	var t Ticket
	err := pool.QueryRow(ctx, `
		INSERT INTO tickets (tenant_id, subject, client_email, priority)
		VALUES ($1, $2, $3, $4)
		RETURNING id, tenant_id, subject, client_email, status, priority, created_at, updated_at`,
		tenantID, subject, clientEmail, priority,
	).Scan(&t.ID, &t.TenantID, &t.Subject, &t.ClientEmail,
		&t.Status, &t.Priority, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return Ticket{}, fmt.Errorf("db.CreateTicket: %w", err)
	}
	return t, nil
}

// DeleteTicket soft-deletes a ticket by setting deleted_at.
func DeleteTicket(ctx context.Context, pool *pgxpool.Pool, tenantID, ticketID int64) error {
	_, err := pool.Exec(ctx, `
		UPDATE tickets
		SET deleted_at=now(), updated_at=now()
		WHERE tenant_id=$1 AND id=$2 AND deleted_at IS NULL`,
		tenantID, ticketID)
	if err != nil {
		return fmt.Errorf("db.DeleteTicket: %w", err)
	}
	return nil
}

// TicketCount returns the total number of non-deleted tickets for a tenant.
func TicketCount(ctx context.Context, pool *pgxpool.Pool, tenantID int64) (int, error) {
	var count int
	err := pool.QueryRow(ctx,
		`SELECT count(*) FROM tickets WHERE tenant_id=$1 AND deleted_at IS NULL`,
		tenantID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("db.TicketCount: %w", err)
	}
	return count, nil
}

// OpenTicketCount returns the number of tickets in 'open' or 'in-progress' state.
func OpenTicketCount(ctx context.Context, pool *pgxpool.Pool, tenantID int64) (int, error) {
	var count int
	err := pool.QueryRow(ctx,
		`SELECT count(*) FROM tickets
		 WHERE tenant_id=$1 AND status IN ('open','in-progress') AND deleted_at IS NULL`,
		tenantID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("db.OpenTicketCount: %w", err)
	}
	return count, nil
}
