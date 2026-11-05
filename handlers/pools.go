package handlers

import (
	"platform/accounting"
	"platform/catalog"
	"platform/db"
	"platform/providers"
)

// Pools holds the three pgxpool connection pools (app, admin, readonly)
// set once by main() after db.Open and db.Migrate succeed.
var Pools *db.Pools

// Domain-layer registries and engines wired into handlers.
var (
	Drivers  *providers.Registry                  = providers.NewDefaultRegistry()
	Catalog  catalog.CatalogRepository            = catalog.PlaceholderCatalogRepository{}
	SubMgr   catalog.SubscriptionBillingManager   = catalog.PlaceholderSubscriptionManager{}
	CostCalc accounting.SupplyCostCalculator      = accounting.PlaceholderSupplyCostCalculator{}
	Pricing  accounting.PricingEngine             = accounting.PlaceholderPricingEngine{}
	Ledger   accounting.MarginLedger              = accounting.PlaceholderMarginLedger{}
)
