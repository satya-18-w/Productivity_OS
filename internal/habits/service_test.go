package habits_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/satya-18-w/productivity-os/internal/categories"
	"github.com/satya-18-w/productivity-os/internal/habits"
	"github.com/satya-18-w/productivity-os/internal/platform/pgtest"
	"github.com/satya-18-w/productivity-os/internal/platform/timezone"
)

type fakeZone struct{}

func (fakeZone) Zone(context.Context, uuid.UUID) (*time.Location, error) { return time.UTC, nil }

func setup(t *testing.T) (habits.Service, *pgxpool.Pool, uuid.UUID) {
	t.Helper()
	pool := pgtest.Pool(t)
	return habits.NewService(pool, fakeZone{}, categories.NewService(pool)), pool, newAccount(t, pool, "owner@test")
}

// mkCategory inserts an active category directly (habits tests only need one to
// exist and be assignable — category CRUD is exercised in the categories package).
func mkCategory(t *testing.T, pool *pgxpool.Pool, acc uuid.UUID, name string) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	require.NoError(t, pool.QueryRow(context.Background(),
		`INSERT INTO categories (account_id, name) VALUES ($1, $2) RETURNING id`, acc, name).Scan(&id))
	return id
}

func archiveCategory(t *testing.T, pool *pgxpool.Pool, id uuid.UUID) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		`UPDATE categories SET archived_at = now() WHERE id = $1`, id)
	require.NoError(t, err)
}

func newAccount(t *testing.T, pool *pgxpool.Pool, email string) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	require.NoError(t, pool.QueryRow(context.Background(),
		`INSERT INTO accounts (email, password_hash, timezone) VALUES ($1, 'x', 'UTC') RETURNING id`,
		email).Scan(&id))
	return id
}

func today() timezone.Date { return timezone.Today(time.UTC) }

func strp(s string) *string { return &s }

func findView(vs []habits.HabitView, id uuid.UUID) (habits.HabitView, bool) {
	for _, v := range vs {
		if v.ID == id {
			return v, true
		}
	}
	return habits.HabitView{}, false
}

func TestCreateHabit(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()

	h, err := svc.CreateHabit(ctx, acc, habits.HabitInput{Name: "  Read 20 pages  "})
	require.NoError(t, err)
	require.Equal(t, "Read 20 pages", h.Name)
	require.Nil(t, h.ArchivedAt)

	_, err = svc.CreateHabit(ctx, acc, habits.HabitInput{Name: "   "})
	var verr *habits.ValidationError
	require.ErrorAs(t, err, &verr)
	require.Contains(t, verr.Fields, "name")
}

func TestCreateHabit_WithTarget(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()

	h, err := svc.CreateHabit(ctx, acc, habits.HabitInput{Name: "Workout", Target: strp(" 30 minutes ")})
	require.NoError(t, err)
	require.Equal(t, "30 minutes", *h.Target, "trimmed")

	// no target, and an all-whitespace target, both leave Target nil
	h2, err := svc.CreateHabit(ctx, acc, habits.HabitInput{Name: "No target"})
	require.NoError(t, err)
	require.Nil(t, h2.Target)

	h3, err := svc.CreateHabit(ctx, acc, habits.HabitInput{Name: "Blank target", Target: strp("   ")})
	require.NoError(t, err)
	require.Nil(t, h3.Target)

	_, err = svc.CreateHabit(ctx, acc, habits.HabitInput{Name: "Too long", Target: strp(strings.Repeat("x", 101))})
	var verr *habits.ValidationError
	require.ErrorAs(t, err, &verr)
	require.Contains(t, verr.Fields, "target")
}

