package main

import "embed"

// migrationsFS embeds every .sql file under migrations/ at build time.
// The //go:embed directive cannot traverse parent directories (..), so
// this declaration lives alongside main.go at the repo root and is
// passed into db.Migrate from main().
//
//go:embed migrations/*.sql
var migrationsFS embed.FS
