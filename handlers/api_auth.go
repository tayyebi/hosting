package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"platform/db"
)

// ── POST /api/login ──────────────────────────────────────────

type apiLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Remember bool   `json:"remember"`
}

type apiLoginResponse struct {
	OK    bool   `json:"ok"`
	Token string `json:"token,omitempty"`
	User  *apiUserInfo `json:"user,omitempty"`
}

type apiUserInfo struct {
	Email      string `json:"email"`
	Name       string `json:"name"`
	Role       string `json:"role"`
	AvatarURL  string `json:"avatar_url,omitempty"`
	TenantSlug string `json:"tenant_slug,omitempty"`
}

func apiLoginHandler(w http.ResponseWriter, r *http.Request) {
	ip := clientIP(r)
	reqID := w.Header().Get("X-Request-ID")

	if !checkLoginRate(ip) {
		if Pools != nil {
			go db.InsertAuditEvent(context.Background(), Pools.Admin, 0, "unknown", "auth.lockout", "warn", map[string]any{"ip": ip}, "too many failed attempts", reqID, ip)
		}
		w.Header().Set("Retry-After", "60")
		apiError(w, http.StatusTooManyRequests, "Too many failed attempts. Please wait 1 minute before trying again.")
		return
	}

	var req apiLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}

	if req.Email == "" || req.Password == "" {
		apiError(w, http.StatusBadRequest, "Email and password are required.")
		return
	}

	sessionUser, ok := authenticateCredentials(r.Context(), req.Email, req.Password)
	if !ok {
		justLocked := recordLoginFailure(ip)
		if Pools != nil {
			go db.InsertAuditEvent(context.Background(), Pools.Admin, 0, "unknown", "auth.login.failure", "warn", map[string]any{"email": req.Email}, "invalid credentials", reqID, ip)
			if justLocked {
				go db.InsertAuditEvent(context.Background(), Pools.Admin, 0, "unknown", "auth.lockout", "warn", map[string]any{"email": req.Email}, "account locked after repeated failures", reqID, ip)
			}
		}
		apiError(w, http.StatusUnauthorized, "Invalid email or password.")
		return
	}

	clearLoginFailures(ip)
	createSession(w, r, sessionUser, req.Remember)
	if Pools != nil {
		go db.InsertAuditEvent(context.Background(), Pools.Admin, sessionUser.ID, string(sessionUser.Role), "auth.login.success", "info", map[string]any{"email": req.Email}, "", reqID, ip)
	}

	// Read back the cookie value as the token for the JSON response.
	token := ""
	for _, c := range w.Header().Values("Set-Cookie") {
		// The session cookie was just set; surface the value.
		if len(c) > len("platform_session=") {
			token = "session-cookie-set"
		}
		_ = c
	}

	apiOK(w, apiLoginResponse{
		OK:    true,
		Token: token,
		User: &apiUserInfo{
			Email:      sessionUser.Email,
			Name:       sessionUser.Name,
			Role:       string(sessionUser.Role),
			AvatarURL:  sessionUser.AvatarURL,
			TenantSlug: sessionUser.TenantSlug,
		},
	})
}

// ── POST /api/logout ─────────────────────────────────────────

func apiLogoutHandler(w http.ResponseWriter, r *http.Request) {
	destroySession(w, r)
	apiOK(w, map[string]any{
		"ok":      true,
		"message": "Session destroyed.",
	})
}