func TestUpdateHabit(t *testing.T) {
	svc, pool, acc := setup(t)
	ctx := context.Background()

	h, err := svc.CreateHabit(ctx, acc, habits.HabitInput{Name: "Meditate", Target: strp("10 minutes")})
	require.NoError(t, err)

	updated, err := svc.UpdateHabit(ctx, acc, h.ID, "Meditate daily", strp("15 minutes"))
	require.NoError(t, err)
	require.Equal(t, "Meditate daily", updated.Name)
	require.Equal(t, "15 minutes", *updated.Target)

	// clearing the target
	updated, err = svc.UpdateHabit(ctx, acc, h.ID, "Meditate daily", nil)
	require.NoError(t, err)
	require.Nil(t, updated.Target)

	// bad name -> validation error, nothing changed
	_, err = svc.UpdateHabit(ctx, acc, h.ID, "   ", nil)
	var verr *habits.ValidationError
	require.ErrorAs(t, err, &verr)
	require.Contains(t, verr.Fields, "name")

	// unknown habit -> not found
	_, err = svc.UpdateHabit(ctx, acc, uuid.New(), "X", nil)
	require.ErrorIs(t, err, habits.ErrHabitNotFound)

	// isolation
	other := newAccount(t, pool, "other-update@test")
	_, err = svc.UpdateHabit(ctx, other, h.ID, "hijacked", nil)
	require.ErrorIs(t, err, habits.ErrHabitNotFound)

	// edit never touches completions/streak — mark today, edit, streak survives
	require.NoError(t, svc.MarkComplete(ctx, acc, h.ID, today()))
	views, err := svc.ListActive(ctx, acc, today())
	require.NoError(t, err)
	v, ok := findView(views, h.ID)
	require.True(t, ok)
	require.Equal(t, 1, v.CurrentStreak)
	_, err = svc.UpdateHabit(ctx, acc, h.ID, "Meditate renamed", strp("20 minutes"))
	require.NoError(t, err)
	views, err = svc.ListActive(ctx, acc, today())
	require.NoError(t, err)
	v, ok = findView(views, h.ID)
	require.True(t, ok)
	require.Equal(t, 1, v.CurrentStreak, "editing name/target does not touch the streak")
	require.Equal(t, "20 minutes", *v.Target)
}

func TestMarkAndUnmark(t *testing.T) {
	svc, pool, acc := setup(t)
	ctx := context.Background()
	h, _ := svc.CreateHabit(ctx, acc, habits.HabitInput{Name: "Meditate"})
	d := today()

	require.NoError(t, svc.MarkComplete(ctx, acc, h.ID, d))
	require.NoError(t, svc.MarkComplete(ctx, acc, h.ID, d)) // idempotent

	var n int
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT count(*) FROM habit_completions WHERE habit_id = $1`, h.ID).Scan(&n))
	require.Equal(t, 1, n)

	require.NoError(t, svc.UnmarkComplete(ctx, acc, h.ID, d))
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT count(*) FROM habit_completions WHERE habit_id = $1`, h.ID).Scan(&n))
	require.Equal(t, 0, n)
}

func TestStreakThroughService(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()
	h, _ := svc.CreateHabit(ctx, acc, habits.HabitInput{Name: "Walk"})
	d := today()

	for _, day := range []timezone.Date{d, d.AddDays(-1), d.AddDays(-2)} {
		require.NoError(t, svc.MarkComplete(ctx, acc, h.ID, day))
	}
	views, err := svc.ListActive(ctx, acc, d)
	require.NoError(t, err)
	v, _ := findView(views, h.ID)
	require.Equal(t, 3, v.CurrentStreak)
	require.True(t, v.CompletedOnDate)
	require.Equal(t, 3, v.Last30Days)

	// break the streak by unmarking yesterday
	require.NoError(t, svc.UnmarkComplete(ctx, acc, h.ID, d.AddDays(-1)))
	views, _ = svc.ListActive(ctx, acc, d)
	v, _ = findView(views, h.ID)
	require.Equal(t, 1, v.CurrentStreak, "only today remains contiguous")
}

