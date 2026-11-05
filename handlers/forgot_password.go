package handlers

import (
	"net/http"
	"time"
)

type forgotPasswordData struct {
	Sent  bool
	Email string
	Error string
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

func ForgotPasswordHandler(w http.ResponseWriter, r *http.Request) {
	if user, ok := getSession(r); ok {
		redirectByRole(w, r, user.Role)
		return
	}

	if r.Method == http.MethodPost {
		if !checkFPRate(clientIP(r)) {
			render(w, forgotPasswordData{Error: "Too many requests. Please wait 15 minutes before trying again."}, "base.html", "forgot_password.html")
			return
		}
		email := r.FormValue("email")
		// Mock: always show "sent" regardless of whether the address exists
		// so we don't leak which addresses are registered.
		render(w, forgotPasswordData{Sent: true, Email: email}, "base.html", "forgot_password.html")
		return
	}

	render(w, forgotPasswordData{}, "base.html", "forgot_password.html")
}
