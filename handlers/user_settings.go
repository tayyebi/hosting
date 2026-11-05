package handlers

import (
	"context"
	"html/template"
	"net/http"
	"path/filepath"
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
		"slice": func(s string, a, b int) string {
			if b > len(s) {
				b = len(s)
			}
			return s[a:b]
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
