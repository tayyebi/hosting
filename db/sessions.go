package db

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SessionRecord is what the session layer needs back from the DB on
// every authenticated request.
type SessionRecord struct {
	UserID     int64
	TenantID   int64 // 0 when the session is a sysadmin session
	Email      string
	Name       string
	GlobalRole string // "sysadmin" | "user"
	TenantSlug string // empty when sysadmin
	TenantRole string // "tenant" | "user" | ""
	ExpiresAt  time.Time
}

// HashSessionToken is the canonical token → token_hash function. We
// store only the hash in the DB so a leaked audit_log or backup can't
// be used to hijack sessions.
func HashSessionToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// CreateSession persists a new session. Uses the admin pool (BYPASSRLS)
// because sessions aren't strictly tenant-isolated — a sysadmin session
// has NULL tenant_id.
func CreateSession(ctx context.Context, admin *pgxpool.Pool, userID int64, tenantID int64, token string, ttl time.Duration, ipAddr, userAgent string) error {
	tokHash := HashSessionToken(token)
	var tid any = tenantID
	if tenantID <= 0 {
		tid = nil
	}
	var ip any = ipAddr
	if ipAddr == "" {
		ip = nil
	}
	_, err := admin.Exec(ctx, `
        INSERT INTO sessions (user_id, tenant_id, token_hash, ip_address, user_agent, expires_at)
        VALUES ($1, $2, $3, $4, $5, now() + $6::interval)`,
		userID, tid, tokHash, ip, userAgent, ttl.String())
	return err
}

// LookupSession resolves a cookie value to the user/tenant record it
// represents. Returns pgx.ErrNoRows for unknown / expired / revoked
// sessions.
func LookupSession(ctx context.Context, admin *pgxpool.Pool, token string) (*SessionRecord, error) {
	tokHash := HashSessionToken(token)
	row := admin.QueryRow(ctx, `
        SELECT s.user_id,
               COALESCE(s.tenant_id, 0),
               u.email,
               COALESCE(u.name, ''),
               u.global_role,
               COALESCE(t.slug, ''),
               COALESCE(utm.role, ''),
               s.expires_at
          FROM sessions s
          JOIN users u ON u.id = s.user_id AND u.deleted_at IS NULL
          LEFT JOIN tenants t ON t.id = s.tenant_id AND t.deleted_at IS NULL
          LEFT JOIN user_tenant_memberships utm
                 ON utm.user_id = u.id
                AND utm.tenant_id = s.tenant_id
                AND COALESCE(utm.active, true) = true
         WHERE s.token_hash = $1
           AND s.revoked_at IS NULL
           AND s.expires_at > now()
         ORDER BY s.created_at DESC
         LIMIT 1`, tokHash)

	var r SessionRecord
	if err := row.Scan(&r.UserID, &r.TenantID, &r.Email, &r.Name,
		&r.GlobalRole, &r.TenantSlug, &r.TenantRole, &r.ExpiresAt); err != nil {
		return nil, err
	}
	return &r, nil
}

// RevokeSession marks the session as revoked. Used on /logout.
func RevokeSession(ctx context.Context, admin *pgxpool.Pool, token string) error {
	_, err := admin.Exec(ctx,
		`UPDATE sessions SET revoked_at = now()
          WHERE token_hash = $1 AND revoked_at IS NULL`,
		HashSessionToken(token))
	return err
}

// RevokeOtherSessions revokes all active sessions for userID except the
// session identified by keepToken. Called after a password change so
// other devices are forced to re-authenticate.
func RevokeOtherSessions(ctx context.Context, admin *pgxpool.Pool, userID int64, keepToken string) error {
	keepHash := HashSessionToken(keepToken)
	_, err := admin.Exec(ctx, `
        UPDATE sessions
           SET revoked_at = now()
         WHERE user_id    = $1
           AND token_hash != $2
           AND revoked_at IS NULL`,
		userID, keepHash)
	return err
}

// UserSession is a row from the sessions table returned for the "active sessions" UI.
type UserSession struct {
	ID        int64
	TokenHash string // first 8 chars shown as an opaque handle
	IPAddress string
	UserAgent string
	CreatedAt time.Time
	ExpiresAt time.Time
	IsCurrent bool
}

// ListUserSessions returns all active (non-revoked, non-expired) sessions for
// the given user. currentTokenHash is used to mark the caller's own session.
func ListUserSessions(ctx context.Context, admin *pgxpool.Pool, userID int64, currentTokenHash string) ([]UserSession, error) {
	rows, err := admin.Query(ctx, `
		SELECT id, token_hash,
		       COALESCE(ip_address::text, ''),
		       COALESCE(user_agent, ''),
		       created_at, expires_at
		  FROM sessions
		 WHERE user_id    = $1
		   AND revoked_at IS NULL
		   AND expires_at > now()
		 ORDER BY created_at DESC
		 LIMIT 20`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []UserSession
	for rows.Next() {
		var s UserSession
		if err := rows.Scan(&s.ID, &s.TokenHash, &s.IPAddress, &s.UserAgent, &s.CreatedAt, &s.ExpiresAt); err != nil {
			return nil, err
		}
		s.IsCurrent = s.TokenHash == currentTokenHash
		out = append(out, s)
	}
	return out, rows.Err()
}

// RevokeSessionByID revokes the session with the given ID, only if it belongs
// to userID (prevents one user revoking another's sessions).
func RevokeSessionByID(ctx context.Context, admin *pgxpool.Pool, userID, sessionID int64) error {
	_, err := admin.Exec(ctx, `
		UPDATE sessions SET revoked_at = now()
		 WHERE id = $1 AND user_id = $2 AND revoked_at IS NULL`,
		sessionID, userID)
	return err
}

// RevokeAllUserSessions revokes every active session for userID, including
// the current one. Used by the "sign out of all devices" UI action.
func RevokeAllUserSessions(ctx context.Context, admin *pgxpool.Pool, userID int64) error {
	_, err := admin.Exec(ctx, `
		UPDATE sessions SET revoked_at = now()
		 WHERE user_id = $1 AND revoked_at IS NULL`, userID)
	return err
}

// PruneExpiredSessions deletes sessions that expired more than 7 days ago or
// were revoked more than 7 days ago. Keeps the table from growing unboundedly.
func PruneExpiredSessions(ctx context.Context, admin *pgxpool.Pool) error {
	_, err := admin.Exec(ctx, `
		DELETE FROM sessions
		WHERE expires_at < now() - interval '7 days'
		   OR (revoked_at IS NOT NULL AND revoked_at < now() - interval '7 days')`)
	return err
}

// ErrSessionNotFound is returned by LookupSession when no live session
// matches. Re-exports pgx.ErrNoRows for ergonomic comparisons.
var ErrSessionNotFound = pgx.ErrNoRows

// Ensure errors package is imported for consumers that want errors.Is.
var _ = errors.Is
