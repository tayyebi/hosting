package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"platform/db"
	"platform/handlers"
)

// statusRecorder wraps ResponseWriter to capture the written HTTP status code.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (sr *statusRecorder) WriteHeader(code int) {
	sr.status = code
	sr.ResponseWriter.WriteHeader(code)
}

// requestLogger logs every request: method, path, status, duration, and request ID.
func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate a unique request ID for correlation.
		var rid [8]byte
		_, _ = rand.Read(rid[:])
		reqID := hex.EncodeToString(rid[:])
		w.Header().Set("X-Request-ID", reqID)

		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		safePath := strings.NewReplacer("\n", "\\n", "\r", "\\r").Replace(r.URL.Path)
		log.Printf("rid=%s %s %s %d %s", reqID, r.Method, safePath, rec.status, fmt.Sprintf("%dms", time.Since(start).Milliseconds()))
	})
}

// limitRequestBody caps POST/PUT/PATCH request bodies to 1 MiB to prevent
// memory exhaustion from oversized payloads.
func limitRequestBody(next http.Handler) http.Handler {
	const maxBytes = 1 << 20 // 1 MiB
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
		}
		next.ServeHTTP(w, r)
	})
}

// requireJSONContentType rejects POST/PUT/PATCH requests to /api/* whose
// Content-Type is not application/json, preventing content-type confusion.
func requireJSONContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if (r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch) &&
			len(r.URL.Path) >= 4 && r.URL.Path[:4] == "/api" {
			ct := r.Header.Get("Content-Type")
			// Only reject if Content-Type is explicitly set to something other than application/json.
			if ct != "" && !strings.HasPrefix(ct, "application/json") {
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusUnsupportedMediaType)
				w.Write([]byte(`{"error":"Unsupported Media Type","message":"Content-Type must be application/json","code":415}`))
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// apiNoCache sets Cache-Control: no-store on all /api/* responses so that
// sensitive JSON data is never stored by intermediary proxies or the browser cache.
func apiNoCache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Path) >= 4 && r.URL.Path[:4] == "/api" {
			w.Header().Set("Cache-Control", "no-store, must-revalidate")
			w.Header().Set("Pragma", "no-cache")
		}
		next.ServeHTTP(w, r)
	})
}

