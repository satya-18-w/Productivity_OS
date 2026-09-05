package reports_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/satya-18-w/productivity-os/internal/categories"
	"github.com/satya-18-w/productivity-os/internal/goals"
	"github.com/satya-18-w/productivity-os/internal/habits"
	"github.com/satya-18-w/productivity-os/internal/platform/pgtest"
	"github.com/satya-18-w/productivity-os/internal/platform/timezone"
	"github.com/satya-18-w/productivity-os/internal/reports"
	"github.com/satya-18-w/productivity-os/internal/tasks"
	"github.com/satya-18-w/productivity-os/internal/timeline"
)

type fakeZone struct{ loc *time.Location }

func (f fakeZone) Zone(context.Context, uuid.UUID) (*time.Location, error) {
	if f.loc == nil {
		return time.UTC, nil
	}
	return f.loc, nil
}

// env wires reports over the real domain modules against one shared pool — the
// composition is exactly what cmd/server does, so this is as much an integration
// test of the wiring as of reports' own logic.
type env struct {
	reports  reports.Service
	timeline timeline.Service
	habits   habits.Service
	tasks    tasks.Service
	pool     *pgxpool.Pool
	acc      uuid.UUID
}

func setup(t *testing.T) env {
	t.Helper()
	return setupZone(t, fakeZone{})
}

func setupZone(t *testing.T, zone fakeZone) env {
	t.Helper()
	pool := pgtest.Pool(t)
	acc := newAccount(t, pool, "owner@test")
	catSvc := categories.NewService(pool)
	goalsSvc := goals.NewService(pool, catSvc)
	tasksSvc := tasks.NewService(pool, catSvc, goalsSvc)
	timelineSvc := timeline.NewService(pool, zone, catSvc, tasksSvc)
	habitsSvc := habits.NewService(pool, zone, catSvc)
	reportsSvc := reports.NewService(timelineSvc, habitsSvc, tasksSvc, zone)
	return env{reportsSvc, timelineSvc, habitsSvc, tasksSvc, pool, acc}
}

func newAccount(t *testing.T, pool *pgxpool.Pool, email string) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	require.NoError(t, pool.QueryRow(context.Background(),
		`INSERT INTO accounts (email, password_hash, timezone) VALUES ($1, 'x', 'UTC') RETURNING id`,
		email).Scan(&id))
	return id
}

func mkCategory(t *testing.T, pool *pgxpool.Pool, acc uuid.UUID, name string) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	require.NoError(t, pool.QueryRow(context.Background(),
		`INSERT INTO categories (account_id, name) VALUES ($1, $2) RETURNING id`, acc, name).Scan(&id))
	return id
}

