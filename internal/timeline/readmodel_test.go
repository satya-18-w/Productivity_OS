package timeline_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
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

// TestTimelineRange_MatchesPerDayTimeline proves the batched range endpoint
// returns exactly what N individual Timeline calls would (MX5-range — the
// optimization the frontend's Week/Month views were built for, docs/left.md).
func TestTimelineRange_MatchesPerDayTimeline(t *testing.T) {
	svc, pool, acc := setup(t)
	ctx := context.Background()
	cat := mkCategory(t, pool, acc, "Deep Work")

	_, _ = svc.AddBlock(ctx, acc, timeline.BlockInput{
		Kind: timeline.Planned, StartsAt: ts("2025-06-15T09:00:00Z"), EndsAt: ts("2025-06-15T11:00:00Z"),
		CategoryID: &cat,
	})
	_, _ = svc.AddBlock(ctx, acc, timeline.BlockInput{
		Kind: timeline.Actual, StartsAt: ts("2025-06-17T09:00:00Z"), EndsAt: ts("2025-06-17T10:00:00Z"),
	})
	// outside the queried range
	_, _ = svc.AddBlock(ctx, acc, timeline.BlockInput{
		Kind: timeline.Planned, StartsAt: ts("2025-07-01T09:00:00Z"), EndsAt: ts("2025-07-01T10:00:00Z"),
	})

	rt, err := svc.TimelineRange(ctx, acc, date(2025, time.June, 15), date(2025, time.June, 21))
	require.NoError(t, err)
	require.Equal(t, date(2025, time.June, 15), rt.From)
	require.Equal(t, date(2025, time.June, 21), rt.To)
	require.Len(t, rt.Days, 7, "one entry per day, inclusive")

	require.Len(t, rt.Days[0].Planned, 1, "June 15")
	require.Equal(t, cat, *rt.Days[0].Planned[0].CategoryID)
	require.Empty(t, rt.Days[1].Planned, "June 16 has nothing")
	require.Empty(t, rt.Days[1].Actual)
	require.Len(t, rt.Days[2].Actual, 1, "June 17")

	// cross-check against Timeline for one of the days
	single, err := svc.Timeline(ctx, acc, date(2025, time.June, 15))
	require.NoError(t, err)
	require.Equal(t, single.Planned, rt.Days[0].Planned)
}

func TestTimelineRange_MidnightBlockAppearsOnBothDays(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()
	_, _ = svc.AddBlock(ctx, acc, timeline.BlockInput{
		Kind: timeline.Actual, StartsAt: ts("2025-06-15T23:00:00Z"), EndsAt: ts("2025-06-16T02:00:00Z"),
	})

	rt, err := svc.TimelineRange(ctx, acc, date(2025, time.June, 15), date(2025, time.June, 16))
	require.NoError(t, err)
	require.Len(t, rt.Days[0].Actual, 1)
	require.Len(t, rt.Days[1].Actual, 1)
}

func TestTimelineRange_ToBeforeFrom(t *testing.T) {
	svc, _, acc := setup(t)
	_, err := svc.TimelineRange(context.Background(), acc, date(2025, time.June, 16), date(2025, time.June, 15))
	requireField(t, err, "to")
}

func TestTimelineRange_ExceedsMaxDays(t *testing.T) {
	svc, _, acc := setup(t)
	_, err := svc.TimelineRange(context.Background(), acc, date(2025, time.January, 1), date(2025, time.December, 31))
	requireField(t, err, "to")
}

func TestTimelineRange_Isolation(t *testing.T) {
	svc, pool, a := setup(t)
	ctx := context.Background()
	b := newAccount(t, pool, "range-iso@test")

	_, _ = svc.AddBlock(ctx, a, timeline.BlockInput{
		Kind: timeline.Actual, StartsAt: ts("2025-06-15T09:00:00Z"), EndsAt: ts("2025-06-15T10:00:00Z"),
	})

	rt, err := svc.TimelineRange(ctx, b, date(2025, time.June, 15), date(2025, time.June, 15))
	require.NoError(t, err)
	require.Empty(t, rt.Days[0].Actual, "B sees none of A's blocks")
}

