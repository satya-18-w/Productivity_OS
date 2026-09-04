package habits_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/satya-18-w/productivity-os/internal/habits"
	"github.com/satya-18-w/productivity-os/internal/platform/pgtest"
	"github.com/satya-18-w/productivity-os/internal/platform/timezone"
)

type fakeZone struct{}

func (fakeZone) Zone(context.Context, uuid.UUID) (*time.Location, error) { return time.UTC, nil }

func setup(t *testing.T) (habits.Service, *pgxpool.Pool, uuid.UUID) {
	t.Helper()
	pool := pgtest.Pool(t)
	return habits.NewService(pool, fakeZone{}), pool, newAccount(t, pool, "owner@test")
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

	h, err := svc.CreateHabit(ctx, acc, "  Read 20 pages  ")
	require.NoError(t, err)
	require.Equal(t, "Read 20 pages", h.Name)
	require.Nil(t, h.ArchivedAt)

	_, err = svc.CreateHabit(ctx, acc, "   ")
	var verr *habits.ValidationError
	require.ErrorAs(t, err, &verr)
	require.Contains(t, verr.Fields, "name")
}

func TestMarkAndUnmark(t *testing.T) {
	svc, pool, acc := setup(t)
	ctx := context.Background()
	h, _ := svc.CreateHabit(ctx, acc, "Meditate")
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
	h, _ := svc.CreateHabit(ctx, acc, "Walk")
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
	h, _ := svc.CreateHabit(ctx, acc, "Journal")
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
	h, _ := svc.CreateHabit(ctx, acc, "Stretch")
	d := today()

	require.NoError(t, svc.MarkComplete(ctx, acc, h.ID, d.AddDays(1))) // future (Q9: allowed)
	require.NoError(t, svc.MarkComplete(ctx, acc, h.ID, d))

	views, _ := svc.ListActive(ctx, acc, d)
	v, _ := findView(views, h.ID)
	require.Equal(t, 1, v.CurrentStreak, "the future completion is stored but does not count")
}

func TestHabitIsolation(t *testing.T) {
	svc, pool, a := setup(t)
	ctx := context.Background()
	b := newAccount(t, pool, "other@test")

	ha, _ := svc.CreateHabit(ctx, a, "A's habit")
	_, _ = svc.CreateHabit(ctx, b, "B's habit")
	d := today()

	bViews, _ := svc.ListActive(ctx, b, d)
	require.Len(t, bViews, 1)
	require.Equal(t, "B's habit", bViews[0].Name)

	require.ErrorIs(t, svc.MarkComplete(ctx, b, ha.ID, d), habits.ErrHabitNotFound)
	require.ErrorIs(t, svc.ArchiveHabit(ctx, b, ha.ID), habits.ErrHabitNotFound)
	require.ErrorIs(t, svc.UnmarkComplete(ctx, b, ha.ID, d), habits.ErrHabitNotFound)
}