func TestArchiveUnarchivePreservesHistory(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()
	h, _ := svc.CreateHabit(ctx, acc, habits.HabitInput{Name: "Journal"})
	d := today()
	require.NoError(t, svc.MarkComplete(ctx, acc, h.ID, d))
	require.NoError(t, svc.MarkComplete(ctx, acc, h.ID, d.AddDays(-1)))

	require.NoError(t, svc.ArchiveHabit(ctx, acc, h.ID))
	active, _ := svc.ListActive(ctx, acc, d)
	_, present := findView(active, h.ID)
	require.False(t, present, "archived habit is hidden from the active list")

	archived, _ := svc.ListArchived(ctx, acc)
	require.Len(t, archived, 1)

	require.NoError(t, svc.UnarchiveHabit(ctx, acc, h.ID))
	active, _ = svc.ListActive(ctx, acc, d)
	v, ok := findView(active, h.ID)
	require.True(t, ok)
	require.Equal(t, 2, v.CurrentStreak, "streak resumes from preserved completions (Q11)")

	require.ErrorIs(t, svc.ArchiveHabit(ctx, acc, uuid.New()), habits.ErrHabitNotFound)
}

func TestFutureCompletionDoesNotInflateStreak(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()
	h, _ := svc.CreateHabit(ctx, acc, habits.HabitInput{Name: "Stretch"})
	d := today()

	require.NoError(t, svc.MarkComplete(ctx, acc, h.ID, d.AddDays(1))) // future (Q9: allowed)
	require.NoError(t, svc.MarkComplete(ctx, acc, h.ID, d))

	views, _ := svc.ListActive(ctx, acc, d)
	v, _ := findView(views, h.ID)
	require.Equal(t, 1, v.CurrentStreak, "the future completion is stored but does not count")
}

func TestHabitCategory(t *testing.T) {
	svc, pool, acc := setup(t)
	ctx := context.Background()
	other := newAccount(t, pool, "other-cat@test")

	mine := mkCategory(t, pool, acc, "Wellness")
	theirs := mkCategory(t, pool, other, "Theirs")
	archived := mkCategory(t, pool, acc, "Old")
	archiveCategory(t, pool, archived)

	h, err := svc.CreateHabit(ctx, acc, habits.HabitInput{Name: "Meditate", CategoryID: &mine})
	require.NoError(t, err)
	require.Equal(t, mine, *h.CategoryID)

	for _, cat := range []uuid.UUID{theirs, archived, uuid.New()} {
		_, err := svc.CreateHabit(ctx, acc, habits.HabitInput{Name: "x", CategoryID: &cat})
		var verr *habits.ValidationError
		require.ErrorAs(t, err, &verr)
		require.Contains(t, verr.Fields, "category_id")
	}

	other2 := mkCategory(t, pool, acc, "Focus")
	require.NoError(t, svc.SetHabitCategory(ctx, acc, h.ID, &other2))
	views, _ := svc.ListActive(ctx, acc, today())
	v, _ := findView(views, h.ID)
	require.Equal(t, other2, *v.CategoryID)

	require.NoError(t, svc.SetHabitCategory(ctx, acc, h.ID, nil))
	views, _ = svc.ListActive(ctx, acc, today())
	v, _ = findView(views, h.ID)
	require.Nil(t, v.CategoryID)

	var verr *habits.ValidationError
	require.ErrorAs(t, svc.SetHabitCategory(ctx, acc, h.ID, &theirs), &verr)
	require.ErrorIs(t, svc.SetHabitCategory(ctx, other, h.ID, &mine), habits.ErrHabitNotFound)
}

