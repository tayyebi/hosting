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
	OK    bool         `json:"ok"`
	Token string       `json:"token,omitempty"`
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

	oc := originFromContext(r)
	var (
		sessionUser User
		ok          bool
	)

	switch oc.Mode {
	case OriginModeSystem:
		u, status := authenticateSysadminCredentials(r.Context(), req.Email, req.Password)
		if status == db.SysadminAuthForbiddenRole {
			recordLoginFailure(ip)
			if Pools != nil {
				go db.InsertAuditEvent(context.Background(), Pools.Admin, u.ID, "user",
					"auth.login.denied", "warn", map[string]any{"email": req.Email, "origin": oc.MatchedDomain},
					"non-sysadmin API login rejected on system origin", reqID, ip)
			}
			apiError(w, http.StatusForbidden, "Only platform system administrators can sign in on this domain.")
			return
		}
		if status == db.SysadminAuthOK {
			sessionUser = u
			ok = true
		}
	case OriginModeTenant:
		if oc.Tenant != nil {
			sessionUser, ok = authenticateTenantCredentials(r.Context(), *oc.Tenant, req.Email, req.Password)
		}
	default:
		sessionUser, ok = authenticateCredentials(r.Context(), req.Email, req.Password)
	}

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

	token := ""
	for _, c := range w.Header().Values("Set-Cookie") {
		if len(c) > len("platform_session=") {
			token = "session-cookie-set"
		}
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
