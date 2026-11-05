package catalog

import "context"

// CatalogRepository manages the catalog of sellable ProductSKUs.
type CatalogRepository interface {
	ListSKUs(ctx context.Context, onlyActive bool) ([]ProductSKU, error)
	GetSKUByCode(ctx context.Context, code string) (ProductSKU, error)
	UpsertSKU(ctx context.Context, sku ProductSKU) (ProductSKU, error)
}

// SubscriptionBillingManager links provisioned resources (root services and
// attached child add-ons like IPs/volumes) to recurring or metered billing lines.
type SubscriptionBillingManager interface {
	AttachLineItem(ctx context.Context, item SubscriptionLineItem) (SubscriptionLineItem, error)
	DetachResourceLineItems(ctx context.Context, resourceID int64) error
	RecordUsage(ctx context.Context, sample UsageSample) error
	ListWorkspaceLineItems(ctx context.Context, workspaceID int64) ([]SubscriptionLineItem, error)
}
