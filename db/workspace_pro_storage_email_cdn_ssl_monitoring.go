package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ── Pro: Storage — S3-Compatible Object Storage ───────────────────────────────

type ProS3BucketDB struct {
	ResourceID    int64
	Name          string
	Status        string
	Region        string
	StorageClass  string
	SizeGB        float64
	ObjectCount   int64
	Versioning    bool
	Encryption    string
	AccessPolicy  string
	Endpoint      string
	CDNEnabled    bool
	LifecycleRule string
	BucketStatus  string
	CreatedAt     time.Time
}

func ListProS3Buckets(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]ProS3BucketDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       b.storage_class, b.size_gb, b.object_count,
		       b.versioning, b.encryption, b.access_policy,
		       b.endpoint, b.cdn_enabled, b.lifecycle_rule, b.bucket_status,
		       r.created_at
		FROM workspace_s3_buckets b
		JOIN resources r ON r.id = b.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListProS3Buckets: %w", err)
	}
	defer rows.Close()
	var out []ProS3BucketDB
	for rows.Next() {
		var v ProS3BucketDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.StorageClass, &v.SizeGB, &v.ObjectCount,
			&v.Versioning, &v.Encryption, &v.AccessPolicy,
			&v.Endpoint, &v.CDNEnabled, &v.LifecycleRule, &v.BucketStatus,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateProS3Bucket(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, region, storageClass, encryption, accessPolicy string,
) (int64, error) {
	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, created_by)
		VALUES ($1,$2,'pro_s3_bucket',$3,'active',$4)
		RETURNING id`,
		tenantID, workspaceID, name, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateProS3Bucket resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_s3_buckets
		       (resource_id, region, storage_class, encryption, access_policy,
		        endpoint)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		resourceID, region, storageClass, encryption, accessPolicy,
		"s3."+region+".cloud")
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateProS3Bucket row: %w", err)
	}
	return resourceID, nil
}

// ── Pro: Email & Messaging — Transactional Email ──────────────────────────────

type ProTransactionalEmailDB struct {
	ResourceID    int64
	Name          string
	Status        string
	Region        string
	Domain        string
	SenderAddress string
	DKIMStatus    string
	SPFStatus     string
	DMARCStatus   string
	DailySent     int
	DailyLimit    int
	DeliveryRate  float64
	BounceRate    float64
	OpenRate      float64
	WebhookURL    string
	APIKeyPrefix  string
	EmailStatus   string
	CreatedAt     time.Time
}

func ListProTransactionalEmails(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]ProTransactionalEmailDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       e.domain, e.sender_address,
		       e.dkim_status, e.spf_status, e.dmarc_status,
		       e.daily_sent, e.daily_limit,
		       e.delivery_rate, e.bounce_rate, e.open_rate,
		       e.webhook_url, e.api_key_prefix, e.email_status,
		       r.created_at
		FROM workspace_transactional_emails e
		JOIN resources r ON r.id = e.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListProTransactionalEmails: %w", err)
	}
	defer rows.Close()
	var out []ProTransactionalEmailDB
	for rows.Next() {
		var v ProTransactionalEmailDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Domain, &v.SenderAddress,
			&v.DKIMStatus, &v.SPFStatus, &v.DMARCStatus,
			&v.DailySent, &v.DailyLimit,
			&v.DeliveryRate, &v.BounceRate, &v.OpenRate,
			&v.WebhookURL, &v.APIKeyPrefix, &v.EmailStatus,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateProTransactionalEmail(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, domain, senderAddress string, dailyLimit int,
) (int64, error) {
	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, created_by)
		VALUES ($1,$2,'pro_transactional_email',$3,'active',$4)
		RETURNING id`,
		tenantID, workspaceID, name, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateProTransactionalEmail resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_transactional_emails
		       (resource_id, domain, sender_address, daily_limit)
		VALUES ($1,$2,$3,$4)`,
		resourceID, domain, senderAddress, dailyLimit)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateProTransactionalEmail row: %w", err)
	}
	return resourceID, nil
}

// ── Pro: Content Delivery — CDN PoP Management ───────────────────────────────

type ProCDNPoPDB struct {
	ResourceID   int64
	Name         string
	Status       string
	Region       string
	Location     string
	Provider     string
	CacheHitRate float64
	BandwidthTB  float64
	RequestsM    float64
	AvgLatencyMs int
	Capacity     string
	TLSVersion   string
	HTTP2        bool
	HTTP3        bool
	PoPStatus    string
	CreatedAt    time.Time
}

func ListProCDNPoPs(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]ProCDNPoPDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       p.location, p.provider,
		       p.cache_hit_rate, p.bandwidth_tb, p.requests_m,
		       p.avg_latency_ms, p.capacity,
		       p.tls_version, p.http2, p.http3, p.pop_status,
		       r.created_at
		FROM workspace_cdn_pops p
		JOIN resources r ON r.id = p.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListProCDNPoPs: %w", err)
	}
	defer rows.Close()
	var out []ProCDNPoPDB
	for rows.Next() {
		var v ProCDNPoPDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Location, &v.Provider,
			&v.CacheHitRate, &v.BandwidthTB, &v.RequestsM,
			&v.AvgLatencyMs, &v.Capacity,
			&v.TLSVersion, &v.HTTP2, &v.HTTP3, &v.PoPStatus,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateProCDNPoP(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, location, provider, capacity, region string,
) (int64, error) {
	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, created_by)
		VALUES ($1,$2,'pro_cdn_pop',$3,'active',$4)
		RETURNING id`,
		tenantID, workspaceID, name, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateProCDNPoP resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_cdn_pops (resource_id, location, provider, capacity)
		VALUES ($1,$2,$3,$4)`,
		resourceID, location, provider, capacity)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateProCDNPoP row: %w", err)
	}
	return resourceID, nil
}

// ── Pro: Security & SSL — TLS Certificates ────────────────────────────────────

type ProSSLCertDB struct {
	ResourceID   int64
	Name         string
	Status       string
	Region       string
	SAN          string
	Issuer       string
	CertType     string
	AutoRenew    bool
	IssuedAt     string
	ExpiresAt    string
	KeyAlgorithm string
	KeySize      int
	Protocol     string
	OCSPStapling bool
	CTLogged     bool
	CertStatus   string
	CreatedAt    time.Time
}

func ListProSSLCerts(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]ProSSLCertDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       c.san, c.issuer, c.cert_type,
		       c.auto_renew, c.issued_at, c.expires_at,
		       c.key_algorithm, c.key_size, c.protocol,
		       c.ocsp_stapling, c.ct_logged, c.cert_status,
		       r.created_at
		FROM workspace_ssl_certs c
		JOIN resources r ON r.id = c.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListProSSLCerts: %w", err)
	}
	defer rows.Close()
	var out []ProSSLCertDB
	for rows.Next() {
		var v ProSSLCertDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.SAN, &v.Issuer, &v.CertType,
			&v.AutoRenew, &v.IssuedAt, &v.ExpiresAt,
			&v.KeyAlgorithm, &v.KeySize, &v.Protocol,
			&v.OCSPStapling, &v.CTLogged, &v.CertStatus,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateProSSLCert(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	domain, san, issuer, certType string, autoRenew bool,
) (int64, error) {
	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, created_by)
		VALUES ($1,$2,'pro_ssl_cert',$3,'active',$4)
		RETURNING id`,
		tenantID, workspaceID, domain, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateProSSLCert resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_ssl_certs
		       (resource_id, san, issuer, cert_type, auto_renew)
		VALUES ($1,$2,$3,$4,$5)`,
		resourceID, san, issuer, certType, autoRenew)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateProSSLCert row: %w", err)
	}
	return resourceID, nil
}

// ── Pro: Monitoring & Alerts — Alert Rules ────────────────────────────────────

type ProMonitoringAlertDB struct {
	ResourceID     int64
	Name           string
	Status         string
	Region         string
	Metric         string
	Condition      string
	Threshold      float64
	Window         string
	Severity       string
	NotifyChannels string
	EvalInterval   string
	State          string
	LastTriggered  string
	Silenced       bool
	AlertStatus    string
	CreatedAt      time.Time
}

func ListProMonitoringAlerts(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]ProMonitoringAlertDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       a.metric, a.condition, a.threshold,
		       a.alert_window, a.severity, a.notify_channels,
		       a.eval_interval, a.state, a.last_triggered,
		       a.silenced, a.alert_status,
		       r.created_at
		FROM workspace_monitoring_alerts a
		JOIN resources r ON r.id = a.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListProMonitoringAlerts: %w", err)
	}
	defer rows.Close()
	var out []ProMonitoringAlertDB
	for rows.Next() {
		var v ProMonitoringAlertDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Metric, &v.Condition, &v.Threshold,
			&v.Window, &v.Severity, &v.NotifyChannels,
			&v.EvalInterval, &v.State, &v.LastTriggered,
			&v.Silenced, &v.AlertStatus,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateProMonitoringAlert(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, metric, condition string, threshold float64, window, severity, notifyChannels string,
) (int64, error) {
	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, created_by)
		VALUES ($1,$2,'pro_monitoring_alert',$3,'active',$4)
		RETURNING id`,
		tenantID, workspaceID, name, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateProMonitoringAlert resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_monitoring_alerts
		       (resource_id, metric, condition, threshold, alert_window,
		        severity, notify_channels)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		resourceID, metric, condition, threshold, window,
		severity, notifyChannels)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateProMonitoringAlert row: %w", err)
	}
	return resourceID, nil
}
