package accounting

import (
	"context"
	"fmt"
	"time"
)

// PlaceholderSupplyCostCalculator provides baseline slice-amortization math and
// acts as the placeholder COGS calculator until historical utilization telemetry is wired.
type PlaceholderSupplyCostCalculator struct{}

func (PlaceholderSupplyCostCalculator) CalculateSupplyCost(_ context.Context, in SupplyCostInput) (SupplyCostBreakdown, error) {
	cur := in.Currency
	if cur == "" {
		cur = "USD"
	}
	out := SupplyCostBreakdown{
		DirectWholesaleCents: in.DirectWholesaleCents,
		OverheadShareCents:   in.SharedOverheadCents,
		Currency:             cur,
	}

	// Standard 50% CPU / 35% RAM / 15% Disk hardware cost attribution when amortizing a host.
	if in.HostMonthlyCostCents > 0 {
		if in.HostTotalVCPUs > 0 && in.AllocatedVCPUs > 0 {
			cpuPool := (in.HostMonthlyCostCents * 50) / 100
			out.ComputeShareCents = (cpuPool * int64(in.AllocatedVCPUs)) / int64(in.HostTotalVCPUs)
		}
		if in.HostTotalRAMMB > 0 && in.AllocatedRAMMB > 0 {
			ramPool := (in.HostMonthlyCostCents * 35) / 100
			out.MemoryShareCents = (ramPool * int64(in.AllocatedRAMMB)) / int64(in.HostTotalRAMMB)
		}
		if in.HostTotalDiskGB > 0 && in.AllocatedDiskGB > 0 {
			diskPool := (in.HostMonthlyCostCents * 15) / 100
			out.StorageShareCents = (diskPool * int64(in.AllocatedDiskGB)) / int64(in.HostTotalDiskGB)
		}
	}

	out.TotalSupplyCostCents = out.ComputeShareCents +
		out.MemoryShareCents +
		out.StorageShareCents +
		out.DirectWholesaleCents +
		out.OverheadShareCents
	return out, nil
}

// PlaceholderPricingEngine implements baseline basis-point markups and cost-floor protection.
type PlaceholderPricingEngine struct{}

func applyBps(amount int64, bps int) int64 {
	return amount + (amount*int64(bps))/10000
}

func (PlaceholderPricingEngine) Quote(_ context.Context, cost SupplyCostBreakdown, pctx PricingContext) (PriceQuote, error) {
	base := cost.TotalSupplyCostCents
	wholesale := applyBps(base, pctx.ProviderMarkupBps+pctx.RegionMarkupBps+pctx.SLASurchargeBps-pctx.TenantResellerDiscBps)
	if pctx.EnforceCostFloor && wholesale < base {
		return PriceQuote{}, fmt.Errorf("%w: wholesale %d < supply cost %d", ErrBelowCostFloor, wholesale, base)
	}

	retailSub := applyBps(wholesale, pctx.TenantRetailMarkupBps-pctx.ClientPromoDiscBps)
	if pctx.EnforceCostFloor && retailSub < wholesale {
		return PriceQuote{}, fmt.Errorf("%w: retail %d < wholesale %d", ErrBelowCostFloor, retailSub, wholesale)
	}

	tax := (retailSub * int64(pctx.TaxRateBps)) / 10000
	return PriceQuote{
		SupplyCostCents:           base,
		CostFloorCents:            base,
		PlatformWholesaleCents:    wholesale,
		TenantRetailSubtotalCents: retailSub,
		TaxCents:                  tax,
		EndUserTotalCents:         retailSub + tax,
		ExpectedPlatformMargin:    wholesale - base,
		ExpectedTenantMargin:      retailSub - wholesale,
		Currency:                  cost.Currency,
	}, nil
}

// PlaceholderMarginLedger is the placeholder double-entry ledger until DB ledger tables are wired.
type PlaceholderMarginLedger struct{}

func (PlaceholderMarginLedger) RecordCost(context.Context, CostEvent) error {
	return fmt.Errorf("%w: PlaceholderMarginLedger.RecordCost", ErrNotImplemented)
}

func (PlaceholderMarginLedger) RecordRevenue(context.Context, RevenueEvent) error {
	return fmt.Errorf("%w: PlaceholderMarginLedger.RecordRevenue", ErrNotImplemented)
}

func (PlaceholderMarginLedger) SummarizeMargins(context.Context, CostCenterKind, time.Time, time.Time) ([]MarginView, error) {
	return nil, fmt.Errorf("%w: PlaceholderMarginLedger.SummarizeMargins", ErrNotImplemented)
}
