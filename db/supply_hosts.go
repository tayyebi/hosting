package db

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"platform/providers"
)

var (
	// ErrNoAssignableInventory indicates no online, ready supply host for this
	// tenant has a free IP address available to fulfill a customer VM purchase.
	ErrNoAssignableInventory = errors.New("no online provisioned supply host with available assignable IPs in tenant inventory")

	// ErrHostHasActiveWorkloads prevents deleting a supply host while customer
	// resources or assigned IPs remain bound to it.
	ErrHostHasActiveWorkloads = errors.New("supply host still has assigned IPs or active customer resources")

	// ErrIPCurrentlyAssigned prevents deleting an IP from the pool while assigned to a VM.
	ErrIPCurrentlyAssigned = errors.New("IP address is currently assigned to a customer resource")
)

// SupplyHostRecord represents a tenant-scoped row in `supply_hosts` plus live inventory counters.
type SupplyHostRecord struct {
	ID                int64
	TenantID          int64
	Name              string
	HostnameOrIP      string
	SSHPort           int
	SSHUser           string
	AuthMethod        string
	SSHCredential     string
	DriverKind        string
	OwnershipType     string
	RegionID          *int64
	RegionSlug        string
	BridgeInterface   string
	StoragePoolPath   string
	CPUCores          int
	RAMMB             int
	DiskGB            int
	KVMAvailable      bool
	OSInfo            string
	HypervisorVersion string
	MonthlyCostCents  int64
	Currency          string
	VirtStackStatus   string // unprovisioned | provisioning | ready | failed
	InventoryStatus   string // pending | online | maintenance | offline
	ProvisionLog      string
	LastError         string
	ProvisionedAt     *time.Time
	LastCheckedAt     *time.Time
	CreatedBy         *int64
	CreatedAt         time.Time
	UpdatedAt         time.Time

	// Aggregated inventory counts
	TotalIPs     int
	AvailableIPs int
	AssignedIPs  int
	ActiveVMs    int
}

// ToProviderHost converts a DB SupplyHostRecord into providers.SupplyHost.
func (h SupplyHostRecord) ToProviderHost() providers.SupplyHost {
	return providers.SupplyHost{
		ID:               h.ID,
		Name:             h.Name,
		HostnameOrIP:     h.HostnameOrIP,
		SSHPort:          h.SSHPort,
		SSHUser:          h.SSHUser,
		AuthMethod:       providers.HostAuthMethod(h.AuthMethod),
		Credential:       h.SSHCredential,
		DriverKind:       providers.DriverKind(h.DriverKind),
		OwnershipType:    providers.OwnershipType(h.OwnershipType),
		RegionSlug:       h.RegionSlug,
		CPUCores:         h.CPUCores,
		RAMMB:            h.RAMMB,
		DiskGB:           h.DiskGB,
		MonthlyCostCents: h.MonthlyCostCents,
		Currency:         h.Currency,
		VirtStackStatus:  providers.VirtStackStatus(h.VirtStackStatus),
		InventoryStatus:  providers.InventoryStatus(h.InventoryStatus),
		BridgeInterface:  h.BridgeInterface,
		StoragePoolPath:  h.StoragePoolPath,
	}
}

// CreateSupplyHostInput carries fields to onboard a new tenant SupplyHost (such as a Linux Nested VPS).
type CreateSupplyHostInput struct {
	TenantID         int64
	Name             string
	HostnameOrIP     string
	SSHPort          int
	SSHUser          string
	AuthMethod       string
	SSHCredential    string
	DriverKind       string
	OwnershipType    string
	RegionSlug       string
	BridgeInterface  string
	StoragePoolPath  string
	CPUCores         int
	RAMMB            int
	DiskGB           int
	MonthlyCostCents int64
	Currency         string
	CreatedBy        int64
}

func ensureRegionID(ctx context.Context, pool *pgxpool.Pool, slug string) (*int64, string) {
	slug = strings.ToLower(strings.TrimSpace(slug))
	slug = strings.ReplaceAll(slug, " ", "-")
	if slug == "" {
		slug = "eu-central-1"
	}
	var id int64
	err := pool.QueryRow(ctx, `
		INSERT INTO regions (slug, display_name)
		VALUES ($1, $2)
		ON CONFLICT (slug) DO UPDATE SET slug = EXCLUDED.slug
		RETURNING id`,
		slug, slug).Scan(&id)
	if err != nil {
		return nil, slug
	}
	return &id, slug
}

