// Package catalog models the unified Product = Resource = Payable-Item layer.
// Every sellable unit — whether a root service (VPS, Shared Hosting, PaaS App,
// Managed K8s) or an attachable add-on (IPv4 address, IPv6 subnet, NVMe volume,
// snapshot, DDoS protection) — is defined by a ProductSKU and instantiated as
// a node in a parent-child PayableResource tree.
package catalog

import (
	"errors"
	"time"

	"platform/providers"
)

var ErrNotImplemented = errors.New("catalog: operation not yet implemented")
var ErrSKUNotFound = errors.New("catalog: SKU not found")

// PayableKind enumerates every category of sellable / provisionable item.
type PayableKind string

const (
	KindVM             PayableKind = "vm"
	KindIPv4           PayableKind = "ipv4"
	KindIPv6Subnet     PayableKind = "ipv6_subnet"
	KindBlockSSD       PayableKind = "block_ssd"
	KindSharedHosting  PayableKind = "shared_hosting"
	KindK8sNode        PayableKind = "k8s_node"
	KindPaaSApp        PayableKind = "paas_app"
	KindSnapshot       PayableKind = "snapshot"
	KindBackupGB       PayableKind = "backup_gb"
	KindDDoSProtection PayableKind = "ddos_protection"
)

// BillingModel determines how a ProductSKU is rated over time.
type BillingModel string

const (
	BillingRecurringFlat BillingModel = "recurring_flat"
	BillingHourlyMetered BillingModel = "hourly_metered"
	BillingUsageCounter  BillingModel = "usage_counter"
	BillingOneTimeSetup  BillingModel = "one_time_setup"
)

// ProvisionTarget maps a ProductSKU onto compatible infrastructure drivers
// and placement constraints.
type ProvisionTarget struct {
	DriverKind         providers.DriverKind
	RequiredRegionSlug string
	VCPUs              int
	MemoryMB           int
	DiskGB             int
	RequiresPublicIPv4 bool
	ProviderTags       []string
}

// ProductSKU is the commercial + technical catalog definition of any sellable item.
type ProductSKU struct {
	ID                     int64
	Code                   string
	Name                   string
	Kind                   PayableKind
	BillingModel           BillingModel
	BaseSupplyCostCents    int64
	SuggestedWholesaleCent int64
	SuggestedRetailCents   int64
	Currency               string
	UnitName               string // e.g. "instance", "ip", "gb-month", "hour"
	Target                 ProvisionTarget
	Active                 bool
}

// PayableResourceNode represents an instantiated resource row (`resources` table)
// as a node in a billable & operational tree (for example, a customer VPS root
// node with child `ipv4` and `block_ssd` nodes via ParentResourceID).
type PayableResourceNode struct {
	ResourceID       int64
	TenantID         int64
	WorkspaceID      int64
	ParentResourceID *int64
	SupplyHostID     *int64
	ExternalRef      string
	SKUCode          string
	Kind             PayableKind
	Name             string
	Status           string
	CreatedAt        time.Time
	Children         []*PayableResourceNode
}

// SubscriptionLineItem binds an instantiated PayableResourceNode to a billing cycle.
type SubscriptionLineItem struct {
	ID               int64
	TenantID         int64
	WorkspaceID      int64
	ResourceID       int64
	ParentResourceID *int64
	SKUCode          string
	Quantity         float64
	UnitPriceCents   int64
	Currency         string
	BillingModel     BillingModel
	PeriodStart      time.Time
	PeriodEnd        time.Time
	Active           bool
}

// UsageSample captures one metered reading (GB-hours, egress TB, API calls)
// for a UsageCounter or HourlyMetered resource.
type UsageSample struct {
	ResourceID int64
	Metric     string
	Value      float64
	RecordedAt time.Time
}