func TestListAllAndAllCompletions(t *testing.T) {
	svc, pool, acc := setup(t)
	ctx := context.Background()
	other := newAccount(t, pool, "other-listall@test")

	active, _ := svc.CreateHabit(ctx, acc, habits.HabitInput{Name: "Meditate"})
	archived, _ := svc.CreateHabit(ctx, acc, habits.HabitInput{Name: "Journal"})
	require.NoError(t, svc.ArchiveHabit(ctx, acc, archived.ID))
	_, _ = svc.CreateHabit(ctx, other, habits.HabitInput{Name: "Not mine"})

	require.NoError(t, svc.MarkComplete(ctx, acc, active.ID, today()))
	require.NoError(t, svc.MarkComplete(ctx, acc, archived.ID, today().AddDays(-1)))

	all, err := svc.ListAll(ctx, acc)
	require.NoError(t, err)
	require.Len(t, all, 2, "active and archived, nothing foreign")

	completions, err := svc.AllCompletions(ctx, acc)
	require.NoError(t, err)
	require.Len(t, completions, 2, "every completion, even the archived habit's")

	byHabit := map[uuid.UUID]timezone.Date{}
	for _, c := range completions {
		byHabit[c.HabitID] = c.Date
	}
	require.Equal(t, today(), byHabit[active.ID])
	require.Equal(t, today().AddDays(-1), byHabit[archived.ID])
}

func findHistoryEntry(entries []habits.HabitHistoryEntry, id uuid.UUID) (habits.HabitHistoryEntry, bool) {
	for _, e := range entries {
		if e.HabitID == id {
			return e, true
		}
	}
	return habits.HabitHistoryEntry{}, false
}

func TestHistory(t *testing.T) {
	svc, pool, acc := setup(t)
	ctx := context.Background()
	other := newAccount(t, pool, "other-history@test")

	active, _ := svc.CreateHabit(ctx, acc, habits.HabitInput{Name: "Meditate"})
	archived, _ := svc.CreateHabit(ctx, acc, habits.HabitInput{Name: "Journal"})
	require.NoError(t, svc.ArchiveHabit(ctx, acc, archived.ID))
	otherHabit, _ := svc.CreateHabit(ctx, other, habits.HabitInput{Name: "Not mine"})

	d := today()
	require.NoError(t, svc.MarkComplete(ctx, acc, active.ID, d))
	require.NoError(t, svc.MarkComplete(ctx, acc, active.ID, d.AddDays(-2)))
	require.NoError(t, svc.MarkComplete(ctx, acc, active.ID, d.AddDays(-10))) // outside the queried range
	require.NoError(t, svc.MarkComplete(ctx, acc, archived.ID, d.AddDays(-1)))
	require.NoError(t, svc.MarkComplete(ctx, other, otherHabit.ID, d))

	from, to := d.AddDays(-3), d
	entries, err := svc.History(ctx, acc, from, to)
	require.NoError(t, err)
	require.Len(t, entries, 2, "the caller's habits only, active and archived, nothing foreign")

	activeEntry, ok := findHistoryEntry(entries, active.ID)
	require.True(t, ok)
	require.False(t, activeEntry.Archived)
	require.ElementsMatch(t, []timezone.Date{d, d.AddDays(-2)}, activeEntry.Completions,
		"the completion 10 days ago is outside the range")

	archivedEntry, ok := findHistoryEntry(entries, archived.ID)
	require.True(t, ok)
	require.True(t, archivedEntry.Archived)
	require.Equal(t, []timezone.Date{d.AddDays(-1)}, archivedEntry.Completions)
}

func TestHistory_ZeroCompletionHabitStillListed(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()
	h, _ := svc.CreateHabit(ctx, acc, habits.HabitInput{Name: "New habit"})

	entries, err := svc.History(ctx, acc, today().AddDays(-5), today())
	require.NoError(t, err)
	entry, ok := findHistoryEntry(entries, h.ID)
	require.True(t, ok)
	require.Empty(t, entry.Completions)
}

