package export_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/satya-18-w/productivity-os/internal/categories"
	"github.com/satya-18-w/productivity-os/internal/export"
	"github.com/satya-18-w/productivity-os/internal/goals"
	"github.com/satya-18-w/productivity-os/internal/habits"
	"github.com/satya-18-w/productivity-os/internal/notes"
	"github.com/satya-18-w/productivity-os/internal/platform/pgtest"
	"github.com/satya-18-w/productivity-os/internal/platform/timezone"
	"github.com/satya-18-w/productivity-os/internal/reviews"
	"github.com/satya-18-w/productivity-os/internal/tasks"
	"github.com/satya-18-w/productivity-os/internal/timeline"
)

type fakeZone struct{}

func (fakeZone) Zone(context.Context, uuid.UUID) (*time.Location, error) { return time.UTC, nil }

// env wires export over the real domain modules against one shared pool —
// composition exactly as cmd/server does it.
type env struct {
	export     export.Service
	categories categories.Service
	timeline   timeline.Service
	tasks      tasks.Service
	habits     habits.Service
	goals      goals.Service
	reviews    reviews.Service
	notes      notes.Service
	pool       *pgxpool.Pool
	acc        uuid.UUID
}

func setup(t *testing.T) env {
	t.Helper()
	pool := pgtest.Pool(t)
	acc := newAccount(t, pool, "owner@test")

	catSvc := categories.NewService(pool)
	goalsSvc := goals.NewService(pool, catSvc)
	tasksSvc := tasks.NewService(pool, catSvc, goalsSvc)
	timelineSvc := timeline.NewService(pool, fakeZone{}, catSvc, tasksSvc)
	habitsSvc := habits.NewService(pool, fakeZone{}, catSvc)
	reviewsSvc := reviews.NewService(pool)
	notesSvc := notes.NewService(pool)

	exportSvc := export.NewService(catSvc, timelineSvc, tasksSvc, habitsSvc, goalsSvc, reviewsSvc, notesSvc)

	return env{
		export: exportSvc, categories: catSvc, timeline: timelineSvc, tasks: tasksSvc,
		habits: habitsSvc, goals: goalsSvc, reviews: reviewsSvc, notes: notesSvc, pool: pool, acc: acc,
	}
}

func newAccount(t *testing.T, pool *pgxpool.Pool, email string) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	require.NoError(t, pool.QueryRow(context.Background(),
		`INSERT INTO accounts (email, password_hash, timezone) VALUES ($1, 'x', 'UTC') RETURNING id`,
		email).Scan(&id))
	return id
}

