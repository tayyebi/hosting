package catalog

import (
	"context"
	"fmt"
)

// BuildResourceTree assembles flat PayableResourceNode slices (loaded from `resources`)
// into a parent-child forest using ParentResourceID, so root services (such as a VPS)
// carry their attached payable children (such as assigned IPv4 addresses or volumes).
func BuildResourceTree(flat []PayableResourceNode) []*PayableResourceNode {
	byID := make(map[int64]*PayableResourceNode, len(flat))
	nodes := make([]*PayableResourceNode, len(flat))
	for i := range flat {
		cp := flat[i]
		cp.Children = nil
		nodes[i] = &cp
		byID[cp.ResourceID] = nodes[i]
	}

	var roots []*PayableResourceNode
	for _, n := range nodes {
		if n.ParentResourceID != nil {
			if parent, ok := byID[*n.ParentResourceID]; ok {
				parent.Children = append(parent.Children, n)
				continue
			}
		}
		roots = append(roots, n)
	}
	return roots
}

// PlaceholderCatalogRepository is the placeholder SKU store until DB-backed SKU
// tables are introduced.
type PlaceholderCatalogRepository struct{}

func (PlaceholderCatalogRepository) ListSKUs(context.Context, bool) ([]ProductSKU, error) {
	return nil, fmt.Errorf("%w: PlaceholderCatalogRepository.ListSKUs", ErrNotImplemented)
}

func (PlaceholderCatalogRepository) GetSKUByCode(_ context.Context, code string) (ProductSKU, error) {
	return ProductSKU{}, fmt.Errorf("%w: PlaceholderCatalogRepository.GetSKUByCode(%s)", ErrNotImplemented, code)
}

func (PlaceholderCatalogRepository) UpsertSKU(context.Context, ProductSKU) (ProductSKU, error) {
	return ProductSKU{}, fmt.Errorf("%w: PlaceholderCatalogRepository.UpsertSKU", ErrNotImplemented)
}

// PlaceholderSubscriptionManager is the placeholder billing-line manager.
type PlaceholderSubscriptionManager struct{}

func (PlaceholderSubscriptionManager) AttachLineItem(context.Context, SubscriptionLineItem) (SubscriptionLineItem, error) {
	return SubscriptionLineItem{}, fmt.Errorf("%w: AttachLineItem", ErrNotImplemented)
}

func (PlaceholderSubscriptionManager) DetachResourceLineItems(context.Context, int64) error {
	return fmt.Errorf("%w: DetachResourceLineItems", ErrNotImplemented)
}

func (PlaceholderSubscriptionManager) RecordUsage(context.Context, UsageSample) error {
	return fmt.Errorf("%w: RecordUsage", ErrNotImplemented)
}

func (PlaceholderSubscriptionManager) ListWorkspaceLineItems(context.Context, int64) ([]SubscriptionLineItem, error) {
	return nil, fmt.Errorf("%w: ListWorkspaceLineItems", ErrNotImplemented)
}