// TestWeek covers the "This Week" grid batching (docs/left.md Phase 6): the ISO
// week's 7 dates, each active habit's streak and week completions, archived habits
// by name only, and isolation.
func TestWeek(t *testing.T) {
	svc, pool, acc := setup(t)
	ctx := context.Background()
	other := newAccount(t, pool, "other-week@test")

	h, err := svc.CreateHabit(ctx, acc, habits.HabitInput{Name: "Workout"})
	require.NoError(t, err)
	archived, err := svc.CreateHabit(ctx, acc, habits.HabitInput{Name: "Old habit"})
	require.NoError(t, err)
	require.NoError(t, svc.ArchiveHabit(ctx, acc, archived.ID))
	otherHabit, err := svc.CreateHabit(ctx, other, habits.HabitInput{Name: "Not mine"})
	require.NoError(t, err)

	d := today()
	require.NoError(t, svc.MarkComplete(ctx, acc, h.ID, d))
	require.NoError(t, svc.MarkComplete(ctx, other, otherHabit.ID, d))

	wv, err := svc.Week(ctx, acc, d)
	require.NoError(t, err)
	require.Len(t, wv.Days, 7)
	require.False(t, d.Before(wv.WeekStart), "requested date is within the returned week")
	require.True(t, d.Before(wv.WeekStart.AddDays(7)), "requested date is within the returned week")

	require.Len(t, wv.Habits, 1, "only the caller's active habit — none of other's")
	entry := wv.Habits[0]
	require.Equal(t, h.ID, entry.HabitID)
	require.Equal(t, "Workout", entry.Name)
	require.Positive(t, entry.CurrentStreak)
	require.Contains(t, entry.Completed, d)

	require.Len(t, wv.Archived, 1)
	require.Equal(t, archived.ID, wv.Archived[0].HabitID)
	require.Equal(t, "Old habit", wv.Archived[0].Name)
}

// TestWeek_StreakLookbackExceedsWeekWindow proves the streak still looks back the
// full window (like ListActive) even though Completed is scoped to the requested
// week's 7 dates only.
func TestWeek_StreakLookbackExceedsWeekWindow(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()
	h, err := svc.CreateHabit(ctx, acc, habits.HabitInput{Name: "Workout"})
	require.NoError(t, err)

	d := today()
	for i := 0; i < 10; i++ {
		require.NoError(t, svc.MarkComplete(ctx, acc, h.ID, d.AddDays(-i)))
	}

	wv, err := svc.Week(ctx, acc, d)
	require.NoError(t, err)
	require.Len(t, wv.Habits, 1)
	require.Equal(t, 10, wv.Habits[0].CurrentStreak, "streak looks back further than the displayed week")
	require.LessOrEqual(t, len(wv.Habits[0].Completed), 7, "Completed is scoped to the week's 7 dates only")
}

func TestHistory_ToBeforeFrom(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()

	_, err := svc.History(ctx, acc, today(), today().AddDays(-1))
	var verr *habits.ValidationError
	require.ErrorAs(t, err, &verr)
	require.Contains(t, verr.Fields, "to")
}

func TestHistory_RangeTooLarge(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()

	_, err := svc.History(ctx, acc, today(), today().AddDays(93)) // 94 days
	var verr *habits.ValidationError
	require.ErrorAs(t, err, &verr)
	require.Contains(t, verr.Fields, "range")

	_, err = svc.History(ctx, acc, today(), today().AddDays(91)) // exactly 92 days
	require.NoError(t, err)
}

func TestCountByCategory(t *testing.T) {
	svc, pool, acc := setup(t)
	ctx := context.Background()
	other := newAccount(t, pool, "other-count@test")

	catA := mkCategory(t, pool, acc, "A")
	catB := mkCategory(t, pool, acc, "B")
	otherCat := mkCategory(t, pool, other, "Other")

	_, _ = svc.CreateHabit(ctx, acc, habits.HabitInput{Name: "1", CategoryID: &catA})
	h2, _ := svc.CreateHabit(ctx, acc, habits.HabitInput{Name: "2", CategoryID: &catA})
	_, _ = svc.CreateHabit(ctx, acc, habits.HabitInput{Name: "3", CategoryID: &catB})
	_, _ = svc.CreateHabit(ctx, acc, habits.HabitInput{Name: "4"}) // uncategorized
	_, _ = svc.CreateHabit(ctx, other, habits.HabitInput{Name: "5", CategoryID: &otherCat})

	counts, err := svc.CountByCategory(ctx, acc)
	require.NoError(t, err)
	require.Equal(t, map[uuid.UUID]int{catA: 2, catB: 1}, counts)

	// archiving one of catA's habits drops the count — archived habits don't count
	require.NoError(t, svc.ArchiveHabit(ctx, acc, h2.ID))
	counts, err = svc.CountByCategory(ctx, acc)
	require.NoError(t, err)
	require.Equal(t, map[uuid.UUID]int{catA: 1, catB: 1}, counts)
}

