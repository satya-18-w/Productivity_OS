package timeline_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/satya-18-w/productivity-os/internal/timeline"
)

func ts(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

func TestAddBlock_PlannedAndActual(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()

	p, err := svc.AddBlock(ctx, acc, timeline.BlockInput{
		Kind: timeline.Planned, StartsAt: ts("2025-06-15T09:00:00Z"), EndsAt: ts("2025-06-15T11:00:00Z"),
	})
	require.NoError(t, err)
	require.Equal(t, timeline.Planned, p.Kind)
	require.Nil(t, p.CategoryID)

	a, err := svc.AddBlock(ctx, acc, timeline.BlockInput{
		Kind: timeline.Actual, StartsAt: ts("2025-06-15T09:30:00Z"), EndsAt: ts("2025-06-15T12:00:00Z"),
	})
	require.NoError(t, err)
	require.Equal(t, timeline.Actual, a.Kind)
	require.NotEqual(t, p.ID, a.ID)
}

func TestAddBlock_MidnightSpanning(t *testing.T) {
	svc, _, acc := setup(t)
	b, err := svc.AddBlock(context.Background(), acc, timeline.BlockInput{
		Kind: timeline.Actual, StartsAt: ts("2025-06-15T23:00:00Z"), EndsAt: ts("2025-06-16T02:00:00Z"),
	})
	require.NoError(t, err)
	require.Equal(t, 3*time.Hour, b.EndsAt.Sub(b.StartsAt))
}

func TestAddBlock_Validation(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()

	_, err := svc.AddBlock(ctx, acc, timeline.BlockInput{
		Kind: "sideways", StartsAt: ts("2025-06-15T09:00:00Z"), EndsAt: ts("2025-06-15T10:00:00Z"),
	})
	requireField(t, err, "kind")

	_, err = svc.AddBlock(ctx, acc, timeline.BlockInput{
		Kind: timeline.Planned, StartsAt: ts("2025-06-15T10:00:00Z"), EndsAt: ts("2025-06-15T09:00:00Z"),
	})
	requireField(t, err, "end")

	_, err = svc.AddBlock(ctx, acc, timeline.BlockInput{
		Kind: timeline.Planned, StartsAt: ts("2025-06-15T10:00:00Z"), EndsAt: ts("2025-06-15T10:00:00Z"),
	})
	requireField(t, err, "end")
}

func TestAddBlock_CategoryMustBeOwnedAndActive(t *testing.T) {
	svc, pool, acc := setup(t)
	ctx := context.Background()
	other := newAccount(t, pool, "other-cat@test")

	mine := mkCategory(t, pool, acc, "Deep Work")
	theirs := mkCategory(t, pool, other, "Theirs")
	archived := mkCategory(t, pool, acc, "Old")
	archiveCategory(t, pool, archived)

	in := func(cat uuid.UUID) timeline.BlockInput {
		return timeline.BlockInput{
			Kind: timeline.Planned, StartsAt: ts("2025-06-15T09:00:00Z"), EndsAt: ts("2025-06-15T10:00:00Z"),
			CategoryID: &cat,
		}
	}

	b, err := svc.AddBlock(ctx, acc, in(mine))
	require.NoError(t, err)
	require.Equal(t, mine, *b.CategoryID)

	_, err = svc.AddBlock(ctx, acc, in(theirs))
	requireField(t, err, "category_id")

	_, err = svc.AddBlock(ctx, acc, in(archived))
	requireField(t, err, "category_id")

	_, err = svc.AddBlock(ctx, acc, in(uuid.New()))
	requireField(t, err, "category_id")
}

func TestAddBlock_TaskLink(t *testing.T) {
	svc, pool, acc := setup(t)
	ctx := context.Background()
	other := newAccount(t, pool, "other-task@test")

	mine := mkTask(t, pool, acc, "Write report")
	theirs := mkTask(t, pool, other, "Theirs")

	b, err := svc.AddBlock(ctx, acc, timeline.BlockInput{
		Kind: timeline.Actual, StartsAt: ts("2025-06-15T09:00:00Z"), EndsAt: ts("2025-06-15T10:00:00Z"),
		TaskID: &mine,
	})
	require.NoError(t, err)
	require.Equal(t, mine, *b.TaskID)
	require.Nil(t, b.CategoryID, "uncategorized task -> no inherited category")

	for _, taskID := range []uuid.UUID{theirs, uuid.New()} {
		_, err := svc.AddBlock(ctx, acc, timeline.BlockInput{
			Kind: timeline.Planned, StartsAt: ts("2025-06-15T09:00:00Z"), EndsAt: ts("2025-06-15T10:00:00Z"),
			TaskID: &taskID,
		})
		requireField(t, err, "task_id")
	}
}

func TestAddBlock_TaskAndCategoryMutuallyExclusive(t *testing.T) {
	svc, pool, acc := setup(t)
	ctx := context.Background()
	task := mkTask(t, pool, acc, "Write report")
	cat := mkCategory(t, pool, acc, "Deep Work")

	_, err := svc.AddBlock(ctx, acc, timeline.BlockInput{
		Kind: timeline.Actual, StartsAt: ts("2025-06-15T09:00:00Z"), EndsAt: ts("2025-06-15T10:00:00Z"),
		TaskID: &task, CategoryID: &cat,
	})
	requireField(t, err, "category_id")
}

func TestEditBlock_TaskLink(t *testing.T) {
	svc, pool, acc := setup(t)
	ctx := context.Background()

	catA := mkCategory(t, pool, acc, "A")
	taskA := mkTaskWithCategory(t, pool, acc, catA, "Task A")
	taskB := mkTask(t, pool, acc, "Task B")

	b, err := svc.AddBlock(ctx, acc, timeline.BlockInput{
		Kind: timeline.Planned, StartsAt: ts("2025-06-15T09:00:00Z"), EndsAt: ts("2025-06-15T10:00:00Z"),
		CategoryID: &catA,
	})
	require.NoError(t, err)

	// link to a task -> the standalone category must be cleared in the same edit
	requireField(t, svc.EditBlock(ctx, acc, b.ID,
		ts("2025-06-15T09:00:00Z"), ts("2025-06-15T10:00:00Z"), &catA, &taskA), "category_id")
	require.NoError(t, svc.EditBlock(ctx, acc, b.ID,
		ts("2025-06-15T09:00:00Z"), ts("2025-06-15T10:00:00Z"), nil, &taskA))

	// re-link to a different task
	require.NoError(t, svc.EditBlock(ctx, acc, b.ID,
		ts("2025-06-15T09:00:00Z"), ts("2025-06-15T10:00:00Z"), nil, &taskB))

	// unlink -> becomes standalone, may then take its own category
	require.NoError(t, svc.EditBlock(ctx, acc, b.ID,
		ts("2025-06-15T09:00:00Z"), ts("2025-06-15T10:00:00Z"), nil, nil))
	require.NoError(t, svc.EditBlock(ctx, acc, b.ID,
		ts("2025-06-15T09:00:00Z"), ts("2025-06-15T10:00:00Z"), &catA, nil))
}

// TestDeleteTask_ClearsBlockLink proves deleting a task clears task_id on its
// linked blocks (ON DELETE SET NULL) without deleting the blocks — MX-TL's
// delete-with-links default (v1.md §7).
func TestDeleteTask_ClearsBlockLink(t *testing.T) {
	svc, pool, acc := setup(t)
	ctx := context.Background()

	task := mkTask(t, pool, acc, "Doomed task")
	b, err := svc.AddBlock(ctx, acc, timeline.BlockInput{
		Kind: timeline.Actual, StartsAt: ts("2025-06-15T09:00:00Z"), EndsAt: ts("2025-06-15T10:00:00Z"),
		TaskID: &task,
	})
	require.NoError(t, err)

	_, delErr := pool.Exec(ctx, `DELETE FROM tasks WHERE id = $1`, task)
	require.NoError(t, delErr)

	all, err := svc.ListAllBlocks(ctx, acc)
	require.NoError(t, err)
	require.Len(t, all, 1)
	require.Equal(t, b.ID, all[0].ID)
	require.Nil(t, all[0].TaskID)
}

func TestEditBlock(t *testing.T) {
	svc, pool, acc := setup(t)
	ctx := context.Background()

	b, _ := svc.AddBlock(ctx, acc, timeline.BlockInput{
		Kind: timeline.Planned, StartsAt: ts("2025-06-15T09:00:00Z"), EndsAt: ts("2025-06-15T10:00:00Z"),
	})
	cat := mkCategory(t, pool, acc, "Reading")

	require.NoError(t, svc.EditBlock(ctx, acc, b.ID, ts("2025-06-15T09:00:00Z"), ts("2025-06-15T11:30:00Z"), &cat, nil))

	requireField(t,
		svc.EditBlock(ctx, acc, b.ID, ts("2025-06-15T11:00:00Z"), ts("2025-06-15T10:00:00Z"), nil, nil), "end")

	require.ErrorIs(t, svc.EditBlock(ctx, acc, uuid.New(),
		ts("2025-06-15T09:00:00Z"), ts("2025-06-15T10:00:00Z"), nil, nil), timeline.ErrBlockNotFound)

	// isolation
	other := newAccount(t, pool, "edit-iso@test")
	require.ErrorIs(t, svc.EditBlock(ctx, other, b.ID,
		ts("2025-06-15T09:00:00Z"), ts("2025-06-15T10:00:00Z"), nil, nil), timeline.ErrBlockNotFound)
}

func TestDeleteBlock(t *testing.T) {
	svc, pool, acc := setup(t)
	ctx := context.Background()

	b, _ := svc.AddBlock(ctx, acc, timeline.BlockInput{
		Kind: timeline.Actual, StartsAt: ts("2025-06-15T09:00:00Z"), EndsAt: ts("2025-06-15T10:00:00Z"),
	})

	other := newAccount(t, pool, "del-iso@test")
	require.ErrorIs(t, svc.DeleteBlock(ctx, other, b.ID), timeline.ErrBlockNotFound)

	require.NoError(t, svc.DeleteBlock(ctx, acc, b.ID))
	require.ErrorIs(t, svc.DeleteBlock(ctx, acc, b.ID), timeline.ErrBlockNotFound)
}

func requireField(t *testing.T, err error, field string) {
	t.Helper()
	var verr *timeline.ValidationError
	require.ErrorAs(t, err, &verr)
	require.Contains(t, verr.Fields, field)
}
