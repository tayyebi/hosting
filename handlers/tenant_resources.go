package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"platform/db"
	"platform/providers"
)

// ── In-memory fallback store (used only when DB pool is nil in unit tests) ───

type TenantResource struct {
	ID           string
	Name         string
	ResourceType string
	Provider     string
	Region       string
	Status       string // "online"|"degraded"|"offline"|"maintenance"|"provisioning"
	IPAddress    string
	Created      string
	Updated      string
}

type tenantResourceStoreType struct {
	mu        sync.RWMutex
	resources map[string][]TenantResource
}

var tenantResourceStore = &tenantResourceStoreType{
	resources: map[string][]TenantResource{},
}

func (s *tenantResourceStoreType) list(slug string) []TenantResource {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.resources[slug]
	out := make([]TenantResource, len(src))
	copy(out, src)
	return out
}

func (s *tenantResourceStoreType) updateStatus(slug, id, status string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, r := range s.resources[slug] {
		if r.ID == id {
			s.resources[slug][i].Status = status
			s.resources[slug][i].Updated = time.Now().Format(time.RFC3339)
			return
		}
	}
}

func (s *tenantResourceStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.resources[slug]
	for i, r := range list {
		if r.ID == id {
			s.resources[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

// ── Category metadata for the Resource Tree ──────────────────────────────────

type resourceCategoryMeta struct {
	Key        string
	Label      string
	BadgeLabel string
	BadgeClass string
	DriverKind string // non-empty for compute supply hosts
}

var resourceCategoryMap = map[string]resourceCategoryMeta{
	"linux-vps":  {Key: "linux-vps", Label: "Linux Nested VPS (QEMU/KVM)", BadgeLabel: "Linux VPS", BadgeClass: "res-type-badge--linux", DriverKind: string(providers.DriverLinuxNestedQEMU)},
	"proxmox":    {Key: "proxmox", Label: "Proxmox Clusters / Nodes", BadgeLabel: "Proxmox", BadgeClass: "res-type-badge--proxmox", DriverKind: string(providers.DriverProxmox)},
	"esxi":       {Key: "esxi", Label: "ESXi Clusters / Nodes", BadgeLabel: "ESXi", BadgeClass: "res-type-badge--esxi", DriverKind: string(providers.DriverVMwareVSphere)},
	"ilo":        {Key: "ilo", Label: "ILO Servers", BadgeLabel: "ILO / BMC", BadgeClass: "res-type-badge--ilo"},
	"linux-vm":   {Key: "linux-vm", Label: "Nested Linux VMs", BadgeLabel: "Linux VM", BadgeClass: "res-type-badge--linux"},
	"win-vm":     {Key: "win-vm", Label: "Nested Windows VMs", BadgeLabel: "Windows VM", BadgeClass: "res-type-badge--win"},
	"cdn-pop":    {Key: "cdn-pop", Label: "CDN PoP Sites", BadgeLabel: "CDN PoP", BadgeClass: "res-type-badge--cdn"},
	"mysql":      {Key: "mysql", Label: "MySQL Clusters", BadgeLabel: "MySQL", BadgeClass: "res-type-badge--mysql"},
	"postgres":   {Key: "postgres", Label: "PostgreSQL Clusters", BadgeLabel: "PostgreSQL", BadgeClass: "res-type-badge--postgres"},
	"redis":      {Key: "redis", Label: "Redis Clusters", BadgeLabel: "Redis", BadgeClass: "res-type-badge--redis"},
	"ceph":       {Key: "ceph", Label: "Ceph Clusters", BadgeLabel: "Ceph", BadgeClass: "res-type-badge--storage"},
	"nfs":        {Key: "nfs", Label: "NFS Servers", BadgeLabel: "NFS", BadgeClass: "res-type-badge--storage"},
	"san":        {Key: "san", Label: "SAN Arrays", BadgeLabel: "SAN", BadgeClass: "res-type-badge--storage"},
	"minio":      {Key: "minio", Label: "Object Storage (MinIO)", BadgeLabel: "MinIO", BadgeClass: "res-type-badge--storage"},
	"ip-pool":    {Key: "ip-pool", Label: "IP Address Pools", BadgeLabel: "IPv4 / IPv6", BadgeClass: "res-type-badge--net"},
	"lb":         {Key: "lb", Label: "Load Balancers", BadgeLabel: "Load Balancer", BadgeClass: "res-type-badge--net"},
	"firewall":   {Key: "firewall", Label: "Firewalls / UTM", BadgeLabel: "Firewall", BadgeClass: "res-type-badge--net"},
	"vpn":        {Key: "vpn", Label: "VPN Gateways", BadgeLabel: "VPN", BadgeClass: "res-type-badge--net"},
	"dns":        {Key: "dns", Label: "DNS Servers", BadgeLabel: "DNS", BadgeClass: "res-type-badge--net"},
	"kafka":      {Key: "kafka", Label: "Kafka Clusters", BadgeLabel: "Kafka", BadgeClass: "res-type-badge--queue"},
	"rabbitmq":   {Key: "rabbitmq", Label: "RabbitMQ Clusters", BadgeLabel: "RabbitMQ", BadgeClass: "res-type-badge--queue"},
	"nats":       {Key: "nats", Label: "NATS Clusters", BadgeLabel: "NATS", BadgeClass: "res-type-badge--queue"},
	"prometheus": {Key: "prometheus", Label: "Prometheus / VictoriaMetrics", BadgeLabel: "Metrics", BadgeClass: "res-type-badge--obs"},
	"grafana":    {Key: "grafana", Label: "Grafana / Kibana", BadgeLabel: "Dashboards", BadgeClass: "res-type-badge--obs"},
	"elk":        {Key: "elk", Label: "ELK Stacks", BadgeLabel: "ELK", BadgeClass: "res-type-badge--obs"},
	"k8s":        {Key: "k8s", Label: "Kubernetes Clusters", BadgeLabel: "Kubernetes", BadgeClass: "res-type-badge--k8s"},
	"docker":     {Key: "docker", Label: "Docker Clusters", BadgeLabel: "Docker", BadgeClass: "res-type-badge--docker"},
}

func categoryForDriverKind(driver string) resourceCategoryMeta {
	switch providers.DriverKind(driver) {
	case providers.DriverProxmox:
		return resourceCategoryMap["proxmox"]
	case providers.DriverVMwareVSphere:
		return resourceCategoryMap["esxi"]
	default:
		return resourceCategoryMap["linux-vps"]
	}
}

func categoryForResourceType(rt string) resourceCategoryMeta {
	k := strings.ToLower(strings.TrimSpace(rt))
	if k == "cdn" {
		k = "cdn-pop"
	}
	if m, ok := resourceCategoryMap[k]; ok {
		return m
	}
	return resourceCategoryMeta{
		Key:        k,
		Label:      rt,
		BadgeLabel: rt,
		BadgeClass: "res-type-badge--storage",
	}
}

// ── View Models for `/t/{slug}/resources/` ───────────────────────────────────

type TenantIPView struct {
	db.SupplyHostIP
	CostDisplay   string
	RetailDisplay string
}

type TenantHostView struct {
	db.SupplyHostRecord
	CategoryKey        string
	BadgeLabel         string
	BadgeClass         string
	MonthlyCostDisplay string
	RAMGBDisplay       string
	IPs                []TenantIPView
}

type TenantInventoryRow struct {
	RowKind        string // "supply_host" | "tenant_resource" | "customer_vm"
	NumericID      int64
	ID             string
	CategoryKey    string
	Name           string
	SubLabel       string
	BadgeLabel     string
	BadgeClass     string
	Status         string // normalized: online | degraded | offline | maintenance | provisioning
	StatusLabel    string
	StatusDotClass string
	StatusTxtClass string
	HostOrIP       string
	Region         string
	CPUDisplay     string
	RAMDisplay     string
	Tags           []string
	DetailJSON     string
}

type tenantDetailPayload struct {
	RowKind           string         `json:"rowKind"`
	ID                int64          `json:"id"`
	Name              string         `json:"name"`
	Meta              string         `json:"meta"`
	StatusLabel       string         `json:"statusLabel"`
	StatusDotClass    string         `json:"statusDotClass"`
	Region            string         `json:"region"`
	HostOrIP          string         `json:"hostOrIP"`
	DriverKind        string         `json:"driverKind,omitempty"`
	VirtStackStatus   string         `json:"virtStackStatus,omitempty"`
	InventoryStatus   string         `json:"inventoryStatus,omitempty"`
	Specs             string         `json:"specs"`
	Bridge            string         `json:"bridge,omitempty"`
	StoragePool       string         `json:"storagePool,omitempty"`
	SupplyCost        string         `json:"supplyCost,omitempty"`
	KVMAvailable      bool           `json:"kvmAvailable"`
	Created           string         `json:"created"`
	ProvisionLog      string         `json:"provisionLog,omitempty"`
	LastError         string         `json:"lastError,omitempty"`
	AvailableIPs      int            `json:"availableIPs"`
	TotalIPs          int            `json:"totalIPs"`
	ActiveVMs         int            `json:"activeVMs"`
	IPs               []detailIPItem `json:"ips,omitempty"`
}

type detailIPItem struct {
	ID             int64  `json:"id"`
	IPAddress      string `json:"ip"`
	CIDRPrefix     int    `json:"cidr"`
	Gateway        string `json:"gateway"`
	Status         string `json:"status"`
	RetailDisplay  string `json:"retail"`
	AssignedVMName string `json:"assignedVM"`
}

type CapacityPoolRow struct {
	Name          string
	TotalDisplay  string
	AllocDisplay  string
	UtilPct       int
	BarClass      string
	HeadroomLabel string
}

type resourcesInnerData struct {
	TenantSlug    string
	ActiveView    string
	Search        string
	StatusFilter  string
	FlashSuccess  string
	FlashError    string

	// Global stats strip
	TotalCount        int
	OnlineCount       int
	DegradedCount     int
	OfflineCount      int
	MaintenanceCount  int
	ProvisioningCount int
	AlertCount        int

	// Tree counts (both leaf keys and top-level category totals)
	GroupCounts map[string]int
	CatCounts   map[string]int

	// Unified inventory rows + supply hosts + assignable IP pool
	Rows         []TenantInventoryRow
	Hosts        []TenantHostView
	IPs          []TenantIPView
	SelectedHost *TenantHostView
	Regions      []string

	// Live capacity planning metrics
	TotalHostVCPUs   int
	AllocVCPUs       int
	ComputeUtilPct   int
	TotalHostRAMGB   int
	AllocRAMGB       int
	RAMUtilPct       int
	TotalHostDiskGB  int
	AllocDiskGB      int
	DiskUtilPct      int
	TotalIPs         int
	AssignedIPs      int
	AvailableIPs     int
	IPUtilPct        int
	CapacityRows     []CapacityPoolRow
}

func statusPresentation(rawStatus, virtStatus string) (norm, label, dotClass, txtClass string) {
	if virtStatus == "provisioning" || rawStatus == "provisioning" {
		return "provisioning", "Provisioning", "res-dot--prov", ""
	}
	if virtStatus == "failed" {
		return "degraded", "Failed", "res-dot--danger", "res-status--danger"
	}
	switch strings.ToLower(strings.TrimSpace(rawStatus)) {
	case "online", "active", "running", "healthy":
		return "online", "Online", "res-dot--ok", ""
	case "degraded", "warning":
		return "degraded", "Degraded", "res-dot--warn", "res-status--warn"
	case "maintenance":
		return "maintenance", "Maintenance", "res-dot--maint", "res-status--maint"
	case "pending":
		return "provisioning", "Pending Setup", "res-dot--warn", "res-status--warn"
	default:
		return "offline", "Offline", "res-dot--danger", "res-status--danger"
	}
}

func utilBarClass(pct int) string {
	if pct >= 85 {
		return "inv-cap-util__fill--danger"
	}
	if pct >= 65 {
		return "inv-cap-util__fill--warn"
	}
	return "inv-cap-util__fill--ok"
}

func parseDollarToCents(s string) int64 {
	s = strings.TrimSpace(strings.TrimPrefix(s, "$"))
	if s == "" {
		return 0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil || f < 0 {
		return 0
	}
	return int64(math.Round(f * 100))
}

func buildTenantResourcesData(ctx context.Context, tenant Tenant, activeView, search, statusFilter string, selectedHostID int64, flashOK, flashErr string) resourcesInnerData {
	if activeView == "" {
		activeView = "all"
	}
	d := resourcesInnerData{
		TenantSlug:   tenant.Slug,
		ActiveView:   activeView,
		Search:       search,
		StatusFilter: statusFilter,
		FlashSuccess: flashOK,
		FlashError:   flashErr,
		GroupCounts:  map[string]int{},
		CatCounts:    map[string]int{},
	}

	regionSet := map[string]bool{}
	addRegion := func(r string) {
		r = strings.TrimSpace(r)
		if r != "" && !regionSet[r] {
			regionSet[r] = true
			d.Regions = append(d.Regions, r)
		}
	}
	tallyStatus := func(norm string, hasErr bool) {
		d.TotalCount++
		switch norm {
		case "online":
			d.OnlineCount++
		case "degraded":
			d.DegradedCount++
			d.AlertCount++
		case "maintenance":
			d.MaintenanceCount++
		case "provisioning":
			d.ProvisioningCount++
		default:
			d.OfflineCount++
		}
		if hasErr && norm != "degraded" {
			d.AlertCount++
		}
	}

	if Pools != nil && tenant.DBID > 0 {
		hosts, err := db.ListSupplyHostsByTenant(ctx, Pools.Admin, tenant.DBID)
		if err != nil {
			log.Printf("resourcesHandler: ListSupplyHostsByTenant: %v", err)
		}
		ips, err := db.ListTenantSupplyHostIPs(ctx, Pools.Admin, tenant.DBID, 0)
		if err != nil {
			log.Printf("resourcesHandler: ListTenantSupplyHostIPs: %v", err)
		}
		genRes, err := db.ListTenantResources(ctx, Pools.Admin, tenant.DBID, "", "")
		if err != nil {
			log.Printf("resourcesHandler: ListTenantResources: %v", err)
		}
		vms, err := db.ListTenantProvisionedVMs(ctx, Pools.Admin, tenant.DBID)
		if err != nil {
			log.Printf("resourcesHandler: ListTenantProvisionedVMs: %v", err)
		}

		ipsByHost := map[int64][]TenantIPView{}
		for _, ip := range ips {
			ipv := TenantIPView{
				SupplyHostIP:  ip,
				CostDisplay:   formatCentsUSD(ip.MonthlyCostCents),
				RetailDisplay: formatCentsUSD(ip.RetailPriceCents),
			}
			d.IPs = append(d.IPs, ipv)
			ipsByHost[ip.HostID] = append(ipsByHost[ip.HostID], ipv)
			d.TotalIPs++
			if ip.Status == "available" {
				d.AvailableIPs++
			} else if ip.Status == "assigned" {
				d.AssignedIPs++
			}
		}
		d.GroupCounts["ip-pool"] = len(d.IPs)

		var totalRAMMB int
		for _, h := range hosts {
			meta := categoryForDriverKind(h.DriverKind)
			hv := TenantHostView{
				SupplyHostRecord:   h,
				CategoryKey:        meta.Key,
				BadgeLabel:         meta.BadgeLabel,
				BadgeClass:         meta.BadgeClass,
				MonthlyCostDisplay: formatCentsUSD(h.MonthlyCostCents),
				RAMGBDisplay:       fmt.Sprintf("%.1f GB", float64(h.RAMMB)/1024.0),
				IPs:                ipsByHost[h.ID],
			}
			d.Hosts = append(d.Hosts, hv)
			d.GroupCounts[meta.Key]++
			addRegion(h.RegionSlug)

			d.TotalHostVCPUs += h.CPUCores
			totalRAMMB += h.RAMMB
			d.TotalHostDiskGB += h.DiskGB

			norm, stLabel, dotCls, txtCls := statusPresentation(h.InventoryStatus, h.VirtStackStatus)
			tallyStatus(norm, h.LastError != "")

			cpuDisp := "—"
			if h.CPUCores > 0 {
				cpuDisp = fmt.Sprintf("%d vCPU", h.CPUCores)
			}
			ramDisp := "—"
			if h.RAMMB > 0 {
				ramDisp = fmt.Sprintf("%.0f GB", math.Max(1, float64(h.RAMMB)/1024.0))
			}

			tags := []string{"virt:" + h.VirtStackStatus, fmt.Sprintf("%d/%d IPs free", h.AvailableIPs, h.TotalIPs)}
			if h.KVMAvailable {
				tags = append(tags, "kvm")
			}

			sub := h.DriverKind
			if h.OSInfo != "" {
				sub = h.OSInfo
			}
			if h.RegionSlug != "" {
				sub = sub + " · " + h.RegionSlug
			}

			var detailIPs []detailIPItem
			for _, hip := range hv.IPs {
				detailIPs = append(detailIPs, detailIPItem{
					ID:             hip.ID,
					IPAddress:      hip.IPAddress,
					CIDRPrefix:     hip.CIDRPrefix,
					Gateway:        hip.Gateway,
					Status:         hip.Status,
					RetailDisplay:  hip.RetailDisplay,
					AssignedVMName: hip.AssignedVMName,
				})
			}
			specs := "Unprobed"
			if h.CPUCores > 0 || h.RAMMB > 0 || h.DiskGB > 0 {
				specs = fmt.Sprintf("%d vCPU · %.0f GB RAM · %d GB Disk", h.CPUCores, float64(h.RAMMB)/1024.0, h.DiskGB)
			}
			payload, _ := json.Marshal(tenantDetailPayload{
				RowKind:         "supply_host",
				ID:              h.ID,
				Name:            h.Name,
				Meta:            sub,
				StatusLabel:     stLabel,
				StatusDotClass:  dotCls,
				Region:          h.RegionSlug,
				HostOrIP:        fmt.Sprintf("%s:%d", h.HostnameOrIP, h.SSHPort),
				DriverKind:      h.DriverKind,
				VirtStackStatus: h.VirtStackStatus,
				InventoryStatus: h.InventoryStatus,
				Specs:           specs,
				Bridge:          h.BridgeInterface,
				StoragePool:     h.StoragePoolPath,
				SupplyCost:      hv.MonthlyCostDisplay + "/mo",
				KVMAvailable:    h.KVMAvailable,
				Created:         h.CreatedAt.UTC().Format("2006-01-02 15:04 UTC"),
				ProvisionLog:    h.ProvisionLog,
				LastError:       h.LastError,
				AvailableIPs:    h.AvailableIPs,
				TotalIPs:        h.TotalIPs,
				ActiveVMs:       h.ActiveVMs,
				IPs:             detailIPs,
			})

			d.Rows = append(d.Rows, TenantInventoryRow{
				RowKind:        "supply_host",
				NumericID:      h.ID,
				ID:             fmt.Sprintf("host-%d", h.ID),
				CategoryKey:    meta.Key,
				Name:           h.Name,
				SubLabel:       sub,
				BadgeLabel:     meta.BadgeLabel,
				BadgeClass:     meta.BadgeClass,
				Status:         norm,
				StatusLabel:    stLabel,
				StatusDotClass: dotCls,
				StatusTxtClass: txtCls,
				HostOrIP:       h.HostnameOrIP,
				Region:         h.RegionSlug,
				CPUDisplay:     cpuDisp,
				RAMDisplay:     ramDisp,
				Tags:           tags,
				DetailJSON:     string(payload),
			})
		}
		d.TotalHostRAMGB = totalRAMMB / 1024

		for i := range d.Hosts {
			if selectedHostID > 0 && d.Hosts[i].ID == selectedHostID {
				d.SelectedHost = &d.Hosts[i]
				break
			}
		}

		// Provisioned customer VMs on this tenant's hosts -> appear under `linux-vm`
		for _, vm := range vms {
			meta := resourceCategoryMap["linux-vm"]
			d.GroupCounts["linux-vm"]++
			addRegion(vm.RegionSlug)
			d.AllocVCPUs += vm.VCPUs
			d.AllocRAMGB += vm.MemoryGB
			d.AllocDiskGB += vm.DiskGB

			norm, stLabel, dotCls, txtCls := statusPresentation(vm.Status, "")
			tallyStatus(norm, false)

			sub := strings.TrimSpace(vm.Distro + " " + vm.Version)
			if vm.HostName != "" {
				sub += " · host " + vm.HostName
			}
			specs := fmt.Sprintf("%d vCPU · %d GB RAM · %d GB Disk", vm.VCPUs, vm.MemoryGB, vm.DiskGB)
			payload, _ := json.Marshal(tenantDetailPayload{
				RowKind:        "customer_vm",
				ID:             vm.ResourceID,
				Name:           vm.Name,
				Meta:           sub,
				StatusLabel:    stLabel,
				StatusDotClass: dotCls,
				Region:         vm.RegionSlug,
				HostOrIP:       vm.IPAddress,
				Specs:          specs,
				Created:        vm.CreatedAt.UTC().Format("2006-01-02 15:04 UTC"),
			})

			d.Rows = append(d.Rows, TenantInventoryRow{
				RowKind:        "customer_vm",
				NumericID:      vm.ResourceID,
				ID:             fmt.Sprintf("vm-%d", vm.ResourceID),
				CategoryKey:    meta.Key,
				Name:           vm.Name,
				SubLabel:       sub,
				BadgeLabel:     meta.BadgeLabel,
				BadgeClass:     meta.BadgeClass,
				Status:         norm,
				StatusLabel:    stLabel,
				StatusDotClass: dotCls,
				StatusTxtClass: txtCls,
				HostOrIP:       vm.IPAddress,
				Region:         vm.RegionSlug,
				CPUDisplay:     fmt.Sprintf("%d vCPU", vm.VCPUs),
				RAMDisplay:     fmt.Sprintf("%d GB", vm.MemoryGB),
				Tags:           []string{"guest-vm", vm.HostName},
				DetailJSON:     string(payload),
			})
		}

		// Generic tenant_resources (all placeholder categories in the Resource Tree)
		for _, gr := range genRes {
			meta := categoryForResourceType(gr.ResourceType)
			d.GroupCounts[meta.Key]++
			addRegion(gr.Region)

			norm, stLabel, dotCls, txtCls := statusPresentation(gr.Status, "")
			tallyStatus(norm, false)

			cpuDisp := "—"
			if gr.CPUCores > 0 {
				cpuDisp = fmt.Sprintf("%d vCPU", gr.CPUCores)
			}
			ramDisp := "—"
			if gr.RAMMB > 0 {
				ramDisp = fmt.Sprintf("%.0f GB", math.Max(1, float64(gr.RAMMB)/1024.0))
			}

			var tags []string
			for _, t := range strings.Split(gr.Tags, ",") {
				t = strings.TrimSpace(t)
				if t != "" {
					tags = append(tags, t)
				}
			}
			if len(tags) == 0 {
				tags = []string{meta.Key}
			}
			sub := gr.Provider
			if sub == "" {
				sub = meta.Label
			}
			if gr.Region != "" {
				sub += " · " + gr.Region
			}

			specs := meta.Label
			if gr.APIURL != "" {
				specs += " · " + gr.APIURL
			}
			payload, _ := json.Marshal(tenantDetailPayload{
				RowKind:         "tenant_resource",
				ID:              gr.ID,
				Name:            gr.Name,
				Meta:            sub,
				StatusLabel:     stLabel,
				StatusDotClass:  dotCls,
				Region:          gr.Region,
				HostOrIP:        gr.IPAddress,
				InventoryStatus: gr.Status,
				Specs:           specs,
				Created:         gr.CreatedAt.UTC().Format("2006-01-02 15:04 UTC"),
			})

			d.Rows = append(d.Rows, TenantInventoryRow{
				RowKind:        "tenant_resource",
				NumericID:      gr.ID,
				ID:             fmt.Sprintf("%d", gr.ID),
				CategoryKey:    meta.Key,
				Name:           gr.Name,
				SubLabel:       sub,
				BadgeLabel:     meta.BadgeLabel,
				BadgeClass:     meta.BadgeClass,
				Status:         norm,
				StatusLabel:    stLabel,
				StatusDotClass: dotCls,
				StatusTxtClass: txtCls,
				HostOrIP:       gr.IPAddress,
				Region:         gr.Region,
				CPUDisplay:     cpuDisp,
				RAMDisplay:     ramDisp,
				Tags:           tags,
				DetailJSON:     string(payload),
			})
		}
	} else {
		for _, lr := range tenantResourceStore.list(tenant.Slug) {
			meta := categoryForResourceType(lr.ResourceType)
			d.GroupCounts[meta.Key]++
			norm, stLabel, dotCls, txtCls := statusPresentation(lr.Status, "")
			tallyStatus(norm, false)
			d.Rows = append(d.Rows, TenantInventoryRow{
				RowKind:        "tenant_resource",
				ID:             lr.ID,
				CategoryKey:    meta.Key,
				Name:           lr.Name,
				SubLabel:       lr.Provider + " · " + lr.Region,
				BadgeLabel:     meta.BadgeLabel,
				BadgeClass:     meta.BadgeClass,
				Status:         norm,
				StatusLabel:    stLabel,
				StatusDotClass: dotCls,
				StatusTxtClass: txtCls,
				HostOrIP:       lr.IPAddress,
				Region:         lr.Region,
				CPUDisplay:     "—",
				RAMDisplay:     "—",
				Tags:           []string{meta.Key},
				DetailJSON:     "{}",
			})
		}
	}

	// Category rollup counts for left tree headers
	d.CatCounts["servers"] = d.GroupCounts["linux-vps"] + d.GroupCounts["proxmox"] + d.GroupCounts["esxi"] + d.GroupCounts["ilo"] + d.GroupCounts["linux-vm"] + d.GroupCounts["win-vm"]
	d.CatCounts["cdn"] = d.GroupCounts["cdn-pop"]
	d.CatCounts["databases"] = d.GroupCounts["mysql"] + d.GroupCounts["postgres"] + d.GroupCounts["redis"]
	d.CatCounts["storage"] = d.GroupCounts["ceph"] + d.GroupCounts["nfs"] + d.GroupCounts["san"] + d.GroupCounts["minio"]
	d.CatCounts["networking"] = d.GroupCounts["ip-pool"] + d.GroupCounts["lb"] + d.GroupCounts["firewall"] + d.GroupCounts["vpn"] + d.GroupCounts["dns"]
	d.CatCounts["queues"] = d.GroupCounts["kafka"] + d.GroupCounts["rabbitmq"] + d.GroupCounts["nats"]
	d.CatCounts["observability"] = d.GroupCounts["prometheus"] + d.GroupCounts["grafana"] + d.GroupCounts["elk"]
	d.CatCounts["containers"] = d.GroupCounts["k8s"] + d.GroupCounts["docker"]

	// Capacity utilization percentages
	if d.TotalHostVCPUs > 0 {
		d.ComputeUtilPct = int(math.Min(100, math.Round(float64(d.AllocVCPUs)*100.0/float64(d.TotalHostVCPUs))))
	}
	if d.TotalHostRAMGB > 0 {
		d.RAMUtilPct = int(math.Min(100, math.Round(float64(d.AllocRAMGB)*100.0/float64(d.TotalHostRAMGB))))
	}
	if d.TotalHostDiskGB > 0 {
		d.DiskUtilPct = int(math.Min(100, math.Round(float64(d.AllocDiskGB)*100.0/float64(d.TotalHostDiskGB))))
	}
	if d.TotalIPs > 0 {
		d.IPUtilPct = int(math.Min(100, math.Round(float64(d.AssignedIPs)*100.0/float64(d.TotalIPs))))
	}

	d.CapacityRows = []CapacityPoolRow{
		{
			Name:          "vCPU (supply hosts)",
			TotalDisplay:  fmt.Sprintf("%d cores", d.TotalHostVCPUs),
			AllocDisplay:  fmt.Sprintf("%d cores", d.AllocVCPUs),
			UtilPct:       d.ComputeUtilPct,
			BarClass:      utilBarClass(d.ComputeUtilPct),
			HeadroomLabel: fmt.Sprintf("%d cores free", maxInt(0, d.TotalHostVCPUs-d.AllocVCPUs)),
		},
		{
			Name:          "RAM (supply hosts)",
			TotalDisplay:  fmt.Sprintf("%d GB", d.TotalHostRAMGB),
			AllocDisplay:  fmt.Sprintf("%d GB", d.AllocRAMGB),
			UtilPct:       d.RAMUtilPct,
			BarClass:      utilBarClass(d.RAMUtilPct),
			HeadroomLabel: fmt.Sprintf("%d GB free", maxInt(0, d.TotalHostRAMGB-d.AllocRAMGB)),
		},
		{
			Name:          "Storage Pool (libvirt / disk)",
			TotalDisplay:  fmt.Sprintf("%d GB", d.TotalHostDiskGB),
			AllocDisplay:  fmt.Sprintf("%d GB", d.AllocDiskGB),
			UtilPct:       d.DiskUtilPct,
			BarClass:      utilBarClass(d.DiskUtilPct),
			HeadroomLabel: fmt.Sprintf("%d GB free", maxInt(0, d.TotalHostDiskGB-d.AllocDiskGB)),
		},
		{
			Name:          "Assignable IP Pool",
			TotalDisplay:  fmt.Sprintf("%d addresses", d.TotalIPs),
			AllocDisplay:  fmt.Sprintf("%d assigned", d.AssignedIPs),
			UtilPct:       d.IPUtilPct,
			BarClass:      utilBarClass(d.IPUtilPct),
			HeadroomLabel: fmt.Sprintf("%d available", d.AvailableIPs),
		},
	}

	// Optional server-side filtering if query params ?q= or ?status= are passed
	if search != "" || statusFilter != "" {
		qLower := strings.ToLower(strings.TrimSpace(search))
		stLower := strings.ToLower(strings.TrimSpace(statusFilter))
		var filtered []TenantInventoryRow
		for _, row := range d.Rows {
			if stLower != "" && stLower != "all" && row.Status != stLower {
				continue
			}
			if qLower != "" {
				hay := strings.ToLower(row.Name + " " + row.SubLabel + " " + row.HostOrIP + " " + row.Region + " " + row.BadgeLabel)
				if !strings.Contains(hay, qLower) {
					continue
				}
			}
			filtered = append(filtered, row)
		}
		d.Rows = filtered
	}

	return d
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ── Auto-Provisioning Orchestrator ───────────────────────────────────────────

func runHostAutoProvisioning(ctx context.Context, host db.SupplyHostRecord, actorID int64) (providers.BootstrapReport, error) {
	if Pools == nil {
		return providers.BootstrapReport{}, errors.New("database unavailable")
	}
	drv, err := Drivers.Compute(providers.DriverKind(host.DriverKind))
	if err != nil {
		_ = db.MarkHostProvisioningFailed(ctx, Pools.Admin, host.ID, err.Error(), "")
		return providers.BootstrapReport{}, err
	}

	_ = db.MarkHostProvisioningStart(ctx, Pools.Admin, host.ID)
	_ = db.WriteSystemLog(ctx, Pools.Admin, "info", "infra-provisioner",
		fmt.Sprintf("Starting virtualization stack auto-provisioning on tenant %d host %s (%s)", host.TenantID, host.Name, host.HostnameOrIP),
		map[string]any{"tenant_id": host.TenantID, "host_id": host.ID, "driver": host.DriverKind})

	report, bootErr := drv.BootstrapHost(ctx, host.ToProviderHost())
	if bootErr != nil {
		_ = db.MarkHostProvisioningFailed(ctx, Pools.Admin, host.ID, bootErr.Error(), report.LogOutput)
		_ = db.WriteSystemLog(ctx, Pools.Admin, "error", "infra-provisioner",
			fmt.Sprintf("Auto-provisioning failed on host %s: %v", host.Name, bootErr),
			map[string]any{"tenant_id": host.TenantID, "host_id": host.ID})
		db.InsertAuditEvent(ctx, Pools.Admin, actorID, "tenant", "tenant.resource.bootstrap.failed", "warn",
			map[string]any{"tenant_id": host.TenantID, "host_id": host.ID, "name": host.Name}, bootErr.Error(), "", "")
		return report, bootErr
	}

	if err := db.MarkHostProvisioningComplete(ctx, Pools.Admin, host.ID, report.Capacity, report.LogOutput); err != nil {
		return report, err
	}
	_ = db.WriteSystemLog(ctx, Pools.Admin, "info", "infra-provisioner",
		fmt.Sprintf("Host %s provisioned and added to tenant %d online inventory (%d vCPUs, %d MB RAM, KVM=%v)",
			host.Name, host.TenantID, report.Capacity.CPUCores, report.Capacity.RAMMB, report.Capacity.KVMAvailable),
		map[string]any{"tenant_id": host.TenantID, "host_id": host.ID})
	db.InsertAuditEvent(ctx, Pools.Admin, actorID, "tenant", "tenant.resource.bootstrap.ready", "info",
		map[string]any{
			"tenant_id":     host.TenantID,
			"host_id":       host.ID,
			"name":          host.Name,
			"cpu_cores":     report.Capacity.CPUCores,
			"ram_mb":        report.Capacity.RAMMB,
			"disk_gb":       report.Capacity.DiskGB,
			"kvm_available": report.Capacity.KVMAvailable,
		}, "", "", "")
	return report, nil
}

// ── Page Renderer ────────────────────────────────────────────────────────────

func renderResourcesPage(w http.ResponseWriter, r *http.Request, tenant Tenant, user User, d resourcesInnerData) {
	innerTmpl, err := template.ParseFiles(filepath.Join("templates", "tenant", "resources_content.html"))
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	var buf bytes.Buffer
	if err := innerTmpl.ExecuteTemplate(&buf, "tenant-page-content", d); err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	shellPaths := []string{
		filepath.Join("templates", "tenant", "base.html"),
		filepath.Join("templates", "tenant", "shell.html"),
	}
	shellTmpl, err := template.ParseFiles(shellPaths...)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	data := TenantShellData{
		User:       user,
		Tenant:     tenant,
		Nav:        tenantNav(tenant.Slug),
		NavGroups:  tenantNavGroups(tenant.Slug),
		ActiveNav:  "resources",
		PageTitle:  "Resources",
		IsSysAdmin: user.Role == RoleSysAdmin,
		Content:    template.HTML(buf.String()),
	}
	if err := shellTmpl.ExecuteTemplate(w, "tenant-base", data); err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
	}
}

func tenantResourcesRedirect(w http.ResponseWriter, r *http.Request, slug string, params url.Values) {
	target := "/t/" + slug + "/resources/"
	if enc := params.Encode(); enc != "" {
		target += "?" + enc
	}
	http.Redirect(w, r, target, http.StatusSeeOther)
}

// ── Tenant HTML Handlers (`/t/{slug}/resources/*`) ───────────────────────────

func resourcesHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := getSession(r)
	tenant, _ := tenantFromContext(r)

	q := r.URL.Query().Get("q")
	status := r.URL.Query().Get("status")
	view := r.URL.Query().Get("view")
	selHostID, _ := strconv.ParseInt(r.URL.Query().Get("host"), 10, 64)
	flashOK := r.URL.Query().Get("ok")
	flashErr := r.URL.Query().Get("err")

	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()

	d := buildTenantResourcesData(ctx, tenant, view, q, status, selHostID, flashOK, flashErr)
	renderResourcesPage(w, r, tenant, user, d)
}

func resourceOnboardFormHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := getSession(r)
	tenant, _ := tenantFromContext(r)
	if Pools == nil || tenant.DBID <= 0 {
		tenantResourcesRedirect(w, r, tenant.Slug, url.Values{"err": {"Database unavailable"}})
		return
	}
	if err := r.ParseForm(); err != nil {
		tenantResourcesRedirect(w, r, tenant.Slug, url.Values{"err": {"Invalid form payload"}})
		return
	}

	category := strings.ToLower(strings.TrimSpace(r.FormValue("category")))
	if category == "" {
		category = "linux-vps"
	}
	meta, known := resourceCategoryMap[category]
	if !known {
		meta = categoryForResourceType(category)
	}

	name := strings.TrimSpace(r.FormValue("name"))
	hostOrIP := strings.TrimSpace(r.FormValue("hostname_or_ip"))
	if name == "" || hostOrIP == "" {
		tenantResourcesRedirect(w, r, tenant.Slug, url.Values{"err": {"Display name and Hostname / IP are required"}})
		return
	}

	region := strings.TrimSpace(r.FormValue("region"))
	if region == "" {
		region = "eu-central-1"
	}
	authMethod := strings.TrimSpace(r.FormValue("auth_method"))
	credential := r.FormValue("credential")
	cpuCores, _ := strconv.Atoi(r.FormValue("cpu_cores"))
	ramGB, _ := strconv.Atoi(r.FormValue("ram_gb"))
	diskGB, _ := strconv.Atoi(r.FormValue("disk_gb"))

	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()

	// Compute supply hosts (`linux-vps`, `proxmox`, `esxi`) go into `supply_hosts`
	if meta.DriverKind != "" {
		sshPort, _ := strconv.Atoi(r.FormValue("ssh_port"))
		costCents := parseDollarToCents(r.FormValue("monthly_cost_usd"))

		host, err := db.CreateSupplyHost(ctx, Pools.Admin, db.CreateSupplyHostInput{
			TenantID:         tenant.DBID,
			Name:             name,
			HostnameOrIP:     hostOrIP,
			SSHPort:          sshPort,
			SSHUser:          r.FormValue("ssh_user"),
			AuthMethod:       authMethod,
			SSHCredential:    credential,
			DriverKind:       meta.DriverKind,
			OwnershipType:    r.FormValue("ownership_type"),
			RegionSlug:       region,
			BridgeInterface:  r.FormValue("bridge_interface"),
			StoragePoolPath:  r.FormValue("storage_pool_path"),
			CPUCores:         cpuCores,
			RAMMB:            ramGB * 1024,
			DiskGB:           diskGB,
			MonthlyCostCents: costCents,
			CreatedBy:        user.ID,
		})
		if err != nil {
			tenantResourcesRedirect(w, r, tenant.Slug, url.Values{
				"view": {meta.Key},
				"err":  {"Failed to onboard host: " + err.Error()},
			})
			return
		}

		db.InsertAuditEvent(ctx, Pools.Admin, user.ID, string(user.Role), "tenant.host.onboarded", "info",
			map[string]any{"tenant_id": tenant.DBID, "host_id": host.ID, "name": host.Name, "driver": host.DriverKind}, "", "", clientIP(r))

		if rawIPs := strings.TrimSpace(r.FormValue("initial_ips")); rawIPs != "" {
			initCIDR, _ := strconv.Atoi(r.FormValue("initial_cidr_prefix"))
			initRetail := parseDollarToCents(r.FormValue("initial_ip_retail_usd"))
			if r.FormValue("initial_ip_retail_usd") == "" {
				initRetail = 200
			}
			if ipEntries, ipErr := db.ParseManualIPList(rawIPs, initCIDR, r.FormValue("initial_gateway"), "", 0, initRetail, "initial pool"); ipErr == nil && len(ipEntries) > 0 {
				_, _ = db.AddTenantSupplyHostIPs(ctx, Pools.Admin, tenant.DBID, host.ID, ipEntries)
			}
		}

		if r.FormValue("auto_provision") == "on" && meta.DriverKind == string(providers.DriverLinuxNestedQEMU) {
			go func(h db.SupplyHostRecord, uid int64) {
				bgCtx, bgCancel := context.WithTimeout(context.Background(), 15*time.Minute)
				defer bgCancel()
				_, _ = runHostAutoProvisioning(bgCtx, h, uid)
			}(host, user.ID)

			tenantResourcesRedirect(w, r, tenant.Slug, url.Values{
				"view": {meta.Key},
				"host": {strconv.FormatInt(host.ID, 10)},
				"ok":   {"Host onboarded and virtualization stack auto-provisioning started"},
			})
			return
		}

		tenantResourcesRedirect(w, r, tenant.Slug, url.Values{
			"view": {meta.Key},
			"host": {strconv.FormatInt(host.ID, 10)},
			"ok":   {meta.BadgeLabel + " onboarded into tenant resources"},
		})
		return
	}

	// All other Resource Tree categories persist in `tenant_resources`
	_, err := db.CreateTenantResourceFull(ctx, Pools.Admin, db.CreateTenantResourceInput{
		TenantID:     tenant.DBID,
		Name:         name,
		ResourceType: meta.Key,
		Provider:     meta.Label,
		Region:       region,
		Status:       "online",
		IPAddress:    hostOrIP,
		APIURL:       r.FormValue("api_url"),
		AuthMethod:   authMethod,
		Credential:   credential,
		Tags:         r.FormValue("tags"),
		CPUCores:     cpuCores,
		RAMMB:        ramGB * 1024,
		DiskGB:       diskGB,
	})
	if err != nil {
		tenantResourcesRedirect(w, r, tenant.Slug, url.Values{
			"view": {meta.Key},
			"err":  {"Failed to onboard resource: " + err.Error()},
		})
		return
	}

	tenantResourcesRedirect(w, r, tenant.Slug, url.Values{
		"view": {meta.Key},
		"ok":   {"Resource " + name + " onboarded"},
	})
}

func tenantHostProbeFormHandler(w http.ResponseWriter, r *http.Request) {
	tenant, _ := tenantFromContext(r)
	hostID, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	ctx, cancel := context.WithTimeout(r.Context(), 25*time.Second)
	defer cancel()

	host, ok, err := db.GetTenantSupplyHostByID(ctx, Pools.Admin, tenant.DBID, hostID)
	if err != nil || !ok {
		tenantResourcesRedirect(w, r, tenant.Slug, url.Values{"err": {"Host not found"}})
		return
	}
	drv, err := Drivers.Compute(providers.DriverKind(host.DriverKind))
	if err != nil {
		tenantResourcesRedirect(w, r, tenant.Slug, url.Values{
			"host": {strconv.FormatInt(hostID, 10)},
			"err":  {err.Error()},
		})
		return
	}
	if err := drv.ValidateConnection(ctx, host.ToProviderHost()); err != nil {
		_ = db.MarkHostProvisioningFailed(ctx, Pools.Admin, hostID, err.Error(), "")
		tenantResourcesRedirect(w, r, tenant.Slug, url.Values{
			"host": {strconv.FormatInt(hostID, 10)},
			"err":  {"Connectivity check failed: " + err.Error()},
		})
		return
	}
	capInfo, err := drv.InspectCapacity(ctx, host.ToProviderHost())
	if err != nil {
		tenantResourcesRedirect(w, r, tenant.Slug, url.Values{
			"host": {strconv.FormatInt(hostID, 10)},
			"err":  {"Capacity probe failed: " + err.Error()},
		})
		return
	}
	_ = db.UpdateHostProbeCapacity(ctx, Pools.Admin, hostID, capInfo)
	tenantResourcesRedirect(w, r, tenant.Slug, url.Values{
		"host": {strconv.FormatInt(hostID, 10)},
		"ok":   {"SSH connectivity verified and hardware capacity updated"},
	})
}

func tenantHostProvisionFormHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := getSession(r)
	tenant, _ := tenantFromContext(r)
	hostID, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	host, ok, err := db.GetTenantSupplyHostByID(ctx, Pools.Admin, tenant.DBID, hostID)
	if err != nil || !ok {
		tenantResourcesRedirect(w, r, tenant.Slug, url.Values{"err": {"Host not found"}})
		return
	}
	_ = db.MarkHostProvisioningStart(ctx, Pools.Admin, hostID)

	go func(h db.SupplyHostRecord, uid int64) {
		bgCtx, bgCancel := context.WithTimeout(context.Background(), 15*time.Minute)
		defer bgCancel()
		_, _ = runHostAutoProvisioning(bgCtx, h, uid)
	}(host, user.ID)

	tenantResourcesRedirect(w, r, tenant.Slug, url.Values{
		"host": {strconv.FormatInt(hostID, 10)},
		"ok":   {"Virtualization stack auto-provisioning started in background"},
	})
}

func tenantHostStatusFormHandler(w http.ResponseWriter, r *http.Request) {
	tenant, _ := tenantFromContext(r)
	hostID, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	_ = r.ParseForm()
	newStatus := strings.TrimSpace(r.FormValue("inventory_status"))
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := db.SetTenantHostInventoryStatus(ctx, Pools.Admin, tenant.DBID, hostID, newStatus); err != nil {
		tenantResourcesRedirect(w, r, tenant.Slug, url.Values{
			"host": {strconv.FormatInt(hostID, 10)},
			"err":  {"Invalid inventory status"},
		})
		return
	}
	tenantResourcesRedirect(w, r, tenant.Slug, url.Values{
		"host": {strconv.FormatInt(hostID, 10)},
		"ok":   {"Host status updated to " + newStatus},
	})
}

func tenantHostDeleteFormHandler(w http.ResponseWriter, r *http.Request) {
	tenant, _ := tenantFromContext(r)
	hostID, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := db.DeleteTenantSupplyHost(ctx, Pools.Admin, tenant.DBID, hostID); err != nil {
		msg := "Failed to remove host"
		if errors.Is(err, db.ErrHostHasActiveWorkloads) {
			msg = "Cannot remove host while assigned IPs or customer VMs are active"
		}
		tenantResourcesRedirect(w, r, tenant.Slug, url.Values{
			"host": {strconv.FormatInt(hostID, 10)},
			"err":  {msg},
		})
		return
	}
	tenantResourcesRedirect(w, r, tenant.Slug, url.Values{"ok": {"Host removed from tenant inventory"}})
}

func tenantAddIPsFormHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := getSession(r)
	tenant, _ := tenantFromContext(r)
	if err := r.ParseForm(); err != nil {
		tenantResourcesRedirect(w, r, tenant.Slug, url.Values{"err": {"Invalid form payload"}})
		return
	}
	hostID, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if hostID <= 0 {
		hostID, _ = strconv.ParseInt(r.FormValue("host_id"), 10, 64)
	}
	returnView := strings.TrimSpace(r.FormValue("return_view"))
	if hostID <= 0 {
		tenantResourcesRedirect(w, r, tenant.Slug, url.Values{"view": {returnView}, "err": {"Select a target server host for the IP pool"}})
		return
	}

	cidr, _ := strconv.Atoi(r.FormValue("cidr_prefix"))
	costCents := parseDollarToCents(r.FormValue("monthly_cost_usd"))
	retailCents := parseDollarToCents(r.FormValue("retail_price_usd"))
	if r.FormValue("retail_price_usd") == "" {
		retailCents = 200
	}

	entries, err := db.ParseManualIPList(
		r.FormValue("ip_list"),
		cidr,
		r.FormValue("gateway"),
		r.FormValue("mac_address"),
		costCents,
		retailCents,
		r.FormValue("notes"),
	)
	if err != nil {
		params := url.Values{"err": {err.Error()}}
		if returnView == "ip-pool" {
			params.Set("view", "ip-pool")
		} else {
			params.Set("host", strconv.FormatInt(hostID, 10))
		}
		tenantResourcesRedirect(w, r, tenant.Slug, params)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()

	if _, err := db.AddTenantSupplyHostIPs(ctx, Pools.Admin, tenant.DBID, hostID, entries); err != nil {
		params := url.Values{"err": {"Failed to add IPs: " + err.Error()}}
		if returnView == "ip-pool" {
			params.Set("view", "ip-pool")
		} else {
			params.Set("host", strconv.FormatInt(hostID, 10))
		}
		tenantResourcesRedirect(w, r, tenant.Slug, params)
		return
	}

	db.InsertAuditEvent(ctx, Pools.Admin, user.ID, string(user.Role), "tenant.ips.added", "info",
		map[string]any{"tenant_id": tenant.DBID, "host_id": hostID, "added_count": len(entries)}, "", "", clientIP(r))

	params := url.Values{"ok": {fmt.Sprintf("Added %d assignable IP(s) to host pool", len(entries))}}
	if returnView == "ip-pool" {
		params.Set("view", "ip-pool")
	} else {
		params.Set("host", strconv.FormatInt(hostID, 10))
	}
	tenantResourcesRedirect(w, r, tenant.Slug, params)
}

func tenantDeleteIPFormHandler(w http.ResponseWriter, r *http.Request) {
	tenant, _ := tenantFromContext(r)
	ipID, _ := strconv.ParseInt(r.PathValue("ip_id"), 10, 64)
	_ = r.ParseForm()
	hostIDStr := r.URL.Query().Get("host")
	returnView := r.URL.Query().Get("view")

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	params := url.Values{}
	if returnView != "" {
		params.Set("view", returnView)
	}
	if hostIDStr != "" {
		params.Set("host", hostIDStr)
	}

	if err := db.DeleteTenantSupplyHostIP(ctx, Pools.Admin, tenant.DBID, ipID); err != nil {
		params.Set("err", "Cannot remove assigned IP address")
		tenantResourcesRedirect(w, r, tenant.Slug, params)
		return
	}
	params.Set("ok", "IP removed from pool")
	tenantResourcesRedirect(w, r, tenant.Slug, params)
}

func resourceStatusHandler(w http.ResponseWriter, r *http.Request) {
	tenant, _ := tenantFromContext(r)
	idStr := r.PathValue("id")
	newStatus := r.FormValue("status")

	if Pools != nil && tenant.DBID > 0 {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err == nil {
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()
			if err := db.UpdateTenantResourceStatus(ctx, Pools.Admin, tenant.DBID, id, newStatus); err != nil {
				log.Printf("resourceStatusHandler: %v", err)
			}
		}
	} else {
		tenantResourceStore.updateStatus(tenant.Slug, idStr, newStatus)
	}
	http.Redirect(w, r, "/t/"+tenant.Slug+"/resources/", http.StatusSeeOther)
}

func resourceDeleteHandler(w http.ResponseWriter, r *http.Request) {
	tenant, _ := tenantFromContext(r)
	idStr := r.PathValue("id")

	if Pools != nil && tenant.DBID > 0 {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err == nil {
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()
			if err := db.DeleteTenantResource(ctx, Pools.Admin, tenant.DBID, id); err != nil {
				log.Printf("resourceDeleteHandler: %v", err)
			}
		}
	} else {
		tenantResourceStore.delete(tenant.Slug, idStr)
	}
	http.Redirect(w, r, "/t/"+tenant.Slug+"/resources/", http.StatusSeeOther)
}

func RegisterTenantResourcesRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /t/{slug}/resources/{$}", tenantMiddleware(resourcesHandler))
	mux.HandleFunc("POST /t/{slug}/resources/onboard/{$}", tenantMiddleware(resourceOnboardFormHandler))
	mux.HandleFunc("POST /t/{slug}/resources/hosts/{id}/probe/{$}", tenantMiddleware(tenantHostProbeFormHandler))
	mux.HandleFunc("POST /t/{slug}/resources/hosts/{id}/provision/{$}", tenantMiddleware(tenantHostProvisionFormHandler))
	mux.HandleFunc("POST /t/{slug}/resources/hosts/{id}/status/{$}", tenantMiddleware(tenantHostStatusFormHandler))
	mux.HandleFunc("POST /t/{slug}/resources/hosts/{id}/delete/{$}", tenantMiddleware(tenantHostDeleteFormHandler))
	mux.HandleFunc("POST /t/{slug}/resources/hosts/{id}/ips/{$}", tenantMiddleware(tenantAddIPsFormHandler))
	mux.HandleFunc("POST /t/{slug}/resources/ips/{$}", tenantMiddleware(tenantAddIPsFormHandler))
	mux.HandleFunc("POST /t/{slug}/resources/ips/{ip_id}/delete/{$}", tenantMiddleware(tenantDeleteIPFormHandler))
	mux.HandleFunc("POST /t/{slug}/resources/{id}/status/{$}", tenantMiddleware(resourceStatusHandler))
	mux.HandleFunc("POST /t/{slug}/resources/{id}/delete/{$}", tenantMiddleware(resourceDeleteHandler))
}
