package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Seed inserts the demo tenants and the platform sysadmin user so that
// the app can authenticate and the Database Manager shows realistic data.
// Idempotent — safe to run on every boot.
func Seed(ctx context.Context, admin *pgxpool.Pool) error {
	if err := seedTenants(ctx, admin); err != nil {
		return err
	}
	if err := seedUsers(ctx, admin); err != nil {
		return err
	}
	if err := seedSystemLogs(ctx, admin); err != nil {
		return err
	}
	if err := seedJobQueues(ctx, admin); err != nil {
		return err
	}
	if err := seedJobWorkers(ctx, admin); err != nil {
		return err
	}
	if err := seedBackgroundJobs(ctx, admin); err != nil {
		return err
	}
	if err := seedReplicas(ctx, admin); err != nil {
		return err
	}
	if err := seedBilling(ctx, admin); err != nil {
		return err
	}
	return nil
}

func seedTenants(ctx context.Context, admin *pgxpool.Pool) error {
	tenants := []struct{ Slug, Name, Domain string }{
		{"buyvm", "BuyVM", "buyvm.net"},
		{"abrnoc", "abrNOC", ""},
		{"hosting", "Hosting", "new.hosting.com"},
	}
	for _, t := range tenants {
		var domain any = t.Domain
		if t.Domain == "" {
			domain = nil
		}
		if _, err := admin.Exec(ctx, `
			INSERT INTO tenants (slug, name, domain, plan)
			VALUES ($1, $2, $3, 'pro')
			ON CONFLICT (slug) DO NOTHING`,
			t.Slug, t.Name, domain,
		); err != nil {
			return fmt.Errorf("db: seed tenant %s: %w", t.Slug, err)
		}
	}
	return nil
}

func seedUsers(ctx context.Context, admin *pgxpool.Pool) error {
	// password_hash is NULL — the in-memory mockUsers map owns credential
	// checks in Phase 1.
	if _, err := admin.Exec(ctx, `
		INSERT INTO users (email, name, global_role)
		VALUES ('sysadmin@platform.io', 'Sys Admin', 'sysadmin')
		ON CONFLICT (email) DO NOTHING`); err != nil {
		return fmt.Errorf("db: seed sysadmin user: %w", err)
	}
	return nil
}

