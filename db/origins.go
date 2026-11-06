package db

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrOriginDomainConflict indicates a domain is already registered as a system
// origin domain or assigned to a tenant.
var ErrOriginDomainConflict = errors.New("domain is already in use")

// ErrInvalidDomain indicates a malformed hostname/domain input.
var ErrInvalidDomain = errors.New("invalid domain name")

// OriginSettings holds global settings for the system-level Origin configuration.
type OriginSettings struct {
	RedirectScheme        string    `json:"redirect_scheme"`
	TrustForwardedHeaders bool      `json:"trust_forwarded_headers"`
	RedirectStatusCode    int       `json:"redirect_status_code"`
	UpdatedAt             time.Time `json:"updated_at"`
}

// DefaultOriginSettings returns the default origin settings when not yet customized.
func DefaultOriginSettings() OriginSettings {
	return OriginSettings{
		RedirectScheme:        "https",
		TrustForwardedHeaders: true,
		RedirectStatusCode:    302,
	}
}

// OriginDomain represents an entry in system_origin_domains.
type OriginDomain struct {
	ID        int64     `json:"id"`
	Domain    string    `json:"domain"`
	Note      string    `json:"note"`
	Position  int       `json:"position"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NormalizeDomain cleans and validates a domain (with optional :port).
// Leading http:// or https:// and a single trailing "/" are stripped for
// operator convenience; paths, query strings, fragments, and userinfo are rejected.
func NormalizeDomain(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", fmt.Errorf("%w: domain is required", ErrInvalidDomain)
	}
	lower := strings.ToLower(s)
	if strings.HasPrefix(lower, "https://") {
		s = s[len("https://"):]
	} else if strings.HasPrefix(lower, "http://") {
		s = s[len("http://"):]
	}
	s = strings.TrimSuffix(s, "/")
	s = strings.TrimSpace(s)
	if s == "" {
		return "", fmt.Errorf("%w: domain is required", ErrInvalidDomain)
	}
	if strings.ContainsAny(s, "/?#@\\ \t\r\n") {
		return "", fmt.Errorf("%w: must contain only a hostname (and optional :port), without path or query", ErrInvalidDomain)
	}

	hostPart := s
	portPart := ""

	if strings.HasPrefix(s, "[") {
		// Bracketed IPv6 with or without port.
		if strings.Contains(s, "]:") {
			h, p, err := net.SplitHostPort(s)
			if err != nil {
				return "", fmt.Errorf("%w: %v", ErrInvalidDomain, err)
			}
			hostPart = h
			portPart = p
		} else if strings.HasSuffix(s, "]") {
			hostPart = strings.TrimSuffix(strings.TrimPrefix(s, "["), "]")
		} else {
			return "", fmt.Errorf("%w: malformed IPv6 address", ErrInvalidDomain)
		}
		ip := net.ParseIP(hostPart)
		if ip == nil {
			return "", fmt.Errorf("%w: invalid IPv6 address", ErrInvalidDomain)
		}
		if portPart != "" {
			if err := validatePort(portPart); err != nil {
				return "", err
			}
			return "[" + strings.ToLower(ip.String()) + "]:" + portPart, nil
		}
		return "[" + strings.ToLower(ip.String()) + "]", nil
	}

	if strings.Count(s, ":") == 1 {
		h, p, err := net.SplitHostPort(s)
		if err != nil {
			return "", fmt.Errorf("%w: %v", ErrInvalidDomain, err)
		}
		hostPart = h
		portPart = p
	} else if strings.Count(s, ":") > 1 {
		return "", fmt.Errorf("%w: IPv6 addresses must be enclosed in brackets", ErrInvalidDomain)
	}

	hostPart = strings.TrimSuffix(strings.ToLower(hostPart), ".")
	if hostPart == "" || len(hostPart) > 253 {
		return "", fmt.Errorf("%w: hostname must be between 1 and 253 characters", ErrInvalidDomain)
	}

	if portPart != "" {
		if err := validatePort(portPart); err != nil {
			return "", err
		}
	}

	// Accept valid IPv4, "localhost", or standard DNS hostname.
	if ip := net.ParseIP(hostPart); ip != nil && ip.To4() != nil {
		if portPart != "" {
			return ip.String() + ":" + portPart, nil
		}
		return ip.String(), nil
	}

	if hostPart != "localhost" {
		if !isValidDNSName(hostPart) {
			return "", fmt.Errorf("%w: must be a valid domain name (e.g. cloud.example.com)", ErrInvalidDomain)
		}
	}

	if portPart != "" {
		return hostPart + ":" + portPart, nil
	}
	return hostPart, nil
}

func validatePort(portPart string) error {
	p, err := strconv.Atoi(portPart)
	if err != nil || p < 1 || p > 65535 {
		return fmt.Errorf("%w: port must be between 1 and 65535", ErrInvalidDomain)
	}
	return nil
}

func isValidDNSName(host string) bool {
	labels := strings.Split(host, ".")
	if len(labels) < 2 {
		return false
	}
	for _, lbl := range labels {
		if len(lbl) < 1 || len(lbl) > 63 {
			return false
		}
		if lbl[0] == '-' || lbl[len(lbl)-1] == '-' {
			return false
		}
		for i := 0; i < len(lbl); i++ {
			c := lbl[i]
			isAlpha := c >= 'a' && c <= 'z'
			isDigit := c >= '0' && c <= '9'
			if !isAlpha && !isDigit && c != '-' {
				return false
			}
		}
	}
	tld := labels[len(labels)-1]
	if len(tld) < 2 {
		return false
	}
	if strings.HasPrefix(tld, "xn--") {
		return len(tld) > 4
	}
	for i := 0; i < len(tld); i++ {
		if tld[i] < 'a' || tld[i] > 'z' {
			return false
		}
	}
	return true
}

// SplitHostAndPort extracts both the normalized full host[:port] and the bare
// host (without port) from an HTTP Host / X-Forwarded-Host value.
func SplitHostAndPort(hostHeader string) (fullHost string, bareHost string) {
	s := strings.TrimSpace(hostHeader)
	if idx := strings.Index(s, ","); idx >= 0 {
		s = strings.TrimSpace(s[:idx])
	}
	s = strings.ToLower(s)
	if s == "" {
		return "", ""
	}
	if strings.HasPrefix(s, "[") {
		if strings.Contains(s, "]:") {
			if h, p, err := net.SplitHostPort(s); err == nil {
				clean := "[" + h + "]"
				return clean + ":" + p, clean
			}
		}
		return s, s
	}
	if strings.Count(s, ":") == 1 {
		if h, p, err := net.SplitHostPort(s); err == nil {
			h = strings.TrimSuffix(h, ".")
			return h + ":" + p, h
		}
	}
	s = strings.TrimSuffix(s, ".")
	return s, s
}

// HostMatchesConfiguredDomain checks whether an incoming HTTP request host
// (fullHost = "host:port" or "host", bareHost = "host") matches a configured
// domain entry. If the configured entry specifies a port, the port must match;
// otherwise matching is performed against bareHost.
func HostMatchesConfiguredDomain(fullHost, bareHost, configured string) bool {
	cfg := strings.ToLower(strings.TrimSpace(configured))
	if cfg == "" || bareHost == "" {
		return false
	}
	cfgHasPort := false
	if strings.HasPrefix(cfg, "[") {
		cfgHasPort = strings.Contains(cfg, "]:")
	} else {
		cfgHasPort = strings.Count(cfg, ":") == 1
	}
	if cfgHasPort {
		return fullHost == cfg
	}
	return bareHost == cfg
}

// ValidateOriginSettings normalizes and validates OriginSettings fields.
func ValidateOriginSettings(s OriginSettings) (OriginSettings, error) {
	s.RedirectScheme = strings.ToLower(strings.TrimSpace(s.RedirectScheme))
	if s.RedirectScheme == "" {
		s.RedirectScheme = "https"
	}
	if s.RedirectScheme != "https" && s.RedirectScheme != "http" {
		return s, fmt.Errorf("redirect_scheme must be 'https' or 'http'")
	}
	if s.RedirectStatusCode == 0 {
		s.RedirectStatusCode = 302
	}
	switch s.RedirectStatusCode {
	case 301, 302, 307, 308:
	default:
		return s, fmt.Errorf("redirect_status_code must be one of 301, 302, 307, or 308")
	}
	return s, nil
}

// GetOriginSettings loads the singleton system_origin_settings row.
func GetOriginSettings(ctx context.Context, pool *pgxpool.Pool) (OriginSettings, error) {
	if pool == nil {
		return DefaultOriginSettings(), nil
	}
	var s OriginSettings
	err := pool.QueryRow(ctx, `
		SELECT redirect_scheme, trust_forwarded_headers, redirect_status_code, updated_at
		FROM system_origin_settings
		WHERE id = 1
	`).Scan(&s.RedirectScheme, &s.TrustForwardedHeaders, &s.RedirectStatusCode, &s.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return DefaultOriginSettings(), nil
	}
	if err != nil {
		return DefaultOriginSettings(), fmt.Errorf("db.GetOriginSettings: %w", err)
	}
	return s, nil
}

// UpdateOriginSettings updates the singleton system_origin_settings row.
func UpdateOriginSettings(ctx context.Context, pool *pgxpool.Pool, in OriginSettings) (OriginSettings, error) {
	validated, err := ValidateOriginSettings(in)
	if err != nil {
		return OriginSettings{}, err
	}
	if pool == nil {
		validated.UpdatedAt = time.Now().UTC()
		return validated, nil
	}
	var out OriginSettings
	err = pool.QueryRow(ctx, `
		INSERT INTO system_origin_settings (id, redirect_scheme, trust_forwarded_headers, redirect_status_code, updated_at)
		VALUES (1, $1, $2, $3, now())
		ON CONFLICT (id) DO UPDATE
		SET redirect_scheme         = EXCLUDED.redirect_scheme,
		    trust_forwarded_headers = EXCLUDED.trust_forwarded_headers,
		    redirect_status_code    = EXCLUDED.redirect_status_code,
		    updated_at              = now()
		RETURNING redirect_scheme, trust_forwarded_headers, redirect_status_code, updated_at
	`, validated.RedirectScheme, validated.TrustForwardedHeaders, validated.RedirectStatusCode).
		Scan(&out.RedirectScheme, &out.TrustForwardedHeaders, &out.RedirectStatusCode, &out.UpdatedAt)
	if err != nil {
		return OriginSettings{}, fmt.Errorf("db.UpdateOriginSettings: %w", err)
	}
	return out, nil
}

// ListOriginDomains returns all configured system origin domains in priority order.
func ListOriginDomains(ctx context.Context, pool *pgxpool.Pool) ([]OriginDomain, error) {
	if pool == nil {
		return nil, nil
	}
	rows, err := pool.Query(ctx, `
		SELECT id, domain, note, position, enabled, created_at, updated_at
		FROM system_origin_domains
		ORDER BY position ASC, id ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("db.ListOriginDomains: %w", err)
	}
	defer rows.Close()

	var list []OriginDomain
	for rows.Next() {
		var d OriginDomain
		if err := rows.Scan(&d.ID, &d.Domain, &d.Note, &d.Position, &d.Enabled, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, fmt.Errorf("db.ListOriginDomains scan: %w", err)
		}
		list = append(list, d)
	}
	return list, rows.Err()
}

