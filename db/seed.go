package db

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	DefaultSysadminEmail    = "tayyebimohammadreza@gmail.com"
	DefaultSysadminName     = "System Administrator"
	DefaultSysadminPassword = "sysadmin123"

	pbkdf2Iterations = 120000
	pbkdf2SaltBytes  = 16
	pbkdf2KeyBytes   = 32
)

// SysadminBootstrapCredentials returns the configured initial super-administrator
// identity from environment variables, falling back to safe local defaults.
func SysadminBootstrapCredentials() (email, name, password string) {
	email = strings.TrimSpace(os.Getenv("SYSADMIN_EMAIL"))
	if email == "" {
		email = DefaultSysadminEmail
	}
	name = strings.TrimSpace(os.Getenv("SYSADMIN_NAME"))
	if name == "" {
		name = DefaultSysadminName
	}
	password = os.Getenv("SYSADMIN_PASSWORD")
	if password == "" {
		password = DefaultSysadminPassword
	}
	return email, name, password
}

// Seed runs idempotent production-ready bootstrap seeders ONLY:
//  1. The System Super Administrator account (so an operator can sign in and
//     onboard infrastructure, IP pools, and tenants).
//  2. Canonical background job queue definitions (with zero fake jobs/workers).
//
// All demo/dummy tenants, fake logs, fake replicas, fake workers, and fake
// billing records have been removed.
func Seed(ctx context.Context, admin *pgxpool.Pool) error {
	if err := seedSuperAdmin(ctx, admin); err != nil {
		return err
	}
	if err := seedJobQueues(ctx, admin); err != nil {
		return err
	}
	return nil
}

func seedSuperAdmin(ctx context.Context, admin *pgxpool.Pool) error {
	email, name, password := SysadminBootstrapCredentials()
	pwHash, err := HashPassword(password)
	if err != nil {
		return fmt.Errorf("db: hash sysadmin password: %w", err)
	}

	// Insert super-admin if missing; if present with NULL/empty password_hash
	// (e.g. upgraded from earlier schema), backfill password_hash without
	// overwriting an operator-changed password.
	if _, err := admin.Exec(ctx, `
		INSERT INTO users (email, name, password_hash, global_role)
		VALUES ($1, $2, $3, 'sysadmin')
		ON CONFLICT (email) DO UPDATE SET
		    global_role   = 'sysadmin',
		    name          = COALESCE(NULLIF(users.name, ''), EXCLUDED.name),
		    password_hash = CASE
		        WHEN users.password_hash IS NULL OR users.password_hash = '' THEN EXCLUDED.password_hash
		        ELSE users.password_hash
		    END`,
		email, name, pwHash,
	); err != nil {
		return fmt.Errorf("db: seed sysadmin user: %w", err)
	}
	return nil
}

func seedJobQueues(ctx context.Context, admin *pgxpool.Pool) error {
	queues := []struct {
		name, desc string
		conc       int
	}{
		{"default", "General-purpose platform operations", 4},
		{"critical", "High-priority control-plane operations", 3},
		{"provisioning", "Infrastructure & host virtualization auto-provisioning", 2},
		{"background", "Long-running maintenance tasks", 2},
		{"mailer", "Transactional email delivery queue", 2},
	}
	for _, q := range queues {
		if _, err := admin.Exec(ctx, `
			INSERT INTO background_job_queues (name, description, concurrency)
			VALUES ($1, $2, $3)
			ON CONFLICT (name) DO NOTHING`,
			q.name, q.desc, q.conc,
		); err != nil {
			return fmt.Errorf("db: seed job queue %s: %w", q.name, err)
		}
	}
	return nil
}

// ── Password Hashing & Authentication (PBKDF2-HMAC-SHA256) ───────────────────

// HashPassword produces a self-describing "pbkdf2_sha256$<iter>$<saltHex>$<dkHex>" string.
func HashPassword(password string) (string, error) {
	salt := make([]byte, pbkdf2SaltBytes)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	dk := pbkdf2SHA256([]byte(password), salt, pbkdf2Iterations, pbkdf2KeyBytes)
	return fmt.Sprintf("pbkdf2_sha256$%d$%s$%s",
		pbkdf2Iterations, hex.EncodeToString(salt), hex.EncodeToString(dk)), nil
}

// VerifyPassword performs a constant-time check against a HashPassword string.
func VerifyPassword(encodedHash, candidate string) bool {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 4 || parts[0] != "pbkdf2_sha256" {
		return false
	}
	iter, err := strconv.Atoi(parts[1])
	if err != nil || iter < 1000 {
		return false
	}
	salt, err := hex.DecodeString(parts[2])
	if err != nil || len(salt) == 0 {
		return false
	}
	expected, err := hex.DecodeString(parts[3])
	if err != nil || len(expected) == 0 {
		return false
	}
	actual := pbkdf2SHA256([]byte(candidate), salt, iter, len(expected))
	return subtle.ConstantTimeCompare(expected, actual) == 1
}

