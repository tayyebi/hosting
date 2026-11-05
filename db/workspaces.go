package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// UserWorkspace is a workspace the user is a member of.
type UserWorkspace struct {
	ID          int64
	Slug        string
	Name        string
	Description string
	TenantSlug  string
	Role        string
	CreatedAt   time.Time
}

// UserInvitation is a pending workspace invitation for the user.
type UserInvitation struct {
	ID           int64
	TenantSlug   string
	InviterEmail string
	Role         string
	ExpiresAt    time.Time
	CreatedAt    time.Time
}

// ListUserWorkspaces returns all workspaces the user is a member of, joined with tenant slug.
func ListUserWorkspaces(ctx context.Context, pool *pgxpool.Pool, userID int64) ([]UserWorkspace, error) {
	rows, err := pool.Query(ctx, `
		SELECT w.id, w.slug, w.name, COALESCE(w.description,''),
		       t.slug, wm.role, w.created_at
		FROM workspace_memberships wm
		JOIN workspaces w ON w.id = wm.workspace_id
		JOIN tenants t ON t.id = w.tenant_id
		WHERE wm.user_id = $1
		  AND w.deleted_at IS NULL
		ORDER BY w.created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("db.ListUserWorkspaces: %w", err)
	}
	defer rows.Close()
	var out []UserWorkspace
	for rows.Next() {
		var w UserWorkspace
		if err := rows.Scan(&w.ID, &w.Slug, &w.Name, &w.Description,
			&w.TenantSlug, &w.Role, &w.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

// ListUserInvitations returns pending (unaccepted) workspace invitations for the user's email.
func ListUserInvitations(ctx context.Context, pool *pgxpool.Pool, email string) ([]UserInvitation, error) {
	rows, err := pool.Query(ctx, `
		SELECT i.id, t.slug, u.email, i.role, i.expires_at, i.created_at
		FROM user_workspace_invitations i
		JOIN tenants t ON t.id = i.tenant_id
		LEFT JOIN users u ON u.id = i.inviter_id
		WHERE i.invitee_email = $1
		  AND i.accepted_at IS NULL
		  AND i.expires_at > now()
		ORDER BY i.created_at DESC`, email)
	if err != nil {
		return nil, fmt.Errorf("db.ListUserInvitations: %w", err)
	}
	defer rows.Close()
	var out []UserInvitation
	for rows.Next() {
		var inv UserInvitation
		var inviterEmail *string
		if err := rows.Scan(&inv.ID, &inv.TenantSlug, &inviterEmail,
			&inv.Role, &inv.ExpiresAt, &inv.CreatedAt); err != nil {
			return nil, err
		}
		if inviterEmail != nil {
			inv.InviterEmail = *inviterEmail
		}
		out = append(out, inv)
	}
	return out, rows.Err()
}

// CreateUserWorkspace inserts a new workspace for the given tenant and adds the
// creating user as owner. Requires the user already has a tenancy row.
func CreateUserWorkspace(ctx context.Context, pool *pgxpool.Pool, userID int64, name, description string) error {
	// Derive slug from name: lowercase, spaces→hyphens, strip non-alphanumeric.
	slug := strings.ToLower(name)
	slug = strings.NewReplacer(" ", "-", "_", "-").Replace(slug)
	var b strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
		}
	}
	slug = strings.Trim(b.String(), "-")
	if slug == "" {
		slug = fmt.Sprintf("ws-%d", userID)
	}

	var wsID int64
	err := pool.QueryRow(ctx, `
		WITH tenant AS (
			SELECT t.id FROM tenants t
			JOIN user_tenant_memberships utm ON utm.tenant_id = t.id
			WHERE utm.user_id = $1
			LIMIT 1
		)
		INSERT INTO workspaces (tenant_id, slug, name, description, created_at, updated_at)
		SELECT id, $2, $3, $4, now(), now() FROM tenant
		RETURNING id`,
		userID, slug, name, description,
	).Scan(&wsID)
	if err != nil {
		return fmt.Errorf("db.CreateUserWorkspace insert: %w", err)
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_memberships (workspace_id, user_id, role, created_at)
		VALUES ($1, $2, 'owner', now())`, wsID, userID)
	return err
}

// WorkspaceNameExistsForUser returns true if the user already has a live workspace with that name.
func WorkspaceNameExistsForUser(ctx context.Context, pool *pgxpool.Pool, userID int64, name string) (bool, error) {
	var exists bool
	err := pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM workspaces w
			JOIN workspace_memberships wm ON wm.workspace_id = w.id
			WHERE wm.user_id = $1
			  AND lower(w.name) = lower($2)
			  AND w.deleted_at IS NULL
		)`, userID, name).Scan(&exists)
	return exists, err
}

// CountUserWorkspaces returns the number of workspaces the user owns or is a member of.
func CountUserWorkspaces(ctx context.Context, pool *pgxpool.Pool, userID int64) (int, error) {
	var count int
	err := pool.QueryRow(ctx, `
		SELECT count(*) FROM workspace_memberships wm
		JOIN workspaces w ON w.id = wm.workspace_id
		WHERE wm.user_id = $1 AND w.deleted_at IS NULL`, userID,
	).Scan(&count)
	return count, err
}

// DeleteUserWorkspace soft-deletes the workspace identified by workspaceID,
// but only if the calling user holds the "owner" role in that workspace.
// Returns (true, nil) on success, (false, nil) if the workspace was not found
// or the user is not an owner.
func DeleteUserWorkspace(ctx context.Context, pool *pgxpool.Pool, userID, workspaceID int64) (bool, error) {
	tag, err := pool.Exec(ctx, `
		UPDATE workspaces
		   SET deleted_at = now()
		 WHERE id = $1
		   AND deleted_at IS NULL
		   AND EXISTS (
		       SELECT 1 FROM workspace_memberships
		        WHERE workspace_id = $1
		          AND user_id = $2
		          AND role = 'owner'
		   )`, workspaceID, userID)
	if err != nil {
		return false, fmt.Errorf("db.DeleteUserWorkspace: %w", err)
	}
	return tag.RowsAffected() == 1, nil
}

// PruneExpiredInvitations deletes workspace invitations that expired more than
// 7 days ago and were never accepted. Keeps user_workspace_invitations tidy.
func PruneExpiredInvitations(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		DELETE FROM user_workspace_invitations
		WHERE accepted_at IS NULL
		  AND expires_at < now() - interval '7 days'`)
	return err
}

// WorkspaceMemberCount returns the number of members in a workspace.
func WorkspaceMemberCount(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) (int, error) {
	var count int
	err := pool.QueryRow(ctx, `
		SELECT count(*) FROM workspace_memberships WHERE workspace_id=$1`, workspaceID,
	).Scan(&count)
	return count, err
}