func seedSystemLogs(ctx context.Context, admin *pgxpool.Pool) error {
	// Skip if already seeded
	var n int
	if err := admin.QueryRow(ctx, `SELECT count(*) FROM system_logs`).Scan(&n); err != nil {
		return fmt.Errorf("db: seed system_logs check: %w", err)
	}
	if n > 0 {
		return nil
	}

	type logSpec struct {
		level   string
		service string
		message string
		fields  string
		ago     time.Duration
	}
	specs := []logSpec{
		{"error", "api-gateway", "Connection refused: upstream host unreachable", `{"host":"node-01","latency_ms":"30042"}`, 2 * time.Minute},
		{"error", "db-proxy", "Database query timeout after 30s", `{"host":"node-02","latency_ms":"30001"}`, 5 * time.Minute},
		{"error", "auth-service", "TLS handshake failed: certificate expired", `{"host":"node-01","latency_ms":"12"}`, 14 * time.Minute},
		{"error", "replica-mgr", "Failed to acquire replica lock after 5 retries", `{"host":"node-03","latency_ms":"891"}`, 27 * time.Minute},
		{"error", "billing-svc", "Stripe API returned 429: rate limit exceeded", `{"host":"node-01","latency_ms":"142"}`, 41 * time.Minute},
		{"warn", "api-gateway", "Slow query detected: 2.4s (threshold 500ms)", `{"latency_ms":"2400"}`, 1 * time.Minute},
		{"warn", "replica-mgr", "Replica lag exceeding 120s — degraded read path", `{"host":"node-02","latency_ms":"38"}`, 3 * time.Minute},
		{"warn", "db-proxy", "Connection pool at 95% capacity", `{"latency_ms":"12"}`, 7 * time.Minute},
		{"warn", "metrics-agent", "Cache miss rate elevated: 34% over last 5m", `{"latency_ms":"38"}`, 11 * time.Minute},
		{"warn", "auth-service", "Retrying request to auth-service (attempt 2/3)", `{"host":"node-01","latency_ms":"142"}`, 18 * time.Minute},
		{"warn", "billing-svc", "JWT expiry within 5 minutes for session abc123", `{"latency_ms":"38"}`, 22 * time.Minute},
		{"warn", "tenant-router", "Memory usage at 87% — approaching limit", `{"host":"node-03","latency_ms":"12"}`, 33 * time.Minute},
		{"info", "api-gateway", "Request completed in 142ms — GET /api/v2/tenants", `{"latency_ms":"142"}`, 30 * time.Second},
		{"info", "tenant-router", "Tenant provisioning started: slug=techcorp", `{"latency_ms":"38"}`, 1 * time.Minute},
		{"info", "replica-mgr", "Replica synced successfully — lag 0ms", `{"latency_ms":"12"}`, 2 * time.Minute},
		{"info", "metrics-agent", "Scheduled backup completed: 2.3 GB", `{"latency_ms":"891"}`, 4 * time.Minute},
		{"info", "api-gateway", "Health check passed — all upstreams healthy", `{"latency_ms":"12"}`, 6 * time.Minute},
		{"info", "auth-service", "New session created for user admin@platform.io", `{"latency_ms":"38"}`, 8 * time.Minute},
		{"info", "api-gateway", "Config reloaded from /etc/platform/config.yaml", `{"latency_ms":"12"}`, 10 * time.Minute},
		{"info", "metrics-agent", "Metric scrape completed: 1247 series", `{"latency_ms":"38"}`, 12 * time.Minute},
		{"info", "tenant-router", "Pod autoscaled from 3 → 5 replicas", `{"latency_ms":"142"}`, 15 * time.Minute},
		{"info", "auth-service", "Certificate renewed for *.platform.io", `{"latency_ms":"12"}`, 20 * time.Minute},
		{"info", "billing-svc", "Invoice created: inv_7f3a1c — tenant acme-corp", `{"latency_ms":"142"}`, 25 * time.Minute},
		{"info", "api-gateway", "Request completed in 58ms — GET /api/v2/workspaces", `{"latency_ms":"58"}`, 28 * time.Minute},
		{"info", "tenant-router", "Tenant provisioning completed: slug=devhub", `{"latency_ms":"891"}`, 35 * time.Minute},
		{"info", "replica-mgr", "Heartbeat received from replica-eu-1", `{"latency_ms":"12"}`, 40 * time.Minute},
		{"info", "billing-svc", "Payment succeeded: ch_9e1b3f — $99.00", `{"latency_ms":"210"}`, 45 * time.Minute},
		{"info", "api-gateway", "Request completed in 320ms — POST /api/v2/tenants", `{"latency_ms":"320"}`, 50 * time.Minute},
		{"debug", "db-proxy", "Cache lookup: key=tenant:acme-corp ttl=284s", `{}`, 45 * time.Second},
		{"debug", "api-gateway", "Outgoing HTTP: POST https://billing.internal/charge 200 88ms", `{}`, 1*time.Minute + 30*time.Second},
		{"debug", "db-proxy", "SQL: SELECT * FROM tenants WHERE slug=$1 [0.8ms]", `{}`, 3*time.Minute + 10*time.Second},
		{"debug", "replica-mgr", "gRPC stream opened: replica-mgr → db-proxy", `{}`, 5 * time.Minute},
		{"debug", "metrics-agent", "Goroutine pool: 42 active / 200 max", `{}`, 8 * time.Minute},
		{"debug", "auth-service", "DNS resolved api.internal → 10.0.1.42 (cached)", `{}`, 12 * time.Minute},
	}

	for i, s := range specs {
		traceID := fmt.Sprintf("%016x", uint64(i+1)*0xdeadbeef)
		ts := time.Now().Add(-s.ago)
		if _, err := admin.Exec(ctx, `
			INSERT INTO system_logs (level, service, message, fields, trace_id, created_at)
			VALUES ($1, $2, $3, $4::jsonb, $5, $6)`,
			s.level, s.service, s.message, s.fields, traceID, ts,
		); err != nil {
			return fmt.Errorf("db: seed system_logs[%d]: %w", i, err)
		}
	}
	return nil
}