// ListEnabledOriginDomains returns only enabled system origin domains in priority order.
func ListEnabledOriginDomains(ctx context.Context, pool *pgxpool.Pool) ([]OriginDomain, error) {
	if pool == nil {
		return nil, nil
	}
	rows, err := pool.Query(ctx, `
		SELECT id, domain, note, position, enabled, created_at, updated_at
		FROM system_origin_domains
		WHERE enabled = true
		ORDER BY position ASC, id ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("db.ListEnabledOriginDomains: %w", err)
	}
	defer rows.Close()

	var list []OriginDomain
	for rows.Next() {
		var d OriginDomain
		if err := rows.Scan(&d.ID, &d.Domain, &d.Note, &d.Position, &d.Enabled, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, fmt.Errorf("db.ListEnabledOriginDomains scan: %w", err)
		}
		list = append(list, d)
	}
	return list, rows.Err()
}

// GetOriginDomain fetches a single system origin domain by ID.
func GetOriginDomain(ctx context.Context, pool *pgxpool.Pool, id int64) (OriginDomain, bool, error) {
	if pool == nil {
		return OriginDomain{}, false, nil
	}
	var d OriginDomain
	err := pool.QueryRow(ctx, `
		SELECT id, domain, note, position, enabled, created_at, updated_at
		FROM system_origin_domains
		WHERE id = $1
	`, id).Scan(&d.ID, &d.Domain, &d.Note, &d.Position, &d.Enabled, &d.CreatedAt, &d.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return OriginDomain{}, false, nil
	}
	if err != nil {
		return OriginDomain{}, false, fmt.Errorf("db.GetOriginDomain: %w", err)
	}
	return d, true, nil
}

// OriginDomainExists reports whether domain is already registered in system_origin_domains.
func OriginDomainExists(ctx context.Context, pool *pgxpool.Pool, domain string, excludeID int64) (bool, error) {
	if pool == nil || domain == "" {
		return false, nil
	}
	var exists bool
	err := pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM system_origin_domains
			WHERE domain = $1 AND ($2::bigint = 0 OR id <> $2)
		)
	`, domain, excludeID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("db.OriginDomainExists: %w", err)
	}
	return exists, nil
}

