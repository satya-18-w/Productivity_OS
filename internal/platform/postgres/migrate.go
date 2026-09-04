package postgres

import (
	"errors"
	"fmt"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5" // registers the "pgx5" scheme
	"github.com/golang-migrate/migrate/v4/source/iofs"

	appdb "github.com/satya-18-w/productivity-os/db"
)

func newMigrator(dbURL string) (*migrate.Migrate, error) {
	src, err := iofs.New(appdb.MigrationsFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("load embedded migrations: %w", err)
	}
	m, err := migrate.NewWithSourceInstance("iofs", src, pgx5URL(dbURL))
	if err != nil {
		return nil, fmt.Errorf("init migrator: %w", err)
	}
	return m, nil
}

func closeMigrator(m *migrate.Migrate) {
	srcErr, dbErr := m.Close()
	_ = srcErr
	_ = dbErr
}

// MigrateUp applies every pending migration. It is safe to run when the database
// is already current.
func MigrateUp(dbURL string) error {
	m, err := newMigrator(dbURL)
	if err != nil {
		return err
	}
	defer closeMigrator(m)

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}

// MigrateDown rolls every migration back. Development and tests only.
func MigrateDown(dbURL string) error {
	m, err := newMigrator(dbURL)
	if err != nil {
		return err
	}
	defer closeMigrator(m)

	if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate down: %w", err)
	}
	return nil
}

// MigrateDrop drops everything in the database, including the migration history.
// Tests only.
func MigrateDrop(dbURL string) error {
	m, err := newMigrator(dbURL)
	if err != nil {
		return err
	}
	defer closeMigrator(m)

	if err := m.Drop(); err != nil {
		return fmt.Errorf("migrate drop: %w", err)
	}
	return nil
}

// pgx5URL rewrites a postgres:// connection string to the scheme the golang-migrate
// pgx/v5 driver registers.
func pgx5URL(u string) string {
	for _, p := range []string{"postgres://", "postgresql://"} {
		if strings.HasPrefix(u, p) {
			return "pgx5://" + strings.TrimPrefix(u, p)
		}
	}
	return u
}
