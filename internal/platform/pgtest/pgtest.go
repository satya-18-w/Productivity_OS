// Package pgtest provides a real-PostgreSQL pool for integration tests. Tests using
// it are skipped unless TEST_DATABASE_URL is set (ADR-0007).
//
// Each test binary (package) gets its own database, derived from the binary name,
// so packages can run in parallel without contending on the same tables. Within a
// package, every public table is truncated between tests.
package pgtest

import (
	"context"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/satya-18-w/productivity-os/internal/platform/postgres"
)

const duplicateDatabase = "42P04"

var nonIdent = regexp.MustCompile(`[^a-z0-9_]+`)

// Pool returns a migrated pool against this package's dedicated test database.
func Pool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	base := os.Getenv("TEST_DATABASE_URL")
	if base == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping integration test")
	}

	dbURL := ensurePackageDB(t, base)

	require.NoError(t, postgres.MigrateUp(dbURL), "apply migrations to the test database")

	pool, err := postgres.Open(context.Background(), dbURL)
	require.NoError(t, err)

	// Truncate on acquire, not on cleanup: a test that deliberately breaks its
	// pool (e.g. an internal-error test) must not fail teardown.
	truncateAll(t, pool)
	t.Cleanup(pool.Close)
	return pool
}

// ensurePackageDB creates (once) a database named after the running test binary
// and returns its connection URL.
func ensurePackageDB(t *testing.T, base string) string {
	t.Helper()

	name := "pos_test_" + nonIdent.ReplaceAllString(
		strings.TrimSuffix(strings.ToLower(filepath.Base(os.Args[0])), ".test"), "_")

	admin, err := url.Parse(base)
	require.NoError(t, err)

	target := *admin
	target.Path = "/" + name

	adminConn := *admin
	adminConn.Path = "/postgres"

	conn, err := pgx.Connect(context.Background(), adminConn.String())
	require.NoError(t, err)
	defer func() { _ = conn.Close(context.Background()) }()

	_, err = conn.Exec(context.Background(), "CREATE DATABASE "+pgx.Identifier{name}.Sanitize())
	if err != nil {
		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) || pgErr.Code != duplicateDatabase {
			require.NoError(t, err)
		}
	}
	return target.String()
}

func truncateAll(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()

	rows, err := pool.Query(ctx, `
		SELECT tablename FROM pg_tables
		WHERE schemaname = 'public' AND tablename <> 'schema_migrations'`)
	require.NoError(t, err)

	var tables []string
	for rows.Next() {
		var name string
		require.NoError(t, rows.Scan(&name))
		tables = append(tables, name)
	}
	require.NoError(t, rows.Err())
	if len(tables) == 0 {
		return
	}

	_, err = pool.Exec(ctx, "TRUNCATE "+strings.Join(tables, ", ")+" RESTART IDENTITY CASCADE")
	require.NoError(t, err)
}
