package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"platform/db"
	"time"
)

type errorPageData struct {
	Code      int
	Title     string
	Message   string
	Hash      string
	Year      int
	Path      string
	Timestamp string
}

var errorTitles = map[int]string{
	400: "Bad Request",
	401: "Unauthorized",
	403: "Forbidden",
	404: "Not Found",
	405: "Method Not Allowed",
	408: "Request Timeout",
	409: "Conflict",
	429: "Too Many Requests",
	500: "Internal Server Error",
	502: "Bad Gateway",
	503: "Service Unavailable",
	504: "Gateway Timeout",
}

var errorMessages = map[int]string{
	400: "The request could not be understood by the server.",
	401: "Authentication is required to access this resource.",
	403: "You don't have permission to access this resource.",
	404: "The page or resource you're looking for doesn't exist.",
	405: "This request method is not supported for this endpoint.",
	408: "The server timed out waiting for the request.",
	409: "The request conflicts with the current state of the server.",
	429: "You've made too many requests. Please slow down.",
	500: "Something went wrong on our end. We're looking into it.",
	502: "An upstream service returned an invalid response.",
	503: "The service is temporarily unavailable. Please try again shortly.",
	504: "An upstream service did not respond in time.",
}

func errorHash() string {
	b := make([]byte, 6)
	rand.Read(b)
	ts := fmt.Sprintf("%x", time.Now().UnixMilli()&0xFFFFFF)
	return ts + "-" + hex.EncodeToString(b)
}

// ServeError renders the error page. Call this instead of http.Error everywhere.
func ServeError(w http.ResponseWriter, r *http.Request, code int, customMessage string) {
	title, ok := errorTitles[code]
	if !ok {
		if code >= 500 {
			title = "Server Error"
		} else {
			title = "Client Error"
		}
	}

	msg := customMessage
	if msg == "" {
		if m, ok := errorMessages[code]; ok {
			msg = m
		} else {
			msg = "An unexpected error occurred."
		}
	}

	now := time.Now()
	data := errorPageData{
		Code:      code,
		Title:     title,
		Message:   msg,
		Hash:      errorHash(),
		Year:      now.Year(),
		Path:      r.URL.Path,
		Timestamp: now.UTC().Format(time.RFC3339),
	}

	tmpl, err := template.ParseFiles(filepath.Join("templates", "error.html"))
	if err != nil {
		// Fallback if template missing
		http.Error(w, fmt.Sprintf("%d %s", code, title), code)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(code)

	// Log 5xx errors to system_logs for observability
	if code >= 500 && Pools != nil {
		go func() {
			ctx2, cancel2 := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel2()
			fields := map[string]any{
				"method": r.Method,
				"path":   r.URL.Path,
				"code":   code,
			}
			_ = db.WriteSystemLog(ctx2, Pools.Admin, "error", "http", fmt.Sprintf("HTTP %d %s %s: %s", code, r.Method, r.URL.Path, msg), fields)
		}()
	}

	tmpl.ExecuteTemplate(w, "error-page", data)
}

// NotFoundHandler is a drop-in for mux.HandleFunc fallback.
func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	ServeError(w, r, http.StatusNotFound, "")
}
