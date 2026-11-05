package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ── Application Services: Messaging — Message Queues ─────────────────────────

type AppsMessagingQueueDB struct {
	ResourceID  int64
	Name        string
	Status      string
	Region      string
	Engine      string
	QueueType   string
	Messages    string
	Throughput  string
	Consumers   int
	Retention   string
	QueueStatus string
	CreatedAt   time.Time
}

func ListAppsMessagingQueues(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]AppsMessagingQueueDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       q.engine, q.queue_type, q.messages,
		       q.throughput, q.consumers, q.retention, q.queue_status,
		       r.created_at
		FROM workspace_apps_messaging_queues q
		JOIN resources r ON r.id = q.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListAppsMessagingQueues: %w", err)
	}
	defer rows.Close()
	var out []AppsMessagingQueueDB
	for rows.Next() {
		var v AppsMessagingQueueDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Engine, &v.QueueType, &v.Messages,
			&v.Throughput, &v.Consumers, &v.Retention, &v.QueueStatus,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateAppsMessagingQueue(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, engine, queueType, retention string,
) (int64, error) {
	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, created_by)
		VALUES ($1,$2,'apps_messaging_queue',$3,'active',$4)
		RETURNING id`,
		tenantID, workspaceID, name, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateAppsMessagingQueue resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_apps_messaging_queues (resource_id, engine, queue_type, retention)
		VALUES ($1,$2,$3,$4)`,
		resourceID, engine, queueType, retention)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateAppsMessagingQueue row: %w", err)
	}
	return resourceID, nil
}

// ── Application Services: Messaging — Event Streaming ────────────────────────

type AppsMessagingStreamDB struct {
	ResourceID     int64
	Name           string
	Status         string
	Region         string
	Engine         string
	Partitions     int
	Throughput     string
	ConsumerGroups int
	Lag            string
	Retention      string
	StreamStatus   string
	CreatedAt      time.Time
}

func ListAppsMessagingStreams(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]AppsMessagingStreamDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       s.engine, s.partitions, s.throughput,
		       s.consumer_groups, s.lag, s.retention, s.stream_status,
		       r.created_at
		FROM workspace_apps_messaging_streaming s
		JOIN resources r ON r.id = s.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListAppsMessagingStreams: %w", err)
	}
	defer rows.Close()
	var out []AppsMessagingStreamDB
	for rows.Next() {
		var v AppsMessagingStreamDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Engine, &v.Partitions, &v.Throughput,
			&v.ConsumerGroups, &v.Lag, &v.Retention, &v.StreamStatus,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateAppsMessagingStream(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, engine, retention string, partitions int,
) (int64, error) {
	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, created_by)
		VALUES ($1,$2,'apps_messaging_stream',$3,'active',$4)
		RETURNING id`,
		tenantID, workspaceID, name, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateAppsMessagingStream resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_apps_messaging_streaming (resource_id, engine, partitions, retention)
		VALUES ($1,$2,$3,$4)`,
		resourceID, engine, partitions, retention)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateAppsMessagingStream row: %w", err)
	}
	return resourceID, nil
}

// ── Application Services: Messaging — Service Bus ─────────────────────────────

type AppsMessagingServiceBusDB struct {
	ResourceID   int64
	Name         string
	Status       string
	Region       string
	Tier         string
	Subscriptions int
	Throughput   string
	DeadLetter   int
	AvgDelivery  string
	MsgSize      string
	BusStatus    string
	CreatedAt    time.Time
}

func ListAppsMessagingServiceBuses(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]AppsMessagingServiceBusDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       b.tier, b.subscriptions, b.throughput,
		       b.dead_letter, b.avg_delivery, b.msg_size, b.bus_status,
		       r.created_at
		FROM workspace_apps_messaging_servicebus b
		JOIN resources r ON r.id = b.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListAppsMessagingServiceBuses: %w", err)
	}
	defer rows.Close()
	var out []AppsMessagingServiceBusDB
	for rows.Next() {
		var v AppsMessagingServiceBusDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Tier, &v.Subscriptions, &v.Throughput,
			&v.DeadLetter, &v.AvgDelivery, &v.MsgSize, &v.BusStatus,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateAppsMessagingServiceBus(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, tier string,
) (int64, error) {
	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, created_by)
		VALUES ($1,$2,'apps_messaging_servicebus',$3,'active',$4)
		RETURNING id`,
		tenantID, workspaceID, name, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateAppsMessagingServiceBus resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_apps_messaging_servicebus (resource_id, tier)
		VALUES ($1,$2)`,
		resourceID, tier)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateAppsMessagingServiceBus row: %w", err)
	}
	return resourceID, nil
}

// ── Application Services: Workflow — Automation ───────────────────────────────

type AppsWorkflowAutomationDB struct {
	ResourceID  int64
	Name        string
	Status      string
	Region      string
	Trigger     string
	Actions     int
	Runs        string
	SuccessRate string
	LastRun     string
	WfStatus    string
	CreatedAt   time.Time
}

func ListAppsWorkflowAutomations(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]AppsWorkflowAutomationDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       w.trigger, w.actions, w.runs,
		       w.success_rate, w.last_run, w.wf_status,
		       r.created_at
		FROM workspace_apps_workflow_automation w
		JOIN resources r ON r.id = w.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListAppsWorkflowAutomations: %w", err)
	}
	defer rows.Close()
	var out []AppsWorkflowAutomationDB
	for rows.Next() {
		var v AppsWorkflowAutomationDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Trigger, &v.Actions, &v.Runs,
			&v.SuccessRate, &v.LastRun, &v.WfStatus,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateAppsWorkflowAutomation(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, trigger string, actions int,
) (int64, error) {
	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, created_by)
		VALUES ($1,$2,'apps_workflow_automation',$3,'active',$4)
		RETURNING id`,
		tenantID, workspaceID, name, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateAppsWorkflowAutomation resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_apps_workflow_automation (resource_id, trigger, actions)
		VALUES ($1,$2,$3)`,
		resourceID, trigger, actions)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateAppsWorkflowAutomation row: %w", err)
	}
	return resourceID, nil
}

// ── Application Services: Workflow — Scheduler / Cron Jobs ───────────────────

type AppsWorkflowSchedulerDB struct {
	ResourceID int64
	Name       string
	Status     string
	Region     string
	Schedule   string
	Timezone   string
	LastRun    string
	Duration   string
	NextRun    string
	Streak     string
	JobStatus  string
	CreatedAt  time.Time
}

func ListAppsWorkflowSchedulers(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]AppsWorkflowSchedulerDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       j.schedule, j.timezone, j.last_run,
		       j.duration, j.next_run, j.streak, j.job_status,
		       r.created_at
		FROM workspace_apps_workflow_scheduler j
		JOIN resources r ON r.id = j.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListAppsWorkflowSchedulers: %w", err)
	}
	defer rows.Close()
	var out []AppsWorkflowSchedulerDB
	for rows.Next() {
		var v AppsWorkflowSchedulerDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Schedule, &v.Timezone, &v.LastRun,
			&v.Duration, &v.NextRun, &v.Streak, &v.JobStatus,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateAppsWorkflowScheduler(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, schedule, timezone string,
) (int64, error) {
	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, created_by)
		VALUES ($1,$2,'apps_workflow_scheduler',$3,'active',$4)
		RETURNING id`,
		tenantID, workspaceID, name, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateAppsWorkflowScheduler resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_apps_workflow_scheduler (resource_id, schedule, timezone)
		VALUES ($1,$2,$3)`,
		resourceID, schedule, timezone)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateAppsWorkflowScheduler row: %w", err)
	}
	return resourceID, nil
}
