package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"platform/db"
)

// ── GET /api/user/settings ─────────────────────────────────

type apiUserProfile struct {
	Email     string `json:"email"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

type apiUserPreferences struct {
	Theme         string `json:"theme"`
	Language      string `json:"language"`
	Timezone      string `json:"timezone"`
	DateFormat    string `json:"date_format"`
	Notifications bool   `json:"notifications_enabled"`
}

type apiUserSecurity struct {
	MFAEnabled     bool   `json:"mfa_enabled"`
	LastLogin      string `json:"last_login"`
	SessionExpiry  string `json:"session_expiry"`
	PasswordChange string `json:"last_password_change"`
}

type apiUserSettingsResponse struct {
	Profile     apiUserProfile     `json:"profile"`
	Preferences apiUserPreferences `json:"preferences"`
	Security    apiUserSecurity    `json:"security"`
}

func apiUserSettingsHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok {
		apiError(w, http.StatusUnauthorized, "Authentication required.")
		return
	}

	resp := apiUserSettingsResponse{
		Profile: apiUserProfile{
			Email:     user.Email,
			Name:      user.Name,
			Role:      string(user.Role),
			AvatarURL: user.AvatarURL,
		},
		Preferences: apiUserPreferences{
			Theme: "system", Language: "en", Timezone: "UTC",
			DateFormat: "YYYY-MM-DD", Notifications: true,
		},
		Security: apiUserSecurity{
			MFAEnabled: false, LastLogin: "Just now",
			SessionExpiry: "8 hours", PasswordChange: "30 days ago",
		},
	}

	if Pools != nil && user.ID > 0 {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		if prefs, err := db.GetUserPreferences(ctx, Pools.App, user.ID); err == nil {
			resp.Preferences.Theme = prefs.Theme
			resp.Preferences.Timezone = prefs.Timezone
			resp.Preferences.Notifications = prefs.EmailNotifications
		}
		if sec, err := db.GetUserSecurity(ctx, Pools.App, user.ID); err == nil {
			resp.Security.MFAEnabled = sec.MFAEnabled
			if sec.LastLoginAt != nil {
				age := time.Since(*sec.LastLoginAt)
				switch {
				case age < time.Minute:
					resp.Security.LastLogin = "Just now"
				case age < time.Hour:
					resp.Security.LastLogin = formatMinutes(int(age.Minutes())) + " ago"
				default:
					resp.Security.LastLogin = formatHours(int(age.Hours())) + " ago"
				}
			}
		}
	}

	apiOK(w, resp)
}

func formatMinutes(m int) string {
	if m == 1 {
		return "1 minute"
	}
	return itoa(m) + " minutes"
}

func formatHours(h int) string {
	if h == 1 {
		return "1 hour"
	}
	return itoa(h) + " hours"
}


// ── POST /api/user/password ────────────────────────────────

// maxPasswordLength bounds submitted passwords so PBKDF2/bcrypt hashing can't
// be abused with megabytes of input.
const maxPasswordLength = 256

type apiChangePasswordRequest struct {
	Current string `json:"current"`
	New     string `json:"new"`
	Confirm string `json:"confirm"`
}

func apiChangePasswordHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok {
		apiError(w, http.StatusUnauthorized, "Authentication required.")
		return
	}

	var req apiChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		apiError(w, http.StatusBadRequest, "Invalid request body: expected a JSON object.")
		return
	}

	// Validate the payload before spending the rate-limit budget (and before
	// running the expensive password hash) so malformed input is free.
	if req.Current == "" {
		apiError(w, http.StatusBadRequest, "Current password is required.")
		return
	}
	if req.New == "" {
		apiError(w, http.StatusBadRequest, "New password is required.")
		return
	}
	if req.Confirm == "" {
		apiError(w, http.StatusBadRequest, "Please confirm the new password.")
		return
	}
	if len(req.New) < 8 {
		apiError(w, http.StatusBadRequest, "New password must be at least 8 characters.")
		return
	}
	if len(req.New) > maxPasswordLength {
		apiError(w, http.StatusBadRequest, "New password must be 256 characters or fewer.")
		return
	}
	if !hasUppercase(req.New) || !hasDigit(req.New) {
		apiError(w, http.StatusBadRequest, "New password must contain at least one uppercase letter and one number.")
		return
	}
	if req.New != req.Confirm {
		apiError(w, http.StatusBadRequest, "New password and confirmation do not match.")
		return
	}
	if req.New == req.Current {
		apiError(w, http.StatusBadRequest, "New password must differ from the current one.")
		return
	}

	if !passwordChangeRateLimiter.allow(fmt.Sprintf("%d", user.ID)) {
		w.Header().Set("Retry-After", "3600")
		apiError(w, http.StatusTooManyRequests, "Too many password change attempts. Try again in 1 hour.")
		return
	}

	changed, err := changeAccountPassword(r.Context(), user, req.Current, req.New)
	if err != nil {
		log.Printf("password change: user %d: %v", user.ID, err)
		apiError(w, http.StatusInternalServerError, "Failed to update password. Please try again.")
		return
	}
	if !changed {
		apiError(w, http.StatusBadRequest, "Current password is incorrect.")
		return
	}

	// Revoke all other active sessions so other devices must re-login.
	if Pools != nil && user.ID > 0 {
		if cookie, err := r.Cookie("platform_session"); err == nil {
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			if err := db.RevokeOtherSessions(ctx, Pools.Admin, user.ID, cookie.Value); err != nil {
				// Non-fatal: log but don't fail the password change.
				log.Printf("password change: revoke other sessions for user %d: %v", user.ID, err)
			}
			cancel()
		}
		go db.InsertAuditEvent(context.Background(), Pools.Admin, user.ID, string(user.Role), "auth.password.change", "info", map[string]any{"email": user.Email}, "", w.Header().Get("X-Request-ID"), clientIP(r))
	}

	apiOK(w, map[string]any{
		"status":  "ok",
		"ok":      true,
		"message": "Password updated successfully.",
	})
}

// ── PATCH /api/user/settings ───────────────────────────────

type apiUserPrefsUpdateRequest struct {
	Theme         *string `json:"theme"`
	Timezone      *string `json:"timezone"`
	Notifications *bool   `json:"notifications_enabled"`
}

func apiUserSettingsPatchHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok {
		apiError(w, http.StatusUnauthorized, "Authentication required.")
		return
	}
	if Pools == nil || user.ID == 0 {
		apiError(w, http.StatusServiceUnavailable, "Database not available.")
		return
	}

	var req apiUserPrefsUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON.")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	prefs, err := db.GetUserPreferences(ctx, Pools.App, user.ID)
	if err != nil {
		apiError(w, http.StatusInternalServerError, "Failed to load preferences.")
		return
	}

	if req.Theme != nil {
		switch *req.Theme {
		case "light", "dark", "system":
			prefs.Theme = *req.Theme
		default:
			apiError(w, http.StatusBadRequest, "Theme must be one of: light, dark, system.")
			return
		}
	}
	if req.Timezone != nil {
		if len(*req.Timezone) > 64 {
			apiError(w, http.StatusBadRequest, "Timezone must not exceed 64 characters.")
			return
		}
		if _, err := time.LoadLocation(*req.Timezone); err != nil {
			apiError(w, http.StatusBadRequest, "Invalid timezone: must be a valid IANA timezone name (e.g. America/New_York).")
			return
		}
		prefs.Timezone = *req.Timezone
	}
	if req.Notifications != nil {
		prefs.EmailNotifications = *req.Notifications
	}

	if err := db.UpdateUserPreferences(ctx, Pools.App, prefs); err != nil {
		apiError(w, http.StatusInternalServerError, "Failed to save preferences.")
		return
	}
	apiOK(w, map[string]string{"status": "ok"})
}
