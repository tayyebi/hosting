package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"platform/db"
)

// ── GET /api/user/sessions ────────────────────────────────────

type apiSessionItem struct {
	ID        int64  `json:"id"`
	Handle    string `json:"handle"`   // first 8 chars of token_hash
	IP        string `json:"ip"`
	UserAgent string `json:"user_agent"`
	CreatedAt string `json:"created_at"`
	ExpiresAt string `json:"expires_at"`
	IsCurrent bool   `json:"is_current"`
}

func apiUserSessionsHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok {
		apiError(w, http.StatusUnauthorized, "Authentication required.")
		return
	}
	if Pools == nil || user.ID == 0 {
		apiError(w, http.StatusServiceUnavailable, "Database not available.")
		return
	}

	if !sessionListRateLimiter.allow(fmt.Sprintf("%d", user.ID)) {
		w.Header().Set("Retry-After", "60")
		apiError(w, http.StatusTooManyRequests, "Too many requests. Retry after 60 seconds.")
		return
	}

	cookie, err := r.Cookie("platform_session")
	if err != nil {
		apiError(w, http.StatusUnauthorized, "Session cookie missing.")
		return
	}
	currentHash := db.HashSessionToken(cookie.Value)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	sessions, err := db.ListUserSessions(ctx, Pools.Admin, user.ID, currentHash)
	if err != nil {
		apiError(w, http.StatusInternalServerError, "Failed to load sessions.")
		return
	}

	items := make([]apiSessionItem, 0, len(sessions))
	for _, s := range sessions {
		handle := s.TokenHash
		if len(handle) > 8 {
			handle = handle[:8]
		}
		items = append(items, apiSessionItem{
			ID:        s.ID,
			Handle:    handle,
			IP:        s.IPAddress,
			UserAgent: s.UserAgent,
			CreatedAt: s.CreatedAt.UTC().Format(time.RFC3339),
			ExpiresAt: s.ExpiresAt.UTC().Format(time.RFC3339),
			IsCurrent: s.IsCurrent,
		})
	}

	apiOK(w, map[string]any{"sessions": items})
}

// ── DELETE /api/user/sessions/{id} ───────────────────────────

func apiUserSessionRevokeHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok {
		apiError(w, http.StatusUnauthorized, "Authentication required.")
		return
	}
	if Pools == nil || user.ID == 0 {
		apiError(w, http.StatusServiceUnavailable, "Database not available.")
		return
	}

	// Extract {id} from the path suffix after the last /.
	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	rawID := parts[len(parts)-1]
	sessionID, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil || sessionID <= 0 {
		apiError(w, http.StatusBadRequest, "Invalid session ID.")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := db.RevokeSessionByID(ctx, Pools.Admin, user.ID, sessionID); err != nil {
		apiError(w, http.StatusInternalServerError, "Failed to revoke session.")
		return
	}

	// If the user just revoked their own session, clear the cookie too.
	if cookie, err2 := r.Cookie("platform_session"); err2 == nil {
		if currentHash := db.HashSessionToken(cookie.Value); true {
			var sid int64
			_ = Pools.Admin.QueryRow(ctx,
				`SELECT id FROM sessions WHERE token_hash=$1 LIMIT 1`, currentHash).Scan(&sid)
			if sid == sessionID {
				w.Header().Set("X-Session-Revoked-Self", "true")
				http.SetCookie(w, &http.Cookie{Name: "platform_session", Value: "", Path: "/", MaxAge: -1})
			}
		}
	}

	_ = json.NewEncoder(w)
	apiOK(w, map[string]string{"status": "revoked"})
}

// ── POST /api/user/sessions/revoke-all ───────────────────────

func apiUserSessionsRevokeAllHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok {
		apiError(w, http.StatusUnauthorized, "Authentication required.")
		return
	}
	if Pools == nil || user.ID == 0 {
		apiError(w, http.StatusServiceUnavailable, "Database not available.")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := db.RevokeAllUserSessions(ctx, Pools.Admin, user.ID); err != nil {
		apiError(w, http.StatusInternalServerError, "Failed to revoke sessions.")
		return
	}

	// Clear the session cookie so the browser is immediately logged out.
	http.SetCookie(w, &http.Cookie{Name: "platform_session", Value: "", Path: "/", MaxAge: -1})

	if Pools != nil {
		go db.InsertAuditEvent(context.Background(), Pools.Admin, user.ID, string(user.Role), "auth.revoke_all_sessions", "info", map[string]any{"email": user.Email}, "", w.Header().Get("X-Request-ID"), clientIP(r))
	}

	apiOK(w, map[string]string{"status": "all_revoked"})
}
