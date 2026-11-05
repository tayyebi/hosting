package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserPreferences mirrors user_preferences table.
type UserPreferences struct {
	UserID             int64
	Theme              string
	Timezone           string
	Locale             string
	EmailNotifications bool
	UpdatedAt          time.Time
}

// UserSecurity holds 2FA and recent-session data for the settings page.
type UserSecurity struct {
	MFAEnabled  bool
	MFAMethod   string // 'totp','webauthn','sms' or ''
	LastLoginAt *time.Time
}

// GetUserPreferences returns the preferences row for a user, creating defaults if absent.
func GetUserPreferences(ctx context.Context, pool *pgxpool.Pool, userID int64) (UserPreferences, error) {
	var p UserPreferences
	err := pool.QueryRow(ctx, `
		SELECT user_id, theme, timezone, locale, email_notifications, updated_at
		FROM user_preferences WHERE user_id=$1`, userID,
	).Scan(&p.UserID, &p.Theme, &p.Timezone, &p.Locale, &p.EmailNotifications, &p.UpdatedAt)
	if err == pgx.ErrNoRows {
		p = UserPreferences{
			UserID:             userID,
			Theme:              "system",
			Timezone:           "UTC",
			Locale:             "en-US",
			EmailNotifications: true,
		}
		_, seedErr := pool.Exec(ctx, `
			INSERT INTO user_preferences (user_id, theme, timezone, locale, email_notifications)
			VALUES ($1,$2,$3,$4,$5) ON CONFLICT (user_id) DO NOTHING`,
			userID, p.Theme, p.Timezone, p.Locale, p.EmailNotifications,
		)
		if seedErr != nil {
			return p, fmt.Errorf("db.GetUserPreferences seed: %w", seedErr)
		}
		return p, nil
	}
	if err != nil {
		return p, fmt.Errorf("db.GetUserPreferences: %w", err)
	}
	return p, nil
}

// UpdateUserPreferences persists preference changes.
func UpdateUserPreferences(ctx context.Context, pool *pgxpool.Pool, p UserPreferences) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO user_preferences (user_id, theme, timezone, locale, email_notifications, updated_at)
		VALUES ($1,$2,$3,$4,$5,now())
		ON CONFLICT (user_id) DO UPDATE SET
			theme=$2, timezone=$3, locale=$4,
			email_notifications=$5, updated_at=now()`,
		p.UserID, p.Theme, p.Timezone, p.Locale, p.EmailNotifications,
	)
	if err != nil {
		return fmt.Errorf("db.UpdateUserPreferences: %w", err)
	}
	return nil
}

// GetUserSecurity returns MFA status and last-login time.
func GetUserSecurity(ctx context.Context, pool *pgxpool.Pool, userID int64) (UserSecurity, error) {
	var s UserSecurity
	// 2FA info
	var method *string
	err := pool.QueryRow(ctx, `
		SELECT enabled, method FROM user_security_2fa WHERE user_id=$1`, userID,
	).Scan(&s.MFAEnabled, &method)
	if err != nil && err != pgx.ErrNoRows {
		return s, fmt.Errorf("db.GetUserSecurity 2fa: %w", err)
	}
	if method != nil {
		s.MFAMethod = *method
	}

	// last login
	var lastLogin *time.Time
	if err2 := pool.QueryRow(ctx, `SELECT last_login_at FROM users WHERE id=$1`, userID).Scan(&lastLogin); err2 == nil {
		s.LastLoginAt = lastLogin
	}
	return s, nil
}

// ListUserPersonalAPIKeys returns active personal API keys for a user.
type PersonalAPIKey struct {
	ID         int64
	Name       string
	LastUsedAt *time.Time
	CreatedAt  time.Time
}

func ListUserPersonalAPIKeys(ctx context.Context, pool *pgxpool.Pool, userID int64) ([]PersonalAPIKey, error) {
	rows, err := pool.Query(ctx, `
		SELECT id, name, last_used_at, created_at
		FROM user_personal_api_keys
		WHERE user_id=$1 AND revoked_at IS NULL
		ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("db.ListUserPersonalAPIKeys: %w", err)
	}
	defer rows.Close()
	var out []PersonalAPIKey
	for rows.Next() {
		var k PersonalAPIKey
		if err := rows.Scan(&k.ID, &k.Name, &k.LastUsedAt, &k.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, k)
	}
	return out, rows.Err()
}

// CountPersonalAPIKeys returns the number of active (non-revoked) personal API keys for a user.
func CountPersonalAPIKeys(ctx context.Context, pool *pgxpool.Pool, userID int64) (int, error) {
	var n int
	err := pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM user_personal_api_keys
		WHERE user_id=$1 AND revoked_at IS NULL`, userID).Scan(&n)
	return n, err
}

// CreatePersonalAPIKey creates a new personal API key for a user.
func CreatePersonalAPIKey(ctx context.Context, pool *pgxpool.Pool, userID int64, name, keyHash string) (PersonalAPIKey, error) {
	var k PersonalAPIKey
	err := pool.QueryRow(ctx, `
		INSERT INTO user_personal_api_keys (user_id, name, key_hash)
		VALUES ($1,$2,$3)
		RETURNING id, name, last_used_at, created_at`,
		userID, name, keyHash,
	).Scan(&k.ID, &k.Name, &k.LastUsedAt, &k.CreatedAt)
	if err != nil {
		return k, fmt.Errorf("db.CreatePersonalAPIKey: %w", err)
	}
	return k, nil
}

// PersonalAPIKeyNameExists returns true if the user already has an active key with the given name.
func PersonalAPIKeyNameExists(ctx context.Context, pool *pgxpool.Pool, userID int64, name string) (bool, error) {
	var exists bool
	err := pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM user_personal_api_keys
			WHERE user_id=$1 AND name=$2 AND revoked_at IS NULL
		)`, userID, name).Scan(&exists)
	return exists, err
}

// RevokePersonalAPIKey marks a personal API key as revoked.
func RevokePersonalAPIKey(ctx context.Context, pool *pgxpool.Pool, userID, keyID int64) error {
	_, err := pool.Exec(ctx, `
		UPDATE user_personal_api_keys SET revoked_at=now()
		WHERE id=$1 AND user_id=$2 AND revoked_at IS NULL`, keyID, userID)
	return err
}
