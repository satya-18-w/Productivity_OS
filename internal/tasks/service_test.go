package tasks_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/satya-18-w/productivity-os/internal/categories"
	"github.com/satya-18-w/productivity-os/internal/goals"
	"github.com/satya-18-w/productivity-os/internal/platform/pgtest"
	"github.com/satya-18-w/productivity-os/internal/platform/timezone"
	"github.com/satya-18-w/productivity-os/internal/tasks"
)

func setup(t *testing.T) (tasks.Service, *pgxpool.Pool, uuid.UUID) {
	t.Helper()
	pool := pgtest.Pool(t)
	catSvc := categories.NewService(pool)
	return tasks.NewService(pool, catSvc, goals.NewService(pool, catSvc)), pool, newAccount(t, pool, "owner@test")
}

// mkGoal inserts a goal directly (tasks tests only need one to exist and be
// assignable — goal CRUD is exercised in the goals package).
func mkGoal(t *testing.T, pool *pgxpool.Pool, acc uuid.UUID, title string) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	require.NoError(t, pool.QueryRow(context.Background(),
		`INSERT INTO goals (account_id, title) VALUES ($1, $2) RETURNING id`,
		acc, title).Scan(&id))
	return id
}

// mkCategory inserts an active category directly (tasks tests only need one to
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

func transitionCount(t *testing.T, pool *pgxpool.Pool, taskID uuid.UUID) int {
	t.Helper()
	var n int
	require.NoError(t, pool.QueryRow(context.Background(),
		`SELECT count(*) FROM task_transitions WHERE task_id = $1`, taskID).Scan(&n))
	return n
}

func requireField(t *testing.T, err error, field string) {
	t.Helper()
	var verr *tasks.ValidationError
	require.ErrorAs(t, err, &verr)
	require.Contains(t, verr.Fields, field)
}

func TestCreateTask(t *testing.T) {
	svc, pool, acc := setup(t)
	ctx := context.Background()

	due := timezone.Date{Year: 2025, Month: 7, Day: 1}
	task, err := svc.CreateTask(ctx, acc, tasks.TaskInput{Title: "  Write spec  ", Description: "details", DueDate: &due})
	require.NoError(t, err)
	require.Equal(t, "Write spec", task.Title)
	require.Equal(t, tasks.Backlog, task.State)
	require.Equal(t, "2025-07-01", task.DueDate.String())
	require.Equal(t, 1, transitionCount(t, pool, task.ID), "creation transition recorded")
}

func TestCreateTask_Validation(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()

	_, err := svc.CreateTask(ctx, acc, tasks.TaskInput{Title: "   "})
	requireField(t, err, "title")

	_, err = svc.CreateTask(ctx, acc, tasks.TaskInput{Title: strings.Repeat("x", 201)})
	requireField(t, err, "title")

	_, err = svc.CreateTask(ctx, acc, tasks.TaskInput{Title: "ok", Description: strings.Repeat("y", 5001)})
	requireField(t, err, "description")
}

func TestUpdateTask(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()
	task, _ := svc.CreateTask(ctx, acc, tasks.TaskInput{Title: "Draft"})

	require.NoError(t, svc.UpdateTask(ctx, acc, task.ID, tasks.TaskInput{Title: "Final", Description: "done thinking"}))

	board, _ := svc.Board(ctx, acc)
	require.Equal(t, "Final", board.Columns[0].Tasks[0].Title)

	require.ErrorIs(t, svc.UpdateTask(ctx, acc, uuid.New(), tasks.TaskInput{Title: "x"}), tasks.ErrTaskNotFound)
}