// AddOriginDomain appends a new system origin domain at the end of the priority list,
// verifying it does not collide with an existing system domain or active tenant domain.
func AddOriginDomain(ctx context.Context, pool *pgxpool.Pool, rawDomain, note string, enabled bool) (OriginDomain, error) {
	domain, err := NormalizeDomain(rawDomain)
	if err != nil {
		return OriginDomain{}, err
	}
	note = strings.TrimSpace(note)
	if len(note) > 200 {
		note = note[:200]
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return OriginDomain{}, fmt.Errorf("db.AddOriginDomain begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var existsInOrigins bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM system_origin_domains WHERE domain = $1)`, domain).Scan(&existsInOrigins); err != nil {
		return OriginDomain{}, fmt.Errorf("db.AddOriginDomain check origins: %w", err)
	}
	if existsInOrigins {
		return OriginDomain{}, fmt.Errorf("%w: %q is already in the system origin list", ErrOriginDomainConflict, domain)
	}

	var existsInTenants bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM tenants WHERE deleted_at IS NULL AND domain = $1)`, domain).Scan(&existsInTenants); err != nil {
		return OriginDomain{}, fmt.Errorf("db.AddOriginDomain check tenants: %w", err)
	}
	if existsInTenants {
		return OriginDomain{}, fmt.Errorf("%w: %q is already assigned to a tenant", ErrOriginDomainConflict, domain)
	}

	var nextPos int
	if err := tx.QueryRow(ctx, `SELECT COALESCE(MAX(position), 0) + 1 FROM system_origin_domains`).Scan(&nextPos); err != nil {
		return OriginDomain{}, fmt.Errorf("db.AddOriginDomain next pos: %w", err)
	}

	var out OriginDomain
	err = tx.QueryRow(ctx, `
		INSERT INTO system_origin_domains (domain, note, position, enabled)
		VALUES ($1, $2, $3, $4)
		RETURNING id, domain, note, position, enabled, created_at, updated_at
	`, domain, note, nextPos, enabled).Scan(
		&out.ID, &out.Domain, &out.Note, &out.Position, &out.Enabled, &out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		return OriginDomain{}, fmt.Errorf("db.AddOriginDomain insert: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return OriginDomain{}, fmt.Errorf("db.AddOriginDomain commit: %w", err)
	}
	return out, nil
}

// UpdateOriginDomain updates a system origin domain's note and enabled state.
func UpdateOriginDomain(ctx context.Context, pool *pgxpool.Pool, id int64, note string, enabled bool) (OriginDomain, bool, error) {
	note = strings.TrimSpace(note)
	if len(note) > 200 {
		note = note[:200]
	}
	var out OriginDomain
	err := pool.QueryRow(ctx, `
		UPDATE system_origin_domains
		SET note = $2, enabled = $3, updated_at = now()
		WHERE id = $1
		RETURNING id, domain, note, position, enabled, created_at, updated_at
	`, id, note, enabled).Scan(
		&out.ID, &out.Domain, &out.Note, &out.Position, &out.Enabled, &out.CreatedAt, &out.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return OriginDomain{}, false, nil
	}
	if err != nil {
		return OriginDomain{}, false, fmt.Errorf("db.UpdateOriginDomain: %w", err)
	}
	return out, true, nil
}

// MoveOriginDomain reorders a domain in system_origin_domains ("top", "up", or "down")
// and renormalizes positions 1..N atomically.
func MoveOriginDomain(ctx context.Context, pool *pgxpool.Pool, id int64, direction string) error {
	direction = strings.ToLower(strings.TrimSpace(direction))
	if direction != "top" && direction != "up" && direction != "down" {
		return fmt.Errorf("invalid move direction %q", direction)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("db.MoveOriginDomain begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	rows, err := tx.Query(ctx, `
		SELECT id FROM system_origin_domains
		ORDER BY position ASC, id ASC
		FOR UPDATE
	`)
	if err != nil {
		return fmt.Errorf("db.MoveOriginDomain lock: %w", err)
	}
	var ids []int64
	for rows.Next() {
		var cur int64
		if err := rows.Scan(&cur); err != nil {
			rows.Close()
			return err
		}
		ids = append(ids, cur)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	idx := -1
	for i, cur := range ids {
		if cur == id {
			idx = i
			break
		}
	}
	if idx == -1 {
		return pgx.ErrNoRows
	}

	switch direction {
	case "top":
		if idx > 0 {
			val := ids[idx]
			copy(ids[1:idx+1], ids[0:idx])
			ids[0] = val
		}
	case "up":
		if idx > 0 {
			ids[idx-1], ids[idx] = ids[idx], ids[idx-1]
		}
	case "down":
		if idx < len(ids)-1 {
			ids[idx+1], ids[idx] = ids[idx], ids[idx+1]
		}
	}

	for pos, curID := range ids {
		if _, err := tx.Exec(ctx, `
			UPDATE system_origin_domains SET position = $2, updated_at = now() WHERE id = $1
		`, curID, pos+1); err != nil {
			return fmt.Errorf("db.MoveOriginDomain update pos: %w", err)
		}
	}

	return tx.Commit(ctx)
}

// ReorderOriginDomains sets positions according to orderedIDs, keeping any omitted
// rows afterward in their existing relative order.
func ReorderOriginDomains(ctx context.Context, pool *pgxpool.Pool, orderedIDs []int64) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("db.ReorderOriginDomains begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	rows, err := tx.Query(ctx, `
		SELECT id FROM system_origin_domains
		ORDER BY position ASC, id ASC
		FOR UPDATE
	`)
	if err != nil {
		return fmt.Errorf("db.ReorderOriginDomains lock: %w", err)
	}
	var existing []int64
	existsMap := make(map[int64]bool)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		existing = append(existing, id)
		existsMap[id] = true
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	seen := make(map[int64]bool, len(existing))
	finalOrder := make([]int64, 0, len(existing))
	for _, id := range orderedIDs {
		if existsMap[id] && !seen[id] {
			seen[id] = true
			finalOrder = append(finalOrder, id)
		}
	}
	for _, id := range existing {
		if !seen[id] {
			seen[id] = true
			finalOrder = append(finalOrder, id)
		}
	}

	for pos, curID := range finalOrder {
		if _, err := tx.Exec(ctx, `
			UPDATE system_origin_domains SET position = $2, updated_at = now() WHERE id = $1
		`, curID, pos+1); err != nil {
			return fmt.Errorf("db.ReorderOriginDomains update pos: %w", err)
		}
	}

	return tx.Commit(ctx)
}

// DeleteOriginDomain removes a system origin domain and compacts remaining positions.
func DeleteOriginDomain(ctx context.Context, pool *pgxpool.Pool, id int64) (bool, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("db.DeleteOriginDomain begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	tag, err := tx.Exec(ctx, `DELETE FROM system_origin_domains WHERE id = $1`, id)
	if err != nil {
		return false, fmt.Errorf("db.DeleteOriginDomain delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return false, nil
	}

	rows, err := tx.Query(ctx, `
		SELECT id FROM system_origin_domains
		ORDER BY position ASC, id ASC
		FOR UPDATE
	`)
	if err != nil {
		return false, fmt.Errorf("db.DeleteOriginDomain reindex query: %w", err)
	}
	var ids []int64
	for rows.Next() {
		var cur int64
		if err := rows.Scan(&cur); err != nil {
			rows.Close()
			return false, err
		}
		ids = append(ids, cur)
	}
	rows.Close()

	for pos, curID := range ids {
		if _, err := tx.Exec(ctx, `UPDATE system_origin_domains SET position = $2 WHERE id = $1`, curID, pos+1); err != nil {
			return false, fmt.Errorf("db.DeleteOriginDomain reindex: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("db.DeleteOriginDomain commit: %w", err)
	}
	return true, nil
}