func seedJobQueues(ctx context.Context, admin *pgxpool.Pool) error {
	queues := []struct{ name, desc string; conc int }{
		{"default", "General-purpose job queue", 4},
		{"critical", "High-priority platform operations", 3},
		{"background", "Long-running background tasks", 2},
		{"mailer", "Email delivery queue", 2},
	}
	for _, q := range queues {
		if _, err := admin.Exec(ctx, `
			INSERT INTO background_job_queues (name, description, concurrency)
			VALUES ($1, $2, $3)
			ON CONFLICT (name) DO NOTHING`,
			q.name, q.desc, q.conc,
		); err != nil {
			return fmt.Errorf("db: seed job queue %s: %w", q.name, err)
		}
	}
	return nil
}

func seedJobWorkers(ctx context.Context, admin *pgxpool.Pool) error {
	var n int
	if err := admin.QueryRow(ctx, `SELECT count(*) FROM background_job_workers`).Scan(&n); err != nil {
		return fmt.Errorf("db: seed workers check: %w", err)
	}
	if n > 0 {
		return nil
	}
	workers := []struct {
		hostname string
		queues   []string
		status   string
	}{
		{"worker-1.platform.internal", []string{"default", "mailer"}, "busy"},
		{"worker-2.platform.internal", []string{"critical"}, "busy"},
		{"worker-3.platform.internal", []string{"critical", "default"}, "busy"},
		{"worker-4.platform.internal", []string{"background"}, "idle"},
		{"worker-5.platform.internal", []string{"background"}, "busy"},
		{"worker-6.platform.internal", []string{"critical"}, "stopped"},
	}
	for _, w := range workers {
		if _, err := admin.Exec(ctx, `
			INSERT INTO background_job_workers (hostname, queues, status, last_heartbeat, started_at)
			VALUES ($1, $2, $3, now(), now() - interval '2 hours')`,
			w.hostname, w.queues, w.status,
		); err != nil {
			return fmt.Errorf("db: seed worker %s: %w", w.hostname, err)
		}
	}
	return nil
}