func ts(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

// seed populates every entity type at least once for acc, including an archived
// category, an archived habit, and both review kinds — round-trip completeness
// fixture.
func seed(t *testing.T, e env, acc uuid.UUID) {
	t.Helper()
	ctx := context.Background()

	active, err := e.categories.Create(ctx, acc, categories.Input{Name: "Deep Work", Colour: "blue"})
	require.NoError(t, err)
	archivedCat, err := e.categories.Create(ctx, acc, categories.Input{Name: "Old"})
	require.NoError(t, err)
	require.NoError(t, e.categories.Archive(ctx, acc, archivedCat.ID))

	_, err = e.timeline.AddBlock(ctx, acc, timeline.BlockInput{
		Kind: timeline.Planned, StartsAt: ts("2025-06-15T09:00:00Z"), EndsAt: ts("2025-06-15T10:00:00Z"), CategoryID: &active.ID,
	})
	require.NoError(t, err)
	_, err = e.timeline.AddBlock(ctx, acc, timeline.BlockInput{
		Kind: timeline.Actual, StartsAt: ts("2025-06-15T09:00:00Z"), EndsAt: ts("2025-06-15T11:00:00Z"),
	})
	require.NoError(t, err)

	_, err = e.tasks.CreateTask(ctx, acc, tasks.TaskInput{Title: "Ship export"})
	require.NoError(t, err)

	activeHabit, err := e.habits.CreateHabit(ctx, acc, habits.HabitInput{Name: "Meditate"})
	require.NoError(t, err)
	archivedHabit, err := e.habits.CreateHabit(ctx, acc, habits.HabitInput{Name: "Old habit"})
	require.NoError(t, err)
	require.NoError(t, e.habits.ArchiveHabit(ctx, acc, archivedHabit.ID))
	require.NoError(t, e.habits.MarkComplete(ctx, acc, activeHabit.ID, timezone.Date{Year: 2025, Month: 6, Day: 15}))
	require.NoError(t, e.habits.MarkComplete(ctx, acc, archivedHabit.ID, timezone.Date{Year: 2025, Month: 6, Day: 10}))

	_, err = e.goals.CreateGoal(ctx, acc, goals.GoalInput{Title: "Ship M8"})
	require.NoError(t, err)

	_, err = e.reviews.SaveDaily(ctx, acc, timezone.Date{Year: 2025, Month: 6, Day: 15}, reviews.Answers{"went_well": "shipped it"})
	require.NoError(t, err)
	_, err = e.reviews.SaveWeekly(ctx, acc, 2025, 24, reviews.Answers{"highlights": "M8"})
	require.NoError(t, err)

	_, err = e.notes.CreateNote(ctx, acc, notes.NoteInput{Title: "Export idea", Body: "ship notes too"})
	require.NoError(t, err)
}

func TestExport_RoundTripCompleteness(t *testing.T) {
	e := setup(t)
	seed(t, e, e.acc)

	out, err := e.export.Export(context.Background(), e.acc)
	require.NoError(t, err)

	require.WithinDuration(t, time.Now(), out.ExportedAt, 10*time.Second)

	require.Len(t, out.Categories, 2, "active + archived category")
	require.Len(t, out.PlannedBlocks, 1)
	require.Len(t, out.ActualBlocks, 1)
	require.Len(t, out.Tasks, 1)
	require.Len(t, out.Habits, 2, "active + archived habit")
	require.Len(t, out.HabitCompletions, 2, "one completion per habit, including the archived one")
	require.Len(t, out.Goals, 1)
	require.Len(t, out.DailyReviews, 1)
	require.Len(t, out.WeeklyReviews, 1)
	require.Len(t, out.Notes, 1)

	require.Equal(t, "Ship export", out.Tasks[0].Title)
	require.Equal(t, "Ship M8", out.Goals[0].Title)
	require.Equal(t, "shipped it", out.DailyReviews[0].Answers["went_well"])
	require.Equal(t, "M8", out.WeeklyReviews[0].Answers["highlights"])
	require.Equal(t, "Export idea", out.Notes[0].Title)
	require.Equal(t, "ship notes too", out.Notes[0].Body)
	require.Zero(t, out.GoalDoneTasks[out.Goals[0].ID], "no linked tasks -> 0")
	require.Zero(t, out.GoalTotalTasks[out.Goals[0].ID])
}

// TestExport_GoalProgress closes the completeness gap flagged during
// MX3-follow-up (docs/export-format.md): GET /api/goals returns derived
// done_tasks/total_tasks live, and the export must match.
func TestExport_GoalProgress(t *testing.T) {
	e := setup(t)
	ctx := context.Background()

	goal, err := e.goals.CreateGoal(ctx, e.acc, goals.GoalInput{Title: "Ship MX-follow-up"})
	require.NoError(t, err)
	t1, err := e.tasks.CreateTask(ctx, e.acc, tasks.TaskInput{Title: "t1", GoalID: &goal.ID})
	require.NoError(t, err)
	_, err = e.tasks.CreateTask(ctx, e.acc, tasks.TaskInput{Title: "t2", GoalID: &goal.ID})
	require.NoError(t, err)
	require.NoError(t, e.tasks.MoveTask(ctx, e.acc, t1.ID, tasks.Done))

	out, err := e.export.Export(ctx, e.acc)
	require.NoError(t, err)
	require.Equal(t, 1, out.GoalDoneTasks[goal.ID])
	require.Equal(t, 2, out.GoalTotalTasks[goal.ID])
}

func TestExport_Isolation(t *testing.T) {
	e := setup(t)
	seed(t, e, e.acc)

	other := newAccount(t, e.pool, "other@test")
	_, err := e.categories.Create(context.Background(), other, categories.Input{Name: "Not mine"})
	require.NoError(t, err)

	out, err := e.export.Export(context.Background(), other)
	require.NoError(t, err)

	require.Len(t, out.Categories, 1, "only other's own category")
	require.Equal(t, "Not mine", out.Categories[0].Name)
	require.Empty(t, out.PlannedBlocks)
	require.Empty(t, out.ActualBlocks)
	require.Empty(t, out.Tasks)
	require.Empty(t, out.Habits)
	require.Empty(t, out.HabitCompletions)
	require.Empty(t, out.Goals)
	require.Empty(t, out.DailyReviews)
	require.Empty(t, out.WeeklyReviews)
	require.Empty(t, out.Notes)
}