// pbkdf2SHA256 implements RFC 2898 / PKCS #5 v2.0 PBKDF2 with HMAC-SHA256
// using only the Go standard library.
func pbkdf2SHA256(password, salt []byte, iter, keyLen int) []byte {
	prf := func(key, data []byte) []byte {
		mac := hmac.New(sha256.New, key)
		mac.Write(data)
		return mac.Sum(nil)
	}
	const hLen = sha256.Size
	numBlocks := (keyLen + hLen - 1) / hLen
	var dk []byte
	var blockBuf [4]byte
	for block := 1; block <= numBlocks; block++ {
		blockBuf[0] = byte(block >> 24)
		blockBuf[1] = byte(block >> 16)
		blockBuf[2] = byte(block >> 8)
		blockBuf[3] = byte(block)

		in := make([]byte, 0, len(salt)+4)
		in = append(in, salt...)
		in = append(in, blockBuf[:]...)
		u := prf(password, in)
		t := make([]byte, len(u))
		copy(t, u)
		for i := 1; i < iter; i++ {
			u = prf(password, u)
			for j := range t {
				t[j] ^= u[j]
			}
		}
		dk = append(dk, t...)
	}
	return dk[:keyLen]
}

// AuthenticatedUser represents a verified user row resolved during login.
type AuthenticatedUser struct {
	ID         int64
	Email      string
	Name       string
	AvatarURL  string
	GlobalRole string // "sysadmin" | "user"
	TenantSlug string // populated if member of a tenant
	TenantRole string // "tenant" | "user"
}

// AuthenticateUser verifies email + password against PostgreSQL `users` and
// resolves primary tenant membership if any.
func AuthenticateUser(ctx context.Context, admin *pgxpool.Pool, email, password string) (AuthenticatedUser, bool, error) {
	var u AuthenticatedUser
	var pwHash *string
	err := admin.QueryRow(ctx, `
		SELECT u.id, u.email, COALESCE(u.name, ''), COALESCE(u.avatar_url, ''),
		       u.global_role, u.password_hash,
		       COALESCE(t.slug, ''), COALESCE(utm.role, '')
		FROM users u
		LEFT JOIN LATERAL (
			SELECT tenant_id, role
			FROM user_tenant_memberships
			WHERE user_id = u.id
			ORDER BY created_at ASC
			LIMIT 1
		) utm ON true
		LEFT JOIN tenants t ON t.id = utm.tenant_id AND t.deleted_at IS NULL
		WHERE u.email = $1 AND u.deleted_at IS NULL`,
		strings.TrimSpace(email),
	).Scan(&u.ID, &u.Email, &u.Name, &u.AvatarURL, &u.GlobalRole, &pwHash, &u.TenantSlug, &u.TenantRole)
	if errors.Is(err, pgx.ErrNoRows) {
		return AuthenticatedUser{}, false, nil
	}
	if err != nil {
		return AuthenticatedUser{}, false, fmt.Errorf("db.AuthenticateUser: %w", err)
	}
	if pwHash == nil || *pwHash == "" || !VerifyPassword(*pwHash, password) {
		return AuthenticatedUser{}, false, nil
	}
	_, _ = admin.Exec(ctx, `UPDATE users SET last_login_at = now() WHERE id = $1`, u.ID)
	return u, true, nil
}

// ChangeUserPassword verifies the user's current password in PostgreSQL and
// updates password_hash atomically.
func ChangeUserPassword(ctx context.Context, admin *pgxpool.Pool, userID int64, currentPassword, newPassword string) (bool, error) {
	var pwHash *string
	err := admin.QueryRow(ctx, `
		SELECT password_hash FROM users WHERE id = $1 AND deleted_at IS NULL`,
		userID).Scan(&pwHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if pwHash == nil || !VerifyPassword(*pwHash, currentPassword) {
		return false, nil
	}
	newHash, err := HashPassword(newPassword)
	if err != nil {
		return false, err
	}
	_, err = admin.Exec(ctx, `
		UPDATE users SET password_hash = $2, updated_at = now() WHERE id = $1`,
		userID, newHash)
	if err != nil {
		return false, err
	}
	return true, nil
}

// SetUserPasswordByEmail sets or resets a user's password hash directly (used by sysadmin owner reset).
func SetUserPasswordByEmail(ctx context.Context, admin *pgxpool.Pool, email, name, globalRole, plainPassword string) (int64, error) {
	pwHash, err := HashPassword(plainPassword)
	if err != nil {
		return 0, err
	}
	if globalRole != "sysadmin" {
		globalRole = "user"
	}
	var id int64
	err = admin.QueryRow(ctx, `
		INSERT INTO users (email, name, password_hash, global_role)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (email) DO UPDATE SET
		    name          = COALESCE(NULLIF($2, ''), users.name),
		    password_hash = EXCLUDED.password_hash,
		    updated_at    = now()
		RETURNING id`,
		strings.TrimSpace(email), strings.TrimSpace(name), pwHash, globalRole,
	).Scan(&id)
	return id, err
}

// LookupUserIDByEmail resolves a user ID by email.
func LookupUserIDByEmail(ctx context.Context, admin *pgxpool.Pool, email string) (int64, error) {
	var id int64
	err := admin.QueryRow(ctx,
		`SELECT id FROM users WHERE email = $1 AND deleted_at IS NULL`,
		email).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// UpsertUser ensures a users row exists for the given email; returns the user id.
func UpsertUser(ctx context.Context, admin *pgxpool.Pool, email, name, globalRole string) (int64, error) {
	if globalRole != "user" && globalRole != "sysadmin" {
		globalRole = "user"
	}
	var id int64
	err := admin.QueryRow(ctx, `
		INSERT INTO users (email, name, global_role)
		VALUES ($1, $2, $3)
		ON CONFLICT (email) DO UPDATE SET
		    name        = EXCLUDED.name,
		    global_role = EXCLUDED.global_role,
		    updated_at  = now()
		RETURNING id`,
		email, name, globalRole).Scan(&id)
	return id, err
}
