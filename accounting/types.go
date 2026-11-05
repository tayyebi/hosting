// Package accounting models the three-layer cost-centric accounting system:
//   - Layer A: Supply Cost (COGS — direct wholesale pass-through, hardware
//     slice amortization, and shared rack/transit/IP overhead)
//   - Layer B: Multi-tier Pricing Rules Engine (SupplyCost -> Platform Wholesale
//     -> Tenant Retail + Tax, guarded by CostFloor)
//   - Layer C: Double-Entry Cost/Revenue Margin Ledger grouped by Cost Center
package accounting

import (
	"errors"
	"time"

	"platform/providers"
)

var ErrNotImplemented = errors.New("accounting: operation not yet implemented")
var ErrBelowCostFloor = errors.New("accounting: computed price falls below supply cost floor")

// ── Layer A: Supply Cost (COGS) ──────────────────────────────────────────────

type SupplyCostModel string

const (
	CostModelDirectPassThrough SupplyCostModel = "direct_pass_through"
	CostModelAmortizedHardware SupplyCostModel = "amortized_hardware"
	CostModelSharedOverhead    SupplyCostModel = "shared_overhead"
)

// SupplyCostInput describes the host/provider economics and slice allocation
// used to compute the true monthly COGS of a resource (VM, IP, volume, etc.).
type SupplyCostInput struct {
	Model                   SupplyCostModel
	HostMonthlyCostCents    int64
	HostTotalVCPUs          int
	HostTotalRAMMB          int
	HostTotalDiskGB         int
	AllocatedVCPUs          int
	AllocatedRAMMB          int
	AllocatedDiskGB         int
	DirectWholesaleCents    int64
	SharedOverheadCents     int64
	Currency                string
}

type SupplyCostBreakdown struct {
	ComputeShareCents    int64
	MemoryShareCents     int64
	StorageShareCents    int64
	DirectWholesaleCents int64
	OverheadShareCents   int64
	TotalSupplyCostCents int64
	Currency             string
}

// ── Layer B: Pricing Rules Engine ────────────────────────────────────────────

// PricingContext supplies the provider, regional, reseller-tier, promotional,
// SLA, and tax factors used to derive wholesale and retail prices from COGS.
type PricingContext struct {
	TenantID              int64
	TenantTier            string // "free" | "starter" | "pro" | "enterprise"
	DriverKind            providers.DriverKind
	OwnershipType         providers.OwnershipType
	RegionSlug            string
	ProviderMarkupBps     int // basis points (e.g. 3500 = +35.00%)
	RegionMarkupBps       int
	TenantResellerDiscBps int
	TenantRetailMarkupBps int
	ClientPromoDiscBps    int
	SLASurchargeBps       int
	TaxRateBps            int
	EnforceCostFloor      bool
}

type PriceQuote struct {
	SupplyCostCents           int64
	CostFloorCents            int64
	PlatformWholesaleCents    int64
	TenantRetailSubtotalCents int64
	TaxCents                  int64
	EndUserTotalCents         int64
	ExpectedPlatformMargin    int64
	ExpectedTenantMargin      int64
	Currency                  string
}

// ── Layer C: Cost-Center Margin Ledger ───────────────────────────────────────

type CostCenterKind string

const (
	CostCenterSupplyHost CostCenterKind = "supply_host"
	CostCenterProvider   CostCenterKind = "provider"
	CostCenterRegion     CostCenterKind = "region"
	CostCenterTenant     CostCenterKind = "tenant"
	CostCenterWorkspace  CostCenterKind = "workspace"
)

// CostEvent records infrastructure COGS accrued against a resource/supply host.
type CostEvent struct {
	ID           int64
	ResourceID   *int64
	SupplyHostID *int64
	TenantID     *int64
	WorkspaceID  *int64
	AmountCents  int64
	Currency     string
	Description  string
	PeriodStart  time.Time
	PeriodEnd    time.Time
	OccurredAt   time.Time
}

// RevenueEvent records wholesale or retail revenue recognized for a resource.
type RevenueEvent struct {
	ID               int64
	ResourceID       *int64
	TenantID         int64
	WorkspaceID      *int64
	InvoiceRef       string
	GrossAmountCents int64
	PaymentFeeCents  int64
	SLACreditCents   int64
	TaxCents         int64
	Currency         string
	OccurredAt       time.Time
}

// MarginView aggregates GrossMargin = Revenue - SupplyCost - PaymentFee - SLACredit
// for any CostCenterKind.
type MarginView struct {
	CenterKind       CostCenterKind
	CenterRef        string
	RevenueCents     int64
	SupplyCostCents  int64
	PaymentFeeCents  int64
	SLACreditCents   int64
	GrossMarginCents int64
	GrossMarginPct   float64
	Currency         string
}
