package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Replica is one row from the replicas table.
type Replica struct {
	ID             int64
	ReplicaID      string
	Host           string
	Region         string
	Provider       string
	Datacenter     string
	Health         string
	MaxWorkspaces  int
	MaxConnections int
	AddedAt        time.Time
	LastHeartbeat  *time.Time
}

// WALReceiverStats is scraped from pg_stat_wal_receiver on a standby.
type WALReceiverStats struct {
	Status          string
	PrimaryHost     string
	ReceivedLSN     string
	LastMsgSendTime time.Time
	LastMsgRecvTime time.Time
	ReplyTime       time.Time
	WriteLSN        string
	FlushLSN        string
	ReplayLSN       string
	SyncLagMs       int64 // approx lag in ms
}

// ListReplicas returns all replicas ordered by added_at.
func ListReplicas(ctx context.Context, pool *pgxpool.Pool) ([]Replica, error) {
	rows, err := pool.Query(ctx, `
		SELECT id, replica_id, host, region, provider, datacenter, health,
		       max_workspaces, max_connections, added_at, last_heartbeat
		FROM replicas
		ORDER BY added_at`)
	if err != nil {
		return nil, fmt.Errorf("db.ListReplicas: %w", err)
	}
	defer rows.Close()
	var out []Replica
	for rows.Next() {
		var r Replica
		if err := rows.Scan(
			&r.ID, &r.ReplicaID, &r.Host, &r.Region, &r.Provider, &r.Datacenter,
			&r.Health, &r.MaxWorkspaces, &r.MaxConnections, &r.AddedAt, &r.LastHeartbeat,
		); err != nil {
			return nil, fmt.Errorf("db.ListReplicas scan: %w", err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// AddReplica inserts a new replica and returns its assigned ID.
func AddReplica(ctx context.Context, pool *pgxpool.Pool, host, region string) (Replica, error) {
	var r Replica
	err := pool.QueryRow(ctx, `
		INSERT INTO replicas (host, region)
		VALUES ($1, $2)
		RETURNING id, replica_id, host, region, provider, datacenter, health,
		          max_workspaces, max_connections, added_at, last_heartbeat`,
		host, region,
	).Scan(
		&r.ID, &r.ReplicaID, &r.Host, &r.Region, &r.Provider, &r.Datacenter,
		&r.Health, &r.MaxWorkspaces, &r.MaxConnections, &r.AddedAt, &r.LastHeartbeat,
	)
	if err != nil {
		return Replica{}, fmt.Errorf("db.AddReplica: %w", err)
	}
	return r, nil
}

// DeleteReplica removes a replica by its string replica_id.
func DeleteReplica(ctx context.Context, pool *pgxpool.Pool, replicaID string) error {
	_, err := pool.Exec(ctx, `DELETE FROM replicas WHERE replica_id = $1`, replicaID)
	if err != nil {
		return fmt.Errorf("db.DeleteReplica: %w", err)
	}
	return nil
}

// ReplicaHostExists reports whether a replica with the given host is registered.
func ReplicaHostExists(ctx context.Context, pool *pgxpool.Pool, host string) (bool, error) {
	var exists bool
	err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM replicas WHERE host = $1)`, host).Scan(&exists)
	return exists, err
}

// GetCurrentReplica returns the replica row whose host matches the provided
// hostname. Returns pgx.ErrNoRows when not found.
func GetCurrentReplica(ctx context.Context, pool *pgxpool.Pool, hostname string) (Replica, error) {
	var r Replica
	err := pool.QueryRow(ctx, `
		SELECT id, replica_id, host, region, provider, datacenter, health,
		       max_workspaces, max_connections, added_at, last_heartbeat
		FROM replicas WHERE host = $1`, hostname,
	).Scan(
		&r.ID, &r.ReplicaID, &r.Host, &r.Region, &r.Provider, &r.Datacenter,
		&r.Health, &r.MaxWorkspaces, &r.MaxConnections, &r.AddedAt, &r.LastHeartbeat,
	)
	if err != nil {
		return Replica{}, err // callers check pgx.ErrNoRows
	}
	return r, nil
}

// GetWALReceiverStats reads pg_stat_wal_receiver. Returns an empty struct (not
// an error) when the server is a primary (no receiver row).
func GetWALReceiverStats(ctx context.Context, pool *pgxpool.Pool) (WALReceiverStats, error) {
	var s WALReceiverStats
	err := pool.QueryRow(ctx, `
		SELECT
			COALESCE(status,'unknown'),
			COALESCE(conninfo,''),
			COALESCE(received_lsn::text,'0/0'),
			COALESCE(last_msg_send_time, now()),
			COALESCE(last_msg_receipt_time, now()),
			COALESCE(reply_time, now()),
			COALESCE(flushed_lsn::text,'0/0'),
			COALESCE(flushed_lsn::text,'0/0'),
			COALESCE(received_lsn::text,'0/0'),
			COALESCE(
			    extract(epoch FROM (now()-last_msg_receipt_time))*1000,
			    0)::bigint
		FROM pg_stat_wal_receiver
		LIMIT 1`).Scan(
		&s.Status, &s.PrimaryHost,
		&s.ReceivedLSN, &s.LastMsgSendTime, &s.LastMsgRecvTime, &s.ReplyTime,
		&s.WriteLSN, &s.FlushLSN, &s.ReplayLSN, &s.SyncLagMs,
	)
	if err == pgx.ErrNoRows {
		// Primary node — no WAL receiver stats
		s.Status = "primary"
		return s, nil
	}
	if err != nil {
		return s, fmt.Errorf("db.GetWALReceiverStats: %w", err)
	}
	return s, nil
}
