package tasks_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/satya-18-w/productivity-os/internal/platform/pgtest"
	"github.com/satya-18-w/productivity-os/internal/platform/timezone"
	"github.com/satya-18-w/productivity-os/internal/tasks"
)

func setup(t *testing.T) (tasks.Service, *pgxpool.Pool, uuid.UUID) {
	t.Helper()
	pool := pgtest.Pool(t)
	return tasks.NewService(pool), pool, newAccount(t, pool, "owner@test")
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