func TestComparison_PerCategoryTotalsAndDifference(t *testing.T) {
	svc, pool, acc := setup(t)
	ctx := context.Background()
	cat := mkCategory(t, pool, acc, "DSA")

	_, _ = svc.AddBlock(ctx, acc, timeline.BlockInput{
		Kind: timeline.Planned, StartsAt: ts("2025-06-15T09:00:00Z"), EndsAt: ts("2025-06-15T11:00:00Z"),
		CategoryID: &cat,
	})
	_, _ = svc.AddBlock(ctx, acc, timeline.BlockInput{
		Kind: timeline.Actual, StartsAt: ts("2025-06-15T09:00:00Z"), EndsAt: ts("2025-06-15T12:00:00Z"),
		CategoryID: &cat,
	})

	cmp, err := svc.Comparison(ctx, acc, date(2025, time.June, 15))
	require.NoError(t, err)
	row, ok := findRow(cmp, "DSA")
	require.True(t, ok)
	require.Equal(t, int64(7200), row.PlannedSeconds)
	require.Equal(t, int64(10800), row.ActualSeconds)
	require.Equal(t, int64(3600), row.DifferenceSeconds)
	require.Equal(t, cat, *row.CategoryID)
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

// TestComparison_TaskLinkedBlockInheritsCategory proves a task-linked block's time
// is attributed to its task's category, not "Uncategorized" (MX-TL analytics
// correctness — the whole point of resolving inheritance in blocksOverlapping).
func TestComparison_TaskLinkedBlockInheritsCategory(t *testing.T) {
	svc, pool, acc := setup(t)
	ctx := context.Background()
	cat := mkCategory(t, pool, acc, "Deep Work")
	task := mkTaskWithCategory(t, pool, acc, cat, "Write report")

	_, err := svc.AddBlock(ctx, acc, timeline.BlockInput{
		Kind: timeline.Actual, StartsAt: ts("2025-06-15T09:00:00Z"), EndsAt: ts("2025-06-15T10:30:00Z"),
		TaskID: &task,
	})
	require.NoError(t, err)

	cmp, err := svc.Comparison(ctx, acc, date(2025, time.June, 15))
	require.NoError(t, err)
	row, ok := findRow(cmp, "Deep Work")
	require.True(t, ok, "task-linked block's time attributed to the task's category")
	require.Equal(t, cat, *row.CategoryID)
	require.Equal(t, int64(5400), row.ActualSeconds)

	_, uncategorized := findRow(cmp, "Uncategorized")
	require.False(t, uncategorized, "must not also appear as Uncategorized")
}

// TestTimeline_TaskLinkedBlockShowsInheritedCategoryName proves the Day timeline
// view resolves the same inheritance for display (v1.md §5).
func TestTimeline_TaskLinkedBlockShowsInheritedCategoryName(t *testing.T) {
	svc, pool, acc := setup(t)
	ctx := context.Background()
	cat := mkCategory(t, pool, acc, "Deep Work")
	task := mkTaskWithCategory(t, pool, acc, cat, "Write report")

	_, err := svc.AddBlock(ctx, acc, timeline.BlockInput{
		Kind: timeline.Planned, StartsAt: ts("2025-06-15T09:00:00Z"), EndsAt: ts("2025-06-15T10:00:00Z"),
		TaskID: &task,
	})
	require.NoError(t, err)

	tl, err := svc.Timeline(ctx, acc, date(2025, time.June, 15))
	require.NoError(t, err)
	require.Len(t, tl.Planned, 1)
	require.Equal(t, cat, *tl.Planned[0].CategoryID)
	require.Equal(t, "Deep Work", *tl.Planned[0].CategoryName)
	require.Equal(t, task, *tl.Planned[0].TaskID)
}

func TestComparison_OverlappingBlocksAreSummed(t *testing.T) {
	svc, pool, acc := setup(t)
	ctx := context.Background()
	cat := mkCategory(t, pool, acc, "Work")

	for _, span := range [][2]string{
		{"2025-06-15T09:00:00Z", "2025-06-15T10:00:00Z"},
		{"2025-06-15T09:30:00Z", "2025-06-15T10:30:00Z"}, // overlaps the first
	} {
		_, err := svc.AddBlock(ctx, acc, timeline.BlockInput{
			Kind: timeline.Actual, StartsAt: ts(span[0]), EndsAt: ts(span[1]), CategoryID: &cat,
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

// findRangeRow returns the range-comparison row for a category name.
func findRangeRow(cmp timeline.RangeComparison, name string) (timeline.CategoryTotals, bool) {
	for _, c := range cmp.Categories {
		if c.CategoryName == name {
			return c, true
		}
	}
	return timeline.CategoryTotals{}, false
}

func TestComparisonRange_SumsAcrossDays(t *testing.T) {
	svc, pool, acc := setup(t)
	ctx := context.Background()
	cat := mkCategory(t, pool, acc, "Work")

	_, _ = svc.AddBlock(ctx, acc, timeline.BlockInput{
		Kind: timeline.Actual, StartsAt: ts("2025-06-15T09:00:00Z"), EndsAt: ts("2025-06-15T10:00:00Z"), CategoryID: &cat,
	})
	_, _ = svc.AddBlock(ctx, acc, timeline.BlockInput{
		Kind: timeline.Actual, StartsAt: ts("2025-06-16T09:00:00Z"), EndsAt: ts("2025-06-16T11:00:00Z"), CategoryID: &cat,
	})
	_, _ = svc.AddBlock(ctx, acc, timeline.BlockInput{
		// outside the queried range
		Kind: timeline.Actual, StartsAt: ts("2025-06-20T09:00:00Z"), EndsAt: ts("2025-06-20T10:00:00Z"), CategoryID: &cat,
	})

	cmp, err := svc.ComparisonRange(ctx, acc, date(2025, time.June, 15), date(2025, time.June, 16))
	require.NoError(t, err)
	require.Equal(t, date(2025, time.June, 15), cmp.From)
	require.Equal(t, date(2025, time.June, 16), cmp.To)
	row, ok := findRangeRow(cmp, "Work")
	require.True(t, ok)
	require.Equal(t, int64(10800), row.ActualSeconds, "1h day 15 + 2h day 16, day 20 excluded")
}

func TestComparisonRange_SingleDayMatchesComparison(t *testing.T) {
	svc, pool, acc := setup(t)
	ctx := context.Background()
	cat := mkCategory(t, pool, acc, "DSA")

	_, _ = svc.AddBlock(ctx, acc, timeline.BlockInput{
		Kind: timeline.Planned, StartsAt: ts("2025-06-15T09:00:00Z"), EndsAt: ts("2025-06-15T11:00:00Z"), CategoryID: &cat,
	})

	day, err := svc.Comparison(ctx, acc, date(2025, time.June, 15))
	require.NoError(t, err)
	rng, err := svc.ComparisonRange(ctx, acc, date(2025, time.June, 15), date(2025, time.June, 15))
	require.NoError(t, err)

	dayRow, _ := findRow(day, "DSA")
	rngRow, _ := findRangeRow(rng, "DSA")
	require.Equal(t, dayRow.PlannedSeconds, rngRow.PlannedSeconds, "from == to matches the single-date query")
}

func TestComparisonRange_ToBeforeFrom(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()

	_, err := svc.ComparisonRange(ctx, acc, date(2025, time.June, 16), date(2025, time.June, 15))
	var verr *timeline.ValidationError
	require.ErrorAs(t, err, &verr)
	require.Contains(t, verr.Fields, "to")
}

func TestComparisonRange_DSTTransitionCountsRealElapsedTime(t *testing.T) {
	ny, err := timezone.LoadLocation("America/New_York")
	require.NoError(t, err)
	svc, _, acc := setupZone(t, ny)
	ctx := context.Background()

	// same spring-forward block as the single-day test, queried through a range
	// that spans the transition date.
	start := time.Date(2025, time.March, 9, 1, 0, 0, 0, ny)
	end := time.Date(2025, time.March, 9, 4, 0, 0, 0, ny)
	_, err = svc.AddBlock(ctx, acc, timeline.BlockInput{Kind: timeline.Actual, StartsAt: start, EndsAt: end})
	require.NoError(t, err)

	cmp, err := svc.ComparisonRange(ctx, acc, date(2025, time.March, 8), date(2025, time.March, 10))
	require.NoError(t, err)
	row, _ := findRangeRow(cmp, "Uncategorized")
	require.Equal(t, int64(7200), row.ActualSeconds, "2 real hours, not 3 wall-clock hours")
}

func TestComparisonRange_Isolation(t *testing.T) {
	svc, pool, a := setup(t)
	ctx := context.Background()
	b := newAccount(t, pool, "range-cmp-iso@test")

	_, _ = svc.AddBlock(ctx, a, timeline.BlockInput{
		Kind: timeline.Actual, StartsAt: ts("2025-06-15T09:00:00Z"), EndsAt: ts("2025-06-15T10:00:00Z"),
	})

	cmp, err := svc.ComparisonRange(ctx, b, date(2025, time.June, 15), date(2025, time.June, 16))
	require.NoError(t, err)
	require.Empty(t, cmp.Categories, "B sees none of A's time")
}

// findDayTotal returns the DayTotal for a date.
func findDayTotal(totals []timeline.DayTotal, d timezone.Date) (timeline.DayTotal, bool) {
	for _, t := range totals {
		if t.Date == d {
			return t, true
		}
	}
	return timeline.DayTotal{}, false
}

func TestDailyActualTotals_PerDaySums(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()

	_, _ = svc.AddBlock(ctx, acc, timeline.BlockInput{
		Kind: timeline.Actual, StartsAt: ts("2025-06-15T09:00:00Z"), EndsAt: ts("2025-06-15T10:00:00Z"),
	})
	_, _ = svc.AddBlock(ctx, acc, timeline.BlockInput{
		Kind: timeline.Actual, StartsAt: ts("2025-06-16T09:00:00Z"), EndsAt: ts("2025-06-16T11:00:00Z"),
	})
	_, _ = svc.AddBlock(ctx, acc, timeline.BlockInput{
		// planned, not actual — must be excluded
		Kind: timeline.Planned, StartsAt: ts("2025-06-15T12:00:00Z"), EndsAt: ts("2025-06-15T13:00:00Z"),
	})

	totals, err := svc.DailyActualTotals(ctx, acc, date(2025, time.June, 15), date(2025, time.June, 16))
	require.NoError(t, err)
	require.Len(t, totals, 2, "one row per day in range, inclusive")

	d15, ok := findDayTotal(totals, date(2025, time.June, 15))
	require.True(t, ok)
	require.Equal(t, int64(3600), d15.ActualSeconds, "1h actual; the planned block is excluded")

	d16, ok := findDayTotal(totals, date(2025, time.June, 16))
	require.True(t, ok)
	require.Equal(t, int64(7200), d16.ActualSeconds)
}

func TestDailyActualTotals_MidnightSpanningBlockSplitsAcrossDays(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()

	_, _ = svc.AddBlock(ctx, acc, timeline.BlockInput{
		Kind: timeline.Actual, StartsAt: ts("2025-06-15T23:00:00Z"), EndsAt: ts("2025-06-16T02:00:00Z"),
	})

	totals, err := svc.DailyActualTotals(ctx, acc, date(2025, time.June, 15), date(2025, time.June, 16))
	require.NoError(t, err)
	d15, _ := findDayTotal(totals, date(2025, time.June, 15))
	d16, _ := findDayTotal(totals, date(2025, time.June, 16))
	require.Equal(t, int64(3600), d15.ActualSeconds, "1h on day N")
	require.Equal(t, int64(7200), d16.ActualSeconds, "2h on day N+1")
}

func TestDailyActualTotals_DSTTransitionDayCountsRealElapsedTime(t *testing.T) {
	ny, err := timezone.LoadLocation("America/New_York")
	require.NoError(t, err)
	svc, _, acc := setupZone(t, ny)
	ctx := context.Background()

	// same spring-forward block as the single-day/range DST tests: 01:00 -> 04:00
	// local, but 02:00-03:00 does not exist, so only 2 real hours elapse.
	start := time.Date(2025, time.March, 9, 1, 0, 0, 0, ny)
	end := time.Date(2025, time.March, 9, 4, 0, 0, 0, ny)
	_, err = svc.AddBlock(ctx, acc, timeline.BlockInput{Kind: timeline.Actual, StartsAt: start, EndsAt: end})
	require.NoError(t, err)

	totals, err := svc.DailyActualTotals(ctx, acc, date(2025, time.March, 8), date(2025, time.March, 10))
	require.NoError(t, err)
	require.Len(t, totals, 3)
	transitionDay, ok := findDayTotal(totals, date(2025, time.March, 9))
	require.True(t, ok)
	require.Equal(t, int64(7200), transitionDay.ActualSeconds, "2 real hours, not 3 wall-clock hours")
}

func TestDailyActualTotals_ToBeforeFrom(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()

	_, err := svc.DailyActualTotals(ctx, acc, date(2025, time.June, 16), date(2025, time.June, 15))
	var verr *timeline.ValidationError
	require.ErrorAs(t, err, &verr)
	require.Contains(t, verr.Fields, "to")
}

func TestDailyActualTotals_Isolation(t *testing.T) {
	svc, pool, a := setup(t)
	ctx := context.Background()
	b := newAccount(t, pool, "daily-total-iso@test")

	_, _ = svc.AddBlock(ctx, a, timeline.BlockInput{
		Kind: timeline.Actual, StartsAt: ts("2025-06-15T09:00:00Z"), EndsAt: ts("2025-06-15T10:00:00Z"),
	})

	totals, err := svc.DailyActualTotals(ctx, b, date(2025, time.June, 15), date(2025, time.June, 15))
	require.NoError(t, err)
	require.Equal(t, int64(0), totals[0].ActualSeconds, "B sees none of A's time")
}

func TestListAllBlocks(t *testing.T) {
	svc, pool, acc := setup(t)
	ctx := context.Background()
	other := newAccount(t, pool, "other-listall@test")

	_, _ = svc.AddBlock(ctx, acc, timeline.BlockInput{
		Kind: timeline.Planned, StartsAt: ts("2025-06-15T09:00:00Z"), EndsAt: ts("2025-06-15T10:00:00Z"),
	})
	_, _ = svc.AddBlock(ctx, acc, timeline.BlockInput{
		// far outside any range other tests would query — proves this is unbounded
		Kind: timeline.Actual, StartsAt: ts("2030-01-01T09:00:00Z"), EndsAt: ts("2030-01-01T10:00:00Z"),
	})
	_, _ = svc.AddBlock(ctx, other, timeline.BlockInput{
		Kind: timeline.Planned, StartsAt: ts("2025-06-15T09:00:00Z"), EndsAt: ts("2025-06-15T10:00:00Z"),
	})

	all, err := svc.ListAllBlocks(ctx, acc)
	require.NoError(t, err)
	require.Len(t, all, 2, "both of the caller's blocks, none of other's, regardless of date")
}

// TestBlocksForTask covers v1.md §7's reverse-lookup view (the "GET
// /api/tasks/{id}/blocks" endpoint) — every block linked to a task, across any
// date, with the task's own category resolved for display.
func TestBlocksForTask(t *testing.T) {
	svc, pool, acc := setup(t)
	ctx := context.Background()
	cat := mkCategory(t, pool, acc, "Deep Work")
	task := mkTaskWithCategory(t, pool, acc, cat, "Write report")
	otherTask := mkTask(t, pool, acc, "Unrelated")

	planned, err := svc.AddBlock(ctx, acc, timeline.BlockInput{
		Kind: timeline.Planned, StartsAt: ts("2025-06-15T09:00:00Z"), EndsAt: ts("2025-06-15T10:00:00Z"),
		TaskID: &task,
	})
	require.NoError(t, err)
	actual, err := svc.AddBlock(ctx, acc, timeline.BlockInput{
		// a much later date — proves this is unbounded, unlike Timeline/TimelineRange
		Kind: timeline.Actual, StartsAt: ts("2025-08-01T14:00:00Z"), EndsAt: ts("2025-08-01T15:00:00Z"),
		TaskID: &task,
	})
	require.NoError(t, err)
	_, err = svc.AddBlock(ctx, acc, timeline.BlockInput{
		Kind: timeline.Planned, StartsAt: ts("2025-06-15T11:00:00Z"), EndsAt: ts("2025-06-15T12:00:00Z"),
		TaskID: &otherTask,
	})
	require.NoError(t, err)

	blocks, err := svc.BlocksForTask(ctx, acc, task)
	require.NoError(t, err)
	require.Len(t, blocks, 2, "both of this task's blocks, ordered by start, not the other task's")
	require.Equal(t, planned.ID, blocks[0].ID)
	require.Equal(t, actual.ID, blocks[1].ID)
	require.Equal(t, cat, *blocks[0].CategoryID, "inherited from the task")
	require.Equal(t, "Deep Work", *blocks[0].CategoryName)
	require.Equal(t, task, *blocks[0].TaskID)
}

func TestBlocksForTask_EmptyAndNotFound(t *testing.T) {
	svc, pool, acc := setup(t)
	ctx := context.Background()
	other := newAccount(t, pool, "other-blocksfortask@test")

	noBlocks := mkTask(t, pool, acc, "No blocks yet")
	blocks, err := svc.BlocksForTask(ctx, acc, noBlocks)
	require.NoError(t, err)
	require.Empty(t, blocks)

	theirs := mkTask(t, pool, other, "Theirs")
	_, err = svc.BlocksForTask(ctx, acc, theirs)
	require.ErrorIs(t, err, timeline.ErrTaskNotFound, "another account's task")

	_, err = svc.BlocksForTask(ctx, acc, uuid.New())
	require.ErrorIs(t, err, timeline.ErrTaskNotFound, "unknown task")
}

func TestCountByCategory(t *testing.T) {
	svc, pool, acc := setup(t)
	ctx := context.Background()
	other := newAccount(t, pool, "other-count@test")

	catA := mkCategory(t, pool, acc, "A")
	catB := mkCategory(t, pool, acc, "B")
	otherCat := mkCategory(t, pool, other, "Other")

	_, _ = svc.AddBlock(ctx, acc, timeline.BlockInput{
		Kind: timeline.Planned, StartsAt: ts("2025-06-15T09:00:00Z"), EndsAt: ts("2025-06-15T10:00:00Z"), CategoryID: &catA,
	})
	_, _ = svc.AddBlock(ctx, acc, timeline.BlockInput{
		Kind: timeline.Actual, StartsAt: ts("2025-06-15T09:00:00Z"), EndsAt: ts("2025-06-15T10:00:00Z"), CategoryID: &catA,
	})
	_, _ = svc.AddBlock(ctx, acc, timeline.BlockInput{
		Kind: timeline.Planned, StartsAt: ts("2025-06-16T09:00:00Z"), EndsAt: ts("2025-06-16T10:00:00Z"), CategoryID: &catB,
	})
	_, _ = svc.AddBlock(ctx, acc, timeline.BlockInput{
		Kind: timeline.Planned, StartsAt: ts("2025-06-17T09:00:00Z"), EndsAt: ts("2025-06-17T10:00:00Z"),
	}) // uncategorized
	_, _ = svc.AddBlock(ctx, other, timeline.BlockInput{
		Kind: timeline.Planned, StartsAt: ts("2025-06-15T09:00:00Z"), EndsAt: ts("2025-06-15T10:00:00Z"), CategoryID: &otherCat,
	})

	counts, err := svc.CountByCategory(ctx, acc)
	require.NoError(t, err)
	require.Equal(t, map[uuid.UUID]int{catA: 2, catB: 1}, counts)
}

// TestCountByCategory_TaskLinkedBlockCountsTowardInheritedCategory proves a
// task-linked block still contributes to its (inherited) category's block total in
// the categories overview, even though it stores no category_id of its own
// (MX-TL analytics correctness).
func TestCountByCategory_TaskLinkedBlockCountsTowardInheritedCategory(t *testing.T) {
	svc, pool, acc := setup(t)
	ctx := context.Background()
	cat := mkCategory(t, pool, acc, "Deep Work")
	task := mkTaskWithCategory(t, pool, acc, cat, "Write report")

	_, err := svc.AddBlock(ctx, acc, timeline.BlockInput{
		Kind: timeline.Planned, StartsAt: ts("2025-06-15T09:00:00Z"), EndsAt: ts("2025-06-15T10:00:00Z"),
		TaskID: &task,
	})
	require.NoError(t, err)
	_, err = svc.AddBlock(ctx, acc, timeline.BlockInput{
		Kind: timeline.Actual, StartsAt: ts("2025-06-16T09:00:00Z"), EndsAt: ts("2025-06-16T10:00:00Z"),
		CategoryID: &cat,
	})
	require.NoError(t, err)

	counts, err := svc.CountByCategory(ctx, acc)
	require.NoError(t, err)
	require.Equal(t, 2, counts[cat], "one direct + one task-linked block, both counted")
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