func ts(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

func date(y int, m time.Month, d int) timezone.Date { return timezone.Date{Year: y, Month: m, Day: d} }

func findCategoryTime(rows []reports.CategoryTimeRow, name string) (reports.CategoryTimeRow, bool) {
	for _, r := range rows {
		if r.CategoryName == name {
			return r, true
		}
	}
	return reports.CategoryTimeRow{}, false
}

func findCategoryTotal(rows []timeline.CategoryTotals, name string) (timeline.CategoryTotals, bool) {
	for _, r := range rows {
		if r.CategoryName == name {
			return r, true
		}
	}
	return timeline.CategoryTotals{}, false
}

func findHabitRow(rows []reports.HabitCompletionRow, id uuid.UUID) (reports.HabitCompletionRow, bool) {
	for _, r := range rows {
		if r.HabitID == id {
			return r, true
		}
	}
	return reports.HabitCompletionRow{}, false
}

func TestTimeByCategory(t *testing.T) {
	e := setup(t)
	ctx := context.Background()
	cat := mkCategory(t, e.pool, e.acc, "Deep Work")

	_, _ = e.timeline.AddBlock(ctx, e.acc, timeline.BlockInput{
		Kind: timeline.Planned, StartsAt: ts("2025-06-15T09:00:00Z"), EndsAt: ts("2025-06-15T11:00:00Z"), CategoryID: &cat,
	})
	_, _ = e.timeline.AddBlock(ctx, e.acc, timeline.BlockInput{
		Kind: timeline.Actual, StartsAt: ts("2025-06-15T09:00:00Z"), EndsAt: ts("2025-06-15T12:00:00Z"), CategoryID: &cat,
	})

	rows, err := e.reports.TimeByCategory(ctx, e.acc, date(2025, time.June, 15), date(2025, time.June, 15))
	require.NoError(t, err)
	row, ok := findCategoryTime(rows, "Deep Work")
	require.True(t, ok)
	require.Equal(t, int64(10800), row.ActualSeconds, "actual only — planned is not part of this report")
}

func TestPlannedVsActualByCategory(t *testing.T) {
	e := setup(t)
	ctx := context.Background()
	cat := mkCategory(t, e.pool, e.acc, "DSA")

	_, _ = e.timeline.AddBlock(ctx, e.acc, timeline.BlockInput{
		Kind: timeline.Planned, StartsAt: ts("2025-06-15T09:00:00Z"), EndsAt: ts("2025-06-15T11:00:00Z"), CategoryID: &cat,
	})
	_, _ = e.timeline.AddBlock(ctx, e.acc, timeline.BlockInput{
		Kind: timeline.Actual, StartsAt: ts("2025-06-15T09:00:00Z"), EndsAt: ts("2025-06-15T12:00:00Z"), CategoryID: &cat,
	})

	rows, err := e.reports.PlannedVsActualByCategory(ctx, e.acc, date(2025, time.June, 15), date(2025, time.June, 15))
	require.NoError(t, err)
	row, ok := findCategoryTotal(rows, "DSA")
	require.True(t, ok)
	require.Equal(t, int64(7200), row.PlannedSeconds)
	require.Equal(t, int64(10800), row.ActualSeconds)
	require.Equal(t, int64(3600), row.DifferenceSeconds)
}

func TestHabitCompletion(t *testing.T) {
	e := setup(t)
	ctx := context.Background()

	h, err := e.habits.CreateHabit(ctx, e.acc, habits.HabitInput{Name: "Meditate"})
	require.NoError(t, err)

	// a 7-day range starting today — the habit is created "today" (test run time),
	// so the range must start no earlier than today or RangeDays would be capped
	// below 7. Future dates are allowed (Q9).
	today := timezone.Today(time.UTC)
	from, to := today, today.AddDays(6)
	require.NoError(t, e.habits.MarkComplete(ctx, e.acc, h.ID, from))
	require.NoError(t, e.habits.MarkComplete(ctx, e.acc, h.ID, from.AddDays(1)))
	require.NoError(t, e.habits.MarkComplete(ctx, e.acc, h.ID, from.AddDays(3)))

	rows, err := e.reports.HabitCompletion(ctx, e.acc, from, to)
	require.NoError(t, err)
	row, ok := findHabitRow(rows, h.ID)
	require.True(t, ok)
	require.Equal(t, 3, row.CompletedDays)
	require.Equal(t, 7, row.RangeDays)
}

func TestHabitCompletion_RangeDaysBoundedByCreation(t *testing.T) {
	e := setup(t)
	ctx := context.Background()
	today := timezone.Today(time.UTC)
	// a 10-day range ending today; the habit is created "today" (test run time), so
	// only 1 of those 10 days falls inside its active window.
	from, to := today.AddDays(-9), today

	h, err := e.habits.CreateHabit(ctx, e.acc, habits.HabitInput{Name: "New habit"})
	require.NoError(t, err)
	require.NoError(t, e.habits.MarkComplete(ctx, e.acc, h.ID, today))

	rows, err := e.reports.HabitCompletion(ctx, e.acc, from, to)
	require.NoError(t, err)
	row, ok := findHabitRow(rows, h.ID)
	require.True(t, ok)
	require.Equal(t, 1, row.CompletedDays)
	require.Equal(t, 1, row.RangeDays, "the habit only existed for 1 of the 10 days in range")
}

func TestHabitCompletion_RangeDaysBoundedByArchive(t *testing.T) {
	e := setup(t)
	ctx := context.Background()
	today := timezone.Today(time.UTC)
	from, to := today, today.AddDays(9) // a 10-day range starting today

	h, err := e.habits.CreateHabit(ctx, e.acc, habits.HabitInput{Name: "Soon archived"})
	require.NoError(t, err)
	require.NoError(t, e.habits.MarkComplete(ctx, e.acc, h.ID, today))
	require.NoError(t, e.habits.ArchiveHabit(ctx, e.acc, h.ID)) // archived "today"

	rows, err := e.reports.HabitCompletion(ctx, e.acc, from, to)
	require.NoError(t, err)
	row, ok := findHabitRow(rows, h.ID)
	require.True(t, ok)
	require.Equal(t, 1, row.CompletedDays)
	require.Equal(t, 0, row.RangeDays,
		"archived the same day it was created — its active window (up to the day before archiving) excludes today")
}

func TestTaskThroughput(t *testing.T) {
	e := setup(t)
	ctx := context.Background()

	a, _ := e.tasks.CreateTask(ctx, e.acc, tasks.TaskInput{Title: "a"})
	b, _ := e.tasks.CreateTask(ctx, e.acc, tasks.TaskInput{Title: "b"})
	require.NoError(t, e.tasks.MoveTask(ctx, e.acc, a.ID, tasks.Done))
	require.NoError(t, e.tasks.MoveTask(ctx, e.acc, b.ID, tasks.Todo))
	require.NoError(t, e.tasks.MoveTask(ctx, e.acc, b.ID, tasks.Done))
	// bounce a out and back into DONE — still counts once
	require.NoError(t, e.tasks.MoveTask(ctx, e.acc, a.ID, tasks.Todo))
	require.NoError(t, e.tasks.MoveTask(ctx, e.acc, a.ID, tasks.Done))

	today := timezone.Today(time.UTC)
	report, err := e.reports.TaskThroughput(ctx, e.acc, today, today)
	require.NoError(t, err)
	require.Equal(t, 2, report.DoneCount)
	require.Equal(t, today, report.From)
	require.Equal(t, today, report.To)
}

func TestDailyActualTotals(t *testing.T) {
	e := setup(t)
	ctx := context.Background()

	_, _ = e.timeline.AddBlock(ctx, e.acc, timeline.BlockInput{
		Kind: timeline.Actual, StartsAt: ts("2025-06-15T09:00:00Z"), EndsAt: ts("2025-06-15T10:00:00Z"),
	})
	_, _ = e.timeline.AddBlock(ctx, e.acc, timeline.BlockInput{
		Kind: timeline.Actual, StartsAt: ts("2025-06-16T09:00:00Z"), EndsAt: ts("2025-06-16T11:00:00Z"),
	})

	totals, err := e.reports.DailyActualTotals(ctx, e.acc, date(2025, time.June, 15), date(2025, time.June, 16))
	require.NoError(t, err)
	require.Len(t, totals, 2)
	require.Equal(t, int64(3600), totals[0].ActualSeconds)
	require.Equal(t, int64(7200), totals[1].ActualSeconds)
}

func TestReports_DSTCrossingRange(t *testing.T) {
	ny, err := timezone.LoadLocation("America/New_York")
	require.NoError(t, err)
	e := setupZone(t, fakeZone{loc: ny})
	ctx := context.Background()

	// 01:00 -> 04:00 local on the spring-forward day: 02:00-03:00 does not exist,
	// so only 2 real hours elapse.
	start := time.Date(2025, time.March, 9, 1, 0, 0, 0, ny)
	end := time.Date(2025, time.March, 9, 4, 0, 0, 0, ny)
	_, err = e.timeline.AddBlock(ctx, e.acc, timeline.BlockInput{Kind: timeline.Actual, StartsAt: start, EndsAt: end})
	require.NoError(t, err)

	from, to := date(2025, time.March, 1), date(2025, time.March, 31) // a one-month range crossing the transition

	rows, err := e.reports.TimeByCategory(ctx, e.acc, from, to)
	require.NoError(t, err)
	row, ok := findCategoryTime(rows, "Uncategorized")
	require.True(t, ok)
	require.Equal(t, int64(7200), row.ActualSeconds, "2 real hours, not 3 wall-clock hours")

	totals, err := e.reports.DailyActualTotals(ctx, e.acc, from, to)
	require.NoError(t, err)
	require.Len(t, totals, 31)
	var transitionDayTotal int64
	for _, d := range totals {
		if d.Date == date(2025, time.March, 9) {
			transitionDayTotal = d.ActualSeconds
		}
	}
	require.Equal(t, int64(7200), transitionDayTotal, "the transition day itself also reflects real elapsed time")
}

func TestReports_ToBeforeFrom(t *testing.T) {
	e := setup(t)
	ctx := context.Background()
	from, to := date(2025, time.June, 16), date(2025, time.June, 15)

	requireRangeError := func(t *testing.T, err error) {
		t.Helper()
		var verr *reports.ValidationError
		require.ErrorAs(t, err, &verr)
		require.Contains(t, verr.Fields, "to")
	}

	_, err := e.reports.TimeByCategory(ctx, e.acc, from, to)
	requireRangeError(t, err)
	_, err = e.reports.PlannedVsActualByCategory(ctx, e.acc, from, to)
	requireRangeError(t, err)
	_, err = e.reports.HabitCompletion(ctx, e.acc, from, to)
	requireRangeError(t, err)
	_, err = e.reports.TaskThroughput(ctx, e.acc, from, to)
	requireRangeError(t, err)
	_, err = e.reports.DailyActualTotals(ctx, e.acc, from, to)
	requireRangeError(t, err)
	_, err = e.reports.Report(ctx, e.acc, from, to)
	requireRangeError(t, err)
}

func TestReport_Combined(t *testing.T) {
	e := setup(t)
	ctx := context.Background()
	cat := mkCategory(t, e.pool, e.acc, "Deep Work")

	// The habit and task below are timestamped "today" (server-generated), so the
	// queried range must include today — use yesterday..today rather than a fixed
	// historical date.
	today := timezone.Today(time.UTC)
	from, to := today.AddDays(-1), today
	blockTime := func(hh, mm int) time.Time {
		return time.Date(from.Year, from.Month, from.Day, hh, mm, 0, 0, time.UTC)
	}

	_, _ = e.timeline.AddBlock(ctx, e.acc, timeline.BlockInput{
		Kind: timeline.Planned, StartsAt: blockTime(9, 0), EndsAt: blockTime(11, 0), CategoryID: &cat,
	})
	_, _ = e.timeline.AddBlock(ctx, e.acc, timeline.BlockInput{
		Kind: timeline.Actual, StartsAt: blockTime(9, 0), EndsAt: blockTime(12, 0), CategoryID: &cat,
	})
	_, _ = e.timeline.AddBlock(ctx, e.acc, timeline.BlockInput{
		// uncategorized actual block — must appear in time_by_category but not planned_vs_actual
		Kind: timeline.Actual, StartsAt: blockTime(13, 0), EndsAt: blockTime(14, 0),
	})
	h, _ := e.habits.CreateHabit(ctx, e.acc, habits.HabitInput{Name: "Meditate"})
	require.NoError(t, e.habits.MarkComplete(ctx, e.acc, h.ID, today))
	task, _ := e.tasks.CreateTask(ctx, e.acc, tasks.TaskInput{Title: "t"})
	require.NoError(t, e.tasks.MoveTask(ctx, e.acc, task.ID, tasks.Done))

	rep, err := e.reports.Report(ctx, e.acc, from, to)
	require.NoError(t, err)
	require.Equal(t, from, rep.From)
	require.Equal(t, to, rep.To)

	// time_by_category keeps Uncategorized (Q8)
	_, hasUncategorized := findCategoryTime(rep.TimeByCategory, "Uncategorized")
	require.True(t, hasUncategorized)
	tbcRow, ok := findCategoryTime(rep.TimeByCategory, "Deep Work")
	require.True(t, ok)
	require.Equal(t, int64(10800), tbcRow.ActualSeconds)

	// planned_vs_actual excludes Uncategorized
	_, hasUncategorizedPVA := findCategoryTotal(rep.PlannedVsActual, "Uncategorized")
	require.False(t, hasUncategorizedPVA, "planned_vs_actual excludes the Uncategorized bucket")
	pvaRow, ok := findCategoryTotal(rep.PlannedVsActual, "Deep Work")
	require.True(t, ok)
	require.Equal(t, int64(7200), pvaRow.PlannedSeconds)
	require.Equal(t, int64(10800), pvaRow.ActualSeconds)

	habitRow, ok := findHabitRow(rep.HabitCompletion, h.ID)
	require.True(t, ok)
	require.Equal(t, 1, habitRow.CompletedDays)

	require.Equal(t, 1, rep.TaskThroughput)
	require.NotEmpty(t, rep.DailyActualTotals)
}

func TestReport_RangeTooLarge(t *testing.T) {
	e := setup(t)
	ctx := context.Background()

	_, err := e.reports.Report(ctx, e.acc, date(2025, time.January, 1), date(2026, time.January, 2)) // 367 days
	var verr *reports.ValidationError
	require.ErrorAs(t, err, &verr)
	require.Contains(t, verr.Fields, "range")

	// exactly at the bound is fine
	_, err = e.reports.Report(ctx, e.acc, date(2025, time.January, 1), date(2026, time.January, 1)) // 366 days
	require.NoError(t, err)
}

func TestReports_Isolation(t *testing.T) {
	e := setup(t)
	ctx := context.Background()
	other := newAccount(t, e.pool, "other@test")

	cat := mkCategory(t, e.pool, e.acc, "Mine")
	_, _ = e.timeline.AddBlock(ctx, e.acc, timeline.BlockInput{
		Kind: timeline.Actual, StartsAt: ts("2025-06-15T09:00:00Z"), EndsAt: ts("2025-06-15T10:00:00Z"), CategoryID: &cat,
	})
	h, _ := e.habits.CreateHabit(ctx, e.acc, habits.HabitInput{Name: "Mine"})
	require.NoError(t, e.habits.MarkComplete(ctx, e.acc, h.ID, date(2025, time.June, 15)))
	task, _ := e.tasks.CreateTask(ctx, e.acc, tasks.TaskInput{Title: "Mine"})
	require.NoError(t, e.tasks.MoveTask(ctx, e.acc, task.ID, tasks.Done))

	from, to := date(2025, time.June, 15), date(2025, time.June, 15)

	timeRows, err := e.reports.TimeByCategory(ctx, other, from, to)
	require.NoError(t, err)
	require.Empty(t, timeRows, "other sees none of the caller's category time")

	habitRows, err := e.reports.HabitCompletion(ctx, other, from, to)
	require.NoError(t, err)
	require.Empty(t, habitRows)

	throughput, err := e.reports.TaskThroughput(ctx, other, from, to)
	require.NoError(t, err)
	require.Equal(t, 0, throughput.DoneCount)

	totals, err := e.reports.DailyActualTotals(ctx, other, from, to)
	require.NoError(t, err)
	require.Equal(t, int64(0), totals[0].ActualSeconds)

	rep, err := e.reports.Report(ctx, other, from, to)
	require.NoError(t, err)
	require.Empty(t, rep.TimeByCategory)
	require.Empty(t, rep.PlannedVsActual)
	require.Empty(t, rep.HabitCompletion)
	require.Equal(t, 0, rep.TaskThroughput)
}