// securityHeaders wraps every response with standard defensive HTTP headers.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Frame-Options", "SAMEORIGIN")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-XSS-Protection", "1; mode=block")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("Permissions-Policy", "geolocation=(), camera=(), microphone=()")
		h.Set("Cross-Origin-Resource-Policy", "same-origin")
		h.Set("Cross-Origin-Opener-Policy", "same-origin")
		h.Set("X-DNS-Prefetch-Control", "off")
		h.Set("Content-Security-Policy",
			"default-src 'self'; "+
				"script-src 'self' 'unsafe-inline'; "+
				"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; "+
				"font-src 'self' https://fonts.gstatic.com; "+
				"img-src 'self' data:; "+
				"connect-src 'self'")
		h.Del("Server")
		h.Del("X-Powered-By")
		// HSTS: only meaningful over TLS; skip in dev to avoid locking out plain-HTTP.
		if os.Getenv("PLATFORM_ENV") == "production" {
			h.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	// ── database ─────────────────────────────────────────────
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pools, err := db.Open(ctx)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pools.Close()

	migSub, err := fs.Sub(migrationsFS, "migrations")
	if err != nil {
		log.Fatalf("embed sub: %v", err)
	}
	if err := db.Migrate(pools.Admin, migSub); err != nil {
		log.Fatalf("migrations: %v", err)
	}
	if err := db.Seed(ctx, pools.Admin); err != nil {
		log.Fatalf("seed: %v", err)
	}
	handlers.Pools = pools
	log.Println("database: migrations applied, seed complete")

	// Background job: prune expired/revoked sessions older than 7 days, hourly.
	go func() {
		for range time.Tick(time.Hour) {
			bctx, bcancel := context.WithTimeout(context.Background(), 30*time.Second)
			if err := db.PruneExpiredSessions(bctx, pools.Admin); err != nil {
				log.Printf("prune sessions: %v", err)
			}
			bcancel()
		}
	}()

	// Background job: prune expired unaccepted workspace invitations older than 7 days, daily.
	go func() {
		for range time.Tick(24 * time.Hour) {
			bctx, bcancel := context.WithTimeout(context.Background(), 30*time.Second)
			if err := db.PruneExpiredInvitations(bctx, pools.App); err != nil {
				log.Printf("prune invitations: %v", err)
			}
			bcancel()
		}
	}()

	// Background job: prune audit_log rows older than AUDIT_LOG_RETENTION_DAYS (default 90), daily.
	auditRetentionDays := 90
	if v := os.Getenv("AUDIT_LOG_RETENTION_DAYS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			auditRetentionDays = n
		}
	}
	go func() {
		for range time.Tick(24 * time.Hour) {
			bctx, bcancel := context.WithTimeout(context.Background(), 30*time.Second)
			if err := db.PruneOldAuditLogs(bctx, pools.Admin, auditRetentionDays); err != nil {
				log.Printf("prune audit_log: %v", err)
			}
			bcancel()
		}
	}()

	// ── HTTP ─────────────────────────────────────────────────
	mux := http.NewServeMux()

	// Static files (no aggressive cache so CSS/JS edits are picked up immediately)
	staticFS := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, must-revalidate")
		staticFS.ServeHTTP(w, r)
	})))

	// Health check — unauthenticated, for load balancers and uptime monitors
	mux.HandleFunc("GET /healthz", handlers.HealthzHandler)

	// Auth
	mux.HandleFunc("/login", handlers.LoginHandler)
	mux.HandleFunc("/logout", handlers.LogoutHandler)
	mux.HandleFunc("/forgot-password", handlers.ForgotPasswordHandler)
	mux.HandleFunc("/sso-guide", handlers.SSOGuideHandler)
	mux.HandleFunc("/status", handlers.StatusPublicHandler)

	// Sysadmin dashboard
	handlers.RegisterSysadminRoutes(mux)

	// Tenant spaces — /t/{slug}/...
	handlers.RegisterTenantDashboardRoutes(mux)
	handlers.RegisterTenantEmployeeRoutes(mux)
	handlers.RegisterTenantBrandRoutes(mux)
	handlers.RegisterTenantLocalizationRoutes(mux)
	handlers.RegisterTenantClientRoutes(mux)
	handlers.RegisterTenantOneTimeAccessRoutes(mux)
	handlers.RegisterTenantAbuseReportsRoutes(mux)
	handlers.RegisterTenantTicketsRoutes(mux)
	handlers.RegisterTenantRefundComplaintsRoutes(mux)
	handlers.RegisterTenantTransactionsRoutes(mux)
	handlers.RegisterTenantPricingRoutes(mux)
	handlers.RegisterTenantTaxRoutes(mux)
	handlers.RegisterTenantSLARoutes(mux)
	handlers.RegisterTenantRegistrationRoutes(mux)
	handlers.RegisterTenantRegisterPageRoutes(mux)
	handlers.RegisterTenantResourcesRoutes(mux)
	handlers.RegisterTenantMonitoringRoutes(mux)
	handlers.RegisterTenantIDSRoutes(mux)
	handlers.RegisterTenantIncidentsRoutes(mux)
	handlers.RegisterTenantStatusPageRoutes(mux)
	handlers.RegisterTenantMaintenanceRoutes(mux)

	// Workspace client area — /t/{slug}/w/{workspace_id}/...
	handlers.RegisterWorkspaceRoutes(mux)

	// Public fancy marketplace — /wizard
	handlers.RegisterWizardRoutes(mux)

	// Per-user routes (any authenticated role)
	handlers.RegisterUserRoutes(mux)
	handlers.RegisterUserWorkspacesRoutes(mux)

	// REST API — /api/*
	handlers.RegisterAPIRoutes(mux)

	// Root + catch-all 404
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			handlers.ServeError(w, r, 404, "")
			return
		}
		handlers.LoginPage(w, r)
	})

	srv := &http.Server{
		Addr:              "0.0.0.0:3000",
		Handler:           requestLogger(securityHeaders(apiNoCache(limitRequestBody(requireJSONContentType(handlers.APIRateLimit(mux)))))),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	log.Println("Platform server starting on :3000")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
