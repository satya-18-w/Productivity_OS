package timeline_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/satya-18-w/productivity-os/internal/platform/timezone"
	"github.com/satya-18-w/productivity-os/internal/timeline"
)

func date(y int, m time.Month, d int) timezone.Date {
	return timezone.Date{Year: y, Month: m, Day: d}
}

// findRow returns the comparison row for a category name.
func findRow(cmp timeline.DayComparison, name string) (timeline.CategoryTotals, bool) {
	for _, c := range cmp.Categories {
		if c.CategoryName == name {
			return c, true
		}
	}
	return timeline.CategoryTotals{}, false
}

func TestTimeline_SplitsPlannedAndActual(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()

	_, _ = svc.AddBlock(ctx, acc, timeline.BlockInput{
		Kind: timeline.Planned, StartsAt: ts("2025-06-15T09:00:00Z"), EndsAt: ts("2025-06-15T11:00:00Z"),
	})
	_, _ = svc.AddBlock(ctx, acc, timeline.BlockInput{
		Kind: timeline.Actual, StartsAt: ts("2025-06-15T09:30:00Z"), EndsAt: ts("2025-06-15T12:00:00Z"),
	})
	// a block on a different day is excluded
	_, _ = svc.AddBlock(ctx, acc, timeline.BlockInput{
		Kind: timeline.Planned, StartsAt: ts("2025-06-16T09:00:00Z"), EndsAt: ts("2025-06-16T10:00:00Z"),
	})

	tl, err := svc.Timeline(ctx, acc, date(2025, time.June, 15))
	require.NoError(t, err)
	require.Len(t, tl.Planned, 1)
	require.Len(t, tl.Actual, 1)
}

func TestTimeline_MidnightBlockAppearsOnBothDays(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()
	_, _ = svc.AddBlock(ctx, acc, timeline.BlockInput{
		Kind: timeline.Actual, StartsAt: ts("2025-06-15T23:00:00Z"), EndsAt: ts("2025-06-16T02:00:00Z"),
	})

	n, _ := svc.Timeline(ctx, acc, date(2025, time.June, 15))
	n1, _ := svc.Timeline(ctx, acc, date(2025, time.June, 16))
	require.Len(t, n.Actual, 1)
	require.Len(t, n1.Actual, 1)
}

func TestComparison_PerCategoryTotalsAndDifference(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()
	cat, _ := svc.CreateCategory(ctx, acc, "DSA")

	_, _ = svc.AddBlock(ctx, acc, timeline.BlockInput{
		Kind: timeline.Planned, StartsAt: ts("2025-06-15T09:00:00Z"), EndsAt: ts("2025-06-15T11:00:00Z"),
		CategoryID: &cat.ID,
	})
	_, _ = svc.AddBlock(ctx, acc, timeline.BlockInput{
		Kind: timeline.Actual, StartsAt: ts("2025-06-15T09:00:00Z"), EndsAt: ts("2025-06-15T12:00:00Z"),
		CategoryID: &cat.ID,
	})

	cmp, err := svc.Comparison(ctx, acc, date(2025, time.June, 15))
	require.NoError(t, err)
	row, ok := findRow(cmp, "DSA")
	require.True(t, ok)
	require.Equal(t, int64(7200), row.PlannedSeconds)
	require.Equal(t, int64(10800), row.ActualSeconds)
	require.Equal(t, int64(3600), row.DifferenceSeconds)
	require.Equal(t, cat.ID, *row.CategoryID)
}

func TestComparison_UncategorizedBucket(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()
	_, _ = svc.AddBlock(ctx, acc, timeline.BlockInput{
		Kind: timeline.Actual, StartsAt: ts("2025-06-15T09:00:00Z"), EndsAt: ts("2025-06-15T10:30:00Z"),
	})

	cmp, _ := svc.Comparison(ctx, acc, date(2025, time.June, 15))
	row, ok := findRow(cmp, "Uncategorized")
	require.True(t, ok)
	require.Nil(t, row.CategoryID)
	require.Equal(t, int64(5400), row.ActualSeconds)
}

func TestComparison_OverlappingBlocksAreSummed(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()
	cat, _ := svc.CreateCategory(ctx, acc, "Work")

	for _, span := range [][2]string{
		{"2025-06-15T09:00:00Z", "2025-06-15T10:00:00Z"},
		{"2025-06-15T09:30:00Z", "2025-06-15T10:30:00Z"}, // overlaps the first
	} {
		_, err := svc.AddBlock(ctx, acc, timeline.BlockInput{
			Kind: timeline.Actual, StartsAt: ts(span[0]), EndsAt: ts(span[1]), CategoryID: &cat.ID,
		})
		require.NoError(t, err)
	}

	cmp, _ := svc.Comparison(ctx, acc, date(2025, time.June, 15))
	row, _ := findRow(cmp, "Work")
	require.Equal(t, int64(7200), row.ActualSeconds, "Q7: sum of durations, not merged wall-clock")
}

func TestComparison_MidnightSpanningBlockIsClippedToTheDate(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()
	_, _ = svc.AddBlock(ctx, acc, timeline.BlockInput{
		Kind: timeline.Planned, StartsAt: ts("2025-06-15T23:00:00Z"), EndsAt: ts("2025-06-16T02:00:00Z"),
	})

	n, _ := svc.Comparison(ctx, acc, date(2025, time.June, 15))
	n1, _ := svc.Comparison(ctx, acc, date(2025, time.June, 16))

	rn, _ := findRow(n, "Uncategorized")
	rn1, _ := findRow(n1, "Uncategorized")
	require.Equal(t, int64(3600), rn.PlannedSeconds, "1h on day N")
	require.Equal(t, int64(7200), rn1.PlannedSeconds, "2h on day N+1")
}

func TestComparison_DSTTransitionCountsRealElapsedTime(t *testing.T) {
	ny, err := timezone.LoadLocation("America/New_York")
	require.NoError(t, err)
	svc, _, acc := setupZone(t, ny)
	ctx := context.Background()

	// 01:00 -> 04:00 local on the spring-forward day: 02:00-03:00 does not exist,
	// so only 2 real hours elapse.
	start := time.Date(2025, time.March, 9, 1, 0, 0, 0, ny)
	end := time.Date(2025, time.March, 9, 4, 0, 0, 0, ny)
	_, err = svc.AddBlock(ctx, acc, timeline.BlockInput{Kind: timeline.Actual, StartsAt: start, EndsAt: end})
	require.NoError(t, err)

	cmp, _ := svc.Comparison(ctx, acc, date(2025, time.March, 9))
	row, _ := findRow(cmp, "Uncategorized")
	require.Equal(t, int64(7200), row.ActualSeconds, "2 real hours, not 3 wall-clock hours")
}

func TestComparison_Isolation(t *testing.T) {
	svc, pool, a := setup(t)
	ctx := context.Background()
	b := newAccount(t, pool, "cmp-iso@test")

	_, _ = svc.AddBlock(ctx, a, timeline.BlockInput{
		Kind: timeline.Actual, StartsAt: ts("2025-06-15T09:00:00Z"), EndsAt: ts("2025-06-15T10:00:00Z"),
	})

	cmp, _ := svc.Comparison(ctx, b, date(2025, time.June, 15))
	require.Empty(t, cmp.Categories, "B sees none of A's time")
}
