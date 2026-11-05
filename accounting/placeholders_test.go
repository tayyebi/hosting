package accounting

import (
	"context"
	"errors"
	"testing"
	"time"

	"platform/providers"
)

func TestCostCentricAccountingPipeline(t *testing.T) {
	ctx := context.Background()
	calc := &PlaceholderSupplyCostCalculator{}
	pricing := &PlaceholderPricingEngine{}
	ledger := &PlaceholderMarginLedger{}

	cogs, err := calc.CalculateSupplyCost(ctx, SupplyCostInput{
		Model:                CostModelAmortizedHardware,
		HostMonthlyCostCents: 2400,
		HostTotalVCPUs:       8,
		HostTotalRAMMB:       16384,
		HostTotalDiskGB:      160,
		AllocatedVCPUs:       2,
		AllocatedRAMMB:       4096,
		AllocatedDiskGB:      40,
		Currency:             "USD",
	})
	if err != nil {
		t.Fatalf("CalculateSupplyCost: %v", err)
	}
	if cogs.TotalSupplyCostCents != 600 {
		t.Fatalf("expected 1/4 slice COGS = 600 cents, got %d", cogs.TotalSupplyCostCents)
	}

	// EnforceCostFloor must reject predatory discount below COGS
	_, err = pricing.Quote(ctx, cogs, PricingContext{
		ClientPromoDiscBps: 5000,
		EnforceCostFloor:   true,
	})
	if !errors.Is(err, ErrBelowCostFloor) {
		t.Fatalf("expected ErrBelowCostFloor, got %v", err)
	}

	quote, err := pricing.Quote(ctx, cogs, PricingContext{
		DriverKind:            providers.DriverLinuxNestedQEMU,
		OwnershipType:         providers.OwnershipLeasedNestedVPS,
		ProviderMarkupBps:     3500,
		TenantRetailMarkupBps: 2500,
		EnforceCostFloor:      true,
	})
	if err != nil {
		t.Fatalf("Quote: %v", err)
	}
	if quote.EndUserTotalCents <= cogs.TotalSupplyCostCents {
		t.Fatalf("expected retail (%d) > COGS (%d)", quote.EndUserTotalCents, cogs.TotalSupplyCostCents)
	}

	now := time.Now().UTC()
	if err := ledger.RecordCost(ctx, CostEvent{AmountCents: cogs.TotalSupplyCostCents, OccurredAt: now}); !errors.Is(err, ErrNotImplemented) {
		t.Fatalf("expected ErrNotImplemented on placeholder ledger, got %v", err)
	}
}
