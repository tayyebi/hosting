package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RegisterUser creates (or updates) a user row with an already-hashed password and
// links the user to the given tenant as a 'tenant' member. Returns the user ID.
// Uses INSERT … ON CONFLICT so the call is idempotent for email re-use: the
// password_hash and name are updated if the account already exists.
func RegisterUser(ctx context.Context, admin *pgxpool.Pool, email, name, passwordHash string, tenantID int64) (int64, error) {
	var userID int64
	err := admin.QueryRow(ctx, `
		INSERT INTO users (email, name, password_hash, global_role)
		VALUES ($1, $2, $3, 'user')
		ON CONFLICT (email) DO UPDATE SET
		    name          = EXCLUDED.name,
		    password_hash = EXCLUDED.password_hash,
		    updated_at    = now()
		RETURNING id`,
		email, name, passwordHash,
	).Scan(&userID)
	if err != nil {
		return 0, fmt.Errorf("db.RegisterUser insert user: %w", err)
	}

	if tenantID > 0 {
		_, err = admin.Exec(ctx, `
			INSERT INTO user_tenant_memberships (user_id, tenant_id, role)
			VALUES ($1, $2, 'tenant')
			ON CONFLICT (user_id, tenant_id) DO NOTHING`,
			userID, tenantID)
		if err != nil {
			return userID, fmt.Errorf("db.RegisterUser membership: %w", err)
		}
	}

	return userID, nil
}

// IsEmailRegistered returns true when the email already exists and belongs to
// the given tenant.
func IsEmailRegistered(ctx context.Context, admin *pgxpool.Pool, email string, tenantID int64) (bool, error) {
	var count int
	err := admin.QueryRow(ctx, `
		SELECT count(*) FROM users u
		JOIN user_tenant_memberships m ON m.user_id = u.id
		WHERE u.email = $1 AND m.tenant_id = $2 AND u.deleted_at IS NULL`,
		email, tenantID,
	).Scan(&count)
	if err != nil && err != pgx.ErrNoRows {
		return false, fmt.Errorf("db.IsEmailRegistered: %w", err)
	}
	return count > 0, nil
}