func seedBackgroundJobs(ctx context.Context, admin *pgxpool.Pool) error {
	var n int
	if err := admin.QueryRow(ctx, `SELECT count(*) FROM background_jobs`).Scan(&n); err != nil {
		return fmt.Errorf("db: seed jobs check: %w", err)
	}
	if n > 0 {
		return nil
	}

	type jobSpec struct {
		queue   string
		kind    string
		status  string
		payload string
		ago     time.Duration
		err     string
	}
	specs := []jobSpec{
		// Running
		{"critical", "tenant.provision", "running", `{"tenant":"acme-corp","plan":"pro"}`, 2 * time.Minute, ""},
		{"background", "backup.create", "running", `{"target":"pg-primary","type":"full"}`, 5 * time.Minute, ""},
		{"default", "billing.sync", "running", `{"tenant":"techcorp","period":"2026-03"}`, 38 * time.Second, ""},
		{"critical", "replica.sync", "running", `{"replica":"replica-eu-2","from":"primary"}`, 7 * time.Minute, ""},
		// Queued
		{"default", "certificate.renew", "queued", `{"domain":"*.devhub.platform.io"}`, 0, ""},
		{"mailer", "email.send", "queued", `{"to":"admin@acme-corp.com","tpl":"invoice"}`, 0, ""},
		{"background", "report.generate", "queued", `{"tenant":"cloudbase","type":"usage","month":"2026-03"}`, 0, ""},
		{"critical", "db.migrate", "queued", `{"version":"20260419_add_replicas","db":"platform"}`, 0, ""},
		{"background", "metrics.flush", "queued", `{"interval":"5m","series":1247}`, 0, ""},
		// Failed
		{"critical", "db.migrate", "failed", `{"version":"20260410_add_audit_log"}`, 65 * time.Minute, "context deadline exceeded after 30s"},
		{"default", "billing.sync", "failed", `{"tenant":"techcorp","period":"2026-02"}`, 42 * time.Minute, "stripe API returned 429: rate limit exceeded"},
		{"critical", "tenant.provision", "failed", `{"tenant":"cloudbase","plan":"pro"}`, 90 * time.Minute, "nil pointer dereference in provisioning step 3"},
		{"background", "backup.create", "failed", `{"target":"pg-primary","type":"full"}`, 120 * time.Minute, "disk quota exceeded on /data/backups"},
		// Succeeded
		{"critical", "tenant.provision", "succeeded", `{"tenant":"acme-corp","plan":"pro"}`, 15 * time.Minute, ""},
		{"default", "billing.sync", "succeeded", `{"tenant":"techcorp","period":"2026-03"}`, 20 * time.Minute, ""},
		{"default", "certificate.renew", "succeeded", `{"domain":"*.acme-corp.platform.io"}`, 30 * time.Minute, ""},
		{"background", "backup.create", "succeeded", `{"target":"pg-primary","type":"full"}`, 45 * time.Minute, ""},
		{"critical", "replica.sync", "succeeded", `{"replica":"replica-eu-1","from":"primary"}`, 60 * time.Minute, ""},
		{"background", "metrics.flush", "succeeded", `{"interval":"5m","series":1247}`, 10 * time.Minute, ""},
		{"mailer", "email.send", "succeeded", `{"to":"ops@techcorp.com","tpl":"alert"}`, 12 * time.Minute, ""},
		{"background", "report.generate", "succeeded", `{"tenant":"acme-corp","type":"usage"}`, 35 * time.Minute, ""},
		{"critical", "db.migrate", "succeeded", `{"version":"20260408_tenant_idx","db":"platform"}`, 80 * time.Minute, ""},
	}

	for i, s := range specs {
		scheduledAt := time.Now().Add(-s.ago)
		var startedAt *time.Time
		if s.status != "queued" {
			t := scheduledAt.Add(2 * time.Second)
			startedAt = &t
		}
		var finishedAt *time.Time
		if s.status == "succeeded" || s.status == "failed" {
			t := scheduledAt.Add(30 * time.Second)
			finishedAt = &t
		}
		var errVal *string
		if s.err != "" {
			errVal = &s.err
		}
		if _, err := admin.Exec(ctx, `
			INSERT INTO background_jobs
			  (queue, kind, payload, status, attempts, max_attempts, last_error,
			   scheduled_at, started_at, finished_at)
			VALUES ($1,$2,$3::jsonb,$4,
			  CASE WHEN $5 != '' THEN 3 ELSE 0 END, 5,
			  $5, $6, $7, $8)`,
			s.queue, s.kind, s.payload, s.status, ptrStrVal(errVal),
			scheduledAt, startedAt, finishedAt,
		); err != nil {
			return fmt.Errorf("db: seed job[%d] %s: %w", i, s.kind, err)
		}
	}
	return nil
}

func ptrStrVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func seedReplicas(ctx context.Context, admin *pgxpool.Pool) error {
	var n int
	if err := admin.QueryRow(ctx, `SELECT count(*) FROM replicas`).Scan(&n); err != nil {
		return fmt.Errorf("db: seed replicas check: %w", err)
	}
	if n > 0 {
		return nil
	}
	replicas := []struct {
		host, region, provider, dc, health string
		ago                                time.Duration
	}{
		{"replica-eu-1.platform.internal", "eu-west-1 (Frankfurt)", "Platform Internal", "Equinix FR5", "healthy", 15 * 24 * time.Hour},
		{"replica-eu-2.platform.internal", "eu-west-1 (Frankfurt)", "Platform Internal", "Equinix FR5", "sync-delayed", 12 * 24 * time.Hour},
		{"replica-us-1.platform.internal", "us-east-1 (New York)", "Platform Internal", "Equinix NY7", "healthy", 9 * 24 * time.Hour},
		{"replica-us-2.platform.internal", "us-east-1 (New York)", "Platform Internal", "Equinix NY7", "overloaded", 7 * 24 * time.Hour},
		{"replica-ap-1.platform.internal", "ap-southeast-1 (Singapore)", "Platform Internal", "Equinix SG3", "out-of-service", 3 * 24 * time.Hour},
		{"replica-ap-2.platform.internal", "ap-southeast-1 (Singapore)", "Platform Internal", "Equinix SG3", "pending", 2 * time.Hour},
	}
	for _, rp := range replicas {
		hb := time.Now().Add(-5 * time.Second)
		if rp.health == "out-of-service" || rp.health == "pending" {
			hb = time.Now().Add(-5 * time.Minute)
		}
		if _, err := admin.Exec(ctx, `
			INSERT INTO replicas (host, region, provider, datacenter, health, added_at, last_heartbeat)
			VALUES ($1,$2,$3,$4,$5, now()-$6::interval, $7)`,
			rp.host, rp.region, rp.provider, rp.dc, rp.health,
			fmt.Sprintf("%d seconds", int(rp.ago.Seconds())), hb,
		); err != nil {
			return fmt.Errorf("db: seed replica %s: %w", rp.host, err)
		}
	}
	return nil
}