func TestMoveTask(t *testing.T) {
	svc, pool, acc := setup(t)
	ctx := context.Background()
	task, _ := svc.CreateTask(ctx, acc, tasks.TaskInput{Title: "Ship it"})

	require.NoError(t, svc.MoveTask(ctx, acc, task.ID, tasks.Todo))
	require.NoError(t, svc.MoveTask(ctx, acc, task.ID, tasks.InProgress))
	require.NoError(t, svc.MoveTask(ctx, acc, task.ID, tasks.Done))

	// same-state move records nothing
	before := transitionCount(t, pool, task.ID)
	require.NoError(t, svc.MoveTask(ctx, acc, task.ID, tasks.Done))
	require.Equal(t, before, transitionCount(t, pool, task.ID))

	// re-open then done again
	require.NoError(t, svc.MoveTask(ctx, acc, task.ID, tasks.InProgress))
	require.NoError(t, svc.MoveTask(ctx, acc, task.ID, tasks.Done))

	// creation + 5 real moves = 6 transitions
	require.Equal(t, 6, transitionCount(t, pool, task.ID))

	var doneEntries int
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT count(*) FROM task_transitions WHERE task_id = $1 AND to_state = 'DONE'`, task.ID).Scan(&doneEntries))
	require.Equal(t, 2, doneEntries)

	// invalid target
	requireField(t, svc.MoveTask(ctx, acc, task.ID, tasks.State("SIDEWAYS")), "state")
	// missing task
	require.ErrorIs(t, svc.MoveTask(ctx, acc, uuid.New(), tasks.Todo), tasks.ErrTaskNotFound)
}

func TestDeleteTask(t *testing.T) {
	svc, pool, acc := setup(t)
	ctx := context.Background()
	task, _ := svc.CreateTask(ctx, acc, tasks.TaskInput{Title: "Temp"})
	_ = svc.MoveTask(ctx, acc, task.ID, tasks.Todo)

	require.NoError(t, svc.DeleteTask(ctx, acc, task.ID))
	require.Equal(t, 0, transitionCount(t, pool, task.ID), "transitions cascade-deleted")
	require.ErrorIs(t, svc.DeleteTask(ctx, acc, task.ID), tasks.ErrTaskNotFound)
}

func TestBoard(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()

	a, _ := svc.CreateTask(ctx, acc, tasks.TaskInput{Title: "first"})
	b, _ := svc.CreateTask(ctx, acc, tasks.TaskInput{Title: "second"})
	require.NoError(t, svc.MoveTask(ctx, acc, b.ID, tasks.Done))

	board, err := svc.Board(ctx, acc)
	require.NoError(t, err)
	require.Len(t, board.Columns, 4)
	require.Equal(t, []tasks.State{tasks.Backlog, tasks.Todo, tasks.InProgress, tasks.Done},
		[]tasks.State{board.Columns[0].State, board.Columns[1].State, board.Columns[2].State, board.Columns[3].State})

	require.Len(t, board.Columns[0].Tasks, 1)
	require.Equal(t, a.ID, board.Columns[0].Tasks[0].ID)
	require.Empty(t, board.Columns[1].Tasks)
	require.Len(t, board.Columns[3].Tasks, 1)
	require.Equal(t, b.ID, board.Columns[3].Tasks[0].ID)
}

func TestTaskCategory(t *testing.T) {
	svc, pool, acc := setup(t)
	ctx := context.Background()
	other := newAccount(t, pool, "other-cat@test")

	mine := mkCategory(t, pool, acc, "Deep Work")
	theirs := mkCategory(t, pool, other, "Theirs")
	archived := mkCategory(t, pool, acc, "Old")
	archiveCategory(t, pool, archived)

	task, err := svc.CreateTask(ctx, acc, tasks.TaskInput{Title: "Ship it", CategoryID: &mine})
	require.NoError(t, err)
	require.Equal(t, mine, *task.CategoryID)

	for _, cat := range []uuid.UUID{theirs, archived, uuid.New()} {
		_, err := svc.CreateTask(ctx, acc, tasks.TaskInput{Title: "x", CategoryID: &cat})
		requireField(t, err, "category_id")
	}

	// change on update
	other2 := mkCategory(t, pool, acc, "Focus")
	require.NoError(t, svc.UpdateTask(ctx, acc, task.ID, tasks.TaskInput{Title: "Ship it", CategoryID: &other2}))
	board, _ := svc.Board(ctx, acc)
	require.Equal(t, other2, *board.Columns[0].Tasks[0].CategoryID)

	// clear on update
	require.NoError(t, svc.UpdateTask(ctx, acc, task.ID, tasks.TaskInput{Title: "Ship it"}))
	board, _ = svc.Board(ctx, acc)
	require.Nil(t, board.Columns[0].Tasks[0].CategoryID)

	// foreign category on update -> 400
	requireField(t, svc.UpdateTask(ctx, acc, task.ID, tasks.TaskInput{Title: "x", CategoryID: &theirs}), "category_id")
}

func TestTaskGoalLink(t *testing.T) {
	svc, pool, acc := setup(t)
	ctx := context.Background()
	other := newAccount(t, pool, "other-goal@test")

	mine := mkGoal(t, pool, acc, "Ship MX3")
	theirs := mkGoal(t, pool, other, "Theirs")

	task, err := svc.CreateTask(ctx, acc, tasks.TaskInput{Title: "Ship it", GoalID: &mine})
	require.NoError(t, err)
	require.Equal(t, mine, *task.GoalID)

	for _, g := range []uuid.UUID{theirs, uuid.New()} {
		_, err := svc.CreateTask(ctx, acc, tasks.TaskInput{Title: "x", GoalID: &g})
		requireField(t, err, "goal_id")
	}

	// change on update
	other2 := mkGoal(t, pool, acc, "Other goal")
	require.NoError(t, svc.UpdateTask(ctx, acc, task.ID, tasks.TaskInput{Title: "Ship it", GoalID: &other2}))
	board, _ := svc.Board(ctx, acc)
	require.Equal(t, other2, *board.Columns[0].Tasks[0].GoalID)

	// clear on update
	require.NoError(t, svc.UpdateTask(ctx, acc, task.ID, tasks.TaskInput{Title: "Ship it"}))
	board, _ = svc.Board(ctx, acc)
	require.Nil(t, board.Columns[0].Tasks[0].GoalID)

	// foreign goal on update -> 400
	requireField(t, svc.UpdateTask(ctx, acc, task.ID, tasks.TaskInput{Title: "x", GoalID: &theirs}), "goal_id")
}

// TestTaskGoalDeleteClearsLink proves deleting a goal clears goal_id on its linked
// tasks (ON DELETE SET NULL) without deleting the tasks — MX3's delete-with-links
// default (v1.md §10).
func TestTaskGoalDeleteClearsLink(t *testing.T) {
	svc, pool, acc := setup(t)
	ctx := context.Background()

	g := mkGoal(t, pool, acc, "Doomed goal")
	task, err := svc.CreateTask(ctx, acc, tasks.TaskInput{Title: "Survives", GoalID: &g})
	require.NoError(t, err)

	_, delErr := pool.Exec(ctx, `DELETE FROM goals WHERE id = $1`, g)
	require.NoError(t, delErr)

	board, err := svc.Board(ctx, acc)
	require.NoError(t, err)
	require.Len(t, board.Columns[0].Tasks, 1)
	require.Equal(t, task.ID, board.Columns[0].Tasks[0].ID)
	require.Nil(t, board.Columns[0].Tasks[0].GoalID)
}

func TestTaskPriority(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()

	high := tasks.High
	task, err := svc.CreateTask(ctx, acc, tasks.TaskInput{Title: "Ship it", Priority: &high})
	require.NoError(t, err)
	require.Equal(t, tasks.High, *task.Priority)

	for _, p := range []tasks.Priority{tasks.Medium, tasks.Low} {
		p := p
		require.NoError(t, svc.UpdateTask(ctx, acc, task.ID, tasks.TaskInput{Title: "Ship it", Priority: &p}))
		board, _ := svc.Board(ctx, acc)
		require.Equal(t, p, *board.Columns[0].Tasks[0].Priority)
	}

	// clear on update
	require.NoError(t, svc.UpdateTask(ctx, acc, task.ID, tasks.TaskInput{Title: "Ship it"}))
	board, _ := svc.Board(ctx, acc)
	require.Nil(t, board.Columns[0].Tasks[0].Priority)

	// invalid value -> 400, on both create and update
	bad := tasks.Priority("URGENT")
	_, err = svc.CreateTask(ctx, acc, tasks.TaskInput{Title: "x", Priority: &bad})
	requireField(t, err, "priority")
	requireField(t, svc.UpdateTask(ctx, acc, task.ID, tasks.TaskInput{Title: "x", Priority: &bad}), "priority")
}

// TestAssignableToAccount covers the MX-TL cross-module check timeline.TaskChecker
// relies on (mirrors categories/goals' identical method).
func TestAssignableToAccount(t *testing.T) {
	svc, pool, acc := setup(t)
	ctx := context.Background()
	other := newAccount(t, pool, "other-assignable@test")

	mine, err := svc.CreateTask(ctx, acc, tasks.TaskInput{Title: "Mine"})
	require.NoError(t, err)
	theirs, err := svc.CreateTask(ctx, other, tasks.TaskInput{Title: "Theirs"})
	require.NoError(t, err)

	ok, err := svc.AssignableToAccount(ctx, acc, mine.ID)
	require.NoError(t, err)
	require.True(t, ok)

	ok, err = svc.AssignableToAccount(ctx, acc, theirs.ID)
	require.NoError(t, err)
	require.False(t, ok, "not the caller's account")

	ok, err = svc.AssignableToAccount(ctx, acc, uuid.New())
	require.NoError(t, err)
	require.False(t, ok, "unknown task")
}

// TestCategoriesForTasks covers the MX-TL bulk lookup timeline uses to resolve a
// task-linked block's inherited category.
func TestCategoriesForTasks(t *testing.T) {
	svc, pool, acc := setup(t)
	ctx := context.Background()
	other := newAccount(t, pool, "other-catfortasks@test")

	cat := mkCategory(t, pool, acc, "Deep Work")
	withCat, err := svc.CreateTask(ctx, acc, tasks.TaskInput{Title: "Categorized", CategoryID: &cat})
	require.NoError(t, err)
	noCat, err := svc.CreateTask(ctx, acc, tasks.TaskInput{Title: "Uncategorized"})
	require.NoError(t, err)
	_, err = svc.CreateTask(ctx, other, tasks.TaskInput{Title: "Other's task"})
	require.NoError(t, err)

	out, err := svc.CategoriesForTasks(ctx, acc)
	require.NoError(t, err)
	require.Len(t, out, 2, "only the caller's tasks")
	require.Equal(t, cat, *out[withCat.ID])
	require.Nil(t, out[noCat.ID])
}

func TestCountByCategory(t *testing.T) {
	svc, pool, acc := setup(t)
	ctx := context.Background()
	other := newAccount(t, pool, "other-count@test")

	catA := mkCategory(t, pool, acc, "A")
	catB := mkCategory(t, pool, acc, "B")
	otherCat := mkCategory(t, pool, other, "Other")

	_, _ = svc.CreateTask(ctx, acc, tasks.TaskInput{Title: "1", CategoryID: &catA})
	_, _ = svc.CreateTask(ctx, acc, tasks.TaskInput{Title: "2", CategoryID: &catA})
	_, _ = svc.CreateTask(ctx, acc, tasks.TaskInput{Title: "3", CategoryID: &catB})
	_, _ = svc.CreateTask(ctx, acc, tasks.TaskInput{Title: "4"}) // uncategorized
	_, _ = svc.CreateTask(ctx, other, tasks.TaskInput{Title: "5", CategoryID: &otherCat})

	counts, err := svc.CountByCategory(ctx, acc)
	require.NoError(t, err)
	require.Equal(t, map[uuid.UUID]int{catA: 2, catB: 1}, counts, "only the caller's tasks, no zero/uncategorized entry")
}

func TestDoneCountInRange(t *testing.T) {
	svc, pool, acc := setup(t)
	ctx := context.Background()
	other := newAccount(t, pool, "other-throughput@test")

	a, _ := svc.CreateTask(ctx, acc, tasks.TaskInput{Title: "a"})
	b, _ := svc.CreateTask(ctx, acc, tasks.TaskInput{Title: "b"})
	_, _ = svc.CreateTask(ctx, acc, tasks.TaskInput{Title: "c"}) // never DONE
	require.NoError(t, svc.MoveTask(ctx, acc, a.ID, tasks.Done))
	require.NoError(t, svc.MoveTask(ctx, acc, b.ID, tasks.Todo))
	require.NoError(t, svc.MoveTask(ctx, acc, b.ID, tasks.Done))
	// bounce a out and back into DONE — still counts once
	require.NoError(t, svc.MoveTask(ctx, acc, a.ID, tasks.Todo))
	require.NoError(t, svc.MoveTask(ctx, acc, a.ID, tasks.Done))

	otherTask, _ := svc.CreateTask(ctx, other, tasks.TaskInput{Title: "other"})
	require.NoError(t, svc.MoveTask(ctx, other, otherTask.ID, tasks.Done))

	from := time.Now().Add(-time.Hour)
	to := time.Now().Add(time.Hour)

	n, err := svc.DoneCountInRange(ctx, acc, from, to)
	require.NoError(t, err)
	require.Equal(t, 2, n, "a and b, each counted once despite a bouncing in and out")

	// a range that excludes every transition -> 0
	n, err = svc.DoneCountInRange(ctx, acc, from.Add(-48*time.Hour), from.Add(-24*time.Hour))
	require.NoError(t, err)
	require.Equal(t, 0, n)

	// isolation: the caller's count never includes other's task
	n, err = svc.DoneCountInRange(ctx, other, from, to)
	require.NoError(t, err)
	require.Equal(t, 1, n)
}

func TestTaskIsolation(t *testing.T) {
	svc, pool, a := setup(t)
	ctx := context.Background()
	other := newAccount(t, pool, "other@test")

	ta, _ := svc.CreateTask(ctx, a, tasks.TaskInput{Title: "A's task"})
	_, _ = svc.CreateTask(ctx, other, tasks.TaskInput{Title: "B's task"})

	// B's board never shows A's task
	bb, _ := svc.Board(ctx, other)
	for _, col := range bb.Columns {
		for _, tk := range col.Tasks {
			require.NotEqual(t, ta.ID, tk.ID)
		}
	}

	require.ErrorIs(t, svc.UpdateTask(ctx, other, ta.ID, tasks.TaskInput{Title: "hijack"}), tasks.ErrTaskNotFound)
	require.ErrorIs(t, svc.MoveTask(ctx, other, ta.ID, tasks.Done), tasks.ErrTaskNotFound)
	require.ErrorIs(t, svc.DeleteTask(ctx, other, ta.ID), tasks.ErrTaskNotFound)

	// A's task is untouched
	ab, _ := svc.Board(ctx, a)
	require.Equal(t, "A's task", ab.Columns[0].Tasks[0].Title)
}
