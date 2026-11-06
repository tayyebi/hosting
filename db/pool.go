// Package db wires the platform to PostgreSQL: pool setup, migrations,
// RLS context helpers, catalog introspection, and the sandboxed ad-hoc
// query runner used by the sysadmin SQL console.
package db

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Pools is the set of connection pools the platform needs. Separating them
// by role means a typo'd DELETE in the sysadmin SQL console can't touch
// anything because Readonly doesn't have the grants to execute it.
type Pools struct {
	// App is the normal application pool. Connects as platform_app; RLS
	// is enforced on every tenant-scoped table.
	App *pgxpool.Pool
	// Admin connects as platform_admin (BYPASSRLS). Used for migrations,
	// sysadmin-only reads across tenants, and the Database Manager's
	// pg_catalog introspection.
	Admin *pgxpool.Pool
	// Readonly connects as platform_readonly (SELECT privileges only).
	// Used by the ad-hoc SQL console to execute sysadmin queries.
	Readonly *pgxpool.Pool
}

// Close closes every pool. Safe to call on a partially-initialised Pools;
// nil fields are skipped.
func (p *Pools) Close() {
	if p == nil {
		return
	}
	if p.App != nil {
		p.App.Close()
	}
	if p.Admin != nil {
		p.Admin.Close()
	}
	if p.Readonly != nil {
		p.Readonly.Close()
	}
}

// Open reads the three DATABASE_URL_* environment variables and opens a
// pool for each. Returns an error if any required URL is missing or any
// pool fails its initial Ping.
//
// Env vars:
//
//	DATABASE_URL           (required) — platform_app
//	DATABASE_URL_ADMIN     (required) — platform_admin (migrations, sysadmin)
//	DATABASE_URL_READONLY  (required) — platform_readonly (SQL console)
//	DB_MAX_CONNS           (optional, default 25)
//	DB_MIN_CONNS           (optional, default 2)
func Open(ctx context.Context) (*Pools, error) {
	appURL := os.Getenv("DATABASE_URL")
	adminURL := os.Getenv("DATABASE_URL_ADMIN")
	roURL := os.Getenv("DATABASE_URL_READONLY")
	if appURL == "" || adminURL == "" || roURL == "" {
		return nil, errors.New("db: DATABASE_URL, DATABASE_URL_ADMIN, and DATABASE_URL_READONLY must all be set")
	}

	maxConns := envInt("DB_MAX_CONNS", 25)
	minConns := envInt("DB_MIN_CONNS", 2)

	app, err := openOne(ctx, appURL, int32(maxConns), int32(minConns))
	if err != nil {
		return nil, fmt.Errorf("db: open app pool: %w", err)
	}
	admin, err := openOne(ctx, adminURL, 5, 1)
	if err != nil {
		app.Close()
		return nil, fmt.Errorf("db: open admin pool: %w", err)
	}
	ro, err := openOne(ctx, roURL, 5, 1)
	if err != nil {
		app.Close()
		admin.Close()
		return nil, fmt.Errorf("db: open readonly pool: %w", err)
	}
	return &Pools{App: app, Admin: admin, Readonly: ro}, nil
}

func openOne(ctx context.Context, url string, maxConns, minConns int32) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}
	cfg.MaxConns = maxConns
	cfg.MinConns = minConns
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.MaxConnIdleTime = 5 * time.Minute
	cfg.HealthCheckPeriod = 30 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("new: %w", err)
	}
	for {
		pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		err = pool.Ping(pingCtx)
		cancel()
		if err == nil {
			return pool, nil
		}
		select {
		case <-ctx.Done():
			pool.Close()
			return nil, fmt.Errorf("ping: %w", err)
		case <-time.After(500 * time.Millisecond):
		}
	}
}

func envInt(name string, fallback int) int {
	v := os.Getenv(name)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}
