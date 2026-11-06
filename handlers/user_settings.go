package handlers

import (
	"context"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"platform/db"
)

// UserSettingsData is passed to the user_settings.html template.
type UserSettingsData struct {
	User    User
	Prefs   db.UserPreferences
	APIKeys []db.PersonalAPIKey
}

func UserSettingsPage(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Accounts created without a display name (e.g. sysadmin bootstrap without
	// SYSADMIN_NAME) still need a usable label for the avatar/breadcrumb.
	if strings.TrimSpace(user.Name) == "" {
		user.Name = user.Email
		if i := strings.IndexByte(user.Name, '@'); i > 0 {
			user.Name = user.Name[:i]
		}
	}

	data := UserSettingsData{
		User: user,
		Prefs: db.UserPreferences{
			UserID:             user.ID,
			Theme:              "system",
			Timezone:           "UTC",
			Locale:             "en-US",
			EmailNotifications: true,
		},
	}

	if Pools != nil && user.ID > 0 {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		if prefs, err := db.GetUserPreferences(ctx, Pools.App, user.ID); err == nil {
			data.Prefs = prefs
		}
		if keys, err := db.ListUserPersonalAPIKeys(ctx, Pools.App, user.ID); err == nil {
			data.APIKeys = keys
		}
	}

	tmpl, err := template.New("user_settings.html").Funcs(template.FuncMap{
		// slice is used for the avatar initial. It must never panic: a missing
		// name or a multi-byte first rune previously made the whole page fail
		// to render.
		"slice": func(s string, a, b int) string {
			if s == "" {
				return "?"
			}
			rs := []rune(s)
			if a < 0 {
				a = 0
			}
			if a > len(rs) {
				a = len(rs)
			}
			if b > len(rs) {
				b = len(rs)
			}
			if b < a {
				b = a
			}
			return string(rs[a:b])
		},
		"fmtTime": func(t time.Time) string {
			return t.Format("Jan 2, 2006")
		},
		"fmtTimeAge": func(tp *time.Time) string {
			if tp == nil {
				return "never"
			}
			age := time.Since(*tp)
			switch {
			case age < time.Minute:
				return "just now"
			case age < time.Hour:
				return itoa(int(age.Minutes())) + "m ago"
			case age < 24*time.Hour:
				return itoa(int(age.Hours())) + "h ago"
			default:
				return tp.Format("Jan 2")
			}
		},
	}).ParseFiles(filepath.Join("templates", "user_settings.html"))
	if err != nil {
		ServeError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	if err := tmpl.ExecuteTemplate(w, "user-settings", data); err != nil {
		ServeError(w, r, http.StatusInternalServerError, err.Error())
	}
}

func RegisterUserRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/u/settings/", UserSettingsPage)
}
