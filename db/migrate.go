package db

import (
	"errors"
	"fmt"
	"io/fs"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

// Migrate applies every pending migration using the admin pool. The
// caller passes in the migrations as an fs.FS — typically an embed.FS
// declared in the main package so `go:embed` can reach the sibling
// migrations/ directory (embed does not allow ../ paths).
//
// Safe to call on every boot; no-op when the DB is already at head.
func Migrate(pool *pgxpool.Pool, migrations fs.FS) error {
	src, err := iofs.New(migrations, ".")
	if err != nil {
		return fmt.Errorf("db: iofs source: %w", err)
	}

	// golang-migrate wants a *sql.DB; pgx exposes one via its stdlib shim.
	sqlDB := stdlib.OpenDBFromPool(pool)
	defer sqlDB.Close()

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("db: postgres driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", src, "postgres", driver)
	if err != nil {
		return fmt.Errorf("db: new migrator: %w", err)
	}

	// Self-healing: if the migrations table is marked dirty from a prior
	// failed run, clear the flag and let the up-files re-apply. Every
	// migration is authored to be idempotent (CREATE ... IF NOT EXISTS,
	// DROP POLICY IF EXISTS, etc.), so replaying is safe.
	if version, dirty, verr := m.Version(); verr == nil && dirty {
		if ferr := m.Force(int(version)); ferr != nil {
			return fmt.Errorf("db: clear dirty flag at v%d: %w", version, ferr)
		}
		// Force() sets the version back to `version` but marks it NOT dirty.
		// To actually re-run v`version`, step back one and let Up() replay.
		if version > 0 {
			if ferr := m.Force(int(version) - 1); ferr != nil {
				return fmt.Errorf("db: rewind past v%d for replay: %w", version, ferr)
			}
		}
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("db: apply migrations: %w", err)
	}
	return nil
}
