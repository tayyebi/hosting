package handlers

import (
	"net/http"
	"strings"
	"time"
)

type forgotPasswordData struct {
	Mode         string // "default" | "system" | "tenant"
	OriginDomain string
	FormAction   string
	BackToSignIn string
	TenantBrand  TenantLoginBrand
	Sent         bool
	Email        string
	Error        string
}

// forgotPasswordRateLimiter — 3 POST attempts per IP per 15 minutes.
var forgotPasswordRateLimiter = newIPRateLimiter(fpMaxAttempts, fpWindow)

const (
	fpMaxAttempts = 3
	fpWindow      = 15 * time.Minute
)

func checkFPRate(ip string) bool {
	return forgotPasswordRateLimiter.allow(ip)
}

func buildForgotPasswordData(r *http.Request, sent bool, email, errMsg string) forgotPasswordData {
	oc := originFromContext(r)
	switch oc.Mode {
	case OriginModeTenant:
		if oc.Tenant != nil {
			return forgotPasswordData{
				Mode:         "tenant",
				FormAction:   "/forgot-password",
				BackToSignIn: "/",
				TenantBrand:  resolveTenantLoginBrand(r.Context(), *oc.Tenant),
				Sent:         sent,
				Email:        email,
				Error:        errMsg,
			}
		}
	case OriginModeSystem:
		return forgotPasswordData{
			Mode:         "system",
			OriginDomain: oc.MatchedDomain,
			FormAction:   "/forgot-password",
			BackToSignIn: "/",
			Sent:         sent,
			Email:        email,
			Error:        errMsg,
		}
	}

	if qSlug := strings.TrimSpace(r.URL.Query().Get("tenant")); qSlug != "" {
		if t, found := store.getBySlug(qSlug); found {
			return forgotPasswordData{
				Mode:         "tenant",
				FormAction:   "/forgot-password?tenant=" + t.Slug,
				BackToSignIn: "/t/" + t.Slug + "/login",
				TenantBrand:  resolveTenantLoginBrand(r.Context(), t),
				Sent:         sent,
				Email:        email,
				Error:        errMsg,
			}
		}
	}

	return forgotPasswordData{
		Mode:         "default",
		FormAction:   "/forgot-password",
		BackToSignIn: "/",
		Sent:         sent,
		Email:        email,
		Error:        errMsg,
	}
}

func ForgotPasswordHandler(w http.ResponseWriter, r *http.Request) {
	if user, ok := getSession(r); ok {
		redirectByUser(w, r, user)
		return
	}

	if r.Method == http.MethodPost {
		if !checkFPRate(clientIP(r)) {
			render(w, buildForgotPasswordData(r, false, r.FormValue("email"), "Too many requests. Please wait 15 minutes before trying again."), "base.html", "forgot_password.html")
			return
		}
		email := r.FormValue("email")
		render(w, buildForgotPasswordData(r, true, email, ""), "base.html", "forgot_password.html")
		return
	}

	render(w, buildForgotPasswordData(r, false, "", ""), "base.html", "forgot_password.html")
}