// CreateSupplyHost registers a new tenant-scoped supply host in inventory.
func CreateSupplyHost(ctx context.Context, pool *pgxpool.Pool, in CreateSupplyHostInput) (SupplyHostRecord, error) {
	if in.TenantID <= 0 {
		return SupplyHostRecord{}, errors.New("tenant_id is required")
	}
	name := strings.TrimSpace(in.Name)
	hostAddr := strings.TrimSpace(in.HostnameOrIP)
	if name == "" || hostAddr == "" {
		return SupplyHostRecord{}, errors.New("name and hostname_or_ip are required")
	}
	port := in.SSHPort
	if port <= 0 || port > 65535 {
		port = 22
	}
	sshUser := strings.TrimSpace(in.SSHUser)
	if sshUser == "" {
		sshUser = "root"
	}
	authMethod := strings.TrimSpace(in.AuthMethod)
	switch authMethod {
	case "ssh_private_key", "ssh_password", "api_token":
	default:
		authMethod = "ssh_private_key"
	}
	driverKind := strings.TrimSpace(in.DriverKind)
	if driverKind == "" {
		driverKind = string(providers.DriverLinuxNestedQEMU)
	}
	ownership := strings.TrimSpace(in.OwnershipType)
	if ownership == "" {
		ownership = string(providers.OwnershipLeasedNestedVPS)
	}
	bridge := strings.TrimSpace(in.BridgeInterface)
	if bridge == "" {
		bridge = "virbr0"
	}
	storagePath := strings.TrimSpace(in.StoragePoolPath)
	if storagePath == "" {
		storagePath = "/var/lib/libvirt/images/platform"
	}
	currency := strings.ToUpper(strings.TrimSpace(in.Currency))
	if currency == "" {
		currency = "USD"
	}

	regionID, regionSlug := ensureRegionID(ctx, pool, in.RegionSlug)
	var createdBy *int64
	if in.CreatedBy > 0 {
		cb := in.CreatedBy
		createdBy = &cb
	}

	var id int64
	err := pool.QueryRow(ctx, `
		INSERT INTO supply_hosts (
			tenant_id, name, hostname_or_ip, ssh_port, ssh_user, auth_method, ssh_credential,
			driver_kind, ownership_type, region_id, region_slug,
			bridge_interface, storage_pool_path,
			cpu_cores, ram_mb, disk_gb,
			monthly_cost_cents, currency, created_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11,
			$12, $13,
			$14, $15, $16,
			$17, $18, $19
		) RETURNING id`,
		in.TenantID, name, hostAddr, port, sshUser, authMethod, in.SSHCredential,
		driverKind, ownership, regionID, regionSlug,
		bridge, storagePath,
		in.CPUCores, in.RAMMB, in.DiskGB,
		in.MonthlyCostCents, currency, createdBy,
	).Scan(&id)
	if err != nil {
		return SupplyHostRecord{}, fmt.Errorf("db.CreateSupplyHost: %w", err)
	}
	rec, _, err := GetTenantSupplyHostByID(ctx, pool, in.TenantID, id)
	return rec, err
}

const supplyHostSelectCols = `
	h.id, h.tenant_id, h.name, h.hostname_or_ip, h.ssh_port, h.ssh_user, h.auth_method, h.ssh_credential,
	h.driver_kind, h.ownership_type, h.region_id, h.region_slug,
	h.bridge_interface, h.storage_pool_path,
	h.cpu_cores, h.ram_mb, h.disk_gb, h.kvm_available,
	h.os_info, h.hypervisor_version,
	h.monthly_cost_cents, h.currency,
	h.virt_stack_status, h.inventory_status,
	h.provision_log, h.last_error,
	h.provisioned_at, h.last_checked_at, h.created_by,
	h.created_at, h.updated_at,
	COALESCE(ip_stats.total_ips, 0),
	COALESCE(ip_stats.avail_ips, 0),
	COALESCE(ip_stats.assigned_ips, 0),
	COALESCE(vm_stats.active_vms, 0)
`

const supplyHostJoins = `
	FROM supply_hosts h
	LEFT JOIN LATERAL (
		SELECT COUNT(*)::int AS total_ips,
		       COUNT(*) FILTER (WHERE status = 'available')::int AS avail_ips,
		       COUNT(*) FILTER (WHERE status = 'assigned')::int AS assigned_ips
		FROM supply_host_ips
		WHERE host_id = h.id
	) ip_stats ON true
	LEFT JOIN LATERAL (
		SELECT COUNT(*)::int AS active_vms
		FROM resources r
		WHERE r.supply_host_id = h.id
		  AND r.parent_resource_id IS NULL
		  AND r.deleted_at IS NULL
		  AND r.status IN ('active', 'provisioning')
	) vm_stats ON true
`

func scanSupplyHost(row pgx.Row) (SupplyHostRecord, error) {
	var h SupplyHostRecord
	err := row.Scan(
		&h.ID, &h.TenantID, &h.Name, &h.HostnameOrIP, &h.SSHPort, &h.SSHUser, &h.AuthMethod, &h.SSHCredential,
		&h.DriverKind, &h.OwnershipType, &h.RegionID, &h.RegionSlug,
		&h.BridgeInterface, &h.StoragePoolPath,
		&h.CPUCores, &h.RAMMB, &h.DiskGB, &h.KVMAvailable,
		&h.OSInfo, &h.HypervisorVersion,
		&h.MonthlyCostCents, &h.Currency,
		&h.VirtStackStatus, &h.InventoryStatus,
		&h.ProvisionLog, &h.LastError,
		&h.ProvisionedAt, &h.LastCheckedAt, &h.CreatedBy,
		&h.CreatedAt, &h.UpdatedAt,
		&h.TotalIPs, &h.AvailableIPs, &h.AssignedIPs, &h.ActiveVMs,
	)
	return h, err
}

