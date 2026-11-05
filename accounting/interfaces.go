package accounting

import (
	"context"
	"time"
)

// SupplyCostCalculator computes the underlying COGS of a resource or host slice.
type SupplyCostCalculator interface {
	CalculateSupplyCost(ctx context.Context, in SupplyCostInput) (SupplyCostBreakdown, error)
}

// PricingEngine derives Platform->Tenant wholesale and Tenant->Client retail
// quotes from SupplyCostBreakdown and PricingContext.
type PricingEngine interface {
	Quote(ctx context.Context, cost SupplyCostBreakdown, pctx PricingContext) (PriceQuote, error)
}

// MarginLedger records cost & revenue events and aggregates profitability per cost center.
type MarginLedger interface {
	RecordCost(ctx context.Context, ev CostEvent) error
	RecordRevenue(ctx context.Context, ev RevenueEvent) error
	SummarizeMargins(ctx context.Context, kind CostCenterKind, from, to time.Time) ([]MarginView, error)
}