func TestCompletionCountsInRange(t *testing.T) {
	svc, pool, acc := setup(t)
	ctx := context.Background()
	other := newAccount(t, pool, "other-range@test")

	h1, _ := svc.CreateHabit(ctx, acc, habits.HabitInput{Name: "Meditate"})
	h2, _ := svc.CreateHabit(ctx, acc, habits.HabitInput{Name: "Walk"})
	d := today()

	require.NoError(t, svc.MarkComplete(ctx, acc, h1.ID, d))
	require.NoError(t, svc.MarkComplete(ctx, acc, h1.ID, d.AddDays(-1)))
	require.NoError(t, svc.MarkComplete(ctx, acc, h1.ID, d.AddDays(-10))) // outside the range
	require.NoError(t, svc.MarkComplete(ctx, acc, h2.ID, d))
	require.NoError(t, svc.ArchiveHabit(ctx, acc, h2.ID))

	otherHabit, _ := svc.CreateHabit(ctx, other, habits.HabitInput{Name: "Other"})
	require.NoError(t, svc.MarkComplete(ctx, other, otherHabit.ID, d))

	from, to := d.AddDays(-2), d
	counts, err := svc.CompletionCountsInRange(ctx, acc, from, to)
	require.NoError(t, err)
	require.Len(t, counts, 2, "active and archived habits both appear")

	byName := map[string]int{}
	for _, c := range counts {
		byName[c.Name] = c.Count
	}
	require.Equal(t, 2, byName["Meditate"])
	require.Equal(t, 1, byName["Walk"], "archived habit still counted")

	// a habit with zero completions in range still appears, with count 0
	h3, _ := svc.CreateHabit(ctx, acc, habits.HabitInput{Name: "Journal"})
	counts, err = svc.CompletionCountsInRange(ctx, acc, from, to)
	require.NoError(t, err)
	found := false
	for _, c := range counts {
		if c.HabitID == h3.ID {
			found = true
			require.Equal(t, 0, c.Count)
		}
	}
	require.True(t, found, "zero-completion habit still listed")

	// to before from -> validation error
	_, err = svc.CompletionCountsInRange(ctx, acc, to, from)
	var verr *habits.ValidationError
	require.ErrorAs(t, err, &verr)
	require.Contains(t, verr.Fields, "to")
}

func TestHabitIsolation(t *testing.T) {
	svc, pool, a := setup(t)
	ctx := context.Background()
	b := newAccount(t, pool, "other@test")

	ha, _ := svc.CreateHabit(ctx, a, habits.HabitInput{Name: "A's habit"})
	_, _ = svc.CreateHabit(ctx, b, habits.HabitInput{Name: "B's habit"})
	d := today()

	bViews, _ := svc.ListActive(ctx, b, d)
	require.Len(t, bViews, 1)
	require.Equal(t, "B's habit", bViews[0].Name)

	require.ErrorIs(t, svc.MarkComplete(ctx, b, ha.ID, d), habits.ErrHabitNotFound)
	require.ErrorIs(t, svc.ArchiveHabit(ctx, b, ha.ID), habits.ErrHabitNotFound)
	require.ErrorIs(t, svc.UnmarkComplete(ctx, b, ha.ID, d), habits.ErrHabitNotFound)
}