// ListSupplyHostsByTenant returns all non-deleted supply hosts for a tenant with live IP & VM counts.
func ListSupplyHostsByTenant(ctx context.Context, pool *pgxpool.Pool, tenantID int64) ([]SupplyHostRecord, error) {
	rows, err := pool.Query(ctx,
		`SELECT `+supplyHostSelectCols+supplyHostJoins+` WHERE h.tenant_id = $1 AND h.deleted_at IS NULL ORDER BY h.id ASC`,
		tenantID)
	if err != nil {
		return nil, fmt.Errorf("db.ListSupplyHostsByTenant: %w", err)
	}
	defer rows.Close()

	var out []SupplyHostRecord
	for rows.Next() {
		h, err := scanSupplyHost(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, rows.Err()
}

// GetTenantSupplyHostByID fetches a single non-deleted supply host scoped to a tenant.
func GetTenantSupplyHostByID(ctx context.Context, pool *pgxpool.Pool, tenantID, id int64) (SupplyHostRecord, bool, error) {
	h, err := scanSupplyHost(pool.QueryRow(ctx,
		`SELECT `+supplyHostSelectCols+supplyHostJoins+` WHERE h.tenant_id = $1 AND h.id = $2 AND h.deleted_at IS NULL`,
		tenantID, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return SupplyHostRecord{}, false, nil
	}
	if err != nil {
		return SupplyHostRecord{}, false, fmt.Errorf("db.GetTenantSupplyHostByID: %w", err)
	}
	return h, true, nil
}

// GetSupplyHostByID fetches a single non-deleted supply host by primary key.
func GetSupplyHostByID(ctx context.Context, pool *pgxpool.Pool, id int64) (SupplyHostRecord, bool, error) {
	h, err := scanSupplyHost(pool.QueryRow(ctx,
		`SELECT `+supplyHostSelectCols+supplyHostJoins+` WHERE h.id = $1 AND h.deleted_at IS NULL`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return SupplyHostRecord{}, false, nil
	}
	if err != nil {
		return SupplyHostRecord{}, false, fmt.Errorf("db.GetSupplyHostByID: %w", err)
	}
	return h, true, nil
}

// MarkHostProvisioningStart transitions a host into `virt_stack_status = 'provisioning'`.
func MarkHostProvisioningStart(ctx context.Context, pool *pgxpool.Pool, hostID int64) error {
	_, err := pool.Exec(ctx, `
		UPDATE supply_hosts
		SET virt_stack_status = 'provisioning',
		    last_error        = '',
		    provision_log     = 'Starting automated virtualization stack installation...' || E'\n',
		    updated_at        = now()
		WHERE id = $1 AND deleted_at IS NULL`, hostID)
	return err
}

// MarkHostProvisioningComplete records successful virtualization stack bootstrap
// and automatically activates the host in `online` inventory.
func MarkHostProvisioningComplete(ctx context.Context, pool *pgxpool.Pool, hostID int64, cap providers.HostCapacity, logText string) error {
	_, err := pool.Exec(ctx, `
		UPDATE supply_hosts
		SET virt_stack_status  = 'ready',
		    inventory_status   = 'online',
		    cpu_cores          = CASE WHEN $2 > 0 THEN $2 ELSE cpu_cores END,
		    ram_mb             = CASE WHEN $3 > 0 THEN $3 ELSE ram_mb END,
		    disk_gb            = CASE WHEN $4 > 0 THEN $4 ELSE disk_gb END,
		    kvm_available      = $5,
		    os_info            = CASE WHEN $6 <> '' THEN $6 ELSE os_info END,
		    hypervisor_version = CASE WHEN $7 <> '' THEN $7 ELSE hypervisor_version END,
		    provision_log      = $8,
		    last_error         = '',
		    provisioned_at     = now(),
		    last_checked_at    = now(),
		    updated_at         = now()
		WHERE id = $1 AND deleted_at IS NULL`,
		hostID, cap.CPUCores, cap.RAMMB, cap.DiskGB, cap.KVMAvailable,
		cap.OSPrettyName, cap.HypervisorVersion, logText,
	)
	return err
}

// MarkHostProvisioningFailed records bootstrap failure details and log output.
func MarkHostProvisioningFailed(ctx context.Context, pool *pgxpool.Pool, hostID int64, errMsg, logText string) error {
	_, err := pool.Exec(ctx, `
		UPDATE supply_hosts
		SET virt_stack_status = 'failed',
		    inventory_status  = CASE WHEN inventory_status = 'online' THEN 'maintenance' ELSE inventory_status END,
		    last_error        = $2,
		    provision_log     = CASE WHEN $3 <> '' THEN $3 ELSE provision_log END,
		    last_checked_at   = now(),
		    updated_at        = now()
		WHERE id = $1 AND deleted_at IS NULL`,
		hostID, errMsg, logText)
	return err
}

// UpdateHostProbeCapacity stores results from an SSH capacity probe.
func UpdateHostProbeCapacity(ctx context.Context, pool *pgxpool.Pool, hostID int64, cap providers.HostCapacity) error {
	_, err := pool.Exec(ctx, `
		UPDATE supply_hosts
		SET cpu_cores          = CASE WHEN $2 > 0 THEN $2 ELSE cpu_cores END,
		    ram_mb             = CASE WHEN $3 > 0 THEN $3 ELSE ram_mb END,
		    disk_gb            = CASE WHEN $4 > 0 THEN $4 ELSE disk_gb END,
		    kvm_available      = $5,
		    os_info            = CASE WHEN $6 <> '' THEN $6 ELSE os_info END,
		    hypervisor_version = CASE WHEN $7 <> '' THEN $7 ELSE hypervisor_version END,
		    last_error         = '',
		    last_checked_at    = now(),
		    updated_at         = now()
		WHERE id = $1 AND deleted_at IS NULL`,
		hostID, cap.CPUCores, cap.RAMMB, cap.DiskGB, cap.KVMAvailable,
		cap.OSPrettyName, cap.HypervisorVersion)
	return err
}

// SetTenantHostInventoryStatus updates scheduling eligibility (`online`, `maintenance`, `offline`, `pending`) for a tenant's host.
func SetTenantHostInventoryStatus(ctx context.Context, pool *pgxpool.Pool, tenantID, hostID int64, status string) error {
	switch status {
	case "pending", "online", "maintenance", "offline":
	default:
		return fmt.Errorf("invalid inventory_status %q", status)
	}
	tag, err := pool.Exec(ctx, `
		UPDATE supply_hosts
		SET inventory_status = $3, updated_at = now()
		WHERE tenant_id = $1 AND id = $2 AND deleted_at IS NULL`, tenantID, hostID, status)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// DeleteTenantSupplyHost soft-deletes a tenant's supply host if it has no assigned IPs or active VMs.
func DeleteTenantSupplyHost(ctx context.Context, pool *pgxpool.Pool, tenantID, hostID int64) error {
	if _, found, err := GetTenantSupplyHostByID(ctx, pool, tenantID, hostID); err != nil {
		return err
	} else if !found {
		return pgx.ErrNoRows
	}

	var busyCount int
	if err := pool.QueryRow(ctx, `
		SELECT (
			(SELECT COUNT(*) FROM supply_host_ips WHERE tenant_id = $1 AND host_id = $2 AND status = 'assigned') +
			(SELECT COUNT(*) FROM resources WHERE tenant_id = $1 AND supply_host_id = $2 AND deleted_at IS NULL AND status IN ('active','provisioning'))
		)::int`, tenantID, hostID).Scan(&busyCount); err != nil {
		return err
	}
	if busyCount > 0 {
		return ErrHostHasActiveWorkloads
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `DELETE FROM supply_host_ips WHERE tenant_id = $1 AND host_id = $2`, tenantID, hostID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE supply_hosts SET deleted_at = now(), inventory_status = 'offline', updated_at = now()
		WHERE tenant_id = $1 AND id = $2 AND deleted_at IS NULL`, tenantID, hostID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// ── Manual Assignable IP Pool (`supply_host_ips`) ────────────────────────────

// SupplyHostIP represents one IPv4/IPv6 entry manually added to a tenant's supply host pool.
type SupplyHostIP struct {
	ID                   int64
	TenantID             int64
	HostID               int64
	HostName             string
	IPAddress            string
	CIDRPrefix           int
	Gateway              string
	MACAddress           string
	IPVersion            int
	MonthlyCostCents     int64
	RetailPriceCents     int64
	Currency             string
	Status               string // available | reserved | assigned | quarantined
	AssignedResourceID   *int64
	AssignedVMResourceID *int64
	AssignedVMName       string
	AssignedAt           *time.Time
	Notes                string
	CreatedAt            time.Time
}

// ManualIPEntry represents one validated IP parsed from manual operator input.
type ManualIPEntry struct {
	IPAddress        string
	CIDRPrefix       int
	IPVersion        int
	Gateway          string
	MACAddress       string
	MonthlyCostCents int64
	RetailPriceCents int64
	Notes            string
}

// ParseManualIPList parses newline-, comma-, or space-separated IP addresses or
// CIDR notations (e.g. "198.51.100.10", "198.51.100.11/24") into validated entries.
func ParseManualIPList(raw string, defaultCIDR int, gateway, mac string, monthlyCostCents, retailPriceCents int64, notes string) ([]ManualIPEntry, error) {
	replacer := strings.NewReplacer(",", "\n", ";", "\n", "\r", "\n", "\t", " ")
	lines := strings.Split(replacer.Replace(raw), "\n")

	if defaultCIDR <= 0 || defaultCIDR > 128 {
		defaultCIDR = 24
	}
	gw := strings.TrimSpace(gateway)
	if gw != "" {
		if _, err := netip.ParseAddr(gw); err != nil {
			return nil, fmt.Errorf("invalid gateway IP %q", gw)
		}
	}
	cleanMAC := strings.TrimSpace(mac)
	if cleanMAC != "" {
		hw, err := net.ParseMAC(cleanMAC)
		if err != nil {
			return nil, fmt.Errorf("invalid MAC address %q", cleanMAC)
		}
		cleanMAC = hw.String()
	}

	seen := make(map[string]bool)
	var out []ManualIPEntry

	for _, line := range lines {
		if idx := strings.IndexByte(line, '#'); idx >= 0 {
			line = line[:idx]
		}
		for _, tok := range strings.Fields(line) {
			tok = strings.TrimSpace(tok)
			if tok == "" {
				continue
			}
			var addr netip.Addr
			cidr := defaultCIDR
			if strings.Contains(tok, "/") {
				pfx, err := netip.ParsePrefix(tok)
				if err != nil {
					return nil, fmt.Errorf("invalid IP/CIDR %q: %w", tok, err)
				}
				addr = pfx.Addr()
				cidr = pfx.Bits()
			} else {
				a, err := netip.ParseAddr(tok)
				if err != nil {
					return nil, fmt.Errorf("invalid IP address %q: %w", tok, err)
				}
				addr = a
			}
			ver := 4
			if addr.Is6() {
				ver = 6
				if cidr < 1 || cidr > 128 {
					cidr = 64
				}
			} else if cidr > 32 {
				cidr = 24
			}
			ipStr := addr.String()
			if seen[ipStr] {
				continue
			}
			seen[ipStr] = true
			out = append(out, ManualIPEntry{
				IPAddress:        ipStr,
				CIDRPrefix:       cidr,
				IPVersion:        ver,
				Gateway:          gw,
				MACAddress:       cleanMAC,
				MonthlyCostCents: monthlyCostCents,
				RetailPriceCents: retailPriceCents,
				Notes:            strings.TrimSpace(notes),
			})
		}
	}
	if len(out) == 0 {
		return nil, errors.New("at least one valid IP address is required")
	}
	return out, nil
}

// AddTenantSupplyHostIPs inserts a batch of manually-entered assignable IPs onto a tenant's server.
func AddTenantSupplyHostIPs(ctx context.Context, pool *pgxpool.Pool, tenantID, hostID int64, entries []ManualIPEntry) ([]SupplyHostIP, error) {
	if _, found, err := GetTenantSupplyHostByID(ctx, pool, tenantID, hostID); err != nil {
		return nil, err
	} else if !found {
		return nil, pgx.ErrNoRows
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	for _, e := range entries {
		retail := e.RetailPriceCents
		if retail < 0 {
			retail = 0
		}
		_, err := tx.Exec(ctx, `
			INSERT INTO supply_host_ips (
				tenant_id, host_id, ip_address, cidr_prefix, gateway, mac_address,
				ip_version, monthly_cost_cents, retail_price_cents, status, notes
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 'available', $10)
			ON CONFLICT (ip_address) DO NOTHING`,
			tenantID, hostID, e.IPAddress, e.CIDRPrefix, e.Gateway, e.MACAddress,
			e.IPVersion, e.MonthlyCostCents, retail, e.Notes,
		)
		if err != nil {
			return nil, fmt.Errorf("db.AddTenantSupplyHostIPs(%s): %w", e.IPAddress, err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return ListTenantSupplyHostIPs(ctx, pool, tenantID, hostID)
}

// ListTenantSupplyHostIPs returns all IPs in `supply_host_ips` for a tenant (filtered to hostID when > 0).
func ListTenantSupplyHostIPs(ctx context.Context, pool *pgxpool.Pool, tenantID, hostID int64) ([]SupplyHostIP, error) {
	query := `
		SELECT ip.id, ip.tenant_id, ip.host_id, h.name, ip.ip_address, ip.cidr_prefix,
		       ip.gateway, ip.mac_address, ip.ip_version,
		       ip.monthly_cost_cents, ip.retail_price_cents, ip.currency,
		       ip.status, ip.assigned_resource_id, ip.assigned_vm_resource_id,
		       COALESCE(vm_res.name, ''), ip.assigned_at, ip.notes, ip.created_at
		FROM supply_host_ips ip
		JOIN supply_hosts h ON h.id = ip.host_id AND h.deleted_at IS NULL
		LEFT JOIN resources vm_res ON vm_res.id = ip.assigned_vm_resource_id
		WHERE ip.tenant_id = $1 AND h.tenant_id = $1
	`
	args := []any{tenantID}
	if hostID > 0 {
		query += ` AND ip.host_id = $2`
		args = append(args, hostID)
	}
	query += ` ORDER BY ip.host_id ASC, ip.id ASC`

	rows, err := pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("db.ListTenantSupplyHostIPs: %w", err)
	}
	defer rows.Close()

	var out []SupplyHostIP
	for rows.Next() {
		var item SupplyHostIP
		if err := rows.Scan(
			&item.ID, &item.TenantID, &item.HostID, &item.HostName, &item.IPAddress, &item.CIDRPrefix,
			&item.Gateway, &item.MACAddress, &item.IPVersion,
			&item.MonthlyCostCents, &item.RetailPriceCents, &item.Currency,
			&item.Status, &item.AssignedResourceID, &item.AssignedVMResourceID,
			&item.AssignedVMName, &item.AssignedAt, &item.Notes, &item.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

// DeleteTenantSupplyHostIP removes an unassigned IP from a tenant's host pool.
func DeleteTenantSupplyHostIP(ctx context.Context, pool *pgxpool.Pool, tenantID, ipID int64) error {
	var status string
	err := pool.QueryRow(ctx, `SELECT status FROM supply_host_ips WHERE tenant_id = $1 AND id = $2`, tenantID, ipID).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		return pgx.ErrNoRows
	}
	if err != nil {
		return err
	}
	if status == "assigned" {
		return ErrIPCurrentlyAssigned
	}
	_, err = pool.Exec(ctx, `DELETE FROM supply_host_ips WHERE tenant_id = $1 AND id = $2 AND status <> 'assigned'`, tenantID, ipID)
	return err
}

// ── Tenant Provisioned Guest VMs View ────────────────────────────────────────

// TenantProvisionedVM summarizes a customer VM provisioned onto one of a tenant's supply hosts.
type TenantProvisionedVM struct {
	ResourceID int64
	Name       string
	Status     string
	IPAddress  string
	HostName   string
	RegionSlug string
	Distro     string
	Version    string
	VCPUs      int
	MemoryGB   int
	DiskGB     int
	CreatedAt  time.Time
}

// ListTenantProvisionedVMs returns non-deleted customer compute_vm resources for a tenant.
func ListTenantProvisionedVMs(ctx context.Context, pool *pgxpool.Pool, tenantID int64) ([]TenantProvisionedVM, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(lvm.public_ip, cvm.ip_address, ''),
		       COALESCE(h.name, ''),
		       COALESCE(h.region_slug, ''),
		       COALESCE(lvm.distro, 'Linux'),
		       COALESCE(lvm.version, ''),
		       COALESCE(lvm.vcpus, cvm.vcpus, 0),
		       COALESCE(lvm.memory_gb, cvm.memory_gb::int, 0),
		       COALESCE(lvm.disk_gb, cvm.disk_gb, 0),
		       r.created_at
		FROM resources r
		LEFT JOIN workspace_linux_vms lvm ON lvm.resource_id = r.id
		LEFT JOIN workspace_compute_vms cvm ON cvm.resource_id = r.id
		LEFT JOIN supply_hosts h ON h.id = r.supply_host_id
		WHERE r.tenant_id = $1
		  AND r.kind = 'compute_vm'
		  AND r.parent_resource_id IS NULL
		  AND r.deleted_at IS NULL
		ORDER BY r.id DESC`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("db.ListTenantProvisionedVMs: %w", err)
	}
	defer rows.Close()

	var out []TenantProvisionedVM
	for rows.Next() {
		var vm TenantProvisionedVM
		if err := rows.Scan(
			&vm.ResourceID, &vm.Name, &vm.Status, &vm.IPAddress,
			&vm.HostName, &vm.RegionSlug, &vm.Distro, &vm.Version,
			&vm.VCPUs, &vm.MemoryGB, &vm.DiskGB, &vm.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, vm)
	}
	return out, rows.Err()
}

// ── Customer VM Scheduling + Atomic Assignable IP Allocation ─────────────────

// CustomerVMOrderRequest describes a customer VM purchase to be scheduled onto
// an `online` + `ready` supply host belonging to the same tenant and assigned an IP from its pool.
type CustomerVMOrderRequest struct {
	TenantID           int64
	WorkspaceID        int64
	CreatedBy          int64
	Name               string
	VMClass            string // "general" | "linux_vm"
	InstanceType       string
	Distro             string
	Version            string
	OSImage            string
	SSHKey             string
	VCPUs              int
	MemoryGB           int
	DiskGB             int
	PreferredRegion    string
	SKUCode            string
	VMSupplyCostCents  int64
	VMRetailPriceCents int64
}

// AllocatedVMOrder holds the persisted parent VM resource, child IP resource,
// chosen SupplyHostRecord, and allocated SupplyHostIP ready for driver execution.
type AllocatedVMOrder struct {
	VMResourceID       int64
	IPResourceID       int64
	Host               SupplyHostRecord
	IP                 SupplyHostIP
	RegionSlug         string
	TotalRetailCents   int64
	MonthlyCostDisplay string
}

// AllocateHostAndIPForVM atomically selects an `online` + `ready` supply host
// belonging to `req.TenantID` with at least one `available` IP (`FOR UPDATE SKIP LOCKED`),
// creates the parent VM `resources` row (`workspace_compute_vms` + `workspace_linux_vms`),
// creates the child payable IP `resources` row (`parent_resource_id`,
// `workspace_assigned_ips`), and marks the IP `assigned`.
func AllocateHostAndIPForVM(ctx context.Context, pool *pgxpool.Pool, req CustomerVMOrderRequest) (AllocatedVMOrder, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return AllocatedVMOrder{}, err
	}
	defer tx.Rollback(ctx)

	prefRegion := strings.ToLower(strings.TrimSpace(req.PreferredRegion))

	// Lock one available IP on an online + ready host owned by req.TenantID, preferring exact region match first.
	var ip SupplyHostIP
	var hostID int64
	err = tx.QueryRow(ctx, `
		SELECT ip.id, ip.tenant_id, ip.host_id, ip.ip_address, ip.cidr_prefix,
		       ip.gateway, ip.mac_address, ip.ip_version,
		       ip.monthly_cost_cents, ip.retail_price_cents, ip.currency
		FROM supply_host_ips ip
		JOIN supply_hosts h ON h.id = ip.host_id
		WHERE ip.tenant_id = $1
		  AND h.tenant_id = $1
		  AND ip.status = 'available'
		  AND ip.ip_version = 4
		  AND h.deleted_at IS NULL
		  AND h.inventory_status = 'online'
		  AND h.virt_stack_status = 'ready'
		ORDER BY
		  CASE WHEN $2 <> '' AND h.region_slug = $2 THEN 0 ELSE 1 END ASC,
		  ip.id ASC
		LIMIT 1
		FOR UPDATE OF ip SKIP LOCKED`,
		req.TenantID, prefRegion,
	).Scan(
		&ip.ID, &ip.TenantID, &hostID, &ip.IPAddress, &ip.CIDRPrefix,
		&ip.Gateway, &ip.MACAddress, &ip.IPVersion,
		&ip.MonthlyCostCents, &ip.RetailPriceCents, &ip.Currency,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return AllocatedVMOrder{}, ErrNoAssignableInventory
	}
	if err != nil {
		return AllocatedVMOrder{}, fmt.Errorf("db.AllocateHostAndIPForVM lock ip: %w", err)
	}

	host, err := scanSupplyHost(tx.QueryRow(ctx,
		`SELECT `+supplyHostSelectCols+supplyHostJoins+` WHERE h.id = $1`, hostID))
	if err != nil {
		return AllocatedVMOrder{}, fmt.Errorf("db.AllocateHostAndIPForVM load host: %w", err)
	}

	var createdByPtr *int64
	if req.CreatedBy > 0 {
		cb := req.CreatedBy
		createdByPtr = &cb
	}

	skuCode := req.SKUCode
	if skuCode == "" {
		skuCode = fmt.Sprintf("vps.nested.%dc-%dg", req.VCPUs, req.MemoryGB)
	}

	totalRetailCents := req.VMRetailPriceCents + ip.RetailPriceCents
	monthlyDisplay := fmt.Sprintf("$%d.%02d", totalRetailCents/100, totalRetailCents%100)

	// 1. Insert root payable VM resource
	var vmResID int64
	err = tx.QueryRow(ctx, `
		INSERT INTO resources (
			tenant_id, workspace_id, region_id, supply_host_id,
			kind, name, status, sku_code,
			supply_cost_cents, retail_price_cents, created_by
		) VALUES ($1, $2, $3, $4, 'compute_vm', $5, 'provisioning', $6, $7, $8, $9)
		RETURNING id`,
		req.TenantID, req.WorkspaceID, host.RegionID, host.ID,
		req.Name, skuCode, req.VMSupplyCostCents, req.VMRetailPriceCents, createdByPtr,
	).Scan(&vmResID)
	if err != nil {
		return AllocatedVMOrder{}, fmt.Errorf("db.AllocateHostAndIPForVM insert vm resource: %w", err)
	}

	instType := req.InstanceType
	if instType == "" {
		instType = fmt.Sprintf("%dc-%dgb", req.VCPUs, req.MemoryGB)
	}
	distro := req.Distro
	if distro == "" {
		distro = "Ubuntu"
	}
	ver := req.Version
	if ver == "" {
		ver = "24.04 LTS"
	}
	osImg := req.OSImage
	if osImg == "" {
		osImg = strings.ToLower(distro) + "-" + ver
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO workspace_compute_vms (
			resource_id, vm_class, instance_type, vcpus, memory_gb,
			ip_address, os_image, disk_gb
		) VALUES ($1, 'general', $2, $3, $4, $5, $6, $7)`,
		vmResID, instType, req.VCPUs, float64(req.MemoryGB),
		ip.IPAddress, osImg, req.DiskGB,
	); err != nil {
		return AllocatedVMOrder{}, fmt.Errorf("db.AllocateHostAndIPForVM insert workspace_compute_vms: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO workspace_linux_vms (
			resource_id, distro, version, vcpus, memory_gb, disk_gb,
			public_ip, private_ip, ssh_key, uptime, monthly_cost, vm_status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, '', $8, '0m', $9, 'provisioning')`,
		vmResID, distro, ver, req.VCPUs, req.MemoryGB, req.DiskGB,
		ip.IPAddress, req.SSHKey, monthlyDisplay,
	); err != nil {
		return AllocatedVMOrder{}, fmt.Errorf("db.AllocateHostAndIPForVM insert workspace_linux_vms: %w", err)
	}

	// 2. Insert child payable IPv4 resource (`parent_resource_id = vmResID`)
	var ipResID int64
	err = tx.QueryRow(ctx, `
		INSERT INTO resources (
			tenant_id, workspace_id, region_id, parent_resource_id, supply_host_id,
			kind, name, status, sku_code, external_ref,
			supply_cost_cents, retail_price_cents, created_by
		) VALUES ($1, $2, $3, $4, $5, 'network_ipv4', $6, 'active', 'addon.ipv4.routed', $7, $8, $9, $10)
		RETURNING id`,
		req.TenantID, req.WorkspaceID, host.RegionID, vmResID, host.ID,
		"ipv4-"+ip.IPAddress, ip.IPAddress,
		ip.MonthlyCostCents, ip.RetailPriceCents, createdByPtr,
	).Scan(&ipResID)
	if err != nil {
		return AllocatedVMOrder{}, fmt.Errorf("db.AllocateHostAndIPForVM insert ip resource: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO workspace_assigned_ips (resource_id, supply_ip_id, ip_address, cidr_prefix, gateway)
		VALUES ($1, $2, $3, $4, $5)`,
		ipResID, ip.ID, ip.IPAddress, ip.CIDRPrefix, ip.Gateway,
	); err != nil {
		return AllocatedVMOrder{}, fmt.Errorf("db.AllocateHostAndIPForVM insert workspace_assigned_ips: %w", err)
	}

	// 3. Mark the manual IP pool row assigned
	if _, err := tx.Exec(ctx, `
		UPDATE supply_host_ips
		SET status                  = 'assigned',
		    assigned_resource_id    = $2,
		    assigned_vm_resource_id = $3,
		    assigned_at             = now(),
		    updated_at              = now()
		WHERE id = $1`,
		ip.ID, ipResID, vmResID,
	); err != nil {
		return AllocatedVMOrder{}, fmt.Errorf("db.AllocateHostAndIPForVM mark ip assigned: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return AllocatedVMOrder{}, err
	}

	ip.HostID = host.ID
	ip.HostName = host.Name
	ip.Status = "assigned"
	return AllocatedVMOrder{
		VMResourceID:       vmResID,
		IPResourceID:       ipResID,
		Host:               host,
		IP:                 ip,
		RegionSlug:         host.RegionSlug,
		TotalRetailCents:   totalRetailCents,
		MonthlyCostDisplay: monthlyDisplay,
	}, nil
}

// FinalizeCustomerVMProvisioning updates the VM resource status after the hypervisor
// driver returns. On driver error, it releases the allocated IP back to the pool
// and marks the failed provisioning attempt cleanly.
func FinalizeCustomerVMProvisioning(ctx context.Context, pool *pgxpool.Pool, vmResourceID int64, externalRef string, driverErr error) error {
	if driverErr == nil {
		tx, err := pool.Begin(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback(ctx)
		if _, err := tx.Exec(ctx, `
			UPDATE resources
			SET status = 'active', external_ref = $2, updated_at = now()
			WHERE id = $1`, vmResourceID, externalRef); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
			UPDATE workspace_linux_vms
			SET vm_status = 'running', uptime = 'Just started'
			WHERE resource_id = $1`, vmResourceID); err != nil {
			return err
		}
		return tx.Commit(ctx)
	}

	_, _, _ = ReleaseCustomerVMAndIP(ctx, pool, vmResourceID)
	return nil
}

// ReleaseCustomerVMAndIP soft-deletes a customer VM and its child payable IP
// resource and returns the IP to `status = 'available'` in `supply_host_ips`.
func ReleaseCustomerVMAndIP(ctx context.Context, pool *pgxpool.Pool, vmResourceID int64) (*SupplyHostRecord, string, error) {
	var hostID *int64
	var externalRef string
	err := pool.QueryRow(ctx, `
		SELECT supply_host_id, COALESCE(external_ref, '')
		FROM resources
		WHERE id = $1 AND deleted_at IS NULL`, vmResourceID,
	).Scan(&hostID, &externalRef)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, "", nil
	}
	if err != nil {
		return nil, "", err
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, "", err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		UPDATE supply_host_ips
		SET status                  = 'available',
		    assigned_resource_id    = NULL,
		    assigned_vm_resource_id = NULL,
		    assigned_at             = NULL,
		    updated_at              = now()
		WHERE assigned_vm_resource_id = $1`, vmResourceID); err != nil {
		return nil, "", err
	}

	if _, err := tx.Exec(ctx, `
		UPDATE resources
		SET deleted_at = now(), status = 'deleted', updated_at = now()
		WHERE id = $1 OR parent_resource_id = $1`, vmResourceID); err != nil {
		return nil, "", err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, "", err
	}

	if hostID != nil && *hostID > 0 {
		if h, ok, hErr := GetSupplyHostByID(ctx, pool, *hostID); hErr == nil && ok {
			return &h, externalRef, nil
		}
	}
	return nil, externalRef, nil
}
