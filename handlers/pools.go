package handlers

import "platform/db"

// Pools holds the three pgxpool connection pools (app, admin, readonly)
// that Phase 1 handlers depend on. Set once by main() after db.Open and
// db.Migrate succeed. Nil until wired.
//
// Handlers that need DB access read from this package-level variable
// rather than receiving the pools through every function signature —
// the existing handler registration API (plain http.HandlerFunc values)
// would otherwise require invasive refactors. This can be tightened in
// a later phase if we adopt a proper request context / DI pattern.
var Pools *db.Pools
