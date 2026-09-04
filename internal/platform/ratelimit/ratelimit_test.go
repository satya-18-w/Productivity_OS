package ratelimit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLimiter(t *testing.T) {
	l := New(3, time.Minute)

	require.True(t, l.Allowed("a"))
	l.Fail("a")
	l.Fail("a")
	require.True(t, l.Allowed("a"), "2 failures still under budget of 3")
	l.Fail("a")
	require.False(t, l.Allowed("a"), "3rd failure trips the limit")

	require.True(t, l.Allowed("b"), "keys are independent")

	l.Reset("a")
	require.True(t, l.Allowed("a"), "reset restores the budget")
}

func TestLimiter_WindowExpiry(t *testing.T) {
	l := New(1, time.Minute)
	now := time.Now()
	l.now = func() time.Time { return now }

	l.Fail("k")
	require.False(t, l.Allowed("k"))

	now = now.Add(61 * time.Second)
	require.True(t, l.Allowed("k"), "a new window resets the count")
}
