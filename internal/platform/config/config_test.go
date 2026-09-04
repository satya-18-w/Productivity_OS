package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLoad_RequiresDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	_, err := Load()
	require.Error(t, err)
	require.Contains(t, err.Error(), "DATABASE_URL")
}

func TestLoad_RejectsUnknownEnv(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("ENV", "staging")
	_, err := Load()
	require.Error(t, err)
	require.Contains(t, err.Error(), "ENV")
}

func TestLoad_RejectsBadDuration(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("SESSION_TTL", "banana")
	_, err := Load()
	require.Error(t, err)
	require.Contains(t, err.Error(), "SESSION_TTL")
}

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("PORT", "")
	t.Setenv("ENV", "")
	t.Setenv("SESSION_TTL", "")
	t.Setenv("SHUTDOWN_GRACE", "")

	c, err := Load()
	require.NoError(t, err)
	require.Equal(t, "8080", c.Port)
	require.Equal(t, "development", c.Env)
	require.Equal(t, 720*time.Hour, c.SessionTTL)
	require.Equal(t, 10*time.Second, c.ShutdownGrace)
	require.False(t, c.IsProduction())
}
