package habits

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/satya-18-w/productivity-os/internal/platform/timezone"
)

func set(dates ...timezone.Date) map[string]struct{} {
	m := make(map[string]struct{}, len(dates))
	for _, d := range dates {
		m[d.String()] = struct{}{}
	}
	return m
}

func TestCurrentStreak(t *testing.T) {
	today := timezone.Date{Year: 2025, Month: time.June, Day: 15}
	d := func(n int) timezone.Date { return today.AddDays(n) }

	require.Equal(t, 0, currentStreak(set(), today), "no completions")
	require.Equal(t, 1, currentStreak(set(today), today), "just today")
	require.Equal(t, 3, currentStreak(set(d(0), d(-1), d(-2)), today), "three in a row ending today")

	require.Equal(t, 2, currentStreak(set(d(-1), d(-2)), today),
		"today not done but yesterday is — anchor on yesterday")

	require.Equal(t, 0, currentStreak(set(d(-2), d(-3)), today),
		"most recent is two days ago — streak is broken")

	require.Equal(t, 2, currentStreak(set(d(0), d(-1), d(-3), d(-4)), today),
		"gap at d-2 stops the count")

	require.Equal(t, 3, currentStreak(set(d(1), d(0), d(-1), d(-2)), today),
		"a future completion (d+1) does not extend the current streak")

	long := make([]timezone.Date, 0, 100)
	for i := 0; i < 100; i++ {
		long = append(long, d(-i))
	}
	require.Equal(t, 100, currentStreak(set(long...), today))
}

func TestCountInRange(t *testing.T) {
	base := timezone.Date{Year: 2025, Month: time.June, Day: 1}
	s := set(base, base.AddDays(2), base.AddDays(5), base.AddDays(40))
	require.Equal(t, 3, countInRange(s, base, base.AddDays(29)))
	require.Equal(t, 0, countInRange(s, base.AddDays(10), base.AddDays(20)))
}