func seedBilling(ctx context.Context, admin *pgxpool.Pool) error {
	// Settings
	var n int
	if err := admin.QueryRow(ctx, `SELECT count(*) FROM billing_settings`).Scan(&n); err != nil {
		return fmt.Errorf("db: seed billing_settings check: %w", err)
	}
	if n == 0 {
		if _, err := admin.Exec(ctx, `
			INSERT INTO billing_settings (url, timeout_s, retry, log_all, enabled)
			VALUES ('https://api.billingpro.io/v2', 10, true, true, true)`); err != nil {
			return fmt.Errorf("db: seed billing_settings: %w", err)
		}
	}

	// Webhooks
	if err := admin.QueryRow(ctx, `SELECT count(*) FROM billing_webhooks`).Scan(&n); err != nil {
		return fmt.Errorf("db: seed billing_webhooks check: %w", err)
	}
	if n == 0 {
		webhooks := []struct{ endpoint, events, status string; ok, total int }{
			{"https://hooks.slack.com/services/T00/B00/abc", "invoice.created, payment.succeeded", "active", 4987, 4997},
			{"https://api.internal/billing-events", "all events", "active", 1240, 1240},
		}
		for _, wh := range webhooks {
			if _, err := admin.Exec(ctx, `
				INSERT INTO billing_webhooks (endpoint, events, status, delivery_ok, delivery_total)
				VALUES ($1,$2,$3,$4,$5)`,
				wh.endpoint, wh.events, wh.status, wh.ok, wh.total,
			); err != nil {
				return fmt.Errorf("db: seed billing_webhook: %w", err)
			}
		}
	}

	// API logs
	if err := admin.QueryRow(ctx, `SELECT count(*) FROM billing_api_logs`).Scan(&n); err != nil {
		return fmt.Errorf("db: seed billing_api_logs check: %w", err)
	}
	if n > 0 {
		return nil
	}
	logs := []struct {
		method, path, tenant string
		code, latency        int
		req, resp            string
		ago                  time.Duration
	}{
		{"POST", "/invoices", "tenant_xk92", 200, 142,
			`POST /invoices HTTP/1.1\nHost: api.billingpro.io\n\n{"tenant_id":"tenant_xk92","amount":4900,"currency":"USD","description":"Pro plan - April 2026","due_date":"2026-04-30"}`,
			`{"id":"inv_7f3a1c","status":"issued","amount":4900,"currency":"USD","due_date":"2026-04-30"}`,
			2 * time.Minute},
		{"GET", "/invoices/inv_7f3a1c", "tenant_xk92", 200, 58,
			`GET /invoices/inv_7f3a1c HTTP/1.1\nHost: api.billingpro.io`,
			`{"id":"inv_7f3a1c","status":"issued","amount":4900,"currency":"USD"}`,
			2*time.Minute + 10*time.Second},
		{"POST", "/charges", "tenant_p1m4", 201, 210,
			`POST /charges HTTP/1.1\nHost: api.billingpro.io\n\n{"tenant_id":"tenant_p1m4","amount":9900,"currency":"USD","payment_method":"pm_card_visa_4242"}`,
			`{"id":"ch_9e1b3f","status":"succeeded","amount":9900,"currency":"USD","captured":true}`,
			11 * time.Minute},
		{"POST", "/invoices", "tenant_br07", 422, 95,
			`POST /invoices HTTP/1.1\nHost: api.billingpro.io\n\n{"tenant_id":"tenant_br07","amount":-100,"currency":"USD"}`,
			`{"error":{"code":"invalid_amount","message":"amount must be a positive integer"}}`,
			34 * time.Minute},
		{"GET", "/health", "system", 503, 10004,
			`GET /health HTTP/1.1\nHost: api.billingpro.io`,
			`{"error":{"code":"service_unavailable","message":"Upstream dependency unreachable."}}`,
			60 * time.Minute},
		{"GET", "/subscriptions", "tenant_xk92", 200, 74,
			`GET /subscriptions?tenant_id=tenant_xk92 HTTP/1.1\nHost: api.billingpro.io`,
			`{"data":[{"id":"sub_c3d8a1","plan":"pro","status":"active","renews_at":"2026-05-14T00:00:00Z"}],"total":1}`,
			2 * time.Hour},
	}
	for _, l := range logs {
		ts := time.Now().Add(-l.ago)
		if _, err := admin.Exec(ctx, `
			INSERT INTO billing_api_logs
			  (method, path, tenant_ref, status_code, latency_ms, request_body, response_body, created_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
			l.method, l.path, l.tenant, l.code, l.latency, l.req, l.resp, ts,
		); err != nil {
			return fmt.Errorf("db: seed billing_api_log: %w", err)
		}
	}
	return nil
}

// LookupUserIDByEmail is used by the session layer to resolve the
// seeded user row for a freshly-authenticated login.
func LookupUserIDByEmail(ctx context.Context, admin *pgxpool.Pool, email string) (int64, error) {
	var id int64
	err := admin.QueryRow(ctx,
		`SELECT id FROM users WHERE email = $1 AND deleted_at IS NULL`,
		email).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// UpsertUser ensures a users row exists for the given email; returns the user id.
func UpsertUser(ctx context.Context, admin *pgxpool.Pool, email, name, globalRole string) (int64, error) {
	if globalRole != "user" && globalRole != "sysadmin" {
		globalRole = "user"
	}
	var id int64
	err := admin.QueryRow(ctx, `
		INSERT INTO users (email, name, global_role)
		VALUES ($1, $2, $3)
		ON CONFLICT (email) DO UPDATE SET
		    name        = EXCLUDED.name,
		    global_role = EXCLUDED.global_role,
		    updated_at  = now()
		RETURNING id`,
		email, name, globalRole).Scan(&id)
	return id, err
}
