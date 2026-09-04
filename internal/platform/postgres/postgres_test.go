package postgres_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/satya-18-w/productivity-os/internal/platform/pgtest"
	"github.com/satya-18-w/productivity-os/internal/platform/postgres"
)

func TestOpen_Unreachable(t *testing.T) {
	_, err := postgres.Open(context.Background(),
		"postgres://nobody:nobody@127.0.0.1:1/nope?sslmode=disable&connect_timeout=1")
	require.Error(t, err)
	require.Contains(t, err.Error(), "unreachable")
}

func TestOpen_BadURL(t *testing.T) {
	_, err := postgres.Open(context.Background(), "://not a url")
	require.Error(t, err)
}

func TestOpenAndHealthy(t *testing.T) {
	pool := pgtest.Pool(t)
	require.True(t, postgres.Healthy(context.Background(), pool))
}

func TestMigrateUp_Idempotent(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	require.NoError(t, postgres.MigrateUp(url))
	require.NoError(t, postgres.MigrateUp(url))
}

func TestMigrate_DownThenUp(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	require.NoError(t, postgres.MigrateUp(url))
	require.NoError(t, postgres.MigrateDown(url))
	require.NoError(t, postgres.MigrateUp(url))

	pool, err := postgres.Open(context.Background(), url)
	require.NoError(t, err)
	defer pool.Close()

	var n int
	require.NoError(t, pool.QueryRow(context.Background(),
		`SELECT count(*) FROM information_schema.tables
		 WHERE table_schema = 'public' AND table_name IN ('accounts', 'sessions')`).Scan(&n))
	require.Equal(t, 2, n)
}
